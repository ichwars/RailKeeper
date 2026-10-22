package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"railkeeper/backend/internal/application"
	ecospkg "railkeeper/backend/internal/ecos"
)

func (a *App) ecosImageOptions(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(35 * time.Second))
	w.Header().Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	result, err := a.digitalCenterWorkspace.ECoSImageOptions(ctx, r.PathValue("id"))
	if err != nil {
		a.respondECoSImageError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) previewECoSImage(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(75 * time.Second))
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseMultipartForm(64 << 10); err != nil || r.MultipartForm == nil {
		respondProblem(w, http.StatusBadRequest, "ecos_image_input", "A 190 × 40 RGB24 BMP and target locomotive are required.")
		return
	}
	defer r.MultipartForm.RemoveAll()
	if len(r.MultipartForm.Value) != 2 || len(r.MultipartForm.Value["objectId"]) != 1 ||
		len(r.MultipartForm.Value["description"]) != 1 || len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["image"]) != 1 {
		respondProblem(w, http.StatusBadRequest, "ecos_image_input", "Unexpected or missing image transfer fields.")
		return
	}
	objectID, err := strconv.Atoi(r.MultipartForm.Value["objectId"][0])
	if err != nil {
		a.respondECoSImageError(w, application.ErrECoSImageInput)
		return
	}
	file, _, err := r.FormFile("image")
	if err != nil {
		a.respondECoSImageError(w, application.ErrECoSImageInput)
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, ecospkg.ImageBMPSize+1))
	if err != nil || len(data) > ecospkg.ImageBMPSize {
		a.respondECoSImageError(w, application.ErrECoSImageInput)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	result, err := a.digitalCenterWorkspace.PreviewECoSImage(ctx, r.PathValue("id"), application.ECoSImagePreviewInput{
		ObjectID: objectID, Description: r.MultipartForm.Value["description"][0], BMP: data,
	}, actorUserID(r))
	if err != nil {
		a.respondECoSImageError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respondJSON(w, http.StatusOK, result)
}

func (a *App) confirmECoSImage(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	// Keep this bounded device operation within its own response deadline rather
	// than losing a successful result at the server's general 30-second limit.
	_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(100 * time.Second))
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var input application.ECoSImageConfirmInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		a.respondECoSImageError(w, application.ErrECoSImageInput)
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		a.respondECoSImageError(w, application.ErrECoSImageInput)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	result, err := a.digitalCenterWorkspace.ConfirmECoSImage(ctx, r.PathValue("id"), input, actorUserID(r))
	if err != nil {
		a.respondECoSImageError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	respondJSON(w, http.StatusOK, result)
}

func (a *App) respondECoSImageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, application.ErrECoSImageInput), errors.Is(err, ecospkg.ErrImageFormat):
		respondProblem(w, http.StatusBadRequest, "ecos_image_input", "Use a valid 190 × 40 RGB24 BMP, description and target locomotive.")
	case errors.Is(err, ecospkg.ErrImageSlotOccupied):
		respondProblem(w, http.StatusConflict, "ecos_image_slot_occupied", "The image slot is now occupied. Create a fresh preview.")
	case errors.Is(err, ecospkg.ErrImageSlotsFull):
		respondProblem(w, http.StatusConflict, "ecos_image_slots_full", "No free user image slot was found.")
	case errors.Is(err, ecospkg.ErrImageWeb):
		respondProblem(w, http.StatusBadGateway, "ecos_image_web", "The ECoS web interface on port 80 is unavailable or incompatible.")
	case errors.Is(err, application.ErrDigitalCenterPreviewStale), errors.Is(err, application.ErrDigitalCenterGrantMismatch),
		errors.Is(err, application.ErrDigitalCenterConfirmationRequired), errors.Is(err, application.ErrDigitalCenterInactive),
		errors.Is(err, application.ErrDigitalCenterNotConfigured), errors.Is(err, application.ErrDigitalCenterWorkspaceUnavailable), errors.Is(err, sql.ErrNoRows):
		a.respondDigitalCenterWorkspaceError(w, err)
	default:
		a.logger.Error("ECoS image operation failed", "error", err)
		respondProblem(w, http.StatusBadGateway, "ecos_image_failed", "The ECoS image operation failed. Check the device and create a fresh preview.")
	}
}
