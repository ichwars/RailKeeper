package api

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestBugReportPayloadContractsRejectAmbiguousOrExcessiveInput(t *testing.T) {
	base := [][2]string{{"reportId", "e58a7637-cf59-4973-a316-a1defafcb5b6"}, {"area", "vehicles"}, {"description", "Synthetic report"}, {"includeDiagnostics", "false"}}
	parse := func(fields [][2]string, screenshot []byte) (application.BugReportSubmissionRequest, error) {
		t.Helper()
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		for _, field := range fields {
			if err := writer.WriteField(field[0], field[1]); err != nil {
				t.Fatal(err)
			}
		}
		if screenshot != nil {
			part, err := writer.CreateFormFile("screenshot", "capture.png")
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
		r := httptest.NewRequest("POST", "/api/v1/bug-reports", &body)
		r.Header.Set("Content-Type", writer.FormDataContentType())
		return parseBugReportMultipart(r)
	}
	screenshot := []byte("synthetic screenshot bytes")
	result, err := parse(append(append([][2]string{}, base...), [2]string{"screenshotConsent", "true"}), screenshot)
	if err != nil || !bytes.Equal(result.ScreenshotBytes, screenshot) || !result.Input.OmitDiagnostics {
		t.Fatalf("payload bytes or diagnostics choice changed: %v", err)
	}
	for _, fields := range [][][2]string{
		append(append([][2]string{}, base...), [2]string{"unknown", "private"}),
		append(append([][2]string{}, base...), [2]string{"area", "accessories"}),
		append(append([][2]string{}, base...), [2]string{"screenshotConsent", "invalid"}),
		{{"reportId", "x"}},
		{{"reportId", "x"}, {"area", "vehicles"}, {"description", "x"}, {"includeDiagnostics", "invalid"}},
		{{"reportId", "x"}, {"area", "vehicles"}, {"description", "x"}, {"clientDiagnostics", "{"}},
		{{"reportId", "x"}, {"area", "vehicles"}, {"description", "x"}, {"clientDiagnostics", "{} {}"}},
	} {
		if _, err := parse(fields, nil); !errors.Is(err, application.ErrBugReportValidation) {
			t.Fatalf("ambiguous payload accepted: %v", err)
		}
	}
	if _, err := parse(base, bytes.Repeat([]byte("x"), application.MaxBugReportScreenshotBytes+1)); !errors.Is(err, application.ErrBugReportScreenshotTooLarge) {
		t.Fatalf("screenshot limit ignored: %v", err)
	}
	if _, err := readBoundedBugReportPart(failingContractReader{}, 4); !errors.Is(err, application.ErrBugReportValidation) {
		t.Fatal("failed screenshot read accepted")
	}
}

func TestBugReportHTTPProblemsRetainRetryAndPrivacySemantics(t *testing.T) {
	app := &App{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for _, c := range []struct {
		err    error
		status int
		code   string
	}{
		{application.ErrBugReportScreenshotTooLarge, 413, "bug_report_screenshot_too_large"},
		{application.ErrBugReportScreenshotInvalid, 400, "bug_report_screenshot_invalid"},
		{application.ErrBugReportValidation, 400, "bug_report_validation_failed"},
		{application.ErrBugReportRateLimited, 429, "bug_report_rate_limited"},
		{application.ErrBugReportStateConflict, 409, "bug_report_submission_conflict"},
		{application.ErrBugReportGitHubUnavailable, 502, "bug_report_github_unavailable"},
		{application.ErrBugReportingUnavailable, 503, "bug_reporting_unavailable"},
		{errors.New("private diagnostic details"), 500, "bug_report_submission_failed"},
	} {
		w := httptest.NewRecorder()
		app.respondBugReportError(w, c.err)
		assertProblem(t, w, c.status, c.code)
		if strings.Contains(w.Body.String(), "private diagnostic details") {
			t.Fatal("private error details exposed")
		}
		if c.status == 429 && w.Header().Get("Retry-After") != "3600" {
			t.Fatal("retry hint lost")
		}
	}
}

type failingContractReader struct{}

func (failingContractReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }
