package application

import (
	"context"
	"encoding/json"

	"railkeeper/backend/internal/infrastructure/vehicledata"
)

func (s *VehicleService) VehiclesForReport(ctx context.Context, ids []string) ([]Vehicle, error) {
	if len(ids) == 0 {
		return []Vehicle{}, nil
	}
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}
	encodedIDs, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	vehicles, err := vehicledata.ReadCoresByID(ctx, s.db, string(encodedIDs))
	if err != nil {
		return nil, err
	}
	if len(vehicles) != len(ids) {
		return nil, ErrVehicleNotFound
	}
	if err := s.attachVehicleRelations(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachVehicleSetMainImages(ctx, vehicles); err != nil {
		return nil, err
	}
	return vehicles, nil
}
