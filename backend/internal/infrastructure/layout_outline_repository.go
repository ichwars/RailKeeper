package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/application"
)

func (r *LayoutRepository) UpdateUnitOutline(
	ctx context.Context,
	unitID string,
	input application.UpdateLayoutUnitOutlineInput,
	actor string,
) (*application.LayoutUnitOutline, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		result, err := tx.ExecContext(ctx, `
UPDATE layout_units SET version=version+1, updated_at=? WHERE id=? AND version=?`,
			now, unitID, input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("version layout unit outline: %w", err)
		}
		if err := requireUpdated(ctx, tx, result, "layout_units", unitID,
			application.ErrLayoutVersionConflict); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
DELETE FROM layout_unit_outline_points WHERE layout_unit_id=?`, unitID); err != nil {
			return fmt.Errorf("clear layout unit outline: %w", err)
		}
		for index, point := range input.Points {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO layout_unit_outline_points(layout_unit_id, point_index, position_x_mm, position_y_mm)
VALUES(?, ?, ?, ?)`, unitID, index, point.XMM, point.YMM); err != nil {
				return fmt.Errorf("insert layout unit outline point: %w", err)
			}
		}
		return writeLayoutAudit(ctx, tx, "LayoutUnitOutlineUpdated", "layout_unit", unitID, actor, now)
	})
	if err != nil {
		return nil, err
	}
	return &application.LayoutUnitOutline{
		LayoutUnitID: unitID, Points: input.Points, Version: input.ExpectedVersion + 1,
	}, nil
}
