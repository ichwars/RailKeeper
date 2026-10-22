package application

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

type fullStoredCVReader struct{ workspaceECoSReaderStub }

func (reader *fullStoredCVReader) ReadStoredCVs(_ context.Context, _ ECoSConnectionInput, _ int, start int) (ECoSStoredCVBatch, error) {
	batch := ECoSStoredCVBatch{}
	for number := start; number < start+32; number++ {
		batch.Values = append(batch.Values, ECoSCVValue{Number: number, Value: number % 256})
	}
	return batch, nil
}

func TestECoSCVWorkspacePreservesAllValuesAndSequentialProgress(t *testing.T) {
	reader := &fullStoredCVReader{workspaceECoSReaderStub{probe: &ECoSRawProbe{Locomotives: []ECoSRawLocomotive{{
		ObjectID: 1002, Name: "Test", Address: 3, Protocol: "DCC28", CVs: []ECoSCVValue{{Number: 8, Value: 151}},
	}}}}}
	repository := &workspaceRepositoryMemory{}
	s := NewDigitalCenterWorkspaceService(repository, &workspaceSettingsReaderStub{value: DigitalCenterSettings{
		ECoS: DigitalProviderSettings{Enabled: true, Host: "ecos.local", Port: "15471"},
	}}, reader, nil, &workspaceVehicleReaderStub{}, nil)
	session, err := s.StartReadSession(t.Context(), "ecos", "admin")
	if err != nil {
		t.Fatal(err)
	}
	item := repository.items[0]
	values, _, err := storedECoSCVSnapshot(item)
	if err != nil || len(values) != 1 || values[0].Value != 151 {
		t.Fatalf("initial CV was dropped: %+v %v", values, err)
	}
	if _, err := s.ReadStoredCVBatch(t.Context(), session.ID, item.ID, 33); err == nil {
		t.Fatal("out-of-order batch accepted")
	}
	for start := 1; start <= 1024; start += 32 {
		item, err = s.ReadStoredCVBatch(t.Context(), session.ID, item.ID, start)
		if err != nil {
			t.Fatal(err)
		}
	}
	// The client lost a committed response: retrying the previous cursor returns
	// the persisted result instead of rejecting or appending duplicate values.
	item, err = s.ReadStoredCVBatch(t.Context(), session.ID, item.ID, 993)
	if err != nil {
		t.Fatal(err)
	}
	values, state, err := storedECoSCVSnapshot(item)
	if err != nil || len(values) != 1024 || state.NextCV != 1025 || state.SourceHost != "ecos.local" {
		t.Fatalf("incomplete snapshot: %d %+v %v", len(values), state, err)
	}
	for i, value := range values {
		if value.Number != i+1 || value.Value != (i+1)%256 {
			t.Fatal("CV lost: " + strconv.Itoa(i+1))
		}
	}
	if !strings.HasSuffix(state.ReadAt, "Z") {
		t.Fatal("missing UTC provenance")
	}
}
