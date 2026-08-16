# RailKeeper Vehicle Search and Spare Parts Guide Design

Date: 2026-08-16  
Status: Approved in conversation  
Documented release: `v0.1.17.6`

## Objective

Publish a complete English and German user-guide chapter for the stable vehicle article-search,
barcode, web-document, and spare-parts workflows. The chapter follows one operational path:

1. prepare reliable search criteria;
2. search and inspect external suggestions;
3. apply selected fields and images to the vehicle draft;
4. save remote content locally;
5. find and import real web documents;
6. extract, maintain, and look up spare parts;
7. recover safely from partial writes.

The pages explain every visible stable action in this scope and clearly distinguish search results,
unsaved local changes, and data already stored in RailKeeper.

## Scope

The chapter owns these stable user workflows:

- article-data web search from a vehicle editor;
- barcode or EAN entry, keyboard scanners, and camera scanning;
- search sources, search trace, result score, detail-page state, conflicts, and source inspection;
- selective transfer of article fields and remote images into a vehicle draft;
- saving selected remote images with the vehicle;
- discovery and import of web documents as local vehicle attachments;
- spare-part extraction from stored attachments;
- manual spare-part creation, update, sorting, linking, and deletion;
- single-part price, link, and availability lookup;
- Piko and Roco manufacturer-overview import;
- roles, persistence, refreshes, partial failures, storage, backup, and recovery.

## Non-goals

The chapter does not explain:

- general accessory inventory or accessory article search;
- master-data editing, manufacturer-domain maintenance, or inventory-number schemes;
- every option on the general Settings page;
- deployment-level OCR installation or environment-variable reference;
- external catalog correctness as authoritative data;
- public sharing, cloud synchronization, or marketplace behavior;
- unpublished or later-version search features.

Those boundaries may be named in plain text or linked only when their owning page is already
published.

## Source and compatibility boundary

Behavior claims come from the stable `v0.1.17.6` tag, not from later `main` development. Primary
source owners include:

- `frontend/src/features/vehicles/useArticleSearchController.ts`;
- `frontend/src/shared/articleSearch/ArticleSearchDialog.tsx`;
- `frontend/src/shared/articleSearch/BarcodeSearchDialog.tsx`;
- `frontend/src/features/vehicles/useVehicleDocumentsController.ts`;
- `frontend/src/features/vehicles/useVehicleSparePartsController.ts`;
- `frontend/src/features/vehicles/VehicleSparePartsTab.tsx`;
- `frontend/src/features/vehicles/vehicleSparePartSearch.ts`;
- `frontend/src/features/vehicles/vehicleSpareParts.ts`;
- `backend/internal/application/article_search*.go`;
- `backend/internal/application/vehicle_spare_parts_service.go`;
- `backend/internal/api/vehicle_operation_handlers.go`;
- remote vehicle image and attachment handlers;
- stable routes, OpenAPI operations, translations, migrations, and backup scope.

## Pages and metadata

Create this language pair:

- `docs/site/guide/vehicles/search-and-spares.md`;
- `docs/site/de/guide/vehicles/search-and-spares.md`.

Both pages use the established user-guide frontmatter contract, name stable version `v0.1.17.6`,
use matching section order, and state the review date as 2026-08-16.

Suggested titles:

- **Article search, web documents, and spare parts**;
- **Artikelsuche, Web-Dokumente und Ersatzteile**.

## Operational model

External pages and extracted values are suggestions. A source URL, high score, recognized
manufacturer domain, price, or availability label does not make the data authoritative. Users must
open the source, check that manufacturer and article identity match, compare conflicts, and select
only values that belong to the physical model.

The chapter separates four states:

| State | Meaning |
| --- | --- |
| Search result | External suggestion, not stored |
| Applied field or selected image | Local vehicle-editor state, not yet stored |
| Imported document or spare-part action | Immediate server-side write |
| Saved vehicle | Core fields and pending images persisted through the vehicle-save workflow |

## Prerequisites, settings, and roles

Article search is enabled by default unless **Article data web search** is disabled in Settings.
The enabled flag and selected sources are browser-local preferences. Default sources are
manufacturer sites, Modellbahn-Fokus catalogs, dealer sites, and general web search. Modellbau Wiki
is an optional fifth source. Older stored defaults are migrated to the current defaults by the
stable frontend.

Normal vehicle search requires:

- manufacturer;
- gauge;
- article number or designation.

An EAN-only barcode search is the exception and can run without those four identity values.
Document search operates on a saved vehicle and uses the same normal identity requirement.

Admin and Editor can use the editable vehicle UI and perform imports or spare-part writes. Viewer
and Planner can inspect stored vehicle, attachment, and spare-part data. The article-search API has
Viewer-level read access, but the stable read-only vehicle UI disables search and application
controls. Image, attachment, and spare-part writes require Editor-level server access. Messe does
not receive the general vehicle workflow.

## Normal article search

The **Model** section contains **Search barcode** and **Search article data**. The normal button is
disabled until its required identity values exist. A search sends manufacturer, article number,
designation, gauge, configured sources, and other non-empty searchable vehicle fields.

The server trims the input, enriches a known manufacturer with configured aliases and preferred
domains, constructs a search plan, and uses a ten-second service timeout. Results are scored,
sorted, deduplicated by URL, and limited to ten. Piko and Roco may also provide direct
manufacturer-specific spare-part results.

The result dialog shows:

- final query;
- active source classes;
- preferred manufacturer domains;
- individual planned search queries;
- result title, source, field count, score, and snippet;
- whether the detail page was loaded, failed, or skipped;
- an external **Open source** action;
- found images;
- conflicts with current fields;
- current and found values with status.

The stable field groups are:

| Group | Visible fields |
| --- | --- |
| Model | Designation, article no., manufacturer, gauge, EAN, railway, epoch, series, vehicle no., subtype, category |
| Mass / construction | Length, weight, color, lettering, load, interior, axles, axle count, traction tires |
| Technology | Adapter, pickups, digital/decoder data, sound, lights, and technical descriptions |
| More data | Description, additional information, production period, list price, and source URL |

RailKeeper removes obviously unusable empty, advertising, cookie, manual-navigation, and implausible
field values before display. This filtering is a quality aid, not proof that the remaining values
are correct.

## Barcode and camera search

**Search barcode** opens an EAN field initialized from the current vehicle form. Users can type or
paste a value, use a keyboard scanner, or start the camera.

Submitting a non-empty value puts it into the local EAN field, leaves the article number unchanged,
closes the barcode dialog, and runs an EAN-only search. The value is not stored until the vehicle is
saved.

Camera behavior:

- it requires a secure browser context such as HTTPS or localhost;
- it requires browser camera support and permission;
- it requests the environment-facing camera when available;
- non-digits are removed from a detected code;
- a detected code must contain at least eight digits;
- detection fills the field but does not submit the search automatically;
- keyboard scanners and manual input remain available if the camera fails.

## Review and apply one result

Each result has independent field and image selections. Found fields are initially selected only
when the corresponding current field is empty. Equal and conflicting existing values are not
selected automatically. Images are never preselected. Failed image previews disappear and are
removed from selection.

**Apply selected fields** applies only the checked, valid fields from that one result. Boolean text
is converted to the stable boolean representation. Selected images become pending remote images in
the vehicle editor. If no image is currently pending, the first selected image becomes the main
image candidate.

Applying closes the result dialog but performs no server write. Users must inspect the complete
vehicle form and pending image list, then use the vehicle's **Create** or **Save changes** action.
Closing the vehicle editor or reloading before that save loses the applied draft changes.

During vehicle save, RailKeeper writes the core vehicle first and downloads selected remote images
sequentially afterward. A later image failure leaves the vehicle and earlier images stored. It does
not roll them back or run the normal final refresh. Users must reload, compare, and retry only
missing images. The media chapter owns image ordering, metadata, maintenance links, previews, and
deletion after import.

## Find and import web documents

On the saved vehicle's **Uploads** tab, **Found documents** uses the current stable article search
and the saved vehicle identity. The result list is deduplicated by URL or title and shows title,
kind, and source.

RailKeeper recognizes an already imported document when its URL occurs in an existing attachment
description. Such a row cannot be selected again and instead offers local download and open
actions. Users can import one remaining row or select all not-yet-imported rows.

Import downloads the external content and stores a real local vehicle attachment. RailKeeper adds
the source and source URL to its description and infers one of these categories:

- `Ersatzteilliste` for spare-parts signals;
- `Anleitung` for manual or instruction signals;
- `Dokumentation` otherwise.

The server accepts only public HTTP(S) locations, rechecks redirects, blocks private or internal
destinations, enforces its configured attachment size and type rules, rejects empty files, and uses
a ten-second remote request timeout. The media chapter owns the full supported-format and normal
attachment-limit reference.

Multi-document import sends requests sequentially. A later failure leaves earlier documents stored
and prevents the normal success refresh and selection reset. Reload, compare the attachment list,
and retry only missing URLs.

## Extract spare parts from an attachment

Every stored attachment has **Extract spare parts**. The action first saves that attachment's
currently edited description, category, and maintenance link. It then asks the server to analyze
the selected attachment, switches to **Spare parts**, and reloads the vehicle.

Extraction requires a saved vehicle article number. The server reads at most 12 MiB from the chosen
attachment and returns at most 80 unique suggestions. Likely spare-parts PDFs, manuals, service
sheets, HTML, and supported text content can be analyzed. PDF text is extracted directly first.
Scanned PDFs may use optional operator-provided OCR; without working OCR they can legitimately
produce no suggestions.

The stable extraction action has no confirmation or row preview. It cleans descriptions, discards
empty or document-title-like rows, removes duplicates already stored on the vehicle, and creates
every remaining candidate sequentially. Attachment metadata and earlier spare parts remain stored
if a later request fails. A source URL that points back to RailKeeper's own attachment download is
not stored as the spare part's external link.

## Maintain spare parts manually

The **Spare parts** tab becomes writable only after the vehicle exists. Its editor contains:

- article number;
- description;
- price as free text;
- external link.

At least article number, description, or link is required; price alone is invalid. Add and update
trim surrounding whitespace. The initial table order is article number ascending. Users can toggle
ascending or descending sorting for article number, description, price, and link.

Creating a part with an existing identity updates the matching entry rather than adding another
row. Identity uses normalized article number first, otherwise normalized URL, otherwise
description. Article-number normalization ignores an `ET` prefix, punctuation, spacing, and case.
The create merge preserves existing non-empty fields and fills only missing values from the new
input. Editing a specific row writes exactly its four submitted fields.

Save, apply, and delete actions persist immediately and reload the complete vehicle on success.
Delete has no additional confirmation. A reload replaces unrelated unsaved vehicle-editor state,
so users must save or intentionally discard other changes first.

## Look up one stored spare part

A stored part with an article number can run **Search this spare part**. The lookup uses the vehicle
manufacturer, the spare-part number, manufacturer and catalog sources, and a Piko or Roco lookup
mode when applicable. The stable UI displays at most five candidates, preferring results with
price, then availability, then link.

Each result can show price, availability, source, and an external link. **Apply price and link**
updates the stored part immediately. Article number and description remain unchanged.
Availability is displayed but not stored. If a chosen candidate has no price, the stable controller
may use another priced candidate's price and URL from the same result list.

Opening the tab also checks linked spare parts once for the current vehicle state. Piko and Roco
use one manufacturer-overview request. Other manufacturers check only the first four externally
linked parts that also have article numbers. The indicator classifies recognizable availability
text as available, limited, unavailable, or unknown. It remains a live suggestion, not inventory
held by RailKeeper.

## Import a Piko or Roco manufacturer overview

**Find available spare parts** is enabled only for:

- a saved vehicle;
- a manufacturer name containing Piko or Roco;
- a non-empty vehicle article number;
- enabled article search;
- Admin or Editor UI access.

RailKeeper searches the manufacturer overview, builds a conservative import plan, updates existing
matches first, and creates missing parts second. Matching prefers normalized spare-part article
number, then description, then URL. Existing article number and description are retained. Existing
price and external URL are also retained; only missing price or URL is filled. Repeated suggestions
are consolidated.

Updates and creates run sequentially without one transaction or confirmation preview. A later
failure leaves earlier rows stored and prevents the normal final refresh. Reload and compare before
retrying. **No missing spare parts found** can mean that all manufacturer entries already match,
not that the manufacturer has no spare parts.

## Persistence and refresh matrix

| Action | Persists data | Full vehicle refresh |
| --- | --- | --- |
| Run article or barcode search | No | No |
| Apply result fields or images | No, local editor only | No |
| Save vehicle after applying a result | Core vehicle, then remote images sequentially | Only after full success |
| Search found documents | No | No |
| Import one web document | Immediately | After success |
| Import selected web documents | Sequentially | Only after full success |
| Extract spare parts | Attachment metadata, then parts sequentially | Only after full success |
| Add, update, apply lookup, or delete one spare part | Immediately | After success |
| Check price or availability | No | No |
| Import Piko/Roco overview | Updates, then creates sequentially | Only after full success |

No later failed request rolls back an earlier successful request in the same sequential workflow.

## Storage and backup

Search responses, result selections, and unsaved vehicle-form changes are transient browser state
and are not part of a backup. Once saved, article-derived vehicle fields, imported images,
attachments and their blobs, and vehicle spare parts belong to local RailKeeper application data
and are included in the application backup scope.

Before a large manufacturer import, document batch, extraction run, or cleanup, users should create
and validate a current application backup. Search URLs and spare-part links still point to external
systems and do not guarantee future external availability.

## Empty, partial, and error states

Both pages cover at least:

| Situation | Documented response |
| --- | --- |
| Article search is disabled | Enable it in Settings for this browser or continue without external suggestions. |
| Normal search button is disabled | Enter manufacturer, gauge, and article number or designation. |
| Camera does not start | Use HTTPS/localhost, check permission and support, or scan manually. |
| No article result | Refine identity values and sources; do not weaken source verification. |
| Detail page failed or was skipped | Open the source and trust only fields that can be verified. |
| A found image disappears | Its preview failed; choose another source image. |
| Applied values vanish | They were local draft state; apply again and save the vehicle. |
| Remote image save partly fails | Reload and compare stored images before retrying. |
| No web documents are found | Check vehicle identity and sources; not every source exposes documents. |
| Document import fails | Check public URL, redirect, type, size, and remote availability. |
| Batch document import partly fails | Earlier documents remain stored; reload and retry missing URLs only. |
| Extraction finds nothing | Check vehicle article number, attachment type/content, text extraction, and optional OCR. |
| Extraction partly fails | Attachment metadata and earlier parts may already be stored; reload and compare. |
| Spare-part form is rejected | Enter article number, description, or link; price alone is insufficient. |
| Single lookup is disabled | Store a spare-part article number and enable article search. |
| Availability is unknown | Treat it as unverified external status and open the source. |
| Manufacturer overview is disabled | It requires Piko or Roco plus the vehicle article number. |
| Manufacturer import partly fails | Reload, compare, and rerun only after understanding stored results. |
| Spare part was deleted without a prompt | Deletion is immediate; recovery requires a suitable backup. |

The stable backend and some frontend branches still emit German errors in English mode. The
English page identifies this as a stable UI limitation.

## Navigation and cross-links

Add exact sidebar entries after the decoder/CV chapter:

```ts
{ text: "Article search, web documents, and spare parts", link: "/guide/vehicles/search-and-spares" }
```

```ts
{ text: "Artikelsuche, Web-Dokumente und Ersatzteile", link: "/de/guide/vehicles/search-and-spares" }
```

Add concise entries to both User Guide landing pages. Add the new chapter to the related-page lists
of the core vehicle, media, maintenance, and decoder/CV chapters. Replace the media chapter's
specialist-boundary wording with a direct transition where appropriate.

The new chapter links only to published owners:

- User Guide overview;
- vehicle inventory and core records;
- vehicle images and attachments;
- vehicle maintenance and condition;
- decoder, functions, and CV data.

Do not create active links to planned settings, master-data, accessory, or administration pages.

## Coverage contract

Change only `vehicle-search-spares` from `planned` to `documented`. Preserve its established page
paths and ownership:

- translations below `settings.articleSearch`;
- translations below `vehicles.articleSearch`, `vehicles.barcode`, and `vehicles.search`;
- translations below `vehicles.spareParts`;
- translations below `vehicles.uploads.extractSpareParts`, `vehicles.uploads.webDocument`, and
  `vehicles.uploads.webDocuments`;
- API prefixes for vehicle image and attachment URL import;
- `/api/v1/vehicles/{id}/spare-parts`;
- `/api/v1/article-search`.

The implementation first proves the missing-page gate after changing coverage status, then adds the
paired pages.

## Verification and review

Verification includes:

1. stable-tag source checks for criteria, sources, roles, fields, defaults, limits, result state,
   local application, persistence, remote imports, extraction, deduplication, lookup, refresh, and
   backup;
2. matching English and German frontmatter and section order;
3. route, navigation, cross-link, unfinished-marker, line-length, and whitespace scans;
4. an intentional missing-page validation failure after the coverage change;
5. `npm.cmd run check` from `docs`, including 19 unit tests, coverage validation, and VitePress
   production build;
6. independent read-only review of source fidelity, parity, data trust, role boundaries, transient
   versus stored state, remote import safety, partial writes, and recovery;
7. correction and focused re-review of every valid finding;
8. pull request from the isolated branch only;
9. merge only when CI, Documentation, Trivy, and CodeQL pass for the reviewed head.

## Expected outcome

Users can move from a reliable vehicle identity to inspected external suggestions, deliberately
apply and save fields and images, store real manufacturer documents, and maintain spare parts
without confusing previews with persisted data. They understand which actions are immediate,
which workflows are sequential and non-atomic, how to verify external sources, and how to recover
after partial failure.
