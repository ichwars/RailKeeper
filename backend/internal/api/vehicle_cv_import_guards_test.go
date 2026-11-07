package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestCVImportRevisionGuardsRollbackAndRejectReplay(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "cv-guards", Password: "cv-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	service := application.NewVehicleService(db)
	v, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "Guarded import", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	aInput := application.VehicleCVValueInput{CVNumber: 1, Value: 3, DecoderProfile: "A", Description: "Keep metadata"}
	a, err := service.CreateCVValue(t.Context(), v.ID, aInput)
	if err != nil {
		t.Fatal(err)
	}
	bInput := application.VehicleCVValueInput{CVNumber: 2, Value: 6, DecoderProfile: "A"}
	b, err := service.CreateCVValue(t.Context(), v.ID, bInput)
	if err != nil {
		t.Fatal(err)
	}
	bInput.Value = 8
	changed, err := service.UpdateCVValue(t.Context(), v.ID, b.ID, bInput)
	if err != nil {
		t.Fatal(err)
	}
	read := func() []application.VehicleCVValue {
		t.Helper()
		values, err := service.ListCVValues(t.Context(), v.ID)
		if err != nil {
			t.Fatal(err)
		}
		return values
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	session := loginRouteTestUser(t, auth, "cv-guards", "cv-test-password")
	path := "/api/v1/vehicles/" + v.ID + "/cv-values/import"
	aInput.Value, aInput.ExpectedRevision = 4, &a.Revision
	bInput.Value, bInput.ExpectedRevision = 9, &b.Revision
	request := application.VehicleCVImportInput{Values: []application.VehicleCVValueInput{aInput, bInput}}
	before := read()
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, request, true), http.StatusConflict, "cv_value_conflict")
	if !reflect.DeepEqual(read(), before) {
		t.Fatal("a stale later row changed the first CV or its history")
	}
	request.Values[1].ExpectedRevision = &changed.Revision
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, request, true), http.StatusOK)
	stored := read()
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, request, true), http.StatusConflict, "cv_value_conflict")
	if !reflect.DeepEqual(read(), stored) {
		t.Fatal("replay changed CV IDs, metadata, revisions or histories")
	}
	absent := ""
	request.Values = []application.VehicleCVValueInput{{CVNumber: 3, Value: 10, ExpectedRevision: &absent}}
	assertStatus(t, authedJSONRequest(t, router, session, http.MethodPost, path, request, true), http.StatusOK)
	created := read()
	assertProblem(t, authedJSONRequest(t, router, session, http.MethodPost, path, request, true), http.StatusConflict, "cv_value_conflict")
	if !reflect.DeepEqual(read(), created) || len(created) != 3 {
		t.Fatal("replaying an absent-identity confirmation created duplicate data")
	}
}

func TestCVImportDataBudgetExcludesOnlyComparisonVersions(t *testing.T) {
	plain := `{"cvNumber":1,"value":3,"description":"<>&` + string(rune(0x2028)) + `"}`
	for _, key := range []string{"expectedRevision", "ExpectedRevision", `\u0065xpectedRevision`,
		"expectedWorkspaceRevision", "ExpectedWorkspaceRevision", `\u0065xpectedWorkspaceRevision`} {
		guarded := strings.TrimSuffix(plain, "}") + `,"` + key + `":"` + strings.Repeat("a", 64) + `"}`
		size, err := cvImportDataSize([]json.RawMessage{json.RawMessage(guarded)})
		if err != nil || size != len(`{"values":[]}`)+len(plain) {
			t.Fatalf("comparison overhead or UTF-8 business bytes changed: %d %v", size, err)
		}
	}
}

func TestCVImportBudgetPreservesRawBusinessEscaping(t *testing.T) {
	for _, description := range []string{
		`"<>&"`, `"\u003c\u003e\u0026"`, `"` + string(rune(0x2028)) + string(rune(0x2029)) + `"`,
		`"\u2028\u2029"`, `"ä😀"`, `"\\u2028"`, `"a\nb"`,
	} {
		plain := `{"cvNumber":1,"value":3,"description":` + description + `}`
		guarded := strings.TrimSuffix(plain, "}") + `,"expectedRevision":"` + strings.Repeat("a", 64) + `"}`
		size, err := cvImportDataSize([]json.RawMessage{json.RawMessage(guarded)})
		if err != nil || size != len(`{"values":[]}`)+len(plain) {
			t.Fatalf("business escaping changed for %q: %d %v", description, size, err)
		}
	}
}

func TestCVImportBodyKeepsBusinessBudgetWhenVersionsAddOverhead(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{Username: "cv-budget", Password: "cv-test-password", Roles: []string{"Editor"}}); err != nil {
		t.Fatal(err)
	}
	service := application.NewVehicleService(db)
	v, err := service.Create(t.Context(), application.CreateVehicleInput{Manufacturer: "Test", Name: "Budget", Gauge: "TT", Category: "Lokomotive", Gattung: "Diesellok"}, "test")
	if err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{AuthService: auth, VehicleService: service})
	session := loginRouteTestUser(t, auth, "cv-budget", "cv-test-password")
	path := "/api/v1/vehicles/" + v.ID + "/cv-values/import"
	// An invalid domain value stops before SQL writes while proving the complete large body was parsed.
	plain := []byte(`{"cvNumber":1,"value":300,"description":"` + strings.Repeat("x", 1000) + `"}`)
	guarded := []byte(strings.TrimSuffix(string(plain), "}") + `,"expectedRevision":""}`)
	count := (maxCVImportDataBytes - len(`{"values":[]}`)) / (len(plain) + 1)
	if count > application.MaxVehicleCVImportValues {
		t.Fatal("fixture exceeds row limit")
	}
	body := func(row []byte, count int) json.RawMessage {
		rows := make([][]byte, count)
		for index := range rows {
			rows[index] = row
		}
		return json.RawMessage(append(append([]byte(`{"values":[`), bytes.Join(rows, []byte(","))...), []byte(`]}`)...))
	}
	request := body(guarded, count)
	if len(request) <= maxCVImportDataBytes {
		t.Fatal("fixture does not cross the original transport limit")
	}
	response := authedJSONRequest(t, router, session, http.MethodPost, path, request, true)
	message := response.Body.String()
	assertProblem(t, response, http.StatusBadRequest, "cv_import_invalid")
	if !strings.Contains(message, "Invalid CV values") {
		t.Fatal("valid business budget was rejected because of comparison overhead")
	}
	response = authedJSONRequest(t, router, session, http.MethodPost, path, body(plain, count+1), true)
	message = response.Body.String()
	assertProblem(t, response, http.StatusBadRequest, "cv_import_invalid")
	if !strings.Contains(message, "exceeds 8 MiB") {
		t.Fatal("business data limit was weakened")
	}
}
