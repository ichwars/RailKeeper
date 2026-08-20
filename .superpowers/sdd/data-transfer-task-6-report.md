# Data Transfer Task 6 Report

## Status

Complete. Approved import previews now apply inside one SQLite transaction across vehicles,
accessories, accessory stock/assets and exhibition lists/entries. The same transaction completes the
revisioned job with compare-and-swap protection and writes one audit event. Any stale target,
changed source hash, altered/incomplete preview, unresolved issue, invalid resolution, locked-list
replacement, row write, job transition or audit failure rolls back the complete import.

The confirm endpoint requires an explicit JSON boolean set to `true`. It returns the existing data
transfer conflict response (HTTP 409) for stale previews and unresolved conflicts. Messe users stay
restricted to exhibition-list-only jobs even when Viewer or Planner is a secondary role. Editor and
Admin remain unrestricted.

## TDD Evidence

### RED 1: repository apply boundary

Command:

```powershell
go test ./internal/infrastructure -run DataTransferApply -count=1
```

Expected failure observed:

```text
repository.ApplyImport undefined
FAIL railkeeper/backend/internal/infrastructure [build failed]
```

The initial tests covered two-row success, audit creation, a constraint failure on row two with no
partial vehicle write, stale `updated_at`, a changed source hash, unresolved issues and a locked
exhibition-list replacement.

### RED 2: application confirmation

Command:

```powershell
go test ./internal/application -run DataTransferConfirmImport -count=1
```

Expected failure observed:

```text
service.ConfirmImport undefined
FAIL railkeeper/backend/internal/application [build failed]
```

### RED 3: HTTP confirmation route

Command:

```powershell
go test ./internal/api -run DataTransferConfirmRoute -count=1
```

Expected failure observed:

```text
got status 404, want 400: {"error":"not_found","message":"API route not found"}
```

### RED 4: per-entry exhibition resolutions

Command:

```powershell
go test ./internal/infrastructure -run DataTransferApplyUsesEach -count=1
```

Expected behavioral failure observed:

```text
resolved exhibition vehicle IDs = []string{"vehicle-a", "vehicle-b"}, want [vehicle-a empty]
```

The fix consumes each persisted link/skip resolution exactly once in preview order.

### RED 5: accessory copy safety

Command:

```powershell
go test ./internal/infrastructure -run DataTransferApplyCopiesAccessory -count=1
```

Expected constraint failure observed:

```text
UNIQUE constraint failed: accessory_assets.inventory_number
```

Copied accessory products now allocate collision-free inventory numbers for both the product and
its individual assets.

### GREEN: focused and concurrency verification

Commands:

```powershell
go test ./internal/application ./internal/infrastructure ./internal/api -run DataTransfer -count=1
go test ./internal/infrastructure -run DataTransferApplyCAS -count=20
```

Result: all packages passed. Twenty repeated concurrent confirmation cases each produced exactly
one committed import and one conflict, with one vehicle and one audit event.

### GREEN: full backend verification

Command:

```powershell
go test ./...
```

Result: all backend packages passed.

The first full run correctly exposed that the new registered route was absent from the OpenAPI
contract. The contract schema and route were added, then the complete suite passed.

## Changed Files

- `backend/internal/infrastructure/data_transfer_apply.go`
- `backend/internal/infrastructure/data_transfer_apply_test.go`
- `backend/internal/application/data_transfer_import.go`
- `backend/internal/application/data_transfer_import_test.go`
- `backend/internal/api/data_transfer_handlers.go`
- `backend/internal/api/data_transfer_handlers_test.go`
- `backend/internal/api/routes.go`
- `openapi/railkeeper.yaml`
- `.superpowers/sdd/data-transfer-task-6-report.md`

## Implementation Notes

- Persisted previews carry the source SHA-256 used to create them. Apply compares that value with
  the revisioned job source hash before any target write.
- Target replacements revalidate the exact previewed `updated_at` value in the apply transaction.
- Record resolutions support replace, copy, skip, use-existing, create and exhibition vehicle
  link/skip behavior. Locked exhibition lists can only be copied or skipped.
- Copy operations allocate local collision-free inventory numbers. Imported database IDs are not
  trusted for new records.
- Accessory stock locations resolve by local ID or root location name and are created transactionally
  when absent. No RailKeeper master-data tables are imported or modified.
- The audit event stores only job-level provenance and record counts, not imported record payloads.

## Concerns

- Replacing an accessory intentionally replaces its imported stock and asset children. Existing
  local references can make SQLite reject such a replacement; that rejection safely rolls back the
  whole import and leaves the job ready for review.
- This task implements backend apply and contract behavior only. The unrelated dirty frontend and
  `facebook-*` files were preserved and are not part of the commit.
