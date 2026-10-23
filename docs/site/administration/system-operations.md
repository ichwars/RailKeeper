---
title: Updates and operation
description: Updates and operation.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Updates and operation

**Goal:** Update the application while retaining your inventory and files. Check the running version first. A GitHub release and the current `main` development branch may differ.

## Before every update

1. Create an [application backup](backup-restore) and keep it outside the installation.
2. As operator, also take a consistent copy of the entire data folder or volume. This private copy includes local authentication data.
3. Record version, configuration and active data path. Read the target version's notes.
4. Plan a short interruption and prevent concurrent changes.

## Update Windows

Download the matching new ZIP from Releases. RailKeeper may offer a matching download under **Settings > General > Updates**. It does not install it automatically.

Close RailKeeper, extract the ZIP into a new application folder and start the new application. Check the data path, sign-in, inventory and pictures. Retain the old application folder and backups until these checks succeed.

Legacy data migration retains its source. If both the old and new locations contain databases, startup stops. Back up both folders separately and establish which collection is current. Do not merge folders.

## Update Docker

Select a verified image tag or digest. In your Compose configuration directory, run:

```sh
docker compose pull
docker compose up -d
docker compose ps
```

Retain the existing data volume. The automatic safety copy before database migrations resides inside the data folder; it does not replace an external backup.

## Check after updating

- Does `/health` respond successfully?
- Do sign-in, the vehicle list and editing a test record work?
- Are pictures, CV provenance and important master data present?
- Can you create and validate a backup?
- Can the server reach the required command stations?

For failures, check logs, the data path and version first. Older application code is not automatically compatible with a migrated database. Restore the complete saved state when rolling back. [Detailed runbook](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md).
