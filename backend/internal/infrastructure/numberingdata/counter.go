package numberingdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/inventory/numbering"
)

type Counter struct{ tx *sql.Tx }

func NewCounter(tx *sql.Tx) *Counter { return &Counter{tx: tx} }
func (c *Counter) Advance(ctx context.Context, category string, next int) error {
	if _, err := c.tx.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET next_number=?, updated_at=?
WHERE category=?
`, next, time.Now().UTC().Format(time.RFC3339), category); err != nil {
		return fmt.Errorf("advance inventory number scheme: %w", err)
	}
	return nil
}
func (c *Counter) Active(
	ctx context.Context,
	category string,
) (*numbering.Scheme, error) {
	var scheme numbering.Scheme
	var active int
	err := c.tx.QueryRowContext(ctx, `
SELECT id, category, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category=? AND active=1
`, category).Scan(
		&scheme.ID,
		&scheme.Category,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, numbering.ErrNotFound
		}
		return nil, fmt.Errorf("read inventory number scheme: %w", err)
	}
	scheme.Active = active == 1
	return &scheme, nil
}
