# Exhibition planning boundaries

`exhibition` owns list and entry use cases, lifecycle transitions, revision checks,
normalization, day overlap, readiness, conflict identities and explicit exceptions.
Its typed repository and mutation ports contain no SQL, filesystem or transport handles.
`infrastructure/exhibitiondata` owns the unchanged SQLite statements, row scanners,
nullable fields and transaction implementation. Application aliases preserve callers,
public records, JSON fields and error identities. No schema migration is introduced.

List creation, header edits, status changes and deletion retain their existing individual
SQL commits. Entry creation, edits and deletion each commit the entry and guarded list
revision in one transaction. Conflict exceptions and their expected list revision also
share one transaction. The use case owns each commit or rollback. Postcommit readback
errors do not undo committed data; stale revision retries remain rejected.

The existing ordering and validation precedence remain intentional. Entry updates check
existence before input validation and read the current revision again. Legacy revision zero
remains supported. Sort-number query failures retain the fallback value. Nullable strings
map to existing empty fields; stored spellings are not normalized during reads. Boolean
flags are true only for the stored value one. Submitted inputs retain their old trimming,
day-scope and availability rules. Historical lifecycle timestamps remain sticky.

Conflict IDs retain their ordered fields, lowercase protocol/address and first twelve SHA256
bytes. Confirmed exceptions remain visible and count in existing readiness/lock confirmation.
Workspace status still overrides the embedded entry status in JSON. The list date range and
all lifecycle transitions remain unchanged. No inventory row or physical device is modified.

Messe, Editor and Admin permissions remain in the existing server routes, with CSRF protection.
Messe can plan entries and exceptions; header/status changes require Editor and deletions Admin.
General inventory access is not added to Messe. Application backups continue to preserve lists,
entries and exceptions while excluding local authentication.

Real SQLite tests inject late list-write, locked-guard, revision and deferred commit failures,
compare every stored table, and retry using the original revision. Additional fixtures verify
lost readbacks, raw legacy fields, stable conflict identities and sticky lifecycle timestamps.
The application SQL ratchet is now empty: any new concrete SQL call fails architecture tests.
