package vehicle

import "context"

type SetRepository interface {
	CheckCreationIdentity(context.Context, string) error
	BeginSet(context.Context) (SetMutation, error)
	Set(context.Context, string) (VehicleSet, error)
	MemberIDs(context.Context, string) ([]string, error)
	MemberImages(context.Context, string) ([]VehicleImage, error)
}
type SetMutation interface {
	VehicleInsertionMutation
	CheckCreationIdentity(context.Context, string) error
	SetInventoryNumberInUse(context.Context, string) (bool, error)
	InsertSet(context.Context, SetWrite) error
	InsertSetMember(context.Context, SetMemberWrite) error
	LinkCreatedSetMember(context.Context, string, string) error
	UpdateSet(context.Context, string, VehicleSetInput, string) (int64, error)
	AuditSetCreated(context.Context, SetAudit) error
	AuditSetUpdated(context.Context, SetAudit) error
	Commit() error
	Rollback() error
}
type SetWrite struct {
	ID, InventoryNumber string
	Input               VehicleSetInput
	At                  string
}
type SetMemberWrite struct {
	SetID, VehicleID, Label string
	Position                int
}
type SetAudit struct{ ID, SetID, Actor, At string }
type SetEnvironment interface {
	LocalizeImages(context.Context, string, []VehicleImageInput) ([]VehicleImageInput, error)
	CleanupImages(context.Context, []VehicleImageInput)
	LoadMembers(context.Context, string) ([]Vehicle, error)
	AttachMainImage(context.Context, *VehicleSet) error
}
type SetService struct {
	repository  SetRepository
	environment SetEnvironment
}

func NewSetService(repository SetRepository, environment SetEnvironment) *SetService {
	return &SetService{repository: repository, environment: environment}
}
