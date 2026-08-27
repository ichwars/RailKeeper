package application

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestPreviewImportPersistsValidationIssues(t *testing.T) {
	repository, service := newDataTransferImportFixture(t)
	job := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	upload := []byte(strings.Join([]string{
		"Inventarnummer;Hersteller;Bezeichnung",
		"RK-001;;BR 218",
		"",
	}, "\n"))

	preview, err := service.UploadAndPreview(t.Context(), job.ID, "vehicles.csv", upload, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if preview.TotalRecords != 1 || preview.ErrorRecords != 1 {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	if len(preview.Issues) == 0 || preview.Issues[0].Code != "missing_manufacturer" {
		t.Fatalf("unexpected issues: %#v", preview.Issues)
	}
	persisted, err := repository.ListIssues(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(persisted) == 0 || persisted[0].Code != "missing_manufacturer" {
		t.Fatalf("issues were not persisted: %#v", persisted)
	}
	loaded, err := repository.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.State != TransferJobReviewRequired || loaded.SourceSHA256 == "" || loaded.Preview["records"] == nil {
		t.Fatalf("preview job was not persisted: %#v", loaded)
	}
	if repository.mutationCount != 1 {
		t.Fatalf("preview used %d atomic import mutations, want 1", repository.mutationCount)
	}
}

func TestPreviewImportReturnsAndAppliesVehicleCSVMapping(t *testing.T) {
	repository, service := newDataTransferImportFixture(t)
	job := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	upload := []byte("Eigene Nummer;Marke;Titel;Spur;Typ;Klasse\nRK-900;Roco;BR 218;H0;Lokomotive;Diesellokomotive\n")

	automatic, err := service.UploadAndPreview(t.Context(), job.ID, "vehicles.csv", upload, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(automatic.CSVMapping) != 6 || automatic.CSVMapping[0].SourceHeader != "Eigene Nummer" ||
		automatic.CSVMapping[0].Origin != CSVMappingUnmapped || len(automatic.VehicleFields) != 88 {
		t.Fatalf("unexpected automatic mapping: %#v, fields=%d", automatic.CSVMapping, len(automatic.VehicleFields))
	}

	columns := append([]DataTransferCSVColumnMapping(nil), automatic.CSVMapping...)
	targets := []string{"inventoryNumber", "manufacturer", "name", "gauge", "category", "gattung"}
	for index := range columns {
		columns[index].TargetField = targets[index]
		columns[index].Origin = CSVMappingManual
	}
	mapped, err := service.UploadAndPreviewWithMapping(t.Context(), job.ID, "vehicles.csv", upload, "editor-1",
		&DataTransferCSVMappingInput{Columns: columns, SaveToProfile: true})
	if err != nil {
		t.Fatal(err)
	}
	if mapped.ReadyRecords != 1 || mapped.ErrorRecords != 0 || len(mapped.Records) != 1 {
		t.Fatalf("manual mapping did not produce ready preview: %#v", mapped)
	}
	var vehicle TransferVehicle
	if err := json.Unmarshal(mapped.Records[0].Data, &vehicle); err != nil {
		t.Fatal(err)
	}
	if vehicle.InventoryNumber != "RK-900" || vehicle.Manufacturer != "Roco" || vehicle.Name != "BR 218" {
		t.Fatalf("manual mapping produced wrong vehicle: %#v", vehicle)
	}
	if repository.profiles[job.ProfileID].Options["csvMapping"] == nil {
		t.Fatalf("mapping was not saved to profile options: %#v", repository.profiles[job.ProfileID].Options)
	}
	loaded, err := repository.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Options["csvMapping"] == nil || loaded.Preview["csvMapping"] == nil {
		t.Fatalf("mapping was not persisted in job snapshot and preview: %#v %#v", loaded.Options, loaded.Preview)
	}
}

func TestDataTransferVehicleCSVFullFieldRoundTrip(t *testing.T) {
	maximumSpeed := 140
	want := TransferVehicle{
		InventoryNumber: "RK-062", Manufacturer: "Roco", ArticleNumber: "73000",
		ArticleSourceURL: "https://example.invalid/73000", Name: "BR 218", Gauge: "H0", Epoch: "IV",
		RailwayCompany: "DB", Category: "Lokomotive", Gattung: "Diesellokomotive", Description: "Test",
		Series: "218", VehicleNumber: "218 001-6", MaximumSpeedKmh: &maximumSpeed, HomeBase: "Bw Hamburg",
		Digital: true, DigitalDecoderNumber: "3", DecoderType: "LokSound 5", DTDecoder: true,
		DTDecoderNumber: "4", ExhibitionReady: true, Exhibition: true, ABCBrakes: true, EAN: "4000000000000",
		ProductionPeriod: "2025", ListPrice: "299,90", AcquisitionType: "Kauf", AcquiredFrom: "Händler",
		PurchasePrice: "249,90", PurchaseDate: "2026-01-02", StorageLocation: "Vitrine",
		StorageDetails: "Fach 2", Condition: "Sehr gut", ConditionDetails: "Eingefahren", Packaging: "OVP",
		LengthMM: "181", WeightG: "540", Color: "Ozeanblau", Lettering: "DB", Load: "",
		Interior: "Führerstand", Axles: "Bo'Bo'", AxleCount: "4", TractionTireCount: "2", Wheelset: "AC",
		CouplingSame: true, CouplingFront: "KKK", CouplingRear: "KKK", PowerPickup: "Schleifer", Adapter: "MTC21",
		DriveEnabled: true, DriveDescription: "Kardan", HeadlightsEnabled: true,
		HeadlightsDescription: "Dreilicht", LightingEnabled: true, LightingDescription: "LED",
		SoundGeneratorEnabled: true, SoundGeneratorDescription: "Diesel", SmokeGeneratorEnabled: true,
		SmokeGeneratorDescription: "Dynamisch", AdditionalInfo: "-Clubbestand", QRCodeEnabled: true,
	}
	payload, err := marshalDataTransferCSV(TransferVehicles, DataTransferSnapshot{Vehicles: []TransferVehicle{want}})
	if err != nil {
		t.Fatal(err)
	}
	gotSnapshot, _, err := parseDataTransferCSV(TransferVehicles, strings.NewReader(string(payload)))
	if err != nil {
		t.Fatal(err)
	}
	if len(gotSnapshot.Vehicles) != 1 {
		t.Fatalf("round-trip vehicles = %d, want 1", len(gotSnapshot.Vehicles))
	}
	got := gotSnapshot.Vehicles[0]
	if got.LengthMM != want.LengthMM || got.WeightG != want.WeightG || got.CouplingRear != want.CouplingRear ||
		!got.SoundGeneratorEnabled || got.AdditionalInfo != want.AdditionalInfo || !got.QRCodeEnabled ||
		got.MaximumSpeedKmh == nil || *got.MaximumSpeedKmh != maximumSpeed {
		t.Fatalf("vehicle round trip lost extended fields: %#v", got)
	}
}

func TestDataTransferVehicleSetCSVRoundTripIgnoresRowOrder(t *testing.T) {
	snapshot := DataTransferSnapshot{
		Vehicles: []TransferVehicle{
			{ID: "vehicle-a", InventoryNumber: "RK-A", Manufacturer: "Roco", Name: "Wagen A", Gauge: "H0",
				Category: "Wagen", Gattung: "Reisezugwagen"},
			{ID: "vehicle-b", InventoryNumber: "RK-B", Manufacturer: "Roco", Name: "Wagen B", Gauge: "H0",
				Category: "Wagen", Gattung: "Reisezugwagen"},
		},
		VehicleSets: []TransferVehicleSet{{
			ID: "set-source", InventoryNumber: "Set-001",
			VehicleSetInput: VehicleSetInput{
				Name: "Rheingold", Manufacturer: "Roco", ArticleNumber: "43000", Gauge: "H0", Epoch: "III",
				RailwayCompany: "DB", Category: "Set", Gattung: "Reisezug", StorageLocation: "Vitrine",
			},
			Members: []TransferVehicleSetMember{
				{SourceVehicleID: "vehicle-a", VehicleInventoryNumber: "RK-A", Position: 1, Label: "Steuerwagen"},
				{SourceVehicleID: "vehicle-b", VehicleInventoryNumber: "RK-B", Position: 2, Label: "Speisewagen"},
			},
		}},
	}
	payload, err := marshalDataTransferCSV(TransferVehicles, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows[0]) != 88 {
		t.Fatalf("vehicle set CSV columns = %d, want 88", len(rows[0]))
	}
	rows[1], rows[2] = rows[2], rows[1]
	var reordered bytes.Buffer
	writer := csv.NewWriter(&reordered)
	writer.Comma = ';'
	writer.WriteAll(rows)
	if err := writer.Error(); err != nil {
		t.Fatal(err)
	}
	got, _, err := parseDataTransferCSV(TransferVehicles, strings.NewReader(reordered.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.VehicleSets) != 1 || len(got.VehicleSets[0].Members) != 2 {
		t.Fatalf("vehicle set CSV round trip = %#v", got.VehicleSets)
	}
	set := got.VehicleSets[0]
	if set.InventoryNumber != "Set-001" || set.Name != "Rheingold" || set.ArticleNumber != "43000" ||
		set.StorageLocation != "Vitrine" || set.Members[0].Position != 1 ||
		set.Members[0].VehicleInventoryNumber != "RK-A" || set.Members[0].Label != "Steuerwagen" ||
		set.Members[1].Position != 2 {
		t.Fatalf("vehicle set CSV values = %#v", set)
	}
}

func TestDataTransferVehicleCSVWithoutSetColumnsCreatesNoSet(t *testing.T) {
	payload := strings.Join([]string{
		"Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung",
		"RK-SOLO;Roco;BR 218;H0;Lokomotive;Diesellokomotive",
	}, "\n")
	snapshot, _, err := parseDataTransferCSV(TransferVehicles, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Vehicles) != 1 || len(snapshot.VehicleSets) != 0 {
		t.Fatalf("single vehicle CSV snapshot = %#v", snapshot)
	}
}

func TestDataTransferVehicleSetCSVPreservesInvalidIntegersForReview(t *testing.T) {
	payload := strings.Join([]string{
		"Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung;" +
			"Set-Inventarnummer;Set-Bezeichnung;Set-Hersteller;Set-Spurweite;Set-Kategorie;Set-Gattung;" +
			"Set-Position;Set-Mitgliederzahl",
		"RK-A;Roco;Wagen A;H0;Wagen;Reisezugwagen;Set-001;Rheingold;Roco;H0;Set;Reisezug;x;zwei",
	}, "\n")
	snapshot, _, err := parseDataTransferCSV(TransferVehicles, strings.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.VehicleSets) != 1 || len(snapshot.VehicleSets[0].Diagnostics) != 2 {
		t.Fatalf("vehicle set diagnostics = %#v", snapshot.VehicleSets)
	}
	if snapshot.VehicleSets[0].Diagnostics[0].Code != "invalid_vehicle_set_position" ||
		snapshot.VehicleSets[0].Diagnostics[1].Code != "invalid_vehicle_set_member_count" {
		t.Fatalf("unexpected diagnostics = %#v", snapshot.VehicleSets[0].Diagnostics)
	}
}

func TestDataTransferCSVProtectionRoundTrip(t *testing.T) {
	for _, value := range []string{"=x", "  =x", "'=x", "''=x", "+x", "-x", "@x"} {
		t.Run(value, func(t *testing.T) {
			encoded := safeDataTransferCSVRow([]string{value})[0]
			decoded := decodeSafeDataTransferCSVValue(strings.TrimSpace(encoded))
			if decoded != value {
				t.Fatalf("CSV protection round trip = %q, want %q (encoded %q)", decoded, value, encoded)
			}
		})
	}
}

func TestDataTransferVehicleCSVRejectsUnknownBooleanValue(t *testing.T) {
	payload := strings.Join([]string{
		"Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung;Digital",
		"RK-BOOL;Roco;BR 218;H0;Lokomotive;Diesellokomotive;treu",
	}, "\n")

	_, _, err := parseDataTransferCSV(TransferVehicles, strings.NewReader(payload))
	if !errors.Is(err, ErrDataTransferValidation) {
		t.Fatalf("unknown boolean error = %v, want validation error", err)
	}
	if err == nil || !strings.Contains(err.Error(), "Digital") || !strings.Contains(err.Error(), "row 2") {
		t.Fatalf("unknown boolean error lacks field or row context: %v", err)
	}
}

func TestPreviewImportRejectsRecordsRejectedByRegularAggregateValidation(t *testing.T) {
	incoming := DataTransferSnapshot{
		Vehicles: []TransferVehicle{{
			InventoryNumber: "RK-INVALID-VEHICLE", Manufacturer: "Roco", Name: "BR 01",
		}},
		Accessories: []TransferAccessory{{
			InventoryNumber: "RK-INVALID-ACCESSORY", Manufacturer: "Viessmann", Name: "Signal",
			Category: "Signal", TrackingMode: "quantity", ArticleType: "invalid", Subtype: "invalid:signal",
			ListPrice: "not-money", PackageQuantity: 1, StockUnit: "piece", InventoryStrategy: "quantity",
		}},
	}
	records, issues := classifyDataTransferImport("job-invalid", incoming, DataTransferSnapshot{}, nil)
	for _, code := range []string{"missing_gauge", "missing_category", "missing_gattung", "invalid_accessory"} {
		assertTransferIssueCode(t, issues, code)
	}
	if records[0].Classification != "error" || records[1].Classification != "error" {
		t.Fatalf("aggregate-invalid records were not blocked: %#v", records)
	}
}

func TestTransferImportRejectsUnknownMasterDataAndUnsupportedVersion(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		payload string
	}{
		{
			name:    "unknown master data area",
			payload: `{"format":"railkeeper-transfer","version":1,"areas":{"vehicles":[],"masterData":[]}}`,
		},
		{
			name:    "unsupported package version",
			payload: `{"format":"railkeeper-transfer","version":4,"areas":{"vehicles":[]}}`,
		},
		{
			name:    "empty areas",
			payload: `{"format":"railkeeper-transfer","version":1,"areas":{}}`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, service := newDataTransferImportFixture(t)
			job := fixtureCreateImportJob(t, service, TransferVehicles, TransferJSON)
			_, err := service.UploadAndPreview(
				t.Context(), job.ID, "transfer.json", []byte(testCase.payload), "editor-1",
			)
			if !errors.Is(err, ErrDataTransferValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestTransferImportAcceptsLegacyPackageVersionOne(t *testing.T) {
	document, err := decodeDataTransferPackage(strings.NewReader(
		`{"format":"railkeeper-transfer","version":1,"createdAt":"2026-01-01T00:00:00Z",` +
			`"areas":{"vehicles":[]}}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if document.Version != DataTransferPackageLegacyVersion || document.Areas.Vehicles == nil {
		t.Fatalf("legacy package decoded incorrectly: %#v", document)
	}
}

func TestTransferImportAcceptsPreviousPackageVersionTwo(t *testing.T) {
	document, err := decodeDataTransferPackage(strings.NewReader(
		`{"format":"railkeeper-transfer","version":2,"createdAt":"2026-01-01T00:00:00Z",` +
			`"areas":{"vehicles":[{"inventoryNumber":"RK-2","manufacturer":"Roco",` +
			`"name":"BR 218","gauge":"H0","category":"Lokomotive","gattung":"Diesellokomotive",` +
			`"digital":false,"dtDecoder":false,"exhibitionReady":false,"exhibition":false,` +
			`"abcBrakes":false,"couplingSame":false,"driveEnabled":false,"headlightsEnabled":false,` +
			`"lightingEnabled":false,"soundGeneratorEnabled":false,"smokeGeneratorEnabled":false,` +
			`"qrCodeEnabled":false}]}}`,
	))
	if err != nil {
		t.Fatal(err)
	}
	if document.Version != DataTransferPackagePreviousVersion || len(document.Areas.Vehicles) != 1 ||
		document.Areas.VehicleSets != nil {
		t.Fatalf("version two package decoded incorrectly: %#v", document)
	}
}

func TestTransferImportRejectsVersionTwoVehicleFieldsInLegacyPackage(t *testing.T) {
	_, err := decodeDataTransferPackage(strings.NewReader(
		`{"format":"railkeeper-transfer","version":1,"areas":{"vehicles":[{` +
			`"inventoryNumber":"RK-LEGACY","manufacturer":"Roco","name":"BR 218","gauge":"H0",` +
			`"lengthMm":"181"}]}}`,
	))
	if !errors.Is(err, ErrDataTransferValidation) {
		t.Fatalf("version 2 field in legacy package error = %v, want validation error", err)
	}
}

func TestPreviewImportClassifiesDuplicatesAndLockedExhibitionReplacement(t *testing.T) {
	repository, service := newDataTransferImportFixture(t)
	repository.snapshot = DataTransferSnapshot{
		Vehicles: []TransferVehicle{{ID: "vehicle-1", InventoryNumber: "RK-001", UpdatedAt: "2026-08-20T10:00:00Z"}},
		Accessories: []TransferAccessory{
			{ID: "accessory-1", InventoryNumber: "RK-A-1", Manufacturer: "Viessmann", ArticleNumber: "4011"},
		},
		ExhibitionLists: []TransferExhibitionList{
			{ID: "list-1", Designation: "Dortmund", Date: "2026-08-20", Locked: true},
		},
	}

	vehicleJob := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	vehiclePreview, err := service.UploadAndPreview(t.Context(), vehicleJob.ID, "vehicles.csv", []byte(
		"Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n"+
			"RK-001;Roco;BR 01;H0;Lokomotive;Dampflokomotive\n"), "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertTransferIssueCode(t, vehiclePreview.Issues, "duplicate_inventory_number")
	if vehiclePreview.Records[0].TargetFingerprint == "" {
		t.Fatal("vehicle duplicate preview has no aggregate target fingerprint")
	}
	persistedVehicleJob, err := repository.GetJob(t.Context(), vehicleJob.ID)
	if err != nil {
		t.Fatal(err)
	}
	persistedPreview, err := json.Marshal(persistedVehicleJob.Preview)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(persistedPreview), `"targetFingerprint"`) {
		t.Fatalf("persisted preview omitted target fingerprint: %s", persistedPreview)
	}

	accessoryJob := fixtureCreateImportJob(t, service, TransferAccessories, TransferCSV)
	accessoryPreview, err := service.UploadAndPreview(t.Context(), accessoryJob.ID, "accessories.csv", []byte(
		"Inventarnummer;Hersteller;Artikelnummer;Bezeichnung;Kategorie;Erfassungsart;Artikelart;Unterart;"+
			"Packungsmenge;Bestandseinheit;Inventarstrategie\n"+
			"RK-A-NEW;Viessmann;4011;Signal;Signal;quantity;other;other:other;1;piece;quantity\n"), "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertTransferIssueCode(t, accessoryPreview.Issues, "matching_manufacturer_article_number")
	if accessoryPreview.Records[0].TargetFingerprint == "" {
		t.Fatal("accessory duplicate preview has no aggregate target fingerprint")
	}

	exhibitionJob := fixtureCreateImportJob(t, service, TransferExhibitionLists, TransferJSON)
	payload, err := json.Marshal(DataTransferPackage{
		Format: DataTransferPackageFormat, Version: DataTransferPackageVersion,
		Areas: DataTransferPackageAreas{ExhibitionLists: []TransferExhibitionList{{
			ID: "list-1", Designation: "Dortmund", Date: "2026-08-20",
			Entries: []TransferExhibitionEntry{
				{VehicleID: "vehicle-1", LocomotiveName: "ID without installation proof"},
				{VehicleID: "foreign-vehicle", VehicleInventoryNumber: "RK-001", LocomotiveName: "Foreign ID"},
			},
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	exhibitionPreview, err := service.UploadAndPreview(
		t.Context(), exhibitionJob.ID, "transfer.json", payload, "editor-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	assertTransferIssueCode(t, exhibitionPreview.Issues, "locked_exhibition_list")
	assertTransferIssueCode(t, exhibitionPreview.Issues, "missing_vehicle_reference")
	assertTransferIssueCode(t, exhibitionPreview.Issues, "exhibition_vehicle_reference")
	if exhibitionPreview.Records[0].TargetFingerprint == "" {
		t.Fatal("exhibition duplicate preview has no aggregate target fingerprint")
	}
	if exhibitionPreview.ErrorRecords != 1 {
		t.Fatalf("locked replacement was not classified as an error: %#v", exhibitionPreview)
	}
}

func TestPreviewImportDoesNotTrustForeignExhibitionListIDWithoutMatchingIdentity(t *testing.T) {
	repository, service := newDataTransferImportFixture(t)
	repository.snapshot = DataTransferSnapshot{ExhibitionLists: []TransferExhibitionList{{
		ID: "list-collision", Designation: "Local show", Date: "2026-08-20", Locked: true,
	}}}
	job := fixtureCreateImportJob(t, service, TransferExhibitionLists, TransferJSON)
	payload, err := json.Marshal(DataTransferPackage{
		Format: DataTransferPackageFormat, Version: DataTransferPackageVersion,
		Areas: DataTransferPackageAreas{ExhibitionLists: []TransferExhibitionList{{
			ID: "list-collision", Designation: "Foreign show", Date: "2027-01-10",
		}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := service.UploadAndPreview(t.Context(), job.ID, "transfer.json", payload, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Issues) != 0 || preview.ReadyRecords != 1 || preview.Records[0].TargetID != "" ||
		preview.Records[0].ProposedAction != "create" {
		t.Fatalf("foreign ID collision was trusted as local identity: %#v", preview)
	}
}

func TestDataTransferPreviewIssuesHaveDeterministicRecordAndEntryIdentity(t *testing.T) {
	incoming := DataTransferSnapshot{
		Vehicles: []TransferVehicle{
			{InventoryNumber: "RK-DUP", Manufacturer: "Roco", Name: "First"},
			{InventoryNumber: "RK-DUP", Manufacturer: "Roco", Name: "Second"},
		},
		ExhibitionLists: []TransferExhibitionList{{
			Designation: "Identity", Date: "2026-08-20", Entries: []TransferExhibitionEntry{
				{VehicleID: "foreign-a", VehicleInventoryNumber: "RK-A"},
				{VehicleID: "foreign-b", VehicleInventoryNumber: "RK-B"},
			},
		}},
	}
	records, issues := classifyDataTransferImport("job-identity", incoming, DataTransferSnapshot{}, nil)
	_, repeated := classifyDataTransferImport("job-identity", incoming, DataTransferSnapshot{}, nil)
	if records[0].RowNumber == nil || *records[0].RowNumber != 1 ||
		records[1].RowNumber == nil || *records[1].RowNumber != 2 {
		t.Fatalf("JSON record rows are not deterministic: %#v", records[:2])
	}
	if len(issues) != len(repeated) {
		t.Fatalf("issue count changed: %d != %d", len(issues), len(repeated))
	}
	seenIDs := map[string]bool{}
	entryFields := map[string]bool{}
	for index, issue := range issues {
		if issue.ID == "" || issue.ID != repeated[index].ID || seenIDs[issue.ID] {
			t.Fatalf("issue identity is not deterministic and unique: %#v / %#v", issues, repeated)
		}
		seenIDs[issue.ID] = true
		if issue.Code == "missing_vehicle_reference" {
			entryFields[issue.Field] = true
		}
	}
	if !entryFields["entries[0].vehicleReference"] || !entryFields["entries[1].vehicleReference"] {
		t.Fatalf("entry issues are not independently keyed: %#v", issues)
	}
}

func TestTransferImportRejectsMalformedOversizedAndMismatchedUploads(t *testing.T) {
	_, service := newDataTransferImportFixture(t)

	malformedJob := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	if _, err := service.UploadAndPreview(t.Context(), malformedJob.ID, "vehicles.csv",
		[]byte("Inventarnummer;Hersteller;Bezeichnung\n\"unterminated"), "editor-1"); !errors.Is(err, ErrDataTransferValidation) {
		t.Fatalf("expected malformed CSV rejection, got %v", err)
	}

	mismatchJob := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	if _, err := service.UploadAndPreview(t.Context(), mismatchJob.ID, "vehicles.json",
		[]byte("Inventarnummer;Hersteller;Bezeichnung\nRK-1;Roco;BR 01\n"), "editor-1"); !errors.Is(err, ErrDataTransferValidation) {
		t.Fatalf("expected MIME/extension mismatch rejection, got %v", err)
	}

	oversizedJob := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	oversized := make([]byte, DataTransferMaxUploadBytes+1)
	if _, err := service.UploadAndPreview(t.Context(), oversizedJob.ID, "vehicles.csv", oversized, "editor-1"); !errors.Is(err, ErrDataTransferUploadTooLarge) {
		t.Fatalf("expected oversized upload rejection, got %v", err)
	}
}

func TestTransferImportEnforcesAreaScopeAndSupportsIssueResolutionAndCancel(t *testing.T) {
	_, service := newDataTransferImportFixture(t)
	job := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	if _, err := service.UploadAndPreview(t.Context(), job.ID, "vehicles.csv", []byte(
		"Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n"+
			"RK-001;;BR 01;H0;Lokomotive;Dampflokomotive\n"),
		"messe-1", TransferExhibitionLists); !errors.Is(err, ErrDataTransferForbidden) {
		t.Fatalf("expected Messe area rejection, got %v", err)
	}

	preview, err := service.UploadAndPreview(t.Context(), job.ID, "vehicles.csv", []byte(
		"Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n"+
			"RK-001;;BR 01;H0;Lokomotive;Dampflokomotive\n"), "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := service.ResolveIssue(t.Context(), job.ID, preview.Issues[0].ID, "skip", "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if resolved.State != TransferJobReady {
		t.Fatalf("resolved job state = %q, want ready", resolved.State)
	}
	cancelled, err := service.CancelJob(t.Context(), job.ID, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if cancelled.State != TransferJobCancelled {
		t.Fatalf("cancelled job state = %q", cancelled.State)
	}
	if repository := service.repository.(*dataTransferImportRepositoryStub); repository.mutationCount != 3 {
		t.Fatalf("upload, resolve, and cancel used %d atomic mutations, want 3", repository.mutationCount)
	}
}

func TestDataTransferConfirmImportRequiresExplicitConfirmationAndReadyResolvedPreview(t *testing.T) {
	repository, service := newDataTransferImportFixture(t)
	job := fixtureCreateImportJob(t, service, TransferVehicles, TransferCSV)
	job.State = TransferJobReady
	job.SourceSHA256 = "source-sha"
	job.Preview = map[string]any{"sourceSha256": "source-sha", "records": []any{}}
	repository.jobs[job.ID] = job

	if _, err := service.ConfirmImport(t.Context(), job.ID, job.Revision, false, "editor-1"); !errors.Is(err, ErrDataTransferValidation) {
		t.Fatalf("ConfirmImport(false) error = %v, want validation", err)
	}
	if repository.applyCount != 0 {
		t.Fatalf("false confirmation applied %d imports", repository.applyCount)
	}

	repository.issues[job.ID] = []DataTransferIssue{{
		ID: "issue-1", JobID: job.ID, Area: TransferVehicles, RecordKey: "RK-1",
		Severity: TransferIssueWarning, Code: "duplicate_inventory_number",
	}}
	if _, err := service.ConfirmImport(t.Context(), job.ID, job.Revision, true, "editor-1"); !errors.Is(err, ErrDataTransferConflict) {
		t.Fatalf("ConfirmImport(unresolved) error = %v, want conflict", err)
	}
	repository.issues[job.ID][0].SelectedResolution = "skip"

	if _, err := service.ConfirmImport(
		t.Context(), job.ID, job.Revision, true, "messe-1", TransferExhibitionLists,
	); !errors.Is(err, ErrDataTransferForbidden) {
		t.Fatalf("ConfirmImport(Messe vehicles) error = %v, want forbidden", err)
	}

	completed, err := service.ConfirmImport(t.Context(), job.ID, job.Revision, true, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if repository.applyCount != 1 || completed.State != TransferJobCompleted ||
		completed.ConfirmedByUserID != "editor-1" {
		t.Fatalf("unexpected confirmed import: count=%d job=%#v", repository.applyCount, completed)
	}
}

func newDataTransferImportFixture(t *testing.T) (*dataTransferImportRepositoryStub, *DataTransferService) {
	t.Helper()
	repository := &dataTransferImportRepositoryStub{
		profiles: map[string]DataTransferProfile{}, jobs: map[string]DataTransferJob{},
		issues: map[string][]DataTransferIssue{},
	}
	return repository, NewDataTransferService(repository, t.TempDir())
}

func fixtureCreateImportJob(
	t *testing.T,
	service *DataTransferService,
	area TransferArea,
	format TransferFormat,
) DataTransferJob {
	t.Helper()
	repository := service.repository.(*dataTransferImportRepositoryStub)
	profileID := "profile-" + string(area) + "-" + string(format)
	repository.profiles[profileID] = DataTransferProfile{
		ID: profileID, Name: profileID, Direction: TransferImport, Format: format,
		Areas: []TransferArea{area}, Enabled: true,
	}
	job, err := service.CreateImportJob(t.Context(), profileID, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	return job
}

func assertTransferIssueCode(t *testing.T, issues []DataTransferIssue, code string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Code == code {
			return
		}
	}
	t.Fatalf("missing issue %q in %#v", code, issues)
}

type dataTransferImportRepositoryStub struct {
	DataTransferRepository
	profiles      map[string]DataTransferProfile
	jobs          map[string]DataTransferJob
	issues        map[string][]DataTransferIssue
	snapshot      DataTransferSnapshot
	nextID        int
	mutationCount int
	applyCount    int
}

func (repository *dataTransferImportRepositoryStub) GetProfile(
	_ context.Context,
	id string,
) (DataTransferProfile, error) {
	profile, found := repository.profiles[id]
	if !found {
		return DataTransferProfile{}, ErrDataTransferNotFound
	}
	return profile, nil
}

func (repository *dataTransferImportRepositoryStub) UpdateProfile(
	_ context.Context,
	profile DataTransferProfile,
) (DataTransferProfile, error) {
	if _, found := repository.profiles[profile.ID]; !found {
		return DataTransferProfile{}, ErrDataTransferNotFound
	}
	repository.profiles[profile.ID] = profile
	return profile, nil
}

func (repository *dataTransferImportRepositoryStub) CreateJob(
	_ context.Context,
	job DataTransferJob,
) (DataTransferJob, error) {
	repository.nextID++
	job.ID = "job-" + string(rune('0'+repository.nextID))
	job.Revision = 1
	repository.jobs[job.ID] = job
	return job, nil
}

func (repository *dataTransferImportRepositoryStub) CompareAndUpdateImportJob(
	_ context.Context,
	mutation DataTransferImportMutation,
) (DataTransferJob, error) {
	current, found := repository.jobs[mutation.Job.ID]
	if !found {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	if current.State != mutation.ExpectedState || current.Revision != mutation.ExpectedRevision {
		return DataTransferJob{}, ErrDataTransferConflict
	}
	repository.mutationCount++
	if mutation.ProfileOptions != nil {
		profile, found := repository.profiles[mutation.ProfileOptions.ProfileID]
		if !found || profile.UpdatedAt != mutation.ProfileOptions.ExpectedUpdatedAt {
			return DataTransferJob{}, ErrDataTransferConflict
		}
		profile.Options = cloneTransferOptions(mutation.ProfileOptions.Options)
		repository.profiles[profile.ID] = profile
	}
	mutation.Job.Revision = current.Revision + 1
	repository.jobs[mutation.Job.ID] = mutation.Job
	if mutation.ReplaceIssues {
		if err := repository.ReplaceIssues(context.Background(), mutation.Job.ID, mutation.Issues); err != nil {
			return DataTransferJob{}, err
		}
	}
	return mutation.Job, nil
}

func (repository *dataTransferImportRepositoryStub) GetJob(
	_ context.Context,
	id string,
) (DataTransferJob, error) {
	job, found := repository.jobs[id]
	if !found {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	return job, nil
}

func (repository *dataTransferImportRepositoryStub) UpdateJob(
	_ context.Context,
	job DataTransferJob,
) (DataTransferJob, error) {
	if _, found := repository.jobs[job.ID]; !found {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	repository.jobs[job.ID] = job
	return job, nil
}

func (repository *dataTransferImportRepositoryStub) ReplaceIssues(
	_ context.Context,
	jobID string,
	issues []DataTransferIssue,
) error {
	stored := make([]DataTransferIssue, len(issues))
	for index, issue := range issues {
		if issue.ID == "" {
			repository.nextID++
			issue.ID = "issue-" + string(rune('0'+repository.nextID))
		}
		stored[index] = issue
	}
	repository.issues[jobID] = stored
	return nil
}

func (repository *dataTransferImportRepositoryStub) ListIssues(
	_ context.Context,
	jobID string,
) ([]DataTransferIssue, error) {
	return append([]DataTransferIssue(nil), repository.issues[jobID]...), nil
}

func (repository *dataTransferImportRepositoryStub) Snapshot(
	_ context.Context,
	_ []TransferArea,
) (DataTransferSnapshot, error) {
	return repository.snapshot, nil
}

func (repository *dataTransferImportRepositoryStub) ApplyImport(
	_ context.Context,
	job DataTransferJob,
	actor string,
) error {
	current, found := repository.jobs[job.ID]
	if !found || current.State != TransferJobReady || current.Revision != job.Revision {
		return ErrDataTransferConflict
	}
	repository.applyCount++
	current.State = TransferJobCompleted
	current.Stage = "completed"
	current.ConfirmedByUserID = actor
	current.Revision++
	repository.jobs[job.ID] = current
	return nil
}

func (repository *dataTransferImportRepositoryStub) ApplyImportWithPolicy(
	ctx context.Context,
	job DataTransferJob,
	actor string,
	_ DataTransferImportPolicy,
) error {
	return repository.ApplyImport(ctx, job, actor)
}

func (repository *dataTransferImportRepositoryStub) ValidateTransferAccessoryReferences(
	_ context.Context,
	_ TransferAccessory,
	_ string,
) error {
	return nil
}
