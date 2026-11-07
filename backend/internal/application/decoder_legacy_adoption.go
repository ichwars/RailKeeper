package application

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type decoderWorkspaceRecord struct {
	ID, VehicleID, Profile string
	InstallationID         sql.NullString
}

// BackfillLegacyDecoderCV records available legacy evidence, without identifying hardware.
func BackfillLegacyDecoderCV(ctx context.Context, db *sql.DB, sourceRoot string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin legacy decoder adoption: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := backfillLegacyDecoderCVTx(ctx, tx, sourceRoot); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit legacy decoder adoption: %w", err)
	}
	return nil
}

func backfillLegacyDecoderCVTx(ctx context.Context, tx *sql.Tx, sourceRoot string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	// Restores and compatible old writers can introduce groups after migration 0077.
	if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO decoder_cv_workspaces
(id, vehicle_id, profile_text, origin, installation_id, created_at, updated_at)
SELECT lower(hex(randomblob(16))), vehicle_id, profile_text, 'legacy', NULL, ?, ? FROM (
 SELECT vehicle_id, COALESCE(decoder_profile,'') AS profile_text FROM vehicle_cv_values
 UNION SELECT vehicle_id, COALESCE(decoder_profile,'') AS profile_text FROM vehicle_cv_files
)`, now, now); err != nil {
		return fmt.Errorf("adopt legacy decoder working groups: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `SELECT id, vehicle_id, profile_text, installation_id
FROM decoder_cv_workspaces w WHERE origin='legacy'
AND NOT EXISTS(SELECT 1 FROM decoder_cv_snapshots s WHERE s.workspace_id=w.id AND s.kind='legacy')
ORDER BY id`)
	if err != nil {
		return fmt.Errorf("list unadopted legacy decoder groups: %w", err)
	}
	defer func() { _ = rows.Close() }()
	workspaces := []decoderWorkspaceRecord{}
	for rows.Next() {
		var item decoderWorkspaceRecord
		if err := rows.Scan(&item.ID, &item.VehicleID, &item.Profile, &item.InstallationID); err != nil {
			return fmt.Errorf("scan legacy decoder group: %w", err)
		}
		workspaces = append(workspaces, item)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, workspace := range workspaces {
		if err := captureDecoderCVSnapshotTx(ctx, tx, workspace, decoderSnapshotInput{Kind: "legacy", RecordedAt: now,
			SourceRoot: sourceRoot, Note: "Übernommener Arbeitsbestand; kein vollständiger Auslesestand. / Adopted working state; not a complete readout."}); err != nil {
			return err
		}
	}
	return nil
}
