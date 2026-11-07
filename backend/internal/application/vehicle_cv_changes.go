package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

const MaxVehicleCVChanges = 500

var ErrVehicleCVConflict = decoder.ErrCVConflict

type VehicleCVChange = decoder.VehicleCVChange

type VehicleCVChangesInput = decoder.VehicleCVChangesInput

func cvRecordValue(item vehicledata.CVRecord) VehicleCVValue { return decoder.CVRecordValue(item) }

func cvRecordInput(vehicleID string, input VehicleCVValueInput) vehicledata.CVRecord {
	return decoder.CVRecordInput(vehicleID, input)
}

func validCVRevision(value string) bool { return decoder.ValidCVRevision(value) }

func checkCVRevision(existing *vehicledata.CVRecord, expected *string) error {
	return cvRuleError(decoder.CheckCVRevision(existing, expected))
}

func beginVehicleCVWrite(ctx context.Context, db *sql.DB, vehicleID string) (*sql.Tx, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin CV write: %w", err)
	}
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM vehicles WHERE id=?`, vehicleID).Scan(&exists); err != nil {
		_ = tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("check CV vehicle: %w", err)
	}
	return tx, nil
}

// All legacy and batch writes validate identity, source ownership and revision in the caller's transaction.
func writeVehicleCVTx(ctx context.Context, tx *sql.Tx, vehicleID string, previous *vehicledata.CVRecord,
	input VehicleCVValueInput, now string) (vehicledata.CVRecord, error) {
	resolved, err := resolveVehicleCVInputTx(ctx, tx, vehicleID, input)
	item := cvRecordInput(vehicleID, resolved)
	if err != nil {
		return item, err
	}
	if input.WorkspaceID != "" && input.ExpectedRevision == nil {
		return item, ErrVehicleValidation
	}
	if err := checkCVRevision(previous, input.ExpectedRevision); err != nil {
		return item, err
	}
	if previous == nil {
		if input.SourceWorkspaceID != "" || input.ExpectedSourceWorkspaceRevision != "" {
			return item, ErrVehicleValidation
		}
	} else {
		if input.WorkspaceID != "" && previous.DecoderProfile != item.DecoderProfile && input.SourceWorkspaceID == "" {
			return item, ErrVehicleValidation
		}
		if err := checkCVDeletionScopeTx(ctx, tx, vehicleID, previous.DecoderProfile, VehicleCVDeletionInput{
			ExpectedRevision: input.ExpectedRevision, WorkspaceID: input.SourceWorkspaceID, ExpectedWorkspaceRevision: input.ExpectedSourceWorkspaceRevision,
		}); err != nil {
			return item, err
		}
	}
	if item.SourceFileID != "" {
		var sourceVehicleID string
		err := tx.QueryRowContext(ctx, `SELECT vehicle_id FROM vehicle_cv_files WHERE id=?`, item.SourceFileID).Scan(&sourceVehicleID)
		// A removed original remains historical evidence on its existing CV, not a newly assigned source.
		retainedMissingSource := errors.Is(err, sql.ErrNoRows) && previous != nil && previous.SourceFileID == item.SourceFileID
		if !retainedMissingSource {
			if errors.Is(err, sql.ErrNoRows) || (err == nil && sourceVehicleID != vehicleID) {
				return item, ErrVehicleValidation
			}
			if err != nil {
				return item, fmt.Errorf("check CV source ownership: %w", err)
			}
		}
	}
	occupied, err := vehicledata.FindCV(ctx, tx, item)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return item, fmt.Errorf("check CV identity: %w", err)
	}
	if err == nil && (previous == nil || occupied.ID != previous.ID) {
		return item, ErrVehicleCVConflict
	}
	item.UpdatedAt = now
	if previous == nil {
		item.ID, item.CreatedAt = randomID(), now
		if err := vehicledata.InsertCV(ctx, tx, item); err != nil {
			return item, fmt.Errorf("create CV: %w", err)
		}
	} else {
		item.ID, item.CreatedAt = previous.ID, previous.CreatedAt
		if item.SourceLine == nil {
			item.SourceLine = previous.SourceLine
		}
		if item.SourceExportedAt == "" {
			item.SourceExportedAt = previous.SourceExportedAt
		}
		if err := vehicledata.UpdateCV(ctx, tx, item); err != nil {
			return item, fmt.Errorf("update CV: %w", err)
		}
		if previous.Value != item.Value {
			if err := vehicledata.InsertCVHistory(ctx, tx, randomID(), vehicleID, item.ID, previous.Value, item.Value, now); err != nil {
				return item, fmt.Errorf("write CV history: %w", err)
			}
		}
	}
	if previous != nil && previous.DecoderProfile != item.DecoderProfile {
		if err := touchDecoderCVWorkspaceTx(ctx, tx, vehicleID, previous.DecoderProfile, now); err != nil {
			return item, err
		}
	}
	if err := touchDecoderCVWorkspaceTx(ctx, tx, vehicleID, item.DecoderProfile, now); err != nil {
		return item, err
	}
	return item, nil
}

// ApplyCVChanges commits only the explicitly submitted manual edits, or none of them.
func (s *VehicleService) ApplyCVChanges(ctx context.Context, vehicleID string, input VehicleCVChangesInput) (int, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if len(input.Changes) == 0 || len(input.Changes) > MaxVehicleCVChanges {
		return 0, ErrVehicleValidation
	}
	seenIDs, seenKeys := map[string]bool{}, map[string]bool{}
	for _, change := range input.Changes {
		if change.ID != strings.TrimSpace(change.ID) || len(change.ID) > 80 || change.Delete == (change.Value != nil) ||
			(change.ID != "" && (seenIDs[change.ID] || !validCVRevision(change.ExpectedRevision))) ||
			(change.ID == "" && (change.Delete || change.ExpectedRevision != "")) {
			return 0, ErrVehicleValidation
		}
		if change.ID == "" && (change.WorkspaceID != "" || change.ExpectedWorkspaceRevision != "") {
			return 0, ErrVehicleValidation
		}
		if change.ID != "" {
			seenIDs[change.ID] = true
		}
		if change.Value != nil {
			value := cleanVehicleCVValueInput(*change.Value)
			key := vehicleCVRequestIdentity(value)
			if !isValidVehicleCVValueInput(value) || seenKeys[key] || value.ExpectedRevision != nil || value.SourceWorkspaceID != "" || value.ExpectedSourceWorkspaceRevision != "" {
				return 0, ErrVehicleValidation
			}
			seenKeys[key] = true
		}
	}
	tx, err := beginVehicleCVWrite(ctx, s.db, vehicleID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	seenTargets := map[string]bool{}
	for _, change := range input.Changes {
		if change.Value == nil {
			continue
		}
		value, err := resolveVehicleCVInputTx(ctx, tx, vehicleID, *change.Value)
		if err != nil {
			return 0, err
		}
		key := vehicleCVIdentity(value)
		if seenTargets[key] {
			return 0, ErrVehicleValidation
		}
		seenTargets[key] = true
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, change := range input.Changes {
		var previous *vehicledata.CVRecord
		if change.ID != "" {
			item, err := vehicledata.ReadCV(ctx, tx, vehicleID, change.ID)
			if errors.Is(err, sql.ErrNoRows) {
				return 0, ErrVehicleCVConflict
			}
			if err != nil {
				return 0, fmt.Errorf("read draft CV: %w", err)
			}
			previous = &item
		}
		if err := checkCVRevision(previous, &change.ExpectedRevision); err != nil {
			return 0, err
		}
		if change.Delete {
			if err := checkCVDeletionScopeTx(ctx, tx, vehicleID, previous.DecoderProfile, VehicleCVDeletionInput{
				ExpectedRevision: &change.ExpectedRevision, WorkspaceID: change.WorkspaceID, ExpectedWorkspaceRevision: change.ExpectedWorkspaceRevision,
			}); err != nil {
				return 0, err
			}
			if err := vehicledata.DeleteCV(ctx, tx, vehicleID, change.ID); err != nil {
				return 0, fmt.Errorf("delete draft CV: %w", err)
			}
			if err := touchDecoderCVWorkspaceTx(ctx, tx, vehicleID, previous.DecoderProfile, now); err != nil {
				return 0, err
			}
		} else {
			value := *change.Value
			value.ExpectedRevision = &change.ExpectedRevision
			value.SourceWorkspaceID, value.ExpectedSourceWorkspaceRevision = change.WorkspaceID, change.ExpectedWorkspaceRevision
			if _, err := writeVehicleCVTx(ctx, tx, vehicleID, previous, value, now); err != nil {
				return 0, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit CV draft: %w", err)
	}
	return len(input.Changes), nil
}

func vehicleCVIdentity(value VehicleCVValueInput) string { return decoder.CVIdentity(value) }
