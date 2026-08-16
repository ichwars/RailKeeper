# RailKeeper Accessories Guide Design

Date: 2026-08-16

Status: Approved in conversation
Documented release: `v0.1.17.6`

## Objective

Publish a complete English and German user-guide section for the stable accessories workspace.
The section guides collectors, clubs, planners, and workshop users through the complete lifecycle
of an accessory article:

1. find and inspect articles in the accessories overview;
2. create or maintain the article record and its technical data;
3. choose the correct stock strategy and manage storage, purchases, and documents;
4. reserve or install stock at a valid target;
5. review usage history, archive obsolete articles, or delete them with the correct authority.

The section explains every stable user-facing action in this scope. It separates product identity,
quantity stock, individually tracked assets, reservations, installations, and historical records so
that users can predict the effect of each write before confirming it.

## Scope

The section owns these stable user workflows:

- accessories overview metrics, search, filters, sorting, selection, and responsive view modes;
- article read, create, edit, duplicate review, archive, restore, and permanent deletion;
- accessory article types, subtypes, common fields, type-specific technical data, and custom fields;
- accessory barcode and external article-data search;
- quantity-based and individually tracked inventory strategies;
- stock by storage location, adjustments, transfers, movements, and minimum-stock state;
- purchases and optional booking of a purchase into stock;
- individual accessory assets and their lifecycle;
- accessory images and documents, including URL import, metadata, download, and deletion;
- reservations, cancellation, installations, condition changes, removal, and disposition;
- allocation summaries and usage history;
- permissions, confirmation steps, stale-resource handling, partial writes, backup, and recovery.

## Non-goals

The section does not explain:

- vehicle records, vehicle article lookup, vehicle attachments, or vehicle spare parts;
- administration of master-data entries or inventory-number schemes;
- the unpublished layout workspace or how to design layouts and layout units;
- deployment-level attachment limits, reverse proxy configuration, or environment variables;
- external catalog data as authoritative truth;
- cloud synchronization, public sharing, marketplace behavior, or later development features.

Existing allocation targets may be named where the accessories workflow exposes them. The guide
does not turn that reference into documentation of the unpublished layout workspace.

## Source and compatibility boundary

Behavior claims come from the stable `v0.1.17.6` tag, not from later `main` development. Primary
source owners include:

- `frontend/src/features/accessories/AccessoriesView.tsx`;
- `frontend/src/features/accessories/ArticleEditorDialog.tsx`;
- the `ArticleCoreTab`, `ArticleStockTab`, `ArticlePurchaseDocumentsTab`, `ArticleSubjectTab`, and
  `ArticleUsageHistoryTab` components;
- accessories overview, editor, search, automation, table, card, and responsive helpers;
- stock, purchase, document, reservation, installation, and asset panels;
- `frontend/src/shared/apiLayoutsAccessories.ts` and the stable API adapter;
- accessories translations in `frontend/src/shared/i18n/en.ts` and `de.ts`;
- `backend/internal/application/accessories*.go` and related inventory, purchase, document, and
  allocation services;
- `backend/internal/api/accessory*_handlers.go` and stable route access rules;
- stable domain types, migrations, OpenAPI operations, tests, and backup scope.

Later article-search refinements or accessories changes on `main` must not be described unless they
are already present in `v0.1.17.6`.

## Information architecture

Create a mirrored four-page section:

- overview and catalogue: `guide/accessories/index.md` and `de/guide/accessories/index.md`;
- records and technical data: `guide/accessories/article-records.md` and
  `de/guide/accessories/article-records.md`;
- stock, purchases, and documents: `guide/accessories/stock-purchases-documents.md` and
  `de/guide/accessories/stock-purchases-documents.md`;
- reservations, installations, and history: `guide/accessories/allocations-history.md` and
  `de/guide/accessories/allocations-history.md`.

Suggested titles:

- **Accessories overview** / **Zubehörübersicht**;
- **Article records and technical data** / **Artikelstammdaten und Fachangaben**;
- **Stock, purchases, and documents** / **Bestand, Käufe und Dokumente**;
- **Reservations, installations, and usage** / **Reservierungen, Einbauten und Verwendung**.

Every page uses the established user-guide frontmatter contract, names stable version
`v0.1.17.6`, uses matching English and German section order, and states the review date as
2026-08-16.

## Page 1: Accessories overview

The overview page introduces the workspace and explains:

- the operational meaning of its metrics and status filters;
- which fields the server search and browser filters inspect;
- how active, archived, low-stock, allocated, and other stable states combine;
- table, card, and compact mobile views;
- stable columns, sorting, and the table selection state, including that `v0.1.17.6` exposes no
  bulk command for the selected accessories;
- opening an article in read-only or edit mode;
- loading, empty, no-result, and stale/error states;
- role-specific read-only messages and the transition to each detailed page.

The page acts as the coverage destination and the hub for the three focused subpages. It does not
repeat their field-by-field or transaction details.

## Page 2: Article records and technical data

The records page documents:

- required and optional common article fields;
- article number, EAN, manufacturer status, URLs, alternative numbers, keywords, compatibility,
  notes, package quantity, stock unit, minimum stock, and inventory strategy;
- configured article types and subtypes, including historical inactive values;
- type-specific stable technical fields and custom subject fields;
- automatic scale and keyword suggestions and when manual edits stop automatic management;
- accessory barcode search and external article-data search as untrusted suggestions;
- which selected search values remain draft state until the article is saved;
- duplicate checking, candidate review, and deliberate duplicate confirmation;
- loss warnings when changing the article type would discard subtype or technical values;
- unsaved-change close confirmation and tab-level validation errors;
- archive, restore, and permanent Admin deletion, including references that can block deletion.

The field reference is grouped by user purpose rather than by component filename. Stable static
choices, configured master data, numeric rules, and normalization behavior are stated precisely.

## Page 3: Stock, purchases, and documents

The stock page starts with the critical inventory-strategy distinction:

| Strategy | Meaning |
| --- | --- |
| Quantity | Interchangeable units are booked as counts by storage location. |
| Individual | Physical items are represented as identifiable assets with their own lifecycle. |

The page then explains:

- initial stock behavior during article creation;
- stock totals, availability, reservations, installations, and minimum-stock warnings;
- positive and negative stock adjustments and their confirmations;
- transfers between active storage locations and movement history;
- individualization of quantity stock where the stable UI permits it;
- creation and editing of individual assets, including inventory identity and lifecycle limits;
- purchase fields, optional immediate stock booking, and purchase history;
- accessory images and other documents, categories, primary-image behavior, metadata, download,
  URL import, and deletion;
- public-URL, redirect, file-type, size, empty-file, and storage-safety rules;
- sequential or multi-step writes, refresh behavior, and recovery after partial failure.

Storage locations are explained only as resources used by stable accessories workflows. CRUD
operations with no published user interface are not presented as visible user actions.

## Page 4: Reservations, installations, and usage

The allocation page explains:

- target kinds exposed by stable `v0.1.17.6` and the boundary around unpublished layouts;
- quantity and individual-asset allocation rules;
- availability checks and valid storage-location or asset choices;
- Planner reservation rights and the difference from Editor installation rights;
- creating and cancelling an active reservation;
- installing with or without a reservation where permitted;
- how a reservation constrains target, location, asset, and quantity choices;
- recording installation condition and notes;
- changing condition, removing an installation, and booking the resulting disposition;
- active and removed installation states;
- allocation summary, reservation state, and usage-history chronology;
- confirmations, validation errors, stale resources, and recovery after a failed refresh.

The page makes clear when a stock count, asset lifecycle, reservation, installation, or history row
changes. It does not imply that a layout target makes the layout editor a published feature.

## Roles and permissions

The section uses one explicit role matrix:

| Capability | Viewer | Planner | Editor | Admin |
| --- | --- | --- | --- | --- |
| Read articles, stock, documents, and history | Yes | Yes | Yes | Yes |
| Create or cancel reservations | No | Yes | Yes | Yes |
| Create or edit articles and technical data | No | No | Yes | Yes |
| Manage stock, purchases, assets, and documents | No | No | Yes | Yes |
| Create, update, or remove installations | No | No | Yes | Yes |
| Archive or restore an article | No | No | Yes | Yes |
| Permanently delete an article | No | No | No | Yes |

The stable UI and server route checks are both audited. Where visible controls and server authority
differ, the guide treats the server rule as authoritative and names the stable UI limitation.
Messe does not receive the general accessories workflow through its role alone.

## Trust, persistence, and safety model

External barcode and article-search results are suggestions. Users must verify manufacturer,
article identity, URLs, technical data, and images before applying them.

The section distinguishes these states:

| State | Persistence |
| --- | --- |
| Search result | External suggestion, not stored |
| Edited article or technical field | Dialog draft until the article save succeeds |
| Stock, purchase, asset, document, reservation, or installation action | Separate immediate server write |
| Usage history | Stored consequence of completed lifecycle actions |

Closing a dirty dialog, changing article type, confirming a duplicate, adjusting stock, cancelling
a reservation, recording an installation, removing an installation, archiving, restoring, and
deleting are documented with their actual stable confirmation behavior. A later failed request
must never be described as rolling back an earlier successful request unless the stable service
uses one transaction that proves that behavior.

## Empty, partial, and error states

The pages cover at least:

- no articles, no search result, no filter match, and article-loading failure;
- missing or inactive master data;
- required common fields and invalid numeric or type-specific values;
- duplicate candidates and deliberate duplicate creation;
- unsaved changes and type-change data loss;
- stale article resources and retry behavior;
- incompatible inventory-strategy operations;
- insufficient quantity, unavailable assets, inactive storage locations, and invalid transfers;
- purchase booking failure and stock/history reconciliation;
- rejected document URLs or files and partial remote imports;
- invalid allocation targets, conflicting reservations, and unavailable reserved resources;
- failed condition change or removal and the need to reload before retrying;
- archive, restore, or deletion blocked by active or historical references;
- role-forbidden writes.

Stable German-only errors visible in the English locale are identified as localization limitations,
not translated into behavior that the UI does not provide.

## Storage and backup

Saved accessory products, technical attributes, quantity stock, movements, purchases, individual
assets, stored documents, reservations, installations, and usage history belong to local
RailKeeper application data. Their precise inclusion in backup and restore is verified against the
stable backup implementation before publication.

Transient search results and unsaved dialog drafts are not backup data. The guide recommends a
current validated backup before large inventory corrections, bulk lifecycle changes, or permanent
deletion. External URLs remain external dependencies even when their metadata is stored locally.

## Navigation and cross-links

Add an **Accessories** / **Zubehör** group to both user-guide sidebars after the vehicle group. The
group links to the overview and its three subpages.

Add the accessories overview to both User Guide landing pages. Add related-page links among the
four accessories pages and from relevant already published pages. Link only to existing published
owners, including:

- User Guide overview;
- Overview, metrics, and data quality;
- vehicle inventory and core records;
- article search, web documents, and spare parts where the boundary needs clarification.

Do not create active links to planned master-data, Settings, backup/restore, layout, or general
administration pages.

## Coverage contract

Change only the `accessories` topic from `planned` to `documented`. Preserve its established
coverage destination and ownership:

- frontend route `/accessories`;
- translations below `accessories`;
- API prefixes below `/api/v1/accessory`;
- `/api/v1/storage-locations`.

The primary English and German coverage paths remain the overview pair. The additional three page
pairs are children of that documented topic.

If the stable source inventory exposes a user-facing accessories API outside the existing prefix
mapping, update the owner map only when the mapping is required for accurate validation and remains
inside the `accessories` topic. Do not absorb unrelated layout or administration owners.

## Verification and review

Verification includes:

1. a stable-tag audit for overview behavior, fields, types, defaults, validation, search, roles,
   stock strategies, movements, purchases, assets, documents, allocations, history, deletion, and
   backup;
2. matching English and German frontmatter, page structure, tables, warnings, and cross-links;
3. an intentional missing-page validation failure after changing the coverage status;
4. route, navigation, unfinished-marker, internal-link, line-length, and whitespace scans;
5. `npm.cmd run check` from `docs`, including unit tests, coverage validation, and the VitePress
   production build;
6. independent read-only review of source fidelity, parity, role boundaries, external-data trust,
   immediate versus draft persistence, stock arithmetic, lifecycle effects, partial writes, and
   recovery guidance;
7. correction and focused re-review of every valid finding;
8. a pull request from the isolated branch only;
9. merge only when all required checks pass for the reviewed head;
10. post-merge verification of GitHub Pages and both language trees.

## Expected outcome

Users can locate an accessory article, understand its product and technical identity, choose the
right stock strategy, maintain purchases and documents, and allocate real stock without confusing
counts, individual assets, reservations, installations, or history. They know which actions are
draft-only, which write immediately, which role is required, and how to recover when an operation
or later refresh fails.
