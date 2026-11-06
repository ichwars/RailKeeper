package vehicledata

import (
	"context"
	"database/sql"
	"fmt"
)

type Record struct {
	ID                        string
	InventoryNumber           string
	Manufacturer              string
	ArticleNumber             string
	ArticleSourceURL          string
	Name                      string
	Gauge                     string
	Epoch                     string
	RailwayCompany            string
	Category                  string
	Gattung                   string
	Description               string
	Series                    string
	VehicleNumber             string
	MaximumSpeedKmh           *int
	HomeBase                  string
	Digital                   bool
	DigitalDecoderNumber      string
	DTDecoder                 bool
	DTDecoderNumber           string
	DecoderType               string
	ExhibitionReady           bool
	Exhibition                bool
	ABCBrakes                 bool
	EAN                       string
	ProductionPeriod          string
	ListPrice                 string
	AcquisitionType           string
	AcquiredFrom              string
	PurchasePrice             string
	PurchaseDate              string
	StorageLocation           string
	StorageDetails            string
	Condition                 string
	ConditionDetails          string
	Packaging                 string
	LengthMM                  string
	WeightG                   string
	Color                     string
	Lettering                 string
	Load                      string
	Interior                  string
	Axles                     string
	AxleCount                 string
	TractionTireCount         string
	Wheelset                  string
	CouplingSame              bool
	CouplingFront             string
	CouplingRear              string
	PowerPickup               string
	Adapter                   string
	DriveEnabled              bool
	DriveDescription          string
	HeadlightsEnabled         bool
	HeadlightsDescription     string
	LightingEnabled           bool
	LightingDescription       string
	SoundGeneratorEnabled     bool
	SoundGeneratorDescription string
	SmokeGeneratorEnabled     bool
	SmokeGeneratorDescription string
	AdditionalInfo            string
	QRCodeEnabled             bool
	CreatedAt                 string
	UpdatedAt                 string
}

// Insert writes through the caller's transaction; aggregate validation and commit stay in the use case.
func Insert(ctx context.Context, tx *sql.Tx, vehicle Record) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO vehicles(
  id, inventory_number, manufacturer, article_number, article_source_url, name, gauge, epoch, railway_company, category, gattung,
  description, series, vehicle_number, maximum_speed_kmh, home_base,
  digital, digital_decoder_number, dt_decoder, dt_decoder_number, decoder_type,
  exhibition_ready, exhibition, abc_brakes, ean, production_period, list_price,
  acquisition_type, acquired_from, purchase_price, purchase_date, storage_location, storage_details, condition, condition_details, packaging,
  length_mm, weight_g, color, lettering, load, interior, axles, axle_count, traction_tire_count, wheelset,
  coupling_same, coupling_front, coupling_rear, power_pickup, adapter,
  drive_enabled, drive_description, headlights_enabled, headlights_description, lighting_enabled, lighting_description,
  sound_generator_enabled, sound_generator_description, smoke_generator_enabled, smoke_generator_description,
  additional_info, qr_code_enabled, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, vehicle.ID, vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.ArticleSourceURL, vehicle.Name, vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description, vehicle.Series, vehicle.VehicleNumber, nullableInt(vehicle.MaximumSpeedKmh), vehicle.HomeBase, boolToInt(vehicle.Digital), vehicle.DigitalDecoderNumber, boolToInt(vehicle.DTDecoder), vehicle.DTDecoderNumber, vehicle.DecoderType, boolToInt(vehicle.ExhibitionReady), boolToInt(vehicle.Exhibition), boolToInt(vehicle.ABCBrakes), vehicle.EAN, vehicle.ProductionPeriod, vehicle.ListPrice, vehicle.AcquisitionType, vehicle.AcquiredFrom, vehicle.PurchasePrice, vehicle.PurchaseDate, vehicle.StorageLocation, vehicle.StorageDetails, vehicle.Condition, vehicle.ConditionDetails, vehicle.Packaging, vehicle.LengthMM, vehicle.WeightG, vehicle.Color, vehicle.Lettering, vehicle.Load, vehicle.Interior, vehicle.Axles, vehicle.AxleCount, vehicle.TractionTireCount, vehicle.Wheelset, boolToInt(vehicle.CouplingSame), vehicle.CouplingFront, vehicle.CouplingRear, vehicle.PowerPickup, vehicle.Adapter, boolToInt(vehicle.DriveEnabled), vehicle.DriveDescription, boolToInt(vehicle.HeadlightsEnabled), vehicle.HeadlightsDescription, boolToInt(vehicle.LightingEnabled), vehicle.LightingDescription, boolToInt(vehicle.SoundGeneratorEnabled), vehicle.SoundGeneratorDescription, boolToInt(vehicle.SmokeGeneratorEnabled), vehicle.SmokeGeneratorDescription, vehicle.AdditionalInfo, boolToInt(vehicle.QRCodeEnabled), vehicle.CreatedAt, vehicle.UpdatedAt); err != nil {
		return fmt.Errorf("insert vehicle: %w", err)
	}
	return nil
}

func Update(ctx context.Context, tx *sql.Tx, vehicle Record) (int64, error) {
	result, err := tx.ExecContext(ctx, `
UPDATE vehicles
SET inventory_number=?, manufacturer=?, article_number=?, article_source_url=?, name=?, gauge=?, epoch=?, railway_company=?, category=?, gattung=?,
    description=?, series=?, vehicle_number=?, maximum_speed_kmh=?, home_base=?,
    digital=?, digital_decoder_number=?, dt_decoder=?, dt_decoder_number=?, decoder_type=?,
    exhibition_ready=?, exhibition=?, abc_brakes=?, ean=?, production_period=?, list_price=?,
    acquisition_type=?, acquired_from=?, purchase_price=?, purchase_date=?, storage_location=?, storage_details=?, condition=?, condition_details=?, packaging=?,
    length_mm=?, weight_g=?, color=?, lettering=?, load=?, interior=?, axles=?, axle_count=?, traction_tire_count=?, wheelset=?,
    coupling_same=?, coupling_front=?, coupling_rear=?, power_pickup=?, adapter=?,
    drive_enabled=?, drive_description=?, headlights_enabled=?, headlights_description=?, lighting_enabled=?, lighting_description=?,
    sound_generator_enabled=?, sound_generator_description=?, smoke_generator_enabled=?, smoke_generator_description=?,
    additional_info=?, qr_code_enabled=?, updated_at=?
WHERE id=?
`, vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.ArticleSourceURL, vehicle.Name, vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description, vehicle.Series, vehicle.VehicleNumber, nullableInt(vehicle.MaximumSpeedKmh), vehicle.HomeBase, boolToInt(vehicle.Digital), vehicle.DigitalDecoderNumber, boolToInt(vehicle.DTDecoder), vehicle.DTDecoderNumber, vehicle.DecoderType, boolToInt(vehicle.ExhibitionReady), boolToInt(vehicle.Exhibition), boolToInt(vehicle.ABCBrakes), vehicle.EAN, vehicle.ProductionPeriod, vehicle.ListPrice, vehicle.AcquisitionType, vehicle.AcquiredFrom, vehicle.PurchasePrice, vehicle.PurchaseDate, vehicle.StorageLocation, vehicle.StorageDetails, vehicle.Condition, vehicle.ConditionDetails, vehicle.Packaging, vehicle.LengthMM, vehicle.WeightG, vehicle.Color, vehicle.Lettering, vehicle.Load, vehicle.Interior, vehicle.Axles, vehicle.AxleCount, vehicle.TractionTireCount, vehicle.Wheelset, boolToInt(vehicle.CouplingSame), vehicle.CouplingFront, vehicle.CouplingRear, vehicle.PowerPickup, vehicle.Adapter, boolToInt(vehicle.DriveEnabled), vehicle.DriveDescription, boolToInt(vehicle.HeadlightsEnabled), vehicle.HeadlightsDescription, boolToInt(vehicle.LightingEnabled), vehicle.LightingDescription, boolToInt(vehicle.SoundGeneratorEnabled), vehicle.SoundGeneratorDescription, boolToInt(vehicle.SmokeGeneratorEnabled), vehicle.SmokeGeneratorDescription, vehicle.AdditionalInfo, boolToInt(vehicle.QRCodeEnabled), vehicle.UpdatedAt, vehicle.ID)
	if err != nil {
		return 0, fmt.Errorf("update vehicle: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read update result: %w", err)
	}
	return affected, nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
