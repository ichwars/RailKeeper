---
title: Install RailKeeper
description: Install RailKeeper.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Install RailKeeper

Choose where your collection will live: a Windows PC or your own server. RailKeeper needs no cloud service. A browser connects to the running application.

## Windows: no additional installation

1. Open [GitHub Releases](https://github.com/ichwars/RailKeeper/releases) and download the Windows x64 ZIP for your chosen version.
2. Extract the entire ZIP into its own application folder.
3. Start `start-railkeeper.bat`.
4. Create your administrator on first access. There is no default password.
5. Check the active data folder under **Settings > General > Storage usage**.

The default is `%LOCALAPPDATA%\RailKeeper\data`, outside the replaceable application folder. An old `data` folder next to the EXE is migrated according to the [update safety rules](system-operations). Do not combine databases.

## Docker: on your computer or server

Git and Docker Compose are required. The [Compose file](https://github.com/ichwars/RailKeeper/blob/main/docker-compose.yml) initially binds the service to the local computer only.

```sh
git clone https://github.com/ichwars/RailKeeper.git
cd RailKeeper
docker compose pull
docker compose up -d
```

Open `http://localhost:8080` on that computer and create your administrator. Persistent data lives in the `railkeeper_data` volume, mounted at `/data` inside the container. Retain that volume when replacing the application.

## Access from other devices

The operator must configure network or reverse-proxy access explicitly. Use HTTPS and the matching public URL for network access. Secure cookies require HTTPS; an intentionally local HTTP test needs matching cookie settings. Keep authentication and role checks enabled when troubleshooting connectivity.

| Setting | Purpose |
| --- | --- |
| `RAILKEEPER_DATA_DIR` | Persistent data folder, outside replaceable application files |
| `RAILKEEPER_PUBLIC_URL` | User-facing address, also used for password-reset links |
| `RAILKEEPER_COOKIE_SECURE` | `true` for HTTPS; must match the actual connection |
| `RAILKEEPER_TRUSTED_PROXY_CIDRS` | Only your own trusted proxy networks |

The full configuration for TLS, upload limits, SMTP, OCR and printers is in the [production runbook](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md). Next: [your first vehicle](../guide/quick-start) and [backups](backup-restore).
