package workshop

import (
	"context"
	"strings"
	"time"

	"railkeeper/backend/internal/credentials"
)

func (s *Service) ListSpareParts(ctx context.Context, vehicleID string) ([]VehicleSparePart, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	if vehicleID == "" {
		return nil, ErrVehicleNotFound
	}
	if err := s.parent(ctx, vehicleID); err != nil {
		return nil, err
	}
	return s.repository.ListSpareParts(ctx, vehicleID)
}
func (s *Service) CreateSparePart(ctx context.Context, vehicleID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	input = cleanVehicleSparePartInput(input)
	if vehicleID == "" || !isValidVehicleSparePartInput(input) {
		return nil, ErrVehicleValidation
	}
	if err := s.parent(ctx, vehicleID); err != nil {
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
		ID:            credentials.RandomID(),
		VehicleID:     vehicleID,
		ArticleNumber: input.ArticleNumber,
		Description:   input.Description,
		Price:         input.Price,
		URL:           input.URL,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := s.repository.CreateSparePart(ctx, sparePart); err != nil {
		return nil, err
	}
	return &sparePart, nil
}
func (s *Service) UpdateSparePart(ctx context.Context, vehicleID, sparePartID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	vehicleID = strings.TrimSpace(vehicleID)
	sparePartID = strings.TrimSpace(sparePartID)
	input = cleanVehicleSparePartInput(input)
	if vehicleID == "" || sparePartID == "" || !isValidVehicleSparePartInput(input) {
		return nil, ErrVehicleValidation
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.repository.UpdateSparePart(ctx, vehicleID, sparePartID, input, now); err != nil {
		return nil, err
	}
	return s.GetSparePart(ctx, vehicleID, sparePartID)
}
func (s *Service) GetSparePart(ctx context.Context, vehicleID, id string) (*VehicleSparePart, error) {
	return s.repository.GetSparePart(ctx, vehicleID, id)
}
func (s *Service) DeleteSparePart(ctx context.Context, vehicleID, sparePartID string) (*VehicleSparePart, error) {
	sparePart, err := s.GetSparePart(ctx, vehicleID, sparePartID)
	if err != nil {
		return nil, err
	}
	if err := s.repository.DeleteSparePart(ctx, vehicleID, sparePartID); err != nil {
		return nil, err
	}
	return sparePart, nil
}
func (s *Service) findDuplicateSparePart(ctx context.Context, vehicleID string, input VehicleSparePartInput) (*VehicleSparePart, error) {
	key := vehicleSparePartDedupKey(input.ArticleNumber, input.Description, input.URL)
	if key == "" {
		return nil, nil
	}
	parts, err := s.repository.ListSpareParts(ctx, vehicleID)
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
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
