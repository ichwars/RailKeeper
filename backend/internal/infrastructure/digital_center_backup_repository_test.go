package infrastructure_test

import (
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestECoSBackupPreviewPersistenceIsAtomic(t *testing.T) {
	db := testDB(t)
	settings := application.NewSettingsService(db)
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{
		ECoS: application.DigitalProviderSettings{Host: "ecos.local", Port: "15471"},
	}); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	session := application.DigitalCenterReadSession{Provider: "ecos", State: application.DigitalCenterSessionReady,
		Host: "ecos.local", Port: 15471, ReadStartedAt: "2026-10-29T07:00:00Z", ReadCompletedAt: "2026-10-29T07:00:00Z"}
	item := application.DigitalCenterWorkItem{CenterObjectID: "1002", Name: "Test", CompareStatus: application.DigitalCompareNew,
		StationStatus: "read", Center: map[string]any{"cvs": []int{}}}
	message := application.DigitalCenterSessionMessage{Severity: application.DigitalCenterMessageInfo,
		Code: application.DigitalCenterMessageReadCompleted, Message: "Offline preview"}
	if _, err := repository.CreateECoSBackupSession(t.Context(), session, []application.DigitalCenterWorkItem{item, item}, message); err == nil {
		t.Fatal("duplicate object IDs accepted")
	}
	for _, table := range []string{"digital_center_read_sessions", "digital_center_work_items", "digital_center_session_messages"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("partial preview in %s: count=%d err=%v", table, count, err)
		}
	}
	saved, err := repository.CreateECoSBackupSession(t.Context(), session, []application.DigitalCenterWorkItem{item}, message)
	if err != nil || saved.ID == "" || saved.Capabilities != (application.DigitalCenterCapabilities{}) {
		t.Fatal(saved, err)
	}
	if entries, err := repository.ListWorkItems(t.Context(), saved.ID); err != nil || len(entries) != 1 {
		t.Fatal(entries, err)
	}
	session.Host = "another.local"
	if _, err := repository.CreateECoSBackupSession(t.Context(), session, []application.DigitalCenterWorkItem{item}, message); err == nil {
		t.Fatal("configuration changed before persistence but import succeeded")
	}
}

func TestECoSBackupContextUsesMatchingPriorStationAndCurrentManualVehicleData(t *testing.T) {
	db := testDB(t)
	vehicles := application.NewVehicleService(db)
	vehicle, err := vehicles.Create(t.Context(), application.CreateVehicleInput{
		Name: "Current manual name", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
		DecoderType: "Current manual model",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := vehicles.UpsertExternalMapping(t.Context(), vehicle.ID, application.VehicleExternalMapInput{
		Provider: "ecos", ExternalID: "1002", ExternalAddress: "3", ExternalProtocol: "DCC128",
	}, "admin"); err != nil {
		t.Fatal(err)
	}
	repository := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	for _, backup := range []bool{false, true} {
		session, err := repository.CreateSession(t.Context(), application.DigitalCenterReadSession{
			Provider: "ecos", State: application.DigitalCenterSessionReady, Host: "ecos.local", Port: 15471,
		})
		if err != nil {
			t.Fatal(err)
		}
		center := map[string]any{"profile": "Saved device model"}
		if backup {
			center = map[string]any{"profile": "Unrelated backup model", "cvRead": map[string]any{"source": "backup"}}
		}
		if err := repository.ReplaceWorkItems(t.Context(), session.ID, []application.DigitalCenterWorkItem{{
			CenterObjectID: "1002", VehicleID: vehicle.ID, Name: "Saved device label", Address: 3, Protocol: "DCC",
			StationStatus: "read", CompareStatus: application.DigitalCompareOK, Center: center,
		}}); err != nil {
			t.Fatal(err)
		}
	}
	contexts, err := repository.LoadECoSBackupContexts(t.Context(), "ecos.local", 15471, []int{1002, 2000})
	if err != nil || len(contexts) != 1 {
		t.Fatal(contexts, err)
	}
	entry := contexts[1002]
	if entry.Item.Center["profile"] != "Saved device model" || entry.Item.Name != "Saved device label" ||
		entry.VehicleID != vehicle.ID || entry.DecoderType != "Current manual model" ||
		entry.VehicleName != "Current manual name" || entry.ExternalAddress != "3" {
		t.Fatal("backup context guessed labels, copied stale manual values or lost mapping", entry)
	}
	if other, err := repository.LoadECoSBackupContexts(t.Context(), "other.local", 15471, []int{1002}); err != nil || len(other) != 0 {
		t.Fatal("another station's labels were reused", other, err)
	}
}
