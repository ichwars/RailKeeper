package api

import (
	"context"
	"net/http"
	"time"
)

func (a *App) readVehicleECoSFunctions(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(45 * time.Second))
	w.Header().Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(r.Context(), 40*time.Second)
	defer cancel()
	result, err := a.digitalCenterWorkspace.ReadVehicleECoSFunctions(ctx, r.PathValue("id"))
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}
