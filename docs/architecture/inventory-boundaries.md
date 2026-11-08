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

`inventory/masterdata` owns the shared records, protected article-type configuration,
input/custom-field definitions, import preparation, sorting, decoder metadata checks
and cache-copy rules. It imports the existing decoder-library parser and shared domain
attribute rules. It has no SQL, HTTP or filesystem dependency.

`infrastructure/masterdatastore` owns the five existing list/snapshot/single-entry/
relation queries through a typed reader port. Its row mapper is also reused by the
existing caller-owned reconciliation transaction. The application retains its service
and transactional writes during the next migration slice, including decoder/seed methods
that require a local service type. This first slice does not complete that write migration.

Cached metadata owns nested maps and arrays. Both active and complete snapshots, and
returned copies, retain their own mutable containers; changing one cannot alter another.
Scalar values, original profile strings and absent-versus-empty arrays retain their exact
representation. The copy shares immutable string contents and uses no JSON roundtrip.
