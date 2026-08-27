package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestCombinedTransferPackageExcludesMasterData(t *testing.T) {
	document := DataTransferPackage{
		Format:  DataTransferPackageFormat,
		Version: DataTransferPackageVersion,
		Areas: DataTransferPackageAreas{
			Vehicles:        []TransferVehicle{{InventoryNumber: "RK-001"}},
			ExhibitionLists: []TransferExhibitionList{{Designation: "Dortmund"}},
		},
	}
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(payload, []byte("masterData")) {
		t.Fatalf("feature package leaked master data: %s", payload)
	}
	if !bytes.Contains(payload, []byte(`"format":"railkeeper-transfer"`)) ||
		!bytes.Contains(payload, []byte(`"version":3`)) {
		t.Fatalf("missing package identity: %s", payload)
	}
}

func TestDataTransferPackageVersionThreeRoundTripsVehicleSet(t *testing.T) {
	snapshot := DataTransferSnapshot{
		Vehicles: []TransferVehicle{
			{ID: "source-a", InventoryNumber: "RK-1", Manufacturer: "Roco", Name: "A", Gauge: "H0",
				Category: "Wagen", Gattung: "Reisezugwagen"},
			{ID: "source-b", InventoryNumber: "RK-2", Manufacturer: "Roco", Name: "B", Gauge: "H0",
				Category: "Wagen", Gattung: "Reisezugwagen"},
		},
		VehicleSets: []TransferVehicleSet{{
			ID: "source-set", InventoryNumber: "Set-1",
			VehicleSetInput: VehicleSetInput{
				Name: "Rheingold", Manufacturer: "Roco", Gauge: "H0", Category: "Set", Gattung: "Reisezug",
			},
			Members: []TransferVehicleSetMember{
				{SourceVehicleID: "source-a", VehicleInventoryNumber: "RK-1", Position: 1, Label: "A-Wagen"},
				{SourceVehicleID: "source-b", VehicleInventoryNumber: "RK-2", Position: 2, Label: "B-Wagen"},
			},
		}},
	}
	payload, err := marshalDataTransferPackage(snapshot, "2026-08-27T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	document, err := decodeDataTransferPackage(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if document.Version != 3 || !reflect.DeepEqual(document.Areas.VehicleSets, snapshot.VehicleSets) {
		t.Fatalf("vehicle set round trip = %#v", document)
	}
}

func TestDataTransferExportNormalizesVehicleSetsOnlyDuringSerialization(t *testing.T) {
	snapshot := validTransferVehicleSetSnapshot()
	snapshot.VehicleSets[0].Members[1].Position = 3
	singleton := snapshot.VehicleSets[0]
	singleton.ID = "source-singleton"
	singleton.InventoryNumber = "Set-Singleton"
	singleton.Members = append([]TransferVehicleSetMember(nil), singleton.Members[:1]...)
	snapshot.VehicleSets = append(snapshot.VehicleSets, singleton)

	payload, err := marshalDataTransferPackage(snapshot, "2026-08-27T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	document, err := decodeDataTransferPackage(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if len(document.Areas.VehicleSets) != 1 {
		t.Fatalf("exported vehicle sets = %#v", document.Areas.VehicleSets)
	}
	set := document.Areas.VehicleSets[0]
	if len(set.Members) != 2 || set.Members[0].Position != 1 || set.Members[1].Position != 2 {
		t.Fatalf("normalized exported members = %#v", set.Members)
	}
	if err := ValidateTransferVehicleSet(set); err != nil {
		t.Fatalf("serialized vehicle set is not importable: %v", err)
	}
	if snapshot.VehicleSets[0].Members[1].Position != 3 || len(snapshot.VehicleSets) != 2 {
		t.Fatalf("serialization mutated source snapshot: %#v", snapshot.VehicleSets)
	}
}

func TestDataTransferExportVehicleCSVUsesStableSemicolonColumns(t *testing.T) {
	snapshot := DataTransferSnapshot{Vehicles: []TransferVehicle{
		{InventoryNumber: "RK-002", Manufacturer: "Märklin", Name: "BR 218; Cargo", Gauge: "H0"},
		{InventoryNumber: "RK-001", Manufacturer: "Roco", Name: "BR 01", Gauge: "H0"},
	}}
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
	wantHeaderPrefix := []string{
		"Inventarnummer", "Hersteller", "Artikel-Nr.", "Quelle / URL", "Bezeichnung", "Spurweite", "Epoche",
		"Bahngesellschaft", "Kategorie", "Gattung", "Beschreibung",
	}
	if !slices.Equal(rows[0][:len(wantHeaderPrefix)], wantHeaderPrefix) {
		t.Fatalf("vehicle CSV header prefix = %#v, want %#v", rows[0][:len(wantHeaderPrefix)], wantHeaderPrefix)
	}
	if rows[1][0] != "RK-001" || rows[1][1] != "Roco" || rows[1][4] != "BR 01" ||
		rows[2][0] != "RK-002" || rows[2][4] != "BR 218; Cargo" {
		t.Fatalf("vehicle CSV rows were not stable and sorted: %#v", rows)
	}
}

func TestDataTransferExportVehicleCSVContainsAllScalarAndSetFields(t *testing.T) {
	payload, err := marshalDataTransferCSV(TransferVehicles, DataTransferSnapshot{Vehicles: []TransferVehicle{{
		InventoryNumber: "RK-001", Manufacturer: "Roco", Name: "BR 01", Gauge: "H0",
	}}})
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(bytes.NewReader(payload))
	reader.Comma = ';'
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || len(rows[0]) != 88 || len(rows[1]) != 88 {
		t.Fatalf("vehicle CSV dimensions = %d rows, %d/%d columns, want 2 rows and 88 columns", len(rows), len(rows[0]), len(rows[1]))
	}
	for _, header := range []string{"Länge (mm)", "Kupplung hinten", "Zusatzinformationen", "QR-Code erstellen"} {
		if !slices.Contains(rows[0], header) {
			t.Fatalf("vehicle CSV header missing %q: %#v", header, rows[0])
		}
	}
}

func TestDataTransferExportCSVNeutralizesSpreadsheetFormulas(t *testing.T) {
	snapshot := DataTransferSnapshot{Vehicles: []TransferVehicle{{
		InventoryNumber: "=HYPERLINK(\"https://example.invalid\")", Manufacturer: "+cmd|' /C calc'!A0",
		ArticleNumber: "-2+3", Name: "@SUM(1+1)", Gauge: "H0", Category: "Lokomotive",
		Gattung: "Dampflokomotive", Description: "ordinary text",
	}}}
	payload, err := marshalDataTransferCSV(TransferVehicles, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, protected := range []string{
		"'=HYPERLINK", "'+cmd", "'-2+3", "'@SUM",
	} {
		if !strings.Contains(text, protected) {
			t.Fatalf("CSV did not neutralize %q: %s", protected, text)
		}
	}
	if !strings.Contains(text, "ordinary text") {
		t.Fatalf("CSV changed ordinary text: %s", text)
	}
}

func TestDataTransferExportAccessoryCSVIncludesCurrentStockAndAssets(t *testing.T) {
	snapshot := DataTransferSnapshot{Accessories: []TransferAccessory{{
		InventoryNumber: "RK-ART-1", Manufacturer: "Viessmann", ArticleNumber: "4011", Name: "Signal",
		Category: "signal", Stock: []TransferAccessoryStock{{LocationID: "loc-1", LocationName: "Schrank", Quantity: 2}},
		Assets: []TransferAccessoryAsset{{ID: "asset-1", InventoryNumber: "RK-ASSET-1", SerialNumber: "S1"}},
	}}}
	payload, err := marshalDataTransferCSV(TransferAccessories, snapshot)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, expected := range []string{"Inventarnummer;Hersteller;Artikelnummer;Bezeichnung", "RK-ART-1", `""locationId"":""loc-1""`, `""id"":""asset-1""`} {
		if !strings.Contains(text, expected) {
			t.Fatalf("accessory CSV missing %q: %s", expected, text)
		}
	}
	for _, excluded := range []string{"purchases", "reservations", "installations", "movements"} {
		if strings.Contains(strings.ToLower(text), excluded) {
			t.Fatalf("accessory CSV leaked %s: %s", excluded, text)
		}
	}
}

func TestDataTransferExportCombinedJSONHasStableOrdering(t *testing.T) {
	snapshot := DataTransferSnapshot{
		Vehicles: []TransferVehicle{
			{ID: "v-2", InventoryNumber: "RK-010"},
			{ID: "v-1", InventoryNumber: "RK-002"},
		},
		Accessories: []TransferAccessory{
			{ID: "a-2", InventoryNumber: "RK-ART-010"},
			{ID: "a-1", InventoryNumber: "RK-ART-002"},
		},
		ExhibitionLists: []TransferExhibitionList{
			{ID: "e-2", Designation: "Zweite", Date: "2026-08-21"},
			{ID: "e-1", Designation: "Erste", Date: "2026-08-20"},
		},
	}
	payload, err := marshalDataTransferPackage(snapshot, "2026-08-20T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	var document DataTransferPackage
	if err := json.Unmarshal(payload, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != DataTransferPackageVersion || document.CreatedAt != "2026-08-20T12:00:00Z" {
		t.Fatalf("unexpected package metadata: %#v", document)
	}
	if document.Areas.Vehicles[0].InventoryNumber != "RK-002" ||
		document.Areas.Accessories[0].InventoryNumber != "RK-ART-002" ||
		document.Areas.ExhibitionLists[0].ID != "e-1" {
		t.Fatalf("unstable package ordering: %#v", document.Areas)
	}
}

func TestDataTransferExportJSONSortsVehicleSetsAndMembers(t *testing.T) {
	snapshot := DataTransferSnapshot{
		Vehicles: []TransferVehicle{
			{ID: "v-1", InventoryNumber: "RK-1"},
			{ID: "v-2", InventoryNumber: "RK-2"},
			{ID: "v-3", InventoryNumber: "RK-3"},
			{ID: "v-4", InventoryNumber: "RK-4"},
		},
		VehicleSets: []TransferVehicleSet{
			{ID: "set-z", InventoryNumber: "Set-010", Members: []TransferVehicleSetMember{
				{SourceVehicleID: "v-2", VehicleInventoryNumber: "RK-2", Position: 2},
				{SourceVehicleID: "v-1", VehicleInventoryNumber: "RK-1", Position: 1},
			}},
			{ID: "set-a", InventoryNumber: "Set-002", Members: []TransferVehicleSetMember{
				{SourceVehicleID: "v-4", VehicleInventoryNumber: "RK-4", Position: 2},
				{SourceVehicleID: "v-3", VehicleInventoryNumber: "RK-3", Position: 1},
			}},
		},
	}
	payload, err := marshalDataTransferPackage(snapshot, "2026-08-27T10:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	document, err := decodeDataTransferPackage(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	if document.Areas.VehicleSets[0].ID != "set-a" || document.Areas.VehicleSets[0].Members[0].Position != 1 ||
		document.Areas.VehicleSets[1].Members[0].Position != 1 {
		t.Fatalf("unstable vehicle set ordering: %#v", document.Areas.VehicleSets)
	}
}

func TestDataTransferExportWritesConfinedArtifactWithSHA256(t *testing.T) {
	repository := &dataTransferExportRepositoryStub{
		profile: DataTransferProfile{
			ID: "profile-1", Name: "Vehicles", Direction: TransferExport, Format: TransferCSV,
			Areas: []TransferArea{TransferVehicles}, Enabled: true,
		},
		snapshot: DataTransferSnapshot{Vehicles: []TransferVehicle{{InventoryNumber: "RK-001"}}},
	}
	dataDir := t.TempDir()
	service := NewDataTransferService(repository, dataDir)
	job, err := service.CreateExportJob(t.Context(), repository.profile.ID, "viewer-1")
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ExecuteExport(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.State != TransferJobCompleted || result.Artifact.SHA256 == "" {
		t.Fatalf("unexpected export result: %#v", result)
	}
	path, err := resolveDataTransferArtifactPath(dataDir, result.Artifact.RelativePath)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(payload)
	if result.Artifact.SHA256 != hex.EncodeToString(digest[:]) || result.Artifact.SizeBytes != int64(len(payload)) {
		t.Fatalf("artifact metadata = %#v, bytes = %d", result.Artifact, len(payload))
	}
	if filepath.Dir(path) != filepath.Join(dataDir, "exports") {
		t.Fatalf("artifact path escaped export directory: %s", path)
	}
}

func TestDataTransferArtifactPathRejectsParentSegments(t *testing.T) {
	_, err := resolveDataTransferArtifactPath(t.TempDir(), "exports/one/../../outside.json")
	if !errors.Is(err, ErrDataTransferArtifactPath) {
		t.Fatalf("expected confined path rejection, got %v", err)
	}
}

func TestDataTransferOpenArtifactRejectsSymlinkedExportDirectory(t *testing.T) {
	dataDir := t.TempDir()
	outsideDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outsideDir, "outside.csv"), []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideDir, filepath.Join(dataDir, "exports")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	repository := &dataTransferExportRepositoryStub{
		job: DataTransferJob{ID: "job-1", Areas: []TransferArea{TransferVehicles}},
		artifact: DataTransferArtifact{
			ID: "artifact-1", JobID: "job-1", RelativePath: "exports/outside.csv",
		},
	}
	service := NewDataTransferService(repository, dataDir)
	_, file, err := service.OpenArtifact(t.Context(), repository.artifact.ID)
	if file != nil {
		_ = file.Close()
	}
	if !errors.Is(err, ErrDataTransferArtifactPath) {
		t.Fatalf("expected symlinked export directory rejection, got %v", err)
	}
}

type dataTransferExportRepositoryStub struct {
	DataTransferRepository
	profile   DataTransferProfile
	job       DataTransferJob
	snapshot  DataTransferSnapshot
	artifact  DataTransferArtifact
	jobNumber int
}

func (stub *dataTransferExportRepositoryStub) GetProfile(_ context.Context, id string) (DataTransferProfile, error) {
	if id != stub.profile.ID {
		return DataTransferProfile{}, ErrDataTransferNotFound
	}
	return stub.profile, nil
}

func (stub *dataTransferExportRepositoryStub) CreateJob(_ context.Context, job DataTransferJob) (DataTransferJob, error) {
	stub.jobNumber++
	job.ID = "job-" + string(rune('0'+stub.jobNumber))
	stub.job = job
	return job, nil
}

func (stub *dataTransferExportRepositoryStub) GetJob(_ context.Context, id string) (DataTransferJob, error) {
	if id != stub.job.ID {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	return stub.job, nil
}

func (stub *dataTransferExportRepositoryStub) UpdateJob(_ context.Context, job DataTransferJob) (DataTransferJob, error) {
	stub.job = job
	return job, nil
}

func (stub *dataTransferExportRepositoryStub) ClaimExportJob(_ context.Context, id string) (DataTransferJob, error) {
	if id != stub.job.ID {
		return DataTransferJob{}, ErrDataTransferNotFound
	}
	if stub.job.State != TransferJobDraft {
		return DataTransferJob{}, ErrDataTransferConflict
	}
	stub.job.State = TransferJobRunning
	stub.job.Stage = "snapshot"
	return stub.job, nil
}

func (stub *dataTransferExportRepositoryStub) Snapshot(
	_ context.Context,
	_ []TransferArea,
) (DataTransferSnapshot, error) {
	return stub.snapshot, nil
}

func (stub *dataTransferExportRepositoryStub) CreateArtifact(
	_ context.Context,
	artifact DataTransferArtifact,
) (DataTransferArtifact, error) {
	artifact.ID = "artifact-1"
	stub.artifact = artifact
	return artifact, nil
}

func (stub *dataTransferExportRepositoryStub) GetArtifact(_ context.Context, id string) (DataTransferArtifact, error) {
	if id != stub.artifact.ID {
		return DataTransferArtifact{}, ErrDataTransferNotFound
	}
	return stub.artifact, nil
}
