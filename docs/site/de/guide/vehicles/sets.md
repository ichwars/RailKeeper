---
title: Ein Wagenset anlegen und ändern
description: Ein Wagenset anlegen und ändern.
audience: user
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Ein Wagenset anlegen und ändern

**Ziel:** Gemeinsame Produktdaten einmal pflegen und jeden Wagen einzeln wiederfinden.
Verwende für die Übung Admin oder Editor und künstliche Daten.

## 1. Ein Zweier-Set anlegen

1. Öffne **Fahrzeugbestand > Neues Fahrzeug** und wähle die Erfassung als **Set**.
2. Nenne das Set **Demo-Wagenset**, Hersteller **Roco**, Artikelnummer **DEMO-SET**, Spurweite **H0**.
3. Wähle eine passende Kategorie und Gattung, beispielsweise Personenwagen und Schnellzugwagen.
4. Lege zwei Mitglieder an: **Demo-Wagen A** und **Demo-Wagen B**, mit unterschiedlichen Fahrzeugnummern.
5. Prüfe die gemeinsamen Angaben und speichere.

<GuideFigure name="vehicle-set" alt="Seterfassung mit gemeinsamen Angaben und einzelnen Mitgliedern." caption="Der Setentwurf bündelt Produktdaten; die Mitglieder behalten eigene Fahrzeugangaben. Künstliche Demo." />

**Erwartung:** Eine eigene Set-Inventarnummer und zwei getrennte Fahrzeug-Inventarnummern.
Die Setnummer kommt aus dem aktiven Set-Schema; die Mitglieder verwenden das passende Fahrzeugschema.

## 2. Mitglieder wiederfinden

Suche im Bestand nach **Demo-Wagenset** und klappe die Setzeile am Pfeil auf.

<GuideFigure name="set-members" alt="Aufgeklapptes Demo-Wagenset mit beiden einzeln geführten Wagen." caption="Das tatsächliche Zweier-Set in der Bestandsliste. Die Suche nach dem Setnamen hält die zugehörigen Mitglieder zusammen." />

Öffne einen Wagen, wenn du seine Fahrzeugnummer oder individuelle Technik prüfen möchtest.
Auf schmaleren Ansichten kann die Aktionstabelle seitlich gescrollt werden.

## 3. Gemeinsame Angaben ändern

Wähle **Bearbeiten** an der Setzeile, nicht an einer Mitgliedszeile.
Ändere die gemeinsame Artikelnummer von **DEMO-SET** zu **DEMO-SET-2**.

<GuideFigure name="set-shared-edit" alt="Setdialog mit geänderter gemeinsamer Artikelnummer vor dem Speichern." caption="Der gemeinsame Datensatz wird bearbeitet. Die individuellen Fahrzeugnummern der beiden Wagen werden dadurch nicht ersetzt." />

Wähle **Speichern**. Öffne das Set erneut und prüfe **DEMO-SET-2**, zwei Mitglieder und die
unterschiedlichen Fahrzeugnummern. Diese gemeinsame Änderung wurde in der Demo tatsächlich gespeichert.

| Angabe | Wo bearbeiten? |
| --- | --- |
| Gemeinsame Artikelnummer, EAN, Kauf- und Verpackungsangaben | Set |
| Fahrzeugnummer oder individuelle technische Besonderheit | Betreffendes Fahrzeug |
| Darstellung des Sets | Upload im Setdialog |
| Bilder eines Wagens | Uploads des betreffenden Fahrzeugs |

## 4. Zusammengehörige Daten übertragen

Ein CSV-/JSON-Transfer kann die Setstruktur übertragen, enthält aber keine Medien und nicht alle
Fahrzeug-Unterdaten. Für einen vollständigen kompatiblen Umzug verwende eine Anwendungssicherung.
[Exportgrenzen](../import-export/exports) · [Wiederherstellung üben](../../administration/restore-exercise).
