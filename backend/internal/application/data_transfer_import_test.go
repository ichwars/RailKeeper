package application

import (
	"context"
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
			payload: `{"format":"railkeeper-transfer","version":2,"areas":{"vehicles":[]}}`,
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
		"Inventarnummer;Hersteller;Bezeichnung\nRK-001;Roco;BR 01\n"), "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertTransferIssueCode(t, vehiclePreview.Issues, "duplicate_inventory_number")

	accessoryJob := fixtureCreateImportJob(t, service, TransferAccessories, TransferCSV)
	accessoryPreview, err := service.UploadAndPreview(t.Context(), accessoryJob.ID, "accessories.csv", []byte(
		"Inventarnummer;Hersteller;Artikelnummer;Bezeichnung\nRK-A-NEW;Viessmann;4011;Signal\n"), "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	assertTransferIssueCode(t, accessoryPreview.Issues, "matching_manufacturer_article_number")

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
	if exhibitionPreview.ErrorRecords != 1 {
		t.Fatalf("locked replacement was not classified as an error: %#v", exhibitionPreview)
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
		"Inventarnummer;Hersteller;Bezeichnung\nRK-001;;BR 01\n"), "messe-1", TransferExhibitionLists); !errors.Is(err, ErrDataTransferForbidden) {
		t.Fatalf("expected Messe area rejection, got %v", err)
	}

	preview, err := service.UploadAndPreview(t.Context(), job.ID, "vehicles.csv", []byte(
		"Inventarnummer;Hersteller;Bezeichnung\nRK-001;;BR 01\n"), "editor-1")
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
	profiles map[string]DataTransferProfile
	jobs     map[string]DataTransferJob
	issues   map[string][]DataTransferIssue
	snapshot DataTransferSnapshot
	nextID   int
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

func (repository *dataTransferImportRepositoryStub) CreateJob(
	_ context.Context,
	job DataTransferJob,
) (DataTransferJob, error) {
	repository.nextID++
	job.ID = "job-" + string(rune('0'+repository.nextID))
	repository.jobs[job.ID] = job
	return job, nil
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
