package application

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

var (
	ErrVehicleCVFileConflict     = errors.New("vehicle CV file or its source references changed")
	ErrVehicleConflict           = errors.New("vehicle changed since the draft was opened")
	ErrVehicleImageConflict      = errors.New("vehicle images changed since the draft was opened")
	ErrVehicleAttachmentConflict = errors.New("vehicle attachment changed since the draft was opened")
)

func recordRevision(value any) string {
	data, _ := json.Marshal(value)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func vehicleCoreRevision(vehicle Vehicle) string {
	var set *VehicleSetSummary
	if vehicle.VehicleSet != nil {
		copy := *vehicle.VehicleSet
		copy.MainImage = nil // Independent media is guarded separately.
		set = &copy
	}
	return recordRevision(struct {
		Core vehicledata.Record
		Set  *VehicleSetSummary
	}{vehicleWriteRecord(vehicle), set})
}

func vehicleImageRevision(image VehicleImage) string {
	return recordRevision(vehicledata.ImageRecord{
		ID: image.ID, VehicleID: image.VehicleID, URL: image.URL, Title: image.Title,
		SourceURL: image.SourceURL, FileName: image.FileName, MimeType: image.MimeType,
		StoragePath: image.StoragePath, ThumbnailPath: image.ThumbnailPath,
		BlobID: image.BlobID, ThumbnailBlobID: image.ThumbnailBlobID,
		MaintenanceID: image.MaintenanceID, IsPrimary: image.IsPrimary, SortOrder: image.SortOrder,
		CreatedAt: image.CreatedAt, UpdatedAt: image.UpdatedAt,
	})
}

func vehicleImagesRevision(images []VehicleImage) string {
	items := make([]struct{ ID, Revision string }, len(images))
	for index, image := range images {
		items[index].ID, items[index].Revision = image.ID, image.Revision
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return recordRevision(items)
}

func vehicleAttachmentRevision(attachment VehicleAttachment) string {
	attachment.Revision = ""
	return recordRevision(struct {
		Attachment          VehicleAttachment
		StoragePath, BlobID string
	}{attachment, attachment.StoragePath, attachment.BlobID})
}

func checkVehicleRevision(actual string, expected *string, conflict error) error {
	if expected == nil {
		return nil // Legacy clients keep their existing request contract.
	}
	if !validCVRevision(*expected) {
		return ErrVehicleValidation
	}
	if actual != *expected {
		return conflict
	}
	return nil
}

func optionalVehicleRevision(expected []string) *string {
	if len(expected) == 0 {
		return nil
	}
	return &expected[0]
}

func (s *VehicleService) cleanupLocalizedVehicleImages(ctx context.Context, images []VehicleImageInput) {
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	blobs := NewFileBlobService(s.db, "")
	for _, image := range images {
		_ = blobs.DeleteIfUnreferenced(cleanupCtx, image.BlobID)
		_ = blobs.DeleteIfUnreferenced(cleanupCtx, image.ThumbnailBlobID)
	}
}
