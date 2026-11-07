package backup

import (
	"bytes"
	"encoding/json"
	"errors"
)

const (
	backupFormat  = "railkeeper-backup"
	backupVersion = 28
)

var (
	ErrBackupInvalid = errors.New("backup invalid")
	ErrBackupPath    = errors.New("backup path invalid")
)

type BackupDocument struct {
	Format    string                      `json:"format"`
	Version   int                         `json:"version"`
	CreatedAt string                      `json:"createdAt"`
	Tables    map[string][]map[string]any `json:"tables"`
	Profiles  []map[string]any            `json:"profiles"`
	Files     []BackupFile                `json:"files"`
}

type BackupFile struct {
	Path          string `json:"path"`
	SizeBytes     int64  `json:"sizeBytes"`
	SHA256        string `json:"sha256"`
	ContentBase64 string `json:"contentBase64"`
}

type BackupImportResult struct {
	RestoredTables int `json:"restoredTables"`
	RestoredRows   int `json:"restoredRows"`
	RestoredFiles  int `json:"restoredFiles"`
}

type BackupValidationResult struct {
	Compatible bool                    `json:"compatible"`
	Format     string                  `json:"format,omitempty"`
	Version    int                     `json:"version"`
	CreatedAt  string                  `json:"createdAt,omitempty"`
	TableCount int                     `json:"tableCount"`
	RowCount   int                     `json:"rowCount"`
	FileCount  int                     `json:"fileCount"`
	FileBytes  int64                   `json:"fileBytes"`
	Tables     []BackupValidationTable `json:"tables"`
	Warnings   []string                `json:"warnings"`
	Errors     []string                `json:"errors"`
}

type BackupValidationTable struct {
	Name           string   `json:"name"`
	Rows           int      `json:"rows"`
	Missing        bool     `json:"missing"`
	UnknownColumns []string `json:"unknownColumns,omitempty"`
}

const Format = backupFormat
const Version = backupVersion

func DecodeBackup(data []byte) (*BackupDocument, error) {
	var doc BackupDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil, ErrBackupInvalid
	}
	return &doc, nil
}
