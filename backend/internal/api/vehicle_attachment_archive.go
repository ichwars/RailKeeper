package api

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"railkeeper/backend/internal/application"
)

const maxAttachmentArchiveBytes int64 = 256 << 20

var errArchiveTooLarge = errors.New("attachment selection exceeds archive size limit")

func (a *App) downloadVehicleAttachmentArchive(w http.ResponseWriter, r *http.Request) {
	ids := r.URL.Query()["file"]
	if len(ids) == 0 || len(ids) > 100 {
		respondProblem(w, http.StatusBadRequest, "attachment_archive_selection", "Select between 1 and 100 attachments.")
		return
	}
	file, err := os.CreateTemp("", "railkeeper-attachments-*.zip")
	if err != nil {
		respondProblem(w, http.StatusInternalServerError, "attachment_archive_failed", "Cannot prepare archive.")
		return
	}
	defer func() { _ = os.Remove(file.Name()) }()
	defer func() { _ = file.Close() }()
	archive := zip.NewWriter(file)
	remaining := maxAttachmentArchiveBytes
	seen := make(map[string]bool, len(ids))
	for index, id := range ids {
		if id == "" || seen[id] {
			respondProblem(w, http.StatusBadRequest, "attachment_archive_selection", "Invalid or duplicate attachment.")
			return
		}
		seen[id] = true
		size, err := a.archiveAttachment(r.Context(), archive, r.PathValue("id"), id, index+1, remaining)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrVehicleNotFound):
				respondProblem(w, http.StatusNotFound, "attachment_not_found", "Attachment not found.")
			case errors.Is(err, errArchiveTooLarge):
				respondProblem(w, http.StatusRequestEntityTooLarge, "attachment_archive_size", "Select up to 256 MiB per archive.")
			default:
				a.logger.Error("attachment archive failed", "error", err)
				respondProblem(w, http.StatusInternalServerError, "attachment_archive_failed", "Cannot prepare archive.")
			}
			return
		}
		remaining -= size
	}
	if err := archive.Close(); err != nil {
		respondProblem(w, http.StatusInternalServerError, "attachment_archive_failed", "Cannot finish archive.")
		return
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="RailKeeper-Unterlagen.zip"`)
	http.ServeContent(w, r, "RailKeeper-Unterlagen.zip", time.Time{}, file)
}

func (a *App) archiveAttachment(ctx context.Context, archive *zip.Writer, vehicleID, id string, index int, remaining int64) (int64, error) {
	attachment, err := a.vehicleService.GetAttachment(ctx, vehicleID, id)
	if err != nil {
		return 0, err
	}
	if attachment.SizeBytes > remaining {
		return 0, errArchiveTooLarge
	}
	var reader io.Reader
	if attachment.BlobID != "" {
		data, err := a.loadFileBlob(ctx, attachment.BlobID)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(data)
	} else {
		fullPath, err := confinedDataPath(a.dataDir, attachment.StoragePath)
		if err != nil {
			return 0, err
		}
		file, err := os.Open(fullPath)
		if err != nil {
			return 0, err
		}
		defer func() { _ = file.Close() }()
		reader = file
	}
	// Numeric prefixes preserve selection order and distinguish identical filenames.
	name := fmt.Sprintf("%03d-%s", index, safeAttachmentFileName(attachment.OriginalName))
	entry, err := archive.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
	if err != nil {
		return 0, err
	}
	size, err := io.Copy(entry, io.LimitReader(reader, remaining+1))
	if size > remaining {
		return 0, errArchiveTooLarge
	}
	return size, err
}
