package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

var (
	errRemoteAccessoryImageTooLarge        = errors.New("remote accessory image too large")
	errRemoteAccessoryImageTypeUnsupported = errors.New("remote accessory image type unsupported")
)

type accessoryDocumentImportURLInput struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPrimary   bool   `json:"isPrimary"`
}

func downloadRemoteAccessoryImage(
	ctx context.Context,
	client *http.Client,
	rawURL string,
	maxBytes int64,
) ([]byte, string, error) {
	// The caller validates the initial URL, while safefetch's transport independently validates
	// every request, redirect, DNS result, and dial target immediately before network access.
	// codeql[go/request-forgery]
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "RailKeeper/0.1 image-fetch")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8")
	response, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return nil, "", fmt.Errorf("remote image returned status %d", response.StatusCode)
	}
	if response.ContentLength > maxBytes {
		return nil, "", errRemoteAccessoryImageTooLarge
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxBytes+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) == 0 {
		return nil, "", errors.New("remote image is empty")
	}
	if int64(len(data)) > maxBytes {
		return nil, "", errRemoteAccessoryImageTooLarge
	}
	mimeType := http.DetectContentType(data)
	if !isAllowedImageMime(mimeType) {
		return nil, "", fmt.Errorf("%w: %s", errRemoteAccessoryImageTypeUnsupported, mimeType)
	}
	return data, mimeType, nil
}

func (a *App) importAccessoryDocumentFromURL(w http.ResponseWriter, r *http.Request) {
	var input accessoryDocumentImportURLInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	input.URL = strings.TrimSpace(input.URL)
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)
	if input.URL == "" {
		respondProblem(w, http.StatusBadRequest, "accessory_image_url_missing", "An image URL is required.")
		return
	}
	if _, err := a.accessoryService.GetProduct(r.Context(), r.PathValue("id")); err != nil {
		a.accessoryError(w, err, "get accessory product for remote image")
		return
	}
	if !isPublicImageURL(r.Context(), input.URL) {
		respondProblem(w, http.StatusBadRequest, "accessory_image_url_invalid",
			"Image URL must be a public HTTP or HTTPS URL.")
		return
	}

	requestCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	data, mimeType, err := downloadRemoteAccessoryImage(
		requestCtx, remoteDocumentHTTPClient(requestCtx), input.URL, a.maxAttachmentBytes,
	)
	if err != nil {
		switch {
		case errors.Is(err, errRemoteAccessoryImageTooLarge):
			respondProblem(w, http.StatusRequestEntityTooLarge, "accessory_document_too_large",
				"Accessory image exceeds the upload limit.")
		case errors.Is(err, errRemoteAccessoryImageTypeUnsupported):
			respondProblem(w, http.StatusUnsupportedMediaType, "accessory_image_type_unsupported",
				"Remote file is not a supported image.")
		default:
			a.logger.Warn("remote accessory image import failed", "url", input.URL, "error", err)
			respondProblem(w, http.StatusBadGateway, "accessory_image_import_failed",
				"Image could not be downloaded.")
		}
		return
	}

	fileName := remoteImageFileName(application.VehicleImageInput{URL: input.URL, Title: input.Title}, mimeType)
	metadata := application.AccessoryDocumentUploadMetadata{
		FileName: fileName, OriginalName: fileName, Category: application.AccessoryDocumentImage,
		MimeType: mimeType, SizeBytes: int64(len(data)), IsPrimary: input.IsPrimary,
	}
	if err := application.ValidateAccessoryDocumentUpload(metadata, a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusUnsupportedMediaType, "accessory_image_type_unsupported",
			"Remote file is not a supported image.")
		return
	}
	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("remote accessory image blob store failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "accessory_image_import_failed",
			"Image could not be stored.")
		return
	}
	document, err := a.accessoryDocumentService.CreateDocument(r.Context(), application.CreateAccessoryDocumentInput{
		ProductID: r.PathValue("id"), FileBlobID: blobID, AccessoryDocumentUploadMetadata: metadata,
		Description: input.Description,
	}, a.maxAttachmentBytes, actorUserID(r))
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		a.accessoryError(w, err, "create remote accessory image")
		return
	}
	respondJSON(w, http.StatusCreated, document)
}
