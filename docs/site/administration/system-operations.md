---
title: Updates and operation
description: Updates and operation.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Updates and operation

**Goal:** Change versions and verify that inventory and files remain complete.
Use a separate test installation for your first exercise.

## 1. Record version and data location

Open **Settings > General > Updates** and record the current version.
A published release may differ from the `main` development branch.
An offered download does not install the update automatically.

<GuideFigure name="update-version" alt="Update settings with the current RailKeeper version and manual check." caption="Actual browser UI from the local demo. This picture does not show a Windows or Docker update being performed." />

On the same settings page, open **Storage usage**. As Admin, record the active data path and
storage mode. Also check the actual startup or Compose configuration.

<GuideFigure name="update-storage" alt="Storage usage with the separate demo's active data path and storage mode." caption="This path belongs to the synthetic test installation. Use the actual paths and volumes shown in your Windows or Docker environment." />

## 2. Back up before replacement

1. Export an application backup and keep it outside the installation.
2. Validate the file without restoring it. A [restore exercise](restore-exercise) on a second
   installation helps establish the recovery procedure.
3. Also take a consistent copy of the complete data directory or volume, including local
   authentication and configuration. Stop the application or use a suitable consistent snapshot.
4. Record startup options, public URL, port, image tag or ZIP version.
5. Read target release notes and plan downtime without simultaneous edits.

## 3. Update Windows

| Step | Action | Check |
| --- | --- | --- |
| Choose a package | Download the matching Windows-x64 ZIP from GitHub Releases | Intended version and architecture |
| Stop the application | Shut down RailKeeper, not just its browser tab | Old process is no longer running |
| Extract separately | Extract the entire ZIP into a new application directory | Startup file and bundled subdirectories are complete |
| Start | Run `start-railkeeper.bat` in the new directory | Browser shows the new version |
| Check data | Verify the same data path and storage mode | No accidental empty inventory |

The default data directory is `%LOCALAPPDATA%\RailKeeper\data`; explicit configuration or portable
mode can differ. Keep the old application directory and backups.
If old and new databases both exist, assess them separately. Do not merge their directories.

## 4. Update Docker

Work in the existing Compose configuration directory. Retain the data volume and environment.
Check the image tag or digest before replacement:

```sh
docker compose config --images
docker compose pull
docker compose up -d
docker compose ps
docker compose logs --tail=100
```

Check configuration and log output for private information before sharing it.
The automatic pre-migration safety copy resides inside the data directory and is not an external backup.

## 5. Recheck the same checkpoints

| Check | Expected result |
| --- | --- |
| Version and `/health` | Intended version and successful health response |
| Authentication | Your existing local account works |
| Inventory and set | Known vehicles and members remain present |
| Picture and attachment | Actually open each, rather than counting thumbnails |
| CVs and functions | Readings, index pages, provenance and manual choices retained |
| Maintenance and master data | Known records remain present |
| New backup | Successful export and validation |
| Station, if used | Reachable from the RailKeeper server; start with a connection test |

For failures, record version, error and affected step. Check data path and logs.
An older application is not automatically compatible with a migrated database.
For rollback, restore the matching application, data and configuration snapshot together.

**Evidence boundary:** The UI and restore exercise were tested in the local demo. Instructions
were checked against startup/migration code and the runbook; these pictures are not Windows or Docker update acceptance.
[Detailed runbook](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md).
