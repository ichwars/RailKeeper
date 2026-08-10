# Freie Planobjekte Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Revisionsgebundene Rechtecke, Ellipsen, Linien und Beschriftungen im maßhaltigen
Gleisplan anlegen, bearbeiten, vergleichen, sichern und wiederherstellen, ohne technische
Gleisanalyse oder Materialbedarf zu verändern.

**Architecture:** `plan_free_objects` bleibt vollständig von `plan_track_objects` getrennt. Ein
schema-versioniertes, diskriminiertes Shape-JSON bildet die vier Formen ab; fokussierte Domain-,
Repository-, Handler- und React-Dateien verhindern weiteres Wachstum der zentralen Trackmodule.

**Tech Stack:** Go, SQLite, React, TypeScript, SVG, Vitest, OpenAPI, RailKeeper-i18n und bestehende
app-eigene UI-Komponenten

## Global Constraints

- Alles bleibt auf `dev/issue-36-advanced-geometry`; kein Push, PR, Merge oder Release.
- Freie Objekte beeinflussen nie Anschlüsse, Issues, Höhen, Stückliste oder Reservierungen.
- Formen sind `rectangle`, `ellipse`, `line` und `label`; Kategorien sind `structure`, `platform`,
  `scenery` und `annotation`.
- Namen enthalten 1 bis 80 Zeichen, Labeltexte 1 bis 120, Schriftgrößen 2 bis 50 mm.
- Alle Maße sind endlich; flächige Maße sind positiv, Linienendpunkte sind nicht der Ursprung.
- Revisionklon, Diff, Backup-Version 13 und OpenAPI bleiben synchron.
- Deutsch und Englisch sowie Hell-/Dunkelmodus und mobile Breiten werden gemeinsam gepflegt.
- Jeder Task folgt red-green-refactor und endet mit einem lokalen Commit.

---

### Task 1: Shape-Domäne und Revisionsvergleich definieren

**Files:**
- Create: `backend/internal/domain/plan_free_object.go`
- Create: `backend/internal/domain/plan_free_object_test.go`
- Create: `backend/internal/domain/plan_free_object_diff.go`
- Create: `backend/internal/domain/plan_free_object_diff_test.go`

**Interfaces:**
- Produces: `FreePlanObjectKind`, `FreePlanObjectCategory`, `FreePlanObjectShape`
- Produces: `PlanFreeObject`, `ValidateFreePlanObjectShape`, `CompareFreePlanObjectRevisions`
- Produces: `PlanFreeObjectChange` using the existing `TrackPlanObjectChangeType` values

- [ ] **Step 1: Failing shape and diff tests write**

Require every valid shape, all category values, invalid schema/non-finite/zero dimensions, empty or
oversized text, unchanged lineage suppression and added/removed/changed ordering:

```go
rectangle := FreePlanObjectShape{
	SchemaVersion: 1, Kind: FreePlanRectangle,
	WidthMM: float64Pointer(300), HeightMM: float64Pointer(80),
}
if err := ValidateFreePlanObjectShape(rectangle); err != nil {
	t.Fatal(err)
}
changed := rectangle
changed.WidthMM = float64Pointer(320)
diff := CompareFreePlanObjectRevisions(
	[]PlanFreeObject{{ID: "old", LineageID: "lineage", Shape: rectangle}},
	[]PlanFreeObject{{ID: "new", LineageID: "lineage", Shape: changed}},
)
if len(diff) != 1 || diff[0].Type != TrackPlanObjectChanged {
	t.Fatalf("unexpected free-object diff: %#v", diff)
}
```

- [ ] **Step 2: Domain RED run**

Run: `cd backend; go test ./internal/domain -run "FreePlan" -count=1`

Expected: FAIL because the free-object types and comparison do not exist.

- [ ] **Step 3: Minimal domain implement**

Define the exact shared model:

```go
type FreePlanObjectShape struct {
	SchemaVersion int                `json:"schemaVersion"`
	Kind          FreePlanObjectKind `json:"kind"`
	WidthMM       *float64           `json:"widthMm,omitempty"`
	HeightMM      *float64           `json:"heightMm,omitempty"`
	EndXMM        *float64           `json:"endXMm,omitempty"`
	EndYMM        *float64           `json:"endYMm,omitempty"`
	Text          string             `json:"text,omitempty"`
	FontSizeMM    *float64           `json:"fontSizeMm,omitempty"`
}

type PlanFreeObject struct {
	ID, LineageID, RevisionID string
	Name                       string
	Category                   FreePlanObjectCategory
	PositionXMM, PositionYMM   float64
	RotationDegrees            float64
	Shape                      FreePlanObjectShape
	Version                    int
	CreatedAt, UpdatedAt       string
}
```

JSON-tag every field with the camel-case names from the spec. `ValidateFreePlanObjectShape` rejects
irrelevant non-nil fields per kind, trims only for validation, uses `math.IsNaN/IsInf`, and returns
`ErrInvalidFreePlanObjectShape`. Compare objects by lineage with 1e-9 tolerance and sorted output.

- [ ] **Step 4: Domain GREEN run**

Run: `cd backend; go test ./internal/domain -run "FreePlan" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

```powershell
git add backend/internal/domain/plan_free_object.go `
  backend/internal/domain/plan_free_object_test.go `
  backend/internal/domain/plan_free_object_diff.go `
  backend/internal/domain/plan_free_object_diff_test.go
git commit -m "feat(planner): define free plan objects"
```

---

### Task 2: Freie Planobjekte migrieren und persistieren

**Files:**
- Create: `backend/migrations/0054_free_plan_objects.sql`
- Create: `backend/internal/infrastructure/free_plan_object_repository.go`
- Create: `backend/internal/infrastructure/free_plan_object_repository_test.go`
- Create: `backend/internal/infrastructure/free_plan_object_migration_test.go`
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/infrastructure/track_planner_repository.go`

**Interfaces:**
- Consumes: `domain.PlanFreeObject`, `domain.ValidateFreePlanObjectShape`
- Extends: `TrackPlan.FreeObjects []domain.PlanFreeObject`
- Produces repository methods: `CreateFreeObject`, `UpdateFreeObject`, `DeleteFreeObject`
- Produces: `GetPlanForFreeObject`

- [ ] **Step 1: Failing migration and repository tests write**

Tests apply migration 0054, require cascading revision deletion, strict JSON hydration, draft-only
CRUD, version conflicts, audit events and `GetPlan` returning an empty non-nil list or persisted
objects. Use inputs introduced in Task 4 as literal repository boundary types only after adding them
to the application file in Step 3.

```go
created, err := repository.CreateFreeObject(ctx, revisionID,
	application.CreateFreePlanObjectInput{
		Name: "Bahnsteig 1", Category: domain.FreePlanPlatform,
		PositionXMM: 200, PositionYMM: 100,
		Shape: domain.FreePlanObjectShape{SchemaVersion: 1, Kind: domain.FreePlanRectangle,
			WidthMM: float64Pointer(500), HeightMM: float64Pointer(70)},
	}, "planner")
if err != nil || created.Version != 1 || created.LineageID != created.ID {
	t.Fatalf("unexpected free object: %#v, %v", created, err)
}
```

- [ ] **Step 2: Persistence RED run**

Run: `cd backend; go test ./internal/infrastructure -run "FreePlan" -count=1`

Expected: FAIL because migration, table and methods are absent.

- [ ] **Step 3: Migration and focused repository implement**

Migration:

```sql
CREATE TABLE plan_free_objects (
  id TEXT PRIMARY KEY,
  lineage_id TEXT NOT NULL,
  revision_id TEXT NOT NULL REFERENCES plan_revisions(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT NOT NULL CHECK(category IN ('structure','platform','scenery','annotation')),
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  rotation_degrees REAL NOT NULL,
  shape_json TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1 CHECK(version >= 1),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(revision_id, lineage_id)
);
CREATE INDEX idx_plan_free_objects_revision ON plan_free_objects(revision_id, created_at, id);
```

In `application/track_planner.go`, add the input structs without service behavior yet:

```go
type CreateFreePlanObjectInput struct {
	Name string `json:"name"`; Category domain.FreePlanObjectCategory `json:"category"`
	PositionXMM float64 `json:"positionXMm"`; PositionYMM float64 `json:"positionYMm"`
	RotationDegrees float64 `json:"rotationDegrees"`; Shape domain.FreePlanObjectShape `json:"shape"`
}
type UpdateFreePlanObjectInput struct {
	CreateFreePlanObjectInput
	ExpectedVersion int `json:"expectedVersion"`
}
```

Implement scan/encode and transactional CRUD in the new repository file. Lock revision state as the
track repository does, normalize no values in persistence, increment versions atomically and write
`PlanFreeObjectCreated|Updated|Deleted` audit events. Extend `GetPlan` to initialize and load
`FreeObjects` ordered by creation time.

- [ ] **Step 4: Persistence GREEN run**

Run: `cd backend; go test ./internal/infrastructure -run "FreePlan" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit Task 2**

```powershell
git add backend/migrations/0054_free_plan_objects.sql `
  backend/internal/infrastructure/free_plan_object_repository.go `
  backend/internal/infrastructure/free_plan_object_repository_test.go `
  backend/internal/infrastructure/free_plan_object_migration_test.go `
  backend/internal/application/track_planner.go `
  backend/internal/infrastructure/track_planner_repository.go
git commit -m "feat(planner): persist free plan objects"
```

---

### Task 3: Revisionklon, Diff und Backup-Version 13 ergänzen

**Files:**
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/application/track_planner_test.go`
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository_test.go`

**Interfaces:**
- Consumes: `CompareFreePlanObjectRevisions`
- Extends: `TrackPlanChangePreview.FreeObjectChanges []domain.PlanFreeObjectChange`
- Produces: cloned free-object lineage and backup version 13 compatibility

- [ ] **Step 1: Failing clone, preview and backup tests write**

Require a cloned object with a new ID, retained lineage and version 1. Require base/current free
changes independent from empty track/material changes. Export and restore all four shapes; create a
version-12 backup without `plan_free_objects` and ensure normalization adds an empty table. Future
version tests advance from 13 to 14.

```go
preview, err := service.ChangePreview(ctx, currentRevisionID)
if err != nil || len(preview.FreeObjectChanges) != 1 ||
	preview.FreeObjectChanges[0].Type != domain.TrackPlanObjectChanged ||
	len(preview.MaterialDeltas) != 0 {
	t.Fatalf("unexpected free-object preview: %#v, %v", preview, err)
}
```

- [ ] **Step 2: Compatibility RED run**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "FreePlan|Backup" -count=1`

Expected: FAIL because clone, preview and backup version 13 are absent.

- [ ] **Step 3: Clone, preview and backup implement**

After cloning track objects, select `name, category, position_x_mm, position_y_mm,
rotation_degrees, shape_json, lineage_id` from the base free-object table and insert new IDs with
the retained lineage, version 1 and current timestamps.

Set `backupVersion = 13`, add `plan_free_objects` to the ordered table list and version policy with
`introduced: 13, required: 13`. During compatibility normalization:

```go
if doc.Version <= 12 {
	if _, exists := doc.Tables["plan_free_objects"]; !exists {
		doc.Tables["plan_free_objects"] = []map[string]any{}
	}
}
```

Populate `FreeObjectChanges` from base and current `TrackPlan.FreeObjects`. Keep the existing track
issue and material comparison unchanged.

- [ ] **Step 4: Compatibility GREEN run**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "FreePlan|Backup" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit Task 3**

```powershell
git add backend/internal/application/track_planner.go `
  backend/internal/application/track_planner_test.go backend/internal/application/backup.go `
  backend/internal/application/backup_test.go `
  backend/internal/infrastructure/layout_revision_repository.go `
  backend/internal/infrastructure/layout_revision_repository_test.go
git commit -m "feat(planner): version free plan objects"
```

---

### Task 4: Service, Planner-Routen und OpenAPI bereitstellen

**Files:**
- Create: `backend/internal/application/free_plan_objects.go`
- Create: `backend/internal/application/free_plan_objects_test.go`
- Create: `backend/internal/api/free_plan_object_handlers.go`
- Create: `backend/internal/api/free_plan_object_handlers_test.go`
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes repository CRUD from Task 2 and shape validation from Task 1
- Produces: `CreateFreeObject`, `UpdateFreeObject`, `DeleteFreeObject` service methods
- Produces: three versioned Planner API routes and complete OpenAPI schemas

- [ ] **Step 1: Failing application, route, role, CSRF and contract tests write**

Require trimming, rotation normalization, valid shape dispatch, rejection of non-finite positions,
invalid category/name/shape/version, Admin/Planner access, Viewer/Editor/Messe denial for writes,
CSRF enforcement, immutable response and conflict response.

```go
created, err := service.CreateFreeObject(ctx, " revision-1 ", CreateFreePlanObjectInput{
	Name: " Bahnsteig 1 ", Category: domain.FreePlanPlatform,
	PositionXMM: 200, PositionYMM: 100, RotationDegrees: -15,
	Shape: rectangle,
}, "planner")
if err != nil || created.Name != "Bahnsteig 1" || created.RotationDegrees != 345 {
	t.Fatalf("free object was not normalized: %#v, %v", created, err)
}
```

- [ ] **Step 2: API RED run**

Run: `cd backend; go test ./internal/application ./internal/api -run "FreePlan|OpenAPI" -count=1`

Expected: FAIL because service methods, handlers, routes and schemas are absent.

- [ ] **Step 3: Focused service and handlers implement**

Extend `TrackPlannerRepository` with the four Task-2 methods. Service normalization uses
`strings.TrimSpace`, `NormalizeTrackRotation`, `validTrackCoordinates`, category validation and
`ValidateFreePlanObjectShape`. Updates copy the normalized embedded create input and require
`ExpectedVersion >= 1`; deletes trim ID and require a positive version.

Handlers follow `decodeLayoutJSON`, `actorUserID`, `trackPlannerError` and `respondJSON` patterns.
Register:

```go
{http.MethodPost, "/api/v1/plan-revisions/{id}/free-objects", routeAccessPlanner,
	(*App).createFreePlanObject, nil},
{http.MethodPut, "/api/v1/plan-free-objects/{id}", routeAccessPlanner,
	(*App).updateFreePlanObject, nil},
{http.MethodDelete, "/api/v1/plan-free-objects/{id}", routeAccessPlanner,
	(*App).deleteFreePlanObject, nil},
```

- [ ] **Step 4: OpenAPI implement and API GREEN run**

Document `FreePlanObjectShape`, `PlanFreeObject`, create/update inputs, change schema,
`TrackPlan.freeObjects`, `TrackPlanChangePreview.freeObjectChanges` and all routes with 400/403/404/
409 responses.

Run: `cd backend; go test ./internal/application ./internal/api -run "FreePlan|OpenAPI" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit Task 4**

```powershell
git add backend/internal/application/free_plan_objects.go `
  backend/internal/application/free_plan_objects_test.go backend/internal/application/track_planner.go `
  backend/internal/api/free_plan_object_handlers.go `
  backend/internal/api/free_plan_object_handlers_test.go backend/internal/api/routes.go `
  backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml
git commit -m "feat(api): manage free plan objects"
```

---

### Task 5: App-eigenen Objekt-Dialog und SVG-Layer bauen

**Files:**
- Create: `frontend/src/features/layouts/FreePlanObjectDialog.tsx`
- Create: `frontend/src/features/layouts/FreePlanObjectDialog.test.tsx`
- Create: `frontend/src/features/layouts/FreePlanObjectLayer.tsx`
- Create: `frontend/src/features/layouts/FreePlanObjectLayer.test.tsx`
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes the Task-4 API routes and schemas
- Produces: `FreePlanObjectDialog` with explicit submit/cancel
- Produces: `FreePlanObjectLayer` with semantic rendering and selection callback

- [ ] **Step 1: Failing type, dialog and layer tests write**

Dialog tests cover create/edit defaults, conditional fields for all four kinds, app-owned selects,
validation, focus/Escape, submit payload and cancel without write. Layer tests assert SVG elements,
category classes, selected non-color outline, rotation transform and accessible names.

```tsx
await user.selectOptionsEquivalent("Form", "Beschriftung");
await user.type(screen.getByRole("textbox", { name: "Text" }), "Gleis 1");
await user.click(screen.getByRole("button", { name: "Planobjekt speichern" }));
expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({
  shape: { schemaVersion: 1, kind: "label", text: "Gleis 1", fontSizeMm: 8 }
}));
```

Use actual `AppSelect` button/option interaction in the implementation test rather than adding a
test-only helper with that example name.

- [ ] **Step 2: Frontend component RED run**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/FreePlanObjectDialog.test.tsx `
  src/features/layouts/FreePlanObjectLayer.test.tsx
```

Expected: FAIL because contracts and components are absent.

- [ ] **Step 3: Contracts and dialog implement**

Add discriminated string unions and a single shape type with optional form-specific properties,
plus create/update API calls. Build the portal dialog using the same focus trap and Escape behavior
as the transition editor. Validate trimmed values before submit and render only fields relevant to
the selected shape.

- [ ] **Step 4: SVG layer, bilingual copy and token styles implement**

Render rectangle, ellipse, line and text inside one transformed `<g role="button">`. Add a
selection outline for each kind and category classes. Reuse `--line`, `--muted`, `--accent`,
`--warning`, panel tokens and font-size tokens; do not add hard-coded theme colors.

- [ ] **Step 5: Component GREEN run and commit**

Run the Task-5 targeted command again. Expected: PASS.

```powershell
git add frontend/src/features/layouts/FreePlanObjectDialog.tsx `
  frontend/src/features/layouts/FreePlanObjectDialog.test.tsx `
  frontend/src/features/layouts/FreePlanObjectLayer.tsx `
  frontend/src/features/layouts/FreePlanObjectLayer.test.tsx `
  frontend/src/shared/apiLayoutsAccessories.ts frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/styles/layouts.css
git commit -m "feat(planner): render free plan objects"
```

---

### Task 6: Freie Objekte in den Gleisplaner integrieren

**Files:**
- Create: `frontend/src/features/layouts/FreePlanObjectInspector.tsx`
- Create: `frontend/src/features/layouts/FreePlanObjectInspector.test.tsx`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.tsx`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.test.tsx`
- Modify: `frontend/src/features/layouts/TrackPlanChangePreviewPanel.tsx`
- Modify: `frontend/src/features/layouts/TrackPlanChangePreviewPanel.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes `TrackPlan.freeObjects`, dialog, layer and CRUD API
- Produces add/edit/drag/rotate/delete workflows and free-object change summary
- Preserves all existing track editing, analysis and reservation behavior

- [ ] **Step 1: Failing integration and regression tests write**

Require centered creation, one API write per drag end, 15-degree rotation, dialog edit, app-owned
delete confirmation, optimistic conflict display, reload recovery, mutually exclusive track/free
selection and no changes to analysis or BOM calls.

```tsx
await user.click(screen.getByRole("button", { name: "Planobjekt hinzufügen" }));
await user.click(screen.getByRole("button", { name: "Planobjekt speichern" }));
expect(api.createFreePlanObject).toHaveBeenCalledWith(revision.id, expect.objectContaining({
  positionXMm: unit.widthMm / 2, positionYMm: unit.heightMm / 2
}));
```

Change-preview tests require separate counts for track and free-object changes.

- [ ] **Step 2: Canvas RED run**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/FreePlanObjectInspector.test.tsx `
  src/features/layouts/TrackPlannerCanvas.test.tsx `
  src/features/layouts/TrackPlanChangePreviewPanel.test.tsx
```

Expected: FAIL because workflows and summary are absent.

- [ ] **Step 3: Inspector and canvas workflows implement**

Keep `selectedID` for tracks and add `selectedFreeObjectID`; every selection clears the other.
Initialize `freeObjects` from `api.trackPlan`. Render `FreePlanObjectLayer` before track groups.
Clamp dragged positions to unit width/height without track snapping. Use the focused inspector for
shape facts and edit/rotate/delete callbacks. Reuse `showError`, `LayoutConfirmDialog` and
`refreshDerived` after writes.

- [ ] **Step 4: Change preview, full tests and build run**

Show `freeObjectChanges.length` independently from track changes. Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/FreePlanObjectInspector.test.tsx `
  src/features/layouts/TrackPlannerCanvas.test.tsx `
  src/features/layouts/TrackPlanChangePreviewPanel.test.tsx
npm.cmd test -- --run
npm.cmd run build
```

Expected: targeted tests, all Vitest files and production build PASS.

- [ ] **Step 5: Commit Task 6**

```powershell
git add frontend/src/features/layouts/FreePlanObjectInspector.tsx `
  frontend/src/features/layouts/FreePlanObjectInspector.test.tsx `
  frontend/src/features/layouts/TrackPlannerCanvas.tsx `
  frontend/src/features/layouts/TrackPlannerCanvas.test.tsx `
  frontend/src/features/layouts/TrackPlanChangePreviewPanel.tsx `
  frontend/src/features/layouts/TrackPlanChangePreviewPanel.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/layouts.css
git commit -m "feat(planner): edit free plan objects"
```

---

### Task 7: Paket I vollständig abnehmen und #36 dokumentieren

**Files:**
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-freie-planobjekte-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-freie-planobjekte.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-free-plan-objects/10-intent.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-free-plan-objects/20-checkpoint.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-free-plan-objects/90-evidence.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-free-plan-objects/99-reflection.md`

**Interfaces:**
- Consumes complete Package I
- Produces local acceptance evidence and a clean #36 implementation branch

- [ ] **Step 1: Full automated verification run**

```powershell
cd backend
go test ./...

cd ..\frontend
npm.cmd test -- --run
npm.cmd run build
```

Expected: all backend packages, all Vitest files and production build PASS.

- [ ] **Step 2: Current server restart and health check**

Rebuild `frontend/dist`, stop only the exact listener on port 18083 and restart with repository data,
migrations, seeds, static directory and repository-local `GOCACHE`. Require `/health` HTTP 200 and
authenticated `codex-test`.

- [ ] **Step 3: Browser acceptance run**

Create one object per shape and verify category rendering. Move and edit the rectangle, rotate the
line, reload and require identical values. Confirm four free objects in the plan, unchanged Tillig
BOM quantities and unchanged track-warning count. Cancel one edit without mutation, delete one
object through the app dialog, verify the change preview and require no console errors in a fresh
tab.

- [ ] **Step 4: Evidence and documentation update**

Record exact commits, test counts, build module count, listener PID, browser values, preserved BOM/
issue counts and the local-only boundary. Mark checkboxes complete only after evidence exists.

- [ ] **Step 5: Documentation commit and clean-state check**

```powershell
git add docs/superpowers/specs/2026-08-10-erweiterte-geometrie-freie-planobjekte-design.md `
  docs/superpowers/plans/2026-08-10-erweiterte-geometrie-freie-planobjekte.md `
  docs/aegis/work/2026-08-10-advanced-geometry-free-plan-objects
git commit -m "docs: record free plan object acceptance"
git diff --check
git status --short
```

Expected: documentation commit succeeds, worktree is clean and no push, PR or merge occurred.
