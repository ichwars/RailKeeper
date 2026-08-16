# Vehicle Operational Data and Valuations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement GitHub issues #78 and #86 by adding vehicle operational data, an accessory per-unit list price, and a cent-exact four-part valuation block on the overview. Issue #82 remains explicitly out of scope.

**Architecture:** Extend the existing forward-only SQLite schema and current vehicle/accessory CRUD paths. Add a focused application service that reads valuation source rows in bulk and calculates all amounts with integer cents. Expose the result through a viewer-only overview endpoint and render it in a dedicated React component whose loading and failure state is independent of the existing overview data.

**Tech Stack:** Go, `database/sql`, SQLite, `net/http`, React, TypeScript, Vite, Vitest, Testing Library, CSS, OpenAPI 3.

## Global Constraints

- Do not change, implement, or close issue #82.
- Preserve existing user data and forward-only migration behavior. Do not rewrite historical migrations.
- Keep monetary storage compatible with existing text fields. Parse for aggregation without rewriting old values.
- Calculate and multiply amounts as `int64` cents. Do not use floating point for persisted or aggregated money.
- Treat blank currency as EUR. Exclude explicit non-EUR accessory purchases and report their count.
- Keep all new vehicle and accessory table columns hidden by default for existing and new preferences.
- Keep Messe users away from valuation data by protecting the endpoint with `routeAccessViewer`.
- Update German and English translations together.
- Use `gofmt` for Go changes and keep TypeScript strict.
- Run each red test before its implementation and each green test after it.

---

### Task 1: Persist and validate the new vehicle operational fields

**Files:**

- Create: `backend/migrations/0056_vehicle_operational_fields.sql`
- Modify: `backend/internal/application/vehicle_types.go`
- Modify: `backend/internal/application/vehicle_validation.go`
- Modify: `backend/internal/application/vehicle_repository.go`
- Modify: `backend/internal/application/vehicle_mutation.go`
- Modify: `backend/internal/application/vehicles.go`
- Test: `backend/internal/application/vehicles_test.go`
- Test: `backend/internal/infrastructure/migrate_test.go`

- [ ] **Step 1: Add failing migration and application tests**

Add tests that create and update a vehicle with both new values, read it back, and find it through
the server-side search. Add invalid cases for speeds `0` and `1001`, plus a `homeBase` longer than
200 Unicode code points. Assert that omitted fields remain valid.

Use these API field shapes in all test fixtures so omitted input and an explicitly invalid zero stay
distinguishable:

```go
MaximumSpeedKmh *int   `json:"maximumSpeedKmh,omitempty"`
HomeBase        string `json:"homeBase,omitempty"`
```

The migration test must verify that an existing vehicle receives safe defaults:

```sql
maximum_speed_kmh INTEGER
home_base TEXT NOT NULL DEFAULT ''
```

Run:

```powershell
cd backend
go test ./internal/application ./internal/infrastructure -run "VehicleOperational|VehicleSearch|Migrate"
```

Expected: FAIL because the fields and columns do not exist.

- [ ] **Step 2: Add migration 0056**

Create the forward-only migration:

```sql
ALTER TABLE vehicles ADD COLUMN maximum_speed_kmh INTEGER
  CHECK (maximum_speed_kmh IS NULL OR maximum_speed_kmh BETWEEN 1 AND 1000);
ALTER TABLE vehicles ADD COLUMN home_base TEXT NOT NULL DEFAULT '';
```

Do not assign a speed to existing vehicles.

- [ ] **Step 3: Extend the vehicle types and validation**

Insert `MaximumSpeedKmh` and `HomeBase` directly after `VehicleNumber` in `Vehicle` and
`CreateVehicleInput`. Use `*int` for the optional speed in both types. Trim `HomeBase`, reject
values over 200 runes, and reject every supplied speed outside `1..1000`. Convert a nil pointer to
SQL `NULL` in persistence.

Trim the home base with `input.HomeBase = strings.TrimSpace(input.HomeBase)` inside
`cleanVehicleInput` and add this validator:

```go
func isValidVehicleOperationalInput(input CreateVehicleInput) bool {
    if utf8.RuneCountInString(input.HomeBase) > 200 {
        return false
    }
    return input.MaximumSpeedKmh == nil ||
        (*input.MaximumSpeedKmh >= 1 && *input.MaximumSpeedKmh <= 1000)
}
```

Call it from both `Create` and `Update` alongside the existing required-field checks and return
`ErrVehicleValidation` when it is false.

- [ ] **Step 4: Extend every explicit vehicle SQL statement and scanner**

Add `maximum_speed_kmh` and `home_base` at the same stable position in list/get SELECTs, INSERT,
UPDATE, argument lists, and scan targets. Pass the optional pointer through a small nullable-value
helper and scan with `sql.NullInt64`, converting a valid value to a fresh `*int` and SQL `NULL` to
nil. Do not collapse a stored value into zero.

Extend the search predicate with:

```sql
OR LOWER(COALESCE(home_base, '')) LIKE LOWER(?)
OR CAST(maximum_speed_kmh AS TEXT) LIKE ?
```

Append matching wildcard arguments in the exact query order. Do not change pagination, ordering,
or other search behavior.

- [ ] **Step 5: Format and rerun focused backend tests**

```powershell
gofmt -w internal/application/vehicle_types.go internal/application/vehicle_validation.go `
  internal/application/vehicle_repository.go internal/application/vehicle_mutation.go `
  internal/application/vehicles.go internal/application/vehicles_test.go `
  internal/infrastructure/migrate_test.go
go test ./internal/application ./internal/infrastructure -run "VehicleOperational|VehicleSearch|Migrate"
```

Expected: PASS.

- [ ] **Step 6: Commit Task 1**

```powershell
git add backend/migrations/0056_vehicle_operational_fields.sql `
  backend/internal/application/vehicle_types.go `
  backend/internal/application/vehicle_validation.go `
  backend/internal/application/vehicle_repository.go `
  backend/internal/application/vehicle_mutation.go `
  backend/internal/application/vehicles.go `
  backend/internal/application/vehicles_test.go `
  backend/internal/infrastructure/migrate_test.go
git commit -m "feat: add vehicle operational data"
```

---

### Task 2: Expose vehicle operational data throughout the frontend and CSV/report flows

**Files:**

- Modify: `frontend/src/shared/api.ts`
- Modify: `frontend/src/features/vehicles/vehicleTransforms.ts`
- Modify: `frontend/src/features/vehicles/vehicleViewModel.ts`
- Modify: `frontend/src/features/vehicles/VehicleModelTab.tsx`
- Modify: `frontend/src/features/vehicles/VehicleReadOnlyView.tsx`
- Modify: `frontend/src/features/vehicles/vehicleReports.ts`
- Modify: `frontend/src/features/vehicles/vehicleTableColumns.ts`
- Modify: `frontend/src/features/importExport/importExportHelpers.tsx`
- Modify: `frontend/src/test/fixtures/vehicles.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Test: `frontend/src/features/vehicles/vehicleTableColumns.test.ts`
- Test: `frontend/src/features/vehicles/VehicleModelTab.test.tsx`
- Test: `frontend/src/features/importExport/importExportHelpers.test.tsx`
- Test: `frontend/src/features/vehicles/vehicleReports.test.ts`

- [ ] **Step 1: Add failing frontend model and flow tests**

Test these behaviors:

- `Vehicle` and `CreateVehicleRequest` accept `maximumSpeedKmh?: number` and `homeBase?: string`.
- The model tab renders the two fields immediately after series and vehicle number.
- Speed input uses `type="number"`, `min={1}`, `max={1000}`, `step={1}` and a visible `km/h` unit.
- Home base uses `maxLength={200}`.
- The read-only/report group is named `vehicle.section.prototypeOperation` and renders both values.
- CSV export contains `maximumSpeedKmh` and `homeBase`; CSV import converts the speed to an integer
  or leaves it absent.
- Both table columns exist but are not in `defaultVehicleTableColumns`.
- Parsing an old stored column preference does not add either new column.

Run:

```powershell
cd frontend
npm.cmd run test:run -- VehicleModelTab vehicleTableColumns importExportHelpers vehicleReports
```

Expected: FAIL because the new fields are absent.

- [ ] **Step 2: Extend shared vehicle types and form/view models**

Add after `vehicleNumber`:

```ts
maximumSpeedKmh?: number;
homeBase?: string;
```

Initialize an empty form with `maximumSpeedKmh: undefined` and `homeBase: ""`. Preserve values in
all vehicle-to-form and form-to-request transforms. When converting the number input, map an empty
value to `undefined`, otherwise use `Number.parseInt(value, 10)`.

- [ ] **Step 3: Implement the approved 2×2 model layout**

Keep the existing responsive form grid. Directly below the row containing series and vehicle
number, add:

```tsx
<label>
  {t("vehicle.field.maximumSpeedKmh")}
  <span className="input-with-unit">
    <input
      type="number"
      min={1}
      max={1000}
      step={1}
      value={vehicle.maximumSpeedKmh ?? ""}
      onChange={(event) => onChange("maximumSpeedKmh", event.target.value === ""
        ? undefined
        : Number.parseInt(event.target.value, 10))}
    />
    <span>km/h</span>
  </span>
</label>
<label>
  {t("vehicle.field.homeBase")}
  <input
    maxLength={200}
    value={vehicle.homeBase ?? ""}
    onChange={(event) => onChange("homeBase", event.target.value)}
  />
</label>
```

Adapt the callback syntax to the existing `VehicleModelTab` props rather than introducing a second
form state mechanism.

- [ ] **Step 4: Add read-only and report group `Vorbild & Betrieb`**

Place the group near the model identity data and render:

```ts
{ label: t("vehicle.field.maximumSpeedKmh"), value: vehicle.maximumSpeedKmh ? `${vehicle.maximumSpeedKmh} km/h` : placeholder },
{ label: t("vehicle.field.homeBase"), value: vehicle.homeBase || placeholder }
```

Use the same ordering in HTML read-only view and generated report. Do not mix `homeBase` with the
physical storage location.

- [ ] **Step 5: Add CSV import/export mappings**

Add stable field IDs `maximumSpeedKmh` and `homeBase` to the vehicle export header and import field
registry. Validate imported speeds as whole numbers from 1 to 1000 before constructing the request.
Keep missing columns backward compatible.

- [ ] **Step 6: Add default-hidden table columns**

Add both keys to `vehicleTableColumnKeys`, map them to the `identity` group, and keep
`defaultVehicleTableColumns` unchanged. For speed display, append `km/h`; for sorting, use the
numeric string padded or compare numerically through the table's existing comparator extension so
`120` sorts after `90`.

- [ ] **Step 7: Add translations**

Use exactly:

```ts
"vehicle.field.maximumSpeedKmh": "Höchstgeschwindigkeit",
"vehicle.field.homeBase": "Heimat-Bw / Einsatzstelle",
"vehicle.section.prototypeOperation": "Vorbild & Betrieb"
```

and:

```ts
"vehicle.field.maximumSpeedKmh": "Maximum speed",
"vehicle.field.homeBase": "Home depot / operating location",
"vehicle.section.prototypeOperation": "Prototype & operation"
```

- [ ] **Step 8: Rerun focused frontend tests and build**

```powershell
npm.cmd run test:run -- VehicleModelTab vehicleTableColumns importExportHelpers vehicleReports
npm.cmd run build
```

Expected: PASS, including strict TypeScript checks.

- [ ] **Step 9: Commit Task 2**

```powershell
git add frontend/src/shared/api.ts frontend/src/features/vehicles `
  frontend/src/features/importExport/importExportHelpers.tsx `
  frontend/src/features/importExport/importExportHelpers.test.tsx `
  frontend/src/test/fixtures/vehicles.ts frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts
git commit -m "feat: expose vehicle operational data"
```

---

### Task 3: Add accessory list price to persistence, API, and article listings

**Files:**

- Create: `backend/migrations/0057_accessory_list_price.sql`
- Modify: `backend/internal/application/accessories.go`
- Modify: `backend/internal/application/accessory_overview.go`
- Modify: `backend/internal/infrastructure/accessory_repository.go`
- Modify: `backend/internal/infrastructure/accessory_article_query.go`
- Test: `backend/internal/application/accessories_test.go`
- Test: `backend/internal/infrastructure/accessory_repository_test.go`
- Test: `backend/internal/application/accessory_overview_test.go`

- [ ] **Step 1: Add failing accessory list-price tests**

Test create, get, update, and article-list round trips using `listPrice: "129.90"`. Add rejected
inputs `-1`, `1.234`, and malformed text. Verify an omitted price stays empty and existing records
migrate safely.

Run:

```powershell
cd backend
go test ./internal/application ./internal/infrastructure -run "Accessory.*ListPrice|AccessoryArticle"
```

Expected: FAIL because `listPrice` is not part of accessory articles.

- [ ] **Step 2: Add migration 0057**

```sql
ALTER TABLE accessory_products ADD COLUMN list_price TEXT NOT NULL DEFAULT '';
```

Do not derive a list price from purchases.

- [ ] **Step 3: Extend accessory domain/API types and validation**

Add `ListPrice string \`json:"listPrice,omitempty"\`` to `AccessoryProduct` and
`AccessoryArticleListItem`, and `ListPrice string \`json:"listPrice"\`` to
`CreateAccessoryProductInput`. Trim the value. Add a narrow exact-string validation helper in
`accessories.go` for this task; Task 5 replaces that helper with the shared money parser after its
table-driven parser tests pass.

Validation rules:

- blank is valid;
- nonnegative only;
- at most two decimal places;
- accepted separators match RailKeeper's price parser;
- persist the trimmed original compatible string, not a float.

- [ ] **Step 4: Extend repository CRUD and article list query**

Add `list_price` to every accessory product SELECT, INSERT, UPDATE, and scan target. Add
`product.list_price` to `article_rows` and scan it into `AccessoryArticleListItem.ListPrice`.
Keep this new column nonsortable because the compatible text storage contains multiple decimal
notations and lexical ordering would be incorrect. Do not add `listPrice` to
`AccessoryArticleSort` or the backend sort map.

- [ ] **Step 5: Format and rerun focused tests**

```powershell
gofmt -w internal/application/accessories.go internal/application/accessory_overview.go `
  internal/infrastructure/accessory_repository.go `
  internal/infrastructure/accessory_article_query.go `
  internal/application/accessories_test.go `
  internal/application/accessory_overview_test.go `
  internal/infrastructure/accessory_repository_test.go
go test ./internal/application ./internal/infrastructure -run "Accessory.*ListPrice|AccessoryArticle"
```

Expected: PASS.

- [ ] **Step 6: Commit Task 3**

```powershell
git add backend/migrations/0057_accessory_list_price.sql `
  backend/internal/application/accessories.go `
  backend/internal/application/accessory_overview.go `
  backend/internal/infrastructure/accessory_repository.go `
  backend/internal/infrastructure/accessory_article_query.go `
  backend/internal/application/accessories_test.go `
  backend/internal/application/accessory_overview_test.go `
  backend/internal/infrastructure/accessory_repository_test.go
git commit -m "feat: add accessory list price"
```

---

### Task 4: Add accessory list price to the editor and default-hidden table column

**Files:**

- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/accessories/articleEditorModel.ts`
- Modify: `frontend/src/features/accessories/ArticleCoreTab.tsx`
- Modify: `frontend/src/features/accessories/articleTableColumns.ts`
- Modify: `frontend/src/features/accessories/ArticleColumnPicker.tsx`
- Modify: `frontend/src/features/accessories/ArticleTable.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Test: `frontend/src/features/accessories/articleEditorModel.test.ts`
- Test: `frontend/src/features/accessories/articleTableColumns.test.ts`
- Test: `frontend/src/features/accessories/ArticleColumnPicker.test.tsx`
- Test: `frontend/src/features/accessories/ArticleTable.test.tsx`

- [ ] **Step 1: Add failing editor and column-preference tests**

Test form/API round trips, the `0.00` value, comma-decimal input, rejection of negative and
unambiguously malformed decimal inputs such as `1.2345`, and display with two locale-aware
decimals. For table preferences assert:

```ts
expect(articleTableColumns).toContain("listPrice");
expect(defaultArticleTableColumns.has("listPrice")).toBe(false);
expect(storedArticleTableColumns(emptyStorage).has("listPrice")).toBe(false);
expect(storedArticleTableColumns(oldPreferenceStorage).has("listPrice")).toBe(false);
```

Run:

```powershell
cd frontend
npm.cmd run test:run -- articleEditorModel articleTableColumns ArticleColumnPicker ArticleTable
```

Expected: FAIL.

- [ ] **Step 2: Extend accessory API and editor types**

Add `listPrice: string` to the article read/write interfaces and editor form model. Preserve the
text until submit, normalize accepted comma input to the API's stored decimal form only through an
exact string conversion, and never call `parseFloat` for money.

- [ ] **Step 3: Add the editor field**

Place the field with the article's commercial core data, labelled "Listenpreis pro Stück" /
"List price per unit". Use the existing app-owned number input where it supports decimal text;
otherwise use an input with `inputMode="decimal"`, `min="0"`, and a validation message. Do not use
native numeric conversion that turns an empty value into zero.

- [ ] **Step 4: Separate all columns from default columns**

Change the definitions to this shape:

```ts
export const articleTableColumns = [
  "image", "inventoryNumber", "manufacturer", "articleNumber", "name",
  "type", "gauge", "listPrice", "stock", "storage"
] as const;

const defaultArticleTableColumnOrder: ArticleTableColumn[] = [
  "image", "inventoryNumber", "manufacturer", "articleNumber", "name",
  "type", "gauge", "stock", "storage"
];

export const defaultArticleTableColumns = new Set(defaultArticleTableColumnOrder);
```

Make `resetArticleTableColumns` and all no-storage/invalid-storage fallbacks use
`defaultArticleTableColumnOrder`. Normalization must not append newly introduced columns. Preserve
the existing safety rule that at least inventory number or name remains visible.

- [ ] **Step 5: Render the table column and picker entry**

Render blank prices as the shared placeholder and nonblank prices through an exact formatter with
two decimals and EUR. Add the picker label without changing existing column order or mobile card
layout. Refactor `ArticleTable`'s header registry from a list of always-sortable pairs to ordered
definitions shaped as `{ key, sort?: AccessoryArticleSort }`; render `listPrice` with a plain
`<th>` and retain the existing sort button for definitions that carry `sort`.

- [ ] **Step 6: Add translations and rerun tests**

Use:

```ts
"accessories.field.listPrice": "Listenpreis pro Stück"
"accessories.field.listPrice": "List price per unit"
```

Then run:

```powershell
npm.cmd run test:run -- articleEditorModel articleTableColumns ArticleColumnPicker ArticleTable
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 7: Commit Task 4**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts `
  frontend/src/features/accessories/articleEditorModel.ts `
  frontend/src/features/accessories/articleEditorModel.test.ts `
  frontend/src/features/accessories/ArticleCoreTab.tsx `
  frontend/src/features/accessories/articleTableColumns.ts `
  frontend/src/features/accessories/articleTableColumns.test.ts `
  frontend/src/features/accessories/ArticleColumnPicker.tsx `
  frontend/src/features/accessories/ArticleColumnPicker.test.tsx `
  frontend/src/features/accessories/ArticleTable.tsx `
  frontend/src/features/accessories/ArticleTable.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: show accessory list price"
```

---

### Task 5: Implement exact money parsing and the valuation application service

**Files:**

- Create: `backend/internal/application/money.go`
- Create: `backend/internal/application/money_test.go`
- Create: `backend/internal/application/overview_valuation.go`
- Create: `backend/internal/application/overview_valuation_test.go`

- [ ] **Step 1: Add failing table-driven money tests**

Cover these exact cases:

```go
tests := []struct {
    input string
    cents int64
    ok    bool
}{
    {"", 0, false},
    {"0", 0, true},
    {"129.90", 12990, true},
    {"129,90", 12990, true},
    {"1.299,90", 129990, true},
    {"1,299.90", 129990, true},
    {"1.299", 129900, true},
    {"-1.00", 0, false},
    {"12.345", 1234500, true},
    {"12.3456", 0, false},
    {"abc", 0, false},
}
```

Also test `formatMoneyCents(0) == "0.00"` and
`formatMoneyCents(129990) == "1299.90"`.

Run:

```powershell
cd backend
go test ./internal/application -run Money
```

Expected: FAIL because the parser does not exist.

- [ ] **Step 2: Implement the integer-cent parser**

Implement `parseMoneyCents(value string) (int64, bool)` by trimming, rejecting signs other than an
optional leading plus, identifying the decimal separator from the last `.` or `,`, validating
thousands grouping, and building the integer from digits. Use `strconv.ParseInt` after removing
validated separators. Reject fractions longer than two digits and guard multiplication overflow.

Implement:

```go
func formatMoneyCents(cents int64) string {
    return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}
```

Do not use `float64`, `strconv.ParseFloat`, or SQL casts for money.

- [ ] **Step 3: Add failing valuation service tests**

Create fixture data that proves all rules in one test:

- vehicle list prices sum independently from vehicle purchase prices;
- invalid and blank vehicle prices add zero;
- accessory list price is multiplied by current quantity;
- quantity stock and active installed quantity are counted exactly once;
- individual assets with `lifecycle_state='retired'` do not increase current list value;
- an archived accessory product with current stock remains valued;
- purchase rows add `quantity * unit_price`;
- blank and `EUR` currency rows are included;
- explicit `USD` and `CHF` rows are excluded and counted;
- an individual asset without `purchase_id` adds its manual purchase price;
- an asset with `purchase_id` does not double count the purchase;
- manual asset purchase cost remains historical even after retirement;
- result strings always have two decimals.

Run:

```powershell
go test ./internal/application -run OverviewValuation
```

Expected: FAIL because the service does not exist.

- [ ] **Step 4: Implement the typed service and bulk reads**

Use this public response:

```go
type OverviewValuation struct {
    VehicleListValue                  string `json:"vehicleListValue"`
    VehiclePurchaseValue              string `json:"vehiclePurchaseValue"`
    AccessoryListValue                string `json:"accessoryListValue"`
    AccessoryPurchaseCost             string `json:"accessoryPurchaseCost"`
    ExcludedForeignCurrencyPurchases  int    `json:"excludedForeignCurrencyPurchases"`
}

type OverviewValuationService struct {
    db *sql.DB
}

func NewOverviewValuationService(db *sql.DB) *OverviewValuationService
func (s *OverviewValuationService) Get(ctx context.Context) (OverviewValuation, error)
```

Use four bulk queries, not per-product queries:

1. `SELECT list_price, purchase_price FROM vehicles`
2. A grouped accessory query returning `list_price` and current owned count. For current count,
   mirror the existing quantity/installation semantics and exclude individual assets whose
   `lifecycle_state='retired'`.
3. `SELECT quantity, unit_price, currency FROM accessory_purchases`
4. `SELECT purchase_price FROM accessory_assets WHERE purchase_id IS NULL`

For every query, close rows and check `rows.Err()`. Wrap errors contextually, for example
`fmt.Errorf("calculate accessory list value: %w", err)`. Use checked integer addition and
multiplication so corrupted or extreme values return an error rather than overflow.

- [ ] **Step 5: Reuse the money parser in accessory validation**

Replace any temporary parser from Task 3. Accessory list-price validation must call the same exact
parser and additionally reject formats whose normalized fraction exceeds two digits. Vehicle and
maintenance legacy values remain read-compatible; do not rewrite them.

- [ ] **Step 6: Format and rerun tests**

```powershell
gofmt -w internal/application/money.go internal/application/money_test.go `
  internal/application/overview_valuation.go `
  internal/application/overview_valuation_test.go internal/application/accessories.go
go test ./internal/application -run "Money|OverviewValuation|Accessory.*ListPrice"
```

Expected: PASS.

- [ ] **Step 7: Commit Task 5**

```powershell
git add backend/internal/application/money.go `
  backend/internal/application/money_test.go `
  backend/internal/application/overview_valuation.go `
  backend/internal/application/overview_valuation_test.go `
  backend/internal/application/accessories.go
git commit -m "feat: calculate exact inventory valuations"
```

---

### Task 6: Expose the valuation endpoint and align the OpenAPI contract

**Files:**

- Create: `backend/internal/api/overview_handlers.go`
- Create: `backend/internal/api/overview_handlers_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/routes.go`
- Modify: `backend/cmd/railkeeper/main.go`
- Modify: `backend/internal/api/router_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

- [ ] **Step 1: Add failing endpoint and authorization tests**

Build a router with `OverviewValuationService` and fixture data. Assert:

- unauthenticated request returns `401`;
- Viewer, Editor, and Admin receive `200`;
- Messe receives `403`;
- response contains all five properties and money strings with two decimals;
- the route table and OpenAPI contain `GET /api/v1/overview/valuation`.

Run:

```powershell
cd backend
go test ./internal/api -run "OverviewValuation|OpenAPI|Route"
```

Expected: FAIL because the route is absent.

- [ ] **Step 2: Wire the service through router configuration**

Add:

```go
OverviewValuationService *application.OverviewValuationService
```

to `api.Config`, add `overviewValuationService` to `App`, and copy it in `NewRouter`. In
`backend/cmd/railkeeper/main.go`, construct it with the shared database:

```go
OverviewValuationService: application.NewOverviewValuationService(db),
```

- [ ] **Step 3: Register the viewer route and thin handler**

Add to `apiRouteSpecs` near the vehicle/accessory read routes:

```go
{http.MethodGet, "/api/v1/overview/valuation", routeAccessViewer, (*App).overviewValuation, nil},
```

The handler returns `503` if the service is unavailable, maps service failures through the existing
JSON error helper, and otherwise responds `200` with the typed result. Do not expose query details
or database errors to the client.

- [ ] **Step 4: Update OpenAPI**

Add the path and an `OverviewValuation` schema with all five required properties. Define monetary
values as strings matching `^[0-9]+\\.[0-9]{2}$` and the excluded count as a nonnegative integer.

In the same contract change, add:

- `maximumSpeedKmh`: integer, minimum 1, maximum 1000, nullable/optional as appropriate;
- `homeBase`: string, maxLength 200;
- accessory `listPrice`: string using the project's existing price conventions;
- the fields to both read and write schemas, not only response schemas.

- [ ] **Step 5: Format and run endpoint/contract tests**

```powershell
gofmt -w internal/api/overview_handlers.go internal/api/overview_handlers_test.go `
  internal/api/router.go internal/api/routes.go internal/api/router_test.go `
  internal/api/openapi_contract_test.go cmd/railkeeper/main.go
go test ./internal/api ./cmd/railkeeper -run "OverviewValuation|OpenAPI|Route"
```

Expected: PASS.

- [ ] **Step 6: Commit Task 6**

```powershell
git add backend/internal/api/overview_handlers.go `
  backend/internal/api/overview_handlers_test.go backend/internal/api/router.go `
  backend/internal/api/routes.go backend/internal/api/router_test.go `
  backend/internal/api/openapi_contract_test.go backend/cmd/railkeeper/main.go `
  openapi/railkeeper.yaml
git commit -m "feat: expose overview valuation API"
```

---

### Task 7: Render dashboard valuation variant B with independent failure handling

**Files:**

- Modify: `frontend/src/shared/api.ts`
- Create: `frontend/src/features/overview/OverviewValuationCard.tsx`
- Create: `frontend/src/features/overview/OverviewValuationCard.test.tsx`
- Modify: `frontend/src/features/overview/OverviewView.tsx`
- Modify: `frontend/src/features/overview/OverviewView.test.tsx`
- Modify: `frontend/src/styles/overview.css`
- Modify: `frontend/src/styles/overrides-responsive.css`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

- [ ] **Step 1: Add failing API/component/overview tests**

Define the expected frontend shape:

```ts
export type OverviewValuation = {
  vehicleListValue: string;
  vehiclePurchaseValue: string;
  accessoryListValue: string;
  accessoryPurchaseCost: string;
  excludedForeignCurrencyPurchases: number;
};
```

Test:

- `1299.90` renders `1.299,90 €` in German and `€1,299.90` in English;
- all four labels and values appear in a 2×2 semantic grid;
- loading does not display invented zero values;
- a valuation request failure renders a localized card error while vehicle KPI data remains;
- the foreign-currency hint appears only when the excluded count is greater than zero;
- the old single `overview.listValue` KPI is gone;
- responsive CSS contains two-column and one-column breakpoints without horizontal overflow.

Run:

```powershell
cd frontend
npm.cmd run test:run -- OverviewValuationCard OverviewView
```

Expected: FAIL.

- [ ] **Step 2: Add API method and exact browser formatter**

Add:

```ts
overviewValuation: () => request<OverviewValuation>("/api/v1/overview/valuation")
```

Format the server decimal string without `Number` or `parseFloat`: validate it against
`/^\d+\.\d{2}$/`, split whole and cents, and add the language-specific grouping separator to the
whole string with `/\B(?=(\d{3})+(?!\d))/g`. Render German as
`${groupedWhole},${cents} €` and English as `€${groupedWhole}.${cents}`. Always preserve exactly two
cents. Export the formatter from the focused component for direct tests.

- [ ] **Step 3: Implement `OverviewValuationCard`**

The component accepts:

```ts
type OverviewValuationCardProps = {
  valuation: OverviewValuation | null;
  loading: boolean;
  error: string;
};
```

Render a heading `overview.valuation.title`, a short basis text, and four labelled value cells:

```ts
[
  ["overview.valuation.vehicleList", valuation.vehicleListValue],
  ["overview.valuation.vehiclePurchase", valuation.vehiclePurchaseValue],
  ["overview.valuation.accessoryList", valuation.accessoryListValue],
  ["overview.valuation.accessoryPurchase", valuation.accessoryPurchaseCost]
]
```

Use `role="status"` for loading and `role="alert"` for the local failure. Keep the card usable in
light/dark mode using existing tokens only.

- [ ] **Step 4: Load valuation independently in `OverviewView`**

Add separate `valuation`, `valuationLoading`, and `valuationError` state. The refresh action starts
both requests, but each request owns its own `.catch` and `.finally`. A valuation failure must not
set the existing vehicle `message` and must not empty `vehicles`.

Remove `stats.totalValue` and the old rounded list-value hero cell. Insert the wider valuation card
in that position while keeping total inventory, digitalization, and maintenance compact. Leave the
existing maintenance-cost formatter behavior unchanged unless a focused test shows it also needs
two decimals.

- [ ] **Step 5: Implement the approved responsive layout**

Use a hero grid where the valuation card spans the width of two ordinary KPI cells on desktop. Its
inner matrix uses two equal columns. At the current tablet breakpoint it remains two columns if
space allows; at the narrow mobile breakpoint it becomes one column. Add `min-width: 0` to grid
children and wrap long German labels. Do not add horizontal document scrolling.

- [ ] **Step 6: Add German and English text**

Use these meanings:

```ts
"overview.valuation.title": "Erfasste Bestandswerte",
"overview.valuation.vehicleList": "Fahrzeuge · Listenwert",
"overview.valuation.vehiclePurchase": "Fahrzeuge · Kaufpreis",
"overview.valuation.accessoryList": "Zubehör · Listenwert",
"overview.valuation.accessoryPurchase": "Zubehör · Kaufkosten",
"overview.valuation.foreignExcluded": "{{count}} Einkauf/Einkäufe in Fremdwährung nicht eingerechnet.",
"overview.valuation.error": "Bestandswerte konnten nicht geladen werden."
```

Provide natural English equivalents and plural handling consistent with the current i18n helper.

- [ ] **Step 7: Rerun focused tests and build**

```powershell
npm.cmd run test:run -- OverviewValuationCard OverviewView
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 8: Commit Task 7**

```powershell
git add frontend/src/shared/api.ts frontend/src/features/overview `
  frontend/src/styles/overview.css frontend/src/styles/overrides-responsive.css `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: show detailed inventory valuations"
```

---

### Task 8: Prove backup compatibility and complete regression verification

**Files:**

- Modify: `backend/internal/application/backup_test.go`
- Modify: `backend/internal/api/smoke_test.go`
- Modify: `docs/superpowers/specs/2026-08-16-vehicle-operational-data-and-valuations-design.md` only if implementation reveals an approved factual correction

- [ ] **Step 1: Add failing backup round-trip coverage**

Create a vehicle with both operational fields and an accessory product with a list price. Export a
backup, restore it into a clean migrated database, and assert all three values survive exactly.
Also restore an older fixture without the new keys/columns and assert defaults are empty without
errors.

Run:

```powershell
cd backend
go test ./internal/application -run "Backup.*Operational|Backup.*ListPrice|BackupRestore"
```

Expected before coverage/compatibility fixes: FAIL.

- [ ] **Step 2: Make only required backup compatibility changes**

The current generic table backup should carry the new columns automatically. If tests pass without
production changes, retain only the regression tests. If older restore validation rejects absent
fields, make the smallest compatibility adjustment and do not include authentication tables or
change the backup security boundary.

- [ ] **Step 3: Extend the smoke test**

Add authenticated `GET /api/v1/overview/valuation` coverage for Admin/Viewer and forbidden coverage
for Messe. Include the new vehicle fields in the create/read smoke flow without weakening CSRF or
role checks.

- [ ] **Step 4: Run all backend checks**

```powershell
gofmt -w internal/application/backup_test.go internal/api/smoke_test.go
go test ./...
```

Expected: PASS.

- [ ] **Step 5: Run all frontend checks**

```powershell
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
```

Expected: all Vitest suites pass and production build succeeds.

- [ ] **Step 6: Review the final diff for scope and generated files**

```powershell
cd ..
git status --short
git diff --check
git diff --stat origin/main...HEAD
git diff origin/main...HEAD -- backend/migrations openapi/railkeeper.yaml `
  backend/internal/application backend/internal/api backend/internal/infrastructure `
  frontend/src/features frontend/src/shared frontend/src/styles
```

Confirm manually:

- no #82 updater/download code changed;
- no `frontend/dist`, `frontend/node_modules`, `data`, `.cache`, secrets, or backups are staged;
- database columns and API/OpenAPI names match exactly;
- no existing stored column preference automatically reveals the new columns;
- valuation money never passes through floating point.

- [ ] **Step 7: Perform local visual acceptance before publishing**

Start the local server with the repo-local cache and inspect:

- vehicle editor and read-only view in German/English;
- vehicle CSV round trip;
- accessory editor and optional table column;
- overview in German/English, light/dark, desktop/mobile;
- valuation loading, foreign-currency hint, and forced API failure;
- long home-base text and empty states.

Do not push, open a PR, merge, close issues, or publish a release until the user has approved this
local result.

- [ ] **Step 8: Commit verification tests and any narrow compatibility fix**

```powershell
git add backend/internal/application/backup_test.go backend/internal/api/smoke_test.go
git commit -m "test: cover valuation data compatibility"
```

If no files changed in Task 8 because all coverage already exists, skip the empty commit.

---

## Completion Criteria

- #78 operational fields persist, validate, search, import/export, report, and render correctly.
- #86 returns four cent-exact EUR valuation strings through one viewer-only endpoint.
- Vehicle and accessory list/purchase values remain separated.
- Accessory purchase costs avoid linked-asset double counting.
- Explicit foreign currencies are excluded transparently.
- The dashboard implements approved variant B without blocking on valuation failures.
- Existing users and old backups remain compatible, with no database overwrite or data-loss path.
- #82 remains untouched.
- `go test ./...`, `npm.cmd run test:run`, `npm.cmd run build`, and `git diff --check` pass.
