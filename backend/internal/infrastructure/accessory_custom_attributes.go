package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

func (r *AccessoryRepository) AccessoryCustomAttributeDefinitions(
	ctx context.Context,
) ([]domain.AccessoryAttributeDefinition, error) {
	rows, err := r.db.QueryContext(ctx, `
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
