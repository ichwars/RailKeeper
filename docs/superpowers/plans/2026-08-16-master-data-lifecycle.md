# Master Data Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make bundled master data update-safe and deactivatable while allowing permanent deletion
only for unused user-created entries.

**Architecture:** Add a server-owned `origin` to each master-data row and reconcile it from a
generated catalog of shipped `(type, key)` pairs during startup. Keep usage detection, lifecycle
commands, import reconciliation, and restore reconciliation in the Go application layer. Expose
read-only origin and management capabilities through the API, then reuse one lifecycle action
component and one historical-option helper across settings, vehicles, accessories, and function
symbols.

**Tech Stack:** Go 1.26, SQLite, React 19, TypeScript 7, Vite 8, Vitest 4, OpenAPI 3, Python 3
standard library for the deterministic bundled-key manifest generator.

## Global Constraints

- #84 and #82 remain outside this branch and PR.
- Migrations are forward-only. Do not rewrite migrations `0001` through `0057`.
- Migration and startup reconciliation must not delete or overwrite existing master-data values.
- Bundled identity is based only on immutable `(type, key)` pairs, never editable labels or IDs.
- `origin` is read-only in normal create and update APIs.
- Unknown reference mappings fail closed and prevent permanent deletion.
- Existing inactive or unmatched persisted values remain readable in vehicle and accessory editors.
- New records can select only active master data.
- Master-data import and backup restore stay transactional and backwards compatible.
- Preserve German and English i18n, dark/light themes, compact desktop layout, and mobile layout.
- Do not commit `frontend/dist`, `frontend/node_modules`, `.cache`, `data`, credentials, or backups.

---

## File Map

### Database and bundled catalog

- Create `backend/migrations/0058_master_data_origin.sql`: add the constrained `origin` column.
- Create `tools/build_bundled_master_data_manifest.py`: build the authoritative shipped-key catalog
  from a clean migrated database plus `master_data.json`.
- Create `backend/seeds/bundled_master_data_manifest.json`: generated sorted `(type, key)` catalog.
- Modify `backend/internal/infrastructure/master_data_seed.go`: load the manifest, insert seed rows
  as bundled, and reconcile matching existing rows without changing user fields or active state.
- Modify `backend/internal/infrastructure/master_data_seed_test.go`: seed/reseed regression coverage.
- Create `backend/internal/infrastructure/master_data_origin_migration_test.go`: migration safety.

### Application lifecycle and persistence

- Modify `backend/internal/application/master_data.go`: origin-aware scans, creates, exports, and
  import entry points.
- Create `backend/internal/application/master_data_lifecycle.go`: origins, capabilities, active-state
  command, reference adapters, transactional permanent deletion, and relation cleanup.
- Create `backend/internal/application/master_data_lifecycle_test.go`: lifecycle and usage tests.
- Modify `backend/internal/application/master_data_custom_fields.go`: reuse generic lifecycle usage
  checks while retaining custom-field value validation.
- Modify `backend/internal/application/master_data_test.go`: document-version and import behavior.
- Modify `backend/internal/application/backup.go`: backup version 16 and restore reconciliation hook.
- Modify `backend/internal/application/backup_article_master_data.go`: preserve origin in legacy
  article master-data snapshots.
- Modify `backend/internal/application/backup_test.go` and
  `backend/internal/application/backup_article_master_data_test.go`: old/new restore coverage.

### HTTP and API contract

- Modify `backend/internal/api/routes.go`: add active-state PATCH route.
- Modify `backend/internal/api/data_handlers.go`: management listing, active-state handler, and
  conflict problem mappings.
- Modify `backend/internal/api/master_data_api_test.go`: management, PATCH, and conflict tests.
- Modify `openapi/railkeeper.yaml`: origin, capabilities, management query, PATCH request, and 409
  responses.
- Modify `frontend/src/shared/api.ts`: matching TypeScript types and API methods.

### Shared frontend behavior

- Create `frontend/src/shared/masterDataOptions.ts`: active plus current-historical option helper.
- Create `frontend/src/shared/masterDataOptions.test.ts`: helper tests.
- Modify `frontend/src/shared/functionSymbols.tsx`: hide inactive symbols except the current value.
- Create `frontend/src/shared/functionSymbols.test.tsx`: inactive-symbol regression tests.
- Modify `frontend/src/features/vehicles/VehiclesView.tsx`,
  `frontend/src/features/vehicles/VehicleModelTab.tsx`, and
  `frontend/src/features/vehicles/vehicleInventoryRenderers.tsx`: load all entries but offer only
  active plus current historical values.
- Modify `frontend/src/features/vehicles/VehicleModelTab.test.tsx`: historical vehicle values.
- Modify `frontend/src/features/accessories/ArticleCoreTab.tsx`: reuse the shared helper.
- Modify `frontend/src/features/accessories/ArticleEditorDialog.test.tsx`: retained inactive values.
- Modify `frontend/src/shared/i18n/de.ts` and `frontend/src/shared/i18n/en.ts`: add the shared inactive
  suffix used by vehicle and function-symbol controls.

### Settings UI and documentation

- Create `frontend/src/features/settings/MasterDataLifecycleActions.tsx`: compact shared actions.
- Create `frontend/src/features/settings/MasterDataLifecycleActions.test.tsx`: action visibility.
- Modify `frontend/src/features/settings/SettingsView.tsx`: managed general master-data table,
  confirmations, status, origin, and lifecycle commands.
- Modify `frontend/src/features/settings/ArticleManagementSettings.tsx`: same lifecycle semantics for
  stock units, article types, subtypes, and custom fields.
- Modify settings tests, German/English i18n, and `frontend/src/styles/settings.css`.
- Modify `docs/site/administration/index.md` and `docs/site/de/administration/index.md`: explain safe
  deactivate/reactivate/delete and update persistence.

---

### Task 1: Add origin storage and deterministic bundled-key reconciliation

**Files:**
- Create: `backend/migrations/0058_master_data_origin.sql`
- Create: `tools/build_bundled_master_data_manifest.py`
- Create: `backend/seeds/bundled_master_data_manifest.json` (generated)
- Create: `backend/internal/infrastructure/master_data_origin_migration_test.go`
- Modify: `backend/internal/infrastructure/master_data_seed.go`
- Modify: `backend/internal/infrastructure/master_data_seed_test.go`

**Interfaces:**
- Consumes: existing migration runner and `SeedMasterData(db, seedsDir)`.
- Produces: `master_data_entries.origin` with values `bundled|custom`; manifest schema
  `{"version":1,"entries":[{"type":"...","key":"..."}]}`.

- [ ] **Step 1: Write the failing migration and reseed tests**

Add a migration test that migrates through `0057`, inserts a custom row, applies `0058`, and proves
the row defaults to custom without changing its active state:

```go
func TestMasterDataOriginMigrationPreservesExistingRowsAsCustom(t *testing.T) {
	root := t.TempDir()
	migrationsDir := filepath.Join(root, "migrations")
	if err := os.Mkdir(migrationsDir, 0o700); err != nil {
		t.Fatal(err)
	}
	copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
		"0057_accessory_list_price.sql")
	db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := infrastructure.Migrate(db, migrationsDir); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
INSERT INTO master_data_entries(
  id, type, key, label, active, sort_order, metadata_json, created_at, updated_at
) VALUES('manufacturer:club', 'manufacturer', 'club', 'Club', 0, 0, '{}', 'now', 'now')`); err != nil {
		t.Fatal(err)
	}
	applyMigrationFile(t, db, "0058_master_data_origin.sql")
	var origin string
	var active int
	if err := db.QueryRow(`
SELECT origin, active FROM master_data_entries
WHERE type='manufacturer' AND key='club'`).Scan(&origin, &active); err != nil {
		t.Fatal(err)
	}
	if origin != "custom" || active != 0 {
		t.Fatalf("origin=%q active=%d", origin, active)
	}
}
```

Extend `master_data_seed_test.go` with a fixture that deactivates and renames one manifest key,
reruns seeding, and asserts `origin='bundled'`, the custom label remains, and `active=0`. Insert a
non-manifest custom row and assert it stays custom.

- [ ] **Step 2: Run the targeted tests and verify failure**

Run from `backend`:

```powershell
go test ./internal/infrastructure -run 'TestMasterDataOrigin|TestSeedMasterData' -count=1
```

Expected: FAIL because migration `0058`, the manifest, and origin-aware seeding do not exist.

- [ ] **Step 3: Add the migration**

Create `0058_master_data_origin.sql`:

```sql
ALTER TABLE master_data_entries
ADD COLUMN origin TEXT NOT NULL DEFAULT 'custom'
  CHECK (origin IN ('bundled', 'custom'));

CREATE INDEX idx_master_data_entries_origin
  ON master_data_entries(origin, type, key);
```

Do not mark all existing rows bundled in SQL. Real databases may contain user-created rows.

- [ ] **Step 4: Add the deterministic manifest generator**

Implement the generator with only Python standard-library modules. It applies every migration to an
in-memory SQLite database, inserts the normal JSON seed, queries unique `(type, key)` pairs, sorts
them, and writes stable UTF-8 JSON:

```python
#!/usr/bin/env python3
import argparse
import json
import sqlite3
from pathlib import Path


def build_manifest(migrations_dir: Path, seed_path: Path) -> dict[str, object]:
    connection = sqlite3.connect(":memory:")
    try:
        for migration in sorted(migrations_dir.glob("*.sql")):
            connection.executescript(migration.read_text(encoding="utf-8"))
        seed = json.loads(seed_path.read_text(encoding="utf-8"))
        for entry in seed["entries"]:
            connection.execute(
                """
                INSERT OR IGNORE INTO master_data_entries(
                  id, type, key, label, active, sort_order, source_url,
                  metadata_json, created_at, updated_at, origin
                ) VALUES(?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'), 'bundled')
                """,
                (
                    entry["id"], entry["type"], entry["key"], entry["label"],
                    int(entry.get("active", True)), entry.get("sortOrder", 0),
                    entry.get("sourceUrl", ""),
                    json.dumps(entry.get("metadata", {}), ensure_ascii=False,
                               separators=(",", ":")),
                ),
            )
        rows = connection.execute(
            "SELECT DISTINCT type, key FROM master_data_entries ORDER BY type, key"
        ).fetchall()
        return {"version": 1, "entries": [
            {"type": type_name, "key": key} for type_name, key in rows
        ]}
    finally:
        connection.close()


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--migrations", type=Path, required=True)
    parser.add_argument("--seed", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    document = build_manifest(args.migrations, args.seed)
    args.output.write_text(
        json.dumps(document, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
        newline="\n",
    )


if __name__ == "__main__":
    main()
```

Generate the committed manifest from the repository root:

```powershell
python tools/build_bundled_master_data_manifest.py `
  --migrations backend/migrations `
  --seed backend/seeds/master_data.json `
  --output backend/seeds/bundled_master_data_manifest.json
```

Run the command twice and verify `git diff` is unchanged after the second run.

- [ ] **Step 5: Make seeding origin-aware without overwriting user state**

Add manifest types and loading to `master_data_seed.go`:

```go
type bundledMasterDataManifest struct {
	Version int                    `json:"version"`
	Entries []bundledMasterDataKey `json:"entries"`
}

type bundledMasterDataKey struct {
	Type string `json:"type"`
	Key  string `json:"key"`
}
```

Insert normal seed rows with `origin='bundled'` and use this conflict clause only:

```sql
ON CONFLICT(type, key) DO UPDATE SET origin='bundled'
```

Then reconcile every manifest pair with:

```sql
UPDATE master_data_entries
SET origin='bundled'
WHERE type=? AND key=?
```

Reject a missing or invalid manifest when `master_data.json` exists. This prevents a new binary from
silently treating shipped entries as custom. Keep the current no-op behavior when the seed directory
contains neither file.

- [ ] **Step 6: Run tests and the full infrastructure package**

```powershell
cd backend
gofmt -w internal/infrastructure/master_data_seed.go `
  internal/infrastructure/master_data_seed_test.go `
  internal/infrastructure/master_data_origin_migration_test.go
go test ./internal/infrastructure -run 'TestMasterDataOrigin|TestSeedMasterData' -count=1
go test ./internal/infrastructure -count=1
```

Expected: PASS. The reseed test must prove label, metadata, sort order, source URL, and inactive state
are unchanged.

- [ ] **Step 7: Commit**

```powershell
git add backend/migrations/0058_master_data_origin.sql `
  backend/internal/infrastructure/master_data_seed.go `
  backend/internal/infrastructure/master_data_seed_test.go `
  backend/internal/infrastructure/master_data_origin_migration_test.go `
  backend/seeds/bundled_master_data_manifest.json `
  tools/build_bundled_master_data_manifest.py
git commit -m "feat: preserve bundled master data origin"
```

### Task 2: Implement origins, capabilities, references, and lifecycle commands

**Files:**
- Create: `backend/internal/application/master_data_lifecycle.go`
- Create: `backend/internal/application/master_data_lifecycle_test.go`
- Modify: `backend/internal/application/master_data.go`
- Modify: `backend/internal/application/master_data_custom_fields.go`

**Interfaces:**
- Consumes: `master_data_entries.origin` from Task 1.
- Produces:
  - `MasterDataOrigin` with `bundled` and `custom` constants.
  - `MasterDataCapabilities{CanDeactivate, CanReactivate, CanDelete}`.
  - `ListForManagement`, `ListAllForManagement`, and `SetActive` methods.
  - errors `ErrMasterDataBundled`, `ErrMasterDataInUse`, and `ErrMasterDataUsageUnknown`.

- [ ] **Step 1: Write failing lifecycle tests**

Create tests for origin scans, custom creation, bundled delete rejection, custom usage, relation
cleanup, management capabilities, and active-state cache invalidation. Use table-driven references:

```go
func TestMasterDataDeleteRejectsBundledAndUsedCustomEntries(t *testing.T) {
	tests := []struct {
		name     string
		typeName string
		key      string
		label    string
		useSQL   string
		useArgs  []any
	}{
		{
			name: "vehicle manufacturer", typeName: "manufacturer", key: "club", label: "Club",
			useSQL: `INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, category, created_at, updated_at
) VALUES('v1', 'RK-LOK-1', ?, 'Test', 'H0', 'Lokomotive', 'now', 'now')`,
			useArgs: []any{"Club"},
		},
		{
			name: "function symbol", typeName: "symbols", key: "club-light", label: "Club light",
			useSQL: `INSERT INTO vehicle_functions(
  id, vehicle_id, function_key, symbol_key, created_at, updated_at
) VALUES('f1', 'v1', 'F0', ?, 'now', 'now')`,
			useArgs: []any{"club-light"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := testDB(t)
			service := application.NewMasterDataService(db)
			active := true
			if _, err := service.Create(t.Context(), test.typeName, application.MasterDataInput{
				Key: test.key, Label: test.label, Active: &active,
			}); err != nil {
				t.Fatal(err)
			}
			if test.typeName == "symbols" {
				if _, err := db.Exec(`
INSERT INTO vehicles(
  id, inventory_number, manufacturer, name, gauge, category, created_at, updated_at
) VALUES('v1', 'RK-LOK-1', 'Club', 'Test', 'H0', 'Lokomotive', 'now', 'now')`); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := db.Exec(test.useSQL, test.useArgs...); err != nil {
				t.Fatal(err)
			}
			if err := service.Delete(t.Context(), test.typeName, test.key); !errors.Is(
				err, application.ErrMasterDataInUse,
			) {
				t.Fatalf("delete error=%v", err)
			}
			if _, err := service.Get(t.Context(), test.typeName, test.key); err != nil {
				t.Fatalf("used entry changed: %v", err)
			}
		})
	}
}
```

Add separate concrete cases for vehicle category, gattung, gauge, epoch, railway company; exhibition
snapshot values; accessory manufacturer, gauge JSON, stock unit, article type, subtype, and custom
attribute keys. Add a recognized zero-reference case for `cv8_manufacturer`.

- [ ] **Step 2: Run the tests and verify failure**

```powershell
cd backend
go test ./internal/application -run 'TestMasterData.*(Origin|Lifecycle|Delete|Capabilities|Active)' -count=1
```

Expected: FAIL because origin fields, lifecycle errors, and methods are undefined.

- [ ] **Step 3: Extend the entry model and all SQL scans**

Add these types to `master_data.go`:

```go
type MasterDataOrigin string

const (
	MasterDataOriginBundled MasterDataOrigin = "bundled"
	MasterDataOriginCustom  MasterDataOrigin = "custom"
)

type MasterDataCapabilities struct {
	CanDeactivate bool `json:"canDeactivate"`
	CanReactivate bool `json:"canReactivate"`
	CanDelete     bool `json:"canDelete"`
}
```

Extend `MasterDataEntry`:

```go
Origin       MasterDataOrigin         `json:"origin"`
Capabilities *MasterDataCapabilities  `json:"capabilities,omitempty"`
```

Add `origin` to every `SELECT`, scanner, `INSERT`, and import row. Normal `Create` always inserts
`custom`; normal `Update` never writes origin. Update cache cloning so the optional capabilities
pointer is copied, not aliased.

- [ ] **Step 4: Implement explicit usage adapters**

Create `master_data_lifecycle.go` with a queryer interface and explicit rules. Values marked
`masterDataUseLabel` compare against the editable label; values marked `masterDataUseKey` compare
against the immutable key:

```go
type masterDataQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type masterDataUsageValue int

const (
	masterDataUseLabel masterDataUsageValue = iota
	masterDataUseKey
)

type masterDataUsageRule struct {
	value masterDataUsageValue
	query string
}

var masterDataUsageRules = map[string][]masterDataUsageRule{
	"manufacturer": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE manufacturer=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE manufacturer=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE manufacturer=? COLLATE NOCASE)`},
	},
	"gauge": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE gauge=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(
  SELECT 1 FROM accessory_products products, json_each(products.gauges_json)
  WHERE json_each.value=? COLLATE NOCASE
)`},
	},
	"epoch": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE epoch=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE epoch=? COLLATE NOCASE)`},
	},
	"railway_company": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE railway_company=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE railway_company=? COLLATE NOCASE)`},
	},
	"vehicle_category": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE category=? COLLATE NOCASE)`},
	},
	"vehicle_gattung": {
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM vehicles WHERE gattung=? COLLATE NOCASE)`},
		{masterDataUseLabel, `SELECT EXISTS(SELECT 1 FROM exhibition_entries WHERE gattung=? COLLATE NOCASE)`},
	},
	"symbols": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM vehicle_functions WHERE symbol_key=?)`},
	},
	"stock_unit": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE stock_unit=?)`},
	},
	"article_type": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE article_type=?)`},
	},
	"accessory_subtype": {
		{masterDataUseKey, `SELECT EXISTS(SELECT 1 FROM accessory_products WHERE subtype=?)`},
	},
	"accessory_custom_field": {
		{masterDataUseKey, `SELECT EXISTS(
  SELECT 1 FROM accessory_product_attributes attributes
  JOIN accessory_products products ON products.id=attributes.product_id
  WHERE attributes.attribute_key=? AND products.article_type='other'
)`},
	},
	"cv8_manufacturer": {},
}
```

Do not treat `master_data_relations` as domain usage. Unknown types return
`ErrMasterDataUsageUnknown` so permanent deletion fails closed.

- [ ] **Step 5: Implement management capabilities and active-state updates**

Add:

```go
func (s *MasterDataService) SetActive(
	ctx context.Context, typeName, key string, active bool,
) (*MasterDataEntry, error)

func (s *MasterDataService) ListForManagement(
	ctx context.Context, typeName string,
) ([]MasterDataEntry, error)

func (s *MasterDataService) ListAllForManagement(
	ctx context.Context,
) (map[string][]MasterDataEntry, error)
```

`SetActive` changes only `active` and `updated_at`, then invalidates the cache. Management methods
set `CanDeactivate=entry.Active`, `CanReactivate=!entry.Active`, and `CanDelete=true` only for an
unused custom entry with a known reference adapter. Normal list methods leave `Capabilities=nil`.

- [ ] **Step 6: Replace physical deletion with one reserved transaction**

Refactor `Delete` to:

1. begin a transaction and reserve the write lock;
2. load origin, type, key, and label;
3. reject bundled entries with `ErrMasterDataBundled`;
4. call the same usage adapter used by capabilities;
5. reject used entries with `ErrMasterDataInUse`;
6. delete relations where the row is parent or child;
7. delete the row, commit, and invalidate the cache.

Remove `deleteAccessoryCustomField`; keep custom-field metadata/value validation used by Update and
Import. Wrap errors with type and key while preserving `errors.Is`.

- [ ] **Step 7: Format and run application tests**

```powershell
cd backend
gofmt -w internal/application/master_data.go `
  internal/application/master_data_custom_fields.go `
  internal/application/master_data_lifecycle.go `
  internal/application/master_data_lifecycle_test.go
go test ./internal/application -run 'TestMasterData.*(Origin|Lifecycle|Delete|Capabilities|Active)' -count=1
go test ./internal/application -count=1
```

Expected: PASS, including custom-field regressions already present in `master_data_test.go`.

- [ ] **Step 8: Commit**

```powershell
git add backend/internal/application/master_data.go `
  backend/internal/application/master_data_custom_fields.go `
  backend/internal/application/master_data_lifecycle.go `
  backend/internal/application/master_data_lifecycle_test.go
git commit -m "feat: enforce master data lifecycle rules"
```

### Task 3: Make master-data export and import origin-safe

**Files:**
- Modify: `backend/internal/application/master_data.go`
- Modify: `backend/internal/application/master_data_test.go`

**Interfaces:**
- Consumes: origin and usage helpers from Task 2.
- Produces: master-data document version 2 and transactional reconciliation for v1/v2 imports.

- [ ] **Step 1: Write failing export/import tests**

Add tests that prove:

```go
func TestMasterDataExportVersionTwoIncludesOrigin(t *testing.T) {
	db := testDB(t)
	if _, err := db.Exec(`
UPDATE master_data_entries SET origin='bundled'
WHERE type='article_type'`); err != nil {
		t.Fatal(err)
	}
	service := application.NewMasterDataService(db)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 2 {
		t.Fatalf("version=%d", doc.Version)
	}
	if doc.Entries["article_type"][0].Origin != application.MasterDataOriginBundled {
		t.Fatalf("origin=%q", doc.Entries["article_type"][0].Origin)
	}
}
```

Also cover: v1 accepted; v3 rejected before mutation; imported `bundled` cannot promote an unknown
key; omitted bundled rows and their bundled relations survive; omitted unused custom rows disappear;
omitted used custom rows reject import before mutation; imported inactive bundled rows stay inactive.

- [ ] **Step 2: Run tests and verify failure**

```powershell
cd backend
go test ./internal/application -run 'TestMasterData(Export|Import)' -count=1
```

Expected: FAIL because export still writes version 1 and import still clears all entries.

- [ ] **Step 3: Build the reconciled desired entry set before mutation**

Set `const masterDataExportVersion = 2` and accept only document versions 1 and 2. Before opening
the write transaction, normalize imported buckets. Inside the reserved transaction, load current
rows into a map keyed by `(type, key)` and construct the desired set with these exact rules:

```go
switch {
case currentExists && current.Origin == MasterDataOriginBundled:
	imported.Origin = MasterDataOriginBundled
case currentExists:
	imported.Origin = MasterDataOriginCustom
default:
	imported.Origin = MasterDataOriginCustom
}
```

For every current row omitted from the document:

- retain it unchanged if bundled;
- reject the whole import if custom and in use;
- otherwise omit it from the desired set so it is deleted.

Do not trust `entry.Origin` from JSON when deciding ownership.

- [ ] **Step 4: Reconcile relations and replace rows transactionally**

Build a relation identity from `(parentType, parentKey, childType, childKey)`. Start with imported
relations, then retain an omitted current relation when both endpoints are retained bundled entries.
Drop relations to custom entries that are legitimately removed. Validate all final relation
endpoints before deleting current rows.

Insert `origin` explicitly when writing desired entries. Keep protected article-type handling,
legacy `article_subtype` normalization, custom-field validation, and all-or-nothing transaction
behavior.

- [ ] **Step 5: Run targeted and full application tests**

```powershell
cd backend
gofmt -w internal/application/master_data.go internal/application/master_data_test.go
go test ./internal/application -run 'TestMasterData(Export|Import)' -count=1
go test ./internal/application -count=1
```

Expected: PASS. Existing import tests must remain green without weakening protected article types.

- [ ] **Step 6: Commit**

```powershell
git add backend/internal/application/master_data.go `
  backend/internal/application/master_data_test.go
git commit -m "feat: reconcile master data imports safely"
```

### Task 4: Preserve origin through backup and restore

**Files:**
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_article_master_data.go`
- Modify: `backend/internal/application/backup_test.go`
- Modify: `backend/internal/application/backup_article_master_data_test.go`

**Interfaces:**
- Consumes: `master_data_entries.origin` and bundled origin constant.
- Produces: backup version 16 with backwards-compatible restore reconciliation.

- [ ] **Step 1: Write failing backup tests**

Add one round-trip test for version 16 and one old-backup test:

```go
func TestBackupRestorePreservesInactiveBundledAndCustomOrigins(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	if _, err := db.Exec(`
UPDATE master_data_entries
SET origin='bundled', active=0, label='Gleismaterial'
WHERE type='article_type' AND key='track'`); err != nil {
		t.Fatal(err)
	}
	active := true
	masterData := application.NewMasterDataService(db)
	if _, err := masterData.Create(t.Context(), "manufacturer", application.MasterDataInput{
		Key: "club", Label: "Club", Active: &active,
	}); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 16 {
		t.Fatalf("version=%d", doc.Version)
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	assertMasterDataOrigin(t, db, "article_type", "track", "bundled", false)
	assertMasterDataOrigin(t, db, "manufacturer", "club", "custom", true)
}

func TestBackupRestoreVersion15ReconcilesCurrentBundledKeys(t *testing.T) {
	dataDir := t.TempDir()
	db := backupTestDB(t, dataDir)
	if _, err := db.Exec(`
UPDATE master_data_entries SET origin='bundled', active=0
WHERE type='article_type' AND key='track'`); err != nil {
		t.Fatal(err)
	}
	service := application.NewBackupService(db, dataDir)
	doc, err := service.Export(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	doc.Version = 15
	for _, row := range doc.Tables["master_data_entries"] {
		delete(row, "origin")
	}
	if _, err := service.Import(t.Context(), doc); err != nil {
		t.Fatal(err)
	}
	assertMasterDataOrigin(t, db, "article_type", "track", "bundled", false)
}
```

Extend legacy article-master-data tests so version 1/2 restores keep the current origin for protected
article types and subtypes.

- [ ] **Step 2: Run tests and verify failure**

```powershell
cd backend
go test ./internal/application -run 'TestBackup.*(Origin|Bundled)|TestLegacy.*MasterData' -count=1
```

Expected: FAIL because backup version is 15 and old rows default to custom after restore.

- [ ] **Step 3: Add restore reconciliation**

Set `backupVersion = 16`. Before clearing tables, read current bundled identities inside the restore
transaction:

```go
type masterDataIdentity struct {
	Type string
	Key  string
}

func readBundledMasterDataIdentities(
	ctx context.Context, tx *sql.Tx,
) ([]masterDataIdentity, error)
```

After generic row insertion and legacy article restoration, run parameterized updates for only those
identities:

```sql
UPDATE master_data_entries
SET origin='bundled'
WHERE type=? AND key=?
```

Rows unknown to the current installation remain custom. Do not change `active`, label, metadata,
sort order, source URL, or timestamps during reconciliation.

- [ ] **Step 4: Preserve origin in legacy article snapshots**

Add `origin string` to `backupArticleMasterDataEntry`, include it in the pre-restore SELECT and
legacy reinsert. Current schema always has the column; old backup document versions affect document
rows, not the live pre-restore schema.

- [ ] **Step 5: Run backup and full application tests**

```powershell
cd backend
gofmt -w internal/application/backup.go `
  internal/application/backup_article_master_data.go `
  internal/application/backup_test.go `
  internal/application/backup_article_master_data_test.go
go test ./internal/application -run 'TestBackup|TestLegacy.*MasterData' -count=1
go test ./internal/application -count=1
```

Expected: PASS for every supported backup version and existing destructive-restore rollback test.

- [ ] **Step 6: Commit**

```powershell
git add backend/internal/application/backup.go `
  backend/internal/application/backup_article_master_data.go `
  backend/internal/application/backup_test.go `
  backend/internal/application/backup_article_master_data_test.go
git commit -m "feat: preserve master data lifecycle in backups"
```

### Task 5: Expose lifecycle management through HTTP and OpenAPI

**Files:**
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/data_handlers.go`
- Modify: `backend/internal/api/master_data_api_test.go`
- Modify: `openapi/railkeeper.yaml`
- Modify: `frontend/src/shared/api.ts`

**Interfaces:**
- Consumes: management and active-state methods from Task 2.
- Produces:
  - `GET ...?management=true` responses with capabilities.
  - `PATCH /api/v1/master-data/{type}/{key}/active` with `{active:boolean}`.
  - TypeScript `origin`, optional `capabilities`, `managedMasterData`,
    `managedMasterDataAll`, and `setMasterDataActive`.

- [ ] **Step 1: Write failing API tests**

Add tests that authenticate an Editor and assert:

```go
response := doAuthedJSON(t, router, http.MethodPatch,
	"/api/v1/master-data/manufacturer/club/active",
	`{"active":false}`, session, cookies, http.StatusOK)
var entry application.MasterDataEntry
if err := json.Unmarshal(response.Body.Bytes(), &entry); err != nil {
	t.Fatal(err)
}
if entry.Active {
	t.Fatal("entry remained active")
}
```

Test `management=true` for bundled, unused custom, and used custom capabilities. Test DELETE returns
409 with `master_data_bundled` or `master_data_in_use`, and malformed PATCH returns 400. Retain 404
and role/CSRF coverage.

- [ ] **Step 2: Run API tests and verify failure**

```powershell
cd backend
go test ./internal/api -run TestMasterDataAPI -count=1
```

Expected: FAIL because the PATCH route and management query do not exist.

- [ ] **Step 3: Implement handlers and stable problem mappings**

Add:

```go
type masterDataActiveInput struct {
	Active *bool `json:"active"`
}
```

Require a non-nil `active` property. For list handlers, dispatch to management methods only when
`management=true`; otherwise keep cached normal listing. Map errors:

```go
case errors.Is(err, application.ErrMasterDataBundled):
	respondProblem(w, http.StatusConflict, "master_data_bundled",
		"Bundled master data cannot be permanently deleted.")
case errors.Is(err, application.ErrMasterDataInUse):
	respondProblem(w, http.StatusConflict, "master_data_in_use",
		"This master data entry is still in use and can only be deactivated.")
```

Register the PATCH route with `routeAccessEditor`. Existing CSRF middleware must protect it.

- [ ] **Step 4: Update OpenAPI and frontend API types**

Add exact schemas:

```yaml
MasterDataCapabilities:
  type: object
  required: [canDeactivate, canReactivate, canDelete]
  properties:
    canDeactivate: { type: boolean }
    canReactivate: { type: boolean }
    canDelete: { type: boolean }
MasterDataActiveInput:
  type: object
  required: [active]
  properties:
    active: { type: boolean }
```

Make `origin` required with enum `[bundled, custom]`; make `capabilities` optional. Add the
management query to both list endpoints, document PATCH, and document 409 DELETE responses. Change
the master-data document example to version 2.

Add matching TypeScript:

```ts
export type MasterDataOrigin = "bundled" | "custom";
export type MasterDataCapabilities = {
  canDeactivate: boolean;
  canReactivate: boolean;
  canDelete: boolean;
};
```

Model `origin?: MasterDataOrigin` and `capabilities?: MasterDataCapabilities` in the frontend adapter.
The OpenAPI response requires origin, while the optional frontend property keeps non-management
fixtures and a briefly stale cached payload backwards compatible. Settings management must use
capabilities and must never derive delete permission from a missing origin.

Add explicit methods rather than another positional boolean:

```ts
managedMasterData: (type: string) =>
  request<MasterDataEntry[]>(`/master-data/${encodeURIComponent(type)}?management=true`),
managedMasterDataAll: () =>
  request<Record<string, MasterDataEntry[]>>("/master-data-all?management=true"),
setMasterDataActive: (type: string, key: string, active: boolean) =>
  request<MasterDataEntry>(
    `/master-data/${encodeURIComponent(type)}/${encodeURIComponent(key)}/active`,
    { method: "PATCH", body: JSON.stringify({ active }) }
  ),
```

- [ ] **Step 5: Format and verify backend/API contract**

```powershell
cd backend
gofmt -w internal/api/routes.go internal/api/data_handlers.go `
  internal/api/master_data_api_test.go
go test ./internal/api -run TestMasterDataAPI -count=1
go test ./... -count=1
cd ..\frontend
npm.cmd run build
```

Expected: PASS. TypeScript build proves frontend types match existing call sites.

- [ ] **Step 6: Commit**

```powershell
git add backend/internal/api/routes.go backend/internal/api/data_handlers.go `
  backend/internal/api/master_data_api_test.go openapi/railkeeper.yaml `
  frontend/src/shared/api.ts
git commit -m "feat: expose master data lifecycle API"
```

### Task 6: Preserve inactive historical options in vehicles, accessories, and symbols

**Files:**
- Create: `frontend/src/shared/masterDataOptions.ts`
- Create: `frontend/src/shared/masterDataOptions.test.ts`
- Modify: `frontend/src/shared/functionSymbols.tsx`
- Create: `frontend/src/shared/functionSymbols.test.tsx`
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Modify: `frontend/src/features/vehicles/VehicleModelTab.tsx`
- Modify: `frontend/src/features/vehicles/vehicleInventoryRenderers.tsx`
- Modify: `frontend/src/features/vehicles/VehicleModelTab.test.tsx`
- Modify: `frontend/src/features/accessories/ArticleCoreTab.tsx`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: all-entry master-data APIs and `MasterDataEntry.active`.
- Produces: `masterDataOptions(entries, currentValues, persistedValue)` and UI behavior that offers
  active values plus only the record's current inactive/unmatched values.

- [ ] **Step 1: Write failing helper and component tests**

Create the helper test:

```ts
it("keeps only active entries plus current inactive and legacy values", () => {
  const options = masterDataOptions(
    [active("piko", "Piko"), inactive("roco", "Roco"), inactive("esu", "ESU")],
    ["Roco", "Legacy"],
    (entry) => entry.label
  );
  expect(options.map(({ value, active }) => ({ value, active }))).toEqual([
    { value: "Piko", active: true },
    { value: "Roco", active: false },
    { value: "Legacy", active: false }
  ]);
});
```

Add vehicle assertions that a current inactive manufacturer renders `Roco (inaktiv)` and an
unrelated inactive manufacturer is absent. Add accessory assertions for manufacturer, gauge, and
stock unit. Add function-symbol tests proving an inactive current symbol remains visible with an
inactive suffix, while inactive unrelated symbols and deactivated fallback keys are absent.

- [ ] **Step 2: Run targeted frontend tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/shared/masterDataOptions.test.ts `
  src/shared/functionSymbols.test.tsx `
  src/features/vehicles/VehicleModelTab.test.tsx `
  src/features/accessories/ArticleEditorDialog.test.tsx
```

Expected: FAIL because the shared helper does not exist and vehicles load only active entries.

- [ ] **Step 3: Implement the shared option helper**

Create:

```ts
export type MasterDataOption = {
  id: string;
  value: string;
  label: string;
  active: boolean;
};

export function masterDataOptions(
  entries: readonly MasterDataEntry[],
  currentValues: readonly string[],
  persistedValue: (entry: MasterDataEntry) => string
): MasterDataOption[] {
  const current = new Set(currentValues.filter(Boolean));
  const options = entries
    .filter((entry) => entry.active || current.has(persistedValue(entry)))
    .map((entry) => ({
      id: entry.id,
      value: persistedValue(entry),
      label: entry.label,
      active: entry.active
    }));
  for (const value of current) {
    if (!options.some((option) => option.value === value)) {
      options.push({ id: `legacy:${value}`, value, label: value, active: false });
    }
  }
  return options;
}
```

Remove the duplicate local helper from `ArticleCoreTab.tsx` and import this function.

- [ ] **Step 4: Load all vehicle entries and pass current values to option rendering**

Change `VehiclesView` from `api.masterDataAll(true)` to `api.masterDataAll()`. Change the renderer
signature to:

```ts
selectOptions: (
  entries: MasterDataEntry[],
  currentValue: string,
  emptyLabel?: string
) => ReactNode;
```

Render `masterDataOptions` with label persistence, append ` (${t("common.inactive")})`, and disable
inactive options. Update every `VehicleModelTab` call with the matching current form value. Keep the
current inactive gattung after category filtering; unrelated inactive gattungen stay unavailable.

- [ ] **Step 5: Correct function-symbol fallback and historical behavior**

When server symbols are loaded, fallback entries must not resurrect an explicitly inactive shipped
symbol. Use fallback symbols only when the server list is empty. Build picker options from active
symbols plus the current inactive/unmatched key. `functionSymbolMetadata` must find the current key
regardless of active state so existing icons remain readable.

Append the localized inactive suffix and disable the current inactive button so it is visible but
cannot be newly selected elsewhere.

Add `"common.inactive": "inaktiv"` in German and `"common.inactive": "inactive"` in English.

- [ ] **Step 6: Run focused and full frontend tests**

```powershell
cd frontend
npm.cmd run test:run -- src/shared/masterDataOptions.test.ts `
  src/shared/functionSymbols.test.tsx `
  src/features/vehicles/VehicleModelTab.test.tsx `
  src/features/accessories/ArticleEditorDialog.test.tsx
npm.cmd run test:run
npm.cmd run build
```

Expected: all tests and build PASS. No TypeScript `any` or unchecked cast is introduced.

- [ ] **Step 7: Commit**

```powershell
git add frontend/src/shared/masterDataOptions.ts `
  frontend/src/shared/masterDataOptions.test.ts `
  frontend/src/shared/functionSymbols.tsx `
  frontend/src/shared/functionSymbols.test.tsx `
  frontend/src/features/vehicles/VehiclesView.tsx `
  frontend/src/features/vehicles/VehicleModelTab.tsx `
  frontend/src/features/vehicles/vehicleInventoryRenderers.tsx `
  frontend/src/features/vehicles/VehicleModelTab.test.tsx `
  frontend/src/features/accessories/ArticleCoreTab.tsx `
  frontend/src/features/accessories/ArticleEditorDialog.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: retain inactive historical master data"
```

### Task 7: Add safe lifecycle controls to both settings areas

**Files:**
- Create: `frontend/src/features/settings/MasterDataLifecycleActions.tsx`
- Create: `frontend/src/features/settings/MasterDataLifecycleActions.test.tsx`
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`
- Modify: `frontend/src/features/settings/ArticleManagementSettings.tsx`
- Modify: `frontend/src/features/settings/ArticleManagementSettings.test.tsx`
- Modify: `frontend/src/features/settings/settingsModel.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/settings.css`

**Interfaces:**
- Consumes: managed API methods and capabilities from Task 5.
- Produces: consistent status/origin/actions, confirmation copy, and lifecycle refresh behavior in
  general and article master-data settings.

- [ ] **Step 1: Write failing lifecycle action tests**

Create component tests for the exact matrix:

```ts
it.each([
  ["bundled active", bundledActive, ["Bearbeiten", "Deaktivieren"]],
  ["bundled inactive", bundledInactive, ["Bearbeiten", "Reaktivieren"]],
  ["used custom", usedCustom, ["Bearbeiten", "Deaktivieren"]],
  ["unused custom", unusedCustom, ["Bearbeiten", "Deaktivieren", "Endgültig löschen"]]
])("shows %s actions", (_name, entry, labels) => {
  render(<MasterDataLifecycleActions entry={entry} disabled={false}
    onEdit={vi.fn()} onSetActive={vi.fn()} onDelete={vi.fn()} />);
  expect(screen.getAllByRole("button").map((button) => button.getAttribute("aria-label")))
    .toEqual(labels);
});
```

Extend SettingsView and ArticleManagementSettings tests to verify managed endpoints, status/origin
labels, muted inactive rows, deactivation confirmation body, direct reactivation, permanent-delete
confirmation, and reload/update after success.

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/MasterDataLifecycleActions.test.tsx `
  src/features/settings/SettingsView.test.tsx `
  src/features/settings/ArticleManagementSettings.test.tsx
```

Expected: FAIL because the lifecycle component and managed API calls are not wired.

- [ ] **Step 3: Implement the compact shared actions**

Use existing transparent icon buttons and Lucide `Pencil`, `Archive`, `ArchiveRestore`, and `Trash2`.
Render edit when allowed; render deactivate/reactivate from capabilities; render danger delete only
when `canDelete` is true. Every button gets translated `aria-label` and `title`.

Do not infer delete permission from `origin` in the browser. Capabilities are authoritative.

- [ ] **Step 4: Update general master-data management**

Load with `api.managedMasterDataAll()` and reload with `api.managedMasterData(activeType)`. Add status
and origin columns using compact text pills. Apply `muted-row` to inactive entries.

Use `api.setMasterDataActive`. Deactivation uses the existing SettingsView confirmation layer with
this meaning in both languages:

```text
The entry will no longer be offered for new records. Existing saved uses remain unchanged.
```

Reactivation executes immediately. Permanent deletion uses danger styling and the phrase
"Permanently delete". Keep `startEdit` and editable metadata behavior unchanged.

- [ ] **Step 5: Update article master-data management**

Load each type with `api.managedMasterData(type)`. Replace the current full-record active update with
`api.setMasterDataActive`. Add origin and status cells and reuse `MasterDataLifecycleActions`.

Pass the SettingsView confirmation callback into `ArticleManagementSettings` so both areas use one
modal implementation. Add `onRemoved(type, key)` behavior that removes the deleted row and resets an
editor targeting it.

- [ ] **Step 6: Add exact German and English copy and compact styles**

Add keys for:

```text
settings.master.status
settings.master.origin
settings.master.origin.bundled
settings.master.origin.custom
settings.master.deactivate
settings.master.reactivate
settings.master.deletePermanently
settings.master.deactivateTitle
settings.master.deactivateBody
settings.master.deleteTitle
settings.master.deleteBody
```

German terminology: `Aktiv`, `Inaktiv`, `RailKeeper`, `Eigener Eintrag`, `Deaktivieren`,
`Reaktivieren`, `Endgültig löschen`. English terminology: `Active`, `Inactive`, `RailKeeper`,
`Custom entry`, `Deactivate`, `Reactivate`, `Permanently delete`.

Keep lifecycle columns narrow, allow the table wrapper to scroll on mobile, and do not add boxed
hover backgrounds to icon buttons.

- [ ] **Step 7: Run frontend tests and build**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/MasterDataLifecycleActions.test.tsx `
  src/features/settings/SettingsView.test.tsx `
  src/features/settings/ArticleManagementSettings.test.tsx
npm.cmd run test:run
npm.cmd run build
```

Expected: PASS in German-default tests and explicit English copy assertions.

- [ ] **Step 8: Commit**

```powershell
git add frontend/src/features/settings/MasterDataLifecycleActions.tsx `
  frontend/src/features/settings/MasterDataLifecycleActions.test.tsx `
  frontend/src/features/settings/SettingsView.tsx `
  frontend/src/features/settings/SettingsView.test.tsx `
  frontend/src/features/settings/ArticleManagementSettings.tsx `
  frontend/src/features/settings/ArticleManagementSettings.test.tsx `
  frontend/src/features/settings/settingsModel.ts `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts `
  frontend/src/styles/settings.css
git commit -m "feat: manage master data lifecycle safely"
```

### Task 8: Document, verify, and visually inspect #83

**Files:**
- Modify: `docs/site/administration/index.md`
- Modify: `docs/site/de/administration/index.md`
- Test: `docs/scripts/validate-docs.test.mjs`

**Interfaces:**
- Consumes: completed lifecycle behavior from Tasks 1 through 7.
- Produces: bilingual operator guidance and final release-quality verification evidence.

- [ ] **Step 1: Add bilingual administration guidance**

Document these facts in matching English and German sections:

- RailKeeper-shipped entries can be deactivated but not permanently deleted.
- Custom unused entries can be permanently deleted.
- Used custom entries can only be deactivated.
- Existing records retain inactive values; new records cannot select them.
- Deactivation and customization survive restart, reseed, update, export/import, and backup/restore.

Do not document #84's Windows path migration or #82's ZIP update flow here.

- [ ] **Step 2: Validate documentation and commit it**

```powershell
cd docs
npm.cmd run check
cd ..
git add docs/site/administration/index.md docs/site/de/administration/index.md
git commit -m "docs: explain master data lifecycle"
```

Expected: documentation validation PASS and both languages contain equivalent lifecycle guidance.

- [ ] **Step 3: Run repository-wide automated verification**

```powershell
cd backend
gofmt -w (Get-ChildItem internal/application,internal/api,internal/infrastructure `
  -Filter *.go -File | Select-Object -ExpandProperty FullName)
go test ./... -count=1
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
cd ..\docs
npm.cmd run check
cd ..
git diff --check origin/main...HEAD
git status --short --branch
```

Expected: backend PASS, all frontend tests PASS, production build PASS, docs PASS, no whitespace
errors, and no uncommitted files.

- [ ] **Step 4: Perform desktop and mobile visual checks**

Run the branch locally with a temporary copy of test data. Check dark and light themes at desktop
and mobile widths:

1. general Stammdaten table shows status, origin, and correct actions;
2. article Stammdaten shows the same lifecycle behavior;
3. deactivate a bundled manufacturer, reload/restart, and confirm it remains inactive;
4. edit a vehicle using it and confirm `inaktiv` remains visible;
5. create a vehicle and confirm the inactive manufacturer is unavailable;
6. repeat for accessory manufacturer, gauge, stock unit, subtype, and custom field;
7. verify an existing inactive function symbol remains readable but cannot be newly selected;
8. verify confirmation dialogs, long German labels, and action menus are not clipped.

Record any visual defect as a failing verification and fix it with a targeted test before continuing.

- [ ] **Step 5: Review the final diff against the specification**

```powershell
git log --oneline origin/main..HEAD
git diff --stat origin/main...HEAD
git diff origin/main...HEAD -- `
  backend/migrations backend/internal/application backend/internal/api `
  backend/internal/infrastructure frontend/src openapi docs/site
```

Confirm every design section has an implementation and test, no #82/#84 code entered the branch,
and generated manifest changes match the generator output.
