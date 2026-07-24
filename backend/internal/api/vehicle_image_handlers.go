package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/safefetch"
)

func effectiveLimit(value, fallback int64) int64 {
	if value <= 0 {
		return fallback
	}
	return value
}

var safeFileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

var allowedAttachmentExtensions = map[string]struct{}{
	".csv":  {},
	".jpeg": {},
	".jpg":  {},
	".json": {},
	".pdf":  {},
	".png":  {},
	".txt":  {},
	".webp": {},
	".xml":  {},
	".zip":  {},
}

func effectiveAttachmentExtensions(input map[string]struct{}) map[string]struct{} {
	if len(input) == 0 {
		return allowedAttachmentExtensions
	}
	out := map[string]struct{}{}
	for extension := range input {
		extension = strings.ToLower(strings.TrimSpace(extension))
		if extension == "" {
			continue
		}
		if !strings.HasPrefix(extension, ".") {
			extension = "." + extension
		}
		if isBlockedAttachmentName("file" + extension) {
			continue
		}
		out[extension] = struct{}{}
	}
	if len(out) == 0 {
		return allowedAttachmentExtensions
	}
	return out
}

func (a *App) localizeVehicleImages(ctx context.Context, vehicleID string, images []application.VehicleImageInput) ([]application.VehicleImageInput, error) {
	out := make([]application.VehicleImageInput, len(images))
	copy(out, images)
	for index, image := range out {
		if image.StoragePath != "" || image.BlobID != "" || !strings.HasPrefix(strings.ToLower(image.URL), "http") {
			continue
		}
		localized, err := a.localizeVehicleImage(ctx, vehicleID, image)
		if err != nil {
			a.logger.Warn("article image localization skipped", "url", image.URL, "error", err)
			continue
		}
		out[index] = localized
	}
	return out, nil
}

func (a *App) localizeVehicleImage(ctx context.Context, vehicleID string, image application.VehicleImageInput) (application.VehicleImageInput, error) {
	if !isPublicImageURL(ctx, image.URL) {
		return image, fmt.Errorf("image url is not public http(s)")
	}
	requestCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, http.MethodGet, image.URL, nil)
	if err != nil {
		return image, err
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 image-fetch")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8")
	client := remoteDocumentHTTPClient(ctx)
	client.Timeout = 6 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		return image, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return image, fmt.Errorf("image fetch returned status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, a.maxImageBytes+1))
	if err != nil || len(data) == 0 || int64(len(data)) > a.maxImageBytes {
		return image, fmt.Errorf("image size invalid")
	}
	mimeType := http.DetectContentType(data)
	if !isAllowedImageMime(mimeType) {
		return image, fmt.Errorf("image type %s is not allowed", mimeType)
	}
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), remoteImageFileName(image, mimeType))
	blobID, err := a.storeFileBlob(ctx, data)
	if err != nil {
		return image, err
	}
	thumbnailBlobID, err := a.createVehicleImageThumbnail(ctx, data, storageName)
	if err != nil {
		a.logger.Warn("image thumbnail skipped", "url", image.URL, "error", err)
	}
	if image.SourceURL == "" {
		image.SourceURL = image.URL
	}
	image.FileName = storageName
	image.MimeType = mimeType
	image.BlobID = blobID
	image.ThumbnailBlobID = thumbnailBlobID
	return image, nil
}

func remoteImageFileName(image application.VehicleImageInput, mimeType string) string {
	extension := ".jpg"
	switch mimeType {
	case "image/png":
		extension = ".png"
	case "image/webp":
		extension = ".webp"
	}
	base := strings.TrimSpace(image.Title)
	if base == "" {
		if parsed, err := url.Parse(image.URL); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" || base == "." || base == "/" {
		base = "artikelbild"
	}
	return safeAttachmentFileName(base + extension)
}

func remoteAttachmentFileName(input importVehicleAttachmentInput, rawURL, mimeType string) string {
	base := strings.TrimSpace(input.Title)
	if base == "" {
		if parsed, err := url.Parse(rawURL); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	if base == "" || base == "." || base == "/" {
		base = "dokument"
	}
	extension := strings.ToLower(filepath.Ext(base))
	if extension == "" {
		extension = attachmentExtensionForMime(mimeType)
		base += extension
	}
	return safeAttachmentFileName(base)
}

func attachmentExtensionForMime(mimeType string) string {
	switch strings.ToLower(strings.Split(mimeType, ";")[0]) {
	case "application/pdf":
		return ".pdf"
	case "application/json", "text/json":
		return ".json"
	case "application/xml", "text/xml":
		return ".xml"
	case "application/zip", "application/x-zip-compressed":
		return ".zip"
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "text/csv":
		return ".csv"
	default:
		return ".txt"
	}
}

func attachmentCategoryForRemoteDocument(fileName, title string) string {
	lower := strings.ToLower(fileName + " " + title)
	if strings.Contains(lower, "ersatzteil") || strings.Contains(lower, "spare") || strings.Contains(lower, "et-blatt") {
		return "Ersatzteilliste"
	}
	if strings.Contains(lower, "anleitung") || strings.Contains(lower, "manual") || strings.Contains(lower, "bedienung") {
		return "Anleitung"
	}
	return "Dokumentation"
}

func isPublicImageURL(ctx context.Context, value string) bool {
	return safefetch.IsPublicHTTPURL(ctx, value)
}

func remoteDocumentHTTPClient(ctx context.Context) *http.Client {
	return safefetch.NewHTTPClient(ctx, safefetch.Options{Timeout: 10 * time.Second, MaxRedirects: 5})
}

func (a *App) uploadVehicleImage(w http.ResponseWriter, r *http.Request) {
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
		respondProblem(w, http.StatusBadRequest, "image_too_large", "Das Bild ist zu gro?.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxImageBytes+1))
	if err != nil || int64(len(data)) > a.maxImageBytes {
		respondProblem(w, http.StatusBadRequest, "image_too_large", "Das Bild ist zu gro?.")
		return
	}
	mimeType := http.DetectContentType(data)
	if !isAllowedImageMime(mimeType) {
		respondProblem(w, http.StatusBadRequest, "image_type_blocked", "Erlaubt sind JPG, PNG und WebP.")
		return
	}
	vehicleID := r.PathValue("id")
	storageName := fmt.Sprintf("%d-%s", time.Now().UTC().UnixNano(), safeAttachmentFileName(header.Filename))
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("image blob write failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_upload_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	thumbnailBlobID, err := a.createVehicleImageThumbnail(r.Context(), data, storageName)
	if err != nil {
		a.logger.Warn("image thumbnail skipped", "file", header.Filename, "error", err)
	}
	image, err := a.vehicleService.CreateImage(r.Context(), vehicleID, application.VehicleImageInput{
		Title:           r.FormValue("title"),
		SourceURL:       r.FormValue("sourceUrl"),
		FileName:        storageName,
		MimeType:        mimeType,
		BlobID:          blobID,
		ThumbnailBlobID: thumbnailBlobID,
		MaintenanceID:   r.FormValue("maintenanceId"),
		IsPrimary:       strings.EqualFold(r.FormValue("isPrimary"), "true"),
	})
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		a.deleteFileBlobIfUnreferenced(r.Context(), thumbnailBlobID)
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		a.logger.Error("image metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_upload_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, image)
}

type importVehicleImageInput struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	SourceURL     string `json:"sourceUrl"`
	MaintenanceID string `json:"maintenanceId"`
	IsPrimary     bool   `json:"isPrimary"`
	SortOrder     int    `json:"sortOrder"`
}

func (a *App) importVehicleImageFromURL(w http.ResponseWriter, r *http.Request) {
	var input importVehicleImageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	if input.URL == "" {
		respondProblem(w, http.StatusBadRequest, "image_url_missing", "Eine Bild-URL ist erforderlich.")
		return
	}
	vehicleID := r.PathValue("id")
	localized, err := a.localizeVehicleImage(r.Context(), vehicleID, application.VehicleImageInput{
		URL:           input.URL,
		Title:         input.Title,
		SourceURL:     input.SourceURL,
		MaintenanceID: input.MaintenanceID,
		IsPrimary:     input.IsPrimary,
		SortOrder:     input.SortOrder,
	})
	if err != nil {
		a.logger.Warn("remote image import failed", "url", input.URL, "error", err)
		respondProblem(w, http.StatusBadGateway, "image_import_failed", "Bild konnte nicht heruntergeladen werden.")
		return
	}
	image, err := a.vehicleService.CreateImage(r.Context(), vehicleID, localized)
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), localized.BlobID)
		a.deleteFileBlobIfUnreferenced(r.Context(), localized.ThumbnailBlobID)
		if fullPath, pathErr := confinedDataPath(a.dataDir, localized.StoragePath); pathErr == nil {
			_ = os.Remove(fullPath)
		}
		if fullPath, pathErr := confinedDataPath(a.dataDir, localized.ThumbnailPath); pathErr == nil {
			_ = os.Remove(fullPath)
		}
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "vehicle_not_found", "Vehicle not found.")
			return
		}
		if errors.Is(err, application.ErrVehicleValidation) {
			respondProblem(w, http.StatusBadRequest, "image_invalid", "Bilddaten sind unvollst?ndig.")
			return
		}
		a.logger.Error("remote image metadata create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_import_failed", "Bild konnte nicht gespeichert werden.")
		return
	}
	respondJSON(w, http.StatusCreated, image)
}

func (a *App) deleteVehicleImage(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.DeleteImage(r.Context(), r.PathValue("id"), r.PathValue("imageID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "image_not_found", "Image not found.")
			return
		}
		if errors.Is(err, application.ErrVehicleImageInUse) {
			respondProblem(w, http.StatusConflict, "image_in_use", "Bild ist mit einem Wartungseintrag verknüpft. Bitte zuerst die Verknüpfung entfernen.")
			return
		}
		a.logger.Error("image delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_delete_failed", "Bild konnte nicht gelöscht werden.")
		return
	}
	a.removeVehicleImageFileIfUnreferenced(r.Context(), image.StoragePath)
	a.removeVehicleImageFileIfUnreferenced(r.Context(), image.ThumbnailPath)
	a.deleteFileBlobIfUnreferenced(r.Context(), image.BlobID)
	a.deleteFileBlobIfUnreferenced(r.Context(), image.ThumbnailBlobID)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) removeVehicleImageFileIfUnreferenced(ctx context.Context, storagePath string) {
	if storagePath == "" {
		return
	}
	references, err := a.vehicleService.ImageFileReferenceCount(ctx, storagePath)
	if err != nil {
		a.logger.Warn("image file reference check failed", "path", storagePath, "error", err)
		return
	}
	if references > 0 {
		return
	}
	if fullPath, err := confinedDataPath(a.dataDir, storagePath); err == nil {
		_ = os.Remove(fullPath)
	}
}

func (a *App) storeFileBlob(ctx context.Context, data []byte) (string, error) {
	if a.fileBlobs == nil {
		return "", errors.New("file blob service is not configured")
	}
	return a.fileBlobs.Store(ctx, data)
}

func (a *App) loadFileBlob(ctx context.Context, blobID string) ([]byte, error) {
	if a.fileBlobs == nil {
		return nil, errors.New("file blob service is not configured")
	}
	return a.fileBlobs.Load(ctx, blobID)
}

func (a *App) deleteFileBlobIfUnreferenced(ctx context.Context, blobID string) {
	if blobID == "" || a.fileBlobs == nil {
		return
	}
	if err := a.fileBlobs.DeleteIfUnreferenced(ctx, blobID); err != nil {
		a.logger.Warn("file blob cleanup failed", "blobID", blobID, "error", err)
	}
}

func serveFileBytes(w http.ResponseWriter, r *http.Request, data []byte, mimeType, disposition, fileName string) {
	if mimeType != "" {
		w.Header().Set("Content-Type", mimeType)
	}
	if fileName != "" {
		w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": cleanOriginalFileName(fileName)}))
	}
	http.ServeContent(w, r, fileName, time.Now().UTC(), bytes.NewReader(data))
}

func (a *App) downloadVehicleImage(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.GetImage(r.Context(), r.PathValue("id"), r.PathValue("imageID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "image_not_found", "Image not found.")
			return
		}
		a.logger.Error("image lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_download_failed", "Bild konnte nicht geladen werden.")
		return
	}
	if image.BlobID != "" {
		data, err := a.loadFileBlob(r.Context(), image.BlobID)
		if err != nil {
			a.logger.Error("image blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "image_download_failed", "Bild konnte nicht geladen werden.")
			return
		}
		serveFileBytes(w, r, data, image.MimeType, "inline", path.Base(image.FileName))
		return
	}
	if image.StoragePath == "" {
		respondProblem(w, http.StatusNotFound, "image_file_missing", "Bilddatei ist nicht lokal gespeichert.")
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, image.StoragePath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "image_path_invalid", "Bild konnte nicht geladen werden.")
		return
	}
	if image.MimeType != "" {
		w.Header().Set("Content-Type", image.MimeType)
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": path.Base(image.FileName)}))
	http.ServeFile(w, r, fullPath)
}

func (a *App) downloadVehicleImageThumbnail(w http.ResponseWriter, r *http.Request) {
	image, err := a.vehicleService.GetImage(r.Context(), r.PathValue("id"), r.PathValue("imageID"))
	if err != nil {
		if errors.Is(err, application.ErrVehicleNotFound) {
			respondProblem(w, http.StatusNotFound, "image_not_found", "Image not found.")
			return
		}
		a.logger.Error("image thumbnail lookup failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "image_thumbnail_failed", "Bildvorschau konnte nicht geladen werden.")
		return
	}
	if image.ThumbnailBlobID != "" {
		data, err := a.loadFileBlob(r.Context(), image.ThumbnailBlobID)
		if err != nil {
			a.logger.Error("image thumbnail blob load failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "image_thumbnail_failed", "Bildvorschau konnte nicht geladen werden.")
			return
		}
		serveFileBytes(w, r, data, "image/jpeg", "inline", strings.TrimSuffix(path.Base(image.FileName), path.Ext(image.FileName))+"-thumb.jpg")
		return
	}
	if image.ThumbnailPath == "" {
		a.downloadVehicleImage(w, r)
		return
	}
	fullPath, err := confinedDataPath(a.dataDir, image.ThumbnailPath)
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "image_path_invalid", "Bildvorschau konnte nicht geladen werden.")
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": strings.TrimSuffix(path.Base(image.FileName), path.Ext(image.FileName)) + "-thumb.jpg"}))
	http.ServeFile(w, r, fullPath)
}

func (a *App) createVehicleImageThumbnail(ctx context.Context, data []byte, storageName string) (string, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	thumb := scaleImageToFit(src, 360, 240)
	var out bytes.Buffer
	if err = jpeg.Encode(&out, thumb, &jpeg.Options{Quality: 82}); err != nil {
		return "", err
	}
	return a.storeFileBlob(ctx, out.Bytes())
}

func scaleImageToFit(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 || (width <= maxWidth && height <= maxHeight) {
		return src
	}
	ratioW := float64(maxWidth) / float64(width)
	ratioH := float64(maxHeight) / float64(height)
	ratio := ratioW
	if ratioH < ratio {
		ratio = ratioH
	}
	dstWidth := max(1, int(float64(width)*ratio))
	dstHeight := max(1, int(float64(height)*ratio))
	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	for y := range dstHeight {
		srcY := bounds.Min.Y + y*height/dstHeight
		for x := range dstWidth {
			srcX := bounds.Min.X + x*width/dstWidth
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}
