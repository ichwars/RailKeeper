package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
)

type dataTransferPersistedPreview struct {
	SourceSHA256 string                                  `json:"sourceSha256"`
	Records      []application.DataTransferPreviewRecord `json:"records"`
}

type dataTransferApplyDB interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type dataTransferApplyResult struct {
	Applied int
	Skipped int
}

func (repository *DataTransferRepository) ApplyImport(
	ctx context.Context,
	job application.DataTransferJob,
	actor string,
) error {
	return repository.ApplyImportWithPolicy(ctx, job, actor, application.DataTransferImportPolicy{
		CanManageExhibitionLists: true,
	})
}

func (repository *DataTransferRepository) ApplyImportWithPolicy(
	ctx context.Context,
	job application.DataTransferJob,
	actor string,
	policy application.DataTransferImportPolicy,
) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin data transfer import: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	preview, issues, err := revalidateTransferPreview(ctx, tx, job)
	if err != nil {
		return err
	}
	result, err := applyTransferRecords(ctx, tx, job, preview.Records, issues, policy)
	if err != nil {
		return err
	}
	now := timestamp()
	completedState := application.TransferJobCompleted
	if job.WarningRecords > 0 || job.ErrorRecords > 0 || result.Skipped > 0 {
		completedState = application.TransferJobCompletedWithWarnings
	}
	update, err := tx.ExecContext(ctx, `
UPDATE data_transfer_jobs
SET state=?, stage='completed', confirmed_by_user_id=NULLIF(?, ''), confirmed_at=?, completed_at=?,
    result_message=?, revision=revision+1, updated_at=?
WHERE id=? AND direction=? AND state=? AND revision=? AND source_sha256=?`, completedState, actor, now, now,
		fmt.Sprintf("Imported %d record(s); skipped %d.", result.Applied, result.Skipped), now, job.ID,
		application.TransferImport, application.TransferJobReady, job.Revision, job.SourceSHA256)
	if err != nil {
		return fmt.Errorf("complete data transfer import: %w", err)
	}
	affected, err := update.RowsAffected()
	if err != nil {
		return fmt.Errorf("read data transfer completion result: %w", err)
	}
	if affected != 1 {
		return dataTransferApplyConflict("import job changed during apply")
	}
	if err := writeTransferAudit(ctx, tx, job, actor, now, result); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit data transfer import: %w", err)
	}
	return nil
}

func revalidateTransferPreview(
	ctx context.Context,
	tx *sql.Tx,
	job application.DataTransferJob,
) (dataTransferPersistedPreview, []application.DataTransferIssue, error) {
	var direction application.TransferDirection
	var state application.TransferJobState
	var revision, totalRecords int
	var sourceSHA, previewJSON string
	err := tx.QueryRowContext(ctx, `
SELECT direction, state, revision, total_records, source_sha256, preview_json
FROM data_transfer_jobs WHERE id=?`, job.ID).Scan(
		&direction, &state, &revision, &totalRecords, &sourceSHA, &previewJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import job no longer exists")
	}
	if err != nil {
		return dataTransferPersistedPreview{}, nil, fmt.Errorf("load data transfer import for apply: %w", err)
	}
	if direction != application.TransferImport || state != application.TransferJobReady ||
		revision != job.Revision || sourceSHA == "" || sourceSHA != job.SourceSHA256 {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import preview is stale")
	}
	preview := dataTransferPersistedPreview{}
	decoder := json.NewDecoder(strings.NewReader(previewJSON))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&preview); err != nil {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("persisted import preview is invalid")
	}
	if preview.SourceSHA256 == "" || preview.SourceSHA256 != sourceSHA || len(preview.Records) != totalRecords {
		return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import source or preview changed")
	}
	issues, err := transferApplyIssues(ctx, tx, job.ID)
	if err != nil {
		return dataTransferPersistedPreview{}, nil, err
	}
	for _, issue := range issues {
		if !validTransferApplyResolution(issue.Code, issue.SelectedResolution) {
			return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("import has unresolved conflicts")
		}
	}
	targetFingerprints, err := currentTransferTargetFingerprints(ctx, tx, preview.Records)
	if err != nil {
		return dataTransferPersistedPreview{}, nil, err
	}
	for _, record := range preview.Records {
		if !slices.Contains(job.Areas, record.Area) {
			return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("preview contains an unselected area")
		}
		if record.TargetID != "" {
			if record.TargetFingerprint == "" ||
				targetFingerprints[record.Area][record.TargetID] != record.TargetFingerprint {
				return dataTransferPersistedPreview{}, nil, dataTransferApplyConflict("preview target changed")
			}
		}
	}
	return preview, issues, nil
}

func transferApplyIssues(
	ctx context.Context,
	db dataTransferApplyDB,
	jobID string,
) ([]application.DataTransferIssue, error) {
	rows, err := db.QueryContext(ctx, dataTransferIssueSelect+`
WHERE job_id=? ORDER BY created_at, id`, jobID)
	if err != nil {
		return nil, fmt.Errorf("load data transfer issues for apply: %w", err)
	}
	defer func() { _ = rows.Close() }()
	issues := []application.DataTransferIssue{}
	for rows.Next() {
		issue, err := scanDataTransferIssue(rows)
		if err != nil {
			return nil, fmt.Errorf("scan data transfer issue for apply: %w", err)
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate data transfer issues for apply: %w", err)
	}
	return issues, nil
}

func validTransferApplyResolution(code, resolution string) bool {
	allowed := map[string]map[string]bool{
		"missing_inventory_number":             {"skip": true},
		"missing_manufacturer":                 {"skip": true},
		"missing_name":                         {"skip": true},
		"missing_gauge":                        {"skip": true},
		"missing_category":                     {"skip": true},
		"missing_gattung":                      {"skip": true},
		"invalid_vehicle":                      {"skip": true},
		"invalid_accessory":                    {"skip": true},
		"duplicate_inventory_number":           {"replace": true, "copy": true, "skip": true},
		"matching_manufacturer_article_number": {"use_existing": true, "create": true, "skip": true},
		"duplicate_exhibition_list":            {"replace": true, "merge": true, "copy": true, "skip": true},
		"locked_exhibition_list":               {"copy": true, "skip": true},
		"exhibition_vehicle_reference":         {"link": true, "skip": true},
		"missing_vehicle_reference":            {"skip": true},
		"duplicate_input_inventory_number":     {"skip": true},
	}
	return allowed[code][resolution]
}

func currentTransferTargetFingerprints(
	ctx context.Context,
	tx *sql.Tx,
	records []application.DataTransferPreviewRecord,
) (map[application.TransferArea]map[string]string, error) {
	wanted := map[application.TransferArea]bool{}
	for _, record := range records {
		if record.TargetID != "" {
			wanted[record.Area] = true
		}
	}
	fingerprints := map[application.TransferArea]map[string]string{}
	if wanted[application.TransferVehicles] {
		vehicles, err := transferVehicleSnapshot(ctx, tx)
		if err != nil {
			return nil, err
		}
		fingerprints[application.TransferVehicles] = map[string]string{}
		for _, vehicle := range vehicles {
			fingerprints[application.TransferVehicles][vehicle.ID] = application.DataTransferTargetFingerprint(vehicle)
		}
	}
	if wanted[application.TransferAccessories] {
		accessories, err := transferAccessorySnapshot(ctx, tx)
		if err != nil {
			return nil, err
		}
		fingerprints[application.TransferAccessories] = map[string]string{}
		for _, accessory := range accessories {
			fingerprints[application.TransferAccessories][accessory.ID] = application.DataTransferTargetFingerprint(accessory)
		}
	}
	if wanted[application.TransferExhibitionLists] {
		lists, err := transferExhibitionSnapshot(ctx, tx)
		if err != nil {
			return nil, err
		}
		fingerprints[application.TransferExhibitionLists] = map[string]string{}
		for _, list := range lists {
			fingerprints[application.TransferExhibitionLists][list.ID] = application.DataTransferTargetFingerprint(list)
		}
	}
	return fingerprints, nil
}

func applyTransferRecords(
	ctx context.Context,
	db *sql.Tx,
	job application.DataTransferJob,
	records []application.DataTransferPreviewRecord,
	issues []application.DataTransferIssue,
	policy application.DataTransferImportPolicy,
) (dataTransferApplyResult, error) {
	result := dataTransferApplyResult{}
	for _, record := range records {
		recordIssues := transferRecordIssues(issues, record)
		action, skip := transferRecordAction(record, recordIssues)
		if skip {
			result.Skipped++
			continue
		}
		if record.Area == application.TransferExhibitionLists && !policy.CanManageExhibitionLists &&
			action != "merge" {
			return dataTransferApplyResult{}, application.ErrDataTransferForbidden
		}
		var err error
		switch record.Area {
		case application.TransferVehicles:
			err = applyTransferVehicle(ctx, db, record, action)
		case application.TransferAccessories:
			err = applyTransferAccessory(ctx, db, record, action)
		case application.TransferExhibitionLists:
			err = applyTransferExhibitionList(ctx, db, record, action, recordIssues)
		default:
			err = dataTransferApplyConflict("preview contains an unsupported area")
		}
		if err != nil {
			return dataTransferApplyResult{}, fmt.Errorf("apply %s record %q: %w", record.Area, record.RecordKey, err)
		}
		result.Applied++
	}
	if result.Applied+result.Skipped != job.TotalRecords {
		return dataTransferApplyResult{}, dataTransferApplyConflict("preview record counts changed")
	}
	return result, nil
}

func transferRecordIssues(
	issues []application.DataTransferIssue,
	record application.DataTransferPreviewRecord,
) []application.DataTransferIssue {
	matched := []application.DataTransferIssue{}
	for _, issue := range issues {
		if issue.Area == record.Area && issue.RecordKey == record.RecordKey &&
			transferRowNumbersEqual(issue.RowNumber, record.RowNumber) {
			matched = append(matched, issue)
		}
	}
	return matched
}

func transferRowNumbersEqual(left, right *int) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func transferRecordAction(
	record application.DataTransferPreviewRecord,
	issues []application.DataTransferIssue,
) (string, bool) {
	action := record.ProposedAction
	for _, issue := range issues {
		switch issue.Code {
		case "missing_inventory_number", "missing_manufacturer", "missing_name", "missing_gauge",
			"missing_category", "missing_gattung", "invalid_vehicle", "invalid_accessory",
			"duplicate_input_inventory_number":
			return action, true
		case "duplicate_inventory_number", "matching_manufacturer_article_number", "duplicate_exhibition_list",
			"locked_exhibition_list":
			if issue.SelectedResolution == "skip" {
				return action, true
			}
			action = issue.SelectedResolution
		}
	}
	return action, false
}

func applyTransferVehicle(
	ctx context.Context,
	db dataTransferApplyDB,
	record application.DataTransferPreviewRecord,
	action string,
) error {
	vehicle := application.TransferVehicle{}
	if err := json.Unmarshal(record.Data, &vehicle); err != nil {
		return dataTransferApplyConflict("vehicle preview data is invalid")
	}
	if err := application.ValidateTransferVehicle(vehicle); err != nil {
		return dataTransferApplyConflict("vehicle preview data violates aggregate validation")
	}
	now := timestamp()
	if action == "copy" {
		inventoryNumber, err := uniqueTransferInventoryNumber(ctx, db, "vehicles", vehicle.InventoryNumber)
		if err != nil {
			return err
		}
		vehicle.InventoryNumber = inventoryNumber
		action = "create"
	}
	arguments := transferVehicleArguments(vehicle, now)
	switch action {
	case "create":
		_, err := db.ExecContext(ctx, `
INSERT INTO vehicles(
  id, inventory_number, manufacturer, article_number, article_source_url, name, gauge, epoch, railway_company,
  category, gattung, description, series, vehicle_number, maximum_speed_kmh, home_base, digital,
  digital_decoder_number, dt_decoder, dt_decoder_number, decoder_type, exhibition_ready, exhibition, abc_brakes,
  ean, production_period, list_price, acquisition_type, acquired_from, purchase_price, purchase_date,
  storage_location, storage_details, condition, condition_details, packaging, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			append([]any{randomID()}, arguments...)...)
		return err
	case "replace":
		if record.TargetID == "" {
			return dataTransferApplyConflict("vehicle replacement target is missing")
		}
		result, err := db.ExecContext(ctx, `
UPDATE vehicles SET inventory_number=?, manufacturer=?, article_number=?, article_source_url=?, name=?, gauge=?,
  epoch=?, railway_company=?, category=?, gattung=?, description=?, series=?, vehicle_number=?, maximum_speed_kmh=?,
  home_base=?, digital=?, digital_decoder_number=?, dt_decoder=?, dt_decoder_number=?, decoder_type=?,
  exhibition_ready=?, exhibition=?, abc_brakes=?, ean=?, production_period=?, list_price=?, acquisition_type=?,
  acquired_from=?, purchase_price=?, purchase_date=?, storage_location=?, storage_details=?, condition=?,
  condition_details=?, packaging=?, updated_at=? WHERE id=?`, append(arguments[0:36], record.TargetID)...)
		if err != nil {
			return err
		}
		return requireApplyUpdate(result, "replace vehicle")
	default:
		return dataTransferApplyConflict("unsupported vehicle resolution")
	}
}

func transferVehicleArguments(vehicle application.TransferVehicle, now string) []any {
	return []any{
		vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.ArticleNumber, vehicle.ArticleSourceURL, vehicle.Name,
		vehicle.Gauge, vehicle.Epoch, vehicle.RailwayCompany, vehicle.Category, vehicle.Gattung, vehicle.Description,
		vehicle.Series, vehicle.VehicleNumber, vehicle.MaximumSpeedKmh, vehicle.HomeBase, vehicle.Digital,
		vehicle.DigitalDecoderNumber, vehicle.DTDecoder, vehicle.DTDecoderNumber, vehicle.DecoderType,
		vehicle.ExhibitionReady, vehicle.Exhibition, vehicle.ABCBrakes, vehicle.EAN, vehicle.ProductionPeriod,
		vehicle.ListPrice, vehicle.AcquisitionType, vehicle.AcquiredFrom, vehicle.PurchasePrice, vehicle.PurchaseDate,
		vehicle.StorageLocation, vehicle.StorageDetails, vehicle.Condition, vehicle.ConditionDetails, vehicle.Packaging,
		now, now,
	}
}

func applyTransferAccessory(
	ctx context.Context,
	db *sql.Tx,
	record application.DataTransferPreviewRecord,
	action string,
) error {
	accessory := application.TransferAccessory{}
	if err := json.Unmarshal(record.Data, &accessory); err != nil {
		return dataTransferApplyConflict("accessory preview data is invalid")
	}
	if err := application.ValidateTransferAccessory(accessory); err != nil {
		return dataTransferApplyConflict("accessory preview data violates aggregate validation")
	}
	if accessory.ManufacturerStatus == "" {
		accessory.ManufacturerStatus = "unknown"
	}
	if accessory.Subtype == "" && accessory.ArticleType == string(domain.AccessoryArticleOther) {
		accessory.Subtype = "other:other"
	}
	referenceTargetID := record.TargetID
	if action == "copy" {
		referenceTargetID = ""
	}
	if err := validateTransferAccessoryReferenceData(ctx, db, accessory, referenceTargetID); err != nil {
		if errors.Is(err, application.ErrAccessoryValidation) {
			return dataTransferApplyConflict("accessory article type or subtype is not active")
		}
		return err
	}
	if action == "use_existing" {
		return nil
	}
	copying := action == "copy"
	if action == "copy" {
		inventoryNumber, err := uniqueTransferInventoryNumber(ctx, db, "accessory_products", accessory.InventoryNumber)
		if err != nil {
			return err
		}
		accessory.InventoryNumber = inventoryNumber
		action = "create"
	}
	now := timestamp()
	productID := record.TargetID
	if action == "create" {
		productID = randomID()
	} else if action != "replace" || productID == "" {
		return dataTransferApplyConflict("unsupported accessory resolution")
	}
	targetStrategy := domain.AccessoryInventoryStrategy(accessory.InventoryStrategy)
	if !targetStrategy.Valid() || domain.AccessoryTrackingMode(accessory.TrackingMode) != targetStrategy.TrackingMode() {
		return dataTransferApplyConflict("imported accessory inventory strategy is invalid")
	}
	if targetStrategy == domain.AccessoryInventoryQuantity && len(accessory.Assets) > 0 {
		return dataTransferApplyConflict("quantity accessory import cannot contain individual assets")
	}
	if action == "replace" {
		currentStrategy, err := accessoryInventoryStrategy(ctx, db, productID)
		if err != nil {
			return err
		}
		if err := validateAccessoryInventoryStrategyTransition(
			ctx, db, productID, currentStrategy, targetStrategy,
		); errors.Is(err, application.ErrAccessoryConflict) {
			return dataTransferApplyConflict("imported accessory inventory strategy conflicts with local allocations")
		} else if err != nil {
			return err
		}
	}
	stock, err := prepareTransferAccessoryStock(ctx, db, productID, targetStrategy, accessory.Stock, now, action == "replace")
	if err != nil {
		return err
	}
	gauges, _ := json.Marshal(nonNilStrings(accessory.Gauges))
	alternatives, _ := json.Marshal(nonNilStrings(accessory.AlternativeNumbers))
	keywords, _ := json.Marshal(nonNilStrings(accessory.Keywords))
	arguments := []any{
		accessory.InventoryNumber, accessory.Manufacturer, accessory.ArticleNumber, accessory.Name, accessory.Category,
		accessory.TrackingMode, accessory.Description, accessory.EAN, accessory.ManufacturerStatus, accessory.ArticleType,
		accessory.Subtype, string(gauges), accessory.Scale, accessory.ListPrice, accessory.PackageQuantity,
		accessory.StockUnit, accessory.MinimumStock, accessory.InventoryStrategy, accessory.ManufacturerURL,
		accessory.ProductURL, string(alternatives), string(keywords), accessory.CompatibilityNotes,
		accessory.InternalNotes, accessory.Archived,
	}
	if action == "create" {
		_, err := db.ExecContext(ctx, `
INSERT INTO accessory_products(
  id, inventory_number, manufacturer, article_number, name, category, tracking_mode, description, ean,
  manufacturer_status, article_type, subtype, gauges_json, scale, list_price, package_quantity, stock_unit,
  minimum_stock, inventory_strategy, manufacturer_url, product_url, alternative_numbers_json, keywords_json,
  compatibility_notes, internal_notes, archived, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			append(append([]any{productID}, arguments...), now, now)...)
		if err != nil {
			return err
		}
	} else {
		result, err := db.ExecContext(ctx, `
UPDATE accessory_products SET inventory_number=?, manufacturer=?, article_number=?, name=?, category=?,
  tracking_mode=?, description=?, ean=?, manufacturer_status=?, article_type=?, subtype=?, gauges_json=?, scale=?,
  list_price=?, package_quantity=?, stock_unit=?, minimum_stock=?, inventory_strategy=?, manufacturer_url=?,
  product_url=?, alternative_numbers_json=?, keywords_json=?, compatibility_notes=?, internal_notes=?, archived=?,
  updated_at=? WHERE id=?`, append(arguments, now, productID)...)
		if err != nil {
			return err
		}
		if err := requireApplyUpdate(result, "replace accessory"); err != nil {
			return err
		}
	}
	for _, level := range stock {
		if _, err := db.ExecContext(ctx, `
INSERT INTO accessory_stock(product_id, location_id, quantity, updated_at) VALUES(?, ?, ?, ?)
ON CONFLICT(product_id, location_id) DO UPDATE SET quantity=excluded.quantity, updated_at=excluded.updated_at`,
			productID, level.LocationID, level.Quantity, now); err != nil {
			return err
		}
	}
	for _, asset := range accessory.Assets {
		if copying && asset.InventoryNumber != "" {
			inventoryNumber, err := uniqueTransferInventoryNumber(
				ctx, db, "accessory_assets", asset.InventoryNumber,
			)
			if err != nil {
				return err
			}
			asset.InventoryNumber = inventoryNumber
		}
		locationID := ""
		var err error
		if asset.StorageLocationID != "" || asset.StorageLocationName != "" {
			locationID, err = transferStorageLocation(ctx, db, asset.StorageLocationID, asset.StorageLocationName, now)
			if err != nil {
				return err
			}
		}
		if action == "replace" {
			matchedID, err := matchingTransferAccessoryAsset(ctx, db, productID, asset)
			if err != nil {
				return err
			}
			if matchedID != "" {
				if err := updateTransferAccessoryAsset(ctx, db, matchedID, asset, locationID, now); err != nil {
					return err
				}
				continue
			}
		}
		if asset.Lifecycle == "reserved" || asset.Lifecycle == "installed" {
			return dataTransferApplyConflict("imported accessory asset has no matching local relationship")
		}
		if _, err := db.ExecContext(ctx, `
INSERT INTO accessory_assets(
  id, product_id, inventory_number, serial_number, condition_state, lifecycle_state, storage_location_id,
  purchase_date, purchase_price, warranty_until, notes, created_at, updated_at
) VALUES(?, ?, NULLIF(?, ''), ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?)`, randomID(), productID,
			asset.InventoryNumber, asset.SerialNumber, asset.Condition, asset.Lifecycle, locationID, asset.PurchaseDate,
			asset.PurchasePrice, asset.WarrantyUntil, asset.Notes, now, now); err != nil {
			return err
		}
	}
	return nil
}

type preparedTransferAccessoryStock struct {
	LocationID string
	Quantity   int
}

func prepareTransferAccessoryStock(
	ctx context.Context,
	db *sql.Tx,
	productID string,
	strategy domain.AccessoryInventoryStrategy,
	stock []application.TransferAccessoryStock,
	now string,
	validateReservations bool,
) ([]preparedTransferAccessoryStock, error) {
	prepared := make([]preparedTransferAccessoryStock, 0, len(stock))
	seenLocations := make(map[string]bool, len(stock))
	for _, level := range stock {
		if level.Quantity < 0 || strategy == domain.AccessoryInventoryIndividual && level.Quantity != 0 {
			return nil, dataTransferApplyConflict("imported accessory stock is incompatible with its inventory strategy")
		}
		locationID, err := transferStorageLocation(ctx, db, level.LocationID, level.LocationName, now)
		if err != nil {
			return nil, err
		}
		if seenLocations[locationID] {
			return nil, dataTransferApplyConflict("imported accessory stock contains a duplicate location")
		}
		seenLocations[locationID] = true
		if validateReservations {
			var reserved int
			if err := db.QueryRowContext(ctx, `
SELECT COALESCE(SUM(quantity), 0)
FROM accessory_reservations
WHERE product_id=? AND location_id=? AND asset_id IS NULL AND status='active'`,
				productID, locationID).Scan(&reserved); err != nil {
				return nil, fmt.Errorf("read reserved accessory stock before import: %w", err)
			}
			if level.Quantity < reserved {
				return nil, dataTransferApplyConflict("imported accessory stock is below its active reserved quantity")
			}
		}
		prepared = append(prepared, preparedTransferAccessoryStock{LocationID: locationID, Quantity: level.Quantity})
	}
	return prepared, nil
}

func matchingTransferAccessoryAsset(
	ctx context.Context,
	db dataTransferApplyDB,
	productID string,
	asset application.TransferAccessoryAsset,
) (string, error) {
	var id string
	if asset.ID != "" {
		err := db.QueryRowContext(ctx, `SELECT id FROM accessory_assets WHERE id=? AND product_id=?`,
			asset.ID, productID).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("match imported accessory asset by id: %w", err)
		}
	}
	if asset.InventoryNumber == "" {
		return "", nil
	}
	err := db.QueryRowContext(ctx, `
SELECT id FROM accessory_assets WHERE product_id=? AND inventory_number=? COLLATE NOCASE`,
		productID, asset.InventoryNumber).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("match imported accessory asset by inventory number: %w", err)
	}
	return id, nil
}

func updateTransferAccessoryAsset(
	ctx context.Context,
	db dataTransferApplyDB,
	assetID string,
	asset application.TransferAccessoryAsset,
	locationID string,
	now string,
) error {
	var activeReservations, activeInstallations int
	var reservationLocation string
	if err := db.QueryRowContext(ctx, `
SELECT
  (SELECT COUNT(*) FROM accessory_reservations WHERE asset_id=? AND status='active'),
	  (SELECT COUNT(*) FROM accessory_installations WHERE asset_id=? AND removed_at IS NULL),
  COALESCE((SELECT MAX(location_id) FROM accessory_reservations WHERE asset_id=? AND status='active'), '')`,
		assetID, assetID, assetID).Scan(&activeReservations, &activeInstallations, &reservationLocation); err != nil {
		return fmt.Errorf("check imported accessory asset relationships: %w", err)
	}
	if activeReservations > 0 && (asset.Lifecycle != "reserved" || locationID != reservationLocation) ||
		activeInstallations > 0 && (asset.Lifecycle != "installed" || locationID != "") ||
		asset.Lifecycle == "reserved" && activeReservations == 0 ||
		asset.Lifecycle == "installed" && activeInstallations == 0 {
		return dataTransferApplyConflict("imported accessory asset conflicts with a local relationship")
	}
	result, err := db.ExecContext(ctx, `
UPDATE accessory_assets
SET inventory_number=NULLIF(?, ''), serial_number=?, condition_state=?, lifecycle_state=?,
    storage_location_id=NULLIF(?, ''), purchase_date=?, purchase_price=?, warranty_until=?, notes=?, updated_at=?
WHERE id=?`, asset.InventoryNumber, asset.SerialNumber, asset.Condition, asset.Lifecycle, locationID,
		asset.PurchaseDate, asset.PurchasePrice, asset.WarrantyUntil, asset.Notes, now, assetID)
	if err != nil {
		return err
	}
	return requireApplyUpdate(result, "update imported accessory asset")
}

func transferStorageLocation(
	ctx context.Context,
	db dataTransferApplyDB,
	sourceID string,
	name string,
	now string,
) (string, error) {
	var id string
	if sourceID != "" {
		err := db.QueryRowContext(ctx, `SELECT id FROM storage_locations WHERE id=?`, sourceID).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", dataTransferApplyConflict("accessory storage location is missing")
	}
	err := db.QueryRowContext(ctx, `
SELECT id FROM storage_locations WHERE parent_id IS NULL AND name=? COLLATE NOCASE`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	id = randomID()
	if _, err := db.ExecContext(ctx, `
INSERT INTO storage_locations(id, name, created_at, updated_at) VALUES(?, ?, ?, ?)`, id, name, now, now); err != nil {
		return "", err
	}
	return id, nil
}

func applyTransferExhibitionList(
	ctx context.Context,
	db dataTransferApplyDB,
	record application.DataTransferPreviewRecord,
	action string,
	issues []application.DataTransferIssue,
) error {
	list := application.TransferExhibitionList{}
	if err := json.Unmarshal(record.Data, &list); err != nil {
		return dataTransferApplyConflict("exhibition preview data is invalid")
	}
	now := timestamp()
	normalizeTransferExhibitionList(&list)
	listID := record.TargetID
	if action == "replace" {
		var locked int
		if listID == "" || db.QueryRowContext(ctx, `SELECT locked FROM exhibition_lists WHERE id=?`, listID).Scan(&locked) != nil || locked != 0 {
			return dataTransferApplyConflict("locked exhibition list cannot be replaced")
		}
		result, err := db.ExecContext(ctx, `
UPDATE exhibition_lists SET
  designation=?, list_date=?, end_date=?, location=?, description=?, organization_notes=?,
  status=?, locked=?, lock_reason=?, locked_at=?, completed_at=?, archived_at=?, revision=revision+1, updated_at=?
WHERE id=? AND locked=0`,
			list.Designation, list.Date, list.EndDate, list.Location, list.Description, list.OrganizationNotes,
			list.Status, list.Locked, list.LockReason, list.LockedAt, list.CompletedAt, list.ArchivedAt, now, listID)
		if err != nil {
			return err
		}
		if err := requireApplyUpdate(result, "replace exhibition list"); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, `DELETE FROM exhibition_entries WHERE list_id=?`, listID); err != nil {
			return err
		}
	} else if action == "merge" {
		var locked int
		if listID == "" || db.QueryRowContext(ctx,
			`SELECT locked FROM exhibition_lists WHERE id=?`, listID).Scan(&locked) != nil || locked != 0 {
			return dataTransferApplyConflict("exhibition list cannot be merged")
		}
	} else if action == "create" || action == "copy" {
		listID = randomID()
		if _, err := db.ExecContext(ctx, `
INSERT INTO exhibition_lists(
  id, designation, list_date, end_date, location, description, organization_notes,
  status, revision, locked, lock_reason, locked_at, completed_at, archived_at, created_at, updated_at
)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?)`, listID, list.Designation, list.Date,
			list.EndDate, list.Location, list.Description, list.OrganizationNotes, list.Status, list.Locked,
			list.LockReason, list.LockedAt, list.CompletedAt, list.ArchivedAt, now, now); err != nil {
			return err
		}
	} else {
		return dataTransferApplyConflict("unsupported exhibition-list resolution")
	}
	for entryIndex, entry := range list.Entries {
		normalizeTransferExhibitionEntry(&entry)
		vehicleID, err := resolveTransferExhibitionVehicle(ctx, db, entry, issues, entryIndex)
		if err != nil {
			return err
		}
		if action == "merge" {
			updated, err := mergeTransferExhibitionEntry(ctx, db, listID, entry, vehicleID, now)
			if err != nil {
				return err
			}
			if updated {
				continue
			}
		}
		if _, err := db.ExecContext(ctx, `
INSERT INTO exhibition_entries(
  id, list_id, vehicle_id, owner, image_url, locomotive_name, gattung, series, manufacturer, epoch,
  railway_company, day_scope, dt_decoder, decoder_number, decoder_type, adapter, interface_name,
  sx_address, analog, availability, function_keys, notes, sort_order, revision, created_at, updated_at
) VALUES(?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?, ?,
		 NULLIF(?, ''), NULLIF(?, ''), ?, 1, ?, ?)`, randomID(), listID, vehicleID, entry.Owner, entry.ImageURL,
			entry.LocomotiveName, entry.Gattung, entry.Series, entry.Manufacturer, entry.Epoch, entry.RailwayCompany,
			entry.DayScope, entry.DTDecoder, entry.DecoderNumber, entry.DecoderType, entry.Adapter,
			entry.InterfaceName, entry.SXAddress, entry.Analog, entry.Availability, entry.FunctionKeys,
			entry.Notes, entry.SortOrder, now, now); err != nil {
			return err
		}
	}
	return nil
}

func mergeTransferExhibitionEntry(
	ctx context.Context,
	db dataTransferApplyDB,
	listID string,
	entry application.TransferExhibitionEntry,
	vehicleID string,
	now string,
) (bool, error) {
	if strings.TrimSpace(entry.ID) == "" {
		return false, nil
	}
	var owner, locomotiveName string
	err := db.QueryRowContext(ctx, `
SELECT owner, locomotive_name FROM exhibition_entries WHERE id=? AND list_id=?`, entry.ID, listID).
		Scan(&owner, &locomotiveName)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !strings.EqualFold(strings.TrimSpace(owner), strings.TrimSpace(entry.Owner)) ||
		!strings.EqualFold(strings.TrimSpace(locomotiveName), strings.TrimSpace(entry.LocomotiveName)) {
		return false, nil
	}
	result, err := db.ExecContext(ctx, `
UPDATE exhibition_entries SET
  vehicle_id=?, owner=?, image_url=NULLIF(?, ''), locomotive_name=?, gattung=?, series=?,
  manufacturer=?, epoch=?, railway_company=?, day_scope=?, dt_decoder=?, decoder_number=NULLIF(?, ''),
  decoder_type=?, adapter=?, interface_name=?, sx_address=?, analog=?, availability=?,
  function_keys=NULLIF(?, ''), notes=NULLIF(?, ''), sort_order=?, revision=revision+1, updated_at=?
WHERE id=? AND list_id=?`, vehicleID, entry.Owner, entry.ImageURL, entry.LocomotiveName, entry.Gattung,
		entry.Series, entry.Manufacturer, entry.Epoch, entry.RailwayCompany, entry.DayScope, entry.DTDecoder,
		entry.DecoderNumber, entry.DecoderType, entry.Adapter, entry.InterfaceName, entry.SXAddress, entry.Analog,
		entry.Availability, entry.FunctionKeys, entry.Notes, entry.SortOrder, now, entry.ID, listID)
	if err != nil {
		return false, err
	}
	if err := requireApplyUpdate(result, "merge exhibition entry"); err != nil {
		return false, err
	}
	return true, nil
}

func normalizeTransferExhibitionList(list *application.TransferExhibitionList) {
	if strings.TrimSpace(list.EndDate) == "" {
		list.EndDate = list.Date
	}
	if list.Status == "" {
		list.Status = application.ExhibitionStatusOpen
		if list.Locked {
			list.Status = application.ExhibitionStatusLocked
		}
	}
	list.Locked = list.Status == application.ExhibitionStatusLocked ||
		list.Status == application.ExhibitionStatusRunning
}

func normalizeTransferExhibitionEntry(entry *application.TransferExhibitionEntry) {
	if strings.TrimSpace(entry.InterfaceName) == "" {
		entry.InterfaceName = entry.Adapter
	}
	if strings.TrimSpace(entry.Availability) == "" {
		entry.Availability = "available"
	}
}

func resolveTransferExhibitionVehicle(
	ctx context.Context,
	db dataTransferApplyDB,
	entry application.TransferExhibitionEntry,
	issues []application.DataTransferIssue,
	entryIndex int,
) (string, error) {
	if entry.VehicleID == "" && entry.VehicleInventoryNumber == "" {
		return "", nil
	}
	var id, inventoryNumber string
	if entry.VehicleID != "" {
		err := db.QueryRowContext(ctx, `SELECT id, inventory_number FROM vehicles WHERE id=?`, entry.VehicleID).Scan(
			&id, &inventoryNumber,
		)
		if err == nil && strings.EqualFold(strings.TrimSpace(inventoryNumber), strings.TrimSpace(entry.VehicleInventoryNumber)) {
			return id, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	var issue *application.DataTransferIssue
	field := fmt.Sprintf("entries[%d].vehicleReference", entryIndex)
	for index := range issues {
		if issues[index].Field == field && (issues[index].Code == "exhibition_vehicle_reference" ||
			issues[index].Code == "missing_vehicle_reference") {
			issue = &issues[index]
			break
		}
	}
	if issue == nil {
		return "", dataTransferApplyConflict("exhibition vehicle resolution is missing")
	}
	if issue.SelectedResolution == "skip" {
		return "", nil
	}
	if entry.VehicleInventoryNumber != "" {
		err := db.QueryRowContext(ctx, `SELECT id FROM vehicles WHERE inventory_number=? COLLATE NOCASE`,
			entry.VehicleInventoryNumber).Scan(&id)
		if err == nil {
			if issue.Code == "exhibition_vehicle_reference" && issue.SelectedResolution == "link" {
				return id, nil
			}
			return "", dataTransferApplyConflict("exhibition vehicle resolution does not match preview")
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	return "", dataTransferApplyConflict("exhibition vehicle target changed")
}

func uniqueTransferInventoryNumber(
	ctx context.Context,
	db dataTransferApplyDB,
	table string,
	base string,
) (string, error) {
	query := map[string]string{
		"vehicles":           `SELECT EXISTS(SELECT 1 FROM vehicles WHERE inventory_number=? COLLATE NOCASE)`,
		"accessory_products": `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE inventory_number=? COLLATE NOCASE)`,
		"accessory_assets":   `SELECT EXISTS(SELECT 1 FROM accessory_assets WHERE inventory_number=? COLLATE NOCASE)`,
	}[table]
	if query == "" {
		return "", dataTransferApplyConflict("unsupported copied record")
	}
	base = strings.TrimSpace(base) + "-COPY"
	for suffix := 1; suffix <= 10_000; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		var exists int
		if err := db.QueryRowContext(ctx, query, candidate).Scan(&exists); err != nil {
			return "", err
		}
		if exists == 0 {
			return candidate, nil
		}
	}
	return "", dataTransferApplyConflict("could not allocate copy inventory number")
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func requireApplyUpdate(result sql.Result, operation string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s rows affected: %w", operation, err)
	}
	if affected != 1 {
		return dataTransferApplyConflict(operation + " target changed")
	}
	return nil
}

func writeTransferAudit(
	ctx context.Context,
	db dataTransferApplyDB,
	job application.DataTransferJob,
	actor string,
	createdAt string,
	result dataTransferApplyResult,
) error {
	details, err := json.Marshal(map[string]any{
		"areas": job.Areas, "appliedRecords": result.Applied, "skippedRecords": result.Skipped,
		"sourceSha256": job.SourceSHA256, "totalRecords": job.TotalRecords,
	})
	if err != nil {
		return fmt.Errorf("encode data transfer audit: %w", err)
	}
	if _, err := db.ExecContext(ctx, `
INSERT INTO audit_logs(id, actor_user_id, action, target_type, target_id, created_at, details_json)
VALUES(?, NULLIF(?, ''), 'data_transfer.import_applied', 'data_transfer_job', ?, ?, ?)`,
		randomID(), actor, job.ID, createdAt, string(details)); err != nil {
		return fmt.Errorf("write data transfer audit: %w", err)
	}
	return nil
}

func dataTransferApplyConflict(message string) error {
	return fmt.Errorf("%w: %s", application.ErrDataTransferConflict, message)
}
