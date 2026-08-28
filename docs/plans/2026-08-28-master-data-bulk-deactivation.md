# Master Data Bulk Deactivation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add safe, atomic bulk deactivation to every editable master-data table.

**Architecture:** A new collection-level PATCH endpoint delegates to an atomic application-service
operation. A focused React selection hook and toolbar provide the same accessible interaction in
general and article master-data tables without moving unrelated Settings code.

**Tech Stack:** Go, SQLite, `net/http`, OpenAPI YAML, React, TypeScript, Vitest, Testing Library, CSS.

## Global Constraints

- Only active entries with `capabilities.canDeactivate` may be selected.
- One request contains 1 to 5,000 unique keys of exactly one master-data type.
- Unknown keys roll back the complete operation.
- Bulk reactivation, deletion, storage locations, background jobs and cross-type selection are out
  of scope.
- Preserve existing German and English translations and the current compact table style.
- Stage only task-owned hunks because unrelated local i18n and architecture edits already exist.

---

### Task 1: Atomic application service

**Files:**
- Modify: `backend/internal/application/master_data_lifecycle.go`
- Test: `backend/internal/application/master_data_lifecycle_test.go`

**Interfaces:**
- Consumes: existing `MasterDataService`, `reserveMasterDataWriteTransaction`,
  `ListForManagement`, `ErrMasterDataValidation`, and `ErrMasterDataNotFound`.
- Produces: `SetActiveMany(ctx context.Context, typeName string, keys []string, active bool)
  ([]MasterDataEntry, error)` and `MaxMasterDataActiveBatchSize = 5000`.

- [ ] **Step 1: Write failing service tests**

Add tests that call `SetActiveMany` with duplicate and whitespace-padded keys, verify both matching
entries become inactive once, and verify the response preserves input order with refreshed
capabilities. Add a second test with one valid and one missing key and assert the valid row remains
active after `ErrMasterDataNotFound`.

```go
updated, err := service.SetActiveMany(ctx, "manufacturer",
    []string{" first ", "second", "first"}, false)
if err != nil { t.Fatal(err) }
if len(updated) != 2 || updated[0].Key != "first" || updated[1].Key != "second" {
    t.Fatalf("unexpected result: %#v", updated)
}
```

- [ ] **Step 2: Run the targeted tests and confirm failure**

Run:

```powershell
cd backend
go test ./internal/application -run MasterDataSetActiveMany -count=1
```

Expected: build failure because `SetActiveMany` does not exist.

- [ ] **Step 3: Implement transactional bulk state updates**

Normalize and deduplicate keys before opening a transaction. Reject empty values and more than
5,000 unique keys. Execute the existing active-state update for each key inside one reserved write
transaction and return `ErrMasterDataNotFound` when any update affects zero rows. Commit, invalidate
the cache, load management entries, and return only requested keys in normalized input order.

```go
const MaxMasterDataActiveBatchSize = 5000

func (s *MasterDataService) SetActiveMany(
    ctx context.Context, typeName string, keys []string, active bool,
) ([]MasterDataEntry, error) {
    // validate, update atomically, commit, invalidate, and return refreshed entries
}
```

- [ ] **Step 4: Run service tests**

Run:

```powershell
cd backend
go test ./internal/application -run 'MasterData(SetActiveMany|SetActiveInvalidatesCache)' -count=1
```

Expected: PASS.

- [ ] **Step 5: Format and commit the service slice**

```powershell
gofmt -w backend/internal/application/master_data_lifecycle.go backend/internal/application/master_data_lifecycle_test.go
git add -- backend/internal/application/master_data_lifecycle.go backend/internal/application/master_data_lifecycle_test.go
git commit -m "feat: Stammdaten atomar gesammelt deaktivieren (#142)"
```

### Task 2: HTTP and OpenAPI contract

**Files:**
- Modify: `backend/internal/api/data_handlers.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/master_data_api_test.go`
- Modify: `backend/internal/api/routes_security_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`
- Modify: `frontend/src/shared/api.ts`

**Interfaces:**
- Consumes: `MasterDataService.SetActiveMany` from Task 1.
- Produces: `PATCH /api/v1/master-data/{type}/active`, schema
  `MasterDataActiveBatchInput`, and frontend method
  `api.setMasterDataActiveMany(type, keys, active): Promise<MasterDataEntry[]>`.

- [ ] **Step 1: Write failing API and contract tests**

Add an authenticated Editor request that sends two keys and asserts both returned entries are
inactive. Add invalid-body, empty-keys, over-limit, and missing-key cases. Extend the route security
table and OpenAPI assertions with the collection path and its `400` and `404` Problem responses.

```go
response := doAuthedJSON(t, router, http.MethodPatch,
    "/api/v1/master-data/manufacturer/active",
    `{"keys":["first","second"],"active":false}`,
    editorSession, editorCookies, http.StatusOK)
```

- [ ] **Step 2: Run the targeted tests and confirm failure**

```powershell
cd backend
go test ./internal/api -run 'MasterData(APIBulk|RouteSecurity|OpenAPI)' -count=1
```

Expected: failure because route, handler, and schema do not exist.

- [ ] **Step 3: Add handler, route, schema, and client method**

Decode the following request shape, require `active`, delegate to `SetActiveMany`, and map validation,
not-found, and internal errors consistently with the existing single-entry handler.

```go
type masterDataActiveBatchInput struct {
    Keys   []string `json:"keys"`
    Active *bool    `json:"active"`
}
```

Register the Editor route, document the request and array response in OpenAPI, and add:

```ts
setMasterDataActiveMany: (type: string, keys: string[], active: boolean) =>
  request<MasterDataEntry[]>(`/master-data/${encodeURIComponent(type)}/active`, {
    method: "PATCH",
    body: JSON.stringify({ keys, active })
  })
```

- [ ] **Step 4: Run API and OpenAPI checks**

```powershell
cd backend
go test ./internal/api -run 'MasterData|OpenAPI' -count=1
```

Expected: PASS.

- [ ] **Step 5: Format and commit the contract slice**

```powershell
gofmt -w backend/internal/api/data_handlers.go backend/internal/api/master_data_api_test.go backend/internal/api/routes_security_test.go backend/internal/api/openapi_contract_test.go
git add -- backend/internal/api/data_handlers.go backend/internal/api/routes.go backend/internal/api/master_data_api_test.go backend/internal/api/routes_security_test.go backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml frontend/src/shared/api.ts
git commit -m "feat: Batch-Endpunkt für Stammdaten bereitstellen (#142)"
```

### Task 3: Reusable accessible selection controls

**Files:**
- Create: `frontend/src/features/settings/MasterDataBulkSelection.tsx`
- Create: `frontend/src/features/settings/MasterDataBulkSelection.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/settings.css`

**Interfaces:**
- Consumes: `MasterDataEntry` and existing i18n/button/table tokens.
- Produces: `useMasterDataBulkSelection(scopeKey, entries, visibleEntries)` and
  `MasterDataBulkToolbar({ count, busy, onDeactivate })`.

- [ ] **Step 1: Write failing hook and toolbar tests**

Cover individual toggles, all-visible toggles, native indeterminate state, exclusion of inactive or
incapable entries, pruning after entry updates, and reset when scope or filter key changes. Assert
the toolbar exposes the selected count and disables its action while busy.

```tsx
expect(screen.getByRole("checkbox", {
  name: "Alle sichtbaren aktiven Einträge auswählen"
})).toHaveProperty("indeterminate", true);
```

- [ ] **Step 2: Run the new test and confirm failure**

```powershell
cd frontend
npm.cmd test -- --run src/features/settings/MasterDataBulkSelection.test.tsx
```

Expected: failure because the module does not exist.

- [ ] **Step 3: Implement selection hook, toolbar, translations, and compact styles**

The hook owns a `Set<string>`, exposes `toggle`, `toggleVisible`, `clear`, selected entries, and
header checked/indeterminate state. The toolbar renders only when `count > 0` and uses the existing
secondary/danger button language without a color-only status.

```ts
export function useMasterDataBulkSelection(
  scopeKey: string,
  entries: MasterDataEntry[],
  visibleEntries: MasterDataEntry[]
): MasterDataBulkSelectionState
```

Add matching German and English keys below `settings.master.*` for selection labels, selected count,
action, dialog title, and dialog body.

- [ ] **Step 4: Run component and i18n tests**

```powershell
cd frontend
npm.cmd test -- --run src/features/settings/MasterDataBulkSelection.test.tsx src/shared/i18n.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit only task-owned hunks**

Stage the new files and CSS normally. Stage only the new translation-key hunks from `de.ts` and
`en.ts`; do not include pre-existing article-search text changes.

```powershell
git add -- frontend/src/features/settings/MasterDataBulkSelection.tsx frontend/src/features/settings/MasterDataBulkSelection.test.tsx frontend/src/styles/settings.css
git add -p -- frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: Stammdaten-Auswahl zugänglich bereitstellen (#142)"
```

### Task 4: Integrate general and article master-data tables

**Files:**
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`
- Modify: `frontend/src/features/settings/ArticleManagementSettings.tsx`
- Modify: `frontend/src/features/settings/ArticleManagementSettings.test.tsx`

**Interfaces:**
- Consumes: `api.setMasterDataActiveMany`, `useMasterDataBulkSelection`, and
  `MasterDataBulkToolbar`.
- Produces: accessible selection columns and one confirmed bulk request per master-data type.

- [ ] **Step 1: Write failing integration tests**

For general data, select one row and then all filtered visible rows, confirm the exact count, assert
one `setMasterDataActiveMany` request, and verify the returned rows become inactive and selection
clears. For article data, verify each table has an independent selection and that a failed request
keeps its selection and shows the error.

```tsx
await user.click(screen.getByRole("checkbox", { name: "Tillig auswählen" }));
await user.click(screen.getByRole("button", { name: "Ausgewählte deaktivieren" }));
await user.click(screen.getByRole("button", { name: "2 Einträge deaktivieren" }));
expect(api.setMasterDataActiveMany).toHaveBeenCalledWith("manufacturer", ["tillig", "roco"], false);
```

- [ ] **Step 2: Run integration tests and confirm failure**

```powershell
cd frontend
npm.cmd test -- --run src/features/settings/SettingsView.test.tsx src/features/settings/ArticleManagementSettings.test.tsx
```

Expected: failure because selection columns and bulk action are absent.

- [ ] **Step 3: Integrate the controls**

Add one nonsortable selection column to general data. Use filtered items for the header action and
the current type plus search as reset scope. In `MasterDataSettingsSection`, use the same hook with
the type as scope, call the batch client after confirmation, merge returned entries through
`onChanged`, and keep selection on error. Pass `busy` to individual lifecycle actions and selection
controls.

- [ ] **Step 4: Run targeted frontend tests and build**

```powershell
cd frontend
npm.cmd test -- --run src/features/settings/SettingsView.test.tsx src/features/settings/ArticleManagementSettings.test.tsx src/features/settings/MasterDataBulkSelection.test.tsx
npm.cmd run build
```

Expected: all tests and TypeScript/Vite build PASS.

- [ ] **Step 5: Commit the UI integration**

```powershell
git add -- frontend/src/features/settings/SettingsView.tsx frontend/src/features/settings/SettingsView.test.tsx frontend/src/features/settings/ArticleManagementSettings.tsx frontend/src/features/settings/ArticleManagementSettings.test.tsx
git commit -m "feat: Stammdaten gesammelt deaktivieren (#142)"
```

### Task 5: Full verification and PR readiness

**Files:**
- Modify: `docs/designs/2026-08-28-master-data-bulk-deactivation-design.md` only if verification
  reveals a contract correction.
- Add: `docs/plans/2026-08-28-master-data-bulk-deactivation.md`

**Interfaces:**
- Consumes: all prior task outputs.
- Produces: a verified, reviewable Issue #142 branch.

- [ ] **Step 1: Run backend, frontend, documentation, and diff checks**

```powershell
cd backend
go test ./...
cd ../frontend
npm.cmd test -- --run
npm.cmd run build
cd ../docs
npm.cmd run check
cd ..
git diff --check main...HEAD
```

Expected: all commands exit 0.

- [ ] **Step 2: Inspect the affected UI at desktop and mobile widths**

Verify dark and light themes, keyboard order, indeterminate selection, long German confirmation
text, table-scoped horizontal scrolling, loading/error states, and disabled controls during a
request. Expected: no page-level overflow and every bulk action remains keyboard reachable.

- [ ] **Step 3: Review scope and repository state**

```powershell
git diff --stat main...HEAD
git diff --name-only main...HEAD
git status --short
```

Expected: only Issue #142 files are committed; pre-existing local architecture and article-search
i18n edits remain unstaged.

- [ ] **Step 4: Commit the plan and any verification-only correction**

```powershell
git add -- docs/plans/2026-08-28-master-data-bulk-deactivation.md
git commit -m "docs: Implementierung für Stammdaten-Batch planen (#142)"
```

- [ ] **Step 5: Push, create the bilingual PR, wait for every check, resolve reviews, and squash merge**

Use the exact verified branch head as the merge precondition. After the squash merge, fetch
`origin/main`, set local `main` to that commit without overwriting unrelated working-tree changes,
and verify both refs match.
