package application

import (
	"context"
	"database/sql"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

var (
	ErrVehicleCreationConflict = vehiclemodel.ErrVehicleCreationConflict
	ErrVehicleCreationIdentity = vehiclemodel.ErrVehicleCreationIdentity
)

type VehicleCreationOptions = vehiclemodel.CreationOptions

func (s *VehicleService) creationIdentity(ctx context.Context, requested string) (string, error) {
	return vehiclemodel.ResolveCreationIdentity(ctx, vehicledata.NewCoreStore(s.db), requested)
}

type creationIdentityReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func checkCreationIdentity(ctx context.Context, reader creationIdentityReader, id string) error {
	return vehicledata.CheckCreationIdentity(ctx, reader, id)
}
