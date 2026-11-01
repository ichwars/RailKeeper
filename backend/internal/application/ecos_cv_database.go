package application

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type ECoSCVDatabase struct {
	Locomotive ECoSLocomotive
	Values     []ECoSCVValue
	Sources    *ECoSCVSourceInfo
	Functions  *ECoSFunctionSnapshot
}

func (s *ECoSService) ReadStoredCVDatabase(ctx context.Context, input ECoSConnectionInput, objectID int) (ECoSCVDatabase, error) {
	result := ECoSCVDatabase{}
	target, err := normalizeECoSInput(input)
	if err != nil {
		return result, err
	}
	web, err := s.storedWeb(ctx, target.Host)
	if err != nil {
		return result, fmt.Errorf("%w: %v", ErrDigitalCenterDeviceOutput, err)
	}
	// Both transports use the same resolved address for the complete snapshot.
	target.Host = web.Address()
	before, err := s.ReadLocomotive(ctx, target, objectID)
	if err != nil {
		return result, err
	}
	backup := s.readStoredBackup(ctx, web)
	return s.readStoredCVSnapshot(ctx, target, web, before, backup)
}

type digitalCenterCVDatabaseReader interface {
	ReadStoredCVDatabase(context.Context, ECoSConnectionInput, int) (ECoSCVDatabase, error)
}

func (service *DigitalCenterWorkspaceService) readStoredCVDatabase(
	ctx context.Context, sessionID, itemID string, restart bool,
) (item DigitalCenterWorkItem, err error) {
	err = service.runECoSOperation(ctx, func() error {
		session, saved, err := service.ecosCVWorkItem(ctx, sessionID, itemID)
		item = saved
		if err != nil {
			return err
		}
		_, state, err := storedECoSCVSnapshot(item)
		if err != nil {
			return err
		}
		if state.Source == "backup" {
			return ErrDigitalCenterCapabilityUnavailable
		}
		center, err := service.configuredCenter(ctx, "ecos")
		if err != nil {
			return err
		}
		if !center.Active || center.Host != session.Host || center.Port != session.Port {
			return ErrDigitalCenterReadNotFresh
		}
		if !restart && state.Source == "database" {
			return nil // A retry after a lost response uses the committed snapshot.
		}
		reader, ok := service.ecos.(digitalCenterCVDatabaseReader)
		if !ok {
			return ErrDigitalCenterCapabilityUnavailable
		}
		objectID, err := strconv.Atoi(item.CenterObjectID)
		if err != nil || objectID < 1 || objectID > maxDigitalCenterObjectID {
			return ErrDigitalCenterDeviceOutput
		}
		snapshot, err := reader.ReadStoredCVDatabase(ctx, ECoSConnectionInput{Host: center.Host, Port: center.Port}, objectID)
		if err != nil {
			return err
		}
		protocol, err := normalizeDigitalCenterProtocol(snapshot.Locomotive.Protocol)
		name, nameErr := normalizeDigitalCenterName(snapshot.Locomotive.Name)
		if err != nil || nameErr != nil || snapshot.Locomotive.ObjectID != objectID || name != item.Name ||
			snapshot.Locomotive.Address != item.Address || protocol != item.Protocol {
			return ErrDigitalCenterReadNotFresh
		}
		values, err := validateStoredECoSCVs(snapshot.Values)
		if err != nil {
			return err
		}
		current, err := service.configuredCenter(ctx, "ecos")
		if err != nil {
			return err
		}
		if !current.Active || current.Host != center.Host || current.Port != center.Port {
			return ErrDigitalCenterReadNotFresh
		}
		item.Center["cvs"] = values
		item.Center["profile"] = snapshot.Locomotive.Profile
		if snapshot.Sources != nil {
			item.Center["cvSources"] = snapshot.Sources
			if len(snapshot.Sources.Conflicts) != 0 {
				item.Center["sourceProfile"] = snapshot.Locomotive.Profile
				item.Center["profile"] = ""
				if item.CompareStatus == DigitalCompareOK {
					item.CompareStatus = DigitalCompareDeviation
				}
			} else {
				delete(item.Center, "sourceProfile")
			}
		}
		if snapshot.Functions != nil {
			snapshot.Functions.Host = session.Host
			item.Center["functionSnapshot"] = snapshot.Functions
		}
		item.Center["cvRead"] = ECoSCVReadState{Source: "database", NextCV: ECoSStoredCVLimit + 1,
			ReadAt: service.now().UTC().Format(time.RFC3339), SourceHost: session.Host}
		item, err = service.repository.UpdateWorkItem(ctx, item)
		return err
	})
	return item, err
}
