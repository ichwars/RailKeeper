# Artikelübersicht im Bestandslayout

**Datum:** 2026-08-09

**Status:** Fachlich freigegeben

## Ziel

Die bestehende Artikelübersicht wird visuell an die Fahrzeugbestandsseite angeglichen. Bild 2 aus
der Nutzerfreigabe ist die verbindliche Referenz. Es entstehen keine neuen Inhalte, Funktionen,
Filter, Ansichten oder Aktionen.

## Verbindliche Anpassungen

- Die Oberzeile `Werkstatt und Sammlung` entfällt.
- Titel, Beschreibung und bestehende Aktion bleiben in der kompakten Bestandskopfzeile.
- Die vier vorhandenen Artikelkennzahlen verwenden dieselbe Kartenhierarchie, Typografie, Höhe,
  Abstände und aktive Darstellung wie die Kennzahlen des Fahrzeugbestands.
- Der umschließende Artikelkasten entfällt. Überschrift, Trefferzahl, Suche, Filter und Tabelle
  stehen auf der transparenten Bestandsfläche.
- Suche und Filter folgen der zweizeiligen Werkzeuganordnung des Fahrzeugbestands. Die bestehende
  Trefferzahl bleibt erhalten und steht am Ende der Filterzeile.
- Die Artikeltabelle übernimmt Tabellenfläche, Zeilenhöhe, Kopfzeile, Abstände und Aktionsausrichtung
  des Fahrzeugbestands. Spalten, Inhalte, Sortierung und Aktionen bleiben unverändert.

## Technische Grenze

Vorhandene RailKeeper-Komponenten, Design-Tokens und Bestandsklassen werden wiederverwendet. API,
Domainmodell, Rollen, Dialoge, Filterlogik und Übersetzungsinhalte werden nicht erweitert.

## Abnahme

- Die Desktopansicht entspricht in Hierarchie und Dichte der freigegebenen Fahrzeugreferenz.
- Dark und Light Theme bleiben lesbar.
- Die bestehende mobile horizontale Tabellenführung bleibt erhalten.
- Es gibt keinen Dokumentüberlauf und keine neuen Browserwarnungen oder Fehler.
- Alle bisherigen Artikelaktionen, Filter und Zustände bleiben funktionsfähig.
