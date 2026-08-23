package api

import (
	"bytes"
	"context"
	"database/sql"
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
	var createdProfile application.DataTransferProfile
	decodeResponse(t, editorCreate, &createdProfile)
	if bytes.Contains(editorCreate.Body.Bytes(), []byte(`"lastUsedAt"`)) {
		t.Fatalf("new profile response unexpectedly contains lastUsedAt: %s", editorCreate.Body.String())
	}
	updatedProfile := layoutRequest(t, router, sessions["editor"], http.MethodPut,
		"/api/v1/data-transfer/profiles/"+createdProfile.ID,
		validDataTransferProfileRequest("Editor updated"), true)
	assertStatus(t, updatedProfile, http.StatusOK)
	disabledProfile := layoutRequest(t, router, sessions["admin"], http.MethodDelete,
		"/api/v1/data-transfer/profiles/"+createdProfile.ID, nil, true)
	assertStatus(t, disabledProfile, http.StatusNoContent)

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
	assertDataTransferAuditActions(t, db,
		"DataTransferProfileCreated", "DataTransferProfileUpdated", "DataTransferProfileDisabled")
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
		[]byte("Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n"+
			"RK-001;;BR 218;H0;Lokomotive;Diesellokomotive\n"))
	assertStatus(t, uploaded, http.StatusOK)
	var preview application.DataTransferPreview
	decodeResponse(t, uploaded, &preview)
	if preview.ErrorRecords != 1 || len(preview.Issues) != 1 {
		t.Fatalf("unexpected import preview: %#v", preview)
	}
	if len(preview.CSVMapping) != 6 || len(preview.VehicleFields) != 62 ||
		preview.CSVMapping[0].SourceHeader != "Inventarnummer" {
		t.Fatalf("missing CSV mapping contract: %#v, fields=%d", preview.CSVMapping, len(preview.VehicleFields))
	}

	uploaded = dataTransferMultipartMappingRequest(t, router, editor,
		"/api/v1/data-transfer/jobs/"+job.ID+"/upload", "vehicles.csv",
		[]byte("Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n"+
			"RK-001;;BR 218;H0;Lokomotive;Diesellokomotive\n"),
		application.DataTransferCSVMappingInput{Columns: preview.CSVMapping, SaveToProfile: true})
	assertStatus(t, uploaded, http.StatusOK)
	decodeResponse(t, uploaded, &preview)

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
	retried := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/retry", nil, true)
	assertStatus(t, retried, http.StatusCreated)
	assertDataTransferAuditActions(t, db, "DataTransferImportJobCreated", "DataTransferImportUploaded",
		"DataTransferImportUploaded",
		"DataTransferIssueResolved", "DataTransferJobCancelled", "DataTransferJobRetried")
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
		[]byte("Inventarnummer;Hersteller;Bezeichnung;Spurweite;Kategorie;Gattung\n"+
			"RK-CONFIRM;Roco;BR 218;H0;Lokomotive;Diesellokomotive\n"))
	assertStatus(t, uploaded, http.StatusOK)
	var preview application.DataTransferPreview
	decodeResponse(t, uploaded, &preview)

	rejected := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/confirm",
		map[string]any{"confirm": false, "expectedRevision": preview.Job.Revision}, true)
	assertProblem(t, rejected, http.StatusBadRequest, "data_transfer_validation")

	missingRevision := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/confirm", map[string]any{"confirm": true}, true)
	assertProblem(t, missingRevision, http.StatusBadRequest, "data_transfer_validation")

	staleRevision := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/confirm",
		map[string]any{"confirm": true, "expectedRevision": preview.Job.Revision - 1}, true)
	assertProblem(t, staleRevision, http.StatusConflict, "data_transfer_conflict")

	confirmed := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+job.ID+"/confirm",
		map[string]any{"confirm": true, "expectedRevision": preview.Job.Revision}, true)
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
	var exhibitionJob application.DataTransferJob
	decodeResponse(t, allowed, &exhibitionJob)
	payload, err := json.Marshal(application.DataTransferPackage{
		Format: application.DataTransferPackageFormat, Version: application.DataTransferPackageVersion,
		CreatedAt: "2026-08-21T00:00:00Z", Areas: application.DataTransferPackageAreas{
			ExhibitionLists: []application.TransferExhibitionList{{
				Designation: "New Messe list", Date: "2026-08-21", Entries: []application.TransferExhibitionEntry{},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	uploaded := dataTransferMultipartRequest(t, router, messe,
		"/api/v1/data-transfer/jobs/"+exhibitionJob.ID+"/upload", "exhibition.json", payload)
	assertStatus(t, uploaded, http.StatusOK)
	var preview application.DataTransferPreview
	decodeResponse(t, uploaded, &preview)
	createList := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+exhibitionJob.ID+"/confirm",
		map[string]any{"confirm": true, "expectedRevision": preview.Job.Revision}, true)
	assertProblem(t, createList, http.StatusForbidden, "data_transfer_forbidden")
}

func TestDataTransferMesseImportCannotReplaceListHeadersOrDeleteEntries(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "messe-merge", Password: "messe-password", Roles: []string{"Messe"},
	}); err != nil {
		t.Fatal(err)
	}
	const now = "2026-08-21T00:00:00Z"
	if _, err := db.Exec(`INSERT INTO exhibition_lists(
		id, designation, list_date, locked, created_at, updated_at
	) VALUES('messe-list', 'Clubtag', '2026-08-21', 0, ?, ?)`, now, now); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO exhibition_entries(
		id, list_id, owner, locomotive_name, day_scope, notes, sort_order, created_at, updated_at
	) VALUES('existing-entry', 'messe-list', 'Club', 'BR 01', 'all', 'Local note', 1, ?, ?)`, now, now); err != nil {
		t.Fatal(err)
	}
	service := application.NewDataTransferService(infrastructure.NewDataTransferRepository(db), t.TempDir())
	profile, err := service.CreateProfile(t.Context(), application.CreateDataTransferProfileInput{
		Name: "Exhibition", Direction: application.TransferImport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists},
	}, "admin-1")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	messe := loginRouteTestUser(t, auth, "messe-merge", "messe-password")
	payload, err := json.Marshal(application.DataTransferPackage{
		Format: application.DataTransferPackageFormat, Version: application.DataTransferPackageVersion,
		CreatedAt: now, Areas: application.DataTransferPackageAreas{
			ExhibitionLists: []application.TransferExhibitionList{{
				ID: "messe-list", Designation: "Clubtag", Date: "2026-08-21", Locked: true,
				Entries: []application.TransferExhibitionEntry{
					{ID: "existing-entry", Owner: "Club", LocomotiveName: "BR 01", DayScope: "all", Notes: "Imported note", SortOrder: 1},
					{ID: "new-entry", Owner: "Guest", LocomotiveName: "BR 03", DayScope: "all", SortOrder: 2},
				},
			}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	createAndResolve := func(t *testing.T, resolution string) application.DataTransferJob {
		t.Helper()
		created := layoutRequest(t, router, messe, http.MethodPost, "/api/v1/data-transfer/jobs/import",
			map[string]any{"profileId": profile.ID}, true)
		assertStatus(t, created, http.StatusCreated)
		var job application.DataTransferJob
		decodeResponse(t, created, &job)
		uploaded := dataTransferMultipartRequest(t, router, messe,
			"/api/v1/data-transfer/jobs/"+job.ID+"/upload", "exhibition.json", payload)
		assertStatus(t, uploaded, http.StatusOK)
		var preview application.DataTransferPreview
		decodeResponse(t, uploaded, &preview)
		if len(preview.Issues) != 1 || preview.Issues[0].Code != "duplicate_exhibition_list" {
			t.Fatalf("unexpected exhibition preview issues: %#v", preview.Issues)
		}
		resolved := layoutRequest(t, router, messe, http.MethodPut,
			"/api/v1/data-transfer/jobs/"+job.ID+"/issues/"+preview.Issues[0].ID,
			map[string]any{"resolution": resolution}, true)
		assertStatus(t, resolved, http.StatusOK)
		decodeResponse(t, resolved, &job)
		return job
	}

	replaceJob := createAndResolve(t, "replace")
	replace := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+replaceJob.ID+"/confirm",
		map[string]any{"confirm": true, "expectedRevision": replaceJob.Revision}, true)
	assertProblem(t, replace, http.StatusForbidden, "data_transfer_forbidden")

	mergeJob := createAndResolve(t, "merge")
	merge := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+mergeJob.ID+"/confirm",
		map[string]any{"confirm": true, "expectedRevision": mergeJob.Revision}, true)
	assertStatus(t, merge, http.StatusOK)

	var locked, entries int
	var note string
	if err := db.QueryRow(`SELECT locked FROM exhibition_lists WHERE id='messe-list'`).Scan(&locked); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM exhibition_entries WHERE list_id='messe-list'`).Scan(&entries); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT notes FROM exhibition_entries WHERE id='existing-entry'`).Scan(&note); err != nil {
		t.Fatal(err)
	}
	if locked != 0 || entries != 2 || note != "Imported note" {
		t.Fatalf("Messe merge changed forbidden list state: locked=%d entries=%d note=%q", locked, entries, note)
	}
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
			preview, err := service.UploadAndPreview(t.Context(), confirmJob.ID, "vehicles.csv", []byte(
				"Inventarnummer;Hersteller;Bezeichnung;Spurweite\nRK-CONFIRM-"+username+";Roco;BR 01;H0\n",
			), "editor-1")
			if err != nil {
				t.Fatal(err)
			}
			forbiddenConfirm := layoutRequest(t, router, session, http.MethodPost,
				"/api/v1/data-transfer/jobs/"+confirmJob.ID+"/confirm",
				map[string]any{"confirm": true, "expectedRevision": preview.Job.Revision}, true)
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

func dataTransferMultipartMappingRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	path string,
	filename string,
	payload []byte,
	mapping application.DataTransferCSVMappingInput,
) *httptest.ResponseRecorder {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	mappingPayload, err := json.Marshal(mapping)
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("mapping", string(mappingPayload)); err != nil {
		t.Fatal(err)
	}
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
	assertDataTransferAuditActions(t, db,
		"DataTransferExportJobCreated", "DataTransferExportExecuted", "DataTransferArtifactDeleted")
}

func assertDataTransferAuditActions(t *testing.T, db *sql.DB, expected ...string) {
	t.Helper()
	rows, err := db.Query(`SELECT action FROM audit_logs WHERE action LIKE 'DataTransfer%'`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	actual := map[string]bool{}
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			t.Fatal(err)
		}
		actual[action] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	for _, action := range expected {
		if !actual[action] {
			t.Errorf("missing data-transfer audit action %q in %#v", action, actual)
		}
	}
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

func TestDataTransferQueryAndRetryRoutesReturnScopedPersistentHistory(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "viewer-query", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-query", Password: "messe-password", Roles: []string{"Messe"}},
		{Username: "editor-query", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "admin-query", Password: "admin-password", Roles: []string{"Admin"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	repository := infrastructure.NewDataTransferRepository(db)
	service := application.NewDataTransferService(repository, t.TempDir())
	vehicleJob, err := repository.CreateJob(t.Context(), application.DataTransferJob{
		ProfileName: "Vehicles", Direction: application.TransferImport, Format: application.TransferCSV,
		Areas: []application.TransferArea{application.TransferVehicles}, State: application.TransferJobCompleted,
		Stage: "completed", TotalRecords: 2, CompletedAt: "2026-08-20T10:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	exhibitionJob, err := repository.CreateJob(t.Context(), application.DataTransferJob{
		ProfileName: "Exhibition", Direction: application.TransferExport, Format: application.TransferJSON,
		Areas: []application.TransferArea{application.TransferExhibitionLists}, State: application.TransferJobCompleted,
		Stage: "completed", TotalRecords: 3, CompletedAt: "2026-08-20T11:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.ReplaceIssues(t.Context(), vehicleJob.ID, []application.DataTransferIssue{{
		JobID: vehicleJob.ID, Area: application.TransferVehicles, Severity: application.TransferIssueWarning,
		Code: "warning", Message: "reviewed",
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.CreateArtifact(t.Context(), application.DataTransferArtifact{
		JobID: exhibitionJob.ID, RelativePath: "exports/exhibition.json", DisplayName: "exhibition.json",
		MIMEType: "application/json", SizeBytes: 13, SHA256: "hash",
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, DataTransferService: service})
	viewer := loginRouteTestUser(t, auth, "viewer-query", "viewer-password")
	messe := loginRouteTestUser(t, auth, "messe-query", "messe-password")
	editor := loginRouteTestUser(t, auth, "editor-query", "editor-password")
	admin := loginRouteTestUser(t, auth, "admin-query", "admin-password")

	summaryResponse := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/data-transfer/summary", nil, true)
	assertStatus(t, summaryResponse, http.StatusOK)
	var summary application.DataTransferSummary
	decodeResponse(t, summaryResponse, &summary)
	if summary.LastExportAt != exhibitionJob.CompletedAt || summary.ArtifactCount != 1 || summary.ArtifactBytes != 13 {
		t.Fatalf("unexpected query summary: %#v", summary)
	}

	listedResponse := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/data-transfer/jobs?direction=import&states=completed&limit=1", nil, true)
	assertStatus(t, listedResponse, http.StatusOK)
	var jobs []application.DataTransferJob
	decodeResponse(t, listedResponse, &jobs)
	if len(jobs) != 1 || jobs[0].ID != vehicleJob.ID {
		t.Fatalf("filtered jobs = %#v", jobs)
	}
	invalidLimit := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/data-transfer/jobs?limit=0", nil, true)
	assertProblem(t, invalidLimit, http.StatusBadRequest, "data_transfer_validation")

	detailResponse := layoutRequest(t, router, viewer, http.MethodGet,
		"/api/v1/data-transfer/jobs/"+vehicleJob.ID, nil, true)
	assertStatus(t, detailResponse, http.StatusOK)
	var details application.DataTransferJobDetails
	decodeResponse(t, detailResponse, &details)
	if details.Job.ID != vehicleJob.ID || len(details.Issues) != 1 {
		t.Fatalf("job details = %#v", details)
	}

	retryResponse := layoutRequest(t, router, viewer, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+vehicleJob.ID+"/retry", nil, true)
	assertProblem(t, retryResponse, http.StatusForbidden, "forbidden")
	retryResponse = layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+vehicleJob.ID+"/retry", nil, true)
	assertStatus(t, retryResponse, http.StatusCreated)
	var retry application.DataTransferJob
	decodeResponse(t, retryResponse, &retry)
	if retry.ID == vehicleJob.ID || retry.State != application.TransferJobDraft || retry.ConfirmedAt != "" {
		t.Fatalf("retry = %#v", retry)
	}
	adminRetry := layoutRequest(t, router, admin, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+vehicleJob.ID+"/retry", nil, true)
	assertStatus(t, adminRetry, http.StatusCreated)

	messeJobsResponse := layoutRequest(t, router, messe, http.MethodGet,
		"/api/v1/data-transfer/jobs", nil, true)
	assertStatus(t, messeJobsResponse, http.StatusOK)
	decodeResponse(t, messeJobsResponse, &jobs)
	if len(jobs) != 1 || jobs[0].ID != exhibitionJob.ID {
		t.Fatalf("messe jobs = %#v", jobs)
	}
	forbiddenDetail := layoutRequest(t, router, messe, http.MethodGet,
		"/api/v1/data-transfer/jobs/"+vehicleJob.ID, nil, true)
	assertProblem(t, forbiddenDetail, http.StatusForbidden, "data_transfer_forbidden")
	messeForbiddenRetry := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+vehicleJob.ID+"/retry", nil, true)
	assertProblem(t, messeForbiddenRetry, http.StatusForbidden, "data_transfer_forbidden")
	messeRetry := layoutRequest(t, router, messe, http.MethodPost,
		"/api/v1/data-transfer/jobs/"+exhibitionJob.ID+"/retry", nil, true)
	assertStatus(t, messeRetry, http.StatusCreated)
}

func validDataTransferProfileRequest(name string) map[string]any {
	return map[string]any{
		"name": name, "direction": "export", "format": "csv", "areas": []string{"vehicles"},
	}
}
