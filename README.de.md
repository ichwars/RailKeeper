# RailKeeper

Deine Modellbahnsammlung lokal verwalten: Fahrzeuge und Sets, Bilder, Dokumente,
Wartung, Zubehör, Wertentwicklung und Decoder/CV. RailKeeper läuft auf Windows
oder mit Docker auf deinem eigenen Server. Eine Zentrale ist dafür nicht erforderlich.

[Download](https://github.com/ichwars/RailKeeper/releases/latest) ·
[Handbuch](https://ichwars.github.io/RailKeeper/de/guide/) ·
[Beispieldaten ausprobieren](https://ichwars.github.io/RailKeeper/de/try) ·
[English](#english)

## Start unter Windows

1. Das Windows-x64-ZIP aus dem [aktuellen Release](https://github.com/ichwars/RailKeeper/releases/latest)
   vollständig in einen eigenen Programmordner entpacken.
2. `start-railkeeper.bat` starten. Den lokalen Einrichtungscode aus
   `%LOCALAPPDATA%\RailKeeper\data\setup-code.txt` verwenden und das eigene Administratorkonto anlegen.
3. Die aktiven Daten liegen standardmäßig unter `%LOCALAPPDATA%\RailKeeper\data`.

Vor einem Update sichern und RailKeeper schließen. Das neue ZIP in einen neuen
Programmordner entpacken, starten und den angezeigten Datenpfad und Bestand prüfen.
Der Download in RailKeeper installiert oder ersetzt keine Dateien automatisch.
[Details und ältere Datenordner](docs/project-overview.de.md#windows-standalone-ohne-installation).

## Start mit Docker

```bash
git clone https://github.com/ichwars/RailKeeper.git
cd RailKeeper
docker compose pull
docker compose up -d
```

Danach [localhost:8080](http://localhost:8080) öffnen. Den lokalen Einmalcode nach dem
[Betriebshandbuch](docs/production-runbook.md#container-einrichten) lesen und das eigene Administratorkonto anlegen.
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

[English quick start and documentation](README.md#english).

## Bestehende Verweise

<a id="überblick"></a>[Überblick](docs/project-overview.de.md#überblick)
<a id="ansichten"></a>[Ansichten](docs/project-overview.de.md#ansichten)
<a id="funktionsumfang"></a>[Funktionsumfang](docs/project-overview.de.md#funktionsumfang)
<a id="schnellstart"></a>[Schnellstart](docs/project-overview.de.md#schnellstart)
<a id="windows-standalone-ohne-installation"></a>[Windows Standalone (ohne Installation)](docs/project-overview.de.md#windows-standalone-ohne-installation)
<a id="docker-compose"></a>[Docker Compose](docs/project-overview.de.md#docker-compose)
<a id="bestehende-docker-installation-aktualisieren"></a>[Bestehende Docker-Installation aktualisieren](docs/project-overview.de.md#bestehende-docker-installation-aktualisieren)
<a id="optionale-umgebungsdatei"></a>[Optionale Umgebungsdatei](docs/project-overview.de.md#optionale-umgebungsdatei)
<a id="optionale-fehlerberichte-direkt-aus-railkeeper"></a>[Optionale Fehlerberichte direkt aus RailKeeper](docs/project-overview.de.md#optionale-fehlerberichte-direkt-aus-railkeeper)
<a id="lokale-entwicklung"></a>[Lokale Entwicklung](docs/project-overview.de.md#lokale-entwicklung)
<a id="optionale-ocr-erkennung-für-gescannte-ersatzteil-pdfs"></a>[Optionale OCR-Erkennung für gescannte Ersatzteil-PDFs](docs/project-overview.de.md#optionale-ocr-erkennung-für-gescannte-ersatzteil-pdfs)
<a id="architektur"></a>[Architektur](docs/project-overview.de.md#architektur)
<a id="sicherheit"></a>[Sicherheit](docs/project-overview.de.md#sicherheit)
<a id="zähler-und-badges"></a>[Zähler und Badges](docs/project-overview.de.md#zähler-und-badges)
<a id="lizenz"></a>[Lizenz](docs/project-overview.de.md#lizenz)
<a id="unterstützung"></a>[Unterstützung](docs/project-overview.de.md#unterstützung)

[AGPL-3.0-only](LICENSE.md) · [Support](SUPPORT.md) · [Sicherheit](SECURITY.md).
