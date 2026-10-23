---
title: Architektur und Beiträge
description: Architektur und Beiträge.
audience: developer
status: development
reviewedVersion: main
lastReviewed: 2026-10-23
---

# Architektur und Beiträge

RailKeeper ist eine lokale, selbst gehostete Anwendung: ein Go-Backend, eine React-/TypeScript-Oberfläche, SQLite und ein privater Dateispeicher. Das Backend liefert API und gebaute Oberfläche aus. Ein Cloud-Konto ist nicht erforderlich.

## Verantwortungen

| Bereich | Ort im Repository | Aufgabe |
| --- | --- | --- |
| Oberfläche | `frontend/src` | Arbeitsabläufe, Vorschauen und übersetzte Texte |
| Backend | `backend/internal` | Rollen, Validierung, Datenzugriff und Geräteadapter |
| Datenbank | `backend/migrations` | Versionierte SQLite-Änderungen |
| Mitgelieferte Daten | `backend/seeds` | Stammdaten und Decoderbibliothek |
| API-Vertrag | `openapi/railkeeper.yaml` | Öffentliche HTTP-Schnittstelle |
| Handbuch | `docs/site` | Deutsch und Englisch unter gleichen relativen Pfaden |

Die [Architekturreferenz](https://github.com/ichwars/RailKeeper/blob/main/docs/architecture.md) beschreibt Modulgrenzen, Startabsicherung und Importentscheidungen genauer. Die [Beitragsregeln](https://github.com/ichwars/RailKeeper/blob/main/CONTRIBUTING.md) erläutern den Arbeitsablauf.

## Lokal entwickeln

Installiere die in `backend/go.mod` und den CI-Dateien angegebene Go-/Node-Version. Installiere Frontend-Abhängigkeiten mit `npm ci`, baue mit `npm run build` und starte das Backend aus `backend` mit `go run ./cmd/railkeeper`. Verwende einen eigenen Datenordner für Tests.

Vor einem Pull Request: Go formatieren, `go test ./...` im Backend und `npm run build` im Frontend ausführen. Für betroffene Abläufe gezielte Tests ergänzen. Dokumentationsänderungen benötigen beide Sprachen und `npm run check` unter `docs`.

## Grenzen erhalten

Rollen und Schutzmaßnahmen werden serverseitig geprüft. Gerätedaten beginnen als lesende Vorschläge; Schreibvorgänge benötigen eine bewusst bestätigte Vorschau. CV-Werte bleiben von ihrer Bedeutung getrennt: Bibliotheksregeln erklären Rohdaten nur bei belegtem Modell-, Firmware- und Indexkontext.

Keine echten Zugangsdaten, Datenbanken oder Sammlungsbilder in Tests und Screenshots übernehmen. Für Fehlerberichte können lokale Diagnose- und Testvariablen existieren; sie sind keine allgemeinen Benutzereinstellungen. Siehe [Sicherheitsrichtlinie](https://github.com/ichwars/RailKeeper/blob/main/SECURITY.md).
