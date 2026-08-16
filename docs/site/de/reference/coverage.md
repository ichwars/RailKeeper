---
title: Dokumentationsabdeckung
description: Zuordnung der RailKeeper-Quellflächen zum zweisprachigen Handbuch.
audience: reference
status: development
reviewedVersion: main
lastReviewed: 2026-08-15
---

# Dokumentationsabdeckung

RailKeeper behandelt die Dokumentationsabdeckung als geprüften Vertrag. Eine schreibgeschützte
Inventur erfasst sichtbare Frontend-Routen, englische und deutsche Übersetzungsschlüssel,
registrierte Go-API-Routen, OpenAPI-Operationen, versionierte Konfigurationsvariablen und bestehende
Projektdokumente. Der Coverage-Validator verlangt anschließend für jede Frontend-Route, jeden
Übersetzungsschlüssel, jede API-Route und jede Umgebungsvariable genau ein verantwortliches Thema.

Die geprüften Zuordnungen stehen in
[`docs/coverage.json`](https://github.com/ichwars/RailKeeper/blob/main/docs/coverage.json). Die
Inventur entsteht nur während der Prüfung und wird nicht als zweite Wahrheitsquelle eingecheckt.

## Abdeckungsstatus

- `planned`: Ziel und Verantwortung stehen fest, das vollständige Seitenpaar folgt in einer
  späteren Dokumentationsetappe.
- `documented`: Die englische und die deutsche Zielseite existieren und erfüllen den Seitenstandard.
- `internal`: Das Thema bleibt absichtlich auf Informationen für Maintainer beschränkt.
- `not-published`: Die Quellfläche existiert, darf aber nicht als stabile öffentliche Funktion
  dargestellt werden.

Ein Thema erhält erst dann den Status `documented`, wenn beide Zielseiten unter identischen relativen
Pfaden liegen, übereinstimmende Prüfmetadaten tragen und die Dokumentationsprüfung bestehen. Die
Anlagenplanung bleibt `not-published`, weil sie nicht zum stabilen, öffentlich dokumentierten
Produktumfang gehört.

## Geprüfte Themen

Die erste Matrix umfasst 21 Themen.

### Benutzerabläufe

- `setup-auth`
- `overview`
- `vehicles-core`
- `vehicle-media`
- `vehicle-maintenance`
- `vehicle-decoder-cv`
- `vehicle-search-spares`
- `accessories`
- `exhibition`
- `import-export`
- `settings-general`

### Administration

- `master-data`
- `users-sessions-security`
- `backup-restore`
- `digital-centers`
- `system-operations`
- `deployment-configuration`

### Entwicklung

- `layouts-unpublished`
- `development-architecture`

### Gemeinsame Referenz

- `releases-support`
- `shared-navigation`

## Ablauf für Beiträge

Wenn sich eine Route, ein API-Bereich, ein Übersetzungsnamensraum, eine Umgebungsvariable oder ein
Handbuchthema ändert, muss `docs/coverage.json` im selben Pull Request angepasst werden. Die
vollständige Prüfung startet im Repository mit:

```powershell
cd docs
npm.cmd run check
```

Der Befehl prüft beide Sprachen, die Prüfmetadaten, die Coverage-Zuordnung und den produktiven
Website-Build.
