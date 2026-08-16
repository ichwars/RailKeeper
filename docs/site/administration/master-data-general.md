---
title: General master data
description: Maintain vehicle classifications, manufacturers, CV8 identities, and function symbols.
audience: admin
status: stable
reviewedVersion: 0.1.19
lastReviewed: 2026-08-16
---

# General master data

Open **Settings > Data > General** to maintain the shared selections used by vehicles, accessories,
decoder data, and exhibitions. Stable v0.1.19 provides eight data types.

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

A symbol consists of a name, optional description, and optional image. The upload accepts SVG,
PNG, JPEG, or WebP and rejects files larger than 1 MiB. The image is stored inside the entry's
metadata, so it is included in master-data export and application backup.

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
cannot be deleted. If a write is rejected for Viewer or Planner, sign in with Admin or Editor
authority instead of retrying the visible control.

## Related pages

- [Master-data administration](./master-data)
- [Article master data and storage locations](./master-data-articles)
- [Inventory-number schemes](./master-data-inventory-numbers)
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19** and was last reviewed on 2026-08-16.
