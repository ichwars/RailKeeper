package application

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	ecospkg "railkeeper/backend/internal/ecos"
)

type backupImportRepository struct {
	workspaceRepositoryMemory
	contexts map[int]ECoSBackupContext
	saves    int
	fail     error
}

func (repository *backupImportRepository) LoadECoSBackupContexts(
	_ context.Context, _ string, _ int, _ []int,
) (map[int]ECoSBackupContext, error) {
	return repository.contexts, nil
}

func (repository *backupImportRepository) CreateECoSBackupSession(
	_ context.Context, session DigitalCenterReadSession, items []DigitalCenterWorkItem, message DigitalCenterSessionMessage,
) (DigitalCenterReadSession, error) {
	if repository.fail != nil {
		return DigitalCenterReadSession{}, repository.fail
	}
	repository.saves++
	session.ID = "imported-backup"
	repository.session, repository.items, repository.messages = session, items, []DigitalCenterSessionMessage{message}
	return session, nil
}

func backupImportFixture() (ecospkg.Backup, ECoSBackupContext) {
	high, low := 16, 3
	return ecospkg.Backup{Serial: 0x12345678, Locomotives: []ecospkg.BackupLocomotive{{
			ObjectID: 1002, Values: []ecospkg.DatabaseCV{{Number: 257, Value: 20, IndexCV31: &high, IndexCV32: &low}},
		}}}, ECoSBackupContext{Item: DigitalCenterWorkItem{
			CenterObjectID: "1002", VehicleID: "vehicle-1", Name: "Saved label", Address: 3, Protocol: "DCC128",
			StationStatus: "read", CompareStatus: DigitalCompareOK, Center: map[string]any{"profile": "Saved decoder"},
		}, VehicleID: "vehicle-1", VehicleName: "Manual name", DecoderType: "Manual decoder",
			ExternalAddress: "3", ExternalProtocol: "DCC"}
}

func TestECoSBackupImportOnlyCreatesOfflinePreviewAndRequiresStationConfirmationForLinks(t *testing.T) {
	backup, identity := backupImportFixture()
	for _, confirmed := range []bool{false, true} {
		repository := &backupImportRepository{contexts: map[int]ECoSBackupContext{1002: identity}}
		service := NewDigitalCenterWorkspaceService(repository, &workspaceSettingsReaderStub{value: DigitalCenterSettings{
			ECoS: DigitalProviderSettings{Host: "ecos.local", Port: "15471"},
		}}, nil, nil, nil, nil)
		service.now = func() time.Time { return time.Date(2026, 10, 29, 7, 0, 0, 0, time.UTC) }
		result, err := service.importDecodedECoSBackup(t.Context(), backup,
			ECoSBackupImportInput{ConfirmStation: confirmed, StationHost: "ecos.local"}, "admin")
		if err != nil || result.Summary.CVValues != 1 || result.Summary.Profiles != 1 ||
			result.Summary.StationSerial != backup.Serial || result.Summary.ImportedAt != "2026-10-29T07:00:00Z" ||
			result.Session.Capabilities != (DigitalCenterCapabilities{}) || repository.saves != 1 {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		item := repository.items[0]
		values, state, err := storedECoSCVSnapshot(item)
		if err != nil || len(values) != 1 || state.Source != "backup" || state.SourceHost != "ecos.local" ||
			state.ReadAt != result.Summary.ImportedAt || item.Name != "Saved label" || *values[0].IndexCV32 != 3 {
			t.Fatalf("snapshot=%+v state=%+v err=%v", item, state, err)
		}
		if confirmed {
			if item.VehicleID != "vehicle-1" || item.RailKeeper["decoderType"] != "Manual decoder" || result.Summary.LinkedVehicles != 1 {
				t.Fatal("confirmed matching existing mapping lost")
			}
		} else if item.VehicleID != "" || result.Summary.LinkedVehicles != 0 {
			t.Fatal("shared backup key and object ID were treated as proof of station identity")
		}
	}
	if identity.Item.Name != "Saved label" || identity.DecoderType != "Manual decoder" {
		t.Fatal("import changed existing identity")
	}
}

func TestECoSBackupImportRejectsBadInputBeforePersistenceAndKeepsUnknownObjectsUnassigned(t *testing.T) {
	backup, identity := backupImportFixture()
	repository := &backupImportRepository{contexts: map[int]ECoSBackupContext{}}
	service := NewDigitalCenterWorkspaceService(repository, &workspaceSettingsReaderStub{value: DigitalCenterSettings{
		ECoS: DigitalProviderSettings{Host: "ecos.local", Port: "15471"},
	}}, nil, nil, nil, nil)
	_, err := service.importDecodedECoSBackup(t.Context(), backup,
		ECoSBackupImportInput{ConfirmStation: true, StationHost: "different.local"}, "admin")
	if !errors.Is(err, ErrECoSBackupStationMismatch) || repository.saves != 0 {
		t.Fatal("mismatched station created preview")
	}
	_, err = service.importDecodedECoSBackup(t.Context(), backup, ECoSBackupImportInput{}, "admin")
	if err != nil || repository.items[0].VehicleID != "" || repository.items[0].Protocol != "" ||
		repository.items[0].Address != 0 || repository.items[0].Name != "ECoS-Objekt 1002" {
		t.Fatal("unknown backup identity was guessed", err)
	}
	for _, change := range []func(*ECoSBackupContext){
		func(c *ECoSBackupContext) { c.ExternalAddress = "4" },
		func(c *ECoSBackupContext) { c.ExternalProtocol = "MFX" },
		func(c *ECoSBackupContext) { c.Item.VehicleID = "different-vehicle" },
		func(c *ECoSBackupContext) { c.Item.StationStatus = "missing" },
	} {
		invalid := identity
		change(&invalid)
		if item := newECoSBackupItem(1002, invalid, true); item.VehicleID != "" {
			t.Fatal("unverified mapping accepted")
		}
	}
	before := repository.saves
	backup.Locomotives[0].Values[0].Value = 256
	if _, err := service.importDecodedECoSBackup(t.Context(), backup, ECoSBackupImportInput{}, "admin"); !errors.Is(err, ErrECoSBackupInvalid) || repository.saves != before {
		t.Fatal("invalid backup partially persisted", err)
	}
	backup.Locomotives[0].Values[0].Value = 20
	repository.fail = errors.New("storage failed")
	if result, err := service.importDecodedECoSBackup(t.Context(), backup, ECoSBackupImportInput{}, "admin"); err == nil || !reflect.DeepEqual(result, ECoSBackupImportResult{}) {
		t.Fatal("persistence failure returned successful preview")
	}
}
