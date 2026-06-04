package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) twoFactorStatus(w http.ResponseWriter, r *http.Request) {
	status, err := a.authService.TwoFactorStatus(r.Context(), actorUserID(r))
	if err != nil {
		handleTwoFactorError(a, w, err, "two_factor_status_failed", "Zwei-Faktor-Status konnte nicht gelesen werden.")
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (a *App) setupTwoFactor(w http.ResponseWriter, r *http.Request) {
	status, err := a.authService.PrepareTwoFactor(r.Context(), actorUserID(r))
	if err != nil {
		handleTwoFactorError(a, w, err, "two_factor_setup_failed", "Zwei-Faktor-Auth konnte nicht vorbereitet werden.")
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (a *App) enableTwoFactor(w http.ResponseWriter, r *http.Request) {
	var input application.TwoFactorEnableInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	status, err := a.authService.EnableTwoFactor(r.Context(), actorUserID(r), input)
	if err != nil {
		handleTwoFactorError(a, w, err, "two_factor_enable_failed", "Zwei-Faktor-Auth konnte nicht aktiviert werden.")
		return
	}
	respondJSON(w, http.StatusOK, status)
}

func (a *App) disableTwoFactor(w http.ResponseWriter, r *http.Request) {
	var input application.TwoFactorDisableInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if err := a.authService.DisableTwoFactor(r.Context(), actorUserID(r), cookieValue(r, "rk_session"), input); err != nil {
		handleTwoFactorError(a, w, err, "two_factor_disable_failed", "Zwei-Faktor-Auth konnte nicht deaktiviert werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func handleTwoFactorError(a *App, w http.ResponseWriter, err error, code, message string) {
	switch {
	case errors.Is(err, application.ErrInvalidLogin):
		respondProblem(w, http.StatusUnauthorized, "invalid_two_factor", "Der Zwei-Faktor-Code oder das Passwort ist nicht korrekt.")
	case errors.Is(err, application.ErrUserNotFound), errors.Is(err, application.ErrUnauthorized):
		respondProblem(w, http.StatusUnauthorized, "unauthorized", "Not logged in.")
	default:
		a.logger.Error(message, "error", err)
		respondProblem(w, http.StatusInternalServerError, code, message)
	}
}
