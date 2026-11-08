package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func CleanExternalMappingInput(input VehicleExternalMapInput) VehicleExternalMapInput {
	input.Provider = strings.ToLower(strings.TrimSpace(input.Provider))
	input.ExternalID = strings.TrimSpace(input.ExternalID)
	input.ExternalName = strings.TrimSpace(input.ExternalName)
	input.ExternalAddress = strings.TrimSpace(input.ExternalAddress)
	input.ExternalProtocol = strings.TrimSpace(input.ExternalProtocol)
	input.SyncStatus = strings.TrimSpace(input.SyncStatus)
	if input.SyncStatus == "" {
		input.SyncStatus = "linked"
	}
	return input
}

func (s *ExternalMappingService) UpsertExternalMapping(ctx context.Context, vehicleID string, input VehicleExternalMapInput, _ string) (*VehicleExternalMap, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = CleanExternalMappingInput(input)
	if vehicleID == "" || input.Provider == "" || input.ExternalID == "" {
		return nil, ErrVehicleValidation
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return s.repository.WriteMapping(ctx, NewExternalMappingWrite(vehicleID, input, now))
}

func (s *ExternalMappingService) RebindExternalMapping(ctx context.Context, vehicleID, previousExternalID string, input VehicleExternalMapInput, actor string) (*VehicleExternalMap, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	previousExternalID = strings.TrimSpace(previousExternalID)
	input = CleanExternalMappingInput(input)
	if previousExternalID == "" || previousExternalID == input.ExternalID {
		return s.UpsertExternalMapping(ctx, vehicleID, input, actor)
	}
	if vehicleID == "" || input.Provider == "" || input.ExternalID == "" {
		return nil, ErrVehicleValidation
	}
	previous, err := s.repository.Mapping(ctx, input.Provider, previousExternalID)
	if err != nil || previous.VehicleID != vehicleID {
		return nil, ErrVehicleExternalMappingConflict
	}
	existing, found, lookupErr := s.repository.FindMapping(ctx, input.Provider, input.ExternalID)
	if lookupErr != nil {
		return nil, fmt.Errorf("check replacement external mapping owner: %w", lookupErr)
	}
	if found {
		if existing.VehicleID != vehicleID {
			return nil, ErrVehicleExternalMappingConflict
		}
		now := time.Now().UTC().Format(time.RFC3339)
		if err := s.repository.ConsolidateMappings(ctx, vehicleID, previousExternalID, input, now); err != nil {
			return nil, err
		}
		return s.repository.Mapping(ctx, input.Provider, input.ExternalID)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if input.SyncStatus == "" {
		input.SyncStatus = "linked"
	}
	if err := s.repository.RebindMapping(ctx, vehicleID, previousExternalID, input, now); err != nil {
		return nil, err
	}
	return s.repository.Mapping(ctx, input.Provider, input.ExternalID)
}

func NewExternalMappingWrite(id string, input VehicleExternalMapInput, at string) ExternalMappingWrite {
	return ExternalMappingWrite{ID: credentials.RandomID(), VehicleID: id, Input: input, At: at}
}
