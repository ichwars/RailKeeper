package vehicledata

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/infrastructure/numberingdata"
	"railkeeper/backend/internal/inventory/numbering"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type coreStore struct{ db *sql.DB }
type coreMutation struct {
	*imageMutation
	*functionMutation
	tx *sql.Tx
}

func NewCoreStore(db *sql.DB) vehiclemodel.CoreRepository        { return &coreStore{db: db} }
func VehicleMutationFor(tx *sql.Tx) vehiclemodel.VehicleMutation { return newCoreMutation(tx) }
func newCoreMutation(tx *sql.Tx) *coreMutation {
	return &coreMutation{imageMutation: &imageMutation{imageReader: &imageReader{query: tx}, tx: tx}, functionMutation: &functionMutation{tx: tx}, tx: tx}
}
func (s *coreStore) BeginVehicle(ctx context.Context) (vehiclemodel.VehicleMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return VehicleMutationFor(tx), nil
}
func (s *coreStore) CheckCreationIdentity(ctx context.Context, id string) error {
	return CheckCreationIdentity(ctx, s.db, id)
}
func (m *coreMutation) CheckCreationIdentity(ctx context.Context, id string) error {
	return CheckCreationIdentity(ctx, m.tx, id)
}
func (m *coreMutation) Counter() numbering.Counter { return numberingdata.NewCounter(m.tx) }
func (m *coreMutation) InventoryNumberInUse(ctx context.Context, number, excluded string) (bool, error) {
	return InventoryNumberInUse(ctx, m.tx, number, excluded)
}
func (m *coreMutation) InsertVehicle(ctx context.Context, record vehiclemodel.Record) error {
	return Insert(ctx, m.tx, record)
}
func (m *coreMutation) UpdateVehicle(ctx context.Context, record vehiclemodel.Record) (int64, error) {
	return Update(ctx, m.tx, record)
}
func (m *coreMutation) ReadCore(ctx context.Context, id string) (*vehiclemodel.Vehicle, error) {
	return ReadCore(ctx, m.tx, id)
}
func (m *coreMutation) WriteSourceMapping(ctx context.Context, write vehiclemodel.ExternalMappingWrite) (*vehiclemodel.VehicleExternalMap, error) {
	return WriteExternalMapping(ctx, m.tx, write)
}
func CheckCreationIdentity(ctx context.Context, reader queryRower, id string) error {
	var exists bool
	if err := reader.QueryRowContext(ctx, `SELECT EXISTS(
SELECT 1 FROM vehicles WHERE id=? UNION ALL SELECT 1 FROM vehicle_sets WHERE id=?)`, id, id).Scan(&exists); err != nil {
		return fmt.Errorf("check vehicle creation identity: %w", err)
	}
	if exists {
		return vehiclemodel.ErrVehicleCreationConflict
	}
	return nil
}

func (m *coreMutation) AuditCreated(ctx context.Context, audit vehiclemodel.VehicleAudit) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleCreated', 'vehicle', ?, ?, '{}')
`, audit.ID, audit.Actor, audit.VehicleID, audit.At); err != nil {
		return fmt.Errorf("write vehicle audit log: %w", err)
	}
	return nil
}
func (m *coreMutation) AuditUpdated(ctx context.Context, audit vehiclemodel.VehicleAudit) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleUpdated', 'vehicle', ?, ?, '{}')
`, audit.ID, audit.Actor, audit.VehicleID, audit.At); err != nil {
		return fmt.Errorf("write vehicle audit log: %w", err)
	}
	return nil
}
func (m *coreMutation) RecordNumberChange(ctx context.Context, change vehiclemodel.VehicleNumberChange) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO inventory_number_history(id, vehicle_id, old_number, new_number, changed_by_user_id, changed_at, reason)
VALUES(?, ?, ?, ?, ?, ?, 'manual_update')
`, change.ID, change.VehicleID, change.Previous, change.Current, change.Actor, change.At); err != nil {
		return fmt.Errorf("write inventory number history: %w", err)
	}
	return nil
}
