package application

import (
	"context"
	"database/sql"
	"strings"
)

type DecoderInstanceInput struct {
	ID               string `json:"id"`
	ExpectedRevision string `json:"expectedRevision"`
	Label            string `json:"label"`
	Manufacturer     string `json:"manufacturer"`
	Model            string `json:"model"`
	SerialNumber     string `json:"serialNumber"`
	Notes            string `json:"notes"`
}

func (s *VehicleService) CreateDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string) (*DecoderConfigurationContext, error) {
	return s.writeDecoderInstance(ctx, vehicleID, input, actor, true)
}

func (s *VehicleService) UpdateDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string) (*DecoderConfigurationContext, error) {
	return s.writeDecoderInstance(ctx, vehicleID, input, actor, false)
}

func (s *VehicleService) writeDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string, create bool) (*DecoderConfigurationContext, error) {
	input.Label, input.Manufacturer = strings.TrimSpace(input.Label), strings.TrimSpace(input.Manufacturer)
	input.Model, input.SerialNumber = strings.TrimSpace(input.Model), strings.TrimSpace(input.SerialNumber)
	input.Notes = strings.TrimSpace(input.Notes)
	if (create && !validDecoderID(input.ID)) || input.ID == "" || len(input.ID) > 80 || input.Label == "" || len(input.Notes) > 4000 {
		return nil, ErrDecoderValidation
	}
	for _, value := range []string{input.Label, input.Manufacturer, input.Model, input.SerialNumber} {
		if len(value) > 160 {
			return nil, ErrDecoderValidation
		}
	}
	action := "DecoderInstanceUpdated"
	if create {
		action = "DecoderInstanceCreated"
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, action, input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			existing := decoderInstanceIn(before, input.ID)
			if create {
				if existing != nil {
					return ErrDecoderConflict
				}
				_, err := tx.ExecContext(ctx, `INSERT INTO decoder_instances
(id,label,manufacturer,model,serial_number,notes,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?)`,
					input.ID, input.Label, input.Manufacturer, input.Model, input.SerialNumber, input.Notes, now, now)
				return err
			}
			if existing == nil {
				return ErrDecoderNotFound
			}
			_, err := tx.ExecContext(ctx, `UPDATE decoder_instances SET label=?,manufacturer=?,model=?,serial_number=?,notes=?,updated_at=? WHERE id=?`,
				input.Label, input.Manufacturer, input.Model, input.SerialNumber, input.Notes, now, input.ID)
			return err
		})
}

type DecoderInstallationInput struct {
	ID               string  `json:"id"`
	ExpectedRevision string  `json:"expectedRevision"`
	DecoderID        string  `json:"decoderId"`
	InstalledAt      *string `json:"installedAt"`
	Notes            string  `json:"notes"`
}

func (s *VehicleService) InstallDecoder(ctx context.Context, vehicleID string, input DecoderInstallationInput, actor string) (*DecoderConfigurationContext, error) {
	if !validDecoderID(input.ID) || input.DecoderID == "" || !validDecoderDates(input.InstalledAt, nil) || len(input.Notes) > 4000 {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderInstalled", input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			if decoderInstanceIn(before, input.DecoderID) == nil {
				return ErrDecoderNotFound
			}
			for _, item := range before.Installations {
				if item.ID == input.ID || (item.DecoderID == input.DecoderID && item.State == "installed") {
					return ErrDecoderConflict
				}
			}
			_, err := tx.ExecContext(ctx, `INSERT INTO decoder_installations
(id,decoder_id,vehicle_id,vehicle_reference_id,vehicle_name,vehicle_inventory_number,state,installed_at,removed_at,notes,recorded_at,updated_at)
VALUES(?,?,?,?,?,?,'installed',?,NULL,?,?,?)`, input.ID, input.DecoderID, before.VehicleID, before.VehicleID, before.VehicleName, before.InventoryNumber,
				input.InstalledAt, strings.TrimSpace(input.Notes), now, now)
			return err
		})
}

type DecoderRemovalInput struct {
	ExpectedRevision string  `json:"expectedRevision"`
	RemovedAt        *string `json:"removedAt"`
	Note             string  `json:"note"`
}

func (s *VehicleService) RemoveDecoder(ctx context.Context, vehicleID, installationID string, input DecoderRemovalInput, actor string) (*DecoderConfigurationContext, error) {
	if !validDecoderDates(nil, input.RemovedAt) || len(input.Note) > 2000 {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderRemoved", installationID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			item := decoderInstallationIn(before, installationID)
			if item == nil || item.VehicleReferenceID != before.VehicleID {
				return ErrDecoderNotFound
			}
			if item.State == "removed" {
				return ErrDecoderConflict
			}
			if !validDecoderDates(item.InstalledAt, input.RemovedAt) {
				return ErrDecoderValidation
			}
			notes := item.Notes
			if note := strings.TrimSpace(input.Note); note != "" {
				if notes != "" {
					notes += "\n"
				}
				notes += note
			}
			if len(notes) > 4000 {
				return ErrDecoderValidation
			}
			_, err := tx.ExecContext(ctx, `UPDATE decoder_installations SET state='removed',removed_at=?,notes=?,updated_at=? WHERE id=?`, input.RemovedAt, notes, now, installationID)
			return err
		})
}

type DecoderInstallationDatesInput struct {
	ExpectedRevision string  `json:"expectedRevision"`
	InstalledAt      *string `json:"installedAt"`
	RemovedAt        *string `json:"removedAt"`
}

func (s *VehicleService) CorrectDecoderInstallationDates(ctx context.Context, vehicleID, installationID string, input DecoderInstallationDatesInput, actor string) (*DecoderConfigurationContext, error) {
	if !validDecoderDates(input.InstalledAt, input.RemovedAt) {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderInstallationDatesCorrected", installationID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			item := decoderInstallationIn(before, installationID)
			if item == nil || item.VehicleReferenceID != before.VehicleID {
				return ErrDecoderNotFound
			}
			if item.State != "removed" && input.RemovedAt != nil {
				return ErrDecoderValidation
			}
			_, err := tx.ExecContext(ctx, `UPDATE decoder_installations SET installed_at=?,removed_at=?,updated_at=? WHERE id=?`, input.InstalledAt, input.RemovedAt, now, installationID)
			return err
		})
}
