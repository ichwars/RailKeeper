package api

import "net/http"

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
