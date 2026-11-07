---
title: Manuelle Wertentwicklung
description: Datierte Fahrzeugbewertungen erfassen, berichtigen und nach Währung auswerten.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-11-07
---

# Manuelle Wertentwicklung

Bewertungen beschreiben deine eigene Einschätzung zu einem bestimmten Datum. Listen- und
Kaufpreise bleiben unverändert. RailKeeper ruft keine Preisportale ab und ergänzt keine
fehlenden früheren Marktwerte.

<GuideFigure name="valuations" alt="Datierte manuelle Fahrzeugbewertung mit Herkunft und erhaltenem Verlauf." caption="Manuelle Bewertung und Berichtigung im Entwicklungsstand. Ausschließlich künstliche Demodaten. Anklicken für die volle Größe." />

## Bewertung erfassen

1. Öffne ein gespeichertes Fahrzeug über **Bearbeiten > Wertentwicklung**.
2. Gib Betrag, dreistelligen Währungscode, Bewertungsdatum und Herkunft an, beispielsweise
   eine eigene Zustandseinschätzung oder einen nachvollziehbaren Beleg.
3. Wähle **Bewertung jetzt speichern**. Diese Schaltfläche speichert ausschließlich die
   Bewertung. Offene Fahrzeug-, Funktions-, CV- und andere Eingaben bleiben ungespeichert.

Gib den Betrag ohne Tausendertrennzeichen ein. Punkt und Komma bedeuten das Dezimalzeichen;
bis zu sechs Nachkommastellen bleiben exakt erhalten. `1234,56` bedeutet 1234,56 Einheiten,
`1.234` bedeutet 1,234 Einheiten. Negative Werte, Exponentialschreibweise und mehr als sechs
Nachkommastellen werden abgewiesen. Währungen werden weder umgerechnet noch vermischt.

Das Bewertungsdatum und die tatsächliche Erfassungszeit werden getrennt angezeigt.
Ein später eingetragener früherer Wert erhält seine tatsächliche heutige Erfassungszeit.
Admin und Editor dürfen erfassen; Viewer darf lesen und exportieren. Messe erhält keinen
Zugriff auf den Fahrzeugbestand oder die Sammlungsauswertung.

## Berichtigen und zurückziehen

**Bewertung berichtigen** bereitet einen zusätzlichen Eintrag vor. Prüfe Angaben und Herkunft,
bevor du speicherst. Der ursprüngliche Eintrag bleibt erhalten und wird als berichtigt markiert.
Zwei Personen können denselben Eintrag nicht unabhängig voneinander berichtigen. Wurde er
inzwischen geändert, lade den Verlauf und vergleiche ihn; deine Eingabe bleibt erhalten.

**Bewertung zurückziehen** verlangt eine neue Begründung. Der Eintrag bleibt im Verlauf,
wird für die Auswertung aber nicht mehr verwendet. Eine ältere weiterhin gültige Bewertung
kann dann wieder der letzte bekannte Wert des Fahrzeugs sein.

Fehlt eine eindeutige Speicherbestätigung, bleibt die Eingabe gesperrt erhalten. Die
Wiederherstellungskennung und die Eingabe bleiben je Konto und Fahrzeug im aktuellen
Browser-Reiter auch nach dem Schließen des Dialogs oder Neuladen der Seite erhalten.
Sie lösen keine automatische Speicherung aus. Kann der Browser diesen ausstehenden Vorgang
nicht sichern, wird die Bewertung nicht gesendet. Wähle
**Gespeicherten Stand prüfen**. RailKeeper liest den vollständigen Verlauf und prüft diese
Eingabe, ohne sie erneut zu senden. Ist sie bereits gespeichert, erscheint die Bestätigung.
Ist sie nicht vorhanden, kannst du die erhaltene Eingabe prüfen und ausdrücklich speichern.

## Sammlung zum Stichtag auswerten

Die Übersicht enthält **Manuell bewertete Sammlung**. Wähle den Stichtag. Es zählt je Fahrzeug
die letzte gültige Bewertung bis einschließlich dieses Datums, mit dem heutigen Stand der
Berichtigungen. Die Karte zeigt bewertete und unbewertete Fahrzeuge und getrennte Summen je
Währung. Fehlende Werte bleiben unbekannt. Sets werden über ihre einzelnen Fahrzeuge gezählt;
Setpreise werden nicht zusätzlich addiert.

## Export und Sicherung

**Verlauf als JSON** enthält sämtliche Einträge des Fahrzeugs, einschließlich älterer Seiten,
Berichtigungen, Rücknahmen, Herkunft und Erfassungszeiten. Der Export dient der vollständigen
Ansicht und Archivierung dieses Verlaufs.

Für eine Wiederherstellung verwende die [Anwendungssicherung](../../administration/backup-restore.md).
Format 28 sichert die Bewertungstabelle vollständig. Ältere Sicherungen erhalten beim Restore
einen leeren Bewertungsverlauf. Flache Fahrzeug-CSV-Dateien ersetzen keine vollständige
Anwendungssicherung. Ältere RailKeeper-Versionen können Format 28 nicht vollständig lesen;
setze die Formatnummer deshalb nicht von Hand herab.
