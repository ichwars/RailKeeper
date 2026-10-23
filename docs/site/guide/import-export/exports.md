---
title: Export your inventory
description: Export your inventory.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Export your inventory

**Goal:** Create a reviewable inventory file for a spreadsheet or another RailKeeper installation. Admin, Editor and Viewer may export. Exhibition-only accounts are limited to exhibition lists.


<GuideFigure name="import-export" alt="Export profiles determine the data area and file format." caption="Export profiles determine the data area and file format. Synthetic demo data, not factory settings. Click for full size." />

## Create an export

1. Open **Import/Export > Create profile** and select **Export**.
2. Choose the required area and format. CSV supports exactly **Vehicles** or **Accessories**. Exhibition lists and multiple areas require **JSON**.
3. Select **New export** and the profile, or start from the enabled profile.
4. Check areas and format, then select **Run export**.
5. Use **Download file** to save the result.

The job and generated file remain traceable in transfer history. An administrator may later delete files. The workspace storage location is on the RailKeeper server, not necessarily on your browser computer.

## Which format fits?

| Purpose | Format |
| --- | --- |
| Review vehicle or accessory fields in a spreadsheet | CSV |
| Transfer several areas and their relationships | RailKeeper JSON |
| Transfer sets with individual members | CSV with set columns or RailKeeper JSON |
| Print a readable list or save PDF | Report in vehicle inventory or exhibition |
| Restore inventory and media | [Application backup](../../administration/backup-restore) |

A search or selection in the vehicle inventory does not automatically become a transfer job selection. Check the scope in the export dialog.

## Limits

An export package is not a complete backup. Vehicle media, CV records, functions, maintenance and local authentication are not part of vehicle transfer. The server versions the JSON package structure and validates it on [import](file-import); do not manually change identifiers, package versions or checksums.

Treat exported files as private collection data. They can contain prices, storage locations and notes.
