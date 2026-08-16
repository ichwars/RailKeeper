package application

import (
	"context"
	"database/sql"
	"fmt"
)

func validateCreateVehicleInput(input CreateVehicleInput) (CreateVehicleInput, error) {
	input = cleanVehicleInput(input)
	if input.Manufacturer == "" || input.Name == "" || input.Gauge == "" || input.Category == "" ||
		input.Gattung == "" {
		return CreateVehicleInput{}, ErrVehicleValidation
	}
	if !isValidVehicleOperationalInput(input) {
		return CreateVehicleInput{}, ErrVehicleOperationalValidation
	}
	return input, nil
}

func (s *VehicleService) insertVehicleTx(
	ctx context.Context,
	tx *sql.Tx,
	vehicleID string,
	input CreateVehicleInput,
	actorUserID string,
	now string,
) (*Vehicle, error) {
	var err error
	if input.InventoryNumber == "" {
		input.InventoryNumber, err = s.nextInventoryNumber(ctx, tx, input.Category)
		if err != nil {
			return nil, err
		}
	} else if err = s.ensureInventoryNumberAvailable(ctx, tx, input.InventoryNumber, ""); err != nil {
		return nil, err
	}

	vehicle := vehicleFromCreateInput(vehicleID, input, now)
	if _, err = tx.ExecContext(ctx, `
INSERT INTO vehicles(
  id, inventory_number, manufacturer, article_number, article_source_url, name, gauge, epoch, railway_company, category, gattung,
  description, series, vehicle_number, maximum_speed_kmh, home_base,
  digital, digital_decoder_number, dt_decoder, dt_decoder_number, decoder_type,
  exhibition_ready, exhibition, abc_brakes, ean, production_period, list_price,
  acquisition_type, acquired_from, purchase_price, purchase_date, storage_location, storage_details, condition, condition_details, packaging,
  length_mm, weight_g, color, lettering, load, interior, axles, axle_count, traction_tire_count, wheelset,
  coupling_same, coupling_front, coupling_rear, power_pickup, adapter,
  drive_enabled, drive_description, headlights_enabled, headlights_description, lighting_enabled, lighting_description,
  sound_generator_enabled, sound_generator_description, smoke_generator_enabled, smoke_generator_description,
  additional_info, qr_code_enabled, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, vehicle.ID, vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.ArticleSourceURL, vehicle.Name, vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description, vehicle.Series, vehicle.VehicleNumber, nullableInt(vehicle.MaximumSpeedKmh), vehicle.HomeBase, boolToInt(vehicle.Digital), vehicle.DigitalDecoderNumber, boolToInt(vehicle.DTDecoder), vehicle.DTDecoderNumber, vehicle.DecoderType, boolToInt(vehicle.ExhibitionReady), boolToInt(vehicle.Exhibition), boolToInt(vehicle.ABCBrakes), vehicle.EAN, vehicle.ProductionPeriod, vehicle.ListPrice, vehicle.AcquisitionType, vehicle.AcquiredFrom, vehicle.PurchasePrice, vehicle.PurchaseDate, vehicle.StorageLocation, vehicle.StorageDetails, vehicle.Condition, vehicle.ConditionDetails, vehicle.Packaging, vehicle.LengthMM, vehicle.WeightG, vehicle.Color, vehicle.Lettering, vehicle.Load, vehicle.Interior, vehicle.Axles, vehicle.AxleCount, vehicle.TractionTireCount, vehicle.Wheelset, boolToInt(vehicle.CouplingSame), vehicle.CouplingFront, vehicle.CouplingRear, vehicle.PowerPickup, vehicle.Adapter, boolToInt(vehicle.DriveEnabled), vehicle.DriveDescription, boolToInt(vehicle.HeadlightsEnabled), vehicle.HeadlightsDescription, boolToInt(vehicle.LightingEnabled), vehicle.LightingDescription, boolToInt(vehicle.SoundGeneratorEnabled), vehicle.SoundGeneratorDescription, boolToInt(vehicle.SmokeGeneratorEnabled), vehicle.SmokeGeneratorDescription, vehicle.AdditionalInfo, boolToInt(vehicle.QRCodeEnabled), vehicle.CreatedAt, vehicle.UpdatedAt); err != nil {
		return nil, fmt.Errorf("insert vehicle: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleCreated', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, vehicle.ID, now); err != nil {
		return nil, fmt.Errorf("write vehicle audit log: %w", err)
	}
	if err = saveVehicleImages(ctx, tx, vehicle.ID, input.Images, now); err != nil {
		return nil, err
	}
	return &vehicle, nil
}

func vehicleFromCreateInput(vehicleID string, input CreateVehicleInput, now string) Vehicle {
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
		QRCodeEnabled: input.QRCodeEnabled, Images: vehicleImagesFromInput(vehicleID, input.Images, now),
		CreatedAt: now, UpdatedAt: now,
	}
}
