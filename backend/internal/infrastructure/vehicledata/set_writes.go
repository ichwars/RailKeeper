package vehicledata

import (
	"context"
	"fmt"
	"strings"

	vehiclemodel "railkeeper/backend/internal/inventory/vehicle"
)

func (m *setMutation) InsertSet(ctx context.Context, write vehiclemodel.SetWrite) error {
	setID, inventoryNumber, input, now := write.ID, write.InventoryNumber, write.Input, write.At
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO vehicle_sets(
  id, inventory_number, name, manufacturer, article_number, article_source_url, gauge, epoch, railway_company,
  category, gattung, description, ean, production_period, list_price, acquisition_type,
  acquired_from, purchase_price, purchase_date, storage_location, storage_details, condition,
  condition_details, packaging, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, setID, inventoryNumber, input.Name, input.Manufacturer, input.ArticleNumber, input.ArticleSourceURL, input.Gauge,
		input.Epoch, input.RailwayCompany, input.Category, input.Gattung, input.Description, input.EAN,
		input.ProductionPeriod, input.ListPrice, input.AcquisitionType, input.AcquiredFrom,
		input.PurchasePrice, input.PurchaseDate, input.StorageLocation, input.StorageDetails,
		input.Condition, input.ConditionDetails, input.Packaging, now, now); err != nil {
		return fmt.Errorf("insert vehicle set: %w", err)
	}
	return nil
}

func (m *setMutation) InsertSetMember(ctx context.Context, member vehiclemodel.SetMemberWrite) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
VALUES(?, ?, ?, ?)
`, member.SetID, member.VehicleID, member.Position, member.Label); err != nil {
		return fmt.Errorf("insert vehicle set member: %w", err)
	}
	return nil
}
func (m *setMutation) LinkCreatedSetMember(ctx context.Context, setID, vehicleID string) error {
	return LinkCreatedSetMember(ctx, m.tx, setID, vehicleID)
}
func (m *setMutation) UpdateSet(ctx context.Context, id string, input vehiclemodel.VehicleSetInput, now string) (int64, error) {
	result, err := m.tx.ExecContext(ctx, `
UPDATE vehicle_sets
SET name=?, manufacturer=?, article_number=?, article_source_url=?, gauge=?, epoch=?, railway_company=?,
    category=?, gattung=?, description=?, ean=?, production_period=?, list_price=?, acquisition_type=?,
    acquired_from=?, purchase_price=?, purchase_date=?, storage_location=?, storage_details=?, condition=?,
    condition_details=?, packaging=?, updated_at=?
WHERE id=?
`, input.Name, input.Manufacturer, input.ArticleNumber, input.ArticleSourceURL, input.Gauge, input.Epoch,
		input.RailwayCompany, input.Category, input.Gattung, input.Description, input.EAN, input.ProductionPeriod,
		input.ListPrice, input.AcquisitionType, input.AcquiredFrom, input.PurchasePrice, input.PurchaseDate,
		input.StorageLocation, input.StorageDetails, input.Condition, input.ConditionDetails, input.Packaging, now, id)
	if err != nil {
		return 0, fmt.Errorf("update vehicle set: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read vehicle set update result: %w", err)
	}
	return affected, nil
}
func (m *setMutation) AuditSetCreated(ctx context.Context, audit vehiclemodel.SetAudit) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleSetCreated', 'vehicle_set', ?, ?, '{}')
`, audit.ID, audit.Actor, audit.SetID, audit.At); err != nil {
		return fmt.Errorf("write vehicle set audit log: %w", err)
	}
	return nil
}
func (m *setMutation) AuditSetUpdated(ctx context.Context, audit vehiclemodel.SetAudit) error {
	if _, err := m.tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleSetUpdated', 'vehicle_set', ?, ?, '{}')
`, audit.ID, audit.Actor, audit.SetID, audit.At); err != nil {
		return fmt.Errorf("write vehicle set update audit log: %w", err)
	}
	return nil
}
func (m *setMutation) SetInventoryNumberInUse(ctx context.Context, inventoryNumber string) (bool, error) {
	var count int
	if err := m.tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vehicle_sets WHERE inventory_number=?
`, strings.TrimSpace(inventoryNumber)).Scan(&count); err != nil {
		return false, fmt.Errorf("check vehicle set inventory number availability: %w", err)
	}
	return count > 0, nil
}
