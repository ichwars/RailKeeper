---
title: Import and export
description: Exchange vehicle, accessory, and exhibition data through reviewed transfer jobs.
audience: user
status: stable
reviewedVersion: 0.1.20
lastReviewed: 2026-08-16
---

# Import and export

The **Import/Export** workspace exchanges vehicle, accessory, and exhibition data without bypassing
RailKeeper's validation and permissions. Persistent profiles define direction, data areas, format,
and options. Each execution creates an auditable job with preview, issues, decisions, result, and,
for exports, a downloadable artifact.

Imports use either one-area CSV or versioned RailKeeper JSON. They remain in review until all
blocking issues are resolved and an authorized user explicitly confirms the transactional apply.
Command-station synchronization, master-data transfer, and application backup and restore remain
separate workflows.

## Access rights

Admin, Editor, Viewer, Planner, and Messe can open **Import/Export**. Pure Messe accounts see and
run only exhibition-list profiles and jobs. The server applies that scope independently of the UI.

Admin and Editor can create or update full transfer profiles and apply imports. Viewer and Planner
can inspect jobs and create exports but cannot apply imported data. Messe can export, review, and
apply exhibition-list transfers within its isolated scope. Only Admin can disable profiles, delete
artifacts, or request that RailKeeper open an artifact folder on the server.

## Choose the correct workflow

| Goal | Workflow |
| --- | --- |
| Transfer one data area as CSV | Create an import or export profile for vehicles or accessories |
| Transfer one or more data areas with relationships | Create a versioned RailKeeper JSON profile |
| Review an import | Open its job, inspect issues, choose resolutions, then confirm the apply |
| Read, compare, or write locomotive data | Use the separately protected **Digital centers** workspace |
| Transfer manufacturers, gauges, categories, symbols, or other master data | Use **Settings > Master data**, not this workspace |
| Preserve or restore application data and stored uploads | Use the Admin-only application backup, not a vehicle export |

## Understand transfer profiles

A transfer profile stores only the reusable selection for a job:

- **Direction**: import reads a file, export creates a file.
- **Areas**: vehicles, accessories, or exhibition lists. CSV supports exactly one area, and
  exhibition lists require RailKeeper JSON.
- **Format**: CSV for tabular single-area exchange, RailKeeper JSON for complete packages that may
  span areas.
- **CSV mapping**: vehicle imports can save source-column assignments to the 62 RailKeeper fields
  as profile defaults.

The **Transfer profiles** table lists import, export, and disabled profiles together. Only enabled
profiles have a start action. An enabled profile name may occur only once per direction, while the
same name may be used once for import and once for export.

## Safe working sequence

1. Ask an Admin to create a current application backup before a large or unfamiliar import.
2. Select an enabled profile whose direction, areas, and format match the intended transfer.
3. For an export, create the job, execute it, and download the artifact after checking its summary.
4. For an import, create the job and upload the matching CSV or RailKeeper JSON file.
5. For CSV, review every detected source column. Map open columns or explicitly ignore them, and
   optionally save the mapping to the profile for recurring files.
6. Inspect the persisted preview and every warning or error. Record a resolution for each issue that
   requires a decision.
7. Confirm only the reviewed snapshot, then inspect the completed job and representative records.

## Important boundaries

- Transfer files are untrusted input. RailKeeper validates package version, hashes, paths, record
  identities, references, controlled values, and role scope before apply.
- CSV is limited to exactly one supported area and cannot carry exhibition lists. Versioned JSON
  can contain several supported areas and their relationships.
- Import preview and issue decisions are persisted on the server. Confirmation applies the reviewed
  snapshot transactionally; a failed apply does not leave a partially imported job.
- Transfer packages never contain users, roles, sessions, password hashes, or other local
  authentication state. Use application backup for full local inventory and upload recovery.
- Digital-center reads and writes never enter a generic transfer job automatically.

## Troubleshooting

| Symptom | Check |
| --- | --- |
| No suitable profile is available | Ask an Editor or Admin to create or enable a profile for the required direction, areas, and format. |
| CSV cannot be selected | Use exactly one vehicle or accessory area, or switch to RailKeeper JSON. Exhibition lists require JSON. |
| A CSV column remains open | Select a RailKeeper field or explicitly ignore the column. The mapping must be complete before validation. |
| An import remains in review | Open the job details and resolve every blocking issue before confirming. |
| Confirmation reports stale state | Reload the job. Another operator changed its revision, state, or issue decisions. |
| A pure Messe account cannot see a profile | Only exhibition-list profiles are visible in the isolated Messe scope. |
| An artifact cannot be downloaded | Reload its job and check whether an Admin deleted the artifact after creation. |
| A failed job offers retry | Review the recorded error and source first. Retry creates a controlled continuation, not an untracked duplicate. |

## Documented RailKeeper version

This chapter documents stable RailKeeper **v0.1.20** and was last reviewed on 2026-08-16.
