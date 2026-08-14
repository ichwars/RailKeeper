# Advanced Geometry Module Port Connections Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Derive module-port connections and provide an explicit server-side snap preview for layout
configurations.

**Architecture:** A pure domain module transforms local ports into configuration coordinates and
derives connections, warnings and snap poses. A focused repository query supplies configuration
placements. Existing configuration persistence remains the only write path.

**Tech Stack:** Go, SQLite, net/http, OpenAPI 3, React, TypeScript, Vitest, Testing Library

## Global Constraints

- All work stays local on `dev/issue-36-advanced-geometry`.
- No new persistence table and no backup-format change.
- Exact connection tolerance is 0.25 mm and 0.5 degrees.
- Snap preview tolerance is 25 mm and 10 degrees.
- Preview never persists or moves a unit automatically.
- Viewer, Editor, Planner and Admin may read analysis. Only Planner and Admin may request previews.
- Messe remains excluded and POST remains CSRF-protected.
- Use TDD and keep German and English translations aligned.

---

### Task 1: Pure module-port geometry

**Files:**
- Create: `backend/internal/domain/layout_port_analysis.go`
- Create: `backend/internal/domain/layout_port_analysis_test.go`

**Interfaces:**
- Produces: `ModulePortPlacement`, `ModulePortAnalysis`, `ModulePortSnapResult`
- Produces: `AnalyzeModulePorts([]ModulePortPlacement)` and
  `FindModulePortSnap(string, TrackPose, []ModulePortPlacement)`

- [x] **Step 1: Write failing domain tests**

Cover translation and rotation, exact compatible connections, open ports, nearby incompatible ports,
archived exclusion, distance/direction boundaries, nearest compatible snap and stable tie-breaking.

- [x] **Step 2: Verify RED**

Run: `cd backend; go test ./internal/domain -run ModulePort -count=1`

Expected: compilation fails because the analysis types and functions do not exist.

- [x] **Step 3: Implement minimal pure geometry**

Reuse `TrackPose`, `TransformTrackPoint` and rotation helpers where their semantics match. Keep JSON
types deterministic and sort connections and issues by unit/port identity.

- [x] **Step 4: Verify GREEN and commit**

Run: `cd backend; go test ./internal/domain -run ModulePort -count=1`

Commit: `feat(layouts): analyze module port geometry`

### Task 2: Application and SQLite context

**Files:**
- Modify: `backend/internal/application/layouts.go`
- Modify: `backend/internal/application/layouts_test.go`
- Modify: `backend/internal/infrastructure/layout_repository.go`
- Modify: `backend/internal/infrastructure/layout_repository_test.go`

**Interfaces:**
- Produces: `LayoutConfigurationPortRepository.LoadConfigurationPortPlacements`
- Produces: `LayoutService.AnalyzeConfigurationPorts` and
  `LayoutService.PreviewConfigurationUnitSnap`

- [x] **Step 1: Write failing service and repository tests**

Assert blank-ID validation, missing configuration mapping, active-port loading, current-pose fallback,
unit membership validation and no mutation during preview.

- [x] **Step 2: Verify RED**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run ConfigurationPort -count=1`

- [x] **Step 3: Implement focused loading and orchestration**

Join `layout_configuration_units`, `layout_units` and active `layout_unit_ports`; return deterministic
placements. Keep calculation in the domain package.

- [x] **Step 4: Verify GREEN and commit**

Commit: `feat(layouts): derive configuration port connections`

### Task 3: HTTP contract and permissions

**Files:**
- Modify: `backend/internal/api/layout_handlers.go`
- Modify: `backend/internal/api/layout_handlers_test.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Produces: `GET /layout-configurations/{id}/port-analysis`
- Produces: `POST /layout-configurations/{id}/unit-snap-preview`

- [x] **Step 1: Write failing handler, role and CSRF tests**

Cover successful analysis/preview, malformed JSON, validation, not found, Viewer/Editor/Planner/Admin/
Messe access and missing CSRF.

- [x] **Step 2: Verify RED**

Run: `cd backend; go test ./internal/api -run 'ConfigurationPort|OpenAPI' -count=1`

- [x] **Step 3: Implement handlers, routes and exact schemas**

Return the shared layout validation/conflict responses and document every result field and enum.

- [x] **Step 4: Verify GREEN and commit**

Commit: `feat(api): expose module port analysis`

### Task 4: Frontend analysis and controlled snap

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`
- Create: `frontend/src/features/layouts/LayoutConfigurationPortAnalysis.tsx`
- Create: `frontend/src/features/layouts/LayoutConfigurationPortAnalysis.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutConfigurationsPanel.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes: Task 3 endpoints
- Produces: typed analysis panel and `An Ports ausrichten` preview action

- [x] **Step 1: Write failing API and component tests**

Cover loading, empty/error states, connection and issue summaries, Viewer read-only behavior, preview
payload, pose transfer into the form and the unsaved-state message.

- [x] **Step 2: Verify RED**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutConfigurationPortAnalysis.test.tsx`

- [x] **Step 3: Implement focused components and responsive styling**

Keep the analysis component read-only. The parent configuration form owns the preview request and
copies the returned pose only when `snapped` is true.

- [x] **Step 4: Verify GREEN and commit**

Commit: `feat(layouts): inspect and align module ports`

### Task 5: Full verification and acceptance

**Files:**
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-modulport-verbindungen-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-modulport-verbindungen.md`
- Modify: `docs/superpowers/specs/2026-08-07-anlagen-zubehoer-gleisplaner-design.md`

- [x] **Step 1: Run full verification**

Run `go test ./...`, `npm.cmd test -- --run`, `npm.cmd run build` and `git diff --check`.

- [x] **Step 2: Restart local server and perform browser QA**

Verify analysis, controlled preview, explicit save, dark theme, narrow layout and browser console.

- [x] **Step 3: Record results and commit**

Commit: `docs: record module port connection acceptance`
