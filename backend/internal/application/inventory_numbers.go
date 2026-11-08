package application

import (
	"context"
	"database/sql"

	"railkeeper/backend/internal/infrastructure/numberingdata"
	"railkeeper/backend/internal/inventory/numbering"
)

var (
	ErrInventoryNumberValidation = numbering.ErrValidation
	ErrInventoryNumberConflict   = numbering.ErrConflict
	ErrInventoryNumberNotFound   = numbering.ErrNotFound
)

type InventoryNumberScheme = numbering.Scheme
type InventoryNumberSchemeInput = numbering.SchemeInput
type InventoryNumberSchemeCreateInput = numbering.CreateInput
type InventoryNumberAvailability = numbering.Availability
type InventoryNumberService = numbering.Service

func NewInventoryNumberService(db *sql.DB) *InventoryNumberService {
	return numbering.New(numberingdata.New(db))
}
func ReserveInventoryNumber(ctx context.Context, tx *sql.Tx, category, fallbackCategory string,
	ensureAvailable InventoryNumberAvailability) (string, error) {
	return numbering.Reserve(ctx, numberingdata.NewCounter(tx), category, fallbackCategory, ensureAvailable)
}
func formatInventoryNumber(prefix string, number, padding int) string {
	return numbering.Format(prefix, number, padding)
}
