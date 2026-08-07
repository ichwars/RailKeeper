# Stage 1 completion evidence

## Completed work packages

- #37 merged and closed.
- #38 merged and closed.
- #40 merged and closed through PRs #49, #50, and #51.
- #39 application contract merged through PR #52.
- #39 transactional SQLite persistence merged through PR #53.
- #39 HTTP, authorization, CSRF, and OpenAPI surface merged through PR #54.
- #41 typed frontend client merged through PR #55.
- #42 product, storage-location, stock, and individual-asset workspace merged through PR #56.
- #42 reservation, installation, condition, removal, and history workspace merged through PR #57.
- #43 layout workspace merged through PR #58.

## Current verification evidence

- `go test ./...`: passed after PR #52 changes.
- `go vet ./...`: passed after PR #52 changes.
- GitHub CI run 307: passed.
- GitHub CodeQL run 108: passed.
- GitHub Trivy run 94: passed.

Further evidence is appended after each merged slice.

## #44 backup version 2 and Stage 1 acceptance

- Backup exports now use version 2 and require every Stage 1 accessory, layout, plan, setup,
  reservation, and installation table.
- Version 1 validation and import accept missing Stage 1 tables and restore them empty.
- Version 2 roundtrip test preserves storage hierarchy, stock, individual assets, layouts, modules,
  plan revision bases, published configuration references, vehicle reservations, and layout-unit
  installations.
- Restore defers SQLite foreign keys inside the existing transaction. Self-references can be
  replaced in any row order and are still enforced at commit.
- Auth, role, session, password, audit, rate-limit, app-setting, and user-setting tables remain
  excluded.
- API smoke matrix covers Admin, Editor, Viewer, Planner, and Messe across vehicle, accessory,
  allocation, storage, and layout categories. Negative Planner-install and Editor-publish checks
  pass.
- `go test ./...`: passed.
- `go vet ./...` with repository-local `GOCACHE`: passed.
- `npm.cmd run test:run -- --reporter=dot`: passed, 25 files and 88 tests.
- `npm.cmd run build`: passed.
- `git diff --check`: passed.
- Real HTTP acceptance on isolated port 18081 exported version 2 with 29 tables and no excluded
  local-auth/settings tables, validated it as compatible, changed the layout, restored 1,338 rows,
  recovered the original layout name, and retained the active admin session.
- Manual role UI acceptance: Editor and Viewer can read layouts but cannot create them; Planner can
  create layouts; Messe sees only the exhibition navigation and cannot reach inventory or layouts.
- Acceptance-discovered repair: pure Messe sessions no longer request general inventory master data.
  They load only permitted symbols, eliminating the misleading role error while preserving isolation.
- Stable documentation now records the Stage 1 architecture, backup v2, and Stage 2 graphical-planner
  direction with Tillig TT Modellgleis first and digital control explicitly out of scope.

## #43 layout workspace UI

- `npm.cmd run test:run -- --reporter=dot`: passed, 24 files and 87 tests.
- Focused layout suite: passed, 8 tests.
- `npm.cmd run build`: passed.
- `git diff --check`: passed.
- Browser QA used the isolated RailKeeper instance on port 18081.
- German and English navigation, labels, forms, status values, and deferred-stage messages: passed.
- Dark theme and mobile viewport 390 x 844: passed without document-level horizontal overflow.
- Interaction proof: created a club layout and module, created a plan variant and draft, submitted it
  for review, published it through the explicit confirmation dialog, and created a setup with module
  position data.
- Browser log after the completed interaction flow: empty.
- Regression repaired: plan loading no longer repeats because the callback depends on the stable
  localized error string instead of the unstable translation function identity.
- Canonical owner: `LayoutPlansPanel`; no fallback or duplicate loading path remains.

## #42 allocation UI slice

- `npm.cmd run test:run -- --reporter=dot`: passed, 23 files and 79 tests.
- `npm.cmd run build`: passed.
- `git diff --check`: passed.
- Browser QA used an isolated temporary RailKeeper data directory on port 18081.
- German and English UI: passed for navigation, labels, empty states, product catalogue, stock,
  reservations, and installation history.
- Light and dark themes: passed.
- Mobile viewport 390 x 844: passed without document-level horizontal overflow.
- Interaction proof: created a storage location, created a Tillig quantity product, booked five
  units through the confirmation dialog, and verified summary/location totals.
- Browser console: no warnings or errors.

## #39 SQLite persistence slice

- Focused allocation tests: passed for quantity reservation, over-reservation, direct installation,
  individual assets, cancellation, all removal states, target mismatch rollback, history, and audit
  privacy.
- Migration preservation test: passed from schema 0039 to 0040 with legacy reservation retained.
- `go test ./...`: passed.
- `go vet ./...`: passed.
- `git diff --check`: passed.
- `go test -race`: not available locally because CGO is disabled; no race result is claimed.
- CI failure in run 309: `ineffassign` found an overwritten lifecycle initializer in the
  installation repository.
- Canonical-owner repair: declare the lifecycle without an unused default, preserving the existing
  exhaustive disposition switch.
- Repair verification: full `go test ./...`, `go vet ./...`, `git diff --check`, CI run 310,
  CodeQL run 111, and Trivy run 97 passed.
- Retirement: no fallback or compatibility path was added; no removal track remains.
