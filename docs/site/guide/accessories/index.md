---
title: Accessories overview
description: Find, filter, inspect, and safely manage accessory articles.
audience: user
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# Accessories overview

**Accessories** is RailKeeper's catalogue and inventory workspace for track, signals, decoders,
electrical components, scenery, lighting, and other model railway articles. It combines product
identity, stored and allocated quantities, images, storage locations, and article lifecycle actions.
This chapter describes stable RailKeeper v0.1.20.

An article is the common product record. Its stock can consist of interchangeable quantities,
individually tracked items, or both. Reservations and installations bind some of that stock to a
vehicle or another configured target. The overview combines those states but does not change stock
by itself.

## Access rights and workspace model

Admin, Editor, Viewer, and Planner users can open **Accessories** and inspect article data. Their
write rights differ:

| Role | Stable overview access |
| --- | --- |
| Viewer | View articles and their stored resources. |
| Planner | View articles and later create or cancel reservations. General article and stock editing remains read-only. |
| Editor | Create and edit articles, archive or restore them, and manage their stock and lifecycle. |
| Admin | Has Editor rights and can additionally delete a completely unused article permanently. |
| Messe | Has no accessories access through the Messe role alone. |

The page shows an explicit read-only note to Viewer and Planner users. Server-side role checks are
authoritative. A visible value or read action does not grant write permission.

The counts and rows come from the local RailKeeper database. Changing search, a filter, or sorting
starts a new server request. The result counter therefore describes the current searched and
filtered list, while the four metrics always summarize all non-archived articles.

## Read the overview metrics

Four metric cards appear above the article list:

| Metric | Stable meaning and action |
| --- | --- |
| **Article inventory** | Number of non-archived article records and distinct article types. Selecting it clears every active filter. |
| **Available** | Total free quantity plus the number of distinct storage locations referenced by active articles. Selecting it shows articles with more than zero available units. |
| **Allocated** | Total active reservations plus active installations. Selecting it shows articles that have either reserved or installed quantity. |
| **Care hints** | Total number of incomplete-data hints across active articles. It is informative and has no filter action. |

**Care hints** is a count of missing fields, not a count of affected articles. One article can
contribute several hints. Stable v0.1.20 checks for missing manufacturer, article number, article
type, and stock unit. Track, signal, decoder, electrical/control, building/equipment, and lighting
articles also contribute a hint when they have no gauge. Landscape consumables and **Other** do not
require a gauge for this metric.

The metric cards ignore the current search and filter selection. Use the result counter below them
for the size of the visible result.

## Search and filter articles

Enter text in **Search articles**. The server performs a case-insensitive substring search in five
fields:

- Inventory number
- Manufacturer
- Article number
- Name
- EAN

Description, keywords, alternative numbers, technical attributes, storage notes, purchase data,
and usage history are not part of the stable free-text search.

The additional filters are:

| Filter | Choices and behavior |
| --- | --- |
| Article type | Types that occur on active articles. |
| Manufacturer | Manufacturers that occur on active articles. |
| Gauge | Gauges that occur on active articles. |
| Status | Available, Reserved, Installed, Maintenance due, Defective, or Archived. |
| Storage location | Active storage locations. An article matches when quantity stock or an individual item is linked to that location. |

Different filter groups combine with AND logic. A row must match the text and every selected
group. **Allocated** from the metric card is the one combined status: it matches **Reserved** OR
**Installed**. The normal list excludes archived articles. Selecting **Archived** replaces that
default and shows archived rows matching the remaining filters.

The status filters mean:

| Status | Included articles |
| --- | --- |
| Available | More than zero units are currently free after active reservations. |
| Reserved | At least one unit is in an active reservation. |
| Installed | At least one unit is in an active installation. |
| Maintenance due | An individual item or active installation has condition **Maintenance due**. |
| Defective | An individual item or active installation has condition **Defective**. |
| Archived | The article record is archived. |

**Reset filters** clears text and all five filter groups. It does not change the selected desktop
view, visible columns, or sorting.

## Choose columns, view, and sorting

At desktop widths, select table or card view. RailKeeper stores that choice in the current
browser's local storage. It is not saved to the user account and does not follow the user to another
browser.

At widths up to 900 pixels, RailKeeper always shows the compact list. The stored table/card choice
remains available when the window becomes wider again.

The table offers these nine data columns:

| Column | Displayed value |
| --- | --- |
| Image | Primary accessory image, or the generic image symbol. |
| Inventory number | RailKeeper's local article identity. |
| Manufacturer | Product manufacturer. |
| Article number | Manufacturer or catalogue number. |
| Name | Article name and entry point to the read-only view. |
| Type / subtype | Configured article classification. |
| Gauge | Every gauge assigned to the article. |
| Stock | Total owned plus free, reserved, and installed quantities. |
| Storage | First storage location and the number of additional locations. |

Every displayed data column is sortable. The initial order is inventory number ascending. Select
the current sort header again to reverse its direction; selecting another header starts that column
ascending. Rows with otherwise equal values use their internal ID as a stable tie-breaker.

The column chooser is available only in table view. Column visibility is also stored in the current
browser. All nine columns start visible. Inventory number and Name cannot both be hidden, so the
table always retains at least one visible article identity.

Card view emphasizes image, identity, type, gauge, storage, and stock. The compact list keeps a
smaller identity and stock summary. Both follow the current server sort but do not expose separate
sort controls.

## Select, open, and manage an article

Table checkboxes select one row or all currently visible rows. Selection is removed when a new
search or filter result no longer contains the article. Stable v0.1.20 provides no bulk action for
this selection. It is visual state only and must not be treated as a pending archive, report, or
stock command.

Select the article name, image, or **View article** to open the read-only article dialog. Depending
on the role and article state, row actions also provide:

| Action | Role and persistence |
| --- | --- |
| View article | Every role with accessories read access. No write. |
| Edit article | Admin or Editor. The article form remains draft state until saved. |
| Archive article | Admin or Editor. Writes immediately without a confirmation and reloads the overview. |
| Restore article | Admin or Editor. Writes immediately without a confirmation and reloads the overview. |
| Delete article | Admin only. Opens a permanent-deletion confirmation. |

Permanent deletion succeeds only for a completely unused article. Non-zero stock, individual
items, stock-movement history, purchases, reservations, installations, or a layout technical
position referencing the article blocks it. Stored article documents do not block an otherwise
unused article; RailKeeper removes their metadata and then attempts to remove their stored blobs as
part of the deletion workflow.

There is no undo for deletion. Create and validate a current application backup before deleting an
important record. Full deletion and recovery rules are covered with the article-record lifecycle.

## Resolve loading, empty, and error states

| Situation | RailKeeper behavior and next step |
| --- | --- |
| Initial load | **Loading articles...** appears until the first request completes. |
| No article exists | **No articles have been created yet.** Admin or Editor can create the first article. |
| Filters have no match | **No articles match the active filters.** Reset or adjust the search and filters. |
| Loading fails before any result | Only the error appears above the empty list area. Check the session and connection, then retry by changing or resetting a filter. |
| Loading fails after data was shown | The error appears while the earlier rows can remain visible. Do not treat them as current; retry the request before acting on the result. |
| Master-data labels fail to load | Rows can fall back to stable translated type/subtype labels. Open the article only after required editor resources load successfully. |
| Archive or restore fails | The error appears above the list. Treat the article as unchanged until a successful reload confirms otherwise. |
| Delete is blocked | The confirmation remains relevant, but the record is not deleted. Resolve every reported stock or usage reference before retrying. |
| Write is forbidden | Sign in with the required Editor or Admin authority. Read access does not override server write checks. |

## Continue with an article workflow

The detailed accessories section separates three workflows:

- create and maintain the article record and its technical data;
- manage stock, purchases, individual items, images, and documents;
- reserve or install stock and read its usage history.

These workflows share one article dialog but have different persistence and permission rules. Save
or intentionally discard unsaved article fields before starting an immediate stock or lifecycle
write.

## Related pages

- [User Guide overview](/guide/)
- [Article records and technical data](./article-records)
- [Stock, purchases, and documents](./stock-purchases-documents)
- [Reservations, installations, and usage](./allocations-history)
- [Overview, metrics, and data quality](/guide/overview/)
- [Vehicle inventory and core records](/guide/vehicles/)
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20** and was last reviewed on 2026-08-16.
