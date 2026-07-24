package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

func NewVehicleService(db *sql.DB) *VehicleService {
	return &VehicleService{db: db}
}

func (s *VehicleService) SetImageLocalizer(localizer VehicleImageLocalizer) {
	s.imageLocalizer = localizer
}

func (s *VehicleService) List(ctx context.Context, query string) ([]Vehicle, error) {
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}

	like := "%" + strings.TrimSpace(query) + "%"
	rows, err := s.db.QueryContext(ctx, `
SELECT id, inventory_number, manufacturer, COALESCE(article_number, ''), COALESCE(article_source_url, ''), name, gauge,
       COALESCE(epoch, ''), COALESCE(railway_company, ''), COALESCE(category, ''), COALESCE(gattung, ''),
       COALESCE(description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''), COALESCE(decoder_type, ''),
       exhibition_ready, exhibition, abc_brakes, COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
       created_at, updated_at
FROM vehicles
WHERE ? = '%%'
   OR inventory_number LIKE ?
   OR manufacturer LIKE ?
   OR article_number LIKE ?
   OR name LIKE ?
ORDER BY updated_at DESC, inventory_number ASC
`, like, like, like, like, like)
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	vehicles := []Vehicle{}
	for rows.Next() {
		var vehicle Vehicle
		var digital int
		var dtDecoder int
		var exhibitionReady int
		var exhibition int
		var abcBrakes int
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
		vehicles = append(vehicles, vehicle)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicles: %w", err)
	}
	if err := s.attachImages(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachAttachments(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachMaintenance(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachSpareParts(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachFunctions(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachCVData(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachExternalMappings(ctx, vehicles); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (s *VehicleService) Get(ctx context.Context, id string) (*Vehicle, error) {
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}

	vehicle, err := s.get(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *VehicleService) resetExpiredExhibitionFlags(ctx context.Context) error {
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
WITH expired_vehicle_ids AS (
  SELECT DISTINCT e.vehicle_id
  FROM exhibition_entries e
  JOIN exhibition_lists l ON l.id = e.list_id
  WHERE e.vehicle_id <> ''
    AND date(l.list_date, '+1 day') <= date('now', 'localtime')
),
active_vehicle_ids AS (
  SELECT DISTINCT e.vehicle_id
  FROM exhibition_entries e
  JOIN exhibition_lists l ON l.id = e.list_id
  WHERE e.vehicle_id <> ''
    AND date(l.list_date, '+1 day') > date('now', 'localtime')
)
UPDATE vehicles
SET exhibition=0, updated_at=?
WHERE exhibition=1
  AND id IN (SELECT vehicle_id FROM expired_vehicle_ids)
  AND id NOT IN (SELECT vehicle_id FROM active_vehicle_ids)
`, now); err != nil {
		return fmt.Errorf("reset expired exhibition flags: %w", err)
	}

	if _, err := s.db.ExecContext(ctx, `
UPDATE vehicles
SET exhibition=0, updated_at=?
WHERE exhibition=1
  AND EXISTS (
    SELECT 1
    FROM exhibition_entries e
    JOIN exhibition_lists l ON l.id = e.list_id
    WHERE e.vehicle_id = ''
      AND date(l.list_date, '+1 day') <= date('now', 'localtime')
      AND (
        (e.decoder_number <> '' AND e.decoder_number = vehicles.digital_decoder_number)
        OR (
          e.locomotive_name <> ''
          AND lower(trim(e.locomotive_name)) = lower(trim(vehicles.name))
          AND (e.manufacturer = '' OR lower(trim(e.manufacturer)) = lower(trim(vehicles.manufacturer)))
        )
      )
  )
  AND NOT EXISTS (
    SELECT 1
    FROM exhibition_entries e
    JOIN exhibition_lists l ON l.id = e.list_id
    WHERE date(l.list_date, '+1 day') > date('now', 'localtime')
      AND (
        e.vehicle_id = vehicles.id
        OR (e.decoder_number <> '' AND e.decoder_number = vehicles.digital_decoder_number)
        OR (
          e.locomotive_name <> ''
          AND lower(trim(e.locomotive_name)) = lower(trim(vehicles.name))
          AND (e.manufacturer = '' OR lower(trim(e.manufacturer)) = lower(trim(vehicles.manufacturer)))
        )
      )
  )
`, now); err != nil {
		return fmt.Errorf("reset legacy expired exhibition flags: %w", err)
	}

	return nil
}
