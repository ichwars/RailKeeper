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
SELECT id, inventory_number, manufacturer, COALESCE(article_number, ''), COALESCE(article_source_url, ''), name, gauge,
       COALESCE(epoch, ''), COALESCE(railway_company, ''), COALESCE(category, ''), COALESCE(gattung, ''),
	       COALESCE(description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
	       maximum_speed_kmh, COALESCE(home_base, ''),
	       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''), COALESCE(decoder_type, ''),
	       exhibition_ready, exhibition, abc_brakes, COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
	       COALESCE(acquisition_type, ''), COALESCE(acquired_from, ''), COALESCE(purchase_price, ''), COALESCE(purchase_date, ''),
	       COALESCE(storage_location, ''), COALESCE(condition, ''), COALESCE(packaging, ''),
	       COALESCE(length_mm, ''), COALESCE(weight_g, ''), COALESCE(color, ''), COALESCE(lettering, ''),
	       COALESCE(load, ''), COALESCE(interior, ''), COALESCE(axles, ''), COALESCE(axle_count, ''),
	       COALESCE(traction_tire_count, ''), COALESCE(wheelset, ''), COALESCE(coupling_front, ''),
	       COALESCE(coupling_rear, ''), COALESCE(power_pickup, ''), COALESCE(adapter, ''),
	       drive_enabled, headlights_enabled, lighting_enabled, sound_generator_enabled, smoke_generator_enabled,
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
	WHERE (? = '' OR EXISTS (
	  SELECT 1 FROM vehicle_set_members scope
	  WHERE scope.vehicle_id=vehicles.id AND scope.vehicle_set_id=?
	))
	AND (
	   ? = '%%'
	   OR inventory_number LIKE ? COLLATE NOCASE
	   OR manufacturer LIKE ? COLLATE NOCASE
	   OR article_number LIKE ? COLLATE NOCASE
	   OR name LIKE ? COLLATE NOCASE
	   OR series LIKE ? COLLATE NOCASE
	   OR vehicle_number LIKE ? COLLATE NOCASE
	   OR decoder_type LIKE ? COLLATE NOCASE
	   OR home_base LIKE ? COLLATE NOCASE
	   OR CAST(maximum_speed_kmh AS TEXT) LIKE ? COLLATE NOCASE
	   OR EXISTS (
	     SELECT 1
	     FROM vehicle_set_members m
	     JOIN vehicle_sets s ON s.id=m.vehicle_set_id
	     WHERE m.vehicle_id=vehicles.id
	       AND (s.inventory_number LIKE ? COLLATE NOCASE
	         OR s.name LIKE ? COLLATE NOCASE
	         OR s.article_number LIKE ? COLLATE NOCASE)
	   )
	)
	ORDER BY updated_at DESC, inventory_number ASC
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
