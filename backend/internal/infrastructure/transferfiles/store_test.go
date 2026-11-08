package transferfiles

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/transfer"
	"railkeeper/backend/internal/transfer/workflow"
)

func assertNoStagedFiles(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(directory, "imports"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("upload left private scratch files behind", len(entries))
	}
}

func TestManagedUploadPreservesOriginalBytesAndRemovesScratchFile(t *testing.T) {
	directory := t.TempDir()
	store := New(directory)
	payload := []byte("Inventarnummer;Bezeichnung\r\nRK-1;Straßenbahn\r\n")
	upload, err := store.StageUpload(t.Context(), "../../job", transfer.TransferCSV,
		`C:\incoming\old.csv`, "text/csv; charset=utf-8", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	metadata := upload.Metadata()
	digest := sha256.Sum256(payload)
	actual, err := io.ReadAll(upload)
	if err != nil || !bytes.Equal(actual, payload) || metadata.Name != "old.csv" ||
		metadata.SHA256 != hex.EncodeToString(digest[:]) || metadata.DetectedMIME == "" {
		t.Fatal("staging altered original content or metadata", metadata, err)
	}
	if err := upload.Close(); err != nil {
		t.Fatal(err)
	}
	assertNoStagedFiles(t, directory)
}

type brokenUploadReader struct{ failure error }

func (r brokenUploadReader) Read(buffer []byte) (int, error) {
	return copy(buffer, "partial private upload"), r.failure
}

type repeatedUploadByte struct{}

func (repeatedUploadByte) Read(buffer []byte) (int, error) {
	for index := range buffer {
		buffer[index] = 'a'
	}
	return len(buffer), nil
}

func TestFailedUploadsAlwaysRemoveStagedBytes(t *testing.T) {
	sentinel := errors.New("synthetic source failure")
	for _, test := range []struct {
		name, filename, mime string
		reader               io.Reader
		want                 error
	}{
		{"reader", "rows.csv", "text/csv", brokenUploadReader{sentinel}, sentinel},
		{"oversized", "rows.csv", "text/csv", io.LimitReader(repeatedUploadByte{}, workflow.DataTransferMaxUploadBytes+1), transfer.ErrDataTransferUploadTooLarge},
		{"mime", "rows.csv", "text/html", strings.NewReader("a;b\n"), transfer.ErrDataTransferValidation},
		{"extension", "rows.exe", "text/csv", strings.NewReader("a;b\n"), transfer.ErrDataTransferValidation},
		{"missing-name", ".", "text/csv", strings.NewReader("a;b\n"), transfer.ErrDataTransferValidation},
	} {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			upload, err := New(directory).StageUpload(t.Context(), "job", transfer.TransferCSV,
				test.filename, test.mime, test.reader)
			if !errors.Is(err, test.want) || upload != nil {
				t.Fatal("invalid upload produced content or lost the cause", err)
			}
			assertNoStagedFiles(t, directory)
		})
	}
}

type artifactRecorder struct {
	failure error
	value   transfer.DataTransferArtifact
}

func (r *artifactRecorder) CreateArtifact(_ context.Context, value transfer.DataTransferArtifact) (transfer.DataTransferArtifact, error) {
	r.value = value
	return value, r.failure
}

func TestArtifactPublicationCompensatesMetadataFailureAndPreservesDownload(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(map[bool]string{true: "metadata-failure", false: "download"}[failed], func(t *testing.T) {
			directory := t.TempDir()
			store := New(directory)
			recorder := &artifactRecorder{}
			if failed {
				recorder.failure = errors.New("synthetic metadata failure")
			}
			payload := []byte("original export bytes\r\n")
			artifact, err := store.WriteExport(t.Context(), recorder, "job", "rows.csv", "text/csv", payload)
			if failed {
				if !errors.Is(err, recorder.failure) {
					t.Fatal("lost metadata failure", err)
				}
				entries, readErr := os.ReadDir(filepath.Join(directory, "exports"))
				if readErr != nil || len(entries) != 0 {
					t.Fatal("uncommitted export left content behind", readErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			file, err := store.OpenArtifact(artifact.RelativePath)
			if err != nil {
				t.Fatal(err)
			}
			metadata, err := file.Metadata()
			if err != nil || metadata.Size != int64(len(payload)) || metadata.ModifiedAt.IsZero() {
				t.Fatal("download metadata lost", metadata, err)
			}
			if _, err := file.Seek(9, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			part, err := io.ReadAll(file)
			if err != nil || !bytes.Equal(part, payload[9:]) {
				t.Fatal("partial download changed bytes", err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
			if _, err := file.Metadata(); err == nil {
				t.Fatal("closed descriptor reported valid metadata")
			}
			if err := store.RemoveArtifact(artifact.RelativePath); err != nil {
				t.Fatal(err)
			}
			if err := store.RemoveArtifact(artifact.RelativePath); err != nil {
				t.Fatal("deleting an already missing file is not idempotent", err)
			}
		})
	}
}
