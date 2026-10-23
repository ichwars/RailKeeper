---
title: Updates und Betrieb
description: Updates und Betrieb.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Updates und Betrieb

**Ziel:** Die Anwendung erneuern, ohne Bestand und Dateien zu verlieren. Prüfe zuerst, welche Version tatsächlich läuft. Ein GitHub-Release und der aktuelle Entwicklungsstand `main` können sich unterscheiden.

## Vor jedem Update

1. Erstelle eine [Anwendungssicherung](backup-restore) und bewahre sie außerhalb der Installation auf.
2. Sichere als Betreiber zusätzlich den vollständigen Datenordner beziehungsweise das Volume konsistent. Diese private Sicherung enthält auch lokale Anmeldedaten.
3. Notiere Version, Konfiguration und aktiven Datenpfad. Lies die Hinweise der Zielversion.
4. Plane eine kurze Unterbrechung und verhindere gleichzeitige Änderungen.

## Windows aktualisieren

Lade das passende neue ZIP aus den Releases herunter. Unter **Einstellungen > Allgemein > Updates** kann RailKeeper einen passenden Download anbieten. Es installiert ihn nicht selbst.

Beende RailKeeper, entpacke das ZIP in einen neuen Programmordner und starte die neue Anwendung. Kontrolliere Datenpfad, Anmeldung, Bestand und Bilder. Bewahre den alten Programmordner und die Sicherungen bis zur erfolgreichen Prüfung auf.

Bei einer Altdatenübernahme bleibt die Quelle erhalten. Existieren am alten und neuen Ort Datenbanken, stoppt der Start. Sichere beide Ordner getrennt und kläre, welcher Bestand aktuell ist. Führe die Ordner nicht zusammen.

## Docker aktualisieren

Wähle einen geprüften Image-Tag oder Digest. Danach im Verzeichnis deiner Compose-Konfiguration:

```sh
docker compose pull
docker compose up -d
docker compose ps
```

Das vorhandene Datenvolume muss erhalten bleiben. Ein automatisches Sicherheitsbackup vor Datenbankmigrationen liegt im Datenordner; es ersetzt keine externe Sicherung.

## Nach dem Update prüfen

- Antwortet `/health` erfolgreich?
- Funktionieren Anmeldung, Fahrzeugliste und die Bearbeitung eines Testdatensatzes?
- Sind Bilder, CVs mit Herkunft und wichtige Stammdaten vorhanden?
- Lässt sich eine Sicherung erstellen und prüfen?
- Sind benötigte Digitalzentralen vom Server aus erreichbar?

Bei Fehlern zuerst Logs, Datenpfad und Versionsstand prüfen. Ein älteres Programm ist nach einer Migration nicht automatisch mit der neuen Datenbank kompatibel. Für eine Rückkehr den gesicherten vollständigen Zustand verwenden. [Ausführliches Runbook](https://github.com/ichwars/RailKeeper/blob/main/docs/production-runbook.md).
