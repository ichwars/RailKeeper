package vehicle

import "context"

type DeleteRepository interface {
	BeginDelete(context.Context) (DeleteMutation, error)
}
type SetMembership struct {
	ID     string
	HasSet bool
}
type DeleteMutation interface {
	Membership(context.Context, string) (SetMembership, error)
	DeleteVehicle(context.Context, string) (int64, error)
	RemoveEmptySet(context.Context, string) error
	AuditDeleted(context.Context, VehicleAudit) error
	Commit() error
	Rollback() error
}
