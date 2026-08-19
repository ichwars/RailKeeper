---
title: Export inventory
description: Create CSV, RailKeeper JSON, and printable vehicle inventory output.
audience: user
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Export inventory

The export actions use the complete vehicle list loaded by **Import/Export**. They do not inherit
searches or filters from the vehicle-inventory page. All three controls remain disabled while the
list is loading or when no vehicles exist.

Admin, Editor, Viewer, and Planner can export. Export does not modify RailKeeper data.

## CSV export

**CSV export** downloads `railkeeper-bestand.csv` with these properties:

- UTF-8 text with a byte-order mark for spreadsheet compatibility;
- semicolon-separated columns;
- the 62 scalar targets listed in [Import vehicle files](/guide/import-export/file-import);
- column labels in the currently selected RailKeeper interface language;
- localized yes/no text for boolean fields;
- double-quote escaping for semicolons, quotes, and line breaks.

This is the broadest stable vehicle-field round trip. A CSV created in one interface language is
recognized through the built-in German and English header aliases. Still review its mapping and
rows before writing it into another installation.

The CSV excludes images, attachment bytes, maintenance, spare parts, function records, CV records,
decoder files, and external mappings.

## RailKeeper JSON export

**JSON export** downloads `railkeeper-bestand.json`. Its top-level shape is:

```json
{
  "format": "railkeeper-vehicles",
  "version": 1,
  "vehicles": []
}
```

The `vehicles` array contains the vehicle objects returned to this page. This makes JSON useful for
inspection and lower-loss data exchange, but it is not an application backup or a complete restore
format.

The stable v0.1.19.2 JSON importer reads only 14 fields from each object. It does not restore all
properties present in the exported object, nested records, or stored file bytes. See the exact
[JSON import subset](/guide/import-export/file-import#json-import-subset) before relying on a
round trip.

## PDF and print view

**PDF/Print view** opens a new browser window and immediately starts the browser print dialog. The
document uses A4 landscape and contains:

- a total, digital, and analog vehicle count;
- generation date and time;
- Inventory number, Manufacturer, Article no., Designation, Gauge, Epoch, Category,
  Digital/analog, and List price for every vehicle.

RailKeeper does not generate a PDF file on the server. Choose **Save as PDF** in the browser print
dialog when the browser or operating system provides that destination.

If nothing opens, allow popups for the RailKeeper address. The print window escapes vehicle text
before inserting it into the report.

## Choosing an export

| Need | Recommended output |
| --- | --- |
| Inspect or edit supported scalar fields in a spreadsheet | CSV |
| Inspect the raw vehicle objects returned to the page | JSON |
| Hand out or archive a compact human-readable inventory list | PDF/Print view |
| Restore application data, uploads, and related records | Admin application backup, not these exports |

Keep exported files according to the sensitivity of the inventory. Article sources, purchase data,
storage locations, values, notes, and other private collection details can be present in CSV or
JSON.

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19.2** and was last reviewed on 2026-08-16.
