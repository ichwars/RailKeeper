# Data Transfer Task 8 Report

## Status

Implemented the strict frontend data-transfer contract and a rendering-free workspace hook. The API
client now covers summary, profile lifecycle, filtered job history and details, import preview and
resolution, export execution, retry, artifact download/deletion, and the capability-gated local
folder action.

The workspace loads dashboard data, selects the first open job, owns filters, selection, dialogs,
mutation state and refreshes, and exposes named actions. A pure Messe role scope is defensively
restricted to exhibition-list profiles, jobs, and composition areas. Admin or Editor roles retain
the complete vehicles, accessories, and exhibition-lists scope. Session roles flow from App through
ImportExportView into the workspace hook.

## TDD Evidence

### RED

Command:

```powershell
cd frontend
npm.cmd run test:run -- src/features/importExport/useDataTransferWorkspace.test.tsx src/app/App.test.tsx
```

Observed expected failures before implementation:

```text
Failed to resolve import "./useDataTransferWorkspace"
TypeError: Cannot read properties of undefined (reading 'join')
```

The failures proved that the workspace hook did not exist and App did not pass session roles to the
import/export workspace.

### GREEN: focused task verification

Command:

```powershell
npm.cmd run test:run -- src/features/importExport/useDataTransferWorkspace.test.tsx src/app/App.test.tsx
```

Result: 2 test files and 8 tests passed. Coverage includes initial dashboard loading and open-job
selection, Messe scope, filters and detail selection, named export/folder actions, every documented
API route shape, multipart upload, artifact URL encoding, and App role propagation.

### GREEN: existing import/export regression suite

Command:

```powershell
npm.cmd run test:run -- src/features/importExport/ImportExportView.test.tsx src/features/importExport/importExportHelpers.test.tsx
```

Result: 2 test files and 18 tests passed.

### GREEN: production build

Command:

```powershell
npm.cmd run build
```

Result: TypeScript project build and Vite production build passed. The focused task and existing
import/export regression commands cover 26 passing tests in total.

## Changed Files

- `frontend/src/features/importExport/dataTransferModel.ts`
- `frontend/src/features/importExport/useDataTransferWorkspace.ts`
- `frontend/src/features/importExport/useDataTransferWorkspace.test.tsx`
- `frontend/src/features/importExport/ImportExportView.tsx`
- `frontend/src/features/importExport/ImportExportView.test.tsx`
- `frontend/src/shared/api.ts`
- `frontend/src/app/App.tsx` (Task 8 hunks only)
- `frontend/src/app/App.test.tsx`
- `.superpowers/sdd/data-transfer-task-8-report.md`

## Self-review

- Transfer unions and payloads match `openapi/railkeeper.yaml`; no master-data area exists.
- The client uses encoded identifiers, repeated state filters, CSRF-protected mutations, multipart
  upload without a manual content-type, and the documented 120-second import timeout.
- The hook contains no JSX and uses no `any`.
- Folder opening requires both Admin role and the runtime capability returned in the summary.
- Existing dirty Digital Centers prototype files, CSS/i18n changes, and `facebook-*` files remain
  untouched and excluded from the commit.

## Concerns

- `frontend/src/shared/api.ts` imports transfer contract types from the feature model. This is the
  narrow dependency implied by the task file split, but a later shared contract module would remove
  the shared-to-feature type dependency if more consumers emerge.
- The default parallel full-suite invocation did not produce a completion summary and was stopped.
  No unrelated persistent Node process was terminated. Focused Task 8 and import/export verification
  plus the complete production build are the completion evidence for this task.
- Vite emits its existing native config-loader warning and the existing large-chunk warning during
  tests/build; neither causes a failure.

## Reviewer Follow-up

The Task 8 review identified four required state and integration fixes. App no longer asserts the
lazy view's prop type. `ImportExportView` now owns the real `{ roles: string[] }` prop contract and
passes it to `useDataTransferWorkspace` while retaining the prototype layout that Task 9 replaces.

Dashboard and detail requests now use independent monotonic request sequences plus an unmount guard.
An older filter response cannot replace a newer dashboard, and an older detail response cannot
replace a newer selection or repopulate a cleared selection. Detail errors have independent loading
and error state, so summary, profiles, jobs, and the selected job remain usable if a detail vanished.

Profile capabilities are now explicit: Editor and Admin can create or update profiles, only Admin
can disable profiles, and pure Messe scope can compose only exhibition-list transfers.

### Follow-up RED

Command:

```powershell
cd frontend
npm.cmd run test:run -- src/features/importExport/useDataTransferWorkspace.test.tsx `
  src/features/importExport/ImportExportView.test.tsx src/app/App.test.tsx
```

Result: 7 expected failures reproduced missing View-to-hook role propagation, combined profile
capabilities, stale dashboard overwrite, stale detail overwrite, detail repopulation after clear,
and dashboard loss after a detail error.

### Follow-up GREEN

The same focused command passed 3 test files and 19 tests. `npm.cmd run build` also passed the strict
TypeScript project build and Vite production build. Per coordination guidance, the full suite was
not rerun in this follow-up.
