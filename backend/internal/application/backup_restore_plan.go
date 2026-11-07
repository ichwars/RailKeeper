package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/infrastructure/backupdata"
)

type legacyBackupRestorePlanner struct{}
type legacyBackupRestoreOperation struct {
	doc         *BackupDocument
	beforeTable func(context.Context, *sql.Tx, string) error
	finish      func(context.Context, *sql.Tx, string) error
}

func BackupRestorePlanner() backupdata.RestorePlanner { return legacyBackupRestorePlanner{} }
func (legacyBackupRestorePlanner) Begin(ctx context.Context, tx *sql.Tx, doc *BackupDocument) (backupdata.RestoreOperation, error) {
	if err := suspendVehicleCatalogProjectionTx(ctx, tx); err != nil {
		return nil, err
	}
	articleMasterData, err := readLegacyRestoreArticleMasterData(ctx, tx, doc.Version)
	if err != nil {
		return nil, err
	}
	bundledMasterData, err := readBundledMasterDataIdentities(ctx, tx)
	if err != nil {
		return nil, err
	}
	bundledFunctionSymbols, err := readBundledFunctionSymbols(ctx, tx)
	if err != nil {
		return nil, err
	}
	articleInventoryScheme, err := readBackupArticleInventoryScheme(ctx, tx)
	if err != nil {
		return nil, err
	}

	return legacyBackupRestoreOperation{doc: doc, beforeTable: func(ctx context.Context, tx *sql.Tx, table string) error {
		if table == "vehicle_sets" {
			if err := prepareBackupVehicleSetInventoryNumbers(ctx, tx, doc); err != nil {
				return err
			}
		}
		if table == "accessory_products" {
			if err := prepareBackupArticleInventoryNumbers(ctx, tx, doc, articleInventoryScheme); err != nil {
				return err
			}
		}
		return nil
	}, finish: func(ctx context.Context, tx *sql.Tx, sourceRoot string) error {
		if err := restoreLegacyArticleMasterData(ctx, tx, articleMasterData); err != nil {
			return err
		}
		if err := restoreBundledFunctionSymbols(ctx, tx, bundledFunctionSymbols); err != nil {
			return err
		}
		if err := restoreMissingDecoderLibrary(ctx, tx); err != nil {
			return err
		}
		if err := reconcileBundledMasterDataIdentities(ctx, tx, bundledMasterData); err != nil {
			return err
		}
		if err := backfillRestoredTrackGeometrySnapshots(ctx, tx); err != nil {
			return err
		}
		if err := restoreLegacyAccessoryProductDefaults(ctx, tx, doc); err != nil {
			return err
		}

		if err := retireRestoredLegacyAllocations(ctx, tx); err != nil {
			return err
		}
		if err := backfillVehicleCatalogTx(ctx, tx); err != nil {
			return err
		}
		if err := restoreVehicleCatalogProjectionTx(ctx, tx, doc.Version < 26); err != nil {
			return err
		}
		if err := backfillLegacyDecoderCVTx(ctx, tx, sourceRoot); err != nil {
			return err
		}
		return nil
	}}, nil
}
func (op legacyBackupRestoreOperation) Rows(ctx context.Context, tx *sql.Tx, table string) ([]map[string]any, error) {
	if err := op.beforeTable(ctx, tx, table); err != nil {
		return nil, err
	}
	rows := backupRowsForRestore(op.doc, table)
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		result = append(result, backupRowForRestore(table, row))
	}
	return result, nil
}
func (op legacyBackupRestoreOperation) ProfileRow(ctx context.Context, tx *sql.Tx, row map[string]any) (map[string]any, error) {
	return backupProfileRowForRestore(ctx, tx, row)
}
func (op legacyBackupRestoreOperation) Finish(ctx context.Context, tx *sql.Tx, root string) error {
	return op.finish(ctx, tx, root)
}
