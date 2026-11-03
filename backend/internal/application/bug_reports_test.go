package application_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"filippo.io/age"
	"filippo.io/age/armor"

	"railkeeper/backend/internal/application"
)

const validBugReportID = "8b1d6340-0f66-472e-9954-58f1e9d88c88"

func validBugReportInput() application.BugReportInput {
	return application.BugReportInput{
		ReportID:          validBugReportID,
		Area:              application.BugReportAreaVehicles,
		Description:       "Das Vorschaubild verschwindet nach dem Speichern.",
		ExpectedBehavior:  "Das Bild bleibt sichtbar.",
		ReproductionSteps: "Fahrzeug öffnen und speichern.",
		ClientDiagnostics: application.BugReportClientDiagnostics{
			BrowserFamily:  "Firefox",
			BrowserMajor:   "128",
			OSFamily:       "Windows",
			OSMajor:        "11",
			ViewportWidth:  1440,
			ViewportHeight: 900,
			Language:       "de",
			TimeZone:       "Europe/Berlin",
			Theme:          "dark",
			Pathname:       "/vehicles",
			Area:           "vehicles",
			Roles:          []string{"Editor"},
		},
	}
}

func TestBugReportValidationAcceptsSupportedAreas(t *testing.T) {
	cases := map[application.BugReportArea]string{
		application.BugReportAreaOverview:       "area:overview",
		application.BugReportAreaVehicles:       "area:vehicles",
		application.BugReportAreaAccessories:    "area:accessories",
		application.BugReportAreaExhibition:     "area:exhibition",
		application.BugReportAreaImportExport:   "area:import-export",
		application.BugReportAreaDigitalCenters: "area:digital-centers",
		application.BugReportAreaSettings:       "area:settings",
		application.BugReportAreaOther:          "area:other",
	}
	for area, wantLabel := range cases {
		t.Run(string(area), func(t *testing.T) {
			input := validBugReportInput()
			input.Area = area
			input.ClientDiagnostics.Area = string(area)
			input.ClientDiagnostics.Pathname = bugReportTestPathname(area)
			validated, err := application.ValidateBugReportInput(input, false)
			if err != nil {
				t.Fatal(err)
			}
			if validated.AreaLabel != wantLabel {
				t.Fatalf("area label = %q, want %q", validated.AreaLabel, wantLabel)
			}
		})
	}
}

func bugReportTestPathname(area application.BugReportArea) string {
	switch area {
	case application.BugReportAreaImportExport:
		return "/data-transfer"
	case application.BugReportAreaDigitalCenters:
		return "/digital-centers"
	case application.BugReportAreaOther:
		return "/overview"
	default:
		return "/" + string(area)
	}
}

func TestBugReportValidationRejectsInvalidFields(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*application.BugReportInput)
		image  bool
	}{
		{name: "report id", mutate: func(input *application.BugReportInput) { input.ReportID = "not-a-uuid" }},
		{name: "area", mutate: func(input *application.BugReportInput) { input.Area = "unknown" }},
		{name: "short description", mutate: func(input *application.BugReportInput) { input.Description = "zu kurz" }},
		{name: "long description", mutate: func(input *application.BugReportInput) { input.Description = strings.Repeat("x", 10001) }},
		{name: "long expected behavior", mutate: func(input *application.BugReportInput) { input.ExpectedBehavior = strings.Repeat("x", 5001) }},
		{name: "long reproduction", mutate: func(input *application.BugReportInput) { input.ReproductionSteps = strings.Repeat("x", 10001) }},
		{name: "invalid email", mutate: func(input *application.BugReportInput) { input.Email = "not an email" }},
		{name: "long email", mutate: func(input *application.BugReportInput) { input.Email = strings.Repeat("a", 250) + "@example.de" }},
		{name: "image without consent", image: true, mutate: func(input *application.BugReportInput) { input.ScreenshotConsent = false }},
		{name: "too many errors", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.Errors = make([]application.BugReportClientError, 31)
		}},
		{name: "query in pathname", mutate: func(input *application.BugReportInput) { input.ClientDiagnostics.Pathname = "/vehicles?secret=1" }},
		{name: "diagnostic area mismatch", mutate: func(input *application.BugReportInput) { input.ClientDiagnostics.Area = "settings" }},
		{name: "concrete pathname id", mutate: func(input *application.BugReportInput) { input.ClientDiagnostics.Pathname = "/vehicles/123456" }},
		{name: "invalid browser family", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.BrowserFamily = "private-host.example"
		}},
		{name: "invalid browser version", mutate: func(input *application.BugReportInput) { input.ClientDiagnostics.BrowserMajor = "140-secret" }},
		{name: "invalid timestamp", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.Errors = []application.BugReportClientError{{
				Timestamp: "yesterday", Category: "api", ErrorClass: "ApiError", ProblemCode: "request_failed",
				Message: "API request failed", Method: "GET", Route: "/vehicles", Status: 500,
			}}
		}},
		{name: "arbitrary error message", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.Errors = []application.BugReportClientError{{
				Timestamp: "2026-08-29T08:00:00Z", Category: "window", ErrorClass: "Error",
				Message: "Bearer ghp_secret host.internal:8080 /etc/passwd",
			}}
		}},
		{name: "unsafe problem code", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.Errors = []application.BugReportClientError{{
				Timestamp: "2026-08-29T08:00:00Z", Category: "api", ErrorClass: "ApiError",
				ProblemCode: "bearer_secret_token", Message: "API request failed", Method: "GET", Route: "/vehicles",
			}}
		}},
		{name: "unsafe route", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.Errors = []application.BugReportClientError{{
				Timestamp: "2026-08-29T08:00:00Z", Category: "network", ErrorClass: "TypeError",
				ProblemCode: "network_error", Message: "Network request failed", Method: "GET",
				Route: "/etc/passwd",
			}}
		}},
		{name: "invalid status", mutate: func(input *application.BugReportInput) {
			input.ClientDiagnostics.Errors = []application.BugReportClientError{{
				Timestamp: "2026-08-29T08:00:00Z", Category: "api", ErrorClass: "ApiError",
				ProblemCode: "request_failed", Message: "API request failed", Method: "GET",
				Route: "/vehicles", Status: 700,
			}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := validBugReportInput()
			test.mutate(&input)
			if _, err := application.ValidateBugReportInput(input, test.image); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestBugReportRenderingIsBilingualDeterministicAndSafe(t *testing.T) {
	input := validBugReportInput()
	input.Description = "  Vorschaubild\nverschwindet nach dem Speichern. Zweiter Satz.  "
	input.Email = "maintainer@example.de"
	input.ScreenshotConsent = true
	input.ClientDiagnostics.Errors = []application.BugReportClientError{{
		Timestamp: "2026-08-29T08:00:00Z", Category: "api", ErrorClass: "ApiError",
		ProblemCode: "request_failed", Message: "API request failed", Method: "GET",
		Route: "/vehicles", Status: 500,
	}}
	validated, err := application.ValidateBugReportInput(input, true)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := application.BuildBugReportIssue(application.BugReportIssueInput{
		Report: validated,
		ServerDiagnostics: application.BugReportServerDiagnostics{
			RailKeeperVersion: "0.1.20", DeploymentType: "container", GoVersion: "go1.26.0",
			OperatingSystem: "linux", Architecture: "amd64", SchemaVersion: 67,
			EnabledDigitalProviderTypes: []string{"ecos"},
		},
		EncryptedContact: "-----BEGIN AGE ENCRYPTED FILE-----\nsecret\n-----END AGE ENCRYPTED FILE-----",
		BaseLabels:       []string{"bug", "source:in-app", "bug"},
		HasScreenshot:    true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if draft.Title != "[Bug][Fahrzeuge] Vorschaubild verschwindet nach dem Speichern" {
		t.Fatalf("title = %q", draft.Title)
	}
	if strings.Join(draft.Labels, ",") != "bug,source:in-app,area:vehicles" {
		t.Fatalf("labels = %#v", draft.Labels)
	}
	for _, required := range []string{
		"Fehlerbeschreibung / Bug description",
		"Erwartetes Verhalten / Expected behavior",
		"Schritte zur Reproduktion / Reproduction steps",
		"Automatisch erfasste Diagnosedaten / Automatically collected diagnostics",
		"Verschlüsselter Kontakt / Encrypted contact",
		"Öffentlicher Screenshot / Public screenshot",
		"Privat melden / Report privately",
		"<!-- railkeeper-bug-report:v1:" + validBugReportID + " -->",
	} {
		if !strings.Contains(draft.Body, required) {
			t.Fatalf("body does not contain %q:\n%s", required, draft.Body)
		}
	}
	if strings.Contains(draft.Body, input.Email) {
		t.Fatal("plaintext email leaked into issue body")
	}
	if !strings.Contains(draft.Body, `"schemaVersion": 67`) {
		t.Fatalf("server diagnostics missing from body:\n%s", draft.Body)
	}
}

func TestBugReportRenderingBoundsUnicodeTitle(t *testing.T) {
	input := validBugReportInput()
	input.Description = strings.Repeat("🚂", 200)
	validated, err := application.ValidateBugReportInput(input, false)
	if err != nil {
		t.Fatal(err)
	}
	draft, err := application.BuildBugReportIssue(application.BugReportIssueInput{Report: validated})
	if err != nil {
		t.Fatal(err)
	}
	if got := len([]rune(draft.Title)); got > 120 {
		t.Fatalf("title has %d code points", got)
	}
}

func TestBugReportContactEncryptionRoundTripsWithConfiguredRecipient(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	encryptor, err := application.NewAgeContactEncryptor(identity.Recipient().String())
	if err != nil {
		t.Fatal(err)
	}
	ciphertext, err := encryptor.Encrypt("maintainer@example.de")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(ciphertext, "-----BEGIN AGE ENCRYPTED FILE-----") {
		t.Fatalf("ciphertext is not armored: %q", ciphertext)
	}
	armorReader := armor.NewReader(strings.NewReader(ciphertext))
	reader, err := age.Decrypt(armorReader, identity)
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plaintext, []byte("maintainer@example.de")) {
		t.Fatalf("plaintext = %q", plaintext)
	}
}

func TestBugReportContactEncryptionRejectsInvalidRecipient(t *testing.T) {
	if _, err := application.NewAgeContactEncryptor("AGE-SECRET-KEY-DO-NOT-ACCEPT"); err == nil {
		t.Fatal("expected invalid recipient error")
	}
}

func TestBugReportServiceCreatesOneIssueAcrossIdempotentRetries(t *testing.T) {
	repository := newMemoryBugReportRepository()
	issues := &fakeBugReportIssueClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: repository, IssueClient: issues, BaseLabels: []string{"bug", "source:in-app"},
	})
	request := application.BugReportSubmissionRequest{Input: validBugReportInput()}

	first, err := service.Submit(t.Context(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := service.Submit(t.Context(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Status != application.BugReportComplete || second.Status != application.BugReportComplete ||
		first.IssueNumber != 1 || second.IssueNumber != 1 || issues.createCalls != 1 {
		t.Fatalf("first=%#v second=%#v createCalls=%d", first, second, issues.createCalls)
	}
}

func TestBugReportServiceUploadsScreenshotAndAddsOnePublicComment(t *testing.T) {
	issues := &fakeBugReportIssueClient{}
	assets := &fakeBugReportAssetClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: newMemoryBugReportRepository(), IssueClient: issues, AssetClient: assets,
		BaseLabels: []string{"bug", "source:in-app"},
	})
	input := validBugReportInput()
	input.ScreenshotConsent = true
	request := application.BugReportSubmissionRequest{
		Input: input, ScreenshotBytes: encodedPNG(t, image.NewRGBA(image.Rect(0, 0, 2, 2))),
		ScreenshotMediaType: "image/png",
	}
	result, err := service.Submit(t.Context(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != application.BugReportComplete || len(assets.createdPaths) != 1 || len(issues.comments) != 1 {
		t.Fatalf("result=%#v paths=%#v comments=%#v", result, assets.createdPaths, issues.comments)
	}
	if !strings.HasSuffix(assets.createdPaths[0], "/"+validBugReportID+".png") ||
		!strings.Contains(issues.comments[0], "railkeeper-screenshot:v1:"+validBugReportID) ||
		!strings.Contains(issues.comments[0], "https://raw.githubusercontent.com/") {
		t.Fatalf("unexpected screenshot side effects: paths=%#v comments=%#v", assets.createdPaths, issues.comments)
	}
}

func TestBugReportServiceContinuesPartialAssetFailureWithoutDuplicateIssue(t *testing.T) {
	repository := newMemoryBugReportRepository()
	issues := &fakeBugReportIssueClient{}
	assets := &fakeBugReportAssetClient{createErr: errors.New("upstream failed")}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: repository, IssueClient: issues, AssetClient: assets,
	})
	input := validBugReportInput()
	input.ScreenshotConsent = true
	request := application.BugReportSubmissionRequest{
		Input: input, ScreenshotBytes: encodedPNG(t, image.NewRGBA(image.Rect(0, 0, 2, 2))),
		ScreenshotMediaType: "image/png",
	}
	partial, err := service.Submit(t.Context(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if partial.Status != application.BugReportPartial || partial.IssueNumber != 1 {
		t.Fatalf("partial result = %#v", partial)
	}
	assets.createErr = nil
	complete, err := service.Submit(t.Context(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if complete.Status != application.BugReportComplete || issues.createCalls != 1 || len(issues.comments) != 1 {
		t.Fatalf("complete=%#v issue calls=%d comments=%d", complete, issues.createCalls, len(issues.comments))
	}
}

func TestBugReportServiceReconcilesIssueAfterUncertainCreateResponse(t *testing.T) {
	repository := newMemoryBugReportRepository()
	issues := &fakeBugReportIssueClient{uncertainCreate: true}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: repository, IssueClient: issues,
	})
	request := application.BugReportSubmissionRequest{Input: validBugReportInput()}
	if _, err := service.Submit(t.Context(), "user-1", request); !errors.Is(err, application.ErrBugReportGitHubUnavailable) {
		t.Fatalf("first submission error = %v", err)
	}
	issues.uncertainCreate = false
	result, err := service.Submit(t.Context(), "user-1", request)
	if err != nil {
		t.Fatal(err)
	}
	if result.IssueNumber != 1 || issues.createCalls != 1 {
		t.Fatalf("result=%#v createCalls=%d", result, issues.createCalls)
	}
}

func TestBugReportServiceRejectsRetryByAnotherUser(t *testing.T) {
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: newMemoryBugReportRepository(), IssueClient: &fakeBugReportIssueClient{},
	})
	request := application.BugReportSubmissionRequest{Input: validBugReportInput()}
	if _, err := service.Submit(t.Context(), "user-1", request); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Submit(t.Context(), "user-2", request); !errors.Is(err, application.ErrBugReportStateConflict) {
		t.Fatalf("other-user retry error = %v", err)
	}
}

func TestBugReportServiceSerializesConcurrentCallsForReportID(t *testing.T) {
	issues := &fakeBugReportIssueClient{createDelay: 20 * time.Millisecond}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: newMemoryBugReportRepository(), IssueClient: issues,
	})
	requests := []application.BugReportSubmissionRequest{
		{Input: validBugReportInput()},
		{Input: validBugReportInput()},
	}
	requests[1].Input.ReportID = "  " + validBugReportID + "  "
	var wait sync.WaitGroup
	errorsSeen := make(chan error, 2)
	for _, request := range requests {
		wait.Add(1)
		go func(request application.BugReportSubmissionRequest) {
			defer wait.Done()
			_, err := service.Submit(t.Context(), "user-1", request)
			errorsSeen <- err
		}(request)
	}
	wait.Wait()
	close(errorsSeen)
	for err := range errorsSeen {
		if err != nil {
			t.Fatal(err)
		}
	}
	if issues.createCalls != 1 {
		t.Fatalf("issue create calls = %d", issues.createCalls)
	}
}

type memoryBugReportRepository struct {
	mu          sync.Mutex
	submissions map[string]application.BugReportSubmission
}

func newMemoryBugReportRepository() *memoryBugReportRepository {
	return &memoryBugReportRepository{submissions: map[string]application.BugReportSubmission{}}
}

func (repository *memoryBugReportRepository) Reserve(
	_ context.Context,
	reportID string,
	actorUserID string,
	_ application.BugReportAdmissionPolicy,
) (bool, bool, application.BugReportSubmission, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if existing, ok := repository.submissions[reportID]; ok {
		return false, true, existing, nil
	}
	now := time.Now().UTC()
	submission := application.BugReportSubmission{
		ReportID: reportID, ActorUserID: actorUserID, State: application.BugReportReserved,
		CreatedAt: now, UpdatedAt: now,
	}
	repository.submissions[reportID] = submission
	return true, true, submission, nil
}

func (repository *memoryBugReportRepository) Get(
	_ context.Context,
	reportID string,
) (application.BugReportSubmission, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	submission, ok := repository.submissions[reportID]
	if !ok {
		return application.BugReportSubmission{}, errors.New("not found")
	}
	return submission, nil
}

func (repository *memoryBugReportRepository) Update(
	_ context.Context,
	reportID string,
	update application.BugReportSubmissionUpdate,
) (application.BugReportSubmission, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	submission := repository.submissions[reportID]
	submission.State = update.State
	submission.GitHubIssueNumber = update.GitHubIssueNumber
	submission.GitHubIssueURL = update.GitHubIssueURL
	submission.AssetURL = update.AssetURL
	submission.AssetSHA = update.AssetSHA
	submission.ErrorCategory = update.ErrorCategory
	submission.UpdatedAt = time.Now().UTC()
	repository.submissions[reportID] = submission
	return submission, nil
}

type fakeBugReportIssueClient struct {
	mu              sync.Mutex
	createCalls     int
	createDelay     time.Duration
	uncertainCreate bool
	issues          []application.GitHubBugReportIssue
	comments        []string
}

func (client *fakeBugReportIssueClient) CreateIssue(
	_ context.Context,
	issue application.BugReportIssue,
) (application.GitHubBugReportIssue, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.createCalls++
	if client.createDelay > 0 {
		time.Sleep(client.createDelay)
	}
	created := application.GitHubBugReportIssue{
		Number: int64(len(client.issues) + 1), URL: "https://github.com/issues/1", Body: issue.Body,
	}
	client.issues = append(client.issues, created)
	if client.uncertainCreate {
		return application.GitHubBugReportIssue{}, errors.New("connection reset")
	}
	return created, nil
}

func (client *fakeBugReportIssueClient) ListRecentIssues(
	_ context.Context,
	_ time.Time,
	_ int,
) ([]application.GitHubBugReportIssue, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	return append([]application.GitHubBugReportIssue{}, client.issues...), nil
}

func (client *fakeBugReportIssueClient) ListIssueComments(
	_ context.Context,
	_ int64,
	_ int,
) ([]application.GitHubBugReportComment, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	result := make([]application.GitHubBugReportComment, 0, len(client.comments))
	for index, body := range client.comments {
		result = append(result, application.GitHubBugReportComment{ID: int64(index + 1), Body: body})
	}
	return result, nil
}

func (client *fakeBugReportIssueClient) CreateIssueComment(
	_ context.Context,
	_ int64,
	body string,
) (application.GitHubBugReportComment, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	client.comments = append(client.comments, body)
	return application.GitHubBugReportComment{ID: int64(len(client.comments)), Body: body}, nil
}

type fakeBugReportAssetClient struct {
	mu           sync.Mutex
	createErr    error
	createdPaths []string
	assets       map[string]application.GitHubBugReportAsset
}

func (client *fakeBugReportAssetClient) CreateFile(
	_ context.Context,
	filePath string,
	_ string,
	_ []byte,
) (application.GitHubBugReportAsset, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.createErr != nil {
		return application.GitHubBugReportAsset{}, client.createErr
	}
	client.createdPaths = append(client.createdPaths, filePath)
	asset := application.GitHubBugReportAsset{
		BlobSHA: "blob", CommitSHA: "commit",
		URL: "https://raw.githubusercontent.com/o/assets/commit/" + filePath,
	}
	if client.assets == nil {
		client.assets = map[string]application.GitHubBugReportAsset{}
	}
	client.assets[filePath] = asset
	return asset, nil
}

func (client *fakeBugReportAssetClient) FindFile(
	_ context.Context,
	filePath string,
) (bool, application.GitHubBugReportAsset, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	asset, ok := client.assets[filePath]
	return ok, asset, nil
}
