package vehicledata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

type setStore struct{ *coreStore }
type setMutation struct{ *coreMutation }

func NewSetStore(db *sql.DB) vehiclemodel.SetRepository {
	return &setStore{coreStore: &coreStore{db: db}}
}
func (s *setStore) BeginSet(ctx context.Context) (vehiclemodel.SetMutation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &setMutation{coreMutation: newCoreMutation(tx)}, nil
}
func (s *setStore) Set(ctx context.Context, id string) (vehiclemodel.VehicleSet, error) {
	var set vehiclemodel.VehicleSet
	err := s.db.QueryRowContext(ctx, `
SELECT id, inventory_number, name, manufacturer, COALESCE(article_number, ''),
       COALESCE(article_source_url, ''), gauge, COALESCE(epoch, ''), COALESCE(railway_company, ''),
       category, gattung, COALESCE(description, ''), COALESCE(ean, ''),
       COALESCE(production_period, ''), COALESCE(list_price, ''), COALESCE(acquisition_type, ''),
       COALESCE(acquired_from, ''), COALESCE(purchase_price, ''), COALESCE(purchase_date, ''),
       COALESCE(storage_location, ''), COALESCE(storage_details, ''), COALESCE(condition, ''),
       COALESCE(condition_details, ''), COALESCE(packaging, ''), created_at, updated_at
FROM vehicle_sets
WHERE id=?
`, id).Scan(
		&set.ID, &set.InventoryNumber, &set.Name, &set.Manufacturer, &set.ArticleNumber,
		&set.ArticleSourceURL, &set.Gauge, &set.Epoch, &set.RailwayCompany, &set.Category, &set.Gattung,
		&set.Description, &set.EAN, &set.ProductionPeriod, &set.ListPrice, &set.AcquisitionType,
		&set.AcquiredFrom, &set.PurchasePrice, &set.PurchaseDate, &set.StorageLocation, &set.StorageDetails,
		&set.Condition, &set.ConditionDetails, &set.Packaging, &set.CreatedAt, &set.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return set, vehiclemodel.ErrVehicleSetNotFound
		}
		return set, fmt.Errorf("get vehicle set: %w", err)
	}

	return set, nil
}
func (s *setStore) MemberIDs(ctx context.Context, id string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT vehicle_id
FROM vehicle_set_members
WHERE vehicle_set_id=?
ORDER BY position ASC
`, id)
	if err != nil {
		return nil, fmt.Errorf("list vehicle set members: %w", err)
	}
	memberIDs := []string{}
	for rows.Next() {
		var memberID string
		if err := rows.Scan(&memberID); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan vehicle set member: %w", err)
		}
		memberIDs = append(memberIDs, memberID)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close vehicle set members: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicle set members: %w", err)
	}
	return memberIDs, nil
}
func (s *setStore) MemberImages(ctx context.Context, id string) ([]vehiclemodel.VehicleImage, error) {
	return ReadImages(ctx, s.db, id)
}
