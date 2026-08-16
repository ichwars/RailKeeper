# Master Data Lifecycle Design (#83)

## Context

RailKeeper currently deletes most master-data rows physically. The bundled seed loader uses
`ON CONFLICT(type, key) DO NOTHING`, so a deleted bundled row is inserted again on the next start
and becomes active. The application also cannot distinguish a RailKeeper-supplied row from a row
created by a user.

This behavior is surprising and unsafe for customized installations. Deactivation must survive
restarts, updates, reseeding, master-data transfer, backup, and restore. Existing vehicle and
accessory values must remain readable even when their master-data entry is inactive.

Issue #83 is implemented separately from the Windows data-directory migration in #84 and the ZIP
update download in #82.

## Goals

- Distinguish bundled and user-created master-data entries.
- Let every entry be deactivated and reactivated.
- Prevent permanent deletion of bundled entries.
- Permit permanent deletion of a user-created entry only when it has no domain usage.
- Preserve existing selections, relationships, user edits, and inactive state across updates.
- Add newly shipped entries as active without changing existing rows.
- Preserve the lifecycle state through master-data export/import and application backup/restore.
- Present clear, separate deactivate, reactivate, and permanent-delete actions.

## Non-goals

- Do not implement the Windows data-directory migration from #84.
- Do not implement update discovery or ZIP download from #82.
- Do not replace existing text-based vehicle fields with foreign keys in this change.
- Do not redesign the complete settings or inventory interfaces.

## Chosen Approach

Store an explicit, server-managed origin on each master-data row. The two supported origins are:

- `bundled`: shipped and owned by RailKeeper
- `custom`: created by a user or imported without a matching RailKeeper-owned key

The origin is returned by read APIs but is not accepted as an editable property in normal create or
update requests. The backend remains authoritative for lifecycle actions and usage checks.

Alternatives were rejected because a separate seed tombstone table would split one lifecycle across
multiple records, while making every row soft-delete-only would contradict the approved ability to
delete unused custom entries.

## Data Model and Bundled Catalog

Add a forward-only migration with an `origin` column on `master_data_entries`. The column is
non-null, constrained to `bundled` or `custom`, and defaults to `custom` so unknown legacy and
imported rows fail safely.

RailKeeper maintains one authoritative bundled-key catalog. It is the union of entries in the
normal master-data seed and the known entries historically introduced by SQL migrations, such as
article types, accessory subtypes, function symbols, CV8 manufacturers, and symbols. The catalog is
identified by immutable `(type, key)` pairs, never by editable labels.

During startup, after migrations and before the master-data cache is warmed, seeding performs two
operations:

1. Insert missing bundled rows as active bundled entries.
2. Mark catalog-matching existing rows as bundled without changing their label, metadata,
   sort order, source URL, or active state.

This keeps existing user changes intact. A newly shipped key is inserted active. An existing
inactive key stays inactive. Future migrations or seeds that introduce bundled data must also add
the immutable key to the catalog.

The migration and startup reconciliation must classify all currently shipped entries while leaving
existing user-created keys as custom. Classification is covered by fixtures that include customized
labels and inactive rows.

## Lifecycle Commands

Use a dedicated active-state command instead of requiring the client to resubmit the complete
editable entry. It accepts only the requested boolean state and updates `updated_at`.

Deactivation and reactivation rules:

- Both bundled and custom entries may be deactivated.
- Deactivation never removes domain values or master-data relations.
- Reactivation restores availability for new selections.
- Both operations invalidate the master-data cache.

Permanent deletion rules:

- Bundled entries are always rejected.
- Custom entries require a server-side domain-usage check.
- Unknown or unsupported reference mappings fail closed and reject deletion.
- Purely administrative relations do not make an otherwise unused custom entry permanent. They are
  removed in the same transaction as the entry.
- The usage check, relation cleanup, and row deletion execute in one reserved SQLite write
  transaction so a concurrent write cannot invalidate the decision.

The backend distinguishes at least these errors:

- bundled entry cannot be permanently deleted
- custom entry is still in use and can only be deactivated
- entry no longer exists

The HTTP layer maps lifecycle conflicts to stable problem codes and conflict responses. It does not
leak SQL details.

## Usage Detection and Historical Values

Usage detection is centralized in the application layer and has explicit adapters for each managed
master-data type. It covers vehicle model fields, accessory product fields and attributes,
decoder/function data, symbol assignments, and other persisted consumers found during
implementation. Values stored as JSON arrays are parsed and compared as values rather than searched
with unsafe substring matching.

Several existing vehicle fields store the selected label as text instead of a foreign key. This
change does not rewrite that storage model. For these fields, the usage adapter compares the actual
persisted selection value used by the UI. Deleting a custom master-data row never clears the copied
historical text from a vehicle or accessory record.

The editor also treats a current persisted value that is absent from the active option set as a
historical inactive value. This covers both a deliberately deactivated entry and legacy text that
no longer has a matching row. The value remains selected and readable until the user chooses a new
active value.

## API Contract

Extend `MasterDataEntry` with the read-only origin. Management responses additionally expose the
allowed lifecycle actions, including whether permanent deletion is currently possible. Usage
capabilities are requested only by management screens so normal active-only option loading does not
pay the cost of cross-domain usage checks.

Add a narrow endpoint for changing active state. Keep the existing delete endpoint, but enforce the
new bundled and in-use rules server-side. Align the Go response types, frontend API adapter, and
`openapi/railkeeper.yaml`.

Create and update inputs cannot assign or change origin. Import and restore use separate trusted
normalization rules described below.

## Settings UI

The general and article master-data management screens show:

- status: Active or Inactive
- origin: RailKeeper or Custom entry
- edit action
- Deactivate or Reactivate action
- Permanently delete action only for an unused custom entry

Bundled and used custom entries do not show a delete action. Inactive rows remain visible in the
management list and use the existing muted-row visual language without relying on color alone.

Deactivation requires a confirmation with this meaning:

> The entry will no longer be offered for new records. Existing saved uses remain unchanged.

Permanent deletion uses danger styling, says "Permanently delete", and explicitly describes that
the unused custom master-data entry will be removed. Reactivation is immediate.

Buttons use compact existing icon-button patterns, localized German and English labels, tooltips,
and accessible names. Desktop and mobile layouts must remain dense and readable.

## Inventory and Accessory Editors

New records receive only active options. When editing an existing record:

- its current inactive value stays selected
- the option label gains an "inactive" suffix
- the historical inactive option cannot be selected for another record
- after choosing a different active value, the inactive value disappears from that editor's option
  set

The same rule applies to vehicles and accessories. Read-only views, tables, reports, and exports
continue to render the stored value without losing it.

## Master-data Export and Import

Increment the standalone master-data document version. New exports contain origin and active state.
Version 1 documents remain accepted.

Import reconciliation is server-authoritative:

- a key already known locally as bundled remains bundled regardless of file content
- an unknown key is custom even if the file claims it is bundled
- current bundled entries omitted from an import are retained rather than physically removed
- imported active states and editable values are applied to matching entries
- custom entries follow the imported document, subject to existing validation and reference safety

The import remains an all-or-nothing transaction and refreshes the cache only after commit.

## Backup and Restore

Increment the application backup version because `master_data_entries` gains a column. Current
backups round-trip origin and active state.

Older supported backups remain restorable. Before destructive restore, RailKeeper retains the
current set of authoritative bundled keys. After table import it reconciles those keys back to
bundled without changing their restored active state or editable values. Missing origin values
default to custom. A backup cannot promote an arbitrary unknown key to bundled.

Existing compatibility validation and transactional restore behavior remain in force. Restore must
not silently reactivate an inactive bundled entry.

## Verification

Backend tests cover:

- migration and startup classification of existing bundled and custom rows
- reseeding an inactive or customized bundled entry without overwriting it
- insertion of a newly shipped bundled entry as active
- rejection of permanent deletion for bundled entries
- deletion of an unused custom entry and cleanup of its administrative relations
- rejection of deletion for used custom entries across every supported reference adapter
- transaction behavior and stable API problem codes
- version 1 and new-version master-data import/export
- old and current application backup restore
- cache invalidation after lifecycle changes

Frontend tests cover:

- status and origin rendering
- correct action availability
- confirmation wording and API calls
- active-only choices for new records
- retained and marked inactive values while editing vehicles and accessories
- German and English labels

Final verification consists of `go test ./...`, the full frontend test suite,
`npm.cmd run build`, and a visual desktop/mobile check of settings, vehicle editing, and accessory
editing.

## Rollout and Safety

Implementation occurs in a dedicated feature branch and PR for #83. The database change is
forward-only. No existing master-data, vehicle, accessory, decoder, function, relation, or upload
row is deleted by migration or seed reconciliation.

The PR must not include #84 or #82. Those issues follow only after #83 is merged and verified.
