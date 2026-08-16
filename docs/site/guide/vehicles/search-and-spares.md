---
title: Article search, web documents, and spare parts
description: Verify external article data, import web documents, and maintain vehicle spare parts.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Article search, web documents, and spare parts

RailKeeper can use external suggestions to complete a vehicle, import reference documents, and
maintain spare parts. Those results are starting points, not authoritative model data. Open the
source and verify the manufacturer and article identity before storing anything.

This workflow has four distinct states:

| State | Meaning |
| --- | --- |
| Search result | External suggestion, not stored |
| Applied field or selected image | Local vehicle-editor state, not yet stored |
| Imported document or spare-part action | Immediate server-side write |
| Saved vehicle | Core fields and pending images persisted through the vehicle-save workflow |

## Prerequisites, settings, and access rights

**Article data web search** is enabled by default unless it was disabled in Settings. Its enabled
flag and selected sources are preferences in the current browser. Another browser or device can
therefore use different settings.

The default sources are manufacturer sites, Modellbahn-Fokus catalogs, dealer sites, and general
web search. Modellbau Wiki is an optional fifth source. Normal article and document searches
require:

- manufacturer;
- gauge;
- article number or designation.

An EAN-only barcode search is the exception and can run without those identity values.

Admin and Editor can use the editable vehicle UI, import files, and change spare parts. Viewer and
Planner can inspect stored vehicles, attachments, and spare parts. The article-search API permits
Viewer-level read access, but the read-only vehicle UI still disables search and apply controls.
Image, attachment, and spare-part writes require Editor-level server access. Messe does not receive
the general vehicle workflow.

## Search for article data

Open a vehicle editor and use **Search article data** in **Model**. The action remains disabled
until manufacturer, gauge, and either article number or designation are present. RailKeeper sends
those values together with other non-empty searchable vehicle fields and the sources selected in
this browser.

The server trims the input and can enrich a known manufacturer with aliases and preferred domains.
It builds several targeted queries, applies a ten-second service timeout, scores and sorts the
results, removes duplicate URLs, and returns at most ten results. Piko and Roco searches can also
include direct manufacturer-specific spare-part results.

The result dialog exposes the evidence behind the suggestions:

- final query, active source classes, preferred domains, and planned search queries;
- result title, source, score, snippet, and number of fields;
- whether the detail page loaded, failed, or was skipped;
- **Open source** for independent checking;
- found images and their previews;
- current and found values, their status, and conflicts.

The stable result groups contain these fields:

| Group | Visible fields |
| --- | --- |
| Model | Designation, article no., manufacturer, gauge, EAN, railway, epoch, series, vehicle no., subtype, category |
| Mass / construction | Length, weight, color, lettering, load, interior, axles, axle count, traction tires |
| Technology | Adapter, pickups, digital/decoder data, sound, lights, and technical descriptions |
| More data | Description, additional information, production period, list price, and source URL |

RailKeeper removes obviously empty, advertising, cookie, manual-navigation, and implausible field
values before display. Filtering and a high score improve review efficiency, but neither proves
that a remaining value belongs to the physical model.

## Search by barcode or camera

Use **Search barcode** in **Model** to open an EAN field initialized from the current vehicle form.
Type or paste the value, use a keyboard scanner, or start the camera.

Submitting a non-empty code copies it into the local EAN field, leaves the article number
unchanged, closes the barcode dialog, and runs an EAN-only search. The EAN is not stored until the
vehicle itself is saved.

Camera scanning requires:

- a secure browser context such as HTTPS or localhost;
- browser camera support and permission;
- a readable code with at least eight digits after non-digits are removed.

RailKeeper requests the environment-facing camera when available. Detection fills the field but
does not submit the search automatically. If permission, browser support, or detection fails, enter
the EAN manually or use a keyboard scanner.

## Review and apply one result

Selections belong to one result at a time. A found field starts selected only when the current
vehicle field is empty. Equal values and conflicts are not selected automatically. Images are
never preselected. An image whose preview fails disappears and is removed from the selection.

Open the source and compare the article identity before selecting anything. Check conflicts and
plausibility across all field groups. **Apply selected fields** then applies only the checked valid
fields from this result. Boolean text is converted into RailKeeper's stable boolean
representation. Selected images become pending remote images. If no image is already pending, the
first selected image becomes the main-image candidate.

Applying closes the result dialog but does not write to the server. Inspect the complete vehicle
form and pending image list, then use **Create** or **Save changes**. Closing the vehicle editor or
reloading before that save loses the applied fields, EAN, and pending images.

Vehicle save writes the core vehicle first, then downloads selected remote images sequentially. If
a later image fails, the vehicle and earlier images remain stored, and the normal final refresh is
not performed. Reload the vehicle, compare its stored images, and retry only the missing ones. See
[Vehicle images and attachments](/guide/vehicles/media) for ordering, metadata, previews,
maintenance links, and deletion after import.

## Find and import web documents

Open the saved vehicle's **Uploads** tab and use **Found documents**. The search uses the vehicle's
stored identity and the normal manufacturer, gauge, and article-number-or-designation requirement.
Results are deduplicated by URL or title and show title, kind, and source.

RailKeeper recognizes a previously imported document when its URL occurs in an attachment
description. Such a row cannot be selected again and offers local download and open actions.
Import one remaining result, or select all results that are not already stored.

An import downloads the external content and stores it as a real local vehicle attachment. Its
description records the source and source URL. RailKeeper infers one of these categories:

- `Ersatzteilliste` for spare-parts signals;
- `Anleitung` for manual or instruction signals;
- `Dokumentation` otherwise.

The server accepts only public HTTP(S) URLs, rechecks redirect destinations, and blocks private or
internal targets. It also enforces the configured attachment size and type rules, rejects empty
files, and applies a ten-second remote request timeout.

Importing several selected documents sends requests sequentially. A later failure leaves earlier
documents stored and prevents the normal final refresh and selection reset. Reload the vehicle,
compare the attachment list, and retry only missing URLs.

## Extract spare parts from an attachment

Every stored attachment offers **Extract spare parts**. The action first saves that selected
attachment's current description, category, and maintenance link. It then analyzes only this
attachment, switches to **Spare parts**, and reloads the vehicle.

Extraction requires a saved vehicle article number. The server reads at most 12 MiB from the
attachment and returns at most 80 unique suggestions. Likely spare-parts PDFs, manuals, service
sheets, HTML, and supported text content can be analyzed. PDF text is extracted directly first.
Scanned PDFs can use optional OCR supplied by the operator. Without working OCR, a scan can
legitimately produce no suggestions.

The stable action has no confirmation or row preview. It cleans descriptions, drops empty and
document-title-like rows, removes duplicates already stored on the vehicle, and creates every
remaining candidate sequentially. A source URL pointing back to RailKeeper's own attachment
download is not stored as an external spare-part link.

Attachment metadata and earlier spare parts remain stored if a later create request fails. Reload
the vehicle and compare its attachment and spare-part data before retrying.

## Maintain spare parts manually

The **Spare parts** tab becomes writable after the vehicle exists. The editor contains article
number, description, price as free text, and external link. At least article number, description,
or link is required. Price alone is invalid.

New and updated values are trimmed. The initial table order is article number ascending. Select a
column heading to toggle ascending or descending order for article number, description, price, or
link.

Creating a part with an existing identity updates the matching entry instead of adding another
row. Identity uses normalized article number first, otherwise normalized URL, otherwise
description. Article-number normalization ignores an `ET` prefix, punctuation, spaces, and case.
The create merge preserves existing non-empty values and fills only fields that are still empty.
Editing a selected row writes exactly its four submitted fields.

Save, apply, and delete actions persist immediately and reload the complete vehicle after success.
Delete has no additional confirmation. Because a full reload replaces unrelated unsaved
vehicle-editor state, save or intentionally discard other edits first.

## Look up one stored spare part

A stored part with an article number offers **Search this spare part**. The lookup uses the vehicle
manufacturer, the spare-part article number, manufacturer and catalog sources, and a Piko or Roco
lookup mode when applicable. The stable UI displays at most five candidates, preferring results
with price, then availability, then link.

Each result can show price, availability, source, and an external link. **Apply price and link**
updates the stored part immediately. Article number and description remain unchanged.
Availability is displayed but not stored. If the selected candidate has no price, the stable
controller can use another priced candidate's price and URL from the same result list.

Opening the tab also checks linked spare parts once for the current vehicle state. Piko and Roco
use one manufacturer-overview request. Other manufacturers check only the first four externally
linked parts that also have article numbers. The indicator classifies recognizable text as
available, limited, unavailable, or unknown. It remains a live suggestion, not local inventory or
a purchase promise.

## Import a Piko or Roco manufacturer overview

**Find available spare parts** is enabled only when all of these conditions are met:

- the vehicle is saved;
- its manufacturer name contains Piko or Roco;
- the vehicle article number is not empty;
- the user has Admin or Editor access to the editable UI.

The button can still be selected while article search is disabled. In that case RailKeeper stops
before the external request, displays the Settings message, and stores nothing.

RailKeeper searches the manufacturer overview and builds a conservative import plan. Matching
prefers normalized spare-part article number, then description, then URL. Repeated suggestions are
consolidated. Existing matches are updated first, and missing parts are created second.

Existing article number and description are retained. Existing price and external URL also win;
only missing price or URL is filled. The updates and creates run sequentially without a transaction
or confirmation preview. A later failure leaves earlier rows stored and prevents the normal final
refresh. Reload and compare before retrying. **No missing spare parts found** can mean that every
manufacturer suggestion already matches a stored row, not that the manufacturer has no parts.

## Protect data during multi-step actions

The timing of persistence and refresh differs by action:

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
Reload, compare stored data, and retry only missing work after any partial failure.

Search responses, result selections, and unsaved form changes are transient browser state and do
not belong to a backup. Once saved, article-derived vehicle fields, imported images, attachment
metadata and blobs, and vehicle spare parts belong to local RailKeeper application data and are
included in the application backup scope.

Before a large document batch, extraction run, manufacturer import, or cleanup, create an
application backup and validate it. External source URLs and spare-part links can still disappear
outside RailKeeper and are not made durable by a local backup.

## Resolve empty, partial, and error states

| Situation | What to do |
| --- | --- |
| Article search is disabled | Enable it in Settings for this browser, or continue without external suggestions. |
| Normal search is disabled | Enter manufacturer, gauge, and article number or designation. |
| Camera does not start | Use HTTPS or localhost, check permission and support, or enter the code manually. |
| No article result appears | Refine the identity and selected sources without weakening source verification. |
| Detail page failed or was skipped | Open the source and use only fields you can verify. |
| A found image disappears | Its preview failed. Select a different source image. |
| Applied values vanish | They were local draft state. Apply them again and save the vehicle. |
| Remote-image save partly fails | Reload and compare stored images before retrying only missing ones. |
| No web documents appear | Check saved vehicle identity and sources. Not every source exposes documents. |
| Document import fails | Check public URL, redirects, type, size, and remote availability. |
| Document batch partly fails | Earlier documents remain stored. Reload and retry only missing URLs. |
| Extraction finds nothing | Check vehicle article number, content, text extraction, and optional OCR. |
| Extraction partly fails | Metadata and earlier parts can already be stored. Reload and compare. |
| Spare-part form is rejected | Enter article number, description, or link. Price alone is insufficient. |
| Single lookup is disabled | Store a spare-part article number and enable article search. |
| Availability is unknown | Treat it as unverified external status and open the source. |
| Manufacturer overview is disabled | It requires Piko or Roco and the saved vehicle article number. |
| Manufacturer import partly fails | Reload, compare, and rerun only after understanding stored results. |
| A spare part was deleted without a prompt | Deletion is immediate. Recovery requires a suitable backup. |

Some stable backend and frontend branches still show German errors in English mode. Treat the
message as a stable UI limitation and use the persistence rules above to determine what may already
have been stored.

## Related pages

- [User Guide overview](/guide/)
- [Vehicle inventory and basic data](/guide/vehicles/)
- [Vehicle images and attachments](/guide/vehicles/media)
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)
- [Decoder, functions, and CV data](/guide/vehicles/decoder-cv)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.17.6** and was last reviewed on 2026-08-16.
