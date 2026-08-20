package api

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDataTransferProfileRoutesEnforceRolesAndMesseScope(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	if _, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Exhibition", Direction: application.TransferExport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists},
	}, "admin-1"); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	sessions := map[string]*application.LoginResult{
		"admin":  loginRouteTestUser(t, auth, "admin", "admin-password"),
		"editor": loginRouteTestUser(t, auth, "editor", "editor-password"),
		"viewer": loginRouteTestUser(t, auth, "viewer", "viewer-password"),
		"messe":  loginRouteTestUser(t, auth, "messe", "messe-password"),
	}

	viewerProfiles := layoutRequest(t, router, sessions["viewer"], http.MethodGet,
		"/api/v1/data-transfer/profiles", nil, true)
	assertStatus(t, viewerProfiles, http.StatusOK)
	var listed []application.DataTransferProfile
	decodeResponse(t, viewerProfiles, &listed)
	if len(listed) != 2 {
		t.Fatalf("viewer profiles = %#v, want all profiles", listed)
	}

	viewerCreate := layoutRequest(t, router, sessions["viewer"], http.MethodPost,
		"/api/v1/data-transfer/profiles", validDataTransferProfileRequest("Viewer"), true)
	assertProblem(t, viewerCreate, http.StatusForbidden, "forbidden")

	editorCreate := layoutRequest(t, router, sessions["editor"], http.MethodPost,
		"/api/v1/data-transfer/profiles", validDataTransferProfileRequest("Editor"), true)
	assertStatus(t, editorCreate, http.StatusCreated)
	var created map[string]json.RawMessage
	decodeResponse(t, editorCreate, &created)
	if _, found := created["lastUsedAt"]; found {
		t.Fatalf("new profile response unexpectedly contains lastUsedAt: %s", editorCreate.Body.String())
	}

	messeProfiles := layoutRequest(t, router, sessions["messe"], http.MethodGet,
		"/api/v1/data-transfer/profiles", nil, true)
	assertStatus(t, messeProfiles, http.StatusOK)
	decodeResponse(t, messeProfiles, &listed)
	if len(listed) != 1 || len(listed[0].Areas) != 1 || listed[0].Areas[0] != application.TransferExhibitionLists {
		t.Fatalf("messe profiles = %#v, want only exhibition-list profiles", listed)
	}

	invalid := layoutRequest(t, router, sessions["editor"], http.MethodPost, "/api/v1/data-transfer/profiles",
		map[string]any{"name": "Master data", "direction": "export", "format": "railkeeper-json",
			"areas": []string{"masterData"}}, true)
	assertProblem(t, invalid, http.StatusBadRequest, "data_transfer_validation")
}

func TestDataTransferImportRoutesUploadResolveAndCancelPersistentPreview(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "editor-import", Password: "editor-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicle import", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	editor := loginRouteTestUser(t, auth, "editor-import", "editor-password")

	created := layoutRequest(t, router, editor, http.MethodPost, "/api/v1/data-transfer/jobs/import",
		map[string]any{"profileId": profile.ID}, true)
	assertStatus(t, created, http.StatusCreated)
	var job application.DataTransferJob
	decodeResponse(t, created, &job)

	uploaded := dataTransferMultipartRequest(t, router, editor,
		"/api/v1/data-transfer/jobs/"+job.ID+"/upload", "vehicles.csv",
		[]byte("Inventarnummer;Hersteller;Bezeichnung\nRK-001;;BR 218\n"))
	assertStatus(t, uploaded, http.StatusOK)
	var preview application.DataTransferPreview
	decodeResponse(t, uploaded, &preview)
	if preview.ErrorRecords != 1 || len(preview.Issues) != 1 {
		t.Fatalf("unexpected import preview: %#v", preview)
	}

	resolved := layoutRequest(t, router, editor, http.MethodPut,
		"/api/v1/data-transfer/jobs/"+job.ID+"/issues/"+preview.Issues[0].ID,
		map[string]any{"resolution": "skip"}, true)
	assertStatus(t, resolved, http.StatusOK)
	decodeResponse(t, resolved, &job)
	if job.State != application.TransferJobReady {
		t.Fatalf("resolved job state = %q", job.State)
	}

	cancelled := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/cancel", nil, true)
	assertStatus(t, cancelled, http.StatusOK)
	decodeResponse(t, cancelled, &job)
	if job.State != application.TransferJobCancelled {
		t.Fatalf("cancelled job state = %q", job.State)
	}
}

func TestDataTransferConfirmRouteRequiresTrueAndAppliesReadyPreview(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "editor-confirm", Password: "editor-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicle import", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	editor := loginRouteTestUser(t, auth, "editor-confirm", "editor-password")
	job, err := service.CreateImportJob(t.Context(), profile.ID, "editor-confirm")
	if err != nil {
		t.Fatal(err)
	}
	uploaded := dataTransferMultipartRequest(t, router, editor,
		"/api/v1/data-transfer/jobs/"+job.ID+"/upload", "vehicles.csv",
		[]byte("Inventarnummer;Hersteller;Bezeichnung;Spurweite\nRK-CONFIRM;Roco;BR 218;H0\n"))
	assertStatus(t, uploaded, http.StatusOK)

	rejected := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/confirm", map[string]any{"confirm": false}, true)
	assertProblem(t, rejected, http.StatusBadRequest, "data_transfer_validation")

	confirmed := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/confirm", map[string]any{"confirm": true}, true)
	assertStatus(t, confirmed, http.StatusOK)
	decodeResponse(t, confirmed, &job)
	if job.State != application.TransferJobCompleted {
		t.Fatalf("confirmed job state = %q", job.State)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM vehicles WHERE inventory_number='RK-CONFIRM'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("confirmed vehicle count = %d, want 1", count)
	}
}

func TestDataTransferImportRoutesEnforceMesseAreaScope(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "messe-import", Password: "messe-password", Roles: []string{"Messe"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	vehicleProfile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	exhibitionProfile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Exhibition", Direction: application.TransferImport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	messe := loginRouteTestUser(t, auth, "messe-import", "messe-password")

	forbidden := layoutRequest(t, router, messe, http.MethodPost, "/api/v1/data-transfer/jobs/import",
		map[string]any{"profileId": vehicleProfile.ID}, true)
	assertProblem(t, forbidden, http.StatusForbidden, "data_transfer_forbidden")
	allowed := layoutRequest(t, router, messe, http.MethodPost, "/api/v1/data-transfer/jobs/import",
		map[string]any{"profileId": exhibitionProfile.ID}, true)
	assertStatus(t, allowed, http.StatusCreated)
}

func TestDataTransferImportWriteScopeKeepsMesseViewerAndPlannerExhibitionOnly(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	users := []application.CreateUserInput{
		{Username: "messe-viewer", Password: "user-password", Roles: []string{"Messe", "Viewer"}},
		{Username: "messe-planner", Password: "user-password", Roles: []string{"Messe", "Planner"}},
		{Username: "messe-editor", Password: "user-password", Roles: []string{"Messe", "Editor"}},
		{Username: "messe-admin", Password: "user-password", Roles: []string{"Messe", "Admin"}},
	}
	for _, user := range users {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	service := application.NewDataTransferService(repository, t.TempDir())
	vehicleProfile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	exhibitionProfile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Exhibition", Direction: application.TransferImport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})

	for _, username := range []string{"messe-viewer", "messe-planner"} {
		t.Run(username, func(t *testing.T) {
			session := loginRouteTestUser(t, auth, username, "user-password")
			forbiddenCreate := layoutRequest(t, router, session, http.MethodPost,
				"/api/v1/data-transfer/jobs/import", map[string]any{"profileId": vehicleProfile.ID}, true)
			assertProblem(t, forbiddenCreate, http.StatusForbidden, "data_transfer_forbidden")
			allowedCreate := layoutRequest(t, router, session, http.MethodPost,
				"/api/v1/data-transfer/jobs/import", map[string]any{"profileId": exhibitionProfile.ID}, true)
			assertStatus(t, allowedCreate, http.StatusCreated)

			uploadJob, err := service.CreateImportJob(t.Context(), vehicleProfile.ID, "editor-1")
			if err != nil {
				t.Fatal(err)
			}
			forbiddenUpload := dataTransferMultipartRequest(t, router, session,
				"/api/v1/data-transfer/jobs/"+uploadJob.ID+"/upload", "vehicles.csv",
				[]byte("Inventarnummer;Hersteller;Bezeichnung\nRK-1;Roco;BR 01\n"))
			assertProblem(t, forbiddenUpload, http.StatusForbidden, "data_transfer_forbidden")

			resolveJob, err := service.CreateImportJob(t.Context(), vehicleProfile.ID, "editor-1")
			if err != nil {
				t.Fatal(err)
			}
			resolveJob.State = application.TransferJobReviewRequired
			if resolveJob, err = repository.UpdateJob(t.Context(), resolveJob); err != nil {
				t.Fatal(err)
			}
			if err := repository.ReplaceIssues(t.Context(), resolveJob.ID, []application.DataTransferIssue{{
				JobID: resolveJob.ID, Area: application.TransferVehicles, RecordKey: "RK-1",
				Severity: application.TransferIssueError, Code: "missing_manufacturer",
				ProposedResolution: "skip",
			}}); err != nil {
				t.Fatal(err)
			}
			issues, err := repository.ListIssues(t.Context(), resolveJob.ID)
			if err != nil {
				t.Fatal(err)
			}
			forbiddenResolve := layoutRequest(t, router, session, http.MethodPut,
				"/api/v1/data-transfer/jobs/"+resolveJob.ID+"/issues/"+issues[0].ID,
				map[string]any{"resolution": "skip"}, true)
			assertProblem(t, forbiddenResolve, http.StatusForbidden, "data_transfer_forbidden")

			cancelJob, err := service.CreateImportJob(t.Context(), vehicleProfile.ID, "editor-1")
			if err != nil {
				t.Fatal(err)
			}
			forbiddenCancel := layoutRequest(t, router, session, http.MethodPost,
				"/api/v1/data-transfer/jobs/"+cancelJob.ID+"/cancel", nil, true)
			assertProblem(t, forbiddenCancel, http.StatusForbidden, "data_transfer_forbidden")

			confirmJob, err := service.CreateImportJob(t.Context(), vehicleProfile.ID, "editor-1")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := service.UploadAndPreview(t.Context(), confirmJob.ID, "vehicles.csv", []byte(
				"Inventarnummer;Hersteller;Bezeichnung;Spurweite\nRK-CONFIRM-"+username+";Roco;BR 01;H0\n",
			), "editor-1"); err != nil {
				t.Fatal(err)
			}
			forbiddenConfirm := layoutRequest(t, router, session, http.MethodPost,
				"/api/v1/data-transfer/jobs/"+confirmJob.ID+"/confirm", map[string]any{"confirm": true}, true)
			assertProblem(t, forbiddenConfirm, http.StatusForbidden, "data_transfer_forbidden")
		})
	}

	for _, username := range []string{"messe-editor", "messe-admin"} {
		t.Run(username, func(t *testing.T) {
			session := loginRouteTestUser(t, auth, username, "user-password")
			created := layoutRequest(t, router, session, http.MethodPost,
				"/api/v1/data-transfer/jobs/import", map[string]any{"profileId": vehicleProfile.ID}, true)
			assertStatus(t, created, http.StatusCreated)
		})
	}
}

func dataTransferMultipartRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	path string,
	filename string,
	payload []byte,
) *httptest.ResponseRecorder {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, path, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("X-CSRF-Token", session.CSRFToken)
	request.AddCookie(&http.Cookie{Name: "rk_session", Value: session.SessionToken})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestDataTransferExportRoutesCreateExecuteDownloadAndDeleteArtifact(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin-export", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "viewer-export", Password: "viewer-password", Roles: []string{"Viewer"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	dataDir := t.TempDir()
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), dataDir)
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	viewer := loginRouteTestUser(t, auth, "viewer-export", "viewer-password")
	admin := loginRouteTestUser(t, auth, "admin-export", "admin-password")

	createdResponse := layoutRequest(t, router, viewer, http.MethodPost,
		"/api/v1/data-transfer/jobs/export", map[string]any{"profileId": profile.ID}, true)
	assertStatus(t, createdResponse, http.StatusCreated)
	var job application.DataTransferJob
	decodeResponse(t, createdResponse, &job)

	executedResponse := layoutRequest(t, router, viewer, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/execute", nil, true)
	assertStatus(t, executedResponse, http.StatusOK)
	var result application.DataTransferExportResult
	decodeResponse(t, executedResponse, &result)
	if result.Job.State != application.TransferJobCompleted || result.OpenFolderAvailable {
		t.Fatalf("unexpected export response: %#v", result)
	}
	conflict := layoutRequest(t, router, viewer, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/execute", nil, true)
	assertProblem(t, conflict, http.StatusConflict, "data_transfer_conflict")

	download := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/data-transfer/artifacts/"+result.Artifact.ID+"/download", nil, true)
	assertStatus(t, download, http.StatusOK)
	if download.Header().Get("Content-Disposition") == "" || download.Body.Len() == 0 {
		t.Fatalf("download lacks headers or content: headers=%v body=%q", download.Header(), download.Body.String())
	}

	deleted := layoutRequest(t, router, admin, http.MethodDelete,
		"/api/v1/data-transfer/artifacts/"+result.Artifact.ID, nil, true)
	assertStatus(t, deleted, http.StatusNoContent)
	matches, err := filepath.Glob(filepath.Join(dataDir, "exports", "*.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("deleted artifact still exists: %v", matches)
	}
	download = layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/data-transfer/artifacts/"+result.Artifact.ID+"/download", nil, true)
	assertProblem(t, download, http.StatusGone, "data_transfer_artifact_deleted")
}

func TestDataTransferExportRoutesEnforceMesseAreaScopeBeforeCreatingJob(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "messe-export", Password: "messe-password", Roles: []string{"Messe"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	vehicleProfile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	exhibitionProfile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Exhibition", Direction: application.TransferExport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	messe := loginRouteTestUser(t, auth, "messe-export", "messe-password")

	forbidden := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/export", map[string]any{"profileId": vehicleProfile.ID}, true)
	assertProblem(t, forbidden, http.StatusForbidden, "data_transfer_forbidden")
	allowed := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/export", map[string]any{"profileId": exhibitionProfile.ID}, true)
	assertStatus(t, allowed, http.StatusCreated)

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM data_transfer_jobs`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("forbidden request persisted a job, count = %d", count)
	}
}

func TestDataTransferExportOpenFolderUsesConfinedExportDirectory(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "admin-folder", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	dataDir := t.TempDir()
	openedPath := ""
	service := application.NewDataTransferService(
		infrastructure.NewDataTransferRepository(db), dataDir,
		application.WithDataTransferFolderOpener(true, func(_ context.Context, path string) error {
			openedPath = path
			return nil
		}),
	)
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	admin := loginRouteTestUser(t, auth, "admin-folder", "admin-password")
	response := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/data-transfer/artifacts/open-folder", nil, true)
	assertStatus(t, response, http.StatusNoContent)
	if openedPath != filepath.Join(dataDir, "exports") {
		t.Fatalf("opened path = %q, want confined exports directory", openedPath)
	}
}

func TestDataTransferProfileRoutesUpdateAndDisableProfiles(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "admin", Password: "admin-password", Roles: []string{"Admin"},
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Vehicles", Direction: application.TransferExport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	session := loginRouteTestUser(t, auth, "admin", "admin-password")

	updated := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/data-transfer/profiles/"+profile.ID,
		map[string]any{"name": "Vehicles and accessories", "direction": "export", "format": "railkeeper-json",
			"areas": []string{"vehicles", "accessories"}}, true)
	assertStatus(t, updated, http.StatusOK)

	disabled := layoutRequest(t, router, session, http.MethodDelete,
		"/api/v1/data-transfer/profiles/"+profile.ID, nil, true)
	assertStatus(t, disabled, http.StatusNoContent)
	profiles, err := service.ListProfiles(t.Context())
	if err != nil || len(profiles) != 1 || profiles[0].Enabled {
		t.Fatalf("unexpected disabled profile state: %#v, %v", profiles, err)
	}
}

func validDataTransferProfileRequest(name string) map[string]any {
	return map[string]any{
		"name": name, "direction": "export", "format": "csv", "areas": []string{"vehicles"},
	}
}
