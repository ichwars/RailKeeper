# Erweiterte Geometrie, Paket C: Gleis-Höhenprofile

**Datum:** 2026-08-10

**Status:** zur lokalen Umsetzung freigegeben, keine Veröffentlichung

## Ziel

Paket C ergänzt jedes Gleisplanobjekt um eine Anfangs- und Endhöhe in Millimetern. RailKeeper
speichert beide Werte revisionssicher und leitet aus der Höhendifferenz und der Kataloglänge die
Steigung in Prozent ab. Damit entsteht das belastbare Fundament für spätere Durchfahrtshöhen und
Ebenenkollisionen.

## Fachliches Modell

`PlanTrackObject` erhält `elevationStartMm` und `elevationEndMm`. Beide Werte dürfen positiv,
null oder negativ sein, müssen aber endlich sein. Bestehende und aus älteren Backups
wiederhergestellte Gleise starten mit 0 mm an beiden Enden.

Die Steigung wird nicht gespeichert. Die Plananalyse liefert je verwendbarer Gleisgeometrie eine
Zeile mit Objekt-ID, Anfangshöhe, Endhöhe, Länge und Steigung:

`gradePercent = (elevationEndMm - elevationStartMm) / lengthMm * 100`

Eine positive Steigung läuft vom ersten zum zweiten Geometrieanschluss aufwärts. Ein negativer Wert
läuft in dieser Richtung abwärts. Die Reihenfolge ist stabil nach Objekt-ID sortiert.

## Persistenz und Kompatibilität

Migration `0049_track_object_elevations.sql` ergänzt die beiden nicht-nullbaren Spalten mit dem
Standardwert 0. Erstellen, Bearbeiten und Klonen von Planobjekten übernehmen beide Werte. Änderungen
verwenden weiterhin die vorhandene optimistische Objektversion und den bestehenden Audit-Eintrag.

Backup-Format 8 exportiert die Höhenfelder. Backups der Formate 1 bis 7 bleiben importierbar; bei
fehlenden Spalten greifen die Datenbank-Standardwerte.

## API und UI

Die bestehenden Create- und Update-Endpunkte werden um beide Höhenfelder erweitert. Die
Plananalyse enthält zusätzlich `grades`.

Der Gleisinspektor verwendet zwei `AppNumberInput`-Felder und einen expliziten Speichern-Button.
Positions- und Rotationsänderungen senden die vorhandenen Höhen unverändert mit. In allen
Revisionszuständen zeigt der Inspektor die berechnete Steigung. Nur Entwürfe sind bearbeitbar.

## Abgrenzung

Nicht Teil dieses Pakets sind Grenzwerte und Warnungen für zulässige Steigungen, durchgehende
Höhenprofile über mehrere Gleise, Durchfahrtshöhen, Ebenenkollisionen, Flexgleisgeometrien und
automatische Geländemodellierung.

## Abnahme

- Höhen bleiben nach Erstellen, Bearbeiten, Klonen und Backup-Wiederherstellung erhalten.
- Nicht endliche Werte werden serverseitig abgewiesen.
- Die Plananalyse berechnet positive, negative und ebene Steigungen deterministisch.
- Ziehen und Drehen eines Gleises verändern dessen Höhen nicht.
- Der Entwurfsinspektor bearbeitet beide Höhen mit App-Komponenten und zeigt die Steigung.
- Veröffentlichte Revisionen zeigen das Höhenprofil nur lesend.
- Backendtests, Frontendtests und Produktionsbuild laufen erfolgreich.
- Die lokale Browserprüfung bestätigt Eingabe, Speicherung und Anzeige.

