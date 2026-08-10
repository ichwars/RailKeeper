# Erweiterte Geometrie, Paket H: echte Übergangsbögen

**Datum:** 2026-08-10

**Status:** unter der bestehenden lokalen Vollfreigabe für Issues 31 bis 36, keine Veröffentlichung

## Ziel

Paket H ergänzt Flexobjekte um mathematisch echte Euler-Spiralen. Ihre Krümmung wächst entlang der
Bogenlänge linear von null bis zum Kehrwert des gewählten Endradius. RailKeeper nennt diese Form
deshalb ausdrücklich Übergangsbogen und grenzt sie von den freien kubischen Bézier-Verläufen aus
Paket G ab.

Ein Übergangsbogen ist weiterhin genau ein physisches TILLIG-83125-Flexgleisstück. Er verändert
keine Nachbargleise automatisch und löst keine vollständige Gleislücke.

## Gewählte Architektur

`PlanTrackObject` erhält neben `flexPath` einen optionalen `transitionPath`. Beide Formen schließen
sich gegenseitig aus und sind nur für Geometriedefinitionen der Art `flex` zulässig. Dadurch bleibt
das vorhandene Flexpfad-Schema Version 1 unverändert und bestehende Daten behalten ihre Bedeutung.

```text
schemaVersion: 1
lengthMm
endRadiusMm
direction: left | right
```

Die effektive Geometrie bleibt der gemeinsame Ausgang für Darstellung, Anschlüsse, Snapping,
Kollision, Höhen, Steigung, Ebenenabstand, Revisionsvergleich und Stückliste. Das bestehende
`EffectiveGeometryForObject` priorisiert genau eine gespeicherte Pfadform und lehnt Objekte mit
beiden Formen als ungültig ab.

## Mathematisches Modell

Für die Bogenlänge `s` zwischen null und `L` gilt mit Vorzeichen `d` für links oder rechts:

```text
k(s) = d · s / (R · L)
theta(s) = d · s² / (2 · R · L)
x(s) = Integral cos(theta(u)) du, u=0..s
y(s) = Integral sin(theta(u)) du, u=0..s
```

Damit startet die Kurve am lokalen Ursprung tangential entlang der positiven X-Achse und endet mit
Radius `R`. Die Endrichtung wird aus `theta(L)` abgeleitet. Die Koordinatenintegrale werden
deterministisch mit Simpson-Integration ausgewertet. Die Polylinie verwendet höchstens 5 mm lange
Segmente und verschärft die Schrittweite anhand der maximalen Krümmung so, dass die geometrische
Sehnenabweichung höchstens 0,05 mm beträgt. Oberhalb 4.096 Segmenten ist die Eingabe ungültig.

Die effektive Länge ist exakt `L`, der kleinste Radius exakt `R`. Alle Werte müssen endlich sein;
Länge und Radius müssen größer als null sein.

## Produkt- und Anlagenregeln

Die verfügbare Länge bleibt 664 mm. Eine Vorschau oberhalb dieser Länge ist sichtbar, aber nicht
übernehmbar. Der wirksame Radiusgrenzwert bleibt das Maximum aus der TILLIG-Empfehlung 543 mm und
`layout.minimumFlexRadiusMm`.

Ein Endradius unter dem Grenzwert erzeugt dieselbe präzise `flex_radius_below_limit`-Warnung wie ein
freier Flexverlauf. Er bleibt nach sichtbarer Warnung bewusst übernehmbar. Dadurch gelten Analyse,
Veröffentlichung und Materiallogik für beide Pfadformen identisch.

## Persistenz und Kompatibilität

Migration `0053_transition_curve_paths.sql` ergänzt
`plan_track_objects.transition_path_json`. Bestehende Zeilen erhalten `NULL`; `flex_path_json` wird
nicht verändert. Repository, Revisionsklon und Scanner persistieren beziehungsweise hydrieren beide
Felder und erzwingen ihre Exklusivität.

Die Backup-Version steigt von 11 auf 12. Backups bis Version 11 erhalten beim Restore ein
`transition_path_json = NULL`. Version-12-Backups bewahren Übergangsbögen vollständig. Die
optimistische Objektversionierung und der bestehende Audit-Schreibpfad bleiben unverändert.

## API

Planobjekt- und Update-Schemas erhalten `transitionPath`. Der neue CSRF-geschützte Planner-Endpunkt

```text
POST /api/v1/plan-track-objects/{id}/transition-preview
```

erwartet Länge, Endradius, Richtung und `expectedVersion`. Er liefert Pfad, effektive Geometrie,
Länge, Radius, Grenzwert, `lengthExceeded`, `radiusBelowLimit` und `applicable`. Die Vorschau schreibt
nichts. Erst das vorhandene Objekt-Update speichert den bestätigten Pfad und entfernt dabei den
anderen Pfadtyp.

Nicht ableitbare Geometrie, eine starre Geometriedefinition, ungültige Richtung und veraltete Version
werden als bestehende Planner-Validierungs-, Not-found- oder Konfliktantworten behandelt.

## Oberfläche

Der Inspector bietet für Flexobjekte zwei app-eigene Aktionen: „Freier Flexverlauf“ und
„Übergangsbogen“. Der neue Dialog verwendet Portal, Fokusfalle, Escape, `AppNumberInput`,
`AppSelect`, bestehende Modal-Tokens und eine SVG-Vorschau. Er zeigt Länge, Endradius, Endrichtung,
Grenzwert sowie Textwarnungen.

„Verlauf vorschlagen“ fordert die maßgebliche Servergeometrie an. „Übergangsbogen übernehmen“ ist
nur bei `applicable` aktiv. Die Übernahme ersetzt einen vorhandenen freien Flexpfad bewusst. Der
freie Flexeditor kann umgekehrt einen Übergangsbogen durch einen geraden Bézier-Ausgangspfad
ersetzen. Abbrechen schreibt nichts.

Deutsch und Englisch werden gemeinsam gepflegt. Persistierte Darstellung und Inspectorwerte kommen
immer aus der effektiven Servergeometrie.

## Fehler- und Grenzfälle

- Beide Pfadtypen gleichzeitig sind ungültig.
- Ein Pfadtyp an einem starren Gleis ist ungültig.
- Länge, Radius oder Richtung außerhalb des Schemas werden abgelehnt.
- Nicht ableitbare Vorschauen liefern keinen partiellen Geometrievertrag.
- Überlänge blockiert die Übernahme, Radiusunterschreitung nicht.
- Höhen- und Lageänderungen erhalten den aktiven Pfadtyp.
- Revisionsklon, Diff, Backup und Restore erhalten beziehungsweise erkennen Übergangsbögen.
- Ein Übergangsbogen zählt weiterhin als genau ein 83125 in der Stückliste.

## Abnahme

- Eine linke und rechte Euler-Spirale besitzen spiegelbildliche Punkte und Endrichtungen.
- Die Krümmung wächst monoton von null bis `1/R`; Länge und Endradius entsprechen der Eingabe.
- Vorschau und Übernahme sind getrennt, Abbrechen verändert nichts.
- 664 mm sind übernehmbar, größere Längen nicht.
- Ein Radius unter 700 mm erzeugt bei entsprechendem Anlagenlimit genau eine Warnung.
- Speichern, Neuladen, Revisionsklon und Backup/Restore erhalten den Pfad.
- Bestehende Bézier-Flexpfade und starre G1-Pläne bleiben unverändert.
- Backend, Frontend, OpenAPI, Produktionsbuild und Browserablauf bestehen.
- Alles bleibt lokal auf `dev/issue-36-advanced-geometry`; kein Push, PR oder Merge.

## Abgrenzung

Nicht Teil von Paket H sind automatische S-Kurven aus zwei Übergangsbögen, konstante Kreisbogenstücke,
vollständiges Lückenrouting, Nachbaränderungen, Flexweichen, Verschnitt und freie dekorative
Planobjekte. Letztere folgen separat als Paket I.
