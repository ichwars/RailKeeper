package application

import (
	"context"
	"fmt"
)

type DigitalCenterAddressConflictError struct {
	ObjectID int
	Name     string
	Address  int
}

func (err *DigitalCenterAddressConflictError) Error() string {
	return "decoder address is already assigned to another ECoS object"
}

type digitalCenterECoSMasterReader interface {
	ListLocomotives(context.Context, ECoSConnectionInput) ([]ECoSLocomotive, error)
}

func (service *DigitalCenterWorkspaceService) checkDigitalCenterAddressConflict(
	ctx context.Context,
	target digitalCenterWriteTarget,
	changes []ECoSLocomotiveSyncChange,
) error {
	if !digitalCenterChangesField(changes, "address") {
		return nil
	}
	reader, ok := service.ecos.(digitalCenterECoSMasterReader)
	if !ok {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	locomotives, err := reader.ListLocomotives(ctx, ECoSConnectionInput{
		Host: target.center.Host,
		Port: target.center.Port,
	})
	if err != nil {
		return fmt.Errorf("check ECoS decoder address ownership: %w", err)
	}
	if len(locomotives) > maxDigitalCenterLocomotives {
		return ErrDigitalCenterDeviceOutput
	}
	targetFound := false
	for _, locomotive := range locomotives {
		if locomotive.ObjectID < 1 || locomotive.ObjectID > maxDigitalCenterObjectID ||
			locomotive.Address < 1 || locomotive.Address > maxDigitalCenterAddress {
			return ErrDigitalCenterDeviceOutput
		}
		if locomotive.ObjectID == target.objectID {
			targetFound = true
			continue
		}
		if locomotive.Address != target.desired.Address {
			continue
		}
		return &DigitalCenterAddressConflictError{
			ObjectID: locomotive.ObjectID,
			Name:     normalizeSafeConflictName(locomotive.Name),
			Address:  target.desired.Address,
		}
	}
	if !targetFound {
		return ErrDigitalCenterDeviceOutput
	}
	return nil
}

func digitalCenterChangesField(changes []ECoSLocomotiveSyncChange, field string) bool {
	for _, change := range changes {
		if change.Field == field {
			return true
		}
	}
	return false
}

func normalizeSafeConflictName(value string) string {
	name, err := normalizeDigitalCenterName(value)
	if err != nil || name == "" {
		return "ECoS-Lok"
	}
	return name
}
