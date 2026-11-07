---
title: Sicherung und Wiederherstellung
description: Sicherung und Wiederherstellung.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-11-07
---

# Sicherung und Wiederherstellung

**Ziel:** Deinen Bestand gegen Datenverlust absichern. Für diese Seite brauchst du die Rolle **Admin**.


<GuideFigure name="backup" alt="Eine ausgewählte Sicherung wird geprüft, bevor eine Wiederherstellung möglich wird." caption="Eine ausgewählte Sicherung wird geprüft, bevor eine Wiederherstellung möglich wird. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Eine Anwendungssicherung erstellen

1. Öffne **Einstellungen > Backup**.
2. Wähle im Bereich **Backup erstellen** die gleichnamige Schaltfläche.
3. Bewahre die heruntergeladene JSON-Datei an einem zweiten Speicherort auf.
4. Prüfe vor einem Update, dass eine aktuelle Sicherung vorhanden ist.

Die Anwendungssicherung enthält Bestand, Zubehör, Wartungen, CVs mit Herkunft, Stammdaten, Ausstellungsdaten und gespeicherte Medien. Benutzerkonten, Passwörter, Sitzungen und lokale Systemeinstellungen sind nicht enthalten. Sie ist deshalb keine vollständige Serversicherung.

Format 27 enthält zusätzliche Tabellen für Decoderexemplare, Einbauten, CV-Arbeitsgruppen und
unveränderliche Teilstände mit eingefrorenen Quellen. Unbekannte Einbaudaten bleiben unbekannt;
ein Erfassungszeitpunkt ist kein Auslesezeitpunkt. Gesicherte Quellen bleiben auch nach dem Entfernen
der ursprünglichen CV-Datei erhalten. Vorhandene Altgruppen werden beim Start oder Restore einmalig
als übernommene Teilstände erhalten, ohne eine Hardwarezuordnung zu erraten. Bereits erfasste Stände
werden dabei nicht neu geschrieben. Die Übernahme liest Dateien aus dem Uploadbereich der Sicherung;
Verweise auf andere lokale Bereiche werden vor dem Wiederherstellen abgewiesen.

Wie Format 26 sichert es Artikelrevisionen, Fahrzeug-/Set-Zuordnungen, individuelle Artikelabweichungen
und den ursprünglichen Zuordnungsmodus. Ungeklärte Katalogvorschläge bleiben beim Wiederherstellen
erhalten. Aktive Artikelangaben müssen mit der gespeicherten Fahrzeugansicht übereinstimmen.
Ältere unterstützte Sicherungen einschließlich Format 26 bleiben einlesbar. Ein älteres Programm kann
das neue Format nicht vollständig wiederherstellen. Ändere deshalb die Formatnummer nicht manuell.
Bewahre vor einer Datenmodellumstellung zusätzlich eine vollständige konsistente Installation und
das dazu passende bisherige Programm auf. Nach neuen Eingaben ist diese ältere Kopie kein
verlustfreier Rückweg.

## Eine Sicherung zunächst prüfen

1. Wähle unter **Backup wiederherstellen** die Sicherungsdatei aus.
2. Warte auf die Prüfung. Kontrolliere Datum, Format, Datensatzanzahl, Warnungen und Fehler.
3. Wenn du nur die Datei prüfen möchtest, beende den Vorgang hier. Die Dateiauswahl allein stellt nichts wieder her.

## Daten bewusst wiederherstellen

::: warning Vorher den aktuellen Stand sichern
Die Wiederherstellung ersetzt die lokalen App-Daten und gespeicherten Medien durch den Inhalt der Datei. Sichere zuerst den jetzigen Bestand. Die vorhandene lokale Anmeldung wird nicht durch Konten aus der Datei ersetzt.
:::

1. Wähle eine kompatible Datei und lies das Prüfergebnis.
2. Gib **WIEDERHERSTELLEN** in das Bestätigungsfeld ein.
3. Wähle **Backup einspielen** und bestätige die Rückfrage.
4. Prüfe danach Fahrzeuge, CVs, Bilder und wichtige Stammdaten.

## Welche Sicherung brauche ich?

| Datei oder Kopie | Zweck |
| --- | --- |
| Bestands-Export | Datenaustausch und Tabellen; kein vollständiger Wiederherstellungsstand |
| Anwendungssicherung | App-Daten zwischen kompatiblen Installationen übertragen oder wiederherstellen |
| vollständige Datei-/Volumensicherung | Serverausfall und Rückkehr zum bisherigen Installationszustand, einschließlich privater lokaler Daten |

Eine Kopie im selben Container schützt nicht vor dessen Verlust. Betreiber sichern Datenbank und Dateispeicher konsistent und bewahren eine Kopie außerhalb des Systems auf. [Updates und Betrieb](system-operations).

[Bebilderte Übung mit einer JSON-Anwendungssicherung](restore-exercise) ·
[Vollständige private Serversicherung getrennt wiederherstellen](server-restore).
