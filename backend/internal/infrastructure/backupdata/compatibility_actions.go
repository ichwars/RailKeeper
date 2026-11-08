package backupdata

import (
	"context"
	"fmt"
	"time"

	"railkeeper/backend/internal/decoder/workflow"
	"railkeeper/backend/internal/infrastructure/catalogdata"
	"railkeeper/backend/internal/infrastructure/masterdatastore"
	"railkeeper/backend/internal/infrastructure/vehicledata"
	"railkeeper/backend/internal/inventory/masterdata"
)

func (m *legacyRestoreStore) SuspendCatalogProjection(ctx context.Context) error {
	return catalogdata.SuspendProjection(ctx, m.tx)
}
func (m *legacyRestoreStore) RestoreCatalogProjection(ctx context.Context, legacy bool) error {
	return catalogdata.RestoreProjection(ctx, m.tx, legacy)
}
func (m *legacyRestoreStore) BackfillVehicleCatalog(ctx context.Context) error {
	return catalogdata.Backfill(ctx, m.tx)
}
func (m *legacyRestoreStore) RestoreMissingDecoderLibrary(ctx context.Context) error {
	return masterdata.SeedDecoderLibrary(ctx, masterdatastore.SeedStoreFor(m.tx), time.Now().UTC().Format(time.RFC3339))
}
func (m *legacyRestoreStore) AdoptLegacyDecoderSnapshots(ctx context.Context, sourceRoot string) error {
	return workflow.AdoptLegacySnapshotsWithin(ctx, vehicledata.LegacySnapshotMutationFor(m.tx, sourceRoot))
}
func (m *legacyRestoreStore) TransferProfileCreatorExists(ctx context.Context, id string) (bool, error) {
	var exists int
	if err := m.tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check restored transfer profile creator: %w", err)
	}
	return exists != 0, nil
}
func (m *legacyRestoreStore) DefaultAccessoryArticleType(ctx context.Context, id string) error {
	if _, err := m.tx.ExecContext(ctx, `UPDATE accessory_products SET article_type='other' WHERE id=?`, id); err != nil {
		return fmt.Errorf("backfill restored legacy accessory article type: %w", err)
	}
	return nil
}
func (m *legacyRestoreStore) DefaultAccessorySubtype(ctx context.Context, id string) error {
	if _, err := m.tx.ExecContext(ctx, `UPDATE accessory_products SET subtype=category WHERE id=?`, id); err != nil {
		return fmt.Errorf("backfill restored legacy accessory subtype: %w", err)
	}
	return nil
}
func (m *legacyRestoreStore) DefaultAccessoryInventoryStrategy(ctx context.Context, id string) error {
	if _, err := m.tx.ExecContext(ctx, `
UPDATE accessory_products
SET inventory_strategy=CASE tracking_mode WHEN 'individual' THEN 'individual' ELSE 'quantity' END
WHERE id=?`, id); err != nil {
		return fmt.Errorf("backfill restored legacy accessory inventory strategy: %w", err)
	}
	return nil
}
