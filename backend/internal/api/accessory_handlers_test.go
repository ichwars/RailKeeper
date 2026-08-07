package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
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
	assertStatus(t, layoutRequest(t, router, session, http.MethodGet,
		"/api/v1/accessory-products?query=Tillig", nil, true), http.StatusOK)
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
