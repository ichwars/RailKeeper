# Bestandsformular: Kompaktraster für Korrektur und Umbuchung

**Status:** Implemented and verified
**Datum:** 2026-08-09
**Bereich:** Artikelübersicht, Dialog „Artikel bearbeiten“, Reiter „Bestand“
**Visuelle Auswahl:** Variante B, vom Nutzer am 2026-08-09 bestätigt

## Ziel

Bestandskorrektur und Umbuchung bleiben auf Desktop nebeneinander, verwenden den verfügbaren Raum
aber entsprechend ihrer Feldanzahl. Das kürzere Korrekturformular darf keine künstlich gestreckten
Abstände oder Leerflächen mehr enthalten.

## Freigegebenes Desktop-Layout

Der gemeinsame Bestandsbereich verwendet ein gewichtetes Zweispaltenraster:

- Die Bestandskorrektur belegt ungefähr ein Drittel der verfügbaren Breite.
- Die Umbuchung belegt ungefähr zwei Drittel der verfügbaren Breite.
- Beide Überschriften beginnen auf derselben Höhe.
- Beide Aktionsschaltflächen schließen auf derselben unteren Linie ab.

Die Bestandskorrektur bleibt ein kompaktes einspaltiges Formular mit dieser Reihenfolge:

1. Lagerort
2. Mengenänderung
3. `Bestand buchen`

Die Umbuchung verwendet innerhalb ihrer breiteren Spalte ein Zweispaltenraster:

1. Zeile: Von Lagerort, Nach Lagerort
2. Zeile: Menge, Notizen
3. Zeile: `Umbuchen` über die gesamte Formularbreite

Die beiden Formulare erreichen dadurch dieselbe kompakte Höhe, ohne Felder oder Zwischenräume zu
strecken. Labels, Eingabefelder und Schaltflächen verwenden weiterhin die bestehenden RailKeeper-
Komponenten, Abstände und Design-Tokens.

## Responsive Verhalten

- Unterhalb von 920 Pixeln stehen Bestandskorrektur und Umbuchung untereinander.
- Das interne Zweispaltenraster der Umbuchung bleibt bestehen, solange ausreichend Breite vorhanden
  ist.
- Unterhalb von 560 Pixeln wird auch die Umbuchung einspaltig.
- In gestapelten Ansichten folgt jede Aktionsschaltfläche direkt auf die zugehörigen Felder.
- Es entsteht kein horizontaler Dokumentüberlauf. Der feste Dialogfooter bleibt erreichbar.

## Technische Umsetzung

Die vorhandenen Formulare und ihre Datenbindung bleiben bestehen. Die Umbuchungsfelder erhalten
einen lokalen Rastercontainer. `frontend/src/styles/accessories.css` definiert ausschließlich die
gewichtete äußere Aufteilung, das interne Umbuchungsraster und die beiden bestehenden Breakpoints.

Die vorherige Gleichverteilung und die Flex-Streckung des Korrekturformulars werden entfernt. Es
entsteht kein neues allgemeines Formularsystem und keine neue UI-Abhängigkeit.

## Abnahmekriterien

- Auf Desktop ist die Umbuchung ungefähr doppelt so breit wie die Bestandskorrektur.
- Die Bestandskorrektur zeigt normale, gleichmäßige Feldabstände ohne vertikale Streckung.
- Von/Nach Lagerort sowie Menge/Notizen bilden im Umbuchungsformular jeweils eine Zeile.
- Beide Desktop-Schaltflächen liegen auf derselben Grundlinie.
- Unter 920 Pixeln stehen die Formulare untereinander.
- Unter 560 Pixeln ist auch die Umbuchung einspaltig.
- Dark und Light Theme zeigen keine Überlagerung, abgeschnittene Beschriftung oder Scrollfalle.
- Bestehende Buchungs- und Umbuchungslogik sowie deren Validierung bleiben unverändert.
- Ergänzte Struktur- und Responsive-Tests sichern die lokalen Rasterklassen.

## Verifikation

- Fokustest des `AccessoryStockPanel` und statischer Responsive-Test.
- Vollständige Frontendtests und Produktionsbuild.
- Browser-QA im Dark und Light Theme auf Desktop sowie bei 390 × 844 Pixeln.
- Direkter Bildvergleich mit dem markierten Nutzer-Screenshot und Variante B des visuellen
  Begleiters.

## Nicht-Ziele

- Keine Änderung an API, Datenmodell, Validierung, Übersetzungen oder Feldinhalten.
- Keine Umgestaltung der Bestandstabelle oder des Bestandsjournals.
- Keine Änderung an Reservierungs-, Einbau- oder Fachangabenformularen.
- Keine Änderung der Dialoggröße oder der festen Aktionsleiste.
