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

Authority starts as `legacy`. Additional tables are initially empty; a separate backfill preserves
each existing representation without inferring article identity. Known set membership may preserve
its common catalog snapshot with explicit member differences. Neither matching manufacturer/article
numbers nor matching descriptions establish a shared article. Backfill is a later, separately reviewed
step. Switching authority to `catalog` requires all writers and their transactional legacy projection;
it is not enabled by this schema change. Missing future API fields must preserve the association.

`tools/regression/verify_upgrade.py` rehearses the schema on a separate copy of a marked synthetic
cold reference backup. It compares every existing table and stored file, permits only the four new
empty tables and appended migration-ledger rows, and checks the automatic SQLite snapshot against
the complete original table manifest. It refuses production/unmarked sources and existing outputs.

Backup format 25 includes all four tables before they are used. Preflight validates complete revision
fields, overrides, revision numbering, origin, timestamps and relationships against the backup itself.
Missing required tables, unknown catalog columns, dangling links and manually lowered format numbers
are incompatible. Formats 1–24 remain supported with missing catalog tables restored empty in this
first step. Local authentication remains outside application backups.

Before migration, retain a consistent full installation, application export and matching old binary.
The migration is additive and the automatic SQLite migration snapshot remains an extra safeguard.
Once new catalog data exists, an old binary or format 24 cannot preserve it. Recovery then uses a
compatible current application or the complete pre-migration installation during a write pause.
Restoring that older installation after new input would discard the newer input; it is not a lossless
rollback. Never lower an exported format number to bypass compatibility checks.
