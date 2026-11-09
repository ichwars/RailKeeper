package api

import (
	"bytes"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestSetImageHTTPContractsPreserveMemberVehiclesAndOriginalBytes(t *testing.T) {
	f := newContractFixture(t)
	set, err := f.vehicles.CreateSet(t.Context(), application.CreateVehicleSetInput{Set: application.VehicleSetInput{Name: "Original set", Manufacturer: "Test", Gauge: "H0", Category: "Wagen", Gattung: "Reisezugwagen"}, Members: []application.CreateVehicleInput{{Name: "Member 1"}, {Name: "Member 2"}}}, f.adminID)
	if err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/vehicle-sets/" + set.ID + "/image"
	image := testPNG(t)
	assertStatus(t, f.upload(t, path, "set.png", image, nil), 201)
	response := f.call(t, "GET", path+"/file", nil)
	assertStatus(t, response, 200)
	if !bytes.Equal(response.Body.Bytes(), image) {
		t.Fatal("set image bytes changed")
	}
	assertStatus(t, f.call(t, "GET", path+"/thumbnail", nil), 200)
	assertProblem(t, f.raw(t, "PUT", "/api/v1/vehicle-sets/"+set.ID+"/main-image", "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "PUT", "/api/v1/vehicle-sets/"+set.ID+"/main-image", map[string]any{"mode": "invalid"}), 400, "vehicle_set_image_validation")
	assertProblem(t, f.call(t, "GET", "/api/v1/vehicle-sets/missing/image/file", nil), 404, "vehicle_set_not_found")
	f.config.FileBlobService = nil
	f.router = NewRouter(f.config)
	assertProblem(t, f.call(t, "GET", path+"/file", nil), 500, "image_download_failed")
	assertProblem(t, f.call(t, "GET", path+"/thumbnail", nil), 500, "image_thumbnail_failed")
	f.config.FileBlobService = application.NewFileBlobService(f.db, f.dataDir)
	f.router = NewRouter(f.config)
	assertStatus(t, f.call(t, "DELETE", path, nil), 204)
	assertProblem(t, f.call(t, "GET", path+"/file", nil), 404, "vehicle_set_image_not_found")
	current, err := f.vehicles.GetSet(t.Context(), set.ID)
	if err != nil || len(current.Members) != 2 {
		t.Fatalf("set image change lost members: %v", err)
	}
}

func TestExhibitionHTTPContractRechecksConflictsAndRetainsEntries(t *testing.T) {
	f := newContractFixture(t)
	f.config.ExhibitionService = application.NewExhibitionService(f.db)
	f.router = NewRouter(f.config)
	response := f.call(t, "POST", "/api/v1/exhibition-lists", map[string]any{"designation": "Original event", "date": "2026-11-09", "endDate": "2026-11-10"})
	assertStatus(t, response, 201)
	id := responseID(t, response)
	base := "/api/v1/exhibition-lists/" + id
	assertStatus(t, f.call(t, "GET", base+"/entries", nil), 200)
	assertStatus(t, f.call(t, "POST", base+"/conflicts/check", nil), 200)
	assertStatus(t, f.call(t, "GET", "/api/v1/exhibition-lists/missing/entries", nil), 404)
	assertStatus(t, f.call(t, "POST", "/api/v1/exhibition-lists/missing/conflicts/check", nil), 404)
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "POST", base+"/conflicts/check", nil, true), 403)
}
