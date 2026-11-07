package api

import (
	"errors"
	"net/http"

	"railkeeper/backend/internal/application"
)

func (a *App) vehicleCatalogContext(w http.ResponseWriter, r *http.Request) {
	result, err := a.vehicleService.CatalogContext(r.Context(), r.PathValue("targetType"), r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleCatalogValidation):
			respondProblem(w, http.StatusBadRequest, "catalog_validation", "Catalog target must be a vehicle or set.")
		case errors.Is(err, application.ErrVehicleCatalogNotFound):
			respondProblem(w, http.StatusNotFound, "catalog_not_found", "Catalog target not found.")
		default:
			respondProblem(w, http.StatusInternalServerError, "catalog_read_failed", "Could not read catalog assignment.")
		}
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) vehicleCatalogBackfillReport(w http.ResponseWriter, r *http.Request) {
	report, err := a.vehicleService.CatalogBackfillReport(r.Context())
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "catalog_report_failed", "Could not read catalog assignments.")
		return
	}
	respondJSON(w, http.StatusOK, report)
}

func (a *App) backfillVehicleCatalog(w http.ResponseWriter, r *http.Request) {
	report, err := a.vehicleService.BackfillLegacyCatalog(r.Context(), actorUserID(r))
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "catalog_backfill_failed", "Could not confirm catalog backfill. Existing inventory fields are unchanged; inspect the mapping report before retrying.")
		return
	}
	respondJSON(w, http.StatusOK, report)
}
