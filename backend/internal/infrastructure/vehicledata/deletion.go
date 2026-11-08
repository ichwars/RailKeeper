package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type deleteStore struct{ db *sql.DB }
type deleteMutation struct{ tx *sql.Tx }

func NewDeleteStore(db *sql.DB) vehiclemodel.DeleteRepository { return &deleteStore{db: db} }
func (s *deleteStore) BeginDelete(ctx context.Context) (vehiclemodel.DeleteMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &deleteMutation{tx: tx}, nil
}
func (m *deleteMutation) Commit() error   { return m.tx.Commit() }
func (m *deleteMutation) Rollback() error { return m.tx.Rollback() }
func (m *deleteMutation) Membership(ctx context.Context, id string) (vehiclemodel.SetMembership, error) {
	var vehicleSetID sql.NullString
	if err := m.tx.QueryRowContext(ctx, `
SELECT vehicle_set_id FROM vehicle_set_members WHERE vehicle_id=?
`, id).Scan(&vehicleSetID); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return vehiclemodel.SetMembership{}, fmt.Errorf("read vehicle set membership: %w", err)
	}

	return vehiclemodel.SetMembership{ID: vehicleSetID.String, HasSet: vehicleSetID.Valid}, nil
}
func (m *deleteMutation) DeleteVehicle(ctx context.Context, id string) (int64, error) {
	result, err := m.tx.ExecContext(ctx, `DELETE FROM vehicles WHERE id=?`, id)
	if err != nil {
		return 0, fmt.Errorf("delete vehicle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read delete result: %w", err)
	}
	return affected, nil
}
func (m *deleteMutation) RemoveEmptySet(ctx context.Context, id string) error {
	if _, err := m.tx.ExecContext(ctx, `
DELETE FROM vehicle_sets
WHERE id=? AND NOT EXISTS (SELECT 1 FROM vehicle_set_members WHERE vehicle_set_id=?)
`, id, id); err != nil {
		return fmt.Errorf("delete empty vehicle set: %w", err)
	}
	return nil
}
func (m *deleteMutation) AuditDeleted(ctx context.Context, audit vehiclemodel.VehicleAudit) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleDeleted', 'vehicle', ?, ?, '{}')
`, audit.ID, audit.Actor, audit.VehicleID, audit.At); err != nil {
		return fmt.Errorf("write vehicle audit log: %w", err)
	}

	return nil
}
