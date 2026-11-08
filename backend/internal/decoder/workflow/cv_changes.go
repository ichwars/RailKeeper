package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/decoder"
)

const MaxVehicleCVChanges = 500

func writeCV(ctx context.Context, tx CVMutation, vehicleID string, previous *decoder.CVRecord,
	input decoder.VehicleCVValueInput, now string) (decoder.CVRecord, error) {
	resolved, err := resolveCVInput(ctx, tx, vehicleID, input)
	item := decoder.CVRecordInput(vehicleID, resolved)
	if err != nil {
		return item, err
	}
	if input.WorkspaceID != "" && input.ExpectedRevision == nil {
		return item, decoder.ErrCVValidation
	}
	if err := decoder.CheckCVRevision(previous, input.ExpectedRevision); err != nil {
		return item, err
	}
	if previous == nil {
		if input.SourceWorkspaceID != "" || input.ExpectedSourceWorkspaceRevision != "" {
			return item, decoder.ErrCVValidation
		}
	} else {
		if input.WorkspaceID != "" && previous.DecoderProfile != item.DecoderProfile && input.SourceWorkspaceID == "" {
			return item, decoder.ErrCVValidation
		}
		if err := checkCVDeletionScope(ctx, tx, vehicleID, previous.DecoderProfile, decoder.VehicleCVDeletionInput{
			ExpectedRevision: input.ExpectedRevision, WorkspaceID: input.SourceWorkspaceID, ExpectedWorkspaceRevision: input.ExpectedSourceWorkspaceRevision,
		}); err != nil {
			return item, err
		}
	}
	if item.SourceFileID != "" {
		sourceVehicleID, found, err := tx.CVSourceVehicle(ctx, item.SourceFileID)
		// A removed original remains historical evidence on its existing CV, not a newly assigned source.
		retainedMissingSource := err == nil && !found && previous != nil && previous.SourceFileID == item.SourceFileID
		if !retainedMissingSource {
			if (err == nil && !found) || (err == nil && sourceVehicleID != vehicleID) {
				return item, decoder.ErrCVValidation
			}
			if err != nil {
				return item, fmt.Errorf("check CV source ownership: %w", err)
			}
		}
	}
	occupied, found, err := tx.FindCV(ctx, item)
	if err != nil {
		return item, fmt.Errorf("check CV identity: %w", err)
	}
	if found && (previous == nil || occupied.ID != previous.ID) {
		return item, decoder.ErrCVConflict
	}
	item.UpdatedAt = now
	if previous == nil {
		item.ID, item.CreatedAt = credentials.RandomID(), now
		if err := tx.InsertCV(ctx, item); err != nil {
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
		if err := tx.UpdateCV(ctx, item); err != nil {
			return item, fmt.Errorf("update CV: %w", err)
		}
		if previous.Value != item.Value {
			if err := tx.InsertCVHistory(ctx, decoder.VehicleCVValueHistory{ID: credentials.RandomID(), VehicleID: vehicleID, CVValueID: item.ID, OldValue: previous.Value, NewValue: item.Value, ChangedAt: now}); err != nil {
				return item, fmt.Errorf("write CV history: %w", err)
			}
		}
	}
	if previous != nil && previous.DecoderProfile != item.DecoderProfile {
		if err := TouchWorkspace(ctx, tx, vehicleID, previous.DecoderProfile, now); err != nil {
			return item, err
		}
	}
	if err := TouchWorkspace(ctx, tx, vehicleID, item.DecoderProfile, now); err != nil {
		return item, err
	}
	return item, nil
}

// ApplyCVChanges commits only the explicitly submitted manual edits, or none of them.
func (s *CVService) ApplyCVChanges(ctx context.Context, vehicleID string, input decoder.VehicleCVChangesInput) (int, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if len(input.Changes) == 0 || len(input.Changes) > MaxVehicleCVChanges {
		return 0, decoder.ErrCVValidation
	}
	seenIDs, seenKeys := map[string]bool{}, map[string]bool{}
	for _, change := range input.Changes {
		if change.ID != strings.TrimSpace(change.ID) || len(change.ID) > 80 || change.Delete == (change.Value != nil) ||
			(change.ID != "" && (seenIDs[change.ID] || !decoder.ValidCVRevision(change.ExpectedRevision))) ||
			(change.ID == "" && (change.Delete || change.ExpectedRevision != "")) {
			return 0, decoder.ErrCVValidation
		}
		if change.ID == "" && (change.WorkspaceID != "" || change.ExpectedWorkspaceRevision != "") {
			return 0, decoder.ErrCVValidation
		}
		if change.ID != "" {
			seenIDs[change.ID] = true
		}
		if change.Value != nil {
			value := decoder.CleanCVValueInput(*change.Value)
			key := decoder.CVRequestIdentity(value)
			if !decoder.ValidCVValueInput(value) || seenKeys[key] || value.ExpectedRevision != nil || value.SourceWorkspaceID != "" || value.ExpectedSourceWorkspaceRevision != "" {
				return 0, decoder.ErrCVValidation
			}
			seenKeys[key] = true
		}
	}
	tx, err := s.repository.BeginVehicleCVWrite(ctx, vehicleID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	seenTargets := map[string]bool{}
	for _, change := range input.Changes {
		if change.Value == nil {
			continue
		}
		value, err := resolveCVInput(ctx, tx, vehicleID, *change.Value)
		if err != nil {
			return 0, err
		}
		key := decoder.CVIdentity(value)
		if seenTargets[key] {
			return 0, decoder.ErrCVValidation
		}
		seenTargets[key] = true
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, change := range input.Changes {
		var previous *decoder.CVRecord
		if change.ID != "" {
			item, found, err := tx.CVValue(ctx, vehicleID, change.ID)
			if err == nil && !found {
				return 0, decoder.ErrCVConflict
			}
			if err != nil {
				return 0, fmt.Errorf("read draft CV: %w", err)
			}
			previous = &item
		}
		if err := decoder.CheckCVRevision(previous, &change.ExpectedRevision); err != nil {
			return 0, err
		}
		if change.Delete {
			if err := checkCVDeletionScope(ctx, tx, vehicleID, previous.DecoderProfile, decoder.VehicleCVDeletionInput{
				ExpectedRevision: &change.ExpectedRevision, WorkspaceID: change.WorkspaceID, ExpectedWorkspaceRevision: change.ExpectedWorkspaceRevision,
			}); err != nil {
				return 0, err
			}
			if err := tx.DeleteCV(ctx, vehicleID, change.ID); err != nil {
				return 0, fmt.Errorf("delete draft CV: %w", err)
			}
			if err := TouchWorkspace(ctx, tx, vehicleID, previous.DecoderProfile, now); err != nil {
				return 0, err
			}
		} else {
			value := *change.Value
			value.ExpectedRevision = &change.ExpectedRevision
			value.SourceWorkspaceID, value.ExpectedSourceWorkspaceRevision = change.WorkspaceID, change.ExpectedWorkspaceRevision
			if _, err := writeCV(ctx, tx, vehicleID, previous, value, now); err != nil {
				return 0, err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit CV draft: %w", err)
	}
	return len(input.Changes), nil
}
