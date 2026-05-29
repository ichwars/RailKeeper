# Production Runbook

Dieses Runbook beschreibt den sicheren Betrieb einer kleinen RailKeeper-Installation mit Docker Compose, Reverse Proxy, TLS und lokalem Datenverzeichnis.

## Zielbild

- RailKeeper läuft als einzelner Container hinter einem Reverse Proxy mit HTTPS.
- SQLite-Datenbank, Uploads und Backups liegen im persistenten Docker-Volume `railkeeper_data` oder in einem explizit gesicherten Host-Verzeichnis.
- Updates werden über GHCR-Images bezogen und erst nach Backup und kurzem Smoke-Test produktiv genutzt.
- Benutzer, Sessions, Rate Limits und Audit Logs bleiben lokal und werden nicht über App-Backups wiederhergestellt.

## Erstinstallation

1. `.env` nur anlegen, wenn Werte überschrieben werden müssen.
2. Container starten:

   ```bash
   docker compose pull
   docker compose up -d
   ```

3. Browser öffnen und den ersten Admin im Setup-Dialog anlegen.
4. Unter Einstellungen prüfen:
   - Version und Update-Status
   - Speicherübersicht
   - Benutzer und Rollen
   - Sicherheitsereignisse
   - Backup-Export

## TLS Und Reverse Proxy

RailKeeper sollte nicht direkt unverschlüsselt im Internet hängen. Empfohlen ist ein Reverse Proxy wie Caddy, Traefik, nginx oder ein NAS-/Router-Proxy.

Mindestanforderungen:

- Extern nur HTTPS anbieten.
- HTTP auf HTTPS umleiten.
- WebSocket-Sonderregeln sind nicht nötig.
- Upload-Limits des Proxys passend zu `RAILKEEPER_MAX_IMAGE_BYTES` und `RAILKEEPER_MAX_ATTACHMENT_BYTES` setzen.
- Bei produktivem HTTPS setzen:

  ```env
  RAILKEEPER_COOKIE_SECURE=true
  ```

Wenn lokal oder nur über HTTP getestet wird, muss `RAILKEEPER_COOKIE_SECURE=false` bleiben, sonst sendet der Browser die Session-Cookies nicht.

## Daten Und Backups

Wichtige Daten liegen im konfigurierten Data-Verzeichnis:

- SQLite-Datenbank
- Uploads für Bilder, Anhänge und Decoder-Dateien
- lokale Laufzeitdaten

Regelmäßiger Ablauf:

1. In RailKeeper unter Einstellungen einen App-Backup-Export erstellen.
2. Zusätzlich das Docker-Volume oder Host-Datenverzeichnis sichern.
3. Backup außerhalb des Hosts ablegen.
4. Vor größeren Updates ein frisches Backup erzeugen.

Restore-Grundsätze:

- App-Backups importieren Fahrzeug-, Stammdaten-, Wartungs-, CV-, Upload- und Ausstellungsdaten.
- Lokale Benutzer, Rollen, Sessions, Rate Limits, Audit Logs und Passwort-Hashes werden bewusst nicht importiert.
- Vor Restore immer die Backup-Validierung ausführen.
- Restore nur mit bewusst bestätigter, aktueller Sicherung durchführen.

## Updates

Standardablauf:

```bash
docker compose pull
docker compose up -d
```

Vorher:

- App-Backup exportieren.
- Optional Volume-/Dateisystem-Backup erstellen.
- Release Notes lesen.
- Prüfen, ob `.env` noch zur neuen Version passt.

Nachher:

- `/health` prüfen.
- Login testen.
- Fahrzeugliste öffnen.
- Ein Fahrzeugdetail öffnen.
- Backup-Validierung mit aktuellem Export testen.
- Audit Log auf ungewöhnliche Fehler prüfen.

Rollback:

1. In `.env` ein bekannt funktionierendes Image setzen, zum Beispiel:

   ```env
   RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:v0.1.12
   ```

2. Container neu ziehen und starten:

   ```bash
   docker compose pull
   docker compose up -d
   ```

3. Wenn Migrationen bereits angewendet wurden, nur mit vorherigem Datenbackup zurückrollen.

## Security Settings

Empfohlene produktive Einstellungen:

SMTP kann im Admin UI unter `Einstellungen > Authentifizierung > SMTP für Passwort-Reset`
gespeichert und per Test-Mail geprüft werden. Die folgenden Variablen bleiben als
Deployment-Defaults nutzbar:

```env
RAILKEEPER_COOKIE_SECURE=true
RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS=.pdf,.jpg,.jpeg,.png,.webp,.txt,.csv,.json,.xml,.zip
RAILKEEPER_PUBLIC_URL=https://railkeeper.example.test
RAILKEEPER_SMTP_HOST=smtp.example.test
RAILKEEPER_SMTP_PORT=587
RAILKEEPER_SMTP_USER=railkeeper@example.test
RAILKEEPER_SMTP_PASSWORD=change-me
RAILKEEPER_SMTP_FROM=railkeeper@example.test
RAILKEEPER_SMTP_TLS=starttls
```

Prüfpunkte:

- Kein Standard-Admin existiert.
- Admin-Konten nur für Administration nutzen.
- Normale Pflege mit Editor-Rolle durchführen.
- Viewer-Rolle für reine Lesekonten.
- Messe-Rolle nur für Ausstellungsworkflow.
- Nicht benötigte Sessions regelmäßig widerrufen.
- Passwortwechsel nach geteilten oder temporären Zugängen durchführen.
- Sicherheitsereignisse nach Setup, Restore und Updates prüfen.

## Lokaler Smoke-Test Nach Änderungen

Für lokale Quellcode-Prüfung:

```bash
cd backend
go test ./...
cd ../frontend
npm.cmd run build
```

Für eine laufende Docker-Installation:

1. Setup/Login prüfen.
2. Fahrzeug anlegen und Detailansicht öffnen.
3. Importpfad mit einer kleinen Testdatei oder Stammdaten-Import prüfen.
4. Backup exportieren und validieren.
5. Viewer-/Messe-Rollen mit eingeschränkten Aktionen prüfen.

## Betriebshinweise

- `/data` privat halten und nicht per Webserver ausliefern.
- Keine Tokens, Passwörter oder privaten Backups in Git committen.
- Externe Artikel-Suchergebnisse als untrusted input behandeln und nur bewusst übernehmen.
- Vor öffentlicher Exposition immer Reverse Proxy, TLS und sichere Cookies prüfen.
