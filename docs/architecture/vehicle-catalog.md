# Vehicle catalog relationships

Catalog articles describe a model; vehicles and sets describe owned inventory. Accessories retain
their independent product/quantity/asset model. The first schema step adds tables without changing
existing vehicle/set rows, reads, writes, IDs, inventory numbers or valuations.

| Table | Meaning |
| --- | --- |
| `catalog_articles` | Independent identity, explicit/manual or identified legacy origin. Manufacturer plus article number is never a deduplication key. |
| `catalog_article_revisions` | Complete immutable catalog snapshot, sequential number, adoption time and note. A correction creates another row. |
| `vehicle_article_links` | Optional revision pinned by one owned vehicle, explicit individual overrides and authority marker. |
| `vehicle_set_article_links` | Equivalent optional relationship for a set; acquisition and valuation remain on the existing set. |

The initial catalog field map is `name`, `manufacturer`, `articleNumber`, `articleSourceUrl`, `gauge`,
`epoch`, `railwayCompany`, `category`, `gattung`, `description`, `ean`, `productionPeriod`, and `listPrice`.
Overrides can explicitly contain an empty string. Every other vehicle/set field, including ownership,
purchase price/date, condition, storage, maintenance, decoder associations and technical modifications,
retains its existing owner. Unknown fields and ownership fields are rejected in catalog JSON.

Authority starts as `legacy`. The separate legacy backfill assigns each unlinked independent vehicle
and set its own article. It uses only existing set membership to pin a member to its set's revision,
with all different member values retained as explicit overrides, including empty strings. Matching
manufacturer/article numbers or descriptions never establish shared article identity. Stable IDs
derive from the original entity; an existing link or immutable revision is never replaced by backfill.
The adoption timestamp is new and is explicitly distinct from an unknown catalog source timestamp.

The same idempotent conversion runs within restore's transaction. Supported older backups retain
their exact legacy inventory and receive the missing catalog relationships. Raw legacy NULLs remain
untouched by migration; catalog JSON follows the existing API's empty-string representation. A later
legacy edit can differ from the initial snapshot; a mapping/difference report identifies that drift.
This initial backfill preserves the legacy authority. The following writer migration activates only
relationships whose revision plus overrides already equal all thirteen legacy fields. Pending values
remain in legacy mode, with their original authority retained separately in `source_authority`, even
when a previous format-25 relationship was marked `catalog`. No pending proposal is applied or lost.

In active `catalog` mode the revision and individual overrides compose the model representation.
The old vehicle/set columns remain its transactional compatibility projection, preserving existing
API responses, exports and valuation rules. All existing inventory writers use SQLite triggers to
retain the pinned immutable revision and update only the edited vehicle/set's overrides in the same
transaction. Ownership-only changes and equivalent NULL/empty catalog values do not rewrite catalog
metadata. Set edits do not propagate to existing members. Newly created set members share their
known set's revision; reused inventory retains its previous identity. Independent new vehicles and
sets receive separate articles. Matching model text never establishes shared identity.

Legacy-mode edits retain unresolved revision/override data. A relationship can become active once
its complete composed representation matches the saved legacy values. Missing new API fields
preserve the existing assignment. Explicit revision adoption is a separate operation; another owned
vehicle never receives a catalog correction automatically.

Administrators can read `GET /api/v1/catalog/legacy-backfill` for the complete assignment/difference
report. A GET never fills missing relationships. `POST` on the same route explicitly completes
missing relationships and requires the existing CSRF protection. Viewer, Editor and Messe roles
cannot use either operation. Repeating a completed POST adds neither catalog rows nor another
backfill audit event. The API reports unlinked entities and drift without silently correcting them;
generated transport types and the frontend adapter use the same contract.

`GET /api/v1/catalog/{targetType}/{id}` exposes one saved vehicle or set's current fields,
pinned revision, individual overrides, original authority and immutable article revisions. It uses
the existing inventory read roles (Viewer, Editor, Admin; no Messe access), performs no backfill
and never applies pending values. All fields are read in one transaction. Its comparison token
includes catalog identity and provenance, so changing an assignment with identical visible fields
still invalidates the previous comparison. Ownership-only edits leave this independent token intact.

`tools/regression/verify_upgrade.py` rehearses the schema on a separate copy of a marked synthetic
cold reference backup. It compares every existing table and stored file, permits only the four new
empty tables and appended migration-ledger rows, and checks the automatic SQLite snapshot against
the complete original table manifest. It refuses production/unmarked sources and existing outputs.
Its explicit `--catalog-backfill` profile additionally checks every generated article, revision and
relationship against an independent projection of the original legacy data. All original tables and
files remain subject to exact comparison. Extra rows, incorrect fields/overrides and an invented
source timestamp fail the comparison. The additional `--catalog-authority` profile independently
derives the exact authority/source-authority conversion, while preserving original revisions,
overrides, timestamps and every other column. The native Windows runner uses both profiles.

Backup format 26 includes all four tables and the original-authority marker. Preflight validates complete revision
fields, overrides, revision numbering, origin, timestamps and relationships against the backup itself.
Missing required tables, unknown catalog columns, dangling links and manually lowered format numbers
are incompatible. Active catalog relationships must exactly compose their legacy projection.
Formats 1–25 remain supported; missing catalog relationships are backfilled only
from the restored legacy representations. Restore suspends exactly the six projection triggers
inside its transaction, imports the complete inventory/catalog, converts old authority markers and
reinstates the triggers before commit. Failure rolls back both data and trigger definitions. Format
26 preserves its supplied markers without reclassification. Local authentication remains outside
application backups.

Before migration, retain a consistent full installation, application export and matching old binary.
The migration is additive and the automatic SQLite migration snapshot remains an extra safeguard.
Once new catalog data exists, an old binary or format 24 cannot preserve it. Recovery then uses a
compatible current application or the complete pre-migration installation during a write pause.
Restoring that older installation after new input would discard the newer input; it is not a lossless
rollback. Never lower an exported format number to bypass compatibility checks.
