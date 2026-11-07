package api

import (
	"bytes"
	"errors"
	"mime"
	"net/http"
	"time"

	"railkeeper/backend/internal/application"
)

func (a *App) decoderConfigurationContext(w http.ResponseWriter, r *http.Request) {
	result, err := a.vehicleService.DecoderConfigurationContext(r.Context(), r.PathValue("id"))
	if err != nil {
		a.respondDecoderConfigurationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) listDecoderCVSnapshots(w http.ResponseWriter, r *http.Request) {
	result, err := a.vehicleService.ListDecoderCVSnapshots(r.Context(), r.URL.Query().Get("vehicleId"))
	if err != nil {
		a.respondDecoderConfigurationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) decoderCVSnapshot(w http.ResponseWriter, r *http.Request) {
	result, err := a.vehicleService.DecoderCVSnapshot(r.Context(), r.PathValue("snapshotID"))
	if err != nil {
		a.respondDecoderConfigurationError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) downloadDecoderSnapshotSource(w http.ResponseWriter, r *http.Request) {
	file, data, err := a.vehicleService.DecoderCVSnapshotSourceContent(r.Context(), r.PathValue("snapshotID"), r.PathValue("sourceID"))
	if err != nil {
		a.respondDecoderConfigurationError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": cleanOriginalFileName(file.OriginalName)}))
	http.ServeContent(w, r, file.FileName, time.Time{}, bytes.NewReader(data))
}

func (a *App) respondDecoderConfigurationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrDecoderNotFound):
		respondProblem(w, http.StatusNotFound, "decoder_not_found", "Decoder configuration or historical source not found.")
	case errors.Is(err, application.ErrDecoderSourceUnavailable):
		respondProblem(w, http.StatusGone, "decoder_source_unavailable", "The recorded original is unavailable. No other file has been substituted.")
	case errors.Is(err, application.ErrDecoderValidation):
		respondProblem(w, http.StatusBadRequest, "decoder_validation", "Use the documented fields, valid dates and an installation in this vehicle.")
	case errors.Is(err, application.ErrDecoderConflict):
		respondProblem(w, http.StatusConflict, "decoder_conflict", "The comparison changed or this identity is already saved. Reload and review before confirming again.")
	default:
		a.logger.Error("decoder configuration failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "decoder_configuration_failed", "Could not complete the decoder request. Reload the saved state before retrying.")
	}
}
