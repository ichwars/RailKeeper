package infrastructure_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferApplyCreatesVehicleSetAtomically(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-IMPORT", "RK-SET-A", "RK-SET-B")
	job := createApplyJobWithVehicleSets(t, repository, "sha-set-create", records, []application.DataTransferVehicleSetPreview{setPreview})

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var setCount, vehicleCount, memberCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets WHERE inventory_number='RK-SET-IMPORT'`).Scan(&setCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles WHERE inventory_number IN ('RK-SET-A', 'RK-SET-B')`).Scan(&vehicleCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
SELECT COUNT(*) FROM vehicle_set_members member
JOIN vehicle_sets vehicle_set ON vehicle_set.id=member.vehicle_set_id
WHERE vehicle_set.inventory_number='RK-SET-IMPORT' AND member.position IN (1, 2)`).Scan(&memberCount); err != nil {
		t.Fatal(err)
	}
	if setCount != 1 || vehicleCount != 2 || memberCount != 2 {
		t.Fatalf("created set/vehicles/members = %d/%d/%d", setCount, vehicleCount, memberCount)
	}
}

func TestDataTransferApplyRollsBackIncompleteVehicleSet(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-ROLLBACK", "RK-ROLLBACK-A", "RK-ROLLBACK-B")
	setPreview.MemberRecordKeys[1] = "missing-record"
	job := createApplyJobWithVehicleSets(t, repository, "sha-set-rollback", records, []application.DataTransferVehicleSetPreview{setPreview})

	err := repository.ApplyImport(t.Context(), job, "editor-1")
	if !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() error = %v, want conflict", err)
	}
	for _, table := range []string{"vehicles", "vehicle_sets", "vehicle_set_members"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s count after rollback = %d", table, count)
		}
	}
	assertApplyJobStillReady(t, repository, job.ID)
}

func TestDataTransferApplyReplacesVehicleSetAndDetachesMissingMembers(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	insertApplyVehicleSet(t, db, "target-set", "RK-SET-REPLACE", []applyVehicleSetMemberFixture{
		{ID: "target-a", InventoryNumber: "RK-REPLACE-A", Position: 1},
		{ID: "target-old", InventoryNumber: "RK-REPLACE-OLD", Position: 2},
	})
	targetSet, targetVehicles := applyVehicleSetSnapshot(t, repository, "target-set")
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-REPLACE", "RK-REPLACE-A", "RK-REPLACE-B")
	records[0].ProposedAction = "replace"
	records[0].TargetID = "target-a"
	records[0].TargetFingerprint = application.DataTransferTargetFingerprint(targetVehicles["RK-REPLACE-A"])
	setPreview.Classification = "warning"
	setPreview.ProposedAction = "replace"
	setPreview.TargetID = targetSet.ID
	setPreview.TargetFingerprint = application.DataTransferTargetFingerprint(targetSet)
	setPreview.Data.Name = "Rheingold aktualisiert"
	job := createApplyJobWithVehicleSets(t, repository, "sha-set-replace", records, []application.DataTransferVehicleSetPreview{setPreview})
	resolveApplyVehicleSetIssue(t, repository, job.ID, setPreview, "duplicate_vehicle_set_inventory_number", "replace")

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := db.QueryRow(`SELECT name FROM vehicle_sets WHERE id='target-set'`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	var importedMembers, detachedVehicle int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_set_members WHERE vehicle_set_id='target-set'`).Scan(&importedMembers); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles WHERE id='target-old'`).Scan(&detachedVehicle); err != nil {
		t.Fatal(err)
	}
	if name != "Rheingold aktualisiert" || importedMembers != 2 || detachedVehicle != 1 {
		t.Fatalf("replace result name/members/detached = %q/%d/%d", name, importedMembers, detachedVehicle)
	}
}

func TestDataTransferApplyCSVReplacePreservesUnmappedVehicleSetFields(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	insertApplyVehicleSet(t, db, "target-set", "RK-SET-CSV", []applyVehicleSetMemberFixture{
		{ID: "target-a", InventoryNumber: "RK-CSV-A", Position: 1, Label: "Lokale Bezeichnung"},
		{ID: "target-old", InventoryNumber: "RK-CSV-OLD", Position: 2},
	})
	if _, err := db.Exec(`
UPDATE vehicle_sets
SET description='Lokale Beschreibung', storage_location='Vitrine 7'
WHERE id='target-set'`); err != nil {
		t.Fatal(err)
	}
	targetSet, targetVehicles := applyVehicleSetSnapshot(t, repository, "target-set")
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-CSV", "RK-CSV-A", "RK-CSV-B")
	records[0].ProposedAction = "replace"
	records[0].TargetID = "target-a"
	records[0].TargetFingerprint = application.DataTransferTargetFingerprint(targetVehicles["RK-CSV-A"])
	setPreview.Classification = "warning"
	setPreview.ProposedAction = "replace"
	setPreview.TargetID = targetSet.ID
	setPreview.TargetFingerprint = application.DataTransferTargetFingerprint(targetSet)
	setPreview.Data.Name = "CSV-Name"
	setPreview.Data.Members[0].Label = ""
	setPreview.Data.Members[1].Label = ""
	job := createApplyJobWithVehicleSets(
		t, repository, "sha-set-csv-preserve", records, []application.DataTransferVehicleSetPreview{setPreview},
	)
	job.Preview["csvMapping"] = []application.DataTransferCSVColumnMapping{
		{TargetField: "inventoryNumber"},
		{TargetField: "manufacturer"},
		{TargetField: "name"},
		{TargetField: "gauge"},
		{TargetField: "category"},
		{TargetField: "gattung"},
		{TargetField: "vehicleSetInventoryNumber"},
		{TargetField: "vehicleSetName"},
		{TargetField: "vehicleSetManufacturer"},
		{TargetField: "vehicleSetGauge"},
		{TargetField: "vehicleSetCategory"},
		{TargetField: "vehicleSetGattung"},
	}
	var err error
	job, err = repository.UpdateJob(t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}
	resolveApplyVehicleSetIssue(
		t, repository, job.ID, setPreview, "duplicate_vehicle_set_inventory_number", "replace",
	)

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var name, description, storageLocation string
	if err := db.QueryRow(`
SELECT name, description, storage_location FROM vehicle_sets WHERE id='target-set'`).Scan(
		&name, &description, &storageLocation,
	); err != nil {
		t.Fatal(err)
	}
	if name != "CSV-Name" || description != "Lokale Beschreibung" || storageLocation != "Vitrine 7" {
		t.Fatalf("CSV replace set fields = %q/%q/%q", name, description, storageLocation)
	}
	var memberLabel string
	if err := db.QueryRow(`
SELECT label FROM vehicle_set_members
WHERE vehicle_set_id='target-set' AND vehicle_id='target-a'`).Scan(&memberLabel); err != nil {
		t.Fatal(err)
	}
	if memberLabel != "Lokale Bezeichnung" {
		t.Fatalf("CSV replace member label = %q, want preserved local label", memberLabel)
	}
}

func TestDataTransferApplyCSVReplaceClearsMappedVehicleSetMemberLabel(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	insertApplyVehicleSet(t, db, "target-set", "RK-SET-CSV-LABEL", []applyVehicleSetMemberFixture{
		{ID: "target-a", InventoryNumber: "RK-CSV-LABEL-A", Position: 1, Label: "Lokale Bezeichnung"},
		{ID: "target-old", InventoryNumber: "RK-CSV-LABEL-OLD", Position: 2},
	})
	targetSet, targetVehicles := applyVehicleSetSnapshot(t, repository, "target-set")
	records, setPreview := applyVehicleSetPreview(
		t, "RK-SET-CSV-LABEL", "RK-CSV-LABEL-A", "RK-CSV-LABEL-B",
	)
	records[0].ProposedAction = "replace"
	records[0].TargetID = "target-a"
	records[0].TargetFingerprint = application.DataTransferTargetFingerprint(targetVehicles["RK-CSV-LABEL-A"])
	setPreview.Classification = "warning"
	setPreview.ProposedAction = "replace"
	setPreview.TargetID = targetSet.ID
	setPreview.TargetFingerprint = application.DataTransferTargetFingerprint(targetSet)
	setPreview.Data.Members[0].Label = ""
	setPreview.Data.Members[1].Label = ""
	job := createApplyJobWithVehicleSets(
		t, repository, "sha-set-csv-clear-label", records, []application.DataTransferVehicleSetPreview{setPreview},
	)
	job.Preview["csvMapping"] = []application.DataTransferCSVColumnMapping{{TargetField: "vehicleSetMemberLabel"}}
	var err error
	job, err = repository.UpdateJob(t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}
	resolveApplyVehicleSetIssue(
		t, repository, job.ID, setPreview, "duplicate_vehicle_set_inventory_number", "replace",
	)

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var memberLabel string
	if err := db.QueryRow(`
SELECT label FROM vehicle_set_members
WHERE vehicle_set_id='target-set' AND vehicle_id='target-a'`).Scan(&memberLabel); err != nil {
		t.Fatal(err)
	}
	if memberLabel != "" {
		t.Fatalf("CSV replace mapped blank member label = %q, want cleared label", memberLabel)
	}
}

func TestDataTransferApplyCopiesVehicleSetWithoutReusingMembers(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	insertApplyVehicleSet(t, db, "target-set", "RK-SET-COPY", []applyVehicleSetMemberFixture{
		{ID: "target-a", InventoryNumber: "RK-COPY-A", Position: 1},
		{ID: "target-b", InventoryNumber: "RK-COPY-B", Position: 2},
	})
	targetSet, targetVehicles := applyVehicleSetSnapshot(t, repository, "target-set")
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-COPY", "RK-COPY-A", "RK-COPY-B")
	for index, inventoryNumber := range []string{"RK-COPY-A", "RK-COPY-B"} {
		records[index].ProposedAction = "replace"
		records[index].TargetID = targetVehicles[inventoryNumber].ID
		records[index].TargetFingerprint = application.DataTransferTargetFingerprint(targetVehicles[inventoryNumber])
	}
	setPreview.Classification = "warning"
	setPreview.ProposedAction = "replace"
	setPreview.TargetID = targetSet.ID
	setPreview.TargetFingerprint = application.DataTransferTargetFingerprint(targetSet)
	job := createApplyJobWithVehicleSets(t, repository, "sha-set-copy", records, []application.DataTransferVehicleSetPreview{setPreview})
	resolveApplyVehicleSetIssue(t, repository, job.ID, setPreview, "duplicate_vehicle_set_inventory_number", "copy")

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var sets, vehicles, originalMembers, copiedMembers int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets`).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_set_members WHERE vehicle_set_id='target-set'`).Scan(&originalMembers); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`
SELECT COUNT(*) FROM vehicle_set_members member
JOIN vehicle_sets vehicle_set ON vehicle_set.id=member.vehicle_set_id
WHERE vehicle_set.inventory_number='RK-SET-COPY-COPY'`).Scan(&copiedMembers); err != nil {
		t.Fatal(err)
	}
	if sets != 2 || vehicles != 4 || originalMembers != 2 || copiedMembers != 2 {
		t.Fatalf("copy result sets/vehicles/original/copied = %d/%d/%d/%d", sets, vehicles, originalMembers, copiedMembers)
	}
}

func TestDataTransferApplySkipsVehicleSetAsGroup(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	insertApplyVehicleSet(t, db, "target-set", "RK-SET-SKIP", []applyVehicleSetMemberFixture{
		{ID: "target-a", InventoryNumber: "RK-SKIP-A", Position: 1},
		{ID: "target-b", InventoryNumber: "RK-SKIP-B", Position: 2},
	})
	targetSet, targetVehicles := applyVehicleSetSnapshot(t, repository, "target-set")
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-SKIP", "RK-SKIP-A", "RK-SKIP-B")
	for index, inventoryNumber := range []string{"RK-SKIP-A", "RK-SKIP-B"} {
		records[index].ProposedAction = "replace"
		records[index].TargetID = targetVehicles[inventoryNumber].ID
		records[index].TargetFingerprint = application.DataTransferTargetFingerprint(targetVehicles[inventoryNumber])
	}
	setPreview.Classification = "warning"
	setPreview.ProposedAction = "replace"
	setPreview.TargetID = targetSet.ID
	setPreview.TargetFingerprint = application.DataTransferTargetFingerprint(targetSet)
	setPreview.Data.Name = "Darf nicht gespeichert werden"
	job := createApplyJobWithVehicleSets(t, repository, "sha-set-skip", records, []application.DataTransferVehicleSetPreview{setPreview})
	resolveApplyVehicleSetIssue(t, repository, job.ID, setPreview, "duplicate_vehicle_set_inventory_number", "skip")

	if err := repository.ApplyImport(t.Context(), job, "editor-1"); err != nil {
		t.Fatal(err)
	}
	var name string
	if err := db.QueryRow(`SELECT name FROM vehicle_sets WHERE id='target-set'`).Scan(&name); err != nil {
		t.Fatal(err)
	}
	var sets, vehicles int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets`).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if name != "Bestehendes Set" || sets != 1 || vehicles != 2 {
		t.Fatalf("skip result name/sets/vehicles = %q/%d/%d", name, sets, vehicles)
	}
}

func TestDataTransferApplyRejectsUnsafeExternalVehicleSetConflict(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDataTransferRepository(db)
	insertApplyVehicleSet(t, db, "target-set", "RK-SET-EXTERNAL", []applyVehicleSetMemberFixture{
		{ID: "target-a", InventoryNumber: "RK-EXTERNAL-A", Position: 1},
		{ID: "target-old", InventoryNumber: "RK-EXTERNAL-OLD", Position: 2},
	})
	insertApplyVehicle(t, db, "outside-b", "RK-EXTERNAL-B")
	targetSet, targetVehicles := applyVehicleSetSnapshot(t, repository, "target-set")
	outside := applyVehicleSnapshot(t, repository, "RK-EXTERNAL-B")
	records, setPreview := applyVehicleSetPreview(t, "RK-SET-EXTERNAL", "RK-EXTERNAL-A", "RK-EXTERNAL-B")
	records[0].ProposedAction = "replace"
	records[0].TargetID = targetVehicles["RK-EXTERNAL-A"].ID
	records[0].TargetFingerprint = application.DataTransferTargetFingerprint(targetVehicles["RK-EXTERNAL-A"])
	records[1].ProposedAction = "replace"
	records[1].TargetID = outside.ID
	records[1].TargetFingerprint = application.DataTransferTargetFingerprint(outside)
	setPreview.Classification = "warning"
	setPreview.ProposedAction = "copy"
	setPreview.TargetID = targetSet.ID
	setPreview.TargetFingerprint = application.DataTransferTargetFingerprint(targetSet)
	job := createApplyJobWithVehicleSets(t, repository, "sha-set-external", records, []application.DataTransferVehicleSetPreview{setPreview})
	resolveApplyVehicleSetIssue(t, repository, job.ID, setPreview, "vehicle_set_member_external_conflict", "replace")

	err := repository.ApplyImport(t.Context(), job, "editor-1")
	if !errors.Is(err, application.ErrDataTransferConflict) {
		t.Fatalf("ApplyImport() error = %v, want conflict", err)
	}
	var sets, vehicles int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicle_sets`).Scan(&sets); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles`).Scan(&vehicles); err != nil {
		t.Fatal(err)
	}
	if sets != 1 || vehicles != 3 {
		t.Fatalf("external conflict changed sets/vehicles = %d/%d", sets, vehicles)
	}
	assertApplyJobStillReady(t, repository, job.ID)
}

func applyVehicleSetPreview(
	t *testing.T,
	setInventoryNumber string,
	firstInventoryNumber string,
	secondInventoryNumber string,
) ([]application.DataTransferPreviewRecord, application.DataTransferVehicleSetPreview) {
	t.Helper()
	records := []application.DataTransferPreviewRecord{
		applyVehicleRecordWithID(t, "source-a", firstInventoryNumber),
		applyVehicleRecordWithID(t, "source-b", secondInventoryNumber),
	}
	set := application.TransferVehicleSet{
		ID: "source-set", InventoryNumber: setInventoryNumber,
		VehicleSetInput: application.VehicleSetInput{
			Name: "Rheingold", Manufacturer: "Roco", Gauge: "H0", Category: "Set", Gattung: "Reisezug",
		},
		Members: []application.TransferVehicleSetMember{
			{SourceVehicleID: "source-a", VehicleInventoryNumber: firstInventoryNumber, Position: 1, Label: "A"},
			{SourceVehicleID: "source-b", VehicleInventoryNumber: secondInventoryNumber, Position: 2, Label: "B"},
		},
	}
	return records, application.DataTransferVehicleSetPreview{
		RecordKey: "vehicle-set:id:source-set", Classification: "ready", ProposedAction: "create",
		MemberRecordKeys: []string{firstInventoryNumber, secondInventoryNumber}, Data: set,
	}
}

func applyVehicleRecordWithID(
	t *testing.T,
	sourceID string,
	inventoryNumber string,
) application.DataTransferPreviewRecord {
	t.Helper()
	data, err := json.Marshal(application.TransferVehicle{
		ID: sourceID, InventoryNumber: inventoryNumber, Manufacturer: "Roco", Name: inventoryNumber,
		Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen",
	})
	if err != nil {
		t.Fatal(err)
	}
	return application.DataTransferPreviewRecord{
		Area: application.TransferVehicles, RecordKey: inventoryNumber, Classification: "ready",
		ProposedAction: "create", Data: data,
	}
}

func createApplyJobWithVehicleSets(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	sourceSHA string,
	records []application.DataTransferPreviewRecord,
	sets []application.DataTransferVehicleSetPreview,
) application.DataTransferJob {
	t.Helper()
	job := createApplyJob(t, repository, sourceSHA, records)
	job.Preview["vehicleSets"] = sets
	updated, err := repository.UpdateJob(t.Context(), job)
	if err != nil {
		t.Fatal(err)
	}
	return updated
}

type applyVehicleSetMemberFixture struct {
	ID              string
	InventoryNumber string
	Position        int
	Label           string
}

func insertApplyVehicleSet(
	t *testing.T,
	db *sql.DB,
	setID string,
	inventoryNumber string,
	members []applyVehicleSetMemberFixture,
) {
	t.Helper()
	for _, member := range members {
		insertApplyVehicle(t, db, member.ID, member.InventoryNumber)
	}
	if _, err := db.Exec(`
INSERT INTO vehicle_sets(
  id, inventory_number, name, manufacturer, gauge, category, gattung, created_at, updated_at
) VALUES(?, ?, 'Bestehendes Set', 'Roco', 'H0', 'Set', 'Reisezug',
  '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`, setID, inventoryNumber); err != nil {
		t.Fatal(err)
	}
	for _, member := range members {
		if _, err := db.Exec(`
INSERT INTO vehicle_set_members(vehicle_set_id, vehicle_id, position, label)
VALUES(?, ?, ?, ?)`, setID, member.ID, member.Position, member.Label); err != nil {
			t.Fatal(err)
		}
	}
}

func insertApplyVehicle(t *testing.T, db *sql.DB, id string, inventoryNumber string) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, category, gattung, created_at, updated_at
) VALUES(?, ?, 'Roco', ?, 'H0', 'Wagen', 'Reisezugwagen',
  '2026-08-27T10:00:00Z', '2026-08-27T10:00:00Z')`, id, inventoryNumber, inventoryNumber); err != nil {
		t.Fatal(err)
	}
}

func applyVehicleSetSnapshot(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	setID string,
) (application.TransferVehicleSet, map[string]application.TransferVehicle) {
	t.Helper()
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	vehicles := make(map[string]application.TransferVehicle, len(snapshot.Vehicles))
	for _, vehicle := range snapshot.Vehicles {
		vehicles[vehicle.InventoryNumber] = vehicle
	}
	for _, set := range snapshot.VehicleSets {
		if set.ID == setID {
			return set, vehicles
		}
	}
	t.Fatalf("vehicle set %q not found", setID)
	return application.TransferVehicleSet{}, nil
}

func applyVehicleSnapshot(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	inventoryNumber string,
) application.TransferVehicle {
	t.Helper()
	snapshot, err := repository.Snapshot(t.Context(), []application.TransferArea{application.TransferVehicles})
	if err != nil {
		t.Fatal(err)
	}
	for _, vehicle := range snapshot.Vehicles {
		if vehicle.InventoryNumber == inventoryNumber {
			return vehicle
		}
	}
	t.Fatalf("vehicle %q not found", inventoryNumber)
	return application.TransferVehicle{}
}

func resolveApplyVehicleSetIssue(
	t *testing.T,
	repository *infrastructure.DataTransferRepository,
	jobID string,
	preview application.DataTransferVehicleSetPreview,
	code string,
	resolution string,
) {
	t.Helper()
	if err := repository.ReplaceIssues(t.Context(), jobID, []application.DataTransferIssue{{
		JobID: jobID, Area: application.TransferVehicles, RecordKey: preview.RecordKey,
		Severity: application.TransferIssueWarning, Code: code, SelectedResolution: resolution,
	}}); err != nil {
		t.Fatal(err)
	}
}
