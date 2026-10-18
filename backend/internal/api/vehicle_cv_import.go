package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) importVehicleCVValues(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8*1024*1024)
	var body struct {
		Values []json.RawMessage `json:"values"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Invalid CV import JSON (maximum 8 MiB).")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must contain one JSON document.")
		return
	}
	if len(body.Values) == 0 || len(body.Values) > application.MaxVehicleCVImportValues {
		respondProblem(w, http.StatusBadRequest, "cv_import_invalid", "Expected 1 to 10000 CV values.")
		return
	}
	input := application.VehicleCVImportInput{}
	for _, raw := range body.Values {
		var required struct {
			CVNumber *int `json:"cvNumber"`
			Value    *int `json:"value"`
		}
		var value application.VehicleCVValueInput
		rowDecoder := json.NewDecoder(bytes.NewReader(raw))
		rowDecoder.DisallowUnknownFields()
		if err := rowDecoder.Decode(&value); err != nil {
			respondProblem(w, http.StatusBadRequest, "cv_import_invalid", "Invalid CV fields or unknown field names.")
			return
		}
		if err := json.Unmarshal(raw, &required); err != nil || required.CVNumber == nil || required.Value == nil {
			respondProblem(w, http.StatusBadRequest, "cv_import_invalid", "Each CV requires an explicit numeric cvNumber and value.")
			return
		}
		input.Values = append(input.Values, value)
	}
	count, err := a.vehicleService.ImportCVValues(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "cv_import_invalid", "Invalid CV values, duplicate keys or incomplete index pair (maximum 10000 values).")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("cv import failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_import_failed", "CV import failed; no values were changed.")
		}
		return
	}
	a.recordAudit(r, "VehicleCVValuesImported", "vehicle", r.PathValue("id"))
	respondJSON(w, http.StatusOK, map[string]int{"imported": count})
}
