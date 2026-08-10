# Evidence

## Automatisierte Prüfung

Ausgeführt am 10. August 2026 auf `dev/issue-36-advanced-geometry`:

```text
cd backend
go test ./...
PASS: 7 Pakete

cd frontend
npm.cmd test -- --run
PASS: 73 Testdateien, 379 Tests

npm.cmd run build
PASS: 2182 Module transformiert
```

Der Build meldete nur die bekannten Hinweise zu `configLoader` und Chunkgröße. `git diff --check`
war sauber. `frontend/dist` und die isolierten QA-Daten unter `.cache` bleiben unversioniert.

## Laufzeit

- Hauptserver: `http://127.0.0.1:18083`, Listener PID 18832, `/health` HTTP 200
- Isolierte Browser-QA während der Abnahme: `http://127.0.0.1:18084`, Listener PID 11812,
  `/health` HTTP 200, anschließend kontrolliert beendet
- Die schreibende Abnahme verwendete eine getrennte Datenbank unter `.cache/stage5-browser-qa`.
  Die Hauptdatenbank wurde nicht für QA-Importe verändert.

## Browserabnahme

Geprüft mit `Stage 5 QA`, `Stage 5 QA Modul`, Planrevision R1:

- Die vorinstallierte Bibliothek `Tillig · TT Modellgleis`, Version 1, zeigte zwei geprüfte
  Geometrien und ihre Quelle.
- Eine RailKeeper-JSON-Datei wurde über den app-eigenen Dateiwähler vollständig geprüft.
- Die Vorschau zeigte Hersteller, System, Spurweite, Maßstab, Version, eine Geometrie und den
  Hinweis, dass externe Prüfstatus zurückgesetzt werden.
- Der bestätigte Import erschien als `Entwurf` und war noch nicht platzierbar.
- Die Admin-Prüfung verlangte einen Nachweis. Danach erschien die Bibliothek als `Geprüft`.
- Der Export wurde über die Oberfläche ohne UI- oder API-Fehler ausgelöst. Backendtests belegen
  zusätzlich den vollständigen JSON-Roundtrip.
- Die freigegebene Geometrie erschien als `RailKeeper QA QA-100 · QA Testgerade` in der Palette.
- Nach Platzierung zeigte der Plan zwei offene Enden und eine Stücklistenposition mit Menge 1.
- Der Materialsaldo zeigte `+1 RailKeeper QA QA-100`.
- Stilllegung lief über den app-eigenen Bestätigungsdialog.
- Nach Stilllegung verschwand QA-100 aus der Palette. Das platzierte Gleis, die zwei offenen Enden,
  Hersteller, Artikelnummer, Name und Stückliste blieben unverändert.
- Die Browserkonsole enthielt keine Fehler oder Warnungen.

Die Browserabnahme deckte zwei Integrationsfehler auf, die vor Abschluss korrigiert und mit
Regressionstests abgesichert wurden: einen wiederholt gestarteten Loader und einen hartcodierten
Hersteller im Planer.

Alle Änderungen blieben lokal. Es erfolgte kein Push, Pull Request, Merge oder Release.
