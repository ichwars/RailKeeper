# Tillig TT Track Planner Package B Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add authoritative endpoint snapping, derived plan validation and a stock-aware bill of materials to the existing revision-bound Tillig TT track planner.

**Architecture:** Pure geometry and plan analysis stay in the domain package. The application service applies snapping before repository writes and combines geometric BOM lines with repository-provided local inventory availability. A read-only analysis endpoint feeds typed React panels, while a matching TypeScript helper provides immediate drag preview without replacing the server result.

**Tech Stack:** Go, SQLite, OpenAPI 3.1, React, TypeScript, SVG, Vitest

## Global Constraints

- All geometry uses millimetres and degrees.
- Snap distance is 8 mm and snap direction tolerance is 5 degrees.
- Connection tolerance is 0.25 mm and 0.5 degrees.
- Only verified geometry is placeable.
- Derived connections, issues and BOM lines are not persisted.
- Open ends and warnings do not block drafts.
- No native browser dialogs, digital layout control or automatic material creation.
- Changes remain local on `dev/issue-32-tillig-track-planner`.

---

### Task 1: Pure geometry, snapping and analysis

**Files:**
- Create: `backend/internal/domain/track_plan_analysis.go`
- Create: `backend/internal/domain/track_plan_analysis_test.go`

**Interfaces:**
- Consumes: `domain.PlanTrackObject`, `domain.TrackPort`, `domain.TrackRoute`.
- Produces: `TransformTrackPoint`, `TransformTrackPort`, `FindTrackSnap`, `AnalyzeTrackPlan`,
  `TrackSnapResult`, `TrackPlanConnection`, `TrackPlanIssue`, `TrackBOMLine`, `TrackPlanAnalysis`.

- [ ] Write table-driven failing tests for 0/90/180-degree point transforms, exact G1 endpoints,
  the 8 mm and 5 degree snap boundaries, nearest compatible candidate selection and normalized pose.
- [ ] Run `go test ./internal/domain -run "TestTrackTransform|TestFindTrackSnap" -count=1` and verify RED.
- [ ] Implement point/port transforms and `FindTrackSnap` using Euclidean distance and circular angle
  distance. Align the moving port direction opposite the target before calculating translation.
- [ ] Run the focused transform and snapping tests and verify GREEN.
- [ ] Write failing tests for two connected G1 pieces, two open ends, a nearby incompatible end,
  collinear overlap, and BOM grouping by geometry version.
- [ ] Implement `AnalyzeTrackPlan` with deterministic ordering and issue severity `warning` or `error`.
- [ ] Run `go test ./internal/domain -run "TestAnalyzeTrackPlan" -count=1` and verify GREEN.
- [ ] Commit as `feat(planner): derive track connections and issues`.

### Task 2: Authoritative service snapping and stock-aware analysis

**Files:**
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/application/track_planner_test.go`
- Modify: `backend/internal/infrastructure/track_planner_repository.go`
- Modify: `backend/internal/infrastructure/track_planner_repository_test.go`

**Interfaces:**
- Produces repository methods `GetPlanForObject(ctx, objectID)` and
  `TrackMaterialAvailability(ctx, bom)`.
- Produces service method `AnalyzePlan(ctx, revisionID)` returning connections, issues, BOM and
  material status.

- [ ] Write failing service tests proving updates snap to the nearest compatible endpoint and remain
  unchanged outside tolerance.
- [ ] Extend the repository interface and implement `GetPlanForObject`; apply `FindTrackSnap` inside
  `UpdateObject` before the existing optimistic write.
- [ ] Run `go test ./internal/application ./internal/infrastructure -run "TestTrackPlannerSnap" -count=1`.
- [ ] Write failing repository tests with Tillig 83101 quantity stock and active reservations.
- [ ] Implement deterministic product matching on manufacturer plus article number and aggregate
  physical, reserved, free and missing quantities without mutating inventory.
- [ ] Run focused application and repository tests and commit as
  `feat(planner): analyze track plan material status`.

### Task 3: Analysis API, OpenAPI and typed client

**Files:**
- Modify: `backend/internal/api/track_planner_handlers.go`
- Modify: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`

**Interfaces:**
- Adds `GET /api/v1/plan-revisions/{id}/track-analysis` for Viewer, Editor, Planner and Admin.
- Adds client method `trackPlanAnalysis(revisionId)` with no unchecked casts.

- [ ] Write failing role, missing-revision and encoded-client-path tests.
- [ ] Add the route and handler using the existing track-planner problem mapping.
- [ ] Run `go test ./internal/api -run "TestTrackPlanner|TestOpenAPI" -count=1`.
- [ ] Add complete OpenAPI schemas for connections, issues, BOM and material status.
- [ ] Add matching TypeScript types and client method, run its Vitest file and commit as
  `feat(planner): expose track plan analysis`.

### Task 4: Magnetic drag preview and status markers

**Files:**
- Modify: `frontend/src/features/layouts/trackPlannerGeometry.ts`
- Modify: `frontend/src/features/layouts/trackPlannerGeometry.test.ts`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.tsx`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.test.tsx`

**Interfaces:**
- Produces `snapTrackPose(moving, objects)` with the same constants and tie-breaking as Go.
- Consumes `api.trackPlanAnalysis` and the authoritative snapped update response.

- [ ] Write failing tests for live snap preview, server correction, connected/open/conflict marker
  classes and read-only published plans.
- [ ] Implement local preview during pointer movement and keep server response authoritative on release.
- [ ] Load analysis after every successful mutation and render non-color-only endpoint symbols.
- [ ] Run the focused geometry and component tests and commit as
  `feat(planner): preview magnetic endpoint snapping`.

### Task 5: Plan-check and material panels

**Files:**
- Create: `frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx`
- Create: `frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes `TrackPlanAnalysis` and optional selected object id.
- Produces compact plan-check and BOM sections with issue selection callbacks.

- [ ] Write failing tests for empty valid plan, open-end warnings, incompatible connections,
  overlaps, grouped quantity, available/reserved/free/missing stock and absent catalog matches.
- [ ] Implement the focused analysis component with buttons that select affected SVG objects.
- [ ] Add German and English copy and responsive CSS using existing tokens.
- [ ] Run all layout tests and `npm.cmd run build`; commit as
  `feat(planner): show plan checks and material demand`.

### Task 6: Full local verification and documentation

**Files:**
- Modify: `docs/superpowers/specs/2026-08-07-anlagen-zubehoer-gleisplaner-design.md`
- Modify: `docs/superpowers/specs/2026-08-09-tillig-tt-gleisplaner-paket-b-design.md`

- [ ] Update the stable status: Paket B covers snapping, derived validation and stock-aware BOM;
  reservation and revision change preview remain Paket C.
- [ ] Run `go test ./...` with the repository-local Go cache.
- [ ] Run `npm.cmd run test -- --run` and `npm.cmd run build`.
- [ ] Rebuild and restart the local server, verify health, login, snapping, issue markers, BOM and no
  browser console errors.
- [ ] Commit documentation as `docs: record track planner package B acceptance`.
