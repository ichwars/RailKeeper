package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"railkeeper/backend/internal/application"
)

func (a *App) listVehicleCVValues(w http.ResponseWriter, r *http.Request) {
	values, err := a.vehicleService.ListCVValues(r.Context(), r.PathValue("id"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("cv value list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_value_list_failed", "CV-Werte konnten nicht geladen werden.")
		return
	}
	respondJSON(w, http.StatusOK, values)
}

func (a *App) createVehicleCVValue(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleCVValueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	value, err := a.vehicleService.CreateCVValue(r.Context(), r.PathValue("id"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "cv_value_invalid", "CV-Nummer muss 1-1024 und Wert 0-255 sein.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
		default:
			a.logger.Error("cv value create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_value_create_failed", "CV-Wert konnte nicht gespeichert werden.")
		}
		return
	}
	respondJSON(w, http.StatusCreated, value)
}

func (a *App) updateVehicleCVValue(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleCVValueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	value, err := a.vehicleService.UpdateCVValue(r.Context(), r.PathValue("id"), r.PathValue("cvValueID"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrVehicleValidation):
			respondProblem(w, http.StatusBadRequest, "cv_value_invalid", "CV-Nummer muss 1-1024 und Wert 0-255 sein.")
		case errors.Is(err, application.ErrVehicleNotFound):
			respondProblem(w, http.StatusNotFound, "cv_value_not_found", "CV value not found.")
		default:
			a.logger.Error("cv value update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_value_update_failed", "CV-Wert konnte nicht aktualisiert werden.")
		}
		return
	}
	respondJSON(w, http.StatusOK, value)
}

func (a *App) deleteVehicleCVValue(w http.ResponseWriter, r *http.Request) {
	if _, err := a.vehicleService.DeleteCVValue(r.Context(), r.PathValue("id"), r.PathValue("cvValueID")); err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "cv_value_not_found", "CV value not found.")
			return
		}
		a.logger.Error("cv value delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_value_delete_failed", "CV-Wert konnte nicht gelöscht werden.")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) previewVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxAttachmentBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_preview_invalid", "CV-Datei konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_missing", "Eine Datei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	originalName := cleanOriginalFileName(header.Filename)
	if header.Size > a.maxAttachmentBytes || isBlockedAttachmentName(originalName) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "cv_file_too_large", "Die Datei ist zu groß.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "cv_file_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if isBlockedAttachmentMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	respondJSON(w, http.StatusOK, esuxPreviewResponse(originalName, int64(len(data)), mimeType, data))
}

func (a *App) uploadVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxAttachmentBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_upload_invalid", "CV-Datei konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "cv_file_missing", "Eine Datei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	originalName := cleanOriginalFileName(header.Filename)
	if header.Size > a.maxAttachmentBytes || isBlockedAttachmentName(originalName) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "cv_file_too_large", "Die Datei ist zu gro?.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "cv_file_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if isBlockedAttachmentMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "cv_file_blocked", "Diese CV-Datei ist nicht erlaubt.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(originalName))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("cv file blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_upload_failed", "CV-Datei konnte nicht gespeichert werden.")
		return
	}
	decoderProfile, description := applyESUXMetadata(originalName, data, r.FormValue("decoderProfile"), r.FormValue("description"))
	cvFile, err := a.vehicleService.CreateCVFile(r.Context(), vehicleID, application.VehicleCVFileInput{
		FileName:       storageName,
		OriginalName:   originalName,
		Description:    description,
		DecoderProfile: decoderProfile,
		MimeType:       mimeType,
		SizeBytes:      int64(len(data)),
		BlobID:         blobID,
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("cv file metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_upload_failed", "CV-Datei konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, cvFile)
}

func (a *App) deleteVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.vehicleService.DeleteCVFile(r.Context(), r.PathValue("id"), r.PathValue("cvFileID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "cv_file_not_found", "CV file not found.")
			return
		}
		a.logger.Error("cv file delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_delete_failed", "CV-Datei konnte nicht gelöscht werden.")
		return
	}
	if fullPath, err := confinedDataPath(a.dataDir, file.StoragePath); err == nil {
		_ = os.Remove(fullPath)
	}
	a.deleteFileBlobIfUnreferenced(r.Context(), file.BlobID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) downloadVehicleCVFile(w http.ResponseWriter, r *http.Request) {
	file, err := a.vehicleService.GetCVFile(r.Context(), r.PathValue("id"), r.PathValue("cvFileID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "cv_file_not_found", "CV file not found.")
			return
		}
		a.logger.Error("cv file download lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "cv_file_download_failed", "CV-Datei konnte nicht geladen werden.")
		return
	}
	if file.MimeType != "" {
		w.Header().Set("Content-Type", file.MimeType)
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": cleanOriginalFileName(file.OriginalName)}))
	if file.BlobID != "" {
		data, err := a.loadFileBlob(r.Context(), file.BlobID)
		if err != nil {
			a.logger.Error("cv file blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "cv_file_download_failed", "CV-Datei konnte nicht geladen werden.")
			return
		}
		http.ServeContent(w, r, file.FileName, time.Now().UTC(), bytes.NewReader(data))
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, file.StoragePath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "cv_file_path_invalid", "CV-Datei konnte nicht geladen werden.")
		return
	}
	http.ServeFile(w, r, fullPath)
}

const maxBackupBytes = 250 * 1024 * 1024
const maxMasterDataImportBytes = 25 * 1024 * 1024
