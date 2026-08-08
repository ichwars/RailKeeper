package api

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"railkeeper/backend/internal/application"
)

const masterDataProtectedMessage = "Standard article type keys cannot be created, changed, or deleted."

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

func (a *App) exportMasterData(w http.ResponseWriter, r *http.Request) {
	doc, err := a.masterDataService.Export(r.Context())
	if err != nil {
		a.logger.Error("master data export failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_export_failed", "Stammdaten konnten nicht exportiert werden.")
		return
	}
	filename := "railkeeper-stammdaten-" + time.Now().UTC().Format("20060102-150405") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	if err := json.NewEncoder(w).Encode(doc); err != nil {
		a.logger.Error("master data encode failed", "error", err)
	}
}

func (a *App) importMasterData(w http.ResponseWriter, r *http.Request) {
	doc, ok := a.readMasterDataImportUpload(w, r)
	if !ok {
		return
	}
	result, err := a.masterDataService.Import(r.Context(), doc)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_import_invalid",
				masterDataValidationMessageOr(err, "Stammdaten-Datei ist ungültig."))
			return
		}
		a.logger.Error("master data import failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_import_failed", "Stammdaten konnten nicht importiert werden.")
		return
	}
	if err := a.masterDataService.RefreshCache(r.Context()); err != nil {
		a.logger.Error("master data cache refresh after import failed", "error", err)
	}
	respondJSON(w, http.StatusOK, result)
}

func (a *App) readMasterDataImportUpload(w http.ResponseWriter, r *http.Request) (*application.MasterDataDocument, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxMasterDataImportBytes+1024*1024)
	if err := r.ParseMultipartForm(maxMasterDataImportBytes); err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei konnte nicht gelesen werden.")
		return nil, false
	}
	if r.MultipartForm != nil {
		defer func() { _ = r.MultipartForm.RemoveAll() }()
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_file_missing", "Eine Stammdaten-Datei ist erforderlich.")
		return nil, false
	}
	defer func() { _ = file.Close() }()
	if header.Size > maxMasterDataImportBytes {
		respondProblem(w, http.StatusBadRequest, "master_data_file_too_large", "Die Stammdaten-Datei ist zu gro?.")
		return nil, false
	}
	data, err := io.ReadAll(io.LimitReader(file, maxMasterDataImportBytes+1))
	if err != nil || int64(len(data)) > maxMasterDataImportBytes {
		respondProblem(w, http.StatusBadRequest, "master_data_file_too_large", "Die Stammdaten-Datei ist zu gro?.")
		return nil, false
	}
	var doc application.MasterDataDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		respondProblem(w, http.StatusBadRequest, "master_data_import_invalid", "Stammdaten-Datei ist ungültig.")
		return nil, false
	}
	return &doc, true
}

func cleanOriginalFileName(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	value = strings.TrimSpace(path.Base(value))
	if value == "" || value == "." || value == "/" {
		return "beilage"
	}
	return value
}

func safeAttachmentFileName(value string) string {
	value = cleanOriginalFileName(value)
	if value == "" {
		return "beilage"
	}
	value = safeFileNamePattern.ReplaceAllString(value, "-")
	value = strings.Trim(value, ".-")
	if value == "" {
		return "beilage"
	}
	return value
}

func confinedDataPath(dataDir, relativePath string) (string, error) {
	base, err := filepath.Abs(dataDir)
	if err != nil {
		return "", err
	}
	target, err := filepath.Abs(filepath.Join(base, relativePath))
	if err != nil {
		return "", err
	}
	if target != base && !strings.HasPrefix(target, base+string(os.PathSeparator)) {
		return "", errors.New("path escapes data directory")
	}
	return target, nil
}

func isBlockedAttachmentName(value string) bool {
	switch strings.ToLower(filepath.Ext(value)) {
	case ".exe", ".bat", ".cmd", ".com", ".scr", ".msi", ".dll", ".ps1", ".vbs", ".js", ".jar", ".sh":
		return true
	default:
		return false
	}
}

func isBlockedAttachmentMime(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.Contains(value, "x-msdownload") ||
		strings.Contains(value, "x-dosexec") ||
		strings.Contains(value, "x-sh") ||
		strings.Contains(value, "javascript") ||
		strings.Contains(value, "ecmascript") ||
		strings.Contains(value, "x-msdos-program")
}

func isAllowedAttachmentUpload(filename, mimeType string) bool {
	return isAllowedAttachmentUploadWithExtensions(filename, mimeType, allowedAttachmentExtensions)
}

func (a *App) isAllowedAttachmentUpload(filename, mimeType string) bool {
	return isAllowedAttachmentUploadWithExtensions(filename, mimeType, a.allowedAttachmentExtensions)
}

func isAllowedAttachmentUploadWithExtensions(filename, mimeType string, extensions map[string]struct{}) bool {
	if isBlockedAttachmentName(filename) || isBlockedAttachmentMime(mimeType) {
		return false
	}
	extension := strings.ToLower(filepath.Ext(filename))
	if _, ok := extensions[extension]; !ok {
		return false
	}
	mimeType = strings.ToLower(strings.TrimSpace(strings.Split(mimeType, ";")[0]))
	switch extension {
	case ".pdf":
		return mimeType == "application/pdf"
	case ".jpg", ".jpeg":
		return mimeType == "image/jpeg"
	case ".png":
		return mimeType == "image/png"
	case ".webp":
		return mimeType == "image/webp"
	case ".zip":
		return mimeType == "application/zip" || mimeType == "application/x-zip-compressed" || mimeType == "application/octet-stream"
	case ".txt", ".csv", ".json", ".xml":
		return strings.HasPrefix(mimeType, "text/") ||
			mimeType == "application/json" ||
			mimeType == "application/xml"
	default:
		return false
	}
}

func isAllowedImageMime(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func (a *App) searchArticleData(w http.ResponseWriter, r *http.Request) {
	var input application.ArticleSearchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	result, err := a.articleSearch.Search(r.Context(), input)
	if err != nil {
		if errors.Is(err, application.ErrArticleSearchValidation) {
			respondProblem(w, http.StatusBadRequest, "article_search_validation", "At least one search field is required.")
			return
		}
		a.logger.Error("article search failed", "error", err)
		respondProblem(w, http.StatusGatewayTimeout, "article_search_failed", "Artikeldaten-Websuche konnte nicht abgeschlossen werden.")
		return
	}

	respondJSON(w, http.StatusOK, result)
}

func (a *App) listInventoryNumberSchemes(w http.ResponseWriter, r *http.Request) {
	schemes, err := a.inventoryNumbers.List(r.Context())
	if err != nil {
		a.logger.Error("inventory number scheme list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_list_failed", "Could not list inventory number schemes.")
		return
	}

	respondJSON(w, http.StatusOK, schemes)
}

func (a *App) createInventoryNumberScheme(w http.ResponseWriter, r *http.Request) {
	var input application.InventoryNumberSchemeCreateInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	scheme, err := a.inventoryNumbers.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "inventory_number_validation", "Category, prefix, next number and padding are required.")
		case errors.Is(err, application.ErrInventoryNumberConflict):
			respondProblem(w, http.StatusConflict, "inventory_number_scheme_exists", "Inventory number scheme already exists.")
		default:
			a.logger.Error("inventory number scheme create failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_create_failed", "Could not create inventory number scheme.")
		}
		return
	}

	respondJSON(w, http.StatusCreated, scheme)
}

func (a *App) updateInventoryNumberScheme(w http.ResponseWriter, r *http.Request) {
	var input application.InventoryNumberSchemeInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	scheme, err := a.inventoryNumbers.Update(r.Context(), r.PathValue("category"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrInventoryNumberValidation):
			respondProblem(w, http.StatusBadRequest, "inventory_number_validation", "Prefix, next number and padding are required.")
		case errors.Is(err, application.ErrInventoryNumberNotFound):
			respondProblem(w, http.StatusNotFound, "inventory_number_scheme_not_found", "Inventory number scheme not found.")
		default:
			a.logger.Error("inventory number scheme update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "inventory_number_scheme_update_failed", "Could not update inventory number scheme.")
		}
		return
	}

	respondJSON(w, http.StatusOK, scheme)
}

func (a *App) listMasterData(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	items, err := a.masterDataService.List(r.Context(), r.PathValue("type"), activeOnly)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", "Master data type is required.")
			return
		}
		a.logger.Error("master data list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) listAllMasterData(w http.ResponseWriter, r *http.Request) {
	activeOnly := r.URL.Query().Get("active") == "true"
	items, err := a.masterDataService.ListAll(r.Context(), activeOnly)
	if err != nil {
		a.logger.Error("master data list all failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_list_failed", "Could not list master data.")
		return
	}

	respondJSON(w, http.StatusOK, items)
}

func (a *App) createMasterData(w http.ResponseWriter, r *http.Request) {
	var input application.MasterDataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	item, err := a.masterDataService.Create(r.Context(), r.PathValue("type"), input)
	if err != nil {
		if errors.Is(err, application.ErrMasterDataProtected) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataProtectedMessage)
			return
		}
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataValidationMessage(err))
			return
		}
		a.logger.Error("master data create failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_create_failed", "Could not create master data.")
		return
	}

	respondJSON(w, http.StatusCreated, item)
}

func (a *App) updateMasterData(w http.ResponseWriter, r *http.Request) {
	var input application.MasterDataInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondProblem(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON.")
		return
	}

	item, err := a.masterDataService.Update(r.Context(), r.PathValue("type"), r.PathValue("key"), input)
	if err != nil {
		switch {
		case errors.Is(err, application.ErrMasterDataProtected):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataProtectedMessage)
		case errors.Is(err, application.ErrMasterDataValidation):
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataValidationMessage(err))
		case errors.Is(err, application.ErrMasterDataNotFound):
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
		default:
			a.logger.Error("master data update failed", "error", err)
			respondProblem(w, http.StatusInternalServerError, "master_data_update_failed", "Could not update master data.")
		}
		return
	}

	respondJSON(w, http.StatusOK, item)
}

func masterDataValidationMessage(err error) string {
	return masterDataValidationMessageOr(err, "Label is required.")
}

func masterDataValidationMessageOr(err error, fallback string) string {
	if err != application.ErrMasterDataValidation {
		return err.Error()
	}
	return fallback
}

func (a *App) deleteMasterData(w http.ResponseWriter, r *http.Request) {
	if err := a.masterDataService.Delete(r.Context(), r.PathValue("type"), r.PathValue("key")); err != nil {
		if errors.Is(err, application.ErrMasterDataProtected) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation", masterDataProtectedMessage)
			return
		}
		if errors.Is(err, application.ErrMasterDataValidation) {
			respondProblem(w, http.StatusBadRequest, "master_data_validation",
				masterDataValidationMessageOr(err, "This master data entry cannot be deleted."))
			return
		}
		if errors.Is(err, application.ErrMasterDataNotFound) {
			respondProblem(w, http.StatusNotFound, "master_data_not_found", "Master data entry not found.")
			return
		}
		a.logger.Error("master data delete failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_delete_failed", "Could not delete master data.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listMasterDataRelations(w http.ResponseWriter, r *http.Request) {
	relations, err := a.masterDataService.Relations(
		r.Context(),
		r.URL.Query().Get("parentType"),
		r.URL.Query().Get("childType"),
	)
	if err != nil {
		a.logger.Error("master data relations list failed", "error", err)
		respondProblem(w, http.StatusInternalServerError, "master_data_relations_failed", "Could not list master data relations.")
		return
	}

	respondJSON(w, http.StatusOK, relations)
}
