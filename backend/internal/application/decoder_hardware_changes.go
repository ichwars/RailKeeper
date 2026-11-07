package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/decoder"
	"railkeeper/backend/internal/infrastructure/vehicledata"
)

type DecoderInstanceInput = decoder.DecoderInstanceInput

func (s *VehicleService) CreateDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string) (*DecoderConfigurationContext, error) {
	return s.writeDecoderInstance(ctx, vehicleID, input, actor, true)
}

func (s *VehicleService) UpdateDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string) (*DecoderConfigurationContext, error) {
	return s.writeDecoderInstance(ctx, vehicleID, input, actor, false)
}

func (s *VehicleService) writeDecoderInstance(ctx context.Context, vehicleID string, input DecoderInstanceInput, actor string, create bool) (*DecoderConfigurationContext, error) {
	input, err := decoder.CleanInstanceInput(input, create)
	if err != nil {
		return nil, err
	}
	action := "DecoderInstanceUpdated"
	if create {
		action = "DecoderInstanceCreated"
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, action, input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanInstance(before, input, now, create)
			if err != nil {
				return err
			}
			return vehicledata.WriteDecoderInstance(ctx, tx, item, create)
		})
}

type DecoderInstallationInput = decoder.DecoderInstallationInput

func (s *VehicleService) InstallDecoder(ctx context.Context, vehicleID string, input DecoderInstallationInput, actor string) (*DecoderConfigurationContext, error) {
	if !decoder.ValidInstallationInput(input) {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderInstalled", input.ID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanInstallation(before, input, now)
			if err != nil {
				return err
			}
			return vehicledata.InsertDecoderInstallation(ctx, tx, item)
		})
}

type DecoderRemovalInput = decoder.DecoderRemovalInput

func (s *VehicleService) RemoveDecoder(ctx context.Context, vehicleID, installationID string, input DecoderRemovalInput, actor string) (*DecoderConfigurationContext, error) {
	if !decoder.ValidRemovalInput(input) {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderRemoved", installationID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanRemoval(before, installationID, input, now)
			if err != nil {
				return err
			}
			return vehicledata.RemoveDecoderInstallation(ctx, tx, item)
		})
}

type DecoderInstallationDatesInput = decoder.DecoderInstallationDatesInput

func (s *VehicleService) CorrectDecoderInstallationDates(ctx context.Context, vehicleID, installationID string, input DecoderInstallationDatesInput, actor string) (*DecoderConfigurationContext, error) {
	if !decoder.ValidDecoderDates(input.InstalledAt, input.RemovedAt) {
		return nil, ErrDecoderValidation
	}
	return s.changeDecoderConfiguration(ctx, vehicleID, input.ExpectedRevision, "DecoderInstallationDatesCorrected", installationID, actor,
		func(tx *sql.Tx, before *DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanInstallationDates(before, installationID, input, now)
			if err != nil {
				return err
			}
			return vehicledata.CorrectDecoderInstallationDates(ctx, tx, item)
		})
}
