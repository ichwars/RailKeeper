package application

import (
	"context"
	"database/sql"
	_ "embed"

	"railkeeper/backend/internal/inventory/masterdata"

	"railkeeper/backend/internal/infrastructure/masterdatastore"
)

// Generated from the artwork manifest; historical migrations remain unchanged.
//
//go:embed function_symbol_ecos_codes.json
var functionSymbolCodesJSON []byte

// SeedFunctionSymbolCodes participates in the caller's seed transaction and retains the original manifest.
func SeedFunctionSymbolCodes(ctx context.Context, tx *sql.Tx, now string) error {
	return masterdata.SeedFunctionSymbolCodes(ctx, masterdatastore.SeedStoreFor(tx), now, functionSymbolCodesJSON)
}
