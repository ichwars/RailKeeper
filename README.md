# RailKeeper

Deine Modellbahnsammlung lokal verwalten: Fahrzeuge und Sets, Bilder, Dokumente,
Wartung, Zubehör, Wertentwicklung und Decoder/CV. RailKeeper läuft auf Windows
oder mit Docker auf deinem eigenen Server. Eine Zentrale ist dafür nicht erforderlich.
Die Beschreibung folgt Main; einzelne Funktionen können dem stabilen Release voraus sein.

[Download](https://github.com/ichwars/RailKeeper/releases/latest) ·
[Handbuch](https://ichwars.github.io/RailKeeper/de/guide/) ·
[Beispieldaten ausprobieren](https://ichwars.github.io/RailKeeper/de/try) ·
[English](#english)

## Start unter Windows

1. Das Windows-x64-ZIP aus dem [aktuellen Release](https://github.com/ichwars/RailKeeper/releases/latest)
   vollständig in einen eigenen Programmordner entpacken.
2. `start-railkeeper.bat` starten und das eigene Administratorkonto anlegen. Falls ein Einrichtungscode
   abgefragt wird, steht er lokal in `%LOCALAPPDATA%\RailKeeper\data\setup-code.txt`.
3. Die aktiven Daten liegen standardmäßig unter `%LOCALAPPDATA%\RailKeeper\data`.

Vor einem Update sichern und RailKeeper schließen. Das neue ZIP in einen neuen
Programmordner entpacken, starten und den angezeigten Datenpfad und Bestand prüfen.
Der Download in RailKeeper installiert oder ersetzt keine Dateien automatisch.
[Details und ältere Datenordner](docs/project-overview.en.md#windows-standalone-no-installation-required).

## Start mit Docker

```bash
git clone https://github.com/ichwars/RailKeeper.git
cd RailKeeper
docker compose pull
docker compose up -d
```

Danach [localhost:8080](http://localhost:8080) öffnen und das eigene Administratorkonto anlegen.
Falls ein Einmalcode abgefragt wird, diesen nach dem
[Betriebshandbuch](docs/production-runbook.md#container-einrichten) lokal lesen.
AMD64 und ARM64 werden unterstützt; auf Raspberry Pi ist ein 64-Bit-System erforderlich.
Der Bestand bleibt im persistenten Docker-Volume. Vor einem Update eine aktuelle,
getrennt gespeicherte Sicherung erstellen und die Wiederherstellung erproben.
[Installation und Upgrade](docs/production-runbook.md) · [Compose-Konfiguration](deploy/README.md).

## Dokumentation und Mitarbeit

| Aufgabe | Einstieg |
| --- | --- |
| Bedienung und Beispielsammlung | [Deutsches Handbuch](https://ichwars.github.io/RailKeeper/de/guide/) |
| Funktionsumfang und Bildschirmbilder | [Ausführlicher Überblick](docs/project-overview.de.md) |
| Entwicklung und Architektur | [Entwicklung](docs/site/de/development/index.md), [Architektur](docs/architecture.md), [Beiträge](CONTRIBUTING.md) |
| Tests und unterstützte Compiler | [Qualitätsprüfungen](docs/quality-gates.md) |
| Betrieb, Diagnose und Wiederherstellung | [Betriebshandbuch](docs/production-runbook.md), [Restore-Übung](docs/site/de/administration/restore-exercise.md) |
| Unabhängig nachvollziehbare Qualifikation | [Build- und Restore-Abnahme](docs/independent-qualification.md) |
| Sicherheit und private Daten | [Sicherheitskonzept](docs/security.md), [Meldungen](SECURITY.md) |

RailKeeper arbeitet lokal mit SQLite und privaten Dateien. App-Backups enthalten den
Bestand, aber keine lokalen Konten oder Anmeldedaten. Für einen vollständigen Serverumzug
wird zusätzlich eine geschützte Installationssicherung benötigt. Geräteänderungen erfolgen
nur nach Vorschau und ausdrücklicher Bestätigung im jeweiligen Arbeitsablauf.

## English

Manage your model railway collection on your own computer or server: vehicles and sets,
pictures, documents, maintenance, accessories, valuation history and decoder/CV data.
Windows and Docker are supported. Collection management does not require a command station.
This description follows Main; some features may be newer than the stable release.

[Download](https://github.com/ichwars/RailKeeper/releases/latest) ·
[Handbook](https://ichwars.github.io/RailKeeper/guide/) ·
[Try sample data](https://ichwars.github.io/RailKeeper/try)

**Windows:** Extract the Windows x64 release ZIP completely and run `start-railkeeper.bat`.
Create your own administrator on first start. If a setup code is requested, read it locally
from `%LOCALAPPDATA%\RailKeeper\data\setup-code.txt`. The default data directory is
`%LOCALAPPDATA%\RailKeeper\data`, outside the replaceable program folder.
Before updating, back up your data, close RailKeeper and extract the new ZIP into a new
program folder. Verify the displayed data path and inventory after starting it.

**Docker:** Use the commands above, then open [localhost:8080](http://localhost:8080).
Create your administrator. If a setup code is requested, read it locally as described in the
[runbook](docs/production-runbook.md#container-einrichten). Images support AMD64 and ARM64;
Raspberry Pi needs a 64-bit operating system.
The Docker volume retains the collection across container replacement.
Create a fresh external backup and practice a separate restore before updating.

| Task | Documentation |
| --- | --- |
| Full features, screenshots and local setup | [Project overview](docs/project-overview.en.md) |
| Build and contribute | [Development](docs/site/development/index.md), [contributing](CONTRIBUTING.md), [quality gates](docs/quality-gates.md) |
| Architecture | [Module boundaries](docs/architecture.md) |
| Operate, upgrade and recover | [Production runbook](docs/production-runbook.md), [restore exercise](docs/site/administration/restore-exercise.md) |
| Independent build/restore qualification | [Handoff checklist](docs/independent-qualification.md) |
| Security | [Security model](docs/security.md), [reporting](SECURITY.md) |

App backups preserve collection data and exclude local authentication. A protected installation
backup is additionally needed for full server recovery. Preview and explicit confirmation guard
device writes. See the overview for the exact scope of each command-station integration.

## Bestehende Verweise / Existing links

<a id="overview"></a>[Overview](docs/project-overview.en.md#overview)
<a id="screens"></a>[Screens](docs/project-overview.en.md#screens)
<a id="highlights"></a>[Highlights](docs/project-overview.en.md#highlights)
<a id="quick-start"></a>[Quick Start](docs/project-overview.en.md#quick-start)
<a id="windows-standalone-no-installation-required"></a>[Windows Standalone (no installation required)](docs/project-overview.en.md#windows-standalone-no-installation-required)
<a id="docker-compose"></a>[Docker Compose](docs/project-overview.en.md#docker-compose)
<a id="update-an-existing-docker-installation"></a>[Update an existing Docker installation](docs/project-overview.en.md#update-an-existing-docker-installation)
<a id="optional-environment-file"></a>[Optional environment file](docs/project-overview.en.md#optional-environment-file)
<a id="optional-in-app-bug-reporting"></a>[Optional in-app bug reporting](docs/project-overview.en.md#optional-in-app-bug-reporting)
<a id="local-development"></a>[Local Development](docs/project-overview.en.md#local-development)
<a id="optional-ocr-for-scanned-spare-parts-pdfs"></a>[Optional OCR for scanned spare-parts PDFs](docs/project-overview.en.md#optional-ocr-for-scanned-spare-parts-pdfs)
<a id="architecture"></a>[Architecture](docs/project-overview.en.md#architecture)
<a id="security"></a>[Security](docs/project-overview.en.md#security)
<a id="counters-and-badges"></a>[Counters And Badges](docs/project-overview.en.md#counters-and-badges)

## License

[AGPL-3.0-only](LICENSE.md) · [Marken und Kennzeichen / Trademarks](TRADEMARKS.md) ·
[Hinweise zu Drittanbietern / Third-party notices](THIRD_PARTY_NOTICES.md).

## Support

[Support und Grenzen / Support and boundaries](SUPPORT.md) ·
[Discussions](https://github.com/ichwars/RailKeeper/discussions) ·
[GitHub Sponsors](https://github.com/sponsors/ichwars) · [Ko-fi](https://ko-fi.com/ichwars).
Freiwillige Beiträge begründen keine Supportansprüche. Voluntary tips do not create support entitlements.
