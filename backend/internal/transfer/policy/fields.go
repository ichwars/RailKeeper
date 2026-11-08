package policy

import (
	"encoding/json"
	"fmt"
)

const dataTransferCSVMappingOption = "csvMapping"

const dataTransferCSVIgnoreTarget = "__ignore__"

var vehicleTransferFields = []VehicleTransferField{
	newVehicleTransferField("inventoryNumber", "Inventarnummer", "Inventory number", VehicleTransferString),
	newVehicleTransferField("manufacturer", "Hersteller", "Manufacturer", VehicleTransferString),
	newVehicleTransferField("articleNumber", "Artikel-Nr.", "Article number", VehicleTransferString, "Artikelnummer"),
	newVehicleTransferField("articleSourceUrl", "Quelle / URL", "Source / URL", VehicleTransferString),
	newVehicleTransferField("name", "Bezeichnung", "Name", VehicleTransferString, "designation"),
	newVehicleTransferField("gauge", "Spurweite", "Gauge", VehicleTransferString),
	newVehicleTransferField("epoch", "Epoche", "Epoch", VehicleTransferString),
	newVehicleTransferField("railwayCompany", "Bahngesellschaft", "Railway company", VehicleTransferString),
	newVehicleTransferField("category", "Kategorie", "Category", VehicleTransferString),
	newVehicleTransferField("gattung", "Gattung", "Class", VehicleTransferString),
	newVehicleTransferField("description", "Beschreibung", "Description", VehicleTransferString),
	newVehicleTransferField("series", "Baureihe", "Series", VehicleTransferString),
	newVehicleTransferField("vehicleNumber", "Fahrzeug-Nr.", "Vehicle number", VehicleTransferString),
	newVehicleTransferField("maximumSpeedKmh", "Höchstgeschwindigkeit", "Maximum speed", VehicleTransferInteger),
	newVehicleTransferField("homeBase", "Heimat-Bw / Einsatzstelle", "Home base", VehicleTransferString),
	newVehicleTransferField("digital", "Digital", "Digital", VehicleTransferBoolean),
	newVehicleTransferField("digitalDecoderNumber", "Digital / Decoder-Nr.", "Digital / decoder number", VehicleTransferString),
	newVehicleTransferField("decoderType", "Decoder-Typ", "Decoder type", VehicleTransferString),
	newVehicleTransferField("dtDecoder", "DT / Decoder", "DT / decoder", VehicleTransferBoolean),
	newVehicleTransferField("dtDecoderNumber", "DT / Decoder-Nr.", "DT / decoder number", VehicleTransferString),
	newVehicleTransferField("exhibitionReady", "Messe tauglich", "Exhibition ready", VehicleTransferBoolean),
	newVehicleTransferField("exhibition", "Ausstellung", "Exhibition", VehicleTransferBoolean),
	newVehicleTransferField("abcBrakes", "ABC-Bremsen", "ABC brakes", VehicleTransferBoolean),
	newVehicleTransferField("ean", "EAN", "EAN", VehicleTransferString),
	newVehicleTransferField("productionPeriod", "Produktionszeit", "Production period", VehicleTransferString),
	newVehicleTransferField("listPrice", "Listenpreis", "List price", VehicleTransferString),
	newVehicleTransferField("acquisitionType", "Erwerbsart", "Acquisition type", VehicleTransferString),
	newVehicleTransferField("acquiredFrom", "Erworben von/bei", "Acquired from", VehicleTransferString),
	newVehicleTransferField("purchasePrice", "Kaufpreis", "Purchase price", VehicleTransferString),
	newVehicleTransferField("purchaseDate", "Kaufdatum", "Purchase date", VehicleTransferString),
	newVehicleTransferField("storageLocation", "Lagerort", "Storage location", VehicleTransferString),
	newVehicleTransferField("storageDetails", "Lagerdetails", "Storage details", VehicleTransferString),
	newVehicleTransferField("condition", "Zustand", "Condition", VehicleTransferString),
	newVehicleTransferField("conditionDetails", "Zustandsdetails", "Condition details", VehicleTransferString),
	newVehicleTransferField("packaging", "Verpackung", "Packaging", VehicleTransferString),
	newVehicleTransferField("lengthMm", "Länge (mm)", "Length (mm)", VehicleTransferString),
	newVehicleTransferField("weightG", "Gewicht (g)", "Weight (g)", VehicleTransferString),
	newVehicleTransferField("color", "Farbe", "Color", VehicleTransferString),
	newVehicleTransferField("lettering", "Beschriftung", "Lettering", VehicleTransferString),
	newVehicleTransferField("load", "Beladung", "Load", VehicleTransferString),
	newVehicleTransferField("interior", "Inneneinrichtung", "Interior", VehicleTransferString),
	newVehicleTransferField("axles", "Achsen", "Axles", VehicleTransferString),
	newVehicleTransferField("axleCount", "Anzahl Achsen", "Axle count", VehicleTransferString),
	newVehicleTransferField("tractionTireCount", "Anzahl Haftreifen", "Traction tire count", VehicleTransferString),
	newVehicleTransferField("wheelset", "Radsatz", "Wheelset", VehicleTransferString),
	newVehicleTransferField("couplingSame", "Kupplung (V=H)", "Same coupling front/rear", VehicleTransferBoolean),
	newVehicleTransferField("couplingFront", "Kupplung vorne", "Front coupling", VehicleTransferString),
	newVehicleTransferField("couplingRear", "Kupplung hinten", "Rear coupling", VehicleTransferString),
	newVehicleTransferField("powerPickup", "Stromabnahme", "Power pickup", VehicleTransferString),
	newVehicleTransferField("adapter", "Adapter / Schnittstelle", "Adapter / interface", VehicleTransferString),
	newVehicleTransferField("driveEnabled", "Antrieb", "Drive", VehicleTransferBoolean),
	newVehicleTransferField("driveDescription", "Antrieb Beschreibung", "Drive description", VehicleTransferString),
	newVehicleTransferField("headlightsEnabled", "Fahrlicht", "Headlights", VehicleTransferBoolean),
	newVehicleTransferField("headlightsDescription", "Fahrlicht Beschreibung", "Headlights description", VehicleTransferString),
	newVehicleTransferField("lightingEnabled", "Beleuchtung", "Lighting", VehicleTransferBoolean),
	newVehicleTransferField("lightingDescription", "Beleuchtung Beschreibung", "Lighting description", VehicleTransferString),
	newVehicleTransferField("soundGeneratorEnabled", "Soundgenerator", "Sound generator", VehicleTransferBoolean),
	newVehicleTransferField("soundGeneratorDescription", "Soundgenerator Beschreibung", "Sound generator description", VehicleTransferString),
	newVehicleTransferField("smokeGeneratorEnabled", "Rauchgenerator", "Smoke generator", VehicleTransferBoolean),
	newVehicleTransferField("smokeGeneratorDescription", "Rauchgenerator Beschreibung", "Smoke generator description", VehicleTransferString),
	newVehicleTransferField("additionalInfo", "Zusatzinformationen", "Additional information", VehicleTransferString),
	newVehicleTransferField("qrCodeEnabled", "QR-Code erstellen", "Create QR code", VehicleTransferBoolean),
}

func newVehicleTransferField(
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
	return cloneVehicleTransferFields(vehicleTransferFields)
}

func cloneVehicleTransferFields(source []VehicleTransferField) []VehicleTransferField {
	fields := make([]VehicleTransferField, len(source))
	copy(fields, source)
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

func PreserveUnmappedTransferVehicleFields(
	incoming TransferVehicle,
	existing TransferVehicle,
	mapping []DataTransferCSVColumnMapping,
) (TransferVehicle, error) {
	mappedFields := make(map[string]bool, len(mapping))
	for _, column := range mapping {
		if column.TargetField != "" {
			mappedFields[column.TargetField] = true
		}
	}
	existingValues, err := transferVehicleJSONFields(existing)
	if err != nil {
		return TransferVehicle{}, err
	}
	incomingValues, err := transferVehicleJSONFields(incoming)
	if err != nil {
		return TransferVehicle{}, err
	}
	for _, field := range vehicleTransferFields {
		if mappedFields[field.Key] {
			continue
		}
		if value, found := existingValues[field.Key]; found {
			incomingValues[field.Key] = value
		} else {
			delete(incomingValues, field.Key)
		}
	}
	payload, err := json.Marshal(incomingValues)
	if err != nil {
		return TransferVehicle{}, fmt.Errorf("encode merged transfer vehicle: %w", err)
	}
	merged := TransferVehicle{}
	if err := json.Unmarshal(payload, &merged); err != nil {
		return TransferVehicle{}, fmt.Errorf("decode merged transfer vehicle: %w", err)
	}
	return merged, nil
}

func transferVehicleJSONFields(vehicle TransferVehicle) (map[string]json.RawMessage, error) {
	payload, err := json.Marshal(vehicle)
	if err != nil {
		return nil, fmt.Errorf("encode transfer vehicle fields: %w", err)
	}
	fields := map[string]json.RawMessage{}
	if err := json.Unmarshal(payload, &fields); err != nil {
		return nil, fmt.Errorf("decode transfer vehicle fields: %w", err)
	}
	return fields, nil
}
