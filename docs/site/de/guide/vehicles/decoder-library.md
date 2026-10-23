---
title: Decoder nachschlagen
description: Decoder nachschlagen.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Decoder nachschlagen

Die Decoderbibliothek erklärt, welche Angaben RailKeeper für ein bestimmtes Decodermodell kennt. Sie ist ein Nachschlagewerk und verändert keine Fahrzeugwerte.


<GuideFigure name="decoder-library" alt="Filtere nach Hersteller, Familie oder Modell; die Detailansicht zeigt Belege und Grenzen." caption="Filtere nach Hersteller, Familie oder Modell; die Detailansicht zeigt Belege und Grenzen. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Ein Modell finden

1. Öffne **Einstellungen > Daten > Allgemeine Stammdaten > Decoderbibliothek**.
2. Wähle links einen Hersteller oder suche nach Modell, Artikelnummer, Herstellerkennung oder CV.
3. Öffne das passende Modell und prüfe Anschlüsse und Artikelnummern.
4. Lies den **Prüfumfang**. Katalogisiert bedeutet nicht, dass bereits jede CV ausgewertet wird.
5. Öffne **CV-Definitionen**, **Erkennung** oder **Quellen**, je nachdem, was du wissen möchtest.

## Was bedeuten die Angaben?

| Anzeige | Bedeutung |
| --- | --- |
| DCC-Grundlagen | allgemeine, dokumentierte Basisbedeutungen |
| CV-Regeln geprüft | für das Profil sind überprüfte Auswertungsregeln vorhanden |
| Herstellerreferenz vorhanden | eine belegte Tabelle oder ein Hinweis kann nachgeschlagen werden; daraus folgt nicht automatisch eine Auswertungsregel |
| teilweise erfasst oder keine passende Unterlage | die verbleibende Grenze ist ausdrücklich dokumentiert |

Ein ähnlicher Modellname oder derselbe Stecker beweist keine identische CV-Belegung. Wähle das genaue Modell und beachte den Firmwarestand der Quelle.

## Was passiert mit meinen gespeicherten CVs?

RailKeeper wertet vorhandene Werte beim Anzeigen anhand der verfügbaren Bibliothek aus. Ein erneuter Import ist für neue Erklärungen normalerweise nicht nötig. Die Bibliothek ersetzt weder fehlende Werte noch eine unbekannte Indexseite. Widersprüche bleiben sichtbar.

Bei ECoS-Speicherwerten kann der hinterlegte Decodertyp zusammen mit passender Herstellerkennung die Zuordnung ermöglichen. Die Herkunft bleibt ECoS. [CV-Werte verstehen](decoder-cv).
