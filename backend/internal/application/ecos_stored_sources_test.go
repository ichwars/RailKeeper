package application

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStoredSourceMergeKeepsPagesAndWithholdsConflictingRegisters(t *testing.T) {
	zero, high, low := 0, 16, 3
	primary := []ECoSCVValue{{Number: 8, Value: 0}, {Number: 257, Value: 1},
		{Number: 257, Value: 2, IndexCV31: &zero, IndexCV32: &zero}}
	secondary := []ECoSCVValue{{Number: 8, Value: 151}, {Number: 257, Value: 2, IndexCV31: &zero, IndexCV32: &zero},
		{Number: 257, Value: 3, IndexCV31: &high, IndexCV32: &low}}
	values, info, err := mergeStoredCVSources(primary, secondary, true, "checked")
	if err != nil || len(values) != 3 || len(info.Conflicts) != 1 || info.BackupAdded != 1 ||
		info.Conflicts[0].Number != 8 || info.Conflicts[0].DatabaseValue != 0 || info.Conflicts[0].BackupValue != 151 {
		t.Fatalf("lost source conflict: %+v %+v %v", values, info, err)
	}
	for _, value := range values {
		if value.Number == 8 {
			t.Fatal("a disputed manufacturer CV remained eligible for import")
		}
	}
	if values[0].IndexCV31 != nil || *values[1].IndexCV31 != 0 || *values[2].IndexCV32 != 3 {
		t.Fatalf("page-less, zero and explicit pages collapsed: %+v", values)
	}
	if primary[0].Value != 0 || secondary[0].Value != 151 {
		t.Fatal("merge rewrote original source evidence")
	}
	_, info, err = mergeStoredCVSources(primary, nil, true, "failed")
	if err != nil || info.Warning == "" || len(info.Sources) != 1 || info.Sources[0] != "database" {
		t.Fatalf("missing backup was silent: %+v %v", info, err)
	}
}

func TestStoredSourceConflictRetainsProfileOnlyAsUnqualifiedEvidence(t *testing.T) {
	item := newDigitalCenterWorkItem(ECoSRawLocomotive{ObjectID: 1002, Name: "Test", Address: 3,
		Protocol: "DCC", Profile: "LokPilot 5 micro DCC", CVSources: &ECoSCVSourceInfo{
			Sources: []string{"database", "backup"}, Conflicts: []ECoSCVSourceConflict{{Number: 8, DatabaseValue: 0, BackupValue: 151}}}})
	if item.Center["profile"] != "" || item.Center["sourceProfile"] != "LokPilot 5 micro DCC" {
		t.Fatal("disputed manufacturer identity still enabled a model-profile fallback")
	}
}

func TestStoredBackupConfigurationIsOptionalBoundedAndConfined(t *testing.T) {
	if _, _, err := readStoredBackupConfiguration(""); !os.IsNotExist(err) {
		t.Fatal("missing configuration unexpectedly required")
	}
	root := t.TempDir()
	private := filepath.Join(root, "private", "ecos-backup")
	if err := os.MkdirAll(private, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(private, "keyring.gpg"), []byte("synthetic-test-ring"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(private, "passphrase"), []byte("synthetic test pass\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	key, passphrase, err := readStoredBackupConfiguration(root)
	if err != nil || string(key) != "synthetic-test-ring" || string(passphrase) != "synthetic test pass" {
		t.Fatal("configuration failed to preserve interior passphrase spaces")
	}
	outside := filepath.Join(t.TempDir(), "outside")
	if err := os.WriteFile(outside, []byte("never-read"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(private, "keyring.gpg")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(private, "keyring.gpg")); err != nil {
		t.Skip("symlinks unavailable")
	}
	if _, _, err := readStoredBackupConfiguration(root); err == nil {
		t.Fatal("private configuration escaped the data directory")
	}
}

type completeStoredRosterStub struct {
	workspaceECoSReaderStub
	reads int
}

func (stub *completeStoredRosterStub) ReadStoredLocomotiveRoster(context.Context, ECoSConnectionInput) (*ECoSRawProbe, error) {
	stub.reads++
	return stub.probe, stub.err
}

func TestStoredRosterPersistsCompleteCVAndFunctionSnapshotsWithoutDiagnosticProbe(t *testing.T) {
	state := &ECoSCVReadState{Source: "database", NextCV: 1025,
		ReadAt: "2026-10-29T10:00:00Z", SourceHost: "ecos.local"}
	functions := &ECoSFunctionSnapshot{ObjectID: 1002, Host: "ecos.local", ReadAt: state.ReadAt,
		Functions: []ECoSFunctionValue{{Index: 1, Description: 12, Momentary: true}}}
	reader := &completeStoredRosterStub{workspaceECoSReaderStub: workspaceECoSReaderStub{probe: &ECoSRawProbe{
		Locomotives: []ECoSRawLocomotive{{ObjectID: 1002, Name: "Test", Address: 3, Protocol: "DCC128",
			Profile: "LokPilot 5 micro DCC", CVs: []ECoSCVValue{{Number: 1, Value: 3}}, CVRead: state,
			CVSources:        &ECoSCVSourceInfo{Sources: []string{"database", "backup"}, BackupStatus: "checked"},
			FunctionSnapshot: functions}},
	}}}
	repository := &workspaceRepositoryMemory{}
	settings := &workspaceSettingsReaderStub{value: DigitalCenterSettings{ECoS: DigitalProviderSettings{
		Enabled: true, Host: "ecos.local", Port: "15471"}}}
	service := NewDigitalCenterWorkspaceService(repository, settings, reader, nil, &workspaceVehicleReaderStub{}, nil)
	session, err := service.StartReadSession(t.Context(), "ecos", "admin")
	if err != nil || session.State != DigitalCenterSessionReady || len(repository.items) != 1 || reader.reads != 1 || reader.probeCalls != 0 {
		t.Fatalf("normal read used diagnostic probes or lost the session: %+v %v", session, err)
	}
	values, saved, err := storedECoSCVSnapshot(repository.items[0])
	if err != nil || len(values) != 1 || saved.NextCV != 1025 || saved.Source != "database" {
		t.Fatalf("first read omitted complete CV snapshot: %+v %+v %v", values, saved, err)
	}
	data, _ := json.Marshal(repository.items[0].Center["functionSnapshot"])
	var snapshot ECoSFunctionSnapshot
	if json.Unmarshal(data, &snapshot) != nil || len(snapshot.Functions) != 1 || !snapshot.Functions[0].Momentary {
		t.Fatal("normal read omitted stored function-key evidence")
	}
	if !session.Capabilities.ReadLocomotives || !session.Capabilities.WriteLocomotives {
		t.Fatal("normal read was turned into an isolated backup workflow")
	}
}
