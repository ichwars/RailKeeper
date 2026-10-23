---
title: Sicherung und Wiederherstellung
description: Sicherung und Wiederherstellung.
audience: admin
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
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

[Bebilderte Wiederherstellungsübung auf einer getrennten Testinstallation](restore-exercise).
