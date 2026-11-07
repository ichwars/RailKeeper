package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"railkeeper/backend/internal/application"
)

func decodeDecoderConfigurationRequest(w http.ResponseWriter, r *http.Request, target any, required ...string) bool {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes))
	if err == nil {
		var fields map[string]json.RawMessage
		decoder := json.NewDecoder(bytes.NewReader(data))
		if err = decoder.Decode(&fields); err == nil && fields != nil && len(fields) == len(required) {
			if err = decoder.Decode(new(any)); err == io.EOF {
				valid := true
				for _, name := range required {
					value, exists := fields[name]
					nullable := name == "installedAt" || name == "removedAt" || name == "installationId"
					if !exists || (!nullable && bytes.Equal(bytes.TrimSpace(value), []byte("null"))) {
						valid = false
						break
					}
				}
				if valid && json.Unmarshal(data, target) == nil {
					return true
				}
			}
		}
	}
	respondProblem(w, http.StatusBadRequest, "decoder_validation", "Use one documented JSON object with every required field (maximum 64 KiB). Unknown dates and an unassigned installation use explicit null.")
	return false
}

func (a *App) respondDecoderConfigurationWrite(w http.ResponseWriter, status int, result *application.DecoderConfigurationContext, err error) {
	if err != nil {
		a.respondDecoderConfigurationError(w, err)
		return
	}
	respondJSON(w, status, result)
}

func (a *App) createDecoderInstance(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderInstanceInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "id", "expectedRevision", "label", "manufacturer", "model", "serialNumber", "notes") {
		return
	}
	result, err := a.vehicleService.CreateDecoderInstance(r.Context(), r.PathValue("id"), input, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusCreated, result, err)
}

func (a *App) updateDecoderInstance(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderInstanceInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "id", "expectedRevision", "label", "manufacturer", "model", "serialNumber", "notes") {
		return
	}
	if input.ID != r.PathValue("decoderID") {
		a.respondDecoderConfigurationError(w, application.ErrDecoderValidation)
		return
	}
	result, err := a.vehicleService.UpdateDecoderInstance(r.Context(), r.PathValue("id"), input, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusOK, result, err)
}

func (a *App) installDecoder(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderInstallationInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "id", "expectedRevision", "decoderId", "installedAt", "notes") {
		return
	}
	result, err := a.vehicleService.InstallDecoder(r.Context(), r.PathValue("id"), input, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusCreated, result, err)
}

func (a *App) removeDecoder(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderRemovalInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "expectedRevision", "removedAt", "note") {
		return
	}
	result, err := a.vehicleService.RemoveDecoder(r.Context(), r.PathValue("id"), r.PathValue("installationID"), input, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusOK, result, err)
}

func (a *App) correctDecoderInstallationDates(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderInstallationDatesInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "expectedRevision", "installedAt", "removedAt") {
		return
	}
	result, err := a.vehicleService.CorrectDecoderInstallationDates(r.Context(), r.PathValue("id"), r.PathValue("installationID"), input, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusOK, result, err)
}

func (a *App) createDecoderCVWorkspace(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderCVWorkspaceInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "id", "expectedRevision", "profile") {
		return
	}
	result, err := a.vehicleService.CreateDecoderCVWorkspace(r.Context(), r.PathValue("id"), input, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusCreated, result, err)
}

func (a *App) assignDecoderCVWorkspace(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderCVAssignmentInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "expectedRevision", "installationId", "snapshotId") {
		return
	}
	result, err := a.vehicleService.AssignDecoderCVWorkspace(r.Context(), r.PathValue("id"), r.PathValue("workspaceID"), input, a.dataDir, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusOK, result, err)
}

func (a *App) captureDecoderCVCheckpoint(w http.ResponseWriter, r *http.Request) {
	var input application.DecoderCVCheckpointInput
	if !decodeDecoderConfigurationRequest(w, r, &input, "id", "expectedRevision", "note") {
		return
	}
	result, err := a.vehicleService.CaptureDecoderCVCheckpoint(r.Context(), r.PathValue("id"), r.PathValue("workspaceID"), input, a.dataDir, actorUserID(r))
	a.respondDecoderConfigurationWrite(w, http.StatusCreated, result, err)
}
