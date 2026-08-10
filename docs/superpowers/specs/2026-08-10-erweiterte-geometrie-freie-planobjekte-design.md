# Erweiterte Geometrie, Paket I: freie Planobjekte

**Datum:** 2026-08-10

**Status:** lokal implementiert und am 10. August 2026 vollständig abgenommen, keine Veröffentlichung

Die Nachweise liegen unter
`docs/aegis/work/2026-08-10-advanced-geometry-free-plan-objects/`.

## Ziel

Paket I ergänzt den maßhaltigen Gleisplan um freie, rein planerische Objekte. Nutzer können Flächen,
Ellipsen, Linien und Beschriftungen für Bahnsteige, Gebäudeumrisse, Anlagenkanten, Landschaft und
Hinweise positionieren, drehen, bearbeiten und revisionssicher speichern.

Freie Planobjekte sind ausdrücklich keine Gleise. Sie erzeugen keine Anschlüsse, offenen Enden,
Kollisionen, Höhenprofile, Steigungen, Ebenenabstände, Stücklistenpositionen oder Reservierungen.

## Gewählte Architektur

Eine eigene Tabelle `plan_free_objects` hält freie Objekte getrennt von `plan_track_objects`.
Dadurch bleiben die Invarianten realer Gleise unverändert: jedes Gleis referenziert weiterhin eine
geprüfte Geometriedefinition, und alle Gleisanalysen konsumieren ausschließlich Trackobjekte.

Zwei verworfene Alternativen sind:

- Nullable Geometrie-IDs in `plan_track_objects`: würde Gleis-, Material- und Anschlusslogik an
  vielen Stellen aufweichen und Fehler durch unvollständige Trackobjekte begünstigen.
- Nur frontendseitige Annotationen: wären nicht revisions-, backup- oder mehrsitzungssicher.

`TrackPlan` liefert künftig zwei explizite Listen: `objects` für Gleise und `freeObjects` für freie
Objekte. Separate Repository-Methoden und fokussierte Handler halten CRUD-Logik aus den bereits
großen Trackdateien heraus.

## Datenmodell

Jedes freie Objekt besitzt:

```text
id, lineageId, revisionId
name
category: structure | platform | scenery | annotation
positionXMm, positionYMm, rotationDegrees
shape
version, createdAt, updatedAt
```

`shape` ist ein schema-versioniertes, diskriminiertes JSON-Objekt:

```text
rectangle: schemaVersion=1, kind=rectangle, widthMm, heightMm
ellipse:   schemaVersion=1, kind=ellipse, widthMm, heightMm
line:      schemaVersion=1, kind=line, endXMm, endYMm
label:     schemaVersion=1, kind=label, text, fontSizeMm
```

Name und Beschriftung werden getrimmt. Namen enthalten 1 bis 80 Zeichen, Labeltexte 1 bis 120.
Breite, Höhe und Schriftgröße müssen endlich und größer als null sein; die Schriftgröße liegt
zwischen 2 und 50 mm. Eine Linie benötigt einen endlichen, vom Ursprung verschiedenen Endpunkt.
Positionen verwenden dieselbe endliche Koordinatengrenze wie Gleisobjekte, Drehungen werden auf
0 bis unter 360 Grad normalisiert.

Die Kategorie steuert ausschließlich die bestehende tokenbasierte Darstellung. Es gibt keinen
freien Farbwähler und keine hart codierten Nutzerfarben. So bleibt die Oberfläche in Hell-/Dunkel-
Modus und bei Export-Screenshots konsistent.

## Persistenz, Revisionen und Backup

Migration `0054_free_plan_objects.sql` erstellt die neue Tabelle mit Fremdschlüssel auf
`plan_revisions`, eindeutiger ID, Lineage-ID, Objektversion und Zeitstempeln. `shape_json` bleibt der
einzige formabhängige Speicher und wird beim Lesen strikt validiert.

Beim Erstellen einer neuen Revision werden freie Objekte mit neuer ID, gleicher Lineage-ID und
Version 1 geklont. Der Änderungsvorschauvertrag erhält `freeObjectChanges` mit `added`, `removed`
und `changed`. Ein Vergleich berücksichtigt Name, Kategorie, Position, Drehung und alle Shape-Felder
mit der etablierten Gleitkommatoleranz.

Die Backup-Version steigt von 12 auf 13. `plan_free_objects` wird exportiert und wiederhergestellt;
Dokumente bis Version 12 erhalten beim Normalisieren eine leere Tabelle. Alte Backups verändern
keine bestehenden Gleis- oder Layoutdaten.

## API und Berechtigungen

Der bestehende Trackplan liefert `freeObjects`. Drei CSRF-geschützte Planner-Routen ergänzen CRUD:

```text
POST   /api/v1/plan-revisions/{id}/free-objects
PUT    /api/v1/plan-free-objects/{id}
DELETE /api/v1/plan-free-objects/{id}?expectedVersion=N
```

Admin und Planner dürfen freie Objekte in Entwürfen ändern. Viewer und Editor dürfen sie nur über
den bestehenden Trackplan lesen; Messe bleibt ausgeschlossen. Update und Delete verlangen eine
aktuelle Objektversion. Published- und Review-Revisionen bleiben unveränderlich. Not-found,
Validation, Immutable und Conflict verwenden die bestehenden Problemtypen des Planers.

OpenAPI dokumentiert Objekt, Shape, Eingaben, Change-Eintrag, Listenfeld und alle drei Routen.

## Oberfläche

Der Gleisplaner erhält eine kompakte Aktion „Planobjekt hinzufügen“. Sie öffnet einen app-eigenen
Portal-Dialog mit `AppSelect`, `AppTextInput` und `AppNumberInput`. Formabhängig erscheinen:

- Rechteck und Ellipse: Breite und Höhe,
- Linie: lokaler Endpunkt X/Y,
- Beschriftung: Text und Schriftgröße.

Kategorie, Name, Position und Drehung sind gemeinsame Felder. Beim Anlegen startet das Objekt in der
Mitte der Anlageneinheit. Bestehende freie Objekte lassen sich im SVG auswählen, frei verschieben,
in 15-Grad-Schritten drehen, über denselben Dialog bearbeiten und über die vorhandene
Bestätigungskomponente löschen.

Eine fokussierte `FreePlanObjectLayer` rendert freie Objekte hinter den Gleisen. Kategorien nutzen
CSS-Klassen und vorhandene Design-Tokens. Linien und Konturen bleiben sichtbar, Flächen dezent und
Beschriftungen lesbar. Auswahl erhält zusätzlich zur Farbe eine gestrichelte Kontur. Lange deutsche
Texte, mobile Breiten, Dunkel-/Hellmodus, leere Zustände und Tastaturfokus werden berücksichtigt.

Der Inspector zeigt Shape, Kategorie, Position, Drehung und relevante Maße. Er trennt freie Objekte
klar von Tillig-Gleisen; Gleisaktionen, Geometriequelle und Materialreservierung erscheinen dort
nicht.

## Datenfluss und Fehlerbehandlung

Anlegen, Bearbeiten, Verschieben und Drehen schreiben jeweils genau einmal mit aktueller Version.
Die API-Antwort ersetzt das lokale Objekt. Danach werden Änderungsvorschau und Planprüfung neu
geladen; die Planprüfung bleibt inhaltlich unverändert. Ein 409-Konflikt nutzt dieselbe sichtbare
Neuladen-Aktion wie Trackobjekte.

Formänderungen schreiben nicht automatisch. Abbrechen und Escape schließen den Dialog ohne
Mutation. Ungültige Maße bleiben im Dialog, serverseitige Probleme erscheinen als Text und native
Browserdialoge werden nicht verwendet.

## Tests und Abnahme

Backendtests decken Shape-Validierung, Normalisierung, CRUD, Rollen, CSRF, Draft-Schutz,
Versionskonflikte, Revisionklon, Diff, Migration, Backup/Restore und OpenAPI ab. Gleisanalyse- und
BOM-Regressionstests beweisen, dass freie Objekte keine technischen Ergebnisse verändern.

Frontendtests decken Dialogfelder, Fokus/Escape, explizites Speichern, Canvasdarstellung,
Auswahl, Verschieben, Drehen, Bearbeiten, Löschen, Konflikte und unveränderte Trackpfade ab.

Die Browserabnahme erstellt je ein Rechteck, eine Ellipse, eine Linie und eine Beschriftung,
verschiebt und bearbeitet mindestens ein Objekt, lädt die Seite neu und prüft unveränderte
Geometrien. Die vorhandene Tillig-Stückliste und die Zahl der Gleiswarnungen müssen identisch
bleiben. Eine frische App-Seite darf keine Konsolenfehler melden.

## Abgrenzung

Nicht Teil von Paket I sind freie Polygone, Bézier-Zeichnen, Bildimporte, Layerreihenfolge,
Gruppierung, Kopieren/Einfügen, Maßketten, Drucklayouts, Kollisionsprüfung gegen freie Objekte,
Inventarverknüpfungen und automatische Anlagensteuerung.

Alles bleibt lokal auf `dev/issue-36-advanced-geometry`. Es erfolgt kein Push, Pull Request, Merge
oder Release.
