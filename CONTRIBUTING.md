# Contributing to RailKeeper

Thank you for improving RailKeeper. Keep changes focused, reviewable, and aligned with its
local-first, self-hosted architecture.

## Before You Start

- Use Go and Node.js versions compatible with `backend/go.mod`, `frontend/package.json`, and the
  current Dockerfile.
- Open an issue or discussion before starting broad architecture, storage, security, or UI-system
  changes.
- Report vulnerabilities privately as described in [SECURITY.md](SECURITY.md).
- Do not commit secrets, local credentials, runtime databases, backups, uploaded data,
  `frontend/dist`, `frontend/node_modules`, or `.cache`.

## Development Checks

Backend:

```powershell
Set-Location backend
go test ./...
```

Frontend:

```powershell
Set-Location frontend
npm.cmd install
npm.cmd run test:run
npm.cmd run build
```

Run `gofmt` after Go changes. Keep backend routes, frontend API types, and
`openapi/railkeeper.yaml` aligned when the API changes. Update German and English translations
together.

## Pull Requests

Explain the problem, the chosen solution, and the verification performed. Include screenshots for
visible UI changes and call out security, migration, backup, restore, or compatibility effects.

By submitting a contribution, you agree to license it under AGPL-3.0-only and confirm that you
have the necessary rights to do so. No copyright assignment or CLA is required.
