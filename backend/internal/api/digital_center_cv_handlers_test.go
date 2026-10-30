package api

import (
	"context"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

type apiStoredCVReader struct{ apiWorkspaceECoSReader }

func (reader *apiStoredCVReader) ReadStoredCVDatabase(_ context.Context, input application.ECoSConnectionInput,
	objectID int) (application.ECoSCVDatabase, error) {
	reader.input = input
	high, low := 16, 3
	return application.ECoSCVDatabase{
		Locomotive: application.ECoSLocomotive{ObjectID: objectID, Name: "Test", Address: 3, Protocol: "DCC28", Profile: "LokPilot 5 micro DCC"},
		Values:     []application.ECoSCVValue{{Number: 257, Value: 20, IndexCV31: &high, IndexCV32: &low}},
	}, nil
}

func (reader *apiStoredCVReader) ReadStoredCVs(_ context.Context, input application.ECoSConnectionInput,
	objectID, startCV int) (application.ECoSStoredCVBatch, error) {
	reader.input = input
	return application.ECoSStoredCVBatch{Values: []application.ECoSCVValue{{Number: startCV, Value: 0}}, Unavailable: 31}, nil
}

func TestECoSStoredCVRoutesUseTrustedSnapshotAndProtectExistingData(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "cv-admin", Password: "test-password", Roles: []string{"Admin"}},
		{Username: "cv-viewer", Password: "test-password", Roles: []string{"Viewer"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	settings := application.NewSettingsService(db)
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{
		ECoS: application.DigitalProviderSettings{Enabled: true, Host: "trusted.local", Port: "15471"},
	}); err != nil {
		t.Fatal(err)
	}
	reader := &apiStoredCVReader{apiWorkspaceECoSReader{probe: application.ECoSRawProbe{
		Locomotives: []application.ECoSRawLocomotive{{ObjectID: 1002, Address: 3, Name: "Test", Protocol: "DCC28"}},
	}}}
	vehicles := application.NewVehicleService(db)
	v, err := vehicles.Create(t.Context(), application.CreateVehicleInput{
		Manufacturer: "Test", Name: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok",
		DecoderType: "Meine manuelle Zuordnung",
	}, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := vehicles.UpsertExternalMapping(t.Context(), v.ID,
		application.VehicleExternalMapInput{Provider: "ecos", ExternalID: "1002"}, "admin"); err != nil {
		t.Fatal(err)
	}
	workspace := application.NewDigitalCenterWorkspaceService(infrastructure.NewDigitalCenterWorkspaceRepository(db),
		settings, reader, nil, vehicles, auth)
	session, err := workspace.StartReadSession(t.Context(), "ecos", "admin")
	if err != nil {
		t.Fatal(err)
	}
	page, err := workspace.ListWorkItems(t.Context(), session.ID, application.DigitalCenterWorkItemFilter{})
	if err != nil {
		t.Fatal(err)
	}
	item := page.Items[0]
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace, VehicleService: vehicles})
	admin := loginRouteTestUser(t, auth, "cv-admin", "test-password")
	viewer := loginRouteTestUser(t, auth, "cv-viewer", "test-password")
	path := "/api/v1/digital-centers/read-sessions/" + session.ID + "/items/" + item.ID
	for _, suffix := range []string{"cv-read", "cv-import"} {
		assertStatus(t, authedJSONRequest(t, router, viewer, http.MethodPost, path+"/"+suffix, map[string]any{}, true), http.StatusForbidden)
		assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/"+suffix, map[string]any{}, false), http.StatusForbidden)
	}
	for _, input := range []map[string]any{{"startCV": 0}, {"startCV": 1025}, {"startCV": 2}, {"startCV": 33, "restart": true}, {"startCV": 1, "host": "attacker.local"}} {
		assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read", input, true), http.StatusBadRequest)
	}
	read := authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read", map[string]any{"startCV": 1}, true)
	assertStatus(t, read, http.StatusOK)
	if reader.input.Host != "trusted.local" {
		t.Fatal("untrusted target")
	}
	assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-import", map[string]any{"vehicleId": v.ID}, true), http.StatusConflict)
	before, _ := vehicles.ListCVValues(t.Context(), v.ID)
	if len(before) != 0 {
		t.Fatal("reading imported CVs implicitly")
	}
	assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-import",
		map[string]any{"vehicleId": v.ID, "values": []int{99}}, true), http.StatusBadRequest)
	for start := 33; start <= 1024; start += 32 {
		r := authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read", map[string]any{"startCV": start}, true)
		assertStatus(t, r, http.StatusOK)
	}
	for _, wantImported := range []int{32, 0} {
		r := authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-import", map[string]any{"vehicleId": v.ID}, true)
		assertStatus(t, r, http.StatusOK)
		var result application.ECoSCVImportResult
		decodeResponse(t, r, &result)
		if result.Imported != wantImported || result.Preserved != 32-wantImported {
			t.Fatalf("result=%+v", result)
		}
	}
	stored, _ := vehicles.ListCVValues(t.Context(), v.ID)
	if len(stored) != 32 || stored[0].Value != 0 {
		t.Fatal("zero value lost")
	}
	assertStatus(t, authedJSONRequest(t, router, viewer, http.MethodPost, path+"/cv-read",
		map[string]any{"startCV": 1, "source": "database"}, true), http.StatusForbidden)
	for _, input := range []map[string]any{
		{"startCV": 1, "source": "unknown"}, {"startCV": 33, "source": "database"},
		{"startCV": 1, "source": "database", "host": "attacker.local"},
	} {
		assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read", input, true), http.StatusBadRequest)
	}
	response := authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read",
		map[string]any{"startCV": 1, "source": "database"}, true)
	assertStatus(t, response, http.StatusOK)
	var exported application.DigitalCenterWorkItem
	decodeResponse(t, response, &exported)
	if response.Header().Get("Cache-Control") != "no-store" || reader.input.Host != "trusted.local" {
		t.Fatal("database export did not use protected configured target")
	}
	beforeImport, _ := vehicles.ListCVValues(t.Context(), v.ID)
	if len(beforeImport) != len(stored) {
		t.Fatal("database read implicitly imported CVs")
	}
	vehicleResponse := authedJSONRequest(t, router, viewer, http.MethodGet, "/api/v1/vehicles/"+v.ID, nil, false)
	assertStatus(t, vehicleResponse, http.StatusOK)
	var withSource application.Vehicle
	decodeResponse(t, vehicleResponse, &withSource)
	if withSource.ECoSDecoderProfile == nil || withSource.ECoSDecoderProfile.Name != "LokPilot 5 micro DCC" ||
		withSource.ECoSDecoderProfile.Host != "trusted.local" || withSource.DecoderType != "Meine manuelle Zuordnung" ||
		len(withSource.CVValues) != len(stored) || len(withSource.ExternalMappings) != 1 ||
		withSource.ExternalMappings[0].ExternalID != "1002" {
		t.Fatalf("cached profile lookup lost provenance or changed inventory: %+v", withSource.ECoSDecoderProfile)
	}
	assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-import",
		map[string]any{"vehicleId": v.ID}, true), http.StatusOK)
	afterImport, _ := vehicles.ListCVValues(t.Context(), v.ID)
	if len(afterImport) != 33 {
		t.Fatal("indexed database value missing")
	}
}
