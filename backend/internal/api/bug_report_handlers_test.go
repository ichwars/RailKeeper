package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestBugReportConfigReturnsExactSafeCapability(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "viewer-report", Password: "viewer-password", Roles: []string{"Viewer"},
	}); err != nil {
		t.Fatal(err)
	}
	capability := BugReportCapability{
		Enabled: true, FallbackURL: "https://github.com/ichwars/RailKeeper/issues/new",
		ContactAvailable: true,
		ServerDiagnostics: &application.BugReportServerDiagnostics{
			RailKeeperVersion: "0.1.20", DeploymentType: "container", GoVersion: "go1.26.0",
			OperatingSystem: "linux", Architecture: "amd64", SchemaVersion: 67,
			EnabledDigitalProviderTypes: []string{"ecos"},
		},
	}
	router := NewRouter(Config{AuthService: auth, BugReportCapability: capability})
	session := loginRouteTestUser(t, auth, "viewer-report", "viewer-password")
	response := authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/bug-reports/config", nil, true)
	assertStatus(t, response, http.StatusOK)
	var body BugReportCapability
	decodeResponse(t, response, &body)
	if !body.Enabled || !body.ContactAvailable || body.MaxScreenshotBytes != application.MaxBugReportScreenshotBytes ||
		body.ServerDiagnostics != nil {
		t.Fatalf("unexpected capability: %#v", body)
	}
	encoded := response.Body.String()
	for _, forbidden := range []string{"token", "repository", "recipient", "hostname", "port"} {
		if strings.Contains(strings.ToLower(encoded), forbidden) {
			t.Fatalf("capability exposed forbidden field %q: %s", forbidden, encoded)
		}
	}
}

func TestBugReportConfigUsesCurrentCapabilityProvider(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "viewer-capability", Password: "viewer-password", Roles: []string{"Viewer"},
	}); err != nil {
		t.Fatal(err)
	}
	enabled := false
	router := NewRouter(Config{
		AuthService: auth,
		BugReportCapabilityControl: BugReportCapabilityControl{
			Provider: func(context.Context) BugReportCapability {
				return BugReportCapability{Enabled: enabled}
			},
		},
	})
	session := loginRouteTestUser(t, auth, "viewer-capability", "viewer-password")

	for _, wantEnabled := range []bool{false, true} {
		enabled = wantEnabled
		response := authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/bug-reports/config", nil, true)
		assertStatus(t, response, http.StatusOK)
		var body BugReportCapability
		decodeResponse(t, response, &body)
		if body.Enabled != wantEnabled {
			t.Fatalf("enabled = %t, want %t", body.Enabled, wantEnabled)
		}
	}
}

func TestBugReportSubmissionCreatesThenReusesIssue(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "editor-report", Password: "editor-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	issues := &apiFakeBugReportIssueClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: infrastructure.NewBugReportRepository(db), IssueClient: issues,
	})
	router := NewRouter(Config{
		AuthService: auth, BugReportService: service, RateLimitService: application.NewRateLimitService(db),
		BugReportCapability: BugReportCapability{Enabled: true},
	})
	session := loginRouteTestUser(t, auth, "editor-report", "editor-password")
	fields := validBugReportMultipartFields("8b1d6340-0f66-472e-9954-58f1e9d88c88")
	first := bugReportMultipartRequest(t, router, session, fields, nil, true)
	assertStatus(t, first, http.StatusCreated)
	second := bugReportMultipartRequest(t, router, session, fields, nil, true)
	assertStatus(t, second, http.StatusOK)
	if issues.createCalls != 1 {
		t.Fatalf("issue create calls = %d", issues.createCalls)
	}
	var result application.BugReportSubmissionResult
	decodeResponse(t, second, &result)
	if result.Status != application.BugReportComplete || result.IssueNumber != 1 {
		t.Fatalf("unexpected retry result: %#v", result)
	}
}

func TestBugReportSubmissionRequiresCSRFAndValidBoundedMultipart(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "viewer-bounds", Password: "viewer-password", Roles: []string{"Viewer"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: infrastructure.NewBugReportRepository(db), IssueClient: &apiFakeBugReportIssueClient{},
	})
	router := NewRouter(Config{AuthService: auth, BugReportService: service, BugReportCapability: BugReportCapability{Enabled: true}})
	session := loginRouteTestUser(t, auth, "viewer-bounds", "viewer-password")
	fields := validBugReportMultipartFields("8b1d6340-0f66-472e-9954-58f1e9d88c89")
	assertProblem(t, bugReportMultipartRequest(t, router, session, fields, nil, false),
		http.StatusForbidden, "csrf_required")
	fields["description"] = strings.Repeat("x", 10_001)
	assertProblem(t, bugReportMultipartRequest(t, router, session, fields, nil, true),
		http.StatusBadRequest, "bug_report_validation_failed")
}

func TestBugReportSubmissionLimitsNewReportIDsPerUser(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "report-rate", Password: "viewer-password", Roles: []string{"Viewer"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: infrastructure.NewBugReportRepository(db), IssueClient: &apiFakeBugReportIssueClient{},
	})
	router := NewRouter(Config{
		AuthService: auth, BugReportService: service, RateLimitService: application.NewRateLimitService(db),
		BugReportCapability: BugReportCapability{Enabled: true},
	})
	session := loginRouteTestUser(t, auth, "report-rate", "viewer-password")
	for index := 0; index < 3; index++ {
		fields := validBugReportMultipartFields(fmt.Sprintf("8b1d6340-0f66-472e-9954-58f1e9d88c8%d", index))
		fields["description"] = "zu kurz"
		assertProblem(t, bugReportMultipartRequest(t, router, session, fields, nil, true),
			http.StatusBadRequest, "bug_report_validation_failed")
	}
	for index := 0; index < 6; index++ {
		reportID := fmt.Sprintf("8b1d6340-0f66-472e-9954-58f1e9d88c9%d", index)
		response := bugReportMultipartRequest(
			t, router, session, validBugReportMultipartFields(reportID), nil, true,
		)
		if index < 5 {
			assertStatus(t, response, http.StatusCreated)
		} else {
			assertProblem(t, response, http.StatusTooManyRequests, "bug_report_rate_limited")
			if response.Header().Get("Retry-After") != "3600" {
				t.Fatalf("Retry-After = %q", response.Header().Get("Retry-After"))
			}
		}
	}
	var submissions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM bug_report_submissions`).Scan(&submissions); err != nil {
		t.Fatal(err)
	}
	if submissions != 5 {
		t.Fatalf("submission rows = %d, want exactly 5 admitted rows", submissions)
	}
}

func TestBugReportSubmissionLimitsAttemptsBeforeValidation(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "report-attempt-rate", Password: "viewer-password", Roles: []string{"Viewer"},
	}); err != nil {
		t.Fatal(err)
	}
	issues := &apiFakeBugReportIssueClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: infrastructure.NewBugReportRepository(db), IssueClient: issues,
	})
	router := NewRouter(Config{
		AuthService: auth, BugReportService: service, RateLimitService: application.NewRateLimitService(db),
		BugReportCapability: BugReportCapability{Enabled: true},
	})
	session := loginRouteTestUser(t, auth, "report-attempt-rate", "viewer-password")
	for index := 0; index < bugReportAttemptLimit; index++ {
		fields := validBugReportMultipartFields(fmt.Sprintf("8b1d6340-0f66-472e-9954-58f1e9d88d%02x", index))
		fields["description"] = "zu kurz"
		assertProblem(t, bugReportMultipartRequest(t, router, session, fields, nil, true),
			http.StatusBadRequest, "bug_report_validation_failed")
	}
	response := bugReportMultipartRequest(
		t, router, session, validBugReportMultipartFields("8b1d6340-0f66-472e-9954-58f1e9d88dff"), nil, true,
	)
	assertProblem(t, response, http.StatusTooManyRequests, "bug_report_rate_limited")
	if response.Header().Get("Retry-After") != "3600" {
		t.Fatalf("Retry-After = %q", response.Header().Get("Retry-After"))
	}
	if issues.createCalls != 0 {
		t.Fatalf("issue create calls = %d, want 0", issues.createCalls)
	}
	var submissions int
	if err := db.QueryRow(`SELECT COUNT(*) FROM bug_report_submissions`).Scan(&submissions); err != nil {
		t.Fatal(err)
	}
	if submissions != 0 {
		t.Fatalf("submission rows = %d, want 0", submissions)
	}
}

func validBugReportMultipartFields(reportID string) map[string]string {
	diagnostics, _ := json.Marshal(application.BugReportClientDiagnostics{
		BrowserFamily: "Firefox", BrowserMajor: "128", OSFamily: "Windows", OSMajor: "11",
		ViewportWidth: 1440, ViewportHeight: 900, Language: "de", TimeZone: "Europe/Berlin",
		Theme: "dark", Pathname: "/vehicles", Area: "vehicles", Roles: []string{"Viewer"},
		Errors: []application.BugReportClientError{},
	})
	return map[string]string{
		"reportId": reportID, "area": "vehicles",
		"description":       "Das Vorschaubild verschwindet nach dem Speichern.",
		"clientDiagnostics": string(diagnostics),
	}
}

func bugReportMultipartRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	fields map[string]string,
	screenshot []byte,
	withCSRF bool,
) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range fields {
		if err := writer.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	if screenshot != nil {
		part, err := writer.CreateFormFile("screenshot", "screenshot.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(screenshot); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/bug-reports", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
	if withCSRF {
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

type apiFakeBugReportIssueClient struct {
	mu          sync.Mutex
	createCalls int
	issues      []application.GitHubBugReportIssue
}

func (client *apiFakeBugReportIssueClient) CreateIssue(
	_ context.Context,
	issue application.BugReportIssue,
) (application.GitHubBugReportIssue, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.createCalls++
	created := application.GitHubBugReportIssue{
		Number: int64(client.createCalls), URL: fmt.Sprintf("https://github.com/issues/%d", client.createCalls),
		Body: issue.Body,
	}
	client.issues = append(client.issues, created)
	return created, nil
}

func (client *apiFakeBugReportIssueClient) ListRecentIssues(
	_ context.Context,
	_ time.Time,
	_ int,
) ([]application.GitHubBugReportIssue, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	return append([]application.GitHubBugReportIssue{}, client.issues...), nil
}

func (client *apiFakeBugReportIssueClient) ListIssueComments(
	_ context.Context,
	_ int64,
	_ int,
) ([]application.GitHubBugReportComment, error) {
	return nil, nil
}

func (client *apiFakeBugReportIssueClient) CreateIssueComment(
	_ context.Context,
	_ int64,
	body string,
) (application.GitHubBugReportComment, error) {
	return application.GitHubBugReportComment{ID: 1, Body: body}, nil
}
