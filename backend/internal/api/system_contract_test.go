package api

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestSystemHTTPContractsMeasureAndOptimizeWithoutLosingInventory(t *testing.T) {
	f := newContractFixture(t)
	vehicleID := f.vehicle(t)
	for _, name := range []string{"uploads/image.png", "uploads/thumbs/thumb.jpg", "uploads/manual.pdf", "uploads/decoder.esux", "other.txt"} {
		p := filepath.Join(f.dataDir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("synthetic-retained-data"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	response := f.call(t, "GET", "/health", nil)
	assertStatus(t, response, 200)
	response = f.call(t, "GET", "/api/v1/system/storage", nil)
	assertStatus(t, response, 200)
	var usage storageUsageResponse
	decodeResponse(t, response, &usage)
	if usage.TotalBytes <= 0 || len(usage.Categories) < 2 || usage.UpdatedAt == "" {
		t.Fatalf("invalid storage usage: %#v", usage)
	}
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "GET", "/api/v1/system/storage", nil, true), 403)
	assertStatus(t, f.call(t, "POST", "/api/v1/system/storage/optimize", nil), 200)
	if _, err := f.vehicles.Get(t.Context(), vehicleID); err != nil {
		t.Fatalf("optimize lost vehicle: %v", err)
	}
	for _, name := range []string{"uploads/image.png", "uploads/manual.pdf", "uploads/decoder.esux", "other.txt"} {
		if data, err := os.ReadFile(filepath.Join(f.dataDir, name)); err != nil || string(data) != "synthetic-retained-data" {
			t.Fatalf("optimize changed original file %s", name)
		}
	}
	t.Setenv("RAILKEEPER_PRINTERS", "Printer A, Printer B,Printer A, ,Printer B")
	t.Setenv("RAILKEEPER_DEFAULT_PRINTER", "printer b")
	response = f.call(t, "GET", "/api/v1/system/printers", nil)
	assertStatus(t, response, 200)
	var printers systemPrintersResponse
	decodeResponse(t, response, &printers)
	if len(printers.Printers) != 2 || !printers.Printers[1].IsDefault || printers.Status != "configured" {
		t.Fatalf("printer contract: %#v", printers)
	}
	t.Setenv("RAILKEEPER_DEFAULT_PRINTER", "")
	if result := printersFromEnv(); !result.Printers[0].IsDefault || result.DefaultPrinter != "Printer A" {
		t.Fatalf("default printer fallback: %#v", result)
	}
	f.config.DatabaseMaintenance = nil
	f.config.DataDir = filepath.Join(f.dataDir, "missing")
	f.router = NewRouter(f.config)
	assertProblem(t, f.call(t, "POST", "/api/v1/system/storage/optimize", nil), 500, "storage_optimize_unavailable")
	assertProblem(t, f.call(t, "GET", "/api/v1/system/storage", nil), 500, "storage_usage_failed")
}

func TestNumberingAndLookupHTTPContractsPreserveCounters(t *testing.T) {
	f := newContractFixture(t)
	path := "/api/v1/inventory-number-schemes"
	assertStatus(t, f.call(t, "GET", path, nil), 200)
	input := map[string]any{"category": "Synthetic workshop", "prefix": "W", "nextNumber": 17, "padding": 5, "active": true}
	response := f.call(t, "POST", path, input)
	assertStatus(t, response, 201)
	var scheme application.InventoryNumberScheme
	decodeResponse(t, response, &scheme)
	if scheme.NextNumber != 17 || scheme.Preview != "W-00017" {
		t.Fatalf("counter changed on create: %#v", scheme)
	}
	assertProblem(t, f.call(t, "POST", path, input), 409, "inventory_number_scheme_exists")
	assertProblem(t, f.call(t, "POST", path, map[string]any{}), 400, "inventory_number_validation")
	assertProblem(t, f.raw(t, "POST", path, "{"), 400, "invalid_json")
	input["nextNumber"] = 18
	response = f.call(t, "PUT", path+"/Synthetic%20workshop", input)
	assertStatus(t, response, 200)
	decodeResponse(t, response, &scheme)
	if scheme.NextNumber != 18 || scheme.Preview != "W-00018" {
		t.Fatalf("counter update lost: %#v", scheme)
	}
	assertProblem(t, f.raw(t, "PUT", path+"/Synthetic%20workshop", "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "PUT", path+"/Synthetic%20workshop", map[string]any{"nextNumber": -1}), 400, "inventory_number_validation")
	assertStatus(t, f.call(t, "GET", "/api/v1/master-data-relations", nil), 200)
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, http.MethodPost, path, input, true), 403)
	assertProblem(t, f.raw(t, "POST", "/api/v1/article-search", "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "POST", "/api/v1/article-search", map[string]any{}), 400, "article_search_validation")
}
