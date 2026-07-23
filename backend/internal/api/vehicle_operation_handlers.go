package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"railkeeper/backend/internal/application"
)

func (a *App) listVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	entries, err := a.vehicleService.ListMaintenance(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("maintenance list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "maintenance_list_failed", "Wartungseintraege konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

func (a *App) createVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleMaintenanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.CreateMaintenance(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "maintenance_invalid", "Wartungseintrag ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("maintenance create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "maintenance_create_failed", "Wartungseintrag konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}

func (a *App) updateVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleMaintenanceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.UpdateMaintenance(r.Context(), r.PathValue("id"), r.PathValue("maintenanceID"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "maintenance_invalid", "Wartungseintrag ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "maintenance_not_found", "Maintenance entry not found.")
		default:
			a.logger.Error("maintenance update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "maintenance_update_failed", "Wartungseintrag konnte nicht aktualisiert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (a *App) deleteVehicleMaintenance(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteMaintenance(r.Context(), r.PathValue("id"), r.PathValue("maintenanceID")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "maintenance_not_found", "Maintenance entry not found.")
			return
		}
		a.logger.Error("maintenance delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "maintenance_delete_failed", "Wartungseintrag konnte nicht gelöscht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listVehicleSpareParts(w http.ResponseWriter, r *http.Request) {
	entries, err := a.vehicleService.ListSpareParts(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("spare part list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "spare_part_list_failed", "Ersatzteile konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, entries)
}

func (a *App) suggestVehicleSpareParts(w http.ResponseWriter, r *http.Request) {
	vehicle, err := a.vehicleService.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("spare part suggestion vehicle lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "spare_part_suggestions_failed", "Ersatzteilvorschlaege konnten nicht geladen werden.")
		return
	}
	articleNumber := strings.TrimSpace(vehicle.ArticleNumber)
	if articleNumber == "" {
		respondJSON(w, http.StatusOK, []application.ArticleSearchSparePart{})
		return
	}
	seen := map[string]bool{}
	suggestions := []application.ArticleSearchSparePart{}
	attachmentID := strings.TrimSpace(r.URL.Query().Get("attachmentId"))
	for _, attachment := range vehicle.Attachments {
		if attachmentID != "" && attachment.ID != attachmentID {
			continue
		}
		if len(suggestions) >= 80 || !looksLikeSparePartAttachment(attachment) {
			continue
		}
		data, err := a.readAttachmentData(r.Context(), attachment, 12*1024*1024)
		if err != nil || len(data) == 0 {
			continue
		}
		downloadURL := "/api/v1/vehicles/" + url.PathEscape(vehicle.ID) + "/attachments/" + url.PathEscape(attachment.ID) + "/download"
		parts := application.ArticleSparePartsFromDocumentData(data, articleNumber, downloadURL)
		for _, part := range parts {
			part.Source = cleanOriginalFileName(attachment.OriginalName)
			if part.URL == "" {
				part.URL = downloadURL
			}
			key := strings.ToLower(part.ArticleNumber + "|" + part.Description + "|" + part.URL)
			if key == "||" || seen[key] {
				continue
			}
			seen[key] = true
			suggestions = append(suggestions, part)
			if len(suggestions) >= 80 {
				break
			}
		}
	}
	respondJSON(w, http.StatusOK, suggestions)
}

func looksLikeSparePartAttachment(attachment application.VehicleAttachment) bool {
	lower := strings.ToLower(attachment.Category + " " + attachment.OriginalName + " " + attachment.Description + " " + attachment.MimeType)
	return strings.Contains(lower, "ersatzteil") ||
		strings.Contains(lower, "spare") ||
		strings.Contains(lower, "et-blatt") ||
		strings.Contains(lower, "serviceblatt") ||
		strings.Contains(lower, "bedienungsanl") ||
		strings.Contains(lower, "manual") ||
		strings.Contains(lower, "pdf")
}

func (a *App) createVehicleSparePart(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleSparePartInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.CreateSparePart(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "spare_part_invalid", "Ersatzteil ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("spare part create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "spare_part_create_failed", "Ersatzteil konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusCreated, entry)
}

func (a *App) updateVehicleSparePart(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleSparePartInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	entry, err := a.vehicleService.UpdateSparePart(r.Context(), r.PathValue("id"), r.PathValue("sparePartID"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "spare_part_invalid", "Ersatzteil ist unvollst?ndig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "spare_part_not_found", "Spare part not found.")
		default:
			a.logger.Error("spare part update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "spare_part_update_failed", "Ersatzteil konnte nicht aktualisiert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, entry)
}

func (a *App) deleteVehicleSparePart(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteSparePart(r.Context(), r.PathValue("id"), r.PathValue("sparePartID")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "spare_part_not_found", "Spare part not found.")
			return
		}
		a.logger.Error("spare part delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "spare_part_delete_failed", "Ersatzteil konnte nicht gel?scht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listVehicleFunctions(w http.ResponseWriter, r *http.Request) {
	functions, err := a.vehicleService.ListFunctions(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("function list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "function_list_failed", "Digitalfunktionen konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, functions)
}

func (a *App) upsertVehicleFunction(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleFunctionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	function, err := a.vehicleService.UpsertFunction(r.Context(), r.PathValue("id"), r.PathValue("functionKey"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "function_invalid", "Digitalfunktion ist ungültig.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("function save failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "function_save_failed", "Digitalfunktion konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, function)
}

func (a *App) deleteVehicleFunction(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteFunction(r.Context(), r.PathValue("id"), r.PathValue("functionKey")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "function_not_found", "Function entry not found.")
			return
		}
		a.logger.Error("function delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "function_delete_failed", "Digitalfunktion konnte nicht gelöscht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
