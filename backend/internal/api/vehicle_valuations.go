package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"railkeeper/backend/internal/valuation"
)

func (a *App) valuationAvailable(w http.ResponseWriter) bool {
	if a.valuationService != nil {
		return true
	}
	respondProblem(w, http.StatusServiceUnavailable, "valuation_unavailable", "Manual valuations are unavailable.")
	return false
}

func (a *App) vehicleValuations(w http.ResponseWriter, r *http.Request) {
	if !a.valuationAvailable(w) {
		return
	}
	offset := 0
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			a.respondValuationError(w, valuation.ErrValidation)
			return
		}
		offset = value
	}
	result, err := a.valuationService.History(r.Context(), r.PathValue("id"), offset)
	if err != nil {
		a.respondValuationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) recordVehicleValuation(w http.ResponseWriter, r *http.Request) {
	if !a.valuationAvailable(w) {
		return
	}
	var input valuation.Input
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 32*1024))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		a.respondValuationError(w, valuation.ErrValidation)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		a.respondValuationError(w, valuation.ErrValidation)
		return
	}
	result, err := a.valuationService.Record(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.respondValuationError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, result)
}

func (a *App) valuationSummary(w http.ResponseWriter, r *http.Request) {
	if !a.valuationAvailable(w) {
		return
	}
	result, err := a.valuationService.Summary(r.Context(), r.URL.Query().Get("asOf"))
	if err != nil {
		a.respondValuationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) exportVehicleValuations(w http.ResponseWriter, r *http.Request) {
	if !a.valuationAvailable(w) {
		return
	}
	result, err := a.valuationService.Export(r.Context(), r.PathValue("id"))
	if err != nil {
		a.respondValuationError(w, err)
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="railkeeper-valuations.json"`)
	respondJSON(w, http.StatusOK, result)
}

func (a *App) respondValuationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, valuation.ErrValidation):
		respondProblem(w, http.StatusBadRequest, "valuation_invalid", "Check date, exact ungrouped amount, currency code, provenance and request identity.")
	case errors.Is(err, valuation.ErrNotFound):
		respondProblem(w, http.StatusNotFound, "valuation_not_found", "Vehicle or valuation not found.")
	case errors.Is(err, valuation.ErrConflict):
		respondProblem(w, http.StatusConflict, "valuation_conflict", "The valuation was corrected or this request identity was already used. Reload and compare before confirming again.")
	default:
		a.logger.Error("manual valuation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "valuation_failed", "Could not complete the valuation request. Reload the saved state before retrying.")
	}
}
