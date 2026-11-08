package transferfiles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/transfer"
	"railkeeper/backend/internal/transfer/workflow"
)

func (s *Store) WriteExport(
	ctx context.Context,
	repository workflow.ArtifactRecorder,
	jobID string,
	displayName string,
	mimeType string,
	payload []byte,
) (artifact transfer.DataTransferArtifact, err error) {
	directory, err := ensureDataTransferExportDirectory(s.dataDir)
	if err != nil {
		return transfer.DataTransferArtifact{}, err
	}
	temporary, err := os.CreateTemp(directory, ".railkeeper-export-*.tmp")
	if err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("create transfer artifact temp file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() {
		_ = temporary.Close()
		if err != nil {
			_ = os.Remove(temporaryPath)
		}
	}()
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(temporary, hash), bytes.NewReader(payload))
	if err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("write transfer artifact: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("sync transfer artifact: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("close transfer artifact: %w", err)
	}
	fileName := safeDataTransferArtifactName(jobID+"-"+credentials.RandomID(), filepath.Ext(displayName))
	relativePath := filepath.ToSlash(filepath.Join("exports", fileName))
	finalPath, err := resolveDataTransferArtifactPath(s.dataDir, relativePath)
	if err != nil {
		return transfer.DataTransferArtifact{}, err
	}
	if filepath.Dir(finalPath) != directory {
		return transfer.DataTransferArtifact{}, transfer.ErrDataTransferArtifactPath
	}
	if err := os.Rename(temporaryPath, finalPath); err != nil {
		return transfer.DataTransferArtifact{}, fmt.Errorf("publish transfer artifact: %w", err)
	}
	artifact, err = repository.CreateArtifact(ctx, transfer.DataTransferArtifact{
		JobID: jobID, RelativePath: relativePath, DisplayName: displayName, MIMEType: mimeType,
		SizeBytes: written, SHA256: hex.EncodeToString(hash.Sum(nil)),
	})
	if err != nil {
		_ = os.Remove(finalPath)
		return transfer.DataTransferArtifact{}, err
	}
	return artifact, nil
}

func ensureDataTransferExportDirectory(dataDir string) (string, error) {
	directory, err := resolveDataTransferArtifactPath(dataDir, "exports")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return "", fmt.Errorf("create transfer export directory: %w", err)
	}
	info, err := os.Lstat(directory)
	if err != nil {
		return "", fmt.Errorf("inspect transfer export directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", transfer.ErrDataTransferArtifactPath
	}
	return directory, nil
}

func resolveDataTransferArtifactPath(dataDir, relativePath string) (string, error) {
	dataDir = strings.TrimSpace(dataDir)
	relativePath = strings.TrimSpace(relativePath)
	if dataDir == "" || relativePath == "" || filepath.IsAbs(relativePath) {
		return "", transfer.ErrDataTransferArtifactPath
	}
	segments := strings.FieldsFunc(relativePath, func(value rune) bool { return value == '/' || value == '\\' })
	if len(segments) == 0 || !strings.EqualFold(segments[0], "exports") {
		return "", transfer.ErrDataTransferArtifactPath
	}
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return "", transfer.ErrDataTransferArtifactPath
		}
	}
	base, err := filepath.Abs(dataDir)
	if err != nil {
		return "", fmt.Errorf("resolve transfer data directory: %w", err)
	}
	exportsDirectory := filepath.Join(base, "exports")
	target, err := filepath.Abs(filepath.Join(base, filepath.FromSlash(strings.ReplaceAll(relativePath, "\\", "/"))))
	if err != nil {
		return "", fmt.Errorf("resolve transfer artifact: %w", err)
	}
	if target != exportsDirectory && !strings.HasPrefix(target, exportsDirectory+string(os.PathSeparator)) {
		return "", transfer.ErrDataTransferArtifactPath
	}
	return target, nil
}

func safeDataTransferArtifactName(jobID, extension string) string {
	var builder strings.Builder
	for _, value := range strings.ToLower(strings.TrimSpace(jobID)) {
		if value >= 'a' && value <= 'z' || value >= '0' && value <= '9' || value == '-' || value == '_' {
			builder.WriteRune(value)
		}
	}
	if builder.Len() == 0 {
		builder.WriteString("export")
	}
	return "railkeeper-" + builder.String() + extension
}
