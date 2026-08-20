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

## Required Data-Safety Follow-up

Reviewer findings were addressed in a follow-up fix commit.

### Aggregate stale protection

Each previewed replacement now persists a deterministic SHA-256 fingerprint of the complete
transferable target aggregate. Vehicles include the full transferable vehicle row. Accessories
include the product, ordered stock with location identity and quantities, and ordered assets with
their transferable fields and IDs. Exhibition lists include the list and all ordered entries.

Apply recomputes these fingerprints from one transaction-local snapshot before any write. Parent
timestamps remain informational but are no longer the safety boundary. Same-second mutations to a
vehicle parent, accessory stock child, or exhibition entry all return the normal data-transfer
conflict and leave the job ready.

RED evidence:

```text
TestDataTransferApplyRejectsSameSecondVehicleMutation: ApplyImport() error = <nil>, want conflict
TestDataTransferApplyRejectsSameSecondAccessoryChildMutation: ApplyImport() error = <nil>, want conflict
TestDataTransferApplyRejectsSameSecondExhibitionEntryMutation: ApplyImport() error = <nil>, want conflict
```

### Provenance-safe accessory merge

Accessory replacement now upserts imported stock without deleting unmatched local stock. Imported
assets match only a local asset on the same product, first by local ID and then by inventory number.
Matched assets retain their local ID and `purchase_id`; only transferable fields change. Unmatched
imports receive new local IDs, while unmatched local assets remain untouched. Active reservation
and installation invariants are checked before changing lifecycle or storage state.

RED evidence:

```text
TestDataTransferApplyAccessoryReplacePreservesPurchaseAndRelationships:
apply accessories record "RK-ART-MERGE": FOREIGN KEY constraint failed
```

The extended invariant test also proved that moving a reserved asset to a different location was
previously accepted and is now rejected with full rollback. A second RED case proved that an
unmatched imported asset could previously enter the local database as Installed without a local
installation. Reserved and Installed lifecycle states now require a matching local relationship.

### Exact deterministic resolution identity

JSON preview records receive stable one-based ordinal row numbers. Issue IDs are deterministic
SHA-256 values over job, area, record key, row number, exact field/entry key, and issue code. Apply
matches record issues by `(area, recordKey, rowNumber)`. Exhibition references use the explicit
`entries[n].vehicleReference` key, never query order or random issue IDs.

RED evidence:

```text
TestDataTransferPreviewIssuesHaveDeterministicRecordAndEntryIdentity:
JSON record rows are not deterministic

TestDataTransferApplyUsesEachExhibitionReferenceResolutionOnce:
resolved exhibition vehicle IDs = ["", "vehicle-b"], want ["vehicle-a", ""]

TestDataTransferApplyBindsDuplicateRecordResolutionsToRowNumber:
row-scoped resolutions created 1 vehicles, want original plus one copy
```

### Follow-up GREEN verification

```powershell
go test ./internal/application ./internal/infrastructure ./internal/api -run DataTransfer -count=1
go test ./internal/infrastructure -run 'DataTransferApply(CAS|RejectsSameSecond|AccessoryReplace|UsesEach|BindsDuplicate)' -count=10
go test ./... -count=1
```

All commands passed. OpenAPI now documents `targetFingerprint` on persisted preview records. No new
database column was necessary because preview records already persist atomically in `preview_json`.

### Follow-up concerns

- Replacement intentionally preserves unmatched local accessory stock and assets because they may
  carry local provenance or relationships absent from a transfer package. It is a conservative
  merge, not destructive reconciliation.
- Imports that conflict with active reservation or installation lifecycle/location truth are
  rejected. Operators must resolve the local relationship or choose a non-replacement resolution.

## Reservation and Strategy Safety Follow-up

The accessory aggregate fingerprint now also covers deterministic local allocation state:
reservations, installations, and installation-condition history. This state is marked `json:"-"`
on transfer accessories, so it contributes to target stale detection without entering export or
import payloads. A reservation added after preview therefore invalidates apply even when product,
stock, and timestamp values did not change.

Accessory replacement now performs these checks in the apply transaction before updating the
product or stock rows:

- The imported strategy and tracking mode must be consistent.
- The existing repository strategy-transition invariant rejects dropping quantity tracking while
  stock, active quantity reservations, or active quantity installations remain.
- The same invariant rejects dropping individual tracking while local assets or their allocation
  relationships remain.
- Every imported stock level must remain at or above the active quantity reserved at that local
  storage location, matching normal stock-adjustment behavior.

RED evidence before implementation:

```text
TestDataTransferApplyRejectsReservationAddedAfterPreview:
ApplyImport() error = <nil>, want reservation fingerprint conflict

TestDataTransferApplyRejectsImportedQuantityBelowActiveReservation:
ApplyImport() error = <nil>, want reserved stock conflict

TestDataTransferApplyRejectsIncompatibleAccessoryStrategyTransitions:
all six quantity, individual, and hybrid transition cases returned <nil>
```

GREEN verification:

```powershell
go test ./internal/application ./internal/infrastructure ./internal/api -run DataTransfer -count=1
go test ./internal/infrastructure -run 'TestDataTransferApply(CASAllowsOneConcurrentConfirmation|RejectsSameSecond|RejectsReservationAddedAfterPreview|RejectsImportedQuantityBelowActiveReservation|RejectsIncompatibleAccessoryStrategyTransitions|AccessoryReplacePreservesPurchaseAndRelationships)' -count=10
go test ./... -count=1
```

All commands passed. The complete backend test run covered every package. No OpenAPI change is
needed because the allocation fingerprint state is deliberately absent from serialized transfer
records; only the existing `targetFingerprint` digest remains exposed.

### Reservation and strategy concerns

- Accessory replacement remains a conservative merge. Stock locations or assets absent from the
  import are preserved, including their local relationships and provenance.
- A concurrent allocation mutation makes the preview stale. The operator must preview again rather
  than applying a decision made against older reservation or installation state.

## Pure Quantity Asset Safety Follow-up

Imports now reject any non-empty individual asset collection when the resulting inventory strategy
is pure `quantity`. This validation runs immediately after strategy/tracking-mode validation and
before resolving locations or writing accessory product, stock, or asset rows. Hybrid
`quantity_later_individual` imports retain their intended support for both inventory forms.

RED evidence before implementation:

```text
TestDataTransferApplyRejectsIndividualAssetWhenCreatingQuantityAccessory:
ApplyImport() error = <nil>, want quantity asset conflict

TestDataTransferApplyRejectsUnmatchedIndividualAssetOnQuantityReplacement:
ApplyImport() error = <nil>, want quantity asset conflict
```

Both regression tests verify full rollback. Create leaves product, asset, and would-be storage
location tables untouched. Quantity-to-quantity replacement preserves the original product and
does not insert the unmatched imported asset.

GREEN verification:

```powershell
go test ./internal/application ./internal/infrastructure ./internal/api -run DataTransfer -count=1
go test ./internal/infrastructure -run 'TestDataTransferApply(CASAllowsOneConcurrentConfirmation|RejectsSameSecond|RejectsReservationAddedAfterPreview|RejectsImportedQuantityBelowActiveReservation|RejectsIncompatibleAccessoryStrategyTransitions|RejectsIndividualAssetWhenCreatingQuantityAccessory|RejectsUnmatchedIndividualAssetOnQuantityReplacement|AccessoryReplacePreservesPurchaseAndRelationships)' -count=10
go test ./... -count=1
```

All commands passed. No API, persistence schema, or OpenAPI contract change was required.
