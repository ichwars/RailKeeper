package vehicledata

import (
	"context"
	"database/sql"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type expiryStore struct{ db *sql.DB }

func NewExpiryStore(db *sql.DB) vehiclemodel.ExpiryRepository { return &expiryStore{db: db} }
func (s *expiryStore) ResetDirectAssignments(ctx context.Context, now string) error {
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

	return nil
}
func (s *expiryStore) ResetLegacyMatches(ctx context.Context, now string) error {

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
