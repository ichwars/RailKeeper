# Artikelverwaltung, Redesign

**Datum:** 2026-08-08

**Status:** Approved follow-up in implementation

**Geltungsbereich:** Artikelverwaltung und zugehörige Einstellungen

## Ziel

RailKeeper ersetzt den bisherigen technisch gegliederten Zubehörbereich durch eine artikelzentrierte
Verwaltung. Anwender suchen, erfassen und pflegen Modellbahnartikel in einer gemeinsamen Übersicht.
Bestand, Käufe, Dokumente, konkrete Einzelstücke und Verwendungen bleiben Bestandteile eines Artikels
und erscheinen nicht mehr als gleichrangige Hauptseiten.

Das Redesign soll wie ein vorhandener RailKeeper-Arbeitsbereich wirken: dicht, ruhig, verlässlich und
auf tägliche Pflege ausgelegt. Die Fahrzeugverwaltung ist das interne Referenzmuster für Seitentitel,
Werkzeugleiste, Tabelle, Dialoge, Typografie, Abstände und transparente Iconaktionen.

Diese Spezifikation ersetzt die UI-spezifischen Zubehörfestlegungen in
`2026-08-07-anlagen-zubehoer-gleisplaner-design.md`. Das vorhandene Backendfundament wird
weiterverwendet und fachlich erweitert. Anlagenverwaltung und Gleisplaner sind nicht Bestandteil dieses
Redesigns.

## Leitprinzipien

1. Der Artikel ist der zentrale Einstiegspunkt.
2. Globale Stammdaten und artikelbezogene Daten werden nicht in derselben Navigation vermischt.
3. Die Oberfläche zeigt nur Felder, die zur gewählten Artikelart passen.
4. Seltene Angaben bleiben erreichbar, dominieren aber nicht die Grundansicht.
5. Lagerbestand, Einzelstücke und Verwendungen werden nachvollziehbar miteinander verbunden.
6. Historie entsteht aus realen Vorgängen und ist keine primäre Arbeitsseite.
7. RailKeeper verwendet app-eigene Bedienelemente statt uneinheitlicher Browserkontrollen.

## Begriffe

- **Artikel:** Gemeinsamer Herstellerartikel mit Artikelnummer, Bezeichnung, Produktdaten und
  Dokumenten.
- **Mengenbestand:** Austauschbare Einheiten eines Artikels an einem oder mehreren Lagerorten.
- **Einzelstück:** Konkretes Exemplar mit eigener Identität, beispielsweise Inventarnummer,
  Seriennummer, Zustand, Garantie oder individuellem Einbauverlauf.
- **Kaufvorgang:** Dokumentierter Zugang mit Bezugsquelle, Preis, Datum, Menge und Belegen.
- **Reservierung:** Vorgemerkte Menge oder vorgemerktes Einzelstück für ein Ziel. Sie verändert die
  Verfügbarkeit, aber nicht den physischen Lagerbestand.
- **Verwendung:** Konkrete Installation in einer Anlage, einem Modul, einem Fahrzeug oder später einem
  Planobjekt.
- **Fachreiter:** Dynamischer Dialogreiter mit technischen Feldern der gewählten Artikelart.

Der Begriff `Einzelobjekt` wird in der Benutzeroberfläche vollständig durch `Einzelstück` ersetzt.

## Navigation

Die Hauptnavigation verwendet folgende Bezeichnungen:

- `Bestand` wird zu `Fahrzeugbestand`.
- `Zubehör` wird zu `Artikelübersicht`.

Innerhalb der Artikelübersicht gibt es keine Untermenüs für Produkte, Lagerorte, Bestand,
Einzelstücke, Reservierungen oder Einbauhistorie.

Lagerorte werden ausschließlich unter `Einstellungen → Artikelverwaltung → Lagerorte` angelegt,
bearbeitet, hierarchisch sortiert, archiviert oder reaktiviert. Im Artikeldialog und in Filtern werden
Lagerorte nur ausgewählt.

Unter `Einstellungen → Artikelverwaltung` liegen außerdem:

- Hersteller,
- Bestandseinheiten,
- Artikelarten und Unterarten,
- optionale kontrollierte Zusatzfelder.

## Artikelübersicht

Die Seite folgt dem Aufbau des Fahrzeugbestands.

### Kopfbereich

Der Kopfbereich enthält:

- Titel `Artikelübersicht`,
- Untertitel `Modellbahnartikel suchen, erfassen und pflegen`,
- primäre Aktion `+ Neuer Artikel` oben rechts.

Viewer und Planner sehen die Aktion nicht als funktionslosen Button. Stattdessen erklärt die Seite an
geeigneter Stelle ihren schreibgeschützten Zugriff. Admin und Editor dürfen neue Artikel anlegen.

### Kennzahlen

Vier kompakte, filterbare Kennzahlen zeigen:

- Anzahl Artikel und Artikelarten,
- frei verfügbare Menge und Anzahl genutzter Lagerorte,
- reservierte und eingebaute Menge,
- Pflegehinweise für fehlende Kerndaten.

Kennzahlen sind globale Werte der Artikelverwaltung. Sie dürfen niemals von einer unsichtbar
ausgewählten Tabellenzeile abhängen.

### Werkzeuge und Filter

Die Seite verwendet ausschließlich eine Tabellenansicht. Eine Kartenansicht und ein Ansichtsumschalter
sind nicht vorgesehen.

Die Werkzeugleiste enthält:

- sofort reagierende Suche nach Hersteller, Artikelnummer, EAN und Bezeichnung,
- Filter für Artikelart, Hersteller, Spurweite, Status und Lagerort,
- sichtbare Ergebnisanzahl,
- Aktion zum Zurücksetzen aktiver Filter.

### Tabelle

Die Grundspalten sind:

| Spalte | Inhalt |
|---|---|
| Artikel | Bezeichnung, Hersteller und Artikelnummer |
| Art / Unterart | Fachliche Einordnung |
| Spur | Eine oder mehrere geeignete Spurweiten |
| Bestand | Gesamt, frei, reserviert und eingebaut in kompakter Form |
| Lagerung | Hauptlagerort oder Anzahl weiterer Lagerorte |
| Aktionen | Ansehen, Bearbeiten und weitere Aktionen |

Alle fachlich sinnvollen Tabellenköpfe sind sortierbar. Der aktive Sortierschlüssel und die Richtung
werden visuell und für assistive Technik ausgezeichnet.

Die Aktionsspalte verwendet transparente Lucide-Iconbuttons nach bestehendem RailKeeper-Muster:

- `Eye` für Ansehen,
- `Pencil` für Bearbeiten,
- `MoreHorizontal` für weitere Aktionen.

Jeder Iconbutton besitzt einen Tooltip, einen zugänglichen Namen und einen sichtbaren Fokuszustand.
Weitere Aktionen enthalten mindestens Archivieren.

## Artikeldialog

`+ Neuer Artikel`, `Ansehen` und `Bearbeiten` öffnen denselben Dialog in unterschiedlichen Modi. Der
Dialog orientiert sich am vorhandenen Fahrzeugdialog und verwendet horizontale Reiter.

### Feste und dynamische Reiter

Vor der ersten Verwendung besitzt der Dialog:

`Artikel | Bestand | Kauf & Dokumente | <Fachreiter>`

Der Fachreiter wird von Artikelart und Unterart bestimmt. Unpassende Fachreiter werden vollständig
ausgeblendet und nicht deaktiviert dargestellt.

Nach der ersten Reservierung oder Installation erscheint zusätzlich:

`Verwendung & Historie`

Der Reiter bleibt nach einem späteren Ausbau sichtbar, weil historische Vorgänge erhalten bleiben.
Reine Bestandsbewegungen erscheinen im Reiter `Bestand` und erzeugen keinen Verwendungsreiter.

### Reiter Artikel

Die direkt sichtbaren Felder sind:

- Produktbild,
- Hersteller,
- Artikelnummer,
- Bezeichnung,
- EAN oder GTIN,
- Herstellerstatus,
- Artikelart und Unterart,
- Spurweite und Maßstab,
- Verpackungseinheit und Bestandseinheit,
- Beschreibung.

`Weitere Angaben` ist ein eingeklappter Bereich für seltene allgemeine Produktdaten, beispielsweise:

- alternative Artikelnummern,
- Erscheinungsjahr oder Produktionszeitraum,
- Hersteller- und Produktlink,
- Schlagwörter,
- allgemeine Kompatibilitätshinweise,
- interne Notizen.

Anlagen-, Fahrzeug- oder einbauspezifische Werte gehören ausdrücklich nicht in `Weitere Angaben`.

### Reiter Bestand

Der Reiter enthält:

- Strategie der Bestandsführung,
- Mindestbestand,
- Bestände je Lagerort,
- Bestandszugang, Korrektur und Umbuchung,
- gespeicherte Einzelstücke,
- Umwandlung einer Mengeneinheit in ein Einzelstück,
- kompaktes Bestandsjournal.

Lagerorte werden hier ausgewählt, aber nicht administriert.

### Reiter Kauf & Dokumente

Der Reiter enthält zwei klar getrennte Bereiche.

**Kauf und Beschaffung:**

- Kaufdatum,
- Bezugsquelle,
- Menge und Ziel-Lagerort,
- Einzelpreis und Gesamtpreis,
- Rechnungsnummer,
- Garantiefrist,
- Notiz.

**Dokumente:**

- Rechnung und Lieferschein,
- Anleitung,
- Datenblatt,
- Grundriss,
- weitere Produktbilder,
- sonstige sichere Anhänge.

Ein als bestandswirksam gespeicherter Kauf erzeugt in derselben Transaktion den Kaufvorgang und den
zugehörigen Lagerzugang.

### Reiter Verwendung & Historie

Der obere Bereich zeigt aktuelle Reservierungen und Installationen. Der historische Bereich zeigt
chronologisch:

- Reservierung,
- bestätigten Einbau,
- Zustandsänderungen,
- Ausbau und dessen Ergebnis.

Eine konkrete Verwendung speichert:

- Zielart und konkretes Ziel,
- Menge oder Einzelstück,
- genauer Einbauort,
- Reservierungs-, Einbau- und Ausbaudatum,
- Zustand,
- Digitaladresse, Decoder-Ausgang oder Anschluss, sofern relevant,
- Verkabelungs-, Einbau- und Ausbauhinweise.

Dasselbe Produkt kann dadurch mehrfach mit unterschiedlichen Einbauorten und technischen
Konfigurationen verwendet werden.

## Dynamische Artikelarten

Die erste erweiterbare Fachsystematik enthält acht Gruppen.

### Gleis

Unterarten umfassen Gerade, Bogen, Flexgleis, Weiche, Kreuzung, Doppelkreuzungsweiche, Übergang und
Prellbock.

Fachwerte umfassen Gleissystem, Länge, Radius, Winkel, Richtung, Herzstückwinkel, Schwellenart,
Profilhöhe, Bettung, Anzahl Anschlüsse und Digitaltauglichkeit.

### Signal

Unterarten umfassen Licht-, Form-, Haupt-, Vor-, Block-, Einfahr-, Ausfahr- und Sperrsignal.

Fachwerte umfassen Vorbild, Epoche, Signalbegriffe, Anzahl LEDs, Bauhöhe, Betriebsspannung AC und DC,
Montageart, Antrieb, integrierten Decoder und passendes Steuermodul.

### Decoder

Unterarten umfassen Lok-, Funktions-, Zubehör-, Schalt-, Servo- und Rückmeldedecoder.

Fachwerte umfassen Schnittstelle, Protokolle, Funktionsausgänge, Motorstrom, Ausgangsstrom,
Summenstrom, RailCom, SUSI, Servoausgänge, Abmessungen und Firmware.

Decoder-CVs bleiben am Fahrzeug oder an der konkreten Verwendung. Sie sind keine allgemeinen
Artikeleigenschaften.

### Elektrik & Steuerung

Unterarten umfassen Weichenantrieb, Rückmelder, Booster, Netzteil, Sensor, Relais, Verteiler und
Bedienelement.

Fachwerte umfassen Ein- und Ausgangsspannung, Strom, Leistung, Anzahl Kanäle, Protokolle,
Anschlussarten, Schutzfunktionen und kompatible Artikel.

### Gebäude & Ausstattung

Unterarten umfassen Gebäude, Bahnsteig, Brücke, Tunnelportal, Straßenfahrzeug, Figur, Straßen- und
Innenausstattung.

Fachwerte umfassen Epoche, Abmessungen, Grundfläche, Material, Bausatz oder Fertigmodell, Teilezahl,
Schwierigkeitsgrad, Beleuchtungsmöglichkeiten und Grundriss.

### Landschaft & Verbrauch

Unterarten umfassen Gras, Streu, Baum, Wasser, Farbe, Klebstoff, Schotter, Draht, Kabel und
Befestigungsmaterial.

Fachwerte umfassen Material, Farbe, Jahreszeit, Inhalt, Einheit, Faser- oder Korngröße, Reichweite,
geeignete Maßstäbe und Sicherheitshinweise.

### Beleuchtung

Unterarten umfassen Leuchte, LED, Lichtleiste, Gebäude- und Effektbeleuchtung.

Fachwerte umfassen Lichtfarbe, Farbtemperatur, Spannung, Strom, AC oder DC, Anzahl LEDs, Dimmung,
Abmessungen und Montageart.

### Sonstiger Artikel

Seltene Artikel verwenden kontrollierte Zusatzfelder mit einem definierten Datentyp:

- Text,
- Zahl mit Einheit,
- Ja oder Nein,
- Datum,
- Einfach- oder Mehrfachauswahl.

Unstrukturierte JSON- oder Freitextsammlungen sind nicht die primäre Darstellung. Fachwerte müssen
filterbar, validierbar und exportierbar bleiben.

## Fachliches Datenmodell

### Gemeinsamer Artikelkern

Der Artikel speichert gemeinsame Hersteller- und Produktdaten. Technische Fachwerte werden typisiert
und anhand der Artikelart validiert. Dokumente, Bestandsvorgänge und Verwendungen referenzieren den
Artikel, ohne seine Felder zu kopieren.

Mögliche Dubletten anhand von normalisiertem Hersteller und Artikelnummer erzeugen vor dem Speichern
eine deutliche Warnung. Eine harte Eindeutigkeitsregel ist nicht vorgesehen, weil Neuauflagen oder
bewusst getrennte Varianten möglich bleiben müssen.

### Bestandsstrategien

Ein Artikel verwendet eine von drei Strategien:

1. **Reiner Mengenbestand:** austauschbare Einheiten, beispielsweise Gleise oder Verbrauchsmaterial.
2. **Konkrete Einzelstücke:** von Beginn an individuell geführte Geräte.
3. **Menge mit späterer Individualisierung:** gelagerte Menge, aus der beim Bedarf ein Einzelstück
   erzeugt werden kann.

Die Individualisierung reduziert den Mengenbestand und erzeugt das Einzelstück atomar. Eine Einheit
darf nie gleichzeitig als Menge und Einzelstück gezählt werden.

### Materialfluss

1. Kauf oder Bestandszugang erhöht den physischen Bestand an einem Lagerort.
2. Reservierung reduziert nur die verfügbare Menge.
3. Bestätigter Einbau reduziert Mengenbestand oder ändert den Zustand eines Einzelstücks.
4. Ausbau führt Material zu einem Lagerort, in Wartung, in den Zustand defekt oder in Ausmusterung.
5. Historische Datensätze bleiben erhalten.

Bestandsmengen dürfen nicht still negativ werden. Zusammengehörige Schreibvorgänge laufen in einer
SQLite-Transaktion.

## Komponenten und Bedienung

Die Artikelübersicht wird in fokussierte Frontendkomponenten getrennt:

- Seitenkopf und Kennzahlen,
- Such- und Filterwerkzeuge,
- sortierbare Artikeltabelle,
- Dialogshell und Modussteuerung,
- feste Reiter,
- dynamischer Fachreiter,
- Verwendungs- und Historienansicht,
- Lagerortverwaltung in den Einstellungen.

Die Oberfläche verwendet vorhandene RailKeeper-Komponenten und ergänzt bei Bedarf fokussierte,
app-eigene Eingabekomponenten. Insbesondere gelten:

- `AppSelect` statt nativer Auswahlbox,
- `AppDateInput` statt nativer Browser-Datumsauswahl,
- app-eigene Text-, Zahlen-, Datei- und Mehrfachauswahl-Shells,
- transparente Lucide-Iconbuttons,
- vorhandene Dialog-, Fokus- und Bestätigungsmuster.

Native Browserkontrollen werden nur verwendet, wenn sie innerhalb einer app-eigenen Shell keine
sichtbaren oder interaktiven Inkonsistenzen verursachen.

## Dialogverhalten und Fehlerbehandlung

- Eingaben bleiben beim Reiterwechsel erhalten.
- Speichern validiert den gesamten Dialog.
- Fehlerhafte Reiter erhalten einen sichtbaren Fehlerindikator.
- Pflichtfehler erscheinen zusätzlich direkt am Feld.
- Schließen mit ungespeicherten Änderungen verlangt eine Bestätigung.
- Ein Wechsel der Artikelart warnt, wenn bereits Fachwerte vorhanden sind.
- Ein fehlgeschlagener Schreibvorgang lässt den Dialog geöffnet und erhält alle Eingaben.
- Fehlende Lagerorte verlinken auf die Lagerorteinstellung.
- Leere Übersichten bieten die Aktion `Ersten Artikel anlegen`.
- Archivierte Artikel und ihre historischen Daten bleiben lesbar.

## Berechtigungen

Alle Rechte werden serverseitig geprüft. Schreibzugriffe bleiben CSRF-geschützt und werden auditiert.

| Rolle | Artikel lesen | Artikel pflegen | Bestand und Käufe pflegen | Reservieren | Einbau bestätigen |
|---|---:|---:|---:|---:|---:|
| Admin | Ja | Ja | Ja | Ja | Ja |
| Editor | Ja | Ja | Ja | Ja | Ja |
| Planner | Ja | Nein | Nein | Ja | Nein |
| Viewer | Ja | Nein | Nein | Nein | Nein |
| Messe | Nein | Nein | Nein | Nein | Nein |

Fehlende Rechte werden erklärt. Die Oberfläche darf zentrale Arbeitsmöglichkeiten nicht
kommentarlos verstecken, wenn dadurch der vorhandene Lesemodus unverständlich wird.

## Internationalisierung und Zugänglichkeit

Alle neuen Texte werden auf Deutsch und Englisch bereitgestellt. Die englischen Begriffe lauten
mindestens `Vehicle inventory`, `Article overview`, `Individual items`, `Usage & history` und
`Purchase & documents`.

Erforderlich sind außerdem:

- vollständige Tastaturbedienung,
- Fokusbindung und Fokusrückgabe für Dialoge,
- sichtbare Fokuszustände,
- textliche Feldfehler zusätzlich zu Farbe,
- zugängliche Namen für Iconbuttons,
- semantische Sortierzustände der Tabelle,
- verständliche Auswahlzustände,
- Prüfung langer deutscher Texte,
- Prüfung in hellem und dunklem Design sowie bei Desktop- und Mobilbreiten.

### Lokalisierte Artikelstammdaten und Bereichsreiter

Standardwerte der Artikelverwaltung werden anhand ihres stabilen Schlüssels in der aktiven Sprache
angezeigt. Das gilt für Bestandseinheiten, Artikelarten und die mitgelieferten Unterarten. Technische
Schlüssel bleiben unverändert sichtbar. Benutzerdefinierte Einträge sowie bewusst umbenannte
Standardwerte werden nicht übersetzt, sondern exakt mit ihrer gespeicherten Bezeichnung angezeigt.

Beim Bearbeiten eines noch unveränderten Standardwertes darf die lokalisierte Anzeige nicht als
sprachgebundene Überschreibung gespeichert werden. Bleibt der lokalisierte Entwurf unverändert, wird
weiterhin die kanonische Standardbezeichnung persistiert. Eine bewusste Texteingabe wird dagegen als
benutzerdefinierte Bezeichnung übernommen.

Die Bereiche `Hersteller`, `Bestandseinheiten`, `Artikelarten und Unterarten`, `Kontrollierte
Zusatzfelder` und `Lagerorte` bilden eine flache Tab-Navigation. Sie verwendet `tablist`, `tab` und
`tabpanel`, kennzeichnet den aktiven Reiter mit `aria-selected` und unterstützt Pfeiltasten sowie
`Home` und `End`. Auf schmalen Ansichten bleibt die Reihenfolge horizontal scrollbar, statt die
Reiter als große Schaltflächen umzubrechen.

## Speicherung, API und Backup

Backend, Frontendclient und OpenAPI-Vertrag bleiben synchron. Neue oder geänderte Fachobjekte
benötigen explizite API-Verträge für:

- Artikelstamm und Fachwerte,
- Kaufvorgänge,
- Bestandsbewegungen,
- Einzelstücke und Individualisierung,
- Verwendung und Historie,
- Artikelstammdaten in den Einstellungen.

Artikel, Fachwerte, Käufe, Dokumente, Bestandsjournal, Einzelstücke, Reservierungen und
Verwendungshistorie sind Bestandteil des Anwendungsbackups. Authentifizierungs- und Auditdaten bleiben
gemäß bestehender Sicherheitsregel ausgeschlossen. Restore prüft die Formatversion vor jeder
Veränderung und erhält die Importierbarkeit bisher unterstützter Backups.

## Teststrategie

Erforderlich sind:

- Go-Tests für Bestandsstrategien, Käufe, Individualisierung, Reservierung, Einbau und Ausbau,
- Transaktions- und Nebenläufigkeitstests gegen negative oder doppelt gezählte Bestände,
- Migrationstests,
- Backup- und Restore-Roundtrips,
- API-Tests für Rollenmatrix, Validierung und CSRF,
- Frontendtests für Suche, Filter, Sortierung, Dialogmodi und dynamische Reiter,
- Frontendtests für Rechtehinweise, Fehlerindikatoren und ungespeicherte Änderungen,
- Abgleich von Backend, Frontendclient und OpenAPI,
- vollständige Go-Test-Suite und Produktionsbuild des Frontends,
- visuelle Prüfung der freigegebenen Übersicht und Dialoge.

## Abnahmekriterien

Das Redesign ist abgenommen, wenn:

1. Die Hauptnavigation `Fahrzeugbestand` und `Artikelübersicht` verwendet.
2. Die Artikelübersicht keine Zubehör-Untermenüs und keine Kartenansicht enthält.
3. Ein Editor über `+ Neuer Artikel` den freigegebenen Dialog öffnet.
4. Der Dialog nur den zur Artikelart passenden Fachreiter zeigt.
5. Ein Tillig-TT-Gleis mit Geometrie- und Bestandsdaten vollständig erfasst werden kann.
6. Ein Signal, Decoder, Gebäude und Landschaftsartikel jeweils nur passende Fachfelder erhalten.
7. Mengenbestand, Einzelstücke und die spätere Individualisierung korrekt funktionieren.
8. Lagerorte ausschließlich in den Einstellungen administriert werden.
9. Kaufvorgänge bei Bedarf atomar einen Lagerzugang erzeugen.
10. `Verwendung & Historie` erst nach Reservierung oder Installation erscheint und erhalten bleibt.
11. Viewer und Planner verständliche Lesemodi erhalten, ohne unzulässige Schreibaktionen.
12. Deutsch, Englisch, Hell, Dunkel, Tastaturbedienung und lange Texte geprüft sind.
13. Backup und Restore sämtliche neuen Fachdaten verlustfrei erhalten.

## Recherchegrundlage

Die Fachfelder wurden anhand realer Herstellerartikel abgeleitet, unter anderem:

- [Tillig TT Modellgleis 83101](https://www.tillig.com/Produkte/produktinfo-83101.html),
- [Tillig TT Bettungsweiche 83816](https://www.tillig.com/Produkte/produktinfo-83816.html),
- [Viessmann Modelltechnik Katalog 2024 bis 2026](https://viessmann-modell.com/media/ed/9d/db/1711099322/Viessmann_Katalog_2024-2025-2026.pdf),
- [ESU LokPilot 5](https://www.esu.eu/en/products/lokpilot/lokpilot-5/),
- [FALLER Bahnhof Mittelstadt](https://www.faller.de/miniaturwelten/rund-um-die-bahn/bahnhoefe/84/bahnhof-mittelstadt),
- [NOCH Streugras Sommerwiese](https://www.noch.de/streugras-sommerwiese/50190/).
