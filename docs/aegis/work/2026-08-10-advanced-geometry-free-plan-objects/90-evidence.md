# Evidence

## Automatisierte Prüfung

Ausgeführt am 10. August 2026 auf `dev/issue-36-advanced-geometry`:

```text
cd backend
go test ./...
PASS: 7 Pakete

cd frontend
npm.cmd test -- --run
PASS: 72 Testdateien, 374 Tests

npm.cmd run build
PASS: 2179 Module transformiert
```

Der Build meldete nur die bereits bekannten Hinweise zu `configLoader` und Chunkgröße. Der lokal
gebaute Ordner `frontend/dist` bleibt unversioniert.

## Server

- Prozess: `railkeeper`, PID 21464
- Adresse: `http://127.0.0.1:18083`
- `/health`: HTTP 200, `{"status":"ok"}`

## Browserabnahme

Geprüft in `QA Bahnhofsmodul`, Revision 2:

- Rechteck `QA Gebäude bearbeitet`, Kategorie Gebäude, 135 × 80 mm, verschoben auf
  398,2474884299394 / 235,71428311925374 mm.
- Ellipse `QA Bahnsteig`, Kategorie Bahnsteig, 180 × 45 mm.
- Linie `QA Landschaft`, Kategorie Landschaft, um 15 Grad gedreht.
- Beschriftung `QA Hinweis`, Kategorie Hinweis, Text `Bahnhof QA`, Schriftgröße 14 mm.
- Alle vier Formen und Kategorieklassen waren gleichzeitig in der SVG-Ebene vorhanden.
- Ein bearbeiteter Name wurde mit `Abbrechen` verworfen, die gespeicherten Daten blieben
  unverändert.
- Nach vollständigem Reload und erneutem Öffnen der Revision waren Namen, Kategorien, Formen,
  Maße, Positionen und Drehungen bytegenau identisch zur vorherigen DOM-Erfassung.
- `QA Hinweis` wurde anschließend über den app-eigenen Bestätigungsdialog gelöscht.
- Die Änderungsvorschau zeigte vor dem Löschen 4, danach 3 Planobjektänderungen.
- Die Tillig-Stückliste blieb bei 2 × 83101 und 1 × 83125.
- Die Planprüfung blieb bei 6 offenen Enden und 0 weiteren Warnungen.
- Planner-Tab und frisch geladener App-Tab meldeten keine Konsolenfehler.

Alle Änderungen und Prüfungen blieben lokal. Es erfolgte kein Push, Pull Request, Merge oder
Release.
