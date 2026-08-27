package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"railkeeper/backend/internal/application"
)

func transferVehicleSetSnapshot(ctx context.Context, tx *sql.Tx) ([]application.TransferVehicleSet, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, inventory_number, name, manufacturer, article_number, article_source_url, gauge, epoch,
       railway_company, category, gattung, description, ean, production_period, list_price,
       acquisition_type, acquired_from, purchase_price, purchase_date, storage_location,
       storage_details, condition, condition_details, packaging, created_at, updated_at
FROM vehicle_sets
ORDER BY inventory_number COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer vehicle sets: %w", err)
	}
	sets := []application.TransferVehicleSet{}
	setIndexes := map[string]int{}
	for rows.Next() {
		set := application.TransferVehicleSet{}
		if err := rows.Scan(
			&set.ID, &set.InventoryNumber, &set.Name, &set.Manufacturer, &set.ArticleNumber,
			&set.ArticleSourceURL, &set.Gauge, &set.Epoch, &set.RailwayCompany, &set.Category,
			&set.Gattung, &set.Description, &set.EAN, &set.ProductionPeriod, &set.ListPrice,
			&set.AcquisitionType, &set.AcquiredFrom, &set.PurchasePrice, &set.PurchaseDate,
			&set.StorageLocation, &set.StorageDetails, &set.Condition, &set.ConditionDetails,
			&set.Packaging, &set.CreatedAt, &set.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan transfer vehicle set: %w", err)
		}
		setIndexes[set.ID] = len(sets)
		sets = append(sets, set)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate transfer vehicle sets: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close transfer vehicle sets: %w", err)
	}

	memberRows, err := tx.QueryContext(ctx, `
SELECT member.vehicle_set_id, member.vehicle_id, vehicle.inventory_number, member.position, member.label
FROM vehicle_set_members member
JOIN vehicles vehicle ON vehicle.id=member.vehicle_id
ORDER BY member.vehicle_set_id, member.position, member.vehicle_id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer vehicle set members: %w", err)
	}
	defer func() { _ = memberRows.Close() }()
	for memberRows.Next() {
		var setID string
		member := application.TransferVehicleSetMember{}
		if err := memberRows.Scan(
			&setID, &member.SourceVehicleID, &member.VehicleInventoryNumber, &member.Position, &member.Label,
		); err != nil {
			return nil, fmt.Errorf("scan transfer vehicle set member: %w", err)
		}
		index, found := setIndexes[setID]
		if !found {
			return nil, fmt.Errorf("transfer vehicle set member references unknown set %q", setID)
		}
		sets[index].Members = append(sets[index].Members, member)
	}
	if err := memberRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer vehicle set members: %w", err)
	}
	return sets, nil
}
