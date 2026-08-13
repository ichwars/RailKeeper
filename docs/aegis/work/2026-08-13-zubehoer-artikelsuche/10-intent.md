# Task Intent: Zubehör-Artikelsuche

## Requested Outcome

Die deutsche Artikelübersicht heißt `Zubehör`. Der Zubehör-Artikel-Dialog erhält die vorhandene
Barcode- und Artikeldatensuche des Fahrzeugbestands mit kontrollierter Feldübernahme und sicherem
Import ausgewählter Trefferbilder.

## Scope

- gemeinsame app-eigene Barcode- und Artikelsuchdialoge,
- Zubehör-Feldadapter und Suchcontroller,
- sicherer Zubehörbildimport aus externen Suchtreffern,
- deutsche und englische Bereichsbezeichnung,
- Frontend-, Backend- und OpenAPI-Tests,
- isolierter Branch `dev/accessory-article-search` auf Basis von `origin/main`.

## Non-Goals

- keine Anlage-Funktionen oder Anlage-Commits,
- keine neuen Suchanbieter,
- keine automatische Artikelart- oder Unterarterkennung,
- keine Cloud-Synchronisierung oder öffentliche Freigabe,
- keine automatische Artikelspeicherung durch die Trefferübernahme.

## Risk Hints

- Die gemeinsame Dialogextraktion darf die Fahrzeugsuche nicht verändern.
- Externe URLs und Dateiinhalte bleiben nicht vertrauenswürdig.
- Teilweise fehlgeschlagene Bildimporte dürfen keinen zweiten Artikel erzeugen.
- Backendroute, Frontendclient und OpenAPI-Vertrag müssen synchron bleiben.
- Der lokale `main` und der Anlage-Branch enthalten nicht freizugebende Anlage-Arbeit.

## Baseline Read Set

- `AGENTS.md`
- `docs/superpowers/specs/2026-08-13-zubehoer-artikelsuche-design.md`
- `docs/superpowers/plans/2026-08-13-zubehoer-artikelsuche.md`
- `frontend/src/features/vehicles/ArticleSearchDialog.tsx`
- `frontend/src/features/vehicles/BarcodeSearchDialog.tsx`
- `frontend/src/features/vehicles/useArticleSearchController.ts`
- `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- `frontend/src/features/accessories/useArticleEditorController.ts`
- `backend/internal/api/vehicle_attachment_handlers.go`
- `backend/internal/api/accessory_document_handlers.go`
- `backend/internal/api/routes.go`
- `openapi/railkeeper.yaml`

## Baseline Usage

- Required refs: vollständig gelesen oder gezielt untersucht.
- Acknowledged refs: Design, Plan, Fahrzeug- und Zubehörsuchpfade, Dateiimport und Vertrag.
- Cited refs: der Implementierungsplan nennt jede betroffene Datei und Verifikation.
- Missing refs: keine.
- Decision: bereit für Task 1 des freigegebenen Plans.

## Impact Statement

Die Änderung betrifft Frontendnavigation, gemeinsame Suchkomponenten, Zubehör-Editorzustand, eine
neue geschützte Bildimport-Route und den OpenAPI-Vertrag. Persistente Zubehördaten ändern sich nur
durch explizites Speichern und ausgewählten Bildimport. Es ist keine Datenbankmigration erforderlich.

## Execution Readiness View

- Intent lock: Zubehörsuche wie Fahrzeugbestand, kein globaler Übersichts-Suchablauf.
- Scope fence: ausschließlich die sieben Tasks des Implementierungsplans.
- Baseline lock: `origin/main` bei Worktree-Erstellung, Commit `5b5ea1d`.
- Compatibility boundary: Fahrzeugverhalten und `/accessories` bleiben stabil.
- Retirement boundary: fahrzeugspezifische Dialogdateien werden erst entfernt, wenn gemeinsame
  Komponenten alle bestehenden Tests erfüllen.
- Task batches: Benennung, gemeinsame UI, Feldadapter, Dialogintegration, Backendimport,
  Bildspeicherung, Gesamtabnahme.
- Test obligations: TDD je Task, vollständige Go- und Frontendsuite, Build, Browser-Smoke-Test.
- Review gates: Rot-Nachweis, Grün-Nachweis und Commit pro Task; GitHub-Prüfungen vor Merge.
- Drift rule: neue Schema-, Migrations-, Anbieter- oder Anlage-Anforderungen führen zurück zur
  Planung statt zu stiller Scope-Erweiterung.
- Completion evidence: Tests, Build, Diff-Isolation, PR-Prüfungen und Merge-Nachweis.

## Stop Conditions

- Done: alle sieben Tasks verifiziert, PR-Prüfungen grün und PR gemergt.
- Needs verification: Code vorhanden, aber eine geforderte lokale oder GitHub-Prüfung fehlt.
- Blocked: Abhängigkeit oder externer Zustand verhindert wiederholt jede sichere Fortsetzung.
- Scope exceeded: eine neue Architektur-, Schema-, Sicherheits- oder Produktentscheidung wird nötig.
