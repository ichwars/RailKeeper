package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"railkeeper/backend/internal/application"
	"time"
)

func (a *App) exportMasterData(w http.ResponseWriter, r *http.Request) {
	doc, err := a.masterDataService.Export(r.Context())
	if err != nil {
		a.logger.Error("master data export failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_export_failed", "Stammdaten konnten nicht exportiert werden.")
		return
	}
	filename := "railkeeper-stammdaten-" + time.Now().UTC().Format("20060102-150405") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		a.logger.Error("master data encode failed", "error", err)
	}
}

func (a *App) importMasterData(w http.ResponseWriter, r *http.Request) {
	doc, ok := a.readMasterDataImportUpload(w, r)
	if !ok {
		return
	}
	result, err := a.masterDataService.Import(r.Context(), doc)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_import_invalid",
				masterDataValidationMessageOr(err, "Stammdaten-Datei ist ungültig."))
			return
		}
		a.logger.Error("master data import failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_import_failed", "Stammdaten konnten nicht importiert werden.")
		return
	}
	if err := a.masterDataService.RefreshCache(r.Context()); err != nil {
		a.logger.Error("master data cache refresh after import failed", "error", err)
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) readMasterDataImportUpload(w http.ResponseWriter, r *http.Request) (*application.MasterDataDocument, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMasterDataImportBytes+1024*1024)
	if err := r.ParseMultipartForm(maxMasterDataImportBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei konnte nicht gelesen werden.")
		return nil, false
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_file_missing", "Eine Stammdaten-Datei ist erforderlich.")
		return nil, false
	}
	defer func() { _ = file.Close() }()
	if header.Size > maxMasterDataImportBytes {
		respondProblem(w, http.StatusBadRequest, "master_data_file_too_large", "Die Stammdaten-Datei ist zu gro?.")
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, maxMasterDataImportBytes+1))
	if err != nil || int64(len(data)) > maxMasterDataImportBytes {
		respondProblem(w, http.StatusBadRequest, "master_data_file_too_large", "Die Stammdaten-Datei ist zu gro?.")
		return nil, false
	}
	var doc application.MasterDataDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei ist ungültig.")
		return nil, false
	}
	return &doc, true
}
