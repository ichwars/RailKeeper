# Evidence Bundle Draft: Zubehör-Artikelsuche

## Baseline

| Evidence | Result |
| --- | --- |
| `go test ./...` | alle Backendpakete erfolgreich |
| `npm.cmd run test:run` | 51 Testdateien, 302 Tests erfolgreich |
| `git status --short` vor Design | sauber |
| Worktreebasis | `origin/main` bei `5b5ea1d` |

## Slice Evidence

| Slice | Red evidence | Green evidence |
| --- | --- | --- |
| Task 1 Benennung | vier erwartete alte Bezeichnungen | 2 Dateien, 16 Tests PASS |
| Task 2 gemeinsame Suche | neuer gemeinsamer Importpfad fehlte | 3 Dateien, 8 Tests PASS; Build PASS |
| Task 3 Zubehöradapter | Adaptermodul fehlte | 1 Datei, 5 Tests PASS |
| Task 4 Dialogintegration | Controller fehlte; Suchaktionen fehlten | 3 Dateien, 77 Tests PASS; Build PASS |
| Task 5 sicherer Bildimport | Downloadhilfe und OpenAPI-Pfad fehlten | API-Paket PASS; `go test ./...` PASS |
| Task 6 Bildpersistenz | API-Methode und Importschritt fehlten | 5 Dateien, 96 Tests PASS; Build PASS |

## Final Local Verification

| Evidence | Result |
| --- | --- |
| `go test ./...` | alle sieben Backendpakete PASS |
| `npm.cmd run test:coverage` | 53 Testdateien, 316 Tests PASS |
| `npm.cmd run build` | TypeScript und Vite-Build PASS |
| `go test ./internal/api -run OpenAPI -count=1` | PASS |
| `go vet ./...` | PASS |
| Branchbasis | `origin/main` bei `5b5ea1d` |
| Anlage-Isolation | keine Anlage-Datei im Branchdiff |
| isolierter Laufzeit-Smoke-Test | Port 18084, `/health` 200, App-Root 200, Testprozess beendet |

`golangci-lint` ist lokal nicht installiert. Der identische konfigurierte Lauf bleibt durch den
GitHub-CI-Job `Backend lint` abgedeckt.

## GitHub CI Follow-up

CodeQL meldete `go/request-forgery` am Erzeugen des externen Bildrequests. Der Befund war ein
Falschalarm für den bewusst flexiblen URL-Import: Der Handler prüft das erste Ziel vorab und der
`safefetch`-Transport prüft Request, Redirect, DNS-Ergebnis und tatsächliches Dial-Ziel erneut.
Diese projektspezifische Sanitizer-Kette wurde am Sink dokumentiert und nur für diese CodeQL-Regel
unterdrückt. Die gezielten Sicherheitsprüfungen und `go test ./...` liefen danach erneut erfolgreich.

## Remaining Evidence

- Rot- und Grün-Nachweis je Task,
- vollständige Backend- und Frontendsuite,
- Frontendbuild,
- Diff- und Anlage-Isolation,
- Browser-Smoke-Test,
- GitHub-Prüfungen und Merge.
