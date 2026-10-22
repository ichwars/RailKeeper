package api

import (
	"context"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

type apiStoredCVReader struct{ apiWorkspaceECoSReader }

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
	router := NewRouter(Config{AuthService: auth, DigitalCenterWorkspace: workspace})
	admin := loginRouteTestUser(t, auth, "cv-admin", "test-password")
	viewer := loginRouteTestUser(t, auth, "cv-viewer", "test-password")
	path := "/api/v1/digital-centers/read-sessions/" + session.ID + "/items/" + item.ID
	for _, suffix := range []string{"cv-read", "cv-import"} {
		assertStatus(t, authedJSONRequest(t, router, viewer, http.MethodPost, path+"/"+suffix, map[string]any{}, true), http.StatusForbidden)
		assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/"+suffix, map[string]any{}, false), http.StatusForbidden)
	}
	for _, input := range []map[string]any{{"startCV": 0}, {"startCV": 1025}, {"startCV": 2}, {"startCV": 1, "host": "attacker.local"}} {
		assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read", input, true), http.StatusBadRequest)
	}
	read := authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-read", map[string]any{"startCV": 1}, true)
	assertStatus(t, read, http.StatusOK)
	if reader.input.Host != "trusted.local" {
		t.Fatal("untrusted target")
	}
	before, _ := vehicles.ListCVValues(t.Context(), v.ID)
	if len(before) != 0 {
		t.Fatal("reading imported CVs implicitly")
	}
	assertStatus(t, authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-import",
		map[string]any{"vehicleId": v.ID, "values": []int{99}}, true), http.StatusBadRequest)
	for _, wantImported := range []int{1, 0} {
		r := authedJSONRequest(t, router, admin, http.MethodPost, path+"/cv-import", map[string]any{"vehicleId": v.ID}, true)
		assertStatus(t, r, http.StatusOK)
		var result application.ECoSCVImportResult
		decodeResponse(t, r, &result)
		if result.Imported != wantImported || result.Preserved != 1-wantImported {
			t.Fatalf("result=%+v", result)
		}
	}
	stored, _ := vehicles.ListCVValues(t.Context(), v.ID)
	if len(stored) != 1 || stored[0].Value != 0 {
		t.Fatal("zero value lost")
	}
}
