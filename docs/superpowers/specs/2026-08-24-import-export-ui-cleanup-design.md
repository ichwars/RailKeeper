# Import/Export UI Cleanup and Cancelled Job Deletion

**Status:** Approved for specification on 2026-08-24

## Goal

Correct the layout defects in the import/export workspace, reduce repeated profile actions to clear
icons, and let administrators permanently remove cancelled transfer jobs without affecting
successful, failed, or active transfers.

## Scope

- Restore a usable vertical scrollbar in the import wizard and prevent lower content from being
  compressed or clipped.
- Correct the shifted columns in the transfer-profile and transfer-history tables.
- Replace the labelled profile run buttons with compact import/export icon actions.
- Add permanent deletion for transfer jobs whose state is exactly `cancelled`.
- Exclude bulk deletion, automatic retention, archiving, undo, and deletion of any other job state.

## Dialog and Table Layout

- Define the import dialog rows explicitly as header, wizard steps, and a bounded scrollable body.
  The body owns vertical scrolling and keeps a stable scrollbar gutter.
- Keep the mapping and review tables independently scrollable when their contents exceed their
  bounded height. Their headers and controls must remain usable at the supported viewport sizes.
- Do not apply block layout to a `td`. Truncated cell content is wrapped in an inner element so the
  browser retains the table-cell geometry and all headers remain aligned with their columns.
- Preserve the existing dense table dimensions, horizontal overflow behavior, dark mode, and mobile
  fallbacks.

## Profile Actions

- Use an icon-only action for enabled profiles that the current role may run.
- Use a file-upload/import icon for import profiles and a file-download/export icon for export
  profiles.
- Preserve the existing action text as `aria-label` and `title` so the purpose remains accessible
  and is shown as a tooltip.
- Keep the existing profile edit menu separate from the run action.

## Cancelled Job Deletion

- The job list and transfer history are two views of the same `data_transfer_jobs` record. Deleting
  a job therefore removes it from both views after the workspace reloads.
- Show a trash icon only for jobs whose state is `cancelled` and only to administrators. Place the
  action in both the job card and the corresponding history row.
- Require an explicit confirmation that names the transfer profile or source file.
- Add `DELETE /api/v1/data-transfer/jobs/{id}`. The route requires an authenticated administrator,
  CSRF protection, and the normal data-transfer area visibility check.
- The application service re-reads the job and rejects every state other than `cancelled` with a
  conflict response. The repository deletes the job in a transaction; existing foreign keys remove
  associated issue and artifact metadata.
- Cancelled import jobs do not produce export artifacts. The operation therefore does not delete
  completed export files or imported inventory data.
- Record `DataTransferJobDeleted` in the audit log after successful deletion. Unknown jobs return
  not found, forbidden roles return forbidden, and non-cancelled jobs return conflict.
- If the deleted job was selected, clear its details before reloading so the UI cannot display stale
  data.

## Error Handling

- Disable destructive controls while another transfer mutation is active.
- Keep the confirmation open when deletion fails and show the server-provided error through the
  existing transfer error presentation.
- Never optimistically remove a job before the server confirms deletion.

## Testing and Verification

- Add application, repository, API, and OpenAPI tests for successful cancelled-job deletion, cascade
  cleanup, Admin-only access, CSRF, area visibility, not-found handling, and rejection of all other
  states.
- Add frontend tests for icon-only profile actions, delete-action visibility, confirmation, API
  invocation, selected-job cleanup, and workspace refresh.
- Extend the focused CSS contract test to cover explicit dialog grid rows, scroll ownership, and the
  absence of block layout on table cells.
- Run the complete Go test suite, frontend test suite and build, OpenAPI checks, and a visual browser
  check at desktop and narrow widths in dark mode.
