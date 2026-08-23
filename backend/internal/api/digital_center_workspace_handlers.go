package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"railkeeper/backend/internal/application"
)

func (a *App) digitalCenterWorkspaceSummary(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	centers, err := a.digitalCenterWorkspace.ListConfiguredCenters(r.Context())
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"centers": centers})
}

func (a *App) startDigitalCenterReadSession(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	session, err := a.digitalCenterWorkspace.StartReadSession(
		r.Context(), r.PathValue("provider"), actorUserID(r),
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, session)
}

func (a *App) getDigitalCenterReadSession(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	session, err := a.digitalCenterWorkspace.GetReadSession(r.Context(), r.PathValue("id"))
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, session)
}

func (a *App) listDigitalCenterWorkItems(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	page, err := optionalPositiveQueryInt(r, "page")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_filter_invalid", "Page must be a positive integer.")
		return
	}
	pageSize, err := optionalPositiveQueryInt(r, "pageSize")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_filter_invalid", "Page size must be a positive integer.")
		return
	}
	items, err := a.digitalCenterWorkspace.ListWorkItems(
		r.Context(),
		r.PathValue("id"),
		application.DigitalCenterWorkItemFilter{
			Query:         r.URL.Query().Get("q"),
			CompareStatus: application.DigitalCenterCompareStatus(r.URL.Query().Get("compareStatus")),
			Page:          page,
			PageSize:      pageSize,
		},
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (a *App) getDigitalCenterWorkItem(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	item, err := a.digitalCenterWorkspace.GetWorkItem(
		r.Context(), r.PathValue("id"), r.PathValue("itemID"),
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (a *App) previewDigitalCenterWrite(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	var input application.DigitalCenterWritePreviewInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	preview, err := a.digitalCenterWorkspace.PreviewWrite(
		r.Context(), r.PathValue("id"), r.PathValue("itemID"), input, actorUserID(r),
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, preview)
}

func (a *App) confirmDigitalCenterWrite(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	var input application.DigitalCenterWriteConfirmInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterWorkspace.ConfirmWrite(
		r.Context(), r.PathValue("id"), r.PathValue("itemID"), input, actorUserID(r),
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) digitalCenterLiveStatus(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	status, err := a.digitalCenterWorkspace.LiveMonitorStatus(r.Context(), r.PathValue("provider"))
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (a *App) startDigitalCenterLiveMonitor(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	status, err := a.digitalCenterWorkspace.StartLiveMonitor(
		r.Context(), r.PathValue("provider"), r.URL.Query().Get("sessionId"),
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (a *App) stopDigitalCenterLiveMonitor(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	status, err := a.digitalCenterWorkspace.StopLiveMonitor(
		r.Context(), r.PathValue("provider"), r.URL.Query().Get("sessionId"),
	)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (a *App) listDigitalCenterSessionMessages(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	messages, err := a.digitalCenterWorkspace.ListSessionMessages(r.Context(), r.PathValue("id"))
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, map[string]any{"messages": messages})
}

func (a *App) requireDigitalCenterWorkspace(w http.ResponseWriter) bool {
	if a.digitalCenterWorkspace != nil {
		return true
	}
	respondProblem(w, http.StatusServiceUnavailable, "digital_center_workspace_unavailable",
		"Digital center workspace is not available.")
	return false
}

func (a *App) respondDigitalCenterWorkspaceError(w http.ResponseWriter, err error) {
	var addressConflict *application.DigitalCenterAddressConflictError
	switch {
	case errors.Is(err, application.ErrDigitalCenterWorkspaceUnavailable):
		respondProblem(w, http.StatusServiceUnavailable, "digital_center_workspace_unavailable",
			"Digital center workspace is not available.")
	case errors.Is(err, application.ErrDigitalCenterNotConfigured):
		respondProblem(w, http.StatusNotFound, "digital_center_not_configured",
			"Digital center is not configured in Settings.")
	case errors.Is(err, application.ErrDigitalCenterInactive):
		respondProblem(w, http.StatusConflict, "digital_center_inactive",
			"Digital center is inactive in Settings.")
	case errors.Is(err, application.ErrDigitalCenterCapabilityUnavailable):
		respondProblem(w, http.StatusBadRequest, "digital_center_capability_unavailable",
			"The selected digital center does not support this operation.")
	case errors.Is(err, application.ErrDigitalCenterLiveStartFailed):
		respondProblem(w, http.StatusBadGateway, "ecos_live_start_failed",
			"ECoS live monitoring could not be started.")
	case errors.Is(err, application.ErrDigitalCenterFilterValidation):
		respondProblem(w, http.StatusBadRequest, "digital_center_filter_invalid",
			"Digital center work-list filter is invalid.")
	case errors.Is(err, application.ErrDigitalCenterReadNotFresh):
		respondProblem(w, http.StatusConflict, "digital_center_read_not_fresh",
			"Create a fresh digital center read before writing.")
	case errors.Is(err, application.ErrDigitalCenterConflictUnresolved):
		respondProblem(w, http.StatusConflict, "digital_center_conflict_unresolved",
			"Resolve the work-item conflict before writing.")
	case errors.As(err, &addressConflict):
		respondProblemDetails(w, http.StatusConflict, "digital_center_address_conflict",
			"The decoder address is already used by another ECoS locomotive.", map[string]any{
				"objectId":       addressConflict.ObjectID,
				"name":           addressConflict.Name,
				"decoderAddress": addressConflict.Address,
			})
	case errors.Is(err, application.ErrDigitalCenterPreviewStale):
		respondProblem(w, http.StatusConflict, "digital_center_write_preview_stale",
			"The write preview is stale. Create a fresh preview.")
	case errors.Is(err, application.ErrDigitalCenterGrantExpired),
		errors.Is(err, application.ErrDigitalCenterGrantConsumed),
		errors.Is(err, application.ErrDigitalCenterGrantActorMismatch),
		errors.Is(err, application.ErrDigitalCenterGrantMismatch):
		respondProblem(w, http.StatusConflict, "digital_center_write_grant_conflict",
			"The write grant is expired, consumed, or does not match this request.")
	case errors.Is(err, application.ErrDigitalCenterWriteFieldUnsupported):
		respondProblem(w, http.StatusBadRequest, "digital_center_write_field_unsupported",
			"The requested field cannot be written for this work item.")
	case errors.Is(err, application.ErrDigitalCenterConfirmationRequired):
		respondProblem(w, http.StatusBadRequest, "digital_center_confirmation_required",
			"Explicit confirmation is required.")
	case errors.Is(err, application.ErrDigitalCenterWriteNoChanges):
		respondProblem(w, http.StatusConflict, "digital_center_write_no_changes",
			"The digital center already has the previewed values.")
	case errors.Is(err, application.ErrDigitalCenterDeviceWrite):
		respondProblem(w, http.StatusBadGateway, "ecos_sync_failed",
			"The ECoS write or verification failed.")
	case errors.Is(err, sql.ErrNoRows):
		respondProblem(w, http.StatusNotFound, "digital_center_workspace_not_found",
			"Digital center workspace resource was not found.")
	default:
		a.logger.Error("digital center workspace operation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "digital_center_workspace_failed",
			"Digital center workspace operation failed.")
	}
}

func optionalPositiveQueryInt(r *http.Request, name string) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(name))
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, application.ErrDigitalCenterFilterValidation
	}
	return parsed, nil
}
