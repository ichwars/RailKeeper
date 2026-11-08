# Backup format, validation and restoration ownership

`internal/backup` owns the existing backup document types, format/version/table policies,
file path/content checks and export/validation/import orchestration. It imports no database,
network or operating-system file API. `filepath` is used only to preserve the existing
cross-platform path grammar. Architecture tests enforce this boundary.

| Owner | Responsibility |
| --- | --- |
| `backup.Service` | Format compatibility, complete validation before restore, application-only export whitelist |
| `backup.Rules` | Existing table/domain checks and strict model-column rules |
| `infrastructure/backupdata.Store` | SQLite schema/snapshot reads, the single restore transaction, staged uploads and compensated file swap |
| `backup.CompatibilityOperation` | Version-specific row/default decisions, numbering, saved master-data authority and restore finalization through typed ports |
| `backupdata` compatibility adapters | Existing SQL and catalog/seed/decoder composition inside the store-owned transaction |

The production startup explicitly connects the store, existing domain rules and legacy
restore planner. `application.NewBackupService` remains a compatibility constructor;
public backup types and errors remain aliases. Existing callers, API/OpenAPI, format 28
retain their contracts. The extraction introduces no new data format or migration.

Snapshot export reads all application tables and transfer profiles in one read transaction.
Profile creator checks use that same transaction, including when SQLite permits only one
connection. Valid creator references remain present; missing local creators remain null.
Only the existence of those users is checked, never their authentication fields. Authentication,
sessions, setup codes, rate limits and audit tables remain outside application backups.
The service also filters unexpected tables returned by an adapter before exporting a document.

Restore first stages validated files. One database transaction then prepares legacy state,
clears/repopulates the known application tables, handles compatible profile references and
finishes the existing catalog/decoder/source reconciliation. The typed compatibility operation
uses adapters bound to that same transaction; they do not open or commit another one. Upload activation precedes database
commit, and a failed commit restores the previous upload directory and database rows.
Successful completion removes the previous/staging directories.

Compatibility reads preserve raw metadata, nullable source URLs, original IDs, integer active
states, origins and timestamps. Older article allocation skips used numbers and retains the
larger existing counter; legacy set allocation keeps its separate activation and sorting rules.
Saved article/symbol rows, missing decoder documentation, bundled origins, archived geometry,
legacy defaults, catalog projection and decoder source adoption retain their original order.
Application compatibility constructors perform no restore SQL. The store still owns the sole
restore commit and compensated file swap. Other application SQL remains explicitly inventoried.

Verification combines all previous backup/API/legacy restore tests with independent tests
that prevent invalid input/schema/domain failures reaching restore. Real SQLite fault tests
fail after deletes, during finalization and on a deferred foreign-key commit. They compare
every application table, transfer profiles, original upload bytes, staging cleanup and local
authentication before/after. Full reference upgrade/restore/browser and native Windows
qualification remain required. Actual physical-device acceptance is separate.
