# Deployment

The default deployment target is Docker Compose with a single container and a persistent `/data` volume.

The Windows alternative is **Windows Standalone (no installation required)**. Its ZIP contains only
the application. Without an explicit `RAILKEEPER_DATA_DIR`, persistent data is stored at
`%LOCALAPPDATA%\RailKeeper\data`, outside the replaceable program directory. Older data beside the
executable is copied once and retained as a source copy. Different old and new databases stop startup
for manual review instead of being selected or combined automatically.

Copy `.env.example` to `.env` only if you want to override operational settings such as upload limits, secure cookies, the GitHub release update endpoint or a manually configured printer list. Docker Compose sets the required container paths for data, migrations, seeds and static files itself.

For production operations, TLS, update, backup, restore and security-setting checklists are maintained in `docs/production-runbook.md`.

## Start from the published image

```bash
docker compose pull
docker compose up -d
```

By default Compose uses:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:latest
```

`latest` tracks stable version tags. Beta releases are published as GitHub
Prereleases and can be pinned explicitly:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:v0.1.15-beta.1
```

The moving `beta` tag tracks the newest prerelease. The moving `edge` tag tracks
pushes to `main` and is intended only for development smoke tests.

For a fixed release, set the image tag in `.env` before pulling:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:v0.1.14
```

Then run:

```bash
docker compose pull
docker compose up -d
```

## Build locally from source

If no published image is available yet, or if you intentionally want to build the checked-out source tree:

```bash
docker compose up -d --build
```

If an older `.env` contains `RAILKEEPER_DATA_DIR`, `RAILKEEPER_MIGRATIONS_DIR`, `RAILKEEPER_SEEDS_DIR` or `RAILKEEPER_STATIC_DIR`, remove those entries before rebuilding. These paths must stay inside the container and are fixed by `docker-compose.yml`.

Docker updates remain:

```bash
docker compose pull
docker compose up -d
```

The `/data` volume is not part of the image replacement. Export an application backup and, before
migration-heavy updates, also back up the volume or its host storage. Automatic pre-migration SQLite
copies contain private authentication and inventory data and do not replace external backups.

## In-app bug reporting

The in-app form requires an issue token/repository and validation of the fixed issue labels.
Without those settings the global action remains an external GitHub link. Prefer a fine-grained
token limited to **Issues: Read and write** on the issue repository, plus metadata read.

Screenshots are a separate optional capability. To enable them, configure a second token limited to
**Contents: Read and write** on a different public asset repository. Without that pair, text reports
remain available and the screenshot field is hidden. Never grant an asset-only token access to the
source repository. An operator may explicitly choose an existing credential for a private test instance;
that is not the recommended credential scope for general installations.

Create these labels in the issue repository before starting RailKeeper:

```text
bug
source:in-app
area:overview
area:vehicles
area:accessories
area:exhibition
area:import-export
area:digital-centers
area:settings
area:other
```

Then configure `.env` with repository names in `owner/name` form. Keep token values out of Git:

```env
RAILKEEPER_GITHUB_ISSUE_TOKEN=
RAILKEEPER_GITHUB_ISSUE_REPOSITORY=github-owner/RailKeeper
RAILKEEPER_GITHUB_ASSET_TOKEN=
RAILKEEPER_GITHUB_ASSET_REPOSITORY=github-owner/railkeeper-bug-assets
RAILKEEPER_BUG_REPORT_LABELS=bug,source:in-app
```

The optional contact field is available only with an age X25519 recipient. Generate and retain the
private identity outside the RailKeeper data directory, backups, repository, and `.env` file:

```bash
age-keygen -o railkeeper-bug-report-identity.txt
age-keygen -y railkeeper-bug-report-identity.txt
```

Set only the printed public `age1...` recipient:

```env
RAILKEEPER_BUG_REPORT_CONTACT_AGE_RECIPIENT=
```

Restart RailKeeper after changing tokens, repositories, labels, or the recipient. Check that the
floating action opens the form. For token rotation, create the replacement token with the same
narrow scope, update `.env`, restart and verify the form, then revoke the old token. If verification
fails, RailKeeper leaves in-app submission disabled and retains the external fallback.

Screenshots are re-decoded and re-encoded before upload, but remain public in the asset repository.
Users must explicitly consent for every selected screenshot. The optional email address is placed
in the issue only as age-encrypted ciphertext. See the
[operator runbook](../docs/production-runbook.md#in-app-fehlerberichte) for recovery and privacy
checks.

## Password reset email

Password reset links are not returned to the browser. SMTP can be configured and tested in the Admin UI under `Einstellungen > Authentifizierung > SMTP für Passwort-Reset`. Configure SMTP in `.env` when deployment defaults are preferred:

```env
RAILKEEPER_PUBLIC_URL=https://railkeeper.example.test
RAILKEEPER_SMTP_HOST=smtp.example.test
RAILKEEPER_SMTP_PORT=587
RAILKEEPER_SMTP_USER=railkeeper@example.test
RAILKEEPER_SMTP_PASSWORD=change-me
RAILKEEPER_SMTP_FROM=railkeeper@example.test
RAILKEEPER_SMTP_TLS=starttls
```

`RAILKEEPER_SMTP_TLS` supports `starttls`, `implicit` and `none`. Without SMTP, local recovery links are written only to the backend log.

The issue token/repository enable text reports. Asset credentials are optional and enable screenshots.
Diagnostic inclusion can be disabled by each reporter; the remembered choice stops local error capture
and omits both browser and server diagnostics. Process memory/uptime and SQLite health are included
only when enabled. No device probes or inventory data are collected.

Der Issue-Token und das Ziel-Repository ermöglichen Textberichte. Asset-Zugangsdaten sind optional
und schalten Screenshots frei. Nutzer können die Diagnose abschalten; die gespeicherte Auswahl
stoppt die lokale Fehlererfassung und lässt Browser- und Serverdiagnose weg. Prozessspeicher, Laufzeit
und SQLite-Zustand werden nur bei eingeschalteter Diagnose ergänzt. Es erfolgen keine Geräteabfragen.
