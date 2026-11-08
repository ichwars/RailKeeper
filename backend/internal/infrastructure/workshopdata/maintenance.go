package workshopdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/domain"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }
func (s *Store) CreateMaintenance(ctx context.Context, maintenance domain.VehicleMaintenance) error {
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_maintenance(id, vehicle_id, kind, status, condition_rating, due_date, completed_at, cost, notes, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, maintenance.ID, maintenance.VehicleID, maintenance.Kind, maintenance.Status, maintenance.ConditionRating, maintenance.DueDate, maintenance.CompletedAt, maintenance.Cost, maintenance.Notes, maintenance.CreatedAt, maintenance.UpdatedAt); err != nil {
		return fmt.Errorf("create vehicle maintenance: %w", err)
	}
	return nil
}
func (s *Store) UpdateMaintenance(ctx context.Context, vehicleID, maintenanceID string, input domain.VehicleMaintenanceInput, now string) error {
	result, err := s.db.ExecContext(ctx, `
UPDATE vehicle_maintenance
SET kind=?, status=?, condition_rating=?, due_date=?, completed_at=?, cost=?, notes=?, updated_at=?
WHERE id=? AND vehicle_id=?
`, input.Kind, input.Status, input.ConditionRating, input.DueDate, input.CompletedAt, input.Cost, input.Notes, now, maintenanceID, vehicleID)
	if err != nil {
		return fmt.Errorf("update vehicle maintenance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle maintenance update result: %w", err)
	}
	if affected == 0 {
		return domain.ErrVehicleNotFound
	}
	return nil
}
func (s *Store) GetMaintenance(ctx context.Context, vehicleID, maintenanceID string) (*domain.VehicleMaintenance, error) {
	var maintenance domain.VehicleMaintenance
	err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, kind, status, COALESCE(condition_rating, ''), COALESCE(due_date, ''), COALESCE(completed_at, ''),
       COALESCE(cost, ''), COALESCE(notes, ''), created_at, updated_at
FROM vehicle_maintenance
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(maintenanceID), strings.TrimSpace(vehicleID)).Scan(
		&maintenance.ID,
		&maintenance.VehicleID,
		&maintenance.Kind,
		&maintenance.Status,
		&maintenance.ConditionRating,
		&maintenance.DueDate,
		&maintenance.CompletedAt,
		&maintenance.Cost,
		&maintenance.Notes,
		&maintenance.CreatedAt,
		&maintenance.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle maintenance: %w", err)
	}
	return &maintenance, nil
}
func (s *Store) DeleteMaintenance(ctx context.Context, vehicleID, maintenanceID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_maintenance WHERE id=? AND vehicle_id=?`, strings.TrimSpace(maintenanceID), strings.TrimSpace(vehicleID))
	if err != nil {
		return fmt.Errorf("delete vehicle maintenance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle maintenance delete result: %w", err)
	}
	if affected == 0 {
		return domain.ErrVehicleNotFound
	}
	return nil
}
func (s *Store) ListMaintenance(ctx context.Context, vehicleID string) ([]domain.VehicleMaintenance, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, kind, status, COALESCE(condition_rating, ''), COALESCE(due_date, ''), COALESCE(completed_at, ''),
       COALESCE(cost, ''), COALESCE(notes, ''), created_at, updated_at
FROM vehicle_maintenance
WHERE vehicle_id=?
ORDER BY
  CASE WHEN status='erledigt' THEN 1 ELSE 0 END ASC,
  CASE WHEN due_date='' THEN 1 ELSE 0 END ASC,
  due_date ASC,
  created_at DESC
`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle maintenance: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []domain.VehicleMaintenance{}
	for rows.Next() {
		var maintenance domain.VehicleMaintenance
		if err := rows.Scan(
			&maintenance.ID,
			&maintenance.VehicleID,
			&maintenance.Kind,
			&maintenance.Status,
			&maintenance.ConditionRating,
			&maintenance.DueDate,
			&maintenance.CompletedAt,
			&maintenance.Cost,
			&maintenance.Notes,
			&maintenance.CreatedAt,
			&maintenance.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle maintenance: %w", err)
		}
		entries = append(entries, maintenance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle maintenance: %w", err)
	}
	return entries, nil
}
