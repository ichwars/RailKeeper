package application

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type ecosFunctionReader interface {
	ReadFunctions(context.Context, ECoSConnectionInput, int) ([]ECoSFunctionValue, error)
	ReadLocomotive(context.Context, ECoSConnectionInput, int) (ECoSLocomotive, error)
}

func (service *DigitalCenterWorkspaceService) ReadVehicleECoSFunctions(
	ctx context.Context, vehicleID string,
) (snapshot ECoSFunctionSnapshot, err error) {
	err = service.runECoSOperation(ctx, func() error {
		center, err := service.configuredCenter(ctx, "ecos")
		if err != nil {
			return err
		}
		if !center.Active {
			return ErrDigitalCenterInactive
		}
		device, ok := service.ecos.(ecosFunctionReader)
		if !ok || service.vehicles == nil {
			return ErrDigitalCenterWorkspaceUnavailable
		}
		vehicles, err := service.vehicles.ListReadOnly(ctx, "")
		if err != nil {
			return err
		}
		var mapping *VehicleExternalMap
		found := false
		for _, vehicle := range vehicles {
			if vehicle.ID == vehicleID {
				found = true
				for _, item := range vehicle.ExternalMappings {
					if item.Provider == "ecos" {
						if mapping != nil {
							return ErrDigitalCenterConflictUnresolved
						}
						copy := item
						mapping = &copy
					}
				}
			}
		}
		if !found {
			return sql.ErrNoRows
		}
		if mapping == nil {
			return ErrDigitalCenterConflictUnresolved
		}
		id, err := strconv.Atoi(mapping.ExternalID)
		if err != nil || id < 1 || id > maxDigitalCenterObjectID {
			return ErrDigitalCenterConflictUnresolved
		}
		target := ECoSConnectionInput{Host: center.Host, Port: center.Port}
		before, err := device.ReadLocomotive(ctx, target, id)
		if err != nil {
			return err
		}
		protocol, protocolErr := normalizeDigitalCenterProtocol(before.Protocol)
		mappedProtocol, mappedProtocolErr := normalizeDigitalCenterProtocol(mapping.ExternalProtocol)
		address, addressErr := strconv.Atoi(mapping.ExternalAddress)
		// Workspace mappings store protocol families (DCC), while the device may
		// report a speed-step variant (DCC128). Compare the same representation.
		if before.ObjectID != id || strings.EqualFold(strings.TrimSpace(before.Protocol), "MULTI") ||
			protocolErr != nil || mappedProtocolErr != nil ||
			(mapping.ExternalAddress != "" && (addressErr != nil || address != before.Address)) ||
			(mapping.ExternalProtocol != "" && mappedProtocol != protocol) {
			return ErrDigitalCenterReadNotFresh
		}
		values, err := device.ReadFunctions(ctx, target, id)
		if err != nil {
			return err
		}
		after, err := device.ReadLocomotive(ctx, target, id)
		if err != nil {
			return err
		}
		if after.ObjectID != before.ObjectID || after.Address != before.Address ||
			after.Protocol != before.Protocol || after.Name != before.Name {
			return ErrDigitalCenterReadNotFresh
		}
		current, err := service.configuredCenter(ctx, "ecos")
		if err != nil {
			return err
		}
		if !current.Active || current.Host != center.Host || current.Port != center.Port {
			return ErrDigitalCenterReadNotFresh
		}
		snapshot = ECoSFunctionSnapshot{ObjectID: id, Host: center.Host,
			ReadAt: service.now().UTC().Format(time.RFC3339), Functions: values}
		return nil
	})
	return snapshot, err
}
