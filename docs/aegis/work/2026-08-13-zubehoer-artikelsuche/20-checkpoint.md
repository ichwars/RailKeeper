# Todo Checkpoint: Zubehör-Artikelsuche

## Current Todo

Task 3: Zubehör-Feldadapter testgetrieben ergänzen.

## Completed Todos

- Design freigegeben und als `f1f9bea` committed.
- Implementierungsplan geprüft und als `7196f61` committed.
- Isolierter Worktree von `origin/main` angelegt.
- Baseline Backend und Frontend vollständig grün geprüft.
- Task 1 als `8eacf57` committed.
- Task 2 gemeinsame Suchmodelle, Präferenzen und Dialoge umgesetzt; Fahrzeugtests und Build grün.

## Active Slice

Slice Card:

- Goal: Zubehörsuchkriterien und kontrollierte Feldzuordnung als reine Funktionen definieren.
- Parent plan/spec: `docs/superpowers/plans/2026-08-13-zubehoer-artikelsuche.md`.
- Files: `accessoryArticleSearch.ts` und zugehöriger Test.
- Boundary: keine UI-Integration, keine Backendänderung, keine Anlage-Datei.
- Verification: gezielter Adaptertest erst rot, dann grün.
- Stop: Task-3-Commit oder fachlich unerwarteter Testfehler.

## Evidence Refs

- Backendbaseline: `go test ./...`, alle Pakete erfolgreich.
- Frontendbaseline: `npm.cmd run test:run`, 51 Testdateien und 302 Tests erfolgreich.
- Worktreebasis: `origin/main` bei `5b5ea1d`.
- Task 1 Rot: vier erwartete Altbezeichnungsfehler.
- Task 1 Grün: zwei Testdateien, 16 Tests erfolgreich.
- Task 2 Rot: gemeinsames Modell fehlte am neuen Importpfad.
- Task 2 Grün: drei Fahrzeug-Testdateien, acht Tests erfolgreich; Produktionsbuild erfolgreich.

## Blocked On

Nichts.

## Next Step

Task-3-Tests für Kriterien, Stammdaten, Konflikte und Feldabbildung schreiben.

## Resume State Hint

Worktree: `C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-article-search`.
Branch: `dev/accessory-article-search`. Task 1 ist committed; Task 2 ist vor dem Commit vollständig grün.

## Drift Check

- Intent: unverändert.
- Scope: unverändert.
- Compatibility: `/accessories` und Fahrzeugfunktion bleiben geschützt.
- New owner or branch: nur der freigegebene isolierte Branch.
- Bounded adaptation: gemeinsame Dialoge verwenden vorerst die bestehenden sichtbaren
  Übersetzungsschlüssel; es entsteht keine doppelte Übersetzungstabelle.
- Decision: continue.
