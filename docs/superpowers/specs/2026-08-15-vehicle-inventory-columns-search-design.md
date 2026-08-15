# Fahrzeugbestand: Spalten und Suche

## Ziel

Die Issues #80 und #81 werden gemeinsam umgesetzt. Benutzer können die sichtbaren Spalten des
Fahrzeugbestands auswählen und anordnen. Die Einstellung wird serverseitig pro Benutzerkonto
gespeichert. Gleichzeitig durchsucht die Bestandssuche zusätzliche Fahrzeugfelder und bietet
Bahngesellschaft, Epoche und Adapter/Schnittstelle als eigene Filter an.

Die Standardansicht bleibt unverändert. Die Umsetzung wird vor einer Veröffentlichung lokal
geprüft.

## Spaltenumfang

Die aktuelle Auswahl bleibt die Standardauswahl und Standardreihenfolge:

1. Bild
2. Inventarnummer
3. Hersteller
4. Artikelnummer
5. Bezeichnung
6. Spur
7. Epoche
8. Ausstellung

Folgende vorhandene kurze Fahrzeugfelder können zusätzlich eingeblendet werden und sind zunächst
ausgeblendet:

- Identifikation und Einordnung: Bahngesellschaft, Kategorie, Gattung, Baureihe, Fahrzeugnummer,
  EAN und Produktionszeitraum
- Digitaltechnik: Digitalstatus, Digital-Decodernummer, D&T-Decoderstatus, D&T-Decodernummer,
  Decoder-Typ, Adapter/Schnittstelle und ABC-Bremse
- Kauf, Lagerung und Zustand: Listenpreis, Erwerbsart, Bezugsquelle, Kaufpreis, Kaufdatum,
  Lagerort, Zustand und Verpackung
- Technische Daten: Länge, Gewicht, Farbe, Beschriftung, Ladung, Inneneinrichtung, Achsen,
  Achsanzahl, Haftreifenanzahl, Radsatz, vordere Kupplung, hintere Kupplung und Stromaufnahme
- Ausstattung: Antrieb, Spitzenlicht, Beleuchtung, Soundgenerator, Rauchgenerator und
  Ausstellungsbereitschaft

Lange Freitexte, Detailbeschreibungen, URLs, interne Zeitstempel und technische Hilfsfelder werden
nicht als Tabellenspalten angeboten.

## Spaltenverwaltung

Ein kompaktes Spaltenmenü ist direkt in der Werkzeugleiste des Fahrzeugbestands erreichbar. Wegen
der Anzahl verfügbarer Felder gliedert es die Auswahl nach fachlichen Bereichen. Sichtbare Spalten
werden zusätzlich in ihrer aktuellen Reihenfolge dargestellt und mit zugänglichen Auf-/Ab-Aktionen
verschoben. Drag-and-drop ist nicht erforderlich.

Die Inventarnummer darf ausgeblendet werden. Ergibt eine Aktion oder eine ungültige gespeicherte
Konfiguration keine sichtbare Datenspalte mehr, wird die Inventarnummer automatisch eingeblendet.
Bild und Ausstellung zählen dabei nicht als alleinige Datenspalten. Die Aktion "Auf Standard
zurücksetzen" stellt Auswahl und Reihenfolge vollständig wieder her.

Neu in späteren Versionen ergänzte Spalten werden nicht an bestehende gespeicherte Konfigurationen
angehängt. Unbekannte, doppelte oder nicht mehr unterstützte Schlüssel werden beim Laden entfernt.

## Speicherung

Die vorhandene API `/api/v1/profile/settings` und die Tabelle `user_settings` werden wiederverwendet.
Die geordnete Spaltenliste wird als JSON unter einem eigenen stabilen Schlüssel gespeichert. Es ist
keine neue Migration und keine installationsweite Einstellung erforderlich.

Die Benutzeroberfläche lädt die Einstellung beim Öffnen des Fahrzeugbestands, normalisiert sie und
verwendet bei fehlender oder unlesbarer Einstellung die Standardauswahl. Änderungen werden
optimistisch dargestellt und anschließend als partielle Profileinstellung gespeichert. Ein
Speicherfehler wird sichtbar gemeldet und überschreibt keine anderen Profileinstellungen.

## Darstellung und Sortierung

Die Desktop-Tabelle rendert Kopf- und Datenzellen aus einer gemeinsamen Spaltendefinition. Kurze
Text-, Datums-, Zahlen- und Ja/Nein-Spalten sind sortierbar. Bild und Ausstellung behalten ihre
speziellen Bedienelemente. Wird die aktive Sortierspalte ausgeblendet, fällt die Sortierung auf die
erste weiterhin sichtbare sortierbare Spalte zurück, bevorzugt auf die Inventarnummer.

Die mobile Liste verwendet dieselbe Auswahl und Reihenfolge, aber eine kompakte Kartenanordnung.
Bild und Aktionen bleiben eigenständige Bereiche. Die ausgewählten Datenfelder erscheinen in der
festgelegten Reihenfolge, ohne horizontalen Dokumentüberlauf. Die bestehende Desktop-Kartenansicht
bleibt als alternative Präsentation erhalten; die Spalteneinstellung steuert die Tabelle und die
mobile Bestandsliste.

## Suche und Filter

Die serverseitige Teilzeichenfolgensuche wird um Baureihe, Fahrzeugnummer und Decoder-Typ
erweitert. Führende und nachgestellte Leerzeichen des Suchbegriffs werden ignoriert. Exakte Treffer
sind durch die Teilzeichenfolgensuche eingeschlossen; Tippfehlertoleranz und eine Mehrwortsyntax
werden nicht ergänzt.

Die vorhandene Filterzeile erhält eigene Auswahlfelder für Bahngesellschaft, Epoche und
Adapter/Schnittstelle. Diese Filter werden wie die bestehenden Bestandsfilter auf die vom Server
gelieferte Ergebnisliste angewendet und lassen sich mit dem Suchbegriff und allen bisherigen Filtern
kombinieren. "Filter zurücksetzen" setzt auch die drei neuen Filter zurück.

Damit alle auswählbaren Spalten und Adapterfilter Werte besitzen, liefert die bestehende
Fahrzeuglisten-API die benötigten kurzen Fahrzeugfelder mit aus. Rollen- und CSRF-Regeln bleiben
unverändert.

## Zubehörbestand

Das bestehende browserlokale Zubehör-Spaltenmenü erhält ausschließlich die Aktion "Auf Standard
zurücksetzen". Sie stellt die vorhandene Standardauswahl wieder her. Reihenfolge, Speicherung und
Spaltenbreiten des Zubehörbestands werden nicht verändert.

## Fehlerbehandlung und Barrierefreiheit

- Ungültige Serverwerte führen zu einer sicheren normalisierten Auswahl statt zu einer leeren oder
  beschädigten Tabelle.
- Fehler beim Laden verwenden die Standardauswahl und zeigen die bestehende Fehlermeldungsfläche.
- Fehler beim Speichern werden gemeldet; die Benutzeraktion bleibt sichtbar, damit sie nicht
  scheinbar verloren geht.
- Das Spaltenmenü lässt sich per Tastatur bedienen, mit Escape schließen und besitzt eindeutige
  deutsche und englische Beschriftungen.
- Ja/Nein-Werte werden textlich dargestellt und nicht ausschließlich über Farbe vermittelt.

## Tests und lokale Abnahme

Backend-Tests decken die zusätzlichen Suchfelder, Teiltreffer, kombinierte Suchergebnisse und die
Benutzertrennung der vorhandenen Profileinstellungen ab. Frontend-Tests decken Normalisierung,
Reihenfolge, vollständiges Ausblenden, unbekannte neue Schlüssel, Reset, Sortierfallback,
serverseitiges Laden/Speichern, neue Filter, Zubehör-Reset und mobile Ausgabe ab.

Vor der lokalen Freigabe werden mindestens `go test ./...`, `npm.cmd run build` und die betroffenen
Frontend-Tests ausgeführt. Danach wird die Fahrzeugübersicht lokal in Deutsch und Englisch, Hell
und Dunkel sowie in Desktop- und Mobilbreite geprüft. Erst nach dieser gemeinsamen Sichtprüfung
wird über Commit, Push, Issue-Abschluss oder Veröffentlichung entschieden.
