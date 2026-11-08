package workshop

import (
	"context"

	"railkeeper/backend/internal/domain"
)

type VehicleMaintenance = domain.VehicleMaintenance
type VehicleMaintenanceInput = domain.VehicleMaintenanceInput
type VehicleSparePart = domain.VehicleSparePart
type VehicleSparePartInput = domain.VehicleSparePartInput

var ErrVehicleValidation = domain.ErrVehicleValidation
var ErrVehicleNotFound = domain.ErrVehicleNotFound

type VehicleLookup func(context.Context, string) error
type Repository interface {
	ListMaintenance(context.Context, string) ([]VehicleMaintenance, error)
	GetMaintenance(context.Context, string, string) (*VehicleMaintenance, error)
	CreateMaintenance(context.Context, VehicleMaintenance) error
	UpdateMaintenance(context.Context, string, string, VehicleMaintenanceInput, string) error
	DeleteMaintenance(context.Context, string, string) error
	ListSpareParts(context.Context, string) ([]VehicleSparePart, error)
	GetSparePart(context.Context, string, string) (*VehicleSparePart, error)
	CreateSparePart(context.Context, VehicleSparePart) error
	UpdateSparePart(context.Context, string, string, VehicleSparePartInput, string) error
	DeleteSparePart(context.Context, string, string) error
}
type Service struct {
	repository Repository
	parent     VehicleLookup
}

func New(repository Repository, parent VehicleLookup) *Service {
	return &Service{repository: repository, parent: parent}
}
