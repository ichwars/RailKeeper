package api

import "net/http"

func (a *App) overviewValuation(w http.ResponseWriter, r *http.Request) {
	if a.overviewValuationService == nil {
		respondProblem(w, http.StatusServiceUnavailable, "overview_valuation_unavailable",
			"Bestandsbewertung ist nicht verfügbar.")
		return
	}
	valuation, err := a.overviewValuationService.Get(r.Context())
	if err != nil {
		a.logger.Error("overview valuation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "overview_valuation_failed",
			"Bestandsbewertung konnte nicht berechnet werden.")
		return
	}
	respondJSON(w, http.StatusOK, valuation)
}
