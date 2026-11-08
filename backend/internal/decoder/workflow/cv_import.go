package workflow

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/decoder"
)

const MaxVehicleCVImportValues = 10000

type VehicleCVImportInput struct {
	Values []decoder.VehicleCVValueInput `json:"values"`
}

func (s *CVService) ImportCVValues(ctx context.Context, vehicleID string, input VehicleCVImportInput) (int, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if len(input.Values) == 0 || len(input.Values) > MaxVehicleCVImportValues {
		return 0, decoder.ErrCVValidation
	}
	seen := map[string]bool{}
	values := make([]decoder.VehicleCVValueInput, len(input.Values))
	for i, value := range input.Values {
		value = decoder.CleanCVValueInput(value)
		key := decoder.CVRequestIdentity(value)
		if !decoder.ValidCVValueInput(value) || seen[key] {
			return 0, decoder.ErrCVValidation
		}
		seen[key], values[i] = true, value
	}
	tx, err := s.repository.BeginVehicleCVWrite(ctx, vehicleID)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()
	seenTargets := map[string]bool{}
	for index, value := range values {
		resolved, err := resolveCVInput(ctx, tx, vehicleID, value)
		if err != nil {
			return 0, err
		}
		key := decoder.CVIdentity(resolved)
		if seenTargets[key] {
			return 0, decoder.ErrCVValidation
		}
		seenTargets[key], values[index] = true, resolved
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, value := range values {
		if err := importCVValue(ctx, tx, vehicleID, value, now); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit cv import: %w", err)
	}
	return len(input.Values), nil
}

func importCVValue(ctx context.Context, tx CVMutation, vehicleID string, value decoder.VehicleCVValueInput, now string) error {
	previous, found, err := tx.FindCV(ctx, decoder.CVRecordInput(vehicleID, value))
	if err != nil {
		return fmt.Errorf("read imported CV: %w", err)
	}
	var existing *decoder.CVRecord
	if found {
		existing = &previous
	}
	_, err = writeCV(ctx, tx, vehicleID, existing, value, now)
	return err
}
