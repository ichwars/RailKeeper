package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

type ECoSCVReadState struct {
	Source      string `json:"source,omitempty"`
	NextCV      int    `json:"nextCV"`
	Unavailable int    `json:"unavailable"`
	ReadAt      string `json:"readAt"`
	SourceHost  string `json:"sourceHost"`
}

type ECoSCVReadInput struct {
	Source  string `json:"source,omitempty"`
	StartCV int    `json:"startCV"`
	Restart bool   `json:"restart,omitempty"`
}

type digitalCenterStoredCVReader interface {
	ReadStoredCVs(context.Context, ECoSConnectionInput, int, int) (ECoSStoredCVBatch, error)
}

type digitalCenterStoredCVImporter interface {
	ImportECoSStoredCVs(context.Context, string, ECoSStoredCVSource) (ECoSCVImportResult, error)
}

// A small batch per request keeps cancellation and progress visible without a background job.
func (service *DigitalCenterWorkspaceService) ReadStoredCVBatch(
	ctx context.Context, sessionID, itemID string, input ECoSCVReadInput,
) (item DigitalCenterWorkItem, err error) {
	if input.Source == "database" {
		if input.StartCV != 1 {
			return item, ErrDigitalCenterFilterValidation
		}
		return service.readStoredCVDatabase(ctx, sessionID, itemID, input.Restart)
	}
	if input.Source != "" {
		return item, ErrDigitalCenterFilterValidation
	}
	startCV := input.StartCV
	err = service.runECoSOperation(ctx, func() error {
		if !validCVNumber(startCV) || (startCV-1)%ECoSStoredCVBatchSize != 0 || (input.Restart && startCV != 1) {
			return ErrDigitalCenterFilterValidation
		}
		var session DigitalCenterReadSession
		session, item, err = service.ecosCVWorkItem(ctx, sessionID, itemID)
		if err != nil {
			return err
		}
		values, state, err := storedECoSCVSnapshot(item)
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
		reader, ok := service.ecos.(digitalCenterStoredCVReader)
		if !ok {
			return ErrDigitalCenterCapabilityUnavailable
		}
		if state.Source == "database" && !input.Restart {
			return ErrDigitalCenterReadNotFresh
		}
		if !input.Restart && startCV < state.NextCV {
			// Includes CV1: a lost first response must not reset another client's
			// progress or depend on the device remaining reachable.
			return nil
		}
		if startCV == 1 {
			values = []ECoSCVValue{}
			state = ECoSCVReadState{NextCV: 1, SourceHost: session.Host,
				ReadAt: service.now().UTC().Format(time.RFC3339)}
		} else if state.NextCV != startCV {
			return ErrDigitalCenterReadNotFresh
		}
		objectID, err := strconv.Atoi(item.CenterObjectID)
		if err != nil {
			return ErrDigitalCenterDeviceOutput
		}
		batch, err := reader.ReadStoredCVs(ctx, ECoSConnectionInput{Host: center.Host, Port: center.Port}, objectID, startCV)
		if err != nil {
			return err
		}
		end := min(startCV+ECoSStoredCVBatchSize, ECoSStoredCVLimit+1)
		if batch.Unavailable < 0 || len(batch.Values)+batch.Unavailable != end-startCV {
			return ErrDigitalCenterDeviceOutput
		}
		seen := map[int]bool{}
		for _, value := range batch.Values {
			if value.Number < startCV || value.Number >= end || !validCVValue(value.Value) || seen[value.Number] ||
				value.IndexCV31 != nil || value.IndexCV32 != nil {
				return ErrDigitalCenterDeviceOutput
			}
			seen[value.Number] = true
		}
		values = append(values, batch.Values...)
		sortECoSCVValues(values)
		state.NextCV = end
		state.Unavailable += batch.Unavailable
		item.Center["cvs"], item.Center["cvRead"] = values, state
		item, err = service.repository.UpdateWorkItem(ctx, item)
		return err
	})
	return item, err
}

func (service *DigitalCenterWorkspaceService) ImportStoredCVs(
	ctx context.Context, sessionID, itemID, vehicleID string,
) (result ECoSCVImportResult, err error) {
	err = service.runECoSOperation(ctx, func() error {
		session, item, err := service.ecosCVWorkItem(ctx, sessionID, itemID)
		if err != nil {
			return err
		}
		values, state, err := storedECoSCVSnapshot(item)
		if err != nil {
			return err
		}
		if state.Source == "backup" && state.SourceHost != session.Host {
			return ErrDigitalCenterDeviceOutput
		}
		if state.Source == "backup" && (item.VehicleID == "" || item.VehicleID != vehicleID) {
			return ErrDigitalCenterConflictUnresolved
		}
		if len(values) == 0 {
			return ErrDigitalCenterFilterValidation
		}
		if state.NextCV != ECoSStoredCVLimit+1 {
			return ErrDigitalCenterReadNotFresh
		}
		importer, ok := service.vehicles.(digitalCenterStoredCVImporter)
		if !ok {
			return ErrDigitalCenterWorkspaceUnavailable
		}
		readAt := state.ReadAt
		if readAt == "" {
			readAt = session.ReadCompletedAt
		}
		result, err = importer.ImportECoSStoredCVs(ctx, vehicleID, ECoSStoredCVSource{
			Host: session.Host, ObjectID: item.CenterObjectID, ReadAt: readAt, Values: values,
			Protocol: item.Protocol,
		})
		return err
	})
	return result, err
}

func (service *DigitalCenterWorkspaceService) ecosCVWorkItem(
	ctx context.Context, sessionID, itemID string,
) (DigitalCenterReadSession, DigitalCenterWorkItem, error) {
	if service.repository == nil {
		return DigitalCenterReadSession{}, DigitalCenterWorkItem{}, ErrDigitalCenterWorkspaceUnavailable
	}
	session, err := service.repository.GetSession(ctx, sessionID)
	if err != nil {
		return session, DigitalCenterWorkItem{}, err
	}
	if session.Provider != "ecos" || session.State != DigitalCenterSessionReady {
		return session, DigitalCenterWorkItem{}, ErrDigitalCenterCapabilityUnavailable
	}
	item, err := service.repository.GetWorkItem(ctx, sessionID, itemID)
	if err == nil && item.StationStatus == "missing" {
		err = ErrDigitalCenterReadNotFresh
	}
	return session, item, err
}

func storedECoSCVSnapshot(item DigitalCenterWorkItem) ([]ECoSCVValue, ECoSCVReadState, error) {
	var snapshot struct {
		Values []ECoSCVValue   `json:"cvs"`
		State  ECoSCVReadState `json:"cvRead"`
	}
	data, err := json.Marshal(item.Center)
	if err == nil {
		err = json.Unmarshal(data, &snapshot)
	}
	if err != nil {
		return nil, snapshot.State, fmt.Errorf("%w: invalid CV snapshot", ErrDigitalCenterDeviceOutput)
	}
	values, err := validateStoredECoSCVs(snapshot.Values)
	if snapshot.State.NextCV != 0 || snapshot.State.Source != "" {
		state := snapshot.State
		_, timeErr := time.Parse(time.RFC3339, state.ReadAt)
		if state.Source == "database" || state.Source == "backup" {
			if state.NextCV != 1025 || state.Unavailable != 0 || timeErr != nil || strings.TrimSpace(state.SourceHost) == "" {
				return nil, state, ErrDigitalCenterDeviceOutput
			}
			return values, state, err
		}
		if state.NextCV < 33 || state.NextCV > ECoSStoredCVLimit+1 || (state.NextCV-1)%32 != 0 ||
			state.Unavailable < 0 || len(values)+state.Unavailable != state.NextCV-1 || timeErr != nil || state.SourceHost == "" || state.Source != "" {
			return nil, state, ErrDigitalCenterDeviceOutput
		}
	}
	return values, snapshot.State, err
}

func validateStoredECoSCVs(values []ECoSCVValue) ([]ECoSCVValue, error) {
	if len(values) > ecospkg.CVDatabaseLimit {
		return nil, ErrDigitalCenterDeviceOutput
	}
	seen := map[string]bool{}
	copyValues := make([]ECoSCVValue, 0, len(values))
	for _, value := range values {
		if !validCVNumber(value.Number) || !validCVValue(value.Value) ||
			(value.IndexCV31 == nil) != (value.IndexCV32 == nil) {
			return nil, ErrDigitalCenterDeviceOutput
		}
		if value.IndexCV31 != nil && (value.Number < 257 || value.Number > 512 || *value.IndexCV31 < 0 ||
			*value.IndexCV31 > 255 || *value.IndexCV32 < 0 || *value.IndexCV32 > 255) {
			return nil, ErrDigitalCenterDeviceOutput
		}
		key := fmt.Sprintf("%d:%d", eCoSCVPage(value), value.Number)
		if seen[key] {
			return nil, ErrDigitalCenterDeviceOutput
		}
		seen[key] = true
		copyValues = append(copyValues, value)
	}
	sortECoSCVValues(copyValues)
	return copyValues, nil
}

func eCoSCVPage(value ECoSCVValue) int {
	if value.IndexCV31 == nil || value.IndexCV32 == nil {
		return -1
	}
	return *value.IndexCV31*256 + *value.IndexCV32
}
