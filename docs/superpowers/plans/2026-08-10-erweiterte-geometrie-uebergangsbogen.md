# Echte Übergangsbögen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** TILLIG-83125-Flexobjekte wahlweise als echte, serverseitig abgeleitete Euler-Spirale
planen, prüfen, speichern und wieder laden.

**Architecture:** Ein eigener `TransitionCurvePath` bleibt exklusiv zum bestehenden Bézier-
`FlexTrackPath`. Ein fokussiertes Domänenmodul leitet die Euler-Spirale numerisch ab; Repository,
API und UI konsumieren dieselbe effektive Geometrie. Vorschau und persistentes Update bleiben
getrennt.

**Tech Stack:** Go, SQLite, React, TypeScript, SVG, Vitest, OpenAPI, RailKeeper-i18n und bestehende
app-eigene UI-Komponenten

## Global Constraints

- Alles bleibt auf `dev/issue-36-advanced-geometry`; kein Push, PR, Merge oder Release.
- Bestehende `flexPath`-Schema-Version 1 und starre G1-Pläne behalten ihre Bedeutung.
- `flexPath` und `transitionPath` schließen sich gegenseitig aus.
- TILLIG 83125 bleibt auf 664 mm und ein physisches Stück pro Planobjekt begrenzt.
- Der Radiusgrenzwert ist das Maximum aus 543 mm Produktwert und Anlagenlimit.
- Überlänge blockiert die Übernahme; Radiusunterschreitung bleibt eine bewusste Warnung.
- Die Domäne ist alleiniger Owner der effektiven Euler-Geometrie.
- Deutsch und Englisch, Backend, Frontend, Backup und OpenAPI bleiben synchron.
- Jeder Task folgt red-green-refactor und endet mit einem lokalen Commit.

---

### Task 1: Euler-Spirale als effektive Geometrie ableiten

**Files:**
- Create: `backend/internal/domain/track_transition_geometry.go`
- Create: `backend/internal/domain/track_transition_geometry_test.go`
- Modify: `backend/internal/domain/track_planner.go`
- Modify: `backend/internal/domain/track_flex_geometry.go`
- Modify: `backend/internal/domain/track_flex_geometry_test.go`
- Modify: `backend/internal/domain/track_plan_revision_diff.go`
- Modify: `backend/internal/domain/track_plan_revision_diff_test.go`

**Interfaces:**
- Produces: `TransitionDirection`, `TransitionCurvePath`, `BuildTransitionTrackGeometry`
- Extends: `PlanTrackObject.TransitionPath *TransitionCurvePath`
- Extends: `EffectiveGeometryForObject` with exclusive path dispatch

- [ ] **Step 1: Failing domain tests write**

Tests require mirrored left/right points, exact length and radius, monotonically increasing sampled
curvature, maximum 5 mm segment length, invalid schema/direction/non-finite values and rejection when
both path types are present:

```go
left := TransitionCurvePath{
	SchemaVersion: 1, LengthMM: 500, EndRadiusMM: 700, Direction: TransitionLeft,
}
effective, err := BuildTransitionTrackGeometry(left)
if err != nil || effective.LengthMM != 500 || effective.MinimumRadiusMM == nil ||
	*effective.MinimumRadiusMM != 700 || len(effective.Geometry.Routes) != 1 {
	t.Fatalf("unexpected transition geometry: %#v, %v", effective, err)
}
right := left
right.Direction = TransitionRight
mirrored, err := BuildTransitionTrackGeometry(right)
if err != nil || !transitionPointsMirrored(
	effective.Geometry.Routes[0].Points, mirrored.Geometry.Routes[0].Points,
) {
	t.Fatalf("transition points are not mirrored: %#v", mirrored)
}
```

- [ ] **Step 2: Domain RED run**

Run: `cd backend; go test ./internal/domain -run "Transition|EffectiveTrack" -count=1`

Expected: FAIL because the transition types and builder do not exist.

- [ ] **Step 3: Minimal transition domain implement**

Define:

```go
type TransitionDirection string

const (
	TransitionLeft  TransitionDirection = "left"
	TransitionRight TransitionDirection = "right"
)

type TransitionCurvePath struct {
	SchemaVersion int                 `json:"schemaVersion"`
	LengthMM      float64             `json:"lengthMm"`
	EndRadiusMM   float64             `json:"endRadiusMm"`
	Direction     TransitionDirection `json:"direction"`
}

func BuildTransitionTrackGeometry(path TransitionCurvePath) (EffectiveTrackGeometry, error)
```

Use `theta(s) = sign*s*s/(2*R*L)`. Choose segment length
`min(5, sqrt(8*R*0.05))`, reject more than 4.096 segments and integrate X/Y per segment with
Simpson weights at start, middle and end. Produce ports `a` and `b`, route `main`, exact effective
length and a pointer to `EndRadiusMM`.

`EffectiveGeometryForObject` rejects both paths together, dispatches transition first when only it is
present and preserves the existing Bézier/fixed fallback. `trackObjectsDiffer` compares all four
transition fields with the established floating tolerance.

- [ ] **Step 4: Domain GREEN run**

Run: `cd backend; go test ./internal/domain -run "Transition|FlexTrack|EffectiveTrack|Revision" -count=1`

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

```powershell
git add backend/internal/domain/track_transition_geometry.go `
  backend/internal/domain/track_transition_geometry_test.go backend/internal/domain/track_planner.go `
  backend/internal/domain/track_flex_geometry.go backend/internal/domain/track_flex_geometry_test.go `
  backend/internal/domain/track_plan_revision_diff.go `
  backend/internal/domain/track_plan_revision_diff_test.go
git commit -m "feat(planner): derive transition curve geometry"
```

---

### Task 2: Übergangsbögen persistieren, klonen und sichern

**Files:**
- Create: `backend/migrations/0053_transition_curve_paths.sql`
- Create: `backend/internal/infrastructure/transition_curve_migration_test.go`
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/infrastructure/track_planner_repository.go`
- Modify: `backend/internal/infrastructure/track_planner_repository_test.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository_test.go`

**Interfaces:**
- Persists: `plan_track_objects.transition_path_json`
- Extends: create/update inputs with `TransitionPath *domain.TransitionCurvePath`
- Produces: backup version 12 and legacy normalization through version 11

- [ ] **Step 1: Failing migration, repository, clone and backup tests write**

The migration test applies migrations through 0052, inserts an existing Bézier object, applies 0053
and checks the new nullable column plus the unchanged JSON. Repository tests create, update and load a
transition object and reject simultaneous paths. Revision tests clone `transition_path_json`.
Backup tests export and restore a transition path and remove the new column from a version-11 copy.

```go
transition := domain.TransitionCurvePath{
	SchemaVersion: 1, LengthMM: 500, EndRadiusMM: 700, Direction: domain.TransitionLeft,
}
created, err := service.CreateObject(ctx, revisionID, application.CreatePlanTrackObjectInput{
	GeometryID: flexGeometryID, PositionXMM: 100, PositionYMM: 100,
	TransitionPath: &transition,
}, "planner-1")
if err != nil || created.TransitionPath == nil || created.FlexPath != nil {
	t.Fatalf("transition path not persisted: %#v, %v", created, err)
}
```

- [ ] **Step 2: Persistence RED run**

Run:

```powershell
cd backend
go test ./internal/application ./internal/infrastructure -run "Transition|BackupVersionTwelve" -count=1
```

Expected: FAIL because migration 0053, repository mapping and backup version 12 are absent.

- [ ] **Step 3: Migration, repository and backup implement**

Migration content:

```sql
ALTER TABLE plan_track_objects ADD COLUMN transition_path_json TEXT;
```

Add `TransitionPath` to plan object inputs. Validate exactly one path for flex geometry and no path
for rigid geometry. `CreateObject`, `UpdateObject`, `trackObjectSelect` and `scanTrackObject` map
`transition_path_json` adjacent to `flex_path_json`; JSON decoding uses the same strict error path.
Clone both JSON columns unchanged.

Set `backupVersion = 12`; future-version tests use 13. For documents through version 11:

```go
if doc.Version <= 11 {
	for _, row := range doc.Tables["plan_track_objects"] {
		row["transition_path_json"] = nil
	}
}
```

- [ ] **Step 4: Persistence GREEN run**

Run:

```powershell
cd backend
go test ./internal/application ./internal/infrastructure -run "Transition|Backup" -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 2**

```powershell
git add backend/migrations/0053_transition_curve_paths.sql `
  backend/internal/infrastructure/transition_curve_migration_test.go `
  backend/internal/application/backup.go backend/internal/application/backup_test.go `
  backend/internal/application/track_planner.go `
  backend/internal/infrastructure/track_planner_repository.go `
  backend/internal/infrastructure/track_planner_repository_test.go `
  backend/internal/infrastructure/layout_revision_repository.go `
  backend/internal/infrastructure/layout_revision_repository_test.go
git commit -m "feat(planner): persist transition curve paths"
```

---

### Task 3: Versionierte Übergangsbogen-Vorschau und OpenAPI bereitstellen

**Files:**
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/application/track_planner_test.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/track_planner_handlers.go`
- Modify: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Produces: `TransitionCurvePreviewInput`, `TransitionCurvePreview`
- Produces: `TrackPlannerService.PreviewTransitionPath`
- Produces: `POST /api/v1/plan-track-objects/{id}/transition-preview`

- [ ] **Step 1: Failing application, API, role, CSRF and contract tests write**

```go
preview, err := service.PreviewTransitionPath(t.Context(), "flex-1", TransitionCurvePreviewInput{
	LengthMM: 500, EndRadiusMM: 700, Direction: domain.TransitionLeft, ExpectedVersion: 3,
})
if err != nil || preview.EffectiveLengthMM != 500 || preview.EffectiveMinimumRadiusMM != 700 ||
	preview.LengthExceeded || preview.RadiusBelowLimit || !preview.Applicable {
	t.Fatalf("unexpected transition preview: %#v, %v", preview, err)
}
```

Also require conflict on stale version, validation for rigid objects and invalid directions, no
repository mutation, Planner/Admin access, Viewer denial, CSRF enforcement and complete OpenAPI
schemas.

- [ ] **Step 2: API RED run**

Run:

```powershell
cd backend
go test ./internal/application ./internal/api -run "TransitionCurvePreview|OpenAPI" -count=1
```

Expected: FAIL because service, route, handler and schemas are absent.

- [ ] **Step 3: Service, handler and contract implement**

Define:

```go
type TransitionCurvePreviewInput struct {
	LengthMM       float64                    `json:"lengthMm"`
	EndRadiusMM    float64                    `json:"endRadiusMm"`
	Direction      domain.TransitionDirection `json:"direction"`
	ExpectedVersion int                       `json:"expectedVersion"`
}

type TransitionCurvePreview struct {
	Path                     domain.TransitionCurvePath `json:"path"`
	EffectiveGeometry        domain.TrackGeometry       `json:"effectiveGeometry"`
	EffectiveLengthMM        float64                    `json:"effectiveLengthMm"`
	EffectiveMinimumRadiusMM float64                    `json:"effectiveMinimumRadiusMm"`
	RadiusLimitMM            float64                    `json:"radiusLimitMm"`
	LengthExceeded           bool                       `json:"lengthExceeded"`
	RadiusBelowLimit         bool                       `json:"radiusBelowLimit"`
	Applicable               bool                       `json:"applicable"`
}
```

The service loads the object and limits, verifies version and flex kind, builds the domain geometry,
rejects incomplete geometry, sets length/radius flags and never writes. Add the Planner route and
thin handler. OpenAPI documents the path, both new schemas, `transitionPath` on object/input and
`left|right` direction.

- [ ] **Step 4: API GREEN run**

Run:

```powershell
cd backend
go test ./internal/application ./internal/api -run "Transition|OpenAPI" -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 3**

```powershell
git add backend/internal/application/track_planner.go `
  backend/internal/application/track_planner_test.go backend/internal/api/routes.go `
  backend/internal/api/track_planner_handlers.go backend/internal/api/track_planner_handlers_test.go `
  backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml
git commit -m "feat(api): preview transition curve paths"
```

---

### Task 4: App-eigenen Übergangsbogen-Editor integrieren

**Files:**
- Create: `frontend/src/features/layouts/TransitionCurveEditorDialog.tsx`
- Create: `frontend/src/features/layouts/TransitionCurveEditorDialog.test.tsx`
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.tsx`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes: transition preview endpoint and `TransitionCurvePath`
- Produces: app-owned preview/apply flow and explicit path conversion

- [ ] **Step 1: Failing API, dialog and canvas tests write**

Dialog tests require server preview, displayed length/radius/direction, warning, disabled overlength,
explicit apply, cancel-without-write and focus behavior. Canvas tests require both Flex actions,
exclusive update payloads, effective render and path preservation on pose/height updates.

```tsx
await user.click(screen.getByRole("button", { name: "Übergangsbogen" }));
await user.click(screen.getByRole("button", { name: "Verlauf vorschlagen" }));
expect(api.previewTransitionCurvePath).toHaveBeenCalledWith(object.id, {
  lengthMm: 500, endRadiusMm: 700, direction: "left", expectedVersion: object.version
});
await user.click(screen.getByRole("button", { name: "Übergangsbogen übernehmen" }));
expect(api.updatePlanTrackObject).toHaveBeenCalledWith(object.id, expect.objectContaining({
  transitionPath: preview.path, flexPath: null
}));
```

- [ ] **Step 2: Frontend RED run**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/TransitionCurveEditorDialog.test.tsx `
  src/features/layouts/TrackPlannerCanvas.test.tsx
```

Expected: FAIL because types, dialog and integration are absent.

- [ ] **Step 3: Frontend contracts and dialog implement**

Add discriminated direction and path types, preview types and
`api.previewTransitionCurvePath(id, input)`. Implement a portal dialog using `AppNumberInput` for
length/radius, `AppSelect` for left/right, the effective server route for SVG preview and the same
focus/escape patterns as `FlexTrackEditorDialog`.

- [ ] **Step 4: Canvas conversion and preservation implement**

Extend plan objects and update payloads with `transitionPath?: TransitionCurvePath | null`. Preserve
the active path on drag, rotation and height changes. Applying a transition sends
`transitionPath: path, flexPath: null`; applying a free path sends `flexPath: path,
transitionPath: null`. Both editor actions remain available and use deterministic straight/default
values when converting.

- [ ] **Step 5: Bilingual copy and styles implement**

Add German and English labels for title, length, end radius, left/right, preview, apply, conversion,
length warning and radius warning. Reuse existing modal tokens and add only transition-specific SVG
and fact-grid selectors under the layout stylesheet.

- [ ] **Step 6: Frontend GREEN, full suite and build run**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/TransitionCurveEditorDialog.test.tsx `
  src/features/layouts/TrackPlannerCanvas.test.tsx
npm.cmd test -- --run
npm.cmd run build
```

Expected: targeted tests, all Vitest files and production build PASS.

- [ ] **Step 7: Commit Task 4**

```powershell
git add frontend/src/features/layouts/TransitionCurveEditorDialog.tsx `
  frontend/src/features/layouts/TransitionCurveEditorDialog.test.tsx `
  frontend/src/shared/apiLayoutsAccessories.ts frontend/src/features/layouts/TrackPlannerCanvas.tsx `
  frontend/src/features/layouts/TrackPlannerCanvas.test.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/styles/layouts.css
git commit -m "feat(planner): edit transition curve paths"
```

---

### Task 5: Paket H vollständig abnehmen und dokumentieren

**Files:**
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-uebergangsbogen-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-uebergangsbogen.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-transition-curves/10-intent.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-transition-curves/20-checkpoint.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-transition-curves/90-evidence.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-transition-curves/99-reflection.md`

**Interfaces:**
- Consumes: complete Package H
- Produces: local acceptance record and clean branch

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

Restart only the exact listener on port 18083 with repository data, migrations, seeds, `dist` and
repository-local `GOCACHE`. Require `/health` HTTP 200 and authenticated `codex-test` session.

- [ ] **Step 3: Browser acceptance run**

Place or select one 83125, preview 500 mm at 700 mm left, apply and reload. Require effective length
500,00 mm, radius 700,00 mm and curved left route. Preview right and require the mirrored route.
Preview 665 mm and require disabled apply. Preview radius 600 mm against layout limit 700 mm and
require one visible, applicable warning. Cancel must preserve the saved left path; BOM stays at one
83125 and a clean new browser tab has no console errors.

- [ ] **Step 4: Evidence and plan checkboxes update**

Record commits, exact test counts, build module count, listener PID, browser values and the explicit
local-only boundary. Mark all plan steps complete only after evidence exists.

- [ ] **Step 5: Documentation commit and clean-state check**

```powershell
git add docs/superpowers/specs/2026-08-10-erweiterte-geometrie-uebergangsbogen-design.md `
  docs/superpowers/plans/2026-08-10-erweiterte-geometrie-uebergangsbogen.md `
  docs/aegis/work/2026-08-10-advanced-geometry-transition-curves
git commit -m "docs: record transition curve acceptance"
git diff --check
git status --short
```

Expected: documentation commit succeeds, worktree is clean and no push, PR or merge occurred.
