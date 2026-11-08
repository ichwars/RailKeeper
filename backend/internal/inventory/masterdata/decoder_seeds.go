package masterdata

import (
	"context"
	"encoding/json"

	"railkeeper/backend/internal/decoderlibrary"
)

// SeedDecoderLibrary adds missing documentation and upgrades only known, unedited bundled metadata.
// Activation, local labels, ordering and source URLs remain the operator's settings.
func SeedDecoderLibrary(ctx context.Context, store DecoderSeedStore, now string) error {
	if err := seedDecoderEntry(ctx, store, now, decoderlibrary.InventoryMasterDataType, decoderlibrary.InventoryKey,
		"Decoderbibliothek", "", 0, decoderlibrary.BundledInventory()); err != nil {
		return err
	}
	for index, profile := range decoderlibrary.Bundled().Profiles {
		if err := seedDecoderEntry(ctx, store, now, decoderlibrary.MasterDataType, profile.ID,
			profile.Name, profile.Sources[0].URL, index, profile); err != nil {
			return err
		}
	}
	for index, reference := range decoderlibrary.BundledReferences() {
		if err := seedDecoderEntry(ctx, store, now, decoderlibrary.ReferenceMasterDataType, reference.ID,
			reference.Title.DE, reference.Sources[0].URL, index, reference); err != nil {
			return err
		}
	}
	return nil
}

func seedDecoderEntry(ctx context.Context, store DecoderSeedStore, now, typeName, key, label, source string, order int, value any) error {
	metadata, err := json.Marshal(value)
	if err != nil {
		return err
	}
	state, found, err := store.DecoderSeedState(ctx, typeName, key)
	if err != nil {
		return err
	}
	if found {
		if state.MetadataJSON == string(metadata) || state.Origin != "bundled" || !decoderlibrary.PreviousBundled(typeName, key, state.MetadataJSON) {
			return nil
		}
		return store.UpgradeDecoderSeed(ctx, SeedMetadataChange{Type: typeName, Key: key, PreviousJSON: state.MetadataJSON, MetadataJSON: string(metadata), At: now})
	}
	return store.InsertDecoderSeed(ctx, DecoderSeedEntry{Type: typeName, Key: key, Label: label, SourceURL: source, SortOrder: order, MetadataJSON: string(metadata), At: now})
}
