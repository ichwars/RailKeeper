package application

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"railkeeper/backend/internal/infrastructure/vehicledata"
	"railkeeper/backend/internal/infrastructure/workshopdata"
	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
	"railkeeper/backend/internal/inventory/workshop"
)

func NewVehicleService(db *sql.DB) *VehicleService {
	rows := workshopdata.New(db)
	service := &VehicleService{db: db, workshopRows: rows}
	service.workshop = workshop.New(rows, func(ctx context.Context, id string) error { _, err := service.Get(ctx, id); return err })
	return service
}

func (s *VehicleService) SetImageLocalizer(localizer VehicleImageLocalizer) {
	s.imageLocalizer = localizer
}

func (s *VehicleService) List(ctx context.Context, query string) ([]Vehicle, error) {
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}
	return s.list(ctx, query, "")
}

func (s *VehicleService) ListReadOnly(ctx context.Context, query string) ([]Vehicle, error) {
	return s.list(ctx, query, "")
}

func (s *VehicleService) list(ctx context.Context, query string, vehicleSetID string) ([]Vehicle, error) {
	vehicles, err := vehicledata.ListCore(ctx, s.db, query, vehicleSetID)
	if err != nil {
		return nil, err
	}
	if err := s.attachImages(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachVehicleSetMainImages(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachAttachments(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachMaintenance(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachSpareParts(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachFunctions(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachCVData(ctx, vehicles); err != nil {
		return nil, err
	}
	if err := s.attachExternalMappings(ctx, vehicles); err != nil {
		return nil, err
	}
	return vehicles, nil
}

func (s *VehicleService) Get(ctx context.Context, id string) (*Vehicle, error) {
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}

	vehicle, err := s.get(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	vehicle.ExternalMappings, err = s.loadVehicleExternalMappings(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	vehicle.ECoSDecoderProfile, err = s.loadECoSDecoderProfile(ctx, vehicle.ID)
	if err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *VehicleService) resetExpiredExhibitionFlags(ctx context.Context) error {
	return vehiclemodel.ResetExpiredExhibitionFlags(ctx, vehicledata.NewExpiryStore(s.db), time.Now().UTC().Format(time.RFC3339))
}
