package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

const MaxVehicleCVChanges = 500

var ErrVehicleCVConflict = errors.New("vehicle CV changed or target identity already exists")

type VehicleCVChange struct {
	ID               string               `json:"id,omitempty"`
	ExpectedRevision string               `json:"expectedRevision"`
	Delete           bool                 `json:"delete,omitempty"`
	Value            *VehicleCVValueInput `json:"value,omitempty"`
}

type VehicleCVChangesInput struct {
	Changes []VehicleCVChange `json:"changes"`
}

func cvRecordRevision(item vehicledata.CVRecord) string {
	data, _ := json.Marshal(item)
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func cvRecordValue(item vehicledata.CVRecord) VehicleCVValue {
	return VehicleCVValue{ID: item.ID, VehicleID: item.VehicleID, CVNumber: item.CVNumber,
		IndexCV31: item.IndexCV31, IndexCV32: item.IndexCV32, Value: item.Value, Description: item.Description,
		Category: item.Category, Protocol: item.Protocol, DecoderProfile: item.DecoderProfile,
		SourceFileID: item.SourceFileID, SourceLine: item.SourceLine, SourceExportedAt: item.SourceExportedAt,
		CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt, Revision: cvRecordRevision(item)}
}

func cvRecordInput(vehicleID string, input VehicleCVValueInput) vehicledata.CVRecord {
	return vehicledata.CVRecord{VehicleID: vehicleID, CVNumber: input.CVNumber, IndexCV31: input.IndexCV31,
		IndexCV32: input.IndexCV32, Value: input.Value, Description: input.Description, Category: input.Category,
		Protocol: input.Protocol, DecoderProfile: input.DecoderProfile, SourceFileID: input.SourceFileID,
		SourceLine: input.SourceLine, SourceExportedAt: input.SourceExportedAt}
}

func validCVRevision(value string) bool {
	data, err := hex.DecodeString(value)
	return err == nil && len(data) == sha256.Size && strings.ToLower(value) == value
}

func checkCVRevision(existing *vehicledata.CVRecord, expected *string) error {
	if expected == nil {
		return nil
	} // Existing clients keep their legacy request contract.
	if *expected != "" && !validCVRevision(*expected) {
		return ErrVehicleValidation
	}
	if existing == nil {
		if *expected != "" {
			return ErrVehicleCVConflict
		}
	} else if cvRecordRevision(*existing) != *expected {
		return ErrVehicleCVConflict
	}
	return nil
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
		if change.ID != "" {
			seenIDs[change.ID] = true
		}
		if change.Value != nil {
			value := cleanVehicleCVValueInput(*change.Value)
			key := vehicleCVRequestIdentity(value)
			if !isValidVehicleCVValueInput(value) || seenKeys[key] || value.ExpectedRevision != nil {
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
			if err := vehicledata.DeleteCV(ctx, tx, vehicleID, change.ID); err != nil {
				return 0, fmt.Errorf("delete draft CV: %w", err)
			}
			if err := touchDecoderCVWorkspaceTx(ctx, tx, vehicleID, previous.DecoderProfile, now); err != nil {
				return 0, err
			}
		} else {
			value := *change.Value
			value.ExpectedRevision = &change.ExpectedRevision
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

func vehicleCVIdentity(value VehicleCVValueInput) string {
	return fmt.Sprintf("%s:%d:%d:%d", value.DecoderProfile, value.CVNumber,
		cvIndexKey(value.IndexCV31), cvIndexKey(value.IndexCV32))
}
