package application

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ECoSStoredCVSource struct {
	Host     string
	ObjectID string
	ReadAt   string
	Protocol string
	Values   []ECoSCVValue
}

type ECoSCVImportResult struct {
	Imported  int `json:"imported"`
	Preserved int `json:"preserved"`
}

// Add memory values atomically. Existing unindexed CVs, under any profile,
// remain authoritative; indexed readings are distinct and are never rewritten.
func (s *VehicleService) ImportECoSStoredCVs(
	ctx context.Context, vehicleID string, source ECoSStoredCVSource,
) (ECoSCVImportResult, error) {
	result := ECoSCVImportResult{}
	values, err := validateStoredECoSCVs(source.Values)
	objectID, idErr := strconv.Atoi(source.ObjectID)
	_, timeErr := time.Parse(time.RFC3339, source.ReadAt)
	if err != nil || len(values) == 0 || idErr != nil || objectID < 1 ||
		objectID > maxDigitalCenterObjectID || timeErr != nil || strings.TrimSpace(source.Host) == "" {
		return result, ErrVehicleValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback() }()
	var matches int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_external_mappings
WHERE vehicle_id=? AND provider='ecos' AND external_id=?`, vehicleID, source.ObjectID).Scan(&matches)
	if err != nil {
		return result, err
	}
	if matches != 1 {
		return result, ErrDigitalCenterConflictUnresolved
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, value := range values {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM vehicle_cv_values
WHERE vehicle_id=? AND cv_number=? AND index_cv31 IS NULL AND index_cv32 IS NULL`,
			vehicleID, value.Number).Scan(&exists); err != nil {
			return ECoSCVImportResult{}, err
		}
		if exists > 0 {
			result.Preserved++
			continue
		}
		input := VehicleCVValueInput{
			CVNumber: value.Number, Value: value.Value, Protocol: source.Protocol, DecoderProfile: "ECoS",
			Category: "ECoS",
			Description: fmt.Sprintf("ECoS @ %s · #%s · CV%d · %s · Cache/Speicher · CV31/CV32: ?",
				source.Host, source.ObjectID, value.Number, source.ReadAt),
		}
		if !isValidVehicleCVValueInput(input) {
			return ECoSCVImportResult{}, ErrVehicleValidation
		}
		if err := importCVValueTx(ctx, tx, vehicleID, input, now); err != nil {
			return ECoSCVImportResult{}, err
		}
		result.Imported++
	}
	if err := tx.Commit(); err != nil {
		return ECoSCVImportResult{}, err
	}
	return result, nil
}
