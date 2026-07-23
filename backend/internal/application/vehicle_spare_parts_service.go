package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

func (s *VehicleService) ListSpareParts(ctx context.Context, vehicleID string) ([]VehicleSparePart, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.loadVehicleSpareParts(ctx, vehicleID)
}

func (s *VehicleService) CreateSparePart(ctx context.Context, vehicleID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleSparePartInput(input)
	if vehicleID == "" || !isValidVehicleSparePartInput(input) {
		return nil, ErrVehicleValidation
	}
	if _, err := s.Get(ctx, vehicleID); err != nil {
		return nil, err
	}
	if existing, err := s.findDuplicateSparePart(ctx, vehicleID, input); err != nil {
		return nil, err
	} else if existing != nil {
		merged := VehicleSparePartInput{
			ArticleNumber: firstNonEmpty(existing.ArticleNumber, input.ArticleNumber),
			Description:   firstNonEmpty(existing.Description, input.Description),
			Price:         firstNonEmpty(existing.Price, input.Price),
			URL:           firstNonEmpty(existing.URL, input.URL),
		}
		return s.UpdateSparePart(ctx, vehicleID, existing.ID, merged)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	sparePart := VehicleSparePart{
		ID:            randomID(),
		VehicleID:     vehicleID,
		ArticleNumber: input.ArticleNumber,
		Description:   input.Description,
		Price:         input.Price,
		URL:           input.URL,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO vehicle_spare_parts(id, vehicle_id, article_number, description, price, url, created_at, updated_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, sparePart.ID, sparePart.VehicleID, sparePart.ArticleNumber, sparePart.Description, sparePart.Price, sparePart.URL, sparePart.CreatedAt, sparePart.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create vehicle spare part: %w", err)
	}
	return &sparePart, nil
}

func (s *VehicleService) findDuplicateSparePart(ctx context.Context, vehicleID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	key := vehicleSparePartDedupKey(input.ArticleNumber, input.Description, input.URL)
	if key == "" {
		return nil, nil
	}
	parts, err := s.loadVehicleSpareParts(ctx, vehicleID)
	if err != nil {
		return nil, err
	}
	for index := range parts {
		if vehicleSparePartDedupKey(parts[index].ArticleNumber, parts[index].Description, parts[index].URL) == key {
			return &parts[index], nil
		}
	}
	return nil, nil
}

func (s *VehicleService) UpdateSparePart(ctx context.Context, vehicleID, sparePartID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	sparePartID = strings.TrimSpace(sparePartID)
	input = cleanVehicleSparePartInput(input)
	if vehicleID == "" || sparePartID == "" || !isValidVehicleSparePartInput(input) {
		return nil, ErrVehicleValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := s.db.ExecContext(ctx, `
UPDATE vehicle_spare_parts
SET article_number=?, description=?, price=?, url=?, updated_at=?
WHERE id=? AND vehicle_id=?
`, input.ArticleNumber, input.Description, input.Price, input.URL, now, sparePartID, vehicleID)
	if err != nil {
		return nil, fmt.Errorf("update vehicle spare part: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle spare part update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return s.GetSparePart(ctx, vehicleID, sparePartID)
}

func (s *VehicleService) GetSparePart(ctx context.Context, vehicleID, sparePartID string) (*VehicleSparePart, error) {
	var sparePart VehicleSparePart
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
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle spare part: %w", err)
	}
	return &sparePart, nil
}

func (s *VehicleService) DeleteSparePart(ctx context.Context, vehicleID, sparePartID string) (*VehicleSparePart, error) {
	sparePart, err := s.GetSparePart(ctx, vehicleID, sparePartID)
	if err != nil {
		return nil, err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM vehicle_spare_parts WHERE id=? AND vehicle_id=?`, strings.TrimSpace(sparePartID), strings.TrimSpace(vehicleID))
	if err != nil {
		return nil, fmt.Errorf("delete vehicle spare part: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle spare part delete result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleNotFound
	}
	return sparePart, nil
}

func (s *VehicleService) loadVehicleSpareParts(ctx context.Context, vehicleID string) ([]VehicleSparePart, error) {
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

	entries := []VehicleSparePart{}
	for rows.Next() {
		var sparePart VehicleSparePart
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
