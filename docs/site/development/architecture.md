---
title: Architecture and contributions
description: Architecture and contributions.
audience: developer
status: development
reviewedVersion: main
lastReviewed: 2026-11-07
---

# Architecture and contributions

RailKeeper is a local, self-hosted application: a Go backend, a React/TypeScript interface, SQLite and private file storage. The backend serves the API and built frontend. No cloud account is required.

## Responsibilities

| Area | Repository location | Purpose |
| --- | --- | --- |
| Interface | `frontend/src` | Workflows, previews and translated text |
| Backend | `backend/internal` | Roles, validation, persistence and device adapters |
| Database | `backend/migrations` | Versioned SQLite changes |
| Bundled data | `backend/seeds` | Master data and decoder library |
| API contract | `openapi/railkeeper.yaml` | Public HTTP interface |
| Handbook | `docs/site` | English and German at matching relative paths |

The [architecture reference](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture.md) details module boundaries, startup safety and import decisions. [Contribution guidelines](https://github.com/ichwars/RailKeeper/blob/main/CONTRIBUTING.md) describe the workflow.

## Develop locally

Install the Go/Node versions specified by `backend/go.mod` and CI configuration. Install frontend dependencies with `npm ci`, build using `npm run build`, then start the backend from `backend` with `go run ./cmd/railkeeper`. Use a separate data folder for tests.

Before a pull request: format Go, run `go test ./...` in the backend and `npm run build` in the frontend. Add focused tests for affected workflows. Documentation changes need both languages and `npm run check` in `docs`.

## Check catalog assignments

Admins can read the mapping/difference report using `GET /api/v1/catalog/legacy-backfill`. It lists
missing relationships, pinned article revisions and individual overrides. Reading it changes no
assignment. An explicitly sent `POST` on the same route completes only missing relationships and
requires the existing CSRF protection. Retries add neither catalog rows nor another backfill audit
event. Existing differences are never silently corrected. Matching assignments use the article
revision plus individual overrides with a transactionally maintained legacy vehicle projection.
Unresolved proposals retain legacy authority and their original assignment mode.

The [catalog-model reference](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/vehicle-catalog.md)
documents field mapping, backups and preservation evidence. Viewer, Editor and Messe cannot use
these administrative inspection operations.

Viewer, Editor and Admin can read a saved vehicle/set assignment using
`GET /api/v1/catalog/{targetType}/{id}` (`vehicle` or `set`). Current inventory fields, the pinned
revision and individual overrides remain separate. Its comparison token detects revision changes
even when visible values are identical. Reading adopts no proposals and fills no missing links.

Editor and Admin confirm new revisions using `POST /catalog/{targetType}/{id}/revisions` and adopt
existing revisions using `POST /catalog/{targetType}/{id}/adoption`, both with CSRF protection.
The comparison token and `selectedFields` are required; an empty selection keeps every saved value.
New revisions use a 32-character hexadecimal identity retained on retries. After HTTP 409, reload
and review the saved assignment before confirming again. Do not automatically modify and replay
the request. Existing vehicle/set write calls retain their previous meaning.

## Read saved decoder evidence

Viewer, Editor and Admin can read `GET /api/v1/vehicles/{id}/decoder-configurations` without
changing the collection. Physical decoders, installation episodes, exact working labels and
recorded partial CV states remain separate. Unknown installation dates stay null.

`GET /api/v1/decoder-configurations/snapshots` lists the private archive, optionally filtered by
the historical `vehicleId`. Snapshot details and retained-source downloads use their snapshot ID
and stay readable after vehicle deletion. Unknown historic CV values and individual source dates
are preserved; a recording time does not claim a complete measurement. Missing original content
returns HTTP 410 rather than another file. Messe cannot use these inventory routes.

The [decoder-model reference](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture/decoder-configurations.md)
describes the model and transaction rules. The OpenAPI contract defines each route and response.

Editor and Admin confirm physical instances, installation/removal, date corrections, working
groups, associations and checkpoints under `/api/v1/vehicles/{id}/decoder-configurations`.
Every command requires CSRF protection, the current context revision and all documented edit
fields. Unknown dates and clearing an association use explicit null. New records retain their
request identity on retries. After HTTP 409, reload and review; never generate a replacement
identity and replay automatically. A checkpoint contains saved CVs and excludes open browser
drafts. Earlier snapshots retain their original hardware descriptions and vehicle identities.

CV writers and source uploads can select an exact working group with `workspaceId` and its
`bindingRevision` as `expectedWorkspaceRevision`. Keep the existing per-row comparison as well.
The server preserves the group's original profile text, including legacy whitespace, and rejects
a changed hardware association/state. Requests without this optional scope keep their existing
profile-cleaning semantics. Source uploads use the same pair as multipart fields.

CV and source-file deletions accept the same paired query fields together with
`expectedRevision`; manual batch deletions carry them on the outer deletion item. A changed
hardware binding rejects the whole operation even when the row itself is unchanged. The exact
stored profile must match the selected group. Legacy requests without scope remain compatible.

## Preserve the boundaries

Roles and protections are enforced by the server. Device data starts as read-only suggestions; writes require an explicitly confirmed preview. CV values remain separate from their meaning: library rules interpret raw data only with documented model, firmware and index context.

Do not include real credentials, databases or collection pictures in tests and screenshots. Local diagnostic/test variables may support bug reports; they are not general user settings. See the [security policy](https://github.com/ichwars/RailKeeper/blob/main/SECURITY.md).
