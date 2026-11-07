package application

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	ErrVehicleCreationConflict = errors.New("vehicle creation identity already exists")
	ErrVehicleCreationIdentity = errors.New("invalid vehicle creation identity")
)

type VehicleCreationOptions struct {
	CreationID      string
	ExternalMapping *VehicleExternalMapInput
}

func (s *VehicleService) creationIdentity(ctx context.Context, requested string) (string, error) {
	if requested == "" {
		return randomID(), nil
	}
	decoded, err := hex.DecodeString(requested)
	if err != nil || len(decoded) != 16 || hex.EncodeToString(decoded) != requested {
		return "", ErrVehicleCreationIdentity
	}
	if err := checkCreationIdentity(ctx, s.db, requested); err != nil {
		return "", err
	}
	return requested, nil
}

type creationIdentityReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func checkCreationIdentity(ctx context.Context, reader creationIdentityReader, id string) error {
	var exists bool
	if err := reader.QueryRowContext(ctx, `SELECT EXISTS(
SELECT 1 FROM vehicles WHERE id=? UNION ALL SELECT 1 FROM vehicle_sets WHERE id=?)`, id, id).Scan(&exists); err != nil {
		return fmt.Errorf("check vehicle creation identity: %w", err)
	}
	if exists {
		return ErrVehicleCreationConflict
	}
	return nil
}
