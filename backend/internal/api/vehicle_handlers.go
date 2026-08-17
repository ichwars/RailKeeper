package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) listVehicles(w http.ResponseWriter, r *http.Request) {
	vehicles, err := a.vehicleService.List(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		a.logger.Error("vehicle list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_list_failed", "Could not list vehicles.")
		return
	}

	respondJSON(w, http.StatusOK, vehicles)
}

func (a *App) createVehicle(w http.ResponseWriter, r *http.Request) {
	var input application.CreateVehicleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	vehicle, err := a.vehicleService.Create(r.Context(), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleOperationalValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_operational_validation",
				"Maximum speed must be a whole number between 1 and 1000 km/h, and home depot / operating location must not exceed 200 characters.")
		case errors.Is(err, application.ErrVehicleValidation), errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_validation", "Manufacturer, name, gauge, category and subtype are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_conflict", "Inventory number already exists.")
		case errors.Is(err, application.ErrInventoryNumberNotFound):
			respondProblem(w, http.StatusBadRequest, "inventory_number_scheme_missing", "No active inventory number scheme is available.")
		default:
			a.logger.Error("vehicle create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "vehicle_create_failed", "Could not create vehicle.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, vehicle)
}

func (a *App) createVehicleSet(w http.ResponseWriter, r *http.Request) {
	var input application.CreateVehicleSetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	vehicleSet, err := a.vehicleService.CreateSet(r.Context(), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleSetValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_set_validation",
				"Set name, manufacturer, gauge, category, subtype and between 2 and 100 members are required.")
		case errors.Is(err, application.ErrVehicleOperationalValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_operational_validation",
				"Maximum speed must be a whole number between 1 and 1000 km/h, and home depot / operating location must not exceed 200 characters.")
		case errors.Is(err, application.ErrVehicleValidation), errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_validation",
				"Every member requires a name and a valid inventory number when one is specified.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_conflict", "Inventory number already exists.")
		case errors.Is(err, application.ErrInventoryNumberNotFound):
			respondProblem(w, http.StatusBadRequest, "inventory_number_scheme_missing",
				"No active inventory number scheme is available.")
		default:
			a.logger.Error("vehicle set create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "vehicle_set_create_failed",
				"Could not create vehicle set.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, vehicleSet)
}

func (a *App) getVehicleSet(w http.ResponseWriter, r *http.Request) {
	vehicleSet, err := a.vehicleService.GetSet(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleSetNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_set_not_found", "Vehicle set not found.")
			return
		}
		a.logger.Error("vehicle set get failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_set_get_failed", "Could not read vehicle set.")
		return
	}
	respondJSON(w, http.StatusOK, vehicleSet)
}

func (a *App) updateVehicleSet(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleSetInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	vehicleSet, err := a.vehicleService.UpdateSet(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleSetValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_set_validation",
				"Set name, manufacturer, gauge, category and subtype are required.")
		case errors.Is(err, application.ErrVehicleSetNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_set_not_found", "Vehicle set not found.")
		default:
			a.logger.Error("vehicle set update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "vehicle_set_update_failed",
				"Could not update vehicle set.")
		}
		return
	}
	respondJSON(w, http.StatusOK, vehicleSet)
}

func (a *App) getVehicle(w http.ResponseWriter, r *http.Request) {
	vehicle, err := a.vehicleService.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("vehicle get failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_get_failed", "Could not read vehicle.")
		return
	}

	respondJSON(w, http.StatusOK, vehicle)
}

func (a *App) updateVehicle(w http.ResponseWriter, r *http.Request) {
	var input application.CreateVehicleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	vehicle, err := a.vehicleService.Update(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleOperationalValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_operational_validation",
				"Maximum speed must be a whole number between 1 and 1000 km/h, and home depot / operating location must not exceed 200 characters.")
		case errors.Is(err, application.ErrVehicleValidation), errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "vehicle_validation", "Manufacturer, name, gauge, category and subtype are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_conflict", "Inventory number already exists.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("vehicle update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "vehicle_update_failed", "Could not update vehicle.")
		}
		return
	}

	respondJSON(w, http.StatusOK, vehicle)
}

func (a *App) upsertVehicleExternalMapping(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleExternalMapInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	mapping, err := a.vehicleService.UpsertExternalMapping(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "external_mapping_validation", "Provider and external id are required.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("vehicle external mapping failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "external_mapping_failed", "Could not save external mapping.")
		}
		return
	}

	respondJSON(w, http.StatusOK, mapping)
}

func (a *App) deleteVehicle(w http.ResponseWriter, r *http.Request) {
	if err := a.vehicleService.Delete(r.Context(), r.PathValue("id"), actorUserID(r)); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("vehicle delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_delete_failed", "Could not delete vehicle.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

const (
	defaultMaxAttachmentBytes = 25 * 1024 * 1024
	defaultMaxImageBytes      = 10 * 1024 * 1024
)
