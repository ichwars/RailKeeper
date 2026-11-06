package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) applyVehicleCVChanges(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
	var body struct {
		Changes []json.RawMessage `json:"changes"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_changes_invalid", "Expected a CV draft JSON document (maximum 2 MiB).")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		respondProblem(w, http.StatusBadRequest, "cv_changes_invalid", "Expected one CV draft JSON document.")
		return
	}
	input := application.VehicleCVChangesInput{}
	if len(body.Changes) == 0 || len(body.Changes) > application.MaxVehicleCVChanges {
		respondProblem(w, http.StatusBadRequest, "cv_changes_invalid", "Expected 1 to 500 CV changes.")
		return
	}
	for _, raw := range body.Changes {
		var change application.VehicleCVChange
		row := json.NewDecoder(bytes.NewReader(raw))
		row.DisallowUnknownFields()
		var required struct {
			ExpectedRevision *string `json:"expectedRevision"`
			Value            *struct {
				CVNumber *int `json:"cvNumber"`
				Value    *int `json:"value"`
			} `json:"value"`
		}
		if err := row.Decode(&change); err != nil {
			respondProblem(w, http.StatusBadRequest, "cv_changes_invalid", "Invalid CV fields or unknown field names.")
			return
		}
		if err := json.Unmarshal(raw, &required); err != nil || required.ExpectedRevision == nil ||
			(!change.Delete && (required.Value == nil || required.Value.CVNumber == nil || required.Value.Value == nil)) {
			respondProblem(w, http.StatusBadRequest, "cv_changes_invalid", "Each change requires an explicit revision and each value an explicit CV number and value.")
			return
		}
		input.Changes = append(input.Changes, change)
	}
	count, err := a.vehicleService.ApplyCVChanges(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "cv_changes_invalid", "Expected 1 to 500 distinct valid CV changes with complete index pairs and revisions.")
		case errors.Is(err, application.ErrVehicleCVConflict):
			respondProblem(w, http.StatusConflict, "cv_value_conflict", "CV data changed or the target identity already exists. No draft changes were saved.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("CV draft failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_changes_failed", "CV draft could not be saved; no draft changes were saved.")
		}
		return
	}
	a.recordAudit(r, "VehicleCVChangesApplied", "vehicle", r.PathValue("id"))
	respondJSON(w, http.StatusOK, map[string]int{"changed": count})
}
