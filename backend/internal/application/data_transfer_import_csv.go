package application

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

var vehicleCSVFieldAliases = vehicleTransferCSVAliases()

var accessoryCSVFieldAliases = transferCSVFieldAliases(map[string][]string{
	"inventoryNumber":   {"Inventarnummer", "inventory number", "inventory_number", "inventoryNumber"},
	"manufacturer":      {"Hersteller", "manufacturer"},
	"articleNumber":     {"Artikelnummer", "article number", "article_number", "articleNumber"},
	"name":              {"Bezeichnung", "Name", "designation"},
	"category":          {"Kategorie", "category"},
	"trackingMode":      {"Erfassungsart", "tracking mode", "tracking_mode", "trackingMode"},
	"description":       {"Beschreibung", "description"},
	"ean":               {"EAN", "ean"},
	"articleType":       {"Artikelart", "article type", "article_type", "articleType"},
	"subtype":           {"Unterart", "subtype"},
	"gauges":            {"Spurweiten", "gauges"},
	"scale":             {"Maßstab", "Massstab", "scale"},
	"listPrice":         {"Listenpreis", "list price", "list_price", "listPrice"},
	"packageQuantity":   {"Packungsmenge", "package quantity", "package_quantity", "packageQuantity"},
	"stockUnit":         {"Bestandseinheit", "stock unit", "stock_unit", "stockUnit"},
	"minimumStock":      {"Mindestbestand", "minimum stock", "minimum_stock", "minimumStock"},
	"inventoryStrategy": {"Inventarstrategie", "inventory strategy", "inventory_strategy", "inventoryStrategy"},
	"stock":             {"Bestand", "stock"},
	"assets":            {"Einzelobjekte", "assets"},
})

func parseDataTransferCSV(
	area TransferArea,
	reader io.Reader,
) (DataTransferSnapshot, []int, error) {
	snapshot, rowNumbers, _, err := parseDataTransferCSVWithMapping(area, reader, nil, nil)
	return snapshot, rowNumbers, err
}

func parseDataTransferCSVWithMapping(
	area TransferArea,
	reader io.Reader,
	requested *DataTransferCSVMappingInput,
	profileDefaults map[string]string,
) (DataTransferSnapshot, []int, []DataTransferCSVColumnMapping, error) {
	payload, err := io.ReadAll(reader)
	if err != nil {
		return DataTransferSnapshot{}, nil, nil, fmt.Errorf("read transfer CSV: %w", err)
	}
	payload = bytes.TrimPrefix(payload, []byte{0xef, 0xbb, 0xbf})
	delimiter, err := detectDataTransferCSVDelimiter(payload)
	if err != nil {
		return DataTransferSnapshot{}, nil, nil, err
	}
	csvReader := csv.NewReader(bytes.NewReader(payload))
	csvReader.Comma = delimiter
	csvReader.FieldsPerRecord = -1
	csvReader.ReuseRecord = false
	rows, err := csvReader.ReadAll()
	if err != nil {
		return DataTransferSnapshot{}, nil, nil, fmt.Errorf("%w: malformed CSV: %v", ErrDataTransferValidation, err)
	}
	if len(rows) == 0 {
		return DataTransferSnapshot{}, nil, nil, fmt.Errorf("%w: CSV header is required", ErrDataTransferValidation)
	}
	var mapping []DataTransferCSVColumnMapping
	aliases := vehicleCSVFieldAliases
	if area == TransferAccessories {
		aliases = accessoryCSVFieldAliases
	} else if area != TransferVehicles {
		return DataTransferSnapshot{}, nil, nil, fmt.Errorf("%w: unsupported CSV area %q", ErrDataTransferValidation, area)
	}
	if area == TransferVehicles {
		mapping = defaultDataTransferCSVMapping(rows[0], profileDefaults)
		if requested != nil {
			mapping = append([]DataTransferCSVColumnMapping(nil), requested.Columns...)
		}
		if err := validateDataTransferCSVMapping(rows[0], mapping); err != nil {
			return DataTransferSnapshot{}, nil, nil, err
		}
		if requested != nil {
			if err := validateRequestedDataTransferCSVMapping(mapping); err != nil {
				return DataTransferSnapshot{}, nil, nil, err
			}
		}
	} else {
		fields, err := resolveDataTransferCSVHeader(rows[0], aliases)
		if err != nil {
			return DataTransferSnapshot{}, nil, nil, err
		}
		mapping = make([]DataTransferCSVColumnMapping, len(fields))
		for index, field := range fields {
			mapping[index] = DataTransferCSVColumnMapping{
				Index: index, SourceHeader: rows[0][index],
				NormalizedHeader: normalizeTransferCSVHeader(rows[0][index]), TargetField: field, Origin: CSVMappingAlias,
			}
		}
	}
	snapshot := DataTransferSnapshot{}
	rowNumbers := []int{}
	vehicleSetIndexes := map[string]int{}
	for index, values := range rows[1:] {
		rowNumber := index + 2
		if transferCSVRowEmpty(values) {
			continue
		}
		if len(values) != len(mapping) {
			return DataTransferSnapshot{}, nil, nil, fmt.Errorf(
				"%w: CSV row %d has %d fields, expected %d", ErrDataTransferValidation,
				rowNumber, len(values), len(mapping),
			)
		}
		row := make(map[string]string, len(mapping))
		for _, column := range mapping {
			if column.TargetField != "" {
				row[column.TargetField] = decodeSafeDataTransferCSVValue(strings.TrimSpace(values[column.Index]))
			}
		}
		switch area {
		case TransferVehicles:
			vehicle, err := transferVehicleFromCSV(row, rowNumber)
			if err != nil {
				return DataTransferSnapshot{}, nil, nil, err
			}
			snapshot.Vehicles = append(snapshot.Vehicles, vehicle)
			set, found, err := transferVehicleSetFromCSV(row, rowNumber, vehicle)
			if err != nil {
				return DataTransferSnapshot{}, nil, nil, err
			}
			if found {
				setKey := transferIdentity(set.InventoryNumber)
				if setKey == "" {
					setKey = fmt.Sprintf("row-%d", rowNumber)
				}
				if setIndex, exists := vehicleSetIndexes[setKey]; exists {
					snapshot.VehicleSets[setIndex].Members = append(
						snapshot.VehicleSets[setIndex].Members,
						set.Members...,
					)
					snapshot.VehicleSets[setIndex].Diagnostics = append(
						snapshot.VehicleSets[setIndex].Diagnostics,
						set.Diagnostics...,
					)
				} else {
					vehicleSetIndexes[setKey] = len(snapshot.VehicleSets)
					snapshot.VehicleSets = append(snapshot.VehicleSets, set)
				}
			}
		case TransferAccessories:
			accessory, err := transferAccessoryFromCSV(row, rowNumber)
			if err != nil {
				return DataTransferSnapshot{}, nil, nil, err
			}
			snapshot.Accessories = append(snapshot.Accessories, accessory)
		}
		rowNumbers = append(rowNumbers, rowNumber)
	}
	for index := range snapshot.VehicleSets {
		slices.SortStableFunc(snapshot.VehicleSets[index].Members, func(left, right TransferVehicleSetMember) int {
			if left.Position != right.Position {
				return left.Position - right.Position
			}
			return strings.Compare(left.VehicleInventoryNumber, right.VehicleInventoryNumber)
		})
	}
	return snapshot, rowNumbers, mapping, nil
}

func transferVehicleSetFromCSV(
	row map[string]string,
	rowNumber int,
	vehicle TransferVehicle,
) (TransferVehicleSet, bool, error) {
	hasSetValue := false
	for _, field := range vehicleSetTransferFields {
		if strings.TrimSpace(row[field.Key]) != "" {
			hasSetValue = true
			break
		}
	}
	if !hasSetValue {
		return TransferVehicleSet{}, false, nil
	}
	position, positionDiagnostic := parseTransferCSVSetInteger(
		row["vehicleSetPosition"], rowNumber, "vehicleSetPosition", "invalid_vehicle_set_position",
	)
	memberCount, memberCountDiagnostic := parseTransferCSVSetInteger(
		row["vehicleSetMemberCount"], rowNumber, "vehicleSetMemberCount", "invalid_vehicle_set_member_count",
	)
	diagnostics := []TransferVehicleSetDiagnostic{}
	if positionDiagnostic != nil {
		diagnostics = append(diagnostics, *positionDiagnostic)
	}
	if memberCountDiagnostic != nil {
		diagnostics = append(diagnostics, *memberCountDiagnostic)
	}
	return TransferVehicleSet{
		InventoryNumber: row["vehicleSetInventoryNumber"],
		VehicleSetInput: VehicleSetInput{
			Name: row["vehicleSetName"], Manufacturer: row["vehicleSetManufacturer"],
			ArticleNumber: row["vehicleSetArticleNumber"], ArticleSourceURL: row["vehicleSetArticleSourceUrl"],
			Gauge: row["vehicleSetGauge"], Epoch: row["vehicleSetEpoch"],
			RailwayCompany: row["vehicleSetRailwayCompany"], Category: row["vehicleSetCategory"],
			Gattung: row["vehicleSetGattung"], Description: row["vehicleSetDescription"],
			EAN: row["vehicleSetEAN"], ProductionPeriod: row["vehicleSetProductionPeriod"],
			ListPrice: row["vehicleSetListPrice"], AcquisitionType: row["vehicleSetAcquisitionType"],
			AcquiredFrom: row["vehicleSetAcquiredFrom"], PurchasePrice: row["vehicleSetPurchasePrice"],
			PurchaseDate: row["vehicleSetPurchaseDate"], StorageLocation: row["vehicleSetStorageLocation"],
			StorageDetails: row["vehicleSetStorageDetails"], Condition: row["vehicleSetCondition"],
			ConditionDetails: row["vehicleSetConditionDetails"], Packaging: row["vehicleSetPackaging"],
		},
		Members: []TransferVehicleSetMember{{
			VehicleInventoryNumber: vehicle.InventoryNumber,
			Position:               position,
			Label:                  row["vehicleSetMemberLabel"],
			SourceRowNumber:        rowNumber,
			DeclaredMemberCount:    memberCount,
		}},
		Diagnostics: diagnostics,
	}, true, nil
}

func parseTransferCSVSetInteger(
	value string,
	rowNumber int,
	field string,
	code string,
) (int, *TransferVehicleSetDiagnostic) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err == nil {
		return parsed, nil
	}
	return 0, &TransferVehicleSetDiagnostic{RowNumber: rowNumber, Field: field, Code: code}
}

func detectDataTransferCSVDelimiter(payload []byte) (rune, error) {
	firstLine := string(payload)
	if newline := strings.IndexByte(firstLine, '\n'); newline >= 0 {
		firstLine = firstLine[:newline]
	}
	firstLine = strings.TrimSuffix(firstLine, "\r")
	best := rune(0)
	bestCount := 0
	for _, candidate := range []rune{';', ',', '\t'} {
		count := strings.Count(firstLine, string(candidate))
		if count > bestCount {
			best = candidate
			bestCount = count
		}
	}
	if bestCount == 0 {
		return 0, fmt.Errorf("%w: CSV delimiter could not be detected", ErrDataTransferValidation)
	}
	return best, nil
}

func transferCSVFieldAliases(fields map[string][]string) map[string]string {
	aliases := map[string]string{}
	for field, names := range fields {
		for _, name := range names {
			aliases[normalizeTransferCSVHeader(name)] = field
		}
	}
	return aliases
}

func resolveDataTransferCSVHeader(header []string, aliases map[string]string) ([]string, error) {
	fields := make([]string, len(header))
	seen := map[string]bool{}
	for index, value := range header {
		field, found := aliases[normalizeTransferCSVHeader(value)]
		if !found {
			return nil, fmt.Errorf("%w: unsupported CSV column %q", ErrDataTransferValidation, value)
		}
		if seen[field] {
			return nil, fmt.Errorf("%w: repeated CSV column %q", ErrDataTransferValidation, value)
		}
		seen[field] = true
		fields[index] = field
	}
	return fields, nil
}

func normalizeTransferCSVHeader(value string) string {
	return strings.Map(func(character rune) rune {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			return unicode.ToLower(character)
		}
		return -1
	}, strings.TrimSpace(strings.TrimPrefix(value, "\ufeff")))
}

func transferCSVRowEmpty(values []string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func decodeSafeDataTransferCSVValue(value string) string {
	if len(value) < 2 || value[0] != '\'' {
		return value
	}
	decoded := value[1:]
	candidate := strings.TrimLeftFunc(decoded, unicode.IsSpace)
	if candidate != "" && (candidate[0] == '\'' || strings.ContainsRune("=+-@", rune(candidate[0]))) {
		return decoded
	}
	return value
}

func transferVehicleFromCSV(row map[string]string, rowNumber int) (TransferVehicle, error) {
	maximumSpeed, err := parseTransferCSVOptionalInteger(row["maximumSpeedKmh"], rowNumber, "Höchstgeschwindigkeit")
	if err != nil {
		return TransferVehicle{}, err
	}
	booleanValues := make(map[string]bool, 13)
	for _, field := range []struct {
		key   string
		label string
	}{
		{key: "digital", label: "Digital"},
		{key: "dtDecoder", label: "D&T-Decoder"},
		{key: "exhibitionReady", label: "Ausstellungsbereit"},
		{key: "exhibition", label: "Ausstellung"},
		{key: "abcBrakes", label: "ABC-Bremsen"},
		{key: "couplingSame", label: "Kupplungen identisch"},
		{key: "driveEnabled", label: "Antrieb vorhanden"},
		{key: "headlightsEnabled", label: "Spitzenlicht vorhanden"},
		{key: "lightingEnabled", label: "Beleuchtung vorhanden"},
		{key: "soundGeneratorEnabled", label: "Soundgenerator vorhanden"},
		{key: "smokeGeneratorEnabled", label: "Rauchgenerator vorhanden"},
		{key: "qrCodeEnabled", label: "QR-Code aktiviert"},
	} {
		parsed, parseErr := parseTransferCSVBoolean(row[field.key], rowNumber, field.label)
		if parseErr != nil {
			return TransferVehicle{}, parseErr
		}
		booleanValues[field.key] = parsed
	}
	return TransferVehicle{
		InventoryNumber: row["inventoryNumber"], Manufacturer: row["manufacturer"],
		ArticleNumber: row["articleNumber"], ArticleSourceURL: row["articleSourceUrl"], Name: row["name"],
		Gauge: row["gauge"], Epoch: row["epoch"], RailwayCompany: row["railwayCompany"],
		Category: row["category"], Gattung: row["gattung"], Description: row["description"],
		Series: row["series"], VehicleNumber: row["vehicleNumber"], MaximumSpeedKmh: maximumSpeed,
		HomeBase: row["homeBase"], Digital: booleanValues["digital"],
		DigitalDecoderNumber: row["digitalDecoderNumber"], DecoderType: row["decoderType"],
		DTDecoder: booleanValues["dtDecoder"], DTDecoderNumber: row["dtDecoderNumber"],
		ExhibitionReady: booleanValues["exhibitionReady"],
		Exhibition:      booleanValues["exhibition"],
		ABCBrakes:       booleanValues["abcBrakes"], EAN: row["ean"],
		ProductionPeriod: row["productionPeriod"], ListPrice: row["listPrice"],
		AcquisitionType: row["acquisitionType"], AcquiredFrom: row["acquiredFrom"],
		PurchasePrice: row["purchasePrice"], PurchaseDate: row["purchaseDate"],
		StorageLocation: row["storageLocation"], StorageDetails: row["storageDetails"],
		Condition: row["condition"], ConditionDetails: row["conditionDetails"], Packaging: row["packaging"],
		LengthMM: row["lengthMm"], WeightG: row["weightG"], Color: row["color"], Lettering: row["lettering"],
		Load: row["load"], Interior: row["interior"], Axles: row["axles"], AxleCount: row["axleCount"],
		TractionTireCount: row["tractionTireCount"], Wheelset: row["wheelset"],
		CouplingSame: booleanValues["couplingSame"], CouplingFront: row["couplingFront"],
		CouplingRear: row["couplingRear"], PowerPickup: row["powerPickup"], Adapter: row["adapter"],
		DriveEnabled: booleanValues["driveEnabled"], DriveDescription: row["driveDescription"],
		HeadlightsEnabled:         booleanValues["headlightsEnabled"],
		HeadlightsDescription:     row["headlightsDescription"],
		LightingEnabled:           booleanValues["lightingEnabled"],
		LightingDescription:       row["lightingDescription"],
		SoundGeneratorEnabled:     booleanValues["soundGeneratorEnabled"],
		SoundGeneratorDescription: row["soundGeneratorDescription"],
		SmokeGeneratorEnabled:     booleanValues["smokeGeneratorEnabled"],
		SmokeGeneratorDescription: row["smokeGeneratorDescription"], AdditionalInfo: row["additionalInfo"],
		QRCodeEnabled: booleanValues["qrCodeEnabled"],
	}, nil
}

func parseTransferCSVOptionalInteger(value string, rowNumber int, field string) (*int, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	parsed, err := parseTransferCSVInteger(value, rowNumber, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseTransferCSVBoolean(value string, rowNumber int, field string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "ja", "yes", "true", "wahr", "digital", "d", "x", "vorhanden":
		return true, nil
	case "", "0", "nein", "no", "false", "falsch", "analog", "a", "nicht vorhanden":
		return false, nil
	}
	return false, fmt.Errorf("%w: invalid %s in CSV row %d", ErrDataTransferValidation, field, rowNumber)
}

func transferAccessoryFromCSV(row map[string]string, rowNumber int) (TransferAccessory, error) {
	packageQuantity, err := parseTransferCSVInteger(row["packageQuantity"], rowNumber, "Packungsmenge")
	if err != nil {
		return TransferAccessory{}, err
	}
	minimumStock, err := parseTransferCSVInteger(row["minimumStock"], rowNumber, "Mindestbestand")
	if err != nil {
		return TransferAccessory{}, err
	}
	stock := []TransferAccessoryStock{}
	if err := decodeTransferCSVJSON(row["stock"], &stock); err != nil {
		return TransferAccessory{}, fmt.Errorf(
			"%w: invalid Bestand JSON in CSV row %d: %v", ErrDataTransferValidation, rowNumber, err,
		)
	}
	assets := []TransferAccessoryAsset{}
	if err := decodeTransferCSVJSON(row["assets"], &assets); err != nil {
		return TransferAccessory{}, fmt.Errorf(
			"%w: invalid Einzelobjekte JSON in CSV row %d: %v", ErrDataTransferValidation, rowNumber, err,
		)
	}
	gauges := []string{}
	for _, gauge := range strings.Split(row["gauges"], ",") {
		if gauge = strings.TrimSpace(gauge); gauge != "" {
			gauges = append(gauges, gauge)
		}
	}
	return TransferAccessory{
		InventoryNumber: row["inventoryNumber"], Manufacturer: row["manufacturer"],
		ArticleNumber: row["articleNumber"], Name: row["name"], Category: row["category"],
		TrackingMode: row["trackingMode"], Description: row["description"], EAN: row["ean"],
		ArticleType: row["articleType"], Subtype: row["subtype"], Gauges: gauges, Scale: row["scale"],
		ListPrice: row["listPrice"], PackageQuantity: packageQuantity, StockUnit: row["stockUnit"],
		MinimumStock: minimumStock, InventoryStrategy: row["inventoryStrategy"], Stock: stock, Assets: assets,
	}, nil
}

func parseTransferCSVInteger(value string, rowNumber int, field string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s in CSV row %d", ErrDataTransferValidation, field, rowNumber)
	}
	return parsed, nil
}

func decodeTransferCSVJSON(value string, target any) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("trailing JSON data")
	}
	return nil
}

func classifyDataTransferImport(
	jobID string,
	incoming DataTransferSnapshot,
	current DataTransferSnapshot,
	rowNumbers map[TransferArea][]int,
) ([]DataTransferPreviewRecord, []DataTransferIssue) {
	records := make([]DataTransferPreviewRecord, 0,
		len(incoming.Vehicles)+len(incoming.Accessories)+len(incoming.ExhibitionLists))
	issues := []DataTransferIssue{}
	currentVehiclesByInventory := map[string]TransferVehicle{}
	currentVehiclesByID := map[string]TransferVehicle{}
	for _, vehicle := range current.Vehicles {
		currentVehiclesByInventory[transferIdentity(vehicle.InventoryNumber)] = vehicle
		if vehicle.ID != "" {
			currentVehiclesByID[vehicle.ID] = vehicle
		}
	}
	seenVehicleInventory := map[string]bool{}
	for index, vehicle := range incoming.Vehicles {
		rowNumber := transferRowNumber(rowNumbers[TransferVehicles], index)
		recordKey := strings.TrimSpace(vehicle.InventoryNumber)
		if recordKey == "" {
			recordKey = fmt.Sprintf("row-%d", transferRowValue(rowNumber, index+1))
		}
		record := newDataTransferPreviewRecord(TransferVehicles, recordKey, rowNumber, vehicle)
		start := len(issues)
		issues = appendRequiredTransferIssues(issues, jobID, TransferVehicles, recordKey, rowNumber,
			vehicle.InventoryNumber, vehicle.Manufacturer, vehicle.Name)
		for _, required := range []struct {
			field, value, code, message string
		}{
			{"gauge", vehicle.Gauge, "missing_gauge", "Gauge is required."},
			{"category", vehicle.Category, "missing_category", "Category is required."},
			{"gattung", vehicle.Gattung, "missing_gattung", "Gattung is required."},
		} {
			if strings.TrimSpace(required.value) == "" {
				issues = append(issues, newTransferIssue(jobID, TransferVehicles, recordKey, rowNumber,
					required.field, TransferIssueError, required.code, required.message, "skip"))
			}
		}
		if err := ValidateTransferVehicle(vehicle); err != nil &&
			strings.TrimSpace(vehicle.Manufacturer) != "" && strings.TrimSpace(vehicle.Name) != "" &&
			strings.TrimSpace(vehicle.Gauge) != "" && strings.TrimSpace(vehicle.Category) != "" &&
			strings.TrimSpace(vehicle.Gattung) != "" {
			issues = append(issues, newTransferIssue(jobID, TransferVehicles, recordKey, rowNumber, "record",
				TransferIssueError, "invalid_vehicle", "Vehicle data violates aggregate validation.", "skip"))
		}
		identity := transferIdentity(vehicle.InventoryNumber)
		if identity != "" && seenVehicleInventory[identity] {
			issues = append(issues, newTransferIssue(jobID, TransferVehicles, recordKey, rowNumber, "inventoryNumber",
				TransferIssueError, "duplicate_input_inventory_number", "Inventory number occurs more than once in the upload.", "skip"))
		} else if target, found := currentVehiclesByInventory[identity]; identity != "" && found {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			record.ProposedAction = "replace"
			issues = append(issues, newTransferIssue(jobID, TransferVehicles, recordKey, rowNumber, "inventoryNumber",
				TransferIssueWarning, "duplicate_inventory_number", "Inventory number already exists.", "replace_or_copy"))
		}
		if identity != "" {
			seenVehicleInventory[identity] = true
		}
		finalizeDataTransferRecord(&record, issues[start:])
		records = append(records, record)
	}

	currentAccessoriesByInventory := map[string]TransferAccessory{}
	currentAccessoriesByArticle := map[string]TransferAccessory{}
	for _, accessory := range current.Accessories {
		currentAccessoriesByInventory[transferIdentity(accessory.InventoryNumber)] = accessory
		if key := transferArticleIdentity(accessory.Manufacturer, accessory.ArticleNumber); key != "" {
			currentAccessoriesByArticle[key] = accessory
		}
	}
	seenAccessoryInventory := map[string]bool{}
	for index, accessory := range incoming.Accessories {
		rowNumber := transferRowNumber(rowNumbers[TransferAccessories], index)
		recordKey := strings.TrimSpace(accessory.InventoryNumber)
		if recordKey == "" {
			recordKey = fmt.Sprintf("row-%d", transferRowValue(rowNumber, index+1))
		}
		record := newDataTransferPreviewRecord(TransferAccessories, recordKey, rowNumber, accessory)
		start := len(issues)
		issues = appendRequiredTransferIssues(issues, jobID, TransferAccessories, recordKey, rowNumber,
			accessory.InventoryNumber, accessory.Manufacturer, accessory.Name)
		if err := ValidateTransferAccessory(accessory); err != nil {
			issues = append(issues, newTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "record",
				TransferIssueError, "invalid_accessory", "Accessory data violates aggregate validation.", "skip"))
		}
		identity := transferIdentity(accessory.InventoryNumber)
		if identity != "" && seenAccessoryInventory[identity] {
			issues = append(issues, newTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "inventoryNumber",
				TransferIssueError, "duplicate_input_inventory_number", "Inventory number occurs more than once in the upload.", "skip"))
		} else if target, found := currentAccessoriesByInventory[identity]; identity != "" && found {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			record.ProposedAction = "replace"
			issues = append(issues, newTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "inventoryNumber",
				TransferIssueWarning, "duplicate_inventory_number", "Inventory number already exists.", "replace_or_copy"))
		} else if target, found := currentAccessoriesByArticle[transferArticleIdentity(accessory.Manufacturer, accessory.ArticleNumber)]; found {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			record.ProposedAction = "use_existing"
			issues = append(issues, newTransferIssue(jobID, TransferAccessories, recordKey, rowNumber, "articleNumber",
				TransferIssueWarning, "matching_manufacturer_article_number",
				"Manufacturer and article number match an existing accessory.", "use_existing"))
		}
		if identity != "" {
			seenAccessoryInventory[identity] = true
		}
		finalizeDataTransferRecord(&record, issues[start:])
		records = append(records, record)
	}

	currentListsByID := map[string]TransferExhibitionList{}
	currentListsByIdentity := map[string]TransferExhibitionList{}
	for _, list := range current.ExhibitionLists {
		if list.ID != "" {
			currentListsByID[list.ID] = list
		}
		currentListsByIdentity[transferExhibitionIdentity(list.Designation, list.Date)] = list
	}
	for index, list := range incoming.ExhibitionLists {
		rowNumber := transferRowNumber(rowNumbers[TransferExhibitionLists], index)
		recordKey := strings.TrimSpace(list.ID)
		if recordKey == "" {
			recordKey = strings.TrimSpace(list.Designation) + "|" + strings.TrimSpace(list.Date)
		}
		record := newDataTransferPreviewRecord(TransferExhibitionLists, recordKey, rowNumber, list)
		start := len(issues)
		if strings.TrimSpace(list.Designation) == "" {
			issues = append(issues, newTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber,
				"designation", TransferIssueError, "missing_name", "Designation is required.", "skip"))
		}
		target, matched := currentListsByID[list.ID]
		if matched && transferExhibitionIdentity(target.Designation, target.Date) !=
			transferExhibitionIdentity(list.Designation, list.Date) {
			matched = false
		}
		if !matched {
			target, matched = currentListsByIdentity[transferExhibitionIdentity(list.Designation, list.Date)]
		}
		if matched {
			record.TargetID = target.ID
			record.TargetUpdatedAt = target.UpdatedAt
			record.TargetFingerprint = DataTransferTargetFingerprint(target)
			if target.Locked {
				record.ProposedAction = "copy"
				issues = append(issues, newTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber, "id",
					TransferIssueError, "locked_exhibition_list", "Locked exhibition lists cannot be replaced.", "copy"))
			} else {
				record.ProposedAction = "replace"
				issues = append(issues, newTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber, "id",
					TransferIssueWarning, "duplicate_exhibition_list",
					"Exhibition list already exists; choose replace or copy.", "replace_or_copy"))
			}
		}
		for entryIndex, entry := range list.Entries {
			if entry.VehicleID == "" && entry.VehicleInventoryNumber == "" {
				continue
			}
			if local, found := currentVehiclesByID[entry.VehicleID]; found &&
				strings.TrimSpace(entry.VehicleInventoryNumber) != "" && strings.EqualFold(
				strings.TrimSpace(entry.VehicleInventoryNumber), strings.TrimSpace(local.InventoryNumber)) {
				continue
			}
			if _, found := currentVehiclesByInventory[transferIdentity(entry.VehicleInventoryNumber)]; found {
				issues = append(issues, newTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber,
					transferExhibitionEntryIssueField(entryIndex), TransferIssueWarning, "exhibition_vehicle_reference",
					"Vehicle reference can be linked by inventory number after confirmation.", "link"))
			} else {
				issues = append(issues, newTransferIssue(jobID, TransferExhibitionLists, recordKey, rowNumber,
					transferExhibitionEntryIssueField(entryIndex), TransferIssueError, "missing_vehicle_reference",
					"Referenced vehicle does not exist in this installation.", "skip"))
			}
		}
		finalizeDataTransferRecord(&record, issues[start:])
		records = append(records, record)
	}
	return records, issues
}

func appendRequiredTransferIssues(
	issues []DataTransferIssue,
	jobID string,
	area TransferArea,
	recordKey string,
	rowNumber *int,
	inventoryNumber string,
	manufacturer string,
	name string,
) []DataTransferIssue {
	if strings.TrimSpace(inventoryNumber) == "" {
		issues = append(issues, newTransferIssue(jobID, area, recordKey, rowNumber, "inventoryNumber",
			TransferIssueError, "missing_inventory_number", "Inventory number is required.", "skip"))
	}
	if strings.TrimSpace(manufacturer) == "" {
		issues = append(issues, newTransferIssue(jobID, area, recordKey, rowNumber, "manufacturer",
			TransferIssueError, "missing_manufacturer", "Manufacturer is required.", "skip"))
	}
	if strings.TrimSpace(name) == "" {
		issues = append(issues, newTransferIssue(jobID, area, recordKey, rowNumber, "name",
			TransferIssueError, "missing_name", "Name is required.", "skip"))
	}
	return issues
}

func newTransferIssue(
	jobID string,
	area TransferArea,
	recordKey string,
	rowNumber *int,
	field string,
	severity TransferIssueSeverity,
	code string,
	message string,
	proposedResolution string,
) DataTransferIssue {
	row := 0
	if rowNumber != nil {
		row = *rowNumber
	}
	identity := fmt.Sprintf("%s\x00%s\x00%s\x00%d\x00%s\x00%s", jobID, area, recordKey, row, field, code)
	digest := sha256.Sum256([]byte(identity))
	return DataTransferIssue{
		ID: hex.EncodeToString(digest[:]), JobID: jobID, Area: area, RecordKey: recordKey,
		RowNumber: rowNumber, Field: field,
		Severity: severity, Code: code, Message: message, ProposedResolution: proposedResolution,
	}
}

func newDataTransferPreviewRecord(
	area TransferArea,
	recordKey string,
	rowNumber *int,
	data any,
) DataTransferPreviewRecord {
	payload, _ := json.Marshal(data)
	return DataTransferPreviewRecord{
		Area: area, RecordKey: recordKey, RowNumber: rowNumber, Classification: "ready",
		ProposedAction: "create", Data: payload,
	}
}

func finalizeDataTransferRecord(record *DataTransferPreviewRecord, issues []DataTransferIssue) {
	record.Classification = "ready"
	for _, issue := range issues {
		if issue.Severity == TransferIssueError {
			record.Classification = "error"
			return
		}
		if issue.Severity == TransferIssueWarning {
			record.Classification = "warning"
		}
	}
}

func countDataTransferPreviewRecords(records []DataTransferPreviewRecord) (ready, warnings, failures int) {
	for _, record := range records {
		switch record.Classification {
		case "ready":
			ready++
		case "warning":
			warnings++
		case "error":
			failures++
		}
	}
	return ready, warnings, failures
}

func transferIdentity(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func transferArticleIdentity(manufacturer, articleNumber string) string {
	manufacturer = transferIdentity(manufacturer)
	articleNumber = transferIdentity(articleNumber)
	if manufacturer == "" || articleNumber == "" {
		return ""
	}
	return manufacturer + "\x00" + articleNumber
}

func transferExhibitionIdentity(designation, date string) string {
	return transferIdentity(designation) + "\x00" + strings.TrimSpace(date)
}

func transferRowNumber(rows []int, index int) *int {
	if index < 0 || index >= len(rows) {
		value := index + 1
		return &value
	}
	value := rows[index]
	return &value
}

func transferExhibitionEntryIssueField(index int) string {
	return fmt.Sprintf("entries[%d].vehicleReference", index)
}

func transferRowValue(rowNumber *int, fallback int) int {
	if rowNumber == nil {
		return fallback
	}
	return *rowNumber
}
