package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto"
	_ "crypto/sha256"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/openpgp"        //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.
	"golang.org/x/crypto/openpgp/packet" //nolint:staticcheck // Legacy ECoS OpenPGP format; bounded, integrity-checked read-only compatibility.

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func syntheticAPIBackup(t *testing.T, dataDir string) []byte {
	t.Helper()
	entity, err := openpgp.NewEntity("Synthetic API backup", "", "fixture@invalid.example",
		&packet.Config{RSABits: 1024, DefaultHash: crypto.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	var keyring bytes.Buffer
	if err := entity.SerializePrivate(&keyring, nil); err != nil {
		t.Fatal(err)
	}
	private := filepath.Join(dataDir, "private", "ecos-backup")
	if err := os.MkdirAll(private, 0o700); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{"keyring.gpg": keyring.Bytes(), "passphrase": []byte("synthetic-passphrase\n")} {
		if err := os.WriteFile(filepath.Join(private, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	serial := make([]byte, 4)
	binary.LittleEndian.PutUint32(serial, 0x12345678)
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	for name, data := range map[string][]byte{
		"root/loks.bin": serial, "root/usrdata/lococonfig/01002.ecs": []byte("lococonfig 0 1 1002 29\n0:1:3\n4099:257:20\n"),
	} {
		if err := tw.WriteHeader(&tar.Header{Name: name, Typeflag: tar.TypeReg, Mode: 0o600, Size: int64(len(data))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	var encrypted bytes.Buffer
	writer, err := openpgp.Encrypt(&encrypted, []*openpgp.Entity{entity}, nil, nil,
		&packet.Config{DefaultCipher: packet.CipherAES128, DefaultHash: crypto.SHA256})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(archive.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return encrypted.Bytes()
}

func TestECoSBackupUploadRequiresAdminCSRFAndCreatesOnlyPreview(t *testing.T) {
	dataDir := t.TempDir()
	db := testRouterDBWithDataDir(t, dataDir)
	auth := application.NewAuthService(db)
	for username, role := range map[string]string{"backup-admin": "Admin", "backup-viewer": "Viewer"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
			Username: username, Password: "test-password", Roles: []string{role},
		}); err != nil {
			t.Fatal(err)
		}
	}
	settings := application.NewSettingsService(db)
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{
		ECoS: application.DigitalProviderSettings{Host: "ecos.local", Port: "15471"},
	}); err != nil {
		t.Fatal(err)
	}
	workspace := application.NewDigitalCenterWorkspaceService(infrastructure.NewDigitalCenterWorkspaceRepository(db),
		settings, nil, nil, application.NewVehicleService(db), auth)
	router := NewRouter(Config{DataDir: dataDir, AuthService: auth, DigitalCenterWorkspace: workspace})
	admin := loginRouteTestUser(t, auth, "backup-admin", "test-password")
	viewer := loginRouteTestUser(t, auth, "backup-viewer", "test-password")
	request := func(user *application.LoginResult, body []byte, contentType string, csrf bool, station string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/digital-centers/ecos/backup-import", bytes.NewReader(body))
		r.Header.Set("Content-Type", contentType)
		r.AddCookie(&http.Cookie{Name: "rk_session", Value: user.SessionToken})
		if csrf {
			r.Header.Set("X-CSRF-Token", user.CSRFToken)
		}
		if station != "" {
			r.Header.Set("X-RailKeeper-Confirm-Station", "true")
			r.Header.Set("X-RailKeeper-Backup-Station", station)
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	assertStatus(t, request(viewer, nil, "application/octet-stream", true, ""), http.StatusForbidden)
	assertStatus(t, request(admin, nil, "application/octet-stream", false, ""), http.StatusForbidden)
	assertProblem(t, request(admin, nil, "application/json", true, ""), http.StatusUnsupportedMediaType, "unsupported_media_type")
	assertProblem(t, request(admin, nil, "application/octet-stream", true, ""), http.StatusServiceUnavailable, "ecos_backup_unavailable")
	backup := syntheticAPIBackup(t, dataDir)
	assertProblem(t, request(admin, bytes.Repeat([]byte{'x'}, (8<<20)+1), "application/octet-stream", true, ""), http.StatusRequestEntityTooLarge, "ecos_backup_too_large")
	assertProblem(t, request(admin, []byte("not an ECoS backup"), "application/octet-stream", true, ""), http.StatusUnprocessableEntity, "ecos_backup_invalid")
	assertProblem(t, request(admin, backup, "application/octet-stream", true, "other.local"), http.StatusConflict, "digital_center_backup_station_mismatch")
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM digital_center_read_sessions").Scan(&count); err != nil || count != 0 {
		t.Fatal("rejected upload persisted a partial session", err)
	}
	response := request(admin, backup, "application/octet-stream", true, "")
	assertStatus(t, response, http.StatusCreated)
	if response.Header().Get("Cache-Control") != "no-store" || strings.Contains(response.Body.String(), "synthetic-passphrase") {
		t.Fatal("missing private response protection")
	}
	var result application.ECoSBackupImportResult
	decodeResponse(t, response, &result)
	if result.Summary.CVValues != 2 || result.Summary.StationSerial != 0x12345678 || result.Summary.LinkedVehicles != 0 ||
		result.Session.Capabilities != (application.DigitalCenterCapabilities{}) {
		t.Fatal(result)
	}
	page, err := workspace.ListWorkItems(t.Context(), result.Session.ID, application.DigitalCenterWorkItemFilter{})
	if err != nil || len(page.Items) != 1 || page.Items[0].VehicleID != "" || page.Items[0].Protocol != "" {
		t.Fatal("unknown backup object assigned or guessed", page, err)
	}
	for _, table := range []string{"vehicles", "vehicle_cv_values", "vehicle_external_mappings"} {
		if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil || count != 0 {
			t.Fatalf("upload wrote inventory table %s", table)
		}
	}
}

func TestECoSBackupPrivateConfigurationIsBoundedAndConfined(t *testing.T) {
	dataDir := t.TempDir()
	private := filepath.Join(dataDir, "private", "ecos-backup")
	if err := os.MkdirAll(private, 0o700); err != nil {
		t.Fatal(err)
	}
	app := &App{dataDir: dataDir}
	if err := os.WriteFile(filepath.Join(private, "keyring.gpg"), bytes.Repeat([]byte{'x'}, (64<<10)+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := app.ecosBackupPrivateConfiguration(); err == nil {
		t.Fatal("oversized private key file accepted")
	}
	if err := os.Remove(filepath.Join(private, "keyring.gpg")); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "keyring.gpg")
	if err := os.WriteFile(outside, []byte("private test"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(private, "keyring.gpg")); err != nil {
		t.Skip("host does not support symlinks")
	}
	if _, _, err := app.ecosBackupPrivateConfiguration(); err == nil {
		t.Fatal("private configuration escaped configured data directory")
	}
}
