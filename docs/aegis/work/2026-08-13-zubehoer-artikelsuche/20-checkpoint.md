# Todo Checkpoint: Zubehör-Artikelsuche

## Current Todo

Task 4: Barcode- und Artikeldatensuche in den Zubehördialog integrieren.

## Completed Todos

- Design freigegeben und als `f1f9bea` committed.
- Implementierungsplan geprüft und als `7196f61` committed.
- Isolierter Worktree von `origin/main` angelegt.
- Baseline Backend und Frontend vollständig grün geprüft.
- Task 1 als `8eacf57` committed.
- Task 2 gemeinsame Suchmodelle, Präferenzen und Dialoge umgesetzt; Fahrzeugtests und Build grün.
- Task 3 Zubehör-Feldadapter mit fünf fachlichen Tests grün umgesetzt.

## Active Slice

Slice Card:

- Goal: Suchblock, Barcodeablauf und Ergebnisdialog in den Zubehör-Artikel-Dialog integrieren.
- Parent plan/spec: `docs/superpowers/plans/2026-08-13-zubehoer-artikelsuche.md`.
- Files: Zubehör-Suchcontroller, Artikel-Tab, Editor-Dialog, View, Tests und Übersetzungen.
- Boundary: noch kein dauerhafter Bildimport, keine Backendänderung, keine Anlage-Datei.
- Verification: gezielte Controller- und Dialogtests erst rot, dann grün.
- Stop: Task-4-Commit oder fachlich unerwarteter Testfehler.

## Evidence Refs

- Backendbaseline: `go test ./...`, alle Pakete erfolgreich.
- Frontendbaseline: `npm.cmd run test:run`, 51 Testdateien und 302 Tests erfolgreich.
- Worktreebasis: `origin/main` bei `5b5ea1d`.
- Task 1 Rot: vier erwartete Altbezeichnungsfehler.
- Task 1 Grün: zwei Testdateien, 16 Tests erfolgreich.
- Task 2 Rot: gemeinsames Modell fehlte am neuen Importpfad.
- Task 2 Grün: drei Fahrzeug-Testdateien, acht Tests erfolgreich; Produktionsbuild erfolgreich.
- Task 3 Rot: Zubehöradaptermodul fehlte.
- Task 3 Grün: eine Testdatei, fünf Tests erfolgreich.

## Blocked On

Nichts.

## Next Step

Task-4-Tests für Suchblock, Barcode, Rollen, Suchzustände und Trefferübernahme schreiben.

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
