package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/application"
)

func (repository *DataTransferRepository) ValidateTransferAccessoryReferences(
	ctx context.Context,
	accessory application.TransferAccessory,
	targetID string,
) error {
	return validateTransferAccessoryReferenceData(ctx, repository.db, accessory, targetID)
}

func validateTransferAccessoryReferenceData(
	ctx context.Context,
	db dataTransferApplyDB,
	accessory application.TransferAccessory,
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
			return application.ErrDataTransferConflict
		}
		if err != nil {
			return fmt.Errorf("load transfer accessory reference target: %w", err)
		}
		state.Current = &current
	}
	return application.ValidateTransferAccessoryReferenceState(accessory, state)
}
