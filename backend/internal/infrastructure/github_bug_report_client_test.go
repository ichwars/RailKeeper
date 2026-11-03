package infrastructure_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestGitHubBugReportIssueClientCreatesLabeledIssueWithScopedHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/repos/ichwars/RailKeeper/issues" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		assertGitHubBugReportHeaders(t, request, "issue-token")
		var payload struct {
			Title  string   `json:"title"`
			Body   string   `json:"body"`
			Labels []string `json:"labels"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		if payload.Title != "[Bug][Fahrzeuge] Bild fehlt" || strings.Join(payload.Labels, ",") != "bug,area:vehicles" {
			t.Fatalf("unexpected issue payload: %#v", payload)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"number":184,"html_url":"https://github.com/ichwars/RailKeeper/issues/184","body":"marker"}`)
	}))
	defer server.Close()

	client, err := infrastructure.NewGitHubBugReportIssueClient(infrastructure.GitHubBugReportClientConfig{
		APIBaseURL: server.URL, Token: "issue-token", Repository: "ichwars/RailKeeper", HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	issue, err := client.CreateIssue(t.Context(), application.BugReportIssue{
		Title: "[Bug][Fahrzeuge] Bild fehlt", Body: "body", Labels: []string{"bug", "area:vehicles"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if issue.Number != 184 || issue.URL != "https://github.com/ichwars/RailKeeper/issues/184" {
		t.Fatalf("unexpected issue: %#v", issue)
	}
}

func TestGitHubBugReportIssueClientListsMarkersAndCreatesComment(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requests++
		assertGitHubBugReportHeaders(t, request, "issue-token")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/o/r/issues":
			if request.URL.Query().Get("per_page") != "30" || request.URL.Query().Get("since") == "" {
				t.Fatalf("unbounded issue query: %s", request.URL.RawQuery)
			}
			_, _ = io.WriteString(response, `[{"number":7,"html_url":"https://example/7","body":"<!-- marker -->"}]`)
		case request.Method == http.MethodGet && request.URL.Path == "/repos/o/r/issues/7/comments":
			_, _ = io.WriteString(response, `[{"id":9,"body":"<!-- screenshot-marker -->"}]`)
		case request.Method == http.MethodPost && request.URL.Path == "/repos/o/r/issues/7/comments":
			_, _ = io.WriteString(response, `{"id":10,"body":"created"}`)
		default:
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.String())
		}
	}))
	defer server.Close()
	client := newTestIssueClient(t, server, "o/r")

	issues, err := client.ListRecentIssues(t.Context(), time.Date(2026, 8, 29, 8, 0, 0, 0, time.UTC), 30)
	if err != nil || len(issues) != 1 || !strings.Contains(issues[0].Body, "marker") {
		t.Fatalf("issues=%#v err=%v", issues, err)
	}
	comments, err := client.ListIssueComments(t.Context(), 7, 100)
	if err != nil || len(comments) != 1 || comments[0].ID != 9 {
		t.Fatalf("comments=%#v err=%v", comments, err)
	}
	if _, err := client.CreateIssueComment(t.Context(), 7, "created"); err != nil {
		t.Fatal(err)
	}
	if requests != 3 {
		t.Fatalf("request count = %d", requests)
	}
}

func TestGitHubBugReportIssueClientPaginatesReconciliationBeyondFirstPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assertGitHubBugReportHeaders(t, request, "issue-token")
		page := request.URL.Query().Get("page")
		response.Header().Set("Content-Type", "application/json")
		switch request.URL.Path {
		case "/repos/o/r/issues":
			if request.URL.Query().Get("per_page") != "100" {
				t.Fatalf("issue page is not bounded: %s", request.URL.RawQuery)
			}
			if page == "1" {
				items := make([]map[string]any, 100)
				for index := range items {
					items[index] = map[string]any{"number": index + 1, "pull_request": map[string]any{"url": "pr"}}
				}
				_ = json.NewEncoder(response).Encode(items)
				return
			}
			_, _ = io.WriteString(response, `[{"number":184,"html_url":"https://example/184","body":"second-page-marker"}]`)
		case "/repos/o/r/issues/184/comments":
			if request.URL.Query().Get("per_page") != "100" {
				t.Fatalf("comment page is not bounded: %s", request.URL.RawQuery)
			}
			if page == "1" {
				items := make([]map[string]any, 100)
				for index := range items {
					items[index] = map[string]any{"id": index + 1, "body": "ordinary"}
				}
				_ = json.NewEncoder(response).Encode(items)
				return
			}
			_, _ = io.WriteString(response, `[{"id":184,"body":"second-page-comment-marker"}]`)
		default:
			t.Fatalf("unexpected request: %s", request.URL.String())
		}
	}))
	defer server.Close()
	client := newTestIssueClient(t, server, "o/r")

	issues, err := client.ListRecentIssues(t.Context(), time.Date(2026, 8, 29, 8, 0, 0, 0, time.UTC), 200)
	if err != nil || len(issues) != 1 || !strings.Contains(issues[0].Body, "second-page-marker") {
		t.Fatalf("issues=%#v err=%v", issues, err)
	}
	comments, err := client.ListIssueComments(t.Context(), 184, 200)
	if err != nil || len(comments) != 101 || !strings.Contains(comments[100].Body, "second-page-comment-marker") {
		t.Fatalf("comments=%d err=%v", len(comments), err)
	}
}

func TestGitHubBugReportAssetClientUploadsBase64AndReturnsImmutableURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPut ||
			request.URL.Path != "/repos/ichwars/railkeeper-bug-assets/contents/reports/2026/08/report.png" {
			t.Fatalf("unexpected request: %s %s", request.Method, request.URL.Path)
		}
		assertGitHubBugReportHeaders(t, request, "asset-token")
		var payload struct {
			Message string `json:"message"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		decoded, err := base64.StdEncoding.DecodeString(payload.Content)
		if err != nil || string(decoded) != "safe-image" {
			t.Fatalf("uploaded content=%q err=%v", decoded, err)
		}
		_, _ = io.WriteString(response, `{"content":{"sha":"blob-sha"},"commit":{"sha":"commit-sha"}}`)
	}))
	defer server.Close()
	client, err := infrastructure.NewGitHubBugReportAssetClient(infrastructure.GitHubBugReportClientConfig{
		APIBaseURL: server.URL, Token: "asset-token", Repository: "ichwars/railkeeper-bug-assets",
		HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	asset, err := client.CreateFile(t.Context(), "reports/2026/08/report.png", "Add report screenshot", []byte("safe-image"))
	if err != nil {
		t.Fatal(err)
	}
	if asset.BlobSHA != "blob-sha" || asset.CommitSHA != "commit-sha" ||
		asset.URL != "https://raw.githubusercontent.com/ichwars/railkeeper-bug-assets/commit-sha/reports/2026/08/report.png" {
		t.Fatalf("unexpected asset: %#v", asset)
	}
}

func TestGitHubBugReportCapabilityRequiresLabelsAndPublicSeparateAssetRepository(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := "issue-token"
		if strings.HasPrefix(request.URL.Path, "/repos/o/assets") {
			token = "asset-token"
		}
		assertGitHubBugReportHeaders(t, request, token)
		switch request.URL.Path {
		case "/repos/o/issues":
			_, _ = io.WriteString(response, `{"private":true,"default_branch":"main"}`)
		case "/repos/o/issues/labels":
			_, _ = io.WriteString(response, `[{"name":"bug"},{"name":"source:in-app"},{"name":"area:vehicles"}]`)
		case "/repos/o/assets":
			_, _ = io.WriteString(response, `{"private":false,"default_branch":"main"}`)
		default:
			t.Fatalf("unexpected capability request: %s", request.URL.Path)
		}
	}))
	defer server.Close()
	issue := newTestIssueClient(t, server, "o/issues")
	asset, err := infrastructure.NewGitHubBugReportAssetClient(infrastructure.GitHubBugReportClientConfig{
		APIBaseURL: server.URL, Token: "asset-token", Repository: "o/assets", HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.ValidateGitHubBugReportCapability(
		t.Context(), issue, asset, []string{"bug", "source:in-app", "area:vehicles"},
	); err != nil {
		t.Fatal(err)
	}
	if err := infrastructure.ValidateGitHubBugReportRepositoryTargets("o/same", "O/SAME"); err == nil {
		t.Fatal("expected equal target repositories to be rejected")
	}
}

func TestGitHubBugReportClientErrorsNeverExposeTokenOrResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Error(response, "remote-secret-body", http.StatusUnauthorized)
	}))
	defer server.Close()
	client := newTestIssueClient(t, server, "o/r")
	_, err := client.CreateIssue(t.Context(), application.BugReportIssue{Title: "title", Body: "body"})
	if err == nil || strings.Contains(err.Error(), "issue-token") || strings.Contains(err.Error(), "remote-secret-body") {
		t.Fatalf("unsafe error: %v", err)
	}
}

func TestGitHubBugReportCapabilityPaginatesLabelsAndFailsOnLaterPageErrors(t *testing.T) {
	for _, failSecondPage := range []bool{false, true} {
		t.Run(fmt.Sprintf("later-page-error=%v", failSecondPage), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				assertGitHubBugReportHeaders(t, request, "issue-token")
				if request.URL.Path == "/repos/o/r" {
					_, _ = io.WriteString(response, `{"private":false}`)
					return
				}
				if request.URL.Path != "/repos/o/r/labels" || request.URL.Query().Get("per_page") != "100" {
					t.Fatalf("unexpected labels request: %s", request.URL.String())
				}
				switch request.URL.Query().Get("page") {
				case "", "1":
					labels := make([]map[string]string, 100)
					for index := range labels {
						labels[index] = map[string]string{"name": fmt.Sprintf("existing-%03d", index)}
					}
					_ = json.NewEncoder(response).Encode(labels)
				case "2":
					if failSecondPage {
						response.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					_, _ = io.WriteString(response, `[{"name":"bug"},{"name":"source:in-app"},{"name":"area:vehicles"}]`)
				default:
					t.Fatalf("unexpected label page: %s", request.URL.RawQuery)
				}
			}))
			defer server.Close()
			err := infrastructure.ValidateGitHubBugReportCapability(t.Context(), newTestIssueClient(t, server, "o/r"), nil,
				[]string{"bug", "source:in-app", "area:vehicles"})
			if failSecondPage {
				if err == nil || !strings.Contains(err.Error(), "status 503") {
					t.Fatalf("later label-page failure must disable reporting: %v", err)
				}
			} else if err != nil {
				t.Fatalf("required labels on a later page must enable reporting: %v", err)
			}
		})
	}
}

func newTestIssueClient(t *testing.T, server *httptest.Server, repository string) *infrastructure.GitHubBugReportIssueClient {
	t.Helper()
	client, err := infrastructure.NewGitHubBugReportIssueClient(infrastructure.GitHubBugReportClientConfig{
		APIBaseURL: server.URL, Token: "issue-token", Repository: repository, HTTPClient: server.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func assertGitHubBugReportHeaders(t *testing.T, request *http.Request, token string) {
	t.Helper()
	if request.Header.Get("Authorization") != "Bearer "+token ||
		request.Header.Get("Accept") != "application/vnd.github+json" ||
		request.Header.Get("X-GitHub-Api-Version") != "2026-03-10" {
		t.Fatalf("unexpected GitHub headers: %#v", request.Header)
	}
}
