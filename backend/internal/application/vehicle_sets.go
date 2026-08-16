package application

import (
	"context"
	"database/sql"
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
	if err = insertVehicleSetTx(ctx, tx, setID, setInput, now); err != nil {
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

	return vehicleSetFromInput(setID, setInput, members, now), nil
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
	return applyVehicleSetFields(setInput, member)
}

func applyVehicleSetFields(setInput VehicleSetInput, member CreateVehicleInput) CreateVehicleInput {
	member.Manufacturer = setInput.Manufacturer
	member.ArticleNumber = setInput.ArticleNumber
	member.ArticleSourceURL = setInput.ArticleSourceURL
	member.Gauge = setInput.Gauge
	member.Epoch = setInput.Epoch
	member.RailwayCompany = setInput.RailwayCompany
	member.Category = setInput.Category
	member.Gattung = setInput.Gattung
	member.Description = setInput.Description
	member.EAN = setInput.EAN
	member.ProductionPeriod = setInput.ProductionPeriod
	member.ListPrice = setInput.ListPrice
	member.AcquisitionType = setInput.AcquisitionType
	member.AcquiredFrom = setInput.AcquiredFrom
	member.PurchasePrice = setInput.PurchasePrice
	member.PurchaseDate = setInput.PurchaseDate
	member.StorageLocation = setInput.StorageLocation
	member.StorageDetails = setInput.StorageDetails
	member.Condition = setInput.Condition
	member.ConditionDetails = setInput.ConditionDetails
	member.Packaging = setInput.Packaging
	return member
}

func (s *VehicleService) vehicleSetInput(ctx context.Context, setID string) (VehicleSetInput, error) {
	var input VehicleSetInput
	err := s.db.QueryRowContext(ctx, `
SELECT name, manufacturer, COALESCE(article_number, ''), COALESCE(article_source_url, ''), gauge,
       COALESCE(epoch, ''), COALESCE(railway_company, ''), category, gattung, COALESCE(description, ''),
       COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
       COALESCE(acquisition_type, ''), COALESCE(acquired_from, ''), COALESCE(purchase_price, ''),
       COALESCE(purchase_date, ''), COALESCE(storage_location, ''), COALESCE(storage_details, ''),
       COALESCE(condition, ''), COALESCE(condition_details, ''), COALESCE(packaging, '')
FROM vehicle_sets
WHERE id=?
`, setID).Scan(
		&input.Name, &input.Manufacturer, &input.ArticleNumber, &input.ArticleSourceURL, &input.Gauge,
		&input.Epoch, &input.RailwayCompany, &input.Category, &input.Gattung, &input.Description,
		&input.EAN, &input.ProductionPeriod, &input.ListPrice, &input.AcquisitionType, &input.AcquiredFrom,
		&input.PurchasePrice, &input.PurchaseDate, &input.StorageLocation, &input.StorageDetails,
		&input.Condition, &input.ConditionDetails, &input.Packaging,
	)
	if err != nil {
		return VehicleSetInput{}, fmt.Errorf("load vehicle set data: %w", err)
	}
	return input, nil
}

func insertVehicleSetTx(
	ctx context.Context,
	tx *sql.Tx,
	setID string,
	input VehicleSetInput,
	now string,
) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicle_sets(
  id, name, manufacturer, article_number, article_source_url, gauge, epoch, railway_company,
  category, gattung, description, ean, production_period, list_price, acquisition_type,
  acquired_from, purchase_price, purchase_date, storage_location, storage_details, condition,
  condition_details, packaging, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, setID, input.Name, input.Manufacturer, input.ArticleNumber, input.ArticleSourceURL, input.Gauge,
		input.Epoch, input.RailwayCompany, input.Category, input.Gattung, input.Description, input.EAN,
		input.ProductionPeriod, input.ListPrice, input.AcquisitionType, input.AcquiredFrom,
		input.PurchasePrice, input.PurchaseDate, input.StorageLocation, input.StorageDetails,
		input.Condition, input.ConditionDetails, input.Packaging, now, now); err != nil {
		return fmt.Errorf("insert vehicle set: %w", err)
	}
	return nil
}

func vehicleSetFromInput(setID string, input VehicleSetInput, members []Vehicle, now string) *VehicleSet {
	return &VehicleSet{
		ID: setID, Name: input.Name, Manufacturer: input.Manufacturer, ArticleNumber: input.ArticleNumber,
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
