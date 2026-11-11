package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func ListCore(ctx context.Context, db *sql.DB, query string, vehicleSetID string) ([]vehiclemodel.Vehicle, error) {
	like := "%" + strings.TrimSpace(query) + "%"
	rows, err := db.QueryContext(ctx, `
SELECT v.id, v.inventory_number, v.manufacturer, COALESCE(v.article_number, ''), COALESCE(v.article_source_url, ''), v.name, v.gauge,
       COALESCE(v.epoch, ''), COALESCE(v.railway_company, ''), COALESCE(v.category, ''), COALESCE(v.gattung, ''),
	       COALESCE(v.description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
	       maximum_speed_kmh, COALESCE(home_base, ''),
	       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''), COALESCE(decoder_type, ''),
	       exhibition_ready, exhibition, abc_brakes, COALESCE(v.ean, ''), COALESCE(v.production_period, ''), COALESCE(v.list_price, ''),
	       COALESCE(v.acquisition_type, ''), COALESCE(v.acquired_from, ''), COALESCE(v.purchase_price, ''), COALESCE(v.purchase_date, ''),
	       COALESCE(v.storage_location, ''), COALESCE(v.condition, ''), COALESCE(v.packaging, ''),
	       COALESCE(length_mm, ''), COALESCE(weight_g, ''), COALESCE(color, ''), COALESCE(lettering, ''),
	       COALESCE(load, ''), COALESCE(interior, ''), COALESCE(axles, ''), COALESCE(axle_count, ''),
	       COALESCE(traction_tire_count, ''), COALESCE(wheelset, ''), COALESCE(coupling_front, ''),
	       COALESCE(coupling_rear, ''), COALESCE(power_pickup, ''), COALESCE(adapter, ''),
	       drive_enabled, headlights_enabled, lighting_enabled, sound_generator_enabled, smoke_generator_enabled,
	       COALESCE(member.vehicle_set_id, ''), COALESCE(sets.name, ''), COALESCE(member.position, 0),
       COALESCE(set_sizes.member_count, 0),
       COALESCE(sets.inventory_number, ''), COALESCE(sets.manufacturer, ''), COALESCE(sets.article_number, ''),
       COALESCE(sets.gauge, ''), COALESCE(sets.epoch, ''), COALESCE(sets.acquisition_type, ''),
       COALESCE(sets.purchase_date, ''), COALESCE(sets.purchase_price, ''), COALESCE(sets.condition, ''),
       v.created_at, v.updated_at
FROM vehicles v
LEFT JOIN vehicle_set_members member ON member.vehicle_id=v.id
LEFT JOIN vehicle_sets sets ON sets.id=member.vehicle_set_id
LEFT JOIN (
    SELECT vehicle_set_id, COUNT(*) AS member_count FROM vehicle_set_members GROUP BY vehicle_set_id
) set_sizes ON set_sizes.vehicle_set_id=member.vehicle_set_id
WHERE (? = '' OR member.vehicle_set_id=?)
  AND (
       ? = '%%'
       OR v.inventory_number LIKE ? COLLATE NOCASE
       OR v.manufacturer LIKE ? COLLATE NOCASE
       OR v.article_number LIKE ? COLLATE NOCASE
       OR v.name LIKE ? COLLATE NOCASE
       OR v.series LIKE ? COLLATE NOCASE
       OR v.vehicle_number LIKE ? COLLATE NOCASE
       OR v.decoder_type LIKE ? COLLATE NOCASE
       OR v.home_base LIKE ? COLLATE NOCASE
       OR CAST(v.maximum_speed_kmh AS TEXT) LIKE ? COLLATE NOCASE
       OR sets.inventory_number LIKE ? COLLATE NOCASE
       OR sets.name LIKE ? COLLATE NOCASE
       OR sets.article_number LIKE ? COLLATE NOCASE
  )
ORDER BY v.updated_at DESC, v.inventory_number ASC
`, vehicleSetID, vehicleSetID, like, like, like, like, like, like, like, like, like, like, like, like, like)
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	vehicles := []vehiclemodel.Vehicle{}
	for rows.Next() {
		var vehicle vehiclemodel.Vehicle
		var digital int
		var dtDecoder int
		var exhibitionReady int
		var exhibition int
		var abcBrakes int
		var driveEnabled int
		var headlightsEnabled int
		var lightingEnabled int
		var soundGeneratorEnabled int
		var smokeGeneratorEnabled int
		var maximumSpeed sql.NullInt64
		var setSummary vehiclemodel.VehicleSetSummary
		if err := rows.Scan(
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
			&vehicle.Condition,
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
			&vehicle.CouplingFront,
			&vehicle.CouplingRear,
			&vehicle.PowerPickup,
			&vehicle.Adapter,
			&driveEnabled,
			&headlightsEnabled,
			&lightingEnabled,
			&soundGeneratorEnabled,
			&smokeGeneratorEnabled,
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
			return nil, fmt.Errorf("scan vehicle: %w", err)
		}
		vehicle.Digital = digital == 1
		vehicle.DTDecoder = dtDecoder == 1
		vehicle.ExhibitionReady = exhibitionReady == 1
		vehicle.Exhibition = exhibition == 1
		vehicle.ABCBrakes = abcBrakes == 1
		vehicle.DriveEnabled = driveEnabled == 1
		vehicle.HeadlightsEnabled = headlightsEnabled == 1
		vehicle.LightingEnabled = lightingEnabled == 1
		vehicle.SoundGeneratorEnabled = soundGeneratorEnabled == 1
		vehicle.SmokeGeneratorEnabled = smokeGeneratorEnabled == 1
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
		vehicles = append(vehicles, vehicle)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicles: %w", err)
	}
	return vehicles, nil
}
