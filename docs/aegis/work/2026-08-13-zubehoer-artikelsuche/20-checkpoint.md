# Todo Checkpoint: Zubehör-Artikelsuche

## Current Todo

Task 1: Bereich in `Zubehör` umbenennen.

## Completed Todos

- Design freigegeben und als `f1f9bea` committed.
- Implementierungsplan geprüft und als `7196f61` committed.
- Isolierter Worktree von `origin/main` angelegt.
- Baseline Backend und Frontend vollständig grün geprüft.

## Active Slice

Slice Card:

- Goal: sichtbare Bereichsbezeichnung ohne Routen- oder Fachbegriffsänderung umstellen.
- Parent plan/spec: `docs/superpowers/plans/2026-08-13-zubehoer-artikelsuche.md`.
- Files: Shell- und Zubehörtests sowie deutsche und englische i18n-Dateien.
- Boundary: keine Suchlogik, keine Backendänderung, keine Anlage-Datei.
- Verification: zwei gezielte Vitest-Dateien erst rot, dann grün.
- Stop: Task-1-Commit oder fachlich unerwarteter Testfehler.

## Evidence Refs

- Backendbaseline: `go test ./...`, alle Pakete erfolgreich.
- Frontendbaseline: `npm.cmd run test:run`, 51 Testdateien und 302 Tests erfolgreich.
- Worktreebasis: `origin/main` bei `5b5ea1d`.

## Blocked On

Nichts.

## Next Step

Task-1-Tests auf `Zubehör` und `Accessories` ändern und den erwarteten Rot-Lauf dokumentieren.

## Resume State Hint

Worktree: `C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-article-search`.
Branch: `dev/accessory-article-search`. Produktionscode wurde noch nicht verändert.

## Drift Check

- Intent: unverändert.
- Scope: unverändert.
- Compatibility: `/accessories` und Fahrzeugfunktion bleiben geschützt.
- New owner or branch: nur der freigegebene isolierte Branch.
- Decision: continue.
