package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"railkeeper/backend/internal/application"
)

const masterDataProtectedMessage = "Standard article type keys cannot be created, changed, or deleted."

func (a *App) listMasterData(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("management") == "true" {
		items, err := a.masterDataService.ListForManagement(r.Context(), r.PathValue("type"))
		if err != nil {
			if errors.Is(err, application.ErrMasterDataValidation) {
				respondProblem(w, http.StatusBadRequest, "master_data_validation", "Master data type is required.")
				return
			}
			a.logger.Error("managed master data list failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
			return
		}
		respondJSON(w, http.StatusOK, items)
		return
	}
	activeOnly := r.URL.Query().Get("active") == "true"
	items, err := a.masterDataService.List(r.Context(), r.PathValue("type"), activeOnly)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", "Master data type is required.")
			return
		}
		a.logger.Error("master data list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) listAllMasterData(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("management") == "true" {
		items, err := a.masterDataService.ListAllForManagement(r.Context())
		if err != nil {
			a.logger.Error("managed master data list all failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
			return
		}
		respondJSON(w, http.StatusOK, items)
		return
	}
	activeOnly := r.URL.Query().Get("active") == "true"
	items, err := a.masterDataService.ListAll(r.Context(), activeOnly)
	if err != nil {
		a.logger.Error("master data list all failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) createMasterData(w http.ResponseWriter, r *http.Request) {
	var input application.MasterDataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	item, err := a.masterDataService.Create(r.Context(), r.PathValue("type"), input)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataProtected) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataProtectedMessage)
			return
		}
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataValidationMessage(err))
			return
		}
		a.logger.Error("master data create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_create_failed", "Could not create master data.")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

func (a *App) updateMasterData(w http.ResponseWriter, r *http.Request) {
	var input application.MasterDataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	item, err := a.masterDataService.Update(r.Context(), r.PathValue("type"), r.PathValue("key"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrMasterDataProtected):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataProtectedMessage)
		case errors.Is(err, application.ErrMasterDataValidation):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataValidationMessage(err))
		case errors.Is(err, application.ErrMasterDataNotFound):
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
		default:
			a.logger.Error("master data update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_update_failed", "Could not update master data.")
		}
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func masterDataValidationMessage(err error) string {
	return masterDataValidationMessageOr(err, "Label is required.")
}

func masterDataValidationMessageOr(err error, fallback string) string {
	if err != application.ErrMasterDataValidation {
		return err.Error()
	}
	return fallback
}

type masterDataActiveInput struct {
	Active *bool `json:"active"`
}

type masterDataActiveBatchInput struct {
	Keys   []string `json:"keys"`
	Active *bool    `json:"active"`
}

func (a *App) setMasterDataActive(w http.ResponseWriter, r *http.Request) {
	var input masterDataActiveInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if input.Active == nil {
		respondProblem(w, http.StatusBadRequest, "master_data_validation", "Active state is required.")
		return
	}
	item, err := a.masterDataService.SetActive(
		r.Context(),
		r.PathValue("type"),
		r.PathValue("key"),
		*input.Active,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrMasterDataValidation):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", "Master data type and key are required.")
		case errors.Is(err, application.ErrMasterDataNotFound):
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
		default:
			a.logger.Error("master data active-state update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_update_failed",
				"Could not update master data active state.")
		}
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (a *App) setMasterDataActiveMany(w http.ResponseWriter, r *http.Request) {
	var input masterDataActiveBatchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	if input.Active == nil {
		respondProblem(w, http.StatusBadRequest, "master_data_validation", "Active state is required.")
		return
	}
	if *input.Active {
		respondProblem(w, http.StatusBadRequest, "master_data_validation", "Bulk reactivation is not supported.")
		return
	}
	items, err := a.masterDataService.SetActiveMany(
		r.Context(),
		r.PathValue("type"),
		input.Keys,
		false,
	)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrMasterDataValidation):
			respondProblem(w, http.StatusBadRequest, "master_data_validation",
				"Master data type and between 1 and 5000 non-empty keys are required.")
		case errors.Is(err, application.ErrMasterDataNotFound):
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
		default:
			a.logger.Error("master data active-state batch update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_update_failed",
				"Could not update master data active state.")
		}
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (a *App) deleteMasterData(w http.ResponseWriter, r *http.Request) {
	if err := a.masterDataService.Delete(r.Context(), r.PathValue("type"), r.PathValue("key")); err != nil {
		switch {
		case errors.Is(err, application.ErrMasterDataBundled):
			respondProblem(w, http.StatusConflict, "master_data_bundled",
				"Bundled master data cannot be permanently deleted.")
		case errors.Is(err, application.ErrMasterDataInUse):
			respondProblem(w, http.StatusConflict, "master_data_in_use",
				"This master data entry is still in use and can only be deactivated.")
		case errors.Is(err, application.ErrMasterDataProtected):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataProtectedMessage)
		case errors.Is(err, application.ErrMasterDataValidation):
			respondProblem(w, http.StatusBadRequest, "master_data_validation",
				masterDataValidationMessageOr(err, "This master data entry cannot be deleted."))
		case errors.Is(err, application.ErrMasterDataNotFound):
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
		default:
			a.logger.Error("master data delete failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_delete_failed", "Could not delete master data.")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listMasterDataRelations(w http.ResponseWriter, r *http.Request) {
	relations, err := a.masterDataService.Relations(
		r.Context(),
		r.URL.Query().Get("parentType"),
		r.URL.Query().Get("childType"),
	)
	if err != nil {
		a.logger.Error("master data relations list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_relations_failed", "Could not list master data relations.")
		return
	}

	respondJSON(w, http.StatusOK, relations)
}
