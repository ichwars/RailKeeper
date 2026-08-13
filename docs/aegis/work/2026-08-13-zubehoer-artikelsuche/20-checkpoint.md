# Todo Checkpoint: Zubehör-Artikelsuche

## Current Todo

Task 5: Sicheren serverseitigen Bildimport für Zubehörartikel ergänzen.

## Completed Todos

- Design freigegeben und als `f1f9bea` committed.
- Implementierungsplan geprüft und als `7196f61` committed.
- Isolierter Worktree von `origin/main` angelegt.
- Baseline Backend und Frontend vollständig grün geprüft.
- Task 1 als `8eacf57` committed.
- Task 2 gemeinsame Suchmodelle, Präferenzen und Dialoge umgesetzt; Fahrzeugtests und Build grün.
- Task 3 Zubehör-Feldadapter mit fünf fachlichen Tests grün umgesetzt.
- Task 4 Suchcontroller, Barcodeablauf, Ergebnisdialog und Bildvormerkung in den Zubehördialog integriert.

## Active Slice

Slice Card:

- Goal: Ausgewählte externe Bilder sicher als Zubehördokument importieren.
- Parent plan/spec: `docs/superpowers/plans/2026-08-13-zubehoer-artikelsuche.md`.
- Files: Backend-Handler, sichere Downloadhilfe, Router, OpenAPI und Tests.
- Boundary: nur öffentliche HTTP(S)-Bilder, bestehende SSRF- und Größenlimits, keine Anlage-Datei.
- Verification: Handler- und Vertragstests erst rot, dann grün.
- Stop: Task-5-Commit oder fachlich unerwarteter Sicherheitsfehler.

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
- Task 4 Rot: Zubehör-Suchcontroller fehlte; danach fehlten die beiden Suchaktionen im Dialog.
- Task 4 Grün: drei gezielte Testdateien, 77 Tests erfolgreich; Produktionsbuild erfolgreich.

## Blocked On

Nichts.

## Next Step

Task-5-Tests für Rollen, URL-Validierung, SSRF, MIME-Typ, Größe und erfolgreichen Import schreiben.

## Resume State Hint

Worktree: `C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-article-search`.
Branch: `dev/accessory-article-search`. Tasks 1 bis 3 sind committed; Task 4 ist vollständig grün.

## Drift Check

- Intent: unverändert.
- Scope: unverändert.
- Compatibility: `/accessories` und Fahrzeugfunktion bleiben geschützt.
- New owner or branch: nur der freigegebene isolierte Branch.
- Bounded adaptation: gemeinsame Dialoge verwenden vorerst die bestehenden sichtbaren
  Übersetzungsschlüssel; es entsteht keine doppelte Übersetzungstabelle.
- Decision: continue.
