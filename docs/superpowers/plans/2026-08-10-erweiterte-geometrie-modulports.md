# Advanced Geometry Module Ports Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist and manage typed module ports as the first reviewable package of Stage 4.

**Architecture:** A focused `layout_unit_ports` table belongs to existing layout units. The layout
application service owns validation and optimistic concurrency, the existing layout repository owns
SQLite operations and audit records, and a focused React panel owns the UI. No connection analysis or
automatic module movement is introduced yet.

**Tech Stack:** Go, SQLite, net/http, OpenAPI 3, React, TypeScript, Vitest, Testing Library

## Global Constraints

- All changes stay local on a new branch stacked on `dev/issue-32-tillig-track-planner`.
- Use migration `0048_layout_unit_ports.sql`; never rewrite historical migrations.
- Planner and Admin may write. Viewer, Editor, Planner and Admin may read. Messe has no access.
- Writes require CSRF protection and optimistic version checks.
- Backup format becomes 7; versions 1 through 6 remain importable.
- No port snapping, compatibility analysis, automatic module ordering or digital control in this package.
- Implement behavior test-first and keep German and English translations aligned.

---

### Task 1: Domain and service contract

**Files:**
- Modify: `backend/internal/domain/layout.go`
- Modify: `backend/internal/application/layouts.go`
- Modify: `backend/internal/application/layouts_test.go`

**Interfaces:**
- Produces: `domain.LayoutUnitPortKind`, `application.LayoutUnitPort`,
  `application.CreateLayoutUnitPortInput`, `application.UpdateLayoutUnitPortInput`
- Produces: `LayoutService.ListUnitPorts`, `CreateUnitPort`, and `UpdateUnitPort`

- [x] **Step 1: Write failing service tests**

Cover trimming and lowercase normalization, direction normalization, non-finite and out-of-bounds
coordinates, invalid kinds, blank interface keys, duplicate-independent service calls, and versions
below one. Extend the repository spy with the three wished-for methods.

- [x] **Step 2: Run the focused tests and verify RED**

Run: `cd backend; go test ./internal/application -run LayoutUnitPort -count=1`

Expected: compilation fails because the port types and service methods do not exist.

- [x] **Step 3: Implement the minimal domain and service behavior**

Define kinds `track`, `power`, `digital`, `feedback`, `accessory`, and `other`. Add repository methods:

```go
ListUnitPorts(context.Context, string) ([]LayoutUnitPort, error)
CreateUnitPort(context.Context, string, CreateLayoutUnitPortInput, string) (*LayoutUnitPort, error)
UpdateUnitPort(context.Context, string, UpdateLayoutUnitPortInput, string) (*LayoutUnitPort, error)
```

Load the owning unit before writes so boundary validation uses its width and height. Normalize the
direction with the existing rotation helper.

- [x] **Step 4: Run focused tests and verify GREEN**

Run: `cd backend; go test ./internal/application -run LayoutUnitPort -count=1`

- [x] **Step 5: Format and commit**

Run: `gofmt -w backend/internal/domain/layout.go backend/internal/application/layouts.go backend/internal/application/layouts_test.go`

Commit: `feat(layouts): define module port service`

### Task 2: SQLite persistence and migration

**Files:**
- Create: `backend/migrations/0048_layout_unit_ports.sql`
- Modify: `backend/internal/infrastructure/layout_repository.go`
- Modify: `backend/internal/infrastructure/layout_repository_test.go`
- Modify: `backend/internal/infrastructure/layout_schema_test.go`

**Interfaces:**
- Consumes: Task 1 port types and repository methods
- Produces: deterministic list/create/update persistence with audit events and version conflicts

- [x] **Step 1: Write failing repository and schema tests**

Assert migration columns, foreign key, case-insensitive unique unit/name constraint, ordered listing,
create audit data, update audit data, version increments, stale-version rejection, and unit lookup.

- [x] **Step 2: Run focused tests and verify RED**

Run: `cd backend; go test ./internal/infrastructure -run 'LayoutUnitPort|LayoutSchema' -count=1`

- [x] **Step 3: Add migration and repository implementation**

Persist IDs as generated UUIDs, booleans as `0/1`, and timestamps with the established repository
helpers. Order active ports first, then case-insensitive name and ID. Map missing rows to
`ErrLayoutNotFound` and stale updates to `ErrLayoutVersionConflict`.

- [x] **Step 4: Run focused tests and verify GREEN**

Run: `cd backend; go test ./internal/infrastructure -run 'LayoutUnitPort|LayoutSchema' -count=1`

- [x] **Step 5: Format and commit**

Commit: `feat(layouts): persist module ports`

### Task 3: HTTP, roles, CSRF and OpenAPI

**Files:**
- Modify: `backend/internal/api/layout_handlers.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/layout_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: Task 1 service methods
- Produces: `GET/POST /layout-units/{id}/ports` and `PUT /layout-unit-ports/{id}`

- [x] **Step 1: Write failing API tests**

Test successful list/create/update responses, malformed JSON, validation and conflict mapping, the
Viewer/Editor/Planner/Admin/Messe matrix, and missing CSRF rejection for both writes.

- [x] **Step 2: Run focused tests and verify RED**

Run: `cd backend; go test ./internal/api -run 'LayoutUnitPort|OpenAPI' -count=1`

- [x] **Step 3: Implement routes, handlers and contract**

Reuse the existing layout error writer and request decoding. Add exact OpenAPI schemas for all port
fields and role/security descriptions. Do not expose a DELETE route.

- [x] **Step 4: Run focused tests and verify GREEN**

Run: `cd backend; go test ./internal/api -run 'LayoutUnitPort|OpenAPI' -count=1`

- [x] **Step 5: Format and commit**

Commit: `feat(api): expose layout module ports`

### Task 4: Backup format 7

**Files:**
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Interfaces:**
- Consumes: `layout_unit_ports` table from Task 2
- Produces: format-7 export/validate/import with formats 1 through 6 still accepted

- [x] **Step 1: Write failing backup tests**

Add a format-7 roundtrip containing a port and a format-6 compatibility test where the table is absent.
Assert restored unit references and all port fields.

- [x] **Step 2: Run focused tests and verify RED**

Run: `cd backend; go test ./internal/application -run Backup -count=1`

- [x] **Step 3: Add the table and compatibility rules**

Raise the current backup version to 7, export/import the table after `layout_units`, require it only
for version 7, and preserve the existing auth-data exclusions.

- [x] **Step 4: Run focused tests and verify GREEN**

Run: `cd backend; go test ./internal/application -run Backup -count=1`

- [x] **Step 5: Format and commit**

Commit: `feat(backup): preserve layout module ports`

### Task 5: Frontend client and module-port panel

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Create: `frontend/src/features/layouts/LayoutModulePortsPanel.tsx`
- Create: `frontend/src/features/layouts/LayoutModulePortsPanel.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutModulesPanel.tsx`
- Modify: `frontend/src/styles/layouts.css`

**Interfaces:**
- Consumes: Task 3 HTTP endpoints
- Produces: typed API methods and a role-aware port list/editor under the selected layout unit

- [x] **Step 1: Write failing component tests**

Cover lazy loading after unit selection, table rendering, no-selection/empty/error states, create,
edit/archive payloads with expected version, app-owned kind select, Viewer read-only rendering and
German labels. Include a long-interface-key fixture.

- [x] **Step 2: Run focused tests and verify RED**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutModulePortsPanel.test.tsx`

- [x] **Step 3: Implement client, panel, integration and responsive styles**

Keep the component focused. `LayoutModulesPanel` passes its selected unit and `canPlan`. The new panel
owns loading and form state. Use `AppSelect`, existing buttons/tokens and a two-column form that
collapses at the existing narrow breakpoint.

- [x] **Step 4: Run focused tests and verify GREEN**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutModulePortsPanel.test.tsx`

- [x] **Step 5: Commit**

Commit: `feat(layouts): manage module ports`

### Task 6: Verification and acceptance record

**Files:**
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-modulports-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-modulports.md`
- Modify: `docs/superpowers/specs/2026-08-07-anlagen-zubehoer-gleisplaner-design.md`

**Interfaces:**
- Consumes: Tasks 1 through 5
- Produces: verified local Package-A acceptance evidence

- [x] **Step 1: Run complete backend tests**

Run: `cd backend; go test ./...`

- [x] **Step 2: Run complete frontend tests and build**

Run: `cd frontend; npm.cmd test -- --run`

Run: `cd frontend; npm.cmd run build`

- [x] **Step 3: Restart the local server and perform browser QA**

Verify create/edit/archive, read-only behavior, custom selects, long text, dark theme, narrow width,
and an empty unit. Confirm no browser console errors.

- [x] **Step 4: Record exact test counts and local acceptance**

Mark every completed plan checkbox and add the date, commands, test counts, build result and browser
findings to the design status. Update the parent design with the local Stage-4 Package-A state.

- [x] **Step 5: Commit the acceptance record**

Commit: `docs: record module port package acceptance`
