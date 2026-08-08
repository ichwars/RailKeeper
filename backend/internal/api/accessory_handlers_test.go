package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryRoutesEnforceRoleAndCSRFBoundaries(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	accessories := application.NewAccessoryService(infrastructure.NewAccessoryRepository(db))
	for _, user := range []application.CreateUserInput{
		{Username: "admin", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "planner", Password: "planner-password", Roles: []string{"Planner"}},
		{Username: "messe", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(Config{AuthService: auth, AccessoryService: accessories})
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin", "admin-password"),
		"editor":  loginRouteTestUser(t, auth, "editor", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer", "viewer-password"),
		"planner": loginRouteTestUser(t, auth, "planner", "planner-password"),
		"messe":   loginRouteTestUser(t, auth, "messe", "messe-password"),
	}

	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodGet, "/api/v1/accessory-products", nil, true)
		want := http.StatusOK
		if role == "messe" {
			want = http.StatusForbidden
		}
		if response.Code != want {
			t.Fatalf("%s read: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-products", map[string]any{
			"manufacturer": "Tillig", "articleNumber": "83-" + role, "name": "Gleis " + role,
			"category": "Gleismaterial", "trackingMode": "quantity",
		}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "editor" {
			want = http.StatusCreated
		}
		if response.Code != want {
			t.Fatalf("%s write: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
	}
	withoutCSRF := layoutRequest(
		t, router, sessions["editor"], http.MethodPost, "/api/v1/accessory-products",
		map[string]any{
			"manufacturer": "Tillig", "articleNumber": "83100", "name": "Gerades Gleis",
			"category": "Gleismaterial", "trackingMode": "quantity",
		}, false,
	)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")
}

func TestAccessoryRoutesCoverCatalogueLocationsAndInventory(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "editor", Password: "editor-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "editor", "editor-password")
	router := NewRouter(Config{
		AuthService:      auth,
		AccessoryService: application.NewAccessoryService(infrastructure.NewAccessoryRepository(db)),
	})

	rootResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/storage-locations", map[string]any{
		"name": "Werkstatt",
	}, true)
	assertStatus(t, rootResponse, http.StatusCreated)
	var root application.StorageLocation
	decodeResponse(t, rootResponse, &root)

	childResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/storage-locations", map[string]any{
		"parentId": root.ID, "name": "Schrank A",
	}, true)
	assertStatus(t, childResponse, http.StatusCreated)
	var child application.StorageLocation
	decodeResponse(t, childResponse, &child)
	assertStatus(t, layoutRequest(t, router, session, http.MethodPut, "/api/v1/storage-locations/"+child.ID,
		map[string]any{"parentId": root.ID, "name": "Schrank A1"}, true), http.StatusOK)
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet, "/api/v1/storage-locations", nil, true),
		http.StatusOK)

	quantityResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-products",
		map[string]any{
			"manufacturer": "Tillig", "articleNumber": "83101", "name": "Gerades Gleis",
			"category": "Gleismaterial", "trackingMode": "quantity",
		}, true)
	assertStatus(t, quantityResponse, http.StatusCreated)
	var quantityProduct application.AccessoryProduct
	decodeResponse(t, quantityResponse, &quantityProduct)
	listResponse := layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/accessory-products?query=Tillig", nil, true)
	assertStatus(t, listResponse, http.StatusOK)
	var listResult application.AccessoryArticleListResult
	decodeResponse(t, listResponse, &listResult)
	if len(listResult.Items) != 1 || listResult.Items[0].ID != quantityProduct.ID ||
		listResult.Metrics.ArticleCount != 1 || len(listResult.FilterOptions.Manufacturers) != 1 {
		t.Fatalf("unexpected article list response: %#v", listResult)
	}
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/accessory-products/"+quantityProduct.ID, nil, true), http.StatusOK)
	assertStatus(t, layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/accessory-products/"+quantityProduct.ID, map[string]any{
			"manufacturer": "Tillig", "articleNumber": "83101", "name": "Gerades Modellgleis",
			"category": "Gleismaterial", "trackingMode": "quantity",
		}, true), http.StatusOK)

	stockPath := "/api/v1/accessory-products/" + quantityProduct.ID + "/stock"
	adjustmentPath := "/api/v1/accessory-products/" + quantityProduct.ID + "/stock-adjustments"
	adjustedResponse := layoutRequest(t, router, session, http.MethodPost, adjustmentPath,
		map[string]any{"locationId": child.ID, "delta": 5}, true)
	assertStatus(t, adjustedResponse, http.StatusOK)
	var stock application.AccessoryStockSummary
	decodeResponse(t, adjustedResponse, &stock)
	if stock.TotalQuantity != 5 {
		t.Fatalf("unexpected stock: %#v", stock)
	}
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet, stockPath, nil, true), http.StatusOK)
	assertProblem(t, layoutRequest(t, router, session, http.MethodPost, adjustmentPath,
		map[string]any{"locationId": child.ID, "delta": -6}, true),
		http.StatusConflict, "accessory_insufficient_stock")

	individualResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-products",
		map[string]any{
			"manufacturer": "Lenz", "articleNumber": "LS150", "name": "Schaltdecoder",
			"category": "Decoder", "trackingMode": "individual",
		}, true)
	assertStatus(t, individualResponse, http.StatusCreated)
	var individualProduct application.AccessoryProduct
	decodeResponse(t, individualResponse, &individualProduct)
	assetsPath := "/api/v1/accessory-products/" + individualProduct.ID + "/assets"
	assetResponse := layoutRequest(t, router, session, http.MethodPost, assetsPath, map[string]any{
		"inventoryNumber": "RK-Z-0001", "condition": "ready", "lifecycle": "stored",
		"storageLocationId": child.ID,
	}, true)
	assertStatus(t, assetResponse, http.StatusCreated)
	var asset application.AccessoryAsset
	decodeResponse(t, assetResponse, &asset)
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet, assetsPath, nil, true), http.StatusOK)
	assertStatus(t, layoutRequest(t, router, session, http.MethodPut, "/api/v1/accessory-assets/"+asset.ID,
		map[string]any{
			"inventoryNumber": "RK-Z-0001", "condition": "maintenance_due", "lifecycle": "maintenance",
			"storageLocationId": child.ID,
		}, true), http.StatusOK)
	assertProblem(t, layoutRequest(t, router, session, http.MethodPost,
		"/api/v1/accessory-products/"+individualProduct.ID+"/stock-adjustments",
		map[string]any{"locationId": child.ID, "delta": 1}, true),
		http.StatusConflict, "accessory_tracking_mode")

	assertProblem(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/accessory-products/missing", nil, true), http.StatusNotFound, "accessory_not_found")
	assertProblem(t, layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-products",
		map[string]any{"name": "Invalid"}, true), http.StatusBadRequest, "accessory_validation")
}

func TestAccessoryArticleRoutesUseActiveConfiguredSubtypeMasterData(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	if _, err := auth.CreateUser(t.Context(), "", application.CreateUserInput{
		Username: "editor", Password: "editor-password", Roles: []string{"Editor"},
	}); err != nil {
		t.Fatal(err)
	}
	session := loginRouteTestUser(t, auth, "editor", "editor-password")
	masterData := application.NewMasterDataService(db)
	active := true
	inactive := false
	if _, err := masterData.Create(t.Context(), "accessory_subtype", application.MasterDataInput{
		Key: "track:club_profile", Label: "Club profile", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := masterData.Create(t.Context(), "accessory_subtype", application.MasterDataInput{
		Key: "track:retired_profile", Label: "Retired profile", Active: &inactive,
	}); err != nil {
		t.Fatal(err)
	}
	router := NewRouter(Config{
		AuthService:      auth,
		AccessoryService: application.NewAccessoryService(infrastructure.NewAccessoryRepository(db)),
	})

	articleInput := func(articleNumber, subtype string) map[string]any {
		return map[string]any{
			"manufacturer": "Club", "articleNumber": articleNumber, "name": "Configured track",
			"category": subtype, "articleType": "track", "subtype": subtype,
			"packageQuantity": 1, "stockUnit": "piece", "inventoryStrategy": "quantity",
		}
	}
	createResponse := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-products",
		articleInput("custom-create", "club_profile"), true)
	assertStatus(t, createResponse, http.StatusCreated)
	var created application.AccessoryProduct
	decodeResponse(t, createResponse, &created)
	if created.Subtype != "track:club_profile" {
		t.Fatalf("configured subtype was not normalized canonically: %#v", created)
	}

	updateResponse := layoutRequest(t, router, session, http.MethodPut,
		"/api/v1/accessory-products/"+created.ID, articleInput("custom-update", "club_profile"), true)
	assertStatus(t, updateResponse, http.StatusOK)

	for name, subtype := range map[string]string{
		"inactive":     "retired_profile",
		"mismatched":   "signal:main",
		"unconfigured": "direct_api_value",
	} {
		t.Run(name, func(t *testing.T) {
			response := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-products",
				articleInput(name, subtype), true)
			assertProblem(t, response, http.StatusBadRequest, "accessory_validation")
		})
	}
}

func TestAccessoryArticleRoutesEnforceRoleAndCSRFForEveryNewEndpoint(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	productPath := "/api/v1/accessory-products/" + fixture.product.ID
	hybridPath := "/api/v1/accessory-products/" + fixture.hybridProduct.ID

	type routeCase struct {
		name       string
		method     string
		path       string
		body       any
		wantEditor int
		multipart  bool
	}
	readCases := []routeCase{
		{name: "catalogue", method: http.MethodGet, path: "/api/v1/accessory-products", wantEditor: http.StatusOK},
		{name: "product", method: http.MethodGet, path: productPath, wantEditor: http.StatusOK},
		{name: "stock movements", method: http.MethodGet, path: productPath + "/stock-movements", wantEditor: http.StatusOK},
		{name: "purchases", method: http.MethodGet, path: productPath + "/purchases", wantEditor: http.StatusOK},
		{name: "documents", method: http.MethodGet, path: productPath + "/documents", wantEditor: http.StatusOK},
		{name: "document", method: http.MethodGet, path: productPath + "/documents/missing", wantEditor: http.StatusNotFound},
		{name: "document download", method: http.MethodGet, path: productPath + "/documents/missing/download", wantEditor: http.StatusNotFound},
		{name: "usage history", method: http.MethodGet, path: productPath + "/usage-history", wantEditor: http.StatusOK},
	}
	for _, testCase := range readCases {
		t.Run("read/"+testCase.name, func(t *testing.T) {
			for role, session := range fixture.sessions {
				response := layoutRequest(t, fixture.router, session, testCase.method, testCase.path, nil, true)
				want := testCase.wantEditor
				if role == "messe" {
					want = http.StatusForbidden
				}
				if response.Code != want {
					t.Fatalf("%s %s %s: got %d, want %d: %s", role, testCase.method, testCase.path,
						response.Code, want, response.Body.String())
				}
			}
		})
	}

	writeCases := []routeCase{
		{name: "duplicate check", method: http.MethodPost, path: "/api/v1/accessory-products/duplicate-check",
			body: map[string]any{"manufacturer": "Tillig", "articleNumber": "83101"}, wantEditor: http.StatusOK},
		{name: "archive", method: http.MethodPost, path: productPath + "/archive", wantEditor: http.StatusOK},
		{name: "restore", method: http.MethodPost, path: productPath + "/restore", wantEditor: http.StatusOK},
		{name: "stock transfer", method: http.MethodPost, path: productPath + "/stock-transfers",
			body: map[string]any{"fromLocationId": fixture.locationA.ID, "toLocationId": fixture.locationB.ID,
				"quantity": 1}, wantEditor: http.StatusOK},
		{name: "purchase", method: http.MethodPost, path: productPath + "/purchases",
			body: map[string]any{"purchasedAt": "2026-08-08", "quantity": 1}, wantEditor: http.StatusCreated},
		{name: "individualization", method: http.MethodPost, path: hybridPath + "/individualizations",
			body: map[string]any{"locationId": fixture.locationA.ID,
				"asset": map[string]any{"condition": "ready", "lifecycle": "stored"}}, wantEditor: http.StatusCreated},
		{name: "document upload", method: http.MethodPost, path: productPath + "/documents",
			wantEditor: http.StatusCreated, multipart: true},
		{name: "document update", method: http.MethodPut, path: productPath + "/documents/missing",
			body: map[string]any{"category": "manual"}, wantEditor: http.StatusNotFound},
		{name: "document delete", method: http.MethodDelete, path: productPath + "/documents/missing",
			wantEditor: http.StatusNotFound},
	}
	for _, testCase := range writeCases {
		t.Run("write/"+testCase.name, func(t *testing.T) {
			for role, session := range fixture.sessions {
				response := accessoryRouteCaseRequest(t, fixture.router, session, testCase, true)
				want := http.StatusForbidden
				if role == "admin" || role == "editor" {
					want = testCase.wantEditor
				}
				if response.Code != want {
					t.Fatalf("%s %s %s: got %d, want %d: %s", role, testCase.method, testCase.path,
						response.Code, want, response.Body.String())
				}
			}
			editorWithoutCSRF := accessoryRouteCaseRequest(
				t, fixture.router, fixture.sessions["editor"], testCase, false,
			)
			assertProblem(t, editorWithoutCSRF, http.StatusForbidden, "csrf_required")
		})
	}
}

func TestAccessoryArticleListParsesApprovedQueryAndRejectsInvalidEnums(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	response := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
		"/api/v1/accessory-products?query=Modell&manufacturer=Tillig&articleType=track&articleType=other"+
			"&gauge=TT&gauge=H0&status=available&status=reserved&locationId="+fixture.locationA.ID+
			"&sort=stock&direction=desc", nil, true)
	assertStatus(t, response, http.StatusOK)
	if !bytes.Contains(response.Body.Bytes(), []byte(`"filters"`)) ||
		bytes.Contains(response.Body.Bytes(), []byte(`"filterOptions"`)) {
		t.Fatalf("article list must expose filters: %s", response.Body.String())
	}
	var result application.AccessoryArticleListResult
	decodeResponse(t, response, &result)
	if len(result.Items) != 1 || result.Items[0].ID != fixture.product.ID || result.Metrics.ArticleCount != 2 {
		t.Fatalf("unexpected filtered list: %#v", result)
	}

	for name, query := range map[string]string{
		"article type": "articleType=invalid",
		"status":       "status=invalid",
		"sort":         "sort=invalid",
		"direction":    "direction=sideways",
	} {
		t.Run(name, func(t *testing.T) {
			invalid := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
				"/api/v1/accessory-products?"+query, nil, true)
			assertProblem(t, invalid, http.StatusBadRequest, "accessory_validation")
		})
	}
}

func TestAccessoryArticleListRejectsInvalidRawGaugeValues(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	for name, query := range map[string]string{
		"empty":      "gauge=",
		"whitespace": "gauge=%20%20",
		"control":    "gauge=%0A",
		"too long":   "gauge=" + strings.Repeat("G", 129),
	} {
		t.Run(name, func(t *testing.T) {
			response := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
				"/api/v1/accessory-products?"+query, nil, true)
			assertProblem(t, response, http.StatusBadRequest, "accessory_validation")
		})
	}

	legitimate := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
		"/api/v1/accessory-products?gauge=Schmalspur%20(750%20mm)", nil, true)
	assertStatus(t, legitimate, http.StatusOK)
}

func TestAccessoryArticleListRejectsInvalidRawScalarFilters(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	filters := []struct {
		name      string
		parameter string
		maxRunes  int
	}{
		{name: "query", parameter: "query", maxRunes: 200},
		{name: "manufacturer", parameter: "manufacturer", maxRunes: 200},
		{name: "location", parameter: "locationId", maxRunes: 128},
	}
	for _, filter := range filters {
		filter := filter
		t.Run(filter.name, func(t *testing.T) {
			tests := map[string]string{
				"repeated":      filter.parameter + "=first&" + filter.parameter + "=second",
				"empty":         filter.parameter + "=",
				"whitespace":    filter.parameter + "=%20%20",
				"invalid UTF-8": filter.parameter + "=%FF",
				"control":       filter.parameter + "=%0A",
				"too long":      filter.parameter + "=" + url.QueryEscape(strings.Repeat("Ä", filter.maxRunes+1)),
			}
			for name, rawQuery := range tests {
				t.Run(name, func(t *testing.T) {
					response := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
						"/api/v1/accessory-products?"+rawQuery, nil, true)
					assertProblem(t, response, http.StatusBadRequest, "accessory_validation")
				})
			}
		})
	}

	for _, rawQuery := range []string{
		"query=" + url.QueryEscape("Märklin Übergang"),
		"manufacturer=" + url.QueryEscape("Česká železnice"),
		"locationId=" + url.QueryEscape("Lager Süd"),
	} {
		response := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
			"/api/v1/accessory-products?"+rawQuery, nil, true)
		assertStatus(t, response, http.StatusOK)
	}
}

func TestAccessoryArticleListRejectsMalformedQueryEncoding(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/accessory-products", nil)
	request.URL.RawQuery = "query=%ZZ"
	request.AddCookie(&http.Cookie{
		Name:  "rk_session",
		Value: fixture.sessions["viewer"].SessionToken,
	})
	response := httptest.NewRecorder()
	fixture.router.ServeHTTP(response, request)
	assertProblem(t, response, http.StatusBadRequest, "accessory_validation")

	validUnicode := layoutRequest(t, fixture.router, fixture.sessions["viewer"], http.MethodGet,
		"/api/v1/accessory-products?query=M%C3%A4rklin%20%C3%9Cbergang", nil, true)
	assertStatus(t, validUnicode, http.StatusOK)
}

func TestAccessoryArticleRoutesCoverDuplicateArchivePurchaseTransferAndIndividualization(t *testing.T) {
	fixture := newAccessoryAPIFixture(t, 1024*1024)
	editor := fixture.sessions["editor"]
	productPath := "/api/v1/accessory-products/" + fixture.product.ID

	duplicate := layoutRequest(t, fixture.router, editor, http.MethodPost,
		"/api/v1/accessory-products/duplicate-check",
		map[string]any{"manufacturer": " Tillig ", "articleNumber": " 83101 "}, true)
	assertStatus(t, duplicate, http.StatusOK)
	var duplicateResult application.AccessoryDuplicateCheckResult
	decodeResponse(t, duplicate, &duplicateResult)
	if len(duplicateResult.Candidates) != 1 || duplicateResult.Candidates[0].ID != fixture.product.ID {
		t.Fatalf("unexpected duplicate result: %#v", duplicateResult)
	}
	createdDuplicate := layoutRequest(t, fixture.router, editor, http.MethodPost, "/api/v1/accessory-products",
		typedAccessoryProductInput("Tillig", "83101", "Confirmed variant", domain.AccessoryInventoryQuantity), true)
	assertStatus(t, createdDuplicate, http.StatusCreated)

	for _, endpoint := range []string{"archive", "archive", "restore", "restore"} {
		response := layoutRequest(t, fixture.router, editor, http.MethodPost, productPath+"/"+endpoint, nil, true)
		assertStatus(t, response, http.StatusOK)
		var product application.AccessoryProduct
		decodeResponse(t, response, &product)
		if product.Archived != (endpoint == "archive") {
			t.Fatalf("%s returned archived=%t", endpoint, product.Archived)
		}
	}
	var archiveAuditCount int
	if err := fixture.db.QueryRow(`SELECT COUNT(*) FROM audit_logs
		WHERE target_type='accessory_product' AND target_id=?
		AND action IN ('AccessoryProductArchived', 'AccessoryProductRestored')`, fixture.product.ID).
		Scan(&archiveAuditCount); err != nil {
		t.Fatal(err)
	}
	if archiveAuditCount != 4 {
		t.Fatalf("archive/restore audit count=%d, want 4", archiveAuditCount)
	}
	assertProblem(t, layoutRequest(t, fixture.router, editor, http.MethodPost,
		"/api/v1/accessory-products/missing/archive", nil, true),
		http.StatusNotFound, "accessory_not_found")

	transfer := layoutRequest(t, fixture.router, editor, http.MethodPost, productPath+"/stock-transfers",
		map[string]any{"fromLocationId": fixture.locationA.ID, "toLocationId": fixture.locationB.ID,
			"quantity": 3, "note": "Move"}, true)
	assertStatus(t, transfer, http.StatusOK)
	var stock application.AccessoryStockSummary
	decodeResponse(t, transfer, &stock)
	if stock.TotalQuantity != 20 || len(stock.Locations) != 2 {
		t.Fatalf("unexpected transferred stock: %#v", stock)
	}
	movements := layoutRequest(t, fixture.router, editor, http.MethodGet, productPath+"/stock-movements", nil, true)
	assertStatus(t, movements, http.StatusOK)
	var journal []application.AccessoryStockMovement
	decodeResponse(t, movements, &journal)
	if len(journal) < 3 {
		t.Fatalf("stock journal missing transfer rows: %#v", journal)
	}

	purchase := layoutRequest(t, fixture.router, editor, http.MethodPost, productPath+"/purchases",
		map[string]any{"purchasedAt": "2026-08-08", "supplier": "Dealer", "quantity": 2,
			"storageLocationId": fixture.locationB.ID, "bookToStock": true, "currency": "eur"}, true)
	assertStatus(t, purchase, http.StatusCreated)
	purchases := layoutRequest(t, fixture.router, editor, http.MethodGet, productPath+"/purchases", nil, true)
	assertStatus(t, purchases, http.StatusOK)
	var purchaseList []application.AccessoryPurchase
	decodeResponse(t, purchases, &purchaseList)
	if len(purchaseList) != 1 || purchaseList[0].Currency != "EUR" {
		t.Fatalf("unexpected purchases: %#v", purchaseList)
	}

	individualized := layoutRequest(t, fixture.router, editor, http.MethodPost,
		"/api/v1/accessory-products/"+fixture.hybridProduct.ID+"/individualizations",
		map[string]any{"locationId": fixture.locationA.ID,
			"asset": map[string]any{"inventoryNumber": "RK-Z-9001", "condition": "ready", "lifecycle": "stored"}}, true)
	assertStatus(t, individualized, http.StatusCreated)
	var asset application.AccessoryAsset
	decodeResponse(t, individualized, &asset)
	if asset.ProductID != fixture.hybridProduct.ID || asset.StorageLocationID != fixture.locationA.ID {
		t.Fatalf("unexpected individualized asset: %#v", asset)
	}
}

func accessoryRouteCaseRequest(
	t *testing.T,
	router http.Handler,
	session *application.LoginResult,
	testCase struct {
		name       string
		method     string
		path       string
		body       any
		wantEditor int
		multipart  bool
	},
	withCSRF bool,
) *httptest.ResponseRecorder {
	t.Helper()
	if testCase.multipart {
		return accessoryMultipartRequest(t, router, session, testCase.method, testCase.path, "manual.pdf",
			[]byte("%PDF-1.7\nexample"), map[string]string{"category": "manual"}, withCSRF)
	}
	return layoutRequest(t, router, session, testCase.method, testCase.path, testCase.body, withCSRF)
}
