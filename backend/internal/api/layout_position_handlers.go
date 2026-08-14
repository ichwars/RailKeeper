package api

import (
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) listLayoutTechnicalPositions(w http.ResponseWriter, r *http.Request) {
	positions, err := a.layoutService.ListTechnicalPositions(r.Context(), r.PathValue("id"))
	if err != nil {
		a.layoutError(w, err, "list layout technical positions")
		return
	}
	respondJSON(w, http.StatusOK, positions)
}

func (a *App) createLayoutTechnicalPosition(w http.ResponseWriter, r *http.Request) {
	var input application.CreateLayoutTechnicalPositionInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	position, err := a.layoutService.CreateTechnicalPosition(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "create layout technical position")
		return
	}
	respondJSON(w, http.StatusCreated, position)
}

func (a *App) updateLayoutTechnicalPosition(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateLayoutTechnicalPositionInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	position, err := a.layoutService.UpdateTechnicalPosition(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "update layout technical position")
		return
	}
	respondJSON(w, http.StatusOK, position)
}
