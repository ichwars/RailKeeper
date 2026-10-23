---
title: Fahrzeugsets
description: Fahrzeugsets.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Fahrzeugsets

Mehrere Wagen aus einer Verpackung gehören zusammen, sollen aber eigene Fahrzeugnummern und Details behalten. Dafür verwendest du ein **Set**.


<GuideFigure name="vehicle-set" alt="Ein Set besitzt gemeinsame Produktdaten und individuell geführte Mitglieder." caption="Ein Set besitzt gemeinsame Produktdaten und individuell geführte Mitglieder. Künstliche Demodaten, keine Werkseinstellungen. Anklicken für die volle Größe." />

## Ein Set anlegen

1. Öffne **Fahrzeugbestand** und **Neues Fahrzeug**.
2. Wähle die Erfassung als **Set**.
3. Trage die gemeinsamen Angaben ein: Bezeichnung, Hersteller, Artikelnummer und Spurweite.
4. Lege mindestens zwei Mitglieder an. Gib jedem Wagen eine eigene Bezeichnung und Fahrzeugnummer.
5. Prüfe die gemeinsamen Kauf- und Verpackungsangaben und speichere das Set.

**Ergebnis:** Das Set und seine Mitglieder gehören zusammen. Die einzelnen Fahrzeuge bleiben einzeln auffindbar und bearbeitbar. RailKeeper vergibt für Set und Fahrzeuge getrennte Inventarnummern.

## Gemeinsame oder einzelne Angaben ändern?

| Angabe | Wo bearbeiten? |
| --- | --- |
| gemeinsame Artikelnummer, EAN, Kauf- oder Verpackungsangaben | Set bearbeiten |
| Fahrzeugnummer und technische Besonderheiten eines Wagens | das betreffende Fahrzeug bearbeiten |
| gemeinsames Bild oder Darstellung des Sets | Upload-Bereich des Sets |

Die Setansicht zeigt die Mitglieder. Prüfe vor dem Speichern, ob du gerade den gemeinsamen Datensatz oder einen einzelnen Wagen änderst. Ein Fahrzeugbild und ein gemeinsames Setbild können verschiedene Aufgaben erfüllen.

## Daten übertragen

Für einen vollständigen Umzug verwende eine [Anwendungssicherung](../../administration/backup-restore). Ein CSV-/JSON-Transfer kann die Setzugehörigkeit mitnehmen, enthält aber keine Medien und nicht alle Fahrzeug-Unterdaten. [Exportgrenzen](../import-export/exports).
