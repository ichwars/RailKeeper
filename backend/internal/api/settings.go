package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) getProfileSettings(w http.ResponseWriter, r *http.Request) {
	if a.settingsService == nil {
		respondJSON(w, http.StatusOK, application.SettingsPayload{Settings: map[string]string{}})
		return
	}
	settings, err := a.settingsService.UserSettings(r.Context(), actorUserID(r))
	if err != nil {
		a.logger.Error("profile settings load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "profile_settings_failed", "Profileinstellungen konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) updateProfileSettings(w http.ResponseWriter, r *http.Request) {
	if a.settingsService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "settings_unavailable", "Einstellungen sind nicht verfügbar.")
		return
	}
	var input application.SettingsPayload
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	settings, err := a.settingsService.UpdateUserSettings(r.Context(), actorUserID(r), input)
	if err != nil {
		if errors.Is(err, application.ErrSettingsValidation) {
			respondProblem(w, http.StatusBadRequest, "settings_validation", "Ungültige Einstellung.")
			return
		}
		a.logger.Error("profile settings update failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "profile_settings_failed", "Profileinstellungen konnten nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) getDigitalSettings(w http.ResponseWriter, r *http.Request) {
	if a.settingsService == nil {
		respondJSON(w, http.StatusOK, application.DigitalCenterSettings{})
		return
	}
	settings, err := a.settingsService.DigitalSettings(r.Context())
	if err != nil {
		a.logger.Error("digital settings load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "digital_settings_failed", "Digitalzentralen-Einstellungen konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (a *App) updateDigitalSettings(w http.ResponseWriter, r *http.Request) {
	if a.settingsService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "settings_unavailable", "Einstellungen sind nicht verfügbar.")
		return
	}
	var input application.DigitalCenterSettings
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	settings, err := a.settingsService.UpdateDigitalSettings(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrSettingsValidation) {
			respondProblem(w, http.StatusBadRequest, "settings_validation", "Ungültige Digitalzentralen-Einstellung.")
			return
		}
		a.logger.Error("digital settings update failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "digital_settings_failed", "Digitalzentralen-Einstellungen konnten nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}
