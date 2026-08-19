---
title: Installation and Administration
description: Install, configure, secure, and operate RailKeeper.
audience: admin
status: stable
reviewedVersion: 0.1.19.2
lastReviewed: 2026-08-16
---

# Installation and Administration

This section covers Windows Standalone and Docker installation, runtime configuration, users and
roles, SMTP, backups and restores, updates, TLS, uploads, OCR, printers, operational checks, and
conservative troubleshooting.

Administration guidance describes the stable v0.1.19.2 runtime and preserves RailKeeper's
local-first, self-hosted security model.

## Administrative workflows

- [Master-data administration](./master-data) covers controlled values, lifecycle, storage
  locations, inventory-number schemes, and the JSON transfer.

## Safe Windows Standalone updates

The Windows Standalone ZIP contains the application only. It never contains a database, uploads,
attachments, thumbnails, or backups. By default, persistent data is stored independently from the
replaceable program folder in `%LOCALAPPDATA%\RailKeeper\data`.

After a successful check under **Settings > General > Updates**, Windows Standalone shows
**Download version X** only when the matching ZIP belongs to the detected GitHub release. The
browser downloads that ZIP directly. RailKeeper does not extract, install, replace, migrate, or
restart anything. If the trusted package is unavailable, the GitHub release page remains linked.

For an update:

1. Download the matching ZIP from the versioned button or the linked release page.
2. Create a current application backup and close RailKeeper.
3. Extract the ZIP into a new program folder. Do not copy a database into that folder.
4. Start the new `RailKeeper.exe`.
5. Sign in, check the active storage location under **Settings > Data storage**, and verify the
   inventory before deleting the previous program folder.

Older Standalone versions stored data in a `data` folder beside `RailKeeper.exe`. On first startup,
RailKeeper copies that legacy data to the safe location and keeps the source unchanged. If both
locations already contain a database, RailKeeper stops before opening or migrating either database
and displays both paths. Close RailKeeper, create separate copies of both folders, and decide which
database is current. Rename the existing safe folder instead of overwriting or merging it, then
copy the chosen complete data folder to `%LOCALAPPDATA%\RailKeeper\data`. Keep both source copies
until the inventory and attachments have been verified.

An explicitly configured `RAILKEEPER_DATA_DIR` always takes precedence and disables automatic
legacy migration. Do not point it at a replaceable program folder or an unreliable removable
drive. Administrators can see the exact active path in **Settings > Data storage**. Opening that
folder in Explorer is available only for a local Windows Standalone instance.

Before pending database migrations, RailKeeper creates a validated private copy under
`safety-backups`. It includes the complete database, including local authentication data. This is a
startup safeguard, not a replacement for an application backup or a filesystem/volume backup.

## Docker updates

Docker keeps persistent data in the mounted `/data` volume. Update the application with:

```powershell
docker compose pull
docker compose up -d
```

Back up the volume before an update and verify `/health`, login, inventory, and backup validation
afterwards. The automatic pre-migration database copy remains inside `/data` and therefore does not
protect against loss of the volume itself.

## Master-data lifecycle

RailKeeper distinguishes between entries shipped with the application and custom entries created
by an administrator or editor. Shipped entries can be edited and deactivated, but cannot be
permanently deleted. An unused custom entry can be permanently deleted. Once a custom entry is in
use, it can only be deactivated.

Deactivated entries are no longer offered for new vehicles, accessories, or other records. Existing
records keep their saved value and continue to display it as inactive, so deactivation does not
invalidate historical inventory data.

Edits and deactivation states survive application restarts, seed reconciliation, updates,
master-data export and import, and application backup and restore. A current RailKeeper backup is
still recommended before extensive master-data changes.

The complete field, permission, numbering, location, and import rules are documented in
[Master-data administration](./master-data).
