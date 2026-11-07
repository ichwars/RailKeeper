package decoder

import (
	"strings"

	"railkeeper/backend/internal/revision"
)

// Field order is part of the existing comparison-token encoding.
type CVBinding struct {
	WorkspaceID, VehicleID, Profile, InstallationID, DecoderID, State string
}

func CVBindingRevision(binding CVBinding) string { return revision.Hash(binding) }

func ValidateCVScope(workspaceID, expected string) error {
	if workspaceID == "" {
		if expected != "" {
			return ErrCVValidation
		}
	} else if !ValidCVRevision(expected) {
		return ErrCVValidation
	}
	return nil
}

func ResolveCVProfile(vehicleID, workspaceID, expected, profile string, binding *CVBinding) (string, error) {
	profile = strings.TrimSpace(profile)
	if err := ValidateCVScope(workspaceID, expected); err != nil {
		return "", err
	}
	if workspaceID == "" {
		return profile, nil
	}
	if binding == nil || binding.WorkspaceID != workspaceID || binding.VehicleID != vehicleID {
		return "", ErrCVValidation
	}
	if CVBindingRevision(*binding) != expected {
		return "", ErrCVConflict
	}
	if strings.TrimSpace(binding.Profile) != profile {
		return "", ErrCVValidation
	}
	return binding.Profile, nil
}

func CheckCVDeletionScope(vehicleID, profile string, input VehicleCVDeletionInput, binding *CVBinding) error {
	if input.WorkspaceID == "" && input.ExpectedWorkspaceRevision == "" {
		return nil
	}
	if input.WorkspaceID == "" || input.ExpectedRevision == nil || !ValidCVRevision(input.ExpectedWorkspaceRevision) {
		return ErrCVValidation
	}
	if binding == nil || binding.WorkspaceID != input.WorkspaceID || binding.VehicleID != vehicleID {
		return ErrCVValidation
	}
	if CVBindingRevision(*binding) != input.ExpectedWorkspaceRevision {
		return ErrCVConflict
	}
	if binding.Profile != profile {
		return ErrCVValidation
	}
	return nil
}
