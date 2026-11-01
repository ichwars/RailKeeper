package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func backupCVWorkItem() DigitalCenterWorkItem {
	zero, high, low := 0, 16, 3
	return DigitalCenterWorkItem{
		ID: "backup-item", SessionID: "backup-session", CenterObjectID: "1002",
		VehicleID: "vehicle-1", Name: "Test", Address: 3, Protocol: "DCC", StationStatus: "read",
		Center: map[string]any{
			"cvs": []ECoSCVValue{
				{Number: 1, Value: 3}, {Number: 257, Value: 12},
				{Number: 257, Value: 13, IndexCV31: &zero, IndexCV32: &zero},
				{Number: 257, Value: 14, IndexCV31: &high, IndexCV32: &low},
			},
			"cvRead": ECoSCVReadState{Source: "backup", NextCV: 1025,
				ReadAt: "2026-10-21T15:00:00+02:00", SourceHost: "ecos.local"},
		},
	}
}

func TestStoredBackupCVSnapshotKeepsPageAddressesAndRequiresCompleteProvenance(t *testing.T) {
	item := backupCVWorkItem()
	values, state, err := storedECoSCVSnapshot(item)
	if err != nil || len(values) != 4 || state.Source != "backup" || state.NextCV != 1025 ||
		state.ReadAt != "2026-10-21T15:00:00+02:00" || state.SourceHost != "ecos.local" {
		t.Fatalf("backup lost data or provenance: %+v %+v %v", values, state, err)
	}
	if values[1].IndexCV31 != nil || *values[2].IndexCV31 != 0 || *values[2].IndexCV32 != 0 ||
		*values[3].IndexCV31 != 16 || *values[3].IndexCV32 != 3 {
		t.Fatalf("backup lost unknown, zero or indexed pages: %+v", values)
	}
	for name, change := range map[string]func(*ECoSCVReadState){
		"no cursor":    func(s *ECoSCVReadState) { s.NextCV = 0 },
		"partial":      func(s *ECoSCVReadState) { s.NextCV = 33 },
		"unavailable":  func(s *ECoSCVReadState) { s.Unavailable = 1 },
		"missing host": func(s *ECoSCVReadState) { s.SourceHost = "" },
		"blank host":   func(s *ECoSCVReadState) { s.SourceHost = "  " },
		"invalid time": func(s *ECoSCVReadState) { s.ReadAt = "not-a-date" },
		"unknown":      func(s *ECoSCVReadState) { s.Source = "unknown" },
	} {
		t.Run(name, func(t *testing.T) {
			invalid := backupCVWorkItem()
			read := invalid.Center["cvRead"].(ECoSCVReadState)
			change(&read)
			invalid.Center["cvRead"] = read
			if _, _, err := storedECoSCVSnapshot(invalid); !errors.Is(err, ErrDigitalCenterDeviceOutput) {
				t.Fatalf("invalid backup accepted: %+v %v", read, err)
			}
		})
	}
	invalid := backupCVWorkItem()
	invalid.Center["cvs"] = []ECoSCVValue{{Number: 1, Value: 256}}
	if _, _, err := storedECoSCVSnapshot(invalid); !errors.Is(err, ErrDigitalCenterDeviceOutput) {
		t.Fatalf("invalid backup CV accepted: %v", err)
	}
}

type backupCVImporter struct {
	workspaceVehicleReaderStub
	importCalls int
	vehicleID   string
	source      ECoSStoredCVSource
}

func (importer *backupCVImporter) ImportECoSStoredCVs(
	_ context.Context, vehicleID string, source ECoSStoredCVSource,
) (ECoSCVImportResult, error) {
	importer.importCalls++
	importer.vehicleID, importer.source = vehicleID, source
	return ECoSCVImportResult{Imported: 1, Preserved: 3}, nil
}

func backupCVWorkspace() (*DigitalCenterWorkspaceService, *workspaceRepositoryMemory, *cvDatabaseStub, *backupCVImporter) {
	repository := &workspaceRepositoryMemory{
		session: DigitalCenterReadSession{ID: "backup-session", Provider: "ecos", State: DigitalCenterSessionReady,
			Host: "ecos.local", Port: 15471, ReadCompletedAt: "2026-10-29T07:00:00Z"},
		items: []DigitalCenterWorkItem{backupCVWorkItem()},
	}
	reader := &cvDatabaseStub{}
	importer := &backupCVImporter{}
	// Backup inspection/import must not require configured settings or a reachable device.
	service := NewDigitalCenterWorkspaceService(repository, nil, reader, nil, importer, nil)
	return service, repository, reader, importer
}

func TestStoredBackupCVSnapshotNeverStartsDeviceReadOrRestart(t *testing.T) {
	for _, source := range []string{"", "database"} {
		for _, restart := range []bool{false, true} {
			service, repository, reader, _ := backupCVWorkspace()
			before := backupCVWorkItem()
			_, err := service.ReadStoredCVBatch(t.Context(), "backup-session", "backup-item",
				ECoSCVReadInput{Source: source, StartCV: 1, Restart: restart})
			if !errors.Is(err, ErrDigitalCenterCapabilityUnavailable) || reader.calls != 0 || reader.reads != 0 {
				t.Fatalf("source=%q restart=%v contacted device: %d %d %v", source, restart, reader.calls, reader.reads, err)
			}
			if !reflect.DeepEqual(before, repository.items[0]) {
				t.Fatal("device read attempt changed backup snapshot")
			}
		}
	}
}

func TestStoredBackupCVSnapshotImportsWithoutRefreshUsingBackupTime(t *testing.T) {
	service, repository, reader, importer := backupCVWorkspace()
	before := backupCVWorkItem()
	result, err := service.ImportStoredCVs(t.Context(), "backup-session", "backup-item", "vehicle-1")
	if err != nil || result.Imported != 1 || result.Preserved != 3 || importer.importCalls != 1 ||
		reader.calls != 0 || reader.reads != 0 || importer.vehicleID != "vehicle-1" ||
		importer.source.ReadAt != "2026-10-21T15:00:00+02:00" || importer.source.Host != "ecos.local" ||
		importer.source.ObjectID != "1002" || importer.source.Protocol != "DCC" {
		t.Fatalf("backup import lost provenance or started refresh: %+v %+v %v", result, importer.source, err)
	}
	expected, _, err := storedECoSCVSnapshot(before)
	if err != nil || !reflect.DeepEqual(importer.source.Values, expected) || !reflect.DeepEqual(repository.items[0], before) {
		t.Fatal("backup import changed snapshot or lost indexed values")
	}
}

func TestStoredBackupCVSnapshotImportRejectsMismatchedStation(t *testing.T) {
	service, repository, reader, importer := backupCVWorkspace()
	state := repository.items[0].Center["cvRead"].(ECoSCVReadState)
	state.SourceHost = "another.local"
	repository.items[0].Center["cvRead"] = state
	_, err := service.ImportStoredCVs(t.Context(), "backup-session", "backup-item", "vehicle-1")
	if !errors.Is(err, ErrDigitalCenterDeviceOutput) || importer.importCalls != 0 || reader.calls != 0 || reader.reads != 0 {
		t.Fatalf("backup from another station accepted: %v", err)
	}
}

func TestStoredBackupCVSnapshotImportRequiresConfirmedVehicleLink(t *testing.T) {
	for _, linked := range []string{"", "another-vehicle"} {
		service, repository, reader, importer := backupCVWorkspace()
		repository.items[0].VehicleID = linked
		_, err := service.ImportStoredCVs(t.Context(), "backup-session", "backup-item", "vehicle-1")
		if !errors.Is(err, ErrDigitalCenterConflictUnresolved) || importer.importCalls != 0 || reader.calls != 0 || reader.reads != 0 {
			t.Fatal("backup without confirmed vehicle link was imported", err)
		}
	}
}
