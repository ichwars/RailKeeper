package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

func applyCatalogRevision(ctx context.Context, tx *sql.Tx, before *VehicleCatalogContext, chosen *VehicleCatalogRevision, selected []string, now string) error {
	target, err := catalogTarget(before.TargetType)
	if err != nil {
		return err
	}
	fields := maps.Clone(before.CurrentFields)
	for _, field := range selected {
		if !validCatalogRequiredField(field, chosen.Fields[field]) {
			return ErrVehicleCatalogValidation
		}
		fields[field] = chosen.Fields[field]
	}
	overrides := VehicleCatalogFields{}
	for field, value := range fields {
		if chosen.Fields[field] != value {
			overrides[field] = value
		}
	}
	raw, err := json.Marshal(overrides)
	if err != nil {
		return err
	}
	source := before.SourceAuthority
	if source == nil && before.Authority == "legacy" {
		legacy := "legacy"
		source = &legacy
	}
	query := `INSERT INTO ` + target.links + `(` + target.column + `,revision_id,overrides_json,authority,source_authority,updated_at)
VALUES(?,?,?,'catalog',?,?) ON CONFLICT(` + target.column + `) DO UPDATE SET revision_id=excluded.revision_id,
overrides_json=excluded.overrides_json,authority=excluded.authority,source_authority=excluded.source_authority,updated_at=excluded.updated_at`
	if _, err := tx.ExecContext(ctx, query, before.TargetID, chosen.ID, string(raw), source, now); err != nil {
		return fmt.Errorf("assign catalog revision: %w", err)
	}
	// Update only changed selected fields. Unselected raw NULLs and every ownership field survive.
	var assignments []string
	var values []any
	for _, field := range vehicleCatalogFieldNames {
		if fields[field] != before.CurrentFields[field] {
			assignments = append(assignments, vehicleCatalogColumns[field]+"=?")
			values = append(values, fields[field])
		}
	}
	if len(assignments) == 0 {
		return nil
	}
	assignments = append(assignments, "updated_at=?")
	values = append(values, now, before.TargetID)
	if _, err := tx.ExecContext(ctx, `UPDATE `+target.table+` SET `+strings.Join(assignments, ",")+` WHERE id=?`, values...); err != nil {
		return fmt.Errorf("project selected catalog fields: %w", err)
	}
	return nil
}
