package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"
)

func prepareBackupVehicleSetInventoryNumbers(
	ctx context.Context,
	tx *sql.Tx,
	doc *BackupDocument,
) error {
	if doc.Version >= 18 || len(doc.Tables["vehicle_sets"]) == 0 {
		return nil
	}

	scheme, err := readBackupVehicleSetInventoryScheme(ctx, tx)
	if errors.Is(err, ErrInventoryNumberNotFound) {
		now := time.Now().UTC().Format(time.RFC3339)
		if _, err = tx.ExecContext(ctx, `
INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
)
VALUES(?, 'Set', 'RK-SET', 1, 6, 1, ?, ?)
ON CONFLICT(category) DO NOTHING
`, randomID(), now, now); err != nil {
			return fmt.Errorf("create legacy vehicle set inventory scheme: %w", err)
		}
		scheme, err = readBackupVehicleSetInventoryScheme(ctx, tx)
	}
	if err != nil {
		return fmt.Errorf("read legacy vehicle set inventory scheme: %w", err)
	}
	if !scheme.Active {
		if _, err = tx.ExecContext(ctx, `
UPDATE inventory_number_schemes SET active=1, updated_at=? WHERE category='Set'
`, time.Now().UTC().Format(time.RFC3339)); err != nil {
			return fmt.Errorf("activate legacy vehicle set inventory scheme: %w", err)
		}
		scheme.Active = true
	}

	originalRows := doc.Tables["vehicle_sets"]
	normalizedRows := make([]map[string]any, len(originalRows))
	for index, row := range originalRows {
		copyRow := make(map[string]any, len(row)+1)
		for column, value := range row {
			copyRow[column] = value
		}
		normalizedRows[index] = copyRow
	}
	sort.SliceStable(normalizedRows, func(left, right int) bool {
		leftCreated := fmt.Sprint(normalizedRows[left]["created_at"])
		rightCreated := fmt.Sprint(normalizedRows[right]["created_at"])
		if leftCreated != rightCreated {
			return leftCreated < rightCreated
		}
		return fmt.Sprint(normalizedRows[left]["id"]) < fmt.Sprint(normalizedRows[right]["id"])
	})
	for index, row := range normalizedRows {
		row["inventory_number"] = formatInventoryNumber(scheme.Prefix, scheme.NextNumber+index, scheme.Padding)
	}
	doc.Tables["vehicle_sets"] = normalizedRows
	if _, err := tx.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET next_number=?, updated_at=?
WHERE category='Set'
`, scheme.NextNumber+len(normalizedRows), time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("advance legacy vehicle set inventory scheme: %w", err)
	}
	return nil
}

func readBackupVehicleSetInventoryScheme(ctx context.Context, tx *sql.Tx) (*InventoryNumberScheme, error) {
	var scheme InventoryNumberScheme
	var active int
	err := tx.QueryRowContext(ctx, `
SELECT id, category, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category='Set'
`).Scan(
		&scheme.ID,
		&scheme.Category,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrInventoryNumberNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read inventory number scheme: %w", err)
	}
	scheme.Active = active == 1
	return &scheme, nil
}
