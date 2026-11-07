package decoder

import "strings"

func CleanWorkspaceInput(input DecoderCVWorkspaceInput) (DecoderCVWorkspaceInput, error) {
	input.Profile = strings.TrimSpace(input.Profile)
	if !ValidDecoderID(input.ID) || input.Profile == "" || len(input.Profile) > 160 {
		return input, ErrValidation
	}
	return input, nil
}

func PlanWorkspace(before *DecoderConfigurationContext, input DecoderCVWorkspaceInput, now string, exists bool) (DecoderCVWorkspace, error) {
	input, err := CleanWorkspaceInput(input)
	if err != nil {
		return DecoderCVWorkspace{}, err
	}
	if exists {
		return DecoderCVWorkspace{}, ErrConflict
	}
	return DecoderCVWorkspace{ID: input.ID, VehicleID: before.VehicleID, Profile: input.Profile,
		Origin: "manual", CreatedAt: now, UpdatedAt: now}, nil
}

func ValidCheckpointInput(input DecoderCVCheckpointInput) bool {
	return ValidDecoderID(input.ID) && len(input.Note) <= 2000
}

func ValidAssignmentInput(input DecoderCVAssignmentInput) bool {
	return ValidDecoderID(input.SnapshotID) && (input.InstallationID == nil || *input.InstallationID != "")
}

// Before is captured as a partial state before the same transaction changes the binding.
type WorkspaceAssignment struct {
	Before DecoderCVWorkspace
	After  DecoderCVWorkspace
}

func PlanWorkspaceAssignment(before *DecoderConfigurationContext, workspaceID string, input DecoderCVAssignmentInput, now string) (WorkspaceAssignment, error) {
	if !ValidAssignmentInput(input) {
		return WorkspaceAssignment{}, ErrValidation
	}
	workspace := WorkspaceIn(before, workspaceID)
	if workspace == nil {
		return WorkspaceAssignment{}, ErrNotFound
	}
	if (workspace.InstallationID == nil && input.InstallationID == nil) ||
		(workspace.InstallationID != nil && input.InstallationID != nil && *workspace.InstallationID == *input.InstallationID) {
		return WorkspaceAssignment{}, ErrConflict
	}
	if input.InstallationID != nil {
		installation := InstallationIn(before, *input.InstallationID)
		if installation == nil || installation.VehicleID == nil || *installation.VehicleID != before.VehicleID || installation.State != "installed" {
			return WorkspaceAssignment{}, ErrValidation
		}
	}
	result := WorkspaceAssignment{Before: *workspace, After: *workspace}
	result.After.InstallationID, result.After.UpdatedAt = input.InstallationID, now
	return result, nil
}
