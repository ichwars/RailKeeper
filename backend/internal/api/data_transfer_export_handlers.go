package api

import (
	"mime"
	"net/http"
)

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
	info, err := file.Metadata()
	if err != nil {
		a.dataTransferError(w, err, "inspect data transfer artifact")
		return
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": artifact.DisplayName})
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", artifact.MIMEType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(w, r, artifact.DisplayName, info.ModifiedAt, file)
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
