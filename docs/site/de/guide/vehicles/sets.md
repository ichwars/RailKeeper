---
title: Ein Wagenset anlegen und ändern
description: Ein Wagenset anlegen und ändern.
audience: user
status: stable
reviewedVersion: 0.2.0
lastReviewed: 2026-11-09
---

# Ein Wagenset anlegen und ändern

**Ziel:** Gemeinsame Produktdaten einmal pflegen und jeden Wagen einzeln wiederfinden.
Verwende für die Übung Admin oder Editor und künstliche Daten.

## 1. Ein Zweier-Set anlegen

1. Öffne **Fahrzeugbestand > Neues Fahrzeug** und wähle die Erfassung als **Set**.
2. Nenne das Set **Demo-Wagenset**, Hersteller **Roco**, Artikelnummer **DEMO-SET**, Spurweite **H0**.
3. Wähle eine passende Kategorie, beispielsweise **Personenwagen**, und die Anzahl **2**.
4. Wähle **Weiter zu Artikeldaten**. Prüfe einen Artikelvorschlag oder gehe ohne Übernahme weiter.
5. Ergänze unter **Fahrzeugdaten** die gemeinsamen Setangaben und die Reiter **Fahrzeug 1** und
   **Fahrzeug 2**. Erfasse passende Gattungen und unterschiedliche Namen beziehungsweise Fahrzeugnummern.
6. Prüfe die Angaben und speichere das Set.

<GuideFigure name="vehicle-set" alt="Erster Schritt der Seterfassung mit gemeinsamen Grunddaten und zwei vorgesehenen Fahrzeugen." caption="Die geführte Erfassung beginnt mit Typ und Grunddaten. Artikeldaten und individuelle Fahrzeuge folgen in eigenen Schritten. Künstliche Demo in v0.2.0." />

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

## 4. Artikelrevisionen vergleichen und übernehmen

Die Setansicht zeigt unter **Artikelrevisionen** die gespeicherten Setangaben und ihre Artikelbasis.
Wähle **Bearbeiten**, um unter **Allgemein** eine Revision oder einen neuen Artikelentwurf zu prüfen.
Markiere nur die Felder, die du jetzt übernehmen möchtest. Individuelle Angaben bleiben zunächst geschützt.
Die Bestätigung speichert ausschließlich den Artikelstand des Sets, ohne Mitgliedsfahrzeuge zu ändern.

Offene Setangaben und der Artikelentwurf bleiben beim Wechsel zu **Upload** erhalten.
Ändert die Bestätigung ein Feld, das du selbst gerade bearbeitest, zeigt der Dialog beide Werte.
Wähle ausdrücklich die gespeicherten Werte oder deinen Entwurf, bevor du die Setangaben speicherst.
**Abbrechen** verwirft offene Setangaben; bestätigte Artikel- und Bildaktionen bleiben gespeichert.

Nach einem unbekannten Speicherergebnis oder fehlgeschlagenem Nachladen bleibt **Speichern** gesperrt.
**Gespeicherte Ansicht erneut laden** liest den aktuellen Setstand, ohne die Bestätigung zu wiederholen.
Prüfe danach gegebenenfalls den Artikelvergleich erneut.
[Auswahl und Artikelrevisionen](./index#artikelrevisionen-gezielt-übernehmen).

## 5. Zusammengehörige Daten übertragen

Ein CSV-/JSON-Transfer kann die Setstruktur übertragen, enthält aber keine Medien und nicht alle
Fahrzeug-Unterdaten. Für einen vollständigen kompatiblen Umzug verwende eine Anwendungssicherung.
[Exportgrenzen](../import-export/exports) · [Wiederherstellung üben](../../administration/restore-exercise).
