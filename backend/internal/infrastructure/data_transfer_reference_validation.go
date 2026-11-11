package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/transfer"
)

func (repository *DataTransferRepository) ValidateTransferAccessoryReferences(
	ctx context.Context, checks []transfer.AccessoryReferenceCheck,
) ([]bool, error) {
	requested := make([][3]string, len(checks))
	for i, check := range checks {
		articleType, subtype := application.TransferAccessoryReferenceKeys(check.Accessory)
		requested[i] = [3]string{string(articleType), subtype, strings.TrimSpace(check.TargetID)}
	}
	payload, err := json.Marshal(requested)
	if err != nil {
		return nil, err
	}
	rows, err := repository.db.QueryContext(ctx, `
WITH requested(position, article_type, subtype, target_id) AS (
  SELECT key, value->>0, value->>1, value->>2 FROM json_each(?)
)
SELECT requested.position, type.key IS NOT NULL, subtype.key IS NOT NULL, product.id IS NOT NULL,
       COALESCE(product.article_type, ''), COALESCE(product.subtype, '')
FROM requested
LEFT JOIN master_data_entries type ON type.type='article_type' AND type.key=requested.article_type AND type.active=1
LEFT JOIN master_data_entries subtype ON subtype.type='accessory_subtype' AND subtype.key=requested.subtype AND subtype.active=1
LEFT JOIN accessory_products product ON product.id=NULLIF(requested.target_id, '')
`, string(payload))
	if err != nil {
		return nil, fmt.Errorf("validate transfer accessory master data: %w", err)
	}
	defer func() { _ = rows.Close() }()
	valid := make([]bool, len(checks))
	for rows.Next() {
		var index int
		var targetFound bool
		var state application.AccessoryProductMutationState
		var current application.AccessoryProduct
		if err := rows.Scan(&index, &state.ArticleTypeActive, &state.SubtypeActive, &targetFound,
			&current.ArticleType, &current.Subtype); err != nil {
			return nil, fmt.Errorf("scan transfer accessory master data: %w", err)
		}
		if requested[index][2] != "" {
			if !targetFound {
				return nil, transfer.ErrDataTransferConflict
			}
			state.Current = &current
		}
		valid[index] = application.ValidateTransferAccessoryReferenceState(checks[index].Accessory, state) == nil
	}
	return valid, rows.Err()
}

func validateTransferAccessoryReferenceData(
	ctx context.Context,
	db dataTransferApplyDB,
	accessory transfer.TransferAccessory,
	targetID string,
) error {
	articleType, subtype := application.TransferAccessoryReferenceKeys(accessory)
	state := application.AccessoryProductMutationState{}
	if err := db.QueryRowContext(ctx, `
SELECT
  EXISTS(SELECT 1 FROM master_data_entries WHERE type='article_type' AND key=? AND active=1),
  EXISTS(SELECT 1 FROM master_data_entries WHERE type='accessory_subtype' AND key=? AND active=1)
`, articleType, subtype).Scan(&state.ArticleTypeActive, &state.SubtypeActive); err != nil {
		return fmt.Errorf("validate transfer accessory master data: %w", err)
	}
	targetID = strings.TrimSpace(targetID)
	if targetID != "" {
		current := application.AccessoryProduct{}
		err := db.QueryRowContext(ctx, `
SELECT article_type, subtype FROM accessory_products WHERE id=?`, targetID).
			Scan(&current.ArticleType, &current.Subtype)
		if errors.Is(err, sql.ErrNoRows) {
			return transfer.ErrDataTransferConflict
		}
		if err != nil {
			return fmt.Errorf("load transfer accessory reference target: %w", err)
		}
		state.Current = &current
	}
	return application.ValidateTransferAccessoryReferenceState(accessory, state)
}
