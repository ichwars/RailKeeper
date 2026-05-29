package api

import (
	"net/http"
	"strconv"

	"railkeeper/backend/internal/application"
)

type auditLogResponse struct {
	Entries []application.AuditLogEntry `json:"entries"`
}

func (a *App) systemAuditLog(w http.ResponseWriter, r *http.Request) {
	entries, err := a.authService.ListAuditLog(r.Context(), auditLimit(r.URL.Query().Get("limit")))
	if err != nil {
		a.logger.Error("audit log list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "audit_log_failed", "Sicherheitsereignisse konnten nicht gelesen werden.")
		return
	}
	respondJSON(w, http.StatusOK, auditLogResponse{Entries: entries})
}

func auditLimit(value string) int {
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}
