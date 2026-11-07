package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"railkeeper/backend/internal/application"
)

const maxCVImportDataBytes = 8 * 1024 * 1024
const maxCVImportRevisionBytes = application.MaxVehicleCVImportValues * (128 + len(`"expectedRevision":"","expectedWorkspaceRevision":""`) + 2)

func (a *App) importVehicleCVValues(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxCVImportDataBytes+maxCVImportRevisionBytes))
	var body struct {
		Values []json.RawMessage `json:"values"`
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Invalid CV import JSON or comparison overhead exceeds its limit.")
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
	dataBytes, err := cvImportDataSize(body.Values)
	if err != nil || dataBytes > maxCVImportDataBytes {
		respondProblem(w, http.StatusBadRequest, "cv_import_invalid", "CV import data exceeds 8 MiB excluding comparison versions.")
		return
	}
	count, err := a.vehicleService.ImportCVValues(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleCVConflict):
			respondProblem(w, http.StatusConflict, "cv_value_conflict", "CV data changed. No values were changed.")
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

// Keep the original business-data budget when row comparison versions add transport overhead.
func cvImportDataSize(rows []json.RawMessage) (int, error) {
	size := len(`{"values":[]}`)
	if len(rows) > 1 {
		size += len(rows) - 1
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	for _, row := range rows {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(row, &fields); err != nil {
			return 0, err
		}
		for key := range fields {
			if strings.EqualFold(key, "expectedRevision") || strings.EqualFold(key, "expectedWorkspaceRevision") {
				delete(fields, key)
			}
		}
		buffer.Reset()
		if err := encoder.Encode(fields); err != nil {
			return 0, err
		}
		size += buffer.Len() - 1 // Encoder appends a newline outside the original JSON object.
		if size > maxCVImportDataBytes {
			return size, nil
		}
	}
	return size, nil
}
