package api

import (
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) vehicleReportData(w http.ResponseWriter, r *http.Request) {
	var input struct {
		VehicleIDs []string `json:"vehicleIds"`
	}
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	if input.VehicleIDs == nil {
		respondProblem(w, http.StatusBadRequest, "vehicle_ids_required", "vehicleIds must be an array.")
		return
	}
	vehicles, err := a.vehicleService.VehiclesForReport(r.Context(), input.VehicleIDs)
	if errors.Is(err, application.ErrVehicleNotFound) {
		respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		return
	}
	if err != nil {
		a.logger.Error("vehicle report data failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_list_failed", "Could not list vehicles.")
		return
	}
	respondJSON(w, http.StatusOK, vehicles)
}
