package api

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationReadHTTPEnforcesInventoryRolesAndKeepsDeletedVehicleArchivePrivate(t *testing.T) {
	db, root := testRouterDB(t), t.TempDir()
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	router := NewRouter(Config{AuthService: auth, VehicleService: service, DataDir: root})
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Name: "Historical vehicle", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("Original CV source\n1;3\n")
	blob, err := application.NewFileBlobService(db, root).Store(t.Context(), content)
	if err != nil {
		t.Fatal(err)
	}
	file, err := service.CreateCVFile(t.Context(), vehicle.ID, application.VehicleCVFileInput{FileName: "original.csv", OriginalName: "original.csv", DecoderProfile: "Working", BlobID: blob, MimeType: "text/csv", SizeBytes: int64(len(content))})
	if err != nil {
		t.Fatal(err)
	}
	value, err := service.CreateCVValue(t.Context(), vehicle.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "Working", SourceFileID: file.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET value=-1 WHERE id=?`, value.ID); err != nil {
		t.Fatal(err)
	}
	current, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	checkpointID, missingID := strings.Repeat("a", 32), strings.Repeat("b", 32)
	if _, err := service.CaptureDecoderCVCheckpoint(t.Context(), vehicle.ID, current.Workspaces[0].ID, application.DecoderCVCheckpointInput{ID: checkpointID, ExpectedRevision: current.Revision}, root, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := service.DeleteCVFile(t.Context(), vehicle.ID, file.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET source_file_id='missing-original' WHERE id=?`, value.ID); err != nil {
		t.Fatal(err)
	}
	current, err = service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CaptureDecoderCVCheckpoint(t.Context(), vehicle.ID, current.Workspaces[0].ID, application.DecoderCVCheckpointInput{ID: missingID, ExpectedRevision: current.Revision}, root, ""); err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, root)
	before, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	contextPath := "/api/v1/vehicles/" + vehicle.ID + "/decoder-configurations"
	archive := "/api/v1/decoder-configurations/snapshots"
	originalPath := archive + "/" + checkpointID + "/sources/" + file.ID + "/download"
	missingPath := archive + "/" + missingID + "/sources/missing-original/download"
	paths := []string{contextPath, archive, archive + "?vehicleId=" + vehicle.ID, archive + "/" + checkpointID, originalPath, missingPath}
	for _, path := range paths {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		assertStatus(t, response, http.StatusUnauthorized)
	}
	var viewer *application.LoginResult
	for _, role := range []string{"Messe", "Viewer", "Editor", "Admin"} {
		if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "decoder-read-" + role, Password: "decoder-read-password", Roles: []string{role}}); err != nil {
			t.Fatal(err)
		}
		session := loginRouteTestUser(t, auth, "decoder-read-"+role, "decoder-read-password")
		if role == "Viewer" {
			viewer = session
		}
		for _, path := range paths {
			response := authedJSONRequest(t, router, session, http.MethodGet, path, nil, false)
			if role == "Messe" {
				assertStatus(t, response, http.StatusForbidden)
				continue
			}
			if path == missingPath {
				assertProblem(t, response, http.StatusGone, "decoder_source_unavailable")
				continue
			}
			assertStatus(t, response, http.StatusOK)
			if path == originalPath {
				if response.Body.String() != string(content) || !strings.Contains(response.Header().Get("Content-Disposition"), "attachment") {
					t.Fatal("private original download changed bytes or disposition")
				}
			} else if strings.Contains(response.Body.String(), `"storagePath"`) || strings.Contains(response.Body.String(), `"blobId"`) {
				t.Fatal("decoder API exposed internal storage references")
			}
		}
	}
	after, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(before.Tables, after.Tables) || !reflect.DeepEqual(before.Files, after.Files) {
		t.Fatal("reading decoder data changed the saved collection")
	}
	if err := service.Delete(t.Context(), vehicle.ID, ""); err != nil {
		t.Fatal(err)
	}
	assertProblem(t, authedJSONRequest(t, router, viewer, http.MethodGet, contextPath, nil, false), http.StatusNotFound, "decoder_not_found")
	response := authedJSONRequest(t, router, viewer, http.MethodGet, archive+"/"+checkpointID, nil, false)
	assertStatus(t, response, http.StatusOK)
	var historical application.DecoderCVSnapshot
	decodeResponse(t, response, &historical)
	if historical.VehicleReferenceID != vehicle.ID || historical.VehicleName != vehicle.Name || historical.Payload.Values[0].Value != -1 {
		t.Fatal("deleted vehicle lost historical identity or unknown values")
	}
	assertStatus(t, authedJSONRequest(t, router, viewer, http.MethodGet, originalPath, nil, false), http.StatusOK)
	assertProblem(t, authedJSONRequest(t, router, viewer, http.MethodGet, archive+"/"+checkpointID+"/sources/missing-original/download", nil, false), http.StatusNotFound, "decoder_not_found")
}
