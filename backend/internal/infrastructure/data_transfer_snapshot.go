package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"railkeeper/backend/internal/application"
)

func (repository *DataTransferRepository) Snapshot(
	ctx context.Context,
	areas []application.TransferArea,
) (application.DataTransferSnapshot, error) {
	tx, err := repository.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return application.DataTransferSnapshot{}, fmt.Errorf("begin transfer snapshot: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	snapshot := application.DataTransferSnapshot{}
	for _, area := range areas {
		switch area {
		case application.TransferVehicles:
			snapshot.Vehicles, err = transferVehicleSnapshot(ctx, tx)
		case application.TransferAccessories:
			snapshot.Accessories, err = transferAccessorySnapshot(ctx, tx)
		case application.TransferExhibitionLists:
			snapshot.ExhibitionLists, err = transferExhibitionSnapshot(ctx, tx)
		default:
			err = fmt.Errorf("%w: unsupported snapshot area %q", application.ErrDataTransferValidation, area)
		}
		if err != nil {
			return application.DataTransferSnapshot{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return application.DataTransferSnapshot{}, fmt.Errorf("commit transfer snapshot: %w", err)
	}
	return snapshot, nil
}

func (repository *DataTransferRepository) GetArtifact(
	ctx context.Context,
	id string,
) (application.DataTransferArtifact, error) {
	artifact := application.DataTransferArtifact{}
	err := repository.db.QueryRowContext(ctx, `
SELECT id, job_id, relative_path, display_name, mime_type, size_bytes, sha256,
       COALESCE(deleted_at, ''), created_at
FROM data_transfer_artifacts
WHERE id=?`, id).Scan(
		&artifact.ID, &artifact.JobID, &artifact.RelativePath, &artifact.DisplayName, &artifact.MIMEType,
		&artifact.SizeBytes, &artifact.SHA256, &artifact.DeletedAt, &artifact.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return application.DataTransferArtifact{}, fmt.Errorf("get transfer artifact: %w", sql.ErrNoRows)
	}
	if err != nil {
		return application.DataTransferArtifact{}, fmt.Errorf("get transfer artifact: %w", err)
	}
	return artifact, nil
}

func transferVehicleSnapshot(ctx context.Context, tx *sql.Tx) ([]application.TransferVehicle, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, inventory_number, manufacturer, COALESCE(article_number, ''), COALESCE(article_source_url, ''),
       name, gauge, COALESCE(epoch, ''), COALESCE(railway_company, ''), COALESCE(category, ''),
       COALESCE(gattung, ''), COALESCE(description, ''), COALESCE(series, ''), COALESCE(vehicle_number, ''),
       maximum_speed_kmh, COALESCE(home_base, ''), digital, COALESCE(digital_decoder_number, ''),
       dt_decoder, COALESCE(dt_decoder_number, ''), COALESCE(decoder_type, ''), exhibition_ready, exhibition,
       abc_brakes, COALESCE(ean, ''), COALESCE(production_period, ''), COALESCE(list_price, ''),
       COALESCE(acquisition_type, ''), COALESCE(acquired_from, ''), COALESCE(purchase_price, ''),
       COALESCE(purchase_date, ''), COALESCE(storage_location, ''), COALESCE(storage_details, ''),
       COALESCE(condition, ''), COALESCE(condition_details, ''), COALESCE(packaging, ''),
       COALESCE(length_mm, ''), COALESCE(weight_g, ''), COALESCE(color, ''), COALESCE(lettering, ''),
       COALESCE(load, ''), COALESCE(interior, ''), COALESCE(axles, ''), COALESCE(axle_count, ''),
       COALESCE(traction_tire_count, ''), COALESCE(wheelset, ''), coupling_same,
       COALESCE(coupling_front, ''), COALESCE(coupling_rear, ''), COALESCE(power_pickup, ''),
       COALESCE(adapter, ''), drive_enabled, COALESCE(drive_description, ''), headlights_enabled,
       COALESCE(headlights_description, ''), lighting_enabled, COALESCE(lighting_description, ''),
       sound_generator_enabled, COALESCE(sound_generator_description, ''), smoke_generator_enabled,
       COALESCE(smoke_generator_description, ''), COALESCE(additional_info, ''), qr_code_enabled,
       created_at, updated_at
FROM vehicles
ORDER BY inventory_number COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer vehicles: %w", err)
	}
	defer func() { _ = rows.Close() }()
	vehicles := []application.TransferVehicle{}
	for rows.Next() {
		vehicle := application.TransferVehicle{}
		var maximumSpeed sql.NullInt64
		var digital, dtDecoder, exhibitionReady, exhibition, abcBrakes int
		var couplingSame, driveEnabled, headlightsEnabled, lightingEnabled int
		var soundGeneratorEnabled, smokeGeneratorEnabled, qrCodeEnabled int
		if err := rows.Scan(
			&vehicle.ID, &vehicle.InventoryNumber, &vehicle.Manufacturer, &vehicle.ArticleNumber,
			&vehicle.ArticleSourceURL, &vehicle.Name, &vehicle.Gauge, &vehicle.Epoch, &vehicle.RailwayCompany,
			&vehicle.Category, &vehicle.Gattung, &vehicle.Description, &vehicle.Series, &vehicle.VehicleNumber,
			&maximumSpeed, &vehicle.HomeBase, &digital, &vehicle.DigitalDecoderNumber, &dtDecoder,
			&vehicle.DTDecoderNumber, &vehicle.DecoderType, &exhibitionReady, &exhibition, &abcBrakes,
			&vehicle.EAN, &vehicle.ProductionPeriod, &vehicle.ListPrice, &vehicle.AcquisitionType,
			&vehicle.AcquiredFrom, &vehicle.PurchasePrice, &vehicle.PurchaseDate, &vehicle.StorageLocation,
			&vehicle.StorageDetails, &vehicle.Condition, &vehicle.ConditionDetails, &vehicle.Packaging,
			&vehicle.LengthMM, &vehicle.WeightG, &vehicle.Color, &vehicle.Lettering, &vehicle.Load,
			&vehicle.Interior, &vehicle.Axles, &vehicle.AxleCount, &vehicle.TractionTireCount, &vehicle.Wheelset,
			&couplingSame, &vehicle.CouplingFront, &vehicle.CouplingRear, &vehicle.PowerPickup, &vehicle.Adapter,
			&driveEnabled, &vehicle.DriveDescription, &headlightsEnabled, &vehicle.HeadlightsDescription,
			&lightingEnabled, &vehicle.LightingDescription, &soundGeneratorEnabled,
			&vehicle.SoundGeneratorDescription, &smokeGeneratorEnabled, &vehicle.SmokeGeneratorDescription,
			&vehicle.AdditionalInfo, &qrCodeEnabled,
			&vehicle.CreatedAt, &vehicle.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transfer vehicle: %w", err)
		}
		if maximumSpeed.Valid {
			value := int(maximumSpeed.Int64)
			vehicle.MaximumSpeedKmh = &value
		}
		vehicle.Digital = digital != 0
		vehicle.DTDecoder = dtDecoder != 0
		vehicle.ExhibitionReady = exhibitionReady != 0
		vehicle.Exhibition = exhibition != 0
		vehicle.ABCBrakes = abcBrakes != 0
		vehicle.CouplingSame = couplingSame != 0
		vehicle.DriveEnabled = driveEnabled != 0
		vehicle.HeadlightsEnabled = headlightsEnabled != 0
		vehicle.LightingEnabled = lightingEnabled != 0
		vehicle.SoundGeneratorEnabled = soundGeneratorEnabled != 0
		vehicle.SmokeGeneratorEnabled = smokeGeneratorEnabled != 0
		vehicle.QRCodeEnabled = qrCodeEnabled != 0
		vehicles = append(vehicles, vehicle)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer vehicles: %w", err)
	}
	return vehicles, nil
}

func transferAccessorySnapshot(ctx context.Context, tx *sql.Tx) ([]application.TransferAccessory, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, inventory_number, manufacturer, article_number, name, category, tracking_mode, description,
       ean, manufacturer_status, article_type, subtype, gauges_json, scale, list_price, package_quantity,
       stock_unit, minimum_stock, inventory_strategy, manufacturer_url, product_url,
       alternative_numbers_json, keywords_json, compatibility_notes, internal_notes, archived,
       created_at, updated_at
FROM accessory_products
ORDER BY inventory_number COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer accessories: %w", err)
	}
	accessories := []application.TransferAccessory{}
	for rows.Next() {
		accessory := application.TransferAccessory{}
		var gauges, alternativeNumbers, keywords string
		var archived int
		if err := rows.Scan(
			&accessory.ID, &accessory.InventoryNumber, &accessory.Manufacturer, &accessory.ArticleNumber,
			&accessory.Name, &accessory.Category, &accessory.TrackingMode, &accessory.Description, &accessory.EAN,
			&accessory.ManufacturerStatus, &accessory.ArticleType, &accessory.Subtype, &gauges, &accessory.Scale,
			&accessory.ListPrice, &accessory.PackageQuantity, &accessory.StockUnit, &accessory.MinimumStock,
			&accessory.InventoryStrategy, &accessory.ManufacturerURL, &accessory.ProductURL, &alternativeNumbers,
			&keywords, &accessory.CompatibilityNotes, &accessory.InternalNotes, &archived, &accessory.CreatedAt,
			&accessory.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan transfer accessory: %w", err)
		}
		if err := decodeTransferStringArray(gauges, &accessory.Gauges); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("decode transfer accessory gauges: %w", err)
		}
		if err := decodeTransferStringArray(alternativeNumbers, &accessory.AlternativeNumbers); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("decode transfer accessory alternative numbers: %w", err)
		}
		if err := decodeTransferStringArray(keywords, &accessory.Keywords); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("decode transfer accessory keywords: %w", err)
		}
		accessory.Archived = archived != 0
		accessories = append(accessories, accessory)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate transfer accessories: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close transfer accessories: %w", err)
	}
	for index := range accessories {
		accessories[index].Stock, err = transferAccessoryStock(ctx, tx, accessories[index].ID)
		if err != nil {
			return nil, err
		}
		accessories[index].Assets, err = transferAccessoryAssets(ctx, tx, accessories[index].ID)
		if err != nil {
			return nil, err
		}
		accessories[index].FingerprintState, err = transferAccessoryFingerprintState(
			ctx, tx, accessories[index].ID,
		)
		if err != nil {
			return nil, err
		}
	}
	return accessories, nil
}

func transferAccessoryStock(
	ctx context.Context,
	tx *sql.Tx,
	productID string,
) ([]application.TransferAccessoryStock, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT stock.location_id, location.name, stock.quantity, stock.updated_at
FROM accessory_stock stock
JOIN storage_locations location ON location.id=stock.location_id
WHERE stock.product_id=?
ORDER BY location.name COLLATE NOCASE, stock.location_id`, productID)
	if err != nil {
		return nil, fmt.Errorf("query transfer accessory stock: %w", err)
	}
	defer func() { _ = rows.Close() }()
	stock := []application.TransferAccessoryStock{}
	for rows.Next() {
		level := application.TransferAccessoryStock{}
		if err := rows.Scan(&level.LocationID, &level.LocationName, &level.Quantity, &level.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan transfer accessory stock: %w", err)
		}
		stock = append(stock, level)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer accessory stock: %w", err)
	}
	return stock, nil
}

func transferAccessoryAssets(
	ctx context.Context,
	tx *sql.Tx,
	productID string,
) ([]application.TransferAccessoryAsset, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT asset.id, COALESCE(asset.inventory_number, ''), asset.serial_number, asset.condition_state,
       asset.lifecycle_state, COALESCE(asset.storage_location_id, ''), COALESCE(location.name, ''),
       asset.purchase_date, asset.purchase_price, asset.warranty_until, asset.notes, asset.created_at, asset.updated_at
FROM accessory_assets asset
LEFT JOIN storage_locations location ON location.id=asset.storage_location_id
WHERE asset.product_id=?
ORDER BY asset.inventory_number COLLATE NOCASE, asset.id`, productID)
	if err != nil {
		return nil, fmt.Errorf("query transfer accessory assets: %w", err)
	}
	defer func() { _ = rows.Close() }()
	assets := []application.TransferAccessoryAsset{}
	for rows.Next() {
		asset := application.TransferAccessoryAsset{}
		if err := rows.Scan(
			&asset.ID, &asset.InventoryNumber, &asset.SerialNumber, &asset.Condition, &asset.Lifecycle,
			&asset.StorageLocationID, &asset.StorageLocationName, &asset.PurchaseDate, &asset.PurchasePrice,
			&asset.WarrantyUntil, &asset.Notes, &asset.CreatedAt, &asset.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transfer accessory asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer accessory assets: %w", err)
	}
	return assets, nil
}

func transferAccessoryFingerprintState(
	ctx context.Context,
	tx *sql.Tx,
	productID string,
) (application.TransferAccessoryFingerprintState, error) {
	state := application.TransferAccessoryFingerprintState{
		Reservations:     []application.TransferAccessoryReservationFingerprint{},
		Installations:    []application.TransferAccessoryInstallationFingerprint{},
		ConditionHistory: []application.TransferAccessoryConditionHistoryFingerprint{},
	}
	rows, err := tx.QueryContext(ctx, `
SELECT id, COALESCE(asset_id, ''), location_id, quantity, status, updated_at
FROM accessory_reservations
WHERE product_id=?
ORDER BY id`, productID)
	if err != nil {
		return state, fmt.Errorf("query transfer accessory reservation fingerprints: %w", err)
	}
	for rows.Next() {
		reservation := application.TransferAccessoryReservationFingerprint{}
		if err := rows.Scan(
			&reservation.ID, &reservation.AssetID, &reservation.LocationID, &reservation.Quantity,
			&reservation.Status, &reservation.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return state, fmt.Errorf("scan transfer accessory reservation fingerprint: %w", err)
		}
		state.Reservations = append(state.Reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return state, fmt.Errorf("iterate transfer accessory reservation fingerprints: %w", err)
	}
	if err := rows.Close(); err != nil {
		return state, fmt.Errorf("close transfer accessory reservation fingerprints: %w", err)
	}

	rows, err = tx.QueryContext(ctx, `
SELECT id, COALESCE(asset_id, ''), source_location_id, quantity, condition_state, installed_at,
       COALESCE(removed_at, '')
FROM accessory_installations
WHERE product_id=?
ORDER BY id`, productID)
	if err != nil {
		return state, fmt.Errorf("query transfer accessory installation fingerprints: %w", err)
	}
	for rows.Next() {
		installation := application.TransferAccessoryInstallationFingerprint{}
		if err := rows.Scan(
			&installation.ID, &installation.AssetID, &installation.SourceLocationID, &installation.Quantity,
			&installation.Condition, &installation.InstalledAt, &installation.RemovedAt,
		); err != nil {
			_ = rows.Close()
			return state, fmt.Errorf("scan transfer accessory installation fingerprint: %w", err)
		}
		state.Installations = append(state.Installations, installation)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return state, fmt.Errorf("iterate transfer accessory installation fingerprints: %w", err)
	}
	if err := rows.Close(); err != nil {
		return state, fmt.Errorf("close transfer accessory installation fingerprints: %w", err)
	}

	rows, err = tx.QueryContext(ctx, `
SELECT history.id, history.installation_id, history.condition_state, history.changed_at
FROM accessory_installation_condition_history history
JOIN accessory_installations installation ON installation.id=history.installation_id
WHERE installation.product_id=?
ORDER BY history.id`, productID)
	if err != nil {
		return state, fmt.Errorf("query transfer accessory condition history fingerprints: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		history := application.TransferAccessoryConditionHistoryFingerprint{}
		if err := rows.Scan(&history.ID, &history.InstallationID, &history.Condition, &history.ChangedAt); err != nil {
			return state, fmt.Errorf("scan transfer accessory condition history fingerprint: %w", err)
		}
		state.ConditionHistory = append(state.ConditionHistory, history)
	}
	if err := rows.Err(); err != nil {
		return state, fmt.Errorf("iterate transfer accessory condition history fingerprints: %w", err)
	}
	return state, nil
}

func transferExhibitionSnapshot(ctx context.Context, tx *sql.Tx) ([]application.TransferExhibitionList, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT id, designation, list_date, end_date, location, description, organization_notes,
       status, locked, lock_reason, locked_at, completed_at, archived_at, created_at, updated_at
FROM exhibition_lists
ORDER BY list_date, designation COLLATE NOCASE, id`)
	if err != nil {
		return nil, fmt.Errorf("query transfer exhibition lists: %w", err)
	}
	lists := []application.TransferExhibitionList{}
	for rows.Next() {
		list := application.TransferExhibitionList{}
		var locked int
		if err := rows.Scan(
			&list.ID, &list.Designation, &list.Date, &list.EndDate, &list.Location, &list.Description,
			&list.OrganizationNotes, &list.Status, &locked, &list.LockReason, &list.LockedAt,
			&list.CompletedAt, &list.ArchivedAt, &list.CreatedAt, &list.UpdatedAt,
		); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan transfer exhibition list: %w", err)
		}
		list.Locked = locked != 0
		lists = append(lists, list)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate transfer exhibition lists: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close transfer exhibition lists: %w", err)
	}
	for index := range lists {
		lists[index].Entries, err = transferExhibitionEntries(ctx, tx, lists[index].ID)
		if err != nil {
			return nil, err
		}
	}
	return lists, nil
}

func transferExhibitionEntries(
	ctx context.Context,
	tx *sql.Tx,
	listID string,
) ([]application.TransferExhibitionEntry, error) {
	rows, err := tx.QueryContext(ctx, `
SELECT entry.id, entry.vehicle_id, COALESCE(vehicle.inventory_number, ''), entry.owner,
       COALESCE(entry.image_url, ''), entry.locomotive_name, entry.gattung, entry.series, entry.manufacturer,
       entry.epoch, entry.railway_company, entry.day_scope, entry.dt_decoder,
       COALESCE(entry.decoder_number, ''), entry.decoder_type, entry.adapter, entry.interface_name,
       entry.sx_address, entry.analog, entry.availability,
       COALESCE(entry.function_keys, ''), COALESCE(entry.notes, ''), entry.sort_order,
       entry.created_at, entry.updated_at
FROM exhibition_entries entry
LEFT JOIN vehicles vehicle ON vehicle.id=entry.vehicle_id
WHERE entry.list_id=?
ORDER BY entry.sort_order, entry.locomotive_name COLLATE NOCASE, entry.id`, listID)
	if err != nil {
		return nil, fmt.Errorf("query transfer exhibition entries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	entries := []application.TransferExhibitionEntry{}
	for rows.Next() {
		entry := application.TransferExhibitionEntry{}
		var dtDecoder, analog int
		if err := rows.Scan(
			&entry.ID, &entry.VehicleID, &entry.VehicleInventoryNumber, &entry.Owner, &entry.ImageURL,
			&entry.LocomotiveName, &entry.Gattung, &entry.Series, &entry.Manufacturer, &entry.Epoch,
			&entry.RailwayCompany, &entry.DayScope, &dtDecoder, &entry.DecoderNumber, &entry.DecoderType,
			&entry.Adapter, &entry.InterfaceName, &entry.SXAddress, &analog, &entry.Availability,
			&entry.FunctionKeys, &entry.Notes, &entry.SortOrder,
			&entry.CreatedAt, &entry.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan transfer exhibition entry: %w", err)
		}
		entry.DTDecoder = dtDecoder != 0
		entry.Analog = analog != 0
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfer exhibition entries: %w", err)
	}
	return entries, nil
}

func decodeTransferStringArray(value string, target *[]string) error {
	if err := json.Unmarshal([]byte(value), target); err != nil {
		return err
	}
	if *target == nil {
		*target = []string{}
	}
	return nil
}

var _ dataTransferSnapshotRepository = (*DataTransferRepository)(nil)

type dataTransferSnapshotRepository interface {
	Snapshot(context.Context, []application.TransferArea) (application.DataTransferSnapshot, error)
	GetArtifact(context.Context, string) (application.DataTransferArtifact, error)
}
