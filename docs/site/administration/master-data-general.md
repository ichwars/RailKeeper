---
title: General master data
description: Maintain vehicle classifications, manufacturers, CV8 identities, and function symbols.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# General master data

Open **Settings > Data > General** to maintain the shared selections used by vehicles, accessories,
decoder data, and exhibitions. The following data types are available.

## Data-type reference

| Data type | Used for | Type-specific fields |
| --- | --- | --- |
| Manufacturers | Vehicle and article manufacturer selection, search-source matching | Name, nominal scales, website, search domains, aliases, source URL |
| Vehicle categories | Broad vehicle classification and inventory-number selection | Name, source URL |
| Vehicle types | More specific vehicle type or `Gattung` | Name, source URL |
| Epochs | Vehicle and exhibition era selection | Name, source URL |
| Gauges | Vehicle and article gauge selection | Name, source URL |
| Railway companies | Vehicle and exhibition operator selection | Name, source URL |
| CV8 manufacturers | Decoder manufacturer identity from CV8 | Name, decimal, binary, hexadecimal, country |
| Function symbols | Images and descriptions for vehicle functions | Name, description, image |

Every entry also has an active state, sort order, origin, immutable key, and timestamps in storage.
The form exposes the fields relevant to the selected type. A new ordinary key is derived from the
label. Editing a label does not change that key.

## Search, sort, and inspect entries

Select a type, then use the local search box. It searches the displayed name, key, source values,
and the type-specific manufacturer, CV8, or symbol metadata. It does not search vehicle or article
records.

The table initially sorts by name. Select a sortable column to switch the sort key and select it
again to reverse the order. It always shows active and inactive entries, their **Bundled** or
**Custom** origin, and only the lifecycle actions allowed for that entry.

Use refresh after another administrator changes the same type. Loading a different type resets
the current edit, search, and sort selection.

## Deactivate multiple entries

Selection checkboxes appear only for active entries that may be deactivated. Select individual
rows or use the table-header checkbox to select all eligible entries currently visible through the
search. Existing selections remain selected when the search term changes, so the action bar always
shows the total selection within the current data type. Changing the data type clears the selection.

**Deactivate selected** opens a confirmation with the data type and count. Only confirming sends
the complete set to the server. RailKeeper processes it atomically: if an entry is missing or the
operation fails, none of the selected entries changes and the selection remains available for
correction. A successful operation clears the selection. Bulk reactivation is not available;
reactivate inactive entries individually.

## Maintain manufacturers

The manufacturer entry controls both selection labels and article-search hints:

- **Nominal scales** accepts the scales associated with the manufacturer.
- **Website** is the canonical manufacturer site.
- **Search domains** narrows preferred article-search sources. If the list is empty and the website
  is valid, the form derives its domain automatically.
- **Aliases** provide alternative manufacturer spellings for matching.
- **Source URL** records the provenance of the master-data entry itself.

Separate multiple scales, domains, or aliases as indicated by the field. Keep only manufacturer or
catalogue domains in the preferred-domain list. Dealer, marketplace, redirect, and SEO domains
reduce search quality. A warning in the table marks manufacturer entries that still need a proper
website review.

Changing a manufacturer label does not rewrite the free-text manufacturer value already stored on
vehicles, accessory articles, or exhibition entries. Review those records when correcting a label
that is already in use.

## Maintain CV8 manufacturers

CV8 identifies the decoder manufacturer through a decimal value from 0 through 255. Entering the
decimal value derives the binary and hexadecimal representations; both remain visible and are
normalized before saving. The optional country value is uppercased and limited to eight
characters.

New entries receive an immutable key in the form `cv8-NNN`, for example `cv8-151`. The displayed
name omits a redundant leading decimal number. RailKeeper links the official NMRA manufacturer-ID
appendix above the list. Use that source before adding or correcting an identity.

## Maintain function symbols

RailKeeper supplies 94 independently authored **Werkstatt-Linie** symbols. Eight are general
fallbacks and 86 retain numeric ECoS function-description codes solely as compatibility metadata.
The graphics themselves are part of RailKeeper and are not copied from the command station.

A symbol consists of a name, optional description, and optional image. Administrators can still
create their own symbol keys and upload SVG, PNG, JPEG, or WebP images up to 1 MiB. The image is
stored inside the entry's metadata, so it is included in master-data export and application backup.

Removing the preview from the form removes the image only when the entry is saved. Existing vehicle
functions keep the symbol key. Deactivating a symbol prevents new normal selections while existing
references remain readable. The exhibition workspace can retrieve active symbols without gaining
general Settings access.

## Create, edit, and retire an entry

1. Select the required data type and search for an existing equivalent.
2. Enter the name and the relevant source or type-specific values.
3. Select **Add**. RailKeeper creates a local **Custom** entry.
4. To correct an entry, select its edit action, change the form, and save.
5. To retire it, use **Deactivate** and confirm. Existing stored uses remain unchanged.
6. Reactivate the entry when it should return to new selections.
7. Use permanent deletion only when RailKeeper offers it for an unused custom entry.

Bundled entries can be edited and deactivated but not deleted. A referenced custom entry also
cannot be deleted. If a write is rejected for Viewer users, sign in with Admin or Editor
authority instead of retrying the visible control.

## Related pages

- [Master-data administration](./master-data)
- [Article master data and storage locations](./master-data-articles)
- [Inventory-number schemes](./master-data-inventory-numbers)
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)


## Decoder library

Navigate **manufacturer → decoder family → model**. The researched directory covers ESU,
ZIMO, Doehler & Haass, Lenz, Kühn, Uhlenbrock, Roco and PIKO, with 40 families, 599 catalogue entries
and 747 documented order-number/interface mappings across the model records. This is a selection of current and
earlier decoders, not a complete manufacturer archive. Each manufacturer lists its coverage
and remaining gaps.

Catalogue entries include hardware types, interface versions and separately ordered sound decoders
with vehicle-specific projects. Hardware generations may share order numbers, which do not identify
the generation alone. Unverified connectors or order numbers remain unrecorded.

| Manufacturer | Entries | Documented order mappings | Remaining scope gaps |
| --- | ---: | ---: | --- |
| ESU | 58 | 132 | V1.0, LokSound classic/V2.x, Select, L V4.0/OEM and individual V2.0 connections |
| ZIMO | 177 | 178 | Early MX620/MX631/MX640–647, MX69/MX690 and further OEM versions |
| Doehler & Haass | 41 | 101 | Further OEM/archive versions; SelecTRIX-only and SUSI modules outside DCC interpretation |
| Lenz | 49 | 54 | Grouped LE/A/E table rows, further revisions and vehicle-specific boards |
| Kühn | 7 | 20 | F060 order number and unambiguous N020/T121/T140 documents |
| Uhlenbrock | 52 | 60 | Contradictory 73800 file, further early IntelliSound/OEM versions |
| Roco | 30 | 30 | Early 107xx, factory and spare-part decoders; Motorola-only 10738 outside DCC |
| PIKO | 185 | 172 | Individual interfaces for 13 G sound articles, earlier sound/OEM/spare-part versions |

Manufacturer references occasionally conflict. ESU 54617/64617 use the newer connection table in
the LokPilot V4.0 manual (11th edition, p. 7); the earlier catalogue label does not become another
connector variant. Kühn F060 is documented as a wired function decoder with an unknown order number.
For DH21A and earlier Lenz decoders, an order number alone does not reliably identify hardware generation.
Sources are linked from each model and order mapping. A large result list does not mean every
generation has been catalogued.

Search model names, product numbers, interfaces, CV8 manufacturer identifiers or CV numbers. Filter by decoder type
and review scope. DCC foundations have their own navigation entry.

Each model provides **Overview, CV definitions, Identification and Sources**. “CV rules reviewed”
covers LokPilot 5 micro DCC, DH18A, DH05C, DH10C, DH16A, DH12A, DH14B, DH21A,
DH21B, DH22A, DH22B, DH24A, DHSP10A, MN180N18, STANDARD+ V2, N45, IntelliDrive 76420,
Roco 10882 (2012 manual) and
PIKO 56500 PluX22. The tables contain 2,091 records including the DCC foundation: 2,088 functional
definitions and three ESU reservation records. Only the
listed registers and memory ranges are reviewed, not complete semantics for every byte or firmware revision. Other models have product
data and references; manufacturer rules are never borrowed from related models. Interfaces or
product numbers that have not been researched remain explicitly unrecorded.

The structured data also preserves source locations, factory defaults, access types, units and
finite allowed-value sets for reviewed table rows. Variable or contradictory entries remain
unresolved. Factory defaults never replace vehicle readings. Additional table fields round-trip
through master data and backups; the existing interface continues to use names, help and bit details.

**CV definitions > Manufacturer CV tables** additionally provides 30,410 source-based CV
associations for 580 of the 599 catalog models. Search by CV number or term; range rows also
match numbers inside the range. Rows identify the source, PDF page and index-page certainty.
Coverage is partial: consult the linked original for model, hardware, sound-project and firmware
conditions. Another 19 models have no CV rows: 15 need matching documentation; four are auxiliary
modules or Motorola receivers without their own DCC CV table. The five legacy Lenz models
GOLD maxi, SILVER mini, SILVER direct, SILVER 21 and LF101XF now include further CV references
from their original manuals. Check the stated edition and generation; plus/successor models remain separate.
LE030 and LE0511A/LE0511D also have original references with printed source pages and different
speed-table boundaries. Indexed original texts are identified as such; the old direct LE0511
PDF URL was unavailable when reviewed.
Roco/Fleischmann 685404/685504/686101/686201 retain the explicit MX622 equivalence from the
manufacturer's 2022 catalogue. The separate ZIMO manual supplies 40 CV entries with source pages.
Check firmware and hardware applicability; OEM factory values remain unspecified.
Five special ESU LokPilot 5 variants and six PIKO XP sound articles also have documented
family-wide firmware CVs. Headings retain their firmware conditions; complete variant/project
tables and project defaults remain missing. PIKO CV48 has a firmware-dependent ABC threshold;
CV30 explicitly retains a version conflict between manufacturer sources. These remain qualified
reference hints.
References enable automatic interpretation only after separate review and never change readings or manual vehicle functions.
Official Lenz CV Editor templates are identified as such; their editor defaults are not hardware
factory settings. References persist as `decoder_reference` master data and round-trip through
master-data transfers and full backups. Restarting preserves local changes.

Roco 10882 has a separate documentation reference; its reviewed definitions are accessible
in model details and search. Its manual names ZIMO with technical CV8=145.
This does not enable rules for other Roco decoders or identification using the retail brand ID.

“View order variants” opens the specific **article number → interface → source** mapping.
The two summarized lists in the overview do not establish pairs. ZIMO order designations also
serve as article numbers. Historical documentation and announced products are marked; this
does not confirm current availability. Roco retail articles and selected earlier PIKO decoders
remain marked as branded entries: the registered brand ID does not establish actual CV8.
An OEM type name does not enable extra CV rules or create a manufacturer conflict on its own.

Sources state their manual revision and scope; the directory states its review date. Facts are
summarized from manufacturer documentation and the NMRA manufacturer register. Linked documents
may change later. New sources become interpretation rules only after review; no background web
fetching is performed.

Admin and Editor can activate or deactivate reviewed profiles. Inactive definitions remain readable
in the library. Vehicle readings and provenance stay unchanged. A free-form profile editor is not
included. Master-data transfer and full backups include `decoder_profile` definitions and the
`decoder_catalog/library` model directory. Restarting preserves local edits and deactivation.

CV8 alone never confirms a model. Documented type rules cover 39 of the catalogued models: D&H uses
CV261, ZIMO CV250 and Lenz CV7. **Identification** shows each model's values and limits. ZIMO
identifies the base type only, without confirming its interface variant. ESU still requires a
matching profile name and CV8.

At startup, new bundled explanations replace only known, unedited previous revisions. Custom
metadata, labels, ordering and deactivation remain unchanged. See
[Decoder library on vehicles](/guide/vehicles/decoder-cv#decoder-library) for identification and scope.
