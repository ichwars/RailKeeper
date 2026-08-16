package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestOverviewValuationEndpointEnforcesViewerAccessAndReturnsExactValues(t *testing.T) {
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
	for _, user := range []application.CreateUserInput{
		{Username: "admin-valuation", Password: "admin-password", Roles: []string{"Admin"}},
		{Username: "editor-valuation", Password: "editor-password", Roles: []string{"Editor"}},
		{Username: "viewer-valuation", Password: "viewer-password", Roles: []string{"Viewer"}},
		{Username: "messe-valuation", Password: "messe-password", Roles: []string{"Messe"}},
	} {
		if _, err := auth.CreateUser(t.Context(), "", user); err != nil {
			t.Fatal(err)
		}
	}
	router := NewRouter(Config{
		AuthService:              auth,
		OverviewValuationService: application.NewOverviewValuationService(db),
	})

	unauthenticated := httptest.NewRecorder()
	router.ServeHTTP(unauthenticated, httptest.NewRequest(http.MethodGet, "/api/v1/overview/valuation", nil))
	if unauthenticated.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d", unauthenticated.Code)
	}

	for _, credentials := range []struct{ username, password string }{
		{"admin-valuation", "admin-password"},
		{"editor-valuation", "editor-password"},
		{"viewer-valuation", "viewer-password"},
	} {
		login, err := auth.Login(t.Context(), application.LoginInput{
			Username: credentials.username, Password: credentials.password,
		})
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest(http.MethodGet, "/api/v1/overview/valuation", nil)
		request.AddCookie(&http.Cookie{Name: "rk_session", Value: login.SessionToken})
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d: %s", credentials.username, response.Code, response.Body.String())
		}
		var valuation application.OverviewValuation
		if err := json.NewDecoder(response.Body).Decode(&valuation); err != nil {
			t.Fatal(err)
		}
		if valuation.VehicleListValue != "0.00" || valuation.VehiclePurchaseValue != "0.00" ||
			valuation.AccessoryListValue != "0.00" || valuation.AccessoryPurchaseCost != "0.00" ||
			valuation.ExcludedForeignCurrencyPurchases != 0 {
			t.Fatalf("unexpected valuation: %+v", valuation)
		}
	}

	messe, err := auth.Login(t.Context(), application.LoginInput{
		Username: "messe-valuation", Password: "messe-password",
	})
	if err != nil {
		t.Fatal(err)
	}
	messeRequest := httptest.NewRequest(http.MethodGet, "/api/v1/overview/valuation", nil)
	messeRequest.AddCookie(&http.Cookie{Name: "rk_session", Value: messe.SessionToken})
	messeResponse := httptest.NewRecorder()
	router.ServeHTTP(messeResponse, messeRequest)
	if messeResponse.Code != http.StatusForbidden {
		t.Fatalf("messe status = %d", messeResponse.Code)
	}
}

func TestOverviewValuationContractAndRouteAreRegistered(t *testing.T) {
	found := false
	for _, route := range apiRouteSpecs() {
		if route.Method == http.MethodGet && route.Path == "/api/v1/overview/valuation" {
			found = route.Access == routeAccessViewer
		}
	}
	if !found {
		t.Fatal("viewer overview valuation route is not registered")
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "openapi", "railkeeper.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	contract := string(data)
	for _, fragment := range []string{
		"  /overview/valuation:\n", "    OverviewValuation:\n",
		"vehicleListValue:", "vehiclePurchaseValue:", "accessoryListValue:",
		"accessoryPurchaseCost:", "excludedForeignCurrencyPurchases:",
	} {
		if !strings.Contains(contract, fragment) {
			t.Errorf("OpenAPI valuation contract is missing %q", fragment)
		}
	}
	for _, schema := range []string{"AccessoryProduct", "AccessoryProductInput", "AccessoryArticleListItem"} {
		if !strings.Contains(openAPIIndentedBlock(t, contract, schema, 4), "listPrice:") {
			t.Errorf("OpenAPI schema %s is missing listPrice", schema)
		}
	}
	for _, schema := range []string{"Vehicle", "CreateVehicleRequest"} {
		block := openAPIIndentedBlock(t, contract, schema, 4)
		if !strings.Contains(block, "maximumSpeedKmh:") || !strings.Contains(block, "homeBase:") {
			t.Errorf("OpenAPI schema %s is missing operational fields", schema)
		}
	}
}
