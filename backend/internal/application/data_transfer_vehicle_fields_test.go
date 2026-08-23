package application

import (
	"errors"
	"testing"
)

func TestVehicleTransferFieldCatalogContainsStable62Fields(t *testing.T) {
	fields := VehicleTransferFields()
	if len(fields) != 62 {
		t.Fatalf("vehicle transfer fields = %d, want 62", len(fields))
	}
	for _, key := range []string{
		"inventoryNumber",
		"maximumSpeedKmh",
		"lengthMm",
		"couplingSame",
		"qrCodeEnabled",
	} {
		if _, ok := VehicleTransferFieldByKey(key); !ok {
			t.Fatalf("missing vehicle transfer field %q", key)
		}
	}
}

func TestDataTransferCSVMappingRecognizesAliasesAndLeavesUnknownColumnsOpen(t *testing.T) {
	mapping := defaultDataTransferCSVMapping([]string{"Inventarnummer", "Vereinsnotiz"}, nil)
	if len(mapping) != 2 {
		t.Fatalf("mapping length = %d, want 2", len(mapping))
	}
	if mapping[0].TargetField != "inventoryNumber" || mapping[0].Origin != CSVMappingAlias {
		t.Fatalf("unexpected automatic mapping: %#v", mapping[0])
	}
	if mapping[1].TargetField != "" || mapping[1].Origin != CSVMappingUnmapped {
		t.Fatalf("unexpected unknown mapping: %#v", mapping[1])
	}
}

func TestDataTransferCSVMappingRecognizesLegacyArticleNumberAndRepairsInvalidDefault(t *testing.T) {
	mapping := defaultDataTransferCSVMapping(
		[]string{"Artikelnummer", "Hersteller"},
		map[string]string{"artikelnummer": "removedField", "hersteller": "removedField"},
	)
	if mapping[0].TargetField != "articleNumber" || mapping[0].Origin != CSVMappingAlias {
		t.Fatalf("legacy article number mapping = %#v", mapping[0])
	}
	if mapping[1].TargetField != "manufacturer" || mapping[1].Origin != CSVMappingAlias {
		t.Fatalf("invalid saved default was not repaired: %#v", mapping[1])
	}
}

func TestDataTransferCSVMappingRejectsDuplicateTargets(t *testing.T) {
	header := []string{"Inventarnummer", "Vereinsnummer"}
	mapping := []DataTransferCSVColumnMapping{
		{Index: 0, SourceHeader: header[0], NormalizedHeader: "inventarnummer", TargetField: "inventoryNumber", Origin: CSVMappingAlias},
		{Index: 1, SourceHeader: header[1], NormalizedHeader: "vereinsnummer", TargetField: "inventoryNumber", Origin: CSVMappingManual},
	}
	if err := validateDataTransferCSVMapping(header, mapping); !errors.Is(err, ErrDataTransferValidation) {
		t.Fatalf("duplicate target error = %v, want validation error", err)
	}
}
