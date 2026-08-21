package api

import (
	"database/sql"
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
