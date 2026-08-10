# Erweiterte Geometrie, Paket E: Konfigurierbare Steigungsgrenzen

**Datum:** 2026-08-10

**Status:** lokal zur Umsetzung freigegeben, keine Veröffentlichung

## Ziel

Paket E ergänzt pro Anlage eine optionale maximale Steigung in Prozent. Überschreitet ein Gleisstück
diesen Wert, erscheint im Planer ein präziser Warnhinweis. Die Prüfung unterstützt die fachliche
Entscheidung, verändert aber weder Geometrie noch Höhenprofil automatisch und blockiert die
Veröffentlichung nicht.

## Gewählter Ansatz

`layouts` erhält das nullable Feld `max_grade_percent`. Ein leerer Wert deaktiviert ausschließlich
die Steigungsgrenzen-Prüfung; die vorhandene Berechnung und Anzeige der tatsächlichen Steigungen
bleibt aktiv. Der Wert gilt für alle Planvarianten und Einheiten der Anlage. Per-Gleis- und
per-Einheit-Grenzwerte werden nicht eingeführt, damit die Einstellung eindeutig und in diesem Paket
klein bleibt.

Die Anlagenanlage und -bearbeitung verwenden das vorhandene `AppNumberInput`. Zulässig sind endliche
Werte größer als 0 und höchstens 100 Prozent. Die Oberfläche bietet Schritte von 0,1 Prozent an und
erklärt, dass ein leeres Feld die Warnung deaktiviert. Das Anlagenprofil zeigt den lokalisierten Wert
oder „Nicht festgelegt“.

## Analyse

Die Domänenanalyse erhält optional `MaxGradePercent`. Für jedes valide Gleisstück mit positiver Länge
wird wie bisher die vorzeichenbehaftete Steigung berechnet. Wenn
`abs(gradePercent) > maxGradePercent` gilt, entsteht der Warnhinweis `grade_limit_exceeded`. Eine
Steigung exakt am Grenzwert ist zulässig; eine kleine numerische Toleranz verhindert instabile
Fließkomma-Vergleiche.

Der Hinweis enthält die Objekt-ID sowie die berechnete Steigung und den gültigen Grenzwert. Auf- und
Abstieg werden damit gleich bewertet, der Detailtext zeigt zur Diagnose weiterhin das Vorzeichen.
Die Änderungsvorschau analysiert Basis- und Arbeitsrevision mit der aktuell für die Anlage
konfigurierten Grenze. Der Grenzwert wird nicht historisch pro Revision dupliziert.

## Persistenz, Backup und API

Migration `0050_layout_max_grade_percent.sql` ergänzt die nullable Spalte mit einer SQLite-Check-
Constraint. Das Layout-Domänenmodell sowie Create- und Update-Eingaben führen `maxGradePercent` als
optionale Zahl. Repository, Handler und OpenAPI bleiben dabei synchron.

Die Backup-Version steigt von 8 auf 9. Neue Exporte enthalten die Spalte über den generischen
Tabellenexport. Backups bis Version 8 enthalten sie nicht und werden mit `NULL` wiederhergestellt.
Damit bleiben alte Anlagen ohne nachträglich erfundene Grenze vollständig kompatibel.

## UI

Die Planprüfung ergänzt einen Zähler für überschrittene Steigungsgrenzen. Der Warnhinweis nennt
Steigung und Grenze lokalisiert, zum Beispiel „Steigung -5,51 % überschreitet Grenzwert 3,00 %“.
Beim Anklicken fokussiert der Planer das betroffene Gleis, dessen Höhenprofil anschließend manuell
korrigiert werden kann. Deutsch und Englisch werden gemeinsam gepflegt.

## Fehler- und Grenzfälle

- Leerer Grenzwert erzeugt keine Grenzwert-Warnungen.
- Werte kleiner oder gleich 0, größer als 100, NaN und unendliche Werte werden serverseitig
  abgelehnt.
- Eine Steigung exakt am Grenzwert erzeugt keine Warnung.
- Auf- und Abstieg werden über den Betrag verglichen.
- Ungültige oder längenlose Geometrien bleiben beim vorhandenen Verhalten und erzeugen keine
  zusätzliche Grenzwert-Warnung.
- Die Sortierung der Hinweise bleibt deterministisch.

## Abgrenzung

Nicht Teil dieses Pakets sind automatische Höhenkorrektur, Grenzwerte je Gleis oder Einheit,
fahrzeugabhängige Zugkraftprofile, Routenauswertung, Übergangsbögen, Durchfahrtshöhen,
Ebenenkollisionen und Flexgleisoptimierung.

## Abnahme

- Anlagen können eine optionale maximale Steigung von mehr als 0 bis einschließlich 100 Prozent
  speichern und wieder entfernen.
- Alte Backups werden ohne Grenzwert wiederhergestellt; neue Backups erhalten Version 9.
- Auf- und Abstieg oberhalb der Grenze erzeugen jeweils einen präzisen Warnhinweis.
- Der exakte Grenzwert und ein nicht gesetzter Grenzwert erzeugen keinen Hinweis.
- Planprüfung und Änderungsvorschau verwenden dieselbe aktuelle Anlagenkonfiguration.
- Die Oberfläche zeigt Zähler und lokalisierten Detailtext und fokussiert das betroffene Gleis.
- Backendtests, Frontendtests, OpenAPI-Vertrag, Produktionsbuild und lokale Browserprüfung laufen
  erfolgreich.
