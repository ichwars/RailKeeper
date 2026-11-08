package application

import (
	"context"

	"railkeeper/backend/internal/infrastructure/catalogdata"
	"railkeeper/backend/internal/inventory/catalog"
)

func (s *VehicleService) CatalogBackfillReport(ctx context.Context) (*VehicleCatalogBackfillReport, error) {
	return catalog.New(catalogdata.New(s.db)).CatalogBackfillReport(ctx)
}

func (s *VehicleService) BackfillLegacyCatalog(ctx context.Context, actor string) (*VehicleCatalogBackfillReport, error) {
	return catalog.New(catalogdata.New(s.db)).BackfillLegacyCatalog(ctx, actor)
}

type VehicleCatalogAssignment = catalog.VehicleCatalogAssignment

type VehicleCatalogBackfillReport = catalog.VehicleCatalogBackfillReport
