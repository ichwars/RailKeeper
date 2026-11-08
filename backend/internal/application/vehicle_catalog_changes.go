package application

import (
	"context"

	"railkeeper/backend/internal/infrastructure/catalogdata"
	"railkeeper/backend/internal/inventory/catalog"
)

func (s *VehicleService) CreateCatalogRevision(ctx context.Context, kind, id string, input VehicleCatalogRevisionInput, actor string) (*VehicleCatalogContext, error) {
	return catalog.New(catalogdata.New(s.db)).CreateCatalogRevision(ctx, kind, id, input, actor)
}

func (s *VehicleService) AdoptCatalogRevision(ctx context.Context, kind, id string, input VehicleCatalogAdoptionInput, actor string) (*VehicleCatalogContext, error) {
	return catalog.New(catalogdata.New(s.db)).AdoptCatalogRevision(ctx, kind, id, input, actor)
}

type VehicleCatalogRevisionInput = catalog.VehicleCatalogRevisionInput

type VehicleCatalogAdoptionInput = catalog.VehicleCatalogAdoptionInput
