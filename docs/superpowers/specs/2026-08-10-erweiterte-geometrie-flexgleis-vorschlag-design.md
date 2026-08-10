# Erweiterte Geometrie, Paket G: Flexgleis-Grundmodell und Verlaufsvorschlag

**Datum:** 2026-08-10

**Status:** lokal zur Umsetzung freigegeben, keine Veröffentlichung

## Ziel

Paket G ergänzt den Gleisplaner um ein reales Flexgleisprodukt und einen ausdrücklich bestätigten
Vorschlagsworkflow für glatte, objektbezogene Gleisverläufe. RailKeeper berechnet einen Verlauf aus
Endpunkt, Endrichtung und zwei Tangentenlängen, zeigt Länge und kleinsten Radius an und übernimmt ihn
erst nach einer Benutzerentscheidung.

Die Funktion ist ein Planungswerkzeug. Sie verlegt kein Gleis automatisch, verändert keine anderen
Planobjekte und behauptet keine physische Genauigkeit jenseits der dokumentierten Produktdaten und
Berechnungstoleranzen.

## Produktbasis

Die erste Flexgeometrie ist TILLIG TT-Modellgleis 83125 „Flexgleis Holzschwelle“. TILLIG nennt eine
Länge von ungefähr 664 mm, Nenngröße TT, Maßstab 1:120 und Spurweite 12 mm:
`https://www.tillig.com/Produkte/produktinfo-83125.html`.

TILLIG empfiehlt in den Gleisbauhinweisen, Flexgleise bei Radien über R543 und als Ausgleichsstücke
zu verwenden: `https://www.tillig.com/gleisplanung2.html`. RailKeeper behandelt 543 mm deshalb als
dokumentierten empfohlenen Mindestradius, nicht als garantierte mechanische Bruchgrenze.

Die Geometriedefinition führt 664 mm als maximal nutzbare Länge eines Planobjekts und 543 mm als
empfohlenen Mindestradius. Die Verpackungseinheit von 20 Stück verändert den Planbedarf nicht. Ein
Planobjekt entspricht einem physischen Flexgleisstück.

## Bewertete Ansätze

### Gewählt: kubische Bézier-Kurve mit serverseitiger Ableitung

Eine kubische Bézier-Kurve bildet einen glatten Verlauf mit kontrollierten Tangenten an beiden Enden.
Gespeichert werden nur die fachlichen Parameter; Route, Länge, Anschlüsse und kleinster Radius werden
deterministisch serverseitig abgeleitet. Das Modell ist kompakt, unterstützt auch S-förmige Verläufe
und fügt sich in die vorhandenen Anschluss-, Höhen- und Kollisionsprüfungen ein.

### Nicht gewählt: freie Polylinie

Eine frei editierbare Punktfolge wäre einfach zu speichern, erlaubt jedoch ungewollte Knicke und
besitzt ohne zusätzliche Glättungsregeln keine belastbare Endtangente oder Radienprüfung. Polylinien
bleiben das interne, abgeleitete Analyseformat, nicht das Benutzermodell.

### Nicht gewählt: vollständiger Klothoidenlöser

Echte Euler-Spiralen bilden vorbildgerechte Übergangsbögen genauer ab. Ein stabiler Löser mit
Richtungs-, Längen- und Radiusbedingungen benötigt jedoch ein eigenständiges Paket und zusätzliche
Interaktionsregeln. Paket G schafft dafür die Anschluss- und Vorschlagsarchitektur, nennt die
Bézier-Kurve aber ausdrücklich nicht Klothoide oder echten Übergangsbogen.

## Domänenmodell

`TrackGeometryKind` erhält `flex`. `TrackGeometryDefinition` erhält das optionale Feld
`minimumRadiusMm`; bei 83125 beträgt `lengthMm` 664 und `minimumRadiusMm` 543. Starre Geometrien
behalten einen leeren Mindestradius.

`PlanTrackObject` erhält optional `flexPath`:

```text
schemaVersion: 1
endXMm, endYMm
endDirectionDegrees
startHandleMm, endHandleMm
```

Der lokale Startpunkt ist immer `P0 = (0, 0)`. Die Fahrtrichtung am Start zeigt entlang der lokalen
positiven X-Achse; der äußere Anschluss A zeigt entsprechend nach 180 Grad. `P3` ist der gespeicherte
Endpunkt. `P1` liegt `startHandleMm` entlang der positiven X-Achse. `P2` liegt
`endHandleMm` entgegen der gespeicherten Endrichtung vor `P3`. Der äußere Anschluss B verwendet
`endDirectionDegrees`.

Ein Flexpfad ist nur für eine Geometriedefinition der Art `flex` zulässig. Eine Flexgeometrie benötigt
immer einen Flexpfad. Alle Werte müssen endlich sein, beide Tangentenlängen müssen größer als 0 sein
und Start und Ende dürfen nicht zusammenfallen. Die Endrichtung wird auf 0 bis unter 360 Grad
normalisiert.

Ein neu platziertes 83125 beginnt als gerader Verlauf mit 664 mm Länge, Endpunkt `(664, 0)`,
Endrichtung 0 Grad und zwei Tangentenlängen von jeweils `664 / 3` mm.

## Effektive Geometrie

Die Bibliotheksgeometrie bleibt unveränderlich. Für jedes Planobjekt liefert die Domäne zusätzlich
eine effektive Geometrie und eine effektive Länge. Bei starren Objekten entsprechen sie unverändert
der Bibliotheksdefinition. Bei Flexobjekten werden sie aus `flexPath` erzeugt.

Die Bézier-Kurve wird adaptiv in eine Polylinie zerlegt. Jedes Teilstück ist höchstens 5 mm lang;
zusätzlich darf die Abweichung zwischen Kurve und Sehne höchstens 0,05 mm betragen. Start und Ende
werden exakt übernommen. Eine feste Obergrenze von 4.096 Teilstücken verhindert pathologische
Eingaben. Kann die Toleranz innerhalb dieser Grenze nicht erreicht werden, ist der Flexpfad ungültig.

Die effektive Länge ist die Summe der abgeleiteten Teilstücke. Der kleinste Radius wird aus der
Kurvenkrümmung an den adaptiv erzeugten Parameterstellen bestimmt. Gerade Abschnitte besitzen einen
unendlichen Radius und begrenzen das Ergebnis nicht. API und Oberfläche kennzeichnen Länge und Radius
als berechnete Werte und zeigen sie mit zwei Nachkommastellen.

Alle vorhandenen Auswertungen verwenden anschließend ausschließlich effektive Route, Anschlüsse und
Länge:

- Anschluss- und Snappingprüfung,
- offene Enden und inkompatible Anschlüsse,
- geometrische Überlappungen,
- Höhenkontinuität und Steigungsberechnung,
- Ebenenabstand an Kreuzungen,
- Revisionsvergleich und Änderungsvorschau.

## Anlagenlimit und Warnungen

Anlagen erhalten das optionale Feld `minimumFlexRadiusMm`. Zulässig sind endliche Werte größer als
0. Der wirksame Grenzwert ist das Maximum aus dem Produktwert und dem Anlagenwert. Ohne Anlagenwert
gilt für 83125 damit die dokumentierte Empfehlung von 543 mm; ein Anlagenwert kann sie verschärfen,
aber nicht unterschreiten.

Ein persistierter Verlauf unterhalb des wirksamen Radiuslimits erzeugt die Warnung
`flex_radius_below_limit`. Sie enthält tatsächlichen Radius und wirksamen Grenzwert, verändert den
Verlauf nicht automatisch und blockiert die Veröffentlichung nicht.

Ein Verlauf mit mehr als 664 mm effektiver Länge kann in der Vorschau erscheinen, wird aber nicht
übernommen. Die Vorschau liefert `lengthExceeded`; die Übernahme bleibt deaktiviert. Für längere
Strecken sind mehrere Flexobjekte erforderlich. Dadurch bleibt die bestehende Stückliste korrekt:
ein Objekt benötigt genau ein Stück 83125.

Verschnitt, Reststücke und das Zusammenfassen mehrerer kurzer Planobjekte auf ein physisches
Flexgleis sind nicht Teil dieses Pakets. Eine spätere Zuschnittoptimierung benötigt ein eigenes
Reststück- und Reservierungsmodell.

## Verlaufsvorschlag

Der Benutzer legt Endpunkt und Endrichtung fest. Optional können die beiden Tangentenlängen direkt
geändert werden. „Verlauf vorschlagen“ ruft eine serverseitige, rein lesende Berechnung auf.

Für den automatischen Vorschlag durchsucht RailKeeper deterministisch positive Tangentenlängen in
einem begrenzten Grob-zu-fein-Raster. Kandidaten über 664 mm werden verworfen. Die verbleibenden
Kandidaten werden lexikografisch bewertet:

1. kleinste Unterschreitung des wirksamen Radiuslimits,
2. kleinste Änderung der Krümmung entlang des Verlaufs,
3. kürzeste effektive Länge,
4. kleinere Start- und danach Endtangente als stabiler Gleichstand.

Die Antwort enthält die vorgeschlagenen Parameter, effektive Route, Länge, kleinsten Radius,
Radiuslimit sowie `lengthExceeded` und `radiusBelowLimit`. Gibt es keinen Kandidaten innerhalb der
Produktlänge, bleibt die Übernahme deaktiviert. Ein Kandidat unterhalb des Radiuslimits darf nach
sichtbarer Warnung bewusst übernommen werden.

Der Vorschlag ist nicht persistent. Erst „Verlauf übernehmen“ schreibt den ausgewählten Flexpfad mit
optimistischer Objektversionierung und Audit-Eintrag. Ein Konflikt lädt wie bei bestehenden
Planobjekten den aktuellen Stand und verlangt eine neue bewusste Entscheidung.

## Oberfläche

Die Bibliothek ergänzt „TILLIG 83125 · Flexgleis Holzschwelle“. Die bestehende Palette zeigt Länge
und Art. Nach der Platzierung rendert der Planer die effektive Route statt einer statischen Linie.

Der Gleisinspektor zeigt für Flexobjekte zusätzlich tatsächliche Länge, kleinsten Radius und
Radiusgrenze. „Flexgleis bearbeiten“ öffnet einen app-eigenen Dialog. Er enthält app-eigene
Zahlenfelder für Endpunkt X/Y, Endrichtung sowie Start- und Endtangente, eine kompakte Vorschau und
die Aktionen „Verlauf vorschlagen“, „Abbrechen“ und „Verlauf übernehmen“.

Der Endpunkt kann auf der Arbeitsfläche gezogen werden. In der Nähe eines freien oder kompatiblen
Gleisports übernimmt die lokale Vorschau dessen Position und Gegenrichtung; die endgültige
serverseitige Berechnung bleibt maßgeblich. Die normale Objektbewegung verschiebt weiterhin den
gesamten Flexverlauf und kann den Startanschluss einrasten lassen.

Eine gestrichelte Vorschau unterscheidet ungespeicherte Geometrie vom gespeicherten Verlauf. Länge
und Radius besitzen Text und Symbol; Warnungen werden nicht allein farblich dargestellt. Deutsch und
Englisch werden gemeinsam gepflegt. Der Dialog bleibt auf Desktop dicht und ist bei schmalen
Ansichten vertikal bedienbar.

## Persistenz, Backup und API

Migration `0052_flex_track_paths.sql`:

- erweitert die zulässigen Geometriearten um `flex` durch einen deterministischen Tabellenumbau,
- ergänzt `track_geometry_definitions.minimum_radius_mm`,
- ergänzt `plan_track_objects.flex_path_json`,
- ergänzt `layouts.minimum_flex_radius_mm`,
- legt die verifizierte Geometriedefinition für TILLIG 83125 an.

Der Tabellenumbau erhält IDs, Fremdschlüssel, Indizes und die bestehende G1-Definition unverändert.
Ein eigener Migrationstest aktualisiert eine Datenbank bis 0051 mit einem bestehenden Planobjekt und
prüft anschließend Fremdschlüssel, G1 sowie 83125.

Die Backup-Version steigt von 10 auf 11. Version 11 enthält die neuen Spalten über den generischen
Tabellenexport. Backups bis Version 10 stellen `minimum_radius_mm`, `flex_path_json` und
`minimum_flex_radius_mm` als `NULL` wieder her. Bestehende Pläne erhalten keinen erfundenen
Flexverlauf.

Frontend, Backend und OpenAPI führen synchron:

- `minimumFlexRadiusMm` am Layout,
- `minimumRadiusMm` an der Geometriedefinition,
- `flexPath`, `effectiveGeometry`, `effectiveLengthMm` und `effectiveMinimumRadiusMm` am Planobjekt,
- Vorschlagseingabe und Vorschlagergebnis,
- `flex_radius_below_limit` mit Radiusdetails.

Der neue CSRF-geschützte Planner-Endpunkt erzeugt eine Vorschau für ein bestehendes Flexobjekt. Das
vorhandene Objekt-Update persistiert den bestätigten `flexPath`; es gibt keinen zweiten Schreibpfad.
Rollen- und Revisionsgrenzen bleiben unverändert.

Beim Klonen einer Revision wird `flex_path_json` unverändert kopiert. Der Revisionsvergleich bewertet
Änderungen am Flexpfad als Objektänderung. Reservierung und Stückliste bleiben objektbezogen.

## Fehler- und Grenzfälle

- Nicht endliche Werte, nicht positive Tangenten oder identische Endpunkte werden abgelehnt.
- Ein Flexpfad an einer starren Geometrie und ein fehlender Pfad an einer Flexgeometrie werden
  abgelehnt.
- Überschreitet die adaptive Ableitung 4.096 Teilstücke, ist die Vorschau ungültig.
- Überlange Verläufe können nicht übernommen werden.
- Radien unter dem Grenzwert bleiben bewusst übernehmbare Warnungen.
- Ein gerader Flexpfad besitzt keinen begrenzenden Radius und erzeugt keine Radiuswarnung.
- Snapping darf nur Position und Endrichtung der Vorschau ändern, niemals andere Planobjekte.
- Ein Updatekonflikt überschreibt keinen inzwischen gespeicherten Verlauf.
- Veröffentlichte Revisionen bleiben unveränderlich.
- Beschädigtes Legacy-JSON erzeugt wie andere beschädigte Geometrie einen Fehlerhinweis und keinen
  stillen Ersatzverlauf.

## Abgrenzung

Nicht Teil von Paket G sind echte Klothoiden, mathematisch garantierte Übergangsbögen,
Verschnittoptimierung, Reststückverwaltung, gemeinsame Materialreservierung mehrerer Flexobjekte,
automatisches Verbinden ganzer Gleislücken, automatische Änderung benachbarter Gleise,
Flexweichen, freie Planobjekte, weitere Flexgleisartikel und digitale Anlagensteuerung.

Echte Übergangsbögen folgen als Paket H auf dem stabilen Vorschlags- und Effektivgeometriemodell.
Freie Planobjekte bleiben Paket I.

## Abnahme

- TILLIG 83125 erscheint als verifizierte Flexgeometrie mit 664 mm und 543 mm Empfehlung.
- Ein gerades Flexobjekt wird platziert, gespeichert, neu geladen, geklont und im Backup
  wiederhergestellt.
- Ein Vorschlag zwischen definiertem Endpunkt und Endrichtung liefert deterministisch dieselben
  Parameter, dieselbe Route, Länge und denselben kleinsten Radius.
- Der Benutzer sieht die Vorschau und übernimmt sie ausdrücklich; Abbrechen verändert das Objekt
  nicht.
- Ein Verlauf über 664 mm kann nicht übernommen werden.
- Ein Verlauf unterhalb des Radiuslimits erzeugt genau eine präzise Warnung, bleibt aber bewusst
  übernehmbar.
- Ein Anlagenlimit oberhalb 543 mm wird wirksam; ein kleinerer Wert senkt die Produktempfehlung nicht.
- Verbindungen, Snapping, Überlappung, Höhenkontinuität, Steigung und Ebenenabstand verwenden die
  effektive Flexroute.
- Die Stückliste zählt jedes Flexobjekt als ein Stück 83125.
- Backendtests, Frontendtests, Produktionsbuild und lokale Browserprüfung laufen erfolgreich.
- Alles bleibt auf `dev/issue-36-advanced-geometry`; kein Push, PR oder Merge.
