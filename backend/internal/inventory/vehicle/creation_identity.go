package vehicle

import (
	"context"
	"encoding/hex"
	"errors"
	"railkeeper/backend/internal/credentials"
)

var (
	ErrVehicleCreationConflict = errors.New("vehicle creation identity already exists")
	ErrVehicleCreationIdentity = errors.New("invalid vehicle creation identity")
)

type CreationOptions struct {
	CreationID      string
	ExternalMapping *VehicleExternalMapInput
}
type CreationIdentityLookup interface {
	CheckCreationIdentity(context.Context, string) error
}

func ResolveCreationIdentity(ctx context.Context, lookup CreationIdentityLookup, requested string) (string, error) {
	if requested == "" {
		return credentials.RandomID(), nil
	}
	decoded, err := hex.DecodeString(requested)
	if err != nil || len(decoded) != 16 || hex.EncodeToString(decoded) != requested {
		return "", ErrVehicleCreationIdentity
	}
	if err := lookup.CheckCreationIdentity(ctx, requested); err != nil {
		return "", err
	}
	return requested, nil
}
