# RailKeeper Quality, Security, and Maintainability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add durable security and quality gates, establish meaningful frontend regression tests,
and then reduce the largest files without changing RailKeeper's behavior or public API.

**Architecture:** Deliver the work as small pull requests. Security automation and the frontend test
foundation are independent and come first. Refactoring starts only after the affected behavior has
tests, stays inside the existing Go packages and React feature structure, and moves one coherent
responsibility at a time.

**Tech Stack:** GitHub Actions, Dependabot, CodeQL, govulncheck, npm audit, golangci-lint, Go 1.25,
React 19, TypeScript 6, Vite 8, Vitest, Testing Library, jsdom.

## Global Constraints

- Preserve the local-first, self-hosted architecture and SQLite storage model.
- Do not change API paths, JSON shapes, database schemas, authorization rules, or visible behavior
  during file-splitting pull requests.
- Keep `frontend/src/shared/api.ts`, backend routes, and `openapi/railkeeper.yaml` aligned.
- Preserve all existing security controls described in `docs/security.md`.
- Every write route remains CSRF-protected and explicitly role-protected server-side.
- External URLs remain restricted through the existing safe-fetch and redirect checks.
- Filesystem paths remain confined through the existing data-directory helper.
- Do not combine broad formatting changes with functional or structural changes.
- Keep generated output, caches, `frontend/dist`, `frontend/node_modules`, runtime data, and local
  credentials out of Git.
- Run `gofmt` after Go moves and keep `.ps1`, `.bat`, and `.cmd` files on CRLF.

---

## Delivery Order

| Priority | Pull request | Deliverable | Dependency |
| --- | --- | --- | --- |
| P0 | PR 1 | Dependabot, CodeQL, root security policy | none |
| P0 | PR 2 | Scheduled dependency audits and Go lint | PR 1 |
| P0 | PR 3 | Frontend test runner and pure-logic tests | none |
| P0 | PR 4 | Frontend workflow tests and coverage gate | PR 3 |
| P1 | PR 5-10 | Incremental `VehiclesView.tsx` extraction | PR 4 |
| P1 | PR 11 | Route authorization metadata and regression test | PR 1 |
| P2 | PR 12-14 | `router.go` split | PR 11 |
| P2 | PR 15-17 | `article_search.go` and `vehicles.go` split | existing Go tests |
| P2 | PR 18 | Trivy container and configuration scan | stable security baseline |

PR 1 and PR 3 may run in parallel. PR 5 starts only after PR 4 is green. Backend file moves remain
separate from frontend extraction so reviewers can isolate regressions.

## Target File Map

### Security and CI

| File | Responsibility |
| --- | --- |
| `.github/dependabot.yml` | Weekly updates for Go modules, npm, Docker, and GitHub Actions |
| `.github/workflows/codeql.yml` | CodeQL analysis for Go and JavaScript/TypeScript |
| `.github/workflows/security-audit.yml` | Scheduled `govulncheck` and runtime npm audit |
| `.github/workflows/ci.yml` | Required backend, frontend test, build, and lint gates |
| `.golangci.yml` | Small, explicit Go lint policy |
| `SECURITY.md` | Supported versions and private vulnerability reporting process |
| `docs/security.md` | RailKeeper security invariants and automated controls |

### Frontend tests and extraction

| File | Responsibility |
| --- | --- |
| `frontend/vitest.config.ts` | jsdom environment, setup file, coverage configuration |
| `frontend/src/test/setup.ts` | jest-dom matchers and deterministic browser cleanup |
| `frontend/src/test/fixtures/vehicles.ts` | Typed vehicle and CV fixtures |
| `frontend/src/features/vehicles/*.test.ts(x)` | Vehicle logic and workflow regression tests |
| `frontend/src/features/importExport/importExportHelpers.test.tsx` | CSV/XML mapping and ECoS merge tests |
| `frontend/src/features/vehicles/useVehicleEditorController.ts` | Selection, modal mode, form and save lifecycle |
| `frontend/src/features/vehicles/useArticleSearchController.ts` | Article and barcode search state |
| `frontend/src/features/vehicles/useVehicleMediaController.ts` | Images and attachments |
| `frontend/src/features/vehicles/useVehicleMaintenanceController.ts` | Maintenance edits and completion |
| `frontend/src/features/vehicles/useVehicleSparePartsController.ts` | Spare-part search, status and imports |
| `frontend/src/features/vehicles/useVehicleDigitalController.ts` | Functions, CV values and CV files |
| `frontend/src/features/vehicles/useVehicleOutputController.ts` | QR, reports and exhibition assignment |
| `frontend/src/features/vehicles/VehicleEditorDialog.tsx` | Editor dialog composition and tab routing |
| `frontend/src/features/vehicles/VehiclesView.tsx` | Page-level orchestration only |

### Backend extraction

| Current file | Target files |
| --- | --- |
| `backend/internal/api/router.go` | `router.go`, `system_handlers.go`, `auth_handlers.go`, `digital_center_handlers.go`, `vehicle_handlers.go`, `vehicle_media_handlers.go`, `vehicle_detail_handlers.go`, `backup_handlers.go`, `master_data_handlers.go`, `http_helpers.go` |
| `backend/internal/application/article_search.go` | `article_search.go`, `article_search_queries.go`, `article_search_ranking.go`, `article_search_html.go`, `article_search_spare_parts.go`, `article_search_documents.go`, `article_search_pdf.go`, `article_search_text.go` |
| `backend/internal/application/vehicles.go` | `vehicles.go`, `vehicle_external_mappings.go`, `vehicle_media.go`, `vehicle_maintenance.go`, `vehicle_spare_parts.go`, `vehicle_functions.go`, `vehicle_cv.go`, `vehicle_inventory_numbers.go`, `vehicle_validation.go` |

## Task 1: Dependabot and Security Policy

**Files:**

- Create: `.github/dependabot.yml`
- Create: `SECURITY.md`
- Modify: `docs/security.md`

**Interfaces:**

- Produces weekly update pull requests for `/backend`, `/frontend`, the root `Dockerfile`, and
  `.github/workflows`.
- Produces a public reporting entry point that links to GitHub private vulnerability reporting.

- [ ] Create `.github/dependabot.yml` with `version: 2` and four weekly entries:
  `gomod` in `/backend`, `npm` in `/frontend`, `docker` in `/`, and `github-actions` in `/`.
- [ ] Schedule all four entries for Monday at `04:00` in `Europe/Berlin` and cap each ecosystem at
  five open update pull requests.
- [ ] Group minor and patch updates per ecosystem; keep major updates in separate pull requests.
- [ ] Create root `SECURITY.md` with these sections: supported versions, private reporting via
  `https://github.com/ichwars/RailKeeper/security/advisories/new`, response expectations, excluded
  public issue reporting, and a link to `docs/security.md`.
- [ ] Extend `docs/security.md` with mandatory invariants: explicit route roles, CSRF on writes,
  negative-path authorization tests, safe external fetching, confined data paths, auth data excluded
  from backups, and session revocation after password changes.
- [ ] Add an "Automated checks" section listing CodeQL, Dependabot, govulncheck, npm audit,
  golangci-lint, backend tests, frontend tests, and frontend builds.
- [ ] Run `git diff --check` and inspect the YAML indentation manually.
- [ ] Commit with `chore(security): add dependency update policy`.

## Task 2: CodeQL and Dependency Audits

**Files:**

- Create: `.github/workflows/codeql.yml`
- Create: `.github/workflows/security-audit.yml`

**Interfaces:**

- CodeQL analyzes `go` and `javascript-typescript` independently.
- Scheduled audits fail on reachable Go vulnerabilities or high-severity runtime npm findings.

- [ ] Create `codeql.yml` for pushes to `main`, pull requests targeting `main`, a weekly schedule,
  and manual dispatch.
- [ ] Set least-privilege workflow permissions: `contents: read`, `security-events: write`,
  `packages: read`, and `actions: read`.
- [ ] Use a matrix with exactly `go` and `javascript-typescript`, then run
  `github/codeql-action/init@v4`, `autobuild@v4`, and `analyze@v4`.
- [ ] Create `security-audit.yml` for lockfile changes, Monday scheduling, and manual dispatch.
- [ ] In its Go job, use `backend/go.mod`, install
  `golang.org/x/vuln/cmd/govulncheck@v1.6.0`, and run `govulncheck ./...` from `backend`.
- [ ] In its npm job, run `npm ci` and `npm audit --omit=dev --audit-level=high` from `frontend`.
- [ ] Write both command outputs to the workflow summary without exposing repository secrets.
- [ ] Add a final scheduled-only reporting job with `issues: write`; on failure, create or update one
  open issue titled `Security audit requires attention` and link the failed workflow run. Do not
  create duplicate issues for repeated failures.
- [ ] Push the branch and verify that both CodeQL matrix jobs and both dependency audit jobs run.
- [ ] Commit with `ci(security): add code and dependency scanning`.

## Task 3: Go Lint Gate

**Files:**

- Create: `.golangci.yml`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**

- The CI lint job scans the module rooted at `backend/go.mod`.
- Lint uses `golangci-lint-action@v9` with `golangci-lint` `v2.11.4`.

- [ ] Create a version-2 golangci-lint config with a five-minute timeout and the standard linter set.
- [ ] Explicitly enable `bodyclose`, `errcheck`, `govet`, `ineffassign`, `misspell`, `nilerr`,
  `noctx`, `staticcheck`, and `unused`.
- [ ] Run the pinned linter locally against `backend` and save the unmodified output in the pull
  request description, not in the repository.
- [ ] Fix findings that are behavior-preserving and covered by tests.
- [ ] For any disputed finding, add a line-scoped `//nolint:<name> // reason` comment; do not disable
  the linter globally.
- [ ] Add a separate `backend-lint` CI job using `actions/setup-go@v6`, `backend/go.mod`, and
  `golangci/golangci-lint-action@v9` with `working-directory: backend`.
- [ ] Run `go test ./...`, `go build ./...`, and the linter from `backend`.
- [ ] Commit with `ci(go): enforce lint checks`.

## Task 4: Frontend Test Foundation

**Files:**

- Modify: `frontend/package.json`
- Modify: `frontend/package-lock.json`
- Create: `frontend/vitest.config.ts`
- Create: `frontend/src/test/setup.ts`
- Create: `frontend/src/test/fixtures/vehicles.ts`
- Create: `frontend/src/shared/i18n.test.ts`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**

- Produces `npm run test`, `npm run test:run`, and `npm run test:coverage`.
- Tests run in jsdom and use the existing German locale by default.

- [ ] Install `vitest`, `@vitest/coverage-v8`, `jsdom`, `@testing-library/react`,
  `@testing-library/user-event`, and `@testing-library/jest-dom` as dev dependencies.
- [ ] Add scripts: `test: vitest`, `test:run: vitest run`, and
  `test:coverage: vitest run --coverage`.
- [ ] Configure jsdom, global test APIs, `src/test/setup.ts`, automatic mock restoration, and v8
  text/JSON/HTML coverage reports in `vitest.config.ts`.
- [ ] In `setup.ts`, import `@testing-library/jest-dom/vitest`; after each test, clear
  `localStorage`, `sessionStorage`, mocks, and DOM state.
- [ ] Add typed minimal fixtures for an analog vehicle, a digital vehicle, maintenance entries,
  images, functions, CV values, and CV files.
- [ ] Add `i18n.test.ts` covering the German default, English selection, interpolation, fallback to
  German, and the language-change browser event; this is the first real test for the new runner.
- [ ] Add `npm run test:run` before `npm run build` in the existing frontend CI job.
- [ ] Run `npm run test:run -- src/shared/i18n.test.ts`, then run the complete suite and build.
- [ ] Commit with `test(frontend): add Vitest foundation`.

## Task 5: Pure Logic Regression Tests

**Files:**

- Create: `frontend/src/features/vehicles/cvImport.test.ts`
- Create: `frontend/src/features/vehicles/articleSearch.test.ts`
- Create: `frontend/src/features/vehicles/vehicleTransforms.test.ts`
- Create: `frontend/src/features/vehicles/useVehicleInventoryController.test.tsx`
- Create: `frontend/src/features/importExport/importExportHelpers.test.tsx`

**Interfaces:**

- Tests lock down behavior that later controller and component extraction must preserve.

- [ ] Test CV import from JSON, semicolon text, duplicate CV/profile keys, invalid values, new rows,
  changed rows, and unchanged rows.
- [ ] Test article-field boolean conversion, selection keys, current values, conflict status, and
  safe source display names.
- [ ] Test vehicle-to-form conversion, primary-image selection, attachment edit state, function edit
  state, exhibition eligibility, and exhibition entry conversion.
- [ ] Test inventory filtering for digital/analog, missing images, maintenance due, missing decoder,
  manufacturer/category/gattung, and exhibition readiness.
- [ ] Test inventory sorting, visible selection, reset behavior, URL filter presets, and localStorage
  view persistence with `renderHook`.
- [ ] Test CSV delimiter detection, quoted CSV parsing, XML parsing, normalized header aliases,
  default column mapping, booleans, import-row creation, ECoS matching, merge changes, and CSV export.
- [ ] Run each new file directly with `npm run test:run -- <path>` while developing.
- [ ] Run the complete frontend suite and build.
- [ ] Commit with `test(frontend): cover vehicle and import logic`.

## Task 6: Frontend Workflow Tests and Coverage Gate

**Files:**

- Create: `frontend/src/features/vehicles/VehiclesView.test.tsx`
- Create: `frontend/src/features/vehicles/VehicleCVTab.test.tsx`
- Create: `frontend/src/features/importExport/ImportExportView.test.tsx`
- Modify: `frontend/vitest.config.ts`
- Modify: `.github/workflows/ci.yml`

**Interfaces:**

- Mocks the exported `api` object at the HTTP boundary; component internals remain unmocked.
- Establishes regression coverage before `VehiclesView.tsx` is split.

- [ ] Add a reusable typed API mock with successful defaults for vehicle and master-data reads.
- [ ] Test initial inventory load, text search, opening a vehicle, create-form required validation,
  successful save, and server-error display in `VehiclesView.test.tsx`.
- [ ] Test CV file preview as a read-only step, explicit import confirmation, duplicate handling, and
  blocked file behavior in `VehicleCVTab.test.tsx`.
- [ ] Test CSV preview, manual column remapping, conflict display, explicit import confirmation, and
  ECoS preview without sync confirmation in `ImportExportView.test.tsx`.
- [ ] Require at least 80 percent lines/statements/functions and 70 percent branches for
  `cvImport.ts`, `articleSearch.ts`, `vehicleTransforms.ts`, and new controller files. Do not claim a
  repository-wide percentage while large legacy views remain outside the threshold.
- [ ] Change frontend CI from `npm run test:run` to `npm run test:coverage`, then run the build.
- [ ] Commit with `test(frontend): cover critical inventory workflows`.

## Task 7: Extract VehiclesView Controllers

**Files:**

- Create the controller files listed in the target file map.
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Modify the related tests from Tasks 5 and 6.

**Interfaces:**

- Each hook owns one state group and returns a named object of state plus commands.
- Hooks receive typed dependencies such as `selected`, `setSelectedDetail`, `reloadVehicles`, and
  `t`; they do not import another controller's mutable state.

- [ ] Extract `useVehicleEditorController` first and move form/mode/modal/save lifecycle with no JSX.
- [ ] Run `VehiclesView.test.tsx`, then the complete frontend suite and build.
- [ ] Commit only that extraction with `refactor(vehicles): extract editor controller`.
- [ ] Repeat the test-move-test-commit cycle for article search, media, maintenance, spare parts,
  digital data, and output/exhibition, one controller per commit.
- [ ] Keep API calls and messages byte-for-byte equivalent unless a test proves a bug and the bug fix
  is submitted separately.
- [ ] Stop and redesign a hook if its public return object exceeds roughly 25 fields; split by user
  workflow instead of passing a giant props object onward.

## Task 8: Extract Vehicle Editor Composition

**Files:**

- Create: `frontend/src/features/vehicles/VehicleEditorDialog.tsx`
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Modify: `frontend/src/features/vehicles/VehiclesView.test.tsx`

**Interfaces:**

- `VehicleEditorDialog` composes the existing model, upload, maintenance, spare-part, function, CV,
  speed-curve, and read-only components.
- `VehiclesView` owns page loading, inventory controller composition, and dialog selection only.

- [ ] Move the editor dialog JSX without changing markup, CSS classes, labels, tab order, or event
  behavior.
- [ ] Replace broad positional props with typed controller result objects grouped by workflow.
- [ ] Verify create, view, edit, delete, CV import, file upload, maintenance, spare parts, QR, report,
  and exhibition flows through the existing tests.
- [ ] Reach the first structural target: `VehiclesView.tsx` below 1,000 lines, no more than 15 direct
  `useState` calls, and every new file below 500 lines where practical.
- [ ] Run `npm run test:coverage` and `npm run build`.
- [ ] Commit with `refactor(vehicles): isolate editor dialog`.

## Task 9: Make Route Authorization Machine-Checkable

**Files:**

- Modify: `backend/internal/api/routes.go`
- Create: `backend/internal/api/routes_security_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`

**Interfaces:**

- Extend `routeSpec` with an explicit access value: `public`, `Admin`, `Editor`, `Viewer`, or `Messe`.
- Route registration and API inventory consume the same metadata, eliminating a duplicate policy
  list.

- [ ] Add an explicit `routeAccess` type and require every `routeSpec` to declare access.
- [ ] Build route registration from the same specs used by the OpenAPI coverage test.
- [ ] Preserve the intentional public routes: health, version, setup status/admin, login, password
  reset request/confirm, logout, and current session.
- [ ] Add a test that fails when an `/api/v1` route has no access declaration.
- [ ] Add a test that fails when a mutating route is public unless it is in the explicit bootstrap
  and authentication allowlist.
- [ ] Add table-driven negative-path requests proving unauthenticated access fails and insufficient
  roles fail for Admin, Editor, Viewer, and Messe routes.
- [ ] Run `go test ./internal/api -run 'Route|OpenAPI|Auth' -count=1`, then `go test ./...`.
- [ ] Commit with `test(api): enforce route authorization metadata`.

## Task 10: Split Backend God Files Mechanically

**Files:**

- Move functions from the three current files into the target files listed above.
- Split `router_test.go`, `article_search_test.go`, and `vehicles_test.go` along the same boundaries.

**Interfaces:**

- Package names, exported types, exported functions, method receivers, SQL, and HTTP behavior remain
  unchanged.

- [ ] Move one responsibility group at a time with no logic edits.
- [ ] Run the nearest package tests after every move.
- [ ] Run `gofmt` only on moved Go files.
- [ ] Keep each commit independently buildable and reviewable.
- [ ] Split `router.go` first, retaining `Config`, `App`, `NewRouter`, top-level middleware wiring, and
  static routing in that file.
- [ ] Split article search second into orchestration, queries, ranking, HTML extraction, spare parts,
  documents, PDF/OCR, and text cleanup.
- [ ] Split vehicles third into core vehicle lifecycle, external mappings, media, maintenance, spare
  parts, functions, CV data, inventory numbering, and validation.
- [ ] Target `router.go` and `article_search.go` below 500 lines and `vehicles.go` below 800 lines;
  keep extracted files below 500 lines where practical.
- [ ] Run `go test ./...`, `go build ./...`, and golangci-lint after each pull request.
- [ ] Use separate commits: `refactor(api): split router handlers`,
  `refactor(search): split article search responsibilities`, and
  `refactor(vehicles): split vehicle persistence responsibilities`.

## Task 11: Add Trivy and Enforce Repository Gates

**Files:**

- Create: `.github/workflows/trivy.yml`
- Modify: `.github/workflows/docker-image.yml`

**Interfaces:**

- Pull requests scan the repository for vulnerable dependencies, secrets, and configuration errors.
- Published container images receive a vulnerability scan before release completion.

- [ ] Add a filesystem/configuration scan on pull requests with SARIF upload to GitHub code scanning.
- [ ] Use `aquasecurity/trivy-action@v0.36.0` and
  `github/codeql-action/upload-sarif@v4`; do not use `master`, `latest`, or another moving action
  reference.
- [ ] Scan the exact image tags produced by `docker-image.yml` and fail on unfixed critical or high
  runtime vulnerabilities after recording justified exceptions in a tracked ignore file.
- [ ] Give SARIF upload only `security-events: write`; all build and scan steps retain
  `contents: read`.
- [ ] Enable GitHub private vulnerability reporting, Dependabot alerts, and Dependabot security
  updates in repository settings.
- [ ] Require these branch checks after one successful main-branch run: backend tests, backend lint,
  frontend coverage, frontend build, both CodeQL languages, and OpenAPI contract tests.
- [ ] Confirm scheduled audit failure creates or updates one issue and that a successful later run
  closes it or marks it resolved.
- [ ] Commit with `ci(security): scan container and configuration`.

## Final Verification

- [ ] From `backend`, run `go test ./...`.
- [ ] From `backend`, run `go build ./...`.
- [ ] Run the pinned golangci-lint configuration against `backend`.
- [ ] From `frontend`, run `npm run test:coverage`.
- [ ] From `frontend`, run `npm run build`.
- [ ] Run `git diff --check`.
- [ ] Confirm `openapi/railkeeper.yaml` still covers every route.
- [ ] Confirm Viewer, Editor, Admin, and Messe negative-path tests pass.
- [ ] Confirm no database migration, API contract change, generated frontend build, runtime data, or
  broad formatting diff entered the refactoring pull requests.

## Definition of Done

- Dependabot covers Go modules, npm, Docker, and GitHub Actions.
- CodeQL analyzes Go and JavaScript/TypeScript on pull requests, main, and schedule.
- Weekly dependency audits cover reachable Go vulnerabilities and high-severity runtime npm issues.
- Root `SECURITY.md` provides a private reporting route and documents supported versions.
- Go lint, backend tests, frontend coverage, and frontend build are required pull-request checks.
- Critical frontend workflows have behavioral tests before their implementation is moved.
- Route authorization is explicit metadata with negative-path regression tests.
- `VehiclesView.tsx`, `router.go`, `article_search.go`, and `vehicles.go` meet the structural targets
  without changing public behavior.

## Reference Versions

- CodeQL workflow syntax: `github/codeql-action/*@v4`.
- golangci-lint action: `golangci/golangci-lint-action@v9` with `v2.11.4`.
- govulncheck: `golang.org/x/vuln/cmd/govulncheck@v1.6.0`.
- Trivy action: `aquasecurity/trivy-action@v0.36.0`.
- Dependabot ecosystems: `gomod`, `npm`, `docker`, and `github-actions`.
