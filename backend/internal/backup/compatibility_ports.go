package backup

import (
	"context"

	"railkeeper/backend/internal/inventory/numbering"
)

type StoredMasterRow struct {
	ID, TypeName, Key, Label                   string
	Active, SortOrder                          int
	SourceURL                                  *string
	MetadataJSON, CreatedAt, UpdatedAt, Origin string
}
type BundledIdentity struct{ TypeName, Key string }
type ArticleInventoryScheme struct {
	ID, Prefix           string
	NextNumber, Padding  int
	Active               bool
	CreatedAt, UpdatedAt string
}

// LegacyRestoreStore uses the existing restore transaction. It owns no independent commit.
type LegacyRestoreStore interface {
	SuspendCatalogProjection(context.Context) error
	LegacyArticleMasterRows(context.Context) ([]StoredMasterRow, error)
	BundledMasterIdentities(context.Context) ([]BundledIdentity, error)
	BundledFunctionSymbols(context.Context) ([]StoredMasterRow, error)
	ArticleScheme(context.Context) (*ArticleInventoryScheme, error)
	InsertArticleScheme(context.Context, ArticleInventoryScheme) error
	AdvanceArticleScheme(context.Context, int, string) error
	VehicleSetScheme(context.Context) (*numbering.Scheme, error)
	CreateLegacyVehicleSetScheme(context.Context, string, string) error
	ActivateLegacyVehicleSetScheme(context.Context, string) error
	AdvanceVehicleSetScheme(context.Context, int, string) error
	RestoreLegacyArticleMasterRows(context.Context, []StoredMasterRow) error
	RestoreBundledFunctionSymbols(context.Context, []StoredMasterRow) error
	RestoreMissingDecoderLibrary(context.Context) error
	ReconcileBundledMasterIdentities(context.Context, []BundledIdentity) error
	BackfillTrackGeometrySnapshots(context.Context) error
	DefaultAccessoryArticleType(context.Context, string) error
	DefaultAccessorySubtype(context.Context, string) error
	DefaultAccessoryInventoryStrategy(context.Context, string) error
	RetireLegacyAllocations(context.Context) error
	BackfillVehicleCatalog(context.Context) error
	RestoreCatalogProjection(context.Context, bool) error
	AdoptLegacyDecoderSnapshots(context.Context, string) error
	TransferProfileCreatorExists(context.Context, string) (bool, error)
}
