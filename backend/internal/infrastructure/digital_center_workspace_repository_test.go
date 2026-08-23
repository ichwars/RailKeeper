package infrastructure_test

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"railkeeper/backend/internal/application"
	"railkeeper/backend/internal/infrastructure"
)

func TestDigitalCenterWorkspaceRepositoryPersistsSessionAndItems(t *testing.T) {
	db := testDB(t)
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos",
		State:    application.DigitalCenterSessionReading,
		Host:     "192.168.2.151",
		Port:     15471,
		Capabilities: application.DigitalCenterCapabilities{
			ReadLocomotives: true,
			LiveMonitor:     true,
		},
		CreatedByUserID: "admin-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	err = repo.ReplaceWorkItems(t.Context(), session.ID, []application.DigitalCenterWorkItem{{
		CenterObjectID: "3",
		VehicleID:      "vehicle-1",
		Name:           "BR 218 402-6",
		Address:        3,
		Protocol:       "DCC",
		CompareStatus:  application.DigitalCompareNew,
		StationStatus:  "active",
		Center:         map[string]any{"speedSteps": 128},
		RailKeeper:     map[string]any{"name": "BR 218 402-6"},
		Proposed:       map[string]any{"decoderAddress": 3},
		Conflicts:      []map[string]any{{"field": "name"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListWorkItems(t.Context(), session.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	item := items[0]
	if item.ID == "" || item.SessionID != session.ID || item.Address != 3 ||
		item.Center["speedSteps"] != float64(128) || item.RailKeeper["name"] != "BR 218 402-6" ||
		item.Proposed["decoderAddress"] != float64(3) || len(item.Conflicts) != 1 {
		t.Fatalf("unexpected persisted item: %#v", item)
	}
	loaded, err := repo.GetWorkItem(t.Context(), session.ID, item.ID)
	if err != nil || loaded.ID != item.ID {
		t.Fatalf("item=%#v err=%v", loaded, err)
	}
	loadedSession, err := repo.GetSession(t.Context(), session.ID)
	if err != nil || !loadedSession.Capabilities.ReadLocomotives || loadedSession.CreatedByUserID != "admin-1" {
		t.Fatalf("session=%#v err=%v", loadedSession, err)
	}
}

func TestDigitalCenterWorkspaceRepositoryReturnsEmptyCollections(t *testing.T) {
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReading,
	})
	if err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListWorkItems(t.Context(), session.ID)
	if err != nil {
		t.Fatal(err)
	}
	messages, err := repo.ListMessages(t.Context(), session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if items == nil || len(items) != 0 || messages == nil || len(messages) != 0 {
		t.Fatalf("nil collections: items=%#v messages=%#v", items, messages)
	}
}

func TestDigitalCenterWorkspaceRepositoryUpdatesSession(t *testing.T) {
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReading,
	})
	if err != nil {
		t.Fatal(err)
	}
	session.State = application.DigitalCenterSessionReady
	session.ReadStartedAt = "2026-08-21T10:00:00Z"
	session.ReadCompletedAt = "2026-08-21T10:00:02Z"
	session.Capabilities.Diagnose = true
	session.UpdatedAt = "2000-01-01T00:00:00Z"
	if err := repo.UpdateSession(t.Context(), session); err != nil {
		t.Fatal(err)
	}
	loaded, err := repo.GetSession(t.Context(), session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.State != application.DigitalCenterSessionReady || loaded.ReadCompletedAt == "" ||
		!loaded.Capabilities.Diagnose || loaded.UpdatedAt == session.UpdatedAt {
		t.Fatalf("unexpected updated session: %#v", loaded)
	}
}

func TestDigitalCenterWorkspaceRepositoryUpdatesWorkItem(t *testing.T) {
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReady,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceWorkItems(t.Context(), session.ID, []application.DigitalCenterWorkItem{{
		CenterObjectID: "3", Name: "Alt", Address: 3, Protocol: "DCC",
		CompareStatus: application.DigitalCompareDeviation, StationStatus: "read",
		Center: map[string]any{"name": "Alt"}, RailKeeper: map[string]any{"name": "Neu"},
	}}); err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListWorkItems(t.Context(), session.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	item := items[0]
	item.Name = "Neu"
	item.Address = 18
	item.Center = map[string]any{"name": "Neu", "decoderAddress": 18}
	item.CompareStatus = application.DigitalCompareOK
	item.Conflicts = []map[string]any{}
	updated, err := repo.UpdateWorkItem(t.Context(), item)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Neu" || updated.Address != 18 || updated.CompareStatus != application.DigitalCompareOK ||
		updated.Center["decoderAddress"] != float64(18) || len(updated.Conflicts) != 0 ||
		updated.UpdatedAt == "" {
		t.Fatalf("updated item=%#v", updated)
	}
}

func TestDigitalCenterWorkspaceRepositoryEnforcesForeignKeysUniquenessAndCascade(t *testing.T) {
	db := testDB(t)
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	if err := repo.ReplaceWorkItems(t.Context(), "missing", []application.DigitalCenterWorkItem{{
		CenterObjectID: "3", CompareStatus: application.DigitalCompareNew,
	}}); err == nil || !strings.Contains(err.Error(), "replace digital center work items") {
		t.Fatalf("foreign key error = %v", err)
	}
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReading,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceWorkItems(t.Context(), session.ID, []application.DigitalCenterWorkItem{
		{CenterObjectID: "3", CompareStatus: application.DigitalCompareNew},
		{CenterObjectID: "3", CompareStatus: application.DigitalCompareDeviation},
	}); err == nil || !strings.Contains(err.Error(), "replace digital center work items") {
		t.Fatalf("uniqueness error = %v", err)
	}
	items, err := repo.ListWorkItems(t.Context(), session.ID)
	if err != nil || len(items) != 0 {
		t.Fatalf("failed replacement was not atomic: items=%#v err=%v", items, err)
	}
	if err := repo.ReplaceWorkItems(t.Context(), session.ID, []application.DigitalCenterWorkItem{{
		CenterObjectID: "3", CompareStatus: application.DigitalCompareNew,
	}}); err != nil {
		t.Fatal(err)
	}
	items, err = repo.ListWorkItems(t.Context(), session.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	if err := repo.AddMessage(t.Context(), application.DigitalCenterSessionMessage{
		SessionID: session.ID, Severity: application.DigitalCenterMessageInfo,
		Code: "read.completed", Message: "Lesen abgeschlossen.",
	}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateWriteGrant(t.Context(), application.DigitalCenterWriteGrant{
		TokenHash: "token", SessionID: session.ID, WorkItemID: items[0].ID, PreviewHash: "preview",
		ActorUserID: "admin-1", ExpiresAt: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(), `DELETE FROM digital_center_read_sessions WHERE id=?`, session.ID); err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{
		"digital_center_work_items", "digital_center_session_messages", "digital_center_write_grants",
	} {
		var count int
		if err := db.QueryRowContext(t.Context(), `SELECT COUNT(*) FROM `+table).Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("%s rows after session delete = %d", table, count)
		}
	}
}

func TestDigitalCenterWorkspaceRepositoryRejectsEmptyReplacementForMissingSession(t *testing.T) {
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	err := repo.ReplaceWorkItems(t.Context(), "missing-session", nil)
	if err == nil || !strings.Contains(err.Error(), "replace digital center work items") ||
		!errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("empty replacement for missing session error = %v", err)
	}
}

func TestDigitalCenterWorkspaceRepositoryRejectsMalformedJSON(t *testing.T) {
	db := testDB(t)
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(db)
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReading,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(t.Context(),
		`UPDATE digital_center_read_sessions SET capability_json='{"readLocomotives":true} trailing' WHERE id=?`,
		session.ID); err != nil {
		t.Fatal(err)
	}
	_, err = repo.GetSession(t.Context(), session.ID)
	if err == nil || !strings.Contains(err.Error(), "decode digital center capabilities") {
		t.Fatalf("malformed capabilities error = %v", err)
	}
}

func TestDigitalCenterWorkspaceRepositoryBoundsMessages(t *testing.T) {
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReading,
	})
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 105; index++ {
		if err := repo.AddMessage(t.Context(), application.DigitalCenterSessionMessage{
			SessionID: session.ID, Severity: application.DigitalCenterMessageInfo,
			Code: "read.completed", Message: "Lesen abgeschlossen.",
		}); err != nil {
			t.Fatal(err)
		}
	}
	messages, err := repo.ListMessages(t.Context(), session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 100 {
		t.Fatalf("messages = %d, want 100", len(messages))
	}
}

func TestDigitalCenterWorkspaceRepositoryRejectsPrivateProtocolMessages(t *testing.T) {
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReading,
	})
	if err != nil {
		t.Fatal(err)
	}
	unsafeMessages := []application.DigitalCenterSessionMessage{
		{Code: "queryObjects(10,name)", Message: "Lesen fehlgeschlagen"},
		{Code: "protocol", Message: "queryObjects(10, name)"},
		{Code: "protocol", Message: "queryObj\x00ects(10, name)"},
		{Code: "protocol", Message: "<REPLY get(10, name)>"},
		{Code: "protocol", Message: "Authorization: Bearer abc.def.ghi"},
		{Code: "protocol", Message: `{"PassWord" : "secret"}`},
		{Code: "protocol", Message: "pass\x00word = secret"},
		{Code: "protocol", Message: "Verbindung fehlgeschlagen", NextAction: "URL mit ?passwd=s3cr3t öffnen"},
		{Code: "protocol", Message: "Verbindung fehlgeschlagen", NextAction: "token = opaque"},
		{Code: "read.failed", Message: "release(1001, view)"},
		{Code: "read.failed", Message: "password hunter2"},
		{Code: "read.failed", Message: "BeA\u200bReR abc.def.ghi"},
		{Code: "read.failed", Message: "To\u200bKeN opaque"},
		{Code: "read.failed", Message: "API\u200b-\x00KEY hunter2"},
		{Code: "read.failed", Message: "Pass\u200bWoRd hunter2"},
	}
	for index, message := range unsafeMessages {
		message.SessionID = session.ID
		message.Severity = application.DigitalCenterMessageError
		err = repo.AddMessage(t.Context(), message)
		if err == nil || !strings.Contains(err.Error(), "raw or private protocol content") {
			t.Fatalf("unsafe message %d error = %v", index, err)
		}
	}
	messages, listErr := repo.ListMessages(t.Context(), session.ID)
	if listErr != nil || len(messages) != 0 {
		t.Fatalf("private message persisted: messages=%#v err=%v", messages, listErr)
	}
	safeMessages := []application.DigitalCenterSessionMessage{
		{Code: "connection.failed", Message: "  Verbindung\x00\r\n fehlgeschlagen.  ", NextAction: "Netzwerk prüfen."},
		{Code: "read.completed", Message: "42 Lokomotiven wurden gelesen."},
		{Code: "parse.failed", Message: "Die gelesenen Daten konnten nicht verarbeitet werden."},
		{Code: "capability.unavailable", Message: "Diese Zentrale unterstützt das Lesen nicht."},
		{Code: "write.unknown", Message: "Der Schreibstatus der Digitalzentrale ist unbekannt."},
		{Code: "live.restart_failed", Message: "Das Live-Monitoring konnte nicht neu gestartet werden."},
	}
	for _, message := range safeMessages {
		message.SessionID = session.ID
		message.Severity = application.DigitalCenterMessageWarning
		if err := repo.AddMessage(t.Context(), message); err != nil {
			t.Fatalf("safe structured message %#v: %v", message, err)
		}
	}
	messages, err = repo.ListMessages(t.Context(), session.ID)
	if err != nil || len(messages) != len(safeMessages) {
		t.Fatalf("safe messages=%#v err=%v", messages, err)
	}
	foundNormalized := false
	for _, message := range messages {
		if message.Code == "connection.failed" {
			foundNormalized = message.Message == "Verbindung fehlgeschlagen." &&
				message.NextAction == "Netzwerk prüfen."
		}
	}
	if !foundNormalized {
		t.Fatalf("safe control characters were not normalized: %#v", messages)
	}
}

func TestDigitalCenterWorkspaceRepositoryConsumesGrantOnceAtomically(t *testing.T) {
	repo, session, item := digitalCenterGrantFixture(t)
	grant := application.DigitalCenterWriteGrant{
		TokenHash: "hashed-token", SessionID: session.ID, WorkItemID: item.ID, PreviewHash: "preview",
		ActorUserID: "admin-1", ExpiresAt: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}
	if err := repo.CreateWriteGrant(t.Context(), grant); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			_, consumeErr := repo.ConsumeWriteGrant(t.Context(), grant.TokenHash, grant.ActorUserID)
			results <- consumeErr
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	successes := 0
	consumed := 0
	for err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, application.ErrDigitalCenterGrantConsumed):
			if !strings.Contains(err.Error(), "consume digital center write grant") {
				t.Fatalf("consumed grant error lacks context: %v", err)
			}
			consumed++
		default:
			t.Fatalf("unexpected consume error: %v", err)
		}
	}
	if successes != 1 || consumed != 1 {
		t.Fatalf("successes=%d consumed=%d", successes, consumed)
	}
}

func TestDigitalCenterWorkspaceRepositoryRejectsGrantForAnotherSessionItem(t *testing.T) {
	repo, firstSession, item := digitalCenterGrantFixture(t)
	secondSession, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReady,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = repo.CreateWriteGrant(t.Context(), application.DigitalCenterWriteGrant{
		TokenHash: "cross-session", SessionID: secondSession.ID, WorkItemID: item.ID,
		PreviewHash: "preview", ActorUserID: "admin-1",
		ExpiresAt: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	})
	if err == nil || !strings.Contains(err.Error(), "create digital center write grant") {
		t.Fatalf("cross-session grant error = %v", err)
	}
	if _, err := repo.ConsumeWriteGrant(t.Context(), "cross-session", "admin-1"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("cross-session grant was persisted: session=%s other=%s err=%v", firstSession.ID, secondSession.ID, err)
	}
}

func TestDigitalCenterWorkspaceRepositoryRejectsExpiredAndMismatchedGrants(t *testing.T) {
	repo, session, item := digitalCenterGrantFixture(t)
	if err := repo.CreateWriteGrant(t.Context(), application.DigitalCenterWriteGrant{
		TokenHash: "expired", SessionID: session.ID, WorkItemID: item.ID, PreviewHash: "preview",
		ActorUserID: "admin-1", ExpiresAt: time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConsumeWriteGrant(t.Context(), "expired", "admin-1"); !errors.Is(err, application.ErrDigitalCenterGrantExpired) {
		t.Fatalf("expired grant error = %v", err)
	} else if !strings.Contains(err.Error(), "consume digital center write grant") {
		t.Fatalf("expired grant error lacks context: %v", err)
	}
	if err := repo.CreateWriteGrant(t.Context(), application.DigitalCenterWriteGrant{
		TokenHash: "actor-bound", SessionID: session.ID, WorkItemID: item.ID, PreviewHash: "preview",
		ActorUserID: "admin-1", ExpiresAt: time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.ConsumeWriteGrant(t.Context(), "actor-bound", "admin-2"); !errors.Is(err, application.ErrDigitalCenterGrantActorMismatch) {
		t.Fatalf("actor mismatch error = %v", err)
	} else if !strings.Contains(err.Error(), "consume digital center write grant") {
		t.Fatalf("actor mismatch error lacks context: %v", err)
	}
}

func digitalCenterGrantFixture(
	t *testing.T,
) (*infrastructure.DigitalCenterWorkspaceRepository, application.DigitalCenterReadSession, application.DigitalCenterWorkItem) {
	t.Helper()
	repo := infrastructure.NewDigitalCenterWorkspaceRepository(testDB(t))
	session, err := repo.CreateSession(t.Context(), application.DigitalCenterReadSession{
		Provider: "ecos", State: application.DigitalCenterSessionReady,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceWorkItems(t.Context(), session.ID, []application.DigitalCenterWorkItem{{
		CenterObjectID: "3", CompareStatus: application.DigitalCompareDeviation,
	}}); err != nil {
		t.Fatal(err)
	}
	items, err := repo.ListWorkItems(t.Context(), session.ID)
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	return repo, session, items[0]
}
