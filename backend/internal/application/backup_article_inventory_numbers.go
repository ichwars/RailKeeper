package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type backupArticleInventoryScheme struct {
	ID         string
	Prefix     string
	NextNumber int
	Padding    int
	Active     bool
	CreatedAt  string
	UpdatedAt  string
}

func readBackupArticleInventoryScheme(
	ctx context.Context,
	tx *sql.Tx,
) (*backupArticleInventoryScheme, error) {
	var scheme backupArticleInventoryScheme
	var active int
	err := tx.QueryRowContext(ctx, `
SELECT id, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category='Artikel'
`).Scan(
		&scheme.ID,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read article inventory scheme before restore: %w", err)
	}
	scheme.Active = active == 1
	return &scheme, nil
}

func prepareBackupArticleInventoryNumbers(
	ctx context.Context,
	tx *sql.Tx,
	doc *BackupDocument,
	fallback *backupArticleInventoryScheme,
) error {
	rows := doc.Tables["accessory_products"]
	if len(rows) == 0 {
		return nil
	}
	scheme, err := ensureBackupArticleInventoryScheme(ctx, tx, fallback)
	if err != nil {
		return err
	}

	used := make(map[string]struct{}, len(rows))
	missing := make([]map[string]any, 0)
	for _, row := range rows {
		inventoryNumber := strings.TrimSpace(fmt.Sprint(row["inventory_number"]))
		if inventoryNumber == "<nil>" {
			inventoryNumber = ""
		}
		if inventoryNumber == "" {
			missing = append(missing, row)
			continue
		}
		if _, duplicate := used[inventoryNumber]; duplicate {
			return fmt.Errorf("prepare article inventory numbers: duplicate %q", inventoryNumber)
		}
		used[inventoryNumber] = struct{}{}
	}
	if len(missing) == 0 {
		return nil
	}
	if !scheme.Active {
		return ErrInventoryNumberNotFound
	}

	sort.SliceStable(missing, func(i, j int) bool {
		leftCreated := fmt.Sprint(missing[i]["created_at"])
		rightCreated := fmt.Sprint(missing[j]["created_at"])
		if leftCreated != rightCreated {
			return leftCreated < rightCreated
		}
		return fmt.Sprint(missing[i]["id"]) < fmt.Sprint(missing[j]["id"])
	})

	next := 1
	for _, row := range missing {
		for {
			candidate := formatInventoryNumber(scheme.Prefix, next, scheme.Padding)
			next++
			if _, exists := used[candidate]; exists {
				continue
			}
			row["inventory_number"] = candidate
			used[candidate] = struct{}{}
			break
		}
	}
	if next < scheme.NextNumber {
		next = scheme.NextNumber
	}
	if _, err := tx.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET next_number=?, updated_at=?
WHERE category='Artikel'
`, next, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("advance restored article inventory scheme: %w", err)
	}
	return nil
}

func ensureBackupArticleInventoryScheme(
	ctx context.Context,
	tx *sql.Tx,
	fallback *backupArticleInventoryScheme,
) (*backupArticleInventoryScheme, error) {
	scheme, err := readBackupArticleInventoryScheme(ctx, tx)
	if err != nil || scheme != nil {
		return scheme, err
	}
	if fallback == nil {
		now := time.Now().UTC().Format(time.RFC3339)
		fallback = &backupArticleInventoryScheme{
			ID: randomID(), Prefix: "RK-ART", NextNumber: 1, Padding: 6, Active: true,
			CreatedAt: now, UpdatedAt: now,
		}
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
) VALUES(?, 'Artikel', ?, ?, ?, ?, ?, ?)
`, fallback.ID, fallback.Prefix, fallback.NextNumber, fallback.Padding, boolToInt(fallback.Active),
		fallback.CreatedAt, fallback.UpdatedAt); err != nil {
		return nil, fmt.Errorf("restore article inventory scheme: %w", err)
	}
	return readBackupArticleInventoryScheme(ctx, tx)
}
