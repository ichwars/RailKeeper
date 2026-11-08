package workflow

import (
	"context"
	"fmt"
	"time"

	"railkeeper/backend/internal/decoder"
)

type ECoSStoredCVSource struct {
	Host, ObjectID, ReadAt, Protocol string
	Values                           []decoder.VehicleCVValueInput
}
type ECoSCVImportResult struct {
	Imported  int `json:"imported"`
	Preserved int `json:"preserved"`
}

// ImportECoSStoredCVs adopts a validated local reading; it never reads or writes hardware.
func (s *CVService) ImportECoSStoredCVs(
	ctx context.Context, vehicleID string, source ECoSStoredCVSource,
) (ECoSCVImportResult, error) {
	result := ECoSCVImportResult{}

	tx, err := s.repository.BeginCVTransaction(ctx)
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback() }()
	matches, err := tx.ECoSMappingCount(ctx, vehicleID, source.ObjectID)
	if err != nil {
		return result, err
	}
	if matches != 1 {
		return result, ErrStoredCVMappingConflict
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, value := range source.Values {
		exists, err := tx.CVAddressExists(ctx, vehicleID, value.CVNumber, value.IndexCV31, value.IndexCV32)
		if err != nil {
			return ECoSCVImportResult{}, err
		}
		if exists {
			result.Preserved++
			continue
		}
		if !decoder.ValidCVIndex(value) {
			return ECoSCVImportResult{}, decoder.ErrCVValidation
		}
		index := "?"
		if value.IndexCV31 != nil {
			index = fmt.Sprintf("%d/%d", *value.IndexCV31, *value.IndexCV32)
		}
		input := decoder.VehicleCVValueInput{
			CVNumber: value.CVNumber, Value: value.Value, Protocol: source.Protocol, DecoderProfile: "ECoS",
			IndexCV31: value.IndexCV31, IndexCV32: value.IndexCV32,
			Category: "ECoS",
			Description: fmt.Sprintf("ECoS @ %s · #%s · CV%d · %s · Cache/Speicher · CV31/CV32: %s",
				source.Host, source.ObjectID, value.CVNumber, source.ReadAt, index),
		}
		if !decoder.ValidCVValueInput(input) {
			return ECoSCVImportResult{}, decoder.ErrCVValidation
		}
		if err := importCVValue(ctx, tx, vehicleID, input, now); err != nil {
			return ECoSCVImportResult{}, err
		}
		result.Imported++
	}
	if err := tx.Commit(); err != nil {
		return ECoSCVImportResult{}, err
	}
	return result, nil
}
