package vehicle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"railkeeper/backend/internal/revision"
)

func FunctionRevision(item VehicleFunction) string {
	item.Revision = ""
	data, _ := json.Marshal(item)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func CoreRevision(vehicle Vehicle) string {
	var set *VehicleSetSummary
	if vehicle.VehicleSet != nil {
		copy := *vehicle.VehicleSet
		copy.MainImage = nil // Independent media is guarded separately.
		set = &copy
	}
	return revision.Hash(struct {
		Core Record
		Set  *VehicleSetSummary
	}{WriteRecord(vehicle), set})
}
func ImageRevision(image VehicleImage) string { return revision.Hash(ImageWriteRecord(image)) }
func ImagesRevision(images []VehicleImage) string {
	items := make([]struct{ ID, Revision string }, len(images))
	for index, image := range images {
		items[index].ID, items[index].Revision = image.ID, image.Revision
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return revision.Hash(items)
}
func AttachmentRevision(attachment VehicleAttachment) string {
	attachment.Revision = ""
	return revision.Hash(struct {
		Attachment          VehicleAttachment
		StoragePath, BlobID string
	}{attachment, attachment.StoragePath, attachment.BlobID})
}
func CheckRevision(actual string, expected *string, conflict error) error {
	if expected == nil {
		return nil // Legacy clients keep their existing request contract.
	}
	if !revision.Valid(*expected) {
		return ErrVehicleValidation
	}
	if actual != *expected {
		return conflict
	}
	return nil
}
func OptionalRevision(expected []string) *string {
	if len(expected) == 0 {
		return nil
	}
	return &expected[0]
}
