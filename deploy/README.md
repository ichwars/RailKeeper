# Deployment

The default deployment target is Docker Compose with a single container and a persistent `/data` volume.

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
