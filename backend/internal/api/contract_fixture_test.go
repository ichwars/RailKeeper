package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"railkeeper/backend/internal/application"
)

// Each contract uses a real migrated database and production middleware, never installation data.
type contractFixture struct {
	db            *sql.DB
	dataDir       string
	auth          *application.AuthService
	vehicles      *application.VehicleService
	adminID       string
	admin, viewer *application.LoginResult
	config        Config
	router        http.Handler
}

func newContractFixture(t *testing.T) *contractFixture {
	t.Helper()
	dir := t.TempDir()
	db := testRouterDBWithDataDir(t, dir)
	f := &contractFixture{db: db, dataDir: dir, auth: application.NewAuthService(db), vehicles: application.NewVehicleService(db)}
	for _, role := range []string{"Admin", "Viewer"} {
		user, err := f.auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: role, Email: role + "@example.test", Password: "synthetic-contract-password", Roles: []string{role}})
		if err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, f.auth, role, "synthetic-contract-password")
		if role == "Admin" {
			f.admin, f.adminID = session, user.ID
		} else {
			f.viewer = session
		}
	}
	f.config = Config{DataDir: dir, AuthService: f.auth, VehicleService: f.vehicles,
		SetupService: application.NewSetupService(db), MasterDataService: application.NewMasterDataService(db),
		SettingsService: application.NewSettingsService(db), InventoryNumbers: application.NewInventoryNumberService(db),
		BackupService: application.NewBackupService(db, dir), FileBlobService: application.NewFileBlobService(db, dir),
		DatabaseMaintenance: application.NewDatabaseMaintenanceService(db, dir),
		SMTPSettingsService: application.NewSMTPSettingsService(db, application.SMTPPasswordResetMailConfig{}, ""),
		Logger:              slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	f.router = NewRouter(f.config)
	return f
}

func (f *contractFixture) vehicle(t *testing.T) string {
	t.Helper()
	v, err := f.vehicles.Create(t.Context(), application.CreateVehicleInput{Name: "Contract vehicle", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	return v.ID
}
func (f *contractFixture) call(t *testing.T, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	return authedJSONRequest(t, f.router, f.admin, method, path, body, true)
}
func (f *contractFixture) raw(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-CSRF-Token", f.admin.CSRFToken)
	r.AddCookie(&http.Cookie{Name: "rk_session", Value: f.admin.SessionToken})
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	return w
}
func (f *contractFixture) upload(t *testing.T, path, name string, data []byte, fields map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if name != "" {
		part, err := writer.CreateFormFile("file", name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = part.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	r := httptest.NewRequest(http.MethodPost, path, &body)
	r.Header.Set("Content-Type", writer.FormDataContentType())
	r.Header.Set("X-CSRF-Token", f.admin.CSRFToken)
	r.AddCookie(&http.Cookie{Name: "rk_session", Value: f.admin.SessionToken})
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, r)
	return w
}
func responseID(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()
	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil || result.ID == "" {
		t.Fatalf("missing response ID: %s (%v)", w.Body, err)
	}
	return result.ID
}
