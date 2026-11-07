package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

type VehicleCVDeletionInput = decoder.VehicleCVDeletionInput

func checkCVDeletionScopeTx(ctx context.Context, tx *sql.Tx, vehicleID, profile string, input VehicleCVDeletionInput) error {
	if input.WorkspaceID == "" && input.ExpectedWorkspaceRevision == "" {
		return nil
	}
	if input.WorkspaceID == "" || !validCVRevision(input.ExpectedWorkspaceRevision) || input.ExpectedRevision == nil {
		return ErrVehicleValidation
	}
	binding, err := readCVBindingTx(ctx, tx, vehicleID, input.WorkspaceID)
	if err != nil {
		return err
	}
	return cvRuleError(decoder.CheckCVDeletionScope(vehicleID, profile, input, &binding))
}

func readCVBindingTx(ctx context.Context, tx *sql.Tx, vehicleID, workspaceID string) (decoder.CVBinding, error) {
	binding, err := vehicledata.ReadCVWorkspaceBinding(ctx, tx, vehicleID, workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		return binding, ErrVehicleValidation
	}
	if err != nil {
		return binding, fmt.Errorf("read CV working group binding: %w", err)
	}
	return binding, nil
}

func resolveCVWorkspaceProfileTx(ctx context.Context, tx *sql.Tx, vehicleID, workspaceID, expected, profile string) (string, error) {
	if err := decoder.ValidateCVScope(workspaceID, expected); err != nil {
		return "", cvRuleError(err)
	}
	var binding *decoder.CVBinding
	if workspaceID != "" {
		value, err := readCVBindingTx(ctx, tx, vehicleID, workspaceID)
		if err != nil {
			return "", err
		}
		binding = &value
	}
	resolved, err := decoder.ResolveCVProfile(vehicleID, workspaceID, expected, profile, binding)
	return resolved, cvRuleError(err)
}

func resolveVehicleCVInputTx(ctx context.Context, tx *sql.Tx, vehicleID string, input VehicleCVValueInput) (VehicleCVValueInput, error) {
	input = cleanVehicleCVValueInput(input)
	if !isValidVehicleCVValueInput(input) {
		return input, ErrVehicleValidation
	}
	profile, err := resolveCVWorkspaceProfileTx(ctx, tx, vehicleID, input.WorkspaceID, input.ExpectedWorkspaceRevision, input.DecoderProfile)
	if err != nil {
		return input, err
	}
	input.DecoderProfile = profile
	return input, nil
}

func vehicleCVRequestIdentity(value VehicleCVValueInput) string {
	return decoder.CVRequestIdentity(value)
}
