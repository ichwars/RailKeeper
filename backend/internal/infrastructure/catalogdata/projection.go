package catalogdata

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"railkeeper/backend/internal/inventory/catalog"
	"railkeeper/backend/migrations"
)

var vehicleCatalogColumns = map[string]string{
	"name": "name", "manufacturer": "manufacturer", "articleNumber": "article_number",
	"articleSourceUrl": "article_source_url", "gauge": "gauge", "epoch": "epoch",
	"railwayCompany": "railway_company", "category": "category", "gattung": "gattung",
	"description": "description", "ean": "ean", "productionPeriod": "production_period", "listPrice": "list_price",
}

func (m *mutation) Project(ctx context.Context, before *catalog.VehicleCatalogContext, chosen *catalog.VehicleCatalogRevision, projection catalog.Projection, now string) error {
	target, err := catalogTarget(before.TargetType)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(projection.Overrides)
	if err != nil {
		return err
	}
	source := projection.SourceAuthority

	query := `INSERT INTO ` + target.links + `(` + target.column + `,revision_id,overrides_json,authority,source_authority,updated_at)
VALUES(?,?,?,'catalog',?,?) ON CONFLICT(` + target.column + `) DO UPDATE SET revision_id=excluded.revision_id,
overrides_json=excluded.overrides_json,authority=excluded.authority,source_authority=excluded.source_authority,updated_at=excluded.updated_at`
	if _, err := m.tx.ExecContext(ctx, query, before.TargetID, chosen.ID, string(raw), source, now); err != nil {
		return fmt.Errorf("assign catalog revision: %w", err)
	}
	// Update only changed selected fields. Unselected raw NULLs and every ownership field survive.
	var assignments []string
	var values []any
	for _, change := range projection.Changes {
		assignments = append(assignments, vehicleCatalogColumns[change.Name]+"=?")
		values = append(values, change.Value)
	}

	if len(assignments) == 0 {
		return nil
	}
	assignments = append(assignments, "updated_at=?")
	values = append(values, now, before.TargetID)
	if _, err := m.tx.ExecContext(ctx, `UPDATE `+target.table+` SET `+strings.Join(assignments, ",")+` WHERE id=?`, values...); err != nil {
		return fmt.Errorf("project selected catalog fields: %w", err)
	}
	return nil
}
func SuspendProjection(ctx context.Context, tx *sql.Tx) error {
	for _, name := range []string{"catalog_vehicle_create", "catalog_vehicle_update", "catalog_vehicle_link", "catalog_set_create", "catalog_set_update", "catalog_set_link"} {
		if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS `+name); err != nil {
			return fmt.Errorf("suspend catalog projection: %w", err)
		}
	}
	return nil
}
func RestoreProjection(ctx context.Context, tx *sql.Tx, legacy bool) error {
	if legacy {
		for _, target := range []struct{ table, column, review string }{
			{"vehicle_article_links", "vehicle_id", "vehicle_catalog_consistency"},
			{"vehicle_set_article_links", "vehicle_set_id", "set_catalog_consistency"},
		} {
			query := `UPDATE ` + target.table + ` SET source_authority=COALESCE(source_authority, authority),
authority=CASE WHEN (SELECT consistent FROM ` + target.review + ` c WHERE c.entity_id=` + target.table + `.` + target.column + `)
THEN 'catalog' ELSE 'legacy' END`
			if _, err := tx.ExecContext(ctx, query); err != nil {
				return fmt.Errorf("preserve legacy catalog authority: %w", err)
			}
		}
	}
	definitions, err := migrations.VehicleCatalogProjectionSQL()
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, definitions); err != nil {
		return fmt.Errorf("restore catalog projection: %w", err)
	}
	return nil
}
