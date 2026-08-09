package api

import (
	"errors"
	"net/http"
	"strconv"

	"railkeeper/backend/internal/application"
)

func (a *App) listTrackGeometries(w http.ResponseWriter, r *http.Request) {
	geometries, err := a.trackPlannerService.ListGeometries(r.Context(), r.URL.Query().Get("gauge"))
	if err != nil {
		a.trackPlannerError(w, err, "list track geometries")
		return
	}
	respondJSON(w, http.StatusOK, geometries)
}

func (a *App) getTrackPlan(w http.ResponseWriter, r *http.Request) {
	plan, err := a.trackPlannerService.GetPlan(r.Context(), r.PathValue("id"))
	if err != nil {
		a.trackPlannerError(w, err, "get track plan")
		return
	}
	respondJSON(w, http.StatusOK, plan)
}

func (a *App) getTrackPlanAnalysis(w http.ResponseWriter, r *http.Request) {
	analysis, err := a.trackPlannerService.AnalyzePlan(r.Context(), r.PathValue("id"))
	if err != nil {
		a.trackPlannerError(w, err, "analyze track plan")
		return
	}
	respondJSON(w, http.StatusOK, analysis)
}

func (a *App) getTrackPlanChangePreview(w http.ResponseWriter, r *http.Request) {
	preview, err := a.trackPlannerService.ChangePreview(r.Context(), r.PathValue("id"))
	if err != nil {
		a.trackPlannerError(w, err, "preview track plan changes")
		return
	}
	respondJSON(w, http.StatusOK, preview)
}

func (a *App) createPlanTrackObject(w http.ResponseWriter, r *http.Request) {
	var input application.CreatePlanTrackObjectInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	object, err := a.trackPlannerService.CreateObject(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.trackPlannerError(w, err, "create track plan object")
		return
	}
	respondJSON(w, http.StatusCreated, object)
}

func (a *App) updatePlanTrackObject(w http.ResponseWriter, r *http.Request) {
	var input application.UpdatePlanTrackObjectInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	object, err := a.trackPlannerService.UpdateObject(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.trackPlannerError(w, err, "update track plan object")
		return
	}
	respondJSON(w, http.StatusOK, object)
}

func (a *App) deletePlanTrackObject(w http.ResponseWriter, r *http.Request) {
	expectedVersion, err := strconv.Atoi(r.URL.Query().Get("expectedVersion"))
	if err != nil {
		a.trackPlannerError(w, application.ErrTrackPlanValidation, "delete track plan object")
		return
	}
	if err := a.trackPlannerService.DeleteObject(
		r.Context(), r.PathValue("id"), expectedVersion, actorUserID(r),
	); err != nil {
		a.trackPlannerError(w, err, "delete track plan object")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) trackPlannerError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrTrackPlanValidation):
		respondProblem(w, http.StatusBadRequest, "track_plan_validation", "Track plan data is invalid.")
	case errors.Is(err, application.ErrTrackPlanNotFound):
		respondProblem(w, http.StatusNotFound, "track_plan_not_found", "Track plan resource not found.")
	case errors.Is(err, application.ErrTrackPlanImmutable):
		respondProblem(w, http.StatusConflict, "track_plan_immutable", "Published track plans are immutable.")
	case errors.Is(err, application.ErrTrackPlanConflict):
		respondProblem(w, http.StatusConflict, "track_plan_conflict", "Track plan data has changed.")
	default:
		a.logger.Error("track planner operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "track_plan_operation_failed", "Track plan operation failed.")
	}
}
