package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type FunctionWrite struct {
	Function       VehicleFunction
	ConfirmedFlags bool
}

func PlanFunctionWrite(vehicleID, key string, input VehicleFunctionInput, now string) FunctionWrite {
	// Requests from the old, mislabelled UI cannot establish either meaning.
	confirmedFlags := input.FlagsVersion == 2
	if !confirmedFlags {
		if input.DirectionDependent != nil {
			input.LegacyDirectionDependent = input.DirectionDependent
		}
		input.DirectionDependent, input.Inverted = nil, nil
	}
	return FunctionWrite{ConfirmedFlags: confirmedFlags, Function: VehicleFunction{
		ID: vehicleID + ":" + key, VehicleID: vehicleID, FunctionKey: key,
		Name: input.Name, SymbolKey: input.SymbolKey, FunctionType: input.FunctionType, Mode: input.Mode,
		DirectionDependent: input.DirectionDependent, Inverted: input.Inverted,
		LegacyDirectionDependent: input.LegacyDirectionDependent, Notes: input.Notes,
		SortOrder: FunctionSortOrder(key), CreatedAt: now, UpdatedAt: now,
	}}
}
func (s *FunctionService) ListFunctions(ctx context.Context, vehicleID string) ([]VehicleFunction, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.repository.Functions(ctx, vehicleID)
}
func (s *FunctionService) UpsertFunction(ctx context.Context, vehicleID, key string, input VehicleFunctionInput) (*VehicleFunction, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	key = NormalizeFunctionKey(key)
	input = CleanVehicleFunctionInput(input)
	if vehicleID == "" || !ValidFunctionKey(key) || !IsValidVehicleFunctionInput(input) {
		return nil, ErrVehicleValidation
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	if err := s.repository.StoreFunction(ctx, PlanFunctionWrite(vehicleID, key, input, time.Now().UTC().Format(time.RFC3339))); err != nil {
		return nil, err
	}
	return s.GetFunction(ctx, vehicleID, key)
}
func (s *FunctionService) GetFunction(ctx context.Context, vehicleID, key string) (*VehicleFunction, error) {
	item, exists, err := s.repository.FindFunction(ctx, strings.TrimSpace(vehicleID), NormalizeFunctionKey(key))
	if err != nil {
		return nil, fmt.Errorf("get vehicle function: %w", err)
	}
	if !exists {
		return nil, ErrVehicleNotFound
	}
	return &item, nil
}
func (s *FunctionService) DeleteFunction(ctx context.Context, vehicleID, key string) (*VehicleFunction, error) {
	function, err := s.GetFunction(ctx, vehicleID, key)
	if err != nil {
		return nil, err
	}
	if err := s.repository.DeleteFunction(ctx, strings.TrimSpace(vehicleID), NormalizeFunctionKey(key)); err != nil {
		return nil, err
	}
	return function, nil
}
func ApplyFunctionChanges(ctx context.Context, mutation FunctionMutation, vehicleID string, changes []VehicleFunctionChange, now string) error {
	if err := ValidateFunctionChanges(changes); err != nil {
		return err
	}
	for _, change := range changes {
		key := NormalizeFunctionKey(change.FunctionKey)
		current, exists, err := mutation.FindFunction(ctx, vehicleID, key)
		if err != nil {
			return fmt.Errorf("read function revision: %w", err)
		}
		if (exists && current.Revision != change.ExpectedRevision) || (!exists && change.ExpectedRevision != "") {
			return ErrVehicleFunctionConflict
		}
		if change.Delete {
			if err := mutation.DeleteDraftFunction(ctx, vehicleID, key); err != nil {
				return err
			}
		} else if err := mutation.StoreFunction(ctx, PlanFunctionWrite(vehicleID, key, CleanVehicleFunctionInput(*change.Value), now)); err != nil {
			return err
		}
	}
	return nil
}
