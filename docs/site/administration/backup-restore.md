---
title: Backup and restore
description: Backup and restore.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Backup and restore

**Goal:** Protect your collection from data loss. This page requires the **Admin** role.


<GuideFigure name="backup" alt="A selected backup is validated before restoring becomes possible." caption="A selected backup is validated before restoring becomes possible. Synthetic demo data, not factory settings. Click for full size." />

## Create an application backup

1. Open **Settings > Backup**.
2. Select **Create backup** in the matching section.
3. Keep the downloaded JSON file in a second storage location.
4. Before updating, check that a current backup exists.

Application backups include inventory, accessories, maintenance, CVs and provenance, master data, exhibition data, and saved media. They exclude user accounts, passwords, sessions, and local system settings. They are not full server backups.

## Check a backup first

1. Select the backup file under **Restore backup**.
2. Wait for validation. Check its date, format, row count, warnings, and errors.
3. If you only wanted to check the file, stop here. Selecting a file alone does not restore anything.

## Deliberately restore data

::: warning Back up the current state first
Restore replaces local application data and saved media with the file's contents. Save the current inventory first. Accounts in the file do not replace the installation's local sign-in data.
:::

1. Select a compatible file and read its validation result.
2. Type **RESTORE** in the confirmation field.
3. Select **Restore backup** and confirm the prompt.
4. Afterwards, check vehicles, CVs, pictures, and important master data.

## Which backup do I need?

| File or copy | Purpose |
| --- | --- |
| inventory export | exchanging tables and records, not a complete recovery state |
| application backup | moving or restoring app data between compatible installations |
| full filesystem/volume backup | server failure recovery and returning to an earlier installation, including private local data |

A copy inside the same container cannot protect against losing that container. Operators back up the database and file storage consistently and keep a copy off the system. [Updates and operation](system-operations).

[Illustrated restore exercise on a separate test installation](restore-exercise).
