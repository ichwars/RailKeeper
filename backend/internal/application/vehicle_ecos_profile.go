package application

import (
	"context"

	"railkeeper/backend/internal/ecosmodel"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type ECoSDecoderProfileEvidence = vehiclemodel.ECoSDecoderProfileEvidence

func (s *VehicleService) loadECoSDecoderProfile(ctx context.Context, vehicleID string) (*ECoSDecoderProfileEvidence, error) {
	observation, err := vehicledata.ReadCachedECoSDecoderProfile(ctx, s.db, vehicleID)
	if err != nil || observation == nil {
		return nil, err
	}
	return ecosmodel.ValidatedDecoderProfile(*observation, maxDigitalCenterProtocol), nil
}
