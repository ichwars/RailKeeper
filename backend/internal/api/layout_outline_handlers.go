package api

import (
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) updateLayoutUnitOutline(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateLayoutUnitOutlineInput
	if !decodeLayoutJSON(w, r, &input) {
		return
	}
	outline, err := a.layoutService.UpdateUnitOutline(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.layoutError(w, err, "update layout unit outline")
		return
	}
	respondJSON(w, http.StatusOK, outline)
}
