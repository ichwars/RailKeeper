# Fahrzeug-Betriebsdaten und Bestandswerte

## Ziel

Die Issues #78 und #86 werden gemeinsam umgesetzt. Fahrzeuge erhalten eine optionale
Vorbild-Höchstgeschwindigkeit und eine optionale Heimat-Bw- beziehungsweise Einsatzstellenangabe.
Das Dashboard zeigt Fahrzeug- und Zubehörwerte getrennt sowie centgenau an.

Issue #82 und der gesamte Updateweg bleiben ausdrücklich außerhalb dieses Vorhabens.

## Fachliche Entscheidungen für Fahrzeugdaten

Die Höchstgeschwindigkeit bezeichnet die Vorbild-Höchstgeschwindigkeit. Die Einheit ist fest
`km/h`; eine Modellgeschwindigkeit oder benutzerdefinierte Betriebsgrenze wird nicht ergänzt. Der
Wert ist optional, ganzzahlig und liegt bei gepflegten Werten zwischen 1 und 1000.

Die Heimatangabe ist ein optionaler Freitext mit der deutschen Bezeichnung
"Heimat-Bw / Einsatzstelle" und der englischen Bezeichnung "Home depot / operating location".
Der Wert wird getrimmt und auf 200 Zeichen begrenzt. Stammdaten, Zeiträume und historische
Zuordnungen sind nicht Bestandteil der ersten Umsetzung.

Die API verwendet die stabilen Felder `maximumSpeedKmh` als optionale Ganzzahl und `homeBase` als
optionalen String. Zwei getrennte vorwärtsgerichtete Datenbankspalten erhalten sichere Leerwerte,
damit bestehende Fahrzeuge ohne Datenänderung gültig bleiben.

## Fahrzeugformular und Darstellung

Beide Angaben liegen im Reiter "Modell" direkt unter Baureihe und Fahrzeugnummer. Auf breiten
Ansichten entsteht ein zusammenhängender Zweizeiler:

1. Baureihe | Fahrzeug-Nr.
2. Höchstgeschwindigkeit | Heimat-Bw / Einsatzstelle

Auf schmalen Ansichten stehen die Eingaben untereinander. Die Höchstgeschwindigkeit zeigt die
feste Einheit `km/h`. Die Heimatangabe wird nicht mit dem physischen Lagerort im Bereich
"Fahrzeug" vermischt. Ebenso liegen beide Werte weder in "Details" noch in "Steuerung" oder
"Fahrkurve".

Die schreibgeschützte Fahrzeugansicht und der Detailreport erhalten eine gemeinsame Gruppe
"Vorbild & Betrieb". Beide Angaben werden außerdem in Fahrzeug-CSV-Import und -Export sowie in die
serverseitige Bestandssuche aufgenommen. Die konfigurierbare Fahrzeugtabelle bietet beide Felder
als neue Spalten an; gemäß der bestehenden Spaltenregeln sind später ergänzte Spalten zunächst
ausgeblendet.

## Zubehör-Listenpreis

Zubehörartikel erhalten einen optionalen "Listenpreis pro Stück". Der Wert ist eine nichtnegative
EUR-Dezimalangabe mit höchstens zwei Nachkommastellen. Er gehört zu den Artikelstammdaten, nicht zu
einem einzelnen Einkauf oder individualisierten Exemplar. Artikelansicht, API, OpenAPI,
Sicherungen und vorhandene strukturierte Artikel-Importe und -Exporte führen das Feld verlustfrei
mit.

Die Zubehörtabelle bietet den Listenpreis als neue optionale Spalte an. Er ist für bestehende
Spaltenkonfigurationen zunächst ausgeblendet. Bestehende Artikel erhalten keinen angenommenen oder
automatisch abgeleiteten Preis.

## Bewertungsmodell

Ein eigener Endpunkt `GET /api/v1/overview/valuation` liefert genau vier Dezimalstrings mit zwei
Nachkommastellen:

- `vehicleListValue`: Summe aller gepflegten Fahrzeug-Listenpreise
- `vehiclePurchaseValue`: Summe aller gepflegten Fahrzeug-Kaufpreise
- `accessoryListValue`: Listenpreis pro Stück multipliziert mit der aktuell vorhandenen Menge
- `accessoryPurchaseCost`: historische erfasste Kaufkosten des Zubehörs

Die aktuell vorhandene Zubehörmenge umfasst den nicht ausgebuchten mengen- und
einzelstückgeführten Bestand. Ein archivierter Artikel mit weiterhin vorhandenem Bestand bleibt in
der Bestandsbewertung enthalten. Ausgebuchte beziehungsweise ausgemusterte Einzelstücke erhöhen
den aktuellen Listenwert nicht.

Die Zubehör-Kaufkosten summieren für jeden erfassten Einkauf `Menge × Stückpreis`. Zusätzlich
werden Kaufpreise manuell erfasster Einzelstücke ohne verknüpften Einkauf berücksichtigt. Ein mit
einem Einkauf verknüpftes Einzelstück wird nicht noch einmal summiert. Die Kaufkosten sind damit
eine historische Ausgabensumme und ausdrücklich keine Lagerbewertung nach FIFO, LIFO oder
Durchschnittspreis.

Fahrzeug- und Zubehörpreise ohne abweichende Währungsangabe werden als EUR behandelt.
Zubehör-Einkäufe mit einer expliziten anderen Währung werden nicht ohne Wechselkurs in EUR
eingerechnet. Die Antwort enthält zusätzlich `excludedForeignCurrencyPurchases` als Anzahl der
ausgeschlossenen Einkäufe, damit die Benutzeroberfläche bei Bedarf einen transparenten Hinweis
anzeigen kann.

## Geldverarbeitung

Alle Summen und Multiplikationen erfolgen serverseitig in ganzzahligen Cent. Ein zentraler Parser
akzeptiert die in RailKeeper bereits vorkommenden Dezimalschreibweisen wie `129.90`, `129,90`,
`1.299,90` und `1,299.90`. Leere Werte zählen als nicht gepflegt. Ungültige oder negative Werte
werden nicht summiert und führen nicht zu Gleitkomma-Rundungsfehlern.

Der Endpunkt liefert normalisierte Dezimalstrings wie `1252.00`. Erst die Benutzeroberfläche
formatiert sie abhängig von der Sprache als beispielsweise `1.252,00 €` oder `€1,252.00`, immer
mit genau zwei Nachkommastellen.

## Backend-Architektur und Berechtigungen

Ein fokussierter Bewertungsdienst koordiniert die vier Summen. Die Persistenzschicht aggregiert
Fahrzeuge, Zubehörbestand, Einkäufe und unverknüpfte Einzelstücke direkt in SQLite. Es entstehen
keine Einzelabfragen pro Zubehörartikel und keine Abhängigkeit von Listenpaginierung oder aktiven
UI-Filtern.

Der HTTP-Handler bleibt dünn und gibt ausschließlich das typisierte Bewertungsobjekt aus. Der
Endpunkt ist für dieselben regulären Rollen erreichbar, die die Übersicht und Bestandsdaten sehen
dürfen. Der Messebetrieb erhält keinen zusätzlichen Zugriff auf Kauf- oder Listenpreise.

## Dashboard, Variante B

Die bestehende Kopfzeile bleibt kompakt. Die bisherige Kachel "Erfasster Listenwert" wird zu einem
breiteren Block "Erfasste Bestandswerte". Gesamtbestand, Digitalisierung und Wartung behalten ihre
Positionen und ihr bisheriges Verhalten.

Auf Desktopbreiten verwendet die Kennzahlenzeile fünf gleich breite Rasterspalten: Gesamtbestand
und Digitalisierung belegen je eine Spalte, der Werteblock zwei Spalten und Wartung die fünfte
Spalte. Damit bleiben alle vier Bereiche in genau einer Zeile. Bis 900 px darf das bestehende
zweispaltige Tablet-Raster umbrechen; bis 640 px stehen die Bereiche einspaltig untereinander.
Inhalt, Kartenhöhe und innere 2×2-Matrix des Werteblocks bleiben unverändert.

Im Werteblock stehen die vier Summen als 2×2-Matrix:

1. Fahrzeuge · Listenwert
2. Fahrzeuge · Kaufpreis
3. Zubehör · Listenwert
4. Zubehör · Kaufkosten

Auf kleinen Breiten wechselt die Matrix zunächst in zwei Spalten und bei Bedarf in eine Spalte,
ohne horizontalen Dokumentüberlauf. Die Beträge werden nie auf ganze Euro gerundet. Ein Hinweis zu
ausgeschlossenen Fremdwährungseinkäufen erscheint nur, wenn der Server einen Wert größer null
meldet.

Der Werteblock wird als fokussierte Komponente aus der bereits großen `OverviewView` ausgelagert.
Die Bewertungsabfrage läuft unabhängig von der bestehenden Fahrzeugabfrage. Schlägt sie fehl,
bleiben alle übrigen Dashboard-Inhalte nutzbar; der Werteblock zeigt einen verständlichen Fehler
und keine erfundenen Nullsummen.

## API, Sicherung und Kompatibilität

Frontend-Typen, Backend-Typen und `openapi/railkeeper.yaml` werden gemeinsam geändert. Die beiden
Fahrzeugfelder und der Zubehör-Listenpreis werden in den bestehenden Lese- und Schreibpfaden
mitgeführt. App-Sicherungen enthalten die neuen Spalten, während ältere Sicherungen und Exporte
weiterhin importierbar bleiben.

Die Migrationen sind vorwärtsgerichtet und ändern keine historischen Migrationen. Bestehende
Preisstrings werden nicht massenhaft umgeschrieben. Die Bewertungslogik normalisiert sie nur beim
Lesen für die Aggregation.

## Fehlerbehandlung und Validierung

- Eine Höchstgeschwindigkeit außerhalb von 1 bis 1000 wird client- und serverseitig abgelehnt.
- Eine Heimatangabe über 200 Zeichen wird serverseitig abgelehnt und im Formular verständlich
  markiert.
- Zubehör-Listenpreise dürfen nicht negativ sein und höchstens zwei Nachkommastellen besitzen.
- Leere optionale Felder bleiben gültig.
- Ungültige historische Preisstrings brechen das Dashboard nicht ab und werden nicht summiert.
- Fremdwährungen werden sichtbar ausgeschlossen statt stillschweigend als EUR behandelt.
- Deutsch und Englisch, Hell und Dunkel sowie Tastaturbedienung werden vollständig berücksichtigt.

## Tests und lokale Abnahme

Die Umsetzung folgt Test-Driven Development. Backend-Tests decken Migrationen, Fahrzeugvalidierung,
Lese- und Schreibpfade, Suche, CSV, Backup-Kompatibilität, Centparser, Mengenmultiplikation,
unverknüpfte Einzelstücke, Doppelzählung und Fremdwährungen ab. API-Vertragstests prüfen die neuen
Felder und den Bewertungsendpunkt.

Frontend-Tests decken Formularposition und Validierung, schreibgeschützte Darstellung, neue
standardmäßig ausgeblendete Spalten, Zubehör-Listenpreis, sprachabhängige Centformatierung,
Werteblock, Fehlerzustand und responsive CSS-Regeln ab.

Vor einer Veröffentlichung werden mindestens `go test ./...`, `npm.cmd run test:run` und
`npm.cmd run build` ausgeführt. Danach werden Fahrzeugformular, Fahrzeugansicht,
Zubehörartikel und Dashboard lokal in Deutsch und Englisch, Hell und Dunkel sowie in Desktop- und
Mobilbreite geprüft. Erst nach dieser Abnahme wird über Push, PR, Issue-Abschluss oder
Veröffentlichung entschieden.
