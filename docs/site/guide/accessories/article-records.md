---
title: Article records and technical data
description: Create accessory records, maintain their identity, and verify technical data.
audience: user
status: stable
reviewedVersion: 0.1.20.1
lastReviewed: 2026-08-16
---

# Article records and technical data

An accessory article is the shared product record behind stock, purchases, documents,
reservations, and installations. Create it once for a manufacturer and catalogue article, then
record quantities or individual items against it. This chapter covers the article identity and the
type-specific data in stable RailKeeper v0.1.20.1.

## Access rights and persistence

Admin and Editor users can create and edit article records. Viewer and Planner users can inspect
them, but article fields remain read-only. A Planner can create or cancel a reservation in the
otherwise read-only dialog; that does not grant permission to change the article itself. The Messe
role has no general accessories access.

The dialog separates **Article**, **Stock**, **Purchase & documents**, the article-type tab, and,
when data exists, **Usage & history**. The main **Save changes** action persists the article form, including
the article type, technical fields, inventory strategy, and minimum stock. Stock movements,
purchases, documents, reservations, and installations use their own actions and can already be
stored while other article fields remain unsaved. Finish or deliberately discard a draft before
starting one of those immediate writes.

Closing the dialog with changed article fields, pending searched images, or an unfinished subform
opens a discard confirmation. In create mode, stock and related resources cannot be persisted until
the core article has been saved once.

## Create an article

1. Open **Accessories** and select **Create article**.
2. On **Article**, select a manufacturer, article type, and subtype, then enter the name.
3. Add catalogue identity, gauges, scale, package quantity, stock unit, and optional reference data.
4. Complete the tab named after the selected article type where technical data is known.
5. On **Stock**, choose the inventory strategy and minimum stock. The detailed consequences are
   covered in [Stock, purchases, and documents](./stock-purchases-documents).
6. Select **Save changes** and resolve any marked tab or duplicate warning.

The required fields are manufacturer, name, subtype, stock unit, and a positive whole package
quantity. Minimum stock must be a whole number of zero or more. Technical numbers must respect the
displayed range and increment. A comma is accepted as the decimal separator.

RailKeeper assigns the article inventory number only during the successful create transaction. It
uses the active **Artikel** inventory-number scheme, whose default format produces values such as
`RK-ART-000001`. A failed create does not consume the reserved number. The field cannot be entered
or changed in the article dialog. If no active scheme exists, creating the article is blocked until
an administrator repairs the inventory-number configuration.

For a new record, RailKeeper selects the first active configured article type. If no active article
type is available, creation remains blocked. The initial values are package quantity 1, stock unit
**Piece**, minimum stock 0, inventory strategy **Quantity**, and manufacturer status **Unknown**.

## Common field reference

| Field | Stable behavior |
| --- | --- |
| Inventory number | Assigned automatically from the **Artikel** scheme and read-only. |
| Manufacturer | Required active master-data selection. A stored inactive or historical value remains visible but cannot be newly selected. |
| Article number | Optional manufacturer or catalogue number. When present, it activates the duplicate check together with manufacturer. |
| Name | Required article name. |
| EAN / GTIN | Optional barcode value and an independent article-search criterion. |
| Manufacturer status | Announced, Available, Discontinued, or Unknown. |
| Article type | Track, Signal, Decoder, Electrical / control, Building / equipment, Landscape consumable, Lighting, or Other, subject to active configuration. |
| Subtype | Required configured subtype belonging to the selected article type. An unchanged historical subtype remains readable and preservable. |
| Gauges | Zero or more active gauge master-data values. Existing inactive values remain displayed. |
| Scale | Optional text. While automatically managed, it follows the first selected active gauge that defines a scale. |
| Package quantity | Required positive whole number describing how many stock units a package contains. |
| Stock unit | Required active master-data value, for example Piece. Existing inactive values remain displayed. |
| Description | Optional free text. |
| Manufacturer URL | Optional manufacturer reference. It is not filled by the stable accessory result picker. |
| Product URL | Optional product reference. Search results can fill it only from an HTTP or HTTPS source. |
| Alternative article numbers | Optional values separated by commas or line breaks. Empty entries are removed. |
| Keywords | Optional comma- or line-break-separated terms. Automatic suggestions combine name, manufacturer, article type, and subtype without case-insensitive duplicates. |
| Compatibility notes | Optional free text for supported systems or products. |
| Internal notes | Optional local notes. |

Leading and trailing whitespace is removed from ordinary text fields. Manually changing **Scale**
or **Keywords** stops the respective automatic suggestion for the current edit session. The
automation runs in create and edit mode, never in read-only view.

Manufacturers, article types, subtypes, gauges, stock units, and custom fields come from master
data. RailKeeper preserves an unchanged inactive value on an existing article, but does not allow
it to be newly added or changed. This makes historical records editable without silently replacing
their classification.

## Choose article type and technical data

The last main tab is named after the selected article type. Its fields are optional unless the
local master-data configuration says otherwise, but every entered value is validated. Whole-number
fields use increment 1; other non-negative numbers have no fixed increment unless listed below.

| Article type | Stable technical fields |
| --- | --- |
| Track | Track system; length and radius in mm; angle 0 to 360° in 0.1° steps; direction Left, Right, or Symmetric; frog angle 0 to 180° in 0.1° steps; sleeper type; rail height in mm; Roadbed; non-negative whole connection count; Digital ready. |
| Signal | Prototype; epochs I to VI; aspects Stop, Proceed, Caution, and Shunting; non-negative whole LED count; height in mm; non-negative AC and DC operating voltage; mounting Surface, Flush, Mast, or Wall; drive Manual, Motor, Servo, or Solenoid; Integrated decoder; control module. |
| Decoder | Interface Wired, NEM 651, NEM 652, PluX16, PluX22, 21MTC, or Next18; protocols DCC, Motorola, Selectrix, mfx, and RailCom; non-negative whole function outputs and servo outputs; non-negative motor, output, and total current in mA; RailCom; SUSI; dimensions; firmware. |
| Electrical / control | Non-negative input voltage, output voltage, current in A, and power in W; non-negative whole channel count; protocols DCC, Motorola, Selectrix, LocoNet, CAN, and S88; connectors Screw, Plug, RJ45, Bus, and Wire; protections Short circuit, Overload, Temperature, and Reverse polarity; compatible article types Track, Signal, Decoder, and Lighting. |
| Building / equipment | Epochs I to VI; dimensions; footprint; material; construction type Kit, Finished model, or Semi-finished kit; non-negative whole part count; difficulty Easy, Medium, or Advanced; lighting options Interior lighting, Platform lighting, Street lighting, and Effect lighting; Floor plan available. |
| Landscape consumable | Material; color; season; non-negative content; content unit Piece, Pack, Meter, Gram, or Milliliter; fibre or grain size; coverage; suitable scales Z, N, TT, H0, 0, 1, and G; safety notes. |
| Lighting | Light color; non-negative color temperature in K, voltage in V, and current in mA; power type AC, DC, or AC or DC; non-negative whole LED count; Dimmable; dimensions; mounting Surface, Flush, Mast, or Wall. |
| Other | Active configured custom fields. Supported kinds are text, number, yes/no, date, single selection, and multiple selection. Selection fields appear only when they have configured choices. |

Boolean fields distinguish an explicit Yes or No from no stored value. Multiple-selection fields
reject unknown or repeated choices. Negative values, values outside a listed range, and values that
do not match a listed increment prevent saving and mark the article-type tab.

Changing the article type clears the subtype. If the change would discard the subtype, a technical
value, or an unfinished number, RailKeeper asks for confirmation. Compatible fields with the same
key and data type remain; incompatible fields are removed only after confirmation. Cancel the
confirmation to keep the current type and values.

For **Other**, the fields are defined by active `accessory_custom_field` master data. A historical
custom value whose definition has become inactive or disappeared is preserved only while it remains
unchanged. It cannot be newly added or edited.

## Use barcode and article-data search

The **Article** tab offers the same reviewed-result workflow used by vehicle article search, with
accessory-specific inputs and target fields. The normal search is available when either:

- EAN is present; or
- manufacturer and the first selected gauge are present, together with either article number or
  name.

RailKeeper sends manufacturer, article number, name, the first gauge, current common fields, and
the current technical attributes to the configured search sources. If article search is disabled in
Settings, the command stops without changing the article.

The result dialog can apply manufacturer, article number, name, EAN, scale, description, product
URL, and gauge. A manufacturer or gauge is selectable only when it matches an active known master-
data value. A selected gauge is added to the existing gauges. Invalid common values and non-HTTP(S)
product URLs are ignored. In stable v0.1.20.1, selectable technical result fields are exposed only
for **Track** articles; other article types can apply the common result groups but not their
technical attributes.

Fields whose current form value is empty start selected in the review. Existing values are not
preselected for replacement. Images are never selected automatically. Select only trusted values,
then apply them to the local draft. Nothing is saved until the article save succeeds.

**Barcode** opens the scanner with the current EAN. A non-empty scanned or manually entered value
replaces the local EAN and immediately starts an EAN-only search. Camera permission and manual-entry
behavior are described in [Article search, web documents, and spare parts](../vehicles/search-and-spares).

Selected search images are queued. After the core article saves, RailKeeper imports them one after
another. The first successful image becomes primary only when the article has no primary accessory
image. A failed image does not roll back the article or earlier images, and later queued images are
still attempted. The dialog remains open with the first import error and keeps failed images
pending; close it only after deciding whether to retry or discard them.

## Review duplicate candidates

When article number is not empty, saving first checks for an exact manufacturer and article-number
match after trimming and without case sensitivity. The current article is excluded while editing.
Name, EAN, gauge, and subtype do not change this match.

If candidates exist, RailKeeper lists them before writing. Select **Cancel** to return to the draft
or **Save anyway** to create or update the captured values. This warning permits intentional
variants; it is not a merge and does not transfer stock, documents, or usage from another article.

## Edit without losing data

Open **Edit article**, change the draft, and use the main **Save changes** action. Validation moves to the
first affected tab in this order: Article, Stock, Purchase & documents, then the article-type tab.
Red tab markers identify tabs that still contain invalid values.

Changing inventory strategy is allowed only when all existing inventory and allocation data remain
representable. In particular, RailKeeper blocks a change from quantity-capable tracking to
individual-only while quantity stock, quantity reservations, or quantity installations exist. It
also blocks removing individual tracking while individual items or their reservation,
installation, or condition history exists. Remove or complete those dependencies deliberately,
then retry. A rejected update leaves the stored article unchanged.

If required master data or custom-field definitions fail to load, the editor marks its resources
as stale and disables affected writes. Use **Retry**, do not replace displayed historical values
from memory, and save only after the resources have loaded. If the article detail itself failed to
load, close and reopen it rather than editing an incomplete fallback.

## Archive, restore, or delete an article

Admin and Editor users can archive or restore an article from the overview. Both actions write
immediately without a confirmation. Archived articles leave the normal list but remain available
through the **Archived** status filter and can be restored.

Only an Admin can permanently delete an article, and RailKeeper asks for confirmation. Deletion is
blocked by non-zero stock, individual items, stock movements, purchases, reservations,
installations, or layout technical positions that reference the article. Article documents alone
do not block deletion; their metadata is removed and RailKeeper then attempts to remove
unreferenced stored blobs. Deletion has no undo, so validate a current backup first.

## Resolve validation and resource errors

| Situation | Next step |
| --- | --- |
| Required field or number is invalid | Open the marked tab and correct every field message. |
| No inventory number can be assigned | Ask an administrator to activate or repair the **Artikel** inventory-number scheme. |
| Article type, subtype, or master data is unavailable | Retry resource loading. Only active configured values can be newly selected. |
| Duplicate candidates appear | Compare manufacturer and article number, then cancel or deliberately save the variant. |
| Strategy change reports a conflict | Resolve incompatible stock, items, reservations, installations, or condition history first. |
| Search cannot start | Add EAN, or complete the required manufacturer/name/article-number/gauge criteria; also verify Settings. |
| Search image import fails | The article may already be saved. Retry or discard only the failed pending images. |
| Save is forbidden | Use an Admin or Editor account. Planner reservation authority is not article-edit authority. |

## Related pages

- [User Guide overview](/guide/)
- [Accessories overview](./)
- [Stock, purchases, and documents](./stock-purchases-documents)
- [Allocations and usage history](./allocations-history)
- [Article search, web documents, and spare parts](../vehicles/search-and-spares)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20.1** and was last reviewed on 2026-08-16.
