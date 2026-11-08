package application

import "railkeeper/backend/internal/inventory/catalog"

type VehicleCatalogFields = catalog.VehicleCatalogFields

func decodeVehicleCatalogFields(raw string, complete bool) (VehicleCatalogFields, error) {
	return catalog.DecodeFields(raw, complete)
}
