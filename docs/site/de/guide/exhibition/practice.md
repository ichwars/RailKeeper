---
title: Einen Fahrtag prüfen, sperren und drucken
description: Einen Fahrtag prüfen, sperren und drucken.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Einen Fahrtag prüfen, sperren und drucken

**Ziel:** Eine gespeicherte Veranstaltung auf Vollständigkeit prüfen und den richtigen Bericht erzeugen.
Verwende für die Übung **Admin** und künstliche Daten. Der Ablauf schreibt nicht zu einer Zentrale.

## 1. Eine Veranstaltung vorbereiten

Wähle **Ausstellung > Neue Veranstaltung**. Erstelle **Demo-Betriebstag** mit Beginn und Ende.
Wähle sie links und füge einen Eintrag mit Besitzer **Demo-Verein**, Lokbezeichnung **Demo F0 F1 F3**,
Adresse **3** und Funktionstasten hinzu. Lasse zum Kennenlernen der Prüfung die **Zentrale / Schnittstelle** zunächst leer.

<GuideFigure name="exhibition-open" alt="Aktuelle Veranstaltung mit einem absichtlich unvollständigen Demo-Eintrag." caption="Die echte Prüfung meldet fehlende Angaben. Eine vorhandene Lok im Bestand ersetzt keinen vollständigen Betriebseintrag." />

## 2. Den Konflikt erkennen und korrigieren

Wähle **Liste sperren**. Die fehlende Schnittstelle verhindert eine stille Sperrung.

<GuideFigure name="exhibition-lock-conflicts" alt="Rückfrage beim Sperren der noch unvollständigen Liste." caption="Die Demo bricht diesen Dialog ab und korrigiert zunächst den Eintrag. Eine Ausnahme würde eine Begründung benötigen und sichtbar bleiben." />

Wähle **Abbrechen**, bearbeite den Eintrag und ergänze **Zentrale / Schnittstelle: Demo-Zentrale**.

<GuideFigure name="exhibition-entry" alt="Eintragsdialog nach dem Ergänzen der Demo-Zentrale." caption="Der Steuerungskontext ist vom Adapter beziehungsweise Decoderstecker getrennt. Speichern ändert nur den Betriebseintrag." />

Speichere und kontrolliere die Konfliktanzeige erneut. Ein fehlendes Bild kann weiterhin **Prüfen** ergeben.

## 3. Planung sperren

Wähle nochmals **Liste sperren**. Ohne die fehlenden Pflichtangaben wird die Sperre direkt gespeichert.

<GuideFigure name="exhibition-locked" alt="Gespeicherte Veranstaltung nach erfolgreicher Sperrung." caption="Einträge sind geschützt. Drucken und zum Status passende weitere Aktionen bleiben möglich." />

## 4. Den gesamten Bericht kontrollieren

Wähle **Drucken** im Veranstaltungskopf. Prüfe im Browser-Druckablauf Veranstaltung, Lok, Besitzer,
Adresse und Funktionstasten. Der Bericht enthält alle Veranstaltungseinträge, unabhängig vom Tabellenfilter.

<GuideFigure name="exhibition-print-report" alt="Tatsächlich erzeugter Bericht der zuvor geprüften Veranstaltung." caption="Das echte Berichts-HTML wurde separat im Browser dargestellt. Kein Betriebssystem-Druckdialog und keine Druckertreiberabnahme." />

Wähle erst danach einen Drucker oder ein PDF-Ziel. Bei Bedarf öffne die Planung mit **Liste entsperren** wieder.
[Status und Rollen](lists-and-locking) · [Einträge und Druckdetails](entries-and-printing).
