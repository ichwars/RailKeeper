package catalogdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/catalog"
	"railkeeper/backend/migrations"
)

func Backfill(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, migrations.LegacyVehicleCatalogSQL); err != nil {
		return fmt.Errorf("backfill vehicle catalog: %w", err)
	}
	return nil
}
func (s *Store) Report(ctx context.Context) ([]catalog.RawAssignment, error) {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	return (&mutation{tx: tx}).Report(ctx)
}
func (m *mutation) Report(ctx context.Context) ([]catalog.RawAssignment, error) {
	rows, err := m.tx.QueryContext(ctx, `
SELECT 'vehicle', v.entity_id, v.fields_json, COALESCE(m.vehicle_set_id, ''),
  COALESCE(r.article_id, ''), COALESCE(l.revision_id, ''), COALESCE(l.authority, 'unlinked'),
  COALESCE(r.fields_json, '{}'), COALESCE(l.overrides_json, '{}')
FROM vehicle_catalog_legacy_fields v LEFT JOIN vehicle_article_links l ON l.vehicle_id=v.entity_id
LEFT JOIN catalog_article_revisions r ON r.id=l.revision_id
LEFT JOIN vehicle_set_members m ON m.vehicle_id=v.entity_id
UNION ALL
SELECT 'set', s.entity_id, s.fields_json, '', COALESCE(r.article_id, ''),
  COALESCE(l.revision_id, ''), COALESCE(l.authority, 'unlinked'),
  COALESCE(r.fields_json, '{}'), COALESCE(l.overrides_json, '{}')
FROM vehicle_set_catalog_legacy_fields s LEFT JOIN vehicle_set_article_links l ON l.vehicle_set_id=s.entity_id
LEFT JOIN catalog_article_revisions r ON r.id=l.revision_id
ORDER BY 1, 2`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	report := []catalog.RawAssignment{}
	for rows.Next() {
		var row catalog.RawAssignment
		assignment := &row.Assignment
		if err := rows.Scan(&assignment.TargetType, &assignment.TargetID, &row.LegacyJSON, &assignment.SetID, &assignment.ArticleID, &assignment.RevisionID, &assignment.Authority, &row.FieldsJSON, &row.OverridesJSON); err != nil {
			return nil, err
		}
		report = append(report, row)
	}
	return report, rows.Err()
}

func (m *mutation) Backfill(ctx context.Context) error { return Backfill(ctx, m.tx) }
func (m *mutation) LinkCount(ctx context.Context) (int, error) {
	var count int
	err := m.tx.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM vehicle_article_links)+(SELECT COUNT(*) FROM vehicle_set_article_links)`).Scan(&count)
	return count, err
}
func (m *mutation) AuditBackfill(ctx context.Context, added int, actor, now string) error {
	details, err := json.Marshal(map[string]int{"linksAdded": added})
	if err != nil {
		return err
	}
	_, err = m.tx.ExecContext(ctx, `INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), 'VehicleCatalogLegacyBackfilled', 'catalog', 'legacy-backfill', ?, ?)`, credentials.RandomID(), actor, now, string(details))
	return err
}
