package workflow

import (
	"context"
	"fmt"
	"time"
)

func BackfillLegacySnapshots(ctx context.Context, repository LegacySnapshotRepository) error {
	tx, err := repository.BeginLegacy(ctx)
	if err != nil {
		return fmt.Errorf("begin legacy decoder adoption: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := AdoptLegacySnapshotsWithin(ctx, tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit legacy decoder adoption: %w", err)
	}
	return nil
}

// AdoptLegacySnapshotsWithin shares the restore/configuration owner's transaction.
func AdoptLegacySnapshotsWithin(ctx context.Context, mutation LegacySnapshotMutation) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := mutation.AdoptLegacyGroups(ctx, now); err != nil {
		return err
	}
	workspaces, err := mutation.PendingLegacyWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		if err := CaptureSnapshot(ctx, mutation, workspace, SnapshotCaptureInput{Kind: "legacy", RecordedAt: now,
			Note: "Übernommener Arbeitsbestand; kein vollständiger Auslesestand. / Adopted working state; not a complete readout.",
		}); err != nil {
			return err
		}
	}
	return nil
}
