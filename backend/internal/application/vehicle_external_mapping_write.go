package application

import (
	"context"
	"database/sql"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type externalMappingWriter interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func writeVehicleExternalMapping(ctx context.Context, writer externalMappingWriter, id string, input VehicleExternalMapInput, now string) (*VehicleExternalMap, error) {
	return vehicledata.WriteExternalMapping(ctx, writer, vehiclemodel.NewExternalMappingWrite(id, input, now))
}
