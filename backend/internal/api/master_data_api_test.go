package api

import (
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

	doAuthedJSON(t, router, http.MethodPost, "/api/v1/master-data/article_type",
		`{"key":"custom","label":"Custom","active":true}`, session, cookies, http.StatusBadRequest)
	doAuthedJSON(t, router, http.MethodPut, "/api/v1/master-data/article_type/track",
		`{"key":"renamed","label":"Gleismaterial","active":true}`, session, cookies, http.StatusBadRequest)
	doAuthedJSON(t, router, http.MethodPut, "/api/v1/master-data/article_type/track",
		`{"label":"Gleismaterial","active":false}`, session, cookies, http.StatusOK)
	doAuthedJSON(t, router, http.MethodDelete, "/api/v1/master-data/article_type/track",
		"", session, cookies, http.StatusBadRequest)
}
