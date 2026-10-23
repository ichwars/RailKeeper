---
title: Architecture and contributions
description: Architecture and contributions.
audience: developer
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
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

## Preserve the boundaries

Roles and protections are enforced by the server. Device data starts as read-only suggestions; writes require an explicitly confirmed preview. CV values remain separate from their meaning: library rules interpret raw data only with documented model, firmware and index context.

Do not include real credentials, databases or collection pictures in tests and screenshots. Local diagnostic/test variables may support bug reports; they are not general user settings. See the [security policy](https://github.com/ichwars/RailKeeper/blob/main/SECURITY.md).
