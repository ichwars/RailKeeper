package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"railkeeper/backend/internal/application"

	"railkeeper/backend/internal/transfer"
	transferpolicy "railkeeper/backend/internal/transfer/policy"
)

func applyTransferVehicle(
	ctx context.Context,
	db *sql.Tx,
	record transfer.DataTransferPreviewRecord,
	action string,
	packageVersion int,
	csvMapping []transfer.DataTransferCSVColumnMapping,
	existingCSVVehicle *transfer.TransferVehicle,
) (transferVehicleApplyResult, error) {
	vehicle := transfer.TransferVehicle{}
	if err := json.Unmarshal(record.Data, &vehicle); err != nil {
		return transferVehicleApplyResult{}, dataTransferApplyConflict("vehicle preview data is invalid")
	}
	if err := application.ValidateTransferVehicle(vehicle); err != nil {
		return transferVehicleApplyResult{}, dataTransferApplyConflict("vehicle preview data violates aggregate validation")
	}
	now := timestamp()
	if action == "copy" {
		inventoryNumber, err := uniqueTransferInventoryNumber(ctx, db, "vehicles", vehicle.InventoryNumber)
		if err != nil {
			return transferVehicleApplyResult{}, err
		}
		vehicle.InventoryNumber = inventoryNumber
		action = "create"
	}
	switch action {
	case "create":
		vehicleID := randomID()
		arguments := transferVehicleArguments(vehicle, now)
		_, err := db.ExecContext(ctx, `
INSERT INTO vehicles(
  id, inventory_number, manufacturer, article_number, article_source_url, name, gauge, epoch, railway_company,
  category, gattung, description, series, vehicle_number, maximum_speed_kmh, home_base, digital,
  digital_decoder_number, dt_decoder, dt_decoder_number, decoder_type, exhibition_ready, exhibition, abc_brakes,
  ean, production_period, list_price, acquisition_type, acquired_from, purchase_price, purchase_date,
  storage_location, storage_details, condition, condition_details, packaging,
  length_mm, weight_g, color, lettering, load, interior, axles, axle_count, traction_tire_count, wheelset,
  coupling_same, coupling_front, coupling_rear, power_pickup, adapter,
  drive_enabled, drive_description, headlights_enabled, headlights_description, lighting_enabled,
  lighting_description, sound_generator_enabled, sound_generator_description, smoke_generator_enabled,
  smoke_generator_description, additional_info, qr_code_enabled, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
         ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			append([]any{vehicleID}, arguments...)...)
		if err != nil {
			return transferVehicleApplyResult{}, err
		}
		return transferVehicleApplyResult{ID: vehicleID, InventoryNumber: vehicle.InventoryNumber, Created: true}, nil
	case "replace":
		if record.TargetID == "" {
			return transferVehicleApplyResult{}, dataTransferApplyConflict("vehicle replacement target is missing")
		}
		if packageVersion == transfer.DataTransferPackageLegacyVersion {
			arguments := transferVehicleArguments(vehicle, now)
			result, err := db.ExecContext(ctx, `
UPDATE vehicles SET inventory_number=?, manufacturer=?, article_number=?, article_source_url=?, name=?, gauge=?,
  epoch=?, railway_company=?, category=?, gattung=?, description=?, series=?, vehicle_number=?, maximum_speed_kmh=?,
  home_base=?, digital=?, digital_decoder_number=?, dt_decoder=?, dt_decoder_number=?, decoder_type=?,
  exhibition_ready=?, exhibition=?, abc_brakes=?, ean=?, production_period=?, list_price=?, acquisition_type=?,
  acquired_from=?, purchase_price=?, purchase_date=?, storage_location=?, storage_details=?, condition=?,
  condition_details=?, packaging=?, updated_at=?
WHERE id=?`, append(append(arguments[0:35], now), record.TargetID)...)
			if err != nil {
				return transferVehicleApplyResult{}, err
			}
			if err := requireApplyUpdate(result, "replace legacy vehicle"); err != nil {
				return transferVehicleApplyResult{}, err
			}
			return transferVehicleApplyResult{ID: record.TargetID, InventoryNumber: vehicle.InventoryNumber}, nil
		}
		if len(csvMapping) > 0 {
			if existingCSVVehicle == nil {
				return transferVehicleApplyResult{}, dataTransferApplyConflict("vehicle replacement target is missing")
			}
			merged, err := transferpolicy.PreserveUnmappedTransferVehicleFields(
				vehicle, *existingCSVVehicle, csvMapping,
			)
			if err != nil {
				return transferVehicleApplyResult{}, err
			}
			vehicle = merged
		}
		arguments := transferVehicleArguments(vehicle, now)
		result, err := db.ExecContext(ctx, `
UPDATE vehicles SET inventory_number=?, manufacturer=?, article_number=?, article_source_url=?, name=?, gauge=?,
  epoch=?, railway_company=?, category=?, gattung=?, description=?, series=?, vehicle_number=?, maximum_speed_kmh=?,
  home_base=?, digital=?, digital_decoder_number=?, dt_decoder=?, dt_decoder_number=?, decoder_type=?,
  exhibition_ready=?, exhibition=?, abc_brakes=?, ean=?, production_period=?, list_price=?, acquisition_type=?,
  acquired_from=?, purchase_price=?, purchase_date=?, storage_location=?, storage_details=?, condition=?,
  condition_details=?, packaging=?, length_mm=?, weight_g=?, color=?, lettering=?, load=?, interior=?, axles=?,
  axle_count=?, traction_tire_count=?, wheelset=?, coupling_same=?, coupling_front=?, coupling_rear=?,
  power_pickup=?, adapter=?, drive_enabled=?, drive_description=?, headlights_enabled=?, headlights_description=?,
  lighting_enabled=?, lighting_description=?, sound_generator_enabled=?, sound_generator_description=?,
  smoke_generator_enabled=?, smoke_generator_description=?, additional_info=?, qr_code_enabled=?, updated_at=?
WHERE id=?`, append(append(arguments[0:62], now), record.TargetID)...)
		if err != nil {
			return transferVehicleApplyResult{}, err
		}
		if err := requireApplyUpdate(result, "replace vehicle"); err != nil {
			return transferVehicleApplyResult{}, err
		}
		return transferVehicleApplyResult{ID: record.TargetID, InventoryNumber: vehicle.InventoryNumber}, nil
	default:
		return transferVehicleApplyResult{}, dataTransferApplyConflict("unsupported vehicle resolution")
	}
}

func transferVehicleArguments(vehicle transfer.TransferVehicle, now string) []any {
	return []any{
		vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.ArticleSourceURL, vehicle.Name,
		vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description,
		vehicle.Series, vehicle.VehicleNumber, vehicle.MaximumSpeedKmh, vehicle.HomeBase, vehicle.Digital,
		vehicle.DigitalDecoderNumber, vehicle.DTDecoder, vehicle.DTDecoderNumber, vehicle.DecoderType,
		vehicle.ExhibitionReady, vehicle.Exhibition, vehicle.ABCBrakes, vehicle.EAN, vehicle.ProductionPeriod,
		vehicle.ListPrice, vehicle.AcquisitionType, vehicle.AcquiredFrom, vehicle.PurchasePrice, vehicle.PurchaseDate,
		vehicle.StorageLocation, vehicle.StorageDetails, vehicle.Condition, vehicle.ConditionDetails, vehicle.Packaging,
		vehicle.LengthMM, vehicle.WeightG, vehicle.Color, vehicle.Lettering, vehicle.Load, vehicle.Interior,
		vehicle.Axles, vehicle.AxleCount, vehicle.TractionTireCount, vehicle.Wheelset, boolToInt(vehicle.CouplingSame),
		vehicle.CouplingFront, vehicle.CouplingRear, vehicle.PowerPickup, vehicle.Adapter,
		boolToInt(vehicle.DriveEnabled), vehicle.DriveDescription, boolToInt(vehicle.HeadlightsEnabled),
		vehicle.HeadlightsDescription, boolToInt(vehicle.LightingEnabled), vehicle.LightingDescription,
		boolToInt(vehicle.SoundGeneratorEnabled), vehicle.SoundGeneratorDescription,
		boolToInt(vehicle.SmokeGeneratorEnabled), vehicle.SmokeGeneratorDescription, vehicle.AdditionalInfo,
		boolToInt(vehicle.QRCodeEnabled),
		now, now,
	}
}
