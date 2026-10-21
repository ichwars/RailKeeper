package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/decoderlibrary"
	"railkeeper/backend/internal/infrastructure"
)

func TestDecoderLibraryAPIAccessAndActivation(t *testing.T) {
	db := testRouterDB(t)
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	setup := application.NewSetupService(db)
	router := NewRouter(Config{SetupService: setup, AuthService: application.NewAuthService(db), MasterDataService: application.NewMasterDataService(db)})
	if err := setup.CreateAdmin(t.Context(), application.CreateAdminInput{Username: "admin", Email: "admin@example.test", Password: "test-password-very-secure"}); err != nil {
		t.Fatal(err)
	}
	admin, cookies := loginTestUser(t, router, "admin", "test-password-very-secure")
	doJSON(t, router, http.MethodGet, "/api/v1/decoder-library", "", nil, http.StatusUnauthorized)
	doJSON(t, router, http.MethodGet, "/api/v1/decoder-library/references/kuehn-n-n025", "", nil, http.StatusUnauthorized)
	for _, role := range []string{"Viewer", "Messe"} {
		body := `{"username":"` + role + `","email":"` + role + `@example.test","password":"test-password-very-secure","roles":["` + role + `"]}`
		doAuthedJSON(t, router, http.MethodPost, "/api/v1/users", body, admin, cookies, http.StatusCreated)
		session, userCookies := loginTestUser(t, router, role, "test-password-very-secure")
		expected := http.StatusOK
		if role == "Messe" {
			expected = http.StatusForbidden
		}
		doAuthedJSON(t, router, http.MethodGet, "/api/v1/decoder-library", "", session, userCookies, expected)
		doAuthedJSON(t, router, http.MethodGet, "/api/v1/decoder-library/references/kuehn-n-n025", "", session, userCookies, expected)
		doAuthedJSON(t, router, http.MethodPatch, "/api/v1/master-data/decoder_reference/cv-ref-kuehn-n025/active", `{"active":false}`, session, userCookies, http.StatusForbidden)
		doAuthedJSON(t, router, http.MethodPatch, "/api/v1/master-data/decoder_profile/dcc-basics/active", `{"active":false}`, session, userCookies, http.StatusForbidden)
	}
	doAuthedJSON(t, router, http.MethodPatch, "/api/v1/master-data/decoder_profile/esu-lokpilot-5-micro-dcc/active", `{"active":false}`, admin, cookies, http.StatusOK)
	response := doAuthedJSON(t, router, http.MethodGet, "/api/v1/decoder-library", "", admin, cookies, http.StatusOK)
	var catalog decoderlibrary.Catalog
	if err := json.Unmarshal(response.Body.Bytes(), &catalog); err != nil || len(catalog.Profiles) != len(decoderlibrary.Bundled().Profiles)-1 {
		t.Fatalf("active catalog: %s %v", response.Body.String(), err)
	}
	if len(catalog.ReferenceCoverage) != len(decoderlibrary.BundledReferences()) {
		t.Fatal("catalog did not expose bounded reference coverage")
	}
	response = doAuthedJSON(t, router, http.MethodGet, "/api/v1/decoder-library/references/kuehn-n-n025", "", admin, cookies, http.StatusOK)
	var references []decoderlibrary.CVReference
	if err := json.Unmarshal(response.Body.Bytes(), &references); err != nil || len(references) != 1 || len(references[0].Rows) == 0 {
		t.Fatal("model reference rows not available")
	}
	response = doAuthedJSON(t, router, http.MethodGet, "/api/v1/decoder-library/references/unknown-model", "", admin, cookies, http.StatusOK)
	if err := json.Unmarshal(response.Body.Bytes(), &references); err != nil || len(references) != 0 {
		t.Fatal("unknown model did not return an empty reference list")
	}
	response = doAuthedJSON(t, router, http.MethodGet, "/api/v1/decoder-library?includeInactive=true", "", admin, cookies, http.StatusOK)
	if err := json.Unmarshal(response.Body.Bytes(), &catalog); err != nil || len(catalog.Profiles) != len(decoderlibrary.Bundled().Profiles) {
		t.Fatalf("complete catalog: %s %v", response.Body.String(), err)
	}
}

func TestDecoderImportPreviewAddsLibraryMatchWithoutChangingReadings(t *testing.T) {
	db := testRouterDB(t)
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	a := &App{masterDataService: application.NewMasterDataService(db)}
	source := []byte("Decoder: LokPilot 5 micro DCC (5.14.183)\nFilename: decoder.esux\nCreated: 10/17/2026 17:41:59\nCV 1 = 3\nCV 7 = 255\nCV 8 = 151\nIndex: 4099 (CV31=16, CV32=3)\nCV 257 = 20")
	preview := esuxPreviewResponse("decoder.txt", int64(len(source)), "text/plain", source)
	before, _ := json.Marshal(preview.SuggestedCVValues)
	a.enrichDecoderPreview(t.Context(), &preview)
	after, _ := json.Marshal(preview.SuggestedCVValues)
	match := preview.DecoderLibraryMatches["LokPilot 5 micro DCC"]
	if string(before) != string(after) || match.Status != "matched" || match.TotalValues != 4 || match.Firmware != "5.14.183" {
		t.Fatalf("preview: %+v", preview)
	}
}

func TestLenzPreviewUsesReviewedProfileWithoutChangingRawValues(t *testing.T) {
	db := testRouterDB(t)
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	app := &App{masterDataService: application.NewMasterDataService(db)}
	for _, code := range []string{"77", "79"} {
		source := "CV,Wert,decoderProfile,indexCV31,indexCV32\n8,99,Zentrale,,\n7," + code +
			",Zentrale,,\n51,128,Zentrale,,\n49,0,Zentrale,,\n145,16,Zentrale,,\n51,3,Zentrale,0,0"
		preview := esuxPreviewResponse("lenz.csv", int64(len(source)), "text/csv", []byte(source))
		before, _ := json.Marshal(preview.SuggestedCVValues)
		app.enrichDecoderPreview(t.Context(), &preview)
		after, _ := json.Marshal(preview.SuggestedCVValues)
		match := preview.DecoderLibraryMatches["Zentrale"]
		if match.Status != "matched" || match.ProfileID != "lenz-standard-plus-v2" ||
			match.Revision != "2026-10-20.2" || match.KnownValues != 5 || match.TotalValues != 6 ||
			match.Firmware != "" || string(before) != string(after) {
			t.Fatalf("Lenz import evidence or raw values changed: %+v", match)
		}
	}
}

func TestAnonymousProfilePreviewRecognizesDocumentedTypeCVs(t *testing.T) {
	db := testRouterDB(t)
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	app := &App{masterDataService: application.NewMasterDataService(db)}
	for _, test := range []struct{ source, status, name string }{
		{"CV,Wert\n8,97\n261,180\n27,3", "matched", "DH18A"},
		{"CV,Wert\n8,97\n261,52\n48,5", "matched", "DH05C"},
		{"CV,Wert\n8,97\n261,102\n48,5", "matched", "DH10C"},
		{"CV,Wert\n8,97\n261,160\n48,5", "matched", "DH16A"},
		{"CV,Wert\n8,97\n261,120\n48,5", "matched", "DH12A"},
		{"CV,Wert\n8,97\n261,141\n48,5", "matched", "DH14B"},
		{"CV,Wert\n8,97\n261,200\n48,5", "matched", "DH21A"},
		{"CV,Wert\n8,97\n261,201\n48,5", "matched", "DH21B"},
		{"CV,Wert\n8,97\n261,202\n48,5", "matched", "DH22A"},
		{"CV,Wert\n8,97\n261,203\n48,5", "matched", "DH22B"},
		{"CV,Wert\n8,97\n261,204\n48,5", "matched", "DH24A"},
		{"CV,Wert\n8,97\n261,110\n48,5", "matched", "DHSP10A"},
		{"CV,Wert\n8,145\n250,127", "family", "MN180"},
		{"CV,Wert\n8,99\n7,77", "matched", "STANDARD+ V2"},
	} {
		preview := esuxPreviewResponse("central-export.csv", int64(len(test.source)), "text/csv", []byte(test.source))
		if len(preview.SuggestedCVValues) == 0 {
			t.Fatalf("CV source not parsed: %+v", preview)
		}
		before, _ := json.Marshal(preview.SuggestedCVValues)
		app.enrichDecoderPreview(t.Context(), &preview)
		match := preview.DecoderLibraryMatches[preview.SuggestedDecoderProfile]
		if match.Status != test.status || match.Identification == nil || match.Identification.Name != test.name {
			t.Fatalf("unexpected identity: %+v", match)
		}
		after, _ := json.Marshal(preview.SuggestedCVValues)
		if string(before) != string(after) {
			t.Fatal("recognition rewrote CV rows")
		}
	}
}

func TestCSVPreviewKeepsProfileAndProtocolBoundaries(t *testing.T) {
	db := testRouterDB(t)
	if err := infrastructure.SeedMasterData(db, "../../seeds"); err != nil {
		t.Fatal(err)
	}
	app := &App{masterDataService: application.NewMasterDataService(db)}
	source := "CV,Wert,decoderProfile,protocol\n8,97,DH export,DCC\n261,180,DH export,DCC\n8,99,Lenz export,DCC\n7,77,Lenz export,DCC\n8,97,MM export,Motorola\n261,180,MM export,Motorola"
	preview := esuxPreviewResponse("central.csv", int64(len(source)), "text/csv", []byte(source))
	app.enrichDecoderPreview(t.Context(), &preview)
	if len(preview.SuggestedCVValues) != 6 || len(preview.DecoderLibraryMatches) != 3 {
		t.Fatalf("CSV profile groups lost: %+v", preview)
	}
	for profile, status := range map[string]string{"DH export": "matched", "Lenz export": "matched", "MM export": "unsupported"} {
		if preview.DecoderLibraryMatches[profile].Status != status {
			t.Fatalf("profile/protocol boundary lost for %s: %+v", profile, preview.DecoderLibraryMatches[profile])
		}
	}
	if preview.SuggestedCVValues[4].Protocol != "Motorola" || preview.SuggestedCVValues[4].DecoderProfile != "MM export" {
		t.Fatal("CSV row protocol/profile changed")
	}
}
