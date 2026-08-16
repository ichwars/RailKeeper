package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestMasterDataAPILifecycleManagement(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	masterData := application.NewMasterDataService(db)
	router := NewRouter(Config{
		SetupService: setup, AuthService: auth, MasterDataService: masterData,
	})
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	adminSession, adminCookies := loginTestUser(t, router, "admin", "very-secure-password")
	doAuthedJSON(t, router, http.MethodPost, "/api/v1/users",
		`{"username":"editor","email":"editor@example.test","password":"editor-secure-password","roles":["Editor"]}`,
		adminSession, adminCookies, http.StatusCreated)
	doAuthedJSON(t, router, http.MethodPost, "/api/v1/users",
		`{"username":"viewer","email":"viewer@example.test","password":"viewer-secure-password","roles":["Viewer"]}`,
		adminSession, adminCookies, http.StatusCreated)
	editorSession, editorCookies := loginTestUser(t, router, "editor", "editor-secure-password")
	viewerSession, viewerCookies := loginTestUser(t, router, "viewer", "viewer-secure-password")

	if _, err := masterData.Create(t.Context(), "manufacturer", application.MasterDataInput{
		Key: "unused", Label: "Unused",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json,
  created_at, updated_at, origin
) VALUES
  ('manufacturer:bundled', 'manufacturer', 'bundled', 'Bundled', 1, 0, '', '{}', 'now', 'now', 'bundled'),
  ('manufacturer:used', 'manufacturer', 'used', 'Used', 1, 0, '', '{}', 'now', 'now', 'custom')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, category, created_at, updated_at
) VALUES('v1', 'RK-1', 'Used', 'Test', 'H0', 'Lokomotive', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}

	response := doAuthedJSON(t, router, http.MethodGet,
		"/api/v1/master-data/manufacturer?management=true", "",
		editorSession, editorCookies, http.StatusOK)
	var entries []application.MasterDataEntry
	if err := json.Unmarshal(response.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	byKey := make(map[string]application.MasterDataEntry, len(entries))
	for _, entry := range entries {
		byKey[entry.Key] = entry
	}
	if byKey["bundled"].Capabilities == nil || byKey["bundled"].Capabilities.CanDelete {
		t.Fatalf("bundled capabilities=%#v", byKey["bundled"].Capabilities)
	}
	if byKey["unused"].Capabilities == nil || !byKey["unused"].Capabilities.CanDelete {
		t.Fatalf("unused capabilities=%#v", byKey["unused"].Capabilities)
	}
	if byKey["used"].Capabilities == nil || byKey["used"].Capabilities.CanDelete {
		t.Fatalf("used capabilities=%#v", byKey["used"].Capabilities)
	}

	allResponse := doAuthedJSON(t, router, http.MethodGet,
		"/api/v1/master-data-all?management=true", "",
		editorSession, editorCookies, http.StatusOK)
	var all map[string][]application.MasterDataEntry
	if err := json.Unmarshal(allResponse.Body.Bytes(), &all); err != nil {
		t.Fatal(err)
	}
	if len(all["manufacturer"]) != len(entries) || all["manufacturer"][0].Capabilities == nil {
		t.Fatalf("all management response=%#v", all["manufacturer"])
	}

	response = doAuthedJSON(t, router, http.MethodPatch,
		"/api/v1/master-data/manufacturer/unused/active", `{"active":false}`,
		editorSession, editorCookies, http.StatusOK)
	var entry application.MasterDataEntry
	if err := json.Unmarshal(response.Body.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Active {
		t.Fatal("entry remained active")
	}
	doAuthedJSON(t, router, http.MethodPatch,
		"/api/v1/master-data/manufacturer/unused/active", `{}`,
		editorSession, editorCookies, http.StatusBadRequest)
	doAuthedJSON(t, router, http.MethodPatch,
		"/api/v1/master-data/manufacturer/missing/active", `{"active":false}`,
		editorSession, editorCookies, http.StatusNotFound)
	doAuthedJSON(t, router, http.MethodPatch,
		"/api/v1/master-data/manufacturer/unused/active", `{"active":true}`,
		viewerSession, viewerCookies, http.StatusForbidden)
	doJSON(t, router, http.MethodPatch,
		"/api/v1/master-data/manufacturer/unused/active", `{"active":true}`,
		editorCookies, http.StatusForbidden)

	assertMasterDataDeleteProblem(t, router, editorSession, editorCookies,
		"bundled", "master_data_bundled")
	assertMasterDataDeleteProblem(t, router, editorSession, editorCookies,
		"used", "master_data_in_use")
}

func assertMasterDataDeleteProblem(
	t *testing.T,
	router http.Handler,
	session application.SessionView,
	cookies []*http.Cookie,
	key, wantCode string,
) {
	t.Helper()
	response := doAuthedJSON(t, router, http.MethodDelete,
		"/api/v1/master-data/manufacturer/"+key, "", session, cookies, http.StatusConflict)
	var problem map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem["error"] != wantCode {
		t.Fatalf("delete %s problem=%#v", key, problem)
	}
}

func TestMasterDataAPIProtectsStandardArticleTypeKeys(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	router := NewRouter(Config{
		SetupService: setup, AuthService: auth,
		MasterDataService: application.NewMasterDataService(db),
	})
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")

	create := doAuthedJSON(t, router, http.MethodPost, "/api/v1/master-data/article_type",
		`{"key":"custom","label":"Custom","active":true}`, session, cookies, http.StatusBadRequest)
	keyChange := doAuthedJSON(t, router, http.MethodPut, "/api/v1/master-data/article_type/track",
		`{"key":"renamed","label":"Gleismaterial","active":true}`, session, cookies, http.StatusBadRequest)
	doAuthedJSON(t, router, http.MethodPut, "/api/v1/master-data/article_type/track",
		`{"label":"Gleismaterial","active":false}`, session, cookies, http.StatusOK)
	deleteResponse := doAuthedJSON(t, router, http.MethodDelete, "/api/v1/master-data/article_type/track",
		"", session, cookies, http.StatusBadRequest)

	for name, response := range map[string]struct {
		body []byte
	}{
		"create":     {body: create.Body.Bytes()},
		"key change": {body: keyChange.Body.Bytes()},
		"delete":     {body: deleteResponse.Body.Bytes()},
	} {
		var problem map[string]string
		if err := json.Unmarshal(response.body, &problem); err != nil {
			t.Fatalf("%s: decode problem: %v", name, err)
		}
		if problem["message"] != "Standard article type keys cannot be created, changed, or deleted." {
			t.Fatalf("%s: unexpected validation message %q", name, problem["message"])
		}
	}
}

func TestMasterDataAPIReportsControlledCustomFieldMetadataValidation(t *testing.T) {
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	router := NewRouter(Config{
		SetupService: setup, AuthService: auth,
		MasterDataService: application.NewMasterDataService(db),
	})
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")

	response := doAuthedJSON(t, router, http.MethodPost, "/api/v1/master-data/accessory_custom_field",
		`{"key":"invalid","label":"Invalid","active":true,"metadata":{}}`,
		session, cookies, http.StatusBadRequest)
	var problem map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem["error"] != "master_data_validation" ||
		problem["message"] != `master data validation failed: custom field "invalid" requires a valid kind` {
		t.Fatalf("unexpected controlled field problem: %#v", problem)
	}
}

func TestMasterDataAPIDeleteReportsReferencedCustomFieldWithoutMutation(t *testing.T) {
	fixture := newReferencedCustomFieldAPIFixture(t)

	response := doAuthedJSON(t, fixture.router, http.MethodDelete,
		"/api/v1/master-data/accessory_custom_field/length", "", fixture.session, fixture.cookies,
		http.StatusConflict)
	var problem map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem["error"] != "master_data_in_use" ||
		problem["message"] != "This master data entry is still in use and can only be deactivated." {
		t.Fatalf("unexpected referenced custom field delete problem: %#v", problem)
	}
	fixture.assertUnchanged(t)
}

func TestMasterDataAPIImportReportsOmittedReferencedCustomFieldWithoutMutation(t *testing.T) {
	fixture := newReferencedCustomFieldAPIFixture(t)
	doc, err := fixture.masterData.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	delete(doc.Entries, "accessory_custom_field")
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}

	response := doAuthedMultipart(t, fixture.router, "/api/v1/master-data/import",
		"railkeeper-master-data.json", data, fixture.session, fixture.cookies, http.StatusBadRequest)
	var problem map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &problem); err != nil {
		t.Fatal(err)
	}
	if problem["error"] != "master_data_import_invalid" ||
		problem["message"] != `master data validation failed: master data entry is in use: omitted referenced master data accessory_custom_field/length` {
		t.Fatalf("unexpected referenced custom field import problem: %#v", problem)
	}
	fixture.assertUnchanged(t)
}

type referencedCustomFieldAPIFixture struct {
	router      http.Handler
	session     application.SessionView
	cookies     []*http.Cookie
	masterData  *application.MasterDataService
	accessories *application.AccessoryService
	productID   string
}

func newReferencedCustomFieldAPIFixture(t *testing.T) referencedCustomFieldAPIFixture {
	t.Helper()
	db := testRouterDB(t)
	setup := application.NewSetupService(db)
	auth := application.NewAuthService(db)
	masterData := application.NewMasterDataService(db)
	accessories := application.NewAccessoryService(infrastructure.NewAccessoryRepository(db))
	router := NewRouter(Config{SetupService: setup, AuthService: auth, MasterDataService: masterData})
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{
		Username: "admin", Email: "admin@example.test", Password: "very-secure-password",
	}); err != nil {
		t.Fatal(err)
	}
	session, cookies := loginTestUser(t, router, "admin", "very-secure-password")
	active := true
	if _, err := masterData.Create(t.Context(), "accessory_custom_field", application.MasterDataInput{
		Key: "length", Label: "Length", Active: &active,
		Metadata: map[string]any{"kind": "number", "unit": "mm"},
	}); err != nil {
		t.Fatal(err)
	}
	value := 12.5
	unit := "mm"
	product, err := accessories.CreateProduct(t.Context(), application.CreateAccessoryProductInput{
		Manufacturer: "Club", Name: "Referenced article", Category: "other",
		ArticleType: domain.AccessoryArticleOther, Subtype: "other:other", PackageQuantity: 1,
		StockUnit: "piece", InventoryStrategy: domain.AccessoryInventoryQuantity,
		Attributes: []domain.AccessoryAttributeValue{{
			Key: "length", Kind: domain.AccessoryAttributeNumber, NumberValue: &value, Unit: &unit,
		}},
	}, "editor")
	if err != nil {
		t.Fatal(err)
	}
	return referencedCustomFieldAPIFixture{
		router: router, session: session, cookies: cookies, masterData: masterData,
		accessories: accessories, productID: product.ID,
	}
}

func (fixture referencedCustomFieldAPIFixture) assertUnchanged(t *testing.T) {
	t.Helper()
	if _, err := fixture.masterData.Get(t.Context(), "accessory_custom_field", "length"); err != nil {
		t.Fatalf("referenced custom field changed: %v", err)
	}
	product, err := fixture.accessories.GetProduct(t.Context(), fixture.productID)
	if err != nil {
		t.Fatal(err)
	}
	if len(product.Attributes) != 1 || product.Attributes[0].Key != "length" {
		t.Fatalf("referenced product attributes changed: %#v", product.Attributes)
	}
}
