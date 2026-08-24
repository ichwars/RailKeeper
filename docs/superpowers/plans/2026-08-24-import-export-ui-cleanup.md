# Import/Export UI Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct import/export scrolling and table alignment, replace profile run labels with icon
actions, and allow administrators to permanently delete only cancelled transfer jobs.

**Architecture:** Keep transfer-job deletion in the existing application service so state and area
rules remain server-side, with a narrow repository delete and an Admin-only REST endpoint. Reuse the
existing transfer confirmation dialog in React, expose deletion through the workspace hook, and keep
the job list and history synchronized by reloading their shared job source after a confirmed delete.

**Tech Stack:** Go 1.25, SQLite, OpenAPI 3, React 19, TypeScript, Vite, Vitest, Testing Library,
Lucide React, CSS.

## Global Constraints

- A job is deletable only when its state is exactly `cancelled`.
- Only administrators may see or invoke job deletion.
- Deletion is permanent, explicitly confirmed, audited, and never removes imported inventory data.
- Successful, failed, draft, reading, review-required, ready, and running jobs remain undeletable.
- Preserve CSRF protection, data-transfer area checks, German and English UI copy, dark mode, and
  the existing dense responsive layout.
- Do not add a migration, retention policy, bulk deletion, archive state, undo flow, or dependency.

## File Map

- `backend/internal/application/data_transfer_types.go`: repository deletion contract.
- `backend/internal/application/data_transfer_profiles.go`: cancelled-state and area policy.
- `backend/internal/application/data_transfer_profiles_test.go`: application policy tests and stub.
- `backend/internal/infrastructure/data_transfer_repository.go`: persistent job deletion.
- `backend/internal/infrastructure/data_transfer_repository_test.go`: cascade and not-found tests.
- `backend/internal/api/data_transfer_handlers.go`: DELETE handler and audit event.
- `backend/internal/api/routes.go`: Admin-only route registration.
- `backend/internal/api/data_transfer_handlers_test.go`: authorization, CSRF, state, and audit tests.
- `backend/internal/api/data_transfer_openapi_test.go`: contract-response assertions.
- `openapi/railkeeper.yaml`: public DELETE operation.
- `frontend/src/shared/api.ts`: typed DELETE request.
- `frontend/src/features/importExport/useDataTransferWorkspace.ts`: capability and mutation state.
- `frontend/src/features/importExport/useDataTransferWorkspace.test.tsx`: hook deletion behavior.
- `frontend/src/features/importExport/TransferProfilesTable.tsx`: icon-only run actions and safe cell
  truncation.
- `frontend/src/features/importExport/TransferJobList.tsx`: cancelled-job trash action.
- `frontend/src/features/importExport/TransferHistoryTable.tsx`: cancelled-history trash action and
  safe cell truncation.
- `frontend/src/features/importExport/ImportExportView.tsx`: confirmation orchestration.
- `frontend/src/features/importExport/ImportExportView.test.tsx`: UI visibility and confirmation.
- `frontend/src/shared/i18n/de.ts`, `frontend/src/shared/i18n/en.ts`: deletion copy.
- `frontend/src/styles/data-transfer-dashboard.css`: aligned cells and compact icon actions.
- `frontend/src/styles/data-transfer-dialogs.css`: bounded dialog scrolling.
- `frontend/src/features/importExport/dataTransferDashboardStyles.test.ts`: CSS and markup contracts.

---

### Task 1: Cancelled-job deletion policy and persistence

**Files:**
- Modify: `backend/internal/application/data_transfer_types.go`
- Modify: `backend/internal/application/data_transfer_profiles.go`
- Test: `backend/internal/application/data_transfer_profiles_test.go`
- Modify: `backend/internal/infrastructure/data_transfer_repository.go`
- Test: `backend/internal/infrastructure/data_transfer_repository_test.go`

**Interfaces:**
- Produces: `DataTransferRepository.DeleteJob(context.Context, string) error`
- Produces: `(*DataTransferService).DeleteJob(context.Context, string, ...TransferArea) error`

- [ ] **Step 1: Write failing application tests**

Add table-driven coverage showing that `cancelled` calls the repository delete once while every
other `TransferJobState` returns `ErrDataTransferConflict`. Add a scoped-area case returning
`ErrDataTransferForbidden`.

```go
func TestDeleteJobAllowsOnlyCancelledTransfers(t *testing.T) {
	repository := &dataTransferQueryRepositoryStub{jobs: []DataTransferJob{{
		ID: "cancelled", Direction: TransferImport, Format: TransferCSV,
		Areas: []TransferArea{TransferVehicles}, State: TransferJobCancelled,
	}}}
	service := NewDataTransferService(repository, t.TempDir())
	if err := service.DeleteJob(t.Context(), "cancelled", TransferVehicles); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repository.deletedJobIDs, []string{"cancelled"}) {
		t.Fatalf("deleted jobs = %#v", repository.deletedJobIDs)
	}
}
```

- [ ] **Step 2: Run the application tests and verify RED**

Run: `cd backend; go test ./internal/application -run 'TestDeleteJob' -count=1`

Expected: compilation fails because `DeleteJob` and `deletedJobIDs` do not exist.

- [ ] **Step 3: Add the service and repository contract**

```go
func (s *DataTransferService) DeleteJob(ctx context.Context, id string, allowedAreas ...TransferArea) error {
	job, err := s.getJob(ctx, id, allowedAreas)
	if err != nil {
		return err
	}
	if job.State != TransferJobCancelled {
		return fmt.Errorf("%w: job is %s", ErrDataTransferConflict, job.State)
	}
	return s.repository.DeleteJob(ctx, job.ID)
}
```

Add `DeleteJob(context.Context, string) error` to `DataTransferRepository` and implement it in the
test stub by recording the ID.

- [ ] **Step 4: Verify the application tests are GREEN**

Run: `cd backend; go test ./internal/application -run 'TestDeleteJob' -count=1`

Expected: PASS.

- [ ] **Step 5: Write failing repository cascade tests**

Create a cancelled job, issue, and artifact metadata, call `DeleteJob`, then assert all three table
counts are zero. Also assert an unknown ID returns `sql.ErrNoRows` through `errors.Is`.

```go
if err := repo.DeleteJob(t.Context(), job.ID); err != nil { t.Fatal(err) }
for _, table := range []string{"data_transfer_jobs", "data_transfer_job_issues", "data_transfer_artifacts"} {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil { t.Fatal(err) }
	if count != 0 { t.Fatalf("%s count = %d", table, count) }
}
```

- [ ] **Step 6: Run repository tests and verify RED**

Run: `cd backend; go test ./internal/infrastructure -run 'TestDataTransferRepositoryDeleteJob' -count=1`

Expected: compilation fails because the concrete repository has no `DeleteJob` method.

- [ ] **Step 7: Implement the persistent delete**

```go
func (repository *DataTransferRepository) DeleteJob(ctx context.Context, id string) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM data_transfer_jobs WHERE id=?`, id)
	if err != nil {
		return fmt.Errorf("delete transfer job: %w", err)
	}
	return requireDataTransferUpdate(result, "delete transfer job")
}
```

Map a zero-row update to `sql.ErrNoRows` in the existing helper or directly in `DeleteJob`, keeping
the API's existing not-found mapping intact.

- [ ] **Step 8: Run focused backend tests and commit**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -count=1`

Expected: PASS.

```powershell
git add backend/internal/application/data_transfer_types.go backend/internal/application/data_transfer_profiles.go backend/internal/application/data_transfer_profiles_test.go backend/internal/infrastructure/data_transfer_repository.go backend/internal/infrastructure/data_transfer_repository_test.go
git commit -m "feat: delete cancelled transfer jobs"
```

### Task 2: Admin-only DELETE API and OpenAPI contract

**Files:**
- Modify: `backend/internal/api/data_transfer_handlers.go`
- Modify: `backend/internal/api/routes.go`
- Test: `backend/internal/api/data_transfer_handlers_test.go`
- Test: `backend/internal/api/data_transfer_openapi_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: `(*DataTransferService).DeleteJob(context.Context, string, ...TransferArea) error`
- Produces: `DELETE /api/v1/data-transfer/jobs/{id}` returning `204 No Content`

- [ ] **Step 1: Write failing route tests**

Add one test fixture with Admin and Editor sessions plus cancelled and completed jobs. Assert:

```go
deleted := layoutRequest(t, router, admin, http.MethodDelete,
	"/api/v1/data-transfer/jobs/"+cancelled.ID, nil, true)
assertStatus(t, deleted, http.StatusNoContent)

forbidden := layoutRequest(t, router, editor, http.MethodDelete,
	"/api/v1/data-transfer/jobs/"+otherCancelled.ID, nil, true)
assertProblem(t, forbidden, http.StatusForbidden, "forbidden")

conflict := layoutRequest(t, router, admin, http.MethodDelete,
	"/api/v1/data-transfer/jobs/"+completed.ID, nil, true)
assertProblem(t, conflict, http.StatusConflict, "data_transfer_conflict")
```

Also call the Admin route without CSRF and expect 403, delete a missing ID and expect 404, and assert
the audit table contains `DataTransferJobDeleted` only for the successful request.

- [ ] **Step 2: Run the API test and verify RED**

Run: `cd backend; go test ./internal/api -run 'TestDataTransferJobDelete' -count=1`

Expected: route returns 404 or 405 because DELETE is not registered.

- [ ] **Step 3: Register and implement the handler**

```go
func (a *App) deleteDataTransferJob(w http.ResponseWriter, r *http.Request) {
	if !a.dataTransferAvailable(w) { return }
	if err := a.dataTransferService.DeleteJob(
		r.Context(), r.PathValue("id"), allowedDataTransferAreas(r)...,
	); err != nil {
		a.dataTransferError(w, err, "delete data transfer job")
		return
	}
	a.recordAudit(r, "DataTransferJobDeleted", "data_transfer_job", r.PathValue("id"))
	w.WriteHeader(http.StatusNoContent)
}
```

Register `{http.MethodDelete, "/api/v1/data-transfer/jobs/{id}", routeAccessAdmin,
(*App).deleteDataTransferJob, nil}` next to the GET route.

- [ ] **Step 4: Verify route tests are GREEN**

Run: `cd backend; go test ./internal/api -run 'TestDataTransferJobDelete' -count=1`

Expected: PASS.

- [ ] **Step 5: Write the failing OpenAPI assertion**

Extend `data_transfer_openapi_test.go` to require the DELETE operation and response codes
`204`, `403`, `404`, and `409`, plus both `sessionCookie` and `csrfHeader` security entries.

- [ ] **Step 6: Run the OpenAPI test and verify RED**

Run: `cd backend; go test ./internal/api -run 'TestDataTransferOpenAPI' -count=1`

Expected: FAIL because `/data-transfer/jobs/{id}` has no DELETE operation.

- [ ] **Step 7: Add the OpenAPI DELETE operation**

```yaml
    delete:
      tags: [DataTransfer]
      summary: Permanently delete a cancelled data transfer job
      description: Admin role required. Jobs in every state except cancelled are rejected.
      security:
        - sessionCookie: []
          csrfHeader: []
      responses:
        "204":
          description: Cancelled transfer job deleted
        "403":
          description: Forbidden
        "404":
          description: Job not found
        "409":
          description: Job is not cancelled
```

Use the existing `Problem` response bodies for error responses.

- [ ] **Step 8: Run API tests and commit**

Run: `cd backend; go test ./internal/api -count=1`

Expected: PASS.

```powershell
git add backend/internal/api/data_transfer_handlers.go backend/internal/api/routes.go backend/internal/api/data_transfer_handlers_test.go backend/internal/api/data_transfer_openapi_test.go openapi/railkeeper.yaml
git commit -m "feat: expose cancelled transfer deletion"
```

### Task 3: Frontend deletion API and workspace state

**Files:**
- Modify: `frontend/src/shared/api.ts`
- Modify: `frontend/src/features/importExport/useDataTransferWorkspace.ts`
- Test: `frontend/src/features/importExport/useDataTransferWorkspace.test.tsx`

**Interfaces:**
- Produces: `api.deleteDataTransferJob(id: string): Promise<void>`
- Produces: `capabilities.canDeleteJobs: boolean`
- Produces: `deleteJob(jobId: string): Promise<void>`

- [ ] **Step 1: Write failing hook tests**

Render the hook as Admin with a selected cancelled job, mock `api.deleteDataTransferJob`, invoke
`deleteJob`, and assert the DELETE call, cleared selection/details, and refreshed summary/profiles/jobs.
Render as Editor and assert `canDeleteJobs` is false.

```tsx
await act(async () => result.current.deleteJob("job-cancelled"));
expect(api.deleteDataTransferJob).toHaveBeenCalledWith("job-cancelled");
expect(result.current.selectedJobId).toBeNull();
expect(result.current.capabilities.canDeleteJobs).toBe(true);
```

- [ ] **Step 2: Run the hook tests and verify RED**

Run: `cd frontend; npm.cmd run test:run -- src/features/importExport/useDataTransferWorkspace.test.tsx`

Expected: TypeScript/test failure because the API method, capability, and hook action do not exist.

- [ ] **Step 3: Add the API method and deletion mutation**

```ts
deleteDataTransferJob: (id: string) =>
  request<void>(`/data-transfer/jobs/${encodeURIComponent(id)}`, { method: "DELETE" }),
```

Add `canDeleteJobs: isAdmin`. Implement a dedicated mutation that waits for the DELETE response,
then clears the selected job only when its ID matches, and finally reloads the workspace. On error,
preserve the selected job and surface the existing error message.

- [ ] **Step 4: Run the hook tests and commit**

Run: `cd frontend; npm.cmd run test:run -- src/features/importExport/useDataTransferWorkspace.test.tsx`

Expected: PASS.

```powershell
git add frontend/src/shared/api.ts frontend/src/features/importExport/useDataTransferWorkspace.ts frontend/src/features/importExport/useDataTransferWorkspace.test.tsx
git commit -m "feat: manage cancelled transfer deletion"
```

### Task 4: Icon actions, deletion confirmation, and synchronized views

**Files:**
- Modify: `frontend/src/features/importExport/TransferProfilesTable.tsx`
- Modify: `frontend/src/features/importExport/TransferJobList.tsx`
- Modify: `frontend/src/features/importExport/TransferHistoryTable.tsx`
- Modify: `frontend/src/features/importExport/ImportExportView.tsx`
- Test: `frontend/src/features/importExport/ImportExportView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/data-transfer-dashboard.css`

**Interfaces:**
- `TransferJobList` and `TransferHistoryTable` consume `canDelete`, `mutating`, and
  `onDeleteRequest(job: DataTransferJob)`.
- `TransferProfilesTable` keeps the existing `onRun(profile)` callback.

- [ ] **Step 1: Write failing component tests**

Add a cancelled import fixture and assert Admin sees two delete actions, one in the job list and one
in history. Assert Editor sees none. Click one action, verify the confirmation names the job, confirm,
and expect `api.deleteDataTransferJob("job-cancelled")`. Assert completed and failed jobs never show
a delete action. Assert profile run actions retain the names `Fahrzeugimport starten` and
`Fahrzeugexport starten` but contain no visible action text.

- [ ] **Step 2: Run the view tests and verify RED**

Run: `cd frontend; npm.cmd run test:run -- src/features/importExport/ImportExportView.test.tsx`

Expected: FAIL because delete controls and confirmation do not exist and run labels are visible.

- [ ] **Step 3: Replace profile run labels with icons**

Select `FileUp` for import and `FileDown` for export. Render the chosen icon inside an `icon-button
transfer-profile-run` button with the existing full action text in both `aria-label` and `title`.
Keep `MoreVertical` as the separate edit action.

- [ ] **Step 4: Add cancelled-job delete actions**

Refactor each job card to an outer non-button container with a full-width selection button and a
sibling trash icon, avoiding nested buttons. In history, place `Trash2` beside `ChevronRight`, stop
event propagation, and render it only when `canDelete && job.state === "cancelled"`.

- [ ] **Step 5: Reuse the transfer confirmation dialog**

```tsx
const [jobPendingDelete, setJobPendingDelete] = useState<DataTransferJob | null>(null);
const deleteAction: TransferPendingAction | null = jobPendingDelete ? {
  title: t("importExport.dashboard.delete.title"),
  body: t("importExport.dashboard.delete.body", {
    name: jobPendingDelete.sourceName || jobPendingDelete.profileName
  }),
  confirmLabel: t("importExport.dashboard.delete.confirm"),
  errorMessage: t("importExport.dashboard.delete.error"),
  dangerous: true,
  run: () => workspace.deleteJob(jobPendingDelete.id)
} : null;
```

Render `TransferConfirmDialog` at view level and pass `setJobPendingDelete` to both job views. Add
German and English keys for title, body, confirm, cancel, error, and the delete-action label.

- [ ] **Step 6: Correct truncation markup and action styling**

Wrap truncated values in `<span className="data-transfer-truncate">` inside ordinary table cells.
Do not attach `data-transfer-truncate` directly to `td`. Style the card wrapper, selection surface,
trash action, 30–32 px profile icon actions, visible focus, danger hover, and the widened history
action column using existing tokens.

- [ ] **Step 7: Run the view tests and commit**

Run: `cd frontend; npm.cmd run test:run -- src/features/importExport/ImportExportView.test.tsx`

Expected: PASS.

```powershell
git add frontend/src/features/importExport/TransferProfilesTable.tsx frontend/src/features/importExport/TransferJobList.tsx frontend/src/features/importExport/TransferHistoryTable.tsx frontend/src/features/importExport/ImportExportView.tsx frontend/src/features/importExport/ImportExportView.test.tsx frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/data-transfer-dashboard.css
git commit -m "fix: refine transfer workspace actions"
```

### Task 5: Dialog scrolling, regression contracts, and full verification

**Files:**
- Modify: `frontend/src/styles/data-transfer-dialogs.css`
- Test: `frontend/src/features/importExport/dataTransferDashboardStyles.test.ts`

**Interfaces:**
- No new runtime interface. This task locks the layout contract around existing class names.

- [ ] **Step 1: Write failing style-contract tests**

Read both transfer CSS files and the two table components. Assert explicit dialog rows, body
vertical scrolling, stable scrollbar gutter, review overflow, and no `td` carrying the truncation
class.

```ts
expect(dialogStyles).toMatch(/\.data-transfer-dialog\s*\{[^}]*grid-template-rows:\s*auto auto minmax\(0, 1fr\)/s);
expect(dialogStyles).toMatch(/\.data-transfer-dialog-body\s*\{[^}]*overflow-y:\s*auto/s);
expect(dialogStyles).toMatch(/\.data-transfer-dialog-body\s*\{[^}]*scrollbar-gutter:\s*stable/s);
expect(dialogStyles).toMatch(/\.transfer-review-wrap\s*\{[^}]*overflow:\s*auto/s);
expect(profileTableSource).not.toMatch(/<td[^>]*className="data-transfer-truncate"/);
expect(historyTableSource).not.toMatch(/<td[^>]*className="data-transfer-truncate"/);
```

- [ ] **Step 2: Run the style test and verify RED**

Run: `cd frontend; npm.cmd run test:run -- src/features/importExport/dataTransferDashboardStyles.test.ts`

Expected: FAIL because the dialog lacks explicit rows/gutter, review overflow is absent, and current
table cells carry the truncation class.

- [ ] **Step 3: Implement bounded scrolling**

```css
.data-transfer-dialog {
  grid-template-rows: auto auto minmax(0, 1fr);
}

.data-transfer-dialog-body {
  overflow-x: hidden;
  overflow-y: auto;
  scrollbar-gutter: stable;
}

.transfer-mapping-table-wrap,
.transfer-review-wrap {
  overflow: auto;
}
```

Retain the current maximum heights and the layer-level fallback scrolling for small viewports.

- [ ] **Step 4: Run focused frontend tests**

Run: `cd frontend; npm.cmd run test:run -- src/features/importExport/dataTransferDashboardStyles.test.ts src/features/importExport/ImportExportView.test.tsx src/features/importExport/useDataTransferWorkspace.test.tsx`

Expected: PASS.

- [ ] **Step 5: Format and run complete automated verification**

Run:

```powershell
gofmt -w backend/internal/application/data_transfer_types.go backend/internal/application/data_transfer_profiles.go backend/internal/application/data_transfer_profiles_test.go backend/internal/infrastructure/data_transfer_repository.go backend/internal/infrastructure/data_transfer_repository_test.go backend/internal/api/data_transfer_handlers.go backend/internal/api/data_transfer_handlers_test.go
cd backend
go test ./...
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
cd ..
git diff --check
```

Expected: Go tests PASS, all frontend tests PASS, Vite build exits 0, and `git diff --check` emits no
output.

- [ ] **Step 6: Perform visual browser verification**

Rebuild the served frontend if needed and inspect `/import-export` at approximately 1480 px and
680 px widths in dark mode. Verify:

- the import dialog body shows a usable vertical scrollbar and its lower controls remain reachable;
- mapping/review tables scroll without compressing their rows;
- profile and history headers align with every body column;
- import/export run icons expose tooltips and visible focus;
- only cancelled jobs expose trash icons in both views;
- cancelling the confirmation keeps the job, confirming removes it from both views;
- completed and failed transfers cannot be deleted.

- [ ] **Step 7: Commit the verified layout fix**

```powershell
git add frontend/src/styles/data-transfer-dialogs.css frontend/src/features/importExport/dataTransferDashboardStyles.test.ts
git commit -m "fix: restore transfer dialog scrolling"
```

- [ ] **Step 8: Review final branch state**

Run: `git status --short --branch; git log --oneline main..HEAD; git diff --stat main...HEAD`

Expected: clean branch with the design, plan, backend deletion, API, frontend interaction, and layout
commits; no generated output or local data is tracked.
