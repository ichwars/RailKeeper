---
title: Vehicle inventory and core records
description: Search, filter, create, maintain, report, and safely delete RailKeeper vehicle records.
audience: user
status: stable
reviewedVersion: 0.1.20.3
lastReviewed: 2026-08-17
---

# Vehicle inventory and core records

**Vehicle inventory** is the central workspace for model railway vehicles. It combines inventory
status, search, filters, table and card views, core vehicle data, QR labels, and printable reports.
This chapter describes the vehicle workflow shipped with RailKeeper v0.1.20.3.

Admin, Editor, Viewer, and Planner users can inspect the inventory. Creating, changing, and deleting
vehicle records requires Admin or Editor. In v0.1.20.3, write controls can still be visible to Viewer
and Planner users. The server rejects their write requests, and RailKeeper displays the error.

Media, maintenance, decoder/CV data, article-data lookup, and spare parts have their own workflows.
This chapter explains their entry points but focuses on the core vehicle record.

## Read the inventory status

Open **Vehicle inventory** in the sidebar. The page heading is **Inventory**. Five status areas sit
above the list:

| Status | Meaning and action |
| --- | --- |
| **Total inventory** | Shows every loaded vehicle, category count, and gauge count. Selecting it clears all inventory filters. |
| **Digitalization** | Shows digital share plus digital and analog counts. Selecting it activates the **Digital** inventory filter. |
| **Maintenance** | Counts incomplete maintenance due today or overdue; the smaller number covers incomplete entries due within the next 14 days. Selecting it activates **Maintenance due**. |
| **Next appointment** | Shows the oldest overdue or nearest upcoming incomplete entry within 14 days. Selecting it opens that vehicle's read-only view. |
| **Image care** | Shows the percentage and count of vehicles with at least one image. Selecting it activates **Without image**. |

The refresh icon reloads the current server search. RailKeeper also reloads the visible inventory
when the browser regains focus, comes online, or returns from a hidden tab.

## Search the inventory

Enter text in **Search inventory**. Each change reloads the vehicle list from the server. Leading
and trailing spaces are ignored, and the search matches substrings in these fields:

- Inventory number
- Manufacturer
- Article number
- Designation
- Series
- Vehicle number
- Decoder type
- Home depot / assignment location
- Maximum speed

It does not search descriptions, railway companies, epochs, categories, decoder numbers, EANs, or
other detail fields in v0.1.20.3. The filters below are applied in the browser to the vehicles
returned by the server search. Search and filters therefore combine.

If loading fails, the error appears above the list. Previously loaded rows can remain visible, so
resolve the error or refresh before treating them as current.

## Filter the inventory

Filter groups combine with AND logic. A vehicle must satisfy every active group.

| Group | Stable v0.1.20.3 choices |
| --- | --- |
| Inventory state | **All**, **Digital**, **Analog**, **With image**, **Without image** |
| Maintenance | **All**, **Maintenance due**, **Without maintenance** |
| Master data | Manufacturer, Category, Subtype, Railway company, Epoch |
| Digital technology | Adapter / interface |
| Operational flag | **Exhibition ready** |
| Overview data gap | **Without article no.**, **Without EAN**, or **Digital without decoder no.** when opened from **Overview** |

The **Without maintenance** label is imprecise in v0.1.20.3. It selects vehicles without a due
maintenance entry, including vehicles that have completed or non-due maintenance history.

Selecting a category limits the subtype choices to distinct subtypes among the currently loaded
server-search results in that category and clears the current subtype selection. **Clear filters**
returns all groups to their defaults and removes an Overview `gap` parameter from the browser
address. The result counter always reports the number after every active filter.

## Choose a view and sort

On a desktop, switch between table and card view with the view icons. RailKeeper stores that choice
in the current browser's local storage. It is not an account-wide setting. At narrow widths, the
interface uses compact, expandable vehicle cards regardless of the desktop preference.

Use **Choose table columns** to control visible columns and their order. The default selection is
Image, Inventory number, Manufacturer, Article number, Designation, Gauge, Epoch, and Exhibition
status. Additional short vehicle fields are available in domain groups. RailKeeper stores this
preference server-side in the user account, so it remains available after a restart and in another
browser.

Inventory number may be hidden. If that would leave no data column visible, RailKeeper restores it
automatically. **Reset to defaults** restores both selection and order. Columns introduced by a
later version remain hidden initially in existing user configurations.

Every offered column except Image and Exhibition status is sortable. If the active sort column is
hidden, RailKeeper falls back to Inventory number where possible or to the first remaining sortable
column.

The initial order is Inventory number ascending with numeric-aware comparison, so `...2` sorts
before `...10`. Select the active header again to reverse its direction. Card and mobile layouts
follow the current table sort but do not expose separate sort controls. Mobile cards use the
configured column selection and reveal additional fields in the configured order when expanded.

In the English locale, the table/card and refresh tooltips remain German in v0.1.20.3. This is a
localization limitation, not a different operation.

## Select vehicles

Table checkboxes select vehicles for reports. The header checkbox selects or clears all vehicles
currently visible after search and filters. Selection can remain in memory when a later filter hides
a row, but the report's **Selection** scope uses only selected vehicles that are still visible.

Card and compact mobile layouts do not show selection checkboxes. Switch to table view when building
a specific report selection.

## Open and inspect a vehicle

Select a vehicle designation, its image, **View**, or the eye icon. RailKeeper loads the complete
record and opens **Vehicle view**. Empty fields and empty sections are omitted. The view can contain:

- product, model, technical, ownership, and control data;
- configured functions;
- images and attachments;
- maintenance entries;
- up to the first 12 displayed CV values plus CV files.

The header actions open **Edit**, an individual detail report, or a QR code. The row quick menu also
opens Uploads, Maintenance, and Spare parts directly. Those specialized editors are covered by their
dedicated chapters. Selecting **Edit** still requires Admin or Editor permission at save time.

## Create a vehicle or set

Admin and Editor users create individual vehicles and related product sets through the same guided
dialog:

1. Under **Type & basics**, choose **Individual vehicle** or **Set** and enter manufacturer,
   designation, article number where known, gauge, and category. A set needs at least two members.
   Its next `RK-SET-*` number is a provisional preview; the server reserves the final number.
2. Under **Article data**, use the embedded barcode or web search, or continue manually. Ranked
   results and grouped field/image review stay inside the creation dialog. External values remain
   user-reviewed suggestions and are imported only when selected.
3. Under **Vehicle details**, enter acquisition data once for the set. Desktop tabs and the mobile
   selector separate **Set & acquisition** from every ordered member. Set values provide defaults,
   while each member can keep its own category, type, model, technical, storage, and condition data.
4. Select **Create** or **Create set**.

For an individual vehicle, leaving **Inventory number** empty uses the active inventory-number
scheme for the selected category. Creation fails if that category has no active scheme. A manual
number must be unique. After creation, the dialog switches to edit mode so the ID-dependent tabs for
functions, speed curve, CV, uploads, maintenance, and spare parts can be edited.

A set is created transactionally: either the set and all members are stored, or none of them are.
Shared article, purchase, storage, condition, and packaging values are stored on the set and used as
defaults for blank member fields during creation. Explicit member values remain unchanged and can be
edited later without modifying the canonical set. A blank member designation becomes the set
designation followed by its position, for example `TEE Roland (1)`. Blank member inventory numbers
are assigned automatically. After creation, the first member opens for further editing.

Every set receives a persistent, unique `RK-SET-*` inventory number. The inventory displays its
canonical row once, even if filters expose only some members, and reports visible versus total
counts. Desktop and mobile keep sets collapsed initially and reveal their indented members on
demand. Set actions open the summary, edit shared data, or duplicate into a new creation draft.

**Save draft** stores the versioned creation state in the current browser. A later dialog offers
resume or discard, and successful creation clears it. Assigning existing vehicles, detaching
members, and deleting a complete set are not available. Article-search data never becomes
authoritative without user selection.

## Core field reference

### Model and product identity

| Fields | Purpose and behavior |
| --- | --- |
| Inventory number | Unique local identity. Blank on create means automatic assignment; blank on update retains the existing number. |
| Article no., Source / URL | Manufacturer/catalog identity and optional source link. The source link is normally populated by lookup or import rather than a plain form input. |
| Manufacturer, Designation, Gauge | Required master/model identity. Manufacturer and gauge use configured master data. |
| Railway company, Epoch | Optional configured master data. |
| Category, Subtype | Both required. Subtype choices depend on the selected category. |
| Description, Series, Vehicle no. | Optional free-text model description and prototype identity. |
| Maximum speed, Home depot / operating location | Optional prototype and operating data. Maximum speed accepts whole values from 1 to 1000 km/h; the operating location is free text. |
| EAN, Production period, List price | Optional commercial data. Production period and list price are stored as trimmed text; the dashboard interprets parseable list prices. |
| Digital, Digital / decoder no. | The switch enables the primary digital decoder-number input. Exhibition eligibility uses this primary number. |
| DT / decoder, DT / decoder no., Decoder type | Optional second decoder flag/number and decoder description. |
| Exhibition ready, Exhibition | Operational flags. **Exhibition ready** is a normal record flag; **Exhibition** is coordinated with exhibition-list handling. |
| ABC brakes | Records ABC braking capability. |

### Technical details

| Fields | Purpose and behavior |
| --- | --- |
| Length (mm), Weight (g) | Free-text numeric inputs for physical dimensions. |
| Color, Lettering, Load, Interior, Axles | Descriptive physical and visual details. |
| Axle count, Traction tire count | Count fields with a numeric keyboard hint; the server stores trimmed text. |
| Wheelset, Power pickup, Adapter / interface | Technical selections from the stable static choice lists below. |
| Coupling (F=R), Front coupling, Rear coupling | When front and rear are equal, the rear value follows the front and cannot be edited separately. |
| Drive, Headlights, Lighting, Sound generator, Smoke generator | Each switch enables its adjacent description input. Switching a feature off does not automatically erase an existing description. |
| Create QR code | Enables the QR button inside the edit form when an inventory number or designation is available. |

### Vehicle, ownership, and condition

| Fields | Purpose and behavior |
| --- | --- |
| Acquisition, from/at | How and from which source the vehicle was acquired. |
| Price, Date | Purchase price as trimmed text and purchase date as a date field. |
| Location, Details | Broad storage location and a more precise shelf, box, or position. |
| Condition, Details | Standard condition choice plus free-text qualification. |
| Packaging | Packaging state. |
| Additional information | Longer notes that do not belong to another core field. |

## Stable select choices

The following choices are static in v0.1.20.3. Their stored values remain German in both UI
locales.

| Field | Choices |
| --- | --- |
| Wheelset | `2-Leiter DC`, `3-Leiter AC`, `NEM`, `RP25`, `Metall`, `Kunststoff` |
| Coupling | `NEM-Schacht`, `Kurzkupplung`, `Bügelkupplung`, `Klauenkupplung`, `Schraubenkupplung` |
| Power pickup | `Schiene`, `Oberleitung`, `Batterie`, `Akku` |
| Adapter / interface | `NEM 651`, `NEM 652`, `PluX12`, `PluX16`, `PluX22`, `MTC21`, `Next18`, `8-polig`, `21-polig` |
| Acquisition | `Kauf`, `Tausch`, `Geschenk`, `Erbe`, `Leihgabe`, `Sonstiges` |
| from/at | `Händler`, `Privat`, `Messe / Börse`, `Online`, `Auktion`, `Hersteller`, `Verein`, `Sonstiges` |
| Location | `Auf Anlage`, `Vitrine`, `Lager`, `Werkstatt`, `Transportbox`, `Ausgeliehen`, `Sonstiges` |
| Condition | `Neu`, `Neuwertig`, `Sehr gut`, `Gut`, `Gebraucht`, `Leichte Gebrauchsspuren`, `Reparaturbedürftig`, `Defekt` |
| Packaging | `Originalverpackung`, `Ersatzverpackung`, `Ohne Verpackung`, `Transportbox`, `Sonstiges` |

Manufacturer, gauge, epoch, railway company, category, and subtype come from the instance's
configured master data rather than these static lists.

## Edit a vehicle

1. Open **Edit** from the table, card, mobile row, quick menu, or read-only view.
2. Change the required or optional fields. Required fields cannot be saved empty.
3. Select **Save changes**.

Changing category clears a subtype that is not related to the new category. Enabling **Coupling
(F=R)** copies the front coupling to the rear; later front changes update both values.

Changing the inventory number requires a new unique value. RailKeeper records the old and new
numbers in inventory-number history and writes a vehicle-update audit entry. A successful save
reloads the inventory and keeps the dialog in edit mode. The server, not the visible button state,
enforces Admin or Editor permission.

## Create QR labels

A QR code needs at least an inventory number or designation. Its plain-text payload contains:

```text
Inventar-Nr.: <inventory number>
Bezeichnung: <designation>
Decoder-Nr.: <decoder number, only when available>
```

RailKeeper uses the primary digital decoder number first, then the DT decoder number. The QR dialog
can download PNG or SVG and open a printable label. The quick menu and read-only view can generate a
QR code whenever identity data exists. The **Create QR code** switch controls only the QR button in
the edit-form Details section in v0.1.20.3.

When either inventory number or designation is empty, the payload writes `-` for that field. At
least one of the two fields must contain a value before RailKeeper creates the QR code.

## Create inventory reports

The dialog starts with **Overview list**, title `Fahrzeugsammlung`, scope **All**, and QR code plus
image enabled. Change these defaults as needed:

1. Optionally search, filter, sort, and select vehicles.
2. Select **Create report** in the inventory toolbar.
3. Choose **Overview list** or **Detail list** and enter a title.
4. Under **Print**, choose **All** or **Selection**.
5. Include or exclude **QR code** and **Image**.
6. Select **Create report**, then use the browser's print dialog to print or save as PDF.

**All** means all currently searched and filtered vehicles in the current sort order, not every
record in the database. **Selection** uses only selected vehicles that remain visible. It is disabled
when no visible vehicle is selected.

The overview list is compact. The detail list loads complete records and can include functions,
maintenance, CV data, images, attachments, and external mappings. Printing one vehicle from its
quick menu or read-only view always creates a detail report with QR code and images enabled.

Reports and some report labels are generated in German even under the English locale in
v0.1.20.3. If no matching vehicle exists, RailKeeper does not create a report. Allow the print window
if the browser blocks popups.

## Exhibition switch boundary

The table's **Exhibition** switch can only be enabled when **Digital** is on and the primary
**Digital / decoder no.** is filled. The DT decoder number alone does not make the vehicle eligible.

Enabling opens an unlocked exhibition-list picker and rejects a duplicate vehicle for the same
owner/name or a duplicate decoder number in that list. The complete operation requires Admin, or a
user who combines Editor and Messe permissions. Disabling requires write permission and resets the
vehicle flag, but it does not delete an existing exhibition-list entry.

## Delete a vehicle

Admin and Editor users can select **Delete** and confirm the vehicle identified by inventory number
and designation. There is no undo or typed confirmation in v0.1.20.3.

Deletion removes the vehicle and cascades through its vehicle-owned database records, including
inventory-number history, images, attachment metadata, maintenance, functions, CV data, external
mappings, and spare parts. Layout or accessory allocation references can block deletion. RailKeeper
writes a `VehicleDeleted` audit entry when deletion succeeds. Export a current backup before removing
important records; the application does not promise physical cleanup of every referenced file.

## Empty, loading, and error states

| Situation | Result and next step |
| --- | --- |
| Initial load | **Loading vehicles from local database...** appears until the request completes. |
| No vehicle exists | **No vehicles available yet.** Admin or Editor can create one or use Import/Export. |
| Search/filter has no match | **No vehicles found for this filter.** Clear filters or change the search. |
| List or detail fails | RailKeeper shows the server error above the inventory. Check the session and connection, then refresh. |
| Required field missing | The Model section opens and lists Manufacturer, Designation, Gauge, Category, or Subtype as missing. |
| Automatic number fails | Ask an administrator to configure an active inventory-number scheme for the category. |
| Manual number conflicts | Choose a different unique inventory number. |
| Save/delete is forbidden | Sign in with Admin or Editor. Visible controls do not override the server role check. |
| Save fails | The error appears above the inventory and the editor remains open. Treat the change as unsaved, check the session and connection, then retry. |
| Delete fails | The error appears above the inventory and the confirmation remains open. Treat the vehicle as not deleted, resolve the reported cause, then retry. |
| Report creation fails | The error appears above the inventory and the report dialog remains open. Check the connection and popup handling, then create it again. |
| Report has no rows | Adjust search/filter or select at least one visible row. |
| Delete is blocked | Remove or change the layout/accessory allocation that still references the vehicle, then retry. |

## Related pages

- [User Guide overview](/guide/)
- [Vehicle images and attachments](/guide/vehicles/media)
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)
- [Decoder, functions, and CV data](/guide/vehicles/decoder-cv)
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)
- [Exhibition workspace](/guide/exhibition/)
- [First setup and sign-in](/guide/getting-started/)
- [Overview, metrics, and data quality](/guide/overview/)
- [Administration overview](/administration/)

## Documented RailKeeper version

This page documents the vehicle workflow shipped with RailKeeper **v0.1.20.3** and was last reviewed on 2026-08-17.
