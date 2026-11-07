package application

import (
	"context"
	"database/sql"
	"fmt"
)

type bundledFunctionSymbol struct {
	id           string
	typeName     string
	key          string
	label        string
	active       int
	sortOrder    int
	sourceURL    sql.NullString
	metadataJSON string
	createdAt    string
	updatedAt    string
	origin       string
}

func backupProfileRowForRestore(
	ctx context.Context,
	tx *sql.Tx,
	row map[string]any,
) (map[string]any, error) {
	restoreRow := make(map[string]any, len(row))
	for column, value := range row {
		restoreRow[column] = value
	}
	creatorID, valid := backupNonEmptyString(restoreRow["created_by_user_id"])
	if !valid {
		restoreRow["created_by_user_id"] = nil
		return restoreRow, nil
	}
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=?)`, creatorID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check restored transfer profile creator: %w", err)
	}
	if exists == 0 {
		restoreRow["created_by_user_id"] = nil
	}
	return restoreRow, nil
}

func readBundledMasterDataIdentities(
	ctx context.Context,
	tx *sql.Tx,
) ([]masterDataIdentity, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT type, key FROM master_data_entries WHERE origin='bundled' ORDER BY type, key`)
	if err != nil {
		return nil, fmt.Errorf("read bundled master data identities: %w", err)
	}
	defer func() { _ = rows.Close() }()
	identities := []masterDataIdentity{}
	for rows.Next() {
		var identity masterDataIdentity
		if err := rows.Scan(&identity.typeName, &identity.key); err != nil {
			return nil, fmt.Errorf("scan bundled master data identity: %w", err)
		}
		identities = append(identities, identity)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundled master data identities: %w", err)
	}
	return identities, nil
}

func readBundledFunctionSymbols(
	ctx context.Context,
	tx *sql.Tx,
) ([]bundledFunctionSymbol, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, type, key, label, active, sort_order, source_url, metadata_json,
       created_at, updated_at, origin
FROM master_data_entries
WHERE type='symbols'
  AND origin='bundled'
  AND json_extract(metadata_json, '$.library')='railkeeper-workshop-line'
ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("read bundled function symbols: %w", err)
	}
	defer func() { _ = rows.Close() }()
	symbols := []bundledFunctionSymbol{}
	for rows.Next() {
		var symbol bundledFunctionSymbol
		if err := rows.Scan(
			&symbol.id,
			&symbol.typeName,
			&symbol.key,
			&symbol.label,
			&symbol.active,
			&symbol.sortOrder,
			&symbol.sourceURL,
			&symbol.metadataJSON,
			&symbol.createdAt,
			&symbol.updatedAt,
			&symbol.origin,
		); err != nil {
			return nil, fmt.Errorf("scan bundled function symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundled function symbols: %w", err)
	}
	return symbols, nil
}

func restoreBundledFunctionSymbols(
	ctx context.Context,
	tx *sql.Tx,
	symbols []bundledFunctionSymbol,
) error {
	for _, symbol := range symbols {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'bundled')
ON CONFLICT(type, key) DO UPDATE SET
  id=excluded.id,
  label=excluded.label,
  active=excluded.active,
  sort_order=excluded.sort_order,
  source_url=excluded.source_url,
  metadata_json=excluded.metadata_json,
  created_at=excluded.created_at,
  updated_at=excluded.updated_at,
  origin='bundled'`,
			symbol.id,
			symbol.typeName,
			symbol.key,
			symbol.label,
			symbol.active,
			symbol.sortOrder,
			symbol.sourceURL,
			symbol.metadataJSON,
			symbol.createdAt,
			symbol.updatedAt,
		); err != nil {
			return fmt.Errorf("restore bundled function symbol %s: %w", symbol.key, err)
		}
	}
	return nil
}

func reconcileBundledMasterDataIdentities(
	ctx context.Context,
	tx *sql.Tx,
	identities []masterDataIdentity,
) error {
	for _, identity := range identities {
		if _, err := tx.ExecContext(ctx, `
UPDATE master_data_entries SET origin='bundled' WHERE type=? AND key=?
`, identity.typeName, identity.key); err != nil {
			return fmt.Errorf("reconcile bundled master data %s/%s: %w",
				identity.typeName, identity.key, err)
		}
	}
	return nil
}

func backupRowForRestore(table string, row map[string]any) map[string]any {
	if table != "master_data_entries" {
		return row
	}
	restoreRow := make(map[string]any, len(row)+1)
	for column, value := range row {
		restoreRow[column] = value
	}
	restoreRow["origin"] = string(MasterDataOriginCustom)
	return restoreRow
}

func backupRowsForRestore(doc *BackupDocument, table string) []map[string]any {
	if table == "vehicle_functions" && doc.Version < 24 {
		return legacyFunctionRowsForRestore(doc.Tables[table])
	}

	rows := doc.Tables[table]
	legacyColumns := []string{}
	if doc.Version <= 10 {
		legacyColumn := map[string]string{
			"track_geometry_definitions": "minimum_radius_mm",
			"plan_track_objects":         "flex_path_json",
			"layouts":                    "minimum_flex_radius_mm",
		}[table]
		if legacyColumn != "" {
			legacyColumns = append(legacyColumns, legacyColumn)
		}
	}
	if doc.Version <= 11 && table == "plan_track_objects" {
		legacyColumns = append(legacyColumns, "transition_path_json")
	}
	if doc.Version <= 13 && table == "plan_track_objects" {
		legacyColumns = append(legacyColumns, "geometry_snapshot_json")
	}
	if len(legacyColumns) == 0 {
		return rows
	}
	normalized := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		copyRow := make(map[string]any, len(row)+len(legacyColumns))
		for column, value := range row {
			copyRow[column] = value
		}
		for _, legacyColumn := range legacyColumns {
			if _, exists := copyRow[legacyColumn]; !exists {
				copyRow[legacyColumn] = nil
			}
		}
		normalized = append(normalized, copyRow)
	}
	return normalized
}

func restoreLegacyAccessoryProductDefaults(
	ctx context.Context,
	tx *sql.Tx,
	doc *BackupDocument,
) error {
	if doc.Version >= 3 {
		return nil
	}
	for _, row := range doc.Tables["accessory_products"] {
		productID, valid := backupNonEmptyString(row["id"])
		if !valid {
			continue
		}
		if _, explicit := row["article_type"]; !explicit {
			if _, err := tx.ExecContext(ctx,
				`UPDATE accessory_products SET article_type='other' WHERE id=?`, productID); err != nil {
				return fmt.Errorf("backfill restored legacy accessory article type: %w", err)
			}
		}
		if _, explicit := row["subtype"]; !explicit {
			if _, err := tx.ExecContext(ctx,
				`UPDATE accessory_products SET subtype=category WHERE id=?`, productID); err != nil {
				return fmt.Errorf("backfill restored legacy accessory subtype: %w", err)
			}
		}
		if _, explicit := row["inventory_strategy"]; !explicit {
			if _, err := tx.ExecContext(ctx, `
UPDATE accessory_products
SET inventory_strategy=CASE tracking_mode WHEN 'individual' THEN 'individual' ELSE 'quantity' END
WHERE id=?`, productID); err != nil {
				return fmt.Errorf("backfill restored legacy accessory inventory strategy: %w", err)
			}
		}
	}
	return nil
}
