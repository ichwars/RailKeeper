package api

import (
	"context"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestBugReportDiagnosticsAreCollectedOnlyWhenRequested(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	_, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "report-diagnostics", Password: "diagnostics-password", Roles: []string{"Messe"},
	})
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	service := application.NewBugReportService(application.BugReportServiceConfig{
		CollectDiagnostics: func(context.Context) (application.BugReportServerDiagnostics, error) {
			calls++
			return application.BugReportServerDiagnostics{SchemaVersion: 72}, nil
		},
	})
	router := NewRouter(Config{AuthService: auth, BugReportService: service})
	session := loginRouteTestUser(t, auth, "report-diagnostics", "diagnostics-password")
	for _, query := range []string{"", "?diagnostics=false", "?diagnostics=true"} {
		response := authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/bug-reports/config"+query, nil, true)
		assertStatus(t, response, http.StatusOK)
		var capability BugReportCapability
		decodeResponse(t, response, &capability)
		if query == "?diagnostics=true" {
			if calls != 1 || capability.ServerDiagnostics == nil || capability.ServerDiagnostics.SchemaVersion != 72 {
				t.Fatalf("explicit preview did not collect diagnostics: calls=%d", calls)
			}
		} else if calls != 0 || capability.ServerDiagnostics != nil {
			t.Fatalf("capability request collected or disclosed diagnostics: calls=%d", calls)
		}
	}
}

func TestBugReportMultipartOptOutDiscardsMalformedDiagnostics(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	_, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "report-optout", Password: "optout-password", Roles: []string{"Viewer"},
	})
	if err != nil {
		t.Fatal(err)
	}
	issues := &apiFakeBugReportIssueClient{}
	service := application.NewBugReportService(application.BugReportServiceConfig{
		Repository: infrastructure.NewBugReportRepository(db), IssueClient: issues,
		CollectDiagnostics: func(context.Context) (application.BugReportServerDiagnostics, error) {
			t.Fatal("opted-out HTTP report triggered collection")
			return application.BugReportServerDiagnostics{}, nil
		},
	})
	router := NewRouter(Config{AuthService: auth, BugReportService: service,
		RateLimitService: application.NewRateLimitService(db), BugReportCapability: BugReportCapability{Enabled: true}})
	session := loginRouteTestUser(t, auth, "report-optout", "optout-password")
	fields := validBugReportMultipartFields("8b1d6340-0f66-472e-9954-58f1e9d88c88")
	fields["includeDiagnostics"] = "false"
	fields["clientDiagnostics"] = "malformed private input"
	assertStatus(t, bugReportMultipartRequest(t, router, session, fields, nil, true), http.StatusCreated)
	if issues.createCalls != 1 {
		t.Fatal("the user-authored report must still be submitted")
	}
}
