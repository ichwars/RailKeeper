# Decoder identity, CV rules and persistence

`internal/decoder` owns physical decoder and installation records, editable CV workspaces,
CV comparison/scope decisions and immutable partial-state validation. It imports neither SQL,
HTTP, filesystem nor device transports. `internal/revision` preserves the existing JSON/SHA-256
comparison encoding. Architecture tests enforce both boundaries.

| Owner | Responsibility |
| --- | --- |
| `decoder` | Input rules, physical identity, installation/removal/date plans, workspace assignment, CV page/source/revision decisions and strict snapshot validation |
| `infrastructure/vehicledata` | Configuration/content reads, binding reads, CV/file record mapping and focused hardware/workspace SQL |
| `decoder/workflow` | Actual configuration, checkpoint, CV/file CRUD and batch/import use cases through semantic transaction ports, revision checks, audit and commit |
| `infrastructure/vehicledata`, `infrastructure/blobdata` | SQLite transaction implementations, confined source-file reads, blob freezing, historical source fallback and legacy adoption within the existing caller transaction |
| `VehicleService` | Compatible public methods, error identities, parent lookup and composition |

Existing application types remain aliases. Public fields, error identities/status mapping,
backup format 28 and existing vehicle/CV records remain unchanged. This extraction introduces no migration, second
CV store or independent writer. Legacy clients can still omit CV comparison tokens; requests
that supply a token retain their optimistic conflict checks. New workspace profiles are trimmed;
stored profile spelling, including whitespace, is never rewritten. Missing index pages remain
different from explicit zero pages.

Hardware plans distinguish a physical decoder instance from its installation episodes and
from a CV workspace. Existing imported decoder identifiers remain editable. Unknown installation
and removal dates stay null; recording time is never substituted for a physical event date.
Removal retains the vehicle reference and original recording time. A date correction does not
silently change an installed decoder into a removed one.

A workspace assignment checks the current installation and captures the previous partial state
before updating its binding. The capture, immutable original-source blobs, assignment and audit
commit together. A failed audit or write rolls everything back. Global workspace and snapshot
identifier checks run against the same transaction, including identifiers on another vehicle.
An already prepared CV request is rejected if its original binding changed; it is never retargeted.

Configuration revisions still include CV histories and private source storage identities.
The exact record field order, sorted reads and explicit hidden path/blob fields are preserved.
Golden tokens cover old CV comparisons and workspace bindings. SQLite integration tests also
change content without changing timestamps, proving stale confirmations remain detectable.

Historical reads and captures run through `decoder/workflow`; its read ports require only rollback,
while mutation ports retain the existing commit owner. Historical snapshots remain readable after
the current vehicle or workspace is removed. CV writes retain value/history/workspace atomicity,
source ownership, removed-original evidence and exact nullable page identities. File creation
retains its original timestamp precision. Value update commits before its separate public readback;
delete returns pre-delete history. ECoS memory adoption preserves existing full addresses under
every profile and performs no device I/O. Remaining seed/backup compatibility and other application
SQL stay explicitly inventoried; the whole application package is not yet SQL-free.

Verification includes independent domain decisions, bounded snapshot fuzzing, existing real-SQLite
atomicity/history/source/restore tests and the same candidate's CV, decoder and archive browser
flows. All device endpoints used by automated tests are simulated; no physical device acceptance
or newly supported programming capability follows from this change.
