package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"railkeeper/backend/internal/domain"
)

const (
	backupFormat  = "railkeeper-backup"
	backupVersion = 15
)

var (
	ErrBackupInvalid = errors.New("backup invalid")
	ErrBackupPath    = errors.New("backup path invalid")
)

type BackupService struct {
	db      *sql.DB
	dataDir string
}

type BackupDocument struct {
	Format    string                      `json:"format"`
	Version   int                         `json:"version"`
	CreatedAt string                      `json:"createdAt"`
	Tables    map[string][]map[string]any `json:"tables"`
	Files     []BackupFile                `json:"files"`
}

type BackupFile struct {
	Path          string `json:"path"`
	SizeBytes     int64  `json:"sizeBytes"`
	SHA256        string `json:"sha256"`
	ContentBase64 string `json:"contentBase64"`
}

type BackupImportResult struct {
	RestoredTables int `json:"restoredTables"`
	RestoredRows   int `json:"restoredRows"`
	RestoredFiles  int `json:"restoredFiles"`
}

type BackupValidationResult struct {
	Compatible bool                    `json:"compatible"`
	Format     string                  `json:"format,omitempty"`
	Version    int                     `json:"version"`
	CreatedAt  string                  `json:"createdAt,omitempty"`
	TableCount int                     `json:"tableCount"`
	RowCount   int                     `json:"rowCount"`
	FileCount  int                     `json:"fileCount"`
	FileBytes  int64                   `json:"fileBytes"`
	Tables     []BackupValidationTable `json:"tables"`
	Warnings   []string                `json:"warnings"`
	Errors     []string                `json:"errors"`
}

type BackupValidationTable struct {
	Name           string   `json:"name"`
	Rows           int      `json:"rows"`
	Missing        bool     `json:"missing"`
	UnknownColumns []string `json:"unknownColumns,omitempty"`
}

var backupTableOrder = []string{
	"master_data_entries",
	"master_data_relations",
	"inventory_number_schemes",
	"file_blobs",
	"storage_locations",
	"vehicles",
	"inventory_number_history",
	"vehicle_external_mappings",
	"vehicle_images",
	"vehicle_attachments",
	"vehicle_maintenance",
	"vehicle_spare_parts",
	"vehicle_functions",
	"vehicle_cv_files",
	"vehicle_cv_values",
	"vehicle_cv_value_history",
	"accessory_products",
	"accessory_product_attributes",
	"accessory_purchases",
	"accessory_stock",
	"accessory_stock_movements",
	"accessory_assets",
	"accessory_documents",
	"layouts",
	"layout_units",
	"layout_unit_ports",
	"layout_unit_outline_points",
	"layout_technical_positions",
	"plan_variants",
	"plan_revisions",
	"track_geometry_libraries",
	"track_geometry_definitions",
	"plan_track_objects",
	"plan_free_objects",
	"layout_configurations",
	"layout_configuration_units",
	"accessory_reservations",
	"accessory_reservation_positions",
	"plan_track_object_reservations",
	"accessory_installations",
	"accessory_installation_positions",
	"accessory_installation_condition_history",
	"exhibition_lists",
	"exhibition_entries",
}

type backupTableVersionPolicy struct {
	introduced int
	required   int
}

var backupTableVersions = map[string]backupTableVersionPolicy{
	"master_data_entries":                      {introduced: 1, required: 1},
	"master_data_relations":                    {introduced: 1, required: 1},
	"inventory_number_schemes":                 {introduced: 1, required: 1},
	"file_blobs":                               {introduced: 1, required: 3},
	"storage_locations":                        {introduced: 2, required: 2},
	"vehicles":                                 {introduced: 1, required: 1},
	"inventory_number_history":                 {introduced: 1, required: 1},
	"vehicle_external_mappings":                {introduced: 1, required: 3},
	"vehicle_images":                           {introduced: 1, required: 1},
	"vehicle_attachments":                      {introduced: 1, required: 1},
	"vehicle_maintenance":                      {introduced: 1, required: 1},
	"vehicle_spare_parts":                      {introduced: 1, required: 3},
	"vehicle_functions":                        {introduced: 1, required: 1},
	"vehicle_cv_files":                         {introduced: 1, required: 1},
	"vehicle_cv_values":                        {introduced: 1, required: 1},
	"vehicle_cv_value_history":                 {introduced: 1, required: 1},
	"accessory_products":                       {introduced: 2, required: 2},
	"accessory_product_attributes":             {introduced: 3, required: 3},
	"accessory_purchases":                      {introduced: 3, required: 3},
	"accessory_stock":                          {introduced: 2, required: 2},
	"accessory_stock_movements":                {introduced: 3, required: 3},
	"accessory_assets":                         {introduced: 2, required: 2},
	"accessory_documents":                      {introduced: 3, required: 3},
	"layouts":                                  {introduced: 2, required: 2},
	"layout_units":                             {introduced: 2, required: 2},
	"layout_unit_ports":                        {introduced: 7, required: 7},
	"layout_unit_outline_points":               {introduced: 4, required: 4},
	"layout_technical_positions":               {introduced: 4, required: 4},
	"plan_variants":                            {introduced: 2, required: 2},
	"plan_revisions":                           {introduced: 2, required: 2},
	"track_geometry_libraries":                 {introduced: 5, required: 5},
	"track_geometry_definitions":               {introduced: 5, required: 5},
	"plan_track_objects":                       {introduced: 5, required: 5},
	"plan_free_objects":                        {introduced: 13, required: 13},
	"layout_configurations":                    {introduced: 2, required: 2},
	"layout_configuration_units":               {introduced: 2, required: 2},
	"accessory_reservations":                   {introduced: 2, required: 2},
	"accessory_reservation_positions":          {introduced: 4, required: 4},
	"plan_track_object_reservations":           {introduced: 6, required: 6},
	"accessory_installations":                  {introduced: 2, required: 2},
	"accessory_installation_positions":         {introduced: 4, required: 4},
	"accessory_installation_condition_history": {introduced: 3, required: 3},
	"exhibition_lists":                         {introduced: 1, required: 3},
	"exhibition_entries":                       {introduced: 1, required: 3},
}

func NewBackupService(db *sql.DB, dataDir string) *BackupService {
	return &BackupService{db: db, dataDir: dataDir}
}

func (s *BackupService) Export(ctx context.Context) (*BackupDocument, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("backup service is not configured")
	}

	doc := &BackupDocument{
		Format:    backupFormat,
		Version:   backupVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Tables:    map[string][]map[string]any{},
		Files:     []BackupFile{},
	}

	for _, table := range backupTableOrder {
		rows, err := s.exportTable(ctx, table)
		if err != nil {
			return nil, err
		}
		doc.Tables[table] = rows
	}

	files, err := s.exportFiles()
	if err != nil {
		return nil, err
	}
	doc.Files = files

	return doc, nil
}

func (s *BackupService) Import(ctx context.Context, doc *BackupDocument) (*BackupImportResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("backup service is not configured")
	}
	if doc == nil {
		return nil, ErrBackupInvalid
	}
	if err := validateBackupFiles(doc.Files); err != nil {
		return nil, err
	}
	validation, err := s.Validate(ctx, doc)
	if err != nil {
		return nil, err
	}
	if !validation.Compatible {
		return nil, ErrBackupInvalid
	}

	stagedFiles, err := s.stageRestoreFiles(doc.Files)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(stagedFiles.root) }()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin backup restore: %w", err)
	}
	if _, err = tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("defer backup restore foreign keys: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()
	articleMasterData, err := readLegacyRestoreArticleMasterData(ctx, tx, doc.Version)
	if err != nil {
		return nil, err
	}
	articleInventoryScheme, err := readBackupArticleInventoryScheme(ctx, tx)
	if err != nil {
		return nil, err
	}

	for i := len(backupTableOrder) - 1; i >= 0; i-- {
		table := backupTableOrder[i]
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+quoteIdentifier(table)); err != nil {
			return nil, fmt.Errorf("clear %s: %w", table, err)
		}
	}

	result := &BackupImportResult{}
	for _, table := range backupTableOrder {
		if table == "accessory_products" {
			if err := prepareBackupArticleInventoryNumbers(ctx, tx, doc, articleInventoryScheme); err != nil {
				return nil, err
			}
		}
		rows := backupRowsForRestore(doc, table)
		if len(rows) == 0 {
			continue
		}
		columns, err := tableColumns(ctx, tx, table)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			inserted, err := insertBackupRow(ctx, tx, table, columns, row)
			if err != nil {
				return nil, err
			}
			if inserted {
				result.RestoredRows++
			}
		}
		result.RestoredTables++
	}
	if err := restoreLegacyArticleMasterData(ctx, tx, articleMasterData); err != nil {
		return nil, err
	}
	if err := backfillRestoredTrackGeometrySnapshots(ctx, tx); err != nil {
		return nil, err
	}
	if err := restoreLegacyAccessoryProductDefaults(ctx, tx, doc); err != nil {
		return nil, err
	}

	uploadsSwap, err := s.replaceUploadsWithStaged(stagedFiles)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		uploadsSwap.rollback()
		return nil, fmt.Errorf("commit backup restore: %w", err)
	}
	committed = true
	uploadsSwap.cleanup()
	result.RestoredFiles = stagedFiles.restoredFiles

	return result, nil
}

func backupRowsForRestore(doc *BackupDocument, table string) []map[string]any {
	rows := doc.Tables[table]
	legacyColumns := []string{}
	if doc.Version <= 10 {
		legacyColumn := map[string]string{
			"track_geometry_definitions": "minimum_radius_mm",
			"plan_track_objects":         "flex_path_json",
			"layouts":                    "minimum_flex_radius_mm",
		}[table]
		if legacyColumn != "" {
			legacyColumns = append(legacyColumns, legacyColumn)
		}
	}
	if doc.Version <= 11 && table == "plan_track_objects" {
		legacyColumns = append(legacyColumns, "transition_path_json")
	}
	if doc.Version <= 13 && table == "plan_track_objects" {
		legacyColumns = append(legacyColumns, "geometry_snapshot_json")
	}
	if len(legacyColumns) == 0 {
		return rows
	}
	normalized := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		copyRow := make(map[string]any, len(row)+len(legacyColumns))
		for column, value := range row {
			copyRow[column] = value
		}
		for _, legacyColumn := range legacyColumns {
			if _, exists := copyRow[legacyColumn]; !exists {
				copyRow[legacyColumn] = nil
			}
		}
		normalized = append(normalized, copyRow)
	}
	return normalized
}

func backfillRestoredTrackGeometrySnapshots(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
UPDATE plan_track_objects
SET geometry_snapshot_json = (
  SELECT json_object(
    'id', geometry.id,
    'libraryId', geometry.library_id,
    'articleNumber', geometry.article_number,
    'name', geometry.name,
    'kind', geometry.kind,
    'lengthMm', geometry.length_mm,
    'minimumRadiusMm', geometry.minimum_radius_mm,
    'geometry', json(geometry.geometry_json),
    'sourceUrl', geometry.source_url,
    'status', geometry.status,
    'createdAt', geometry.created_at
  )
  FROM track_geometry_definitions geometry
  WHERE geometry.id = plan_track_objects.geometry_id
)
WHERE geometry_snapshot_json IS NULL`); err != nil {
		return fmt.Errorf("backfill restored track geometry snapshots: %w", err)
	}
	return nil
}

func restoreLegacyAccessoryProductDefaults(
	ctx context.Context,
	tx *sql.Tx,
	doc *BackupDocument,
) error {
	if doc.Version >= 3 {
		return nil
	}
	for _, row := range doc.Tables["accessory_products"] {
		productID, valid := backupNonEmptyString(row["id"])
		if !valid {
			continue
		}
		if _, explicit := row["article_type"]; !explicit {
			if _, err := tx.ExecContext(ctx,
				`UPDATE accessory_products SET article_type='other' WHERE id=?`, productID); err != nil {
				return fmt.Errorf("backfill restored legacy accessory article type: %w", err)
			}
		}
		if _, explicit := row["subtype"]; !explicit {
			if _, err := tx.ExecContext(ctx,
				`UPDATE accessory_products SET subtype=category WHERE id=?`, productID); err != nil {
				return fmt.Errorf("backfill restored legacy accessory subtype: %w", err)
			}
		}
		if _, explicit := row["inventory_strategy"]; !explicit {
			if _, err := tx.ExecContext(ctx, `
UPDATE accessory_products
SET inventory_strategy=CASE tracking_mode WHEN 'individual' THEN 'individual' ELSE 'quantity' END
WHERE id=?`, productID); err != nil {
				return fmt.Errorf("backfill restored legacy accessory inventory strategy: %w", err)
			}
		}
	}
	return nil
}

func (s *BackupService) Validate(ctx context.Context, doc *BackupDocument) (*BackupValidationResult, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("backup service is not configured")
	}

	result := &BackupValidationResult{
		Tables:   []BackupValidationTable{},
		Warnings: []string{},
		Errors:   []string{},
	}
	if doc == nil {
		result.Errors = append(result.Errors, "Backup-Dokument fehlt.")
		return finishBackupValidation(result), nil
	}

	result.Format = doc.Format
	result.Version = doc.Version
	result.CreatedAt = doc.CreatedAt
	result.FileCount = len(doc.Files)
	for _, file := range doc.Files {
		result.FileBytes += file.SizeBytes
	}
	if doc.Format != backupFormat {
		result.Errors = append(result.Errors, "Backup-Format wird nicht unterstützt.")
	}
	if doc.Version < 1 || doc.Version > backupVersion {
		result.Errors = append(result.Errors, "Backup-Version wird nicht unterstützt.")
	}
	if err := validateBackupFiles(doc.Files); err != nil {
		result.Errors = append(result.Errors, "Backup-Dateien sind unvollständig oder beschädigt.")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin backup validation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	knownTables := map[string]struct{}{}
	for _, table := range backupTableOrder {
		knownTables[table] = struct{}{}
		rows, exists := doc.Tables[table]
		item := BackupValidationTable{Name: table, Rows: len(rows), Missing: !exists}
		if !exists {
			if backupTableOptional(doc.Version, table) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Optionale Tabelle %s fehlt im Backup und wird leer wiederhergestellt.", table))
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("Tabelle %s fehlt im Backup.", table))
			}
			result.Tables = append(result.Tables, item)
			continue
		}
		result.TableCount++
		result.RowCount += len(rows)

		columns, err := tableColumns(ctx, tx, table)
		if err != nil {
			return nil, err
		}
		unknownColumns := map[string]struct{}{}
		for _, row := range rows {
			for column := range row {
				if !columns[column] {
					unknownColumns[column] = struct{}{}
				}
			}
		}
		item.UnknownColumns = sortedKeys(unknownColumns)
		if len(item.UnknownColumns) > 0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Tabelle %s enthält unbekannte Spalten, die beim Restore ignoriert werden.", table))
		}
		if table == "accessory_product_attributes" {
			result.Errors = append(result.Errors,
				validateBackupAccessoryProductAttributes(
					rows,
					doc.Tables["accessory_products"],
					doc.Tables["master_data_entries"],
					doc.Version,
				)...)
		}
		if table == "master_data_entries" && doc.Version >= 3 {
			result.Errors = append(result.Errors, validateBackupProtectedArticleTypes(rows)...)
		}
		result.Tables = append(result.Tables, item)
	}
	for table := range doc.Tables {
		if _, ok := knownTables[table]; !ok {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Unbekannte Tabelle %s wird beim Restore ignoriert.", table))
		}
	}

	return finishBackupValidation(result), nil
}

func validateBackupProtectedArticleTypes(rows []map[string]any) []string {
	entries := make([]MasterDataEntry, 0, len(standardArticleTypeKeys))
	for _, row := range rows {
		typeName, valid := backupNonEmptyString(row["type"])
		if !valid || typeName != standardArticleType {
			continue
		}
		key, _ := backupNonEmptyString(row["key"])
		label, _ := backupNonEmptyString(row["label"])
		entries = append(entries, MasterDataEntry{Type: typeName, Key: key, Label: label})
	}
	if err := validateProtectedArticleTypes(entries, true); err != nil {
		return []string{"Tabelle master_data_entries enthält keine gültige vollständige Artikelarten-Konfiguration."}
	}
	return nil
}

func backupTableOptional(version int, table string) bool {
	policy, known := backupTableVersions[table]
	return known && (version < policy.introduced || version < policy.required)
}

func validateBackupAccessoryProductAttributes(
	attributeRows []map[string]any,
	productRows []map[string]any,
	masterDataRows []map[string]any,
	backupVersion int,
) []string {
	productTypes := make(map[string]domain.AccessoryArticleType, len(productRows))
	for _, row := range productRows {
		productID, productIDValid := backupNonEmptyString(row["id"])
		articleType, articleTypeValid := backupNonEmptyString(row["article_type"])
		if productIDValid && articleTypeValid {
			productTypes[productID] = domain.AccessoryArticleType(articleType)
		}
	}

	validationErrors := []string{}
	controlledDefinitions := []domain.AccessoryAttributeDefinition{}
	if backupVersion >= 3 {
		controlledDefinitions, validationErrors = backupControlledAccessoryAttributeDefinitions(masterDataRows)
	}
	attributesByProduct := map[string][]domain.AccessoryAttributeValue{}
	for index, row := range attributeRows {
		if err := validateBackupAccessoryProductAttribute(row); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes enthält in Zeile %d ungültige Attributdaten: %v",
				index+1,
				err,
			))
			continue
		}
		productID, _ := backupNonEmptyString(row["product_id"])
		attributesByProduct[productID] = append(attributesByProduct[productID],
			backupDomainAccessoryAttribute(row))
	}

	productIDs := make([]string, 0, len(attributesByProduct))
	for productID := range attributesByProduct {
		productIDs = append(productIDs, productID)
	}
	sort.Strings(productIDs)
	for _, productID := range productIDs {
		articleType, exists := productTypes[productID]
		if !exists {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes verweist auf Produkt %s ohne gültigen article_type.",
				productID,
			))
			continue
		}
		var err error
		if backupVersion >= 3 && articleType == domain.AccessoryArticleOther {
			err = validateBackupControlledAccessoryAttributeValues(
				attributesByProduct[productID],
				controlledDefinitions,
			)
		} else {
			err = domain.ValidateAccessoryAttributeValues(articleType, attributesByProduct[productID])
		}
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes enthält für Produkt %s ungültige Attributdaten: %v",
				productID,
				err,
			))
		}
	}
	return validationErrors
}

func backupControlledAccessoryAttributeDefinitions(
	rows []map[string]any,
) ([]domain.AccessoryAttributeDefinition, []string) {
	definitions := []domain.AccessoryAttributeDefinition{}
	validationErrors := []string{}
	seen := map[string]struct{}{}
	for index, row := range rows {
		typeName, valid := backupNonEmptyString(row["type"])
		if !valid || typeName != accessoryCustomField {
			continue
		}
		key, keyValid := backupNonEmptyString(row["key"])
		active, activeValid := backupBooleanValue(row["active"])
		metadataJSON, metadataValid := row["metadata_json"].(string)
		metadata := map[string]any{}
		var metadataErr error
		if metadataValid {
			metadataErr = json.Unmarshal([]byte(metadataJSON), &metadata)
		}
		if !keyValid || !activeValid || !metadataValid || metadataErr != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes kann wegen ungültiger Custom-Field-Definition in Zeile %d nicht validiert werden.",
				index+1,
			))
			continue
		}
		if _, duplicate := seen[key]; duplicate {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes enthält eine doppelte Custom-Field-Definition %s.", key,
			))
			continue
		}
		seen[key] = struct{}{}
		definition, err := ParseAccessoryCustomAttributeDefinition(key, active, metadata)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf(
				"Tabelle accessory_product_attributes kann Custom-Field-Definition %s nicht validieren: %v",
				key,
				err,
			))
			continue
		}
		definitions = append(definitions, definition)
	}
	return definitions, validationErrors
}

func validateBackupControlledAccessoryAttributeValues(
	values []domain.AccessoryAttributeValue,
	definitions []domain.AccessoryAttributeDefinition,
) error {
	if err := domain.ValidateAccessoryAttributeValues(domain.AccessoryArticleOther, values); err != nil {
		return err
	}
	activeDefinitions := make([]domain.AccessoryAttributeDefinition, 0, len(definitions))
	activeKeys := make(map[string]struct{}, len(definitions))
	inactiveKeys := make(map[string]struct{}, len(definitions))
	for _, definition := range definitions {
		if definition.Active {
			activeDefinitions = append(activeDefinitions, definition)
			activeKeys[definition.Key] = struct{}{}
		} else {
			inactiveKeys[definition.Key] = struct{}{}
		}
	}
	activeValues := make([]domain.AccessoryAttributeValue, 0, len(values))
	for _, value := range values {
		if _, active := activeKeys[value.Key]; active {
			activeValues = append(activeValues, value)
			continue
		}
		if _, inactive := inactiveKeys[value.Key]; !inactive {
			return fmt.Errorf("%w: undefined controlled key %q",
				domain.ErrAccessoryAttributeValidation, value.Key)
		}
	}
	return domain.ValidateControlledAccessoryAttributeValues(activeValues, activeDefinitions)
}

func backupDomainAccessoryAttribute(row map[string]any) domain.AccessoryAttributeValue {
	key, _ := backupNonEmptyString(row["attribute_key"])
	valueType, _ := backupNonEmptyString(row["value_type"])
	attribute := domain.AccessoryAttributeValue{
		Key:  key,
		Kind: domain.AccessoryAttributeKind(valueType),
	}
	if unit, ok := row["unit"].(string); ok {
		attribute.Unit = &unit
	}
	switch attribute.Kind {
	case domain.AccessoryAttributeText:
		value := row["text_value"].(string)
		attribute.TextValue = &value
	case domain.AccessoryAttributeNumber:
		value, _ := backupNumberValue(row["number_value"])
		attribute.NumberValue = &value
	case domain.AccessoryAttributeBoolean:
		value, _ := backupBooleanValue(row["boolean_value"])
		attribute.BooleanValue = &value
	case domain.AccessoryAttributeDate:
		value := row["date_value"].(string)
		attribute.DateValue = &value
	case domain.AccessoryAttributeSingleSelect:
		attribute.OptionValues = []string{row["single_select_value"].(string)}
	case domain.AccessoryAttributeMultiSelect:
		_ = json.Unmarshal([]byte(row["multi_select_value"].(string)), &attribute.OptionValues)
	}
	return attribute
}

func backupNumberValue(value any) (float64, bool) {
	switch typed := value.(type) {
	case json.Number:
		number, err := typed.Float64()
		return number, err == nil && !math.IsNaN(number) && !math.IsInf(number, 0)
	case float64:
		return typed, !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case float32:
		return float64(typed), !math.IsNaN(float64(typed)) && !math.IsInf(float64(typed), 0)
	case int:
		return float64(typed), true
	case int8:
		return float64(typed), true
	case int16:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case uint:
		return float64(typed), true
	case uint8:
		return float64(typed), true
	case uint16:
		return float64(typed), true
	case uint32:
		return float64(typed), true
	case uint64:
		return float64(typed), true
	default:
		return 0, false
	}
}

func backupBooleanValue(value any) (bool, bool) {
	if typed, ok := value.(bool); ok {
		return typed, true
	}
	number, valid := backupNumberValue(value)
	if !valid || (number != 0 && number != 1) {
		return false, false
	}
	return number == 1, true
}

func validateBackupAccessoryProductAttribute(row map[string]any) error {
	if _, valid := backupNonEmptyString(row["product_id"]); !valid {
		return errors.New("product_id fehlt oder ist ungültig")
	}
	if _, valid := backupNonEmptyString(row["attribute_key"]); !valid {
		return errors.New("attribute_key fehlt oder ist ungültig")
	}
	valueType, valueTypeValid := backupNonEmptyString(row["value_type"])
	valueColumnByType := map[string]string{
		"text":          "text_value",
		"number":        "number_value",
		"boolean":       "boolean_value",
		"date":          "date_value",
		"single_select": "single_select_value",
		"multi_select":  "multi_select_value",
	}
	matchingColumn, supported := valueColumnByType[valueType]
	if !valueTypeValid || !supported {
		return fmt.Errorf("value_type %q wird nicht unterstützt", valueType)
	}
	valueColumns := []string{
		"text_value", "number_value", "boolean_value", "date_value", "single_select_value", "multi_select_value",
	}
	for _, column := range valueColumns {
		if column == matchingColumn {
			if row[column] == nil {
				return fmt.Errorf("%s muss gesetzt sein", column)
			}
			continue
		}
		if row[column] != nil {
			return fmt.Errorf("%s darf für value_type %q nicht gesetzt sein", column, valueType)
		}
	}
	if row["unit"] != nil {
		if valueType != "number" {
			return fmt.Errorf("unit ist nur für value_type number erlaubt")
		}
		if _, ok := row["unit"].(string); !ok {
			return errors.New("unit muss eine Zeichenfolge sein")
		}
	}

	value := row[matchingColumn]
	switch valueType {
	case "text", "date", "single_select":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("%s muss eine Zeichenfolge sein", matchingColumn)
		}
	case "number":
		if !validBackupNumber(value) {
			return errors.New("number_value muss eine endliche Zahl sein")
		}
	case "boolean":
		if !validBackupBoolean(value) {
			return errors.New("boolean_value muss true, false, 0 oder 1 sein")
		}
	case "multi_select":
		encoded, ok := value.(string)
		if !ok {
			return errors.New("multi_select_value muss ein JSON-String sein")
		}
		var options []string
		if err := json.Unmarshal([]byte(encoded), &options); err != nil || len(options) == 0 {
			return errors.New("multi_select_value muss ein nicht-leeres JSON-Array aus Zeichenfolgen sein")
		}
	}
	return nil
}

func backupNonEmptyString(value any) (string, bool) {
	text, ok := value.(string)
	text = strings.TrimSpace(text)
	return text, ok && text != ""
}

func validBackupNumber(value any) bool {
	_, valid := backupNumberValue(value)
	return valid
}

func validBackupBoolean(value any) bool {
	_, valid := backupBooleanValue(value)
	return valid
}

func (s *BackupService) exportTable(ctx context.Context, table string) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM "+quoteIdentifier(table))
	if err != nil {
		return nil, fmt.Errorf("export %s: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("read %s columns: %w", table, err)
	}

	out := []map[string]any{}
	for rows.Next() {
		values := make([]any, len(columns))
		targets := make([]any, len(columns))
		for i := range values {
			targets[i] = &values[i]
		}
		if err := rows.Scan(targets...); err != nil {
			return nil, fmt.Errorf("scan %s: %w", table, err)
		}
		row := map[string]any{}
		for i, column := range columns {
			row[column] = normalizeBackupValue(table, column, values[i])
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s: %w", table, err)
	}
	return out, nil
}

func (s *BackupService) exportFiles() ([]BackupFile, error) {
	uploadsDir := filepath.Join(s.dataDir, "uploads")
	if _, err := os.Stat(uploadsDir); errors.Is(err, os.ErrNotExist) {
		return []BackupFile{}, nil
	}

	files := []BackupFile{}
	if err := filepath.WalkDir(uploadsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(s.dataDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if err := validateBackupFilePath(relative); err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		files = append(files, BackupFile{
			Path:          relative,
			SizeBytes:     int64(len(data)),
			SHA256:        hex.EncodeToString(sum[:]),
			ContentBase64: base64.StdEncoding.EncodeToString(data),
		})
		return nil
	}); err != nil {
		return nil, fmt.Errorf("export backup files: %w", err)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

type stagedRestoreFiles struct {
	root          string
	uploadsDir    string
	restoredFiles int
}

func (s *BackupService) stageRestoreFiles(files []BackupFile) (*stagedRestoreFiles, error) {
	if err := os.MkdirAll(s.dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("prepare restore data directory: %w", err)
	}
	root := filepath.Join(s.dataDir, ".restore-staging-"+randomID())
	uploadsDir := filepath.Join(root, "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create restore staging directory: %w", err)
	}

	staged := &stagedRestoreFiles{root: root, uploadsDir: uploadsDir}
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.RemoveAll(root)
		}
	}()

	for _, file := range files {
		if err := validateBackupFilePath(file.Path); err != nil {
			return nil, err
		}
		data, err := base64.StdEncoding.DecodeString(file.ContentBase64)
		if err != nil {
			return nil, ErrBackupInvalid
		}
		sum := sha256.Sum256(data)
		if file.SHA256 != "" && !strings.EqualFold(file.SHA256, hex.EncodeToString(sum[:])) {
			return nil, ErrBackupInvalid
		}
		relative := strings.TrimPrefix(pathClean(file.Path), "uploads/")
		absTarget, err := confinedChildPath(uploadsDir, relative)
		if err != nil {
			return nil, err
		}
		if err := os.MkdirAll(filepath.Dir(absTarget), 0o755); err != nil {
			return nil, fmt.Errorf("create restore staging directory: %w", err)
		}
		if err := os.WriteFile(absTarget, data, 0o600); err != nil {
			return nil, fmt.Errorf("stage restore file: %w", err)
		}
		staged.restoredFiles++
	}
	cleanup = false
	return staged, nil
}

type uploadsRestoreSwap struct {
	uploadsDir string
	backupDir  string
	hadUploads bool
	replaced   bool
}

func (s *BackupService) replaceUploadsWithStaged(staged *stagedRestoreFiles) (*uploadsRestoreSwap, error) {
	uploadsDir := filepath.Join(s.dataDir, "uploads")
	swap := &uploadsRestoreSwap{
		uploadsDir: uploadsDir,
		backupDir:  filepath.Join(s.dataDir, ".restore-uploads-backup-"+randomID()),
	}

	if _, err := os.Stat(uploadsDir); err == nil {
		swap.hadUploads = true
		if err := os.Rename(uploadsDir, swap.backupDir); err != nil {
			return nil, fmt.Errorf("stage current uploads: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("inspect uploads directory: %w", err)
	}

	if err := os.Rename(staged.uploadsDir, uploadsDir); err != nil {
		if swap.hadUploads {
			_ = os.Rename(swap.backupDir, uploadsDir)
		}
		return nil, fmt.Errorf("activate restored uploads: %w", err)
	}
	swap.replaced = true
	return swap, nil
}

func (s *uploadsRestoreSwap) rollback() {
	if s == nil || !s.replaced {
		return
	}
	_ = os.RemoveAll(s.uploadsDir)
	if s.hadUploads {
		_ = os.Rename(s.backupDir, s.uploadsDir)
	}
}

func (s *uploadsRestoreSwap) cleanup() {
	if s == nil || !s.hadUploads {
		return
	}
	_ = os.RemoveAll(s.backupDir)
}

func confinedChildPath(baseDir, relativePath string) (string, error) {
	base, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}
	target := filepath.Join(baseDir, filepath.FromSlash(relativePath))
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if absTarget != base && !strings.HasPrefix(absTarget, base+string(os.PathSeparator)) {
		return "", ErrBackupPath
	}
	return absTarget, nil
}

func tableColumns(ctx context.Context, tx *sql.Tx, table string) (map[string]bool, error) {
	rows, err := tx.QueryContext(ctx, "PRAGMA table_info("+quoteIdentifier(table)+")")
	if err != nil {
		return nil, fmt.Errorf("read %s schema: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	columns := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return nil, fmt.Errorf("scan %s schema: %w", table, err)
		}
		columns[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate %s schema: %w", table, err)
	}
	return columns, nil
}

func insertBackupRow(ctx context.Context, tx *sql.Tx, table string, allowedColumns map[string]bool, row map[string]any) (bool, error) {
	columns := make([]string, 0, len(row))
	for column := range row {
		if allowedColumns[column] {
			columns = append(columns, column)
		}
	}
	sort.Strings(columns)
	if len(columns) == 0 {
		return false, nil
	}

	placeholders := make([]string, len(columns))
	values := make([]any, len(columns))
	for i, column := range columns {
		placeholders[i] = "?"
		value, err := normalizeImportValue(table, column, row[column])
		if err != nil {
			return false, err
		}
		values[i] = value
	}

	query := "INSERT INTO " + quoteIdentifier(table) +
		" (" + strings.Join(quoteIdentifiers(columns), ", ") + ")" +
		" VALUES (" + strings.Join(placeholders, ", ") + ")"
	if _, err := tx.ExecContext(ctx, query, values...); err != nil {
		return false, fmt.Errorf("insert %s: %w", table, err)
	}
	return true, nil
}

func normalizeBackupValue(table, column string, value any) any {
	switch typed := value.(type) {
	case []byte:
		if table == "file_blobs" && column == "data" {
			return base64.StdEncoding.EncodeToString(typed)
		}
		return string(typed)
	default:
		return typed
	}
}

func normalizeImportValue(table, column string, value any) (any, error) {
	if table == "file_blobs" && column == "data" {
		if encoded, ok := value.(string); ok {
			decoded, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				return nil, ErrBackupInvalid
			}
			return decoded, nil
		}
		return nil, ErrBackupInvalid
	}
	switch typed := value.(type) {
	case json.Number:
		if intValue, err := typed.Int64(); err == nil {
			return intValue, nil
		}
		if floatValue, err := typed.Float64(); err == nil {
			return floatValue, nil
		}
		return typed.String(), nil
	default:
		return typed, nil
	}
}

func finishBackupValidation(result *BackupValidationResult) *BackupValidationResult {
	result.Compatible = len(result.Errors) == 0
	return result
}

func sortedKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func quoteIdentifiers(values []string) []string {
	out := make([]string, len(values))
	for i, value := range values {
		out[i] = quoteIdentifier(value)
	}
	return out
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func validateBackupFilePath(value string) error {
	value = filepath.ToSlash(strings.TrimSpace(value))
	if value == "" || filepath.IsAbs(value) || strings.Contains(value, "\x00") {
		return ErrBackupPath
	}
	cleaned := pathClean(value)
	if cleaned != value || strings.HasPrefix(cleaned, "../") || cleaned == ".." {
		return ErrBackupPath
	}
	if !strings.HasPrefix(cleaned, "uploads/") {
		return ErrBackupPath
	}
	return nil
}

func validateBackupFiles(files []BackupFile) error {
	for _, file := range files {
		if err := validateBackupFilePath(file.Path); err != nil {
			return err
		}
		data, err := base64.StdEncoding.DecodeString(file.ContentBase64)
		if err != nil {
			return ErrBackupInvalid
		}
		if file.SizeBytes > 0 && int64(len(data)) != file.SizeBytes {
			return ErrBackupInvalid
		}
		sum := sha256.Sum256(data)
		if file.SHA256 != "" && !strings.EqualFold(file.SHA256, hex.EncodeToString(sum[:])) {
			return ErrBackupInvalid
		}
	}
	return nil
}

func pathClean(value string) string {
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(value)))
	if cleaned == "." {
		return ""
	}
	return cleaned
}

func DecodeBackup(data []byte) (*BackupDocument, error) {
	var doc BackupDocument
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&doc); err != nil {
		return nil, ErrBackupInvalid
	}
	return &doc, nil
}
