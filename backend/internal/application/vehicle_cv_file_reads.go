package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

const vehicleCVFileSelect = vehicledata.CVFileSelect

func cvFileRevision(file VehicleCVFile, sourceReferences string) string {
	return decoder.CVFileRevision(file, sourceReferences)
}

func scanVehicleCVFile(row interface{ Scan(...any) error }) (*VehicleCVFile, error) {
	return vehicledata.ScanCVFile(row)
}

func readVehicleCVFile(ctx context.Context, query queryRower, vehicleID, fileID string) (*VehicleCVFile, error) {
	file, err := scanVehicleCVFile(query.QueryRowContext(ctx, vehicleCVFileSelect+`WHERE id=? AND vehicle_id=?`,
		strings.TrimSpace(fileID), strings.TrimSpace(vehicleID)))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVehicleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get vehicle CV file: %w", err)
	}
	return file, nil
}
