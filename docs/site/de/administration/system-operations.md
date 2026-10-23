---
title: Updates und Betrieb
description: Updates und Betrieb.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Updates und Betrieb

**Ziel:** Eine Version wechseln und anschließend nachvollziehbar prüfen, ob Bestand und Dateien
vollständig sind. Verwende für die erste Übung eine getrennte Testinstallation.

## 1. Version und Datenpfad festhalten

Öffne **Einstellungen > Allgemein > Updates** und notiere die aktuelle Version.
Ein Release und der Entwicklungsstand `main` können voneinander abweichen.
Der angebotene Download installiert nichts automatisch.

<GuideFigure name="update-version" alt="Updatebereich mit aktueller RailKeeper-Version und manueller Updateprüfung." caption="Die echte Browseroberfläche der lokalen Demo. Dieses Bild zeigt keine Durchführung eines Windows- oder Docker-Updates." />

Öffne auf derselben Einstellungsseite **Speichernutzung**. Notiere als Admin den angezeigten
aktiven Datenpfad und den Speichermodus. Prüfe zusätzlich deine Start- oder Compose-Konfiguration.

<GuideFigure name="update-storage" alt="Speichernutzung mit aktivem Datenpfad und Speichermodus der getrennten Demo." caption="Der Pfad gehört zur künstlichen Testinstallation. Bei Windows beziehungsweise Docker gelten die tatsächlich dort angezeigten Pfade und Volumes." />

## 2. Vor dem Austausch sichern

1. Erstelle eine Anwendungssicherung und bewahre sie außerhalb der Installation auf.
2. Prüfe die Datei ohne Restore. Eine [Wiederherstellungsübung](restore-exercise) auf einer zweiten
   Installation zeigt, ob du den Rückweg beherrschst.
3. Sichere zusätzlich den vollständigen Datenordner oder das Volume konsistent, einschließlich
   lokaler Anmeldung und Konfiguration. Stoppe dafür die Anwendung oder verwende einen dafür
   vorgesehenen konsistenten Snapshot.
4. Notiere Startparameter, öffentliche URL, Port, Image-Tag beziehungsweise ZIP-Version.
5. Lies die Hinweise der Zielversion und plane eine Unterbrechung ohne parallele Änderungen.

## 3. Windows aktualisieren

| Schritt | Aktion | Kontrolle |
| --- | --- | --- |
| Paket wählen | Passendes Windows-x64-ZIP aus den GitHub-Releases laden | Gewünschte Version und Architektur |
| Anwendung beenden | RailKeeper sauber schließen; nicht nur den Browser-Tab | Der bisherige Prozess läuft nicht mehr |
| Neu entpacken | Das gesamte ZIP in einen neuen Programmordner entpacken | Startdatei und mitgelieferte Unterordner vollständig |
| Starten | `start-railkeeper.bat` im neuen Ordner starten | Browser zeigt die neue Version |
| Daten prüfen | Gleichen Datenpfad und Speichermodus kontrollieren | Kein versehentlich leerer neuer Bestand |

Der Standard-Datenordner ist `%LOCALAPPDATA%\RailKeeper\data`; ein ausdrücklich konfigurierter
Pfad oder portabler Modus kann abweichen. Bewahre alten Programmordner und Sicherungen auf.
Existieren alte und neue Datenbanken gleichzeitig, kläre ihren Stand getrennt. Vermische keine Ordner.

## 4. Docker aktualisieren

Arbeite im Verzeichnis der vorhandenen Compose-Konfiguration. Behalte Datenvolume und
Umgebungsvariablen bei. Prüfe Image-Tag oder Digest vor dem Austausch:

```sh
docker compose config --images
docker compose pull
docker compose up -d
docker compose ps
docker compose logs --tail=100
```

Gib Ausgaben der Konfiguration und Logs nur nach Prüfung auf private Angaben weiter.
Ein automatisches Sicherheitsbackup vor Migrationen liegt im Datenordner und ersetzt keine externe Kopie.

## 5. Nach dem Update dieselben Kontrollpunkte prüfen

| Kontrolle | Erwartung |
| --- | --- |
| Version und `/health` | Gewünschter Stand, erfolgreiche Zustandsprüfung |
| Anmeldung | Eigenes vorhandenes Konto funktioniert |
| Bestand und Set | Bekannte Fahrzeuge und Mitglieder vorhanden |
| Bild und Dokument | Tatsächlich öffnen, nicht nur Vorschaubild zählen |
| CVs und Funktionen | Werte, Indexseiten, Herkunft und manuelle Angaben erhalten |
| Wartung und Stammdaten | Bekannte Datensätze vorhanden |
| Neue Sicherung | Export und Dateiprüfung erfolgreich |
| Zentrale, falls genutzt | Vom RailKeeper-Server erreichbar; zunächst Verbindung testen |

Bei Problemen notiere Version, Fehlermeldung und betroffenen Schritt. Prüfe Datenpfad und Logs.
Nach einer Migration ist ein älteres Programm nicht automatisch mit der neuen Datenbank kompatibel.
Für eine Rückkehr stelle den zusammengehörigen gesicherten Programm-, Daten- und Konfigurationsstand wieder her.

**Prüfgrenze:** Oberfläche und Restore-Übung sind mit der lokalen Demo geprüft. Diese Anleitung
ist mit Start-/Migrationscode und Runbook abgeglichen; die Bilder belegen keine Windows- oder Docker-Updateabnahme.
[Ausführliches Runbook](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md).
