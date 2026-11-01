package application

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// Generated from the artwork manifest; historical migrations remain unchanged.
//
//go:embed function_symbol_ecos_codes.json
var functionSymbolCodesJSON []byte

// SeedFunctionSymbolCodes corrects only the compatibility code of pristine bundled metadata.
// Artwork, local labels, activation, ordering and vehicle assignments are preserved.
func SeedFunctionSymbolCodes(ctx context.Context, tx *sql.Tx, now string) error {
	var document struct {
		Symbols []struct {
			Key            string `json:"key"`
			PreviousSHA256 string `json:"previousSHA256"`
			ECoSCode       *int   `json:"ecosCode"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(functionSymbolCodesJSON, &document); err != nil {
		return fmt.Errorf("parse bundled function symbol codes: %w", err)
	}
	for _, symbol := range document.Symbols {
		var previous, origin string
		err := tx.QueryRowContext(ctx, `SELECT metadata_json,origin FROM master_data_entries
 WHERE type='symbols' AND key=?`, symbol.Key).Scan(&previous, &origin)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read bundled function symbol code: %w", err)
		}
		var metadata map[string]any
		if origin != "bundled" || json.Unmarshal([]byte(previous), &metadata) != nil {
			continue
		}
		canonical, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		digest := sha256.Sum256(canonical)
		if hex.EncodeToString(digest[:]) != symbol.PreviousSHA256 {
			continue
		}
		delete(metadata, "ecosCode")
		if symbol.ECoSCode != nil {
			metadata["ecosCode"] = *symbol.ECoSCode
		}
		updated, err := json.Marshal(metadata)
		if err != nil {
			return err
		}
		if string(updated) == string(canonical) {
			continue
		}
		if _, err := tx.ExecContext(ctx, `UPDATE master_data_entries SET metadata_json=?,updated_at=?
 WHERE type='symbols' AND key=? AND origin='bundled' AND metadata_json=?`,
			string(updated), now, symbol.Key, previous); err != nil {
			return fmt.Errorf("correct bundled function symbol code: %w", err)
		}
	}
	return nil
}
