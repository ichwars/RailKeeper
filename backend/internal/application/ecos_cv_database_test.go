package application

import (
	"context"
	"reflect"
	"testing"
)

type cvDatabaseStub struct {
	fullStoredCVReader
	snapshot ECoSCVDatabase
	reads    int
	failDB   bool
	target   ECoSConnectionInput
}

func (reader *cvDatabaseStub) ReadStoredCVDatabase(_ context.Context, input ECoSConnectionInput, _ int) (ECoSCVDatabase, error) {
	reader.reads++
	reader.target = input
	if reader.failDB {
		return ECoSCVDatabase{}, ErrDigitalCenterDeviceOutput
	}
	return reader.snapshot, nil
}

func TestCVDatabaseWorkspaceRetainsPagesRetriesAndTarget(t *testing.T) {
	high, low, next := 16, 3, 8
	reader := &cvDatabaseStub{snapshot: ECoSCVDatabase{
		Locomotive: ECoSLocomotive{ObjectID: 1002, Name: " Test   Lok ", Address: 3, Protocol: "DCC128", Profile: "LokPilot 5 micro DCC"},
		Values: []ECoSCVValue{{Number: 257, Value: 20, IndexCV31: &high, IndexCV32: &low},
			{Number: 257, Value: 1, IndexCV31: &high, IndexCV32: &next}},
	}}
	reader.probe = &ECoSRawProbe{Locomotives: []ECoSRawLocomotive{{ObjectID: 1002, Name: "Test Lok", Address: 3, Protocol: "DCC28"}}}
	repository := &workspaceRepositoryMemory{}
	settings := &workspaceSettingsReaderStub{value: DigitalCenterSettings{ECoS: DigitalProviderSettings{
		Enabled: true, Host: "ecos.local", Port: "15471",
	}}}
	service := NewDigitalCenterWorkspaceService(repository, settings, reader, nil, &workspaceVehicleReaderStub{}, nil)
	session, err := service.StartReadSession(t.Context(), "ecos", "admin")
	if err != nil {
		t.Fatal(err)
	}
	id := repository.items[0].ID
	read := func(restart bool) (DigitalCenterWorkItem, error) {
		return service.ReadStoredCVBatch(t.Context(), session.ID, id, ECoSCVReadInput{StartCV: 1, Source: "database", Restart: restart})
	}
	item, err := read(false)
	values, state, decodeErr := storedECoSCVSnapshot(item)
	if err != nil || decodeErr != nil || state.Source != "database" || state.NextCV != 1025 ||
		len(values) != 2 || *values[0].IndexCV32 != 3 || *values[1].IndexCV32 != 8 ||
		reader.target.Host != "ecos.local" || item.Center["profile"] != "LokPilot 5 micro DCC" {
		t.Fatalf("snapshot=%+v state=%+v %v %v", values, state, err, decodeErr)
	}
	reader.failDB = true
	if _, err := read(false); err != nil || reader.reads != 1 {
		t.Fatalf("retry repeated device read: %d %v", reader.reads, err)
	}
	if _, err := read(true); err == nil {
		t.Fatal("failed replacement accepted")
	}
	saved, _ := repository.GetWorkItem(t.Context(), session.ID, id)
	unchanged, _, _ := storedECoSCVSnapshot(saved)
	if !reflect.DeepEqual(values, unchanged) {
		t.Fatal("failed export replaced committed values")
	}
	reader.failDB = false
	reader.snapshot.Locomotive.Address = 4
	if _, err := read(true); err != ErrDigitalCenterReadNotFresh {
		t.Fatalf("changed locomotive accepted: %v", err)
	}
	reader.snapshot.Locomotive.Address = 3
	settings.value.ECoS.Host = "another.local"
	if _, err := read(true); err != ErrDigitalCenterReadNotFresh {
		t.Fatalf("changed station accepted: %v", err)
	}
	if _, err := service.ReadStoredCVBatch(t.Context(), session.ID, id,
		ECoSCVReadInput{StartCV: 1}); err == nil {
		t.Fatal("legacy batch appended to indexed database snapshot")
	}
}

func TestStoredCVValidationKeepsPageZeroDistinctAndRejectsBadAddresses(t *testing.T) {
	zero, high, low := 0, 16, 3
	values := []ECoSCVValue{{Number: 257, Value: 1},
		{Number: 257, Value: 2, IndexCV31: &zero, IndexCV32: &zero},
		{Number: 257, Value: 3, IndexCV31: &high, IndexCV32: &low}}
	if parsed, err := validateStoredECoSCVs(values); err != nil || len(parsed) != 3 {
		t.Fatal(parsed, err)
	}
	for _, invalid := range []ECoSCVValue{
		{Number: 1, Value: 0, IndexCV31: &high, IndexCV32: &low},
		{Number: 257, Value: 0, IndexCV31: &high},
		{Number: 257, Value: 3, IndexCV31: &high, IndexCV32: &low},
	} {
		if _, err := validateStoredECoSCVs(append(append([]ECoSCVValue{}, values...), invalid)); err == nil {
			t.Fatalf("invalid or duplicate address accepted: %+v", invalid)
		}
	}
}
