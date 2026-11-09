package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"railkeeper/backend/internal/application"
	"time"
)

func (a *App) exportBackup(w http.ResponseWriter, r *http.Request) {
	backup, err := a.backupService.Export(r.Context())
	if err != nil {
		a.logger.Error("backup export failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "backup_export_failed", "Backup konnte nicht erstellt werden.")
		return
	}

	filename := "railkeeper-backup-" + time.Now().UTC().Format("20060102-150405") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	if err := json.NewEncoder(w).Encode(backup); err != nil {
		a.logger.Error("backup encode failed", "error", err)
	}
}

func (a *App) restoreBackup(w http.ResponseWriter, r *http.Request) {
	backup, ok := a.readBackupUpload(w, r)
	if !ok {
		return
	}

	result, err := a.backupService.Import(r.Context(), backup)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrBackupInvalid), errors.Is(err, application.ErrBackupPath):
			respondProblem(w, http.StatusBadRequest, "backup_restore_invalid", "Backup-Datei ist ungültig.")
		default:
			a.logger.Error("backup restore failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "backup_restore_failed", "Backup konnte nicht wiederhergestellt werden.")
		}
		return
	}
	if a.masterDataService != nil {
		if err := a.masterDataService.RefreshCache(r.Context()); err != nil {
			a.logger.Error("master data cache refresh after backup restore failed", "error", err)
		}
	}
	if a.fileBlobs != nil {
		if err := a.fileBlobs.MigrateFilesystemBlobs(r.Context()); err != nil {
			a.logger.Error("backup restore blob migration failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "backup_restore_failed", "Backup konnte nicht wiederhergestellt werden.")
			return
		}
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) validateBackup(w http.ResponseWriter, r *http.Request) {
	backup, ok := a.readBackupUpload(w, r)
	if !ok {
		return
	}
	result, err := a.backupService.Validate(r.Context(), backup)
	if err != nil {
		a.logger.Error("backup validation failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "backup_validation_failed", "Backup konnte nicht geprüft werden.")
		return
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) readBackupUpload(w http.ResponseWriter, r *http.Request) (*application.BackupDocument, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBackupBytes+1024*1024)
	if err := r.ParseMultipartForm(maxBackupBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "backup_restore_invalid", "Backup-Datei konnte nicht gelesen werden.")
		return nil, false
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "backup_file_missing", "Eine Backup-Datei ist erforderlich.")
		return nil, false
	}
	defer func() { _ = file.Close() }()
	if header.Size > maxBackupBytes {
		respondProblem(w, http.StatusBadRequest, "backup_file_too_large", "Die Backup-Datei ist zu gro?.")
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, maxBackupBytes+1))
	if err != nil || int64(len(data)) > maxBackupBytes {
		respondProblem(w, http.StatusBadRequest, "backup_file_too_large", "Die Backup-Datei ist zu gro?.")
		return nil, false
	}

	backup, err := application.DecodeBackup(data)
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "backup_restore_invalid", "Backup-Datei ist ungültig.")
		return nil, false
	}
	return backup, true
}
