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
Legacy fields remain authoritative throughout this phase. Switching to `catalog` requires all
writers and their transactional legacy projection and is a separate step. Missing future API fields
must preserve the association.

`tools/regression/verify_upgrade.py` rehearses the schema on a separate copy of a marked synthetic
cold reference backup. It compares every existing table and stored file, permits only the four new
empty tables and appended migration-ledger rows, and checks the automatic SQLite snapshot against
the complete original table manifest. It refuses production/unmarked sources and existing outputs.
Its explicit `--catalog-backfill` profile additionally checks every generated article, revision and
relationship against an independent projection of the original legacy data. All original tables and
files remain subject to exact comparison. Extra rows, incorrect fields/overrides and an invented
source timestamp fail the comparison. The native Windows reference runner uses the same profile.

Backup format 25 includes all four tables before they are used. Preflight validates complete revision
fields, overrides, revision numbering, origin, timestamps and relationships against the backup itself.
Missing required tables, unknown catalog columns, dangling links and manually lowered format numbers
are incompatible. Formats 1–24 remain supported; missing catalog relationships are backfilled only
from the restored legacy representations. Local authentication remains outside application backups.

Before migration, retain a consistent full installation, application export and matching old binary.
The migration is additive and the automatic SQLite migration snapshot remains an extra safeguard.
Once new catalog data exists, an old binary or format 24 cannot preserve it. Recovery then uses a
compatible current application or the complete pre-migration installation during a write pause.
Restoring that older installation after new input would discard the newer input; it is not a lossless
rollback. Never lower an exported format number to bypass compatibility checks.
