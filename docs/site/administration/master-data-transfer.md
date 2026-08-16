---
title: Master-data transfer
description: Export and reconcile RailKeeper master data with the versioned JSON document.
audience: admin
status: stable
reviewedVersion: 0.1.19
lastReviewed: 2026-08-16
---

# Master-data transfer

The **Master-data transfer** box under **Settings > Import/Export** moves controlled entries and
their relations between RailKeeper instances. Both download and upload are Admin-only operations.
This is a specialized configuration transfer, not an application backup and not the vehicle-file
Import/Export workspace.

## What the document contains

The downloaded file is named `railkeeper-stammdaten-YYYYMMDD-HHMMSS.json`. Stable v0.1.19 exports
format `railkeeper-master-data`, version 2, with:

- the creation timestamp;
- every active and inactive general and article master-data entry;
- keys, labels, source URLs, metadata, sort order, origin, and timestamps;
- configured master-data relations, such as vehicle category-to-type mappings;
- embedded function-symbol images because they are entry metadata.

It does not contain vehicles, accessory articles, stock, storage locations, inventory-number
schemes, users, roles, sessions, or authentication data. Use application backup and restore for a
recoverable inventory and operational-data transfer. Local authentication remains separate there
as well.

## Export master data

1. Open **Settings > Import/Export** as an Admin.
2. In **Master-data transfer**, select the download action.
3. Store the JSON with the related RailKeeper version and a short change note.
4. Keep a current application backup separately.

Export is read-only. It includes custom edits and inactive states present at download time.

## Understand import reconciliation

Import is a full desired-state reconciliation, not an additive merge. RailKeeper accepts document
versions 1 and 2 and validates the complete document before replacing entries and relations in one
database transaction.

The stable rules are:

- A matching existing bundled identity remains **Bundled**, regardless of the imported origin.
- An unknown identity becomes **Custom**, even if the document claims it is bundled.
- Bundled entries omitted from the document are retained with their current state.
- Unused custom entries omitted from the document are removed.
- If an omitted custom entry is referenced, the entire import is rejected before mutation.
- Standard article-type keys remain protected and must form a valid complete configuration.
- If older documents omit article types or accessory subtypes, the current protected data is
  preserved according to the compatibility rules.
- Relations must be unique and point to entries that exist after reconciliation.
- An omitted current relation is retained only when both endpoints are bundled; other omitted
  relations are removed.
- Custom-field definitions and their stored article references must remain valid.

Only after every check succeeds does RailKeeper replace the master-data tables and refresh the
runtime cache. A failed validation or database operation leaves the previous committed state
unchanged.

## Import safely

1. Create and validate a current application backup.
2. Export the current master data as a separate rollback reference.
3. Compare the incoming file with the current export, especially omitted custom entries, inactive
   states, article types, subtypes, and relations.
4. Select the JSON file in **Master-data transfer**.
5. Select upload and confirm the destructive reconciliation.
6. Wait for the reported imported-entry and imported-relation counts.
7. Reload **Settings > Data** and verify representative general and article data.
8. Create or edit one affected vehicle and accessory article before accepting the transfer.

The stable UI has no dry-run preview. The upload limit is 25 MiB. A malformed file, unsupported
version, duplicate identity, broken relation, protected article-type change, or referenced omitted
custom entry rejects the import.

## Recover from a rejected or incorrect transfer

For a rejection, leave the instance running, read the error, correct the source document or its
origin instance, export again, and retry. Do not remove referenced values merely to make an import
pass.

If a technically valid import produced the wrong desired state, import the previously exported
master-data file. If inventory context, number schemes, locations, or other application data also
need recovery, use the validated application backup instead.

## Related pages

- [Master-data administration](./master-data)
- [General master data](./master-data-general)
- [Article master data and storage locations](./master-data-articles)
- [Import and export workspace](/guide/import-export/)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.19** and was last reviewed on 2026-08-16.
