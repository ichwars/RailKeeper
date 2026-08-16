# Vehicle Creation And Inventory UI Correction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the mismatched vehicle-set inventory and creation UI with the approved desktop and mobile hierarchy, including persistent `RK-SET` inventory numbers and embedded article-data review.

**Architecture:** Extend the existing modular monolith without replacing the vehicle or article-search domains. The backend owns canonical set identity and shared data; the vehicle list exposes one canonical nested set projection per member. The frontend derives grouped rows/cards from that projection and runs creation through a reducer-backed wizard whose article-search controller remains shared with existing edit contexts.

**Tech Stack:** Go 1.25, `net/http`, SQLite migrations, React 19, TypeScript 7, Vite 8, Vitest, Testing Library, CSS design tokens, OpenAPI 3.

## Global Constraints

- Work only in `C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\vehicle-ui-correction` on branch `dev/vehicle-ui-correction`.
- Keep one correction branch and one pull request with reviewable commits.
- Use migration `0060_vehicle_set_inventory_number.sql`; do not rewrite migration `0059_vehicle_sets.sql`.
- Use backup format version `18`; versions `1` through `17` remain importable.
- Default set numbering is category `Set`, prefix `RK-SET`, padding `6`, first number `1`.
- Set and member inventory numbers are server-assigned and remain read-only previews during creation.
- Do not invent lifecycle/status values and do not expose complete-set deletion.
- Preserve existing article-search providers, ranking, enrichment, extraction, URL safety, and standalone edit/detail dialog.
- Preserve vehicle-based report, exhibition, filter, sort, and selection behavior.
- Maintain German and English translations, keyboard semantics, dark/light themes, and mobile touch targets of at least 44 by 44 CSS pixels.
- Do not commit `frontend/dist`, `frontend/node_modules`, `.cache`, data files, screenshots, local credentials, or generated packages.
- Before merge, show the running implementation to the user at the agreed desktop and mobile states; merge and release only after that checkpoint and all checks pass.

---

## Planned File Structure

### Backend

- Create `backend/migrations/0060_vehicle_set_inventory_number.sql`: schema, default scheme, deterministic backfill, constraints.
- Create `backend/internal/infrastructure/vehicle_set_inventory_number_migration_test.go`: migration behavior and constraints.
- Modify `backend/internal/application/vehicle_types.go`: canonical `VehicleSetSummary`, nested vehicle projection, set identity, not-found error.
- Modify `backend/internal/application/vehicle_sets.go`: reserve set number, read set, update shared data transactionally.
- Modify `backend/internal/application/vehicle_sets_test.go`: number reservation, read/update, rollback, projection tests.
- Modify `backend/internal/application/vehicles.go` and `vehicle_repository.go`: select and scan canonical set summaries.
- Modify `backend/internal/api/vehicle_handlers.go`, `routes.go`, and API tests: GET/PATCH set endpoints and role/CSRF behavior.
- Modify `openapi/railkeeper.yaml` and `openapi_contract_test.go`: exact public contract.
- Create `backend/internal/application/backup_vehicle_sets.go`: version-17 restore normalization and set scheme recovery.
- Modify `backend/internal/application/backup.go` and `backup_test.go`: backup version 18 compatibility.

### Frontend inventory

- Modify `frontend/src/shared/api.ts`: `VehicleSetSummary`, `VehicleSet`, GET/PATCH methods.
- Modify `frontend/src/features/vehicles/vehicleSetGroups.ts`: canonical grouping model and visible/total member counts.
- Modify `frontend/src/features/vehicles/vehicleTableColumns.ts`: non-sortable `type` presentation column and new defaults.
- Create `frontend/src/features/vehicles/VehicleSetInventoryRow.tsx`: aligned desktop set row and tri-state selection.
- Create `frontend/src/features/vehicles/VehicleSetInventoryMobileCard.tsx`: dedicated mobile set card.
- Create `frontend/src/features/vehicles/vehicleSetDuplicate.ts`: convert a loaded set into a number-free wizard prefill.
- Modify `VehicleInventoryTable.tsx`, `VehicleInventoryMobileList.tsx`, `VehicleInventoryPanel.tsx`, and `VehiclesView.tsx`: integration and set actions.
- Create `VehicleSetSummaryDialog.tsx` and `VehicleSetEditorDialog.tsx`: real View/Edit/Duplicate behavior without deletion.
- Modify inventory tests and `frontend/src/styles/vehicle-inventory.css`.

### Frontend wizard

- Create `frontend/src/features/vehicles/vehicleCreateWizardState.ts`: reducer, validation, member preservation, versioned local draft.
- Create `VehicleCreateWizardShell.tsx`, `VehicleCreateStepBasics.tsx`, `VehicleCreateStepArticle.tsx`, `VehicleCreateArticleResults.tsx`, `VehicleCreateArticleReview.tsx`, `VehicleCreateStepDetails.tsx`, and `VehicleSetDetailsTabs.tsx`.
- Modify `useArticleSearchController.ts`: exported controller contract usable by embedded and standalone presentations.
- Replace `VehicleCreateWizard.tsx` with orchestration and payload mapping only.
- Modify `VehicleEditorDialog.tsx` and `VehiclesView.tsx`: pass controller state to creation, retain dialogs outside creation.
- Modify `frontend/src/styles/vehicle-create-wizard.css`, `de.ts`, `en.ts`, and focused tests.

### Documentation and release

- Modify English and German vehicle guides and `docs/coverage.json` if source coverage changes.
- Create `docs/releases/v0.1.20.md` and update `backend/cmd/railkeeper/main.go` only after visual approval.

---

### Task 1: Persist and reserve set inventory numbers

**Files:**
- Create: `backend/migrations/0060_vehicle_set_inventory_number.sql`
- Create: `backend/internal/infrastructure/vehicle_set_inventory_number_migration_test.go`
- Modify: `backend/internal/application/vehicle_types.go`
- Modify: `backend/internal/application/vehicle_sets.go`
- Test: `backend/internal/application/vehicle_sets_test.go`

**Interfaces:**
- Consumes: `ReserveInventoryNumber(ctx, tx, category, fallbackCategory, ensureAvailable)` from `inventory_numbers.go`.
- Produces: `VehicleSet.InventoryNumber string`, `(*VehicleService).nextVehicleSetInventoryNumber`, and a transactional set number used by all later API/UI tasks.

- [ ] **Step 1: Write the failing migration and reservation tests**

Add tests that seed two pre-0060 sets in reverse ID order, apply the migration, and assert exact values and constraints:

```go
func TestVehicleSetInventoryNumberMigrationBackfillsDeterministically(t *testing.T) {
	db := migratedVehicleSetInventoryNumberTestDB(t)
	seedVehicleSetBeforeInventoryNumberMigration(t, db, "set-b", "2026-08-02T00:00:00Z")
	seedVehicleSetBeforeInventoryNumberMigration(t, db, "set-a", "2026-08-01T00:00:00Z")

	applyMigrationFile(t, db, "0060_vehicle_set_inventory_number.sql")

	assertText(t, db, `SELECT inventory_number FROM vehicle_sets WHERE id='set-a'`, "RK-SET-000001")
	assertText(t, db, `SELECT inventory_number FROM vehicle_sets WHERE id='set-b'`, "RK-SET-000002")
	assertText(t, db, `SELECT CAST(next_number AS TEXT) FROM inventory_number_schemes WHERE category='Set'`, "3")
	expectConstraintFailure(t, db, `UPDATE vehicle_sets SET inventory_number='' WHERE id='set-a'`)
	expectConstraintFailure(t, db, `UPDATE vehicle_sets SET inventory_number='RK-SET-000001' WHERE id='set-b'`)
}
```

Extend `vehicle_sets_test.go` with a successful create followed by a rolled-back conflicting create and another successful create. Assert set numbers `RK-SET-000001` then `RK-SET-000002`, proving the rollback consumed nothing.

- [ ] **Step 2: Run focused tests and verify failure**

Run:

```powershell
cd backend
go test ./internal/infrastructure ./internal/application -run 'VehicleSetInventoryNumber|CreateVehicleSetAssigns|CreateVehicleSetRollsBackSetNumber'
```

Expected: FAIL because migration `0060` and `VehicleSet.InventoryNumber` do not exist.

- [ ] **Step 3: Add the migration and transactional reservation**

Use the established article-number migration pattern, including an existing custom `Set` scheme:

```sql
ALTER TABLE vehicle_sets ADD COLUMN inventory_number TEXT NOT NULL DEFAULT '';

INSERT INTO inventory_number_schemes(
  id, category, prefix, next_number, padding, active, created_at, updated_at
)
VALUES(
  lower(hex(randomblob(16))), 'Set', 'RK-SET', 1, 6, 1, datetime('now'), datetime('now')
)
ON CONFLICT(category) DO NOTHING;

WITH ranked AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at, id) - 1 AS number_offset
  FROM vehicle_sets
)
UPDATE vehicle_sets
SET inventory_number = (
  SELECT printf('%s-%0*d', scheme.prefix, scheme.padding, scheme.next_number + ranked.number_offset)
  FROM ranked
  JOIN inventory_number_schemes scheme ON scheme.category = 'Set'
  WHERE ranked.id = vehicle_sets.id
)
WHERE inventory_number = '';

UPDATE inventory_number_schemes
SET next_number = next_number + (SELECT COUNT(*) FROM vehicle_sets), updated_at = datetime('now')
WHERE category = 'Set';

CREATE UNIQUE INDEX ux_vehicle_sets_inventory_number ON vehicle_sets(inventory_number);

CREATE TRIGGER vehicle_sets_inventory_number_required_insert
BEFORE INSERT ON vehicle_sets WHEN TRIM(NEW.inventory_number) = ''
BEGIN SELECT RAISE(ABORT, 'vehicle set inventory number required'); END;

CREATE TRIGGER vehicle_sets_inventory_number_required_update
BEFORE UPDATE OF inventory_number ON vehicle_sets WHEN TRIM(NEW.inventory_number) = ''
BEGIN SELECT RAISE(ABORT, 'vehicle set inventory number required'); END;
```

Add `InventoryNumber string \`json:"inventoryNumber"\`` to `VehicleSet`. In `CreateSet`, reserve before `insertVehicleSetTx` and pass the reserved value into both the INSERT and `vehicleSetFromInput`:

```go
setInventoryNumber, err := s.nextVehicleSetInventoryNumber(ctx, tx)
if err != nil {
	return nil, err
}
if err = insertVehicleSetTx(ctx, tx, setID, setInventoryNumber, setInput, now); err != nil {
	return nil, err
}
```

`nextVehicleSetInventoryNumber` must call `ReserveInventoryNumber(ctx, tx, "Set", "", ...)`; the availability callback queries `vehicle_sets` and returns `ErrInventoryNumberConflict` for a match.

- [ ] **Step 4: Format and run focused tests**

Run:

```powershell
cd backend
gofmt -w internal/infrastructure/vehicle_set_inventory_number_migration_test.go internal/application/vehicle_types.go internal/application/vehicle_sets.go internal/application/vehicle_sets_test.go
go test ./internal/infrastructure ./internal/application -run 'VehicleSetInventoryNumber|CreateVehicleSetAssigns|CreateVehicleSetRollsBackSetNumber'
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add backend/migrations/0060_vehicle_set_inventory_number.sql backend/internal/infrastructure/vehicle_set_inventory_number_migration_test.go backend/internal/application/vehicle_types.go backend/internal/application/vehicle_sets.go backend/internal/application/vehicle_sets_test.go
git commit -m "feat: assign persistent vehicle set numbers"
```

### Task 2: Expose canonical set read, update, and vehicle-list projection

**Files:**
- Modify: `backend/internal/application/vehicle_types.go`
- Modify: `backend/internal/application/vehicle_sets.go`
- Modify: `backend/internal/application/vehicles.go`
- Modify: `backend/internal/application/vehicle_repository.go`
- Modify: `backend/internal/application/vehicle_sets_test.go`
- Modify: `backend/internal/api/vehicle_handlers.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/routes_security_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: persistent `VehicleSet.InventoryNumber` from Task 1.
- Produces: `Vehicle.VehicleSet *VehicleSetSummary`, `GetSet(ctx, id)`, `UpdateSet(ctx, id, input, actorUserID)`, GET/PATCH `/api/v1/vehicle-sets/{id}`.

- [ ] **Step 1: Write failing service, route, authorization, and contract tests**

Define the expected nested projection explicitly in `vehicle_sets_test.go`:

```go
listed, err := service.List(ctx, "")
if err != nil { t.Fatal(err) }
if listed[0].VehicleSet == nil || listed[0].VehicleSet.InventoryNumber != "RK-SET-000001" {
	t.Fatalf("missing canonical set projection: %#v", listed[0].VehicleSet)
}
if listed[0].VehicleSet.MemberCount != 2 || listed[0].VehicleSet.Position != 1 {
	t.Fatalf("wrong membership projection: %#v", listed[0].VehicleSet)
}
```

Add tests that `GetSet` returns ordered members, `UpdateSet` changes shared fields on the canonical row and member compatibility snapshots, and a forced member-update failure rolls back both. Add route cases proving Viewer can GET, Viewer cannot PATCH, Editor can PATCH only with CSRF, and missing IDs return `vehicle_set_not_found`.

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd backend
go test ./internal/application ./internal/api -run 'VehicleSet|OpenAPI.*VehicleSet'
```

Expected: FAIL because the nested projection and GET/PATCH endpoints are absent.

- [ ] **Step 3: Add exact domain interfaces and service behavior**

Add the summary without removing temporary flat fields:

```go
type VehicleSetSummary struct {
	ID              string `json:"id"`
	InventoryNumber string `json:"inventoryNumber"`
	Name            string `json:"name"`
	Manufacturer    string `json:"manufacturer"`
	ArticleNumber   string `json:"articleNumber,omitempty"`
	Gauge           string `json:"gauge"`
	Epoch           string `json:"epoch,omitempty"`
	AcquisitionType string `json:"acquisitionType,omitempty"`
	PurchaseDate    string `json:"purchaseDate,omitempty"`
	PurchasePrice   string `json:"purchasePrice,omitempty"`
	Condition       string `json:"condition,omitempty"`
	MemberCount     int    `json:"memberCount"`
	Position        int    `json:"position"`
}
```

Add `VehicleSet *VehicleSetSummary \`json:"vehicleSet,omitempty"\`` to `Vehicle`. Extend list/get SELECTs with canonical `vehicle_sets` columns and populate both the nested object and the existing flat compatibility fields.

Implement:

```go
func (s *VehicleService) GetSet(ctx context.Context, id string) (*VehicleSet, error)
func (s *VehicleService) UpdateSet(ctx context.Context, id string, input VehicleSetInput, actorUserID string) (*VehicleSet, error)
```

`UpdateSet` cleans and validates input, begins one transaction, updates `vehicle_sets`, applies `applyVehicleSetFields` to every ordered member snapshot, writes `VehicleSetUpdated`, commits, then returns `GetSet`. Return `ErrVehicleSetNotFound` when no canonical row exists.

- [ ] **Step 4: Add handlers, routes, and OpenAPI**

Register:

```go
{http.MethodGet, "/api/v1/vehicle-sets/{id}", routeAccessViewer, (*App).getVehicleSet, nil},
{http.MethodPatch, "/api/v1/vehicle-sets/{id}", routeAccessEditor, (*App).updateVehicleSet, nil},
```

Document `VehicleSetSummary`, nested `vehicleSet`, `inventoryNumber`, GET response, PATCH input/response, `404 vehicle_set_not_found`, and the existing CSRF-protected write behavior in `openapi/railkeeper.yaml`.

- [ ] **Step 5: Format, run tests, and commit**

```powershell
cd backend
gofmt -w internal/application/vehicle_types.go internal/application/vehicle_sets.go internal/application/vehicles.go internal/application/vehicle_repository.go internal/application/vehicle_sets_test.go internal/api/vehicle_handlers.go internal/api/routes.go internal/api/routes_security_test.go internal/api/openapi_contract_test.go
go test ./internal/application ./internal/api
cd ..
git add backend/internal/application backend/internal/api openapi/railkeeper.yaml
git commit -m "feat: expose canonical vehicle set data"
```

Expected: all application and API tests PASS.

### Task 3: Preserve set inventory numbers in backup version 18

**Files:**
- Create: `backend/internal/application/backup_vehicle_sets.go`
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Interfaces:**
- Consumes: migration 0060 and category `Set` scheme.
- Produces: version-18 round trip and deterministic version-17 restore normalization.

- [ ] **Step 1: Write failing round-trip and legacy restore tests**

Add one test exporting a set and asserting `backup.Version == 18` plus the exact `inventory_number`. Add one version-17 document with two set rows that omit `inventory_number`; restore and assert deterministic `RK-SET-000001/000002` order and `next_number == 3`.

```go
if backup.Version != 18 {
	t.Fatalf("expected backup version 18, got %d", backup.Version)
}
if got := backup.Tables["vehicle_sets"][0]["inventory_number"]; got != "RK-SET-000001" {
	t.Fatalf("set inventory number missing from backup: %v", got)
}
```

- [ ] **Step 2: Run the focused tests and verify failure**

```powershell
cd backend
go test ./internal/application -run 'BackupVersionEighteen|LegacyVersionSeventeenVehicleSet'
```

Expected: FAIL with version 17 or missing `inventory_number` during restore.

- [ ] **Step 3: Implement legacy normalization**

Set `backupVersion = 18`. Before restoring `vehicle_sets`, call:

```go
func prepareBackupVehicleSetInventoryNumbers(
	ctx context.Context,
	tx *sql.Tx,
	doc *BackupDocument,
) error
```

For versions below 18, clone every set row, assign values using the active restored `Set` scheme in stable `created_at`, `id` order, update `doc.Tables["vehicle_sets"]`, and advance the scheme in the same restore transaction. Do not mutate the caller's original row maps. If the backup has no `Set` scheme, insert the default scheme before assigning numbers.

- [ ] **Step 4: Run backup and full backend tests**

```powershell
cd backend
gofmt -w internal/application/backup_vehicle_sets.go internal/application/backup.go internal/application/backup_test.go
go test ./internal/application -run 'BackupVersionEighteen|LegacyVersionSeventeenVehicleSet'
go test ./...
```

Expected: both commands PASS.

- [ ] **Step 5: Commit**

```powershell
git add backend/internal/application/backup_vehicle_sets.go backend/internal/application/backup.go backend/internal/application/backup_test.go
git commit -m "feat: back up vehicle set inventory numbers"
```

### Task 4: Add frontend set contracts, grouping, and the type column

**Files:**
- Modify: `frontend/src/shared/api.ts`
- Modify: `frontend/src/features/vehicles/vehicleSetGroups.ts`
- Modify: `frontend/src/features/vehicles/vehicleSetGroups.test.ts`
- Modify: `frontend/src/features/vehicles/vehicleTableColumns.ts`
- Modify: `frontend/src/features/vehicles/vehicleTableColumns.test.ts`
- Modify: `frontend/src/features/vehicles/useVehicleColumnPreferences.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: backend `VehicleSetSummary` and set GET/PATCH routes.
- Produces: typed `VehicleInventorySetGroup`, `type` column, `api.vehicleSet`, and `api.updateVehicleSet` for presentation tasks.

- [ ] **Step 1: Write failing type-column and grouping tests**

Build a fixture where two visible members carry the same canonical summary with `memberCount: 4`. Assert the group uses canonical set data, preserves explicit member positions, and reports `visibleMemberCount: 2`, `totalMemberCount: 4`. Assert old saved preferences remain unchanged while new defaults start with `type`.

```ts
expect(grouped[0]).toMatchObject({
  kind: "set",
  set: { id: "set-1", inventoryNumber: "RK-SET-000001", memberCount: 4 },
  visibleMemberCount: 2,
  totalMemberCount: 4
});
expect(defaultVehicleTableColumns).toEqual([
  "type", "image", "inventoryNumber", "manufacturer", "articleNumber", "name", "gauge", "epoch", "exhibition"
]);
```

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleSetGroups.test.ts src/features/vehicles/vehicleTableColumns.test.ts src/features/vehicles/useVehicleColumnPreferences.test.tsx
```

Expected: FAIL because `Vehicle.vehicleSet`, `type`, and visible/total counts do not exist.

- [ ] **Step 3: Implement exact frontend contracts**

Add:

```ts
export type VehicleSetSummary = {
  id: string;
  inventoryNumber: string;
  name: string;
  manufacturer: string;
  articleNumber?: string;
  gauge: string;
  epoch?: string;
  acquisitionType?: string;
  purchaseDate?: string;
  purchasePrice?: string;
  condition?: string;
  memberCount: number;
  position: number;
};
```

Add `vehicleSet?: VehicleSetSummary` to `Vehicle`, `inventoryNumber` to `VehicleSet`, and API methods:

```ts
vehicleSet: (id: string) => request<VehicleSet>(`/vehicle-sets/${encodeURIComponent(id)}`),
updateVehicleSet: (id: string, input: VehicleSetInput) =>
  request<VehicleSet>(`/vehicle-sets/${encodeURIComponent(id)}`, {
    method: "PATCH",
    body: JSON.stringify(input)
  }),
```

Use this group union:

```ts
export type VehicleInventoryGroup =
  | { kind: "single"; vehicle: Vehicle }
  | {
      kind: "set";
      id: string;
      set: VehicleSetSummary;
      members: Vehicle[];
      visibleMemberCount: number;
      totalMemberCount: number;
    };

export type VehicleInventorySetGroup = Extract<VehicleInventoryGroup, { kind: "set" }>;
```

Add `type` to the column keys, identity group, label translations, and make it presentation-only by excluding it from `VehicleSortableColumn` and `isVehicleDataColumn`.

- [ ] **Step 4: Run tests and build**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleSetGroups.test.ts src/features/vehicles/vehicleTableColumns.test.ts src/features/vehicles/useVehicleColumnPreferences.test.tsx
npm.cmd run build
```

Expected: tests PASS and TypeScript build succeeds.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/shared/api.ts frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/features/vehicles/vehicleSetGroups.ts frontend/src/features/vehicles/vehicleSetGroups.test.ts frontend/src/features/vehicles/vehicleTableColumns.ts frontend/src/features/vehicles/vehicleTableColumns.test.ts frontend/src/features/vehicles/useVehicleColumnPreferences.test.tsx
git commit -m "feat: model vehicle set inventory groups"
```

### Task 5: Implement real set actions and dialogs

**Files:**
- Create: `frontend/src/features/vehicles/VehicleSetSummaryDialog.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetSummaryDialog.test.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetEditorDialog.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetEditorDialog.test.tsx`
- Create: `frontend/src/features/vehicles/vehicleSetDuplicate.ts`
- Create: `frontend/src/features/vehicles/vehicleSetDuplicate.test.ts`
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: `api.vehicleSet(id)`, `api.updateVehicleSet(id, input)` from Task 4.
- Produces: `VehicleCreatePrefill`, `vehicleSetDuplicatePrefill`, and `onOpenSet`, `onEditSet`,
  `onDuplicateSet` callbacks passed to desktop and mobile set rows.

- [ ] **Step 1: Write failing dialog tests**

Test loading, error, ordered members, Viewer read-only behavior, Editor save, and duplicate
transformation. The duplicate helper must clear every inventory number:

```ts
const prefill = vehicleSetDuplicatePrefill(setFixture);
expect(prefill.kind).toBe("set");
expect(prefill.shared.inventoryNumber).toBe("");
expect(prefill.members.every((member) => member.inventoryNumber === "")).toBe(true);
```

- [ ] **Step 2: Run tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetSummaryDialog.test.tsx src/features/vehicles/VehicleSetEditorDialog.test.tsx src/features/vehicles/vehicleSetDuplicate.test.ts
```

Expected: FAIL because the components do not exist.

- [ ] **Step 3: Implement summary and editor dialogs**

`VehicleSetSummaryDialog` loads by ID on open, renders shared identity/acquisition fields and ordered member links, and exposes View/Edit/Duplicate based on callbacks. `VehicleSetEditorDialog` edits only `VehicleSetInput`; it must never expose set/member inventory number fields as writable.

Use explicit props:

```ts
type VehicleSetDialogProps = {
  setId: string;
  canEdit: boolean;
  onClose: () => void;
  onUpdated: (set: VehicleSet) => void;
  onDuplicate: (prefill: VehicleCreatePrefill) => void;
};
```

Define and implement:

```ts
export type VehicleCreatePrefill = {
  kind: "set";
  shared: CreateVehicleRequest;
  members: CreateVehicleRequest[];
};

export function vehicleSetDuplicatePrefill(set: VehicleSet): VehicleCreatePrefill
```

Build `shared` from the canonical set fields and build members with `vehicleToForm`. Clear the set
number, every member inventory number, vehicle IDs, attachments, maintenance records, CV files, and
local stored-image identities. Keep copyable catalogue fields and physical member details. Reuse
only external HTTP(S) image suggestions that will pass through the existing safe import pipeline.
In `VehiclesView`,
keep one `selectedVehicleSetID` and one mode (`"view" | "edit" | null`), reload vehicles after
update, and pass the prefill into the create wizard. Do not add a delete action.

- [ ] **Step 4: Run dialog tests and the frontend build**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetSummaryDialog.test.tsx src/features/vehicles/VehicleSetEditorDialog.test.tsx src/features/vehicles/vehicleSetDuplicate.test.ts src/features/vehicles/VehiclesView.test.tsx
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/features/vehicles/VehicleSetSummaryDialog.tsx frontend/src/features/vehicles/VehicleSetSummaryDialog.test.tsx frontend/src/features/vehicles/VehicleSetEditorDialog.tsx frontend/src/features/vehicles/VehicleSetEditorDialog.test.tsx frontend/src/features/vehicles/vehicleSetDuplicate.ts frontend/src/features/vehicles/vehicleSetDuplicate.test.ts frontend/src/features/vehicles/VehiclesView.tsx frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: add vehicle set view and edit actions"
```

### Task 6: Replace the desktop set separator with an aligned inventory row

**Files:**
- Create: `frontend/src/features/vehicles/VehicleSetInventoryRow.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetInventoryRow.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryTable.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryTable.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryPanel.tsx`
- Modify: `frontend/src/features/vehicles/vehicleInventoryRenderers.tsx`
- Modify: `frontend/src/styles/vehicle-inventory.css`

**Interfaces:**
- Consumes: `VehicleInventorySetGroup` and set action callbacks from Tasks 4 and 5.
- Produces: aligned desktop hierarchy with tri-state set selection and visible tree connectors.

- [ ] **Step 1: Write failing row and integration tests**

Assert the set row has one cell per configured column rather than `colSpan`, shows `RK-SET-000001`, uses a mixed checkbox when one of two visible members is selected, and clicking it sends both visible member IDs.

```ts
expect(screen.getByRole("checkbox", { name: /RK-SET-000001/ })).toHaveProperty("indeterminate", true);
await user.click(screen.getByRole("checkbox", { name: /RK-SET-000001/ }));
expect(onToggleSetSelection).toHaveBeenCalledWith(["member-1", "member-2"]);
expect(container.querySelector(".vehicle-set-inventory-row td[colspan]")).toBeNull();
```

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetInventoryRow.test.tsx src/features/vehicles/VehicleInventoryTable.test.tsx
```

Expected: FAIL because the row component and set callbacks do not exist.

- [ ] **Step 3: Implement the row and selection semantics**

Use props:

```ts
type VehicleSetInventoryRowProps = {
  group: VehicleInventorySetGroup;
  columns: readonly VehicleTableColumn[];
  collapsed: boolean;
  selectedVehicleIDs: ReadonlySet<string>;
  onToggleCollapsed: () => void;
  onToggleSelection: (vehicleIDs: string[]) => void;
  onOpen: (setId: string) => void;
  onEdit?: (setId: string) => void;
  onDuplicate?: (setId: string) => void;
};
```

Set `checkboxRef.current.indeterminate = selectedCount > 0 && selectedCount < group.members.length`. Render canonical fields for `inventoryNumber`, manufacturer, article number, name, gauge, epoch, acquisition/condition columns; render blank member-only columns. The `type` cell contains a non-color-only set badge, the acquisition type when present, and the expand button. Member type cells say `Vehicle`/`Fahrzeug`; their inventory cells show the vehicle number as compact secondary text. Put acquisition date/price in a secondary row inside the name cell without changing table column alignment.

Render member rows with `.vehicle-set-child-row` plus a tree connector pseudo-element. The existing View/Edit/Delete/quick actions remain unchanged for physical vehicles.

- [ ] **Step 4: Run tests and build**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetInventoryRow.test.tsx src/features/vehicles/VehicleInventoryTable.test.tsx src/features/vehicles/vehicleInventoryRenderers.test.ts
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/features/vehicles/VehicleSetInventoryRow.tsx frontend/src/features/vehicles/VehicleSetInventoryRow.test.tsx frontend/src/features/vehicles/VehicleInventoryTable.tsx frontend/src/features/vehicles/VehicleInventoryTable.test.tsx frontend/src/features/vehicles/VehicleInventoryPanel.tsx frontend/src/features/vehicles/vehicleInventoryRenderers.tsx frontend/src/styles/vehicle-inventory.css
git commit -m "feat: align vehicle sets in desktop inventory"
```

### Task 7: Add the dedicated mobile set hierarchy

**Files:**
- Create: `frontend/src/features/vehicles/VehicleSetInventoryMobileCard.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetInventoryMobileCard.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx`
- Modify: `frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts`
- Modify: `frontend/src/styles/vehicle-inventory.css`

**Interfaces:**
- Consumes: canonical groups and set action callbacks from Tasks 4 to 6.
- Produces: mobile set card and indented member-card tree with no horizontal table dependency.

- [ ] **Step 1: Write failing mobile tests**

At a semantic level, assert the card exposes set number/name/member count, expand state, View/Edit/Duplicate actions, and member cards only when expanded. In the CSS regression test, require one-column mobile layout and 44-pixel action targets.

```ts
expect(screen.getByText("RK-SET-000001")).toBeVisible();
expect(screen.getByText(/2 von 4/)).toBeVisible();
expect(screen.queryByText("RK-WAG-000001")).toBeNull();
await user.click(screen.getByRole("button", { name: /Set aufklappen/ }));
expect(screen.getByText("RK-WAG-000001")).toBeVisible();
```

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetInventoryMobileCard.test.tsx src/features/vehicles/VehicleInventoryMobileList.test.tsx src/features/vehicles/vehicleInventoryResponsive.test.ts
```

Expected: FAIL because the dedicated card does not exist.

- [ ] **Step 3: Implement the mobile component**

Use the same action contract as the desktop row. Render canonical set number/name, manufacturer plus article number, visible/total count, acquisition summary, and a compact action bar. Keep physical member rendering in `VehicleInventoryMobileCard`, nested below a visible connector and indentation. Do not render the desktop table or force horizontal scrolling below the mobile breakpoint.

- [ ] **Step 4: Run tests and build**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetInventoryMobileCard.test.tsx src/features/vehicles/VehicleInventoryMobileList.test.tsx src/features/vehicles/vehicleInventoryResponsive.test.ts
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/features/vehicles/VehicleSetInventoryMobileCard.tsx frontend/src/features/vehicles/VehicleSetInventoryMobileCard.test.tsx frontend/src/features/vehicles/VehicleInventoryMobileList.tsx frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts frontend/src/styles/vehicle-inventory.css
git commit -m "feat: add mobile vehicle set hierarchy"
```

### Task 8: Introduce the wizard reducer and versioned local drafts

**Files:**
- Create: `frontend/src/features/vehicles/vehicleCreateWizardState.ts`
- Create: `frontend/src/features/vehicles/vehicleCreateWizardState.test.ts`
- Modify: `frontend/src/features/vehicles/VehicleCreateWizard.tsx`

**Interfaces:**
- Consumes: `CreateVehicleRequest`, `VehicleSetInput`, `PendingArticleImage`, and
  `VehicleCreatePrefill` from Task 5.
- Produces: `VehicleCreateWizardState`, `VehicleCreateWizardAction`, `vehicleCreateWizardReducer`, `loadVehicleCreateDraft`, `saveVehicleCreateDraft`, and `clearVehicleCreateDraft`.

- [ ] **Step 1: Write failing reducer and draft tests**

Cover step transitions, explicit article substates, member preservation, invalid draft rejection, and successful version-1 restore:

```ts
const populated = stateWithMembers([
  emptyMember(),
  { ...emptyMember(), name: "Speisewagen" },
  emptyMember()
]);
const reduced = vehicleCreateWizardReducer(populated, { type: "set-member-count", count: 2 });
expect(reduced.pendingMemberReduction).toEqual({ requestedCount: 2, populatedIndexes: [1] });

localStorage.setItem(vehicleCreateDraftKey, JSON.stringify({ version: 99, savedAt: "now", state: {} }));
expect(loadVehicleCreateDraft()).toEqual({ kind: "invalid" });
```

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleCreateWizardState.test.ts
```

Expected: FAIL because the state module does not exist.

- [ ] **Step 3: Implement the explicit state model**

Use these discriminants:

```ts
export type VehicleCreationKind = "single" | "set";
export type VehicleCreateStep = "basics" | "article" | "details";
export type VehicleCreateArticleStage = "input" | "results" | "review";

export type VehicleSetMemberDraft = {
  form: CreateVehicleRequest;
  touched: boolean;
};

export type VehicleCreateWizardState = {
  kind: VehicleCreationKind;
  step: VehicleCreateStep;
  articleStage: VehicleCreateArticleStage;
  shared: CreateVehicleRequest;
  members: VehicleSetMemberDraft[];
  selectedResultIndex: number | null;
  activeDetailsTab: "set" | `member:${number}`;
  pendingMemberReduction: null | { requestedCount: number; populatedIndexes: number[] };
};
```

Store `{ version: 1, savedAt: ISO string, state }` under `railkeeper.vehicleCreateDraft.v1`. Validate the full discriminated shape before returning it. Storage exceptions return `{ kind: "error" }` and never block creation.
Initialize the reducer either from an empty form or a `VehicleCreatePrefill`; remove the old
`VehicleSetMemberDraft` declaration from `VehicleCreateWizard.tsx` to avoid a circular import.

- [ ] **Step 4: Run tests and commit**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleCreateWizardState.test.ts
cd ..
git add frontend/src/features/vehicles/vehicleCreateWizardState.ts frontend/src/features/vehicles/vehicleCreateWizardState.test.ts frontend/src/features/vehicles/VehicleCreateWizard.tsx
git commit -m "refactor: model vehicle creation as a wizard state"
```

Expected: PASS.

### Task 9: Build the responsive wizard shell and basic-data step

**Files:**
- Create: `frontend/src/features/vehicles/VehicleCreateWizardShell.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateWizardShell.test.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateStepBasics.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateStepBasics.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleCreateWizard.tsx`
- Modify: `frontend/src/features/vehicles/VehicleEditorDialog.tsx`
- Modify: `frontend/src/styles/vehicle-create-wizard.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: reducer from Task 8 and `api.inventoryNumberSchemes()`.
- Produces: approved vertical desktop rail, compact mobile progress, Step 1 fields, provisional set-number preview, and safe member-count confirmation.

- [ ] **Step 1: Write failing shell and basics tests**

Assert ordered-list current-step semantics, destination-specific buttons, radio semantics, exact basics fields, disabled set option for ECoS, provisional `RK-SET-000001`, minimum two members, and confirmation before removing populated members.

```ts
expect(screen.getByRole("radio", { name: /Set/ })).toHaveAttribute("aria-checked", "true");
expect(screen.getByText("RK-SET-000001")).toHaveTextContent(/vorläufig/i);
expect(screen.getByRole("button", { name: /Weiter zu Artikeldaten/ })).toBeEnabled();
```

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleCreateWizardShell.test.tsx src/features/vehicles/VehicleCreateStepBasics.test.tsx src/features/vehicles/VehicleCreateWizard.test.tsx
```

Expected: FAIL because the approved shell and step components do not exist.

- [ ] **Step 3: Implement shell and Step 1**

`VehicleCreateWizardShell` owns one labelled dialog surface, fixed header/footer, scrollable body, desktop rail summaries, and mobile top progress. `VehicleCreateStepBasics` renders type cards followed by manufacturer, designation, article number, gauge, category, and set count. Load the active `Set` scheme and format the preview with:

```ts
export function inventoryNumberPreview(scheme: InventoryNumberScheme) {
  return `${scheme.prefix}-${String(scheme.nextNumber).padStart(scheme.padding, "0")}`;
}
```

Label it provisional because another transaction can reserve it first. A user-confirmed reduction dispatches `confirm-member-reduction`; cancel leaves all member drafts untouched.
If no active `Set` scheme exists, render the established inventory-scheme error, omit the numeric
preview, and disable progression for set creation while leaving single-vehicle creation available.

- [ ] **Step 4: Run tests and build**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleCreateWizardShell.test.tsx src/features/vehicles/VehicleCreateStepBasics.test.tsx src/features/vehicles/VehicleCreateWizard.test.tsx
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/features/vehicles/VehicleCreateWizardShell.tsx frontend/src/features/vehicles/VehicleCreateWizardShell.test.tsx frontend/src/features/vehicles/VehicleCreateStepBasics.tsx frontend/src/features/vehicles/VehicleCreateStepBasics.test.tsx frontend/src/features/vehicles/VehicleCreateWizard.tsx frontend/src/features/vehicles/VehicleEditorDialog.tsx frontend/src/styles/vehicle-create-wizard.css frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: rebuild vehicle wizard shell and basics"
```

### Task 10: Embed article search results and grouped review in Step 2

**Files:**
- Modify: `frontend/src/features/vehicles/useArticleSearchController.ts`
- Modify: `frontend/src/features/vehicles/useArticleSearchController.test.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateStepArticle.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateArticleResults.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateArticleReview.tsx`
- Create: `frontend/src/features/vehicles/VehicleCreateStepArticle.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleCreateWizard.tsx`
- Modify: `frontend/src/features/vehicles/VehicleEditorDialog.tsx`
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Modify: `frontend/src/styles/vehicle-create-wizard.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: existing sanitized `ArticleSearchResponse`, selection-key helpers, field groups, and wizard article stage.
- Produces: exported `ArticleSearchController`, embedded input/results/review, while standalone `ArticleSearchDialog` remains for non-create contexts.

- [ ] **Step 1: Write failing controller and embedded-flow tests**

Export the controller return type and test that `run` no longer requires opening a portal. In the wizard test, drive input to results, select a result, review grouped conflicts, toggle a field and image, apply, and continue manually after an error.

```ts
await user.click(screen.getByRole("button", { name: /Artikeldaten suchen/ }));
expect(await screen.findByRole("heading", { name: /Suchergebnisse/ })).toBeVisible();
await user.click(screen.getByRole("button", { name: /Roco 6280002 auswählen/ }));
expect(screen.getByRole("heading", { name: /Datenübernahme prüfen/ })).toBeVisible();
expect(screen.queryByRole("dialog", { name: /Artikeldaten-Websuche/ })).toBeNull();
```

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/useArticleSearchController.test.tsx src/features/vehicles/VehicleCreateStepArticle.test.tsx
```

Expected: FAIL because create mode still launches the standalone dialog.

- [ ] **Step 3: Separate controller behavior from presentation**

Export:

```ts
export type ArticleSearchController = ReturnType<typeof useArticleSearchController>;
```

Keep `state.response`, `state.loading`, `state.error`, selected fields/images, barcode state, and all commands unchanged. In `VehiclesView`, pass the controller into `VehicleEditorDialog`; render `ArticleSearchDialog` and `BarcodeSearchDialog` only when `mode !== "create"`.

- [ ] **Step 4: Implement the three embedded substates**

`VehicleCreateStepArticle` renders barcode/manufacturer-article/manual input. Results render ranked source cards with image, score, detail trace, summary, and source link. Selecting a result dispatches `select-article-result` and enters review. Review renders the exact groups:

```ts
const createReviewGroups = [
  "identification",
  "railway",
  "technical",
  "description",
  "media"
] as const;
```

Desktop uses compact tables; mobile uses stacked current/found cards under the same semantic checkbox labels. Every group shows field count and conflict count. `Apply selected fields` calls the existing controller command and advances; `Continue without import` preserves form data and advances.

- [ ] **Step 5: Run tests, build, and commit**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/useArticleSearchController.test.tsx src/features/vehicles/VehicleCreateStepArticle.test.tsx src/features/vehicles/VehicleCreateWizard.test.tsx
npm.cmd run build
cd ..
git add frontend/src/features/vehicles/useArticleSearchController.ts frontend/src/features/vehicles/useArticleSearchController.test.tsx frontend/src/features/vehicles/VehicleCreateStepArticle.tsx frontend/src/features/vehicles/VehicleCreateArticleResults.tsx frontend/src/features/vehicles/VehicleCreateArticleReview.tsx frontend/src/features/vehicles/VehicleCreateStepArticle.test.tsx frontend/src/features/vehicles/VehicleCreateWizard.tsx frontend/src/features/vehicles/VehicleEditorDialog.tsx frontend/src/features/vehicles/VehiclesView.tsx frontend/src/styles/vehicle-create-wizard.css frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: embed article review in vehicle creation"
```

Expected: tests PASS and build succeeds.

### Task 11: Replace the legacy final form with set/member detail tabs

**Files:**
- Create: `frontend/src/features/vehicles/VehicleCreateStepDetails.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetDetailsTabs.tsx`
- Create: `frontend/src/features/vehicles/VehicleSetDetailsTabs.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleCreateWizard.tsx`
- Modify: `frontend/src/features/vehicles/VehicleCreateWizard.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleModelTab.tsx`
- Modify: `frontend/src/features/vehicles/VehicleFormFields.tsx`
- Modify: `frontend/src/styles/vehicle-create-wizard.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: reducer state, shared set form, ordered member drafts, existing vehicle field controls.
- Produces: `Set & acquisition` plus member tabs/selector, grouped accordions, valid single/set payloads, and draft save/resume/discard UI.

- [ ] **Step 1: Write failing tab, inheritance, payload, and draft UI tests**

Assert keyboard tab semantics, add-member action, inherited shared fields are not editable in member tabs, member-specific changes stay isolated, images attach to the intended member, and set/member numbers are read-only previews.

```ts
expect(screen.getByRole("tab", { name: /Set & Anschaffung/ })).toHaveAttribute("aria-selected", "true");
await user.click(screen.getByRole("tab", { name: /Wagen 2/ }));
expect(screen.getByLabelText(/Fahrzeugnummer/)).toHaveValue("50 50 23-11 011-5");
expect(screen.queryByRole("textbox", { name: /Hersteller/ })).toBeNull();
```

Test that Save draft writes version 1, reopening offers Resume/Discard, successful creation clears the key, and storage failure displays a non-blocking message.

- [ ] **Step 2: Run focused tests and verify failure**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetDetailsTabs.test.tsx src/features/vehicles/VehicleCreateWizard.test.tsx
```

Expected: FAIL because Step 3 still embeds the entire legacy `VehicleModelTab`.

- [ ] **Step 3: Implement focused details components**

`VehicleCreateStepDetails` owns grouped accordions. For sets, `VehicleSetDetailsTabs` exposes `Set & acquisition`, one tab per ordered member, and `Add vehicle`; mobile replaces the horizontal row with an app-owned select. Shared fields live only in the set panel. Member panels render designation, vehicle number, technical/equipment/coupling/decoder/notes fields by composing focused controls extracted from `VehicleModelTab`, not by duplicating field definitions.
Each member panel starts with a read-only inherited-context summary for manufacturer, article number,
gauge, epoch, category, and subtype. Accordion headers show their actual field counts. Decoder groups
show existing CV values and file names as read-only previews; creation does not invent server IDs for
CV records.

Use accessible tab IDs:

```ts
const tabId = (value: VehicleCreateWizardState["activeDetailsTab"]) => `vehicle-create-tab-${value}`;
const panelId = (value: VehicleCreateWizardState["activeDetailsTab"]) => `vehicle-create-panel-${value}`;
```

ArrowLeft/ArrowRight move focus and selection among enabled desktop tabs. Every panel uses `role="tabpanel"` and `aria-labelledby`.

- [ ] **Step 4: Wire payloads and draft lifecycle**

Map shared fields once into `CreateVehicleSetRequest.set`; map each member-specific draft into `members` while assigning imported images to the selected owner. Keep all `inventoryNumber` values empty so the backend reserves them. Single creation continues using `CreateVehicleRequest`.

On close with meaningful data, offer Save draft/Discard/Continue editing. An invalid or incompatible
stored draft produces a localized explanation and offers a clean start. On successful
`createVehicle` or `createVehicleSet`, call `clearVehicleCreateDraft()` before closing or switching
to the created record.

- [ ] **Step 5: Run tests, build, and commit**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleSetDetailsTabs.test.tsx src/features/vehicles/VehicleCreateWizard.test.tsx src/features/vehicles/VehicleModelTab.test.tsx src/features/vehicles/VehiclesView.test.tsx
npm.cmd run build
cd ..
git add frontend/src/features/vehicles/VehicleCreateStepDetails.tsx frontend/src/features/vehicles/VehicleSetDetailsTabs.tsx frontend/src/features/vehicles/VehicleSetDetailsTabs.test.tsx frontend/src/features/vehicles/VehicleCreateWizard.tsx frontend/src/features/vehicles/VehicleCreateWizard.test.tsx frontend/src/features/vehicles/VehicleModelTab.tsx frontend/src/features/vehicles/VehicleFormFields.tsx frontend/src/styles/vehicle-create-wizard.css frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: add set and member detail steps"
```

Expected: PASS.

### Task 12: Verify filtering, sorting, selection, accessibility, and long text

**Files:**
- Modify: `frontend/src/features/vehicles/useVehicleInventoryController.ts`
- Modify: `frontend/src/features/vehicles/useVehicleInventoryController.test.tsx`
- Modify: `frontend/src/features/vehicles/vehicleSetGroups.test.ts`
- Modify: `frontend/src/features/vehicles/VehicleInventoryTable.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleCreateWizard.test.tsx`
- Modify: `frontend/src/styles/vehicle-inventory.css`
- Modify: `frontend/src/styles/vehicle-create-wizard.css`

**Interfaces:**
- Consumes: completed inventory and wizard components.
- Produces: regression coverage for the behavioral and accessibility acceptance contract.

- [ ] **Step 1: Add failing integration tests for edge behavior**

Cover these exact cases:

```ts
const setSummary = (id: string, position: number, memberCount: number): VehicleSetSummary => ({
  id,
  inventoryNumber: `RK-SET-${id}`,
  name: id,
  manufacturer: "Roco",
  gauge: "H0",
  memberCount,
  position
});

it("shows a canonical set once when only one member matches a filter", () => {
  const groups = groupVehicleInventory([
    vehicle("member-3", {
      vehicleSet: {
        id: "set-1",
        inventoryNumber: "RK-SET-000001",
        name: "TEE Roland",
        manufacturer: "Märklin",
        gauge: "H0",
        memberCount: 4,
        position: 3
      }
    })
  ]);
  expect(groups).toHaveLength(1);
  expect(groups[0]).toMatchObject({
    kind: "set",
    visibleMemberCount: 1,
    totalMemberCount: 4,
    members: [{ id: "member-3" }]
  });
});

it("keeps explicit member positions after sorted input determines group order", () => {
  const groups = groupVehicleInventory([
    vehicle("member-2", { vehicleSet: setSummary("set-b", 2, 2) }),
    vehicle("member-a", { vehicleSet: setSummary("set-a", 1, 1) }),
    vehicle("member-1", { vehicleSet: setSummary("set-b", 1, 2) })
  ]);
  expect(groups.map((group) => group.kind === "set" ? group.id : group.vehicle.id))
    .toEqual(["set-b", "set-a"]);
  expect(groups[0].kind === "set" ? groups[0].members.map((member) => member.id) : [])
    .toEqual(["member-1", "member-2"]);
});
```

Add component assertions that set selection sends only the rendered member IDs, invalid submission
focuses the first labelled field in the named step, mixed selection exposes the native
`indeterminate` property, and ArrowLeft/ArrowRight changes the active details tab. Use complete local
fixtures in the respective test files and do not call the network.

- [ ] **Step 2: Run focused tests and confirm any failures**

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/useVehicleInventoryController.test.tsx src/features/vehicles/vehicleSetGroups.test.ts src/features/vehicles/VehicleInventoryTable.test.tsx src/features/vehicles/VehicleInventoryMobileList.test.tsx src/features/vehicles/VehicleCreateWizard.test.tsx
```

Expected: new assertions may FAIL until controller selection and focus behavior are aligned.

- [ ] **Step 3: Make the minimal controller and CSS corrections**

Keep search/filter logic vehicle-based. Derive set presence only after `sortedVehicles` is finalized. Set selection receives the group's visible member IDs, never `memberCount` or hidden IDs. Add `overflow-wrap: anywhere`, `min-width: 0`, visible `:focus-visible`, non-color tree borders/icons, and 44-pixel mobile controls using existing tokens.

- [ ] **Step 4: Run the complete frontend suite and build**

```powershell
cd frontend
npm.cmd run test:run
npm.cmd run build
```

Expected: all frontend tests PASS and production build succeeds with only known Vite chunk-size warnings.

- [ ] **Step 5: Commit**

```powershell
git add frontend/src/features/vehicles frontend/src/styles/vehicle-inventory.css frontend/src/styles/vehicle-create-wizard.css
git commit -m "test: cover vehicle set UI behavior"
```

### Task 13: Update stable documentation and run repository verification

**Files:**
- Modify: `docs/site/guide/vehicles/index.md`
- Modify: `docs/site/de/guide/vehicles/index.md`
- Modify: `docs/coverage.json` only when validation reports changed source coverage.
- Modify: any touched Go/TS/CSS files only for failures found by the checks.

**Interfaces:**
- Consumes: completed feature behavior.
- Produces: bilingual operator documentation and a fully verified candidate for browser review.

- [ ] **Step 1: Document the real workflow in English and German**

Document persisted set numbers, parent/member hierarchy, set View/Edit/Duplicate actions, the three wizard steps, embedded article review, draft behavior, and mobile presentation. State explicitly that complete-set deletion is not available and article data remains a user-reviewed suggestion.

- [ ] **Step 2: Run format and backend tests**

```powershell
cd backend
gofmt -w internal/application/vehicle_types.go internal/application/vehicle_sets.go internal/application/vehicles.go internal/application/vehicle_repository.go internal/application/backup_vehicle_sets.go internal/application/backup.go internal/api/vehicle_handlers.go internal/api/routes.go internal/api/routes_security_test.go internal/api/openapi_contract_test.go internal/infrastructure/vehicle_set_inventory_number_migration_test.go cmd/railkeeper/main.go
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run frontend and documentation checks**

```powershell
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
cd ..\docs
npm.cmd test
npm.cmd run check
```

Expected: 0 failing tests, successful frontend build, successful VitePress build. If documentation validation names source coverage drift, update only the exact matching entries in `docs/coverage.json` and rerun `npm.cmd run check`.

- [ ] **Step 4: Run Windows package validation**

```powershell
cd ..
powershell.exe -ExecutionPolicy Bypass -File .\tools\build_windows_standalone.ps1
```

Expected: the package validation succeeds and no generated package is staged.

- [ ] **Step 5: Review and commit documentation**

```powershell
git status --short
git diff --check
git add docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md docs/coverage.json
git commit -m "docs: explain vehicle set inventory workflow"
```

Before committing, omit `docs/coverage.json` from `git add` when it did not change.

### Task 14: Perform fixed-viewport visual acceptance and user checkpoint

**Files:**
- Modify: affected frontend TSX/CSS/i18n files only when a visible mismatch is found.
- Do not add: screenshots, browser traces, `frontend/dist`, or local test data.

**Interfaces:**
- Consumes: verified local build and approved reference images.
- Produces: browser-checked candidate explicitly approved by the user before merge/release.

- [ ] **Step 1: Start the candidate with fresh static assets**

```powershell
cd frontend
npm.cmd run build
$repo='C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\vehicle-ui-correction'
$env:RAILKEEPER_ADDR=':8092'
$env:RAILKEEPER_DATA_DIR="$repo\.cache\vehicle-ui-data"
$env:RAILKEEPER_MIGRATIONS_DIR="$repo\backend\migrations"
$env:RAILKEEPER_SEEDS_DIR="$repo\backend\seeds"
$env:RAILKEEPER_STATIC_DIR="$repo\frontend\dist"
$env:GOCACHE="$repo\.cache\go-build"
Start-Process -FilePath 'go' -ArgumentList 'run','./cmd/railkeeper' -WorkingDirectory "$repo\backend" -WindowStyle Hidden -PassThru
```

Expected: `/health` returns success and `/vehicles` serves the new build.

- [ ] **Step 2: Capture the required states at exact viewports**

Use the in-app browser at 1365 by 768 and 390 by 844. Inspect inventory desktop/mobile, wizard Step 1, article results, article review, Step 3 set/member, empty/loading/search-error/validation-error, long German text, light theme, and dark theme. Compare each state beside the approved reference, not from memory.

- [ ] **Step 3: Correct structural mismatches and repeat**

Fix only evidence-backed mismatches in hierarchy, alignment, density, wrapping, scrolling, focus, or mobile composition. After every correction run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles
npm.cmd run build
```

Expected: PASS, then recapture the corrected state at the same viewport.

- [ ] **Step 4: Present the running checkpoint to the user**

Tell the user exactly which desktop/mobile states are ready at `http://127.0.0.1:8092/vehicles` and pause for visual approval. Do not merge, tag, or release before explicit approval.

- [ ] **Step 5: Commit final visual corrections after approval**

```powershell
git add frontend/src/features/vehicles frontend/src/styles/vehicle-inventory.css frontend/src/styles/vehicle-create-wizard.css frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "fix: align vehicle UI with approved layouts"
```

If the visual pass required no code changes, do not create an empty commit.

### Task 15: Prepare v0.1.20, publish the PR, merge green checks, and release

**Files:**
- Modify: `backend/cmd/railkeeper/main.go`
- Create: `docs/releases/v0.1.20.md`
- Modify: release index/coverage files only when the existing documentation validator requires them.

**Interfaces:**
- Consumes: explicit visual approval and all green local checks.
- Produces: PR for issue #102, merged `main`, tag `v0.1.20`, and GitHub release/package workflow.

- [ ] **Step 1: Bump the application version and add bilingual release notes**

Change only the production constant in `backend/cmd/railkeeper/main.go`:

```go
version = "0.1.20"
```

Create `docs/releases/v0.1.20.md` describing set inventory numbers, corrected desktop/mobile inventory hierarchy, embedded three-step creation flow, article-data review, draft handling, shared set editing, compatibility, and migration/backup notes in German and English.

- [ ] **Step 2: Run the final clean verification**

```powershell
cd backend
go test ./...
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
cd ..\docs
npm.cmd run check
cd ..
powershell.exe -ExecutionPolicy Bypass -File .\tools\build_windows_standalone.ps1
git diff --check
git status --short
```

Expected: every command PASS; the v0.1.20 Windows package builds; status contains only intentional
source/docs changes and no generated output.

- [ ] **Step 3: Commit and push the branch**

```powershell
git add backend/cmd/railkeeper/main.go docs/releases/v0.1.20.md
git commit -m "chore: prepare v0.1.20"
git push -u origin dev/vehicle-ui-correction
```

- [ ] **Step 4: Create and monitor the pull request**

```powershell
gh pr create --base main --head dev/vehicle-ui-correction --title "Correct vehicle set inventory and creation UI" --body "Closes #102. Adds persistent set inventory numbers, canonical set editing, corrected desktop/mobile hierarchy, and the approved three-step creation flow with embedded article review."
gh pr checks --watch
```

Expected: every required check is green. If a check fails, inspect the failing job, fix it on the branch, rerun the relevant local command, commit, push, and wait again.

- [ ] **Step 5: Merge and release only while green**

```powershell
gh pr merge --squash --delete-branch
git fetch origin main --tags
git tag -a v0.1.20 origin/main -m "RailKeeper v0.1.20"
git push origin v0.1.20
gh release create v0.1.20 --title "RailKeeper v0.1.20" --notes-file docs/releases/v0.1.20.md
```

Expected: PR merged into `main`, tag and release exist, and the tagged Windows and Docker workflows
complete successfully. Monitor both without manual identifiers:

```powershell
$releaseRuns = gh run list --branch v0.1.20 --limit 10 --json databaseId,workflowName --jq '.[] | select(.workflowName == "Windows Standalone Package" or .workflowName == "Docker Image") | .databaseId'
if (@($releaseRuns).Count -ne 2) { throw 'Expected v0.1.20 Windows and Docker release runs.' }
foreach ($releaseRun in $releaseRuns) { gh run watch $releaseRun --exit-status }
```

Do not put the PC into standby.
