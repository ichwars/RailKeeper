package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
)

const layoutTechnicalPositionSelect = `
SELECT id, layout_unit_id, label, kind, position_x_mm, position_y_mm, rotation_degrees,
       COALESCE(product_id, ''), description, version, archived, created_at, updated_at
FROM layout_technical_positions`

func (r *LayoutRepository) ListTechnicalPositions(
	ctx context.Context,
	layoutUnitID string,
) ([]application.LayoutTechnicalPosition, error) {
	exists, err := layoutRecordExists(ctx, r.db, "layout_units", layoutUnitID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, application.ErrLayoutNotFound
	}

	rows, err := r.db.QueryContext(
		ctx,
		layoutTechnicalPositionSelect+` WHERE layout_unit_id=? ORDER BY label COLLATE NOCASE, id`,
		layoutUnitID,
	)
	if err != nil {
		return nil, fmt.Errorf("list layout technical positions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	positions := []application.LayoutTechnicalPosition{}
	for rows.Next() {
		position, err := scanLayoutTechnicalPosition(rows)
		if err != nil {
			return nil, fmt.Errorf("scan layout technical position: %w", err)
		}
		positions = append(positions, *position)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate layout technical positions: %w", err)
	}
	return positions, nil
}

func (r *LayoutRepository) CreateTechnicalPosition(
	ctx context.Context,
	layoutUnitID string,
	input application.CreateLayoutTechnicalPositionInput,
	actor string,
) (*application.LayoutTechnicalPosition, error) {
	now := timestamp()
	position := &application.LayoutTechnicalPosition{
		ID: randomID(), LayoutUnitID: layoutUnitID, Label: input.Label, Kind: input.Kind,
		PositionXMM: input.PositionXMM, PositionYMM: input.PositionYMM,
		RotationDegrees: input.RotationDegrees, ProductID: input.ProductID,
		Description: input.Description, Version: 1, Archived: input.Archived,
		CreatedAt: now, UpdatedAt: now,
	}

	err := r.withTx(ctx, func(tx *sql.Tx) error {
		exists, err := recordExists(ctx, tx, "layout_units", layoutUnitID)
		if err != nil {
			return err
		}
		if !exists {
			return application.ErrLayoutNotFound
		}
		if err := requireTechnicalPositionProduct(ctx, tx, input.ProductID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO layout_technical_positions(
  id, layout_unit_id, label, kind, position_x_mm, position_y_mm, rotation_degrees,
  product_id, description, version, archived, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, 1, ?, ?, ?)`,
			position.ID, position.LayoutUnitID, position.Label, position.Kind,
			position.PositionXMM, position.PositionYMM, position.RotationDegrees,
			position.ProductID, position.Description, boolToInt(position.Archived), now, now); err != nil {
			return fmt.Errorf("insert layout technical position: %w", err)
		}
		return writeLayoutAudit(
			ctx, tx, "LayoutTechnicalPositionCreated", "layout_technical_position", position.ID, actor, now,
		)
	})
	if err != nil {
		return nil, err
	}
	return position, nil
}

func (r *LayoutRepository) UpdateTechnicalPosition(
	ctx context.Context,
	id string,
	input application.UpdateLayoutTechnicalPositionInput,
	actor string,
) (*application.LayoutTechnicalPosition, error) {
	now := timestamp()
	err := r.withTx(ctx, func(tx *sql.Tx) error {
		if err := requireTechnicalPositionProduct(ctx, tx, input.ProductID); err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
UPDATE layout_technical_positions
SET label=?, kind=?, position_x_mm=?, position_y_mm=?, rotation_degrees=?,
    product_id=NULLIF(?, ''), description=?, archived=?, version=version+1, updated_at=?
WHERE id=? AND version=?`, input.Label, input.Kind, input.PositionXMM, input.PositionYMM,
			input.RotationDegrees, input.ProductID, input.Description, boolToInt(input.Archived),
			now, id, input.ExpectedVersion)
		if err != nil {
			return fmt.Errorf("update layout technical position: %w", err)
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("read layout technical position update result: %w", err)
		}
		if affected == 0 {
			exists, err := layoutRecordExists(ctx, tx, "layout_technical_positions", id)
			if err != nil {
				return err
			}
			if !exists {
				return application.ErrLayoutPositionNotFound
			}
			return application.ErrLayoutPositionVersionConflict
		}
		return writeLayoutAudit(
			ctx, tx, "LayoutTechnicalPositionUpdated", "layout_technical_position", id, actor, now,
		)
	})
	if err != nil {
		return nil, err
	}
	return r.getTechnicalPosition(ctx, id)
}

func (r *LayoutRepository) getTechnicalPosition(
	ctx context.Context,
	id string,
) (*application.LayoutTechnicalPosition, error) {
	position, err := scanLayoutTechnicalPosition(
		r.db.QueryRowContext(ctx, layoutTechnicalPositionSelect+` WHERE id=?`, id),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, application.ErrLayoutPositionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get layout technical position: %w", err)
	}
	return position, nil
}

func scanLayoutTechnicalPosition(scanner rowScanner) (*application.LayoutTechnicalPosition, error) {
	position := &application.LayoutTechnicalPosition{}
	var archived int
	err := scanner.Scan(
		&position.ID, &position.LayoutUnitID, &position.Label, &position.Kind,
		&position.PositionXMM, &position.PositionYMM, &position.RotationDegrees,
		&position.ProductID, &position.Description, &position.Version, &archived,
		&position.CreatedAt, &position.UpdatedAt,
	)
	position.Archived = archived != 0
	return position, err
}

func requireTechnicalPositionProduct(ctx context.Context, tx *sql.Tx, productID string) error {
	if productID == "" {
		return nil
	}
	exists, err := layoutRecordExists(ctx, tx, "accessory_products", productID)
	if err != nil {
		return err
	}
	if !exists {
		return application.ErrLayoutPositionProductNotFound
	}
	return nil
}

type layoutQueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func layoutRecordExists(ctx context.Context, db layoutQueryRower, table, id string) (bool, error) {
	query := map[string]string{
		"layout_units":               `SELECT COUNT(*) FROM layout_units WHERE id=?`,
		"layout_technical_positions": `SELECT COUNT(*) FROM layout_technical_positions WHERE id=?`,
		"accessory_products":         `SELECT COUNT(*) FROM accessory_products WHERE id=?`,
	}[table]
	if query == "" {
		return false, fmt.Errorf("unsupported layout twin table %q", table)
	}
	var count int
	if err := db.QueryRowContext(ctx, query, id).Scan(&count); err != nil {
		return false, fmt.Errorf("check %s record: %w", table, err)
	}
	return count > 0, nil
}
