package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *VehicleService) Delete(ctx context.Context, id, actorUserID string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrVehicleNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete vehicle: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var vehicleSetID sql.NullString
	if err = tx.QueryRowContext(ctx, `
SELECT vehicle_set_id FROM vehicle_set_members WHERE vehicle_id=?
`, id).Scan(&vehicleSetID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read vehicle set membership: %w", err)
	}

	result, err := tx.ExecContext(ctx, `DELETE FROM vehicles WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete vehicle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if affected == 0 {
		_ = tx.Rollback()
		return ErrVehicleNotFound
	}
	if vehicleSetID.Valid {
		if _, err = tx.ExecContext(ctx, `
DELETE FROM vehicle_sets
WHERE id=? AND NOT EXISTS (SELECT 1 FROM vehicle_set_members WHERE vehicle_set_id=?)
`, vehicleSetID.String, vehicleSetID.String); err != nil {
			return fmt.Errorf("delete empty vehicle set: %w", err)
		}
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleDeleted', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, id, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("write vehicle audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit delete vehicle: %w", err)
	}

	return nil
}

func (s *VehicleService) get(ctx context.Context, id string) (*Vehicle, error) {
	var vehicle Vehicle
	var digital int
	var dtDecoder int
	var exhibitionReady int
	var exhibition int
	var abcBrakes int
	var couplingSame int
	var driveEnabled int
	var headlightsEnabled int
	var lightingEnabled int
	var soundGeneratorEnabled int
	var smokeGeneratorEnabled int
	var qrCodeEnabled int
	var maximumSpeed sql.NullInt64
	var setSummary VehicleSetSummary
	if err := s.db.QueryRowContext(ctx, `
SELECT id, inventory_number, manufacturer, COALESCE(article_number, ''), COALESCE(article_source_url, ''), name, gauge,
       COALESCE(epoch, ''), COALESCE(railway_company, ''), COALESCE(category, ''), COALESCE(gattung, ''),
       COALESCE(description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
       maximum_speed_kmh, COALESCE(home_base, ''),
       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''), COALESCE(decoder_type, ''),
       exhibition_ready, exhibition, abc_brakes, COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
       COALESCE(acquisition_type, ''), COALESCE(acquired_from, ''), COALESCE(purchase_price, ''), COALESCE(purchase_date, ''),
       COALESCE(storage_location, ''), COALESCE(storage_details, ''), COALESCE(condition, ''), COALESCE(condition_details, ''), COALESCE(packaging, ''),
       COALESCE(length_mm, ''), COALESCE(weight_g, ''), COALESCE(color, ''), COALESCE(lettering, ''),
       COALESCE(load, ''), COALESCE(interior, ''), COALESCE(axles, ''), COALESCE(axle_count, ''),
       COALESCE(traction_tire_count, ''), COALESCE(wheelset, ''),
       coupling_same, COALESCE(coupling_front, ''), COALESCE(coupling_rear, ''), COALESCE(power_pickup, ''), COALESCE(adapter, ''),
       drive_enabled, COALESCE(drive_description, ''), headlights_enabled, COALESCE(headlights_description, ''),
	       lighting_enabled, COALESCE(lighting_description, ''), sound_generator_enabled, COALESCE(sound_generator_description, ''),
	       smoke_generator_enabled, COALESCE(smoke_generator_description, ''), COALESCE(additional_info, ''), qr_code_enabled,
	       COALESCE((SELECT m.vehicle_set_id FROM vehicle_set_members m WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.name FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT m.position FROM vehicle_set_members m WHERE m.vehicle_id=vehicles.id), 0),
	       COALESCE((SELECT COUNT(*) FROM vehicle_set_members m WHERE m.vehicle_set_id=(SELECT own.vehicle_set_id FROM vehicle_set_members own WHERE own.vehicle_id=vehicles.id)), 0),
	       COALESCE((SELECT s.inventory_number FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.manufacturer FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.article_number FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.gauge FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.epoch FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.acquisition_type FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.purchase_date FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.purchase_price FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       COALESCE((SELECT s.condition FROM vehicle_set_members m JOIN vehicle_sets s ON s.id=m.vehicle_set_id WHERE m.vehicle_id=vehicles.id), ''),
	       created_at, updated_at
FROM vehicles
WHERE id=?
`, id).Scan(
		&vehicle.ID,
		&vehicle.InventoryNumber,
		&vehicle.Manufacturer,
		&vehicle.ArticleNumber,
		&vehicle.ArticleSourceURL,
		&vehicle.Name,
		&vehicle.Gauge,
		&vehicle.Epoch,
		&vehicle.RailwayCompany,
		&vehicle.Category,
		&vehicle.Gattung,
		&vehicle.Description,
		&vehicle.Series,
		&vehicle.VehicleNumber,
		&maximumSpeed,
		&vehicle.HomeBase,
		&digital,
		&vehicle.DigitalDecoderNumber,
		&dtDecoder,
		&vehicle.DTDecoderNumber,
		&vehicle.DecoderType,
		&exhibitionReady,
		&exhibition,
		&abcBrakes,
		&vehicle.EAN,
		&vehicle.ProductionPeriod,
		&vehicle.ListPrice,
		&vehicle.AcquisitionType,
		&vehicle.AcquiredFrom,
		&vehicle.PurchasePrice,
		&vehicle.PurchaseDate,
		&vehicle.StorageLocation,
		&vehicle.StorageDetails,
		&vehicle.Condition,
		&vehicle.ConditionDetails,
		&vehicle.Packaging,
		&vehicle.LengthMM,
		&vehicle.WeightG,
		&vehicle.Color,
		&vehicle.Lettering,
		&vehicle.Load,
		&vehicle.Interior,
		&vehicle.Axles,
		&vehicle.AxleCount,
		&vehicle.TractionTireCount,
		&vehicle.Wheelset,
		&couplingSame,
		&vehicle.CouplingFront,
		&vehicle.CouplingRear,
		&vehicle.PowerPickup,
		&vehicle.Adapter,
		&driveEnabled,
		&vehicle.DriveDescription,
		&headlightsEnabled,
		&vehicle.HeadlightsDescription,
		&lightingEnabled,
		&vehicle.LightingDescription,
		&soundGeneratorEnabled,
		&vehicle.SoundGeneratorDescription,
		&smokeGeneratorEnabled,
		&vehicle.SmokeGeneratorDescription,
		&vehicle.AdditionalInfo,
		&qrCodeEnabled,
		&vehicle.VehicleSetID,
		&vehicle.VehicleSetName,
		&vehicle.VehicleSetPosition,
		&vehicle.VehicleSetSize,
		&setSummary.InventoryNumber,
		&setSummary.Manufacturer,
		&setSummary.ArticleNumber,
		&setSummary.Gauge,
		&setSummary.Epoch,
		&setSummary.AcquisitionType,
		&setSummary.PurchaseDate,
		&setSummary.PurchasePrice,
		&setSummary.Condition,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	vehicle.Digital = digital == 1
	vehicle.DTDecoder = dtDecoder == 1
	vehicle.ExhibitionReady = exhibitionReady == 1
	vehicle.Exhibition = exhibition == 1
	vehicle.ABCBrakes = abcBrakes == 1
	vehicle.CouplingSame = couplingSame == 1
	vehicle.DriveEnabled = driveEnabled == 1
	vehicle.HeadlightsEnabled = headlightsEnabled == 1
	vehicle.LightingEnabled = lightingEnabled == 1
	vehicle.SoundGeneratorEnabled = soundGeneratorEnabled == 1
	vehicle.SmokeGeneratorEnabled = smokeGeneratorEnabled == 1
	vehicle.QRCodeEnabled = qrCodeEnabled == 1
	if maximumSpeed.Valid {
		value := int(maximumSpeed.Int64)
		vehicle.MaximumSpeedKmh = &value
	}
	if vehicle.VehicleSetID != "" {
		setSummary.ID = vehicle.VehicleSetID
		setSummary.Name = vehicle.VehicleSetName
		setSummary.MemberCount = vehicle.VehicleSetSize
		setSummary.Position = vehicle.VehicleSetPosition
		vehicle.VehicleSet = &setSummary
	}
	images, err := s.loadVehicleImages(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Images = images
	if vehicle.VehicleSet != nil {
		_, _, _, mainImage, err := s.resolveSetMainImage(ctx, vehicle.VehicleSet.ID)
		if err != nil {
			return nil, err
		}
		vehicle.VehicleSet.MainImage = mainImage
	}
	attachments, err := s.loadVehicleAttachments(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Attachments = attachments
	maintenance, err := s.loadVehicleMaintenance(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Maintenance = maintenance
	spareParts, err := s.loadVehicleSpareParts(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.SpareParts = spareParts
	functions, err := s.loadVehicleFunctions(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.Functions = functions
	cvValues, err := s.loadVehicleCVValues(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.CVValues = cvValues
	cvFiles, err := s.loadVehicleCVFiles(ctx, id)
	if err != nil {
		return nil, err
	}
	vehicle.CVFiles = cvFiles

	return &vehicle, nil
}
