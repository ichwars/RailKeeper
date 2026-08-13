# Todo Checkpoint: Zubehör-Artikelsuche

## Current Todo

Task 7: GitHub-Übergabe nach grüner lokaler Gesamtabnahme.

## Completed Todos

- Design freigegeben und als `f1f9bea` committed.
- Implementierungsplan geprüft und als `7196f61` committed.
- Isolierter Worktree von `origin/main` angelegt.
- Baseline Backend und Frontend vollständig grün geprüft.
- Task 1 als `8eacf57` committed.
- Task 2 gemeinsame Suchmodelle, Präferenzen und Dialoge umgesetzt; Fahrzeugtests und Build grün.
- Task 3 Zubehör-Feldadapter mit fünf fachlichen Tests grün umgesetzt.
- Task 4 Suchcontroller, Barcodeablauf, Ergebnisdialog und Bildvormerkung in den Zubehördialog integriert.
- Task 5 serverseitigen URL-Bildimport mit Rollen-, SSRF-, Redirect-, MIME- und Größenprüfung ergänzt.
- Task 6 Bildentwürfe werden nach dem Speichern importiert; Teilerfolge bleiben wiederholbar und erzeugen keine Dublette.
- Task 7 lokale Testmatrix, Build, OpenAPI, Vet und Diff-Isolation erfolgreich; Markdown-Whitespace bereinigt.

## Active Slice

Slice Card:

- Goal: Abschlusscommit, Push, PR, GitHub-Prüfungen und Merge.
- Parent plan/spec: `docs/superpowers/plans/2026-08-13-zubehoer-artikelsuche.md`.
- Files: keine weiteren Produktionsänderungen geplant.
- Boundary: Anlage bleibt vollständig ausgeschlossen.
- Verification: GitHub-CI und PR-Diff.
- Stop: erfolgreicher Merge oder ein echter externer Blocker.

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
- Task 5 Rot: URL-Downloadhilfe, Sentinel-Fehler und OpenAPI-Pfad fehlten.
- Task 5 Grün: API-Paket und vollständige Backendsuite erfolgreich.
- Task 6 Rot: API-Methode fehlte; Bildentwürfe wurden nach Stammdatenspeicherung nicht importiert.
- Task 6 Grün: fünf gezielte Testdateien, 96 Tests erfolgreich; Build erfolgreich; Primärbild- und Sitzungswechseltest grün.
- Task 7 lokal: `go test ./...` PASS; 53 Frontend-Testdateien und 316 Tests mit Coverage PASS;
  Frontendbuild PASS; OpenAPI-Vertrag und `go vet ./...` mit lokalem Go-Cache PASS.
- Isolationsprüfung: Basis `origin/main` bei `5b5ea1d`; keine Anlage-Datei im Branchdiff.

## Blocked On

Nichts.

## Next Step

Verifikationsnachtrag committen, Branch pushen, PR erstellen und CI bis zum Merge beobachten.

## Resume State Hint

Worktree: `C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-article-search`.
Branch: `dev/accessory-article-search`. Tasks 1 bis 6 sind committed; lokale Gesamtabnahme ist grün.

## Drift Check

- Intent: unverändert.
- Scope: unverändert.
- Compatibility: `/accessories` und Fahrzeugfunktion bleiben geschützt.
- New owner or branch: nur der freigegebene isolierte Branch.
- Bounded adaptation: gemeinsame Dialoge verwenden vorerst die bestehenden sichtbaren
  Übersetzungsschlüssel; es entsteht keine doppelte Übersetzungstabelle.
- Bounded adaptation: der sichere Downloader ist als kleine testbare API-Hilfe extrahiert;
  der bestehende `safefetch`-Client bleibt für DNS- und Redirectschutz zuständig.
- Decision: continue.
