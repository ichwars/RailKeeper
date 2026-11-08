# Inventory rules and storage boundaries

Inventory migration proceeds by use case. The remaining concrete SQL inventory in
`internal/architecture/application_sql_baseline.json` names each unfinished owner;
completed entries are removed and architecture checks reject their reintroduction.

## Numbering

`inventory/numbering` owns scheme normalization, validation, category fallback,
formatting and the bounded search for an available number. Typed scheme and counter
ports contain no database types. `infrastructure/numberingdata` owns actual SQL,
nullable/scanner mapping and timestamps for stored schemes and counter advancement.
The application constructor and existing transport records remain compatible aliases.

The counter adapter receives the inventory mutation's existing transaction. It never
begins, commits or rolls back that transaction. Number availability, advancement and
the owning vehicle/set/accessory creation therefore retain the caller's atomicity.
Conflict candidates advance only the local search until a free candidate is found;
storage or availability failures never acknowledge a reservation. The existing limit
of 500 candidates, fallback order, prefixes, padding, IDs and recorded counters remain.
An omitted fallback retains its original behavior rather than silently choosing another
category. The embedded input field name remains compatible with existing Go callers.

## Overview values

`inventory/overview` owns collection-value calculations through a typed row source.
`infrastructure/overviewdata` owns the four existing queries: standalone vehicles and
nonempty vehicle sets, accessory ownership, purchases and manual individual-asset prices.
Set membership and the distinction between stock, individual assets and installations
remain in those exact queries. Query rows preserve their stored strings.

`inventory/money` owns the existing integer-cent parsing, formatting and checked
arithmetic. Recorded price formats and euro decorations retain their previous rules.
Unparseable/absent prices remain uncounted, foreign-currency purchases retain their
explicit excluded count, and negative/overflow arithmetic returns an error. A failed
source never yields an acknowledged partial collection total.

Immutable valuation history and corrections remain owned by `valuation`. The overview
continues to sum recorded acquisition/list prices and does not create history events.
No schema or JSON-contract change is required by these boundaries.

## Master-data rules and reads

`inventory/masterdata` owns shared records, protected article-type configuration,
input/custom-field definitions, import/reconciliation decisions, sorting, decoder metadata
checks, cache ownership and the actual use cases. Its typed repository includes semantic
read and mutation operations. It has no SQL, HTTP or filesystem dependency.

`infrastructure/masterdatastore` owns existing reads and writes in focused adapters.
Custom-field changes, active-state batches, deletion and import each retain one transaction
owner. Reference/usage callbacks use that same transaction after reserving the write lock.
Original entry/relationship identities, bundled origins, raw metadata and timestamps retain
their rules. Injected later failures roll back earlier updates, deletions and inserts.

After a successful write the core invalidates its cache before final readback, preserving
the invalidation even if readback fails. A rejected write retains the valid previous cache.
Management capability calculations remain conservative for unknown reference types.

The application service is local composition embedding the real core, preserving its extra
decoder/library methods. Its database field remains only for explicitly inventoried legacy
decoder/symbol seeds, which belong to their separate migration. It executes no master-data
read/write SQL. Backup compatibility owns its separate bundled-identity row type until that
legacy restore adapter is migrated.

Cached metadata owns nested maps and arrays. Both active and complete snapshots, and
returned copies, retain their own mutable containers; changing one cannot alter another.
Scalar values, original profile strings and absent-versus-empty arrays retain their exact
representation. The copy shares immutable string contents and uses no JSON roundtrip.

## Maintenance and spare parts

`inventory/workshop` owns maintenance defaults, date/cost validation, spare-part identity
and duplicate merging, and the actual CRUD use cases. Typed parent lookup and storage
ports keep those decisions independent of SQL. The application composes the parent lookup
with its existing aggregate read, preserving the previous lookup and failure behavior.
Aggregate assembly calls the row adapter directly to avoid recursive parent lookup.

`infrastructure/workshopdata` owns the ten existing queries and mutations. Every record
lookup, update and deletion retains its vehicle scope. Existing IDs, creation times,
price strings, source URLs, row ordering and empty dates retain their original behavior.
A duplicate spare part keeps its existing identity and recorded fields according to the
original merge rules. A database failure cannot be acknowledged as an empty list or a
successful mutation. Dedicated adapter tests verify cross-vehicle isolation and outages.

Shared transport records and the original validation/not-found error identities live in
`domain`; the application retains compatible aliases. This split changes neither schema
nor JSON contracts and introduces no additional transaction or filesystem effect.

## Vehicle and set catalog revisions

`inventory/catalog` owns field decoding and validation, revision creation/adoption,
preservation of unresolved proposals, selected-field projection and comparison reports.
Its repository and mutation ports contain no SQL types. Repeated adoption of an already
selected revision returns the existing context without writes, audit or commit.

`infrastructure/catalogdata` owns catalog reads and mutations, revision numbering,
audit records and projection-trigger maintenance. One transaction covers the current
comparison, any retained proposal, article/revision creation, selected-field projection,
audit and final context read. The use case requests commit only after all succeed.
Failure rolls back the same transaction. Read snapshots retain their previous transaction
boundaries and comparison-token encoding.

Projection updates only changed selected fields in their original order. Unselected raw
NULLs, acquisition/ownership fields, other vehicles and set members retain their values.
Individual overrides, original authority markers and unknown source times are preserved.
The existing immutable revisions and all JSON shapes remain compatible aliases.

Legacy backfill and trigger suspension/restoration also execute in this adapter. Restore
passes its existing transaction to those helpers; they never begin or commit a separate
one. Backfill reports cross the port as typed raw rows before pure comparison/aggregation;
malformed stored fields fail the report instead of acknowledging a partial result.

## Vehicle records, rules and core reads

`inventory/vehicle` owns shared vehicle/set/media records, normalization, operational and
function-input rules, input projections and the existing core/media comparison encodings.
Application records and `vehicledata` write rows remain compatible aliases. Field names,
JSON tags, embedded field names, stored price strings and error identities do not change.
Golden tests derived from the previous record definitions protect outstanding draft tokens,
including private file references and the separate treatment of set main images.

`infrastructure/vehicledata` owns the concrete single-vehicle and list queries, including
nullable field mapping and set summaries. Query text, filter scope and ordering remain
unchanged. Aggregate enrichment retains its existing application owner and read order.
Vehicle/set mutation, set-image transactions and exhibition-expiry effects still have their
explicit remaining entries; this shared foundation does not claim those migrations complete.

## Vehicle images

The independent `inventory/vehicle.ImageService` owns image creation/deletion, parent and
maintenance-reference decisions, first/selected primary image rules, optional draft guards
and acknowledgement tokens. Typed image/mutation ports contain no SQL. The existing parent
callback preserves the aggregate lookup and exhibition-expiry behavior before image creation.

`vehicledata` executes the actual scoped image reads, append/delete, primary-flag updates
and commits. The existing revision check, primary updates, image mutation and final list
token read retain one transaction. A stale comparison or later failure rolls it all back.
Maintenance ownership is checked again within that same transaction before image creation.

Whole-vehicle image replacement calls the independent `ReplaceImages` policy through a
narrow adapter wrapping the vehicle mutation's existing transaction. It never starts or
commits another one. Matching by recorded ID and then URL retains original IDs, creation
times, private paths, blob references and file metadata. Stored zero order values retain
their previous position rule; changed titles and source descriptions keep their old rules.
Private-file metadata in the input cannot replace the original retained metadata.

Set-image mutations remain explicitly inventoried under their existing owner until their
separate follow-up slice completes.

## Vehicle documents and private file references

`inventory/vehicle.AttachmentService` owns document creation, scoped metadata edits and
deletion through typed repository/mutation ports. Existing parent lookup, maintenance
ownership, optional revision guards and acknowledgement encodings keep their previous order.
Metadata update/deletion retain their transaction, including readback before commit for an
update. The original single-statement document creation keeps its existing boundary.

`vehicledata` owns the unchanged scoped reads/inserts/updates/deletes and the private-path
reference query. Filesystem path normalization and legacy NULL handling remain there;
blank paths still return zero without a query. The count includes original images,
thumbnails, documents and CV files, preserving shared legacy files until their final
reference is removed. No file activation or deletion is added to these SQL transactions.

Real SQLite fault tests remove a just-updated row from its transaction's readback. The
failed use case rolls back both metadata changes and that removal, retaining every original
field and private file reference. A separate scoped deletion test preserves another
vehicle's document. Existing HTTP revision and shared-file deletion tests remain in CI.
