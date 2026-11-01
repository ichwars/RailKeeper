package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"

	ecospkg "railkeeper/backend/internal/ecos"
)

type ECoSCVSourceConflict struct {
	Number        int    `json:"number"`
	IndexCV31     *int   `json:"indexCV31,omitempty"`
	IndexCV32     *int   `json:"indexCV32,omitempty"`
	DatabaseValue int    `json:"databaseValue"`
	BackupValue   int    `json:"backupValue"`
	BackupSource  string `json:"backupSource,omitempty"`
	RawCacheValue *int   `json:"rawCacheValue,omitempty"`
}

type ECoSCVSourceInfo struct {
	Sources           []string               `json:"sources"`
	BackupStatus      string                 `json:"backupStatus"`
	BackupAdded       int                    `json:"backupAdded"`
	Conflicts         []ECoSCVSourceConflict `json:"conflicts"`
	Warning           string                 `json:"warning,omitempty"`
	StoredCacheStatus string                 `json:"storedCacheStatus,omitempty"`
	UnverifiedValues  []ECoSCVValue          `json:"unverifiedValues,omitempty"`
}

type storedECoSWeb interface {
	Address() string
	Read(context.Context, int, string) ([]ecospkg.DatabaseCV, error)
	ReadBackup(context.Context) ([]byte, error)
}

type storedECoSBackup struct {
	status       string
	values       map[int][]ECoSCVValue
	storedStatus string
	unverified   map[int][]ECoSCVValue
}

func (s *ECoSService) storedWeb(ctx context.Context, host string) (storedECoSWeb, error) {
	if s.storedWebFactory != nil {
		return s.storedWebFactory(ctx, host)
	}
	return ecospkg.NewCVDatabaseReader(ctx, host)
}

func (s *ECoSService) SetBackupDataDir(dataDir string) {
	s.backupMu.Lock()
	defer s.backupMu.Unlock()
	s.backupDataDir = dataDir
}

func (service *DigitalCenterWorkspaceService) SetECoSBackupDataDir(dataDir string) {
	if reader, ok := service.ecos.(interface{ SetBackupDataDir(string) }); ok {
		reader.SetBackupDataDir(dataDir)
	}
}

func (s *ECoSService) readStoredBackup(ctx context.Context, web storedECoSWeb) storedECoSBackup {
	s.backupMu.RLock()
	dataDir := s.backupDataDir
	s.backupMu.RUnlock()
	key, passphrase, err := readStoredBackupConfiguration(dataDir)
	defer clear(key)
	defer clear(passphrase)
	if errors.Is(err, os.ErrNotExist) || dataDir == "" {
		return storedECoSBackup{status: "not_configured"}
	}
	if err != nil {
		return storedECoSBackup{status: "failed"}
	}
	data, err := web.ReadBackup(ctx)
	if err != nil {
		return storedECoSBackup{status: "failed"}
	}
	backup, err := ecospkg.DecodeBackup(bytes.NewReader(data), key, passphrase)
	if err != nil {
		return storedECoSBackup{status: "failed"}
	}
	result := storedECoSBackup{status: "checked", values: make(map[int][]ECoSCVValue),
		storedStatus: backup.StoredLocomotivesStatus, unverified: make(map[int][]ECoSCVValue)}
	for _, locomotive := range backup.Locomotives {
		result.values[locomotive.ObjectID] = storedDatabaseCVs(locomotive.Values)
	}
	for _, locomotive := range backup.StoredLocomotives {
		result.unverified[locomotive.ObjectID] = storedDatabaseCVs(locomotive.UnverifiedCVs)
	}
	return result
}

func readStoredBackupConfiguration(dataDir string) ([]byte, []byte, error) {
	if dataDir == "" {
		return nil, nil, os.ErrNotExist
	}
	root, err := os.OpenRoot(dataDir)
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
			return nil, ErrECoSBackupUnavailable
		}
		data, err := io.ReadAll(io.LimitReader(file, limit+1))
		if err != nil || int64(len(data)) > limit {
			clear(data)
			return nil, ErrECoSBackupUnavailable
		}
		return data, nil
	}
	key, err := read("keyring.gpg", 64<<10)
	if err != nil {
		return nil, nil, err
	}
	passphrase, err := read("passphrase", 4<<10)
	if err != nil {
		clear(key)
		return nil, nil, ErrECoSBackupUnavailable
	}
	passphrase = bytes.TrimRight(passphrase, "\r\n")
	if strings.TrimSpace(string(passphrase)) == "" {
		clear(key)
		clear(passphrase)
		return nil, nil, ErrECoSBackupUnavailable
	}
	return key, passphrase, nil
}

func storedDatabaseCVs(values []ecospkg.DatabaseCV) []ECoSCVValue {
	result := make([]ECoSCVValue, 0, len(values))
	for _, value := range values {
		result = append(result, ECoSCVValue{Number: value.Number, Value: value.Value,
			IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32})
	}
	return result
}

func mergeStoredCVSources(database, backup []ECoSCVValue, databaseOK bool, backupStatus string) ([]ECoSCVValue, *ECoSCVSourceInfo, error) {
	info := &ECoSCVSourceInfo{Sources: []string{}, BackupStatus: backupStatus, Conflicts: []ECoSCVSourceConflict{}}
	if databaseOK {
		info.Sources = append(info.Sources, "database")
	}
	if backupStatus == "checked" && len(backup) > 0 {
		info.Sources = append(info.Sources, "backup")
	}
	if backupStatus == "failed" {
		info.Warning = "Die gespeicherte ECoS-Sicherung konnte nicht geprüft werden. / The stored ECoS backup could not be checked."
	}
	if backupStatus == "not_configured" {
		info.Warning = "Die interne Sicherungsauswertung ist nicht eingerichtet. / Internal backup inspection is not configured."
	}
	if !databaseOK {
		info.Warning = "Der ECoS-Datenbankexport konnte nicht gelesen werden. / The ECoS database export could not be read."
	}
	primary, err := validateStoredECoSCVs(database)
	if err != nil {
		return nil, info, err
	}
	secondary, err := validateStoredECoSCVs(backup)
	if err != nil {
		return nil, info, err
	}
	type address struct{ number, page int }
	values := map[address]ECoSCVValue{}
	for _, value := range primary {
		values[address{value.Number, eCoSCVPage(value)}] = value
	}
	for _, value := range secondary {
		key := address{value.Number, eCoSCVPage(value)}
		if existing, found := values[key]; found {
			if existing.Value != value.Value {
				info.Conflicts = append(info.Conflicts, ECoSCVSourceConflict{Number: value.Number,
					IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32, DatabaseValue: existing.Value, BackupValue: value.Value})
				delete(values, key) // Never silently import either side of an unresolved source conflict.
			}
			continue
		}
		values[key] = value
		info.BackupAdded++
	}
	result := make([]ECoSCVValue, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	validated, err := validateStoredECoSCVs(result)
	return validated, info, err
}
