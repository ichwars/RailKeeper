package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"railkeeper/backend/internal/infrastructure"
	"railkeeper/backend/internal/startup"
)

func TestPrepareStartupConflictSkipsDatabaseMigrationsSeedsAndRouter(t *testing.T) {
	runtimeConfig := startup.RuntimeConfig{
		DataDir:       filepath.Join(t.TempDir(), "safe"),
		LegacyDataDir: filepath.Join(t.TempDir(), "legacy"),
		MigrationsDir: filepath.Join(t.TempDir(), "migrations"),
		SeedsDir:      filepath.Join(t.TempDir(), "seeds"),
	}
	calls := []string{}
	dependencies := startupDependencies{
		ResolveLegacyData: func(
			context.Context, string, string, startup.LegacyMigrationOptions,
		) (startup.LegacyMigrationResult, error) {
			calls = append(calls, "resolve")
			return startup.LegacyMigrationResult{
				Status:  startup.LegacyConflict,
				DataDir: runtimeConfig.DataDir,
				Conflict: &startup.LegacyConflictInfo{
					SafePath: runtimeConfig.DataDir, LegacyPath: runtimeConfig.LegacyDataDir,
				},
			}, nil
		},
		Preflight:  func() error { calls = append(calls, "preflight"); return nil },
		PathExists: func(string) bool { calls = append(calls, "exists"); return true },
		OpenSQLite: func(string) (*sql.DB, error) {
			calls = append(calls, "open")
			return nil, nil
		},
		MigrateSafely: func(
			context.Context, *sql.DB, string, string, infrastructure.MigrationSafetyOptions,
		) (infrastructure.MigrationSafetyResult, error) {
			calls = append(calls, "migrate")
			return infrastructure.MigrationSafetyResult{}, nil
		},
		SeedRoles:      func(*sql.DB) error { calls = append(calls, "seed-roles"); return nil },
		SeedMasterData: func(*sql.DB, string) error { calls = append(calls, "seed-master"); return nil },
		BuildHandler: func(
			context.Context, *sql.DB, applicationDataPaths, StartupState,
		) (http.Handler, error) {
			calls = append(calls, "router")
			return http.NotFoundHandler(), nil
		},
	}

	result, err := prepareStartup(context.Background(), runtimeConfig, "0.1.18", dependencies)
	if err != nil {
		t.Fatalf("prepareStartup() error = %v", err)
	}
	if result.Conflict == nil || result.Handler == nil || result.Database != nil {
		t.Fatalf("unexpected conflict result: %#v", result)
	}
	if !reflect.DeepEqual(calls, []string{"resolve"}) {
		t.Fatalf("conflict startup calls = %v, want only legacy resolution", calls)
	}
}

func TestPrepareStartupUsesResolvedSafePathEverywhereAndCapturesExistenceBeforeOpen(t *testing.T) {
	root := t.TempDir()
	safeDir := filepath.Join(root, "safe")
	legacyDir := filepath.Join(root, "legacy")
	runtimeConfig := startup.RuntimeConfig{
		DataDir: safeDir, LegacyDataDir: legacyDir,
		MigrationsDir: filepath.Join(root, "migrations"), SeedsDir: filepath.Join(root, "seeds"),
	}
	receipt := &startup.MigrationReceipt{SourcePath: legacyDir, TargetPath: safeDir}
	calls := []string{}
	var capturedPaths applicationDataPaths
	var capturedState StartupState
	dependencies := startupDependencies{
		ResolveLegacyData: func(
			_ context.Context, gotSafe, gotLegacy string, options startup.LegacyMigrationOptions,
		) (startup.LegacyMigrationResult, error) {
			calls = append(calls, "resolve")
			if gotSafe != safeDir || gotLegacy != legacyDir || options.Version != "0.1.18" {
				t.Fatalf("legacy inputs = %q, %q, %#v", gotSafe, gotLegacy, options)
			}
			return startup.LegacyMigrationResult{
				Status: startup.LegacyMigrated, DataDir: safeDir, Receipt: receipt,
			}, nil
		},
		Preflight: func() error {
			calls = append(calls, "preflight")
			return nil
		},
		PathExists: func(path string) bool {
			calls = append(calls, "exists")
			if path != filepath.Join(safeDir, "railkeeper.db") {
				t.Fatalf("database existence path = %q", path)
			}
			return true
		},
		OpenSQLite: func(dataDir string) (*sql.DB, error) {
			calls = append(calls, "open")
			if dataDir != safeDir {
				t.Fatalf("OpenSQLite data dir = %q", dataDir)
			}
			return nil, nil
		},
		MigrateSafely: func(
			_ context.Context,
			_ *sql.DB,
			dataDir string,
			migrationsDir string,
			options infrastructure.MigrationSafetyOptions,
		) (infrastructure.MigrationSafetyResult, error) {
			calls = append(calls, "migrate")
			if dataDir != safeDir || migrationsDir != runtimeConfig.MigrationsDir || !options.DatabaseExisted {
				t.Fatalf("migration inputs = %q, %q, %#v", dataDir, migrationsDir, options)
			}
			return infrastructure.MigrationSafetyResult{BackupPath: filepath.Join(safeDir, "safety.db")}, nil
		},
		SeedRoles: func(*sql.DB) error {
			calls = append(calls, "seed-roles")
			return nil
		},
		SeedMasterData: func(_ *sql.DB, seedsDir string) error {
			calls = append(calls, "seed-master")
			if seedsDir != runtimeConfig.SeedsDir {
				t.Fatalf("seed dir = %q", seedsDir)
			}
			return nil
		},
		BuildHandler: func(
			_ context.Context, _ *sql.DB, paths applicationDataPaths, state StartupState,
		) (http.Handler, error) {
			calls = append(calls, "router")
			capturedPaths = paths
			capturedState = state
			return http.NotFoundHandler(), nil
		},
	}

	result, err := prepareStartup(context.Background(), runtimeConfig, "0.1.18", dependencies)
	if err != nil {
		t.Fatalf("prepareStartup() error = %v", err)
	}
	if result.Conflict != nil || result.Handler == nil || result.State.Receipt != receipt {
		t.Fatalf("unexpected successful result: %#v", result)
	}
	if result.State.Runtime.DataDir != safeDir ||
		result.State.SafetyBackupPath != filepath.Join(safeDir, "safety.db") {
		t.Fatalf("unexpected startup state: %#v", result.State)
	}
	wantPaths := applicationDataPaths{
		DataDir: safeDir, DatabasePath: filepath.Join(safeDir, "railkeeper.db"),
		BlobDataDir: safeDir, BackupDataDir: safeDir, APIDataDir: safeDir,
	}
	if capturedPaths != wantPaths {
		t.Fatalf("application data paths = %#v, want %#v", capturedPaths, wantPaths)
	}
	if capturedState.Runtime.DataDir != safeDir || capturedState.Receipt != receipt {
		t.Fatalf("handler startup state = %#v", capturedState)
	}
	wantCalls := []string{
		"resolve", "preflight", "exists", "open", "migrate", "seed-roles", "seed-master", "router",
	}
	if !reflect.DeepEqual(calls, wantCalls) {
		t.Fatalf("successful startup calls = %v, want %v", calls, wantCalls)
	}
}

func TestPrepareStartupClosesDatabaseWhenLaterStartupFails(t *testing.T) {
	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	dependencies := startupDependencies{
		ResolveLegacyData: startup.ResolveLegacyData,
		Preflight:         func() error { return nil },
		PathExists:        func(string) bool { return true },
		OpenSQLite:        func(string) (*sql.DB, error) { return database, nil },
		MigrateSafely: func(
			context.Context, *sql.DB, string, string, infrastructure.MigrationSafetyOptions,
		) (infrastructure.MigrationSafetyResult, error) {
			return infrastructure.MigrationSafetyResult{}, nil
		},
		SeedRoles:      func(*sql.DB) error { return nil },
		SeedMasterData: func(*sql.DB, string) error { return nil },
		BuildHandler: func(
			context.Context, *sql.DB, applicationDataPaths, StartupState,
		) (http.Handler, error) {
			return nil, errors.New("router failed")
		},
	}

	_, err = prepareStartup(context.Background(), startup.RuntimeConfig{DataDir: t.TempDir()}, "0.1.18", dependencies)
	if err == nil {
		t.Fatal("expected handler construction failure")
	}
	if pingErr := database.PingContext(context.Background()); pingErr == nil {
		t.Fatal("database remained open after startup failure")
	}
}

func TestStorageFolderCommandUsesExplorerWithOnlyConfiguredPath(t *testing.T) {
	dataPath := `C:\Users\Ada\AppData\Local\RailKeeper\data`
	command := storageFolderCommand(context.Background(), dataPath)
	if len(command.Args) != 2 || command.Args[0] != "explorer.exe" || command.Args[1] != dataPath {
		t.Fatalf("storage folder command args = %v", command.Args)
	}
}
