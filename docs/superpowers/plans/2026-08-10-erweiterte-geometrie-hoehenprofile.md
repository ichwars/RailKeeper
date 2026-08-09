# Umsetzungsplan: Gleis-Höhenprofile

## 1. Domänenanalyse

- [x] Erwartete Steigungsberechnung zuerst als Domänentest ergänzen.
- [x] Höhenfelder und deterministische `TrackGrade`-Analyse implementieren.

## 2. Anwendung und Persistenz

- [x] Validierung nicht endlicher Höhen zuerst testen.
- [x] Create-, Update- und Repository-Pfade um beide Höhen erweitern.
- [x] Migration 0049 mit Standardwerten ergänzen und Schema testen.
- [x] Das Klonen einer Revision einschließlich Höhen testen und implementieren.

## 3. Backup und Vertrag

- [x] Backup-Rundlauf der Höhen sowie Kompatibilität von Version 7 testen.
- [x] Backup-Format auf Version 8 anheben.
- [x] Backendantworten und OpenAPI-Schemas um Höhen und `grades` erweitern.

## 4. Planer-UI

- [x] UI-Test für Bearbeitung, unveränderte Höhen bei Bewegung und lesende Anzeige ergänzen.
- [x] API-Typen, Gleisinspektor und deutsch/englische Texte erweitern.
- [x] Bestehende App-Komponenten und das kompakte Planer-Raster verwenden.

## 5. Verifikation

- [x] Gezielte Backend- und Frontendtests ausführen.
- [x] Vollständige Backendtests, Frontendtests und Produktionsbuild ausführen.
- [x] Server neu starten und Höhenprofil im Browser lokal abnehmen.
- [x] Abnahmeergebnisse im Design dokumentieren.
