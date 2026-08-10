package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

const maxTrackLibraryImportBytes = 4 * 1024 * 1024

func (a *App) listTrackLibraries(w http.ResponseWriter, r *http.Request) {
	libraries, err := a.trackLibraryService.List(r.Context())
	if err != nil {
		a.trackLibraryError(w, err, "list track libraries")
		return
	}
	respondJSON(w, http.StatusOK, libraries)
}

func (a *App) exportTrackLibrary(w http.ResponseWriter, r *http.Request) {
	doc, err := a.trackLibraryService.Export(r.Context(), r.PathValue("id"))
	if err != nil {
		a.trackLibraryError(w, err, "export track library")
		return
	}
	w.Header().Set("Content-Disposition", `attachment; filename="railkeeper-track-library.json"`)
	respondJSON(w, http.StatusOK, doc)
}

func (a *App) previewTrackLibraryImport(w http.ResponseWriter, r *http.Request) {
	var doc domain.TrackLibraryPackage
	if !decodeTrackLibraryJSON(w, r, &doc) {
		return
	}
	preview, err := a.trackLibraryService.PreviewImport(r.Context(), doc)
	if err != nil {
		a.trackLibraryError(w, err, "preview track library import")
		return
	}
	respondJSON(w, http.StatusOK, preview)
}

func (a *App) importTrackLibrary(w http.ResponseWriter, r *http.Request) {
	var input application.ImportTrackLibraryInput
	if !decodeTrackLibraryJSON(w, r, &input) {
		return
	}
	library, err := a.trackLibraryService.Import(r.Context(), input, actorUserID(r))
	if err != nil {
		a.trackLibraryError(w, err, "import track library")
		return
	}
	respondJSON(w, http.StatusCreated, library)
}

func (a *App) updateTrackLibraryStatus(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateTrackLibraryStatusInput
	if !decodeTrackLibraryJSON(w, r, &input) {
		return
	}
	library, err := a.trackLibraryService.UpdateStatus(
		r.Context(), r.PathValue("id"), input, actorUserID(r),
	)
	if err != nil {
		a.trackLibraryError(w, err, "update track library status")
		return
	}
	respondJSON(w, http.StatusOK, library)
}

func decodeTrackLibraryJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxTrackLibraryImportBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must contain one JSON document.")
		return false
	}
	return true
}

func (a *App) trackLibraryError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrTrackLibraryValidation):
		respondProblem(w, http.StatusBadRequest, "track_library_validation", "Track library data is invalid.")
	case errors.Is(err, application.ErrTrackLibraryConflict):
		respondProblem(w, http.StatusConflict, "track_library_conflict", "Track library version already exists.")
	case errors.Is(err, application.ErrTrackLibraryNotFound):
		respondProblem(w, http.StatusNotFound, "track_library_not_found", "Track library was not found.")
	default:
		a.logger.Error("track library operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "track_library_operation_failed", "Track library operation failed.")
	}
}
