package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/decoderlibrary"
)

// SeedDecoderLibrary adds missing documentation and upgrades only known, unedited bundled metadata.
// Activation, local labels, ordering and source URLs remain the operator's settings.
func SeedDecoderLibrary(ctx context.Context, tx *sql.Tx, now string) error {
	if err := seedDecoderEntry(ctx, tx, now, decoderlibrary.InventoryMasterDataType, decoderlibrary.InventoryKey,
		"Decoderbibliothek", "", 0, decoderlibrary.BundledInventory()); err != nil {
		return err
	}
	for index, profile := range decoderlibrary.Bundled().Profiles {
		if err := seedDecoderEntry(ctx, tx, now, decoderlibrary.MasterDataType, profile.ID,
			profile.Name, profile.Sources[0].URL, index, profile); err != nil {
			return err
		}
	}
	for index, reference := range decoderlibrary.BundledReferences() {
		if err := seedDecoderEntry(ctx, tx, now, decoderlibrary.ReferenceMasterDataType, reference.ID,
			reference.Title.DE, reference.Sources[0].URL, index, reference); err != nil {
			return err
		}
	}
	return nil
}

func seedDecoderEntry(ctx context.Context, tx *sql.Tx, now, typeName, key, label, source string, order int, value any) error {
	metadata, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var previous, origin string
	err = tx.QueryRowContext(ctx, `SELECT metadata_json,origin FROM master_data_entries WHERE type=? AND key=?`, typeName, key).
		Scan(&previous, &origin)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read decoder library seed: %w", err)
	}
	if err == nil {
		if previous == string(metadata) || origin != "bundled" || !decoderlibrary.PreviousBundled(typeName, key, previous) {
			return nil
		}
		_, err = tx.ExecContext(ctx, `UPDATE master_data_entries SET metadata_json=?,updated_at=?
 WHERE type=? AND key=? AND metadata_json=? AND origin='bundled'`, string(metadata), now, typeName, key, previous)
		if err != nil {
			return fmt.Errorf("upgrade bundled decoder documentation: %w", err)
		}
		return nil
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO master_data_entries
 (id,type,key,label,active,sort_order,source_url,metadata_json,created_at,updated_at,origin)
 VALUES(?,?,?,?,1,?,?,?,?,?,'bundled')`, typeName+":"+key, typeName, key, label, order, source, string(metadata), now, now)
	if err != nil {
		return fmt.Errorf("seed decoder documentation: %w", err)
	}
	return nil
}
