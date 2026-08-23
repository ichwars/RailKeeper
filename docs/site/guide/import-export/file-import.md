---
title: Import vehicle files
description: Map, validate, review, and save vehicle rows from CSV, TSV, XML, or JSON.
audience: user
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# Import vehicle files

File import is a review workflow, not a direct database load. RailKeeper reads the selected file,
maps its columns, validates each row against the vehicles already loaded, and lets an authorized
user choose which rows to create or update.

Use an Editor or Admin account to save. Viewer and Planner can inspect the preview but the server
rejects their vehicle writes.

## Supported input formats

| Extension | Stable interpretation |
| --- | --- |
| `.csv` | Uses the most frequent delimiter in the first non-empty line: tab, semicolon, or comma. A tie between semicolon and comma prefers semicolon. |
| `.tsv` | Always uses a tab delimiter. |
| `.xml` | Reads elements named `vehicle`, `fahrzeug`, `locomotive`, or `lok`. If none exist, RailKeeper uses the most frequently repeated nested element as the record shape. Attributes and direct text children become columns. |
| `.json` | Accepts a vehicle array or an object with a `vehicles` array. The stable importer copies only the JSON subset listed below. |

Delimited values may be quoted with double quotes. Doubled quotes inside a quoted field are
unescaped. Leading and trailing whitespace is removed from every parsed cell, and completely empty
rows are ignored.

An invalid JSON or XML document is not a supported import. XML without usable records reports an
empty table. Choose the corrected file again after fixing it.

## JSON import subset

The JSON export wraps the vehicle list as `{ "format": "railkeeper-vehicles", "version": 1,
"vehicles": [...] }`. The stable JSON importer does not validate `format` or `version` and does not
round-trip every property. It builds its review table from these 14 fields only:

| Identity | Classification | Operation and value |
| --- | --- | --- |
| Inventory number, manufacturer, article number, designation | Gauge, epoch, railway company, category, subtype | Maximum speed, home depot / operating location, digital, digital / decoder number, list price |

Use CSV when you need the complete supported 62-field scalar exchange. Use an application backup
when you need restorable application data and uploads.

## Column mapping

The first row supplies the column names. RailKeeper normalizes case, whitespace, German umlauts,
common punctuation, and known German or English aliases before looking for a target field.

After loading a file:

1. Under **Detected CSV mapping**, review every source header.
2. Assign every useful open column to a RailKeeper field.
3. Explicitly set an irrelevant source column to **Ignore column**.
4. For recurring files, optionally select **Save mapping to profile**.
5. Choose **Validate mapping**, then check the recalculated import review.

One target field can belong to only one source column. Targets already in use are unavailable for
other columns. Mapping validation rereads the same file on the server and replaces the persistent
preview, so review conflicts and corrections again afterwards.

## All 62 CSV and table targets

The complete stable CSV export uses the same targets and can therefore be imported without manual
remapping. Other files may use any subset.

| Group | Supported fields |
| --- | --- |
| Identity and catalogue | Inventory number; Manufacturer; Article no.; Source / URL; Designation; Gauge; Epoch; Railway company; Category; Subtype; Description; Series; Vehicle no.; EAN; Production period; List price |
| Operation, digital, and exhibition | Maximum speed; Home depot / operating location; Digital; Digital / decoder no.; Decoder type; DT / decoder; DT / decoder no.; Exhibition ready; Exhibition; ABC brakes; Create QR code |
| Acquisition, storage, and condition | Acquisition type; Acquired from; Purchase price; Purchase date; Storage location; Storage details; Condition; Condition details; Packaging |
| Physical and mechanical data | Length (mm); Weight (g); Color; Lettering; Load; Interior; Axles; Axle count; Traction tire count; Wheelset; Coupling (F=R); Front coupling; Rear coupling; Power pickup; Adapter / interface |
| Equipment and notes | Drive; Drive description; Headlights; Headlight description; Lighting; Lighting description; Sound generator; Sound generator description; Smoke generator; Smoke generator description; Additional information |

Images and other nested or file-backed records are not targets.

## Boolean values

For Digital, DT / decoder, Exhibition ready, Exhibition, ABC brakes, Coupling (F=R), Drive,
Headlights, Lighting, Sound generator, Smoke generator, and Create QR code, these non-empty values
mean **yes** regardless of case:

`1`, `ja`, `yes`, `true`, `wahr`, `digital`, `d`, `x`, `vorhanden`

Every other non-empty value becomes **no**. An empty cell is ignored. Prefer the explicit `yes` and
`no` values created by RailKeeper's CSV export to avoid ambiguous source data.

## Validation and initial selection

A new row needs Manufacturer, Designation, Gauge, Category, and Subtype before the preview allows
selection. The server may apply additional vehicle rules when saving.

Two field-specific checks run in the browser:

- Maximum speed must be an integer from 1 through 1000 km/h.
- Home depot / operating location must contain at most 200 Unicode characters.

An invalid row is marked as an error, is deselected, and cannot be selected until the visible
problem is corrected or its source mapping is changed. The review table directly edits only
Inventory number, Manufacturer, Article no., Designation, Gauge, Category, and Subtype. Correct
other values in the source file or through the column mapping, then reload when necessary.

## New vehicles and detected duplicates

RailKeeper compares a non-empty inventory number case-insensitively with the current vehicle list.

| Result | Default behavior |
| --- | --- |
| No inventory-number match and no validation issue | Action **New**, row selected |
| Existing inventory number | Action **Update**, warning, row not selected |
| Existing inventory number plus another validation issue | Action **Update**, error, row blocked |

For an update, expand the field preview and inspect **same**, **fills empty field**, and
**overwrites**. Only mapped, non-empty imported strings, valid numbers, and parsed booleans are
applied. Empty source cells do not erase existing strings. A non-empty boolean token that is not in
the yes list explicitly applies **no**.

Changing a duplicate row to **New** does not create a second vehicle with the same inventory
number. RailKeeper marks the row as an error instead.

## Save the reviewed selection

1. Correct the visible core fields and review every update comparison.
2. Select only rows that should be written.
3. Choose **Save selection**.
4. Wait until each successful row shows **saved**.
5. Read any row-specific error and verify already saved vehicles before retrying.

Rows are processed sequentially. Each selected row calls either vehicle creation or vehicle update.
If one row fails, it becomes an error and RailKeeper continues with later selected rows. The
operation is therefore not atomic and there is no bulk undo.

File imports do not create function mappings, CV entries, ECoS mappings, images, attachments,
maintenance, spare parts, or decoder files. Add those through their dedicated vehicle workflows.

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.20** and was last reviewed on 2026-08-16.
