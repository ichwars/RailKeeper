package api

import (
	"context"
	"errors"
	"net/http"
	"railkeeper/backend/internal/application"
	"strconv"
	"strings"
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

func (a *App) deleteDataTransferJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) {
		return
	}
	if err := a.dataTransferService.DeleteJob(
		r.Context(), r.PathValue("id"), allowedDataTransferAreas(r)...,
	); err != nil {
		a.dataTransferError(w, err, "delete data transfer job")
		return
	}
	a.recordAudit(r, "DataTransferJobDeleted", "data_transfer_job", r.PathValue("id"))
	w.WriteHeader(http.StatusNoContent)
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
		case "Admin", "Editor", "Viewer":
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
