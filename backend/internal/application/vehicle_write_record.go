package application

import (
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func vehicleWriteRecord(vehicle Vehicle) vehicledata.Record { return vehiclemodel.WriteRecord(vehicle) }
