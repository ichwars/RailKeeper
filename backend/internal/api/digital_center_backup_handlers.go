package api

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
	ecospkg "railkeeper/backend/internal/ecos"
)

func (a *App) importECoSBackup(w http.ResponseWriter, r *http.Request) {
	if !a.requireDigitalCenterWorkspace(w) {
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/octet-stream" {
		respondProblem(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "An ECoS backup file is required.")
		return
	}
	confirmation := r.Header.Get("X-RailKeeper-Confirm-Station")
	if confirmation != "" && confirmation != "true" && confirmation != "false" {
		respondProblem(w, http.StatusBadRequest, "invalid_request", "Invalid station confirmation.")
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, ecospkg.EncryptedBackupLimit)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var oversized *http.MaxBytesError
		if errors.As(err, &oversized) {
			respondProblem(w, http.StatusRequestEntityTooLarge, "ecos_backup_too_large", "The ECoS backup is too large.")
		} else {
			respondProblem(w, http.StatusUnprocessableEntity, "ecos_backup_invalid", "The ECoS backup could not be read.")
		}
		return
	}
	keyring, passphrase, err := a.ecosBackupPrivateConfiguration()
	if err != nil {
		respondProblem(w, http.StatusServiceUnavailable, "ecos_backup_unavailable", "ECoS backup decryption is not configured.")
		return
	}
	defer clear(keyring)
	defer clear(passphrase)
	ctx, cancel := context.WithTimeout(r.Context(), 45*time.Second)
	defer cancel()
	result, err := a.digitalCenterWorkspace.ImportECoSBackup(ctx, bytes.NewReader(data), keyring, passphrase,
		application.ECoSBackupImportInput{ConfirmStation: confirmation == "true",
			StationHost: r.Header.Get("X-RailKeeper-Backup-Station")}, actorUserID(r))
	if err != nil {
		switch {
		case errors.Is(err, application.ErrECoSBackupInvalid):
			respondProblem(w, http.StatusUnprocessableEntity, "ecos_backup_invalid", "The ECoS backup is invalid or damaged.")
		case errors.Is(err, application.ErrECoSBackupUnavailable):
			respondProblem(w, http.StatusServiceUnavailable, "ecos_backup_unavailable", "ECoS backup decryption is unavailable.")
		case errors.Is(err, application.ErrECoSBackupStationMismatch):
			respondProblem(w, http.StatusConflict, "digital_center_backup_station_mismatch", "The configured ECoS has changed. Review the station confirmation.")
		default:
			a.respondDigitalCenterWorkspaceError(w, err)
		}
		return
	}
	a.recordAudit(r, "ECoSBackupPreviewImported", "digital_center_read_session", result.Session.ID)
	respondJSON(w, http.StatusCreated, result)
}

func (a *App) ecosBackupPrivateConfiguration() ([]byte, []byte, error) {
	root, err := os.OpenRoot(a.dataDir)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = root.Close() }()
	read := func(name string, limit int64) ([]byte, error) {
		file, err := root.Open("private/ecos-backup/" + name)
		if err != nil {
			return nil, err
		}
		defer func() { _ = file.Close() }()
		info, err := file.Stat()
		if err != nil || !info.Mode().IsRegular() || info.Size() < 1 || info.Size() > limit {
			return nil, application.ErrECoSBackupUnavailable
		}
		data, err := io.ReadAll(io.LimitReader(file, limit+1))
		if err != nil || int64(len(data)) > limit {
			clear(data)
			return nil, application.ErrECoSBackupUnavailable
		}
		return data, nil
	}
	keyring, err := read("keyring.gpg", 64<<10)
	if err != nil {
		return nil, nil, err
	}
	passphrase, err := read("passphrase", 4<<10)
	if err != nil {
		clear(keyring)
		return nil, nil, err
	}
	// Provisioning may append a newline; interior spaces are part of the secret.
	passphrase = bytes.TrimRight(passphrase, "\r\n")
	if strings.TrimSpace(string(passphrase)) == "" {
		clear(keyring)
		clear(passphrase)
		return nil, nil, application.ErrECoSBackupUnavailable
	}
	return keyring, passphrase, nil
}
