package vehicledata

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/revision"
)

type decoderWorkspaceSource struct {
	Value      decoder.VehicleCVFile
	Path, Blob string
}

type decoderWorkspaceContent struct {
	Values    []CVRecord
	Histories []decoder.VehicleCVValueHistory
	Files     []decoderWorkspaceSource
}

func loadDecoderWorkspaceContent(ctx context.Context, tx *sql.Tx, vehicleID string, workspaces []decoder.DecoderCVWorkspace) error {
	if len(workspaces) == 0 {
		return nil
	}
	content := make(map[string]*decoderWorkspaceContent, len(workspaces))
	for _, workspace := range workspaces {
		content[workspace.Profile] = &decoderWorkspaceContent{
			Values: []CVRecord{}, Histories: []decoder.VehicleCVValueHistory{}, Files: []decoderWorkspaceSource{},
		}
	}

	const profiles = "(SELECT profile_text FROM decoder_cv_workspaces WHERE vehicle_id=?)"
	args := []any{vehicleID, vehicleID}
	if err := ReadDecoderRows(ctx, tx, CVSelect+` WHERE vehicle_id=? AND decoder_profile IN `+profiles+` ORDER BY id`,
		args, func(rows *sql.Rows) error {
			item, err := ScanCV(rows)
			if err == nil {
				group := content[item.DecoderProfile]
				group.Values = append(group.Values, item)
			}
			return err
		}); err != nil {
		return err
	}
	if err := ReadDecoderRows(ctx, tx, `
SELECT v.decoder_profile,h.id,h.cv_value_id,h.vehicle_id,h.old_value,h.new_value,h.changed_at
FROM vehicle_cv_value_history h JOIN vehicle_cv_values v ON v.id=h.cv_value_id
WHERE v.vehicle_id=? AND v.decoder_profile IN `+profiles+` ORDER BY h.id`, args, func(rows *sql.Rows) error {
		var profile string
		var item decoder.VehicleCVValueHistory
		if err := rows.Scan(&profile, &item.ID, &item.CVValueID, &item.VehicleID, &item.OldValue, &item.NewValue, &item.ChangedAt); err != nil {
			return err
		}
		group := content[profile]
		group.Histories = append(group.Histories, item)
		return nil
	}); err != nil {
		return err
	}
	if err := ReadDecoderRows(ctx, tx, CVFileSelect+` WHERE vehicle_id=? AND decoder_profile IN `+profiles+` ORDER BY id`,
		args, func(rows *sql.Rows) error {
			file, err := ScanCVFile(rows)
			if err == nil {
				group := content[file.DecoderProfile]
				group.Files = append(group.Files, decoderWorkspaceSource{*file, file.StoragePath, file.BlobID})
			}
			return err
		}); err != nil {
		return err
	}
	for i := range workspaces {
		workspace := &workspaces[i]
		group := content[workspace.Profile]
		workspace.ValueCount, workspace.SourceCount = len(group.Values), len(group.Files)
		workspace.ContentRevision = revision.Hash(group)
	}
	return nil
}
