package api

import (
	"errors"
	"net/http"
	"strconv"

	"railkeeper/backend/internal/application"
)

func (a *App) listSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := a.authService.ListSessions(r.Context(), sessionLimit(r.URL.Query().Get("limit")))
	if err != nil {
		a.logger.Error("session list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "session_list_failed", "Sitzungen konnten nicht gelesen werden.")
		return
	}
	respondJSON(w, http.StatusOK, sessions)
}

func (a *App) revokeSession(w http.ResponseWriter, r *http.Request) {
	if err := a.authService.RevokeSession(r.Context(), actorUserID(r), r.PathValue("id")); err != nil {
		if errors.Is(err, application.ErrSessionNotFound) {
			respondProblem(w, http.StatusNotFound, "session_not_found", "Sitzung wurde nicht gefunden.")
			return
		}
		a.logger.Error("session revoke failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "session_revoke_failed", "Sitzung konnte nicht widerrufen werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func sessionLimit(value string) int {
	limit, err := strconv.Atoi(value)
	if err != nil || limit <= 0 {
		return 200
	}
	if limit > 200 {
		return 200
	}
	return limit
}
