---
title: Try file imports with example data
description: Try file imports with example data.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Try file imports with example data

**Goal:** Import two vehicles, map a column yourself and resolve an inventory-number conflict.
You need **Admin** or **Editor**. Use a test inventory: confirmation creates real records in
whichever installation you use.

All data is synthetic. These files contain no pictures, CVs or decoder projects.
Use the [separate CV example](../vehicles/cv-import-example) for decoder settings.

## Download the files

| File | Purpose |
| --- | --- |
| <GuideDownload file="vehicle-import-demo.csv">Vehicle CSV</GuideDownload> | Two new vehicles; map the extra `myNote` column manually. |
| <GuideDownload file="vehicle-import-conflict.csv">CSV with one conflict</GuideDownload> | After the first import: one existing inventory number and a new coach. |
| <GuideDownload file="vehicle-import-demo.json">RailKeeper vehicle JSON</GuideDownload> | Two additional vehicles in the versioned exchange format. |

Keep the original file extensions. The CSV is UTF-8 with semicolon separators.
The JSON file is a version 3 RailKeeper transfer package, not an application backup.

## 1. Create a CSV profile and preview

1. Open **Import/Export > Create profile**.
2. Name it **Demo CSV**, select **Import**, **Vehicles** and **CSV**, then create the profile.
3. Select **New import**, choose **Demo CSV** and open `vehicle-import-demo.csv`.
4. Wait for **Detected CSV mapping**. No vehicles have been imported yet.

<GuideFigure name="import-csv-mapping" alt="CSV mapping with the deliberately unmapped myNote column." caption="Known columns are detected. myNote is deliberately left open. Synthetic demo data; click for full size." />

## 2. Map a column yourself

Choose **Description** for `myNote`. Enable **Save mapping to profile** if you want to reuse it
for the second file. Select **Validate mapping**. Every column needs a target or an explicit
ignore decision.

Review these two new inventory numbers:

| Inventory number | Name |
| --- | --- |
| `DEMO-LOK-001` | Demo-Lok 01 |
| `DEMO-WAG-001` | Demo-Wagen 01 |

<GuideFigure name="import-csv-review" alt="Reviewed CSV containing two new vehicles before import." caption="Two new records are ready. Only the import button applies them. Synthetic demo data." />

## 3. Confirm and check the result

Select **Import 2 records**. The dialog closes after successful confirmation.
Open the completed job, then check both vehicles in inventory. Their description reads
**Nur Übungsdaten / Synthetic example**, as assigned from the source column.

<GuideFigure name="import-csv-result" alt="Job details after a successful CSV import." caption="The completed job shows the result; also check the vehicles in inventory. Synthetic demo data." />

## 4. Deliberately skip a conflict

1. Start another **Demo CSV** import using `vehicle-import-conflict.csv`.
2. Review column mapping again. If you did not save it to the profile, map `myNote` to
   **Description** again. Continue to review.
3. `DEMO-LOK-001` has a conflict because that inventory number already exists.
   The file now names it **Demo-Lok 01 Update**.
4. For this row, choose **Do not import this record**.
5. Check the new coach `DEMO-WAG-002` and select **Import 1 record**.

<GuideFigure name="import-csv-conflict" alt="An inventory-number conflict before a resolution is selected." caption="An existing inventory number needs a decision. A second row can create a new record. Synthetic demo data." />

**Expected result:** The inventory contains three CSV demo vehicles. The original locomotive
still has the name **Demo-Lok 01**. No additional locomotive shares its inventory number.
Replacing or importing a copy are different choices with different results.

## 5. Try the JSON package

Create another import profile, **Demo JSON**, with **Vehicles** and **JSON**.
Under **New import**, choose that profile and `vehicle-import-demo.json`.
No CSV mapping is required. Review the two records and confirm their import.

<GuideFigure name="import-json-review" alt="Preview of the versioned JSON example package with two vehicles." caption="RailKeeper JSON is validated as a package. Its inventory numbers differ from the CSV example. Synthetic demo data." />

The inventory now also contains `DEMO-JSON-001` and `DEMO-JSON-002`.
Running all steps once in an initially empty demo inventory creates **five demo vehicles**.
Repeating the exercise with existing demo inventory numbers produces conflicts again.

[Import rules and troubleshooting](file-import) · [Exports versus backups](exports)
