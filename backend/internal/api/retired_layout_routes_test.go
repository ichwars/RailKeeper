package api

import (
	"net/http"
	"testing"
)

func TestRetiredPlanningRoutesAreAbsent(t *testing.T) {
	router, sessions, _, _, _ := accessoryAllocationRouter(t)
	for _, path := range []string{
		"/api/v1/layouts", "/api/v1/layout-units/old/ports",
		"/api/v1/plan-revisions/old/track-plan", "/api/v1/track-libraries",
	} {
		for _, method := range []string{http.MethodGet, http.MethodPost} {
			response := authedJSONRequest(t, router, sessions["editor"], method, path, nil, true)
			assertStatus(t, response, http.StatusNotFound)
		}
	}
}

func TestAccessoryWritesRejectRetiredTargetFields(t *testing.T) {
	router, sessions, product, location, vehicle := accessoryAllocationRouter(t)
	for _, field := range []string{"layoutId", "layoutUnitId", "technicalPositionId"} {
		for _, path := range []string{"/api/v1/accessory-reservations", "/api/v1/accessory-installations"} {
			body := map[string]any{"productId": product.ID, "quantity": 1,
				"vehicleId": vehicle.ID, field: "retired-target"}
			if path == "/api/v1/accessory-reservations" {
				body["locationId"] = location.ID
			} else {
				body["sourceLocationId"] = location.ID
			}
			response := authedJSONRequest(t, router, sessions["editor"], http.MethodPost, path, body, true)
			assertProblem(t, response, http.StatusBadRequest, "invalid_json")
		}
	}
}
