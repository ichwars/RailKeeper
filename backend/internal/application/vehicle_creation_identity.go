package application

import vehiclemodel "railkeeper/backend/internal/inventory/vehicle"

var (
	ErrVehicleCreationConflict = vehiclemodel.ErrVehicleCreationConflict
	ErrVehicleCreationIdentity = vehiclemodel.ErrVehicleCreationIdentity
)

type VehicleCreationOptions = vehiclemodel.CreationOptions
