package api

import (
	"net/http"
	"testing"
)

func TestWorkshopHTTPContractsPreserveEntriesAndEnforcePermissions(t *testing.T) {
	f := newContractFixture(t)
	id := f.vehicle(t)
	base := "/api/v1/vehicles/" + id
	for _, contract := range []struct {
		suffix, invalidCode string
		input               map[string]any
	}{
		{"maintenance", "maintenance_invalid", map[string]any{"kind": "Wartung", "status": "geplant", "notes": "Original workshop note"}},
		{"spare-parts", "spare_part_invalid", map[string]any{"articleNumber": "ET-123", "description": "Original motor", "price": "12.50"}},
	} {
		t.Run(contract.suffix, func(t *testing.T) {
			path := base + "/" + contract.suffix
			assertStatus(t, f.call(t, "GET", path, nil), 200)
			assertProblem(t, f.raw(t, "POST", path, "{"), 400, "invalid_json")
			assertProblem(t, f.call(t, "POST", path, map[string]any{"dueDate": "invalid"}), 400, contract.invalidCode)
			assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "POST", path, contract.input, true), 403)
			assertStatus(t, authedJSONRequest(t, f.router, f.admin, "POST", path, contract.input, false), 403)
			created := f.call(t, "POST", path, contract.input)
			assertStatus(t, created, 201)
			entryPath := path + "/" + responseID(t, created)
			assertProblem(t, f.raw(t, "PUT", entryPath, "{"), 400, "invalid_json")
			assertProblem(t, f.call(t, "PUT", entryPath, map[string]any{"dueDate": "invalid"}), 400, contract.invalidCode)
			updated := f.call(t, "PUT", entryPath, contract.input)
			assertStatus(t, updated, 200)
			assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "DELETE", entryPath, nil, true), 403)
			assertStatus(t, f.call(t, "GET", path, nil), 200)
			assertStatus(t, f.call(t, "DELETE", entryPath, nil), 204)
			assertStatus(t, f.call(t, "DELETE", entryPath, nil), 404)
			assertStatus(t, f.call(t, "PUT", entryPath, contract.input), 404)
			assertStatus(t, f.call(t, "GET", "/api/v1/vehicles/missing/"+contract.suffix, nil), 404)
			assertStatus(t, f.call(t, "POST", "/api/v1/vehicles/missing/"+contract.suffix, contract.input), 404)
		})
	}
	path := base + "/functions"
	assertStatus(t, f.call(t, "GET", path, nil), 200)
	input := map[string]any{"name": "Cab light", "behavior": "Momentary", "symbol": "light"}
	assertStatus(t, f.call(t, "PUT", path+"/F1", input), 200)
	var entries []map[string]any
	decodeResponse(t, f.call(t, "GET", path, nil), &entries)
	if len(entries) != 1 || entries[0]["name"] != "Cab light" {
		t.Fatalf("stored function lost: %#v", entries)
	}
	assertProblem(t, f.raw(t, "PUT", path+"/F1", "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "PUT", path+"/invalid", input), 400, "function_invalid")
	assertStatus(t, f.call(t, "PUT", "/api/v1/vehicles/missing/functions/F1", input), 404)
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, http.MethodDelete, path+"/F1", nil, true), 403)
	assertStatus(t, f.call(t, "DELETE", path+"/F1", nil), 204)
	assertStatus(t, f.call(t, "DELETE", path+"/F1", nil), 404)
	assertStatus(t, f.call(t, "GET", "/api/v1/vehicles/missing/functions", nil), 404)
	assertStatus(t, f.call(t, "GET", base+"/spare-parts/suggestions", nil), 200)
	assertStatus(t, f.call(t, "GET", "/api/v1/vehicles/missing/spare-parts/suggestions", nil), 404)
}
