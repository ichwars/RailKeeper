package workflow

import (
	"context"

	"railkeeper/backend/internal/decoder"
)

func checkCVDeletionScope(ctx context.Context, tx CVMutation, vehicleID, profile string, input decoder.VehicleCVDeletionInput) error {
	if input.WorkspaceID == "" && input.ExpectedWorkspaceRevision == "" {
		return nil
	}
	if input.WorkspaceID == "" || !decoder.ValidCVRevision(input.ExpectedWorkspaceRevision) || input.ExpectedRevision == nil {
		return decoder.ErrCVValidation
	}
	binding, err := tx.CVBinding(ctx, vehicleID, input.WorkspaceID)
	if err != nil {
		return err
	}
	return decoder.CheckCVDeletionScope(vehicleID, profile, input, &binding)
}

func resolveCVWorkspaceProfile(ctx context.Context, tx CVMutation, vehicleID, workspaceID, expected, profile string) (string, error) {
	if err := decoder.ValidateCVScope(workspaceID, expected); err != nil {
		return "", err
	}
	var binding *decoder.CVBinding
	if workspaceID != "" {
		value, err := tx.CVBinding(ctx, vehicleID, workspaceID)
		if err != nil {
			return "", err
		}
		binding = &value
	}
	resolved, err := decoder.ResolveCVProfile(vehicleID, workspaceID, expected, profile, binding)
	return resolved, err
}

func resolveCVInput(ctx context.Context, tx CVMutation, vehicleID string, input decoder.VehicleCVValueInput) (decoder.VehicleCVValueInput, error) {
	input = decoder.CleanCVValueInput(input)
	if !decoder.ValidCVValueInput(input) {
		return input, decoder.ErrCVValidation
	}
	profile, err := resolveCVWorkspaceProfile(ctx, tx, vehicleID, input.WorkspaceID, input.ExpectedWorkspaceRevision, input.DecoderProfile)
	if err != nil {
		return input, err
	}
	input.DecoderProfile = profile
	return input, nil
}
