package application_test

import (
	"context"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestUserSettingsRemainIsolatedAndMergePartialUpdates(t *testing.T) {
	db := testDB(t)
	service := application.NewSettingsService(db)
	ctx := context.Background()

	for _, id := range []string{"user-a", "user-b"} {
		if _, err := db.Exec(`
INSERT INTO users(id, username, password_hash, created_at)
VALUES(?, ?, 'hash', '2026-08-15T00:00:00Z')
`, id, id); err != nil {
			t.Fatal(err)
		}
	}

	_, err := service.UpdateUserSettings(ctx, "user-a", application.SettingsPayload{Settings: map[string]string{
		"railkeeper.vehicles.tableColumns": `["series","inventoryNumber"]`,
		"theme":                            "dark",
	}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.UpdateUserSettings(ctx, "user-a", application.SettingsPayload{Settings: map[string]string{
		"railkeeper.vehicles.tableColumns": `["inventoryNumber"]`,
	}})
	if err != nil {
		t.Fatal(err)
	}

	userA, err := service.UserSettings(ctx, "user-a")
	if err != nil {
		t.Fatal(err)
	}
	userB, err := service.UserSettings(ctx, "user-b")
	if err != nil {
		t.Fatal(err)
	}
	if userA.Settings["theme"] != "dark" ||
		userA.Settings["railkeeper.vehicles.tableColumns"] != `["inventoryNumber"]` {
		t.Fatalf("unexpected merged settings: %#v", userA.Settings)
	}
	if len(userB.Settings) != 0 {
		t.Fatalf("settings leaked to user-b: %#v", userB.Settings)
	}
}
