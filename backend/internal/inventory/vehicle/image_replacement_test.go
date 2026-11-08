package vehicle

import (
	"context"
	"errors"
	"testing"
)

type imageReplacementProbe struct {
	existing         map[string]ImageMetadata
	inserted         []ImageRecord
	cleared          bool
	maintenanceError error
}

func (p *imageReplacementProbe) ExistingImages(context.Context, string) (map[string]ImageMetadata, error) {
	return p.existing, nil
}
func (p *imageReplacementProbe) ClearImages(context.Context, string) error {
	p.cleared = true
	return nil
}
func (p *imageReplacementProbe) InsertImage(_ context.Context, row ImageRecord) error {
	p.inserted = append(p.inserted, row)
	return nil
}
func (p *imageReplacementProbe) MaintenanceBelongs(context.Context, string, string) (bool, error) {
	return p.maintenanceError == nil, p.maintenanceError
}

func TestImageReplacementRetainsOriginalIdentityPrivateFilesAndStoredZeroOrder(t *testing.T) {
	metadata := ImageMetadata{ID: "original", URL: "https://example.invalid/original", FileName: "original.png", MimeType: "image/png",
		StoragePath: "private/original", ThumbnailPath: "private/thumbnail", BlobID: "blob", ThumbnailBlobID: "thumbnail-blob", CreatedAt: "recorded", SortOrder: 0, Position: 0}
	for _, byID := range []bool{true, false} {
		probe := &imageReplacementProbe{existing: map[string]ImageMetadata{metadata.ID: metadata, metadata.URL: metadata}}
		input := VehicleImageInput{URL: metadata.URL, FileName: "replacement.png", StoragePath: "private/replacement", BlobID: "replacement-blob", Title: "new title"}
		if byID {
			input.ID = metadata.ID
			input.URL = "https://example.invalid/new-link"
		}
		if err := ReplaceImages(t.Context(), probe, "vehicle", []VehicleImageInput{input}, "later"); err != nil {
			t.Fatal(err)
		}
		if !probe.cleared || len(probe.inserted) != 1 {
			t.Fatal("replacement not written", probe)
		}
		row := probe.inserted[0]
		if row.ID != metadata.ID || row.CreatedAt != metadata.CreatedAt || row.StoragePath != metadata.StoragePath || row.BlobID != metadata.BlobID ||
			row.ThumbnailPath != metadata.ThumbnailPath || row.ThumbnailBlobID != metadata.ThumbnailBlobID || row.FileName != metadata.FileName ||
			row.MimeType != metadata.MimeType || row.SortOrder != 0 || row.Title != "new title" || row.UpdatedAt != "later" ||
			row.URL != "/api/v1/vehicles/vehicle/images/original/file" || probe.existing[metadata.ID] != metadata {
			t.Fatal("original image metadata replaced", row, metadata)
		}
	}
}

func TestImageReplacementReturnsMaintenanceFailureToItsTransactionOwner(t *testing.T) {
	failure := errors.New("maintenance lookup unavailable")
	probe := &imageReplacementProbe{maintenanceError: failure}
	err := ReplaceImages(t.Context(), probe, "vehicle", []VehicleImageInput{{URL: "https://example.invalid/image", MaintenanceID: "maintenance"}}, "later")
	if !errors.Is(err, failure) || !probe.cleared || len(probe.inserted) != 0 {
		t.Fatal("failed replacement acknowledged", probe, err)
	}
}
