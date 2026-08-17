---
title: Import and export
description: Exchange vehicle inventory safely and understand the separate ECoS workflow.
audience: user
status: stable
reviewedVersion: 0.1.19.1
lastReviewed: 2026-08-16
---

# Import and export

The **Import/Export** workspace exchanges vehicle inventory without bypassing RailKeeper's normal
validation and permissions. File imports first become a review table. Exports create a CSV, a
RailKeeper JSON file, or a browser print view. An independently controlled ECoS workflow can read
locomotive data and hand one locomotive at a time to the vehicle editor.

This chapter documents stable RailKeeper v0.1.19.1. It covers vehicle exchange only. Master-data
transfer, command-station configuration, and application backup and restore remain separate
administrative tasks.

## Access rights

Admin, Editor, Viewer, and Planner can open **Import/Export**. A pure Messe account remains in the
isolated exhibition workspace and cannot open this page. The server still checks every read and
write independently.

| Action | Admin | Editor | Viewer | Planner |
| --- | --- | --- | --- | --- |
| Load the current vehicle list | Yes | Yes | Yes | Yes |
| Export CSV, JSON, or a print view | Yes | Yes | Yes | Yes |
| Parse a file and inspect its local preview | Yes | Yes | Yes | Yes |
| Create or update vehicles from the preview | Yes | Yes | No | No |
| Configure, read, or write an ECoS | Yes | No | No | No |

The page does not hide the **Save selection** button from Viewer or Planner. The protected vehicle
API nevertheless rejects their write. Use an Editor or Admin account for an actual file import.

## Choose the correct workflow

| Goal | Workflow |
| --- | --- |
| Review CSV, TSV, XML, or RailKeeper JSON and create or update vehicles | [Import vehicle files](/guide/import-export/file-import) |
| Download the current vehicle list or print a compact inventory report | [Export inventory](/guide/import-export/exports) |
| Read locomotives, static functions, and CV values from an ESU ECoS | [ECoS locomotive sync](/guide/import-export/ecos-sync) |
| Transfer manufacturers, gauges, categories, symbols, or other master data | Use **Settings > Master data**, not this workspace |
| Preserve or restore application data and stored uploads | Use the Admin-only application backup, not a vehicle export |

## Safe working sequence

1. Ask an Admin to create a current application backup before a large or unfamiliar import.
2. Export CSV as a readable comparison file when you want to inspect the current scalar vehicle
   fields in a spreadsheet.
3. Load the source file and resolve every required field, validation message, and duplicate.
4. Select only rows you have checked. Pay particular attention to fields marked as overwrites.
5. Save the selection, then inspect both saved and failed rows before leaving the page.
6. Open representative vehicles and confirm their core data after a large import.

## Important boundaries

- File exchange covers vehicle fields. It does not restore images, attachment bytes, maintenance,
  spare parts, decoder files, function mappings, CV records, or application users.
- The file preview is calculated in the browser. A row is written only when **Save selection** is
  used by an authorized account.
- Selected rows are saved one after another, not in one all-or-nothing transaction. A later error
  does not undo vehicles already marked **saved**.
- CSV is the broadest stable round-trip format for the 62 supported scalar fields. The JSON import
  intentionally reads a smaller field subset even when the JSON export contains more properties.
- ECoS data uses a separate work list and vehicle-editor handoff. It never enters the generic file
  import table automatically.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| Export buttons are disabled | Wait until the vehicle list has loaded. All export buttons remain disabled when the inventory is empty. |
| A column is shown as open | Assign its target in **Column mapping**, or leave it on **Ignore** when it is not needed. |
| A duplicate row is not selected | Review its update preview and select it explicitly only when the detected inventory-number match is correct. |
| Saving fails although the preview looked valid | Confirm the account is Admin or Editor, then read the row-specific server error. Server validation remains authoritative. |
| Some rows were saved before another failed | This is expected for the sequential import. Recheck saved rows and retry only the failed or still-open rows. |
| The print view does not open | Allow popups for the RailKeeper address and try again. |

## Documented RailKeeper version

This chapter documents stable RailKeeper **v0.1.19.1** and was last reviewed on 2026-08-16.
