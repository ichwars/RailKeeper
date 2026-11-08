# Transfer contracts and transaction ownership

`internal/transfer` owns the shared exchange records, profile/job repository contract,
CSV mapping records, package encoding for versions 1–3 and target fingerprint versions 1–2.
It imports the standard library and shared domain records. It performs no SQL, filesystem,
HTTP or device operations. Architecture tests enforce this boundary.

`transfer/policy` owns field catalogs, CSV mapping/decoding, package decoding, preview
classification, conflict/group decisions, snapshot ordering and export encoding. It has
no SQL, HTTP or filesystem dependency. Typed aggregate-validation callbacks reuse the
same vehicle/accessory rules as ordinary inventory writes; absent callbacks fail closed.
Set normalization/required-field rules are shared through `domain`.

`transfer/workflow` owns profile/job use cases, permitted areas, explicit confirmation,
preview persistence and failure/retry decisions. Its typed file port exposes managed
upload readers, artifact readers and publication. It has no SQL, HTTP, filesystem
or concrete OS-file types. Optional import/export repository capabilities remain
explicit; the base profile/job port does not falsely advertise them.

`infrastructure/transferfiles` owns private import staging and export files, MIME
inspection, hashes, upload limits, path checks and cleanup. Export publication keeps
the temporary-file/rename protocol and removes published content if recording its
metadata fails. Upload readers remove their scratch file on close. Failed/oversized
uploads also clean the immutable scratch pathname, even when named return values
have already been replaced by an error return. Artifact readers retain seek and
modification metadata for HTTP Range/conditional downloads.

`application` retains compatibility aliases and delegates fingerprint/name identity functions
to that package. Existing JSON field names, field order, omitted fields, selected empty areas,
Unicode case folding and stored preview fingerprints remain compatible. Shared set input and
exhibition status types live in `domain`; both application and transfer use the same types.

| Owner | Responsibility |
| --- | --- |
| `transfer` | Exchange types, empty/omitted area distinction, legacy encoding and exact comparison identities |
| `transfer/policy` | Parsing, field preservation, preview/group decisions, allowed resolutions and export encoding |
| `transfer/workflow` | Profile/job orchestration, user decisions, permitted areas and typed file ports |
| `infrastructure/transferfiles` | Private staging/publication, byte metadata, bounded uploads, path confinement and cleanup |
| `application.DataTransferService` | Compatible alias and composition of the workflow, validators and file adapter |
| `infrastructure.DataTransferRepository` | Actual SQLite reads, revision checks, import application, job transitions and audit |
| Existing vehicle/accessory validators | The same aggregate rules used by regular inventory writes and transfer application |

Profile/job persistence and snapshot reads use the independent contracts directly.
Missing-record errors retain both the domain sentinel and the legacy SQL sentinel,
allowing the pure workflow to avoid database dependencies without breaking old callers.
The legacy apply adapter still calls existing application aggregate validators; the entire
infrastructure package is therefore not claimed to be independent of application yet. A future
move of those validators must preserve their shared rules rather than duplicate them.

The existing repository remains the only import transaction owner. Vehicle, accessory
and exhibition writers now live in focused adapter files; set writers retain their
shared caller transaction. The apply coordinator owns exactly one begin/commit, checked
by architecture tests. Preview and apply share the same allowed-resolution policy.
The writer split preserves every existing SQL literal (70 occurrences), while existing
SQLite/API tests retain the actual behavior and rollback checks.

Target comparisons, vehicle/set/accessory/exhibition mutations, job state and audit
still commit together or roll back together. These boundaries introduce no migration,
changed confirmation or new format. Partial transfer packages retain their scope.

The verification combines fixed fingerprints captured from the previous implementation,
selected empty area encoding, Unicode identity tests and existing SQLite/API tests for
legacy package import, concurrent confirmation, changed targets and rollback. Full backend,
OpenAPI and reference preservation checks continue to apply.
