package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"railkeeper/backend/internal/application"
)

func (a *App) readDigitalCenterCVBatch(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	var input application.ECoSCVReadInput
	if !decodeDigitalCenterCVRequest(w, r, &input) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(95 * time.Second))
	w.Header().Set("Cache-Control", "no-store")
	item, err := a.digitalCenterWorkspace.ReadStoredCVBatch(ctx,
		r.PathValue("id"), r.PathValue("itemID"), input)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (a *App) importDigitalCenterCVs(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	var input struct {
		VehicleID string `json:"vehicleId"`
	}
	if !decodeDigitalCenterCVRequest(w, r, &input) {
		return
	}
	if input.VehicleID == "" {
		a.respondDigitalCenterWorkspaceError(w, application.ErrDigitalCenterFilterValidation)
		return
	}
	result, err := a.digitalCenterWorkspace.ImportStoredCVs(r.Context(),
		r.PathValue("id"), r.PathValue("itemID"), input.VehicleID)
	if err != nil {
		a.respondDigitalCenterWorkspaceError(w, err)
		return
	}
	a.recordAudit(r, "VehicleCVValuesImported", "vehicle", input.VehicleID)
	respondJSON(w, http.StatusOK, result)
}

func decodeDigitalCenterCVRequest(w http.ResponseWriter, r *http.Request, input any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Invalid CV request.")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Expected one JSON document.")
		return false
	}
	return true
}
