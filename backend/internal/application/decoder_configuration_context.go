package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

var (
	ErrDecoderValidation = errors.New("invalid decoder configuration")
	ErrDecoderNotFound   = errors.New("decoder configuration not found")
	ErrDecoderConflict   = errors.New("decoder configuration comparison changed")
)

func (s *VehicleService) DecoderConfigurationContext(ctx context.Context, vehicleID string) (*DecoderConfigurationContext, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return readDecoderConfigurationContextTx(ctx, tx, strings.TrimSpace(vehicleID))
}

func readDecoderConfigurationContextTx(ctx context.Context, tx *sql.Tx, vehicleID string) (*DecoderConfigurationContext, error) {
	result := &DecoderConfigurationContext{VehicleID: vehicleID, Instances: []DecoderInstance{},
		Installations: []DecoderInstallation{}, Workspaces: []DecoderCVWorkspace{}, Snapshots: []DecoderCVSnapshotSummary{}}
	if err := tx.QueryRowContext(ctx, `SELECT name,inventory_number FROM vehicles WHERE id=?`, vehicleID).
		Scan(&result.VehicleName, &result.InventoryNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDecoderNotFound
		}
		return nil, err
	}
	if err := readDecoderRows(ctx, tx, `SELECT id,label,manufacturer,model,serial_number,notes,created_at,updated_at
FROM decoder_instances ORDER BY id`, nil, func(rows *sql.Rows) error {
		var item DecoderInstance
		if err := rows.Scan(&item.ID, &item.Label, &item.Manufacturer, &item.Model, &item.SerialNumber,
			&item.Notes, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		result.Instances = append(result.Instances, item)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := readDecoderRows(ctx, tx, `SELECT id,decoder_id,vehicle_id,vehicle_reference_id,vehicle_name,
vehicle_inventory_number,state,installed_at,removed_at,notes,recorded_at,updated_at
FROM decoder_installations ORDER BY id`, nil, func(rows *sql.Rows) error {
		var item DecoderInstallation
		if err := rows.Scan(&item.ID, &item.DecoderID, &item.VehicleID, &item.VehicleReferenceID, &item.VehicleName,
			&item.VehicleInventoryNumber, &item.State, &item.InstalledAt, &item.RemovedAt, &item.Notes,
			&item.RecordedAt, &item.UpdatedAt); err != nil {
			return err
		}
		result.Installations = append(result.Installations, item)
		return nil
	}); err != nil {
		return nil, err
	}
	if err := readDecoderRows(ctx, tx, `SELECT id,vehicle_id,profile_text,origin,installation_id,created_at,updated_at
FROM decoder_cv_workspaces WHERE vehicle_id=? ORDER BY id`, []any{vehicleID}, func(rows *sql.Rows) error {
		var item DecoderCVWorkspace
		if err := rows.Scan(&item.ID, &item.VehicleID, &item.Profile, &item.Origin, &item.InstallationID, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return err
		}
		result.Workspaces = append(result.Workspaces, item)
		return nil
	}); err != nil {
		return nil, err
	}
	for index := range result.Workspaces {
		if err := decoderWorkspaceContentTx(ctx, tx, &result.Workspaces[index]); err != nil {
			return nil, err
		}
		_, revision, err := readCVWorkspaceBindingTx(ctx, tx, vehicleID, result.Workspaces[index].ID)
		if err != nil {
			return nil, err
		}
		result.Workspaces[index].BindingRevision = revision
	}
	snapshots, err := readDecoderSnapshotSummaries(ctx, tx, vehicleID)
	if err != nil {
		return nil, err
	}
	result.Snapshots = snapshots
	result.Revision = recordRevision(result)
	return result, nil
}

func readDecoderRows(ctx context.Context, tx *sql.Tx, statement string, args []any, scan func(*sql.Rows) error) error {
	rows, err := tx.QueryContext(ctx, statement, args...)
	if err != nil {
		return fmt.Errorf("read decoder configuration: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		if err := scan(rows); err != nil {
			return fmt.Errorf("scan decoder configuration: %w", err)
		}
	}
	return rows.Err()
}

func decoderWorkspaceContentTx(ctx context.Context, tx *sql.Tx, workspace *DecoderCVWorkspace) error {
	values := []vehicledata.CVRecord{}
	histories := []VehicleCVValueHistory{}
	files := []struct {
		Value      VehicleCVFile
		Path, Blob string
	}{}
	if err := readDecoderRows(ctx, tx, vehicledata.CVSelect+` WHERE vehicle_id=? AND decoder_profile=? ORDER BY id`,
		[]any{workspace.VehicleID, workspace.Profile}, func(rows *sql.Rows) error {
			item, err := vehicledata.ScanCV(rows)
			if err == nil {
				values = append(values, item)
			}
			return err
		}); err != nil {
		return err
	}
	if err := readDecoderRows(ctx, tx, `SELECT h.id,h.cv_value_id,h.vehicle_id,h.old_value,h.new_value,h.changed_at
FROM vehicle_cv_value_history h JOIN vehicle_cv_values v ON v.id=h.cv_value_id
WHERE v.vehicle_id=? AND v.decoder_profile=? ORDER BY h.id`, []any{workspace.VehicleID, workspace.Profile}, func(rows *sql.Rows) error {
		var item VehicleCVValueHistory
		if err := rows.Scan(&item.ID, &item.CVValueID, &item.VehicleID, &item.OldValue, &item.NewValue, &item.ChangedAt); err != nil {
			return err
		}
		histories = append(histories, item)
		return nil
	}); err != nil {
		return err
	}
	if err := readDecoderRows(ctx, tx, vehicleCVFileSelect+` WHERE vehicle_id=? AND decoder_profile=? ORDER BY id`,
		[]any{workspace.VehicleID, workspace.Profile}, func(rows *sql.Rows) error {
			file, err := scanVehicleCVFile(rows)
			if err != nil {
				return err
			}
			files = append(files, struct {
				Value      VehicleCVFile
				Path, Blob string
			}{*file, file.StoragePath, file.BlobID})
			return nil
		}); err != nil {
		return err
	}
	workspace.ValueCount, workspace.SourceCount = len(values), len(files)
	workspace.ContentRevision = recordRevision(struct {
		Values    []vehicledata.CVRecord
		Histories []VehicleCVValueHistory
		Files     any
	}{values, histories, files})
	return nil
}

func readDecoderSnapshotSummaries(ctx context.Context, tx *sql.Tx, vehicleID string) ([]DecoderCVSnapshotSummary, error) {
	result := []DecoderCVSnapshotSummary{}
	err := readDecoderRows(ctx, tx, `SELECT id,workspace_id,vehicle_reference_id,vehicle_name,
vehicle_inventory_number,profile_text,installation_id,kind,scope,note,recorded_at,
json_array_length(payload_json,'$.values') FROM decoder_cv_snapshots
WHERE (?='' OR vehicle_reference_id=?) ORDER BY recorded_at DESC,id DESC`, []any{vehicleID, vehicleID}, func(rows *sql.Rows) error {
		var item DecoderCVSnapshotSummary
		if err := rows.Scan(&item.ID, &item.WorkspaceID, &item.VehicleReferenceID, &item.VehicleName,
			&item.VehicleInventoryNumber, &item.Profile, &item.InstallationID, &item.Kind, &item.Scope, &item.Note, &item.RecordedAt, &item.ValueCount); err != nil {
			return err
		}
		result = append(result, item)
		return nil
	})
	return result, err
}
