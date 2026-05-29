package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestLocalSmokeLoginVehicleImportBackupAndRoles(t *testing.T) {
	dataDir := t.TempDir()
	db := testRouterDBWithDataDir(t, dataDir)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	vehicles := application.NewVehicleService(db)
	masterData := application.NewMasterDataService(db)
	backup := application.NewBackupService(db, dataDir)
	exhibition := application.NewExhibitionService(db)
	router := NewRouter(Config{
		DataDir:           dataDir,
		SetupService:      setup,
		AuthService:       auth,
		VehicleService:    vehicles,
		MasterDataService: masterData,
		BackupService:     backup,
		ExhibitionService: exhibition,
	})

	doJSON(t, router, http.MethodPost, "/api/v1/setup/admin", `{"username":"admin","email":"admin@example.test","password":"very-secure-password"}`, nil, http.StatusCreated)
	adminSession, adminCookies := loginTestUser(t, router, "admin", "very-secure-password")

	doJSON(t, router, http.MethodGet, "/api/v1/vehicles", "", nil, http.StatusUnauthorized)
	doJSON(t, router, http.MethodPost, "/api/v1/vehicles", `{"manufacturer":"Piko","name":"BR 118","gauge":"H0","category":"Lokomotive","gattung":"Diesellok"}`, adminCookies, http.StatusForbidden)

	rolesResponse := doJSON(t, router, http.MethodGet, "/api/v1/roles", "", adminCookies, http.StatusOK)
	var roles []application.RoleView
	if err := json.NewDecoder(rolesResponse.Body).Decode(&roles); err != nil {
		t.Fatal(err)
	}
	if len(roles) < 4 {
		t.Fatalf("expected seeded roles, got %#v", roles)
	}

	doAuthedJSON(t, router, http.MethodPost, "/api/v1/users", `{"username":"viewer","email":"viewer@example.test","password":"viewer-secure-password","roles":["Viewer"]}`, adminSession, adminCookies, http.StatusCreated)
	doAuthedJSON(t, router, http.MethodPost, "/api/v1/users", `{"username":"messe","email":"messe@example.test","password":"messe-secure-password","roles":["Messe"]}`, adminSession, adminCookies, http.StatusCreated)
	createVehicleResponse := doAuthedJSON(t, router, http.MethodPost, "/api/v1/vehicles", `{"manufacturer":"Piko","name":"BR 118","gauge":"H0","category":"Lokomotive","gattung":"Diesellok"}`, adminSession, adminCookies, http.StatusCreated)
	var vehicle application.Vehicle
	if err := json.NewDecoder(createVehicleResponse.Body).Decode(&vehicle); err != nil {
		t.Fatal(err)
	}
	if vehicle.ID == "" || vehicle.InventoryNumber == "" {
		t.Fatalf("expected created vehicle identity, got %#v", vehicle)
	}

	listResponse := doJSON(t, router, http.MethodGet, "/api/v1/vehicles", "", adminCookies, http.StatusOK)
	var listed []application.Vehicle
	if err := json.NewDecoder(listResponse.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != vehicle.ID {
		t.Fatalf("expected vehicle list to include created vehicle, got %#v", listed)
	}

	masterDataExport := doJSON(t, router, http.MethodGet, "/api/v1/master-data/export", "", adminCookies, http.StatusOK).Body.Bytes()
	doAuthedMultipart(t, router, "/api/v1/master-data/import", "railkeeper-master-data.json", masterDataExport, adminSession, adminCookies, http.StatusOK)

	backupExport := doJSON(t, router, http.MethodGet, "/api/v1/backup/export", "", adminCookies, http.StatusOK).Body.Bytes()
	validateResponse := doAuthedMultipart(t, router, "/api/v1/backup/validate", "railkeeper-backup.json", backupExport, adminSession, adminCookies, http.StatusOK)
	var validation application.BackupValidationResult
	if err := json.NewDecoder(validateResponse.Body).Decode(&validation); err != nil {
		t.Fatal(err)
	}
	if !validation.Compatible || validation.RowCount == 0 {
		t.Fatalf("expected compatible backup validation, got %#v", validation)
	}

	doAuthedJSON(t, router, http.MethodDelete, "/api/v1/vehicles/"+vehicle.ID, "", adminSession, adminCookies, http.StatusNoContent)
	doJSON(t, router, http.MethodGet, "/api/v1/vehicles", "", adminCookies, http.StatusOK)
	restoreResponse := doAuthedMultipart(t, router, "/api/v1/backup/restore", "railkeeper-backup.json", backupExport, adminSession, adminCookies, http.StatusOK)
	var restore application.BackupImportResult
	if err := json.NewDecoder(restoreResponse.Body).Decode(&restore); err != nil {
		t.Fatal(err)
	}
	if restore.RestoredRows == 0 {
		t.Fatalf("expected backup restore to restore rows, got %#v", restore)
	}
	restoredResponse := doJSON(t, router, http.MethodGet, "/api/v1/vehicles", "", adminCookies, http.StatusOK)
	var restored []application.Vehicle
	if err := json.NewDecoder(restoredResponse.Body).Decode(&restored); err != nil {
		t.Fatal(err)
	}
	if len(restored) != 1 || restored[0].ID != vehicle.ID {
		t.Fatalf("expected restore to bring vehicle back, got %#v", restored)
	}

	viewerSession, viewerCookies := loginTestUser(t, router, "viewer", "viewer-secure-password")
	doJSON(t, router, http.MethodGet, "/api/v1/vehicles", "", viewerCookies, http.StatusOK)
	doAuthedJSON(t, router, http.MethodPost, "/api/v1/vehicles", `{"manufacturer":"Roco","name":"V 200","gauge":"H0","category":"Lokomotive","gattung":"Diesellok"}`, viewerSession, viewerCookies, http.StatusForbidden)

	messeSession, messeCookies := loginTestUser(t, router, "messe", "messe-secure-password")
	doJSON(t, router, http.MethodGet, "/api/v1/exhibition-lists", "", messeCookies, http.StatusOK)
	doJSON(t, router, http.MethodGet, "/api/v1/vehicles", "", messeCookies, http.StatusForbidden)
	doAuthedJSON(t, router, http.MethodPost, "/api/v1/exhibition-lists", `{"designation":"Leipzig 2026","date":"2026-05-12"}`, messeSession, messeCookies, http.StatusForbidden)
}

func doJSON(t *testing.T, router http.Handler, method, target, body string, cookies []*http.Cookie, want int) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s %s: expected %d, got %d: %s", method, target, want, response.Code, response.Body.String())
	}
	return response
}

func doAuthedJSON(t *testing.T, router http.Handler, method, target, body string, session application.SessionView, cookies []*http.Cookie, want int) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("%s %s: expected %d, got %d: %s", method, target, want, response.Code, response.Body.String())
	}
	return response
}

func doAuthedMultipart(t *testing.T, router http.Handler, target, fileName string, fileData []byte, session application.SessionView, cookies []*http.Cookie, want int) *httptest.ResponseRecorder {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(fileData); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	request := httptest.NewRequest(http.MethodPost, target, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != want {
		t.Fatalf("POST %s: expected %d, got %d: %s", target, want, response.Code, response.Body.String())
	}
	return response
}
