# Architecture

RailKeeper is a small modular monolith. It is deployed as one process, but the code is separated by responsibility.

## Boundaries

- `api`: HTTP transport, request validation, response mapping
- `application`: use cases, transactions, authorization decisions
- `domain`: vehicle inventory model and domain rules
- `infrastructure`: SQLite, migrations, seed loading and filesystem adapters

## API Contract

`openapi/railkeeper.yaml` is the public contract. The frontend currently uses a small hand-written adapter in `frontend/src/shared/api.ts`; generated types can replace it later once the API stops changing rapidly.

## Runtime

The production runtime is a Go binary that serves:

- `/api/v1/*` for JSON APIs
- `/health` for container health checks
- static frontend files from `RAILKEEPER_STATIC_DIR`

Node.js is only used to build the frontend.

## Scope Decisions

- Vehicles are the core inventory aggregate.
- Accessories are intentionally excluded from the MVP.
- Article data web search is a core module. It already uses an adapter boundary so provider-specific logic can be replaced later.
- SQLite remains the default database because it keeps local installation, backup, and restore simple.
- Attachments are stored on the filesystem below the configured data directory; metadata stays in SQLite.
- Backup and restore are application-level use cases: they export inventory data and uploads while intentionally excluding local authentication data such as users, sessions and password hashes.
