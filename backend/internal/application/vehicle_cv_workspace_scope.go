package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type VehicleCVDeletionInput struct {
	ExpectedRevision          *string
	WorkspaceID               string
	ExpectedWorkspaceRevision string
}

func checkCVDeletionScopeTx(ctx context.Context, tx *sql.Tx, vehicleID, profile string, input VehicleCVDeletionInput) error {
	if input.WorkspaceID == "" && input.ExpectedWorkspaceRevision == "" {
		return nil
	}
	if input.WorkspaceID == "" || !validCVRevision(input.ExpectedWorkspaceRevision) || input.ExpectedRevision == nil {
		return ErrVehicleValidation
	}
	stored, revision, err := readCVWorkspaceBindingTx(ctx, tx, vehicleID, input.WorkspaceID)
	if err != nil {
		return err
	}
	if revision != input.ExpectedWorkspaceRevision {
		return ErrVehicleCVConflict
	}
	if stored != profile {
		return ErrVehicleValidation
	}
	return nil
}

func readCVWorkspaceBindingTx(ctx context.Context, tx *sql.Tx, vehicleID, workspaceID string) (string, string, error) {
	var profile, installation, decoder, state string
	err := tx.QueryRowContext(ctx, `SELECT w.profile_text,COALESCE(w.installation_id,''),COALESCE(i.decoder_id,''),COALESCE(i.state,'')
FROM decoder_cv_workspaces w LEFT JOIN decoder_installations i ON i.id=w.installation_id
WHERE w.id=? AND w.vehicle_id=?`, workspaceID, vehicleID).Scan(&profile, &installation, &decoder, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrVehicleValidation
	}
	if err != nil {
		return "", "", fmt.Errorf("read CV working group binding: %w", err)
	}
	revision := recordRevision(struct{ WorkspaceID, VehicleID, Profile, InstallationID, DecoderID, State string }{workspaceID, vehicleID, profile, installation, decoder, state})
	return profile, revision, nil
}

func resolveCVWorkspaceProfileTx(ctx context.Context, tx *sql.Tx, vehicleID, workspaceID, expected, profile string) (string, error) {
	profile = strings.TrimSpace(profile)
	if workspaceID == "" {
		if expected != "" {
			return "", ErrVehicleValidation
		}
		return profile, nil
	}
	if !validCVRevision(expected) {
		return "", ErrVehicleValidation
	}
	stored, revision, err := readCVWorkspaceBindingTx(ctx, tx, vehicleID, workspaceID)
	if err != nil {
		return "", err
	}
	if revision != expected {
		return "", ErrVehicleCVConflict
	}
	if strings.TrimSpace(stored) != profile {
		return "", ErrVehicleValidation
	}
	return stored, nil
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

// Scoped inputs are resolved exactly inside the transaction before checking mixed scoped/legacy targets.
func vehicleCVRequestIdentity(value VehicleCVValueInput) string {
	if value.WorkspaceID != "" {
		return fmt.Sprintf("workspace:%q:%d:%d:%d", value.WorkspaceID, value.CVNumber, cvIndexKey(value.IndexCV31), cvIndexKey(value.IndexCV32))
	}
	return "profile:" + vehicleCVIdentity(value)
}
