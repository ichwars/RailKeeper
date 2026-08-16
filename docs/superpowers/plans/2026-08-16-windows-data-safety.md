# Windows Data Safety Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement issue #84 so Windows Standalone keeps persistent data outside the replaceable
program directory, migrates legacy data losslessly, and creates a validated SQLite safety copy
before pending migrations.

**Architecture:** A new `backend/internal/startup` package resolves runtime paths and performs the
legacy-data state machine. Focused SQLite snapshot and migration-planning helpers remain in
`backend/internal/infrastructure`. The command coordinates these units, while authenticated system
APIs expose storage transparency without accepting filesystem paths from clients.

**Tech Stack:** Go 1.26, `database/sql`, modernc SQLite, React 19, TypeScript, Vitest, PowerShell,
OpenAPI, VitePress.

## Global Constraints

- Windows Standalone defaults to `%LOCALAPPDATA%\RailKeeper\data`.
- A non-empty `RAILKEEPER_DATA_DIR` always wins and disables automatic legacy migration.
- Docker continues to use `/data`; non-Windows server defaults remain unchanged.
- The legacy source directory is never modified or deleted.
- Different old and new databases are never merged or selected automatically.
- No pending migration runs unless a complete SQLite snapshot passes `PRAGMA integrity_check`.
- Safety copies contain authentication data and remain private.
- The Windows ZIP contains no `data` directory or user-data artifacts.
- UI and documentation use “Windows Standalone (ohne Installation)” and the English equivalent.
- Preserve #83 master-data `origin` and `active` state through migration and schema upgrades.
- Do not add dependencies or rewrite historical SQL migrations.

---

## Planned file structure

- `backend/internal/startup/runtime_config.go`: pure runtime/path resolution.
- `backend/internal/startup/runtime_config_test.go`: Windows, configured, Docker/server path cases.
- `backend/internal/startup/legacy_data.go`: legacy decision table and atomic promotion.
- `backend/internal/startup/legacy_data_test.go`: copy, conflict, equivalence, interruption, and #83
  preservation tests.
- `backend/internal/startup/conflict_page.go`: loopback-only static conflict handler.
- `backend/internal/startup/conflict_page_test.go`: asserts safety copy and absence of app endpoints.
- `backend/internal/infrastructure/sqlite_snapshot.go`: consistent snapshot and integrity validation.
- `backend/internal/infrastructure/sqlite_snapshot_test.go`: WAL and corrupt-copy coverage.
- `backend/internal/infrastructure/migration_safety.go`: migration inspection and pre-migration gate.
- `backend/internal/infrastructure/migration_safety_test.go`: pending/current/new/failure coverage.
- `backend/cmd/railkeeper/startup.go`: startup orchestration extracted from `main.go`.
- `backend/cmd/railkeeper/startup_test.go`: typed startup outcomes and dependency wiring.
- `backend/internal/api/storage_info_handlers.go`: storage information, Explorer, acknowledgement.
- `frontend/src/features/settings/StorageLocationPanel.tsx`: storage-mode and migration-receipt UI.
- `tools/windows_package_validation.ps1`: reusable staged-directory and ZIP validation.
- `tools/test_windows_package_validation.ps1`: self-contained validation regression checks.
- `tools/build_windows_standalone.ps1`: renamed package builder without user data.
- `.github/workflows/windows-standalone.yml`: renamed Windows package workflow.

### Task 1: Resolve runtime and persistent data paths

**Files:**
- Create: `backend/internal/startup/runtime_config.go`
- Create: `backend/internal/startup/runtime_config_test.go`
- Modify: `backend/cmd/railkeeper/main.go`

**Interfaces:**
- Consumes: process arguments, `runtime.GOOS`, executable directory, current directory, and
  environment lookup.
- Produces:

```go
type StorageMode string

const (
    StorageModeWindowsStandalone StorageMode = "windows_standalone"
    StorageModeConfigured        StorageMode = "configured"
    StorageModeServer            StorageMode = "server"
)

type RuntimeInputs struct {
    GOOS             string
    Args             []string
    ExecutableDir    string
    WorkingDir       string
    LookupEnv        func(string) (string, bool)
    PathExists       func(string) bool
    JoinPath         func(...string) string
    AbsPath          func(string) (string, error)
}

type RuntimeConfig struct {
    Standalone             bool
    StorageMode            StorageMode
    DataDir                string
    LegacyDataDir          string
    MigrationsDir          string
    SeedsDir               string
    StaticDir              string
    AddrDefault            string
    OpenDataFolderSupported bool
}

func ResolveRuntimeConfig(inputs RuntimeInputs) (RuntimeConfig, error)
```

- [ ] **Step 1: Write failing table-driven path-resolution tests**

Cover these exact cases in `runtime_config_test.go`:

```go
tests := []struct {
    name       string
    goos       string
    args       []string
    env        map[string]string
    bundled    bool
    wantMode   startup.StorageMode
    wantData   string
    wantLegacy string
}{
    {"new Windows standalone", "windows", []string{"--standalone"},
        map[string]string{"LOCALAPPDATA": `C:\Users\Ada\AppData\Local`}, true,
        startup.StorageModeWindowsStandalone,
        `C:\Users\Ada\AppData\Local\RailKeeper\data`, `C:\RailKeeper\data`},
    {"legacy portable flag stays safe", "windows", []string{"--portable"},
        map[string]string{"LOCALAPPDATA": `C:\Users\Ada\AppData\Local`}, true,
        startup.StorageModeWindowsStandalone,
        `C:\Users\Ada\AppData\Local\RailKeeper\data`, `C:\RailKeeper\data`},
    {"explicit directory wins", "windows", []string{"--standalone"},
        map[string]string{"LOCALAPPDATA": `C:\Users\Ada\AppData\Local`,
            "RAILKEEPER_DATA_DIR": `D:\RailKeeperUSB\data`}, true,
        startup.StorageModeConfigured, `D:\RailKeeperUSB\data`, ""},
    {"server default unchanged", "linux", nil, map[string]string{}, false,
        startup.StorageModeServer, `data`, ""},
}
```

Also require errors for missing/blank `LOCALAPPDATA` in bundled Windows mode and prove there is no
fallback to `<exe>\data`.

- [ ] **Step 2: Run the focused test and verify RED**

Run: `cd backend; go test ./internal/startup -run TestResolveRuntimeConfig -count=1`

Expected: FAIL because the startup package and resolver do not exist.

- [ ] **Step 3: Implement the pure resolver**

Implement the declared types. Production inputs use `filepath.Join` and `filepath.Abs`; tests inject
Windows path joining so Windows behavior remains testable on Linux CI. Treat `--standalone`, `--portable`,
`RAILKEEPER_PORTABLE=true`, or the packaged `web/index.html` + `migrations` + `seeds` layout as
standalone runtime selection. Check `RAILKEEPER_DATA_DIR` before Windows default selection. Use
`filepath.Clean` and `filepath.Abs`; reject an empty Windows LocalAppData value. Keep the existing
directory defaults for migrations, seeds, static files, address, and development/server data.

- [ ] **Step 4: Run startup resolver and command tests**

Run:

```powershell
cd backend
go test ./internal/startup ./cmd/railkeeper -count=1
```

Expected: PASS. Existing command behavior remains green except that standalone data defaults to
LocalAppData.

- [ ] **Step 5: Commit runtime resolution**

```powershell
git add backend/internal/startup/runtime_config.go `
  backend/internal/startup/runtime_config_test.go backend/cmd/railkeeper/main.go
git commit -m "feat: resolve safe Windows data paths"
```

### Task 2: Create and validate consistent SQLite snapshots

**Files:**
- Create: `backend/internal/infrastructure/sqlite_snapshot.go`
- Create: `backend/internal/infrastructure/sqlite_snapshot_test.go`
- Modify: `backend/internal/infrastructure/sqlite.go`

**Interfaces:**
- Consumes: an open source `*sql.DB` or an existing DB path and a non-existent target path.
- Produces:

```go
func CreateSQLiteSnapshot(ctx context.Context, source *sql.DB, targetPath string) error
func CreateSQLiteSnapshotFromPath(ctx context.Context, sourcePath, targetPath string) error
func ValidateSQLiteSnapshot(ctx context.Context, snapshotPath string) error
func SQLiteSnapshotsEquivalent(ctx context.Context, leftPath, rightPath, tempDir string) (bool, error)
```

- [ ] **Step 1: Write failing WAL snapshot tests**

Create a source DB through `OpenSQLite`, create a table, insert one row, force a WAL-backed write,
then call `CreateSQLiteSnapshot`. Open the result separately and assert the row exists and
`PRAGMA integrity_check` returns `ok`. Add tests that reject an existing non-empty target and a
corrupt DB. Add an equivalence test for two logically identical DBs and a difference test after one
extra insert.

- [ ] **Step 2: Run the snapshot tests and verify RED**

Run: `cd backend; go test ./internal/infrastructure -run 'Snapshot|Equivalent' -count=1`

Expected: FAIL with undefined snapshot functions.

- [ ] **Step 3: Implement `VACUUM INTO` snapshots and validation**

Use `VACUUM INTO ?` on the source connection. Require that the target does not exist, create only
its parent with mode `0700`, and set `PRAGMA synchronous=FULL` on the source connection before the
snapshot. Close and reopen the target independently, read-only, then require exactly one
`PRAGMA integrity_check` row equal to `ok`. Delete an invalid target on failure.

For path sources, open the existing DB with a read-only SQLite URI, set one open connection, and
close it after snapshotting. For equivalence, snapshot each DB into separate random files below
`tempDir`, validate both, calculate SHA-256 over the normalized snapshot bytes, and delete both temp
files. Any inability to prove equality returns `false` with an error; callers must treat that as a
conflict.

This follows SQLite's documented guarantee that `VACUUM INTO` produces a consistent snapshot of the
original database and includes committed WAL state.

- [ ] **Step 4: Run focused and complete infrastructure tests**

Run:

```powershell
cd backend
go test ./internal/infrastructure -run 'Snapshot|Equivalent' -count=1
go test ./internal/infrastructure -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the snapshot primitive**

```powershell
git add backend/internal/infrastructure/sqlite_snapshot.go `
  backend/internal/infrastructure/sqlite_snapshot_test.go backend/internal/infrastructure/sqlite.go
git commit -m "feat: create validated SQLite snapshots"
```

### Task 3: Gate pending migrations behind a safety copy

**Files:**
- Create: `backend/internal/infrastructure/migration_safety.go`
- Create: `backend/internal/infrastructure/migration_safety_test.go`
- Modify: `backend/internal/infrastructure/sqlite.go`

**Interfaces:**
- Consumes: current DB, whether the DB existed before open, active data directory, migration files,
  clock, and snapshot function.
- Produces:

```go
type MigrationPlan struct {
    Applied []string
    Pending []string
    From    string
    To      string
}

type MigrationSafetyResult struct {
    Plan       MigrationPlan
    BackupPath string
}

type MigrationSafetyOptions struct {
    DatabaseExisted bool
    Now             func() time.Time
    Snapshot        func(context.Context, *sql.DB, string) error
}

func InspectMigrationPlan(ctx context.Context, db *sql.DB, migrationsDir string) (MigrationPlan, error)
func MigrateSafely(ctx context.Context, db *sql.DB, dataDir, migrationsDir string,
    options MigrationSafetyOptions) (MigrationSafetyResult, error)
```

- [ ] **Step 1: Write failing migration-gate tests**

Cover: new DB with pending migrations and no copy; existing current DB and no copy; existing DB with
one pending migration and a validated copy; snapshot callback failure with unchanged
`schema_migrations`; corrupt snapshot validation with unchanged schema; deterministic filename
containing `from`, `to`, and UTC timestamp.

Use a test migration that adds a marker table and assert the table is absent after forced backup
failure.

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `cd backend; go test ./internal/infrastructure -run 'MigrationPlan|MigrateSafely' -count=1`

Expected: FAIL with undefined migration-safety types.

- [ ] **Step 3: Implement inspection and gated migration**

Read and sort `.sql` files using the same rules as `Migrate`. Detect whether
`schema_migrations` exists through `sqlite_schema` without creating it. Treat every file as pending
when the table is absent. When `DatabaseExisted` and `Pending` are both true, create
`<dataDir>/safety-backups/railkeeper-pre-migration-<from>-to-<to>-<UTC>.db`, validate it, reopen it,
and compare its applied migration list to the source plan. Only then call the existing transactional
migration application loop. Return the copy path for logging and UI diagnostics.

Keep `Migrate` as a compatibility wrapper for existing tests and non-production helpers; production
startup must call `MigrateSafely`.

- [ ] **Step 4: Run migration and backend tests**

Run:

```powershell
cd backend
go test ./internal/infrastructure -count=1
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit the migration gate**

```powershell
git add backend/internal/infrastructure/migration_safety.go `
  backend/internal/infrastructure/migration_safety_test.go backend/internal/infrastructure/sqlite.go
git commit -m "feat: back up databases before migrations"
```

### Task 4: Migrate legacy standalone data atomically

**Files:**
- Create: `backend/internal/startup/legacy_data.go`
- Create: `backend/internal/startup/legacy_data_test.go`

**Interfaces:**
- Consumes: safe/legacy directories, application version, deterministic test hooks, and Task 2
  snapshot functions.
- Produces:

```go
type LegacyMigrationStatus string

const (
    LegacyReady    LegacyMigrationStatus = "ready"
    LegacyMigrated LegacyMigrationStatus = "migrated"
    LegacyConflict LegacyMigrationStatus = "conflict"
)

type MigrationReceipt struct {
    SourcePath     string `json:"sourcePath"`
    TargetPath     string `json:"targetPath"`
    MigratedAt     string `json:"migratedAt"`
    Version        string `json:"version"`
    FilesVerified int    `json:"filesVerified"`
    Acknowledged  bool   `json:"acknowledged"`
}

type LegacyConflictInfo struct {
    SafePath   string
    LegacyPath string
    Reason     string
}

type LegacyMigrationOptions struct {
    Version      string
    Now          func() time.Time
    RandomSuffix func() string
}

type LegacyMigrationResult struct {
    Status   LegacyMigrationStatus
    DataDir  string
    Receipt  *MigrationReceipt
    Conflict *LegacyConflictInfo
}

func ResolveLegacyData(ctx context.Context, safeDataDir, legacyDataDir string,
    options LegacyMigrationOptions) (LegacyMigrationResult, error)
```

- [ ] **Step 1: Write the legacy state-machine tests**

Test every decision-table row. For migration success create nested uploads, a backup file, an
unknown future file, and a migrated RailKeeper DB. Record source hashes before the call and prove
they are unchanged afterward. Assert the target receipt and complete manifest.

Create an integration fixture with migration `0058`, set bundled manufacturer `tillig` to
`active=0`, migrate it, and assert `origin='bundled'` and `active=0` in the target.

Add failure cases for a symlink/reparse point, non-empty target without DB, injected copy failure,
hash mismatch, corrupt source DB, and atomic rename failure. Assert no active target exists and only
the owned staging directory is removed.

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `cd backend; go test ./internal/startup -run 'Legacy|MasterData' -count=1`

Expected: FAIL with undefined migration types.

- [ ] **Step 3: Implement safe copy, manifest, receipt, and promotion**

Use a staging sibling named `.railkeeper-migration-<random>` under the safe directory's parent.
Reject directory entries whose `os.FileMode` includes `ModeSymlink`; on Windows also reject reparse
points detected by the platform helper. Snapshot `railkeeper.db` through Task 2 and skip source
`railkeeper.db`, `railkeeper.db-wal`, and `railkeeper.db-shm` during ordinary file copying.

Copy regular files with `O_CREATE|O_EXCL`, preserve relative paths only, use mode `0600`, hash while
copying, and compare path/type/size/SHA-256 manifests. Build the source manifest before and after
the copy and fail if it changed, preventing activation when another process mutates the legacy
directory concurrently. Validate the DB, write
`.railkeeper-legacy-migration.json`, close every handle, remove an empty target directory if present,
then `os.Rename(staging, safeDataDir)`. Never remove `legacyDataDir`.

Whenever the safe directory is selected, read and validate an existing migration receipt so its
administrator notice survives later restarts until acknowledgement.

When both DBs exist, use `SQLiteSnapshotsEquivalent`. Equality selects the safe directory without
copying. Inequality or indeterminate comparison returns `LegacyConflict` without filesystem writes
outside owned temporary comparison files.

- [ ] **Step 4: Run startup and complete backend tests**

Run:

```powershell
cd backend
go test ./internal/startup -count=1
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit legacy migration**

```powershell
git add backend/internal/startup/legacy_data.go backend/internal/startup/legacy_data_test.go
git commit -m "feat: migrate legacy Windows data safely"
```

### Task 5: Orchestrate safe startup and serve the conflict page

**Files:**
- Create: `backend/cmd/railkeeper/startup.go`
- Create: `backend/cmd/railkeeper/startup_test.go`
- Create: `backend/internal/startup/conflict_page.go`
- Create: `backend/internal/startup/conflict_page_test.go`
- Modify: `backend/cmd/railkeeper/main.go`

**Interfaces:**
- Consumes: `RuntimeConfig`, `LegacyMigrationResult`, `MigrateSafely`, listener, logger.
- Produces:

```go
func ConflictHandler(info startup.LegacyConflictInfo) http.Handler

type StartupState struct {
    Runtime         startup.RuntimeConfig
    Receipt         *startup.MigrationReceipt
    SafetyBackupPath string
}
```

- [ ] **Step 1: Write failing conflict and orchestration tests**

Assert `GET /` returns a German/English safety explanation containing escaped absolute paths,
`Cache-Control: no-store`, and status 409. Assert `/api/v1/setup/status`, `/api/v1/version`, and an
unknown route do not expose normal JSON APIs. Assert the page has no form, delete, merge, overwrite,
or continue action.

For orchestration, inject a conflict and prove `OpenSQLite`, `MigrateSafely`, seeds, and router
construction are never called. Inject successful migration and prove the resolved safe path is used
for DB, blobs, backups, and `api.Config.DataDir`. Verify `DatabaseExisted` is captured before open.

- [ ] **Step 2: Run the focused tests and verify RED**

Run:

```powershell
cd backend
go test ./internal/startup ./cmd/railkeeper -run 'Conflict|Startup' -count=1
```

Expected: FAIL because the handler and orchestration seam do not exist.

- [ ] **Step 3: Extract orchestration and implement the minimal safety page**

Move startup path selection, legacy resolution, database existence check, `OpenSQLite`,
`MigrateSafely`, seeds, and service wiring from `main.go` into focused functions in `startup.go`.
Keep healthcheck handling and process exit in `main.go`.

Build the conflict page with `html/template`, no external assets or scripts, escaped paths, inline
accessible CSS, German first and English second, and the exact safe manual sequence from the spec.
Serve it only on the already loopback-bound standalone listener. Log conflict metadata without DB
contents. Apply the existing automatic browser-open behavior.

- [ ] **Step 4: Run command, startup, and full backend tests**

Run:

```powershell
cd backend
gofmt -w cmd/railkeeper/main.go cmd/railkeeper/startup.go cmd/railkeeper/startup_test.go `
  internal/startup/conflict_page.go internal/startup/conflict_page_test.go
go test ./cmd/railkeeper ./internal/startup -count=1
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit startup safety**

```powershell
git add backend/cmd/railkeeper/main.go backend/cmd/railkeeper/startup.go `
  backend/cmd/railkeeper/startup_test.go backend/internal/startup/conflict_page.go `
  backend/internal/startup/conflict_page_test.go
git commit -m "feat: gate startup on Windows data safety"
```

### Task 6: Expose storage location and migration receipt safely

**Files:**
- Create: `backend/internal/api/storage_info_handlers.go`
- Create: `backend/internal/api/storage_info_handlers_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/cmd/railkeeper/startup.go`
- Modify: `openapi/railkeeper.yaml`
- Modify: `frontend/src/shared/api.ts`
- Create: `frontend/src/features/settings/StorageLocationPanel.tsx`
- Create: `frontend/src/features/settings/StorageLocationPanel.test.tsx`
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/settings.css`

**Interfaces:**
- Consumes: resolved storage mode/path and optional receipt from Tasks 1, 4, and 5.
- Produces API:

```text
GET  /api/v1/system/storage/info
POST /api/v1/system/storage/open-folder
POST /api/v1/system/storage/migration-receipt/acknowledge
```

```ts
export type StorageLocationInfo = {
  dataPath: string;
  mode: "windows_standalone" | "configured" | "server";
  openFolderAvailable: boolean;
  migrationReceipt?: {
    sourcePath: string;
    targetPath: string;
    migratedAt: string;
    version: string;
    filesVerified: number;
    acknowledged: boolean;
  };
};
```

- [ ] **Step 1: Write failing backend API tests**

Assert all three routes require Admin and writes require CSRF. `GET` returns the exact resolved path
and mode. The open action accepts an empty body, passes only configured `DataDir` to an injected
`func(context.Context, string) error`, and returns 409 `storage_folder_open_unavailable` when the
capability is false. The acknowledgement updates only `Acknowledged` in the receipt using atomic
temp-write + rename and never writes to `SourcePath`.

- [ ] **Step 2: Run API tests and verify RED**

Run: `cd backend; go test ./internal/api -run StorageInfo -count=1`

Expected: FAIL because the routes and handler types do not exist.

- [ ] **Step 3: Implement API, OpenAPI, and Explorer capability**

Add storage info and opener dependencies to `api.Config`. Implement the Windows opener in command
wiring with `exec.CommandContext(ctx, "explorer.exe", dataDir).Start()`. Do not accept any path,
query, or filename from HTTP. Return capability false for every non-Windows-Standalone runtime.
Define all request/response and problem codes in OpenAPI.

- [ ] **Step 4: Write failing frontend component tests**

Mock `api.storageLocationInfo`. Assert the card displays absolute path and translated mode, renders
“Datenordner öffnen” only when available, calls `api.openStorageFolder`, shows the retained source
and target after migration, and calls `api.acknowledgeStorageMigration` only after “Verstanden”.
Cover long German paths and unavailable/error states.

- [ ] **Step 5: Implement the focused storage-location component**

Add the three typed API methods, render `StorageLocationPanel` inside the existing storage card,
reuse existing panel/button tokens, and keep the path selectable and horizontally wrappable. Add
complete German and English copy. Do not add another top-level settings tab.

- [ ] **Step 6: Run API, frontend tests, and build**

Run:

```powershell
cd backend
go test ./internal/api ./cmd/railkeeper -count=1
cd ..\frontend
npm.cmd test -- --run StorageLocationPanel SettingsView
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 7: Commit storage transparency**

```powershell
git add backend/internal/api/storage_info_handlers.go `
  backend/internal/api/storage_info_handlers_test.go backend/internal/api/router.go `
  backend/internal/api/routes.go backend/cmd/railkeeper/startup.go openapi/railkeeper.yaml `
  frontend/src/shared/api.ts frontend/src/features/settings/StorageLocationPanel.tsx `
  frontend/src/features/settings/StorageLocationPanel.test.tsx `
  frontend/src/features/settings/SettingsView.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/styles/settings.css
git commit -m "feat: show persistent storage location"
```

### Task 7: Build a data-free Windows Standalone package

**Files:**
- Create: `tools/windows_package_validation.ps1`
- Create: `tools/test_windows_package_validation.ps1`
- Create: `tools/build_windows_standalone.ps1`
- Delete: `tools/build_windows_portable.ps1`
- Create: `.github/workflows/windows-standalone.yml`
- Delete: `.github/workflows/windows-portable.yml`
- Modify: `deploy/windows/start-railkeeper.bat`
- Modify: `deploy/windows/README-Windows.txt`

**Interfaces:**
- Consumes: staged package directory and produced ZIP.
- Produces PowerShell functions:

```powershell
function Assert-RailKeeperPackageDirectory([string]$PackageDir)
function Assert-RailKeeperPackageArchive([string]$ZipPath)
```

- [ ] **Step 1: Write failing package-validation regression script**

Create temporary good and bad package trees. The good tree contains only EXE, web, migrations,
seeds, BAT, and README. Bad cases contain `data`, `railkeeper.db`, `railkeeper.db-wal`, `uploads`,
and `backups`. Create ZIPs for good and bad cases. Require good checks to return normally and each
bad check to throw with the offending relative path. Clean the exact temporary root in `finally`.

- [ ] **Step 2: Run validation regression and verify RED**

Run: `pwsh -NoProfile -File tools/test_windows_package_validation.ps1`

Expected: FAIL because the validation module does not exist.

- [ ] **Step 3: Implement validation and rename the builder/workflow**

Reject directory segments named `data`, `uploads`, `attachments`, `thumbnails`, or `backups`, and
extensions/names `.db`, `.db-wal`, `.db-shm`. Validate both the staging tree and entries read through
`System.IO.Compression.ZipFile::OpenRead`. Do not extract the archive for validation.

Rename package directory to `RailKeeper-Windows-Standalone`, retain asset filename
`RailKeeper-windows-x64-v<version>.zip`, remove creation of `data`, invoke `--standalone` in BAT,
and update workflow names/artifacts. Run directory validation before compression and archive
validation afterward.

- [ ] **Step 4: Run validation and build a real package**

Run:

```powershell
pwsh -NoProfile -File tools/test_windows_package_validation.ps1
pwsh -NoProfile -File tools/build_windows_standalone.ps1
```

Expected: PASS. The ZIP exists and validation reports no forbidden entries. Do not add `dist`,
`.cache`, or package output to Git.

- [ ] **Step 5: Commit package safety**

```powershell
git add tools/windows_package_validation.ps1 tools/test_windows_package_validation.ps1 `
  tools/build_windows_standalone.ps1 tools/build_windows_portable.ps1 `
  .github/workflows/windows-standalone.yml .github/workflows/windows-portable.yml `
  deploy/windows/start-railkeeper.bat deploy/windows/README-Windows.txt
git commit -m "build: create data-free Windows standalone package"
```

### Task 8: Document, verify, and close issue #84 scope

**Files:**
- Modify: `README.md`
- Modify: `deploy/README.md`
- Modify: `docs/architecture.md`
- Modify: `docs/security.md`
- Modify: `docs/coverage.json`
- Modify: `docs/site/administration/index.md`
- Modify: `docs/site/de/administration/index.md`
- Modify: `deploy/windows/README-Windows.txt`

**Interfaces:**
- Consumes: final behavior from Tasks 1-7.
- Produces: complete bilingual operational guidance and verification evidence.

- [ ] **Step 1: Update all stable documentation surfaces**

Document exact default path, explicit portable-data configuration and warning, automatic legacy
migration, retained source, two-DB conflict, safety-copy location and privacy, Explorer action,
Docker `/data`, and the limits of automatic protection. Replace default-package “Portable” wording
with “Windows Standalone (no installation required)” without changing historical release notes.

- [ ] **Step 2: Run terminology and documentation checks**

Run:

```powershell
rg -n "RailKeeper Portable|windows-portable|RailKeeper-Portable" README.md deploy docs/site `
  .github tools backend frontend
cd docs
npm.cmd run check
```

Expected: `rg` finds only explicit legacy/deprecated compatibility explanations. Documentation tests
and VitePress build pass.

- [ ] **Step 3: Run the complete automated verification matrix**

Run:

```powershell
cd backend
go test ./... -count=1
cd ..\frontend
npm.cmd test -- --run
npm.cmd run build
cd ..
pwsh -NoProfile -File tools/test_windows_package_validation.ps1
pwsh -NoProfile -File tools/build_windows_standalone.ps1
git diff --check origin/main...HEAD
git status --short
```

Expected: all Go packages pass; all frontend tests pass; frontend build succeeds; package validation
and real package build succeed; diff check is empty; status contains only intentional tracked source
changes and ignored build output.

- [ ] **Step 4: Perform isolated Windows acceptance tests**

Use temporary application, LocalAppData, and legacy roots. Never point at the user's real data.
Verify new install, successful legacy migration, retained source, interrupted migration cleanup,
different-DB safety page, equivalent-DB selection, forced safety-backup failure, pending migration,
inactive bundled manufacturer preservation, storage settings, Explorer capability, dark/light, and
mobile/desktop layouts. After populating the safe default with vehicles, a modified manufacturer,
users, uploads, and backups, delete and re-extract only the temporary program directory and prove
the replacement executable reads the unchanged safe data. Record exact temp paths and delete only
those verified test roots afterward.

- [ ] **Step 5: Commit documentation and final refinements**

```powershell
git add README.md deploy docs backend frontend openapi tools .github
git commit -m "docs: explain safe Windows data storage"
```

- [ ] **Step 6: Re-run clean-tree verification**

Run: `git status -sb`

Expected: branch is clean and ahead of its base only by the intentional #84 commits. Do not publish
or merge until the user reviews the local Windows acceptance result.
