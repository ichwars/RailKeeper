package api

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"

	"railkeeper/backend/internal/application"
)

type dataTransferScopeKey struct{}

type createDataTransferExportJobRequest struct {
	ProfileID string `json:"profileId"`
}

type createDataTransferImportJobRequest struct {
	ProfileID string `json:"profileId"`
}

type resolveDataTransferIssueRequest struct {
	Resolution string `json:"resolution"`
}

type confirmDataTransferImportRequest struct {
	Confirm *bool `json:"confirm"`
}

func (a *App) createDataTransferImportJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input createDataTransferImportJobRequest
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	job, err := a.dataTransferService.CreateImportJob(
		r.Context(), input.ProfileID, actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "create data transfer import job")
		return
	}
	respondJSON(w, http.StatusCreated, job)
}

func (a *App) uploadDataTransferImport(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, application.DataTransferMaxUploadBytes+1024*1024)
	reader, err := r.MultipartReader()
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "data_transfer_upload_invalid",
			"Import upload must be valid multipart data.")
		return
	}
	part, err := reader.NextPart()
	if err != nil || part.FormName() != "file" || strings.TrimSpace(part.FileName()) == "" {
		respondProblem(w, http.StatusBadRequest, "data_transfer_upload_missing", "Import file is required.")
		return
	}
	defer func() { _ = part.Close() }()
	preview, err := a.dataTransferService.UploadAndPreviewReader(
		r.Context(), r.PathValue("id"), part.FileName(), part.Header.Get("Content-Type"), part,
		actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "preview data transfer import")
		return
	}
	respondJSON(w, http.StatusOK, preview)
}

func (a *App) resolveDataTransferIssue(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input resolveDataTransferIssueRequest
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	job, err := a.dataTransferService.ResolveIssue(
		r.Context(), r.PathValue("id"), r.PathValue("issueID"), input.Resolution,
		actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "resolve data transfer issue")
		return
	}
	respondJSON(w, http.StatusOK, job)
}

func (a *App) cancelDataTransferJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	job, err := a.dataTransferService.CancelJob(
		r.Context(), r.PathValue("id"), actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "cancel data transfer job")
		return
	}
	respondJSON(w, http.StatusOK, job)
}

func (a *App) confirmDataTransferImport(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input confirmDataTransferImportRequest
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	if input.Confirm == nil {
		respondProblem(w, http.StatusBadRequest, "data_transfer_validation", "Import confirmation is required.")
		return
	}
	job, err := a.dataTransferService.ConfirmImport(
		r.Context(), r.PathValue("id"), *input.Confirm, actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "confirm data transfer import")
		return
	}
	respondJSON(w, http.StatusOK, job)
}

func (a *App) createDataTransferExportJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input createDataTransferExportJobRequest
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	job, err := a.dataTransferService.CreateExportJob(
		r.Context(), input.ProfileID, actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "create data transfer export job")
		return
	}
	respondJSON(w, http.StatusCreated, job)
}

func (a *App) executeDataTransferExport(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	result, err := a.dataTransferService.ExecuteExport(
		r.Context(), r.PathValue("id"), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "execute data transfer export")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) downloadDataTransferArtifact(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	artifact, file, err := a.dataTransferService.OpenArtifact(
		r.Context(), r.PathValue("id"), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "download data transfer artifact")
		return
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		a.dataTransferError(w, err, "inspect data transfer artifact")
		return
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": artifact.DisplayName})
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", artifact.MIMEType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, artifact.DisplayName, info.ModTime(), file)
}

func (a *App) deleteDataTransferArtifact(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	if err := a.dataTransferService.DeleteArtifact(r.Context(), r.PathValue("id")); err != nil {
		a.dataTransferError(w, err, "delete data transfer artifact")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) openDataTransferArtifactFolder(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	if storageRequestHasInput(r) {
		respondProblem(w, http.StatusBadRequest, "data_transfer_folder_request_invalid",
			"The export-folder action does not accept a path or request body.")
		return
	}
	if err := a.dataTransferService.OpenArtifactFolder(r.Context()); err != nil {
		a.dataTransferError(w, err, "open data transfer artifact folder")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listDataTransferProfiles(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	profiles, err := a.dataTransferService.ListProfiles(r.Context())
	if err != nil {
		a.dataTransferError(w, err, "list data transfer profiles")
		return
	}
	if dataTransferMesseOnly(r) {
		profiles = filterMesseDataTransferProfiles(profiles)
	}
	respondJSON(w, http.StatusOK, profiles)
}

func (a *App) createDataTransferProfile(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input application.CreateDataTransferProfileInput
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	profile, err := a.dataTransferService.CreateProfile(r.Context(), input, actorUserID(r))
	if err != nil {
		a.dataTransferError(w, err, "create data transfer profile")
		return
	}
	respondJSON(w, http.StatusCreated, profile)
}

func (a *App) updateDataTransferProfile(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	var input application.UpdateDataTransferProfileInput
	if !decodeBoundedJSON(w, r, &input) {
		return
	}
	profile, err := a.dataTransferService.UpdateProfile(r.Context(), r.PathValue("id"), input)
	if err != nil {
		a.dataTransferError(w, err, "update data transfer profile")
		return
	}
	respondJSON(w, http.StatusOK, profile)
}

func (a *App) disableDataTransferProfile(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	if _, err := a.dataTransferService.DisableProfile(r.Context(), r.PathValue("id")); err != nil {
		a.dataTransferError(w, err, "disable data transfer profile")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) dataTransferAvailable(w http.ResponseWriter) bool {
	if a.dataTransferService != nil {
		return true
	}
	respondProblem(w, http.StatusServiceUnavailable, "data_transfer_unavailable", "Data transfer is not available.")
	return false
}

func (a *App) dataTransferError(w http.ResponseWriter, err error, action string) {
	switch {
	case errors.Is(err, application.ErrDataTransferUploadTooLarge):
		respondProblem(w, http.StatusRequestEntityTooLarge, "data_transfer_upload_too_large",
			"Import file exceeds the upload limit.")
	case errors.Is(err, application.ErrDataTransferValidation):
		respondProblem(w, http.StatusBadRequest, "data_transfer_validation", err.Error())
	case errors.Is(err, application.ErrDataTransferForbidden):
		respondProblem(w, http.StatusForbidden, "data_transfer_forbidden", "The selected data area is not available to this role.")
	case errors.Is(err, application.ErrDataTransferArtifactDeleted):
		respondProblem(w, http.StatusGone, "data_transfer_artifact_deleted", "The export artifact was deleted.")
	case errors.Is(err, application.ErrDataTransferConflict):
		respondProblem(w, http.StatusConflict, "data_transfer_conflict", "The data transfer state changed. Refresh and retry.")
	case errors.Is(err, application.ErrDataTransferOpenUnavailable):
		respondProblem(w, http.StatusConflict, "data_transfer_folder_open_unavailable",
			"Opening the export folder is unavailable in this runtime.")
	case errors.Is(err, application.ErrDataTransferArtifactPath):
		respondProblem(w, http.StatusBadRequest, "data_transfer_artifact_path_invalid", "The export artifact path is invalid.")
	case errors.Is(err, application.ErrDataTransferNotFound):
		respondProblem(w, http.StatusNotFound, "data_transfer_not_found", "Data transfer resource not found.")
	default:
		a.logger.Error("data transfer operation failed", "action", action, "error", err)
		respondProblem(w, http.StatusInternalServerError, "data_transfer_operation_failed", "Data transfer operation failed.")
	}
}

func allowedDataTransferAreas(r *http.Request) []application.TransferArea {
	if dataTransferMesseOnly(r) {
		return []application.TransferArea{application.TransferExhibitionLists}
	}
	return nil
}

func withDataTransferScope(r *http.Request, messeOnly bool) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), dataTransferScopeKey{}, messeOnly))
}

func dataTransferMesseOnly(r *http.Request) bool {
	messeOnly, _ := r.Context().Value(dataTransferScopeKey{}).(bool)
	return messeOnly
}

func dataTransferMesseScope(roles []string) bool {
	hasMesse := false
	for _, role := range roles {
		switch role {
		case "Admin", "Editor", "Planner", "Viewer":
			return false
		case "Messe":
			hasMesse = true
		}
	}
	return hasMesse
}

func filterMesseDataTransferProfiles(profiles []application.DataTransferProfile) []application.DataTransferProfile {
	filtered := make([]application.DataTransferProfile, 0, len(profiles))
	for _, profile := range profiles {
		if len(profile.Areas) == 1 && profile.Areas[0] == application.TransferExhibitionLists {
			filtered = append(filtered, profile)
		}
	}
	return filtered
}
