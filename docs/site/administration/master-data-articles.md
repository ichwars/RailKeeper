---
title: Article master data and storage locations
description: Configure accessory units, classifications, custom fields, and hierarchical storage locations.
audience: admin
status: stable
reviewedVersion: 0.1.19
lastReviewed: 2026-08-16
---

# Article master data and storage locations

Open **Settings > Data > Article data** for the controlled values used by accessory articles. The
page separates stock units, article classification, custom fields, and storage locations. Admin and
Editor can change them. Viewer and Planner receive an explicit read-only view.

## Stock units

Stock units describe what article quantities count, for example pieces, packs, metres, or grams.
Each entry has an immutable key, editable label, active state, and origin.

Create a unit only when the existing list cannot express the stock. Deactivating a unit removes it
from new article selections, while articles already using it retain the historical value. A custom
unit can be permanently deleted only when no article references its key.

## Article types and subtypes

RailKeeper ships eight protected article-type keys:

- Track
- Signal
- Decoder
- Electrical / control
- Building / equipment
- Landscape consumable
- Lighting
- Other

Their labels and active states can be maintained, but v0.1.19 does not allow creating or deleting
article-type keys. This protects the matching technical-data contracts. Deactivating a type removes
it from new article selection without rewriting existing articles.

Subtypes refine one of those types. Admin and Editor can create, edit, deactivate, reactivate, and,
while unused, delete a custom subtype. Its key must begin with the parent article-type key and a
colon, for example `track:turnout`. RailKeeper uses that prefix to decide which subtypes belong to
the selected article type.

Changing or deactivating a classification never migrates article records automatically. Review the
affected articles and decide whether their historical classification should remain or be edited
individually.

## Controlled custom fields

Custom fields appear only on articles of type **Other**. Each definition has a key, label, active
state, and one of these data kinds:

| Kind | Configuration and resulting value |
| --- | --- |
| Text | Free text |
| Number | Numeric value and optional unit |
| Yes / No | Explicit boolean value |
| Date | Date value |
| Single selection | Exactly one configured option |
| Multiple selection | Zero or more configured options |

Single- and multiple-selection definitions require a non-empty comma-separated option list. The UI
removes empty options; the server rejects duplicate options. Choose the kind carefully: once a custom
field is referenced by stored **Other** article attributes, its key cannot be deleted. Historical
values whose definition becomes inactive remain readable but cannot be newly added or changed.

## Storage locations

Storage locations form a hierarchy independent from the controlled master-data table. Each
location has a required name, optional parent, optional description, and archived state. The list
shows the full path, for example `Store room / Cabinet A / Drawer 3`.

Names must be unique among siblings without regard to letter case. A location cannot be its own
parent or be moved beneath one of its descendants. The parent picker excludes those invalid
choices.

There is no permanent-delete action in v0.1.19. Archive a location to retire it and reactivate it
when required. Stock and individual items can use a location only when that location and every
ancestor are active. Archiving a parent therefore makes its entire branch unavailable for new
stock operations without erasing stored location references.

To build a safe hierarchy:

1. Create the top-level room or store.
2. Create cabinets or shelves with that top-level parent.
3. Add bins or drawers under the appropriate branch.
4. Verify the full displayed path before assigning stock.
5. Archive obsolete leaves first, then their empty parents.

Storage locations are included in application backup and restore, but not in the standalone
master-data JSON transfer.

## Lifecycle and conflict handling

The action column reflects server-calculated capabilities. Bundled entries never show permanent
deletion. Referenced custom entries also lose that action. Deactivation and deletion require a
confirmation; reactivation writes immediately.

If a save reports a conflict, check for a duplicate sibling location, a protected article type, a
referenced custom value, or an invalid custom-field option list. A rejected operation leaves the
stored configuration unchanged.

## Related pages

- [Master-data administration](./master-data)
- [General master data](./master-data-general)
- [Article records and technical data](/guide/accessories/article-records)
- [Stock, purchases, and documents](/guide/accessories/stock-purchases-documents)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19** and was last reviewed on 2026-08-16.
