# Production Runbook

Dieses Runbook beschreibt den sicheren Betrieb einer kleinen RailKeeper-Installation mit Docker Compose, Reverse Proxy, TLS und lokalem Datenverzeichnis.

## Zielbild

- RailKeeper läuft als einzelner Container hinter einem Reverse Proxy mit HTTPS.
- SQLite-Datenbank, Uploads und Backups liegen im persistenten Docker-Volume `railkeeper_data` oder in einem explizit gesicherten Host-Verzeichnis.
- Updates werden über GHCR-Images bezogen und erst nach Backup und kurzem Smoke-Test produktiv genutzt.
- Benutzer, Sessions, Rate Limits, Audit Logs und lokale Einstellungen bleiben lokal und werden nicht über App-Backups wiederhergestellt.

## Erstinstallation

### Prozessorarchitektur und Raspberry Pi

Der Docker-Workflow baut und prüft `linux/amd64` und `linux/arm64` auf nativen Runnern.
Erst nach erfolgreichen Startprüfungen mit frischer SQLite-Datenbank, Frontend-Prüfung und
Sicherheitsscans für beide Architekturen werden gemeinsame Image-Tags veröffentlicht.
Docker wählt daraus automatisch die passende Architektur. Auf ARM64 ein eventuell gesetztes
`platform: linux/amd64` aus der eigenen Compose-Datei entfernen.

Auf dem Raspberry Pi ist ein 64-Bit-ARM-Betriebssystem erforderlich. `uname -m` muss dort
`aarch64` liefern; ein 32-Bit-Betriebssystem wird nicht unterstützt. Ältere Release-Tags erhalten
nicht rückwirkend ARM64-Images. Ab v0.1.22 liefern `latest` und die versionierten stabilen
Release-Tags beide Architekturen. `edge` folgt der Entwicklung auf `main`. Die verfügbaren
Architekturen vor einem Wechsel prüfen:

```bash
docker buildx imagetools inspect ghcr.io/ichwars/railkeeper:latest
```

### Container einrichten

1. `.env` nur anlegen, wenn Werte überschrieben werden müssen.
2. Container starten:

   ```bash
   docker compose pull
   docker compose up -d
   ```

3. Den einmaligen Einrichtungscode lokal auslesen:

   ```bash
   docker compose exec railkeeper cat /data/setup-code.txt
   ```

   Bei direktem Serverstart liegt `setup-code.txt` im konfigurierten Datenverzeichnis.
   Windows Standalone verwendet standardmäßig:

   ```powershell
   Get-Content "$env:LOCALAPPDATA\RailKeeper\data\setup-code.txt"
   ```

   Ein explizites `RAILKEEPER_DATA_DIR` hat Vorrang. Die Datei bleibt lokal und ist bei
   Linux-Starts nur für den Dateieigentümer lesbar. Code nicht veröffentlichen oder in
   Fehlerberichte übernehmen. Ein beschädigter Code kann bei gestopptem Server und noch
   fehlenden Konten durch Entfernen ausschließlich dieser Datei neu erzeugt werden.
4. Browser öffnen und den Code zusammen mit den eigenen Kontodaten im Setup-Dialog eingeben.
   Verbrauch des Codes, Admin, Rollen und Audit werden atomar gespeichert. Vorhandene
   Benutzer überspringen die Code-Erzeugung vollständig. Ein verbrauchter Code bleibt
   ungültig, auch wenn die Datei noch vorhanden ist. App-Backups und Diagnoseexporte
   enthalten diesen Code nicht.
5. Unter Einstellungen prüfen:
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
- Upload-Limits des Proxys passend zu `RAILKEEPER_MAX_IMAGE_MB` und `RAILKEEPER_MAX_ATTACHMENT_MB` setzen.
- Wenn der Proxy die Client-IP über `X-Forwarded-For` übermittelt, seine internen Netze explizit
  als kommaseparierte CIDRs in `RAILKEEPER_TRUSTED_PROXY_CIDRS` eintragen. Ohne diese Einstellung
  ignoriert RailKeeper Forwarding-Header für Rate-Limits.
- Bei produktivem HTTPS die öffentliche URL setzen. Der Standard `auto` aktiviert dann
  sichere Cookies:

  ```env
  RAILKEEPER_PUBLIC_URL=https://railkeeper.example.test
  RAILKEEPER_COOKIE_SECURE=auto
  RAILKEEPER_TRUSTED_PROXY_CIDRS=172.18.0.0/16
  ```

`auto` verwendet ausschließlich diese konfigurierte URL, keine Forwarding-Header. Ohne
öffentliche URL oder mit einer HTTP-URL bleibt lokaler HTTP-/LAN-Betrieb möglich.
Explizites `true` beziehungsweise `false` bleibt wirksam. Bei ausschließlich lokalem HTTP
darf kein `true` gesetzt sein, sonst sendet der Browser die Session-Cookies nicht.
Die Ersteinrichtung zeigt eine Abweichung zwischen tatsächlicher Verbindung und
Cookie-Einstellung an. Ein aus einer älteren Compose-Konfiguration übernommenes `false`
überschreibt die Automatik weiterhin und sollte bei HTTPS überprüft werden.

## ECoS-Lokbilder direkt übertragen

Für den direkten Bild-Upload benötigt RailKeeper zusätzlich zur TCP-Protokollverbindung
(standardmäßig Port 15471) Zugriff auf die ECoS-Weboberfläche über TCP-Port 80. Bei einer
ausgehenden Firewall beide Ports gezielt für die konfigurierte Zentrale freigeben. Die Anwendung
ändert keine Firewall-Regeln. Passwortgeschützte oder abweichende Webformulare werden derzeit
nicht unterstützt und führen zu einer verständlichen Fehlermeldung.

Der Ablauf befindet sich bei einem gespeicherten Fahrzeug unter Uploads → Zentralenbild erstellen
→ Direkt an ECoS übertragen. Nur Admins können schreiben. Vor jedem Upload werden Ziel-Lok,
freier Bildplatz und Vorher-/Nachher-Vorschau bestätigt. Während der Übertragung die Bildbibliothek
und Lokdaten der ECoS nicht parallel bearbeiten. Bei einem Teilerfolg oder Verbindungsabbruch
Bildplatz und Lokzuordnung auf der ECoS prüfen; RailKeeper wiederholt keine Schreibzugriffe und
startet die Zentrale nicht neu. Details: [ECoS-Bildübertragung](architecture/ecos-image-transfer.md).

## Daten Und Backups

Wichtige Daten liegen im konfigurierten Data-Verzeichnis:

- SQLite-Datenbank
- Uploads für Bilder, Anhänge und Decoder-Dateien
- lokale Laufzeitdaten

Regelmäßiger Ablauf:

1. In RailKeeper unter Einstellungen einen App-Backup-Export erstellen.
2. Zusätzlich das Docker-Volume oder Host-Datenverzeichnis konsistent vollständig sichern.
   Dafür Schreibzugriffe pausieren und die Anwendung stoppen. Zur vollständigen Sicherung gehören
   gespeicherte Dateien, lokale Konten und die private Startkonfiguration; das passende Image getrennt erhalten.
3. Backup außerhalb des Hosts ablegen.
4. Vor größeren Updates ein frisches Backup erzeugen.

Restore-Grundsätze:

- App-Backups importieren Fahrzeug-, Stammdaten-, Wartungs-, CV-, Upload- und Ausstellungsdaten.
- Lokale Benutzer, Rollen, Sessions, Rate Limits, Audit Logs, Passwort-Hashes, SMTP-/System- und Benutzereinstellungen werden bewusst nicht importiert.
- Vor Restore immer die Backup-Validierung ausführen.
- Restore nur mit bewusst bestätigter, aktueller Sicherung durchführen.

## Updates

### Begrenzte Passwortprüfung

Login-Anfragen warten bei belegten Arbeitsplätzen höchstens eine Sekunde, ohne während
dieser Wartezeit eine weitere Argon2-Berechnung zu starten. Die Warteschlange ist begrenzt.
Bei Überlast antwortet der Server mit 429 und Retry-After 1; das bestehende IP-/Kontobudget
bleibt bei seinem Fünf-Minuten-Fenster. Passwort, zweiter Faktor und Dummy-Prüfung bleiben erforderlich.

| Variable | Standard | Zulässiger Bereich |
| --- | --- | --- |
| RAILKEEPER_LOGIN_PARALLEL | 2 | 1–4, begrenzt durch das Speicherbudget |
| RAILKEEPER_LOGIN_QUEUE | 8 | 0–64 wartende Anfragen |
| RAILKEEPER_LOGIN_WAIT_MS | 1000 | 0–2000; 0 weist belegte Kapazität sofort ab |
| RAILKEEPER_LOGIN_MEMORY_MB | 512 | 256–1024 MiB, mindestens Parallelität ×256 MiB |

Das Budget berücksichtigt den maximal akzeptierten gespeicherten Hash mit 256 MiB,
nicht nur den Standardhash mit 64 MiB. Es wird kein Speicher vorab reserviert; das Budget ist keine RSS-Obergrenze. Zusätzlich
benötigen Prozess, SQLite und Betriebssystem Speicher; auf kleinen Hosts Parallelität 1
mit mindestens 256 MiB Budget wählen. Bestehende gültige Hashes werden nicht abgeschwächt.
Compose leitet diese Werte weiter; Windows-/direkte Starts verwenden dieselben Variablen.
Ungültige Werte beenden den Start vor dem Öffnen der Datenbank. Eine Warteschlange
schützt vor Lastspitzen, nicht vor jeder dauerhaften Überlast.

Im isolierten Referenzlauf mit drei Laststößen zu je 16 Anfragen und
256-MiB-Hashes liefen höchstens zwei Ableitungen gleichzeitig. Der gemessene
Spitzenspeicher des Testprozesses betrug rund 1,2 GiB, einschließlich noch nicht
freigegebener Argon2-Speicherbereiche. Für diese Konfiguration ist entsprechend
Prozessreserve einzuplanen; das 512-MiB-Budget begrenzt gleichzeitig aktive
Passwortarbeit, nicht den gesamten Prozessspeicher.

### Aktualisieren

Standardablauf:

```bash
docker compose pull
docker compose up -d
```

Vorher:

- App-Backup exportieren.
- Vollständige konsistente Installationssicherung erstellen und ihre Wiederherstellung an einer isolierten Kopie prüfen.
- Zielinstallation, tatsächliche Daten-/Uploadpfade, Konfiguration und aktiven Image-Digest feststellen.
- Release Notes lesen.
- Prüfen, ob `.env` noch zur neuen Version passt.

Nachher:

- `/health` prüfen.
- Login testen.
- Fahrzeugliste öffnen.
- Ein Fahrzeugdetail öffnen.
- Backup-Validierung mit aktuellem Export testen.
- Audit Log auf ungewöhnliche Fehler prüfen.

Rückweg:

Ein altes Image ist nach einer Migration kein belegter verlustfreier Rückweg. Starte es nicht
ungeprüft auf der aktuellen Datenbank. Schema 80 und App-Backupformat 28 enthalten unter anderem
Decoderexemplare, Einbauzeiträume, unveränderliche CV-Teilstände und manuelle Bewertungsverläufe.
Schema 78/79 ergänzen private Anmeldeschutzdaten, die nicht in App-Backups gehören. Ältere Modelle können diese
Angaben nicht vollständig ausdrücken. Die Formatnummer einer Sicherung darf nicht herabgesetzt werden.

| Situation | Vorgehen |
| --- | --- |
| Unveränderter Datenzustand, kompatible vorherige Version nachgewiesen | Die geprüfte vorherige Software verwenden; Daten erhalten. |
| Migration fehlgeschlagen, seit Sicherung keine neuen Nutzereingaben | Schreibpause beibehalten, vollständige geprüfte Installationssicherung mit ihrer passenden alten Software wiederherstellen und kontrollieren. |
| Nach der Sicherung wurden neue Daten bestätigt | Aktuellen Zustand zusätzlich vollständig sichern und vorwärts reparieren oder eine vollständige geprüfte Übernahme aller Änderungen durchführen. Die alte Sicherung allein verliert neue Eingaben. |

Die automatische Sicherung vor einer Migration schützt die Datenbank zusätzlich. Sie ersetzt
keine vollständige Installation mit Dateien, Konten und Konfiguration. Details stehen in der
[Serversicherungs-Übung](site/de/administration/server-restore.md) und im
[Modernisierungsablauf](modernization-rollout.md). Veröffentlichung und Produktivumstellung
bleiben gesonderte Betreiberaufträge.

## Release-Veröffentlichung

Nach erfolgreicher `CI` für einen Push auf `main` prüft **Publish Release** die Laufzeitversion
in `backend/cmd/railkeeper/main.go` und die passende Release-Datei, beispielsweise
`docs/releases/v0.1.22.md`.
Der geprüfte Commit muss weiterhin der aktuelle Stand von `main` sein. Für eine noch nicht
veröffentlichte Version erstellt der Workflow Tag und GitHub-Release mit den zweisprachigen
Release-Notizen und startet die vorhandenen Docker- und Windows-Workflows für diesen Tag.

Bestehende Releases und Tags werden nicht überschrieben. Fehlen passende Release-Notizen,
schlägt die CI fehl oder verweist ein bestehender Tag auf einen anderen Commit, wird keine neue
Veröffentlichung ausgeführt. Ein erneuter Versuch desselben Release-Laufs kann die Paket-Builds
nach einer unterbrochenen Veröffentlichung erneut starten, solange `main` noch auf denselben
Commit zeigt. Bei inzwischen weitergelaufenem `main` die beiden Paket-Workflows manuell am
bestehenden Release-Tag starten. Der manuelle Release-Start auf `main` prüft ebenfalls zuerst
dessen erfolgreiche CI. Der Windows-Upload erhält die vorhandenen Release-Notizen.

Das repositoryeigene `GITHUB_TOKEN` benötigt nur `contents: write` zum Veröffentlichen und
`actions: write` zum Starten der Paket-Builds. Tag-Schreibvorgänge mit diesem Token lösen keine
Push-Workflows aus; deshalb werden die beiden Builds ausdrücklich per `workflow_dispatch`
gestartet. Vor einem produktiven Update den erfolgreichen Docker-Build und die Verfügbarkeit
des gewünschten Windows-Pakets im Release prüfen.

## Beta-Releases

Beta-Versionen werden mit einem Prerelease-Tag veröffentlicht, zum Beispiel:

```text
v0.1.15-beta.1
```

Der Windows-Portable-Workflow markiert solche Tags automatisch als GitHub Prerelease.
Der Docker-Workflow setzt für stabile Versionstags `latest`, für Prerelease-Tags
zusätzlich `beta`. Normale Pushes auf `main` erzeugen den beweglichen `edge`-Kanal. Bei einer
neuen, geprüften Version veröffentlicht **Publish Release** zusätzlich das Release; dessen Paket-Build setzt
für stabile Versionen `latest` beziehungsweise für Prereleases `beta`.

Beta-Installation per Docker:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:v0.1.15-beta.1
```

Alternativ kann für fortlaufende Beta-Tests der bewegliche Kanal verwendet werden:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:beta
```

Vor jedem Beta-Test:

- App-Backup exportieren.
- Optional Volume-/Dateisystem-Backup erstellen.
- Release Notes und bekannte Risiken lesen.
- Bei Migrationen prüfen, ob ein Rollback nur mit Datenbackup möglich ist.

Nach jedem Beta-Test:

- `/health` prüfen.
- Login, Fahrzeugliste und Fahrzeugdetail testen.
- Import/Export und Backup-Validierung smoke-testen.
- Audit Log auf Fehler prüfen.

## Security Settings

Empfohlene produktive Einstellungen:

SMTP kann im Admin UI unter `Einstellungen > Authentifizierung > SMTP für Passwort-Reset`
gespeichert und per Test-Mail geprüft werden. Die folgenden Variablen bleiben als
Deployment-Defaults nutzbar:

```env
RAILKEEPER_COOKIE_SECURE=auto
RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS=.pdf,.jpg,.jpeg,.png,.webp,.txt,.csv,.json,.xml,.zip
RAILKEEPER_PUBLIC_URL=https://railkeeper.example.test
RAILKEEPER_SMTP_HOST=smtp.example.test
RAILKEEPER_SMTP_PORT=587
RAILKEEPER_SMTP_USER=railkeeper@example.test
RAILKEEPER_SMTP_PASSWORD=change-me
RAILKEEPER_SMTP_FROM=railkeeper@example.test
RAILKEEPER_SMTP_TLS=starttls
```

Passwort-Reset-Mails werden nur mit einer gültigen HTTP(S)-Adresse in
`RAILKEEPER_PUBLIC_URL` beziehungsweise im Feld "Öffentliche URL" der SMTP-Einstellungen
versendet. Der eingehende `Host`-Header wird nicht für Links in E-Mails verwendet.

Prüfpunkte:

- Kein Standard-Admin existiert.
- Admin-Konten nur für Administration nutzen.
- Normale Pflege mit Editor-Rolle durchführen.
- Viewer-Rolle für reine Lesekonten.
- Messe-Rolle nur für Ausstellungsworkflow.
- Nicht benötigte Sessions regelmäßig widerrufen.
- Passwortwechsel nach geteilten oder temporären Zugängen durchführen.
- Sicherheitsereignisse nach Setup, Restore und Updates prüfen.

## In-App-Fehlerberichte

Die Funktion ist optional. Die vollständige GitHub- und age-Einrichtung steht in den
[Bereitstellungshinweisen](../deploy/README.md#in-app-bug-reporting). Für den Betrieb gelten
zusätzlich folgende Regeln:

Nutzer können die Diagnose abschalten. Die gespeicherte Auswahl stoppt die lokale Fehlererfassung
und lässt Browser- und Serverdiagnose weg. Prozessspeicher, Laufzeit und SQLite-Zustand werden nur
bei eingeschalteter Diagnose ergänzt. Es erfolgen keine Geräteabfragen.

- Für allgemeine Installationen das Issue-Token auf das Issue-Repository begrenzen. Es benötigt
  **Issues: Read and write**.
- Das optionale Asset-Token darf nur auf das getrennte öffentliche Screenshot-Repository zugreifen und
  benötigt ausschließlich **Contents: Read and write**.
- Die private age-Identität bleibt außerhalb des RailKeeper-Datenverzeichnisses, der App-Backups,
  des Repositorys und der `.env`. RailKeeper erhält nur den öffentlichen Empfänger.
- Screenshots sind trotz Bereinigung öffentlich. Keine Screenshots mit Personenbezug,
  Zugangsdaten, privaten Dokumenten oder nicht zur Veröffentlichung bestimmten Oberflächen senden.

RailKeeper prüft beim Start den lesenden Zugriff auf die Repositories und sämtliche benötigten Labels.
Scheitert dieser Zugriff, ist ein konfiguriertes Asset-Repository privat, fehlen Labels oder ist GitHub nicht erreichbar,
bleibt das In-App-Formular deaktiviert. Die Schaltfläche führt dann zum normalen GitHub-Issue-Link.
Schreibrechte werden ohne Testübermittlung nicht geprüft; fehlende Schreibrechte zeigen sich erst beim Versand.
Ohne optionale Asset-Konfiguration bleiben Textberichte verfügbar. Die Ursache wird nur als sichere
Fehlerkategorie protokolliert, Tokens erscheinen nicht im Log.

Prüfung nach Einrichtung oder Rotation:

1. RailKeeper neu starten und Startprotokoll auf die Kategorie `github_capability` prüfen.
2. Als Viewer oder Editor anmelden und **Fehler melden** öffnen.
3. Prüfen, dass Bereich, Beschreibung und Datenhinweis angezeigt werden.
4. Einen Testbericht ohne Screenshot absenden und Labels sowie zweisprachigen Issue-Inhalt prüfen.
5. Optional einen unkritischen Test-Screenshot mit Zustimmung senden und den öffentlichen,
   commitgebundenen Link im Issue prüfen.
6. Test-Issue und Test-Asset nach der Prüfung nach den Projektregeln bereinigen.

Ein neuer Benutzer kann höchstens fünf neue Bericht-IDs pro Stunde erstellen. Ein erneuter Versuch
mit derselben Bericht-ID zählt nicht als neuer Bericht. Wird das Issue erstellt, aber Screenshot oder
Kommentar schlagen fehl, zeigt RailKeeper einen Teilstatus und verwendet beim erneuten Versuch
dieselbe Bericht-ID und dasselbe Issue. Dadurch entstehen keine beabsichtigten Duplikate.

Die optionale E-Mail-Adresse steht im Issue als ASCII-armierter age-Block. Den Block einschließlich
Kopf- und Fußzeile in eine Datei kopieren und ausschließlich auf einem vertrauenswürdigen System
entschlüsseln:

```bash
age --decrypt -i railkeeper-bug-report-identity.txt contact.age
```

Token-Rotation:

1. Ersatz-Token mit identischem, engem Repository-Scope erstellen.
2. Wert in `.env` ersetzen und RailKeeper neu starten.
3. Formular und einen kontrollierten Testbericht prüfen.
4. Erst danach das alte Token widerrufen.

Bei vermutetem Token-Abfluss das betroffene Token sofort auf GitHub widerrufen, RailKeeper neu
starten und GitHub-Auditdaten sowie Repository-Änderungen prüfen. Ein widerrufenes oder ungültiges
Token deaktiviert das Formular beim nächsten Start. Das jeweils andere Token ebenfalls rotieren,
wenn eine gemeinsame Ablage kompromittiert sein könnte.

Der Diagnosevertrag ist bewusst eng. Enthalten sind RailKeeper- und Go-Version, Betriebssystem und
Architektur, Schema-Version, Bereitstellungsart, Browserfamilie und Hauptversion, Darstellungsmodus,
Sprache, Zeitzone, Viewport, aktueller Pfad und Bereich, Rollen, aktivierte
Digitalzentralen-Typen sowie höchstens 30 bereinigte Browserfehler. Nicht enthalten sind:

- Bestandsdaten und Bestandszahlen zu Fahrzeugen, Zubehör, Anlagen oder Ausstellungen
- Statistiken, Gerätezahlen, Zustände, Adressen, Hosts oder Ports verbundener Geräte
- Druckernamen, Seriennummern, Zugangsdaten, Tokens, Cookies oder API-Schlüssel
- Anfrageinhalte, Webhook-URLs, Hostnamen der Benutzer, lokale Pfade oder Benutzernamen

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
