package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *CoreService) Create(ctx context.Context, input CreateVehicleInput, actorUserID string) (*Vehicle, error) {
	return s.CreateWithOptions(ctx, input, CreationOptions{}, actorUserID)
}

// CreateWithExternalMapping binds the source identity in the vehicle transaction.
// Its unique existing mapping prevents a repeated adoption from creating another vehicle.
func (s *CoreService) CreateWithExternalMapping(
	ctx context.Context, input CreateVehicleInput, mapping VehicleExternalMapInput, actorUserID string,
) (*Vehicle, error) {
	return s.CreateWithOptions(ctx, input, CreationOptions{ExternalMapping: &mapping}, actorUserID)
}

func (s *CoreService) CreateWithOptions(
	ctx context.Context, input CreateVehicleInput, options CreationOptions, actorUserID string,
) (*Vehicle, error) {
	input, err := ValidateCreateInput(input)
	if err != nil {
		return nil, err
	}
	if options.ExternalMapping != nil {
		cleaned := CleanExternalMappingInput(*options.ExternalMapping)
		if cleaned.Provider == "" || cleaned.ExternalID == "" {
			return nil, ErrVehicleValidation
		}
		options.ExternalMapping = &cleaned
	}
	vehicleID, err := ResolveCreationIdentity(ctx, s.repository, options.CreationID)
	if err != nil {
		return nil, err
	}
	if len(input.Images) > 0 {
		input.Images, err = s.environment.LocalizeImages(ctx, vehicleID, input.Images)
		if err != nil {
			return nil, err
		}
	}

	defer s.environment.CleanupImages(ctx, input.Images)

	tx, err := s.repository.BeginVehicle(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin create vehicle: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if options.CreationID != "" {
		if err := tx.CheckCreationIdentity(ctx, vehicleID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	vehicle, err := InsertVehicle(ctx, tx, vehicleID, input, actorUserID, now)
	if err != nil {
		return nil, err
	}
	var savedMapping *VehicleExternalMap
	if options.ExternalMapping != nil {
		savedMapping, err = tx.WriteSourceMapping(ctx, NewExternalMappingWrite(vehicle.ID, *options.ExternalMapping, now))
		if err != nil {
			return nil, err
		}
	}

	vehicle, err = tx.ReadCore(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	images, err := tx.Images(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	vehicle.ImagesRevision = ImagesRevision(images)
	if savedMapping != nil {
		vehicle.ExternalMappings = []VehicleExternalMap{*savedMapping}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create vehicle: %w", err)
	}
	if err = s.environment.CompleteCreated(ctx, vehicle); err != nil {
		return nil, err
	}

	return vehicle, nil
}

func (s *CoreService) Update(ctx context.Context, id string, input CreateVehicleInput, actorUserID string) (*Vehicle, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrVehicleNotFound
	}

	existing, err := s.environment.Existing(ctx, id)
	if err != nil {
		return nil, err
	}

	replaceImages := input.Images != nil
	input = CleanVehicleInput(input)
	if input.InventoryNumber == "" {
		input.InventoryNumber = existing.InventoryNumber
	}
	if input.Manufacturer == "" || input.Name == "" || input.Gauge == "" || input.Category == "" ||
		input.Gattung == "" {
		return nil, ErrVehicleValidation
	}
	if !IsValidVehicleOperationalInput(input) {
		return nil, ErrVehicleOperationalValidation
	}
	if err := ValidateFunctionChanges(input.FunctionChanges); err != nil {
		return nil, err
	}
	if replaceImages && len(input.Images) > 0 {
		input.Images, err = s.environment.LocalizeImages(ctx, id, input.Images)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	vehicle := Vehicle{
		ID:                        id,
		VehicleSetID:              existing.VehicleSetID,
		VehicleSetName:            existing.VehicleSetName,
		VehicleSetPosition:        existing.VehicleSetPosition,
		VehicleSetSize:            existing.VehicleSetSize,
		InventoryNumber:           input.InventoryNumber,
		Manufacturer:              input.Manufacturer,
		ArticleNumber:             input.ArticleNumber,
		ArticleSourceURL:          input.ArticleSourceURL,
		Name:                      input.Name,
		Gauge:                     input.Gauge,
		Epoch:                     input.Epoch,
		RailwayCompany:            input.RailwayCompany,
		Category:                  input.Category,
		Gattung:                   input.Gattung,
		Description:               input.Description,
		Series:                    input.Series,
		VehicleNumber:             input.VehicleNumber,
		MaximumSpeedKmh:           input.MaximumSpeedKmh,
		HomeBase:                  input.HomeBase,
		Digital:                   input.Digital,
		DigitalDecoderNumber:      input.DigitalDecoderNumber,
		DTDecoder:                 input.DTDecoder,
		DTDecoderNumber:           input.DTDecoderNumber,
		DecoderType:               input.DecoderType,
		ExhibitionReady:           input.ExhibitionReady,
		Exhibition:                input.Exhibition,
		ABCBrakes:                 input.ABCBrakes,
		EAN:                       input.EAN,
		ProductionPeriod:          input.ProductionPeriod,
		ListPrice:                 input.ListPrice,
		AcquisitionType:           input.AcquisitionType,
		AcquiredFrom:              input.AcquiredFrom,
		PurchasePrice:             input.PurchasePrice,
		PurchaseDate:              input.PurchaseDate,
		StorageLocation:           input.StorageLocation,
		StorageDetails:            input.StorageDetails,
		Condition:                 input.Condition,
		ConditionDetails:          input.ConditionDetails,
		Packaging:                 input.Packaging,
		LengthMM:                  input.LengthMM,
		WeightG:                   input.WeightG,
		Color:                     input.Color,
		Lettering:                 input.Lettering,
		Load:                      input.Load,
		Interior:                  input.Interior,
		Axles:                     input.Axles,
		AxleCount:                 input.AxleCount,
		TractionTireCount:         input.TractionTireCount,
		Wheelset:                  input.Wheelset,
		CouplingSame:              input.CouplingSame,
		CouplingFront:             input.CouplingFront,
		CouplingRear:              input.CouplingRear,
		PowerPickup:               input.PowerPickup,
		Adapter:                   input.Adapter,
		DriveEnabled:              input.DriveEnabled,
		DriveDescription:          input.DriveDescription,
		HeadlightsEnabled:         input.HeadlightsEnabled,
		HeadlightsDescription:     input.HeadlightsDescription,
		LightingEnabled:           input.LightingEnabled,
		LightingDescription:       input.LightingDescription,
		SoundGeneratorEnabled:     input.SoundGeneratorEnabled,
		SoundGeneratorDescription: input.SoundGeneratorDescription,
		SmokeGeneratorEnabled:     input.SmokeGeneratorEnabled,
		SmokeGeneratorDescription: input.SmokeGeneratorDescription,
		AdditionalInfo:            input.AdditionalInfo,
		QRCodeEnabled:             input.QRCodeEnabled,
		Images:                    ImagesFromInput(id, input.Images, now),
		CreatedAt:                 existing.CreatedAt,
		UpdatedAt:                 now,
	}

	defer s.environment.CleanupImages(ctx, input.Images)

	tx, err := s.repository.BeginVehicle(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin update vehicle: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, err = tx.ReadCore(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := CheckRevision(existing.Revision, input.ExpectedRevision, ErrVehicleConflict); err != nil {
		return nil, err
	}
	if replaceImages {
		images, err := tx.Images(ctx, id)
		if err != nil {
			return nil, err
		}
		if err := CheckRevision(ImagesRevision(images), input.ExpectedImagesRevision, ErrVehicleImageConflict); err != nil {
			return nil, err
		}
	}

	if vehicle.InventoryNumber != existing.InventoryNumber {
		if err = EnsureInventoryNumberAvailable(ctx, tx, vehicle.InventoryNumber, vehicle.ID); err != nil {
			return nil, err
		}
	}

	affected, err := tx.UpdateVehicle(ctx, WriteRecord(vehicle))
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		_ = tx.Rollback()
		return nil, ErrVehicleNotFound
	}

	if vehicle.InventoryNumber != existing.InventoryNumber {
		if err = tx.RecordNumberChange(ctx, VehicleNumberChange{ID: credentials.RandomID(), VehicleID: vehicle.ID, Previous: existing.InventoryNumber, Current: vehicle.InventoryNumber, Actor: actorUserID, At: now}); err != nil {
			return nil, err
		}
	}

	if err = tx.AuditUpdated(ctx, VehicleAudit{ID: credentials.RandomID(), VehicleID: vehicle.ID, Actor: actorUserID, At: now}); err != nil {
		return nil, err
	}
	if replaceImages {
		if err = ReplaceImages(ctx, tx, vehicle.ID, input.Images, now); err != nil {
			return nil, err
		}
	}

	if err = ApplyFunctionChanges(ctx, tx, vehicle.ID, input.FunctionChanges, now); err != nil {
		return nil, err
	}

	stored, err := tx.ReadCore(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle = *stored
	images, err := tx.Images(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	vehicle.ImagesRevision = ImagesRevision(images)

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update vehicle: %w", err)
	}
	if err = s.environment.CompleteUpdated(ctx, &vehicle); err != nil {
		return nil, err
	}

	return &vehicle, nil
}
