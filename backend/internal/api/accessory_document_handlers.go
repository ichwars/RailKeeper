package api

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"railkeeper/backend/internal/application"
)

const accessoryMultipartOverheadBytes int64 = 1024 * 1024

func (a *App) listAccessoryDocuments(w http.ResponseWriter, r *http.Request) {
	productID := r.PathValue("id")
	if _, err := a.accessoryService.GetProduct(r.Context(), productID); err != nil {
		a.accessoryError(w, err, "get accessory product for documents")
		return
	}
	documents, err := a.accessoryDocumentService.ListDocuments(r.Context(), productID)
	if err != nil {
		a.accessoryError(w, err, "list accessory documents")
		return
	}
	respondJSON(w, http.StatusOK, documents)
}

func (a *App) getAccessoryDocument(w http.ResponseWriter, r *http.Request) {
	document, ok := a.accessoryDocumentForProduct(w, r)
	if !ok {
		return
	}
	respondJSON(w, http.StatusOK, document)
}

func (a *App) uploadAccessoryDocument(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		a.maxAttachmentBytes+accessoryMultipartOverheadBytes,
	)
	if err := r.ParseMultipartForm(a.maxAttachmentBytes); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			respondProblem(w, http.StatusRequestEntityTooLarge, "accessory_document_too_large",
				"Accessory document exceeds the upload limit.")
			return
		}
		respondProblem(w, http.StatusBadRequest, "accessory_document_multipart_invalid",
			"Accessory document upload must be valid multipart data.")
		return
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}

	category := application.AccessoryDocumentCategory(strings.TrimSpace(r.FormValue("category")))
	isPrimary, valid := parseOptionalAccessoryBoolean(r.FormValue("isPrimary"))
	if !valid || !category.Valid() || (isPrimary && category != application.AccessoryDocumentImage) {
		respondProblem(w, http.StatusBadRequest, "accessory_document_metadata_invalid",
			"Accessory document metadata is invalid.")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "accessory_document_file_required",
			"Accessory document file is required.")
		return
	}
	defer func() { _ = file.Close() }()
	if header.Size > a.maxAttachmentBytes {
		respondProblem(w, http.StatusRequestEntityTooLarge, "accessory_document_too_large",
			"Accessory document exceeds the upload limit.")
		return
	}
	data, err := io.ReadAll(io.LimitReader(file, a.maxAttachmentBytes+1))
	if err != nil {
		a.logger.Error("accessory document read failed", "error", err)
		respondProblem(w, http.StatusBadRequest, "accessory_document_read_failed",
			"Accessory document could not be read.")
		return
	}
	if int64(len(data)) > a.maxAttachmentBytes {
		respondProblem(w, http.StatusRequestEntityTooLarge, "accessory_document_too_large",
			"Accessory document exceeds the upload limit.")
		return
	}
	if len(data) == 0 {
		respondProblem(w, http.StatusBadRequest, "accessory_document_empty",
			"Accessory document must not be empty.")
		return
	}

	originalName := header.Filename
	mimeType := http.DetectContentType(data)
	metadata := application.AccessoryDocumentUploadMetadata{
		FileName: originalName, OriginalName: originalName, Category: category,
		MimeType: mimeType, SizeBytes: int64(len(data)), IsPrimary: isPrimary,
	}
	if err := application.ValidateAccessoryDocumentUpload(metadata, a.maxAttachmentBytes); err != nil {
		respondProblem(w, http.StatusUnsupportedMediaType, "accessory_document_type_unsupported",
			"Accessory document file type is not supported.")
		return
	}

	blobID, err := a.storeFileBlob(r.Context(), data)
	if err != nil {
		a.logger.Error("accessory document blob store failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "accessory_document_store_failed",
			"Accessory document could not be stored.")
		return
	}
	document, err := a.accessoryDocumentService.CreateDocument(r.Context(), application.CreateAccessoryDocumentInput{
		ProductID: r.PathValue("id"), FileBlobID: blobID, AccessoryDocumentUploadMetadata: metadata,
		Description: r.FormValue("description"),
	}, a.maxAttachmentBytes, actorUserID(r))
	if err != nil {
		a.deleteFileBlobIfUnreferenced(r.Context(), blobID)
		a.accessoryError(w, err, "create accessory document")
		return
	}
	respondJSON(w, http.StatusCreated, document)
}

func (a *App) updateAccessoryDocument(w http.ResponseWriter, r *http.Request) {
	var input application.UpdateAccessoryDocumentInput
	if !decodeAccessoryJSON(w, r, &input) {
		return
	}
	if _, ok := a.accessoryDocumentForProduct(w, r); !ok {
		return
	}
	document, err := a.accessoryDocumentService.UpdateDocument(
		r.Context(), r.PathValue("documentID"), input, actorUserID(r),
	)
	if err != nil {
		a.accessoryError(w, err, "update accessory document")
		return
	}
	respondJSON(w, http.StatusOK, document)
}

func (a *App) deleteAccessoryDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := a.accessoryDocumentForProduct(w, r); !ok {
		return
	}
	if err := a.accessoryDocumentService.DeleteDocument(
		r.Context(), r.PathValue("documentID"), actorUserID(r),
	); err != nil {
		a.accessoryError(w, err, "delete accessory document")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) downloadAccessoryDocument(w http.ResponseWriter, r *http.Request) {
	document, ok := a.accessoryDocumentForProduct(w, r)
	if !ok {
		return
	}
	data, err := a.loadFileBlob(r.Context(), document.FileBlobID)
	if err != nil {
		a.logger.Error("accessory document blob load failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "accessory_document_download_failed",
			"Accessory document could not be loaded.")
		return
	}
	disposition := "attachment"
	if strings.HasPrefix(strings.ToLower(document.MimeType), "image/") {
		disposition = "inline"
	}
	w.Header().Set("Content-Type", document.MimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition,
		map[string]string{"filename": cleanOriginalFileName(document.OriginalName)}))
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (a *App) accessoryDocumentForProduct(
	w http.ResponseWriter,
	r *http.Request,
) (*application.AccessoryDocument, bool) {
	document, err := a.accessoryDocumentService.GetDocument(r.Context(), r.PathValue("documentID"))
	if err != nil {
		a.accessoryError(w, err, "get accessory document")
		return nil, false
	}
	if document.ProductID != r.PathValue("id") {
		a.accessoryError(w, application.ErrAccessoryNotFound, "get accessory document")
		return nil, false
	}
	return document, true
}

func parseOptionalAccessoryBoolean(value string) (bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, true
	}
	parsed, err := strconv.ParseBool(value)
	return parsed, err == nil
}
