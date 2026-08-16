package api

import (
	"context"
	"net/http"
	"sync"

	"railkeeper/backend/internal/startup"
)

type StorageLocationInfo struct {
	DataPath            string                    `json:"dataPath"`
	Mode                startup.StorageMode       `json:"mode"`
	OpenFolderAvailable bool                      `json:"openFolderAvailable"`
	MigrationReceipt    *startup.MigrationReceipt `json:"migrationReceipt,omitempty"`
}

type StorageLocationConfig struct {
	DataPath             string
	Mode                 startup.StorageMode
	OpenFolderAvailable  bool
	MigrationReceipt     *startup.MigrationReceipt
	OpenFolder           func(context.Context, string) error
	AcknowledgeMigration func(context.Context, string) (*startup.MigrationReceipt, error)
}

type storageLocationState struct {
	mu                   sync.RWMutex
	info                 StorageLocationInfo
	openFolder           func(context.Context, string) error
	acknowledgeMigration func(context.Context, string) (*startup.MigrationReceipt, error)
}

func newStorageLocationState(config StorageLocationConfig, fallbackDataPath string) *storageLocationState {
	dataPath := config.DataPath
	if dataPath == "" {
		dataPath = fallbackDataPath
	}
	mode := config.Mode
	if mode == "" {
		mode = startup.StorageModeServer
	}
	return &storageLocationState{
		info: StorageLocationInfo{
			DataPath: dataPath, Mode: mode,
			OpenFolderAvailable: config.OpenFolderAvailable,
			MigrationReceipt:    cloneMigrationReceipt(config.MigrationReceipt),
		},
		openFolder:           config.OpenFolder,
		acknowledgeMigration: config.AcknowledgeMigration,
	}
}

func (a *App) storageLocationInfo(response http.ResponseWriter, _ *http.Request) {
	a.storageLocation.mu.RLock()
	info := a.storageLocation.info
	info.MigrationReceipt = cloneMigrationReceipt(info.MigrationReceipt)
	a.storageLocation.mu.RUnlock()
	respondJSON(response, http.StatusOK, info)
}

func (a *App) openStorageFolder(response http.ResponseWriter, request *http.Request) {
	if storageRequestHasInput(request) {
		respondProblem(
			response,
			http.StatusBadRequest,
			"storage_folder_request_invalid",
			"The storage folder action does not accept a path or request body.",
		)
		return
	}
	a.storageLocation.mu.RLock()
	dataPath := a.storageLocation.info.DataPath
	available := a.storageLocation.info.OpenFolderAvailable
	opener := a.storageLocation.openFolder
	a.storageLocation.mu.RUnlock()
	if !available || opener == nil {
		respondProblem(
			response,
			http.StatusConflict,
			"storage_folder_open_unavailable",
			"Opening the storage folder is unavailable in this runtime.",
		)
		return
	}
	if err := opener(request.Context(), dataPath); err != nil {
		a.logger.Error("storage folder open failed", "error", err)
		respondProblem(
			response,
			http.StatusInternalServerError,
			"storage_folder_open_failed",
			"The storage folder could not be opened.",
		)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (a *App) acknowledgeStorageMigration(response http.ResponseWriter, request *http.Request) {
	if storageRequestHasInput(request) {
		respondProblem(
			response,
			http.StatusBadRequest,
			"storage_migration_request_invalid",
			"The migration acknowledgement does not accept a path or request body.",
		)
		return
	}
	a.storageLocation.mu.Lock()
	defer a.storageLocation.mu.Unlock()
	if a.storageLocation.info.MigrationReceipt == nil || a.storageLocation.acknowledgeMigration == nil {
		respondProblem(
			response,
			http.StatusConflict,
			"storage_migration_receipt_unavailable",
			"No migration receipt is available for acknowledgement.",
		)
		return
	}
	updated, err := a.storageLocation.acknowledgeMigration(
		request.Context(), a.storageLocation.info.DataPath,
	)
	if err != nil {
		a.logger.Error("storage migration acknowledgement failed", "error", err)
		respondProblem(
			response,
			http.StatusInternalServerError,
			"storage_migration_acknowledgement_failed",
			"The migration acknowledgement could not be stored.",
		)
		return
	}
	a.storageLocation.info.MigrationReceipt = cloneMigrationReceipt(updated)
	response.WriteHeader(http.StatusNoContent)
}

func storageRequestHasInput(request *http.Request) bool {
	return len(request.URL.Query()) != 0 || request.ContentLength > 0 || len(request.TransferEncoding) != 0
}

func cloneMigrationReceipt(receipt *startup.MigrationReceipt) *startup.MigrationReceipt {
	if receipt == nil {
		return nil
	}
	clone := *receipt
	return &clone
}
