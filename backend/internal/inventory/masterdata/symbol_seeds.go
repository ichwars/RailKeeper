package masterdata

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// SeedFunctionSymbolCodes corrects only the compatibility code of pristine bundled metadata.
// Artwork, local labels, activation, ordering and vehicle assignments are preserved.
func SeedFunctionSymbolCodes(ctx context.Context, store FunctionSymbolSeedStore, now string, manifest []byte) error {
	var document struct {
		Symbols []struct {
			Key            string `json:"key"`
			PreviousSHA256 string `json:"previousSHA256"`
			ECoSCode       *int   `json:"ecosCode"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(manifest, &document); err != nil {
		return fmt.Errorf("parse bundled function symbol codes: %w", err)
	}
	for _, symbol := range document.Symbols {
		state, found, err := store.FunctionSymbolSeedState(ctx, symbol.Key)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		previous, origin := state.MetadataJSON, state.Origin

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
		if err := store.CorrectFunctionSymbolSeed(ctx, SeedMetadataChange{Type: "symbols", Key: symbol.Key, PreviousJSON: previous, MetadataJSON: string(updated), At: now}); err != nil {
			return err
		}
	}
	return nil
}
