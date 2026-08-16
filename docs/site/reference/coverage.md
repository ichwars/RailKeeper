---
title: Documentation Coverage
description: See how RailKeeper source surfaces map to the bilingual handbook.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-08-15
---

# Documentation coverage

RailKeeper treats documentation coverage as a reviewed contract. A read-only inventory scans
visible frontend routes, English and German translation keys, registered Go API routes, OpenAPI
operations, tracked configuration variables, and existing project documents. The coverage
validator then requires every frontend route, translation key, API route, and environment variable
to have one owning topic.

The reviewed assignments live in
[`docs/coverage.json`](https://github.com/ichwars/RailKeeper/blob/main/docs/coverage.json). The
inventory is generated during validation and is not committed as a second source of truth.

## Coverage states

- `planned`: The destination and owner are agreed, but the complete page pair belongs to a later
  documentation stage.
- `documented`: Both English and German destination pages exist and pass the page standard.
- `internal`: The topic is intentionally limited to maintainer-facing material.
- `not-published`: The source exists, but it must not be presented as a stable public feature.

A topic can move to `documented` only after both destination pages use identical relative paths,
carry matching review metadata, and pass the documentation check. Layout planning remains
`not-published` because it is not part of the stable user-facing product scope.

## Reviewed topics

The initial matrix contains 21 topics.

### User workflows

- `setup-auth`
- `overview`
- `vehicles-core`
- `vehicle-media`
- `vehicle-maintenance`
- `vehicle-decoder-cv`
- `vehicle-search-spares`
- `accessories`
- `exhibition`
- `import-export`
- `settings-general`

### Administration

- `master-data`
- `users-sessions-security`
- `backup-restore`
- `digital-centers`
- `system-operations`
- `deployment-configuration`

### Development

- `layouts-unpublished`
- `development-architecture`

### Shared reference

- `releases-support`
- `shared-navigation`

## Contributor workflow

When a route, API area, translation namespace, environment variable, or handbook topic changes,
update `docs/coverage.json` in the same pull request. Run the complete check from the repository:

```powershell
cd docs
npm.cmd run check
```

The command verifies both languages, review metadata, coverage ownership, and the production site
build.
