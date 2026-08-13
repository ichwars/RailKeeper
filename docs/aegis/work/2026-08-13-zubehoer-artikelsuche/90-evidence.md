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

## Remaining Evidence

- Rot- und Grün-Nachweis je Task,
- vollständige Backend- und Frontendsuite,
- Frontendbuild,
- Diff- und Anlage-Isolation,
- Browser-Smoke-Test,
- GitHub-Prüfungen und Merge.
