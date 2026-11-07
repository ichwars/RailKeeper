package application

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/migrations"
)

func suspendVehicleCatalogProjectionTx(ctx context.Context, tx *sql.Tx) error {
	for _, name := range []string{"catalog_vehicle_create", "catalog_vehicle_update", "catalog_vehicle_link", "catalog_set_create", "catalog_set_update", "catalog_set_link"} {
		if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS `+name); err != nil {
			return fmt.Errorf("suspend catalog projection: %w", err)
		}
	}
	return nil
}

func restoreVehicleCatalogProjectionTx(ctx context.Context, tx *sql.Tx, legacy bool) error {
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
