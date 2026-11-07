package api

import (
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestDecoderConfigurationDeletionScopeHTTPPreservesOriginalsAndRejectsIncompleteComparisons(t *testing.T) {
	db, root := testRouterDB(t), t.TempDir()
	auth, service := application.NewAuthService(db), application.NewVehicleService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "delete-scope", Password: "synthetic-delete-scope-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "delete-scope", "synthetic-delete-scope-password")
	blobs := application.NewFileBlobService(db, root)
	router := NewRouter(Config{AuthService: auth, VehicleService: service, FileBlobService: blobs, DataDir: root})
	vehicle, err := service.Create(t.Context(), application.CreateVehicleInput{Name: "Delete scope", Manufacturer: "Test", Gauge: "H0", Category: "Lokomotive", Gattung: "Diesellok"}, "")
	if err != nil {
		t.Fatal(err)
	}
	first, err := service.CreateCVValue(t.Context(), vehicle.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "Legacy"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE vehicle_cv_values SET decoder_profile=' Legacy '; UPDATE decoder_cv_workspaces SET profile_text=' Legacy '`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateCVValue(t.Context(), vehicle.ID, application.VehicleCVValueInput{CVNumber: 1, Value: 8, DecoderProfile: "Legacy"}); err != nil {
		t.Fatal(err)
	}
	current, err := service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	var group, wrong application.DecoderCVWorkspace
	for _, item := range current.Workspaces {
		if item.Profile == " Legacy " {
			group = item
		} else {
			wrong = item
		}
	}
	content := []byte("CV;Value\n1;3\n")
	blob, err := blobs.Store(t.Context(), content)
	if err != nil {
		t.Fatal(err)
	}
	file, err := service.CreateCVFile(t.Context(), vehicle.ID, application.VehicleCVFileInput{FileName: "original.csv", OriginalName: "original.csv",
		DecoderProfile: group.Profile, WorkspaceID: group.ID, ExpectedWorkspaceRevision: group.BindingRevision, BlobID: blob})
	if err != nil {
		t.Fatal(err)
	}
	values, err := service.ListCVValues(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range values {
		if value.ID == first.ID {
			*first = value
		}
	}
	query := func(revision string, item application.DecoderCVWorkspace) string {
		return "?" + url.Values{"expectedRevision": {revision}, "workspaceId": {item.ID}, "expectedWorkspaceRevision": {item.BindingRevision}}.Encode()
	}
	prefix := "/api/v1/vehicles/" + vehicle.ID
	rowPath, filePath := prefix+"/cv-values/"+first.ID, prefix+"/cv-files/"+file.ID
	backup := application.NewBackupService(db, root)
	before, err := backup.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []string{"?workspaceId=" + group.ID, "?expectedWorkspaceRevision=" + group.BindingRevision,
		"?workspaceId=&expectedWorkspaceRevision=", query(first.Revision, group) + "&workspaceId=duplicate", query(first.Revision, wrong)} {
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, rowPath+bad, nil, true), http.StatusBadRequest)
		assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, filePath+bad, nil, true), http.StatusBadRequest)
	}
	after, err := backup.Export(t.Context())
	if err != nil || !reflect.DeepEqual(before.Tables, after.Tables) || !reflect.DeepEqual(before.Files, after.Files) {
		t.Fatal("invalid comparisons changed live or frozen data")
	}
	current, err = service.DecoderConfigurationContext(t.Context(), vehicle.ID)
	if err != nil {
		t.Fatal(err)
	}
	current, err = service.CreateDecoderInstance(t.Context(), vehicle.ID, application.DecoderInstanceInput{ID: strings.Repeat("a", 32), ExpectedRevision: current.Revision, Label: "Individual unit"}, "")
	if err != nil {
		t.Fatal(err)
	}
	current, err = service.InstallDecoder(t.Context(), vehicle.ID, application.DecoderInstallationInput{ID: strings.Repeat("b", 32), DecoderID: strings.Repeat("a", 32), ExpectedRevision: current.Revision}, "")
	if err != nil {
		t.Fatal(err)
	}
	installation := strings.Repeat("b", 32)
	current, err = service.AssignDecoderCVWorkspace(t.Context(), vehicle.ID, group.ID, application.DecoderCVAssignmentInput{ExpectedRevision: current.Revision,
		InstallationID: &installation, SnapshotID: strings.Repeat("c", 32)}, root, "")
	if err != nil {
		t.Fatal(err)
	}
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, rowPath+query(first.Revision, group), nil, true), http.StatusConflict, "cv_value_conflict")
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodDelete, filePath+query(file.Revision, group), nil, true), http.StatusConflict, "cv_file_conflict")
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, prefix+"/cv-values/changes", application.VehicleCVChangesInput{Changes: []application.VehicleCVChange{
		{ID: first.ID, ExpectedRevision: first.Revision, Delete: true, WorkspaceID: group.ID, ExpectedWorkspaceRevision: group.BindingRevision},
	}}, true), http.StatusConflict)
	for _, item := range current.Workspaces {
		if item.ID == group.ID {
			group = item
		}
	}
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, rowPath+query(first.Revision, group), nil, true), http.StatusNoContent)
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodDelete, filePath+query(file.Revision, group), nil, true), http.StatusNoContent)
	source := authedJSONRequest(t, router, session, http.MethodGet, "/api/v1/decoder-configurations/snapshots/"+strings.Repeat("c", 32)+"/sources/"+file.ID+"/download", nil, false)
	assertStatus(t, source, http.StatusOK)
	if string(content) != source.Body.String() {
		t.Fatal("live deletion changed the frozen original download")
	}
	remaining, err := service.ListCVValues(t.Context(), vehicle.ID)
	if err != nil || len(remaining) != 1 || remaining[0].DecoderProfile != "Legacy" || remaining[0].Value != 8 {
		t.Fatal("raw group deletion changed the distinct trimmed group")
	}
}
