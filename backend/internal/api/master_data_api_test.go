package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
)

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
