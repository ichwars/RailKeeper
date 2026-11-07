package decoder

import "strings"

func CleanInstanceInput(input DecoderInstanceInput, create bool) (DecoderInstanceInput, error) {
	input.Label, input.Manufacturer = strings.TrimSpace(input.Label), strings.TrimSpace(input.Manufacturer)
	input.Model, input.SerialNumber = strings.TrimSpace(input.Model), strings.TrimSpace(input.SerialNumber)
	input.Notes = strings.TrimSpace(input.Notes)
	if (create && !ValidDecoderID(input.ID)) || input.ID == "" || input.Label == "" || len(input.Notes) > 4000 {
		return input, ErrValidation
	}
	for _, value := range []string{input.Label, input.Manufacturer, input.Model, input.SerialNumber} {
		if len(value) > 160 {
			return input, ErrValidation
		}
	}
	return input, nil
}

func PlanInstance(before *DecoderConfigurationContext, input DecoderInstanceInput, now string, create bool) (DecoderInstance, error) {
	input, err := CleanInstanceInput(input, create)
	if err != nil {
		return DecoderInstance{}, err
	}
	existing := InstanceIn(before, input.ID)
	if create && existing != nil {
		return DecoderInstance{}, ErrConflict
	}
	if !create && existing == nil {
		return DecoderInstance{}, ErrNotFound
	}
	item := DecoderInstance{ID: input.ID, Label: input.Label, Manufacturer: input.Manufacturer,
		Model: input.Model, SerialNumber: input.SerialNumber, Notes: input.Notes, CreatedAt: now, UpdatedAt: now}
	if existing != nil {
		item.CreatedAt = existing.CreatedAt
	}
	return item, nil
}

func ValidInstallationInput(input DecoderInstallationInput) bool {
	return ValidDecoderID(input.ID) && input.DecoderID != "" && ValidDecoderDates(input.InstalledAt, nil) && len(input.Notes) <= 4000
}

func PlanInstallation(before *DecoderConfigurationContext, input DecoderInstallationInput, now string) (DecoderInstallation, error) {
	if !ValidInstallationInput(input) {
		return DecoderInstallation{}, ErrValidation
	}
	if InstanceIn(before, input.DecoderID) == nil {
		return DecoderInstallation{}, ErrNotFound
	}
	for _, item := range before.Installations {
		if item.ID == input.ID || (item.DecoderID == input.DecoderID && item.State == "installed") {
			return DecoderInstallation{}, ErrConflict
		}
	}
	vehicleID := before.VehicleID
	return DecoderInstallation{ID: input.ID, DecoderID: input.DecoderID, VehicleID: &vehicleID,
		VehicleReferenceID: vehicleID, VehicleName: before.VehicleName, VehicleInventoryNumber: before.InventoryNumber,
		State: "installed", InstalledAt: input.InstalledAt, Notes: strings.TrimSpace(input.Notes), RecordedAt: now, UpdatedAt: now}, nil
}

func ValidRemovalInput(input DecoderRemovalInput) bool {
	return ValidDecoderDates(nil, input.RemovedAt) && len(input.Note) <= 2000
}

func PlanRemoval(before *DecoderConfigurationContext, installationID string, input DecoderRemovalInput, now string) (DecoderInstallation, error) {
	if !ValidRemovalInput(input) {
		return DecoderInstallation{}, ErrValidation
	}
	item := InstallationIn(before, installationID)
	if item == nil || item.VehicleReferenceID != before.VehicleID {
		return DecoderInstallation{}, ErrNotFound
	}
	if item.State == "removed" {
		return DecoderInstallation{}, ErrConflict
	}
	if !ValidDecoderDates(item.InstalledAt, input.RemovedAt) {
		return DecoderInstallation{}, ErrValidation
	}
	result := *item
	if note := strings.TrimSpace(input.Note); note != "" {
		if result.Notes != "" {
			result.Notes += "\n"
		}
		result.Notes += note
	}
	if len(result.Notes) > 4000 {
		return DecoderInstallation{}, ErrValidation
	}
	result.State, result.RemovedAt, result.UpdatedAt = "removed", input.RemovedAt, now
	return result, nil
}

func PlanInstallationDates(before *DecoderConfigurationContext, installationID string, input DecoderInstallationDatesInput, now string) (DecoderInstallation, error) {
	if !ValidDecoderDates(input.InstalledAt, input.RemovedAt) {
		return DecoderInstallation{}, ErrValidation
	}
	item := InstallationIn(before, installationID)
	if item == nil || item.VehicleReferenceID != before.VehicleID {
		return DecoderInstallation{}, ErrNotFound
	}
	if item.State != "removed" && input.RemovedAt != nil {
		return DecoderInstallation{}, ErrValidation
	}
	result := *item
	result.InstalledAt, result.RemovedAt, result.UpdatedAt = input.InstalledAt, input.RemovedAt, now
	return result, nil
}
