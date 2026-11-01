package application_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestVehicleECoSProfileUsesSavedMatchingStationWithoutChangingManualData(t *testing.T) {
	db := backupTestDB(t, t.TempDir())
	service := application.NewVehicleService(db)
	settings := application.NewSettingsService(db)
	configuration := application.DigitalCenterSettings{ECoS: application.DigitalProviderSettings{
		Enabled: true, Host: "ecos.test", Port: "15471",
	}}
	if _, err := settings.UpdateDigitalSettings(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Test", Name: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
		DecoderType: "Meine manuelle Zuordnung",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.UpsertExternalMapping(t.Context(), vehicle.ID, application.VehicleExternalMapInput{
		Provider: "ecos", ExternalID: "1002", ExternalAddress: "3", ExternalProtocol: "DCC128",
	}, "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ImportCVValues(t.Context(), vehicle.ID, application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{
		{CVNumber: 8, Value: 97, DecoderProfile: "Manual", Description: "Original"},
	}}); err != nil {
		t.Fatal(err)
	}
	manual, _ := service.ListCVValues(t.Context(), vehicle.ID)
	insert := func(id, status, body, date string, address int) {
		t.Helper()
		if _, err := db.Exec(`INSERT INTO digital_center_read_sessions
(id,provider,state,host,port,read_completed_at,created_at,updated_at) VALUES(?,'ecos','ready','ecos.test',15471,?,?,?)`,
			id, date, date, date); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(`INSERT INTO digital_center_work_items
(id,session_id,center_object_id,vehicle_id,name,decoder_address,protocol,compare_status,station_status,center_json,created_at,updated_at)
VALUES(?,?,'1002',?,'Test',?,'DCC','ok',?,?,?,?)`, id, id, vehicle.ID, address, status, body, date, date); err != nil {
			t.Fatal(err)
		}
	}
	read := func() *application.Vehicle {
		t.Helper()
		result, err := service.Get(t.Context(), vehicle.ID)
		if err != nil {
			t.Fatal(err)
		}
		if result.DecoderType != "Meine manuelle Zuordnung" || !reflect.DeepEqual(manual, result.CVValues) {
			t.Fatal("source lookup changed manual decoder/CV data")
		}
		if len(result.ExternalMappings) != 1 || result.ExternalMappings[0].ExternalID != "1002" {
			t.Fatal("vehicle detail omitted the saved station mapping needed by the function dialog")
		}
		return result
	}
	if read().ECoSDecoderProfile != nil {
		t.Fatal("profile guessed without a saved comparison")
	}
	data, _ := json.Marshal(map[string]any{"profile": "LokPilot 5 micro DCC", "cvRead": map[string]any{
		"source": "database", "nextCV": 1025, "unavailable": 0, "sourceHost": "ecos.test", "readAt": "2026-10-27T07:01:00Z",
	}})
	insert("first", "read", string(data), "2026-10-27T07:00:00Z", 3)
	source := read().ECoSDecoderProfile
	if source == nil || source.Name != "LokPilot 5 micro DCC" || source.ObjectID != "1002" ||
		source.Host != "ecos.test" || source.ReadAt != "2026-10-27T07:01:00Z" {
		t.Fatalf("missing source context: %+v", source)
	}
	insert("completed-db", "incomplete", string(data), "2026-10-27T07:01:30Z", 3)
	if read().ECoSDecoderProfile == nil {
		t.Fatal("verified complete database export lost its profile because of older raw-probe errors")
	}
	backup, _ := json.Marshal(map[string]any{"profile": "Unrelated backup model", "cvRead": map[string]any{
		"source": "backup", "nextCV": 1025, "unavailable": 0, "sourceHost": "ecos.test", "readAt": "2026-10-20T07:00:00Z",
	}})
	for _, status := range []string{"read", "incomplete", "missing"} {
		insert("backup-"+status, status, string(backup), "2026-10-27T07:01:40Z", 3)
		profile := read().ECoSDecoderProfile
		if profile == nil || profile.Name != source.Name || profile.ReadAt != source.ReadAt {
			t.Fatalf("separate backup workspace displaced current decoder model: %+v", profile)
		}
	}
	var conflict map[string]any
	if err := json.Unmarshal(data, &conflict); err != nil {
		t.Fatal(err)
	}
	conflict["cvSources"] = map[string]any{"sources": []string{"database", "backup"},
		"backupStatus": "checked", "conflicts": []map[string]any{{"number": 8, "databaseValue": 0, "backupValue": 151}}}
	conflictData, _ := json.Marshal(conflict)
	insert("source-conflict", "read", string(conflictData), "2026-10-27T07:01:50Z", 3)
	if read().ECoSDecoderProfile != nil {
		t.Fatal("latest source conflict revived an older model profile or treated disputed CV8 as simply absent")
	}
	insert("reconciled", "read", string(data), "2026-10-27T07:01:55Z", 3)
	if read().ECoSDecoderProfile == nil {
		t.Fatal("a later non-conflicting observation failed to restore profile context")
	}
	configuration.ECoS.Host = "another.test"
	if _, err := settings.UpdateDigitalSettings(t.Context(), configuration); err != nil {
		t.Fatal(err)
	}
	if read().ECoSDecoderProfile != nil {
		t.Fatal("profile from another configured station accepted")
	}
	configuration.ECoS.Host = "ecos.test"
	_, _ = settings.UpdateDigitalSettings(t.Context(), configuration)
	insert("newer", "missing", string(data), "2026-10-27T07:02:00Z", 3)
	if read().ECoSDecoderProfile != nil {
		t.Fatal("stale profile revived after a newer missing snapshot")
	}
	insert("changed", "read", string(data), "2026-10-27T07:03:00Z", 4)
	if read().ECoSDecoderProfile != nil {
		t.Fatal("profile with a different mapped address accepted")
	}
	insert("malformed", "read", "not-json", "2026-10-27T07:04:00Z", 3)
	if read().ECoSDecoderProfile != nil {
		t.Fatal("malformed cache accepted")
	}
}
