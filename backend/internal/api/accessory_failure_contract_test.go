package api

import (
	"net/http/httptest"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestAccessoryHTTPFailuresDoNotAcknowledgeUnknownFieldsOrUnavailableStorage(t *testing.T) {
	f := newContractFixture(t)
	repository := infrastructure.NewAccessoryRepository(f.db)
	app := &App{logger: f.config.Logger, accessoryService: application.NewAccessoryService(repository), accessoryAllocationService: application.NewAccessoryAllocationService(repository)}
	if err := f.db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, handler := range []routeHandler{(*App).getAccessoryStock, (*App).listAccessoryStockMovements, (*App).listAccessoryPurchases, (*App).listAccessoryAssets, (*App).listStorageLocations, (*App).listAccessoryReservations, (*App).listAccessoryInstallations, (*App).getAccessoryAllocationSummary, (*App).getAccessoryUsageHistory} {
		r := httptest.NewRequest("GET", "/contract", nil)
		r.SetPathValue("id", strings.Repeat("1", 32))
		w := httptest.NewRecorder()
		handler(app, w, r)
		assertProblem(t, w, 500, "accessory_operation_failed")
	}
	for _, handler := range []routeHandler{(*App).adjustAccessoryStock, (*App).transferAccessoryStock, (*App).createAccessoryPurchase, (*App).individualizeAccessoryProduct, (*App).createAccessoryAsset, (*App).updateAccessoryAsset, (*App).createStorageLocation, (*App).updateStorageLocation, (*App).removeAccessoryInstallation, (*App).updateAccessoryInstallationCondition} {
		r := httptest.NewRequest("POST", "/contract", strings.NewReader(`{"unexpectedField":"must not become a write"}`))
		r.SetPathValue("id", strings.Repeat("1", 32))
		w := httptest.NewRecorder()
		handler(app, w, r)
		assertProblem(t, w, 400, "invalid_json")
	}
}
