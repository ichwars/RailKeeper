# Safe Windows data storage and update download design

## Scope

This design implements GitHub issues #84 and #82 in that order. Issue #84 separates replaceable
Windows application files from persistent user data, migrates existing standalone installations
without modifying the source, and creates a validated SQLite safety copy before schema migrations.
Issue #82 then adds a user-initiated download of the matching Windows ZIP.

The implementation must preserve the local-first deployment model, SQLite, the existing Docker
volume, explicit `RAILKEEPER_DATA_DIR` configurations, authentication data, uploads, backups, and
the active/inactive master-data state introduced by #83.

## Goals

- Store new Windows Standalone data under `%LOCALAPPDATA%\RailKeeper\data` by default.
- Keep the downloaded application directory fully replaceable.
- Migrate a legacy `data` directory beside `RailKeeper.exe` without modifying or deleting it.
- Stop safely when old and new locations contain different databases.
- Create and validate a complete SQLite safety copy before pending migrations of an existing DB.
- Show administrators the active data path, storage mode, and legacy-migration result.
- Offer the exact Windows ZIP through an explicit button only in Windows Standalone mode.
- Keep Docker and explicitly configured data directories unchanged.

## Non-goals

- RailKeeper does not install, extract, replace, or restart itself.
- RailKeeper does not merge two databases.
- RailKeeper does not delete the legacy directory or automatic migration safety copies.
- RailKeeper does not add cloud storage, a background updater, or a second launcher executable.
- RailKeeper does not promise protection from manual deletion of the persistent data directory,
  storage failure, malware, or missing external backups.

## Terminology

- **Windows Standalone (no installation required):** the downloadable Windows package. Its program
  files are replaceable and its default data directory is outside the program directory.
- **Explicit portable-data mode:** a user deliberately sets `RAILKEEPER_DATA_DIR`, possibly to a
  directory beside the executable or on removable media. Deleting that location deletes the data.
- **Legacy standalone installation:** the existing package layout with
  `<program directory>\data\railkeeper.db`.
- **Safe data directory:** `%LOCALAPPDATA%\RailKeeper\data`.
- **Application directory:** the directory containing `RailKeeper.exe`, `web`, `migrations`, and
  `seeds`.

The product UI and documentation stop calling the default Windows package "Portable". The old
`--portable` argument remains a deprecated compatibility alias for recognizing the bundled runtime
layout; it no longer selects a data directory beside the executable.

## Architecture

The startup safety chain uses focused units with explicit inputs and results:

1. `runtimeconfig` resolves the packaged runtime mode, executable directory, configured paths, and
   Windows storage mode without opening SQLite.
2. `LegacyDataMigrator` inspects the safe and legacy locations and either selects the safe path,
   migrates the legacy directory, or returns a typed conflict result.
3. `MigrationSafetyBackup` checks whether an existing database has pending migrations, creates a
   consistent safety copy, validates it, and only then authorizes migrations.
4. `main.go` coordinates the units, starts the normal application on success, or serves the local
   safety page for a legacy conflict.
5. The system API exposes storage information and the permitted Explorer action to administrators.
6. The version API exposes a Windows download only when runtime mode, target version, asset name,
   and HTTPS origin all match.

The new components live under `backend/internal/infrastructure` or a focused startup package, not
inside the API handlers. `main.go` remains orchestration code.

## Runtime and path resolution

Resolution follows this precedence:

1. A non-empty `RAILKEEPER_DATA_DIR` always wins on every operating system. No legacy migration or
   Windows default-path change runs. The mode is reported as explicitly configured.
2. A bundled Windows runtime uses `%LOCALAPPDATA%\RailKeeper\data`.
3. Docker and other server deployments retain the existing `./data` fallback. Docker continues to
   set `/data` explicitly.

A bundled runtime is detected through the existing `web`, `migrations`, and `seeds` layout or the
new `--standalone` argument. `--portable` and `RAILKEEPER_PORTABLE=true` remain compatibility aliases
for bundled runtime behavior only. They never silently opt into data beside the executable.

If `%LOCALAPPDATA%` is unavailable or resolves to an unusable path, Windows Standalone stops with a
clear startup error. It must not fall back to the application directory.

## Legacy migration decision table

Before creating a new database, RailKeeper evaluates
`%LOCALAPPDATA%\RailKeeper\data\railkeeper.db` and
`<application directory>\data\railkeeper.db`:

| Safe DB | Legacy DB | Result |
| --- | --- | --- |
| absent | absent | Use the safe directory and create a new installation there. |
| present | absent | Use the safe directory. |
| absent | present | Run the lossless legacy migration. |
| present | present, equivalent | Use the safe directory and retain the legacy copy. |
| present | present, different | Do not open either as the active app DB; serve the safety page. |

An existing non-empty safe directory without a safe database is never merged with legacy content.
It is treated as a conflict because its files may be user data. An empty safe directory may be
removed immediately before atomic promotion.

## Lossless legacy migration

The migrator creates a randomly named sibling staging directory below
`%LOCALAPPDATA%\RailKeeper`. It never stages inside the legacy application directory.

1. Open the legacy SQLite database read-only and create a consistent database snapshot in the
   staging directory. The snapshot mechanism must account for committed WAL content.
2. Copy every non-database file and directory from the legacy data root without following links
   outside that root. This includes uploads, attachments, thumbnails, exports, backups, and unknown
   future files.
3. Compare the source and staged non-database manifests using relative path, type, size, and SHA-256.
4. Open the staged database independently and require `PRAGMA integrity_check` to return `ok`.
5. Verify critical application state, including users, vehicles, uploads, backups, and the #83
   master-data origin/active fields when their tables exist.
6. Write a migration receipt containing source path, target path, timestamp, application version,
   and verification summary.
7. Flush and close all staging resources.
8. Atomically rename the staging directory to the safe data directory.

Any failure removes only the incomplete staging directory. The source remains byte-for-byte
untouched. A destination is never promoted before all checks pass. Unexpected interruption may
leave an identifiable staging directory, which the next run validates as incomplete and removes
before starting a new attempt; it is never selected as active data.

Database equivalence for the two-database case is determined from independent consistent SQLite
snapshots and SHA-256, not from timestamps or raw live DB/WAL file copies. When equivalence cannot
be proven, RailKeeper treats the databases as different.

## Conflict safety page

For Windows Standalone, the listener is already local-only. A typed legacy conflict starts a minimal
HTTP handler instead of the normal API and frontend. It displays:

- that RailKeeper found two different databases and changed neither one;
- the exact legacy and safe paths;
- that automatic merging is intentionally unavailable;
- safe manual choices: close RailKeeper, back up both directories, then rename or move the unwanted
  safe directory before retrying, or explicitly configure the intended directory;
- a reminder not to delete either copy before verifying the chosen installation.

The normal frontend, setup endpoints, authentication, migrations, seeds, and file services are not
started. The page contains no delete, overwrite, merge, or choose-and-continue action. It is served
only on the loopback listener and automatically opened under the same rules as Windows Standalone.

## Safety copy before database migrations

The production startup path determines whether the database existed before `OpenSQLite`. A new empty
database does not need a pre-migration copy. For an existing database:

1. Read the available migration files and applied `schema_migrations` versions without modifying the
   schema.
2. If no migration is pending, continue without creating a copy.
3. If migrations are pending, create
   `safety-backups\railkeeper-pre-migration-<from>-to-<to>-<UTC>.db` through a SQLite-consistent
   snapshot operation.
4. Open the copy independently and require `PRAGMA integrity_check` to return `ok`.
5. Confirm that the copied migration state matches the source migration state.
6. Only after validation may the existing transactional migration runner apply migrations.

The copy contains the complete SQLite database, including users, password hashes, roles, sessions,
settings, audit data, and master-data lifecycle state. It is therefore private operational data and
must never be served or included in public diagnostics. Automatic deletion or retention pruning is
out of scope for this first implementation.

If snapshot creation or validation fails, no pending migration executes. Startup ends with a clear
error that includes the data path and failed safety-copy path but never credentials or DB contents.

## Storage transparency and administration API

The authenticated Admin settings area shows:

- the actual absolute data directory;
- storage mode: Windows Standalone, explicitly configured, Docker/server, or development;
- whether Explorer opening is available;
- the last legacy-migration receipt, including the retained source path.

An Admin-only, CSRF-protected POST action opens the active data directory in Windows Explorer. The
backend accepts no path from the request. It uses only the resolved application data directory and
is available only for a local Windows Standalone runtime. Docker and remote server installations
return an unavailable capability and do not execute operating-system commands.

After a successful legacy migration, administrators see a prominent informational notice until an
administrator acknowledges it. Acknowledgement updates only the migration receipt in the safe data
directory; it never touches the legacy source.

The OpenAPI contract, backend API types, frontend client, German translation, and English translation
change together.

## Windows package

The package build and workflow use "Windows Standalone" terminology. The archive retains the stable
asset shape `RailKeeper-windows-x64-v<version>.zip` so update selection can be deterministic.

The package contains only replaceable runtime content:

- `RailKeeper.exe`
- `web`
- `migrations`
- `seeds`
- `start-railkeeper.bat`
- `README-Windows.txt`

It contains no `data` directory and no `.db`, `.db-wal`, `.db-shm`, upload, attachment, thumbnail,
or backup content. The build script validates the staged directory before compression and inspects
the finished archive before reporting success.

`start-railkeeper.bat` invokes `RailKeeper.exe --standalone`. The bundled Windows instructions
describe the safe default location and the optional explicit portable-data mode separately.

## Update download

The release lookup remains read-only. The backend selects a downloadable package only when all of
these conditions hold:

- the running mode is Windows Standalone;
- a newer version is available;
- the asset name exactly matches `RailKeeper-windows-x64-v<target version>.zip` after safe version
  normalization;
- the asset URL uses HTTPS;
- the host and path identify the configured RailKeeper GitHub repository release download.

The backend no longer exposes the first arbitrary release asset as a Windows download. Custom update
sources may provide version and release-page information, but they do not receive a Windows package
capability unless they meet the same trusted GitHub rules.

When a package is available, settings show `Version X herunterladen` / `Download version X`. The
user click navigates directly to the GitHub release asset so normal browser download behavior and
GitHub content disposition apply. No confirmation dialog is added. RailKeeper does not claim the
download is installed or complete.

The release-page link remains visible as the fallback. Docker and other server modes show their
existing release information and documented Docker update instructions, never the Windows ZIP
button. Offline, invalid release, invalid asset, and missing asset states remain non-destructive and
produce understandable messages.

## Security constraints

- All migration and backup paths are derived from trusted runtime configuration and confined to
  their expected roots.
- Directory-copy code does not follow links outside the legacy root.
- Staging names are random and created with restrictive permissions.
- Existing targets are never overwritten through rename.
- Conflict and failure paths do not initialize a new DB or expose normal app routes.
- The Explorer action is Admin-only, CSRF-protected, capability-gated, and pathless.
- Download URLs are returned only after exact asset-name and GitHub HTTPS validation.
- Safety copies and migration receipts are private and excluded from static serving.

## Verification

### Unit and integration tests

- Windows path resolution with injected OS, environment, executable path, and LocalAppData path.
- Explicit `RAILKEEPER_DATA_DIR` precedence on Windows, Linux, and Docker-like configurations.
- New Windows installation in the safe directory.
- Complete legacy migration with nested unknown files and no source mutation.
- Interrupted copy, hash mismatch, invalid SQLite copy, non-empty destination, and atomic-promotion
  failure.
- Equivalent and different two-database cases.
- Conflict handler exposes only the safety page and not setup or normal API routes.
- Pending versus current migrations and new versus existing databases.
- WAL-backed source changes included in the pre-migration safety copy.
- Forced snapshot or validation failure prevents every pending migration.
- Users, roles, sessions, vehicles, uploads, backups, and a deactivated bundled manufacturer survive
  legacy migration and a subsequent schema migration.
- Storage information and Explorer API authorization/capability behavior.
- Exact release asset selection, untrusted URL rejection, Docker suppression, release fallback, and
  offline behavior.
- Frontend download, storage path, migration notice, acknowledgement, German, and English states.

### Package and browser verification

- Build a real Windows archive and inspect its entries for forbidden data content.
- Start the package with isolated LocalAppData and legacy directories.
- Exercise new install, successful legacy migration, conflict page, retained source, settings path,
  Explorer capability, and download-button states.
- Check desktop/mobile layouts, dark/light themes, long German text, loading, error, and unavailable
  states.

### Baseline checks

- `go test ./... -count=1`
- `npm.cmd test -- --run`
- `npm.cmd run build`
- documentation validation and VitePress build
- Windows package build and archive-content validation
- OpenAPI/backend/frontend contract alignment

## Documentation and release communication

Update the bundled Windows README, root README, deployment configuration, administrator guide,
update guide, security documentation, release notes, and English/German user documentation. All
surfaces use the same terms and explain:

- the program directory may be replaced while safe default data remains;
- explicit portable-data mode can lose data when its directory is deleted;
- the legacy source remains after automatic migration;
- automatic pre-migration copies do not replace external backups;
- Docker continues to update with `docker compose pull` and `docker compose up -d`;
- the ZIP button downloads only and performs no installation.

## Delivery order

1. Runtime/path resolution and typed startup outcomes.
2. Legacy migration, validation, conflict page, and #83 preservation integration test.
3. Pre-migration SQLite safety copy and failure gates.
4. Storage transparency, Explorer action, and migration acknowledgement.
5. Windows Standalone packaging and forbidden-content checks.
6. Trusted release-asset selection and download UI from #82.
7. Full documentation, package QA, and end-to-end verification.

Issue #84 must pass all relevant acceptance checks before #82 is described or released as a safe
Windows update path.
