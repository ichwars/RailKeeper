package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"railkeeper/backend/internal/application"
)

func decodeCatalogRequest(w http.ResponseWriter, r *http.Request, input any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(input); err != nil {
		return err
	}
	if err := decoder.Decode(new(struct{})); err != io.EOF {
		return errors.New("catalog request must contain one JSON object")
	}
	return nil
}

func (a *App) createVehicleCatalogRevision(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleCatalogRevisionInput
	if err := decodeCatalogRequest(w, r, &input); err != nil {
		respondProblem(w, http.StatusBadRequest, "catalog_validation", "Use one valid catalog JSON object with documented fields.")
		return
	}
	result, err := a.vehicleService.CreateCatalogRevision(r.Context(), r.PathValue("targetType"), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.respondCatalogWriteError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, result)
}

func (a *App) adoptVehicleCatalogRevision(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleCatalogAdoptionInput
	if err := decodeCatalogRequest(w, r, &input); err != nil {
		respondProblem(w, http.StatusBadRequest, "catalog_validation", "Use one valid catalog JSON object with documented fields.")
		return
	}
	result, err := a.vehicleService.AdoptCatalogRevision(r.Context(), r.PathValue("targetType"), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.respondCatalogWriteError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) respondCatalogWriteError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrVehicleCatalogConflict):
		respondProblem(w, http.StatusConflict, "catalog_conflict", "The comparison changed or this revision is already saved. Reload the assignment and review before confirming again.")
	case errors.Is(err, application.ErrVehicleCatalogValidation):
		respondProblem(w, http.StatusBadRequest, "catalog_validation", "Provide a current comparison, valid revision identity, complete catalog fields and explicit field selection.")
	case errors.Is(err, application.ErrVehicleCatalogNotFound):
		respondProblem(w, http.StatusNotFound, "catalog_not_found", "Catalog target or article revision not found.")
	default:
		a.logger.Error("catalog change failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "catalog_write_failed", "Could not confirm catalog change. Reload and review the saved assignment before retrying.")
	}
}
