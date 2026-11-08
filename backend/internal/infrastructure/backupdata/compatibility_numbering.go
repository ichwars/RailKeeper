package backupdata

import (
	"context"

	"railkeeper/backend/internal/backup"
)

func (m *legacyRestoreStore) AdvanceArticleScheme(ctx context.Context, next int, at string) error {
	_, err := m.tx.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET next_number=?, updated_at=?
WHERE category='Artikel'
`, next, at)
	return err
}
func (m *legacyRestoreStore) InsertArticleScheme(ctx context.Context, entry backup.ArticleInventoryScheme) error {
	active := 0
	if entry.Active {
		active = 1
	}
	_, err := m.tx.ExecContext(ctx, `
INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
) VALUES(?, 'Artikel', ?, ?, ?, ?, ?, ?)
`, entry.ID, entry.Prefix, entry.NextNumber, entry.Padding, active, entry.CreatedAt, entry.UpdatedAt)
	return err
}
func (m *legacyRestoreStore) CreateLegacyVehicleSetScheme(ctx context.Context, id, at string) error {
	_, err := m.tx.ExecContext(ctx, `
INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
)
VALUES(?, 'Set', 'RK-SET', 1, 6, 1, ?, ?)
ON CONFLICT(category) DO NOTHING
`, id, at, at)
	return err
}
func (m *legacyRestoreStore) ActivateLegacyVehicleSetScheme(ctx context.Context, at string) error {
	_, err := m.tx.ExecContext(ctx, `
UPDATE inventory_number_schemes SET active=1, updated_at=? WHERE category='Set'
`, at)
	return err
}
func (m *legacyRestoreStore) AdvanceVehicleSetScheme(ctx context.Context, next int, at string) error {
	_, err := m.tx.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET next_number=?, updated_at=?
WHERE category='Set'
`, next, at)
	return err
}
