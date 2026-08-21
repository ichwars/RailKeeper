# Data Transfer and Digital Centers Design

Status: approved design, 20 August 2026

## Goal

RailKeeper receives two dedicated operational workspaces:

1. Import/Export follows the selected data-transfer dashboard reference and manages persistent
   profiles, reviewable transfer jobs, transfer history, and local export artifacts.
2. Digital Centers follows the selected three-column operational reference and combines passive
   live monitoring, locomotive comparison, diagnosis, and controlled synchronization.

Both workspaces retain RailKeeper's current colors, tokens, typography, and local-first deployment
model. The reference images define the desktop topology.

## Source of truth and user decisions

- Import/Export visual reference:
  `C:\Users\droth\.codex\generated_images\01a01b90-522c-7693-b405-a54de717a377\exec-072ae6d7-1eda-49c2-99a7-cf6bfd5f2a74.png`
- Digital Centers visual reference:
  `C:\Users\droth\Documents\Codex\RailKeeper\design-specs\digitalzentralen\digitalzentralen-referenz.png`
- Digital Centers product concept:
  `C:\Users\droth\Documents\Codex\RailKeeper\design-specs\digitalzentralen\digitalzentralen-konzept.md`
- Import/Export covers vehicles, accessories, and exhibition lists.
- Import/Export must not import or export master data. Master-data transfer remains exclusively in
  Settings.
- Settings > Digital Centers remains unchanged and is the sole configuration surface for command
  stations.
- Add and edit actions in the Digital Centers workspace navigate to Settings > Digital Centers.
- Profiles, jobs, previews, history, and artifact metadata are persistent.

## Non-goals

- Do not move, remove, or redesign master-data transfer in Settings.
- Do not move, remove, or redesign command-station configuration in Settings.
- Do not turn the live monitor into a train or turnout controller.
- Do not add cloud synchronization, multi-tenant storage, public sharing, or automatic device
  writes.
- Do not include users, roles, sessions, password hashes, rate limits, audit logs, or master-data
  tables in feature-level transfer packages.

## Architecture

### Data-transfer module

Add a focused backend application module for data transfer. HTTP handlers remain thin, the
application service owns validation and transaction boundaries, and infrastructure adapters own
SQLite and filesystem access.

Migration `0061_data_transfer.sql` adds:

- `data_transfer_profiles`: name, direction, format, selected areas, options, enabled state,
  ownership metadata, and timestamps.
- `data_transfer_jobs`: profile snapshot, direction, state, current stage, source name, source hash,
  package version, record counts, creator, timestamps, and terminal result.
- `data_transfer_job_issues`: job, area, record key or row number, field, severity, issue code,
  readable message, proposed resolution, and selected resolution.
- `data_transfer_artifacts`: job, confined relative path, display name, MIME type, byte size, digest,
  and timestamps.

Completed jobs are the transfer history. The implementation must not maintain a second divergent
history table. Profiles keep their own last-used timestamp, while every job stores a complete
profile snapshot so later profile edits do not rewrite history.

### Transfer package

RailKeeper JSON uses this versioned envelope:

```json
{
  "format": "railkeeper-transfer",
  "version": 1,
  "createdAt": "2026-08-20T12:00:00Z",
  "areas": {
    "vehicles": [],
    "accessories": [],
    "exhibitionLists": []
  }
}
```

Only selected keys are emitted. `masterData` is not a supported area and is rejected at this API
boundary. Unknown top-level or area keys produce a validation error rather than being ignored.

Supported combinations:

| Area | CSV | RailKeeper JSON |
| --- | --- | --- |
| Vehicles | Yes | Yes |
| Accessories | Yes | Yes |
| Exhibition lists | No | Yes |
| Multiple areas | No | Yes |

An exhibition-list package contains the list header and its entries. It references labels needed to
render the list but does not contain separately importable master-data tables.

### Import transaction

Imports use two server-side phases:

1. Upload, type detection, structural validation, parsing, field mapping, duplicate detection, and
   persistent preview creation.
2. Explicit confirmation, stale-preview validation, and one SQLite transaction for every approved
   change in the job.

The job stores a cryptographic hash of the uploaded file and a snapshot of the proposed changes.
Before commit, the service rechecks relevant business keys and update timestamps. A changed file,
changed mapping, changed conflict decision, or stale target invalidates the previous confirmation.
Any write error rolls back the complete job.

Vehicle and accessory imports support create and controlled update decisions. Exhibition-list
imports support create, replace-after-confirmation, and import-as-copy. Locked lists cannot be
replaced. Vehicle links in exhibition entries are resolved by stable IDs only when the source is
known to be the same installation; otherwise they are proposed by inventory number and require
confirmation.

### Export execution

Exports snapshot the profile, calculate the exact selected record count, validate the format/area
combination, and create an artifact under the configured data directory. Artifact paths are stored
as confined relative paths and are revalidated before download or deletion.

The local-storage panel lists real artifacts. A native "open folder" action is exposed only when the
runtime can perform it safely on the local host. Other deployments show the confined path and
download actions without claiming that a server folder can be opened in the browser.

### Backup compatibility

The existing application backup and restore contract remains distinct from feature-level transfer.
Transfer profiles are included in future application backups as configuration. Jobs, issues,
history, and generated artifact bytes are operational records and remain excluded. Existing backup
versions continue to restore without requiring data-transfer records.

## Import/Export workspace

### Desktop topology

The page follows the selected reference:

1. Header with `DATENTRANSFER`, title, subtitle, `Neuer Import`, and `Neuer Export`.
2. Full-width four-cell summary strip for open jobs, selected record total, latest export, and local
   storage.
3. Left column with job filters and job cards.
4. Center column with export profiles above transfer history.
5. Right column with selected-job details above local storage.

The reference's master-data examples are replaced with exhibition-list examples. A complete profile
contains vehicles, accessories, and exhibition lists.

### Job list and details

Filters cover all, open, and completed jobs. Each card shows direction, area, filename or profile,
state, ready/total counts, and progress. Required states are draft, reading, review required, ready,
running, completed, completed with warnings, failed, and cancelled.

Selecting a card updates the right-hand detail panel. The stepper represents real stages such as
file, mapping, review, and completion. The primary action always advances to the next safe stage. It
never writes merely because the job was selected or reopened.

### Import flow

`Neuer Import` opens a guided dialog:

1. Select an import profile or one allowed area.
2. Select or drop a supported file.
3. Review structural validation and field mapping.
4. Resolve duplicates, conflicts, and invalid records.
5. Review the final change summary.
6. Confirm the transaction explicitly.

Import profiles may store area, format, mappings, and default conflict suggestions. They never
store a prior approval and never auto-confirm a later job.

### Export flow and profiles

`Neuer Export` accepts an existing export profile or an ad hoc configuration. The dialog shows
areas, format, options, and final record count before execution. Invalid combinations are disabled
with a reason.

Profiles can be created, edited, disabled, and executed. Re-executing a profile creates a new job
with a profile snapshot. It does not mutate the previous history entry.

### History and artifacts

History rows show time, direction, areas, record count, result, filename, and actions. Details open
the persisted job summary. Retry creates a new job with the previous settings but no inherited
import confirmation. Export artifacts can be downloaded again or explicitly deleted. Deleting an
artifact retains the history metadata and marks the file as unavailable.

### Responsive behavior

At narrower widths, the job list, center workspace, and detail panel stack in that order. The detail
panel may become a drawer after job selection. Tables use controlled horizontal scrolling only when
column reduction would remove core information. Long German labels and filenames remain accessible
through truncation plus title or detail views.

## Digital Centers workspace

### Separation from Settings

Create a dedicated frontend feature and route for Digital Centers. Do not render `SettingsView` in
a standalone mode. The existing Settings tab, its configuration forms, connection tests, and stored
settings remain unchanged.

The workspace reads configured command stations through existing server APIs. Add, edit, and
configuration actions navigate to Settings > Digital Centers with the relevant station selected
when possible.

### Desktop topology

The page follows the selected reference:

1. Header with `DIGITALBETRIEB`, title, and subtitle.
2. Full-width active-station toolbar with station selector, connection state, host and port, live
   monitor control, read action, and context menu.
3. Left column with configured stations.
4. Largest center column with the locomotive worklist.
5. Right column with Live Status, Diagnosis, and Messages tabs plus the write-lock panel.

### Reading and comparison

`Daten lesen` tests the connection and reads supported locomotive master data into a session-bound
worklist. It writes neither RailKeeper data nor station data. Search, quick filters, advanced
filters, pagination, refresh, and row actions operate on that worklist.

Comparison states are OK, deviation, missing in station, new, and conflict. A row detail compares
RailKeeper and station values with source and timestamp. New station objects are never matched by
name alone. Address, protocol, manufacturer information, and other supported fields may contribute
to a proposal, but the user confirms the mapping.

### Live status, diagnosis, and messages

The live monitor is passive. It may observe message rate, block or response counters, and masked
events, but it sends no driving or switching command. The pulse graph is rendered from live data in
a canvas-based chart, not as decorative CSS or a fabricated static asset.

Connection loss marks the display as interrupted and stops presenting old values as live. Diagnosis
shows latency, device and protocol versions, adapter capabilities, last successful communication,
and readable protocol errors. Messages include time, severity, station, explanation, and a useful
next action.

### Write protection and synchronization

Writing is locked by default. A one-operation authorization requires:

1. A successful connection and device identification.
2. A fresh read of the current station state.
3. A field-level preview.
4. Resolved mappings and conflicts.
5. Explicit confirmation by an authorized user.

The authorization expires after completion, cancellation, connection loss, or a short server-side
timeout. CV writes require a separate confirmation. Unsupported fields are skipped with a reason.
After every write, RailKeeper reads the affected values again, compares the result, and records an
audit event.

### Responsive behavior

On medium screens, the status column moves below the worklist. On small screens, the station list
becomes a collapsible selector above the worklist and the right-side tabs become a full-width panel.
Core locomotive columns remain available through deliberate column reduction and controlled table
scrolling.

## API shape

The OpenAPI contract, backend routes, and frontend API adapter change together. The data-transfer
surface includes:

- profile list, create, update, disable, and delete operations;
- job list and detail operations;
- import upload, preview, issue-resolution, confirm, and cancel operations;
- export job creation and execution;
- artifact download and deletion;
- summary counts and recent-history queries.

Digital Centers uses existing connection and adapter APIs where they already fit. New endpoints are
limited to worklist persistence, field-level comparison, write authorization, verified write
execution, diagnosis, and session messages. Configuration remains on the existing Settings API.

## Authorization and security

- Viewer-equivalent roles may run permitted exports and inspect permitted history.
- Editor and Admin may create import jobs and confirm inventory changes.
- Command-station writes require the existing privileged role boundary and a server-side operation
  token tied to station, user, preview, and expiry.
- Messe may use exhibition-list-only profiles according to its existing exhibition permissions. It
  cannot obtain general vehicle or accessory exports through profile composition.
- All writes remain CSRF protected and audited.
- Upload limits, extension checks, MIME detection, executable blocking, package version checks, and
  data-directory confinement apply before parsing or persistence.
- Device replies, imported values, filenames, and external identifiers are untrusted input.

## Error behavior

Errors identify the cause and the next safe action. Import issues include area, record or row, field,
severity, and code. Connection failures distinguish unreachable host, timeout, protocol error,
incompatible version, and missing permission. A technical detail view supplements the readable
message without replacing it.

Failed import transactions leave no partial data. Interrupted exports leave no registered artifact
unless the final file was atomically moved into place. Disconnected monitors stop live timestamps.
Expired write authorizations return the user to a fresh read and preview.

## Testing and verification

Backend coverage includes:

- profile validation and persistence;
- package parsing and version rejection;
- per-area export and combined JSON export;
- vehicle, accessory, and exhibition-list import previews;
- conflict and stale-preview detection;
- transaction rollback on any record failure;
- artifact path confinement and atomic creation;
- roles, CSRF, Messe isolation, and audit records;
- command-station read-only behavior, authorization expiry, writes, and verification reads.

Frontend coverage includes:

- routing and unchanged Settings tabs;
- profile and job dialogs;
- job selection, filters, details, history, and retry behavior;
- import mapping, conflict resolution, and explicit confirmation;
- station selection, worklist filters, pagination, tabs, and write-lock states;
- loading, empty, warning, failed, interrupted, and completed states;
- German and English labels.

Verification requires `go test ./...`, `npm.cmd run build`, targeted frontend tests, and browser checks
for both workspaces in light and dark themes at desktop and smaller widths. Design QA compares each
rendered page with its reference at the same viewport and state. No P0, P1, or P2 visual difference
may remain before handoff.

## Delivery sequence

1. Add migration, domain types, repositories, services, routes, and OpenAPI for data transfer.
2. Add transactional packages for vehicles, accessories, and exhibition lists.
3. Build the Import/Export workspace against the real APIs.
4. Restore Settings to its unchanged command-station configuration role.
5. Add operational Digital Centers services and the dedicated workspace.
6. Complete security, regression, responsive, browser, and design-QA verification.

The sequence may be divided into reviewable commits, but each exposed workflow must remain coherent
and must not present a write action before its server-side safety controls exist.
