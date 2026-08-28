# Architecture

RailKeeper is a small modular monolith. It is deployed as one process, but the code is separated by
responsibility.

## Boundaries

- `api`: HTTP transport, request validation, response mapping
- `application`: use cases, transactions, authorization decisions
- `domain`: vehicle inventory model and domain rules
- `infrastructure`: SQLite, migrations, seed loading and filesystem adapters

## API Contract

`openapi/railkeeper.yaml` is the public contract. The frontend currently uses a small hand-written adapter
in `frontend/src/shared/api.ts`; generated types can replace it later once the API stops changing rapidly.

## Runtime

The production runtime is a Go binary that serves:

- `/api/v1/*` for JSON APIs
- `/health` for container health checks
- static frontend files from `RAILKEEPER_STATIC_DIR`

Node.js is only used to build the frontend.

## Persistent storage and startup safety

- Windows Standalone resolves its default data directory to `%LOCALAPPDATA%\RailKeeper\data`.
- Docker keeps the fixed `/data` volume. Explicit `RAILKEEPER_DATA_DIR` configuration takes
  precedence and disables automatic legacy-directory migration.
- A legacy Windows `data` directory beside the executable is copied into a verified sibling staging
  directory. SQLite state is captured through a consistent snapshot, ordinary files are compared by
  path, type, size, and SHA-256, and the complete staging directory is promoted atomically. The
  source is never modified or removed.
- Two different databases produce a typed conflict outcome. Only a loopback safety page starts;
  migrations, seeds, file services, authentication routes, and the normal frontend remain inactive.
- Pending schema migrations on an existing database require a validated SQLite snapshot under
  `safety-backups` before any migration statement runs.
- The Admin storage API exposes only the already resolved path and capabilities. It never accepts a
  filesystem path from the browser.

## Scope Decisions

- Vehicles are the core inventory aggregate. A vehicle set owns shared catalogue and acquisition data,
  while its ordered members remain ordinary vehicles with their own inventory numbers, technical data,
  maintenance, CV values, uploads, and exhibition assignments.
- Accessories are a separate catalogue and inventory area. Quantity stock and individually tracked assets
  share storage locations, reservations, installations, condition, and history workflows.
- Private and club layouts contain modules, segments, baseboards, areas, setup configurations, plan
  variants, and immutable published plan revisions.
- The Planner role owns layout structure and plan publication. Editors manage inventory and physical
  installations. Viewers have read access. Messe remains isolated from general inventory and layout APIs.
- The Stage 1 layout workspace uses structured forms. Graphical track planning and digital control are not
  part of this stage.
- Article data web search is a core module. It already uses an adapter boundary so provider-specific logic
  can be replaced later. It remains a synchronous request with one truthful in-progress state; simulated
  phases and polling are not part of the current contract. See `docs/architecture/article-search-progress.md`.
- Intellibox 3 keeps Z21-compatible UDP and LocoNet over TCP as separate transports. Only the bounded,
  read-only Z21 connection test and diagnostics are available today. See
  `docs/architecture/intellibox3-transports.md` for the validation status and hardware test boundary.
- Z21 locomotive reads are targeted status polls for decoder addresses already known to RailKeeper, not a
  roster import. Only the confirmed address enters the comparison workspace; driving and function state is
  discarded. See `docs/architecture/z21-read-scope.md` for protocol evidence, bounds, and hardware gaps.
- SQLite remains the default database because it keeps local installation, backup, and restore simple.
- Attachments are stored on the filesystem below the configured data directory; metadata stays in SQLite.
- Backup version 20 covers vehicle sets and their main-image assignments, vehicles, accessories, storage,
  layouts, plan revisions, setups, reservations, installations, exhibitions, transfer profiles, conflict
  exceptions, and uploads. Older supported versions remain importable; tables introduced later are
  restored empty. Restore defers foreign-key checks inside its transaction so self-references can be
  replaced atomically and are still validated at commit.
- Backup and restore intentionally exclude local authentication and installation data such as users,
  roles, sessions, password hashes, audit logs, SMTP settings, and user UI settings.
