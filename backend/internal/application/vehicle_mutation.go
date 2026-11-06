package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func (s *VehicleService) Create(ctx context.Context, input CreateVehicleInput, actorUserID string) (*Vehicle, error) {
	input, err := validateCreateVehicleInput(input)
	if err != nil {
		return nil, err
	}
	vehicleID := randomID()
	if s.imageLocalizer != nil && len(input.Images) > 0 {
		input.Images, err = s.imageLocalizer(ctx, vehicleID, input.Images)
		if err != nil {
			return nil, err
		}
	}

	defer s.cleanupLocalizedVehicleImages(ctx, input.Images)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create vehicle: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	vehicle, err := s.insertVehicleTx(ctx, tx, vehicleID, input, actorUserID, now)
	if err != nil {
		return nil, err
	}

	vehicle, err = readVehicleCore(ctx, tx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	images, err := readVehicleImages(ctx, tx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	vehicle.ImagesRevision = vehicleImagesRevision(images)

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create vehicle: %w", err)
	}
	vehicle.Functions, err = s.loadVehicleFunctions(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}

	return vehicle, nil
}

func (s *VehicleService) Update(ctx context.Context, id string, input CreateVehicleInput, actorUserID string) (*Vehicle, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrVehicleNotFound
	}

	existing, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}

	replaceImages := input.Images != nil
	input = cleanVehicleInput(input)
	if input.InventoryNumber == "" {
		input.InventoryNumber = existing.InventoryNumber
	}
	if input.Manufacturer == "" || input.Name == "" || input.Gauge == "" || input.Category == "" ||
		input.Gattung == "" {
		return nil, ErrVehicleValidation
	}
	if !isValidVehicleOperationalInput(input) {
		return nil, ErrVehicleOperationalValidation
	}
	if err := validateVehicleFunctionChanges(input.FunctionChanges); err != nil {
		return nil, err
	}
	if s.imageLocalizer != nil && replaceImages && len(input.Images) > 0 {
		input.Images, err = s.imageLocalizer(ctx, id, input.Images)
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
		Images:                    vehicleImagesFromInput(id, input.Images, now),
		CreatedAt:                 existing.CreatedAt,
		UpdatedAt:                 now,
	}

	defer s.cleanupLocalizedVehicleImages(ctx, input.Images)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update vehicle: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	existing, err = readVehicleCore(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := checkVehicleRevision(existing.Revision, input.ExpectedRevision, ErrVehicleConflict); err != nil {
		return nil, err
	}
	if replaceImages {
		images, err := readVehicleImages(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if err := checkVehicleRevision(vehicleImagesRevision(images), input.ExpectedImagesRevision, ErrVehicleImageConflict); err != nil {
			return nil, err
		}
	}

	if vehicle.InventoryNumber != existing.InventoryNumber {
		if err = s.ensureInventoryNumberAvailable(ctx, tx, vehicle.InventoryNumber, vehicle.ID); err != nil {
			return nil, err
		}
	}

	affected, err := vehicledata.Update(ctx, tx, vehicleWriteRecord(vehicle))
	if err != nil {
		return nil, err
	}

	if affected == 0 {
		_ = tx.Rollback()
		return nil, ErrVehicleNotFound
	}

	if vehicle.InventoryNumber != existing.InventoryNumber {
		if _, err = tx.ExecContext(ctx, `
INSERT INTO inventory_number_history(id, vehicle_id, old_number, new_number, changed_by_user_id, changed_at, reason)
VALUES(?, ?, ?, ?, ?, ?, 'manual_update')
`, randomID(), vehicle.ID, existing.InventoryNumber, vehicle.InventoryNumber, actorUserID, now); err != nil {
			return nil, fmt.Errorf("write inventory number history: %w", err)
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleUpdated', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, vehicle.ID, now); err != nil {
		return nil, fmt.Errorf("write vehicle audit log: %w", err)
	}
	if replaceImages {
		if err = saveVehicleImages(ctx, tx, vehicle.ID, input.Images, now); err != nil {
			return nil, err
		}
	}

	if err = applyVehicleFunctionChanges(ctx, tx, vehicle.ID, input.FunctionChanges, now); err != nil {
		return nil, err
	}

	stored, err := readVehicleCore(ctx, tx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle = *stored
	images, err := readVehicleImages(ctx, tx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	vehicle.ImagesRevision = vehicleImagesRevision(images)

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update vehicle: %w", err)
	}
	if vehicle.VehicleSet != nil {
		_, _, _, mainImage, err := s.resolveSetMainImage(ctx, vehicle.VehicleSet.ID)
		if err != nil {
			return nil, err
		}
		vehicle.VehicleSet.MainImage = mainImage
	}
	attachments, err := s.loadVehicleAttachments(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Attachments = attachments
	maintenance, err := s.loadVehicleMaintenance(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Maintenance = maintenance
	functions, err := s.loadVehicleFunctions(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.Functions = functions
	cvValues, err := s.loadVehicleCVValues(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.CVValues = cvValues
	cvFiles, err := s.loadVehicleCVFiles(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.CVFiles = cvFiles
	externalMappings, err := s.loadVehicleExternalMappings(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.ExternalMappings = externalMappings
	vehicle.ECoSDecoderProfile, err = s.loadECoSDecoderProfile(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}

	return &vehicle, nil
}
