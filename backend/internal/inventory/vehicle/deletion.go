package vehicle

import (
	"context"
	"fmt"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func Delete(ctx context.Context, repository DeleteRepository, id, actor string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrVehicleNotFound
	}
	tx, err := repository.BeginDelete(ctx)
	if err != nil {
		return fmt.Errorf("begin delete vehicle: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	membership, err := tx.Membership(ctx, id)
	if err != nil {
		return err
	}
	affected, err := tx.DeleteVehicle(ctx, id)
	if err != nil {
		return err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return ErrVehicleNotFound
	}
	if membership.HasSet {
		if err := tx.RemoveEmptySet(ctx, membership.ID); err != nil {
			return err
		}
	}
	if err := tx.AuditDeleted(ctx, VehicleAudit{ID: credentials.RandomID(), Actor: actor, VehicleID: id, At: time.Now().UTC().Format(time.RFC3339)}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete vehicle: %w", err)
	}
	return nil
}
