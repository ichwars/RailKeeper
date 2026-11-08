package application

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	"railkeeper/backend/internal/infrastructure/workshopdata"
	"railkeeper/backend/internal/inventory/workshop"
)

func NewVehicleService(db *sql.DB) *VehicleService {
	rows := workshopdata.New(db)
	service := &VehicleService{db: db, workshopRows: rows}
	service.workshop = workshop.New(rows, func(ctx context.Context, id string) error { _, err := service.Get(ctx, id); return err })
	return service
}

func (s *VehicleService) SetImageLocalizer(localizer VehicleImageLocalizer) {
	s.imageLocalizer = localizer
}

func (s *VehicleService) List(ctx context.Context, query string) ([]Vehicle, error) {
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}
	return s.list(ctx, query, "")
}

func (s *VehicleService) ListReadOnly(ctx context.Context, query string) ([]Vehicle, error) {
	return s.list(ctx, query, "")
}

func (s *VehicleService) list(ctx context.Context, query string, vehicleSetID string) ([]Vehicle, error) {
	vehicles, err := vehicledata.ListCore(ctx, s.db, query, vehicleSetID)
	if err != nil {
		return nil, err
	}
	if err := s.attachImages(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachVehicleSetMainImages(ctx, vehicles); err != nil {
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
	vehicle.ExternalMappings, err = s.loadVehicleExternalMappings(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.ECoSDecoderProfile, err = s.loadECoSDecoderProfile(ctx, vehicle.ID)
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
