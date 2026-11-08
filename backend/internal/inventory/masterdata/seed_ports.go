package masterdata

import "context"

type SeedState struct{ MetadataJSON, Origin string }
type SeedMetadataChange struct{ Type, Key, PreviousJSON, MetadataJSON, At string }
type DecoderSeedEntry struct {
	Type, Key, Label, SourceURL, MetadataJSON, At string
	SortOrder                                     int
}
type DecoderSeedStore interface {
	DecoderSeedState(context.Context, string, string) (SeedState, bool, error)
	UpgradeDecoderSeed(context.Context, SeedMetadataChange) error
	InsertDecoderSeed(context.Context, DecoderSeedEntry) error
}
type FunctionSymbolSeedStore interface {
	FunctionSymbolSeedState(context.Context, string) (SeedState, bool, error)
	CorrectFunctionSymbolSeed(context.Context, SeedMetadataChange) error
}
type SeedStore interface {
	DecoderSeedStore
	FunctionSymbolSeedStore
}
