package application

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestClassifyDataTransferVehicleSetStructure(t *testing.T) {
	tests := []struct {
		name string
		code string
		edit func(*DataTransferSnapshot)
	}{
		{name: "missing inventory number", code: "missing_vehicle_set_inventory_number", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].InventoryNumber = ""
		}},
		{name: "invalid required metadata", code: "invalid_vehicle_set", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Name = ""
		}},
		{name: "too small", code: "vehicle_set_too_small", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members = snapshot.VehicleSets[0].Members[:1]
		}},
		{name: "duplicate position", code: "duplicate_vehicle_set_position", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members[1].Position = 1
		}},
		{name: "position gap", code: "non_contiguous_vehicle_set_positions", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members[1].Position = 3
		}},
		{name: "member count mismatch", code: "vehicle_set_member_count_mismatch", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members[0].DeclaredMemberCount = 3
			snapshot.VehicleSets[0].Members[0].SourceRowNumber = 2
		}},
		{name: "conflicting CSV metadata", code: "conflicting_vehicle_set_metadata", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members[1].SourceSetInput = snapshot.VehicleSets[0].VehicleSetInput
			snapshot.VehicleSets[0].Members[1].SourceSetInput.Name = "Anderer Name"
			snapshot.VehicleSets[0].Members[1].SourceRowNumber = 3
		}},
		{name: "duplicate member", code: "duplicate_vehicle_set_member", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members[1].SourceVehicleID = "source-a"
			snapshot.VehicleSets[0].Members[1].VehicleInventoryNumber = "RK-A"
		}},
		{name: "missing member reference", code: "missing_vehicle_set_member_reference", edit: func(snapshot *DataTransferSnapshot) {
			snapshot.VehicleSets[0].Members[1].SourceVehicleID = "missing"
			snapshot.VehicleSets[0].Members[1].VehicleInventoryNumber = "RK-MISSING"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			incoming := validTransferVehicleSetSnapshot()
			test.edit(&incoming)
			records, _ := classifyDataTransferImport("job-set", incoming, DataTransferSnapshot{}, nil)
			previews, issues := classifyDataTransferVehicleSets(
				"job-set", incoming, DataTransferSnapshot{}, records,
			)
			if len(previews) != 1 || previews[0].Classification != "error" {
				t.Fatalf("set previews = %#v", previews)
			}
			codes := make([]string, 0, len(issues))
			for _, issue := range issues {
				codes = append(codes, issue.Code)
			}
			if !slices.Contains(codes, test.code) {
				t.Fatalf("issue codes = %#v, want %q", codes, test.code)
			}
		})
	}
}

func TestClassifyDataTransferVehicleSetReadyAndCountsMembersOnly(t *testing.T) {
	incoming := validTransferVehicleSetSnapshot()
	records, vehicleIssues := classifyDataTransferImport("job-set", incoming, DataTransferSnapshot{}, nil)
	previews, setIssues := classifyDataTransferVehicleSets("job-set", incoming, DataTransferSnapshot{}, records)
	if len(vehicleIssues) != 0 || len(setIssues) != 0 || len(previews) != 1 ||
		previews[0].Classification != "ready" || len(previews[0].MemberRecordKeys) != 2 {
		t.Fatalf("valid set classification = records %#v, previews %#v, issues %#v/%#v",
			records, previews, vehicleIssues, setIssues)
	}
	ready, warnings, failures := countDataTransferPreviewRecords(records)
	if ready != 2 || warnings != 0 || failures != 0 {
		t.Fatalf("vehicle record counts = %d/%d/%d", ready, warnings, failures)
	}
}

func TestClassifyDataTransferVehicleSetConflicts(t *testing.T) {
	incoming := validTransferVehicleSetSnapshot()
	current := validTransferVehicleSetSnapshot()
	current.VehicleSets[0].ID = "target-set"
	current.VehicleSets[0].UpdatedAt = "2026-08-27T10:00:00Z"
	current.Vehicles[0].ID = "target-a"
	current.Vehicles[1].ID = "target-b"
	current.VehicleSets[0].Members[0].SourceVehicleID = "target-a"
	current.VehicleSets[0].Members[1].SourceVehicleID = "target-b"
	records, _ := classifyDataTransferImport("job-set", incoming, current, nil)
	previews, issues := classifyDataTransferVehicleSets("job-set", incoming, current, records)
	if len(previews) != 1 || previews[0].TargetID != "target-set" ||
		previews[0].TargetFingerprint == "" || previews[0].ProposedAction != "replace" {
		t.Fatalf("duplicate set preview = %#v", previews)
	}
	assertTransferIssueCode(t, issues, "duplicate_vehicle_set_inventory_number")

	current.VehicleSets = nil
	records, _ = classifyDataTransferImport("job-external", incoming, current, nil)
	previews, issues = classifyDataTransferVehicleSets("job-external", incoming, current, records)
	if len(previews) != 1 || previews[0].ProposedAction != "copy" {
		t.Fatalf("external member conflict preview = %#v", previews)
	}
	assertTransferIssueCode(t, issues, "vehicle_set_member_external_conflict")

	current = validTransferVehicleSetSnapshot()
	current.VehicleSets[0].ID = "target-set"
	current.Vehicles[0].ID = "target-a"
	current.Vehicles[1].ID = "target-b"
	current.VehicleSets[0].Members = current.VehicleSets[0].Members[:1]
	current.VehicleSets[0].Members[0].SourceVehicleID = "target-a"
	records, _ = classifyDataTransferImport("job-target-external", incoming, current, nil)
	previews, issues = classifyDataTransferVehicleSets("job-target-external", incoming, current, records)
	if len(previews) != 1 || previews[0].ProposedAction != "copy" ||
		previews[0].TargetID != "target-set" || previews[0].TargetFingerprint == "" {
		t.Fatalf("target set external conflict preview = %#v", previews)
	}
	assertTransferIssueCode(t, issues, "vehicle_set_member_external_conflict")
}

func TestPreviewImportPersistsVehicleSetGroupsWithoutChangingRecordCounts(t *testing.T) {
	repository, service := newDataTransferImportFixture(t)
	job := fixtureCreateImportJob(t, service, TransferVehicles, TransferJSON)
	incoming := validTransferVehicleSetSnapshot()
	payload, err := json.Marshal(DataTransferPackage{
		Format: DataTransferPackageFormat, Version: DataTransferPackageVersion,
		Areas: DataTransferPackageAreas{Vehicles: incoming.Vehicles, VehicleSets: incoming.VehicleSets},
	})
	if err != nil {
		t.Fatal(err)
	}
	preview, err := service.UploadAndPreview(t.Context(), job.ID, "vehicle-set.json", payload, "editor-1")
	if err != nil {
		t.Fatal(err)
	}
	if preview.TotalRecords != 2 || preview.ReadyRecords != 2 || preview.WarningRecords != 0 ||
		preview.ErrorRecords != 0 || len(preview.VehicleSets) != 1 ||
		preview.VehicleSets[0].Classification != "ready" {
		t.Fatalf("vehicle set preview = %#v", preview)
	}
	persisted, err := repository.GetJob(t.Context(), job.ID)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(persisted.Preview)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Preview["vehicleSets"] == nil || !json.Valid(encoded) {
		t.Fatalf("persisted vehicle set preview = %s", encoded)
	}
}

func TestVehicleSetConflictSuppressesMemberLevelDuplicateIssues(t *testing.T) {
	incoming := validTransferVehicleSetSnapshot()
	current := validTransferVehicleSetSnapshot()
	current.VehicleSets[0].ID = "target-set"
	current.Vehicles[0].ID = "target-a"
	current.Vehicles[1].ID = "target-b"
	current.VehicleSets[0].Members[0].SourceVehicleID = "target-a"
	current.VehicleSets[0].Members[1].SourceVehicleID = "target-b"
	records, issues := classifyDataTransferImport("job-set", incoming, current, nil)
	sets, setIssues := classifyDataTransferVehicleSets("job-set", incoming, current, records)
	issues = append(issues, setIssues...)
	records, issues = suppressDataTransferVehicleSetMemberConflicts(records, issues, sets)
	for _, issue := range issues {
		if issue.Code == "duplicate_inventory_number" {
			t.Fatalf("member conflict was not suppressed: %#v", issues)
		}
	}
	assertTransferIssueCode(t, issues, "duplicate_vehicle_set_inventory_number")
	if records[0].Classification != "ready" || records[1].Classification != "ready" {
		t.Fatalf("member classifications = %#v", records)
	}
}

func validTransferVehicleSetSnapshot() DataTransferSnapshot {
	setInput := VehicleSetInput{
		Name: "Rheingold", Manufacturer: "Roco", Gauge: "H0", Category: "Set", Gattung: "Reisezug",
	}
	return DataTransferSnapshot{
		Vehicles: []TransferVehicle{
			{ID: "source-a", InventoryNumber: "RK-A", Manufacturer: "Roco", Name: "Wagen A", Gauge: "H0",
				Category: "Wagen", Gattung: "Reisezugwagen"},
			{ID: "source-b", InventoryNumber: "RK-B", Manufacturer: "Roco", Name: "Wagen B", Gauge: "H0",
				Category: "Wagen", Gattung: "Reisezugwagen"},
		},
		VehicleSets: []TransferVehicleSet{{
			ID: "source-set", InventoryNumber: "Set-001", VehicleSetInput: setInput,
			Members: []TransferVehicleSetMember{
				{SourceVehicleID: "source-a", VehicleInventoryNumber: "RK-A", Position: 1,
					DeclaredMemberCount: 2, SourceSetInput: setInput},
				{SourceVehicleID: "source-b", VehicleInventoryNumber: "RK-B", Position: 2,
					DeclaredMemberCount: 2, SourceSetInput: setInput},
			},
		}},
	}
}
