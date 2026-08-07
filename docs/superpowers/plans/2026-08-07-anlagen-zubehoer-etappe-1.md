# Anlagen und Zubehör, Etappe 1, Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Das gemeinsame Fundament für Anlagen, Module, Planrevisionen, Zubehörbestand,
Reservierungen und Einbauhistorie als vollständig getestete Backend- und Frontend-Funktion liefern.

**Architecture:** Zwei neue, voneinander getrennte Fachbereiche verwenden Domain-Typen und
Repository-Schnittstellen in `application` sowie SQLite-Implementierungen in `infrastructure`.
Explizite HTTP-Routen verbinden die Services mit zwei eigenständigen React-Features. Der Planer
verwaltet in Etappe 1 nur Anlagenstruktur und Revisionsmetadaten, noch keine Gleisgeometrie.

**Tech Stack:** Go 1.26.5, `net/http`, SQLite über `modernc.org/sqlite`, React 19.2.8,
TypeScript 7.0.2, Vite 8.2.0, Vitest 4.1.10, Testing Library, OpenAPI 3.1.

## Global Constraints

- RailKeeper bleibt lokal, selbst gehostet, SQLite-basiert und als ein Prozess deploybar.
- Es werden keine Steuerbefehle an Digitalzentralen, Decoder oder Rückmeldesysteme gesendet.
- `Planner` darf Pläne verwalten und Bestand reservieren, aber weder Bestand korrigieren noch
  Installationen bestätigen.
- `Editor` darf Zubehör, Bestand, Installationen und operative Zustände verwalten, aber keine
  Planrevision verändern oder veröffentlichen.
- `Viewer` liest Anlagen und Zubehör; `Messe` erhält keinen allgemeinen Inventarzugriff.
- `Admin` darf alle neuen Funktionen verwenden.
- Alle Berechtigungen werden serverseitig geprüft. Schreibzugriffe bleiben CSRF-geschützt.
- Planung, Reservierung, Einbau und Betriebszustand bleiben getrennte Zustandsachsen.
- Veröffentlichte Planrevisionen sind unveränderlich. Änderungen beginnen als neuer Entwurf.
- Bereits unterstützte Backup-Version 1 bleibt importierbar und erzeugt leere neue Fachbereiche.
- Backend-Routen, `frontend/src/shared/api.ts` und `openapi/railkeeper.yaml` bleiben synchron.
- Deutsch und Englisch werden gleichzeitig ergänzt.
- Neue Styles verwenden vorhandene Tokens und werden in fokussierten Feature-Dateien gehalten.
- Kein grafischer Planer, keine Tillig-Geometrie, keine Höhenberechnung und kein Fremdimport in
  Etappe 1.
- Jeder Pull Request aktualisiert sein GitHub-Arbeitspaket und verweist mit `Refs #<issue>` darauf.
- Nach dem Merge wird das Arbeitspaket geschlossen und das übergeordnete Tracking-Issue aktualisiert.

---

## Delivery Order

| PR | Arbeitspaket | Ergebnis | Abhängigkeit |
|---|---|---|---|
| 1 | Rollen und Schema | Planner-Rolle und vollständiges persistentes Fundament | keine |
| 2 | Anlagen-Backend | Anlagen, Einheiten, Aufbauten und Planrevisionen per API | PR 1 |
| 3 | Zubehör-Backend | Produkte, Lagerorte, Mengenbestand und Einzelobjekte per API | PR 1 |
| 4 | Materialfluss | Reservieren, Einbauen, Ausbauen und Historie | PR 2 und PR 3 |
| 5 | API-Vertrag | OpenAPI und typisierter Frontend-Client | PR 2 bis PR 4 |
| 6 | Zubehör-Oberfläche | Nutzbarer Zubehörbereich in Deutsch und Englisch | PR 5 |
| 7 | Anlagen-Oberfläche | Nutzbarer Anlagenbereich in Deutsch und Englisch | PR 5 |
| 8 | Backup und Abnahme | Versionskompatibilität, Roundtrip und Gesamtprüfung | PR 6 und PR 7 |

PR 2 und PR 3 können nach PR 1 parallel entstehen. PR 6 und PR 7 können nach PR 5 parallel
entstehen. Jeder PR bleibt einzeln testbar und reviewbar.

## Target File Map

### Backend

| Datei | Verantwortung |
|---|---|
| `backend/migrations/0039_layout_accessory_foundation.sql` | Tabellen, Constraints, Indizes und Planner-Rolle |
| `backend/internal/domain/accessory.go` | Zubehör-, Bestands-, Zustands- und Zieltypen |
| `backend/internal/domain/layout.go` | Anlagen-, Einheiten-, Plan- und Revisionszustände |
| `backend/internal/application/accessories.go` | AccessoryService, Eingaben, Repository-Vertrag |
| `backend/internal/application/accessory_inventory.go` | Bestand, Einzelobjekte und Lagerorte |
| `backend/internal/application/accessory_allocations.go` | Reservierungs- und Installationsabläufe |
| `backend/internal/application/layouts.go` | LayoutService, Eingaben, Repository-Vertrag |
| `backend/internal/application/layout_revisions.go` | Entwurf, Prüfung und Veröffentlichung |
| `backend/internal/infrastructure/accessory_repository.go` | SQLite-Persistenz und Transaktionen für Zubehör |
| `backend/internal/infrastructure/layout_repository.go` | SQLite-Persistenz und Transaktionen für Anlagen |
| `backend/internal/api/accessory_handlers.go` | Zubehör-, Bestand-, Reservierungs- und Einbau-HTTP |
| `backend/internal/api/layout_handlers.go` | Anlagen-, Modul-, Aufbau- und Revisions-HTTP |
| `backend/internal/api/routes.go` | Explizite Routen und Rollen |
| `backend/internal/api/router.go` | Service-Injektion |
| `backend/cmd/railkeeper/main.go` | Repository- und Service-Verdrahtung |
| `backend/internal/application/backup.go` | Backupversion 2 und neue Tabellen |

### Frontend und Vertrag

| Datei | Verantwortung |
|---|---|
| `openapi/railkeeper.yaml` | Öffentlicher API-Vertrag für Etappe 1 |
| `frontend/src/shared/api.ts` | Typen und HTTP-Methoden |
| `frontend/src/shared/i18n/de.ts` | Deutsche Texte |
| `frontend/src/shared/i18n/en.ts` | Englische Texte |
| `frontend/src/app/App.tsx` | Lazy Routes und Zugriffsschutz |
| `frontend/src/app/Shell.tsx` | Navigation für Zubehör und Anlagen |
| `frontend/src/features/accessories/AccessoriesView.tsx` | Bereichsorchestrierung |
| `frontend/src/features/accessories/AccessoryProductsPanel.tsx` | Produktkatalog und Bearbeitung |
| `frontend/src/features/accessories/AccessoryStockPanel.tsx` | Lagerorte, Bestand und Einzelobjekte |
| `frontend/src/features/accessories/AccessoryAllocationsPanel.tsx` | Reservierungen und Einbauhistorie |
| `frontend/src/features/layouts/LayoutsView.tsx` | Anlagenliste und Auswahl |
| `frontend/src/features/layouts/LayoutWorkspace.tsx` | Anlagenarbeitsmappe für Etappe 1 |
| `frontend/src/features/layouts/LayoutModulesPanel.tsx` | Anlageneinheiten |
| `frontend/src/features/layouts/LayoutConfigurationsPanel.tsx` | Aufbaukonfigurationen |
| `frontend/src/features/layouts/LayoutPlansPanel.tsx` | Varianten und Revisionsmetadaten |
| `frontend/src/styles/accessories.css` | Zubehördarstellung |
| `frontend/src/styles/layouts.css` | Anlagendarstellung |

## Task 1: Planner-Rolle und persistentes Fundament

**Files:**

- Create: `backend/migrations/0039_layout_accessory_foundation.sql`
- Create: `backend/internal/domain/accessory.go`
- Create: `backend/internal/domain/layout.go`
- Modify: `backend/internal/infrastructure/sqlite.go`
- Modify: `backend/internal/application/auth.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/routes_security_test.go`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**

- Produces die Rolle `Planner` und `routeAccessPlanner`.
- `Viewer`-Leserechte schließen `Editor`, `Planner` und `Admin` ein.
- Produces die Tabellen, die Tasks 2 bis 4 verwenden.

- [ ] **Step 1: Write failing role tests.**

  Ergänze `SeedRoles`- und Autorisierungstests mit den exakten Erwartungen:

  ```go
  if !hasRole([]string{"Planner"}, "Viewer") {
      t.Fatal("Planner must inherit Viewer read access")
  }
  if hasRole([]string{"Planner"}, "Editor") {
      t.Fatal("Planner must not inherit Editor write access")
  }
  if hasRole([]string{"Editor"}, "Planner") {
      t.Fatal("Editor must not inherit Planner write access")
  }
  ```

- [ ] **Step 2: Run the focused tests and verify failure.**

  Run: `cd backend; go test ./internal/application ./internal/infrastructure ./internal/api`

  Expected: FAIL because `Planner` and `routeAccessPlanner` do not exist.

- [ ] **Step 3: Add the Planner role without broadening Editor rights.**

  Change `SeedRoles` to seed exactly `Admin`, `Editor`, `Viewer`, `Messe`, `Planner`. Extend
  `hasRole` so required `Viewer` also accepts exact `Planner`. Add `routeAccessPlanner` and include
  it in the route metadata security test. Admin continues to satisfy every role through
  `RequireAnyRole`.

- [ ] **Step 4: Add migration 0039 with exact relational rules.**

  Create these tables with text UUID primary keys, UTC timestamp text columns, foreign keys and
  named `CHECK` constraints:

  ```sql
  storage_locations(id, parent_id, name, description, archived, created_at, updated_at)
  accessory_products(id, manufacturer, article_number, name, category, tracking_mode,
                     description, created_at, updated_at)
  accessory_stock(product_id, location_id, quantity, updated_at)
  accessory_assets(id, product_id, inventory_number, serial_number, condition_state,
                   lifecycle_state, storage_location_id, purchase_date, purchase_price,
                   warranty_until, notes, created_at, updated_at)
  layouts(id, name, kind, gauge, scale, description, version, archived, created_at, updated_at)
  layout_units(id, layout_id, name, kind, owner_label, width_mm, height_mm, version,
               archived, created_at, updated_at)
  plan_variants(id, layout_unit_id, name, description, archived, created_at, updated_at)
  plan_revisions(id, variant_id, revision_number, status, base_revision_id, version,
                 created_by, published_by, published_at, created_at, updated_at)
  layout_configurations(id, layout_id, name, description, version, archived,
                        created_at, updated_at)
  layout_configuration_units(id, configuration_id, unit_id, plan_revision_id,
                             position_x_mm, position_y_mm, rotation_degrees, sort_order)
  accessory_reservations(id, product_id, asset_id, location_id, quantity, layout_id,
                         layout_unit_id, status, note, created_by, created_at, updated_at)
  accessory_installations(id, product_id, asset_id, source_location_id, quantity, vehicle_id,
                          layout_id, layout_unit_id, condition_state, installed_by,
                          installed_at, removed_by, removed_at, removal_disposition, notes)
  ```

  Enforce `tracking_mode IN ('quantity','individual')`, non-negative stock, reservation states
  `active|fulfilled|cancelled`, plan states `draft|review|published|archived`, lifecycle states
  `stored|reserved|installed|maintenance|retired`, and condition states
  `ready|maintenance_due|defective|unknown`. A reservation has exactly one of `layout_id` and
  `layout_unit_id`. An installation has exactly one of `vehicle_id`, `layout_id` and
  `layout_unit_id`. Individual allocations use quantity `1`.

- [ ] **Step 5: Add indexes and uniqueness rules.**

  Add unique location names per parent, unique product article numbers per manufacturer when the
  article number is not empty, unique asset inventory numbers when present, unique revision numbers
  per variant, one active published revision per variant, one stock row per product and location,
  and indexes for every foreign key used in list queries.

- [ ] **Step 6: Add domain enums and validation.**

  Define named string types and constants in `domain/accessory.go` and `domain/layout.go`. Add pure
  `Valid()` methods and `ValidateAllocationTarget(vehicleID, layoutID, unitID string) error`.
  Table-test every accepted and rejected enum value plus zero, multiple and valid targets.

- [ ] **Step 7: Verify migration and security baseline.**

  Run: `cd backend; go test ./internal/domain ./internal/infrastructure ./internal/application ./internal/api`

  Expected: PASS. Then run `gofmt` on changed Go files and `git diff --check`.

- [ ] **Step 8: Commit and update the GitHub work item.**

  Commit: `feat: add layout and accessory foundation`

## Task 2: Anlagen-Backend und Revisionsablauf

**Files:**

- Create: `backend/internal/application/layouts.go`
- Create: `backend/internal/application/layout_revisions.go`
- Create: `backend/internal/application/layouts_test.go`
- Create: `backend/internal/infrastructure/layout_repository.go`
- Create: `backend/internal/infrastructure/layout_repository_test.go`
- Create: `backend/internal/api/layout_handlers.go`
- Create: `backend/internal/api/layout_handlers_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/cmd/railkeeper/main.go`

**Interfaces:**

- Produces `application.LayoutService`.
- Read routes require Viewer, Planner writes require Planner.
- Produces Anlagen, Einheiten, Konfigurationen, Varianten und unveränderliche Revisionen.

- [ ] **Step 1: Define exact application types and repository contract.**

  Add `Layout`, `LayoutUnit`, `LayoutConfiguration`, `ConfigurationUnit`, `PlanVariant`,
  `PlanRevision` and matching input types. Every mutable input contains `ExpectedVersion int` on
  update. Define errors `ErrLayoutValidation`, `ErrLayoutNotFound`, `ErrLayoutVersionConflict`,
  `ErrPlanRevisionImmutable` and `ErrPlanRevisionConflict`.

  ```go
  type LayoutRepository interface {
      ListLayouts(context.Context) ([]Layout, error)
      GetLayout(context.Context, string) (*Layout, error)
      CreateLayout(context.Context, CreateLayoutInput, string) (*Layout, error)
      UpdateLayout(context.Context, string, UpdateLayoutInput, string) (*Layout, error)
      ListUnits(context.Context, string) ([]LayoutUnit, error)
      CreateUnit(context.Context, string, CreateLayoutUnitInput, string) (*LayoutUnit, error)
      UpdateUnit(context.Context, string, UpdateLayoutUnitInput, string) (*LayoutUnit, error)
      ListConfigurations(context.Context, string) ([]LayoutConfiguration, error)
      SaveConfiguration(context.Context, string, SaveLayoutConfigurationInput, string) (*LayoutConfiguration, error)
      ListVariants(context.Context, string) ([]PlanVariant, error)
      CreateVariant(context.Context, string, CreatePlanVariantInput, string) (*PlanVariant, error)
      CreateDraft(context.Context, string, CreatePlanRevisionInput, string) (*PlanRevision, error)
      SubmitRevision(context.Context, string, int, string) (*PlanRevision, error)
      PublishRevision(context.Context, string, int, string) (*PlanRevision, error)
  }
  ```

- [ ] **Step 2: Write failing service tests.**

  Cover private and club layouts, each unit kind, stale expected versions, configuration membership,
  monotonically increasing revision numbers, draft to review, draft or review to published,
  archiving the previously published revision and rejection of published revision mutation.

- [ ] **Step 3: Implement validation and orchestration.**

  Trim user text, require name, gauge and scale, accept layout kinds `private|club`, unit kinds
  `baseboard|module|segment|area`, reject negative dimensions, and normalize rotations into
  `[0,360)`. Keep SQL out of `application`.

- [ ] **Step 4: Implement the SQLite repository transactionally.**

  Use `UPDATE ... WHERE id=? AND version=?` for optimistic locking. Publication must archive the
  prior published revision and publish the target in one transaction. Write audit actions
  `LayoutCreated`, `LayoutUpdated`, `LayoutUnitCreated`, `LayoutConfigurationSaved`,
  `PlanDraftCreated`, `PlanRevisionSubmitted` and `PlanRevisionPublished` in the same transaction.

- [ ] **Step 5: Add explicit HTTP routes.**

  ```text
  GET    /api/v1/layouts                                      Viewer
  POST   /api/v1/layouts                                      Planner
  GET    /api/v1/layouts/{id}                                 Viewer
  PUT    /api/v1/layouts/{id}                                 Planner
  GET    /api/v1/layouts/{id}/units                           Viewer
  POST   /api/v1/layouts/{id}/units                           Planner
  PUT    /api/v1/layout-units/{id}                            Planner
  GET    /api/v1/layouts/{id}/configurations                  Viewer
  POST   /api/v1/layouts/{id}/configurations                  Planner
  PUT    /api/v1/layout-configurations/{id}                   Planner
  GET    /api/v1/layout-units/{id}/plan-variants              Viewer
  POST   /api/v1/layout-units/{id}/plan-variants              Planner
  POST   /api/v1/plan-variants/{id}/revisions                 Planner
  POST   /api/v1/plan-revisions/{id}/submit                   Planner
  POST   /api/v1/plan-revisions/{id}/publish                  Planner
  ```

  Map validation to 400, missing records to 404, stale versions to 409 and immutable revisions to
  409. Return stable problem codes such as `layout_version_conflict`.

- [ ] **Step 6: Test role and CSRF boundaries through the router.**

  Verify Viewer read only, Editor read only, Planner read and plan writes, Admin all, Messe denied,
  and every write rejected without a valid CSRF token.

- [ ] **Step 7: Wire services and run tests.**

  Construct `infrastructure.NewLayoutRepository(db)`, then `application.NewLayoutService(repo)` in
  `main.go`. Add `LayoutService` to `api.Config` and `App`. Run `cd backend; go test ./...`.

- [ ] **Step 8: Commit and update the GitHub work item.**

  Commit: `feat(layouts): add structure and plan revisions`

## Task 3: Zubehörkatalog, Lager und Einzelobjekte

**Files:**

- Create: `backend/internal/application/accessories.go`
- Create: `backend/internal/application/accessory_inventory.go`
- Create: `backend/internal/application/accessories_test.go`
- Create: `backend/internal/infrastructure/accessory_repository.go`
- Create: `backend/internal/infrastructure/accessory_repository_test.go`
- Create: `backend/internal/api/accessory_handlers.go`
- Create: `backend/internal/api/accessory_handlers_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/cmd/railkeeper/main.go`

**Interfaces:**

- Produces `application.AccessoryService` for products, locations, stock and assets.
- Viewer and Planner read; Editor changes inventory; Admin has all rights.

- [ ] **Step 1: Define exact types and repository contract.**

  Add `AccessoryProduct`, `StorageLocation`, `AccessoryStockSummary`, `AccessoryAsset`,
  `StockAdjustmentInput` and their create/update inputs. Monetary values remain decimal strings,
  consistent with vehicles. Define errors for validation, not found, conflict, insufficient stock
  and invalid tracking mode.

  ```go
  type AccessoryRepository interface {
      ListProducts(context.Context, string) ([]AccessoryProduct, error)
      GetProduct(context.Context, string) (*AccessoryProduct, error)
      CreateProduct(context.Context, CreateAccessoryProductInput, string) (*AccessoryProduct, error)
      UpdateProduct(context.Context, string, UpdateAccessoryProductInput, string) (*AccessoryProduct, error)
      ListLocations(context.Context) ([]StorageLocation, error)
      CreateLocation(context.Context, CreateStorageLocationInput, string) (*StorageLocation, error)
      AdjustStock(context.Context, string, StockAdjustmentInput, string) (*AccessoryStockSummary, error)
      GetStock(context.Context, string) (*AccessoryStockSummary, error)
      ListAssets(context.Context, string) ([]AccessoryAsset, error)
      CreateAsset(context.Context, string, CreateAccessoryAssetInput, string) (*AccessoryAsset, error)
      UpdateAsset(context.Context, string, UpdateAccessoryAssetInput, string) (*AccessoryAsset, error)
  }
  ```

- [ ] **Step 2: Write failing inventory tests.**

  Cover trimmed required fields, quantity versus individual mode, duplicate manufacturer/article,
  hierarchical locations, prevention of location cycles, positive and negative stock adjustments,
  rejection of negative balances, unique asset inventory numbers, valid condition transitions and
  rejection of assets for quantity products.

- [ ] **Step 3: Implement application validation.**

  Require manufacturer, name, category and tracking mode. Permit blank article number. Permit stock
  adjustments only for `quantity` products and asset creation only for `individual` products.
  Reject self-parenting and descendant-parenting locations.

- [ ] **Step 4: Implement SQLite persistence and audit.**

  Stock adjustment uses `INSERT ... ON CONFLICT ... DO UPDATE` inside a transaction after checking
  the resulting quantity is non-negative. Audit `AccessoryProductCreated`,
  `AccessoryProductUpdated`, `StorageLocationCreated`, `AccessoryStockAdjusted`,
  `AccessoryAssetCreated` and `AccessoryAssetUpdated` with quantity deltas but without sensitive
  free-text notes.

- [ ] **Step 5: Add HTTP routes.**

  ```text
  GET    /api/v1/accessory-products                            Viewer
  POST   /api/v1/accessory-products                            Editor
  GET    /api/v1/accessory-products/{id}                       Viewer
  PUT    /api/v1/accessory-products/{id}                       Editor
  GET    /api/v1/accessory-products/{id}/stock                 Viewer
  POST   /api/v1/accessory-products/{id}/stock-adjustments     Editor
  GET    /api/v1/accessory-products/{id}/assets                Viewer
  POST   /api/v1/accessory-products/{id}/assets                Editor
  PUT    /api/v1/accessory-assets/{id}                         Editor
  GET    /api/v1/storage-locations                             Viewer
  POST   /api/v1/storage-locations                             Editor
  ```

- [ ] **Step 6: Add API and role tests.**

  Test all error mappings, Viewer and Planner read access, Editor inventory writes, Planner write
  denial, Admin access, Messe denial and CSRF rejection.

- [ ] **Step 7: Wire services and verify.**

  Construct the repository and service in `main.go`, inject them through `api.Config`, run `gofmt`,
  `cd backend; go test ./...`, and `git diff --check`.

- [ ] **Step 8: Commit and update the GitHub work item.**

  Commit: `feat(accessories): add catalog and inventory`

## Task 4: Reservierung, Einbau und Ausbau

**Files:**

- Create: `backend/internal/application/accessory_allocations.go`
- Create: `backend/internal/application/accessory_allocations_test.go`
- Modify: `backend/internal/infrastructure/accessory_repository.go`
- Modify: `backend/internal/infrastructure/accessory_repository_test.go`
- Modify: `backend/internal/api/accessory_handlers.go`
- Modify: `backend/internal/api/accessory_handlers_test.go`
- Modify: `backend/internal/api/routes.go`

**Interfaces:**

- Planner und Editor können reservieren und stornieren.
- Nur Editor bestätigt Einbau, Ausbau und Betriebszustand.
- Produces berechnete Werte `owned`, `stored`, `reserved`, `installed`, `available`, `missing`.

- [ ] **Step 1: Extend the repository contract.**

  Add methods `ListReservations`, `CreateReservation`, `CancelReservation`, `ListInstallations`,
  `Install`, `RemoveInstallation` and `UpdateInstallationCondition`. Inputs use the explicit target:

  ```go
  type AllocationTargetInput struct {
      VehicleID    string `json:"vehicleId"`
      LayoutID     string `json:"layoutId"`
      LayoutUnitID string `json:"layoutUnitId"`
  }
  ```

- [ ] **Step 2: Write failing lifecycle tests.**

  Cover quantity reservation without stock movement, availability reduction, asset reservation,
  over-reservation rejection, cancellation, installation fulfilling a reservation, direct
  installation, atomic stock decrement, active installation uniqueness for an asset, condition
  change while installed, removal to storage, removal to maintenance, defect marking, retirement
  and full immutable history.

- [ ] **Step 3: Implement target validation.**

  Require exactly one target. Verify the referenced vehicle, layout or unit exists inside the same
  transaction before allocation. A unit target must belong to its referenced layout when both are
  supplied by an internal call. Public input still accepts only one concrete target.

- [ ] **Step 4: Implement reservation transactions.**

  Quantity reservation calculates available stock as stored minus active reservations. Individual
  reservation locks one stored asset. Cancellation changes status to `cancelled`; rows are never
  deleted. Planner actions write actor and audit metadata.

- [ ] **Step 5: Implement installation and removal transactions.**

  Installing quantity stock decrements the chosen stock row and creates an active installation.
  Installing an asset clears its storage location and sets lifecycle `installed`. Fulfilling a
  reservation changes it to `fulfilled` in the same transaction. Removal closes the active history
  row and either restores stock, sets asset lifecycle `maintenance`, records defect, or retires the
  asset according to `removalDisposition`.

- [ ] **Step 6: Add routes with mixed role authorization.**

  ```text
  GET    /api/v1/accessory-reservations                        Viewer
  POST   /api/v1/accessory-reservations                        EditorOrPlanner
  POST   /api/v1/accessory-reservations/{id}/cancel            EditorOrPlanner
  GET    /api/v1/accessory-installations                       Viewer
  POST   /api/v1/accessory-installations                       Editor
  POST   /api/v1/accessory-installations/{id}/remove           Editor
  PUT    /api/v1/accessory-installations/{id}/condition        Editor
  ```

  Add `routeAccessEditorOrPlanner` and resolve it to `RequireAnyRole("Editor", "Planner")` in route
  registration. Include this access type in negative authorization regression tests.

- [ ] **Step 7: Run service, route and full backend tests.**

  Run focused allocation tests first, then `cd backend; go test ./...`. Expected: PASS with no race
  between stock mutation, reservation state and history creation.

- [ ] **Step 8: Commit and update the GitHub work item.**

  Commit: `feat(accessories): track reservations and installations`

## Task 5: OpenAPI-Vertrag und typisierter Frontend-Client

**Files:**

- Modify: `openapi/railkeeper.yaml`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `frontend/src/shared/api.ts`
- Create: `frontend/src/shared/apiLayoutsAccessories.test.ts`

**Interfaces:**

- Exposes every route and JSON shape from Tasks 2 to 4.
- Produces typed `api.layouts*`, `api.accessory*`, `api.reservation*` and `api.installation*` methods.

- [ ] **Step 1: Add failing contract assertions.**

  Extend the Go contract test to require every new path and every new schema. Add Vitest request
  tests that assert encoded IDs, search queries, HTTP methods, JSON bodies and CSRF header behavior.

- [ ] **Step 2: Add OpenAPI tags and schemas.**

  Add tags `Accessories` and `Layouts`. Define schemas for every type and input from Tasks 2 to 4,
  including enum values, required properties, integer minimums, decimal string fields and 409
  problem responses for conflicts.

- [ ] **Step 3: Add all path operations.**

  Mirror the exact route list from Tasks 2 to 4. Document role expectations in operation
  descriptions and include `sessionCookie` security. Do not claim Messe access.

- [ ] **Step 4: Add frontend types without unchecked casts.**

  Reproduce the API JSON shapes as strict TypeScript types. Model enums as string unions and targets
  as a union that permits exactly one of `vehicleId`, `layoutId` or `layoutUnitId`.

- [ ] **Step 5: Add client methods.**

  Use the existing `request<T>` helper, `encodeURIComponent` for path IDs and explicit request bodies.
  Do not add a second HTTP abstraction.

- [ ] **Step 6: Verify contract and client.**

  Run `cd backend; go test ./internal/api`, then
  `cd ../frontend; npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts`, and finally
  `npm.cmd run build`.

- [ ] **Step 7: Commit and update the GitHub work item.**

  Commit: `feat(api): expose layout and accessory contracts`

## Task 6: Zubehör-Oberfläche

**Files:**

- Create: `frontend/src/features/accessories/AccessoriesView.tsx`
- Create: `frontend/src/features/accessories/AccessoryProductsPanel.tsx`
- Create: `frontend/src/features/accessories/AccessoryStockPanel.tsx`
- Create: `frontend/src/features/accessories/AccessoryAllocationsPanel.tsx`
- Create: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Create: `frontend/src/styles/accessories.css`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/app/Shell.tsx`
- Modify: `frontend/src/app/styles.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**

- Produces die Navigation `Zubehör / Accessories` für Admin, Editor, Viewer und Planner.
- Mutationen erscheinen nur für Admin und Editor; Reservieren zusätzlich für Planner.

- [ ] **Step 1: Write failing workflow tests.**

  Mock only the exported `api` boundary. Test initial load, search, empty/error/loading states,
  product creation, stock adjustment, asset creation, reservation by Planner, hidden installation
  action for Planner, installation and removal by Editor, and long German labels.

- [ ] **Step 2: Add route and navigation access.**

  Extend `AppView` with `accessories`, path `/accessories`, lazy load and role-aware view selection.
  Planner must land on an allowed general view, while a user with only Messe never sees the item.

- [ ] **Step 3: Build the dense operational layout.**

  Use one page header, compact search and tabs `Produkte`, `Bestand`, `Einzelobjekte`,
  `Reservierungen`, `Einbauhistorie`. Use existing buttons, inputs, tables, panels and design tokens.
  Do not use marketing cards or oversized typography.

- [ ] **Step 4: Implement product and stock workflows.**

  Product forms expose manufacturer, article number, name, category, tracking mode and description.
  Quantity products show per-location balance and adjustment dialog. Individual products show assets
  with inventory number, serial number, lifecycle, condition, purchase and warranty data.

- [ ] **Step 5: Implement allocation workflows.**

  Reservation requires product or asset, target and quantity. Installation requires source location,
  optional reservation and target. Removal requires destination or disposition. Every destructive or
  history-changing action has explicit confirmation and retains server errors in the dialog.

- [ ] **Step 6: Add complete German and English text.**

  Add matching translation keys for navigation, headings, tabs, fields, enums, actions, empty states,
  confirmations and error summaries. Extend the existing i18n parity test to reject a key present in
  only one locale.

- [ ] **Step 7: Verify behavior and presentation.**

  Run the focused Vitest file, complete frontend tests and build. Check desktop/mobile, light/dark,
  Viewer, Planner and Editor manually with long German values.

- [ ] **Step 8: Commit and update the GitHub work item.**

  Commit: `feat(frontend): add accessory workspace`

## Task 7: Anlagen-Oberfläche

**Files:**

- Create: `frontend/src/features/layouts/LayoutsView.tsx`
- Create: `frontend/src/features/layouts/LayoutWorkspace.tsx`
- Create: `frontend/src/features/layouts/LayoutModulesPanel.tsx`
- Create: `frontend/src/features/layouts/LayoutConfigurationsPanel.tsx`
- Create: `frontend/src/features/layouts/LayoutPlansPanel.tsx`
- Create: `frontend/src/features/layouts/LayoutsView.test.tsx`
- Create: `frontend/src/styles/layouts.css`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/app/Shell.tsx`
- Modify: `frontend/src/app/styles.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**

- Produces die Navigation `Anlagen / Layouts` für Admin, Editor, Viewer und Planner.
- Planner und Admin bearbeiten Struktur und Revisionen; Editor und Viewer lesen.

- [ ] **Step 1: Write failing workflow tests.**

  Cover private and club layout lists, creating a layout, adding modules, configuring a setup,
  creating a plan variant and draft, submitting and publishing a revision, stale-version conflict,
  read-only Editor/Viewer behavior and Messe denial.

- [ ] **Step 2: Add route and navigation integration.**

  Extend `AppView` with `layouts`, path `/layouts`, lazy loading and sidebar entry. Preserve stored
  sidebar ordering and filtering when older preferences do not contain the new view.

- [ ] **Step 3: Build the Anlagenarbeitsmappe.**

  Implement `Übersicht | Planer | Module | Aufbauten | Technik | Wartung | Dokumente`. In Etappe 1
  the active functional tabs are Übersicht, Module, Aufbauten and Planer metadata. Technik, Wartung
  and Dokumente display an honest localized empty state for later stages, not inactive fake controls.

- [ ] **Step 4: Implement module and setup management.**

  Support baseboard, module, segment and area, optional owner label, dimensions and archived state.
  Setup editing selects units and published plan revisions plus numeric position and rotation. It is
  a structured form in Etappe 1, not a graphical canvas.

- [ ] **Step 5: Implement revision workflow.**

  Show revision number, status, author and timestamps. Planner can create a draft from the active
  revision, mark it for review and publish it. Publishing requires an explicit summary confirmation.
  A 409 conflict keeps local form data and offers reload instead of silently overwriting it.

- [ ] **Step 6: Add German and English copy plus focused styles.**

  Use matching keys for all layout kinds, unit kinds, plan states, actions and empty states. Keep the
  interface dense, calm and compatible with light/dark and long German text.

- [ ] **Step 7: Verify behavior and presentation.**

  Run the focused Vitest file, complete frontend tests and build. Manually check compact and club
  layouts, all relevant roles, desktop/mobile and both themes.

- [ ] **Step 8: Commit and update the GitHub work item.**

  Commit: `feat(frontend): add layout workspace`

## Task 8: Backup, Restore und Etappe-1-Abnahme

**Files:**

- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`
- Modify: `backend/internal/api/smoke_test.go`
- Modify: `docs/architecture.md`
- Modify: `docs/roadmap.md`
- Modify: `frontend/src/features/importExport/ImportExportView.test.tsx`

**Interfaces:**

- Produces Backupversion 2 mit allen neuen Fachdaten.
- Consumes Version 1 und Version 2; Version 1 lässt neue Tabellen leer.
- Schließt alle Abnahmekriterien der Designspezifikation ab.

- [ ] **Step 1: Write failing backup compatibility tests.**

  Create data for product, stock, asset, layout, unit, configuration, plan revision, reservation and
  installation. Assert a version-2 export contains every new table, restore reproduces all links and
  auth data remains excluded. Build a literal version-1 fixture and assert successful validation and
  restore with empty new tables.

- [ ] **Step 2: Version the backup format.**

  Export `Version: 2`. Accept versions `1` and `2`. For version 1, treat every table introduced by
  migration 0039 as optional and empty. For version 2, require all new tables. Keep unknown tables and
  columns warnings unchanged.

- [ ] **Step 3: Order tables for foreign-key-safe restore.**

  Insert parents before children and clear children before parents. The required order is locations,
  products, stock/assets, layouts, units, variants/revisions, configurations/placements,
  reservations and installations, with vehicles restored before vehicle-targeted installations.

- [ ] **Step 4: Add end-to-end role smoke coverage.**

  Exercise one successful route per new read/write category for Admin, Editor, Viewer and Planner,
  plus negative Planner installation, negative Editor publish and complete Messe isolation.

- [ ] **Step 5: Update stable documentation.**

  Document the new application boundaries, role, backup content and current Etappe-1 limitations in
  `docs/architecture.md`. Move accessories and layout foundation from deferred to delivered in
  `docs/roadmap.md`; keep graphical planning in the next stage.

- [ ] **Step 6: Run the complete verification baseline.**

  ```powershell
  cd backend
  go test ./...

  cd ..\frontend
  npm.cmd run test:run
  npm.cmd run build
  ```

  Expected: all commands exit 0. Also run `git diff --check` and inspect OpenAPI/backend/frontend
  route parity.

- [ ] **Step 7: Perform manual acceptance.**

  Confirm every workflow listed under `Abnahme von Etappe 1` in the design spec. Export a backup,
  change data, restore it and verify products, stock, assets, layouts, modules, revisions,
  reservations and installation history.

- [ ] **Step 8: Commit, close work items and update the roadmap issue.**

  Commit: `feat: complete layout and accessory foundation`

  Close the eight Etappe-1 work items only after their PRs are merged. Close the Etappe-1 stage issue
  after the full acceptance run, then mark Etappe 2 as the next planned stage in the tracking issue.

## Plan Self-Review Result

- Every Etappe-1 requirement from the approved design has a task and an acceptance check.
- Graphical geometry, Tillig data, Flexgleis, elevations and foreign imports remain outside scope.
- Planner, Editor, Viewer, Messe and Admin rights are explicit at route, UI and test level.
- Database, application, API, client, UI, backup, documentation and audit changes are covered.
- The eight tasks match independently reviewable pull requests and public GitHub work items.
