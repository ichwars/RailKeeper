package application

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestMasterDataRefreshCannotLoseConcurrentUpdateInvalidation(t *testing.T) {
	db := masterDataCacheTestDB(t)
	before := masterDataCacheEntry("Before", true)
	insertMasterDataCacheEntry(t, db, before)

	refreshRead := make(chan struct{})
	releaseRefresh := make(chan struct{})
	var loaderMu sync.Mutex
	loads := 0
	service := newMasterDataService(db, func(ctx context.Context) (map[string][]MasterDataEntry, error) {
		loaderMu.Lock()
		loads++
		load := loads
		loaderMu.Unlock()
		if load == 1 {
			close(refreshRead)
			<-releaseRefresh
			return map[string][]MasterDataEntry{"test": {before}}, nil
		}
		return loadMasterDataSnapshot(ctx, db)
	})

	refreshDone := make(chan error, 1)
	go func() { refreshDone <- service.RefreshCache(context.Background()) }()
	<-refreshRead

	active := false
	updateDone := make(chan error, 1)
	go func() {
		_, err := service.Update(context.Background(), "test", "entry", MasterDataInput{
			Label: "After", Active: &active,
		})
		updateDone <- err
	}()
	waitForMasterDataLabel(t, db, "After")
	select {
	case err := <-updateDone:
		t.Fatalf("update invalidation escaped blocked refresh: %v", err)
	default:
	}

	close(releaseRefresh)
	if err := <-refreshDone; err != nil {
		t.Fatal(err)
	}
	if err := <-updateDone; err != nil {
		t.Fatal(err)
	}

	all, err := service.ListAll(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if got := all["test"]; len(got) != 1 || got[0].Label != "After" || got[0].Active {
		t.Fatalf("concurrent update was hidden by stale refresh cache: %#v", got)
	}
}

func TestMasterDataRefreshDerivesActiveEntriesFromOneSnapshot(t *testing.T) {
	entry := masterDataCacheEntry("Snapshot", true)
	loads := 0
	service := newMasterDataService(masterDataCacheTestDB(t), func(context.Context) (map[string][]MasterDataEntry, error) {
		loads++
		return map[string][]MasterDataEntry{"test": {entry}}, nil
	})
	if err := service.RefreshCache(context.Background()); err != nil {
		t.Fatal(err)
	}
	if loads != 1 {
		t.Fatalf("refresh loaded %d snapshots, want exactly one", loads)
	}
	all, _ := service.ListAll(context.Background(), false)
	active, _ := service.ListAll(context.Background(), true)
	if len(all["test"]) != 1 || len(active["test"]) != 1 ||
		all["test"][0].Label != active["test"][0].Label || !active["test"][0].Active {
		t.Fatalf("all and active caches came from different snapshots: all=%#v active=%#v", all, active)
	}
}

func TestMasterDataRefreshErrorPreservesPreviousCache(t *testing.T) {
	entry := masterDataCacheEntry("Cached", true)
	wantErr := errors.New("snapshot failed")
	loads := 0
	service := newMasterDataService(masterDataCacheTestDB(t), func(context.Context) (map[string][]MasterDataEntry, error) {
		loads++
		if loads > 1 {
			return nil, wantErr
		}
		return map[string][]MasterDataEntry{"test": {entry}}, nil
	})
	if err := service.RefreshCache(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.RefreshCache(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("refresh error=%v, want %v", err, wantErr)
	}
	all, err := service.ListAll(context.Background(), false)
	if err != nil {
		t.Fatal(err)
	}
	if loads != 2 || len(all["test"]) != 1 || all["test"][0].Label != "Cached" {
		t.Fatalf("failed refresh discarded previous cache: loads=%d cache=%#v", loads, all)
	}
}

func masterDataCacheTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", t.TempDir()+"/master-data.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`
CREATE TABLE master_data_entries (
  id TEXT PRIMARY KEY, type TEXT NOT NULL, key TEXT NOT NULL, label TEXT NOT NULL, active INTEGER NOT NULL,
  sort_order INTEGER NOT NULL, source_url TEXT, metadata_json TEXT NOT NULL, created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL, UNIQUE(type, key)
)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		t.Fatal(err)
	}
	return db
}

func masterDataCacheEntry(label string, active bool) MasterDataEntry {
	return MasterDataEntry{
		ID: "test:entry", Type: "test", Key: "entry", Label: label, Active: active,
		Metadata: map[string]any{}, CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
	}
}

func insertMasterDataCacheEntry(t *testing.T, db *sql.DB, entry MasterDataEntry) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, source_url, metadata_json, created_at, updated_at
) VALUES(?, ?, ?, ?, ?, ?, '', '{}', ?, ?)`, entry.ID, entry.Type, entry.Key, entry.Label,
		boolToInt(entry.Active), entry.SortOrder, entry.CreatedAt, entry.UpdatedAt); err != nil {
		t.Fatal(err)
	}
}

func waitForMasterDataLabel(t *testing.T, db *sql.DB, want string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		var label string
		if err := db.QueryRow(`SELECT label FROM master_data_entries WHERE type='test' AND key='entry'`).Scan(&label); err == nil && label == want {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("master-data label never became %q", want)
}
