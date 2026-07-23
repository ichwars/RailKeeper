package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

type importVehicleAttachmentInput struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Category      string `json:"category"`
	MaintenanceID string `json:"maintenanceId"`
}

func (a *App) uploadVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, a.maxAttachmentBytes+1024*1024)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "attachment_upload_invalid", "Beilage konnte nicht gelesen werden.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "attachment_missing", "Eine Datei ist erforderlich.")
		return
	}
	defer func() { _ = file.Close() }()
	originalName := cleanOriginalFileName(header.Filename)
	if header.Size > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "attachment_too_large", "Die Datei ist zu gro?.")
		return
	}
	if isBlockedAttachmentName(originalName) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Ausführbare Dateien sind nicht erlaubt.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "attachment_too_large", "Die Datei ist zu gro?.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "attachment_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if isBlockedAttachmentMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Ausführbare Dateien sind nicht erlaubt.")
		return
	}
	if !a.isAllowedAttachmentUpload(originalName, mimeType) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Erlaubt sind PDF, TXT, CSV, JSON, XML, ZIP sowie JPG, PNG und WebP.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(originalName))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("attachment blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_upload_failed", "Beilage konnte nicht gespeichert werden.")
		return
	}
	attachment, err := a.vehicleService.CreateAttachment(r.Context(), vehicleID, application.VehicleAttachmentInput{
		FileName:      storageName,
		OriginalName:  originalName,
		Description:   r.FormValue("description"),
		Category:      r.FormValue("category"),
		MimeType:      mimeType,
		SizeBytes:     int64(len(data)),
		BlobID:        blobID,
		MaintenanceID: r.FormValue("maintenanceId"),
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("attachment metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_upload_failed", "Beilage konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, attachment)
}

func (a *App) importVehicleAttachmentFromURL(w http.ResponseWriter, r *http.Request) {
	var input importVehicleAttachmentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" || !isPublicImageURL(r.Context(), input.URL) {
		respondProblem(w, http.StatusBadRequest, "attachment_url_invalid", "Dokument-URL ist nicht erreichbar oder nicht erlaubt.")
		return
	}
	requestCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, input.URL, nil)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "attachment_url_invalid", "Dokument-URL ist ung?ltig.")
		return
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 document-fetch")
	req.Header.Set("Accept", "application/pdf,text/plain,application/json,application/xml,text/xml,application/zip,image/*;q=0.8,*/*;q=0.4")
	client := remoteDocumentHTTPClient(r.Context())
	resp, err := client.Do(req)
	if err != nil {
		a.logger.Warn("remote attachment fetch failed", "url", input.URL, "error", err)
		respondProblem(w, http.StatusBadGateway, "attachment_fetch_failed", "Dokument konnte nicht heruntergeladen werden.")
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		respondProblem(w, http.StatusBadGateway, "attachment_fetch_failed", "Dokument konnte nicht heruntergeladen werden.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, a.maxAttachmentBytes+1))
	if err != nil || int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusBadRequest, "attachment_too_large", "Die Datei ist zu gro?.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "attachment_empty", "Leere Dateien sind nicht erlaubt.")
		return
	}
	mimeType := http.DetectContentType(data)
	if headerMime := strings.TrimSpace(resp.Header.Get("Content-Type")); headerMime != "" && (strings.Contains(headerMime, "pdf") || strings.Contains(headerMime, "zip") || strings.Contains(headerMime, "xml")) {
		mimeType = strings.Split(headerMime, ";")[0]
	}
	originalName := remoteAttachmentFileName(input, input.URL, mimeType)
	if isBlockedAttachmentName(originalName) || isBlockedAttachmentMime(mimeType) || !a.isAllowedAttachmentUpload(originalName, mimeType) {
		respondProblem(w, http.StatusBadRequest, "attachment_type_blocked", "Erlaubt sind PDF, TXT, CSV, JSON, XML, ZIP sowie JPG, PNG und WebP.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(originalName))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("remote attachment blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_import_failed", "Dokument konnte nicht gespeichert werden.")
		return
	}
	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = attachmentCategoryForRemoteDocument(originalName, input.Title)
	}
	attachment, err := a.vehicleService.CreateAttachment(r.Context(), vehicleID, application.VehicleAttachmentInput{
		FileName:      storageName,
		OriginalName:  originalName,
		Description:   strings.TrimSpace(input.Description),
		Category:      category,
		MimeType:      mimeType,
		SizeBytes:     int64(len(data)),
		BlobID:        blobID,
		MaintenanceID: strings.TrimSpace(input.MaintenanceID),
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		if errors.Is(err, application.ErrVehicleValidation) {
			respondProblem(w, http.StatusBadRequest, "attachment_invalid", "Beilage ist unvollst?ndig.")
			return
		}
		a.logger.Error("remote attachment metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_import_failed", "Dokument konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, attachment)
}

func (a *App) updateVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	var input application.VehicleAttachmentUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	attachment, err := a.vehicleService.UpdateAttachment(r.Context(), r.PathValue("id"), r.PathValue("attachmentID"), input)
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			return
		}
		a.logger.Error("attachment update failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_update_failed", "Beilage konnte nicht aktualisiert werden.")
		return
	}
	respondJSON(w, http.StatusOK, attachment)
}

func (a *App) deleteVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	attachment, err := a.vehicleService.DeleteAttachment(r.Context(), r.PathValue("id"), r.PathValue("attachmentID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			return
		}
		a.logger.Error("attachment delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_delete_failed", "Beilage konnte nicht gelöscht werden.")
		return
	}
	if fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath); err == nil {
		_ = os.Remove(fullPath)
	}
	a.deleteFileBlobIfUnreferenced(r.Context(), attachment.BlobID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) downloadVehicleAttachment(w http.ResponseWriter, r *http.Request) {
	attachment, err := a.vehicleService.GetAttachment(r.Context(), r.PathValue("id"), r.PathValue("attachmentID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			return
		}
		a.logger.Error("attachment download lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "attachment_download_failed", "Beilage konnte nicht geladen werden.")
		return
	}
	if attachment.MimeType != "" {
		w.Header().Set("Content-Type", attachment.MimeType)
	}
	disposition := "attachment"
	inlinePreview := r.URL.Query().Get("inline") == "true" && canPreviewAttachmentInline(attachment.MimeType, attachment.OriginalName)
	if inlinePreview {
		disposition = "inline"
		if shouldSandboxInlineAttachment(attachment.MimeType, attachment.OriginalName) {
			w.Header().Set("Content-Security-Policy", "sandbox; default-src 'none'; img-src 'self' data: blob:; style-src 'unsafe-inline'")
		}
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": cleanOriginalFileName(attachment.OriginalName)}))
	if attachment.BlobID != "" {
		data, err := a.loadFileBlob(r.Context(), attachment.BlobID)
		if err != nil {
			a.logger.Error("attachment blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "attachment_download_failed", "Beilage konnte nicht geladen werden.")
			return
		}
		http.ServeContent(w, r, attachment.FileName, time.Now().UTC(), bytes.NewReader(data))
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "attachment_path_invalid", "Beilage konnte nicht geladen werden.")
		return
	}
	http.ServeFile(w, r, fullPath)
}

func canPreviewAttachmentInline(mimeType, fileName string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	fileName = strings.ToLower(strings.TrimSpace(fileName))
	if strings.Contains(mimeType, "pdf") || strings.HasPrefix(mimeType, "image/") || strings.HasPrefix(mimeType, "text/") {
		return true
	}
	if strings.Contains(mimeType, "json") || strings.Contains(mimeType, "xml") {
		return true
	}
	switch path.Ext(fileName) {
	case ".pdf", ".jpg", ".jpeg", ".png", ".webp", ".txt", ".csv", ".json", ".xml", ".html", ".htm":
		return true
	default:
		return false
	}
}

func shouldSandboxInlineAttachment(mimeType, fileName string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	fileName = strings.ToLower(strings.TrimSpace(fileName))
	if strings.HasPrefix(mimeType, "text/html") || strings.HasSuffix(fileName, ".html") || strings.HasSuffix(fileName, ".htm") {
		return true
	}
	return false
}

func (a *App) readAttachmentData(ctx context.Context, attachment application.VehicleAttachment, maxBytes int64) ([]byte, error) {
	if attachment.BlobID != "" {
		data, err := a.loadFileBlob(ctx, attachment.BlobID)
		if err != nil {
			return nil, err
		}
		if maxBytes > 0 && int64(len(data)) > maxBytes {
			return nil, fmt.Errorf("attachment blob exceeds read limit")
		}
		return data, nil
	}
	fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("attachment file exceeds read limit")
	}
	return data, nil
}
