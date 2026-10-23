package application

import (
	"errors"
	"strings"
	"testing"
)

func TestECoSWorkspaceReadsLocomotivesAlongsideConsists(t *testing.T) {
	repository := &workspaceRepositoryMemory{}
	reader := &workspaceECoSReaderStub{probe: &ECoSRawProbe{Locomotives: []ECoSRawLocomotive{
		{ObjectID: 1001, Name: "Testlok", Address: 3, Protocol: "DCC28", CVs: []ECoSCVValue{{Number: 8, Value: 151}}},
		{ObjectID: 1002, Name: "Testverbund", Protocol: "MULTI"},
		{ObjectID: 1003, Name: "Zweite Lok", Address: 44, Protocol: "DCC128"},
	}}}
	service := NewDigitalCenterWorkspaceService(repository,
		&workspaceSettingsReaderStub{value: configuredECoSSettings()}, reader, nil,
		&workspaceVehicleReaderStub{}, nil)
	session, err := service.StartReadSession(t.Context(), "ecos", "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	if session.State != DigitalCenterSessionReady || len(repository.items) != 2 {
		t.Fatalf("session=%#v items=%#v, want both individually addressed locomotives", session, repository.items)
	}
	if repository.items[0].CenterObjectID != "1001" || repository.items[1].CenterObjectID != "1003" {
		t.Fatalf("worklist includes wrong objects: %#v", repository.items)
	}
	cvs := repository.items[0].Center["cvs"].([]ECoSCVValue)
	if len(cvs) != 1 || cvs[0].Number != 8 || cvs[0].Value != 151 {
		t.Fatalf("CVs lost: %#v", cvs)
	}
	if len(repository.messages) != 1 || repository.messages[0].Code != DigitalCenterMessageReadCompleted ||
		!strings.Contains(repository.messages[0].Message, "1 ECoS-Mehrfachtraktionen mit Protokoll MULTI") ||
		!strings.Contains(repository.messages[0].Message, "1 ECoS consists with protocol MULTI") {
		t.Fatalf("missing consist exclusion notice: %#v", repository.messages)
	}
}

func TestNormalizeDigitalCenterRosterOnlySkipsKnownECoSConsists(t *testing.T) {
	valid := ECoSRawLocomotive{ObjectID: 1001, Address: 3, Protocol: "DCC28"}
	consist := ECoSRawLocomotive{ObjectID: 1002, Protocol: "MULTI"}
	tests := []struct {
		name        string
		provider    string
		locomotives []ECoSRawLocomotive
		wantError   bool
		wantSkipped int
	}{
		{name: "only consists", provider: "ecos", locomotives: []ECoSRawLocomotive{consist}, wantSkipped: 1},
		{name: "ordinary address missing", provider: "ecos", locomotives: []ECoSRawLocomotive{{ObjectID: 1001, Protocol: "DCC28"}, consist}, wantError: true},
		{name: "unknown protocol", provider: "ecos", locomotives: []ECoSRawLocomotive{{ObjectID: 1001, Protocol: "MULTI2"}}, wantError: true},
		{name: "invalid consist ID", provider: "ecos", locomotives: []ECoSRawLocomotive{{Protocol: "MULTI"}}, wantError: true},
		{name: "duplicate consist", provider: "ecos", locomotives: []ECoSRawLocomotive{consist, consist}, wantError: true},
		{name: "ID shared with locomotive", provider: "ecos", locomotives: []ECoSRawLocomotive{valid, {ObjectID: 1001, Protocol: "MULTI"}}, wantError: true},
		{name: "invalid consist name", provider: "ecos", locomotives: []ECoSRawLocomotive{{ObjectID: 1001, Protocol: "MULTI", Name: string([]byte{0xff})}}, wantError: true},
		{name: "other providers stay strict", provider: "cs3", locomotives: []ECoSRawLocomotive{consist}, wantError: true},
		{name: "count limit includes consists", provider: "ecos", locomotives: make([]ECoSRawLocomotive, maxDigitalCenterLocomotives+1), wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			locos, skipped, err := normalizeDigitalCenterRoster(test.provider, &ECoSRawProbe{Locomotives: test.locomotives})
			if test.wantError {
				if !errors.Is(err, ErrDigitalCenterDeviceOutput) || len(locos) != 0 {
					t.Fatalf("locos=%#v err=%v", locos, err)
				}
				return
			}
			if err != nil || skipped != test.wantSkipped || len(locos) != len(test.locomotives)-skipped {
				t.Fatalf("locos=%#v skipped=%d err=%v", locos, skipped, err)
			}
		})
	}
}
