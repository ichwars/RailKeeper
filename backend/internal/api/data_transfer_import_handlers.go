package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"railkeeper/backend/internal/application"
	"strings"
)

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
