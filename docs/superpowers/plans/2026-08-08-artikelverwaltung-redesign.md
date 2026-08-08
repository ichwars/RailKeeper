# Artikelverwaltung Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the technically split accessory area with a bilingual, article-centred inventory workflow that unifies product data, stock, purchases, documents, individual items, reservations, and installations.

**Architecture:** Extend the existing accessory foundation instead of introducing a second inventory subsystem. Keep article orchestration in `application`, SQLite transactions and queries in `infrastructure`, thin role-protected HTTP handlers in `api`, and a focused React feature composed around one overview and one tabbed editor dialog. Reuse generic master data and file blobs, preserve existing accessory IDs and history, and keep version-2 backups restorable while exporting the new schema as version 3.

**Tech Stack:** Go, `net/http`, SQLite via `modernc.org/sqlite`, React 19, TypeScript 7, Vite 8, Vitest, Testing Library, Lucide React, OpenAPI 3.

## Global Constraints

- Work in `C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\layout-accessory-foundation` on branch `dev/stage1-acceptance`.
- Preserve local-first operation, one Go process, SQLite, CSRF protection, audit logging, and the existing Admin/Editor/Planner/Viewer/Messe boundaries.
- Keep the current accessory, reservation, installation, and storage-location IDs. Migration `0041` must be forward-only and deterministic.
- Do not rebuild `accessory_products` merely to add the third stock strategy. Keep the legacy `tracking_mode` column for persistence compatibility and add `inventory_strategy`; map `quantity_later_individual` to legacy mode `quantity`.
- Drop the hard unique manufacturer/article-number index and replace it with a non-unique lookup index. Duplicate matches produce a warning response, not a database conflict.
- Use `AppSelect`, `AppDateInput`, and focused app-owned input shells. Native inputs may remain inside those shells but must not expose browser-specific styling or interaction.
- Use the existing vehicle dialog, table, focus, confirmation, tooltip, and transparent icon-button conventions.
- New UI text must exist in German and English. User-facing text must say `Einzelstück` and `Individual item`, never `Einzelobjekt`.
- Keep `frontend/src/shared/api.ts`, backend routes, and `openapi/railkeeper.yaml` synchronized in the same task that changes an endpoint.
- Do not implement layout planning or digital layout control. Existing vehicle/layout/layout-unit IDs are only usage targets.
- Do not stage `.superpowers/`, `frontend/dist`, `frontend/node_modules`, `data`, or `.cache`.
- Run all `go test` commands from `backend` and all `npm.cmd` commands from `frontend` unless a step states otherwise.

## Target Interfaces

Use these names consistently across application models, JSON, frontend types, and OpenAPI:

```go
type AccessoryInventoryStrategy string

const (
	AccessoryInventoryQuantity AccessoryInventoryStrategy = "quantity"
	AccessoryInventoryIndividual AccessoryInventoryStrategy = "individual"
	AccessoryInventoryQuantityLaterIndividual AccessoryInventoryStrategy = "quantity_later_individual"
)

type AccessoryArticleType string

const (
	AccessoryArticleTrack AccessoryArticleType = "track"
	AccessoryArticleSignal AccessoryArticleType = "signal"
	AccessoryArticleDecoder AccessoryArticleType = "decoder"
	AccessoryArticleElectrical AccessoryArticleType = "electrical_control"
	AccessoryArticleBuilding AccessoryArticleType = "building_equipment"
	AccessoryArticleLandscape AccessoryArticleType = "landscape_consumable"
	AccessoryArticleLighting AccessoryArticleType = "lighting"
	AccessoryArticleOther AccessoryArticleType = "other"
)
```

```ts
export type AccessoryInventoryStrategy =
  | "quantity"
  | "individual"
  | "quantity_later_individual";

export type AccessoryArticleType =
  | "track"
  | "signal"
  | "decoder"
  | "electrical_control"
  | "building_equipment"
  | "landscape_consumable"
  | "lighting"
  | "other";

export type AccessoryAttributeValue =
  | { key: string; kind: "text"; textValue: string }
  | { key: string; kind: "number"; numberValue: number; unit?: string }
  | { key: string; kind: "boolean"; booleanValue: boolean }
  | { key: string; kind: "date"; dateValue: string }
  | { key: string; kind: "single_select"; optionValues: [string] }
  | { key: string; kind: "multi_select"; optionValues: string[] };
```

The overview endpoint returns one stable shape:

```ts
export type AccessoryArticleListItem = {
  id: string;
  manufacturer: string;
  articleNumber?: string;
  name: string;
  articleType: AccessoryArticleType;
  subtype: string;
  gauges: string[];
  inventoryStrategy: AccessoryInventoryStrategy;
  archived: boolean;
  owned: number;
  available: number;
  reserved: number;
  installed: number;
  locationNames: string[];
  hasUsageHistory: boolean;
  careHintCount: number;
  updatedAt: string;
};
```

The list query accepts `query`, repeated `articleType`, `manufacturer`, repeated `gauge`, repeated `status`, `locationId`, `sort`, and `direction`. Valid status values are `available`, `reserved`, `installed`, `maintenance_due`, `defective`, and `archived`. Valid sort keys are `article`, `type`, `gauge`, `stock`, `storage`, and `updatedAt`; valid directions are `asc` and `desc`. Manufacturer status values are `announced`, `available`, `discontinued`, and `unknown`.

---

### Task 1: Add the forward-only article persistence model

**Files:**

- Create: `backend/migrations/0041_article_management_redesign.sql`
- Modify: `backend/internal/infrastructure/layout_accessory_schema_test.go`
- Create: `backend/internal/infrastructure/article_management_migration_test.go`

- [x] Add a failing migration test that starts at migration 0040, inserts one quantity product, one individual product, stock, an asset, a reservation, and an installation, applies 0041, and asserts every original ID and relation still exists.

```go
func TestArticleManagementMigrationPreservesAccessoryFoundation(t *testing.T) {
	db := migratedTestDB(t, 40)
	seedLegacyAccessoryGraph(t, db)
	applyMigrationFile(t, db, "0041_article_management_redesign.sql")
	assertRowCount(t, db, "accessory_products", 2)
	assertText(t, db, `SELECT inventory_strategy FROM accessory_products WHERE id='quantity-product'`, "quantity")
	assertText(t, db, `SELECT inventory_strategy FROM accessory_products WHERE id='individual-product'`, "individual")
	assertForeignKeyCheck(t, db)
}
```

- [x] Run `go test ./internal/infrastructure -run 'TestArticleManagementMigration'`; expect failure because migration 0041 does not exist.
- [x] Add columns to `accessory_products`: `ean`, `manufacturer_status`, `article_type`, `subtype`, `gauges_json`, `scale`, `package_quantity`, `stock_unit`, `minimum_stock`, `inventory_strategy`, `manufacturer_url`, `product_url`, `alternative_numbers_json`, `keywords_json`, `compatibility_notes`, `internal_notes`, and `archived`. Backfill `subtype=category`, infer `inventory_strategy` from legacy `tracking_mode`, and set `article_type='other'` for legacy rows.
- [x] Drop `ux_accessory_products_article`; create a non-unique normalized lookup index on manufacturer and article number plus indexes for article type, archive state, and EAN.
- [x] Create `accessory_product_attributes` with a typed-value check that permits exactly the columns matching `text`, `number`, `boolean`, `date`, `single_select`, or `multi_select`.
- [x] Create `accessory_stock_movements` with `movement_type` values `purchase`, `adjustment`, `transfer_in`, `transfer_out`, `individualization`, `installation`, and `removal`; include quantity, location, source type/ID, actor, note, and timestamp.
- [x] Create `accessory_purchases` and `accessory_documents`. Documents reference `file_blobs`; purchases reference product and optional destination location.
- [x] Add nullable `purchase_id` to `accessory_assets`. Add exact placement, digital address, decoder output, connection, and wiring-note columns to reservations/installations without changing the existing target XOR rules.
- [x] Seed generic master data entries for the eight article types and these subtype keys:
  - `track`: `straight`, `curve`, `flex`, `turnout`, `crossing`, `double_slip`, `transition`, `buffer_stop`
  - `signal`: `light`, `semaphore`, `main`, `distant`, `block`, `entry`, `exit`, `shunting`
  - `decoder`: `locomotive`, `function`, `accessory`, `switching`, `servo`, `feedback`
  - `electrical_control`: `turnout_drive`, `feedback`, `booster`, `power_supply`, `sensor`, `relay`, `distribution`, `control_element`
  - `building_equipment`: `building`, `platform`, `bridge`, `tunnel_portal`, `road_vehicle`, `figure`, `street_equipment`, `interior_equipment`
  - `landscape_consumable`: `grass`, `scatter`, `tree`, `water`, `paint`, `adhesive`, `ballast`, `wire`, `cable`, `fastener`
  - `lighting`: `lamp`, `led`, `light_strip`, `building_lighting`, `effect_lighting`
  - `other`: `other`
- [x] Seed stock units `piece`, `pack`, `meter`, `gram`, `milliliter` and controlled field kinds `text`, `number`, `boolean`, `date`, `single_select`, `multi_select`. Use deterministic IDs derived from the entry keys.
- [x] Extend the schema test to assert the new tables, indexes, checks, and seed counts.
- [x] Run `go test ./internal/infrastructure -run 'TestArticleManagementMigration|TestLayoutAccessorySchema'`; expect PASS.
- [x] Commit: `git add backend/migrations/0041_article_management_redesign.sql backend/internal/infrastructure/layout_accessory_schema_test.go backend/internal/infrastructure/article_management_migration_test.go && git commit -m "feat: add article management schema"`

### Task 2: Define article types, field validation, and duplicate warnings

**Files:**

- Modify: `backend/internal/domain/accessory.go`
- Create: `backend/internal/domain/accessory_attributes.go`
- Modify: `backend/internal/domain/accessory_test.go`
- Create: `backend/internal/domain/accessory_attributes_test.go`
- Modify: `backend/internal/application/accessories.go`
- Modify: `backend/internal/application/accessories_test.go`

- [x] Add table-driven failing tests for all three inventory strategies, all eight article types, the typed attribute union, permitted standard keys, and custom fields allowed only for `other`.
- [x] Define the standard field catalog in Go. The exact keys are:
  - `track`: `trackSystem`, `lengthMm`, `radiusMm`, `angleDegrees`, `direction`, `frogAngleDegrees`, `sleeperType`, `railHeightMm`, `roadbed`, `connectionCount`, `digitalReady`
  - `signal`: `prototype`, `epoch`, `aspects`, `ledCount`, `heightMm`, `voltageAC`, `voltageDC`, `mounting`, `driveType`, `integratedDecoder`, `controlModule`
  - `decoder`: `interface`, `protocols`, `functionOutputs`, `motorCurrentMa`, `outputCurrentMa`, `totalCurrentMa`, `railCom`, `susi`, `servoOutputs`, `dimensions`, `firmware`
  - `electrical_control`: `inputVoltage`, `outputVoltage`, `currentA`, `powerW`, `channelCount`, `protocols`, `connectors`, `protections`, `compatibleArticles`
  - `building_equipment`: `epoch`, `dimensions`, `footprint`, `material`, `constructionType`, `partCount`, `difficulty`, `lightingOptions`, `floorPlanAvailable`
  - `landscape_consumable`: `material`, `color`, `season`, `content`, `contentUnit`, `fiberOrGrainSize`, `coverage`, `suitableScales`, `safetyNotes`
  - `lighting`: `lightColor`, `colorTemperatureK`, `voltage`, `currentMa`, `powerType`, `ledCount`, `dimmable`, `dimensions`, `mounting`
- [x] Expand `AccessoryProduct` and input models with the common-core fields and `[]AccessoryAttributeValue`. Define the request/result types for the explicit pre-save duplicate check; Task 3 connects them to the catalogue query so no interface depends on types introduced by a later task:

```go
type AccessoryDuplicateCheckInput struct {
	Manufacturer  string `json:"manufacturer"`
	ArticleNumber string `json:"articleNumber"`
	ExcludeID     string `json:"excludeId"`
}

type AccessoryDuplicateCheckResult struct {
	Candidates []AccessoryDuplicateCandidate `json:"candidates"`
}
```

- [x] Validate required core fields, article-type/subtype consistency, standard attribute kinds, positive package quantity, non-negative minimum stock, and normalized arrays. Preserve a clear `ErrAccessoryValidation` result.
- [x] Run `go test ./internal/domain ./internal/application -run 'Accessory|Article'`; expect PASS.
- [x] Commit: `git add backend/internal/domain/accessory.go backend/internal/domain/accessory_attributes.go backend/internal/domain/accessory_test.go backend/internal/domain/accessory_attributes_test.go backend/internal/application/accessories.go backend/internal/application/accessories_test.go && git commit -m "feat: define typed article catalog"`

### Task 3: Implement the filterable article catalogue and global metrics

**Files:**

- Modify: `backend/internal/infrastructure/accessory_repository.go`
- Modify: `backend/internal/infrastructure/accessory_repository_test.go`
- Create: `backend/internal/infrastructure/accessory_article_query.go`
- Create: `backend/internal/application/accessory_overview.go`
- Create: `backend/internal/application/accessory_overview_test.go`

- [x] Add failing repository tests covering search by manufacturer/article number/EAN/name, each filter, every sort key and direction, archived visibility, multiple locations, and a product with mixed reserved/installed amounts.
- [x] Implement `AccessoryArticleListQuery`, `AccessoryArticleListResult`, and `AccessoryOverviewMetrics`. Metrics must be global and computed independently from table selection:

```go
type AccessoryOverviewMetrics struct {
	ArticleCount      int `json:"articleCount"`
	ArticleTypeCount  int `json:"articleTypeCount"`
	Available         int `json:"available"`
	LocationCount     int `json:"locationCount"`
	Reserved          int `json:"reserved"`
	Installed         int `json:"installed"`
	CareHintCount     int `json:"careHintCount"`
}
```

- [x] Split the repository interface by capability now that the catalogue query types exist:

```go
type AccessoryCatalogRepository interface {
	ListArticles(context.Context, AccessoryArticleListQuery) (*AccessoryArticleListResult, error)
	GetProduct(context.Context, string) (*AccessoryProduct, error)
	FindDuplicateCandidates(context.Context, string, string, string) ([]AccessoryDuplicateCandidate, error)
	CreateProduct(context.Context, CreateAccessoryProductInput, string) (*AccessoryProduct, error)
	UpdateProduct(context.Context, string, UpdateAccessoryProductInput, string) (*AccessoryProduct, error)
}
```

- [x] Build parameterized SQL with an explicit sort-key whitelist. Never interpolate raw request values. Use aggregate subqueries for quantity stock, assets, active reservations, active installations, storage names, and `hasUsageHistory`.
- [x] Define care hints as missing manufacturer, article number, article type, gauge for gauge-relevant types, or stock unit. Return the per-row count and global count.
- [x] Load typed attributes with one batched query for the returned product IDs, not one query per product.
- [x] Add tests proving the duplicate lookup ignores the current product during update and returns variants without preventing save.
- [x] Add `AccessoryService.CheckDuplicateProducts`, trim and normalize its input, and return candidates without converting equality into `ErrAccessoryConflict`.
- [x] Run `go test ./internal/infrastructure ./internal/application -run 'AccessoryArticle|AccessoryOverview|Duplicate'`; expect PASS.
- [x] Commit: `git add backend/internal/infrastructure/accessory_repository.go backend/internal/infrastructure/accessory_repository_test.go backend/internal/infrastructure/accessory_article_query.go backend/internal/application/accessory_overview.go backend/internal/application/accessory_overview_test.go && git commit -m "feat: query article overview"`

### Task 4: Make purchases, stock movements, transfers, and individualization transactional

**Files:**

- Modify: `backend/internal/application/accessory_inventory.go`
- Create: `backend/internal/application/accessory_inventory_test.go`
- Create: `backend/internal/application/accessory_purchases.go`
- Create: `backend/internal/application/accessory_purchases_test.go`
- Modify: `backend/internal/infrastructure/accessory_inventory_repository.go`
- Modify: `backend/internal/infrastructure/accessory_repository_test.go`

- [x] Add failing application/repository tests for manual adjustment journals, transfers, quantity purchases, individual purchases, hybrid individualization, insufficient stock rollback, and concurrent decrements that never produce negative stock.
- [x] Introduce these commands:

```go
type TransferAccessoryStockInput struct {
	FromLocationID string `json:"fromLocationId"`
	ToLocationID   string `json:"toLocationId"`
	Quantity       int    `json:"quantity"`
	Note           string `json:"note"`
}

type CreateAccessoryPurchaseInput struct {
	PurchasedAt      string `json:"purchasedAt"`
	Supplier         string `json:"supplier"`
	Quantity         int    `json:"quantity"`
	UnitPrice        string `json:"unitPrice"`
	Currency         string `json:"currency"`
	InvoiceNumber    string `json:"invoiceNumber"`
	WarrantyUntil    string `json:"warrantyUntil"`
	StorageLocationID string `json:"storageLocationId"`
	BookToStock      bool   `json:"bookToStock"`
	Notes            string `json:"notes"`
}

type IndividualizeAccessoryInput struct {
	LocationID string                    `json:"locationId"`
	Asset      CreateAccessoryAssetInput `json:"asset"`
}
```

- [x] Extend the repository with `ListStockMovements`, `TransferStock`, `ListPurchases`, `CreatePurchase`, and `Individualize`. All mutating methods receive the actor and write audit entries.
- [x] Make adjustment insert a movement in the same transaction. Make transfer decrement and increment with paired movement rows in one transaction.
- [x] Make a booked quantity/hybrid purchase create the purchase, increment stock, and create a purchase movement atomically. Make a booked individual purchase create the requested number of stored assets, linked by `purchase_id`, atomically.
- [x] Allow individualization only for `quantity_later_individual`. Lock through the SQLite write transaction, decrement one unit from the chosen location, create one stored asset, and write an `individualization` movement. Roll back all writes on any conflict.
- [x] Keep reservations availability-only. Installation/removal continue to own physical movement and must write journal rows without double counting.
- [x] Run `go test ./internal/application ./internal/infrastructure -run 'Purchase|StockMovement|Transfer|Individualiz|Concurrent'`; expect PASS.
- [x] Commit: `git add backend/internal/application/accessory_inventory.go backend/internal/application/accessory_inventory_test.go backend/internal/application/accessory_purchases.go backend/internal/application/accessory_purchases_test.go backend/internal/infrastructure/accessory_inventory_repository.go backend/internal/infrastructure/accessory_repository_test.go && git commit -m "feat: add transactional article inventory"`

### Task 5: Add article documents and complete usage history

**Files:**

- Create: `backend/internal/application/accessory_documents.go`
- Create: `backend/internal/application/accessory_documents_test.go`
- Create: `backend/internal/infrastructure/accessory_document_repository.go`
- Modify: `backend/internal/application/accessory_allocations.go`
- Modify: `backend/internal/application/accessory_allocations_test.go`
- Modify: `backend/internal/infrastructure/accessory_allocation_repository.go`
- Modify: `backend/internal/infrastructure/accessory_installation_repository.go`
- Modify: `backend/internal/infrastructure/accessory_allocation_repository_test.go`
- Modify: `backend/internal/application/file_blobs.go`
- Create: `backend/internal/application/file_blobs_test.go`

- [x] Add failing tests for safe article document metadata, file-blob reference retention/deletion, reservation placement, installation technical data, condition events, removal result, and history remaining visible after removal.
- [x] Define document categories `invoice`, `delivery_note`, `manual`, `data_sheet`, `floor_plan`, `image`, and `other`; enforce the existing upload size, extension, executable, MIME, and path-confinement rules.
- [x] Store document metadata in `accessory_documents` and bytes in `file_blobs`. Image documents carry `is_primary`; the full product response exposes the primary document download URL as `primaryImageUrl`. Update `FileBlobService` reference counting, filesystem migration checks, and cleanup so an article document can never orphan or prematurely delete a blob.
- [x] Expand reservation/installation inputs with `placement`, `digitalAddress`, `decoderOutput`, `connection`, and `wiringNotes`. Keep decoder CVs out of product attributes.
- [x] Add `AccessoryUsageHistory` that merges reservations, installations, condition changes, and removals in descending timestamp order. Do not include pure stock movements.
- [x] Keep `hasUsageHistory=true` once any reservation or installation row exists, regardless of current status.
- [x] Run `go test ./internal/application ./internal/infrastructure -run 'AccessoryDocument|AccessoryUsage|FileBlob|Allocation|Installation'`; expect PASS.
- [x] Commit: `git add backend/internal/application/accessory_documents.go backend/internal/application/accessory_documents_test.go backend/internal/infrastructure/accessory_document_repository.go backend/internal/application/accessory_allocations.go backend/internal/application/accessory_allocations_test.go backend/internal/infrastructure/accessory_allocation_repository.go backend/internal/infrastructure/accessory_installation_repository.go backend/internal/infrastructure/accessory_allocation_repository_test.go backend/internal/application/file_blobs.go backend/internal/application/file_blobs_test.go && git commit -m "feat: add article documents and usage history"`

### Task 6: Publish the complete role-safe API and OpenAPI contract

**Files:**

- Modify: `backend/internal/api/routes.go`
- Modify: `backend/internal/api/accessory_handlers.go`
- Modify: `backend/internal/api/accessory_handlers_test.go`
- Modify: `backend/internal/api/accessory_allocation_handlers.go`
- Modify: `backend/internal/api/accessory_allocation_handlers_test.go`
- Create: `backend/internal/api/accessory_document_handlers.go`
- Create: `backend/internal/api/accessory_document_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

- [x] Add failing route tests for Admin, Editor, Planner, Viewer, and Messe on every new read/write endpoint, including missing CSRF on writes. Assert Messe gets 403 throughout.
- [x] Change `GET /accessory-products` to return `{items, metrics, filters}` and parse the approved list query. Keep `GET /accessory-products/{id}` as the full editor payload.
- [x] Add endpoints:
  - `POST /accessory-products/duplicate-check`
  - `POST /accessory-products/{id}/archive`
  - `POST /accessory-products/{id}/restore`
  - `GET /accessory-products/{id}/stock-movements`
  - `POST /accessory-products/{id}/stock-transfers`
  - `GET|POST /accessory-products/{id}/purchases`
  - `POST /accessory-products/{id}/individualizations`
  - `GET|POST /accessory-products/{id}/documents`
  - `GET|PUT|DELETE /accessory-products/{id}/documents/{documentID}`
  - `GET /accessory-products/{id}/documents/{documentID}/download`
  - `GET /accessory-products/{id}/usage-history`
- [x] Protect catalogue, stock, purchase, asset, document, and installation writes with `routeAccessEditor`; reservations remain `routeAccessEditorOrPlanner`; all article reads use `routeAccessViewer`, which excludes Messe.
- [x] Return 400 for invalid field/type combinations, 404 for missing IDs, 409 for insufficient stock or state conflicts, and 413/415 for rejected uploads. Duplicate check returns candidate rows with 200; create/update may still save a confirmed duplicate and never return 409 solely for manufacturer/article-number equality.
- [x] Document every request/response schema, query enum, multipart upload, role behavior, and problem code in OpenAPI. Extend the OpenAPI contract test to require each route and schema.
- [x] Run `go test ./internal/api -run 'Accessory|OpenAPI|Route'`; expect PASS.
- [x] Commit: `git add backend/internal/api/routes.go backend/internal/api/accessory_handlers.go backend/internal/api/accessory_handlers_test.go backend/internal/api/accessory_allocation_handlers.go backend/internal/api/accessory_allocation_handlers_test.go backend/internal/api/accessory_document_handlers.go backend/internal/api/accessory_document_handlers_test.go backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml && git commit -m "feat: expose article management api"`

### Task 7: Preserve all article data in backup version 3

**Files:**

- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

- [x] Add a failing round-trip test that exports and restores product attributes, purchases, documents and blob bytes, movements, individual items, reservations, installations, and removed usage history.
- [x] Set `backupVersion = 3`. Add the new tables to `backupTableOrder` after their parent tables and before dependent usage tables.
- [x] Keep version-1 and version-2 documents valid. Mark version-3 tables optional only when validating older versions, not when validating a version-3 backup.
- [x] Verify restore preflight rejects future versions and malformed typed attributes before deleting or replacing existing application data.
- [x] Assert authentication, sessions, password hashes, rate limits, and audit logs remain excluded.
- [x] Run `go test ./internal/application -run 'Backup'`; expect PASS.
- [x] Commit: `git add backend/internal/application/backup.go backend/internal/application/backup_test.go && git commit -m "feat: include article management in backups"`

### Task 8: Extend the typed frontend API and app-owned controls

**Files:**

- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/shared/apiLayoutsAccessories.test.ts`
- Modify: `frontend/src/shared/api.ts`
- Create: `frontend/src/shared/ui/AppTextInput.tsx`
- Create: `frontend/src/shared/ui/AppTextInput.test.tsx`
- Create: `frontend/src/shared/ui/AppNumberInput.tsx`
- Create: `frontend/src/shared/ui/AppNumberInput.test.tsx`
- Create: `frontend/src/shared/ui/AppFilePicker.tsx`
- Create: `frontend/src/shared/ui/AppFilePicker.test.tsx`
- Create: `frontend/src/shared/ui/AppMultiSelect.tsx`
- Create: `frontend/src/shared/ui/AppMultiSelect.test.tsx`
- Modify: `frontend/src/styles/forms-controls.css`

- [x] Add failing API tests for URL encoding, repeated filters, sort/direction, multipart documents, purchase creation, transfer, individualization, archive/restore, and history.
- [x] Replace the old product type with the target interfaces from this plan. Use discriminated unions for attribute values and write inputs. Do not use `any` or unchecked casts.
- [x] Add app-owned text, number, file, and multi-select shells with labels, help, errors, disabled/read-only state, described-by wiring, focus styling, and forwarded refs. `AppNumberInput` exposes a string to preserve incomplete decimal input and parses only on submit.
- [x] Keep hidden native file input and keyboard-operable trigger inside `AppFilePicker`; expose chosen filename and a clear action. Keep native text/number semantics inside the styled shells.
- [x] Add all new API methods to `createLayoutsAccessoriesAPI` and re-export types through `api.ts`.
- [x] Run `npm.cmd run test:run -- src/shared/apiLayoutsAccessories.test.ts src/shared/ui/AppTextInput.test.tsx src/shared/ui/AppNumberInput.test.tsx src/shared/ui/AppFilePicker.test.tsx src/shared/ui/AppMultiSelect.test.tsx`; expect PASS.
- [x] Run `npm.cmd run build`; expect `tsc -b && vite build` to exit 0.
- [x] Commit: `git add frontend/src/shared/apiLayoutsAccessories.ts frontend/src/shared/apiLayoutsAccessories.test.ts frontend/src/shared/api.ts frontend/src/shared/ui/AppTextInput.tsx frontend/src/shared/ui/AppTextInput.test.tsx frontend/src/shared/ui/AppNumberInput.tsx frontend/src/shared/ui/AppNumberInput.test.tsx frontend/src/shared/ui/AppFilePicker.tsx frontend/src/shared/ui/AppFilePicker.test.tsx frontend/src/shared/ui/AppMultiSelect.tsx frontend/src/shared/ui/AppMultiSelect.test.tsx frontend/src/styles/forms-controls.css && git commit -m "feat: add typed article ui controls"`

### Task 9: Move storage locations and article master data into Settings

**Files:**

- Modify: `frontend/src/features/settings/settingsModel.ts`
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Create: `frontend/src/features/settings/SettingsView.test.tsx`
- Create: `frontend/src/features/settings/ArticleManagementSettings.tsx`
- Create: `frontend/src/features/settings/ArticleManagementSettings.test.tsx`
- Move: `frontend/src/features/accessories/AccessoryLocationsPanel.tsx` to `frontend/src/features/settings/StorageLocationsSettings.tsx`
- Move: `frontend/src/features/accessories/accessoryLocations.ts` to `frontend/src/features/settings/storageLocations.ts`
- Move: `frontend/src/features/accessories/accessoryLocations.test.ts` to `frontend/src/features/settings/storageLocations.test.ts`
- Modify: `frontend/src/styles/settings.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

- [x] Add failing settings tests that open `Einstellungen → Artikelverwaltung`, manage a hierarchical location, and show manufacturer, unit, type/subtype, and controlled custom-field sections.
- [x] Add `articleManagement` to `SettingsTab` and `settingsTabs`. Mount `ArticleManagementSettings` from `SettingsView` without adding more state to the already large central settings component.
- [x] Reuse the existing generic master-data API for `manufacturer`, `accessory_unit`, `accessory_type`, `accessory_subtype`, and `accessory_custom_field`. Standard type keys remain stable; settings may change labels/active state, not their keys. Admin/Editor may mutate locations and article master data; Planner/Viewer receive the same understandable read-only presentation.
- [x] Move location create/edit/archive/reactivate UI into `StorageLocationsSettings`. Keep parent-cycle and archived-parent errors visible.
- [x] Remove every storage-location administration link or form from the accessory feature. Selection controls remain in article filters and dialogs.
- [x] Run `npm.cmd run test:run -- src/features/settings/SettingsView.test.tsx src/features/settings/ArticleManagementSettings.test.tsx src/features/settings/storageLocations.test.ts`; expect PASS.
- [x] Commit: `git add -A frontend/src/features/settings frontend/src/features/accessories/AccessoryLocationsPanel.tsx frontend/src/features/accessories/accessoryLocations.ts frontend/src/features/accessories/accessoryLocations.test.ts frontend/src/styles/settings.css frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts && git commit -m "feat: move article settings into settings"`

### Task 10: Replace accessory subpages with the article overview table

**Files:**

- Modify: `frontend/src/app/Shell.tsx`
- Create: `frontend/src/app/Shell.test.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Create: `frontend/src/features/accessories/ArticleOverviewHeader.tsx`
- Create: `frontend/src/features/accessories/ArticleMetrics.tsx`
- Create: `frontend/src/features/accessories/ArticleToolbar.tsx`
- Create: `frontend/src/features/accessories/ArticleTable.tsx`
- Create: `frontend/src/features/accessories/ArticleTable.test.tsx`
- Create: `frontend/src/features/accessories/useArticleOverview.ts`
- Create: `frontend/src/features/accessories/useArticleOverview.test.tsx`
- Delete: `frontend/src/features/accessories/AccessoryProductsPanel.tsx`
- Modify: `frontend/src/styles/accessories.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

- [x] Add failing tests for `Fahrzeugbestand`/`Artikelübersicht`, no accessory tab list, no card toggle, global metrics, instant search, each filter, filter reset, result count, empty/error/loading states, and viewer/planner read-only explanation.
- [x] Implement a single article overview composed from the focused components above. Do not retain invisible product selection state.
- [x] Render the four approved compact metric panels. Clicking a metric may set a visible filter but must never depend on row selection.
- [x] Render only the table columns `Artikel`, `Art / Unterart`, `Spur`, `Bestand`, `Lagerung`, and `Aktionen`.
- [x] Make sortable headers real buttons with `aria-sort` on their `th`. Preserve active sort key and direction in `useArticleOverview`.
- [x] Add transparent `Eye`, `Pencil`, and `MoreHorizontal` action buttons with localized accessible names and tooltips. The menu contains archive/restore. Hide write actions for Viewer/Planner while preserving view.
- [x] Make `+ Neuer Artikel` and `Ersten Artikel anlegen` available only to Admin/Editor. Long manufacturer/name/location text must truncate with an accessible full-value tooltip.
- [x] Run `npm.cmd run test:run -- src/app/Shell.test.tsx src/features/accessories/AccessoriesView.test.tsx src/features/accessories/ArticleTable.test.tsx src/features/accessories/useArticleOverview.test.tsx`; expect PASS.
- [x] Commit: `git add -A frontend/src/app/Shell.tsx frontend/src/app/Shell.test.tsx frontend/src/features/accessories frontend/src/styles/accessories.css frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts && git commit -m "feat: build article overview"`

### Task 11: Build the shared create, view, and edit article dialog

**Files:**

- Create: `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- Create: `frontend/src/features/accessories/ArticleEditorDialog.test.tsx`
- Create: `frontend/src/features/accessories/useArticleEditorController.ts`
- Create: `frontend/src/features/accessories/useArticleEditorController.test.tsx`
- Create: `frontend/src/features/accessories/articleEditorModel.ts`
- Create: `frontend/src/features/accessories/ArticleCoreTab.tsx`
- Create: `frontend/src/features/accessories/ArticleStockTab.tsx`
- Create: `frontend/src/features/accessories/ArticlePurchaseDocumentsTab.tsx`
- Create: `frontend/src/features/accessories/ArticleUsageHistoryTab.tsx`
- Refactor: `frontend/src/features/accessories/AccessoryStockPanel.tsx`
- Refactor: `frontend/src/features/accessories/AccessoryReservationsPanel.tsx`
- Refactor: `frontend/src/features/accessories/AccessoryInstallationsPanel.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/styles/accessories.css`

- [x] Add failing dialog tests for create/view/edit modes, focus trap and return, tab persistence, dirty-close confirmation, whole-form validation, tab error badges, pre-save duplicate warning confirmation, and failed-save value preservation.
- [x] Mirror `VehicleEditorDialog` structure with one modal shell and horizontal tabs. Fixed tabs are `article`, `stock`, and `purchaseDocuments`; append exactly one subject tab; append `usageHistory` only when `hasUsageHistory` is true.
- [x] In `ArticleCoreTab`, render the approved common fields and a collapsed `Weitere Angaben`. Keep product links, alternate numbers, keywords, compatibility, and internal notes there. Never put placement, target, digital address, or wiring there.
- [x] In `ArticleStockTab`, integrate strategy, minimum stock, locations, adjustment, transfer, individual items, hybrid individualization, and a compact stock journal. Location controls select only.
- [x] In `ArticlePurchaseDocumentsTab`, separate purchase entry/list from safe document upload/list. A booked purchase must submit one command, not a client-side purchase followed by a second stock request.
- [x] In `ArticleUsageHistoryTab`, show current reservations/installations first and chronological history below. Reuse and reshape the current reservation/installation components instead of duplicating their domain logic.
- [x] Disable every form control in view mode. Planner can create/cancel reservations only; Admin/Editor can perform all mutations; Viewer is read-only.
- [x] Run `npm.cmd run test:run -- src/features/accessories/ArticleEditorDialog.test.tsx src/features/accessories/useArticleEditorController.test.tsx src/features/accessories/AccessoriesView.test.tsx`; expect PASS.
- [x] Commit: `git add -A frontend/src/features/accessories frontend/src/styles/accessories.css && git commit -m "feat: add article editor workflow"`

### Task 12: Implement the dynamic subject tab and controlled custom fields

**Files:**

- Create: `frontend/src/features/accessories/articleTypeFields.ts`
- Create: `frontend/src/features/accessories/articleTypeFields.test.ts`
- Create: `frontend/src/features/accessories/ArticleSubjectTab.tsx`
- Create: `frontend/src/features/accessories/ArticleSubjectTab.test.tsx`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- Modify: `frontend/src/features/accessories/articleEditorModel.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

- [x] Add failing tests proving each of the eight article types renders only its own field definitions, changing type warns before discarding values, and `other` supports only configured typed custom fields.
- [x] Encode the exact field keys from Task 2 in a readonly registry. Each definition contains `key`, `kind`, label key, optional help key, optional unit/options, and validation bounds. TypeScript must fail compilation if a registry uses an unsupported kind.
- [x] Render one `ArticleSubjectTab` using the registry and app-owned controls. Do not create eight separate page tabs or eight mostly duplicate form components.
- [x] For `other`, load active `accessory_custom_field` master data. Respect text, number+unit, boolean, date, single-select, and multi-select kinds and their configured options.
- [x] Add a fixture test for Tillig TT Modellgleis 83101 with article number, TT gauge, track system, straight subtype, length, connections, package/unit, stock location, and quantity.
- [x] Run `npm.cmd run test:run -- src/features/accessories/articleTypeFields.test.ts src/features/accessories/ArticleSubjectTab.test.tsx src/features/accessories/ArticleEditorDialog.test.tsx`; expect PASS.
- [x] Commit: `git add frontend/src/features/accessories/articleTypeFields.ts frontend/src/features/accessories/articleTypeFields.test.ts frontend/src/features/accessories/ArticleSubjectTab.tsx frontend/src/features/accessories/ArticleSubjectTab.test.tsx frontend/src/features/accessories/ArticleEditorDialog.tsx frontend/src/features/accessories/articleEditorModel.ts frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts && git commit -m "feat: add typed article subject fields"`

### Task 13: Complete integration, accessibility, visual QA, and cleanup

**Files:**

- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/styles/accessories.css`
- Modify: `frontend/src/styles/forms-controls.css`
- Modify: `frontend/src/styles/settings.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `docs/superpowers/specs/2026-08-08-artikelverwaltung-redesign-design.md`

- [x] Add one frontend integration test that creates the Tillig article, books a purchase into stock, individualizes one unit, reserves and installs it, then verifies `Verwendung & Historie` remains visible after removal.
- [x] Add read-only integration coverage for Planner, Viewer, and Messe. Messe must never receive the route/view; Planner may reserve but cannot mutate product, stock, purchase, document, asset, or installation data.
- [x] Search and remove old UI concepts: `rg -n "Einzelobjekt|Products|Produkte|accessories.tabs|Karten|card view" frontend/src`. Expected result: no obsolete user-facing accessory navigation or view toggle.
- [x] Run `gofmt -w` on every changed Go file.
- [x] Run backend verification from `backend`: `go test ./...`; expect PASS.
- [x] Run frontend verification from `frontend`: `npm.cmd run test:run`; expect PASS.
- [x] Run frontend production build: `npm.cmd run build`; expect exit code 0.
- [x] Run `git diff --check`; expect no output.
- [x] Start the local app using the documented repo-local `GOCACHE`, then visually verify `/accessories` and `/settings?tab=articleManagement` at desktop and mobile widths in German/English and light/dark themes.
- [x] Visual acceptance checklist:
  - table-only overview, no submenus or view switch
  - sortable headers and right-aligned action column
  - restrained typography and aligned controls
  - transparent icon hover/focus states
  - dialog focus trap, dirty close, long German labels, error indicators
  - one dynamic subject tab
  - conditional usage/history tab
  - storage administration only in Settings
- [x] Update the design spec status to `Implemented and verified` only after all automated and visual checks pass.
- [x] Commit: `git add backend frontend openapi docs/superpowers/specs/2026-08-08-artikelverwaltung-redesign-design.md && git commit -m "test: verify article management redesign"`

## Final Delivery Gate

- [x] Confirm `git status --short` contains no generated output, local data, caches, or `.superpowers/` staging.
- [x] Confirm the OpenAPI path inventory equals the registered article routes.
- [x] Confirm a version-2 backup validates and restores, and a version-3 round-trip preserves every new article table and blob.
- [x] Confirm no quantity can become negative and no unit is counted as both quantity and individual item.
- [x] Confirm all 13 acceptance criteria in the approved specification have an automated test or an explicit visual-check result.
- [x] Prepare the PR summary in German and English with screenshots of overview, create dialog, dynamic Tillig track fields, and Settings storage locations.
