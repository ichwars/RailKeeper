# Tillig TT Track Planner Package C Implementation Plan

**Goal:** Add deterministic revision comparison and confirmed plan-object material reservation.

**Architecture:** Stable object lineage enables derived revision diffs. Existing accessory
reservations remain authoritative; a dedicated link table adds plan-object scope and active-link
uniqueness. Reservation batches are transactional and explicit.

## Global constraints

- Changes remain local on `dev/issue-32-tillig-track-planner`.
- No automatic reservation, installation, publication, or native browser dialog.
- Existing backups 1 through 5 remain importable.
- Viewer, Editor, Planner, and Admin may read previews; only Planner and Admin may reserve.

### Task 1: Lineage and revision diff domain

- [x] Add migration 0046 with `plan_track_objects.lineage_id` and reservation-link table.
- [x] Preserve lineage when cloning revisions and initialize it for new objects.
- [x] Write RED tests for added, removed, changed, unchanged, material delta, issue delta, and no-base
  revisions.
- [x] Implement pure deterministic diff types and helpers in the domain/application boundary.

### Task 2: Change-preview repository, API, and client

- [x] Add repository reads for base plans and affected configurations.
- [x] Add `GET /plan-revisions/{id}/track-change-preview` with role and missing-revision tests.
- [x] Document all schemas in OpenAPI and add the typed client method.
- [x] Build a compact preview panel with object, material, warning, and configuration summaries.

### Task 3: Transactional plan reservations

- [x] Add RED repository tests for quantity, individual asset, insufficient stock, stale object,
  mismatched product, duplicate active link, cancellation, and all-or-nothing batches.
- [x] Implement the batch service and `POST /plan-revisions/{id}/track-reservations`.
- [x] Return refreshed material and active plan-reservation status.
- [x] Extend backup/restore to format 6 with version 5 compatibility.

### Task 4: Confirmed reservation UI

- [x] Add a RailKeeper dialog that loads eligible products, storage sources, and optional assets.
- [x] Require explicit confirmation and show resulting free/missing quantities.
- [x] Refresh analysis and preview after success; keep errors and selection intact on failure.
- [x] Cover Planner/Admin write access and Viewer/Editor read-only behavior.

### Task 5: Acceptance

- [x] Update Stage-3 status and package documentation.
- [x] Run all Go tests, all frontend tests, and the production build.
- [x] Rebuild the local server and verify diff, confirmation, reservation, cancellation status, and an
  empty browser console.
