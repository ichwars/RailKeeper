package application

import (
	"context"

	"railkeeper/backend/internal/infrastructure/catalogdata"
	"railkeeper/backend/internal/inventory/catalog"
)

func (s *VehicleService) CatalogContext(ctx context.Context, kind, id string) (*VehicleCatalogContext, error) {
	return catalog.New(catalogdata.New(s.db)).CatalogContext(ctx, kind, id)
}

type VehicleCatalogRevision = catalog.VehicleCatalogRevision

type VehicleCatalogContext = catalog.VehicleCatalogContext

var (
	ErrVehicleCatalogValidation = catalog.ErrVehicleCatalogValidation
	ErrVehicleCatalogNotFound   = catalog.ErrVehicleCatalogNotFound
	ErrVehicleCatalogConflict   = catalog.ErrVehicleCatalogConflict
)
