# Tillig-TT-Gleisplaner, Paket A Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eine verifizierte Tillig-TT-Geometrie und revisionsgebundene Gleisobjekte maßhaltig in einem
lokalen Planentwurf platzieren, verschieben, drehen und wieder entfernen.

**Architecture:** Geometriedefinitionen sind unveränderliche, quellenbezogene Bibliotheksdaten.
Planobjekte referenzieren eine konkrete Geometrie-ID und gehören exakt zu einer Planrevision. Ein
eigener `TrackPlannerService` hält Planerlogik aus der bestehenden Anlagenverwaltung heraus; das
Frontend erhält eine fokussierte SVG-Arbeitsfläche im Register `Planer`.

**Tech Stack:** Go, SQLite, OpenAPI 3, React, TypeScript, SVG, Vitest

## Global Constraints

- RailKeeper bleibt lokal, selbst gehostet und SQLite-basiert.
- RailKeeper plant und dokumentiert, sendet aber keine digitalen Steuerbefehle.
- Nur geometrisch geprüfte Bibliothekseinträge sind platzierbar.
- Veröffentlichte Planrevisionen bleiben unveränderlich.
- Alle Schreibzugriffe sind CSRF-geschützt, rollenbasiert und auditiert.
- Millimeter sind die interne Geometrieeinheit.
- Paket A enthält kein Snapping, keine Verbindungserkennung, keine Stückliste und kein Flexgleis.
- Als erster Herstellerdatensatz gilt Tillig 83101, Gleisstück G1, gerade 166 mm, TT 1:120,
  Spurweite 12 mm, Quelle `https://www.tillig.com/Produkte/produktinfo-83101.html`.
- Änderungen und Commits bleiben lokal. Kein Push, keine PR und kein Merge.

---

### Task 1: Versionierte Geometriebibliothek und Planobjektschema

**Files:**
- Create: `backend/migrations/0045_track_planner_foundation.sql`
- Create: `backend/internal/domain/track_planner.go`
- Create: `backend/internal/domain/track_planner_test.go`
- Create: `backend/internal/infrastructure/track_planner_schema_test.go`

**Interfaces:**
- Produces: `TrackGeometryKind`, `TrackGeometryStatus`, `TrackPoint`, `TrackPort`, `TrackRoute`,
  `TrackGeometryDefinition` and `PlanTrackObject`.
- Persists: `track_geometry_libraries`, `track_geometry_definitions`, `plan_track_objects`.

- [ ] **Step 1: Write failing domain tests**

Test that only `straight`, `curve`, `turnout` and `crossing` kinds are valid in Paket A, only
`verified` definitions are placeable, rotation normalizes into `[0, 360)`, and the G1 route has two
ports at `(0,0)` and `(166,0)`.

- [ ] **Step 2: Run the domain test and verify failure**

Run: `cd backend; go test ./internal/domain -run TestTrack -count=1`

Expected: FAIL because the track planner domain types do not exist.

- [ ] **Step 3: Add the migration and domain types**

The migration creates:

```sql
CREATE TABLE track_geometry_libraries (
  id TEXT PRIMARY KEY,
  manufacturer TEXT NOT NULL,
  track_system TEXT NOT NULL,
  gauge TEXT NOT NULL,
  scale TEXT NOT NULL,
  version TEXT NOT NULL,
  source_url TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'verified', 'retired')),
  created_at TEXT NOT NULL,
  UNIQUE (manufacturer, track_system, gauge, version)
);

CREATE TABLE track_geometry_definitions (
  id TEXT PRIMARY KEY,
  library_id TEXT NOT NULL,
  article_number TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('straight', 'curve', 'turnout', 'crossing')),
  length_mm REAL NOT NULL CHECK (length_mm > 0),
  geometry_json TEXT NOT NULL,
  source_url TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'verified', 'retired')),
  created_at TEXT NOT NULL,
  FOREIGN KEY (library_id) REFERENCES track_geometry_libraries(id) ON DELETE RESTRICT,
  UNIQUE (library_id, article_number)
);

CREATE TABLE plan_track_objects (
  id TEXT PRIMARY KEY,
  revision_id TEXT NOT NULL,
  geometry_id TEXT NOT NULL,
  position_x_mm REAL NOT NULL,
  position_y_mm REAL NOT NULL,
  rotation_degrees REAL NOT NULL CHECK (rotation_degrees >= 0 AND rotation_degrees < 360),
  version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (revision_id) REFERENCES plan_revisions(id) ON DELETE CASCADE,
  FOREIGN KEY (geometry_id) REFERENCES track_geometry_definitions(id) ON DELETE RESTRICT
);
```

Seed one verified library and one verified geometry. `geometry_json` uses schema version 1:

```json
{
  "schemaVersion": 1,
  "ports": [
    {"id": "a", "xMm": 0, "yMm": 0, "directionDegrees": 180},
    {"id": "b", "xMm": 166, "yMm": 0, "directionDegrees": 0}
  ],
  "routes": [
    {"id": "main", "points": [{"xMm": 0, "yMm": 0}, {"xMm": 166, "yMm": 0}]}
  ]
}
```

- [ ] **Step 4: Test migration tables, foreign keys and immutable source data**

Verify that unknown revisions/geometries fail, invalid rotation fails, and deleting a referenced
geometry fails. Verify the Tillig G1 source URL, 166 mm length and both ports after JSON decoding.

- [ ] **Step 5: Run tests and commit**

Run: `cd backend; go test ./internal/domain ./internal/infrastructure -run "TestTrack" -count=1`

Expected: PASS.

Commit: `feat(planner): add verified track geometry foundation`

---

### Task 2: Planer-Service und revisionssichere Persistenz

**Files:**
- Create: `backend/internal/application/track_planner.go`
- Create: `backend/internal/application/track_planner_test.go`
- Create: `backend/internal/infrastructure/track_planner_repository.go`
- Create: `backend/internal/infrastructure/track_planner_repository_test.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository.go`

**Interfaces:**
- Produces: `TrackPlannerRepository`, `TrackPlannerService`, `TrackPlan`,
  `CreatePlanTrackObjectInput`, `UpdatePlanTrackObjectInput`.
- Methods: `ListGeometries(ctx, gauge)`, `GetPlan(ctx, revisionID)`,
  `CreateObject(ctx, revisionID, input, actor)`, `UpdateObject(ctx, id, input, actor)`,
  `DeleteObject(ctx, id, expectedVersion, actor)`.

- [ ] **Step 1: Write failing service tests**

Cover trimmed inputs, finite coordinates, normalized rotation, positive versions, verified geometry
requirement, draft-only writes and typed validation/conflict/immutable errors.

- [ ] **Step 2: Run the service test and verify failure**

Run: `cd backend; go test ./internal/application -run TestTrackPlanner -count=1`

Expected: FAIL because `TrackPlannerService` does not exist.

- [ ] **Step 3: Implement the focused service and repository**

Use explicit errors:

```go
var (
    ErrTrackPlanValidation = errors.New("track plan validation failed")
    ErrTrackPlanNotFound   = errors.New("track plan object not found")
    ErrTrackPlanImmutable  = errors.New("track plan revision is immutable")
    ErrTrackPlanConflict   = errors.New("track plan object version conflict")
)
```

All mutations run in one transaction, verify revision status `draft`, verify geometry and library
status `verified`, increment object versions and write `PlanTrackObjectCreated`,
`PlanTrackObjectUpdated` or `PlanTrackObjectDeleted` audit events.

- [ ] **Step 4: Clone base-revision objects when creating a draft**

Inside the existing `CreateDraft` transaction, copy every `plan_track_objects` row from
`base_revision_id` into the new draft with fresh IDs and version 1. Do not copy objects from unrelated
variants or draft bases.

- [ ] **Step 5: Test persistence and commit**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "TestTrackPlanner|TestPlanRevision" -count=1`

Expected: PASS.

Commit: `feat(planner): persist revision-bound track objects`

---

### Task 3: Rollen-API, OpenAPI und typisierter Client

**Files:**
- Create: `backend/internal/api/track_planner_handlers.go`
- Create: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/cmd/railkeeper/main.go`
- Modify: `openapi/railkeeper.yaml`
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`

**Interfaces:**
- GET `/api/v1/track-geometries?gauge=TT` for Viewer, Editor, Planner and Admin.
- GET `/api/v1/plan-revisions/{id}/track-plan` for Viewer, Editor, Planner and Admin.
- POST `/api/v1/plan-revisions/{id}/track-objects` for Planner and Admin.
- PUT `/api/v1/plan-track-objects/{id}` for Planner and Admin.
- DELETE `/api/v1/plan-track-objects/{id}?expectedVersion=1` for Planner and Admin.

- [ ] **Step 1: Write failing route and client tests**

Cover Viewer reads, Planner writes, Admin writes, Editor/Messe write denial, CSRF denial, invalid input,
immutable revision, stale version and encoded IDs in the TypeScript client.

- [ ] **Step 2: Run tests and verify failure**

Run: `cd backend; go test ./internal/api -run TestTrackPlannerRoutes -count=1`

Run: `cd frontend; npm.cmd test -- --run src/shared/apiLayoutsAccessories.test.ts`

Expected: both fail because routes and client methods do not exist.

- [ ] **Step 3: Implement handlers, wiring and OpenAPI schemas**

Keep handlers limited to decoding, service calls and problem mapping. Add the service to `api.Config`
and `api.App`; construct it from a dedicated infrastructure repository in `main.go`.

- [ ] **Step 4: Implement the typed client**

Expose `trackGeometries`, `trackPlan`, `createPlanTrackObject`, `updatePlanTrackObject` and
`deletePlanTrackObject`. Avoid unchecked casts and duplicate transport types.

- [ ] **Step 5: Run contract tests and commit**

Run: `cd backend; go test ./internal/api -run "TestTrackPlannerRoutes|TestOpenAPI" -count=1`

Run: `cd frontend; npm.cmd test -- --run src/shared/apiLayoutsAccessories.test.ts`

Expected: PASS.

Commit: `feat(planner): expose track plan API`

---

### Task 4: Maßhaltige SVG-Arbeitsfläche im Planer-Register

**Files:**
- Create: `frontend/src/features/layouts/TrackPlannerCanvas.tsx`
- Create: `frontend/src/features/layouts/TrackPlannerCanvas.test.tsx`
- Create: `frontend/src/features/layouts/trackPlannerGeometry.ts`
- Create: `frontend/src/features/layouts/trackPlannerGeometry.test.ts`
- Modify: `frontend/src/features/layouts/LayoutPlansPanel.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes the Task-3 client types and methods.
- Produces an SVG canvas whose viewBox uses the selected unit dimensions in millimetres.

- [ ] **Step 1: Write failing geometry and component tests**

Cover exact G1 length, object transform, Planner/Admin edit controls, Viewer read-only mode, palette
filtering to verified TT geometries, placement, pointer movement, 15-degree rotation buttons, deletion,
loading, empty and conflict states.

- [ ] **Step 2: Run tests and verify failure**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/trackPlannerGeometry.test.ts src/features/layouts/TrackPlannerCanvas.test.tsx`

Expected: FAIL because the editor modules do not exist.

- [ ] **Step 3: Implement geometry helpers and SVG rendering**

Render every route as SVG `polyline`. Apply each plan object's translation and rotation as one SVG
group transform. Preserve millimetres in the viewBox. Render ports as small non-color-only endpoint
markers. Do not introduce HTML geometry overlays.

- [ ] **Step 4: Implement explicit plan editing**

Opening a draft with `Plan bearbeiten` is the explicit edit action. Palette activation creates G1 at
the unit centre. Pointer movement updates the local draft and saves on pointer release. Rotation uses
visible `-15°` and `+15°` actions. Deletion requires the existing app-owned confirmation dialog.
Published/review/archived revisions open with `Plan ansehen` and expose no mutation controls.

- [ ] **Step 5: Add German/English copy and responsive CSS**

Keep the canvas dominant, palette compact and inspector calm. On narrow screens stack palette, canvas
and inspector without native browser dialogs.

- [ ] **Step 6: Run tests, build and commit**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/trackPlannerGeometry.test.ts src/features/layouts/TrackPlannerCanvas.test.tsx src/features/layouts/LayoutsView.test.tsx`

Run: `cd frontend; npm.cmd run build`

Expected: PASS.

Commit: `feat(planner): place verified Tillig track in drafts`

---

### Task 5: Backup-Version 5 und Gesamtprüfung

**Files:**
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`
- Modify: `backend/internal/infrastructure/layout_accessory_schema_test.go`
- Modify: `docs/superpowers/specs/2026-08-07-anlagen-zubehoer-gleisplaner-design.md`

**Interfaces:**
- Backup version 5 requires all three Paket-A tables.
- Versions 1 through 4 remain importable with empty track-planner tables.

- [ ] **Step 1: Write failing backup roundtrip and compatibility tests**

Export a verified library, G1 geometry, draft and published plan objects. Mutate the database, restore
the export and verify IDs, geometry JSON, coordinates, rotations and revision links. Validate that a
version-4 document without Paket-A tables remains compatible and a version-5 document without any of
the three tables is rejected before mutation.

- [ ] **Step 2: Run tests and verify failure**

Run: `cd backend; go test ./internal/application -run "TestBackupVersionFive|TestBackupVersionFourWithoutTrackPlanner" -count=1`

Expected: FAIL because backup version 5 is not defined.

- [ ] **Step 3: Register version-5 tables and update stable documentation**

Preserve authentication exclusions. Do not export audit events. Document that Paket A intentionally
contains only verified straight G1 and that snapping, validation and BOM are Paket B.

- [ ] **Step 4: Run complete verification**

Run: `cd backend; go test -p=1 ./...`

Run: `cd frontend; npm.cmd test -- --run`

Run: `cd frontend; npm.cmd run build`

Run: `git diff --check`

Expected: all commands pass.

- [ ] **Step 5: Perform local browser acceptance and commit**

Verify dark/light desktop rendering, draft-only actions, placement at 166 mm, movement, rotation,
deletion confirmation and persistence after reload. Keep the local server on port 18083 and leave the
authenticated Planer view open.

Commit: `test(planner): verify first Tillig geometry slice`
