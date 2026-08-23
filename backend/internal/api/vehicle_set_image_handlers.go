package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

func (a *App) setVehicleSetMainImage(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleSetMainImageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	set, err := a.vehicleService.SetSetMainImage(r.Context(), r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.respondVehicleSetImageError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, set)
}

func (a *App) uploadVehicleSetImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxImageBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxImageBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "image_upload_invalid", "Bild konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "image_missing", "Eine Bilddatei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	if header.Size > a.maxImageBytes {
		respondProblem(w, http.StatusBadRequest, "image_too_large", "Das Bild ist zu groß.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxImageBytes+1))
	if err != nil || int64(len(data)) > a.maxImageBytes {
		respondProblem(w, http.StatusBadRequest, "image_too_large", "Das Bild ist zu groß.")
		return
	}
	mimeType := http.DetectContentType(data)
	if !isAllowedImageMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "image_type_blocked", "Erlaubt sind JPG, PNG und WebP.")
		return
	}
	if err := validateVehicleImageDimensions(data); err != nil {
		respondProblem(w, http.StatusBadRequest, "image_dimensions_invalid", "Bildabmessungen überschreiten das erlaubte Limit.")
		return
	}
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(header.Filename))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("vehicle set image blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_upload_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	thumbnailBlobID, thumbnailErr := a.createVehicleImageThumbnail(r.Context(), data, storageName)
	if thumbnailErr != nil {
		a.logger.Warn("vehicle set image thumbnail skipped", "file", header.Filename, "error", thumbnailErr)
	}
	set, replaced, err := a.vehicleService.UpsertSetImage(r.Context(), r.PathValue("id"), application.VehicleSetImageInput{
		FileName: storageName, MimeType: mimeType, BlobID: blobID, ThumbnailBlobID: thumbnailBlobID,
	}, actorUserID(r))
	for _, replacedBlobID := range replaced {
		a.deleteFileBlobIfUnreferenced(r.Context(), replacedBlobID)
	}
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		a.deleteFileBlobIfUnreferenced(r.Context(), thumbnailBlobID)
		a.respondVehicleSetImageError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, set)
}

func (a *App) deleteVehicleSetImage(w http.ResponseWriter, r *http.Request) {
	blobIDs, err := a.vehicleService.DeleteSetImage(r.Context(), r.PathValue("id"), actorUserID(r))
	if err != nil {
		a.respondVehicleSetImageError(w, err)
		return
	}
	for _, blobID := range blobIDs {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) downloadVehicleSetImage(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.GetSetImage(r.Context(), r.PathValue("id"))
	if err != nil {
		a.respondVehicleSetImageError(w, err)
		return
	}
	data, err := a.loadFileBlob(r.Context(), image.BlobID)
	if err != nil {
		a.logger.Error("vehicle set image blob load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_download_failed", "Bild konnte nicht geladen werden.")
		return
	}
	serveFileBytes(w, r, data, image.MimeType, "inline", path.Base(image.FileName))
}

func (a *App) downloadVehicleSetImageThumbnail(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.GetSetImage(r.Context(), r.PathValue("id"))
	if err != nil {
		a.respondVehicleSetImageError(w, err)
		return
	}
	if image.ThumbnailBlobID == "" {
		a.downloadVehicleSetImage(w, r)
		return
	}
	data, err := a.loadFileBlob(r.Context(), image.ThumbnailBlobID)
	if err != nil {
		a.logger.Error("vehicle set image thumbnail load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_thumbnail_failed", "Bildvorschau konnte nicht geladen werden.")
		return
	}
	fileName := strings.TrimSuffix(path.Base(image.FileName), path.Ext(image.FileName)) + "-thumb.jpg"
	serveFileBytes(w, r, data, "image/jpeg", "inline", fileName)
}

func (a *App) respondVehicleSetImageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrVehicleSetNotFound):
		respondProblem(w, http.StatusNotFound, "vehicle_set_not_found", "Vehicle set not found.")
	case errors.Is(err, application.ErrVehicleSetImageNotFound):
		respondProblem(w, http.StatusNotFound, "vehicle_set_image_not_found", "Set image not found.")
	case errors.Is(err, application.ErrVehicleSetImageValidation):
		respondProblem(w, http.StatusBadRequest, "vehicle_set_image_validation", "Das ausgewählte Setbild ist ungültig.")
	default:
		a.logger.Error("vehicle set image operation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "vehicle_set_image_failed", "Das Setbild konnte nicht verarbeitet werden.")
	}
}
