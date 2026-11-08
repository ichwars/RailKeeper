package transferfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"railkeeper/backend/internal/transfer"
	"railkeeper/backend/internal/transfer/workflow"
)

func (s *Store) storeDataTransferUpload(
	jobID string,
	source io.Reader,
) (path string, digest string, mimeType string, err error) {
	directory, err := ensureDataTransferImportDirectory(s.dataDir)
	if err != nil {
		return "", "", "", err
	}
	file, err := os.CreateTemp(directory, safeDataTransferImportName(jobID)+"-*.upload")
	if err != nil {
		return "", "", "", fmt.Errorf("create transfer upload temp file: %w", err)
	}
	temporaryPath := file.Name()
	path = temporaryPath
	remove := true
	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close transfer upload: %w", closeErr)
		}
		if remove || err != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(file, hash), io.LimitReader(source, workflow.DataTransferMaxUploadBytes+1))
	if err != nil {
		return "", "", "", fmt.Errorf("store transfer upload: %w", err)
	}
	if written > workflow.DataTransferMaxUploadBytes {
		return "", "", "", transfer.ErrDataTransferUploadTooLarge
	}
	if err := file.Sync(); err != nil {
		return "", "", "", fmt.Errorf("sync transfer upload: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", "", fmt.Errorf("rewind transfer upload: %w", err)
	}
	prefix := make([]byte, 512)
	read, readErr := file.Read(prefix)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		return "", "", "", fmt.Errorf("inspect transfer upload: %w", readErr)
	}
	remove = false
	return path, hex.EncodeToString(hash.Sum(nil)), http.DetectContentType(prefix[:read]), nil
}

func ensureDataTransferImportDirectory(dataDir string) (string, error) {
	base, err := filepath.Abs(strings.TrimSpace(dataDir))
	if err != nil || strings.TrimSpace(dataDir) == "" {
		return "", transfer.ErrDataTransferUploadPath
	}
	directory := filepath.Join(base, "imports")
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return "", fmt.Errorf("create transfer import directory: %w", err)
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return "", fmt.Errorf("inspect transfer import directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", transfer.ErrDataTransferUploadPath
	}
	return directory, nil
}

func safeDataTransferImportName(jobID string) string {
	name := strings.TrimSuffix(safeDataTransferArtifactName(jobID, ""), "")
	return strings.TrimPrefix(name, "railkeeper-")
}

func validateDataTransferUploadName(format transfer.TransferFormat, sourceName string) (string, error) {
	sourceName = strings.TrimSpace(strings.ReplaceAll(sourceName, "\\", "/"))
	sourceName = filepath.Base(sourceName)
	if sourceName == "" || sourceName == "." {
		return "", fmt.Errorf("%w: upload filename is required", transfer.ErrDataTransferValidation)
	}
	extension := strings.ToLower(filepath.Ext(sourceName))
	if format == transfer.TransferCSV && extension != ".csv" || format == transfer.TransferJSON && extension != ".json" {
		return "", fmt.Errorf("%w: upload extension does not match job format", transfer.ErrDataTransferValidation)
	}
	return sourceName, nil
}

func validateDataTransferUploadMIME(format transfer.TransferFormat, declared, detected string) error {
	declared = normalizeTransferMIME(declared)
	detected = normalizeTransferMIME(detected)
	allowedDeclared := map[transfer.TransferFormat]map[string]bool{
		transfer.TransferCSV: {
			"": true, "application/octet-stream": true, "text/csv": true, "text/plain": true,
			"application/vnd.ms-excel": true,
		},
		transfer.TransferJSON: {"": true, "application/octet-stream": true, "application/json": true, "text/plain": true},
	}
	allowedDetected := map[transfer.TransferFormat]map[string]bool{
		transfer.TransferCSV:  {"text/plain": true, "text/csv": true, "application/octet-stream": true},
		transfer.TransferJSON: {"application/json": true, "text/plain": true, "application/octet-stream": true},
	}
	if !allowedDeclared[format][declared] || !allowedDetected[format][detected] {
		return fmt.Errorf("%w: upload MIME type does not match job format", transfer.ErrDataTransferValidation)
	}
	return nil
}

func normalizeTransferMIME(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if separator := strings.IndexByte(value, ';'); separator >= 0 {
		value = strings.TrimSpace(value[:separator])
	}
	return value
}
