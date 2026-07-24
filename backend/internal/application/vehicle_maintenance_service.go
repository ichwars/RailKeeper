package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *VehicleService) ListMaintenance(ctx context.Context, vehicleID string) ([]VehicleMaintenance, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleMaintenance(ctx, vehicleID)
}

func (s *VehicleService) CreateMaintenance(ctx context.Context, vehicleID string, input VehicleMaintenanceInput) (*VehicleMaintenance, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleMaintenanceInput(input)
	if vehicleID == "" || !isValidVehicleMaintenanceInput(input) {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	maintenance := VehicleMaintenance{
		ID:              randomID(),
		VehicleID:       vehicleID,
		Kind:            input.Kind,
		Status:          input.Status,
		ConditionRating: input.ConditionRating,
		DueDate:         input.DueDate,
		CompletedAt:     input.CompletedAt,
		Cost:            input.Cost,
		Notes:           input.Notes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_maintenance(id, vehicle_id, kind, status, condition_rating, due_date, completed_at, cost, notes, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, maintenance.ID, maintenance.VehicleID, maintenance.Kind, maintenance.Status, maintenance.ConditionRating, maintenance.DueDate, maintenance.CompletedAt, maintenance.Cost, maintenance.Notes, maintenance.CreatedAt, maintenance.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle maintenance: %w", err)
	}
	return &maintenance, nil
}

func (s *VehicleService) UpdateMaintenance(ctx context.Context, vehicleID, maintenanceID string, input VehicleMaintenanceInput) (*VehicleMaintenance, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	maintenanceID = strings.TrimSpace(maintenanceID)
	input = cleanVehicleMaintenanceInput(input)
	if vehicleID == "" || maintenanceID == "" || !isValidVehicleMaintenanceInput(input) {
		return nil, ErrVehicleValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `
UPDATE vehicle_maintenance
SET kind=?, status=?, condition_rating=?, due_date=?, completed_at=?, cost=?, notes=?, updated_at=?
WHERE id=? AND vehicle_id=?
`, input.Kind, input.Status, input.ConditionRating, input.DueDate, input.CompletedAt, input.Cost, input.Notes, now, maintenanceID, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("update vehicle maintenance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle maintenance update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return s.GetMaintenance(ctx, vehicleID, maintenanceID)
}

func (s *VehicleService) GetMaintenance(ctx context.Context, vehicleID, maintenanceID string) (*VehicleMaintenance, error) {
	var maintenance VehicleMaintenance
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
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle maintenance: %w", err)
	}
	return &maintenance, nil
}

func (s *VehicleService) DeleteMaintenance(ctx context.Context, vehicleID, maintenanceID string) (*VehicleMaintenance, error) {
	maintenance, err := s.GetMaintenance(ctx, vehicleID, maintenanceID)
	if err != nil {
		return nil, err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_maintenance WHERE id=? AND vehicle_id=?`, strings.TrimSpace(maintenanceID), strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("delete vehicle maintenance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle maintenance delete result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return maintenance, nil
}

func (s *VehicleService) loadVehicleMaintenance(ctx context.Context, vehicleID string) ([]VehicleMaintenance, error) {
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

	entries := []VehicleMaintenance{}
	for rows.Next() {
		var maintenance VehicleMaintenance
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
