package application

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func validDecoderID(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 16 && hex.EncodeToString(decoded) == value
}

func validDecoderDates(installed, removed *string) bool {
	for _, date := range []*string{installed, removed} {
		if date != nil {
			if _, err := time.Parse(time.DateOnly, *date); err != nil {
				return false
			}
		}
	}
	return installed == nil || removed == nil || *installed <= *removed
}

func (s *VehicleService) changeDecoderConfiguration(ctx context.Context, vehicleID, expected, action, targetID, actor string,
	apply func(*sql.Tx, *DecoderConfigurationContext, string) error) (*DecoderConfigurationContext, error) {
	if !validCVRevision(expected) {
		return nil, ErrDecoderValidation
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	before, err := readDecoderConfigurationContextTx(ctx, tx, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, err
	}
	if before.Revision != expected {
		return nil, ErrDecoderConflict
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	if err := apply(tx, before, now); err != nil {
		return nil, err
	}
	details, err := json.Marshal(map[string]string{"vehicleId": before.VehicleID})
	if err != nil {
		return nil, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_logs
(id,actor_user_id,action,target_type,target_id,created_at,details_json)
VALUES(?,NULLIF(?,''),?,'decoder_configuration',?,?,?)`, randomID(), actor, action, targetID, now,
		string(details)); err != nil {
		return nil, fmt.Errorf("audit decoder configuration: %w", err)
	}
	after, err := readDecoderConfigurationContextTx(ctx, tx, before.VehicleID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit decoder configuration: %w", err)
	}
	return after, nil
}

func decoderInstanceIn(result *DecoderConfigurationContext, id string) *DecoderInstance {
	for index := range result.Instances {
		if result.Instances[index].ID == id {
			return &result.Instances[index]
		}
	}
	return nil
}

func decoderInstallationIn(result *DecoderConfigurationContext, id string) *DecoderInstallation {
	for index := range result.Installations {
		if result.Installations[index].ID == id {
			return &result.Installations[index]
		}
	}
	return nil
}
