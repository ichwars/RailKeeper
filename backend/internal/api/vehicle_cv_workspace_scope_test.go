package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationCVScopeHTTPKeepsRawLabelsAndCleansRejectedUploadBlobs(t *testing.T) {
	db, root := testRouterDB(t), t.TempDir()
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "scoped-cv", Password: "scoped-cv-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "scoped-cv", "scoped-cv-password")
	router := NewRouter(Config{AuthService: auth, VehicleService: service, FileBlobService: application.NewFileBlobService(db, root), DataDir: root})
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Name: "Scoped HTTP", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateCVValue(t.Context(), vehicle.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "Legacy"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET decoder_profile=' Legacy '; UPDATE decoder_cv_workspaces SET profile_text=' Legacy '`); err != nil {
		t.Fatal(err)
	}
	current, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	workspace := current.Workspaces[0]
	empty := ""
	input := application.VehicleCVValueInput{CVNumber: 2, Value: 8, DecoderProfile: "Legacy", WorkspaceID: workspace.ID, ExpectedWorkspaceRevision: workspace.BindingRevision, ExpectedRevision: &empty}
	prefix := "/api/v1/vehicles/" + vehicle.ID
	response := authedJSONRequest(t, router, session, http.MethodPost, prefix+"/cv-values", input, true)
	assertStatus(t, response, http.StatusCreated)
	var created application.VehicleCVValue
	decodeResponse(t, response, &created)
	if created.DecoderProfile != " Legacy " {
		t.Fatal("HTTP scope normalized the stored profile")
	}
	input.CVNumber = 3
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, prefix+"/cv-values/import", application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{input}}, true), http.StatusOK)
	upload := func(content string) *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		for key, value := range map[string]string{"decoderProfile": "Legacy", "workspaceId": workspace.ID, "expectedWorkspaceRevision": workspace.BindingRevision} {
			if err := writer.WriteField(key, value); err != nil {
				t.Fatal(err)
			}
		}
		part, err := writer.CreateFormFile("file", "scope.csv")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(http.MethodPost, prefix+"/cv-files", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		request.Header.Set("X-CSRF-Token", session.CSRFToken)
		request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}
	response = upload("CV;Value\n1;3\n")
	assertStatus(t, response, http.StatusCreated)
	var file application.VehicleCVFile
	decodeResponse(t, response, &file)
	if file.DecoderProfile != workspace.Profile {
		t.Fatal("upload inferred another group")
	}
	if _, err := db.Exec(`UPDATE decoder_cv_workspaces SET profile_text=' Changed '`); err != nil {
		t.Fatal(err)
	}
	backup := application.NewBackupService(db, root)
	before, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, upload("Different source that must be removed after rejection"), http.StatusConflict, "cv_file_conflict")
	input.CVNumber = 4
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, prefix+"/cv-values", input, true), http.StatusConflict, "cv_value_conflict")
	after, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(before.Tables, after.Tables) || !reflect.DeepEqual(before.Files, after.Files) {
		t.Fatal("rejected scope kept an orphan blob or changed CV/source data")
	}
}
