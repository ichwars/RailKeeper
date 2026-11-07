# Transfer contracts and transaction ownership

`internal/transfer` owns the shared exchange records, profile/job repository contract,
CSV mapping records, package encoding for versions 1–3 and target fingerprint versions 1–2.
It imports the standard library and shared domain records. It performs no SQL, filesystem,
HTTP or device operations. Architecture tests enforce this boundary.

`application` retains compatibility aliases and delegates fingerprint/name identity functions
to that package. Existing JSON field names, field order, omitted fields, selected empty areas,
Unicode case folding and stored preview fingerprints remain compatible. Shared set input and
exhibition status types live in `domain`; both application and transfer use the same types.

| Owner | Responsibility |
| --- | --- |
| `transfer` | Exchange types, empty/omitted area distinction, legacy encoding and exact comparison identities |
| `application.DataTransferService` | File limits, preview workflow, user decisions and permitted import/export areas |
| `infrastructure.DataTransferRepository` | Actual SQLite reads, revision checks, import application, job transitions and audit |
| Existing vehicle/accessory validators | The same aggregate rules used by regular inventory writes and transfer application |

Profile/job persistence and snapshot reads now use the independent contracts directly.
The legacy apply adapter still calls existing application aggregate validators; the entire
infrastructure package is therefore not claimed to be independent of application yet. A future
move of those validators must preserve their shared rules rather than duplicate them.

The existing repository remains the only import transaction owner. Target comparisons,
vehicle/set/accessory/exhibition mutations, job state and audit still commit together or roll
back together. This contract extraction adds no writer, migration, changed confirmation or
new data format. It does not broaden the objects contained in a partial transfer package.

The verification combines fixed fingerprints captured from the previous implementation,
selected empty area encoding, Unicode identity tests and existing SQLite/API tests for
legacy package import, concurrent confirmation, changed targets and rollback. Full backend,
OpenAPI and reference preservation checks continue to apply.
