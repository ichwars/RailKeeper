package numberingdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"railkeeper/backend/internal/credentials"
	"railkeeper/backend/internal/inventory/numbering"
)

type Store struct{ db *sql.DB }

func New(db *sql.DB) *Store { return &Store{db: db} }
func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func (s *Store) List(ctx context.Context) ([]numbering.Scheme, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, category, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
ORDER BY
  CASE category
    WHEN 'Fahrzeug' THEN 0
    WHEN 'Lokomotive' THEN 1
    WHEN 'Wagen' THEN 2
    ELSE 3
  END,
  category ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list inventory number schemes: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := []numbering.Scheme{}
	for rows.Next() {
		scheme, err := scanScheme(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, scheme)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inventory number schemes: %w", err)
	}
	return out, nil
}

func (s *Store) Create(ctx context.Context, category string, schemeInput numbering.SchemeInput) (*numbering.Scheme, error) {
	var existing int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM inventory_number_schemes WHERE category=?`, category).Scan(&existing); err != nil {
		return nil, fmt.Errorf("check inventory number scheme: %w", err)
	}
	if existing > 0 {
		return nil, numbering.ErrConflict
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO inventory_number_schemes(id, category, prefix, next_number, padding, active, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, credentials.RandomID(), category, schemeInput.Prefix, schemeInput.NextNumber, schemeInput.Padding, boolToInt(schemeInput.Active), now, now); err != nil {
		return nil, fmt.Errorf("create inventory number scheme: %w", err)
	}
	return s.get(ctx, category)
}

func (s *Store) Update(ctx context.Context, category string, input numbering.SchemeInput) (*numbering.Scheme, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `
UPDATE inventory_number_schemes
SET prefix=?, next_number=?, padding=?, active=?, updated_at=?
WHERE category=?
`, input.Prefix, input.NextNumber, input.Padding, boolToInt(input.Active), now, category)
	if err != nil {
		return nil, fmt.Errorf("update inventory number scheme: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read inventory number scheme update result: %w", err)
	}
	if affected == 0 {
		return nil, numbering.ErrNotFound
	}

	return s.get(ctx, category)
}

func (s *Store) get(ctx context.Context, category string) (*numbering.Scheme, error) {
	var scheme numbering.Scheme
	var active int
	if err := s.db.QueryRowContext(ctx, `
SELECT id, category, prefix, next_number, padding, active, created_at, updated_at
FROM inventory_number_schemes
WHERE category=?
`, category).Scan(
		&scheme.ID,
		&scheme.Category,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, numbering.ErrNotFound
		}
		return nil, fmt.Errorf("get inventory number scheme: %w", err)
	}
	scheme.Active = active == 1
	scheme.Preview = numbering.Format(scheme.Prefix, scheme.NextNumber, scheme.Padding)
	return &scheme, nil
}

type schemeScanner interface {
	Scan(dest ...any) error
}

func scanScheme(scanner schemeScanner) (numbering.Scheme, error) {
	var scheme numbering.Scheme
	var active int
	if err := scanner.Scan(
		&scheme.ID,
		&scheme.Category,
		&scheme.Prefix,
		&scheme.NextNumber,
		&scheme.Padding,
		&active,
		&scheme.CreatedAt,
		&scheme.UpdatedAt,
	); err != nil {
		return numbering.Scheme{}, fmt.Errorf("scan inventory number scheme: %w", err)
	}
	scheme.Active = active == 1
	scheme.Preview = numbering.Format(scheme.Prefix, scheme.NextNumber, scheme.Padding)
	return scheme, nil
}
