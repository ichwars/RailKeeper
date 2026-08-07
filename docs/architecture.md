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

## Scope Decisions

- Vehicles are the core inventory aggregate.
- Accessories are a separate catalogue and inventory area. Quantity stock and individually tracked assets
  share storage locations, reservations, installations, condition, and history workflows.
- Private and club layouts contain modules, segments, baseboards, areas, setup configurations, plan
  variants, and immutable published plan revisions.
- The Planner role owns layout structure and plan publication. Editors manage inventory and physical
  installations. Viewers have read access. Messe remains isolated from general inventory and layout APIs.
- The Stage 1 layout workspace uses structured forms. Graphical track planning and digital control are not
  part of this stage.
- Article data web search is a core module. It already uses an adapter boundary so provider-specific logic
  can be replaced later.
- SQLite remains the default database because it keeps local installation, backup, and restore simple.
- Attachments are stored on the filesystem below the configured data directory; metadata stays in SQLite.
- Backup version 2 covers vehicles, accessories, storage, layouts, plan revisions, setups, reservations,
  installations, exhibitions, and uploads. Version 1 remains importable with the Stage 1 tables restored
  empty. Restore defers foreign-key checks inside its transaction so self-references can be replaced
  atomically and are still validated at commit.
- Backup and restore intentionally exclude local authentication and installation data such as users,
  roles, sessions, password hashes, audit logs, SMTP settings, and user UI settings.
