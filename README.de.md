<p align="center">
  <img src="frontend/public/brand/railkeeper-logo.png" alt="RailKeeper" width="360">
</p>

<h1 align="center">RailKeeper</h1>

<p align="center">
  Selbst gehostetes Bestands-, Dokumentations- und Betriebswerkzeug für Modelleisenbahnsammlungen.
</p>

<p align="center">
  <a href="https://github.com/ichwars/RailKeeper/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/ichwars/RailKeeper?style=for-the-badge"></a>
  <a href="https://github.com/ichwars/RailKeeper/actions/workflows/ci.yml"><img alt="CI" src="https://img.shields.io/github/actions/workflow/status/ichwars/RailKeeper/ci.yml?branch=main&style=for-the-badge"></a>
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white">
  <img alt="Docker" src="https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white">
  <img alt="React" src="https://img.shields.io/badge/React-UI-61DAFB?style=for-the-badge&logo=react&logoColor=111">
  <img alt="SQLite" src="https://img.shields.io/badge/SQLite-local--first-003B57?style=for-the-badge&logo=sqlite&logoColor=white">
  <a href="LICENSE.md"><img alt="Lizenz" src="https://img.shields.io/badge/License-AGPL--3.0--only-7AC943?style=for-the-badge"></a>
  <a href="https://github.com/ichwars/RailKeeper/releases"><img alt="Downloads" src="https://img.shields.io/github/downloads/ichwars/RailKeeper/total?style=for-the-badge"></a>
</p>

<p align="center">
  <a href="README.md">English</a> · <strong>Deutsch</strong> ·
  <a href="https://ichwars.github.io/RailKeeper/de/">Dokumentation</a>
</p>

## Überblick

RailKeeper ist eine vollständige, selbst gehostete Anwendung, die Modelleisenbahnfahrzeuge,
Decoderdaten, Bilder, Dokumente, Wartungen, Ausstellungslisten und Importe in einem lokalen
Arbeitsbereich verwaltet. Sie läuft als einzelner Go-Dienst mit integriertem React-Frontend und
speichert sämtliche Betriebsdaten in SQLite.

Das Projekt richtet sich an private Sammlungen, Vereine und kleine Werkstätten, die ein ernsthaftes
Bestandssystem ohne Cloud-Abhängigkeit benötigen. RailKeeper hält Daten, Uploads und Backups unter
deiner Kontrolle und bietet gleichzeitig moderne Abläufe wie Artikelsuche im Web, strukturierte
Importprüfung, ECoS-Auslesung und die Suche nach neuen Releases.

Das Handbuch bündelt die vollständige Anleitung für Benutzer, Betrieb und Entwicklung. Die
Abdeckung wird in englischer und deutscher Sprache direkt gegen die Projektquellen geprüft.

## Funktionsumfang

- Responsive Betriebsübersicht mit Bestand, Werten, Wartung, Datenqualität und Zentralenstatus in einer Ansicht
- Profilbasierter Import und Export für Fahrzeuge, Zubehör und Ausstellungslisten mit Vorschau,
  Konfliktlösung, persistenten Aufträgen, Transferverlauf und lokalen Dateien. Stammdaten verbleiben
  ausschließlich unter Einstellungen.
- Eigener Digitalzentralen-Arbeitsplatz mit konfigurierten Zentralen, Lokvergleich, Live-Status,
  Diagnose und Synchronisierung mit Vorschau vor jedem Schreibvorgang
- Persistenter Ausstellungsbetrieb mit Veranstaltungen, Teilnehmern, mehrtägiger Fahrzeugzuordnung,
  Bereitschaftsprüfung, Konfliktbehandlung, Bildern, Funktionstasten, Druck und Listensperre
- Lokaler Bestand mit SQLite, Uploads und JSON-Backups
- Fahrzeugdatensätze mit Modelldaten, technischen Feldern, Eigentumsangaben, Bildern, Anhängen, QR-Codes und übersichtlichen Leseansichten
- Artikelbestand mit generierten Inventarnummern, sortierbarer Auswahlspalte, Mengen- oder Einzelverfolgung, Lagerortbeständen, Reservierungen, Einbauten, Dokumenten und Nutzungshistorie
- Fundament für Anlagen, Module, Aufbauten und Planrevisionen mit Versionskonfliktschutz; der
  Arbeitsbereich bleibt vorübergehend aus der Hauptnavigation ausgeblendet, solange er weiter
  verfeinert wird
- Websuche nach Artikeldaten mit konfigurierbaren Quellen, Barcode-/EAN-Eingabe,
  ZXing-Kamerascanner, typisierten Gleis-Fachangaben und ausdrücklicher feldweiser Prüfung
- PDF-Berichtsdialog für Bestandsübersicht und Detaillisten mit auswählbaren Fahrzeugen, QR-Codes und Bildern
- Responsiver Bestandsablauf mit mobil optimierten Dialogen, Filtern und Kameraersatz für die Barcode-Eingabe
- Der kontrollierte ECoS-Lokabgleich liest Lokdaten, CV-Werte und statische Funktionstasten und
  schreibt Name, Adresse und Protokoll erst nach Vorschau und Bestätigung. Geschwindigkeit,
  Richtung, aktive Funktionszustände und ECoS-Anlagenobjektmanager werden nicht überwacht. Z21 per
  UDP und CS3 per HTTP bleiben reine Verbindungsadapter.
- Decoder-Funktionszuordnung von F0 bis F31 mit Symbolbibliothek und gespeicherten SVG-/PNG-Grafiken
- Strukturierte CV-Werte, CV-Import und -Export, Decoderprofile, NMRA-CV8-Herstellerstammdaten und ESU-/LokProgrammer-Dateimetadaten
- Wartungen, Zustandshistorie und durchsuchbare Dokumentation je Fahrzeug
- Lokale Authentifizierung mit Ersteinrichtung, E-Mail-Adresse, Rollen, Sitzungen, Passwortänderung, tokenbasiertem Passwort-Reset und Audit-Log
- Benutzerspezifische Reihenfolge und Sichtbarkeit der Sidebar-Einträge
- Stammdatenverwaltung für Hersteller, Spurweiten, Epochen, Kategorien, Unterarten, Bahngesellschaften und Symbole
- Docker-Compose-Bereitstellung mit gehärtetem Laufzeitcontainer und persistentem `/data`-Volume
- Integrierte GitHub-Release-Prüfung mit Release-Notizen und benutzergesteuerter Aktualisierung

## Ansichten

RailKeeper ist um responsive Arbeitsansichten statt Marketingseiten aufgebaut. Die folgenden
Aufnahmen zeigen die aktuelle deutsche Oberfläche und konzentrieren sich auf die neuesten Abläufe:

| Responsive Übersicht | Profilbasierter Import/Export |
| --- | --- |
| ![Responsive RailKeeper-Übersicht auf Deutsch](docs/screenshots/overview-de.png) | ![Profilbasierter Import und Export in RailKeeper auf Deutsch](docs/screenshots/import-export-de.png) |
| Digitalzentralen | Ausstellungsbetrieb |
| ![RailKeeper-Arbeitsplatz für Digitalzentralen auf Deutsch](docs/screenshots/digital-centers-de.png) | ![RailKeeper-Arbeitsplatz für Ausstellungen auf Deutsch](docs/screenshots/exhibition-de.png) |

Weitere Detailabläufe:

| Artikel-Websuche | Ersatzteilsuche | Decoder-Geschwindigkeitskurve |
| --- | --- | --- |
| ![RailKeeper Artikelsuche mit Produktdetails](docs/screenshots/search_product_details.png) | ![RailKeeper Ersatzteilsuche und Fahrzeugteileliste](docs/screenshots/spare_parts_search.png) | ![RailKeeper Decoder-Geschwindigkeitskurve](docs/screenshots/speed-performance_curve.png) |

- **Übersicht** für Bestand, Wert, Wartung, Datenqualität und Zentralenstatus
- **Fahrzeugbestand** für Fahrzeugsuche, Filter, Leseansichten, Berichte, Bearbeitung, Uploads, CVs und Funktionstasten
- **Zubehör** für Artikeldaten- und Barcodesuche, sortierbare Bestandsdaten, Lagerbestände,
  Reservierungen, Einbauten, Dokumente und Historie
- **Ausstellung** für persistente Veranstaltungen, Teilnehmer, Fahrtage, Bereitschaft und Konflikte
- **Import/Export** für wiederverwendbare Profile und kontrollierte Übertragungen von Fahrzeugen,
  Zubehör und Ausstellungslisten
- **Digitalzentralen** für Auslesen, Vergleich, Diagnose, Live-Monitoring und bestätigte Schreibvorgänge
- **Einstellungen** für Stammdaten, Darstellung, Backups, Aktualisierungen und Authentifizierung

## Schnellstart

### Windows Standalone (ohne Installation)

Lade das Windows-x64-ZIP aus einem Release herunter, entpacke es vollständig und starte:

```text
start-railkeeper.bat
```

RailKeeper läuft lokal ohne zusätzliche Software. Persistente Daten liegen standardmäßig außerhalb
des austauschbaren Programmordners unter `%LOCALAPPDATA%\RailKeeper\data`. Das ZIP enthält bewusst
keine Datenbank, Uploads, Backups oder einen Ordner `data`. Das Ersetzen des entpackten
Programmordners ersetzt deshalb nicht die aktiven Benutzerdaten.

Beim ersten Start nach einer älteren Portable- oder Standalone-Version kopiert RailKeeper einen
vorhandenen Ordner `data` neben `RailKeeper.exe` in den sicheren Pfad und lässt die Quelle
unverändert. Enthalten beide Pfade unterschiedliche Datenbanken oder sonstige persistente Dateien,
stoppt der Start auf einer lokalen Sicherheitsseite und verändert keine der beiden Kopien. Den
aktiven absoluten Datenpfad und eine beibehaltene Migrationsquelle sehen Administratoren unter
**Einstellungen > Allgemein**. Setze `RAILKEEPER_DATA_DIR` nur für einen ausdrücklich gewünschten
alternativen Datenpfad. Ein Pfad im Programmordner oder auf einem Wechseldatenträger kann weiterhin
verloren gehen, wenn dieser Speicherort gelöscht oder ersetzt wird. Externe Backups bleiben nötig.

Erkennt Windows Standalone unter **Einstellungen > Allgemein > Updates** ein neueres Release, nennt
die Schaltfläche die verfügbare Version und startet den Download des passenden GitHub-ZIP. RailKeeper
entpackt, installiert, ersetzt oder startet dabei nichts automatisch neu. Erstelle eine Sicherung,
beende RailKeeper, entpacke das heruntergeladene ZIP in einen neuen Programmordner und prüfe nach dem
Start den angezeigten Datenpfad sowie den Bestand. Fehlt ein vertrauenswürdiges passendes ZIP, nutze
stattdessen die verlinkte GitHub-Release-Seite.

### Docker Compose

```bash
git clone https://github.com/ichwars/RailKeeper.git
cd RailKeeper
docker compose pull
docker compose up -d
```

Öffne:

```text
http://localhost:8080
```

Beim ersten Start öffnet RailKeeper die Einrichtung. Lege dort das erste Administratorkonto an.
RailKeeper enthält keine Standardzugangsdaten.

### Bestehende Docker-Installation aktualisieren

```bash
git pull
docker compose pull
docker compose up -d
```

SQLite-Datenbank, Uploads und lokale Dateien bleiben im Docker-Volume `railkeeper_data` erhalten.

Um statt `latest` ein bestimmtes Release festzulegen, trage Folgendes in `.env` ein:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:v0.1.20.1
```

Wenn du bewusst den ausgecheckten Quellstand bauen möchtest, verwende:

```bash
docker compose up -d --build
```

### Optionale Umgebungsdatei

Kopiere `.env.example` nur dann nach `.env`, wenn du Betriebseinstellungen wie sichere Cookies,
Upload-Limits, Druckerkonfiguration oder den GitHub-Release-Endpunkt überschreiben möchtest.

Diese Containerpfade dürfen in Docker Compose nicht überschrieben werden:

```env
RAILKEEPER_DATA_DIR=/data
RAILKEEPER_MIGRATIONS_DIR=/app/migrations
RAILKEEPER_SEEDS_DIR=/app/seeds
RAILKEEPER_STATIC_DIR=/app/web
```

## Lokale Entwicklung

Backend:

```bash
cd backend
go test ./...
go run ./cmd/railkeeper
```

Frontend:

```bash
cd frontend
npm ci
npm run build
```

Die Produktionslaufzeit stellt das gebaute Frontend aus `frontend/dist` bereit.

Erstelle ein Windows-Standalone-Paket:

```powershell
.\tools\build_windows_standalone.ps1
```

Das Skript baut das Frontend, kompiliert `RailKeeper.exe` für Windows x64, weist Benutzerdaten im
Paketbaum und ZIP zurück und erstellt
`dist\windows-standalone\RailKeeper-windows-x64-v<version>.zip`.

Nützliche lokale Standardwerte:

```env
RAILKEEPER_ADDR=:8080
RAILKEEPER_DATA_DIR=./data
RAILKEEPER_MIGRATIONS_DIR=./backend/migrations
RAILKEEPER_SEEDS_DIR=./backend/seeds
RAILKEEPER_STATIC_DIR=./frontend/dist
RAILKEEPER_COOKIE_SECURE=false
RAILKEEPER_UPDATE_CHECK_URL=https://api.github.com/repos/ichwars/RailKeeper/releases/latest
```

Optionale SMTP-Einstellungen für Passwort-Reset-E-Mails lassen sich in der Admin-Oberfläche unter
`Einstellungen > Authentifizierung > SMTP für Passwort-Reset` konfigurieren. Die folgenden
Umgebungsvariablen bleiben als Vorgaben für die Bereitstellung verfügbar:

```env
RAILKEEPER_PUBLIC_URL=https://railkeeper.example.test
RAILKEEPER_SMTP_HOST=smtp.example.test
RAILKEEPER_SMTP_PORT=587
RAILKEEPER_SMTP_USER=railkeeper@example.test
RAILKEEPER_SMTP_PASSWORD=change-me
RAILKEEPER_SMTP_FROM=railkeeper@example.test
RAILKEEPER_SMTP_TLS=starttls
```

Passwort-Reset-E-Mails werden nur versendet, wenn `RAILKEEPER_PUBLIC_URL` oder die in der
Admin-Oberfläche gespeicherte öffentliche URL eine gültige HTTP(S)-Origin ist. RailKeeper leitet
versendete Reset-Links niemals aus dem `Host`-Header der Anfrage ab.

Ist SMTP nicht konfiguriert, gibt der Browser keine Passwort-Reset-Links zurück. Nur für die lokale
Wiederherstellung schreibt das Backend den Link in das Serverprotokoll.

### Optionale OCR-Erkennung für gescannte Ersatzteil-PDFs

RailKeeper liest textbasierte PDF-Ersatzteillisten direkt. Installiere für gescannte PDFs ohne
Textebene entweder `ocrmypdf` oder sowohl `pdftoppm` als auch `tesseract` auf dem Host und halte die
Werkzeuge in `PATH` verfügbar. Der OCR-Ersatz wird nur verwendet, wenn die integrierte
PDF-Texterkennung keinen verwertbaren Ersatzteiltext findet.

```env
RAILKEEPER_PDF_OCR=on
RAILKEEPER_PDF_OCR_MAX_PAGES=4
```

Setze `RAILKEEPER_PDF_OCR=off`, um den Ersatz ausdrücklich zu deaktivieren.

## Architektur

```text
backend/
  cmd/railkeeper/          Go-Einstiegspunkt
  internal/api/            HTTP-Routen, Middleware und Antwortabbildung
  internal/application/    Anwendungsfälle, Validierung, Backup und Transaktionen
  internal/infrastructure/ SQLite, Migrationen und Seed-Laden
  migrations/              SQLite-Schemamigrationen
  seeds/                   Stammdaten-Seed als JSON
frontend/
  src/app/                 Shell, Routing und globale Stile
  src/features/            Einrichtung, Auth, Fahrzeuge, Ausstellung, Import/Export, Digitalzentralen, Einstellungen
  src/shared/              API-Adapter, i18n und gemeinsame Frontend-Typen
openapi/
  railkeeper.yaml          API-Vertrag
deploy/
  README.md                Hinweise zur Bereitstellung
docs/
  architecture.md
  production-runbook.md
  roadmap.md
  security.md
```

## Sicherheit

RailKeeper ist für vertrauenswürdige, selbst gehostete Umgebungen vorgesehen. Die
Standardinstallation vermeidet dennoch die üblichen Fehler:

- kein Standard-Administratorkonto
- Argon2id-Passworthashing
- HTTP-only-Sitzungscookies
- SameSite-Cookies und CSRF-Schutz
- Rollenprüfungen für Viewer-, Editor-, Admin- und Messe-Abläufe
- Ratenbegrenzung für Einrichtung, Anmeldung und Sitzungen
- Passwort-Reset-Links per E-Mail, wenn SMTP konfiguriert ist
- Audit-Log für relevante Sicherheits- und Datenaktionen
- Upload-Größenbegrenzung und Sperre ausführbarer Anhänge
- Laufzeitdaten werden von Git ignoriert

Setze für HTTPS-Bereitstellungen:

```env
RAILKEEPER_COOKIE_SECURE=true
```

## Zähler und Badges

Die README enthält einen GitHub-Zähler für Release-Downloads. GitHub stellt weder einen
zuverlässigen öffentlichen README-Aufrufzähler noch einen allgemeinen Installationszähler für
selbst gehostete Docker-Bereitstellungen bereit. Dafür wären Drittanbietertracking,
Paketregistermetriken oder ausdrücklich aktivierte Telemetrie erforderlich. RailKeeper aktiviert
nichts davon.

## Lizenz

RailKeeper wird unter `AGPL-3.0-only` veröffentlicht. Siehe [LICENSE.md](LICENSE.md).

RailKeeper ist eine lokale, selbst gehostete Anwendung. Die AGPL-3.0 hält Weiterentwicklungen offen
und verpflichtet Betreiber einer über ein Netzwerk bereitgestellten, veränderten Fassung, ihren
Nutzern den korrespondierenden Quellcode zugänglich zu machen. Damit schützt sie die dauerhafte
Offenheit des Projekts besser als die bisherige freizügige Lizenz. Die AGPL erlaubt kommerzielle
Nutzung. Bereits veröffentlichte Fassungen behalten die Lizenzbedingungen, unter denen sie
veröffentlicht wurden.

Die Softwarelizenz überträgt keine Rechte an Projektkennzeichen oder fremden Marken, Grafiken,
Dokumentationen oder Protokollrechten. Siehe [TRADEMARKS.md](TRADEMARKS.md) und
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md).

## Unterstützung

Freiwillige Zuwendungen helfen bei Entwicklungs- und Projektkosten. Sie begründen keinen Anspruch
auf kostenpflichtige Funktionen, Support, Reaktionszeiten oder besonderen Zugang.

- [GitHub Sponsors](https://github.com/sponsors/ichwars)
- [Ko-fi](https://ko-fi.com/ichwars)

Supportwege und Abgrenzungen stehen in [SUPPORT.md](SUPPORT.md).
