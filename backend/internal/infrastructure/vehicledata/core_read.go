package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

const vehicleCoreSelect = `
SELECT v.id, v.inventory_number, v.manufacturer, COALESCE(v.article_number, ''), COALESCE(v.article_source_url, ''), v.name, v.gauge,
       COALESCE(v.epoch, ''), COALESCE(v.railway_company, ''), COALESCE(v.category, ''), COALESCE(v.gattung, ''),
       COALESCE(v.description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
       maximum_speed_kmh, COALESCE(home_base, ''),
       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''), COALESCE(decoder_type, ''),
       exhibition_ready, exhibition, abc_brakes, COALESCE(v.ean, ''), COALESCE(v.production_period, ''), COALESCE(v.list_price, ''),
       COALESCE(v.acquisition_type, ''), COALESCE(v.acquired_from, ''), COALESCE(v.purchase_price, ''), COALESCE(v.purchase_date, ''),
       COALESCE(v.storage_location, ''), COALESCE(v.storage_details, ''), COALESCE(v.condition, ''), COALESCE(v.condition_details, ''), COALESCE(v.packaging, ''),
       COALESCE(length_mm, ''), COALESCE(weight_g, ''), COALESCE(color, ''), COALESCE(lettering, ''),
       COALESCE(load, ''), COALESCE(interior, ''), COALESCE(axles, ''), COALESCE(axle_count, ''),
       COALESCE(traction_tire_count, ''), COALESCE(wheelset, ''),
       coupling_same, COALESCE(coupling_front, ''), COALESCE(coupling_rear, ''), COALESCE(power_pickup, ''), COALESCE(adapter, ''),
       drive_enabled, COALESCE(drive_description, ''), headlights_enabled, COALESCE(headlights_description, ''),
	       lighting_enabled, COALESCE(lighting_description, ''), sound_generator_enabled, COALESCE(sound_generator_description, ''),
	       smoke_generator_enabled, COALESCE(smoke_generator_description, ''), COALESCE(additional_info, ''), qr_code_enabled,
	       COALESCE(member.vehicle_set_id, ''), COALESCE(sets.name, ''), COALESCE(member.position, 0),
       (SELECT COUNT(*) FROM vehicle_set_members WHERE vehicle_set_id=member.vehicle_set_id),
       COALESCE(sets.inventory_number, ''), COALESCE(sets.manufacturer, ''), COALESCE(sets.article_number, ''),
       COALESCE(sets.gauge, ''), COALESCE(sets.epoch, ''), COALESCE(sets.acquisition_type, ''),
       COALESCE(sets.purchase_date, ''), COALESCE(sets.purchase_price, ''), COALESCE(sets.condition, ''),
       v.created_at, v.updated_at
FROM vehicles v
LEFT JOIN vehicle_set_members member ON member.vehicle_id=v.id
LEFT JOIN vehicle_sets sets ON sets.id=member.vehicle_set_id
`

func ReadCore(ctx context.Context, query queryRower, id string) (*vehiclemodel.Vehicle, error) {
	vehicle, err := scanVehicleCore(query.QueryRowContext(ctx, vehicleCoreSelect+`WHERE v.id=?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, vehiclemodel.ErrVehicleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	return vehicle, nil
}

func ReadCoresByID(ctx context.Context, query vehicleMediaQuery, vehicleIDsJSON string) ([]vehiclemodel.Vehicle, error) {
	rows, err := query.QueryContext(ctx, vehicleCoreSelect+`
JOIN json_each(?) requested ON requested.value=v.id
ORDER BY requested.key`, vehicleIDsJSON)
	if err != nil {
		return nil, fmt.Errorf("list requested vehicles: %w", err)
	}
	defer func() { _ = rows.Close() }()
	vehicles := []vehiclemodel.Vehicle{}
	for rows.Next() {
		vehicle, err := scanVehicleCore(rows)
		if err != nil {
			return nil, fmt.Errorf("scan requested vehicle: %w", err)
		}
		vehicles = append(vehicles, *vehicle)
	}
	return vehicles, rows.Err()
}

func scanVehicleCore(row interface{ Scan(...any) error }) (*vehiclemodel.Vehicle, error) {
	var vehicle vehiclemodel.Vehicle
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
	var setSummary vehiclemodel.VehicleSetSummary
	if err := row.Scan(
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
		return nil, err
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
	vehicle.Revision = vehiclemodel.CoreRevision(vehicle)
	return &vehicle, nil
}
