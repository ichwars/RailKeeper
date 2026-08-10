# Gleisbibliotheken und Austauschformate Implementation Plan

**Goal:** Issue #33 mit einem sicheren, versionierten RailKeeper-JSON-Austausch für zusätzliche
Gleisbibliotheken und unveränderlichen Geometriesnapshots bestehender Pläne abschließen.

**Architecture:** Neue fokussierte Domain-, Application-, Repository-, Handler- und
React-Komponenten erweitern den vorhandenen `TrackPlannerService`, ohne zentrale Planerdateien mit
Importlogik zu belasten. Migration 0055 ergänzt Prüfmetadaten und Geometriesnapshots.

**Constraints:** Alles bleibt lokal. Fremdformate ohne verifizierbare Spezifikation werden nicht
implementiert. Importdaten sind untrusted, Status wird immer auf `draft` zurückgesetzt.

## Task 1: Austauschdomäne und Validierung

- [x] RailKeeper-Bibliotheksdokument, Vorschau, Statusinput und Limits definieren
- [x] Strikte Validierung für Format, Schema, Metadaten, URL, Duplikate und Geometrien testen
- [x] Domain-Tests grün ausführen
- [x] Lokal committen

## Task 2: Migration und unveränderliche Snapshots

- [x] Migration 0055 für Prüfmetadaten und `geometry_snapshot_json` schreiben
- [x] Bestehende Planobjekte deterministisch zurückfüllen
- [x] Platzierung, Lesen und Revisionsklon auf Snapshots umstellen
- [x] Regression beweist unveränderte Geometrie nach Definition- oder Statusänderung
- [x] Backupversion 14 und Restore-Kompatibilität bis Version 13 ergänzen
- [x] Infrastruktur- und Backuptests grün ausführen und lokal committen

## Task 3: Bibliotheksworkflow im Backend

- [x] Repository für Liste, Export, Konfliktprüfung, Draftimport und Statuswechsel testen
- [x] Servicevorschau normalisiert ohne Mutation
- [x] Bestätigter Import bleibt `draft`, bestätigte Admin-Prüfung setzt alle Definitionen
  `verified`, Stilllegung entfernt sie aus der Palette
- [x] Audit-Ereignisse und eindeutige Versionen prüfen
- [x] Lokal committen

## Task 4: API und OpenAPI

- [x] Handler-, Rollen-, CSRF-, Größen- und Fehlervertragstests schreiben
- [x] Fünf versionierte Routen registrieren
- [x] OpenAPI-Pfade und vollständige Schemas ergänzen
- [x] OpenAPI-Vertrag und Backendtests grün ausführen
- [x] Lokal committen

## Task 5: Strikter Client und Planeroberfläche

- [x] TypeScript-Verträge und API-Adapter testen
- [x] Kompaktes Bibliotheks-Panel mit Status, Quelle, Definitionen und Export bauen
- [x] App-eigene Importvorschau und Prüf-/Stilllegungsdialoge bauen
- [x] Deutsche und englische Texte sowie responsive Tokenstile ergänzen
- [x] Komponenten- und Integrationstests grün ausführen
- [x] Lokal committen

## Task 6: Vollständige Abnahme

- [ ] `go test ./...`
- [ ] `npm.cmd test -- --run`
- [ ] `npm.cmd run build`
- [ ] aktuellen lokalen Server neu starten und `/health` prüfen
- [ ] Browser: Export, Vorschau, Draftimport, Freigabe, Palette und Stilllegung prüfen
- [ ] Browser: Stückliste und vorhandene Revision bleiben nach Bibliotheksstatus unverändert
- [ ] Browserkonsolen auf Fehler prüfen
- [ ] Evidenz unter `docs/aegis/work/2026-08-10-track-libraries-exchange/` dokumentieren
- [ ] Plan abhaken, lokal committen und sauberen Arbeitsbaum prüfen
