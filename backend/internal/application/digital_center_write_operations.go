package application

import (
	"context"
	"strconv"
	"strings"
)

type digitalCenterECoSCreator interface {
	CreateLocomotive(context.Context, ECoSLocomotiveCreateInput) (*ECoSLocomotiveCreateResult, error)
}

type digitalCenterVehicleMappingWriter interface {
	UpsertExternalMapping(context.Context, string, VehicleExternalMapInput, string) (*VehicleExternalMap, error)
}

type digitalCenterVehicleMappingRebinder interface {
	RebindExternalMapping(
		context.Context, string, string, VehicleExternalMapInput, string,
	) (*VehicleExternalMap, error)
}

func (service *DigitalCenterWorkspaceService) applyDigitalCenterWrite(
	ctx context.Context,
	target digitalCenterWriteTarget,
) (bool, int, error) {
	if target.operation == DigitalCenterWriteCreate {
		creator, ok := service.ecos.(digitalCenterECoSCreator)
		if !ok {
			return false, 0, ErrDigitalCenterWorkspaceUnavailable
		}
		created, err := creator.CreateLocomotive(ctx, ECoSLocomotiveCreateInput{
			Host: target.center.Host, Port: target.center.Port, Desired: target.desired, Confirm: true,
		})
		if err != nil {
			return false, 0, err
		}
		if created == nil || !created.Applied || created.ObjectID < 1 || created.ObjectID > maxDigitalCenterObjectID {
			return false, 0, ErrECoSWriteStateUnknown
		}
		return true, created.ObjectID, nil
	}
	writer, ok := service.ecos.(digitalCenterECoSWriter)
	if !ok {
		return false, 0, ErrDigitalCenterWorkspaceUnavailable
	}
	synced, err := writer.SyncLocomotive(ctx, ECoSLocomotiveSyncInput{
		Host: target.center.Host, Port: target.center.Port, ObjectID: target.objectID,
		Desired: target.desired, Confirm: true,
	})
	if err != nil {
		return false, 0, err
	}
	return synced != nil && synced.Applied, target.objectID, nil
}

func (service *DigitalCenterWorkspaceService) persistDigitalCenterMapping(
	ctx context.Context,
	target digitalCenterWriteTarget,
	verified ECoSLocomotive,
	actor string,
	syncStatus string,
) error {
	input := VehicleExternalMapInput{
		Provider: target.center.Provider, ExternalID: target.item.CenterObjectID,
		ExternalName: verified.Name, ExternalAddress: strconv.Itoa(verified.Address),
		ExternalProtocol: verified.Protocol, SyncStatus: syncStatus,
	}
	if target.operation == DigitalCenterWriteCreate {
		rebinder, ok := service.vehicles.(digitalCenterVehicleMappingRebinder)
		if !ok {
			return ErrDigitalCenterWorkspaceUnavailable
		}
		_, err := rebinder.RebindExternalMapping(ctx, target.item.VehicleID, target.previousObjectID,
			input, strings.TrimSpace(actor))
		return err
	}
	writer, ok := service.vehicles.(digitalCenterVehicleMappingWriter)
	if !ok {
		return ErrDigitalCenterWorkspaceUnavailable
	}
	_, err := writer.UpsertExternalMapping(ctx, target.item.VehicleID, input, strings.TrimSpace(actor))
	return err
}

func normalizeDigitalCenterWriteOperation(value DigitalCenterWriteOperation) (DigitalCenterWriteOperation, error) {
	switch value {
	case "", DigitalCenterWriteUpdate:
		return DigitalCenterWriteUpdate, nil
	case DigitalCenterWriteCreate:
		return DigitalCenterWriteCreate, nil
	default:
		return "", ErrDigitalCenterWriteFieldUnsupported
	}
}

func equalDigitalCenterWriteFields(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func digitalCenterWriteObjectID(target digitalCenterWriteTarget) string {
	if target.operation == DigitalCenterWriteCreate && target.objectID == 0 {
		return ""
	}
	if target.objectID > 0 {
		return strconv.Itoa(target.objectID)
	}
	return strings.TrimSpace(target.item.CenterObjectID)
}
