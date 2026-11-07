package vehicledata

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder"
)

// All operations use the caller's transaction, including its revision check and audit.
func WriteDecoderInstance(ctx context.Context, tx *sql.Tx, item decoder.DecoderInstance, create bool) error {
	if create {
		_, err := tx.ExecContext(ctx, `INSERT INTO decoder_instances
(id,label,manufacturer,model,serial_number,notes,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`,
			item.ID, item.Label, item.Manufacturer, item.Model, item.SerialNumber, item.Notes, item.CreatedAt, item.UpdatedAt)
		return err
	}
	_, err := tx.ExecContext(ctx, `UPDATE decoder_instances SET label=?,manufacturer=?,model=?,serial_number=?,notes=?,updated_at=? WHERE id=?`,
		item.Label, item.Manufacturer, item.Model, item.SerialNumber, item.Notes, item.UpdatedAt, item.ID)
	return err
}

func InsertDecoderInstallation(ctx context.Context, tx *sql.Tx, item decoder.DecoderInstallation) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO decoder_installations
(id,decoder_id,vehicle_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,state,installed_at,removed_at,notes,recorded_at,updated_at)
VALUES(?,?,?,?,?,?,'installed',?,NULL,?,?,?)`, item.ID, item.DecoderID, item.VehicleID, item.VehicleReferenceID, item.VehicleName, item.VehicleInventoryNumber,
		item.InstalledAt, item.Notes, item.RecordedAt, item.UpdatedAt)
	return err
}

func RemoveDecoderInstallation(ctx context.Context, tx *sql.Tx, item decoder.DecoderInstallation) error {
	_, err := tx.ExecContext(ctx, `UPDATE decoder_installations SET state='removed',removed_at=?,notes=?,updated_at=? WHERE id=?`, item.RemovedAt, item.Notes, item.UpdatedAt, item.ID)
	return err
}

func CorrectDecoderInstallationDates(ctx context.Context, tx *sql.Tx, item decoder.DecoderInstallation) error {
	_, err := tx.ExecContext(ctx, `UPDATE decoder_installations SET installed_at=?,removed_at=?,updated_at=? WHERE id=?`, item.InstalledAt, item.RemovedAt, item.UpdatedAt, item.ID)
	return err
}

func DecoderWorkspaceExists(ctx context.Context, tx *sql.Tx, id, vehicleID, profile string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM decoder_cv_workspaces WHERE id=? OR (vehicle_id=? AND profile_text=?))`, id, vehicleID, profile).Scan(&exists)
	return exists, err
}

func InsertDecoderWorkspace(ctx context.Context, tx *sql.Tx, item decoder.DecoderCVWorkspace) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO decoder_cv_workspaces
(id,vehicle_id,profile_text,origin,installation_id,created_at,updated_at) VALUES(?,?,?,'manual',NULL,?,?)`, item.ID, item.VehicleID, item.Profile, item.CreatedAt, item.UpdatedAt)
	return err
}

func AssignDecoderWorkspace(ctx context.Context, tx *sql.Tx, item decoder.DecoderCVWorkspace) error {
	_, err := tx.ExecContext(ctx, `UPDATE decoder_cv_workspaces SET installation_id=?,updated_at=? WHERE id=?`, item.InstallationID, item.UpdatedAt, item.ID)
	return err
}

func DecoderSnapshotExists(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	var exists bool
	err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM decoder_cv_snapshots WHERE id=?)`, id).Scan(&exists)
	return exists, err
}
