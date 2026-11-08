package vehicle

import (
	"context"

	"railkeeper/backend/internal/inventory/numbering"
)

type CoreRepository interface {
	CheckCreationIdentity(context.Context, string) error
	BeginVehicle(context.Context) (VehicleMutation, error)
}
type VehicleInsertionMutation interface {
	ImageReplacementMutation
	FunctionMutation
	Counter() numbering.Counter
	InventoryNumberInUse(context.Context, string, string) (bool, error)
	InsertVehicle(context.Context, Record) error
	AuditCreated(context.Context, VehicleAudit) error
}
type VehicleMutation interface {
	VehicleInsertionMutation
	CheckCreationIdentity(context.Context, string) error
	UpdateVehicle(context.Context, Record) (int64, error)
	ReadCore(context.Context, string) (*Vehicle, error)
	Images(context.Context, string) ([]VehicleImage, error)
	WriteSourceMapping(context.Context, ExternalMappingWrite) (*VehicleExternalMap, error)
	AuditUpdated(context.Context, VehicleAudit) error
	RecordNumberChange(context.Context, VehicleNumberChange) error
	Commit() error
	Rollback() error
}
type VehicleAudit struct{ ID, VehicleID, Actor, At string }
type VehicleNumberChange struct{ ID, VehicleID, Previous, Current, Actor, At string }

// CoreEnvironment composes file compensation and separately owned aggregate reads.
type CoreEnvironment interface {
	Existing(context.Context, string) (*Vehicle, error)
	LocalizeImages(context.Context, string, []VehicleImageInput) ([]VehicleImageInput, error)
	CleanupImages(context.Context, []VehicleImageInput)
	CompleteCreated(context.Context, *Vehicle) error
	CompleteUpdated(context.Context, *Vehicle) error
}
type CoreService struct {
	repository  CoreRepository
	environment CoreEnvironment
}

func NewCoreService(repository CoreRepository, environment CoreEnvironment) *CoreService {
	return &CoreService{repository: repository, environment: environment}
}
