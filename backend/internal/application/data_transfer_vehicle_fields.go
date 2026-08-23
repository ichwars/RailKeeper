package application

import (
	"fmt"
	"strings"
)

type VehicleTransferValueKind string

const (
	VehicleTransferString  VehicleTransferValueKind = "string"
	VehicleTransferInteger VehicleTransferValueKind = "integer"
	VehicleTransferBoolean VehicleTransferValueKind = "boolean"
)

type VehicleTransferField struct {
	Key     string                   `json:"key"`
	LabelDE string                   `json:"labelDE"`
	LabelEN string                   `json:"labelEN"`
	Kind    VehicleTransferValueKind `json:"kind"`
	Aliases []string                 `json:"aliases"`
}

type DataTransferCSVMappingOrigin string

const (
	CSVMappingAlias    DataTransferCSVMappingOrigin = "alias"
	CSVMappingProfile  DataTransferCSVMappingOrigin = "profile"
	CSVMappingManual   DataTransferCSVMappingOrigin = "manual"
	CSVMappingIgnored  DataTransferCSVMappingOrigin = "ignored"
	CSVMappingUnmapped DataTransferCSVMappingOrigin = "unmapped"
)

type DataTransferCSVColumnMapping struct {
	Index            int                          `json:"index"`
	SourceHeader     string                       `json:"sourceHeader"`
	NormalizedHeader string                       `json:"normalizedHeader"`
	TargetField      string                       `json:"targetField"`
	Origin           DataTransferCSVMappingOrigin `json:"origin"`
}

type DataTransferCSVMappingInput struct {
	Columns       []DataTransferCSVColumnMapping `json:"columns"`
	SaveToProfile bool                           `json:"saveToProfile"`
}

const dataTransferCSVMappingOption = "csvMapping"
const dataTransferCSVIgnoreTarget = "__ignore__"

var vehicleTransferFields = []VehicleTransferField{
	vehicleTransferField("inventoryNumber", "Inventarnummer", "Inventory number", VehicleTransferString),
	vehicleTransferField("manufacturer", "Hersteller", "Manufacturer", VehicleTransferString),
	vehicleTransferField("articleNumber", "Artikel-Nr.", "Article number", VehicleTransferString, "Artikelnummer"),
	vehicleTransferField("articleSourceUrl", "Quelle / URL", "Source / URL", VehicleTransferString),
	vehicleTransferField("name", "Bezeichnung", "Name", VehicleTransferString, "designation"),
	vehicleTransferField("gauge", "Spurweite", "Gauge", VehicleTransferString),
	vehicleTransferField("epoch", "Epoche", "Epoch", VehicleTransferString),
	vehicleTransferField("railwayCompany", "Bahngesellschaft", "Railway company", VehicleTransferString),
	vehicleTransferField("category", "Kategorie", "Category", VehicleTransferString),
	vehicleTransferField("gattung", "Gattung", "Class", VehicleTransferString),
	vehicleTransferField("description", "Beschreibung", "Description", VehicleTransferString),
	vehicleTransferField("series", "Baureihe", "Series", VehicleTransferString),
	vehicleTransferField("vehicleNumber", "Fahrzeug-Nr.", "Vehicle number", VehicleTransferString),
	vehicleTransferField("maximumSpeedKmh", "Höchstgeschwindigkeit", "Maximum speed", VehicleTransferInteger),
	vehicleTransferField("homeBase", "Heimat-Bw / Einsatzstelle", "Home base", VehicleTransferString),
	vehicleTransferField("digital", "Digital", "Digital", VehicleTransferBoolean),
	vehicleTransferField("digitalDecoderNumber", "Digital / Decoder-Nr.", "Digital / decoder number", VehicleTransferString),
	vehicleTransferField("decoderType", "Decoder-Typ", "Decoder type", VehicleTransferString),
	vehicleTransferField("dtDecoder", "DT / Decoder", "DT / decoder", VehicleTransferBoolean),
	vehicleTransferField("dtDecoderNumber", "DT / Decoder-Nr.", "DT / decoder number", VehicleTransferString),
	vehicleTransferField("exhibitionReady", "Messe tauglich", "Exhibition ready", VehicleTransferBoolean),
	vehicleTransferField("exhibition", "Ausstellung", "Exhibition", VehicleTransferBoolean),
	vehicleTransferField("abcBrakes", "ABC-Bremsen", "ABC brakes", VehicleTransferBoolean),
	vehicleTransferField("ean", "EAN", "EAN", VehicleTransferString),
	vehicleTransferField("productionPeriod", "Produktionszeit", "Production period", VehicleTransferString),
	vehicleTransferField("listPrice", "Listenpreis", "List price", VehicleTransferString),
	vehicleTransferField("acquisitionType", "Erwerbsart", "Acquisition type", VehicleTransferString),
	vehicleTransferField("acquiredFrom", "Erworben von/bei", "Acquired from", VehicleTransferString),
	vehicleTransferField("purchasePrice", "Kaufpreis", "Purchase price", VehicleTransferString),
	vehicleTransferField("purchaseDate", "Kaufdatum", "Purchase date", VehicleTransferString),
	vehicleTransferField("storageLocation", "Lagerort", "Storage location", VehicleTransferString),
	vehicleTransferField("storageDetails", "Lagerdetails", "Storage details", VehicleTransferString),
	vehicleTransferField("condition", "Zustand", "Condition", VehicleTransferString),
	vehicleTransferField("conditionDetails", "Zustandsdetails", "Condition details", VehicleTransferString),
	vehicleTransferField("packaging", "Verpackung", "Packaging", VehicleTransferString),
	vehicleTransferField("lengthMm", "Länge (mm)", "Length (mm)", VehicleTransferString),
	vehicleTransferField("weightG", "Gewicht (g)", "Weight (g)", VehicleTransferString),
	vehicleTransferField("color", "Farbe", "Color", VehicleTransferString),
	vehicleTransferField("lettering", "Beschriftung", "Lettering", VehicleTransferString),
	vehicleTransferField("load", "Beladung", "Load", VehicleTransferString),
	vehicleTransferField("interior", "Inneneinrichtung", "Interior", VehicleTransferString),
	vehicleTransferField("axles", "Achsen", "Axles", VehicleTransferString),
	vehicleTransferField("axleCount", "Anzahl Achsen", "Axle count", VehicleTransferString),
	vehicleTransferField("tractionTireCount", "Anzahl Haftreifen", "Traction tire count", VehicleTransferString),
	vehicleTransferField("wheelset", "Radsatz", "Wheelset", VehicleTransferString),
	vehicleTransferField("couplingSame", "Kupplung (V=H)", "Same coupling front/rear", VehicleTransferBoolean),
	vehicleTransferField("couplingFront", "Kupplung vorne", "Front coupling", VehicleTransferString),
	vehicleTransferField("couplingRear", "Kupplung hinten", "Rear coupling", VehicleTransferString),
	vehicleTransferField("powerPickup", "Stromabnahme", "Power pickup", VehicleTransferString),
	vehicleTransferField("adapter", "Adapter / Schnittstelle", "Adapter / interface", VehicleTransferString),
	vehicleTransferField("driveEnabled", "Antrieb", "Drive", VehicleTransferBoolean),
	vehicleTransferField("driveDescription", "Antrieb Beschreibung", "Drive description", VehicleTransferString),
	vehicleTransferField("headlightsEnabled", "Fahrlicht", "Headlights", VehicleTransferBoolean),
	vehicleTransferField("headlightsDescription", "Fahrlicht Beschreibung", "Headlights description", VehicleTransferString),
	vehicleTransferField("lightingEnabled", "Beleuchtung", "Lighting", VehicleTransferBoolean),
	vehicleTransferField("lightingDescription", "Beleuchtung Beschreibung", "Lighting description", VehicleTransferString),
	vehicleTransferField("soundGeneratorEnabled", "Soundgenerator", "Sound generator", VehicleTransferBoolean),
	vehicleTransferField("soundGeneratorDescription", "Soundgenerator Beschreibung", "Sound generator description", VehicleTransferString),
	vehicleTransferField("smokeGeneratorEnabled", "Rauchgenerator", "Smoke generator", VehicleTransferBoolean),
	vehicleTransferField("smokeGeneratorDescription", "Rauchgenerator Beschreibung", "Smoke generator description", VehicleTransferString),
	vehicleTransferField("additionalInfo", "Zusatzinformationen", "Additional information", VehicleTransferString),
	vehicleTransferField("qrCodeEnabled", "QR-Code erstellen", "Create QR code", VehicleTransferBoolean),
}

func vehicleTransferField(
	key string,
	labelDE string,
	labelEN string,
	kind VehicleTransferValueKind,
	extraAliases ...string,
) VehicleTransferField {
	aliases := []string{key, labelDE, labelEN}
	aliases = append(aliases, extraAliases...)
	return VehicleTransferField{Key: key, LabelDE: labelDE, LabelEN: labelEN, Kind: kind, Aliases: aliases}
}

func VehicleTransferFields() []VehicleTransferField {
	fields := make([]VehicleTransferField, len(vehicleTransferFields))
	copy(fields, vehicleTransferFields)
	for index := range fields {
		fields[index].Aliases = append([]string(nil), fields[index].Aliases...)
	}
	return fields
}

func VehicleTransferFieldByKey(key string) (VehicleTransferField, bool) {
	for _, field := range vehicleTransferFields {
		if field.Key == key {
			field.Aliases = append([]string(nil), field.Aliases...)
			return field, true
		}
	}
	return VehicleTransferField{}, false
}

func defaultDataTransferCSVMapping(
	header []string,
	profileDefaults map[string]string,
) []DataTransferCSVColumnMapping {
	aliases := vehicleTransferCSVAliases()
	mapping := make([]DataTransferCSVColumnMapping, len(header))
	usedTargets := map[string]bool{}
	for index, sourceHeader := range header {
		normalized := normalizeTransferCSVHeader(sourceHeader)
		target, hasProfileDefault := profileDefaults[normalized]
		if hasProfileDefault && target != dataTransferCSVIgnoreTarget {
			if _, valid := VehicleTransferFieldByKey(target); !valid {
				hasProfileDefault = false
				target = ""
			}
		}
		origin := CSVMappingProfile
		if !hasProfileDefault {
			target = aliases[normalized]
			origin = CSVMappingAlias
		}
		if hasProfileDefault && target == dataTransferCSVIgnoreTarget {
			target = ""
			origin = CSVMappingIgnored
		}
		if target == "" && origin != CSVMappingIgnored {
			origin = CSVMappingUnmapped
		} else if target != "" && usedTargets[target] {
			target = ""
			origin = CSVMappingUnmapped
		} else if target != "" {
			usedTargets[target] = true
		}
		mapping[index] = DataTransferCSVColumnMapping{
			Index: index, SourceHeader: sourceHeader, NormalizedHeader: normalized,
			TargetField: target, Origin: origin,
		}
	}
	return mapping
}

func dataTransferCSVMappingDefaults(options map[string]any) map[string]string {
	defaults := map[string]string{}
	value, ok := options[dataTransferCSVMappingOption]
	if !ok {
		return defaults
	}
	values, ok := value.(map[string]any)
	if !ok {
		if typed, typedOK := value.(map[string]string); typedOK {
			for source, target := range typed {
				defaults[normalizeTransferCSVHeader(source)] = target
			}
		}
		return defaults
	}
	for source, value := range values {
		target, ok := value.(string)
		if ok {
			defaults[normalizeTransferCSVHeader(source)] = strings.TrimSpace(target)
		}
	}
	return defaults
}

func dataTransferOptionsWithCSVMapping(
	options map[string]any,
	mapping []DataTransferCSVColumnMapping,
) map[string]any {
	updated := cloneTransferOptions(options)
	defaults := make(map[string]any, len(mapping))
	for _, column := range mapping {
		switch {
		case column.TargetField != "":
			defaults[column.NormalizedHeader] = column.TargetField
		case column.Origin == CSVMappingIgnored:
			defaults[column.NormalizedHeader] = dataTransferCSVIgnoreTarget
		}
	}
	updated[dataTransferCSVMappingOption] = defaults
	return updated
}

func vehicleTransferCSVAliases() map[string]string {
	aliases := make(map[string]string, len(vehicleTransferFields)*3)
	for _, field := range vehicleTransferFields {
		for _, alias := range field.Aliases {
			aliases[normalizeTransferCSVHeader(alias)] = field.Key
		}
	}
	return aliases
}

func validateDataTransferCSVMapping(header []string, mapping []DataTransferCSVColumnMapping) error {
	if len(header) == 0 || len(mapping) != len(header) {
		return fmt.Errorf("%w: mapping must contain every CSV column", ErrDataTransferValidation)
	}
	seenIndexes := map[int]bool{}
	seenTargets := map[string]bool{}
	validOrigins := map[DataTransferCSVMappingOrigin]bool{
		CSVMappingAlias: true, CSVMappingProfile: true, CSVMappingManual: true,
		CSVMappingIgnored: true, CSVMappingUnmapped: true,
	}
	for _, column := range mapping {
		if column.Index < 0 || column.Index >= len(header) || seenIndexes[column.Index] {
			return fmt.Errorf("%w: invalid or repeated CSV mapping index", ErrDataTransferValidation)
		}
		seenIndexes[column.Index] = true
		if column.SourceHeader != header[column.Index] ||
			column.NormalizedHeader != normalizeTransferCSVHeader(column.SourceHeader) {
			return fmt.Errorf("%w: CSV mapping header changed", ErrDataTransferValidation)
		}
		if !validOrigins[column.Origin] {
			return fmt.Errorf("%w: invalid CSV mapping origin", ErrDataTransferValidation)
		}
		if column.TargetField == "" {
			if column.Origin != CSVMappingIgnored && column.Origin != CSVMappingUnmapped {
				return fmt.Errorf("%w: empty CSV target must be ignored or unmapped", ErrDataTransferValidation)
			}
			continue
		}
		if _, ok := VehicleTransferFieldByKey(column.TargetField); !ok {
			return fmt.Errorf("%w: unsupported CSV target %q", ErrDataTransferValidation, column.TargetField)
		}
		if seenTargets[column.TargetField] {
			return fmt.Errorf("%w: repeated CSV target %q", ErrDataTransferValidation, column.TargetField)
		}
		seenTargets[column.TargetField] = true
	}
	return nil
}
