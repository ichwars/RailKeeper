package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"railkeeper/backend/internal/application"
)

type eCoSLocomotiveSyncRequest struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	VehicleID string `json:"vehicleId"`
	ObjectID  int    `json:"objectId"`
	DryRun    bool   `json:"dryRun"`
	Confirm   bool   `json:"confirm"`
}

func (a *App) testECoSConnection(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.ecosService.TestConnection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "ecos_validation", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) probeECoSLocomotiveRaw(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	probe, err := a.ecosService.ProbeLocomotiveRaw(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_raw_probe_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, probe)
}

func (a *App) countECoSLocomotives(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	summary, err := a.ecosService.CountLocomotives(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_locomotive_count_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, summary)
}

func (a *App) syncECoSLocomotive(w http.ResponseWriter, r *http.Request) {
	var request eCoSLocomotiveSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if request.Confirm {
		respondProblem(w, http.StatusConflict, "digital_center_write_grant_required",
			"Create a fresh workspace preview and confirm it with its write grant.")
		return
	}
	vehicle, err := a.vehicleService.Get(r.Context(), request.VehicleID)
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("vehicle read for ecos sync failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_read_failed", "Could not read vehicle.")
		return
	}

	objectID := request.ObjectID
	mapping := vehicleECoSMappingForSync(vehicle, objectID)
	if objectID <= 0 && mapping != nil {
		objectID = parsePositiveIntText(mapping.ExternalID)
	}
	if objectID <= 0 {
		respondProblem(w, http.StatusBadRequest, "ecos_object_required", "ECoS object ID is required.")
		return
	}

	desired := application.ECoSLocomotiveSyncDesired{
		Name:     vehicle.Name,
		Address:  parsePositiveIntText(vehicle.DigitalDecoderNumber),
		Protocol: "",
	}
	if mapping != nil {
		if address := parsePositiveIntText(mapping.ExternalAddress); address > 0 {
			desired.Address = address
		}
		desired.Protocol = mapping.ExternalProtocol
	}

	result, err := a.ecosService.SyncLocomotive(r.Context(), application.ECoSLocomotiveSyncInput{
		Host:     request.Host,
		Port:     request.Port,
		ObjectID: objectID,
		Desired:  desired,
		DryRun:   request.DryRun,
		Confirm:  request.Confirm,
	})
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_sync_failed", err.Error())
		return
	}

	if (request.Confirm && len(result.Changes) == 0) || result.Applied {
		address := ""
		if desired.Address > 0 {
			address = strconv.Itoa(desired.Address)
		}
		if _, err := a.vehicleService.UpsertExternalMapping(r.Context(), vehicle.ID, application.VehicleExternalMapInput{
			Provider:         "ecos",
			ExternalID:       strconv.Itoa(objectID),
			ExternalName:     desired.Name,
			ExternalAddress:  address,
			ExternalProtocol: desired.Protocol,
			SyncStatus:       "synced",
		}, actorUserID(r)); err != nil {
			a.logger.Warn("ecos sync mapping update failed", "vehicleID", vehicle.ID, "objectID", objectID, "error", err)
		}
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) eCoSLiveStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, a.ecosService.LiveStatus())
}

func (a *App) startECoSLive(w http.ResponseWriter, r *http.Request) {
	var input application.ECoSConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	status, err := a.ecosService.StartLive(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadGateway, "ecos_live_start_failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, status)
}

func (a *App) stopECoSLive(w http.ResponseWriter, r *http.Request) {
	status := a.ecosService.StopLive()
	respondJSON(w, http.StatusOK, status)
}

func (a *App) testZ21Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.TestZ21Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) probeZ21Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.ProbeZ21Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) testIntellibox3Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.TestIntellibox3Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) probeIntellibox3Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.ProbeIntellibox3Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) testCS3Connection(w http.ResponseWriter, r *http.Request) {
	var input application.DigitalCenterConnectionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	result, err := a.digitalCenterService.TestCS3Connection(r.Context(), input)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "digital_center_validation", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, result)
}
