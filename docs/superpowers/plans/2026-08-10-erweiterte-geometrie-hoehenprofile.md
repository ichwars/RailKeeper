# Umsetzungsplan: Gleis-Höhenprofile

## 1. Domänenanalyse

- [ ] Erwartete Steigungsberechnung zuerst als Domänentest ergänzen.
- [ ] Höhenfelder und deterministische `TrackGrade`-Analyse implementieren.

## 2. Anwendung und Persistenz

- [ ] Validierung nicht endlicher Höhen zuerst testen.
- [ ] Create-, Update- und Repository-Pfade um beide Höhen erweitern.
- [ ] Migration 0049 mit Standardwerten ergänzen und Schema testen.
- [ ] Das Klonen einer Revision einschließlich Höhen testen und implementieren.

## 3. Backup und Vertrag

- [ ] Backup-Rundlauf der Höhen sowie Kompatibilität von Version 7 testen.
- [ ] Backup-Format auf Version 8 anheben.
- [ ] Backendantworten und OpenAPI-Schemas um Höhen und `grades` erweitern.

## 4. Planer-UI

- [ ] UI-Test für Bearbeitung, unveränderte Höhen bei Bewegung und lesende Anzeige ergänzen.
- [ ] API-Typen, Gleisinspektor und deutsch/englische Texte erweitern.
- [ ] Bestehende App-Komponenten und das kompakte Planer-Raster verwenden.

## 5. Verifikation

- [ ] Gezielte Backend- und Frontendtests ausführen.
- [ ] Vollständige Backendtests, Frontendtests und Produktionsbuild ausführen.
- [ ] Server neu starten und Höhenprofil im Browser lokal abnehmen.
- [ ] Abnahmeergebnisse im Design dokumentieren.

