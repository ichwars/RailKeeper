package application

import "testing"

func TestUnverifiedCacheQuarantinesDisagreementButNeverSuppliesMissingCVs(t *testing.T) {
	exported := []ECoSCVValue{{Number: 1, Value: 3}, {Number: 17, Value: 0}, {Number: 29, Value: 8}}
	raw := []ECoSCVValue{{Number: 1, Value: 3}, {Number: 7, Value: 0}, {Number: 8, Value: 151},
		{Number: 17, Value: 209}, {Number: 29, Value: 34}}
	values, info, err := mergeStoredCVSources(exported, nil, true, "checked")
	if err != nil {
		t.Fatal(err)
	}
	values, err = inspectUnverifiedCVCache(values, exported, raw, "available", info)
	if err != nil || len(values) != 1 || values[0].Number != 1 || values[0].Value != 3 || len(info.Conflicts) != 2 {
		t.Fatalf("disputed defaults or raw-only bytes remained importable: %+v %+v %v", values, info, err)
	}
	if len(info.UnverifiedValues) != 1 || info.UnverifiedValues[0].Number != 8 || info.UnverifiedValues[0].Value != 151 {
		t.Fatalf("missing raw evidence was lost, imported or given fabricated zero bytes: %+v", info.UnverifiedValues)
	}
	if info.StoredCacheStatus != "available" || info.Conflicts[0].Number != 17 ||
		info.Conflicts[0].DatabaseValue != 0 || info.Conflicts[0].BackupValue != 209 || info.Conflicts[0].BackupSource != "stored_cache" {
		t.Fatalf("source distinction or contradictory CV17 evidence lost: %+v", info)
	}
	for _, conflict := range info.Conflicts {
		if conflict.Number == 7 {
			t.Fatal("an unmarked raw zero became evidence of a CV reading")
		}
	}
}

func TestUnverifiedCacheKeepsThreeSourcesAndQuarantinesConfigurationOnlyDefaults(t *testing.T) {
	database := []ECoSCVValue{{Number: 17, Value: 0}}
	configuration := []ECoSCVValue{{Number: 17, Value: 5}, {Number: 18, Value: 0}}
	values, info, err := mergeStoredCVSources(database, configuration, true, "checked")
	if err != nil {
		t.Fatal(err)
	}
	values, err = inspectUnverifiedCVCache(values, storedCVReference(database, configuration),
		[]ECoSCVValue{{Number: 17, Value: 209}, {Number: 18, Value: 51}}, "available", info)
	if err != nil || len(values) != 0 || len(info.Conflicts) != 2 || len(info.UnverifiedValues) != 0 {
		t.Fatalf("an unsafe configuration default survived the cache comparison: %+v %+v %v", values, info, err)
	}
	if info.Conflicts[0].DatabaseValue != 0 || info.Conflicts[0].BackupValue != 5 ||
		info.Conflicts[0].RawCacheValue == nil || *info.Conflicts[0].RawCacheValue != 209 {
		t.Fatal("third source overwrote earlier source evidence")
	}
	if countStoredBackupAdditions(values, database) != 0 {
		t.Fatal("quarantined additions were reported as imported backup values")
	}
}

func TestSourceConflictsMarkMatchingRowsForReviewWithoutReplacingIdentityConflicts(t *testing.T) {
	loco := ECoSRawLocomotive{ObjectID: 1002, Name: "Test", Address: 3, Protocol: "DCC",
		CVSources: &ECoSCVSourceInfo{Conflicts: []ECoSCVSourceConflict{{Number: 17, DatabaseValue: 0, BackupValue: 209}}}}
	vehicle := mappedDigitalCenterVehicle("vehicle-1", "ecos", "1002", "3", "DCC")
	vehicle.Name = "Test"
	items := compareDigitalCenterLocomotives("ecos", []ECoSRawLocomotive{loco}, []Vehicle{vehicle})
	if len(items) != 1 || items[0].CompareStatus != DigitalCompareDeviation || items[0].VehicleID != vehicle.ID {
		t.Fatalf("CV source problem looked identical or lost its existing mapping: %+v", items)
	}
	unassigned := compareDigitalCenterLocomotives("ecos", []ECoSRawLocomotive{loco}, nil)
	if len(unassigned) != 1 || unassigned[0].CompareStatus != DigitalCompareNew {
		t.Fatal("source conflict replaced the new-vehicle classification")
	}
	firstCandidate := mappedDigitalCenterVehicle("vehicle-1", "ecos", "old-1", "3", "DCC")
	secondCandidate := mappedDigitalCenterVehicle("vehicle-2", "ecos", "old-2", "3", "DCC")
	ambiguous := compareDigitalCenterLocomotives("ecos", []ECoSRawLocomotive{loco}, []Vehicle{firstCandidate, secondCandidate})
	if ambiguous[0].CompareStatus != DigitalCompareConflict {
		t.Fatal("source conflict hid an existing identity conflict")
	}
}
