package application

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/infrastructure/vehicledata"
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
	if err := validateVehicleFunctionChanges(input.FunctionChanges); err != nil {
		return CreateVehicleInput{}, err
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
	if err = vehicledata.Insert(ctx, tx, vehicleWriteRecord(vehicle)); err != nil {
		return nil, err
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
	if err = applyVehicleFunctionChanges(ctx, tx, vehicle.ID, input.FunctionChanges, now); err != nil {
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
