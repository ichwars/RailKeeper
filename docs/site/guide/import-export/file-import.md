---
title: Import vehicle files
description: Import vehicle files.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Import vehicle files

**Goal:** Review an inventory file before deliberately importing it. Vehicle imports require Admin or Editor. Create a [backup](../../administration/backup-restore) before a large or unfamiliar import.


Try the [CSV and JSON examples with a guided walkthrough](examples).


<GuideFigure name="import-dialog" alt="Start a new import and choose a matching profile." caption="Start a new import and choose a matching profile. Synthetic demo data, not factory settings. Click for full size." />

## 1. Create a matching profile

Open **Import/Export > Create profile**. Give it a useful name such as “Vehicles from spreadsheet”. Select **Import**, the **Vehicles** area and **CSV** or **JSON**.

| Source | Matching format |
| --- | --- |
| Vehicle table with column headers | CSV, exactly one area |
| Package from the current RailKeeper export | RailKeeper JSON, including multiple areas and vehicle sets |
| LokProgrammer CV list or decoder project | In the vehicle under **CV > Files & import**, not here |
| ECoS or CS3 locomotive list | In **Command stations**, not as a file transfer |

The current dialog accepts CSV or versioned RailKeeper JSON packages. An arbitrary XML, TSV or JSON file is not such a package. Convert external spreadsheets to CSV first. The previous instructions for direct XML/TSV import and a limited 14-field JSON import no longer describe this workflow.

## 2. Choose the file and map columns

1. Select **New import** and your import profile.
2. Use **Browse** to select the file. This creates the preview; it does not import vehicles yet.
3. For a vehicle CSV, inspect **Detected CSV mapping**.
4. Assign every source column to a target field or explicitly choose **Ignore column**.
5. Select **Validate mapping** or **Continue to review**.

A target field can be mapped only once. You may save a recurring mapping to the profile. Changing the mapping recalculates the preview, so review its conflicts again.

## 3. Review the preview and conflicts

Check counts, names and inventory numbers. Read every warning. Existing records or conflicting references offer appropriate decisions, such as use existing data, replace, create a copy or skip. Not every decision is available for every conflict.

Detected vehicle sets appear as groups. Review their inventory numbers, members and order together. A set decision applies to the entire group; vehicles are not silently removed from other sets.

Correct invalid source data in the file and upload it again. Resolve every blocking conflict before confirming.

## 4. Confirm the import

Confirm only the fully reviewed preview. The server applies the reviewed revision in one transaction. A failed confirmation leaves no partial import. Check the completed job and a few vehicles in the inventory afterward.

If the job changed meanwhile, RailKeeper may require another review or upload. Confirm only the new preview. **Close** leaves the dialog; **Cancel job** ends the job and retains its history.

## What is transferred?

Vehicle files transfer basic fields and documented set relationships. They do not include vehicle pictures, CV rows, functions, maintenance or decoder files. Use the relevant vehicle workflows or an application backup for those. The [transfer overview](./) explains permissions and limits.
