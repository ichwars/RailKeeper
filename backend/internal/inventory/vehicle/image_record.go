package vehicle

func ImageWriteRecord(image VehicleImage) ImageRecord {
	return ImageRecord{
		ID: image.ID, VehicleID: image.VehicleID, URL: image.URL, Title: image.Title,
		SourceURL: image.SourceURL, FileName: image.FileName, MimeType: image.MimeType,
		StoragePath: image.StoragePath, ThumbnailPath: image.ThumbnailPath,
		BlobID: image.BlobID, ThumbnailBlobID: image.ThumbnailBlobID,
		MaintenanceID: image.MaintenanceID, IsPrimary: image.IsPrimary, SortOrder: image.SortOrder,
		CreatedAt: image.CreatedAt, UpdatedAt: image.UpdatedAt,
	}
}
