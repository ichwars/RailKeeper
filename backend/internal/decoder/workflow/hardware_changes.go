package workflow

import (
	"context"

	"railkeeper/backend/internal/decoder"
)

func (s *ConfigurationService) CreateDecoderInstance(ctx context.Context, vehicleID string, input decoder.DecoderInstanceInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	return s.writeDecoderInstance(ctx, vehicleID, input, actor, true)
}

func (s *ConfigurationService) UpdateDecoderInstance(ctx context.Context, vehicleID string, input decoder.DecoderInstanceInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	return s.writeDecoderInstance(ctx, vehicleID, input, actor, false)
}

func (s *ConfigurationService) writeDecoderInstance(ctx context.Context, vehicleID string, input decoder.DecoderInstanceInput, actor string, create bool) (*decoder.DecoderConfigurationContext, error) {
	input, err := decoder.CleanInstanceInput(input, create)
	if err != nil {
		return nil, err
	}
	action := "DecoderInstanceUpdated"
	if create {
		action = "DecoderInstanceCreated"
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, action, input.ID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanInstance(before, input, now, create)
			if err != nil {
				return err
			}
			return tx.WriteDecoderInstance(ctx, item, create)
		})
}

func (s *ConfigurationService) InstallDecoder(ctx context.Context, vehicleID string, input decoder.DecoderInstallationInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	if !decoder.ValidInstallationInput(input) {
		return nil, decoder.ErrValidation
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, "DecoderInstalled", input.ID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanInstallation(before, input, now)
			if err != nil {
				return err
			}
			return tx.InsertDecoderInstallation(ctx, item)
		})
}

func (s *ConfigurationService) RemoveDecoder(ctx context.Context, vehicleID, installationID string, input decoder.DecoderRemovalInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	if !decoder.ValidRemovalInput(input) {
		return nil, decoder.ErrValidation
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, "DecoderRemoved", installationID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanRemoval(before, installationID, input, now)
			if err != nil {
				return err
			}
			return tx.RemoveDecoderInstallation(ctx, item)
		})
}

func (s *ConfigurationService) CorrectDecoderInstallationDates(ctx context.Context, vehicleID, installationID string, input decoder.DecoderInstallationDatesInput, actor string) (*decoder.DecoderConfigurationContext, error) {
	if !decoder.ValidDecoderDates(input.InstalledAt, input.RemovedAt) {
		return nil, decoder.ErrValidation
	}
	return s.change(ctx, vehicleID, input.ExpectedRevision, "DecoderInstallationDatesCorrected", installationID, actor,
		func(tx ConfigurationMutation, before *decoder.DecoderConfigurationContext, now string) error {
			item, err := decoder.PlanInstallationDates(before, installationID, input, now)
			if err != nil {
				return err
			}
			return tx.CorrectDecoderInstallationDates(ctx, item)
		})
}
