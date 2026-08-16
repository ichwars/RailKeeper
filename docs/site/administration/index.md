---
title: Installation and Administration
description: Install, configure, secure, and operate RailKeeper.
audience: admin
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Installation and Administration

This section covers Windows Portable and Docker installation, runtime configuration, users and
roles, SMTP, backups and restores, updates, TLS, uploads, OCR, printers, operational checks, and
conservative troubleshooting.

Administration guidance describes the stable v0.1.17.6 runtime and preserves RailKeeper's
local-first, self-hosted security model.

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
