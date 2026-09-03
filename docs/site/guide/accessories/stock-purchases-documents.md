---
title: Stock, purchases, and documents
description: Manage accessory quantities, individual items, purchases, images, and documents.
audience: user
status: stable
reviewedVersion: 0.1.20.4
lastReviewed: 2026-08-16
---

# Stock, purchases, and documents

RailKeeper can count interchangeable accessory units, track physical items individually, or move
units from counted stock into individual tracking when needed. Purchases and documents belong to
the same article but have their own immediate save actions. This chapter describes the stable
behavior of RailKeeper v0.1.20.3.

Admin and Editor users can manage these resources. Viewer and Planner users can inspect them, but a
Planner's reservation authority does not include stock, purchase, item, or document changes. Save
the article record once before creating any related resource.

## Choose an inventory strategy

The strategy is stored with the article through the main **Save changes** action. It controls which stock
commands and records are available:

| Strategy | Counted stock | Individual items | Purchase booked to stock |
| --- | --- | --- | --- |
| **Quantity** | Adjust and transfer quantities by storage location. | Not available. | Adds the purchase quantity to the selected location and writes a Purchase movement. |
| **Individual** | No quantity adjustment or transfer. | Create and maintain one record per physical item. | Creates one stored item per purchased unit, with condition Unknown. It creates no counted stock movement. |
| **Quantity, individual later** | Adjust and transfer quantities. | Convert one available unit at a time into an individual item. | Adds counted stock like Quantity. Purchased units can be individualized later. |

Choose **Quantity** for interchangeable material, **Individual** when every unit needs its own
identity or condition, and the hybrid strategy when packages begin as quantities but selected units
later need lifecycle tracking.

Changing strategy can be blocked by existing stock and allocations. Quantity values, quantity
reservations, or quantity installations prevent a change to individual-only tracking. Existing
items and their reservation, installation, or condition history prevent removing individual
tracking. A rejected change writes nothing.

**Minimum stock** is a non-negative whole-number planning value. It is saved with the article and
does not itself create, reserve, or order stock.

## Read stock and availability

The **Stock** tab lists the counted quantity for every referenced storage location. The total is the
sum of those rows. Individual items appear in a separate table and are not added to this counted
total.

Counted stock is owned physical quantity. **Available** quantity is lower when active quantity
reservations exist. RailKeeper therefore prevents an adjustment, transfer, or individualization
from reducing a location below its actively reserved quantity. Installed counted units have
already been removed from stored quantity.

The stock journal is newest first and contains these stable movement types:

| Movement | Quantity and source |
| --- | --- |
| Purchase | Positive purchased quantity at the destination, linked to the purchase. Quantity and hybrid strategies only. |
| Adjustment | The signed manual delta at one location. |
| Transfer out | Negative quantity at the source, linked to the same transfer as Transfer in. |
| Transfer in | Positive quantity at the destination. |
| Individualization | `-1` at the source, linked to the newly created item. |
| Installation | Negative installed quantity at the source for quantity-based installation. |
| Removal | Positive returned quantity at the destination when a quantity installation is removed back to stock. |

The journal shows date, movement type, signed quantity, and note. A transfer writes two rows in one
transaction. Reservation creation and cancellation do not write stock movements because they
change availability, not owned quantity.

## Adjust quantity stock

Use **Adjust stock** for a physical correction that is not a purchase or transfer:

1. Select an active storage location.
2. Enter a non-zero whole-number delta. A positive value adds units; a negative value removes
   units.
3. Select **Book** and review the confirmation containing the signed quantity and article name.
4. Confirm to write the new location quantity, one Adjustment journal row, and the audit event in a
   single transaction.

The action is available for Quantity and hybrid articles only. RailKeeper rejects an inactive
location, zero, a non-whole value, or a reduction below active reservations. Canceling the
confirmation writes nothing. After success, all article resources are reloaded.

Use purchases for incoming goods whose supplier or invoice should be recorded. A manual positive
adjustment creates no purchase record. There is no note field on the stable adjustment command.

## Transfer stock between locations

Select different active source and destination locations, enter a positive whole quantity, and
optionally add a note. **Transfer** opens a confirmation. Confirming atomically:

- subtracts the quantity from the source without touching reserved units;
- adds it to the destination;
- writes linked Transfer out and Transfer in journal rows with the same optional note; and
- reloads the article resources.

The transfer fails without changing either location when the source lacks enough available
quantity, both locations are equal, a location or its ancestor is archived, or any required value
is missing. The command exists only for Quantity and hybrid articles.

## Manage individual items

Individual and hybrid articles show the item table. For an Individual article, **Create item** adds
a new physical record without changing counted stock. For a hybrid article,
**Individualize** consumes exactly one available counted unit from the selected location and creates
the item in the same transaction. It also writes an Individualization movement of `-1`.

| Item field | Stable rule |
| --- | --- |
| Inventory number | Optional item identity. When supplied, it must be unique without regard to case across accessory items. It is separate from the article inventory number. |
| Serial number | Optional manufacturer serial number. |
| Condition | Ready, Maintenance due, Defective, or Unknown. Default is Ready for manual entry. |
| Lifecycle | Stored, Maintenance, or Retired. Default is Stored. Reserved and Installed are controlled by allocation commands and cannot be chosen here. |
| Storage location | Required active location in the stable form. A hybrid item is taken from this same location. |
| Purchase date | Optional calendar date in `YYYY-MM-DD` storage format. |
| Purchase price | Optional non-negative number, entered in 0.01 steps. |
| Warranty until | Optional calendar date. |
| Notes | Optional local text. |

**Save item** always opens a confirmation. For a hybrid article, insufficient available quantity
cancels the whole individualization, including the new item. Empty inventory and serial numbers are
allowed, but a duplicate non-empty inventory number is rejected.

Use the edit icon to load an existing item into the same form. Saving the edit also requires
confirmation and changes only that item. An item whose lifecycle is Reserved or Installed has no
edit action in this table. Change it through the matching reservation or installation workflow so
that lifecycle, location, and history remain consistent.

## Record purchases

The purchase list is a stable, newest-first history. Stable v0.1.20.3 provides an add action but no
purchase edit or delete action.

| Purchase field | Stable rule |
| --- | --- |
| Purchase date | Required valid date, defaulting to today. |
| Supplier | Optional text. |
| Quantity | Required positive whole number. |
| Unit price | Optional value. A numeric value produces the displayed quantity-times-unit-price preview. |
| Currency | The stable form uses EUR. The stored API value, when present, is a three-letter uppercase code. |
| Invoice number | Optional text. |
| Warranty until | Optional valid date. |
| Book to stock | When clear, only the purchase record is created. When selected, an active destination is required and stock is created in the same transaction. |
| Storage location | Required only for Book to stock. |
| Notes | Optional text; copied to the Purchase movement for counted stock. |

Selecting **Book purchase** writes immediately without a confirmation. The effect depends on the
article strategy:

| Strategy and option | Atomic result |
| --- | --- |
| Any strategy, Book to stock off | One purchase record; no stock, item, or movement change. |
| Quantity, Book to stock on | One purchase, the full quantity added at the location, and one linked Purchase movement. |
| Hybrid, Book to stock on | Same as Quantity. Individualize selected units later. |
| Individual, Book to stock on | One purchase and one Stored item per unit. Each item inherits purchase date, unit price, warranty, location, and purchase link; condition starts Unknown. No quantity movement is written. |

Purchase creation and optional stock booking run in one database transaction. If any part fails,
neither the purchase nor its stock/items is stored. After success, the form resets and RailKeeper
reloads the related resources.

## Manage images and documents

The document area lists the original file name and category. Every user with accessory read access
can download a document. Admin and Editor users can upload, choose the primary image, and delete.

| Operation | Confirmation and effect |
| --- | --- |
| Upload | No confirmation. Select a file, category, and optional description. The first uploaded Image becomes primary when no primary image exists. Resources reload after success. |
| Make primary image | No confirmation. Available on a non-primary Image. It clears the previous primary flag and sets this image in one transaction, then reloads. |
| Download | No write and no confirmation. Images may display inline; other types download as attachments. |
| Delete | Destructive confirmation. It removes the document metadata, then deletes the stored blob only when no other record references it, and reloads. Deleting the primary image does not automatically promote another image. |

The categories are Invoice, Delivery note, Manual, Data sheet, Floor plan, Image, and Other. Only an
Image can be primary.

Accepted upload extensions are `.pdf`, `.jpg`, `.jpeg`, `.png`, `.webp`, `.zip`, `.txt`, `.csv`,
`.json`, and `.xml`, and detected content must match the extension. JSON, CSV, and XML receive
additional structural checks. Empty files, unsafe names, executable or script types, HTML, MIME
mismatches, and files above the configured attachment limit are rejected. The default limit is
25 MB through `RAILKEEPER_MAX_ATTACHMENT_MB`. Uploaded blobs stay inside RailKeeper's local data
model and downloads require an authorized application request.

Images selected in article-data search use a separate URL import. RailKeeper accepts only public
HTTP or HTTPS addresses, rejects localhost and private or internal addresses before connecting and
on redirects, allows at most five redirects, uses a ten-second request limit, and accepts only
detected JPEG, PNG, or WebP content within the same size limit. The idempotency key prevents the
same queued image from being stored twice for the article.

Accessory articles, counted stock, movements, items, purchases, document metadata, and stored file
blobs are included in the application backup. Keep current validated backups before destructive
document or inventory cleanup.

## Protect data during immediate writes

Stock commands, item saves, purchase creation, document upload, primary-image changes, and document
deletion persist independently of the main article **Save changes** action. Their confirmations protect the
command shown above, not unsaved article fields elsewhere in the dialog.

After each successful sub-action, RailKeeper requests stock, journal, items, purchases, documents,
reservations, installations, history, locations, vehicles, and layout resources again. Those reads
can partially fail after the write has already committed. In that case:

1. Treat the command as potentially successful. Do not submit it again from stale data.
2. Read the resource error shown by the editor. RailKeeper marks resources stale and disables
   further stock, document, reservation, and installation writes.
3. Select **Retry** until a complete reload succeeds.
4. Verify the stock row, item, purchase, document, and journal before deciding whether another write
   is needed.

The same rule matters after document deletion: metadata deletion commits before cleanup of an
unreferenced blob. An error during cleanup can therefore follow an already removed list entry.
Reload before attempting to delete again.

## Resolve stock and document errors

| Situation | Next step |
| --- | --- |
| Adjustment or transfer reports insufficient stock | Check active reservations at the source. Use only the free quantity. |
| Location is missing or rejected | Select an active location whose parent chain is also active. Storage-location administration is separate from this article dialog. |
| Individualization fails | Confirm that the article is hybrid and the selected location has at least one unreserved counted unit. |
| Item inventory number conflicts | Use a different item number or edit the existing item. Letter case does not create a distinct number. |
| Reserved or installed item cannot be edited | Complete or cancel its allocation through the reservation or installation controls. |
| Purchase is rejected | Check purchase date, positive whole quantity, warranty date, and destination when booking to stock. |
| File type is unsupported | Use an accepted extension with matching real content; renaming a file is not sufficient. |
| File is too large | Reduce it below the configured attachment limit or ask the operator to review the limit. |
| Remote image is rejected | Use a public HTTP(S) JPEG, PNG, or WebP URL that does not redirect to a private address. |
| Write succeeded but refresh failed | Do not repeat it. Retry the resource load and verify the journal or resource list first. |

## Related pages

- [User Guide overview](/guide/)
- [Accessories overview](./)
- [Article records and technical data](./article-records)
- [Reservations, installations, and usage](./allocations-history)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20.3** and was last reviewed on 2026-08-16.
