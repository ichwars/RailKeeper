package api

import (
	"net/http"
	"strconv"

	"railkeeper/backend/internal/application"
)

func (a *App) createFreePlanObject(w http.ResponseWriter, r *http.Request) {
	var input application.CreateFreePlanObjectInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	object, err := a.trackPlannerService.CreateFreeObject(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.trackPlannerError(w, err, "create free plan object")
		return
	}
	respondJSON(w, http.StatusCreated, object)
}

func (a *App) updateFreePlanObject(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateFreePlanObjectInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	object, err := a.trackPlannerService.UpdateFreeObject(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.trackPlannerError(w, err, "update free plan object")
		return
	}
	respondJSON(w, http.StatusOK, object)
}

func (a *App) deleteFreePlanObject(w http.ResponseWriter, r *http.Request) {
	expectedVersion, err := strconv.Atoi(r.URL.Query().Get("expectedVersion"))
	if err != nil {
		a.trackPlannerError(w, application.ErrTrackPlanValidation, "delete free plan object")
		return
	}
	if err := a.trackPlannerService.DeleteFreeObject(
		r.Context(), r.PathValue("id"), expectedVersion, actorUserID(r),
	); err != nil {
		a.trackPlannerError(w, err, "delete free plan object")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
