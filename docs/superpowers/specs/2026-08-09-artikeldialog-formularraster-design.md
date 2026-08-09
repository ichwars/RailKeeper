# Artikeldialog: Formularraster und Checkbox-Ausrichtung

**Status:** Zur schriftlichen Prüfung
**Datum:** 2026-08-09
**Bereich:** Artikelübersicht, Dialog „Artikel bearbeiten“

## Ziel

Die Formulare in den Reitern `Bestand` und `Fachangaben` erhalten eine ruhige, konsistente
Anordnung. Die fachlichen Abläufe, Validierung, Datenbindung und vorhandenen Komponenten bleiben
unverändert.

## Freigegebene Anordnung

### Bestandskorrektur und Umbuchung

Die beiden Formulare bleiben nebeneinander. Beide Spalten erhalten dieselbe Breite und Höhe.
Überschriften beginnen auf derselben Rasterzeile, Feldabstände sind identisch und beide primären
Schaltflächen schließen ihre Spalte auf derselben unteren Linie ab. Pflichtmarkierungen bleiben
Bestandteil des jeweiligen Labels und dürfen nicht in eine eigene Zeile umbrechen.

Unterhalb von 920 Pixeln werden die Formulare wie bisher untereinander dargestellt. Die
Schaltfläche folgt dann direkt auf die zugehörigen Felder.

### Reservieren und Einbauen

Die Eingabeformulare erhalten auf Desktop ein Zweispaltenraster. Zusammengehörige Felder fließen
zeilenweise von links nach rechts. Technische Platzierungsfelder sind Teil desselben Rasters und
bilden keinen einspaltigen Unterblock mehr. Überschriften, Statuszusammenfassungen,
Formularaktionen und fachlich breite Hinweise dürfen beide Spalten belegen.

Die vorhandenen bedingten Felder für Reservierung, Mengen-, Einzelstück- und Hybridbestand bleiben
erhalten. Das Raster muss auch bei ein- oder ausgeblendeten Feldern ohne Leerzeilen oder versetzte
Breiten funktionieren. Unterhalb von 920 Pixeln wechseln beide Formulare auf eine Spalte.

### Boolean-Fachangaben

Boolean-Felder verwenden dieselbe vertikale Struktur wie Text- und Auswahlfelder: Label oben,
darunter eine Kontrollzeile in Eingabefeldhöhe. Checkbox und Hinweistext werden innerhalb dieser
Kontrollzeile vertikal zentriert. Dadurch liegen Checkboxen wie `Bettung` und `Digitaltauglich`
auf derselben Grundhöhe wie die benachbarten Eingabefelder.

## Technische Grenzen

- Bestehende RailKeeper-Komponenten, Design-Tokens und Breakpoints werden wiederverwendet.
- Keine Änderung an API, Datenmodell, Validierung, Feldreihenfolge oder Übersetzungen.
- Keine neue UI-Abhängigkeit und kein neues allgemeines Rastersystem.
- Änderungen bleiben in den betroffenen Zubehörformularen, ihren Tests und
  `frontend/src/styles/accessories.css`.

## Abnahmekriterien

- Bestandskorrektur und Umbuchung sind auf Desktop gleich breit und unten bündig.
- Reservierungs- und Einbauformulare zeigen ihre Felder auf Desktop in zwei gleich breiten Spalten.
- Auf schmalen Ansichten sind alle betroffenen Formulare einspaltig und ohne horizontalen
  Dokumentüberlauf bedienbar.
- Boolean-Felder sind mit benachbarten Eingabefeldern vertikal ausgerichtet.
- Pflichtmarkierungen stehen direkt im Label.
- Dark und Light Theme zeigen keine Überlagerung, abgeschnittene Beschriftung oder Scrollfalle.
- Bestehende Reservierungs-, Einbau-, Umbuchungs- und Fachangaben-Tests bleiben grün; ergänzte
  Strukturtests sichern die neuen Klassen und responsiven Regeln.

## Nicht-Ziele

- Keine fachliche Umordnung oder Umbenennung von Feldern.
- Keine Änderung der Dialoggröße oder der festen Aktionsleiste.
- Keine Neugestaltung der Bestands-, Reservierungs- oder Einbauprozesse.
