package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/startup"
)

func TestStorageInfoRoutesRequireAdminAndReturnResolvedConfiguration(t *testing.T) {
	database := testRouterDB(t)
	setup := application.NewSetupService(database)
	auth := application.NewAuthService(database)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "admin-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "editor", Password: "editor-secure-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	dataPath := `C:\Users\Ada\AppData\Local\RailKeeper\data`
	receipt := &startup.MigrationReceipt{
		SourcePath: `C:\RailKeeper\data`, TargetPath: dataPath,
		MigratedAt: "2026-08-16T14:30:00Z", Version: "0.1.19", FilesVerified: 17,
	}
	router := NewRouter(Config{
		SetupService: setup,
		AuthService:  auth,
		StorageLocation: StorageLocationConfig{
			DataPath: dataPath, Mode: startup.StorageModeWindowsStandalone,
			OpenFolderAvailable: true, MigrationReceipt: receipt,
		},
	})
	adminSession, adminCookies := loginTestUser(t, router, "admin", "admin-secure-password")
	_, editorCookies := loginTestUser(t, router, "editor", "editor-secure-password")

	doJSON(t, router, http.MethodGet, "/api/v1/system/storage/info", "", nil, http.StatusUnauthorized)
	doJSON(t, router, http.MethodGet, "/api/v1/system/storage/info", "", editorCookies, http.StatusForbidden)
	response := doAuthedJSON(
		t, router, http.MethodGet, "/api/v1/system/storage/info", "", adminSession, adminCookies, http.StatusOK,
	)
	var info StorageLocationInfo
	if err := json.NewDecoder(response.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	want := StorageLocationInfo{
		DataPath: dataPath, Mode: startup.StorageModeWindowsStandalone,
		OpenFolderAvailable: true, MigrationReceipt: receipt,
	}
	if !reflect.DeepEqual(info, want) {
		t.Fatalf("storage info = %#v, want %#v", info, want)
	}
}

func TestStorageInfoOpenFolderUsesOnlyConfiguredPathAndRequiresCSRF(t *testing.T) {
	database := testRouterDB(t)
	setup := application.NewSetupService(database)
	auth := application.NewAuthService(database)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin-open", Email: "admin-open@example.test", Password: "admin-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	dataPath := `C:\Users\Ada\AppData\Local\RailKeeper\data`
	openedPath := ""
	router := NewRouter(Config{
		SetupService: setup,
		AuthService:  auth,
		StorageLocation: StorageLocationConfig{
			DataPath: dataPath, Mode: startup.StorageModeWindowsStandalone,
			OpenFolderAvailable: true,
			OpenFolder: func(_ context.Context, path string) error {
				openedPath = path
				return nil
			},
		},
	})
	session, cookies := loginTestUser(t, router, "admin-open", "admin-secure-password")

	withoutCSRF := doJSON(
		t, router, http.MethodPost, "/api/v1/system/storage/open-folder", "", cookies, http.StatusForbidden,
	)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
	doAuthedJSON(
		t, router, http.MethodPost, "/api/v1/system/storage/open-folder", "", session, cookies, http.StatusNoContent,
	)
	if openedPath != dataPath {
		t.Fatalf("opened path = %q, want configured %q", openedPath, dataPath)
	}

	openedPath = ""
	injected := doAuthedJSON(
		t, router, http.MethodPost,
		"/api/v1/system/storage/open-folder?path=C:%5Cwrong", "", session, cookies, http.StatusBadRequest,
	)
	assertProblem(t, injected, http.StatusBadRequest, "storage_folder_request_invalid")
	if openedPath != "" {
		t.Fatalf("query path reached opener: %q", openedPath)
	}
}

func TestStorageInfoOpenFolderReportsUnavailableCapability(t *testing.T) {
	database := testRouterDB(t)
	setup := application.NewSetupService(database)
	auth := application.NewAuthService(database)
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin-server", Email: "admin-server@example.test", Password: "admin-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{
		SetupService: setup,
		AuthService:  auth,
		StorageLocation: StorageLocationConfig{
			DataPath: "/data", Mode: startup.StorageModeServer, OpenFolderAvailable: false,
		},
	})
	session, cookies := loginTestUser(t, router, "admin-server", "admin-secure-password")
	response := doAuthedJSON(
		t, router, http.MethodPost, "/api/v1/system/storage/open-folder", "", session, cookies,
		http.StatusConflict,
	)
	assertProblem(t, response, http.StatusConflict, "storage_folder_open_unavailable")
}

func TestStorageInfoAcknowledgementUpdatesOnlySafeReceiptAtomically(t *testing.T) {
	dataDir := t.TempDir()
	legacyDir := t.TempDir()
	legacySentinel := filepath.Join(legacyDir, "keep.txt")
	if err := os.WriteFile(legacySentinel, []byte("legacy-unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	receipt := &startup.MigrationReceipt{
		SourcePath: legacyDir,
		TargetPath: dataDir,
		MigratedAt: time.Date(2026, 8, 16, 14, 30, 0, 0, time.UTC).Format(time.RFC3339),
		Version:    "0.1.19", FilesVerified: 21,
	}
	receiptData, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(dataDir, ".railkeeper-legacy-migration.json")
	if err = os.WriteFile(receiptPath, append(receiptData, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	database := testRouterDB(t)
	setup := application.NewSetupService(database)
	auth := application.NewAuthService(database)
	if err = setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin-ack", Email: "admin-ack@example.test", Password: "admin-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{
		SetupService: setup,
		AuthService:  auth,
		StorageLocation: StorageLocationConfig{
			DataPath: dataDir, Mode: startup.StorageModeWindowsStandalone,
			MigrationReceipt: receipt, AcknowledgeMigration: startup.AcknowledgeMigrationReceipt,
		},
	})
	session, cookies := loginTestUser(t, router, "admin-ack", "admin-secure-password")

	withoutCSRF := doJSON(
		t, router, http.MethodPost, "/api/v1/system/storage/migration-receipt/acknowledge", "", cookies,
		http.StatusForbidden,
	)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
	doAuthedJSON(
		t, router, http.MethodPost, "/api/v1/system/storage/migration-receipt/acknowledge", "", session,
		cookies, http.StatusNoContent,
	)

	legacyContent, err := os.ReadFile(legacySentinel)
	if err != nil || string(legacyContent) != "legacy-unchanged" {
		t.Fatalf("legacy source changed: %q, %v", legacyContent, err)
	}
	updatedData, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var updated startup.MigrationReceipt
	if err = json.Unmarshal(updatedData, &updated); err != nil {
		t.Fatal(err)
	}
	want := *receipt
	want.Acknowledged = true
	if !reflect.DeepEqual(updated, want) {
		t.Fatalf("updated receipt = %#v, want %#v", updated, want)
	}
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != ".railkeeper-legacy-migration.json" {
		t.Fatalf("acknowledgement left unexpected files: %v", entries)
	}

	infoResponse := doAuthedJSON(
		t, router, http.MethodGet, "/api/v1/system/storage/info", "", session, cookies, http.StatusOK,
	)
	var info StorageLocationInfo
	if err = json.NewDecoder(infoResponse.Body).Decode(&info); err != nil {
		t.Fatal(err)
	}
	if info.MigrationReceipt == nil || !info.MigrationReceipt.Acknowledged {
		t.Fatalf("storage info did not retain acknowledgement: %#v", info)
	}
}
