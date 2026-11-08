package application_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/configuration"
	"railkeeper/backend/internal/infrastructure/settingsdata"
)

func TestConfigurationSMTPFallbackPasswordRetentionAndExplicitClearing(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	fallback := application.SMTPPasswordResetMailConfig{Host: " smtp.fallback.test ", Username: " fallback-user ", Password: "  synthetic fallback credential  ", From: "sender@example.test"}
	service := application.NewSMTPSettingsService(db, fallback, " https://railkeeper.test/// ")
	initial, err := service.Get(t.Context())
	if err != nil || !initial.Enabled || initial.Host != "smtp.fallback.test" || initial.Port != "587" || initial.TLSMode != "starttls" || initial.PublicURL != "https://railkeeper.test" || !initial.PasswordConfigured {
		t.Fatal("environment fallback changed", err)
	}
	mailer, url, err := service.EffectiveMailer(t.Context())
	if err != nil || mailer == nil || url != initial.PublicURL {
		t.Fatal("fallback mailer construction failed", err)
	}
	input := application.SMTPSettingsInput{Enabled: true, Host: " smtp.saved.test ", Port: "2525", Username: " saved-user ", From: "saved@example.test", PublicURL: "https://saved.test/", TLSMode: " NONE "}
	saved, err := service.Update(t.Context(), input)
	if err != nil || saved.Host != "smtp.saved.test" || saved.Username != "saved-user" || saved.PublicURL != "https://saved.test" || saved.TLSMode != "none" || !saved.PasswordConfigured {
		t.Fatal("saved SMTP configuration changed", err)
	}
	password := func() string {
		t.Helper()
		var value string
		if err := db.QueryRow(`SELECT value FROM app_settings WHERE key='smtp.password'`).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	if password() != fallback.Password {
		t.Fatal("blank input did not retain exact fallback password")
	}
	body, err := json.Marshal(saved)
	if err != nil {
		t.Fatal(err)
	}
	var exposed map[string]json.RawMessage
	if err := json.Unmarshal(body, &exposed); err != nil {
		t.Fatal(err)
	}
	if _, exists := exposed["password"]; exists || bytes.Contains(body, []byte(fallback.Password)) {
		t.Fatal("SMTP response exposed a password")
	}
	input.Password = "  synthetic replacement credential  "
	if _, err := service.Update(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	if password() != input.Password {
		t.Fatal("password whitespace was normalized")
	}
	input.ClearPassword = true
	cleared, err := service.Update(t.Context(), input)
	if err != nil || cleared.PasswordConfigured || password() != "" {
		t.Fatal("explicit clear did not win over supplied password", err)
	}
	input = application.SMTPSettingsInput{}
	disabled, err := service.Update(t.Context(), input)
	if err != nil || disabled.Enabled || disabled.PasswordConfigured {
		t.Fatal("stored disabled configuration revived fallback", err)
	}
	if mailer, _, err := service.EffectiveMailer(t.Context()); err != nil || mailer != nil {
		t.Fatal("disabled mailer was constructed", err)
	}
	if _, err := db.Exec(`DELETE FROM app_settings WHERE key LIKE 'smtp.%'`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO app_settings(key,value,updated_at)VALUES('smtp.enabled','false','recorded')`); err != nil {
		t.Fatal(err)
	}
	partial, err := service.Get(t.Context())
	if err != nil || partial.Host != "" || partial.Username != "" || partial.PasswordConfigured || partial.Enabled {
		t.Fatal("partial stored values were merged with environment credentials", err)
	}
}

func TestConfigurationSMTPValidationAndWriteFailurePreserveAllRows(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	service := application.NewSMTPSettingsService(db, application.SMTPPasswordResetMailConfig{}, "")
	valid := application.SMTPSettingsInput{Enabled: true, Host: "smtp.test", Port: "587", From: "sender@example.test", PublicURL: "https://railkeeper.test", TLSMode: "starttls", Password: "synthetic credential"}
	before := databaseRowFingerprints(t, db)
	for _, kind := range []string{"tls", "port", "host", "sender", "url"} {
		input := valid
		switch kind {
		case "tls":
			input.TLSMode = "unsupported"
		case "port":
			input.Port = "not-a-port"
		case "host":
			input.Host = ""
		case "sender":
			input.From = "invalid-sender"
		case "url":
			input.PublicURL = ""
		}
		if result, err := service.Update(t.Context(), input); result != nil || !errors.Is(err, application.ErrSMTPSettingsValidation) {
			t.Fatal("invalid SMTP configuration accepted", kind, err)
		}
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("validation modified settings or authentication")
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_smtp_setting BEFORE INSERT ON app_settings WHEN NEW.key='smtp.tls_mode'
BEGIN SELECT RAISE(FAIL,'synthetic SMTP write failure'); END`); err != nil {
		t.Fatal(err)
	}
	if result, err := service.Update(t.Context(), valid); result != nil || err == nil || !strings.Contains(err.Error(), "synthetic SMTP write failure") {
		t.Fatal("write failure not observed", err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || db.Stats().InUse != 0 {
		t.Fatal("failed SMTP write left partial rows or a transaction")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_smtp_setting`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(t.Context(), valid); err != nil {
		t.Fatal("SMTP retry failed", err)
	}
}

type configurationReadFault struct {
	configuration.Repository
	reads int
}

func (r *configurationReadFault) SMTPSettings(ctx context.Context) (map[string]string, error) {
	r.reads++
	if r.reads == 2 {
		return nil, errors.New("synthetic postcommit read failure")
	}
	return r.Repository.SMTPSettings(ctx)
}

func TestConfigurationLostSMTPReadbackRetainsCommittedSettingsAndAllowsRetry(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	repository := &configurationReadFault{Repository: settingsdata.New(db)}
	service := configuration.NewSMTPSettingsService(repository, configuration.SMTPConfig{}, "")
	input := configuration.SMTPSettingsInput{Enabled: true, Host: "smtp.test", From: "sender@example.test", PublicURL: "https://railkeeper.test", Password: "synthetic retained credential"}
	if result, err := service.Update(t.Context(), input); result != nil || err == nil || !strings.Contains(err.Error(), "postcommit read failure") {
		t.Fatal("lost readback was acknowledged", err)
	}
	var host string
	if err := db.QueryRow(`SELECT value FROM app_settings WHERE key='smtp.host'`).Scan(&host); err != nil || host != input.Host {
		t.Fatal("successful commit was undone after readback failure", err)
	}
	if _, err := service.Update(t.Context(), input); err != nil {
		t.Fatal("retry after lost readback failed", err)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM app_settings WHERE key LIKE 'smtp.%'`).Scan(&count); err != nil || count != 8 {
		t.Fatal("retry duplicated SMTP settings", count, err)
	}
}

func TestConfigurationCancellationPreservesPreferencesAndMaintenanceData(t *testing.T) {
	root := t.TempDir()
	db := backupTestDB(t, root)
	db.SetMaxOpenConns(1)
	settings := application.NewSettingsService(db)
	smtp := application.NewSMTPSettingsService(db, application.SMTPPasswordResetMailConfig{}, "")
	maintenance := application.NewDatabaseMaintenanceService(db, root)
	before := databaseRowFingerprints(t, db)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	checks := []func() error{
		func() error { _, err := settings.UserSettings(ctx, "user"); return err },
		func() error {
			_, err := settings.UpdateUserSettings(ctx, "user", application.SettingsPayload{})
			return err
		},
		func() error { _, err := settings.DigitalSettings(ctx); return err },
		func() error {
			_, err := settings.UpdateDigitalSettings(ctx, application.DigitalCenterSettings{})
			return err
		},
		func() error { _, err := smtp.Get(ctx); return err },
		func() error { _, err := smtp.Update(ctx, application.SMTPSettingsInput{}); return err },
		func() error { _, err := maintenance.Optimize(ctx); return err },
	}
	for i, check := range checks {
		if err := check(); !errors.Is(err, context.Canceled) {
			t.Fatal("cancellation was not propagated", i, err)
		}
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("cancelled operation changed rows")
	}
	result, err := maintenance.Optimize(t.Context())
	if err != nil || result == nil {
		t.Fatal("maintenance retry failed", err)
	}
	expected := result.BeforeBytes - result.AfterBytes
	if expected < 0 {
		expected = 0
	}
	if result.ReclaimedBytes != expected {
		t.Fatal("invalid reclaimed size")
	}
	if _, err := time.Parse(time.RFC3339, result.OptimizedAt); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
		t.Fatal("optimization changed stored rows or authentication")
	}
	badRoot := filepath.Join(root, "not-a-directory")
	if err := os.WriteFile(badRoot, []byte("fixture"), 0600); err != nil {
		t.Fatal(err)
	}
	if result, err := application.NewDatabaseMaintenanceService(db, badRoot).Optimize(t.Context()); result != nil || err == nil {
		t.Fatal("invalid storage root accepted")
	}
}

func TestConfigurationNilServicesRetainReadAndFallbackContracts(t *testing.T) {
	var settings *application.SettingsService
	value, err := settings.UserSettings(t.Context(), "user")
	if err != nil || value == nil || value.Settings == nil || len(value.Settings) != 0 {
		t.Fatal("nil preference read changed", err)
	}
	if _, err := settings.UpdateUserSettings(t.Context(), "user", application.SettingsPayload{}); !errors.Is(err, application.ErrSettingsValidation) {
		t.Fatal(err)
	}
	digital, err := settings.DigitalSettings(t.Context())
	if err != nil || digital.Provider != "ecos" || digital.ECoS.Port != "15471" || digital.Z21.Port != "21105" || digital.CS3.Port != "80" {
		t.Fatal("default digital settings changed", err)
	}
	if _, err := settings.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{}); !errors.Is(err, application.ErrSettingsValidation) {
		t.Fatal(err)
	}
	var smtp *application.SMTPSettingsService
	if value, err := smtp.Get(t.Context()); err != nil || value == nil || value.Enabled || value.PasswordConfigured {
		t.Fatal("nil SMTP read changed", err)
	}
	if mailer, url, err := smtp.EffectiveMailer(t.Context()); err != nil || mailer != nil || url != "" {
		t.Fatal("nil mailer changed", err)
	}
	fallback := application.NewSMTPSettingsService(nil, application.SMTPPasswordResetMailConfig{Host: "smtp.test", From: "sender@example.test"}, "https://railkeeper.test/")
	if mailer, url, err := fallback.EffectiveMailer(t.Context()); err != nil || mailer == nil || url != "https://railkeeper.test" {
		t.Fatal("DB-free fallback changed", err)
	}
	if _, err := fallback.Update(t.Context(), application.SMTPSettingsInput{}); !errors.Is(err, application.ErrSMTPSettingsValidation) {
		t.Fatal(err)
	}
	var maintenance *application.DatabaseMaintenanceService
	if result, err := maintenance.Optimize(t.Context()); result != nil || err == nil {
		t.Fatal("nil maintenance accepted")
	}
}

func TestConfigurationPreferenceValidationAndSQLFailureKeepExistingUserState(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`INSERT INTO users(id,username,password_hash,created_at)
VALUES('owner','owner','synthetic-hash','2026-01-01T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	service := application.NewSettingsService(db)
	initial := application.SettingsPayload{Settings: map[string]string{"theme": "dark", "navigation": `{"order":["vehicles","decoder"],"hidden":["exhibition"]}`}}
	if _, err := service.UpdateUserSettings(t.Context(), "owner", initial); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	for _, invalid := range []application.SettingsPayload{
		{Settings: map[string]string{"theme": "light", " ": "invalid"}},
		{Settings: map[string]string{"theme": "light", strings.Repeat("x", 161): "invalid"}},
		{Settings: map[string]string{"theme": "light", "invalid": strings.Repeat("x", 8001)}},
	} {
		if result, err := service.UpdateUserSettings(t.Context(), "owner", invalid); result != nil || !errors.Is(err, application.ErrSettingsValidation) {
			t.Fatal("invalid preferences accepted", err)
		}
		if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) {
			t.Fatal("invalid preference batch overwrote existing state")
		}
	}
	if _, err := service.UpdateDigitalSettings(t.Context(), application.DigitalCenterSettings{Provider: "unknown"}); !errors.Is(err, application.ErrSettingsValidation) {
		t.Fatal("invalid digital provider accepted", err)
	}
	if _, err := db.Exec(`CREATE TRIGGER fail_user_preference BEFORE INSERT ON user_settings WHEN NEW.key='theme'
BEGIN SELECT RAISE(FAIL,'synthetic preference failure'); END`); err != nil {
		t.Fatal(err)
	}
	update := application.SettingsPayload{Settings: map[string]string{" theme ": "light", "new-option": "preserved"}}
	if result, err := service.UpdateUserSettings(t.Context(), "owner", update); result != nil || err == nil || !strings.Contains(err.Error(), "synthetic preference failure") {
		t.Fatal("preference write failure not observed", err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || db.Stats().InUse != 0 {
		t.Fatal("failed preference write left partial state")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_user_preference`); err != nil {
		t.Fatal(err)
	}
	saved, err := service.UpdateUserSettings(t.Context(), "owner", update)
	if err != nil || saved.Settings["theme"] != "light" || saved.Settings["navigation"] != initial.Settings["navigation"] || saved.Settings["new-option"] != "preserved" {
		t.Fatal("preference retry lost existing navigation or key normalization", err)
	}
}

func TestConfigurationDeferredCommitFailurePreservesSettingsAndAllowsRetry(t *testing.T) {
	db := testDB(t)
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE settings_commit_guard(id TEXT PRIMARY KEY,user_id TEXT REFERENCES users(id) DEFERRABLE INITIALLY DEFERRED);
CREATE TRIGGER fail_settings_commit AFTER INSERT ON app_settings WHEN NEW.key='smtp.tls_mode'
BEGIN INSERT INTO settings_commit_guard(id,user_id)VALUES('guard','missing-user'); END`); err != nil {
		t.Fatal(err)
	}
	before := databaseRowFingerprints(t, db)
	service := application.NewSMTPSettingsService(db, application.SMTPPasswordResetMailConfig{}, "")
	input := application.SMTPSettingsInput{Enabled: true, Host: "smtp.test", From: "sender@example.test", PublicURL: "https://railkeeper.test"}
	if result, err := service.Update(t.Context(), input); result != nil || err == nil || !strings.Contains(err.Error(), "commit smtp settings update") {
		t.Fatal("deferred commit failure was not observed", err)
	}
	if !reflect.DeepEqual(before, databaseRowFingerprints(t, db)) || db.Stats().InUse != 0 {
		t.Fatal("failed settings commit retained partial data or a transaction")
	}
	if _, err := db.Exec(`DROP TRIGGER fail_settings_commit`); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Update(t.Context(), input); err != nil {
		t.Fatal("retry after failed commit did not succeed", err)
	}
}
