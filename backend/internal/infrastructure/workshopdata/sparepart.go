package workshopdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"railkeeper/backend/internal/domain"
)

func (s *Store) CreateSparePart(ctx context.Context, sparePart domain.VehicleSparePart) error {
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_spare_parts(id, vehicle_id, article_number, description, price, url, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, sparePart.ID, sparePart.VehicleID, sparePart.ArticleNumber, sparePart.Description, sparePart.Price, sparePart.URL, sparePart.CreatedAt, sparePart.UpdatedAt); err != nil {
		return fmt.Errorf("create vehicle spare part: %w", err)
	}
	return nil
}
func (s *Store) UpdateSparePart(ctx context.Context, vehicleID, sparePartID string, input domain.VehicleSparePartInput, now string) error {
	result, err := s.db.ExecContext(ctx, `
UPDATE vehicle_spare_parts
SET article_number=?, description=?, price=?, url=?, updated_at=?
WHERE id=? AND vehicle_id=?
`, input.ArticleNumber, input.Description, input.Price, input.URL, now, sparePartID, vehicleID)
	if err != nil {
		return fmt.Errorf("update vehicle spare part: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle spare part update result: %w", err)
	}
	if affected == 0 {
		return domain.ErrVehicleNotFound
	}
	return nil
}
func (s *Store) GetSparePart(ctx context.Context, vehicleID, sparePartID string) (*domain.VehicleSparePart, error) {
	var sparePart domain.VehicleSparePart
	err := s.db.QueryRowContext(ctx, `
SELECT id, vehicle_id, COALESCE(article_number, ''), COALESCE(description, ''), COALESCE(price, ''), COALESCE(url, ''), created_at, updated_at
FROM vehicle_spare_parts
WHERE id=? AND vehicle_id=?
`, strings.TrimSpace(sparePartID), strings.TrimSpace(vehicleID)).Scan(
		&sparePart.ID,
		&sparePart.VehicleID,
		&sparePart.ArticleNumber,
		&sparePart.Description,
		&sparePart.Price,
		&sparePart.URL,
		&sparePart.CreatedAt,
		&sparePart.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle spare part: %w", err)
	}
	return &sparePart, nil
}
func (s *Store) DeleteSparePart(ctx context.Context, vehicleID, sparePartID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_spare_parts WHERE id=? AND vehicle_id=?`, strings.TrimSpace(sparePartID), strings.TrimSpace(vehicleID))
	if err != nil {
		return fmt.Errorf("delete vehicle spare part: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read vehicle spare part delete result: %w", err)
	}
	if affected == 0 {
		return domain.ErrVehicleNotFound
	}
	return nil
}
func (s *Store) ListSpareParts(ctx context.Context, vehicleID string) ([]domain.VehicleSparePart, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, vehicle_id, COALESCE(article_number, ''), COALESCE(description, ''), COALESCE(price, ''), COALESCE(url, ''), created_at, updated_at
FROM vehicle_spare_parts
WHERE vehicle_id=?
ORDER BY article_number COLLATE NOCASE ASC, created_at DESC
`, strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("list vehicle spare parts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := []domain.VehicleSparePart{}
	for rows.Next() {
		var sparePart domain.VehicleSparePart
		if err := rows.Scan(&sparePart.ID, &sparePart.VehicleID, &sparePart.ArticleNumber, &sparePart.Description, &sparePart.Price, &sparePart.URL, &sparePart.CreatedAt, &sparePart.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan vehicle spare part: %w", err)
		}
		entries = append(entries, sparePart)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle spare parts: %w", err)
	}
	return entries, nil
}
