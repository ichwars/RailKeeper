# Technical follow-up R0–R9

This contract continues the quality work after the November 2026 reassessment. Existing
inventory, valuations, decoder/CV data, authentication and public contracts must survive
each step. Runtime evidence and private backups remain outside Git.

## Scope and completion

| Package | Deliverable | Required evidence |
| --- | --- | --- |
| R0 | Owned persistence inventory, transaction boundaries and regression guard | Every remaining file has a stage and target owner; new SQL execution in application fails architecture tests. |
| R1 | Bounded login waiting with validated configuration | Burst, queue exhaustion, cancellation, timeout and resource-bound tests; supported hashes remain valid. |
| R2 | Separate budgets for cryptographically recognized login origins | No IP-based trust; full password/2FA required; revocation, recovery, audit and auth-backup isolation tested. |
| R3 | Account, session, password and 2FA use cases behind persistence ports | Last-admin rules and atomic mutation/session revocation/audit preserved; completed SQL inventory removed. |
| R4 | Transfer policy and per-entity persistence separated | One import commit owns rows, job state and audit; preview revisions and replay guards remain. |
| R5 | Remaining inventory/catalog, decoder/CV, backup and configuration boundaries | Each named exception is migrated; domain rules and persistence have actual independent owners. |
| R6 | Clear initial view selection and everyday labels | Existing preferences preserved; all permitted functions reachable; DE/EN, keyboard and responsive checks. |
| R7 | Behavioral test reserve and supported additional analysis | Measurement scope/floors retained; new packages included; typed lint only with verified compiler support. |
| R8 | Concise entry documentation and reproducible operations | Details stay available; reviewable changes; build/restore instructions usable without session knowledge. |
| R9 | Qualified final source and safe KI test deployment | Same candidate identity across platforms; current installation snapshot, isolated verification and preserved live data. |

Human pilot participation, independent human review and physical device acceptance remain
R10 evidence. The technical R0–R9 work prepares those activities without inventing their
results. Offline capture and multi-site replication remain separate product decisions.

## Persistence inventory

The initial inventory contains 57 production files and 305 concrete database API calls,
including transaction creation. Fifty-six files issue Query/Exec calls. The baseline records
47 functions that create transactions. It is a source inventory, not a claim that all
operations share one transaction or that lexical counts prove semantic safety.

`backend/internal/architecture/application_sql_baseline.json` identifies every remaining
file, the follow-up package and intended persistence owner. Transaction function names
record the present owners. The architecture test rejects new files with concrete SQL,
increased call counts and stale completed entries. As migrations land, the inventory must
shrink. Renaming files or wrapping SQL without moving the responsibility does not complete R5.

| Existing responsibility | Current atomic boundary | Target and protected contract |
| --- | --- | --- |
| Login | `authdata.Store.CommitLogin` | Account recheck, TOTP consumption, hash upgrade, session, audit and budget reset commit together. |
| Accounts and sessions | AuthService mutation methods in the inventory | R3 account store owns each account/role change and its session revocation/audit; the last administrator remains protected. |
| Setup | SetupService/setup-authorization transaction | Exactly one setup-code consumption, account, roles and audit; existing installations do not reinitialize. |
| Vehicle/catalog edits | VehicleService mutation and catalog transaction functions | R5a preserves revision checks, article/set relationships, creation identities, audit and the caller's single commit. |
| Decoder/CV | VehicleService mutation/checkpoint transaction functions | R5b keeps profile bytes, absent/zero index identity, original source bytes and frozen partial-state descriptions. |
| Application restore | `backupdata.Store.Restore` | Staging precedes row replacement; file activation is compensated if commit fails. Legacy callbacks use that same transaction. |
| Transfer import | `DataTransferRepository.ApplyImportWithPolicy` | Preview/fingerprint revalidation, entity changes, completed job revision and audit use one commit. |
| Exhibition | ExhibitionService workspace transactions | Roles, entries, print/report data and revisions remain consistent. |
| Settings and SMTP | Current per-setting persistence operations | Existing local configuration and authentication exclusions remain; do not invent broad transactions across external I/O. |
| Maintenance and seeds | Existing exclusive maintenance or seed transaction | Preserve maintenance locking and SQLite operations that cannot run inside a transaction; no generic transaction wrapper. |

Adapters may accept an existing caller-owned transaction while a feature is being migrated.
They must not open nested commits. Network calls must not be introduced into SQL transactions.
Source freezing retains its existing atomicity until a separately tested equivalent design exists.

## Data and behavior preservation

- Structural packages normally require no schema or data migration. A necessary auth-only
  migration is additive, uses the next free number and remains outside application backups.
- Keep IDs, links, decimal values, currencies, original profile spelling, unknown physical
  dates, CV pages, source provenance, immutable histories and file bytes.
- Never treat purchase/list prices as historical valuations. A correction or withdrawal
  adds history; it does not overwrite an original entry.
- Preserve role checks, CSRF, bounded password hashing, one-use TOTP, session revocation,
  private data paths, conflict handling and retained drafts after uncertain responses.
- Each delivery captures the then-current installation and matching software. An earlier
  backup is not permission to overwrite later user changes. Compare against the immediate
  cutover snapshot before reopening the service.

## Build and analysis policy

The module and release builds use Go 1.27.4. `.github/scripts/check_go_toolchain.py` verifies
that their pins match and checks the dedicated minimum-compiler CI path. The qualified
golangci-lint 2.14.0 analyzer uses the same compiler.

Frontend lint continues to require hook rules and strict TypeScript compilation. Additional
type-aware rules must support the actual project compiler. Unsupported versions, broad rule
disables, reduced measurement scope or mechanical `void` insertions are not substitutes.
The current compatibility decision and any remaining tool limitation must be recorded honestly.

Behavioral reserve is a measured target, not a lowered gate: approximately 83/78/78/83 for
frontend statements/branches/functions/lines and 77 percent total backend coverage. Existing
module floors and critical failure cases remain mandatory.

## Review and delivery

Keep semantic changes separate from mechanical moves. Describe the concrete behavior,
transaction owner, compatibility and actual validation in every PR, German first and English
second. Complete required checks; do not rerun successful identical suites without a changed
source or unresolved concern. New modules retain architecture and coverage checks.

R9 qualifies Linux, native Windows and AMD64/ARM64 images for one selected final source.
An isolated restore and the synthetic valuation/decoder workflows precede the KI update.
Retain protected full installation backups outside Git and verify configuration, database
integrity, original rows/files, build identity and protected routes after cutover.

## Bounded login work (R1)

Login work retains two active operations by default, with at most eight queued
requests and a one-second wait. Deployment flags are validated before startup;
invalid or insufficient password-memory budgets fail explicitly. Cancellation
never frees an active derivation early. A full queue responds with HTTP 429 and
`Retry-After: 1`; the five-minute account budget remains unchanged.

The real-Argon2 burst and repeated maximum-memory tests exercise queue recovery
and the accepted 256-MiB hash limit. Resource measurements inform installation
sizing; the password-work budget is not a process RSS cap. See the
[operator configuration](production-runbook.md#begrenzte-passwortprüfung).

## Recognized browser budgets (R2)

An opaque random 256-bit browser token is issued only with a committed, complete
password and second-factor login. A per-account HttpOnly/SameSite cookie follows
the installation's Secure policy and expires after thirty days. SQLite stores
only token hashes, account-bound budget keys and a credential-state fingerprint;
at most ten browsers per account are retained. These auth-only records are never
part of application backups. Restore preserves the target installation's auth.

Unknown origins retain the ten-attempt/five-minute account budget. A recognized
browser uses an independent ten-attempt budget and does not clear an attacker's
unknown-origin window. One queue slot is reserved when there are at least two;
mixed queued classes alternate with FIFO order within each class. Recognition
lookups have separate bounded concurrency and a one-second deadline. All origins
still require the password and any second factor and obey IP/global work limits.

Credential and role changes revoke proofs. TOTP consumption alone does not.
An authenticated administrator can reset a user's budgets and revoke proofs in
one CSRF-protected, audited transaction. Existing passwords and sessions survive.

Threat limits: a new browser has no protected proof until a complete login.
Shared NAT can still exhaust the unchanged IP limit. A stolen proof can consume
its own protected budget but cannot authenticate; distributed sustained overload
can still exhaust global capacity. Recovery requires an existing admin session,
another permitted origin, or expiry of the five-minute window. Independent human
security review remains part of R10, alongside the documented automated checks.

## Account and session boundaries (R3)

`authentication/management` owns account validation, role inheritance, session/CSRF
policy, password/reset decisions and second-factor orchestration. It uses typed
persistence ports and has no SQL, HTTP or filesystem dependencies. Its own coverage
floor preserves the former application floor of 78%; the critical login module
retains its independent 98% floor and the total measurement scope is unchanged.

`authdata` owns account/role mutations, session persistence, reset requests and
second-factor commits in focused adapters. Password derivation happens before
account-update transactions. `application.AuthService` composes these modules and
preserves existing transport types through aliases. The three completed SQL files
leave the inventory: 54 files and 254 concrete database calls remain for R5.

Account mutations retain one SQLite owner. Session revocation and logout now include
successful audit in that transaction. Reset requests cannot invalidate an older
link when replacement insertion/audit fails. Reset completion rechecks the token,
expiry and one-use state after beginning the transaction. Password changes compare
the originally verified hash before writing. Rejected missing-user/last-admin paths
release their transaction. Fault injection and concurrent reset tests cover these
additional atomicity guarantees separately from the structural move.

## Transfer rules and import writers (R4, first part)

`transfer/policy` owns concrete parsing, field mapping/preservation, preview classification,
set grouping, conflict resolutions and output encoding. Ordinary writes and transfer reuse
the same set normalization/required-field rules; vehicle/accessory aggregate validation
remains an explicit port to the existing shared validators until R5a. Missing validators
fail closed. Helpers remain private unless needed by another layer or contract test.

Vehicle, accessory and exhibition import writers are separate adapters. Set membership,
preview/fingerprint checks, job completion and audit retain the caller's single transaction.
An architecture guard rejects nested transaction owners. Every original SQL literal remains
unchanged.

## Transfer workflows and private files (R4, second part)

`transfer/workflow` owns actual profile/job orchestration through typed repository
and file ports. `infrastructure/transferfiles` owns upload staging and export effects;
`application.DataTransferService` is compatible composition. Architecture tests reject
concrete I/O imports in the completed facade and in the pure workflow. Optional
repository capabilities remain explicit, and file metadata/seek preserve partial downloads.

File failures retain compensation: an export without committed metadata is removed,
and preview readers are closed on parsing/persistence failures. New injected read and
oversize tests exposed a pre-existing scratch-file leak caused by overwritten named
returns; cleanup now uses the immutable temporary pathname. Existing stored bytes,
formats, confirmation rules and transaction ownership remain unchanged. The new workflow
and file adapter inherit the previous application coverage floor of 78% with the same
whole-backend measurement scope.

## Numbering and collection overview (R5a, first package)

`inventory/numbering` owns actual validation, fallback/formatting and bounded reservation
decisions behind semantic scheme/counter ports. `numberingdata` executes the existing
queries and counter advancement in the caller's transaction. `inventory/overview` owns
actual collection arithmetic and error handling through typed rows; `overviewdata`
retains the four existing SQL queries. Shared legacy money rules live in `inventory/money`.

The two completed application files leave the SQL inventory: 52 files and 243 concrete
calls remain for the named R5 owners. IDs, counters, membership-based price selection,
stored price strings and currency exclusions retain their existing semantics. Immutable
valuation-history events remain distinct. Each extracted module retains the former 78%
application coverage floor, with the complete backend still instrumented.

## Master-data policy and reads (R5a, next slice)

Shared master-data models, protected-type/custom-field/import preparation and decoder
metadata validation live in the independent `inventory/masterdata` module. Five actual
SQL reads move to `masterdatastore` behind a typed reader port. Transactional writes
and service orchestration remain explicit for the next slice: 52 files/238 application
SQL calls still remain, with two calls in `master_data.go`.

A deterministic ownership test exposed shared nested metadata in cached responses.
Cache copies now own mutable maps/arrays independently, including both active and
complete views. Recorded number types, profile strings and missing/empty values are
preserved. Structural extraction and this behavior correction are separate commits.

## Master-data use cases and transactions (R5a, write slice)

The independent master-data core owns actual create/update/import, management,
cache and lifecycle decisions through semantic mutation ports. Focused SQL adapters
own custom-field, batch, deletion and reconciliation transactions, with usage/reference
callbacks bound to the same reserved write transaction. Existing SQL literals remain
identical. Cache invalidation precedes readback after a successful write; failed writes
retain the previous cache.

The four completed master-data files leave the shrinking SQL inventory: 48 files and
217 concrete calls remain, including separately assigned legacy decoder/symbol seeds.
Application composition retains a local service receiver for those existing methods.
Architecture tests guard the four transaction owners. Additional SQL-trigger failures
verify rollback of earlier batch rows, cleared/imported records and removed relations,
including identities, raw metadata and timestamps.
