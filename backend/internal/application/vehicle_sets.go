package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const maxVehicleSetMembers = 100

func (s *VehicleService) CreateSet(
	ctx context.Context,
	input CreateVehicleSetInput,
	actorUserID string,
) (*VehicleSet, error) {
	setInput := cleanVehicleSetInput(input.Set)
	if !isValidVehicleSetInput(setInput) || len(input.Members) < 2 || len(input.Members) > maxVehicleSetMembers {
		return nil, ErrVehicleSetValidation
	}

	preparedMembers := make([]CreateVehicleInput, len(input.Members))
	memberIDs := make([]string, len(input.Members))
	for index, member := range input.Members {
		member = mergeVehicleSetMember(setInput, member, index)
		var err error
		member, err = validateCreateVehicleInput(member)
		if err != nil {
			return nil, err
		}
		memberIDs[index] = randomID()
		preparedMembers[index] = member
	}

	for index, member := range preparedMembers {
		var err error
		if s.imageLocalizer != nil && len(member.Images) > 0 {
			member.Images, err = s.imageLocalizer(ctx, memberIDs[index], member.Images)
			if err != nil {
				return nil, err
			}
		}
		preparedMembers[index] = member
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create vehicle set: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	now := time.Now().UTC().Format(time.RFC3339)
	setID := randomID()
	setInventoryNumber, err := s.nextVehicleSetInventoryNumber(ctx, tx)
	if err != nil {
		return nil, err
	}
	if err = insertVehicleSetTx(ctx, tx, setID, setInventoryNumber, setInput, now); err != nil {
		return nil, err
	}

	members := make([]Vehicle, 0, len(preparedMembers))
	for index, memberInput := range preparedMembers {
		member, createErr := s.insertVehicleTx(ctx, tx, memberIDs[index], memberInput, actorUserID, now)
		if createErr != nil {
			return nil, createErr
		}
		if _, err = tx.ExecContext(ctx, `
INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
VALUES(?, ?, ?, ?)
`, setID, member.ID, index+1, member.Name); err != nil {
			return nil, fmt.Errorf("insert vehicle set member: %w", err)
		}
		member.VehicleSetID = setID
		member.VehicleSetName = setInput.Name
		member.VehicleSetPosition = index + 1
		member.VehicleSetSize = len(preparedMembers)
		member.VehicleSet = vehicleSetSummary(setID, setInventoryNumber, setInput, len(preparedMembers), index+1)
		members = append(members, *member)
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleSetCreated', 'vehicle_set', ?, ?, '{}')
`, randomID(), actorUserID, setID, now); err != nil {
		return nil, fmt.Errorf("write vehicle set audit log: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create vehicle set: %w", err)
	}

	for index := range members {
		images, loadErr := s.loadVehicleImages(ctx, members[index].ID)
		if loadErr != nil {
			return nil, loadErr
		}
		members[index].Images = images
	}

	return vehicleSetFromInput(setID, setInventoryNumber, setInput, members, now), nil
}

func (s *VehicleService) GetSet(ctx context.Context, id string) (*VehicleSet, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrVehicleSetNotFound
	}

	var set VehicleSet
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
			return nil, ErrVehicleSetNotFound
		}
		return nil, fmt.Errorf("get vehicle set: %w", err)
	}

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
	if err := s.resetExpiredExhibitionFlags(ctx); err != nil {
		return nil, err
	}
	loadedMembers, err := s.list(ctx, "", id)
	if err != nil {
		return nil, fmt.Errorf("load vehicle set members: %w", err)
	}
	membersByID := make(map[string]Vehicle, len(loadedMembers))
	for _, member := range loadedMembers {
		membersByID[member.ID] = member
	}
	set.Members = make([]Vehicle, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		member, found := membersByID[memberID]
		if !found {
			return nil, fmt.Errorf("get vehicle set member %s: %w", memberID, ErrVehicleNotFound)
		}
		set.Members = append(set.Members, member)
	}
	return &set, nil
}

func (s *VehicleService) UpdateSet(
	ctx context.Context,
	id string,
	input VehicleSetInput,
	actorUserID string,
) (*VehicleSet, error) {
	id = strings.TrimSpace(id)
	input = cleanVehicleSetInput(input)
	if id == "" {
		return nil, ErrVehicleSetNotFound
	}
	if !isValidVehicleSetInput(input) {
		return nil, ErrVehicleSetValidation
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update vehicle set: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := tx.ExecContext(ctx, `
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
		return nil, fmt.Errorf("update vehicle set: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read vehicle set update result: %w", err)
	}
	if affected == 0 {
		return nil, ErrVehicleSetNotFound
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleSetUpdated', 'vehicle_set', ?, ?, '{}')
`, randomID(), actorUserID, id, now); err != nil {
		return nil, fmt.Errorf("write vehicle set update audit log: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update vehicle set: %w", err)
	}
	return s.GetSet(ctx, id)
}

func (s *VehicleService) nextVehicleSetInventoryNumber(ctx context.Context, tx *sql.Tx) (string, error) {
	return ReserveInventoryNumber(ctx, tx, "Set", "", func(candidate string) error {
		return s.ensureVehicleSetInventoryNumberAvailable(ctx, tx, candidate)
	})
}

func (s *VehicleService) ensureVehicleSetInventoryNumberAvailable(
	ctx context.Context,
	tx *sql.Tx,
	inventoryNumber string,
) error {
	var count int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM vehicle_sets WHERE inventory_number=?
`, strings.TrimSpace(inventoryNumber)).Scan(&count); err != nil {
		return fmt.Errorf("check vehicle set inventory number availability: %w", err)
	}
	if count > 0 {
		return ErrInventoryNumberConflict
	}
	return nil
}

func cleanVehicleSetInput(input VehicleSetInput) VehicleSetInput {
	input.Name = strings.TrimSpace(input.Name)
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.ArticleSourceURL = strings.TrimSpace(input.ArticleSourceURL)
	input.Gauge = strings.TrimSpace(input.Gauge)
	input.Epoch = strings.TrimSpace(input.Epoch)
	input.RailwayCompany = strings.TrimSpace(input.RailwayCompany)
	input.Category = strings.TrimSpace(input.Category)
	input.Gattung = strings.TrimSpace(input.Gattung)
	input.Description = strings.TrimSpace(input.Description)
	input.EAN = strings.TrimSpace(input.EAN)
	input.ProductionPeriod = strings.TrimSpace(input.ProductionPeriod)
	input.ListPrice = strings.TrimSpace(input.ListPrice)
	input.AcquisitionType = strings.TrimSpace(input.AcquisitionType)
	input.AcquiredFrom = strings.TrimSpace(input.AcquiredFrom)
	input.PurchasePrice = strings.TrimSpace(input.PurchasePrice)
	input.PurchaseDate = strings.TrimSpace(input.PurchaseDate)
	input.StorageLocation = strings.TrimSpace(input.StorageLocation)
	input.StorageDetails = strings.TrimSpace(input.StorageDetails)
	input.Condition = strings.TrimSpace(input.Condition)
	input.ConditionDetails = strings.TrimSpace(input.ConditionDetails)
	input.Packaging = strings.TrimSpace(input.Packaging)
	return input
}

func isValidVehicleSetInput(input VehicleSetInput) bool {
	return input.Name != "" && input.Manufacturer != "" && input.Gauge != "" && input.Category != "" &&
		input.Gattung != ""
}

func mergeVehicleSetMember(setInput VehicleSetInput, member CreateVehicleInput, index int) CreateVehicleInput {
	if strings.TrimSpace(member.Name) == "" {
		member.Name = fmt.Sprintf("%s (%d)", setInput.Name, index+1)
	}
	return applyVehicleSetDefaults(setInput, member)
}

func applyVehicleSetDefaults(setInput VehicleSetInput, member CreateVehicleInput) CreateVehicleInput {
	member.Manufacturer = vehicleSetDefault(member.Manufacturer, setInput.Manufacturer)
	member.ArticleNumber = vehicleSetDefault(member.ArticleNumber, setInput.ArticleNumber)
	member.ArticleSourceURL = vehicleSetDefault(member.ArticleSourceURL, setInput.ArticleSourceURL)
	member.Gauge = vehicleSetDefault(member.Gauge, setInput.Gauge)
	member.Epoch = vehicleSetDefault(member.Epoch, setInput.Epoch)
	member.RailwayCompany = vehicleSetDefault(member.RailwayCompany, setInput.RailwayCompany)
	member.Category = vehicleSetDefault(member.Category, setInput.Category)
	member.Gattung = vehicleSetDefault(member.Gattung, setInput.Gattung)
	member.Description = vehicleSetDefault(member.Description, setInput.Description)
	member.EAN = vehicleSetDefault(member.EAN, setInput.EAN)
	member.ProductionPeriod = vehicleSetDefault(member.ProductionPeriod, setInput.ProductionPeriod)
	member.ListPrice = vehicleSetDefault(member.ListPrice, setInput.ListPrice)
	member.AcquisitionType = vehicleSetDefault(member.AcquisitionType, setInput.AcquisitionType)
	member.AcquiredFrom = vehicleSetDefault(member.AcquiredFrom, setInput.AcquiredFrom)
	member.PurchasePrice = vehicleSetDefault(member.PurchasePrice, setInput.PurchasePrice)
	member.PurchaseDate = vehicleSetDefault(member.PurchaseDate, setInput.PurchaseDate)
	member.StorageLocation = vehicleSetDefault(member.StorageLocation, setInput.StorageLocation)
	member.StorageDetails = vehicleSetDefault(member.StorageDetails, setInput.StorageDetails)
	member.Condition = vehicleSetDefault(member.Condition, setInput.Condition)
	member.ConditionDetails = vehicleSetDefault(member.ConditionDetails, setInput.ConditionDetails)
	member.Packaging = vehicleSetDefault(member.Packaging, setInput.Packaging)
	return member
}

func vehicleSetDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func insertVehicleSetTx(
	ctx context.Context,
	tx *sql.Tx,
	setID string,
	inventoryNumber string,
	input VehicleSetInput,
	now string,
) error {
	if _, err := tx.ExecContext(ctx, `
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

func vehicleSetFromInput(
	setID string,
	inventoryNumber string,
	input VehicleSetInput,
	members []Vehicle,
	now string,
) *VehicleSet {
	return &VehicleSet{
		ID: setID, InventoryNumber: inventoryNumber, Name: input.Name, Manufacturer: input.Manufacturer,
		ArticleNumber:    input.ArticleNumber,
		ArticleSourceURL: input.ArticleSourceURL, Gauge: input.Gauge, Epoch: input.Epoch,
		RailwayCompany: input.RailwayCompany, Category: input.Category, Gattung: input.Gattung,
		Description: input.Description, EAN: input.EAN, ProductionPeriod: input.ProductionPeriod,
		ListPrice: input.ListPrice, AcquisitionType: input.AcquisitionType, AcquiredFrom: input.AcquiredFrom,
		PurchasePrice: input.PurchasePrice, PurchaseDate: input.PurchaseDate,
		StorageLocation: input.StorageLocation, StorageDetails: input.StorageDetails,
		Condition: input.Condition, ConditionDetails: input.ConditionDetails, Packaging: input.Packaging,
		Members: members, CreatedAt: now, UpdatedAt: now,
	}
}

func vehicleSetSummary(
	setID string,
	inventoryNumber string,
	input VehicleSetInput,
	memberCount int,
	position int,
) *VehicleSetSummary {
	return &VehicleSetSummary{
		ID: setID, InventoryNumber: inventoryNumber, Name: input.Name, Manufacturer: input.Manufacturer,
		ArticleNumber: input.ArticleNumber, Gauge: input.Gauge, Epoch: input.Epoch,
		AcquisitionType: input.AcquisitionType, PurchaseDate: input.PurchaseDate,
		PurchasePrice: input.PurchasePrice, Condition: input.Condition,
		MemberCount: memberCount, Position: position,
	}
}
