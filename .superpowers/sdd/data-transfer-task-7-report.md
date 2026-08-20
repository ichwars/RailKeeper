# Data Transfer Task 7 Report

## Status

Implemented dashboard summary, scoped and filtered job history, job details, safe terminal-job retry,
profile-only backup coverage, and the complete public data-transfer contract. Messe queries, details,
retry, artifact counts, and latest-export metadata remain restricted to exhibition-list-only jobs.

Retry creates a new unapproved draft. It copies only the transfer selection and profile name, uses
the current package version and actor, and does not carry the original profile ID, source, preview,
record counters, issues, confirmation, completion result, or artifacts.

Backups expose an optional top-level `profiles` collection. They do not include transfer jobs,
issues, or artifact content. Missing creator users are normalized to `null` during export and again
during restore. An older backup without the profile section preserves local transfer profiles.

## TDD Evidence

### RED 1: summary and retry service

Command:

```powershell
go test ./internal/application -run 'Test(DataTransferSummary|Retry|BackupRoundTripRestoresDataTransferProfiles|BackupRestoreWithoutProfiles)'
```

Expected compile failures observed:

```text
service.Summary undefined
service.RetryJob undefined
FAIL railkeeper/backend/internal/application [build failed]
```

### RED 2: query and retry routes

Command:

```powershell
go test ./internal/api -run TestDataTransferQueryAndRetryRoutesReturnScopedPersistentHistory
```

Expected route failure observed:

```text
got status 404, want 200: {"error":"not_found","message":"API route not found"}
```

### RED 3: creator normalization

Command:

```powershell
go test ./internal/application -run TestBackupRoundTripRestoresDataTransferProfilesWithoutOperationalHistory
```

Expected behavioral failure observed:

```text
exported missing local creator = "missing-user", want nil
```

### RED 4: bounded history limit

Command:

```powershell
go test ./internal/api -run TestDataTransferQueryAndRetryRoutesReturnScopedPersistentHistory
```

Expected contract failure observed:

```text
got status 200, want 400
```

An explicit `limit=0` no longer bypasses the documented 1 to 200 history bound.

### GREEN: focused verification

Commands:

```powershell
go test ./internal/application -run 'Test(DataTransferSummary|Retry|BackupRoundTripRestoresDataTransferProfiles|BackupRestoreWithoutProfiles)'
go test ./internal/api -run TestDataTransferQueryAndRetryRoutesReturnScopedPersistentHistory
go test ./internal/api -run 'TestOpenAPI|TestDataTransfer'
```

Result: all focused tests passed.

### GREEN: full backend verification

Command:

```powershell
go test ./...
```

Result: all backend packages passed. The first full run correctly detected the registered summary
route before it existed in OpenAPI. After completing the contract, the full suite passed.

## Independent Review Follow-up

The completion review found and this task fixed three contract issues:

- Current backups now always serialize `profiles`, including `"profiles":[]`. This distinguishes a
  current empty snapshot, which clears destination profiles, from a legacy backup without the field,
  which preserves local profiles.
- Backup preflight now validates profile columns, required values, duplicate IDs, direction, format,
  area selection, CSV constraints, options JSON, enabled state, nullable creator/last-use values, and
  timestamps before reporting compatibility.
- OpenAPI 3.1 expresses issue `rowNumber` as `type: [integer, "null"]`.

Focused review-fix verification and the subsequent full backend run both passed.

## Changed Files

- `backend/internal/application/data_transfer_profiles.go`
- `backend/internal/application/data_transfer_profiles_test.go`
- `backend/internal/application/data_transfer_types.go`
- `backend/internal/application/backup.go`
- `backend/internal/application/backup_test.go`
- `backend/internal/infrastructure/data_transfer_repository.go`
- `backend/internal/api/data_transfer_handlers.go`
- `backend/internal/api/data_transfer_handlers_test.go`
- `backend/internal/api/routes.go`
- `openapi/railkeeper.yaml`
- `.superpowers/sdd/data-transfer-task-7-report.md`

## Implementation Notes

- Open jobs are draft, reading, review-required, ready, or running jobs. Selected records are the
  sum of their persisted record counts.
- Deleted artifacts are retained in job details as history but excluded from summary count and
  byte totals.
- History supports `profileId`, `direction`, repeated or comma-separated `states`, and a 1 to 200
  `limit`, defaulting to 100.
- Scoped history is filtered before applying the limit, so hidden jobs cannot reduce Messe result
  counts.
- No migration was added. Reserved migration 0062 remains untouched.
- Transfer areas are exactly `vehicles`, `accessories`, and `exhibitionLists`. Feature transfer
  packages continue to exclude master data and authentication/security tables.

## Concerns

- The summary and detail queries enumerate persisted artifacts and filter them in the application
  layer. This is appropriate for the current local-first scale, but a repository-level scoped
  aggregate may be preferable if histories become very large.
- Retry intentionally accepts only terminal states. An operator must cancel or complete an open job
  before creating a new attempt.
- Existing unrelated dirty frontend and `facebook-*` files were preserved and excluded from the
  task commit.

## Preview Enum Contract Follow-up

An independent contract review found that `DataTransferPreviewRecord.proposedAction` can emit
`create`, `replace`, `use_existing`, and `copy`, while OpenAPI documented only `create`. It also
found that issue suggestions can emit `replace_or_copy`, which is not itself a selectable
resolution.

The contract now separates these concepts:

- `TransferProposedAction` documents all four emitted preview actions.
- `TransferProposedResolution` documents suggestions, including `replace_or_copy`.
- `TransferIssueResolution` remains the exact selectable request enum and excludes
  `replace_or_copy`.

RED evidence:

```text
TestOpenAPIDataTransferPreviewEnumsMatchRuntimeValues:
OpenAPI block TransferProposedAction is missing
```

GREEN verification:

```powershell
go test ./internal/api -run TestOpenAPIDataTransferPreviewEnumsMatchRuntimeValues
go test ./internal/application -run DataTransfer
go test ./...
```

The semantic regression test marshals representative runtime preview records and issues, then
checks their emitted values against the dedicated OpenAPI enums. It also asserts that
`replace_or_copy` cannot enter the selectable resolution schema.
