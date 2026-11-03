package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

const (
	bugReportFallbackURL     = "https://github.com/ichwars/RailKeeper/issues/new"
	maxBugReportRequestBytes = application.MaxBugReportScreenshotBytes + 200*1024
	bugReportAttemptLimit    = 15
)

type BugReportCapability struct {
	ScreenshotAvailable          bool                                    `json:"screenshotAvailable"`
	Enabled                      bool                                    `json:"enabled"`
	FallbackURL                  string                                  `json:"fallbackUrl"`
	MaxScreenshotBytes           int                                     `json:"maxScreenshotBytes"`
	AcceptedScreenshotTypes      []string                                `json:"acceptedScreenshotTypes"`
	IncludedDiagnosticCategories []string                                `json:"includedDiagnosticCategories"`
	ExcludedDiagnosticCategories []string                                `json:"excludedDiagnosticCategories"`
	ContactAvailable             bool                                    `json:"contactAvailable"`
	ServerDiagnostics            *application.BugReportServerDiagnostics `json:"serverDiagnostics,omitempty"`
}

type BugReportCapabilityControl struct {
	Provider   func(context.Context) BugReportCapability
	Invalidate func()
}

func normalizeBugReportCapability(capability BugReportCapability) BugReportCapability {
	if capability.FallbackURL == "" {
		capability.FallbackURL = bugReportFallbackURL
	}
	capability.MaxScreenshotBytes = application.MaxBugReportScreenshotBytes
	capability.AcceptedScreenshotTypes = []string{"image/png", "image/jpeg", "image/webp"}
	capability.IncludedDiagnosticCategories = []string{
		"railkeeper", "runtime", "browser", "presentation", "navigation", "roles",
		"enabledDigitalProviderTypes", "clientErrors", "processHealth", "databaseHealth",
	}
	capability.ExcludedDiagnosticCategories = []string{
		"inventoryData", "inventoryCounts", "connectedDevices", "addresses", "credentials",
		"requestContent", "localPaths",
	}
	return capability
}

func (app *App) currentBugReportCapability(ctx context.Context) BugReportCapability {
	if app.bugReportCapabilityControl.Provider != nil {
		return normalizeBugReportCapability(app.bugReportCapabilityControl.Provider(ctx))
	}
	return app.bugReportCapability
}

func (app *App) bugReportConfig(response http.ResponseWriter, request *http.Request) {
	capability := app.currentBugReportCapability(request.Context())
	if request.URL.Query().Get("diagnostics") == "true" && app.bugReportService != nil {
		diagnostics, err := app.bugReportService.Diagnostics(request.Context())
		if err != nil {
			respondProblem(response, http.StatusServiceUnavailable, "bug_report_diagnostics_unavailable", "Diagnostics could not be collected.")
			return
		}
		capability.ServerDiagnostics = &diagnostics
	} else {
		capability.ServerDiagnostics = nil
	}
	respondJSON(response, http.StatusOK, capability)
}

func (app *App) submitBugReport(response http.ResponseWriter, request *http.Request) {
	if app.bugReportService == nil || !app.currentBugReportCapability(request.Context()).Enabled {
		respondProblem(response, http.StatusServiceUnavailable, "bug_reporting_unavailable",
			"In-app bug reporting is not configured.")
		return
	}
	allowed, err := app.rateLimits.Allow(
		request.Context(), "bug_report_attempt", actorUserID(request), bugReportAttemptLimit, time.Hour,
	)
	if err != nil {
		app.logger.Error("bug report attempt admission failed", "error_category", "rate_limit")
		respondProblem(response, http.StatusInternalServerError, "bug_report_submission_failed",
			"The bug report could not be submitted.")
		return
	}
	if !allowed {
		response.Header().Set("Retry-After", "3600")
		respondProblem(response, http.StatusTooManyRequests, "bug_report_rate_limited",
			"Too many bug report attempts. Please try again later.")
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, maxBugReportRequestBytes)
	parsed, err := parseBugReportMultipart(request)
	if err != nil {
		app.respondBugReportError(response, err)
		return
	}
	if len(parsed.ScreenshotBytes) > 0 && !app.currentBugReportCapability(request.Context()).ScreenshotAvailable {
		app.respondBugReportError(response, application.ErrBugReportValidation)
		return
	}
	result, err := app.bugReportService.Submit(
		request.Context(), actorUserID(request), parsed,
		application.BugReportAdmissionPolicy{
			Scope: "bug_report_new", Limit: 5, Window: time.Hour,
		},
	)
	if err != nil {
		if errors.Is(err, application.ErrBugReportGitHubUnavailable) &&
			app.bugReportCapabilityControl.Invalidate != nil {
			app.bugReportCapabilityControl.Invalidate()
		}
		app.respondBugReportError(response, err)
		return
	}
	if result.Status == application.BugReportPartial && app.bugReportCapabilityControl.Invalidate != nil {
		app.bugReportCapabilityControl.Invalidate()
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	respondJSON(response, status, result)
}

func parseBugReportMultipart(request *http.Request) (application.BugReportSubmissionRequest, error) {
	reader, err := request.MultipartReader()
	if err != nil {
		return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
	}
	fields := map[string]string{}
	seen := map[string]bool{}
	var screenshot []byte
	var screenshotMediaType string
	for {
		part, nextErr := reader.NextPart()
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if nextErr != nil {
			var maxBytesError *http.MaxBytesError
			if errors.As(nextErr, &maxBytesError) {
				return application.BugReportSubmissionRequest{}, application.ErrBugReportScreenshotTooLarge
			}
			return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
		}
		name := part.FormName()
		if seen[name] || !knownBugReportMultipartField(name) {
			_ = part.Close()
			return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
		}
		seen[name] = true
		if name == "screenshot" {
			screenshot, err = readBoundedBugReportPart(part, application.MaxBugReportScreenshotBytes)
			if err == nil {
				screenshotMediaType, _, err = mime.ParseMediaType(part.Header.Get("Content-Type"))
			}
		} else {
			fields[name], err = readBoundedBugReportTextPart(part, bugReportTextFieldLimit(name))
		}
		_ = part.Close()
		if err != nil {
			return application.BugReportSubmissionRequest{}, err
		}
	}
	if fields["reportId"] == "" || fields["area"] == "" || fields["description"] == "" {
		return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
	}
	includeDiagnostics := true
	if value := fields["includeDiagnostics"]; value != "" {
		includeDiagnostics, err = strconv.ParseBool(value)
		if err != nil {
			return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
		}
	}
	var diagnostics application.BugReportClientDiagnostics
	if includeDiagnostics {
		decoder := json.NewDecoder(strings.NewReader(fields["clientDiagnostics"]))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&diagnostics); err != nil {
			return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
		}
		if err := ensureBugReportJSONEnd(decoder); err != nil {
			return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
		}
	}
	consent := false
	if value := fields["screenshotConsent"]; value != "" {
		consent, err = strconv.ParseBool(value)
		if err != nil {
			return application.BugReportSubmissionRequest{}, application.ErrBugReportValidation
		}
	}
	return application.BugReportSubmissionRequest{
		Input: application.BugReportInput{
			OmitDiagnostics: !includeDiagnostics, ReportID: fields["reportId"], Area: application.BugReportArea(fields["area"]),
			Description: fields["description"], ExpectedBehavior: fields["expectedBehavior"],
			ReproductionSteps: fields["reproductionSteps"], Email: fields["email"],
			ClientDiagnostics: diagnostics, ScreenshotConsent: consent,
		},
		ScreenshotBytes: screenshot, ScreenshotMediaType: screenshotMediaType,
	}, nil
}

func knownBugReportMultipartField(name string) bool {
	switch name {
	case "reportId", "area", "description", "expectedBehavior", "reproductionSteps", "email",
		"clientDiagnostics", "includeDiagnostics", "screenshotConsent", "screenshot":
		return true
	default:
		return false
	}
}

func bugReportTextFieldLimit(name string) int {
	switch name {
	case "reportId":
		return 64
	case "area":
		return 32
	case "description", "reproductionSteps":
		return 40_000
	case "expectedBehavior":
		return 20_000
	case "email":
		return 254
	case "clientDiagnostics":
		return 64 * 1024
	case "screenshotConsent", "includeDiagnostics":
		return 8
	default:
		return 0
	}
}

func readBoundedBugReportPart(part io.Reader, limit int) ([]byte, error) {
	value, err := io.ReadAll(io.LimitReader(part, int64(limit)+1))
	if err != nil {
		return nil, application.ErrBugReportValidation
	}
	if len(value) > limit {
		return nil, application.ErrBugReportScreenshotTooLarge
	}
	return value, nil
}

func readBoundedBugReportTextPart(part io.Reader, limit int) (string, error) {
	value, err := io.ReadAll(io.LimitReader(part, int64(limit)+1))
	if err != nil || len(value) > limit {
		return "", application.ErrBugReportValidation
	}
	return string(value), nil
}

func ensureBugReportJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
}

func (app *App) respondBugReportError(response http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrBugReportScreenshotTooLarge):
		respondProblem(response, http.StatusRequestEntityTooLarge, "bug_report_screenshot_too_large",
			"The screenshot is too large.")
	case errors.Is(err, application.ErrBugReportScreenshotInvalid):
		respondProblem(response, http.StatusBadRequest, "bug_report_screenshot_invalid",
			"The screenshot is invalid.")
	case errors.Is(err, application.ErrBugReportValidation):
		respondProblem(response, http.StatusBadRequest, "bug_report_validation_failed", "The bug report is invalid.")
	case errors.Is(err, application.ErrBugReportRateLimited):
		response.Header().Set("Retry-After", "3600")
		respondProblem(response, http.StatusTooManyRequests, "bug_report_rate_limited",
			"Too many new bug reports. Please try again later.")
	case errors.Is(err, application.ErrBugReportStateConflict):
		respondProblem(response, http.StatusConflict, "bug_report_submission_conflict",
			"The bug report belongs to another submission.")
	case errors.Is(err, application.ErrBugReportGitHubUnavailable):
		respondProblem(response, http.StatusBadGateway, "bug_report_github_unavailable",
			"GitHub is currently unavailable.")
	case errors.Is(err, application.ErrBugReportingUnavailable):
		respondProblem(response, http.StatusServiceUnavailable, "bug_reporting_unavailable",
			"In-app bug reporting is not configured.")
	default:
		app.logger.Error("bug report submission failed", "error_category", "internal")
		respondProblem(response, http.StatusInternalServerError, "bug_report_submission_failed",
			"The bug report could not be submitted.")
	}
}
