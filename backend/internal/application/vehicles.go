package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrVehicleValidation = errors.New("vehicle validation failed")
	ErrVehicleNotFound   = errors.New("vehicle not found")
)

type VehicleService struct {
	db *sql.DB
}

type Vehicle struct {
	ID                   string `json:"id"`
	InventoryNumber      string `json:"inventoryNumber"`
	Manufacturer         string `json:"manufacturer"`
	ArticleNumber        string `json:"articleNumber,omitempty"`
	Name                 string `json:"name"`
	Gauge                string `json:"gauge"`
	Epoch                string `json:"epoch,omitempty"`
	RailwayCompany       string `json:"railwayCompany,omitempty"`
	Category             string `json:"category,omitempty"`
	Gattung              string `json:"gattung,omitempty"`
	Description          string `json:"description,omitempty"`
	Series               string `json:"series,omitempty"`
	VehicleNumber        string `json:"vehicleNumber,omitempty"`
	Digital              bool   `json:"digital"`
	DigitalDecoderNumber string `json:"digitalDecoderNumber,omitempty"`
	DTDecoder            bool   `json:"dtDecoder"`
	DTDecoderNumber      string `json:"dtDecoderNumber,omitempty"`
	ExhibitionReady      bool   `json:"exhibitionReady"`
	ABCBrakes            bool   `json:"abcBrakes"`
	EAN                  string `json:"ean,omitempty"`
	ProductionPeriod     string `json:"productionPeriod,omitempty"`
	ListPrice            string `json:"listPrice,omitempty"`
	CreatedAt            string `json:"createdAt"`
	UpdatedAt            string `json:"updatedAt"`
}

type CreateVehicleInput struct {
	InventoryNumber      string `json:"inventoryNumber"`
	Manufacturer         string `json:"manufacturer"`
	ArticleNumber        string `json:"articleNumber"`
	Name                 string `json:"name"`
	Gauge                string `json:"gauge"`
	Epoch                string `json:"epoch"`
	RailwayCompany       string `json:"railwayCompany"`
	Category             string `json:"category"`
	Gattung              string `json:"gattung"`
	Description          string `json:"description"`
	Series               string `json:"series"`
	VehicleNumber        string `json:"vehicleNumber"`
	Digital              bool   `json:"digital"`
	DigitalDecoderNumber string `json:"digitalDecoderNumber"`
	DTDecoder            bool   `json:"dtDecoder"`
	DTDecoderNumber      string `json:"dtDecoderNumber"`
	ExhibitionReady      bool   `json:"exhibitionReady"`
	ABCBrakes            bool   `json:"abcBrakes"`
	EAN                  string `json:"ean"`
	ProductionPeriod     string `json:"productionPeriod"`
	ListPrice            string `json:"listPrice"`
}

func NewVehicleService(db *sql.DB) *VehicleService {
	return &VehicleService{db: db}
}

func (s *VehicleService) List(ctx context.Context, query string) ([]Vehicle, error) {
	like := "%" + strings.TrimSpace(query) + "%"
	rows, err := s.db.QueryContext(ctx, `
SELECT id, inventory_number, manufacturer, COALESCE(article_number, ''), name, gauge,
       COALESCE(epoch, ''), COALESCE(railway_company, ''), COALESCE(category, ''), COALESCE(gattung, ''),
       COALESCE(description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''),
       exhibition_ready, abc_brakes, COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
       created_at, updated_at
FROM vehicles
WHERE ? = '%%'
   OR inventory_number LIKE ?
   OR manufacturer LIKE ?
   OR article_number LIKE ?
   OR name LIKE ?
ORDER BY updated_at DESC, inventory_number ASC
`, like, like, like, like, like)
	if err != nil {
		return nil, fmt.Errorf("list vehicles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	vehicles := []Vehicle{}
	for rows.Next() {
		var vehicle Vehicle
		var digital int
		var dtDecoder int
		var exhibitionReady int
		var abcBrakes int
		if err := rows.Scan(
			&vehicle.ID,
			&vehicle.InventoryNumber,
			&vehicle.Manufacturer,
			&vehicle.ArticleNumber,
			&vehicle.Name,
			&vehicle.Gauge,
			&vehicle.Epoch,
			&vehicle.RailwayCompany,
			&vehicle.Category,
			&vehicle.Gattung,
			&vehicle.Description,
			&vehicle.Series,
			&vehicle.VehicleNumber,
			&digital,
			&vehicle.DigitalDecoderNumber,
			&dtDecoder,
			&vehicle.DTDecoderNumber,
			&exhibitionReady,
			&abcBrakes,
			&vehicle.EAN,
			&vehicle.ProductionPeriod,
			&vehicle.ListPrice,
			&vehicle.CreatedAt,
			&vehicle.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vehicle: %w", err)
		}
		vehicle.Digital = digital == 1
		vehicle.DTDecoder = dtDecoder == 1
		vehicle.ExhibitionReady = exhibitionReady == 1
		vehicle.ABCBrakes = abcBrakes == 1
		vehicles = append(vehicles, vehicle)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate vehicles: %w", err)
	}
	return vehicles, nil
}

func (s *VehicleService) Get(ctx context.Context, id string) (*Vehicle, error) {
	vehicle, err := s.get(ctx, strings.TrimSpace(id))
	if err != nil {
		return nil, err
	}
	return vehicle, nil
}

func (s *VehicleService) Create(ctx context.Context, input CreateVehicleInput, actorUserID string) (*Vehicle, error) {
	input = cleanVehicleInput(input)
	if input.Manufacturer == "" || input.Name == "" || input.Gauge == "" {
		return nil, ErrVehicleValidation
	}
	if input.InventoryNumber == "" {
		next, err := s.nextInventoryNumber(ctx)
		if err != nil {
			return nil, err
		}
		input.InventoryNumber = next
	}

	now := time.Now().UTC().Format(time.RFC3339)
	vehicle := Vehicle{
		ID:                   randomID(),
		InventoryNumber:      input.InventoryNumber,
		Manufacturer:         input.Manufacturer,
		ArticleNumber:        input.ArticleNumber,
		Name:                 input.Name,
		Gauge:                input.Gauge,
		Epoch:                input.Epoch,
		RailwayCompany:       input.RailwayCompany,
		Category:             input.Category,
		Gattung:              input.Gattung,
		Description:          input.Description,
		Series:               input.Series,
		VehicleNumber:        input.VehicleNumber,
		Digital:              input.Digital,
		DigitalDecoderNumber: input.DigitalDecoderNumber,
		DTDecoder:            input.DTDecoder,
		DTDecoderNumber:      input.DTDecoderNumber,
		ExhibitionReady:      input.ExhibitionReady,
		ABCBrakes:            input.ABCBrakes,
		EAN:                  input.EAN,
		ProductionPeriod:     input.ProductionPeriod,
		ListPrice:            input.ListPrice,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create vehicle: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
INSERT INTO vehicles(
  id, inventory_number, manufacturer, article_number, name, gauge, epoch, railway_company, category, gattung,
  description, series, vehicle_number, digital, digital_decoder_number, dt_decoder, dt_decoder_number,
  exhibition_ready, abc_brakes, ean, production_period, list_price, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, vehicle.ID, vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.Name, vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description, vehicle.Series, vehicle.VehicleNumber, boolToInt(vehicle.Digital), vehicle.DigitalDecoderNumber, boolToInt(vehicle.DTDecoder), vehicle.DTDecoderNumber, boolToInt(vehicle.ExhibitionReady), boolToInt(vehicle.ABCBrakes), vehicle.EAN, vehicle.ProductionPeriod, vehicle.ListPrice, vehicle.CreatedAt, vehicle.UpdatedAt); err != nil {
		return nil, fmt.Errorf("insert vehicle: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleCreated', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, vehicle.ID, now); err != nil {
		return nil, fmt.Errorf("write vehicle audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create vehicle: %w", err)
	}

	return &vehicle, nil
}

func (s *VehicleService) Update(ctx context.Context, id string, input CreateVehicleInput, actorUserID string) (*Vehicle, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrVehicleNotFound
	}

	existing, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}

	input = cleanVehicleInput(input)
	if input.InventoryNumber == "" {
		input.InventoryNumber = existing.InventoryNumber
	}
	if input.Manufacturer == "" || input.Name == "" || input.Gauge == "" {
		return nil, ErrVehicleValidation
	}

	now := time.Now().UTC().Format(time.RFC3339)
	vehicle := Vehicle{
		ID:                   id,
		InventoryNumber:      input.InventoryNumber,
		Manufacturer:         input.Manufacturer,
		ArticleNumber:        input.ArticleNumber,
		Name:                 input.Name,
		Gauge:                input.Gauge,
		Epoch:                input.Epoch,
		RailwayCompany:       input.RailwayCompany,
		Category:             input.Category,
		Gattung:              input.Gattung,
		Description:          input.Description,
		Series:               input.Series,
		VehicleNumber:        input.VehicleNumber,
		Digital:              input.Digital,
		DigitalDecoderNumber: input.DigitalDecoderNumber,
		DTDecoder:            input.DTDecoder,
		DTDecoderNumber:      input.DTDecoderNumber,
		ExhibitionReady:      input.ExhibitionReady,
		ABCBrakes:            input.ABCBrakes,
		EAN:                  input.EAN,
		ProductionPeriod:     input.ProductionPeriod,
		ListPrice:            input.ListPrice,
		CreatedAt:            existing.CreatedAt,
		UpdatedAt:            now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin update vehicle: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `
UPDATE vehicles
SET inventory_number=?, manufacturer=?, article_number=?, name=?, gauge=?, epoch=?, railway_company=?, category=?, gattung=?,
    description=?, series=?, vehicle_number=?, digital=?, digital_decoder_number=?, dt_decoder=?, dt_decoder_number=?,
    exhibition_ready=?, abc_brakes=?, ean=?, production_period=?, list_price=?, updated_at=?
WHERE id=?
`, vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.Name, vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description, vehicle.Series, vehicle.VehicleNumber, boolToInt(vehicle.Digital), vehicle.DigitalDecoderNumber, boolToInt(vehicle.DTDecoder), vehicle.DTDecoderNumber, boolToInt(vehicle.ExhibitionReady), boolToInt(vehicle.ABCBrakes), vehicle.EAN, vehicle.ProductionPeriod, vehicle.ListPrice, vehicle.UpdatedAt, vehicle.ID)
	if err != nil {
		return nil, fmt.Errorf("update vehicle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read update result: %w", err)
	}
	if affected == 0 {
		_ = tx.Rollback()
		return nil, ErrVehicleNotFound
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleUpdated', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, vehicle.ID, now); err != nil {
		return nil, fmt.Errorf("write vehicle audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit update vehicle: %w", err)
	}

	return &vehicle, nil
}

func (s *VehicleService) Delete(ctx context.Context, id, actorUserID string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrVehicleNotFound
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin delete vehicle: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx, `DELETE FROM vehicles WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete vehicle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read delete result: %w", err)
	}
	if affected == 0 {
		_ = tx.Rollback()
		return ErrVehicleNotFound
	}

	if _, err = tx.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, ?, 'VehicleDeleted', 'vehicle', ?, ?, '{}')
`, randomID(), actorUserID, id, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return fmt.Errorf("write vehicle audit log: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit delete vehicle: %w", err)
	}

	return nil
}

func (s *VehicleService) get(ctx context.Context, id string) (*Vehicle, error) {
	var vehicle Vehicle
	var digital int
	var dtDecoder int
	var exhibitionReady int
	var abcBrakes int
	if err := s.db.QueryRowContext(ctx, `
SELECT id, inventory_number, manufacturer, COALESCE(article_number, ''), name, gauge,
       COALESCE(epoch, ''), COALESCE(railway_company, ''), COALESCE(category, ''), COALESCE(gattung, ''),
       COALESCE(description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
       digital, COALESCE(digital_decoder_number, ''), dt_decoder, COALESCE(dt_decoder_number, ''),
       exhibition_ready, abc_brakes, COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
       created_at, updated_at
FROM vehicles
WHERE id=?
`, id).Scan(
		&vehicle.ID,
		&vehicle.InventoryNumber,
		&vehicle.Manufacturer,
		&vehicle.ArticleNumber,
		&vehicle.Name,
		&vehicle.Gauge,
		&vehicle.Epoch,
		&vehicle.RailwayCompany,
		&vehicle.Category,
		&vehicle.Gattung,
		&vehicle.Description,
		&vehicle.Series,
		&vehicle.VehicleNumber,
		&digital,
		&vehicle.DigitalDecoderNumber,
		&dtDecoder,
		&vehicle.DTDecoderNumber,
		&exhibitionReady,
		&abcBrakes,
		&vehicle.EAN,
		&vehicle.ProductionPeriod,
		&vehicle.ListPrice,
		&vehicle.CreatedAt,
		&vehicle.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrVehicleNotFound
		}
		return nil, fmt.Errorf("get vehicle: %w", err)
	}
	vehicle.Digital = digital == 1
	vehicle.DTDecoder = dtDecoder == 1
	vehicle.ExhibitionReady = exhibitionReady == 1
	vehicle.ABCBrakes = abcBrakes == 1

	return &vehicle, nil
}

func (s *VehicleService) nextInventoryNumber(ctx context.Context) (string, error) {
	var next int
	if err := s.db.QueryRowContext(ctx, `
SELECT COALESCE(MAX(CAST(SUBSTR(inventory_number, 8) AS INTEGER)), 0) + 1
FROM vehicles
WHERE inventory_number LIKE 'RK-FAH-%'
`).Scan(&next); err != nil {
		return "", fmt.Errorf("next inventory number: %w", err)
	}
	return fmt.Sprintf("RK-FAH-%06d", next), nil
}

func cleanVehicleInput(input CreateVehicleInput) CreateVehicleInput {
	input.InventoryNumber = strings.TrimSpace(input.InventoryNumber)
	input.Manufacturer = strings.TrimSpace(input.Manufacturer)
	input.ArticleNumber = strings.TrimSpace(input.ArticleNumber)
	input.Name = strings.TrimSpace(input.Name)
	input.Gauge = strings.TrimSpace(input.Gauge)
	input.Epoch = strings.TrimSpace(input.Epoch)
	input.RailwayCompany = strings.TrimSpace(input.RailwayCompany)
	input.Category = strings.TrimSpace(input.Category)
	input.Gattung = strings.TrimSpace(input.Gattung)
	input.Description = strings.TrimSpace(input.Description)
	input.Series = strings.TrimSpace(input.Series)
	input.VehicleNumber = strings.TrimSpace(input.VehicleNumber)
	input.DigitalDecoderNumber = strings.TrimSpace(input.DigitalDecoderNumber)
	input.DTDecoderNumber = strings.TrimSpace(input.DTDecoderNumber)
	input.EAN = strings.TrimSpace(input.EAN)
	input.ProductionPeriod = strings.TrimSpace(input.ProductionPeriod)
	input.ListPrice = strings.TrimSpace(input.ListPrice)
	return input
}
