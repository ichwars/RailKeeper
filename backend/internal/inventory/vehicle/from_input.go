package vehicle

func FromCreateInput(vehicleID string, input CreateVehicleInput, now string) Vehicle {
	return Vehicle{
		ID: vehicleID, InventoryNumber: input.InventoryNumber, Manufacturer: input.Manufacturer,
		ArticleNumber: input.ArticleNumber, ArticleSourceURL: input.ArticleSourceURL, Name: input.Name,
		Gauge: input.Gauge, Epoch: input.Epoch, RailwayCompany: input.RailwayCompany, Category: input.Category,
		Gattung: input.Gattung, Description: input.Description, Series: input.Series,
		VehicleNumber: input.VehicleNumber, MaximumSpeedKmh: input.MaximumSpeedKmh, HomeBase: input.HomeBase,
		Digital: input.Digital, DigitalDecoderNumber: input.DigitalDecoderNumber, DTDecoder: input.DTDecoder,
		DTDecoderNumber: input.DTDecoderNumber, DecoderType: input.DecoderType,
		ExhibitionReady: input.ExhibitionReady, Exhibition: input.Exhibition, ABCBrakes: input.ABCBrakes,
		EAN: input.EAN, ProductionPeriod: input.ProductionPeriod, ListPrice: input.ListPrice,
		AcquisitionType: input.AcquisitionType, AcquiredFrom: input.AcquiredFrom,
		PurchasePrice: input.PurchasePrice, PurchaseDate: input.PurchaseDate,
		StorageLocation: input.StorageLocation, StorageDetails: input.StorageDetails,
		Condition: input.Condition, ConditionDetails: input.ConditionDetails, Packaging: input.Packaging,
		LengthMM: input.LengthMM, WeightG: input.WeightG, Color: input.Color, Lettering: input.Lettering,
		Load: input.Load, Interior: input.Interior, Axles: input.Axles, AxleCount: input.AxleCount,
		TractionTireCount: input.TractionTireCount, Wheelset: input.Wheelset,
		CouplingSame: input.CouplingSame, CouplingFront: input.CouplingFront, CouplingRear: input.CouplingRear,
		PowerPickup: input.PowerPickup, Adapter: input.Adapter, DriveEnabled: input.DriveEnabled,
		DriveDescription: input.DriveDescription, HeadlightsEnabled: input.HeadlightsEnabled,
		HeadlightsDescription: input.HeadlightsDescription, LightingEnabled: input.LightingEnabled,
		LightingDescription: input.LightingDescription, SoundGeneratorEnabled: input.SoundGeneratorEnabled,
		SoundGeneratorDescription: input.SoundGeneratorDescription,
		SmokeGeneratorEnabled:     input.SmokeGeneratorEnabled,
		SmokeGeneratorDescription: input.SmokeGeneratorDescription, AdditionalInfo: input.AdditionalInfo,
		QRCodeEnabled: input.QRCodeEnabled, Images: ImagesFromInput(vehicleID, input.Images, now),
		CreatedAt: now, UpdatedAt: now,
	}
}
func ImagesFromInput(vehicleID string, images []VehicleImageInput, now string) []VehicleImage {
	cleaned := CleanVehicleImageInputs(images)
	out := make([]VehicleImage, 0, len(cleaned))
	for index, image := range cleaned {
		sortOrder := image.SortOrder
		if sortOrder == 0 {
			sortOrder = index
		}
		out = append(out, VehicleImage{
			ID:              image.ID,
			VehicleID:       vehicleID,
			URL:             image.URL,
			Title:           image.Title,
			SourceURL:       image.SourceURL,
			FileName:        image.FileName,
			MimeType:        image.MimeType,
			StoragePath:     image.StoragePath,
			ThumbnailPath:   image.ThumbnailPath,
			BlobID:          image.BlobID,
			ThumbnailBlobID: image.ThumbnailBlobID,
			MaintenanceID:   image.MaintenanceID,
			IsPrimary:       image.IsPrimary,
			SortOrder:       sortOrder,
			CreatedAt:       now,
		})
	}
	return out
}
func WithImageURLs(image VehicleImage) VehicleImage {
	if image.StoragePath != "" || image.BlobID != "" {
		image.URL = "/api/v1/vehicles/" + image.VehicleID + "/images/" + image.ID + "/file"
	}
	if image.ThumbnailPath != "" || image.ThumbnailBlobID != "" {
		image.ThumbnailURL = "/api/v1/vehicles/" + image.VehicleID + "/images/" + image.ID + "/thumbnail"
	}
	return image
}
