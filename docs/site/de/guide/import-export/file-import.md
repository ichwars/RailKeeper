---
title: Fahrzeugdateien importieren
description: Fahrzeugdateien importieren.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Fahrzeugdateien importieren

**Ziel:** Eine Bestandsdatei erst prüfen und anschließend bewusst übernehmen. Du brauchst Admin- oder Editor-Rechte für Fahrzeugimporte. Für eine große oder unbekannte Datei zuerst eine [Sicherung](../../administration/backup-restore) erstellen.


Zum Ausprobieren: [CSV- und JSON-Beispiele mit Schrittfolge](examples).


<GuideFigure name="import-dialog" alt="Starte einen neuen Import und wähle das passende Profil." caption="Starte einen neuen Import und wähle das passende Profil. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## 1. Ein passendes Profil anlegen

Öffne **Import/Export > Profil anlegen**. Vergib einen verständlichen Namen, etwa „Fahrzeuge aus Tabelle“. Wähle **Import**, den Bereich **Fahrzeuge** und das Format **CSV** oder **JSON**.

| Quelle | Geeignetes Format |
| --- | --- |
| Fahrzeugtabelle mit Spaltenüberschriften | CSV, genau ein Bereich |
| Paket aus dem aktuellen RailKeeper-Export | RailKeeper-JSON, auch mehrere Bereiche und Fahrzeugsets |
| LokProgrammer-CV-Liste oder Decoderprojekt | Im Fahrzeug unter **CV > Dateien & Import**, nicht hier |
| Lokliste aus ECoS oder CS3 | Im Arbeitsbereich **Digitalzentralen**, nicht als Dateitransfer |

Der aktuelle Transferdialog nimmt CSV oder versionierte RailKeeper-JSON-Pakete an. Eine beliebige XML-, TSV- oder JSON-Datei ist kein solches Paket. Wandle eine fremde Tabelle zunächst in CSV um. Die frühere Anleitung zum direkten XML/TSV-Import und zum eingeschränkten 14-Felder-JSON-Import gilt für diesen Ablauf nicht mehr.

## 2. Datei auswählen und Spalten zuordnen

1. Wähle **Neuer Import** und dein Importprofil.
2. Wähle die Datei unter **Durchsuchen**. Dadurch wird die Vorschau erstellt, noch kein Fahrzeug importiert.
3. Prüfe bei einer Fahrzeug-CSV die **Erkannte CSV-Zuordnung**.
4. Weise jede Quellspalte einem Zielfeld zu oder wähle ausdrücklich **Spalte ignorieren**.
5. Wähle **Zuordnung prüfen** beziehungsweise **Weiter zur Prüfung**.

Ein Zielfeld darf nur einmal zugeordnet sein. Bei wiederkehrenden Dateien kannst du die Zuordnung im Profil speichern. Eine neue Zuordnung berechnet die Vorschau erneut; prüfe danach auch die Konflikte erneut.

## 3. Vorschau und Konflikte prüfen

Kontrolliere Anzahl, Namen und Inventarnummern. Lies jede Warnung. Bei vorhandenen Datensätzen oder widersprüchlichen Verweisen bietet RailKeeper passende Entscheidungen an, beispielsweise vorhandene Daten verwenden, ersetzen, eine Kopie anlegen oder überspringen. Es sind nicht in jedem Konflikt alle Entscheidungen möglich.

Erkannte Fahrzeugsets erscheinen als zusammengehörige Gruppen. Prüfe Set-Inventarnummer, Mitglieder und Reihenfolge gemeinsam. Eine Set-Aktion gilt für die ganze Gruppe; ein Fahrzeug wird nicht unbemerkt aus einem anderen Set entfernt.

Fehlerhafte Quelldaten in der Datei korrigieren und erneut hochladen. Ein blockierender Konflikt muss vor der Übernahme gelöst sein.

## 4. Übernahme bestätigen

Bestätige erst die vollständig geprüfte Vorschau. Der Server übernimmt den geprüften Stand in einer Transaktion. Scheitert diese Übernahme, entsteht kein Teilimport. Kontrolliere danach den abgeschlossenen Auftrag und einige Fahrzeuge im Bestand.

Wurde der Auftrag zwischenzeitlich geändert, kann RailKeeper eine neue Prüfung oder einen erneuten Upload verlangen. Bestätige dann nur die neue Vorschau. **Schließen** verlässt den Dialog; **Auftrag abbrechen** beendet den Auftrag und erhält seinen Verlauf.

## Was wird übertragen?

Fahrzeugdateien übertragen Grunddaten und dokumentierte Set-Beziehungen. Sie enthalten keine Fahrzeugbilder, CV-Zeilen, Funktionen, Wartungen oder Decoderdateien. Dafür die jeweiligen Fahrzeugbereiche oder eine Anwendungssicherung verwenden. Die [Transferübersicht](./) erläutert Berechtigungen und Grenzen.
