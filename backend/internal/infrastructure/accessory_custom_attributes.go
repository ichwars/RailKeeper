package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (r *AccessoryRepository) AccessoryCustomAttributeDefinitions(
	ctx context.Context,
) ([]domain.AccessoryAttributeDefinition, error) {
	return accessoryCustomAttributeDefinitions(ctx, r.db)
}

func accessoryCustomAttributeDefinitions(
	ctx context.Context,
	queryer accessoryRowsQueryer,
) ([]domain.AccessoryAttributeDefinition, error) {
	rows, err := queryer.QueryContext(ctx, `
SELECT key, active, metadata_json
FROM master_data_entries
WHERE type='accessory_custom_field'
ORDER BY sort_order, key`)
	if err != nil {
		return nil, fmt.Errorf("list accessory custom attribute definitions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	definitions := []domain.AccessoryAttributeDefinition{}
	for rows.Next() {
		var key, metadataJSON string
		var active bool
		if err := rows.Scan(&key, &active, &metadataJSON); err != nil {
			return nil, fmt.Errorf("scan accessory custom attribute definition: %w", err)
		}
		var metadata map[string]any
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("parse accessory custom attribute definition %q: %w", key, err)
		}
		definition, err := application.ParseAccessoryCustomAttributeDefinition(key, active, metadata)
		if err != nil {
			return nil, fmt.Errorf("parse accessory custom attribute definition %q: %w", key, err)
		}
		definitions = append(definitions, definition)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate accessory custom attribute definitions: %w", err)
	}
	return definitions, nil
}

func accessoryProductMutationState(
	ctx context.Context,
	tx *sql.Tx,
	productID string,
	input application.CreateAccessoryProductInput,
) (application.AccessoryProductMutationState, error) {
	state := application.AccessoryProductMutationState{}
	if err := tx.QueryRowContext(ctx, `
SELECT
  EXISTS(SELECT 1 FROM master_data_entries WHERE type='article_type' AND key=? AND active=1),
  EXISTS(SELECT 1 FROM master_data_entries WHERE type='accessory_subtype' AND key=? AND active=1)
`, input.ArticleType, input.Subtype).Scan(&state.ArticleTypeActive, &state.SubtypeActive); err != nil {
		return state, fmt.Errorf("read accessory product master data lifecycle: %w", err)
	}
	definitions, err := accessoryCustomAttributeDefinitions(ctx, tx)
	if err != nil {
		return state, err
	}
	state.CustomAttributeDefinitions = definitions
	if productID == "" {
		return state, nil
	}

	current, err := scanAccessoryProduct(tx.QueryRowContext(ctx, accessoryProductSelect+` WHERE id=?`, productID))
	if errors.Is(err, sql.ErrNoRows) {
		return state, application.ErrAccessoryNotFound
	}
	if err != nil {
		return state, fmt.Errorf("read current accessory product for validation: %w", err)
	}
	attributes, err := loadAccessoryAttributes(ctx, tx, []string{productID})
	if err != nil {
		return state, err
	}
	current.Attributes = attributes[productID]
	state.Current = current
	return state, nil
}

func reserveAccessoryWriteTransaction(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `UPDATE master_data_entries SET updated_at=updated_at WHERE 0`); err != nil {
		return fmt.Errorf("reserve accessory write transaction: %w", err)
	}
	return nil
}
