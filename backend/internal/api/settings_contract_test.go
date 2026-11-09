package api

import (
	"net/http"
	"strings"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestSettingsHTTPContractsRoundTripAndIsolateUsers(t *testing.T) {
	f := newContractFixture(t)
	profile := "/api/v1/profile/settings"
	digital := "/api/v1/system/digital-settings"
	input := map[string]any{"settings": map[string]string{"railkeeper.settings.theme": "dark", "railkeeper.sidebar.Admin": "[\"vehicles\"]"}}
	assertStatus(t, f.call(t, "PUT", profile, input), 200)
	var saved application.SettingsPayload
	decodeResponse(t, f.call(t, "GET", profile, nil), &saved)
	if saved.Settings["railkeeper.settings.theme"] != "dark" {
		t.Fatalf("settings not persisted: %#v", saved)
	}
	var other application.SettingsPayload
	decodeResponse(t, authedJSONRequest(t, f.router, f.viewer, "GET", profile, nil, true), &other)
	if len(other.Settings) != 0 {
		t.Fatalf("settings leaked between users: %#v", other)
	}
	assertProblem(t, f.call(t, "PUT", profile, map[string]any{"settings": map[string]string{"": "invalid"}}), 400, "settings_validation")
	assertProblem(t, f.raw(t, "PUT", profile, "{"), 400, "invalid_json")
	assertStatus(t, f.call(t, "GET", digital, nil), 200)
	d := map[string]any{"provider": "ecos", "ecos": map[string]any{"enabled": true, "host": "192.0.2.1", "port": "15471"}}
	assertStatus(t, f.call(t, "PUT", digital, d), 200)
	var settings application.DigitalCenterSettings
	decodeResponse(t, f.call(t, "GET", digital, nil), &settings)
	if settings.Provider != "ecos" || settings.ECoS.Host != "192.0.2.1" {
		t.Fatalf("digital configuration lost: %#v", settings)
	}
	assertProblem(t, f.call(t, "PUT", digital, map[string]any{"provider": "unknown"}), 400, "settings_validation")
	assertProblem(t, f.raw(t, "PUT", digital, "{"), 400, "invalid_json")
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, "PUT", digital, d, true), 403)
	f.config.SettingsService = nil
	f.router = NewRouter(f.config)
	for _, path := range []string{profile, digital} {
		assertStatus(t, f.call(t, "GET", path, nil), 200)
		assertProblem(t, f.call(t, "PUT", path, input), 503, "settings_unavailable")
	}
}

func TestSMTPHTTPContractsNeverReturnStoredPassword(t *testing.T) {
	f := newContractFixture(t)
	path := "/api/v1/system/smtp"
	assertStatus(t, f.call(t, "GET", path, nil), 200)
	assertProblem(t, f.raw(t, "PUT", path, "{"), 400, "invalid_json")
	assertProblem(t, f.raw(t, "POST", path+"/test", "{"), 400, "invalid_json")
	assertProblem(t, f.call(t, "POST", path+"/test", map[string]string{"recipient": "test@example.test"}), 400, "smtp_disabled")
	assertProblem(t, f.call(t, "PUT", path, map[string]any{"enabled": true}), 400, "smtp_settings_validation")
	input := map[string]any{"enabled": false, "host": "mail.example.test", "port": "587", "from": "test@example.test", "tlsMode": "starttls", "password": "synthetic-private-value"}
	for _, method := range []string{"PUT", "GET"} {
		response := f.call(t, method, path, input)
		assertStatus(t, response, 200)
		if strings.Contains(response.Body.String(), "synthetic-private-value") {
			t.Fatal("SMTP response disclosed password")
		}
		var result application.SMTPSettings
		decodeResponse(t, response, &result)
		if !result.PasswordConfigured {
			t.Fatal("configured password was not retained")
		}
	}
	input["password"] = ""
	input["clearPassword"] = true
	var cleared application.SMTPSettings
	response := f.call(t, "PUT", path, input)
	assertStatus(t, response, 200)
	decodeResponse(t, response, &cleared)
	if cleared.PasswordConfigured {
		t.Fatal("explicit password removal was ignored")
	}
	assertStatus(t, authedJSONRequest(t, f.router, f.viewer, http.MethodGet, path, nil, true), 403)
	f.config.SMTPSettingsService = nil
	f.router = NewRouter(f.config)
	assertStatus(t, f.call(t, "GET", path, nil), 200)
	for _, method := range []string{"PUT", "POST"} {
		endpoint := path
		if method == "POST" {
			endpoint += "/test"
		}
		assertProblem(t, f.call(t, method, endpoint, input), 503, "smtp_settings_unavailable")
	}
}
