package infrastructure_test

import (
	"testing"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalCenterReadCacheRetainsSuccessfulReadsUntilSuccessfulReplacement(t *testing.T) {
	db := testDB(t)
	repository := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	if saved, err := repository.LatestReadySession(t.Context(), "ecos", "center.local", 15471); err != nil || saved != nil {
		t.Fatalf("empty cache = %#v, %v", saved, err)
	}
	baseline := application.DigitalCenterReadSession{
		Provider: "ecos", Host: "center.local", Port: 15471,
		State:           application.DigitalCenterSessionReady,
		Capabilities:    application.DigitalCenterCapabilities{ReadLocomotives: true},
		ReadCompletedAt: "2026-11-08T10:00:00Z",
	}
	first, err := repository.CreateSession(t.Context(), baseline)
	if err != nil {
		t.Fatal(err)
	}
	for _, change := range []struct {
		name  string
		apply func(*application.DigitalCenterReadSession)
	}{
		{"failed", func(s *application.DigitalCenterReadSession) { s.State = application.DigitalCenterSessionFailed }},
		{"interrupted", func(s *application.DigitalCenterReadSession) { s.State = application.DigitalCenterSessionInterrupted }},
		{"reading", func(s *application.DigitalCenterReadSession) { s.State = application.DigitalCenterSessionReading }},
		{"other-provider", func(s *application.DigitalCenterReadSession) { s.Provider = "z21" }},
		{"other-host", func(s *application.DigitalCenterReadSession) { s.Host = "other.local" }},
		{"other-port", func(s *application.DigitalCenterReadSession) { s.Port = 21105 }},
		{"incomplete", func(s *application.DigitalCenterReadSession) { s.ReadCompletedAt = "" }},
		{"offline-backup", func(s *application.DigitalCenterReadSession) { s.Capabilities.ReadLocomotives = false }},
	} {
		t.Run(change.name, func(t *testing.T) {
			session := baseline
			session.ReadCompletedAt = "2026-11-08T11:00:00Z"
			change.apply(&session)
			if _, err := repository.CreateSession(t.Context(), session); err != nil {
				t.Fatal(err)
			}
			// A new adapter has no in-memory state and must recover the same stored result.
			reopened := infrastructure.NewDigitalCenterWorkspaceRepository(db)
			saved, err := reopened.LatestReadySession(t.Context(), "ecos", "center.local", 15471)
			if err != nil || saved == nil || saved.ID != first.ID {
				t.Fatalf("previous successful cache = %#v, %v", saved, err)
			}
		})
	}
	second, err := repository.CreateSession(t.Context(), baseline)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := repository.LatestReadySession(t.Context(), "ecos", "center.local", 15471)
	if err != nil || saved == nil || saved.ID != second.ID {
		t.Fatalf("new completed read at the same timestamp = %#v, %v", saved, err)
	}
	retained, err := repository.GetSession(t.Context(), first.ID)
	if err != nil || retained != first {
		t.Fatalf("original read was changed: %#v, %v", retained, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.LatestReadySession(t.Context(), "ecos", "center.local", 15471); err == nil {
		t.Fatal("storage failure must not be reported as an empty cache")
	}
}
