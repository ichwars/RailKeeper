package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
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
	Confirm          *bool `json:"confirm"`
	ExpectedRevision *int  `json:"expectedRevision"`
}

func (a *App) dataTransferSummary(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	summary, err := a.dataTransferService.Summary(r.Context(), allowedDataTransferAreas(r)...)
	if err != nil {
		a.dataTransferError(w, err, "summarize data transfers")
		return
	}
	respondJSON(w, http.StatusOK, summary)
}

func (a *App) listDataTransferJobs(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	filter, err := dataTransferJobFilter(r)
	if err != nil {
		a.dataTransferError(w, err, "validate data transfer job filters")
		return
	}
	jobs, err := a.dataTransferService.ListJobs(r.Context(), filter, allowedDataTransferAreas(r)...)
	if err != nil {
		a.dataTransferError(w, err, "list data transfer jobs")
		return
	}
	respondJSON(w, http.StatusOK, jobs)
}

func (a *App) getDataTransferJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	details, err := a.dataTransferService.GetJobDetails(
		r.Context(), r.PathValue("id"), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "get data transfer job")
		return
	}
	respondJSON(w, http.StatusOK, details)
}

func (a *App) retryDataTransferJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	job, err := a.dataTransferService.RetryJob(
		r.Context(), r.PathValue("id"), actorUserID(r), allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "retry data transfer job")
		return
	}
	a.recordAudit(r, "DataTransferJobRetried", "data_transfer_job", job.ID)
	respondJSON(w, http.StatusCreated, job)
}

func dataTransferJobFilter(r *http.Request) (application.DataTransferJobFilter, error) {
	query := r.URL.Query()
	filter := application.DataTransferJobFilter{
		ProfileID: strings.TrimSpace(query.Get("profileId")),
		Direction: application.TransferDirection(strings.TrimSpace(query.Get("direction"))),
		Limit:     100,
	}
	if value := strings.TrimSpace(query.Get("limit")); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 || limit > 200 {
			return application.DataTransferJobFilter{}, application.ErrDataTransferValidation
		}
		filter.Limit = limit
	}
	stateValues := append([]string{}, query["states"]...)
	stateValues = append(stateValues, query["state"]...)
	for _, value := range stateValues {
		for _, state := range strings.Split(value, ",") {
			state = strings.TrimSpace(state)
			if state == "" {
				return application.DataTransferJobFilter{}, application.ErrDataTransferValidation
			}
			filter.States = append(filter.States, application.TransferJobState(state))
		}
	}
	return filter, nil
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
	a.recordAudit(r, "DataTransferImportJobCreated", "data_transfer_job", job.ID)
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
	var mapping *application.DataTransferCSVMappingInput
	var fileName, fileMIMEType string
	var filePayload []byte
	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			respondProblem(w, http.StatusBadRequest, "data_transfer_upload_invalid",
				"Import upload must be valid multipart data.")
			return
		}
		switch {
		case part.FormName() == "mapping" && strings.TrimSpace(part.FileName()) == "" && mapping == nil:
			payload, readErr := io.ReadAll(io.LimitReader(part, 256*1024+1))
			_ = part.Close()
			if readErr != nil || len(payload) > 256*1024 {
				respondProblem(w, http.StatusBadRequest, "data_transfer_mapping_invalid", "CSV mapping is invalid.")
				return
			}
			mapping = &application.DataTransferCSVMappingInput{}
			decoder := json.NewDecoder(bytes.NewReader(payload))
			decoder.DisallowUnknownFields()
			if err := decoder.Decode(mapping); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
				respondProblem(w, http.StatusBadRequest, "data_transfer_mapping_invalid", "CSV mapping is invalid.")
				return
			}
		case part.FormName() == "file" && strings.TrimSpace(part.FileName()) != "" && fileName == "":
			fileName = part.FileName()
			fileMIMEType = part.Header.Get("Content-Type")
			filePayload, err = io.ReadAll(io.LimitReader(part, application.DataTransferMaxUploadBytes+1))
			_ = part.Close()
			if err != nil {
				a.dataTransferError(w, err, "read data transfer upload")
				return
			}
			if int64(len(filePayload)) > application.DataTransferMaxUploadBytes {
				a.dataTransferError(w, application.ErrDataTransferUploadTooLarge, "preview data transfer import")
				return
			}
		default:
			_ = part.Close()
			respondProblem(w, http.StatusBadRequest, "data_transfer_upload_invalid",
				"Import upload contains an unexpected or repeated multipart field.")
			return
		}
	}
	if fileName == "" {
		respondProblem(w, http.StatusBadRequest, "data_transfer_upload_missing", "Import file is required.")
		return
	}
	preview, err := a.dataTransferService.UploadAndPreviewReader(
		r.Context(), r.PathValue("id"), fileName, fileMIMEType, bytes.NewReader(filePayload),
		actorUserID(r), mapping, allowedDataTransferAreas(r)...,
	)
	if err != nil {
		a.dataTransferError(w, err, "preview data transfer import")
		return
	}
	a.recordAudit(r, "DataTransferImportUploaded", "data_transfer_job", preview.Job.ID)
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
	a.recordAudit(r, "DataTransferIssueResolved", "data_transfer_issue", r.PathValue("issueID"))
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
	a.recordAudit(r, "DataTransferJobCancelled", "data_transfer_job", job.ID)
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
	if input.Confirm == nil || input.ExpectedRevision == nil || *input.ExpectedRevision < 1 {
		respondProblem(w, http.StatusBadRequest, "data_transfer_validation",
			"Import confirmation and the reviewed revision are required.")
		return
	}
	job, err := a.dataTransferService.ConfirmImportWithPolicy(
		r.Context(), r.PathValue("id"), *input.ExpectedRevision, *input.Confirm, actorUserID(r),
		application.DataTransferImportPolicy{CanManageExhibitionLists: !dataTransferMesseOnly(r)},
		allowedDataTransferAreas(r)...,
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
	a.recordAudit(r, "DataTransferExportJobCreated", "data_transfer_job", job.ID)
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
	a.recordAudit(r, "DataTransferExportExecuted", "data_transfer_job", result.Job.ID)
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
	a.recordAudit(r, "DataTransferArtifactDeleted", "data_transfer_artifact", r.PathValue("id"))
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
	a.recordAudit(r, "DataTransferProfileCreated", "data_transfer_profile", profile.ID)
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
	a.recordAudit(r, "DataTransferProfileUpdated", "data_transfer_profile", profile.ID)
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
	a.recordAudit(r, "DataTransferProfileDisabled", "data_transfer_profile", r.PathValue("id"))
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
