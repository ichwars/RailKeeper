package backup

import (
	"context"

	"railkeeper/backend/internal/inventory/masterdata"
)

type CompatibilityOperation struct {
	doc               *BackupDocument
	store             LegacyRestoreStore
	articleMasterRows []StoredMasterRow
	bundledIdentities []BundledIdentity
	bundledSymbols    []StoredMasterRow
	articleScheme     *ArticleInventoryScheme
}

func BeginCompatibilityRestore(ctx context.Context, store LegacyRestoreStore, doc *BackupDocument) (*CompatibilityOperation, error) {
	if err := store.SuspendCatalogProjection(ctx); err != nil {
		return nil, err
	}
	op := &CompatibilityOperation{doc: doc, store: store}
	var err error
	if doc.Version < 3 {
		op.articleMasterRows, err = store.LegacyArticleMasterRows(ctx)
		if err != nil {
			return nil, err
		}
		keys := map[string]bool{}
		for _, row := range op.articleMasterRows {
			if row.TypeName == masterdata.StandardArticleType {
				keys[row.Key] = true
			}
		}
		expected := masterdata.StandardArticleTypeKeys()
		if len(keys) != len(expected) {
			return nil, ErrBackupInvalid
		}
		for _, key := range expected {
			if !keys[key] {
				return nil, ErrBackupInvalid
			}
		}
	}
	op.bundledIdentities, err = store.BundledMasterIdentities(ctx)
	if err != nil {
		return nil, err
	}
	op.bundledSymbols, err = store.BundledFunctionSymbols(ctx)
	if err != nil {
		return nil, err
	}
	op.articleScheme, err = store.ArticleScheme(ctx)
	if err != nil {
		return nil, err
	}
	return op, nil
}

func (op *CompatibilityOperation) Rows(ctx context.Context, table string) ([]map[string]any, error) {
	if table == "vehicle_sets" {
		if err := prepareVehicleSetInventoryNumbers(ctx, op.store, op.doc); err != nil {
			return nil, err
		}
	}
	if table == "accessory_products" {
		if err := prepareArticleInventoryNumbers(ctx, op.store, op.doc, op.articleScheme); err != nil {
			return nil, err
		}
	}
	rows := rowsForCompatibilityRestore(op.doc, table)
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		result = append(result, rowForCompatibilityRestore(table, row))
	}
	return result, nil
}

func (op *CompatibilityOperation) ProfileRow(ctx context.Context, row map[string]any) (map[string]any, error) {
	restored := make(map[string]any, len(row))
	for column, value := range row {
		restored[column] = value
	}
	creatorID, valid := BackupNonEmptyString(restored["created_by_user_id"])
	if !valid {
		restored["created_by_user_id"] = nil
		return restored, nil
	}
	exists, err := op.store.TransferProfileCreatorExists(ctx, creatorID)
	if err != nil {
		return nil, err
	}
	if !exists {
		restored["created_by_user_id"] = nil
	}
	return restored, nil
}

func (op *CompatibilityOperation) Finish(ctx context.Context, sourceRoot string) error {
	if err := op.store.RestoreLegacyArticleMasterRows(ctx, op.articleMasterRows); err != nil {
		return err
	}
	if err := op.store.RestoreBundledFunctionSymbols(ctx, op.bundledSymbols); err != nil {
		return err
	}
	if err := op.store.RestoreMissingDecoderLibrary(ctx); err != nil {
		return err
	}
	if err := op.store.ReconcileBundledMasterIdentities(ctx, op.bundledIdentities); err != nil {
		return err
	}
	if err := op.store.BackfillTrackGeometrySnapshots(ctx); err != nil {
		return err
	}
	if err := op.restoreLegacyAccessoryDefaults(ctx); err != nil {
		return err
	}
	if err := op.store.RetireLegacyAllocations(ctx); err != nil {
		return err
	}
	if err := op.store.BackfillVehicleCatalog(ctx); err != nil {
		return err
	}
	if err := op.store.RestoreCatalogProjection(ctx, op.doc.Version < 26); err != nil {
		return err
	}
	return op.store.AdoptLegacyDecoderSnapshots(ctx, sourceRoot)
}

func (op *CompatibilityOperation) restoreLegacyAccessoryDefaults(ctx context.Context) error {
	if op.doc.Version >= 3 {
		return nil
	}
	for _, row := range op.doc.Tables["accessory_products"] {
		productID, valid := BackupNonEmptyString(row["id"])
		if !valid {
			continue
		}
		if _, explicit := row["article_type"]; !explicit {
			if err := op.store.DefaultAccessoryArticleType(ctx, productID); err != nil {
				return err
			}
		}
		if _, explicit := row["subtype"]; !explicit {
			if err := op.store.DefaultAccessorySubtype(ctx, productID); err != nil {
				return err
			}
		}
		if _, explicit := row["inventory_strategy"]; !explicit {
			if err := op.store.DefaultAccessoryInventoryStrategy(ctx, productID); err != nil {
				return err
			}
		}
	}
	return nil
}
