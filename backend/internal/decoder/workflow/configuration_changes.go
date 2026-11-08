package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/decoder"
)

func (s *ConfigurationService) Context(ctx context.Context, vehicleID string) (*decoder.DecoderConfigurationContext, error) {
	tx, err := s.repository.BeginRead(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return tx.Context(ctx, strings.TrimSpace(vehicleID))
}
func (s *ConfigurationService) change(ctx context.Context, vehicleID, expected, action, targetID, actor string,
	apply func(ConfigurationMutation, *decoder.DecoderConfigurationContext, string) error) (*decoder.DecoderConfigurationContext, error) {
	if !decoder.ValidCVRevision(expected) {
		return nil, decoder.ErrValidation
	}
	tx, err := s.repository.BeginChange(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := tx.Context(ctx, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, err
	}
	if before.Revision != expected {
		return nil, decoder.ErrConflict
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := apply(tx, before, now); err != nil {
		return nil, err
	}
	details, err := json.Marshal(map[string]string{"vehicleId": before.VehicleID})
	if err != nil {
		return nil, err
	}
	if err := tx.AuditConfiguration(ctx, ConfigurationAudit{ID: credentials.RandomID(), Actor: actor, Action: action, TargetID: targetID, At: now, Details: string(details)}); err != nil {
		return nil, err
	}

	after, err := tx.Context(ctx, before.VehicleID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit decoder configuration: %w", err)
	}
	return after, nil
}

func TouchWorkspace(ctx context.Context, writer WorkspaceWriter, vehicleID, profile, now string) error {
	return writer.TouchWorkspace(ctx, WorkspaceTouch{ID: credentials.RandomID(), VehicleID: vehicleID, Profile: profile, At: now})
}
