package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
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
		!bytes.Contains(payload, []byte(`"version":1`)) {
		t.Fatalf("missing package identity: %s", payload)
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
	wantPrefix := strings.Join([]string{
		"Inventarnummer;Hersteller;Artikelnummer;Bezeichnung;Spurweite;Epoche;Bahngesellschaft;Kategorie;Gattung;Beschreibung",
		"RK-001;Roco;;BR 01;H0;;;;;",
		`RK-002;Märklin;;"BR 218; Cargo";H0;;;;;`,
	}, "\n")
	if !strings.HasPrefix(string(payload), wantPrefix) {
		t.Fatalf("vehicle CSV = %q, want prefix %q", payload, wantPrefix)
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
