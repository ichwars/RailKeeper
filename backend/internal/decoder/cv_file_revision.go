package decoder

import "railkeeper/backend/internal/revision"

func CVFileRevision(file VehicleCVFile, sourceReferences string) string {
	file.Revision = ""
	return revision.Hash(struct {
		File                                  VehicleCVFile
		StoragePath, BlobID, SourceReferences string
	}{file, file.StoragePath, file.BlobID, sourceReferences})
}
