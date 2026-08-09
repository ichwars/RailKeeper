package api

import (
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/domain"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryAllocationRoutesEnforceMixedRolesAndCSRF(t *testing.T) {
	router, sessions, product, location, layout := accessoryAllocationRouter(t)

	readPaths := []string{
		"/api/v1/accessory-reservations",
		"/api/v1/accessory-installations",
		"/api/v1/accessory-products/" + product.ID + "/allocation-summary",
		"/api/v1/accessory-products/" + product.ID + "/usage-history",
	}
	for _, path := range readPaths {
		for role, session := range sessions {
			response := layoutRequest(t, router, session, http.MethodGet, path, nil, true)
			want := http.StatusOK
			if role == "messe" {
				want = http.StatusForbidden
			}
			if response.Code != want {
				t.Fatalf("%s GET %s: got %d, want %d: %s", role, path, response.Code, want, response.Body.String())
			}
		}
	}

	reservationIDs := map[string]string{}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-reservations",
			map[string]any{
				"productId": product.ID, "locationId": location.ID, "quantity": 1, "layoutId": layout.ID,
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "editor" || role == "planner" {
			want = http.StatusCreated
		}
		if response.Code != want {
			t.Fatalf("%s reserve: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
		if want == http.StatusCreated {
			var reservation application.AccessoryReservation
			decodeResponse(t, response, &reservation)
			reservationIDs[role] = reservation.ID
		}
	}

	withoutCSRF := layoutRequest(t, router, sessions["planner"], http.MethodPost,
		"/api/v1/accessory-reservations", map[string]any{
			"productId": product.ID, "locationId": location.ID, "quantity": 1, "layoutId": layout.ID,
		}, false)
	assertProblem(t, withoutCSRF, http.StatusForbidden, "csrf_required")

	for _, role := range []string{"admin", "editor", "planner"} {
		path := "/api/v1/accessory-reservations/" + reservationIDs[role] + "/cancel"
		assertStatus(t, layoutRequest(t, router, sessions[role], http.MethodPost, path, nil, true), http.StatusOK)
	}
	for _, role := range []string{"viewer", "messe"} {
		path := "/api/v1/accessory-reservations/not-allowed/cancel"
		assertStatus(t, layoutRequest(t, router, sessions[role], http.MethodPost, path, nil, true),
			http.StatusForbidden)
	}

	installations := map[string]application.AccessoryInstallation{}
	for role, session := range sessions {
		response := layoutRequest(t, router, session, http.MethodPost, "/api/v1/accessory-installations",
			map[string]any{
				"productId": product.ID, "sourceLocationId": location.ID, "quantity": 1,
				"layoutId": layout.ID, "condition": "ready",
			}, true)
		want := http.StatusForbidden
		if role == "admin" || role == "editor" {
			want = http.StatusCreated
		}
		if response.Code != want {
			t.Fatalf("%s install: got %d, want %d: %s", role, response.Code, want, response.Body.String())
		}
		if want == http.StatusCreated {
			var installation application.AccessoryInstallation
			decodeResponse(t, response, &installation)
			installations[role] = installation
		}
	}

	conditionPath := "/api/v1/accessory-installations/" + installations["admin"].ID + "/condition"
	assertStatus(t, layoutRequest(t, router, sessions["admin"], http.MethodPut, conditionPath,
		map[string]any{"condition": "maintenance_due"}, true), http.StatusOK)
	removePath := "/api/v1/accessory-installations/" + installations["editor"].ID + "/remove"
	assertStatus(t, layoutRequest(t, router, sessions["editor"], http.MethodPost, removePath,
		map[string]any{"disposition": "stored", "storageLocationId": location.ID}, true), http.StatusOK)
	for _, role := range []string{"planner", "viewer", "messe"} {
		assertStatus(t, layoutRequest(t, router, sessions[role], http.MethodPut,
			"/api/v1/accessory-installations/not-allowed/condition",
			map[string]any{"condition": "ready"}, true), http.StatusForbidden)
		assertStatus(t, layoutRequest(t, router, sessions[role], http.MethodPost,
			"/api/v1/accessory-installations/not-allowed/remove",
			map[string]any{"disposition": "maintenance"}, true), http.StatusForbidden)
	}
}

func TestAccessoryAllocationRoutesCoverLifecycleAndSummary(t *testing.T) {
	router, sessions, product, location, layout := accessoryAllocationRouter(t)
	editor := sessions["editor"]

	reservationResponse := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/accessory-reservations", map[string]any{
			"productId": product.ID, "locationId": location.ID, "quantity": 4,
			"layoutId": layout.ID, "note": "Bahnhof",
		}, true)
	assertStatus(t, reservationResponse, http.StatusCreated)
	var reservation application.AccessoryReservation
	decodeResponse(t, reservationResponse, &reservation)

	installResponse := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/accessory-installations", map[string]any{
			"reservationId": reservation.ID, "productId": product.ID, "sourceLocationId": location.ID,
			"quantity": 4, "layoutId": layout.ID, "condition": "ready", "notes": "Montiert",
		}, true)
	assertStatus(t, installResponse, http.StatusCreated)
	var installation application.AccessoryInstallation
	decodeResponse(t, installResponse, &installation)

	assertStatus(t, layoutRequest(t, router, editor, http.MethodGet,
		"/api/v1/accessory-reservations?productId="+product.ID, nil, true), http.StatusOK)
	assertStatus(t, layoutRequest(t, router, editor, http.MethodGet,
		"/api/v1/accessory-installations?productId="+product.ID, nil, true), http.StatusOK)

	summaryResponse := layoutRequest(t, router, editor, http.MethodGet,
		"/api/v1/accessory-products/"+product.ID+"/allocation-summary", nil, true)
	assertStatus(t, summaryResponse, http.StatusOK)
	var summary application.AccessoryAllocationSummary
	decodeResponse(t, summaryResponse, &summary)
	if summary.Owned != 40 || summary.Stored != 36 || summary.Reserved != 0 || summary.Installed != 4 ||
		summary.Available != 36 || summary.Missing != 0 {
		t.Fatalf("unexpected allocation summary: %#v", summary)
	}

	removeResponse := layoutRequest(t, router, editor, http.MethodPost,
		"/api/v1/accessory-installations/"+installation.ID+"/remove",
		map[string]any{"disposition": "maintenance", "notes": "Prüfstand"}, true)
	assertStatus(t, removeResponse, http.StatusOK)

	historyResponse := layoutRequest(t, router, editor, http.MethodGet,
		"/api/v1/accessory-products/"+product.ID+"/usage-history", nil, true)
	assertStatus(t, historyResponse, http.StatusOK)
	var history application.AccessoryUsageHistory
	decodeResponse(t, historyResponse, &history)
	eventTypes := map[application.AccessoryUsageEventType]bool{}
	for _, event := range history.Events {
		eventTypes[event.Type] = true
	}
	if history.ProductID != product.ID || len(history.Events) != 3 ||
		!eventTypes[application.AccessoryUsageReservation] ||
		!eventTypes[application.AccessoryUsageInstallation] ||
		!eventTypes[application.AccessoryUsageRemoval] {
		t.Fatalf("unexpected usage history: %#v", history)
	}

	invalid := layoutRequest(t, router, editor, http.MethodPost, "/api/v1/accessory-reservations",
		map[string]any{"productId": product.ID, "locationId": location.ID, "quantity": 1}, true)
	assertProblem(t, invalid, http.StatusBadRequest, "accessory_validation")
}

func accessoryAllocationRouter(t *testing.T) (
	http.Handler,
	map[string]*application.LoginResult,
	*application.AccessoryProduct,
	*application.StorageLocation,
	*application.Layout,
) {
	t.Helper()
	db := testRouterDB(t)
	auth := application.NewAuthService(db)
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
	sessions := map[string]*application.LoginResult{
		"admin":   loginRouteTestUser(t, auth, "admin", "admin-password"),
		"editor":  loginRouteTestUser(t, auth, "editor", "editor-password"),
		"viewer":  loginRouteTestUser(t, auth, "viewer", "viewer-password"),
		"planner": loginRouteTestUser(t, auth, "planner", "planner-password"),
		"messe":   loginRouteTestUser(t, auth, "messe", "messe-password"),
	}

	repository := infrastructure.NewAccessoryRepository(db)
	accessories := application.NewAccessoryService(repository)
	product, err := accessories.CreateProduct(t.Context(), application.CreateAccessoryProductInput{
		Manufacturer: "Tillig", ArticleNumber: "83101", Name: "Gerades Gleis",
		Category: "Gleismaterial", TrackingMode: domain.AccessoryTrackingModeQuantity,
	}, "seed")
	if err != nil {
		t.Fatal(err)
	}
	location, err := accessories.CreateLocation(t.Context(), application.CreateStorageLocationInput{Name: "Lager"}, "seed")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := accessories.AdjustStock(t.Context(), product.ID, application.StockAdjustmentInput{
		LocationID: location.ID, Delta: 40,
	}, "seed"); err != nil {
		t.Fatal(err)
	}
	layouts := application.NewLayoutService(infrastructure.NewLayoutRepository(db))
	layout, err := layouts.CreateLayout(t.Context(), application.CreateLayoutInput{
		Name: "Heimanlage", Kind: domain.LayoutKindPrivate, Gauge: "TT", Scale: "1:120",
	}, "seed")
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter(Config{
		AuthService:                auth,
		AccessoryService:           accessories,
		AccessoryAllocationService: application.NewAccessoryAllocationService(repository),
	})
	return router, sessions, product, location, layout
}
