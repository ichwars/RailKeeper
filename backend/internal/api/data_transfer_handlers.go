package api

import (
	"context"
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

type dataTransferScopeKey struct{}

func (a *App) listDataTransferProfiles(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	profiles, err := a.dataTransferService.ListProfiles(r.Context())
	if err != nil {
		a.dataTransferError(w, err, "list data transfer profiles")
		return
	}
	if dataTransferMesseOnly(r) {
		profiles = filterMesseDataTransferProfiles(profiles)
	}
	respondJSON(w, http.StatusOK, profiles)
}

func (a *App) createDataTransferProfile(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input application.CreateDataTransferProfileInput
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	profile, err := a.dataTransferService.CreateProfile(r.Context(), input, actorUserID(r))
	if err != nil {
		a.dataTransferError(w, err, "create data transfer profile")
		return
	}
	respondJSON(w, http.StatusCreated, profile)
}

func (a *App) updateDataTransferProfile(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input application.UpdateDataTransferProfileInput
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	profile, err := a.dataTransferService.UpdateProfile(r.Context(), r.PathValue("id"), input)
	if err != nil {
		a.dataTransferError(w, err, "update data transfer profile")
		return
	}
	respondJSON(w, http.StatusOK, profile)
}

func (a *App) disableDataTransferProfile(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	if _, err := a.dataTransferService.DisableProfile(r.Context(), r.PathValue("id")); err != nil {
		a.dataTransferError(w, err, "disable data transfer profile")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) dataTransferAvailable(w http.ResponseWriter) bool {
	if a.dataTransferService != nil {
		return true
	}
	respondProblem(w, http.StatusServiceUnavailable, "data_transfer_unavailable", "Data transfer is not available.")
	return false
}

func (a *App) dataTransferError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrDataTransferValidation):
		respondProblem(w, http.StatusBadRequest, "data_transfer_validation", err.Error())
	case errors.Is(err, application.ErrDataTransferNotFound):
		respondProblem(w, http.StatusNotFound, "data_transfer_not_found", "Data transfer profile not found.")
	default:
		a.logger.Error("data transfer operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "data_transfer_operation_failed", "Data transfer operation failed.")
	}
}

func withDataTransferScope(r *http.Request, messeOnly bool) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), dataTransferScopeKey{}, messeOnly))
}

func dataTransferMesseOnly(r *http.Request) bool {
	messeOnly, _ := r.Context().Value(dataTransferScopeKey{}).(bool)
	return messeOnly
}

func dataTransferMesseScope(roles []string) bool {
	hasMesse := false
	for _, role := range roles {
		switch role {
		case "Admin", "Editor", "Planner", "Viewer":
			return false
		case "Messe":
			hasMesse = true
		}
	}
	return hasMesse
}

func filterMesseDataTransferProfiles(profiles []application.DataTransferProfile) []application.DataTransferProfile {
	filtered := make([]application.DataTransferProfile, 0, len(profiles))
	for _, profile := range profiles {
		if len(profile.Areas) == 1 && profile.Areas[0] == application.TransferExhibitionLists {
			filtered = append(filtered, profile)
		}
	}
	return filtered
}
