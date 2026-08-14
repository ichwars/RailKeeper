# Erweiterte Geometrie, Paket F: Ebenenabstand an Gleiskreuzungen

**Datum:** 2026-08-10

**Status:** lokal umgesetzt und abgenommen, keine Veröffentlichung

## Ziel

Paket F erkennt an geometrischen Gleiskreuzungen einen zu kleinen vertikalen Abstand zwischen zwei
Gleisebenen. Die Anlage erhält dafür einen optionalen Mindestabstand in Millimetern. Eine
Unterschreitung erscheint als präziser, anklickbarer Warnhinweis, verändert aber weder Geometrie noch
Höhenprofil automatisch und blockiert die Veröffentlichung nicht.

Der berechnete Wert ist ausdrücklich der vertikale Abstand zwischen den Schienenoberkanten beider
Planobjekte. Er ersetzt keine physische Durchfahrtshöhenprüfung gegen Brückenunterkanten,
Lichtraumprofile oder konkrete Fahrzeuge. Diese fachlich umfangreichere Prüfung benötigt freie
Bauwerksobjekte und folgt in einem späteren Paket.

## Bewertete Ansätze

### Gewählt: Schnittpunktprüfung der Gleismittellinien

Die vorhandenen zweidimensionalen Routen werden segmentweise auf echte innere Schnittpunkte geprüft.
An jedem Schnittpunkt interpoliert RailKeeper die Höhe beider Gleise und vergleicht deren absoluten
Abstand mit dem Anlagenlimit. Dieser Ansatz nutzt die bereits persistierten Höhenprofile, liefert
reproduzierbare Ergebnisse und bleibt unabhängig von einer noch nicht vorhandenen 3D-Darstellung.

### Nicht gewählt: Dreidimensionale Hüllkörper

Hüllkörper könnten Brücken, Bettungen, Oberleitungen und Lichtraumprofile genauer abbilden. Dafür
fehlen derzeit jedoch belastbare Abmessungen und freie Bauwerksobjekte. Eine pauschale Hülle würde
Scheingenauigkeit erzeugen und den Umfang dieses Pakets deutlich überschreiten.

### Nicht gewählt: Ausschließlich manuelle Messpunkte

Manuelle Messpunkte wären fachlich kontrollierbar, würden vorhandene geometrische und vertikale
Planinformationen aber nicht nutzen. Sie bleiben eine sinnvolle spätere Ergänzung für Bauwerke,
ersetzen jedoch nicht die automatische Ebenenkollisionsprüfung.

## Anlagenkonfiguration

`layouts` erhält das nullable Feld `minimum_track_clearance_mm`. Ein leerer Wert deaktiviert nur die
Ebenenabstandsprüfung. Zulässig sind endliche Werte größer als 0. Der Wert gilt für alle Varianten,
Revisionen und Einheiten der Anlage; Grenzwerte je Plan oder Gleis werden nicht eingeführt.

Anlegen und Bearbeiten einer Anlage verwenden das vorhandene `AppNumberInput` mit der Beschriftung
„Mindestabstand kreuzender Gleise (mm)“. Ein Hilfetext erklärt die Messung zwischen
Schienenoberkanten und die Deaktivierung durch Leeren des Feldes. Das Anlagenprofil zeigt den
lokalisierten Wert oder „Nicht festgelegt“.

## Analysierbare Geometrien

Die automatische Prüfung wird nur ausgeführt, wenn beide Planobjekte eine eindeutig
interpolierbare Geometrie besitzen:

- genau zwei geordnete Anschlüsse,
- genau eine zusammenhängende Route,
- mindestens zwei Routenpunkte,
- Routenanfang und -ende entsprechen in dieser oder umgekehrter Reihenfolge den beiden Anschlüssen,
- eine positive Gesamtlänge der Route.

Die Zuordnung von Routenenden zu Anschlüssen verwendet die vorhandene Geometrietoleranz von 0,25 mm.
Verläuft die Route entgegen der Anschlussreihenfolge, wird die Höheninterpolation entsprechend
umgedreht. Mehrport-, Mehrfachrouten- und nicht eindeutig zuordenbare Geometrien werden in diesem
Paket ohne spekulative Auswertung übersprungen.

## Schnittpunkte und Höheninterpolation

RailKeeper vergleicht die transformierten Routensegmente verschiedener Planobjekte. Ein
analysierbarer Kreuzungspunkt muss mehr als 0,25 mm von den Endpunkten beider Segmente entfernt
liegen. Gemeinsame Endpunkte und reguläre Gleisverbindungen werden dadurch nicht als
Ebenenkollision bewertet. Kollineare Segmente bleiben Bestandteil der vorhandenen
Überlappungsprüfung und erzeugen keinen zusätzlichen Ebenenabstandshinweis.

Für jedes Objekt wird die Strecke vom Routenanfang bis zum Schnittpunkt durch die gesamte
Routenlänge geteilt. Mit diesem Anteil wird linear zwischen `elevationStartMm` und
`elevationEndMm` interpoliert. Bei umgekehrter Routenausrichtung werden Anfangs- und Endhöhe vor der
Interpolation vertauscht. Der tatsächliche Ebenenabstand ist der Betrag der Differenz beider
interpolierter Höhen.

Kreuzen sich dasselbe Objektpaar mehrfach, entsteht höchstens ein Warnhinweis. Maßgeblich ist der
Schnittpunkt mit dem kleinsten Ebenenabstand; bei Gleichstand entscheiden die Koordinaten
deterministisch.

## Warnhinweis und Toleranz

Wenn `clearanceMm < minimumTrackClearanceMm` gilt, entsteht der Warnhinweis
`insufficient_clearance`. Der exakte Grenzwert ist zulässig. Eine kleine numerische Toleranz
verhindert instabile Fließkomma-Vergleiche, besitzt aber keine fachliche Bedeutung.

Der Hinweis enthält beide Objekt-IDs, den tatsächlichen Abstand, den Anlagen-Grenzwert und den
zweidimensionalen Schnittpunkt. Er hat den Schweregrad `warning`. Die Änderungsvorschau analysiert
Basis- und Arbeitsrevision mit dem aktuell für die Anlage konfigurierten Grenzwert, analog zur
Steigungsgrenze. Der Wert wird nicht historisch in Revisionen dupliziert.

## Persistenz, Backup und API

Migration `0051_layout_minimum_track_clearance.sql` ergänzt die nullable SQLite-Spalte mit einer
Check-Constraint für positive Werte. Layout-Modell, Create- und Update-Eingaben, Repository,
Handler, Frontend-Typen und OpenAPI führen `minimumTrackClearanceMm` synchron.

Die Backup-Version steigt von 9 auf 10. Neue Exporte enthalten den Wert über den generischen
Tabellenexport. Backups bis Version 9 enthalten die Spalte nicht und stellen Anlagen mit `NULL`
wieder her. Ein alter Export erhält dadurch keinen erfundenen Standardwert.

`TrackPlanIssue` erhält die optionalen Felder `clearanceMm`, `clearanceLimitMm`, `intersectionXMm`
und `intersectionYMm`. Der Warncode wird ebenfalls in der Änderungsvorschau und im OpenAPI-Vertrag
geführt.

## Oberfläche

Die Planprüfung ergänzt einen Zähler für Abstandsunterschreitungen. Der lokalisierte Detailtext
lautet beispielsweise „Ebenenabstand 25,00 mm unterschreitet Grenzwert 40,00 mm“. Beim Anklicken
fokussiert der Planer das erste betroffene Gleis; beide Objekt-IDs bleiben im Warnhinweis enthalten.
Deutsch und Englisch werden gemeinsam gepflegt.

Der Planer bietet keine automatische Höhenkorrektur an. Der Benutzer entscheidet, welches Gleis
angehoben, abgesenkt oder geometrisch verschoben wird, und pflegt dessen Höhenprofil mit den
vorhandenen app-eigenen Eingabefeldern.

## Fehler- und Grenzfälle

- Ein leerer Anlagenwert erzeugt keine Ebenenabstandshinweise.
- Werte kleiner oder gleich 0, NaN und unendliche Werte werden serverseitig abgelehnt.
- Ein Abstand exakt am Grenzwert erzeugt keinen Hinweis.
- Negative und ansteigende Höhenprofile werden linear und vorzeichenrichtig interpoliert.
- Reguläre Anschlusskontakte und kollineare Überlappungen erzeugen keinen zusätzlichen Hinweis.
- Mehrport-, Mehrfachrouten-, längenlose und nicht eindeutig zuordenbare Geometrien werden nicht
  spekulativ bewertet.
- Mehrere Kreuzungen desselben Objektpaars erzeugen nur den Hinweis für den kleinsten Abstand.
- Die Sortierung der Hinweise und die Auswahl bei Gleichstand bleiben deterministisch.

## Abgrenzung

Nicht Teil dieses Pakets sind Brückenunterkanten, Gleiskörperdicken, Oberleitung,
fahrzeugabhängige Lichtraumprofile, freie Bauwerksobjekte, dreidimensionale Darstellung,
automatische Höhenkorrektur, Flexgleisoptimierung und Übergangsbögen.

## Abnahme

- Eine Anlage kann einen positiven Mindestabstand speichern und wieder entfernen.
- Alte Backups werden ohne Grenzwert wiederhergestellt; neue Backups erhalten Version 10.
- Zwei eindeutig interpolierbare Gleise mit innerem Routenschnittpunkt erzeugen unterhalb der
  konfigurierten Grenze genau einen präzisen Warnhinweis.
- Der exakte Grenzwert, ein nicht gesetzter Grenzwert, ein Anschlusskontakt und eine kollineare
  Überlappung erzeugen keinen Ebenenabstandshinweis.
- Ansteigende und abfallende Profile liefern am Schnittpunkt die korrekte interpolierte Höhe.
- Mehrdeutige Geometrien werden übersprungen, ohne bestehende Warnungen zu verändern.
- Planprüfung und Änderungsvorschau verwenden dieselbe aktuelle Anlagenkonfiguration.
- Oberfläche, API und OpenAPI zeigen tatsächlichen Abstand, Grenzwert und betroffene Objekte
  konsistent; Anklicken fokussiert das erste Gleis.
- Backendtests, Frontendtests, Produktionsbuild und lokale Browserprüfung laufen erfolgreich.

## Lokales Abnahmeprotokoll

- Branch: `dev/issue-36-advanced-geometry`, ausschließlich lokal, kein Push, PR oder Merge.
- Backend: `go test ./...` erfolgreich, 7 Pakete.
- Frontend: `npm.cmd test -- --run` erfolgreich, 66 Testdateien und 356 Tests.
- Produktionsbuild: `npm.cmd run build` erfolgreich, 2.173 Module transformiert.
- Anlagenprofil: Mindestabstand von 40,00 mm gespeichert und nach Neuladen weiterhin sichtbar.
- Browserprüfung: Zwei Tillig-G1 kreuzen sich rechtwinklig. Bei 25,00 mm Ebenenabstand und
  40,00 mm Grenzwert erscheint genau eine Abstandsunterschreitung mit den korrekten Werten.
- Hinweisnavigation: Der Warnhinweis fokussiert das betroffene Gleis im Gleisinspektor.
- Korrektur: Nach Anheben des oberen Gleises auf exakt 40,00 mm verschwindet der Hinweis und bleibt
  nach Neuladen entfernt.
- Laufzeit: lokaler Server auf `127.0.0.1:18083`, Prozess-ID 2032, `/health` antwortet mit HTTP 200.
- Sitzung: als `codex-test` angemeldet, keine Warnungen oder Fehler in der Browserkonsole.
