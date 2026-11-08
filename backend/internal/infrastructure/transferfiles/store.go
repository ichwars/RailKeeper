// Package transferfiles owns the private filesystem effects of import/export jobs.
package transferfiles

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"railkeeper/backend/internal/transfer"
	"railkeeper/backend/internal/transfer/workflow"
)

type Store struct{ dataDir string }

func New(dataDir string) *Store { return &Store{dataDir: strings.TrimSpace(dataDir)} }

var _ workflow.FileStore = (*Store)(nil)

func (s *Store) StageUpload(
	_ context.Context,
	jobID string,
	format transfer.TransferFormat,
	sourceName, declaredMIME string,
	source io.Reader,
) (workflow.Upload, error) {
	name, err := validateDataTransferUploadName(format, sourceName)
	if err != nil {
		return nil, err
	}
	path, digest, detected, err := s.storeDataTransferUpload(jobID, source)
	if err != nil {
		return nil, err
	}
	if err := validateDataTransferUploadMIME(format, declaredMIME, detected); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("open transfer upload: %w", err)
	}
	return &managedUpload{File: file, metadata: workflow.UploadMetadata{
		Name: name, SHA256: digest, DetectedMIME: detected,
	}}, nil
}

type managedUpload struct {
	*os.File
	metadata workflow.UploadMetadata
}

func (u *managedUpload) Metadata() workflow.UploadMetadata { return u.metadata }

func (u *managedUpload) Close() error {
	closeErr := u.File.Close()
	removeErr := os.Remove(u.File.Name())
	if errors.Is(removeErr, os.ErrNotExist) {
		removeErr = nil
	}
	return errors.Join(closeErr, removeErr)
}

func (s *Store) OpenArtifact(relativePath string) (workflow.ArtifactReader, error) {
	if _, err := ensureDataTransferExportDirectory(s.dataDir); err != nil {
		return nil, err
	}
	path, err := resolveDataTransferArtifactPath(s.dataDir, relativePath)
	if err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, fmt.Errorf("inspect transfer artifact: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, transfer.ErrDataTransferArtifactPath
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transfer artifact: %w", err)
	}
	return &artifactReader{File: file}, nil
}

type artifactReader struct{ *os.File }

func (r *artifactReader) Metadata() (workflow.ArtifactMetadata, error) {
	info, err := r.File.Stat()
	if err != nil {
		return workflow.ArtifactMetadata{}, err
	}
	return workflow.ArtifactMetadata{Size: info.Size(), ModifiedAt: info.ModTime()}, nil
}

func (s *Store) RemoveArtifact(relativePath string) error {
	if _, err := ensureDataTransferExportDirectory(s.dataDir); err != nil {
		return err
	}
	path, err := resolveDataTransferArtifactPath(s.dataDir, relativePath)
	if err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect transfer artifact for deletion: %w", err)
	}
	if err == nil {
		if !info.Mode().IsRegular() {
			return transfer.ErrDataTransferArtifactPath
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("delete transfer artifact: %w", err)
		}
	}
	return nil
}

func (s *Store) ExportDirectory() (string, error) {
	directory, err := ensureDataTransferExportDirectory(s.dataDir)
	if err != nil {
		return "", err
	}
	resolved, err := resolveDataTransferArtifactPath(s.dataDir, "exports")
	if err != nil || resolved != directory {
		return "", transfer.ErrDataTransferArtifactPath
	}
	return resolved, nil
}
