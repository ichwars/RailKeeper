# Vehicle Inventory Columns and Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement GitHub issues #80 and #81 with server-persisted, reorderable vehicle inventory columns, expanded search, three new filters, responsive rendering, and an accessory-column reset.

**Architecture:** Reuse the existing `/api/v1/profile/settings` key/value API for an ordered JSON column list. Keep column definitions, normalization, rendering, persistence, and UI controls in focused frontend modules; extend the existing vehicle list query for the additional search and display fields. Apply the new domain filters in the existing inventory controller so they combine with the server-side search result and all current filters.

**Tech Stack:** Go 1.x, SQLite, React 19, TypeScript 7, Vite 8, Vitest 4, Testing Library, existing RailKeeper CSS tokens and i18n.

## Global Constraints

- The current table columns and order remain the default.
- Newly introduced columns remain hidden for existing saved preferences.
- Inventory number may be hidden, but normalization must restore it when no data column remains.
- Column widths are not configurable.
- Preferences are server-side and isolated per local user account; no installation-wide defaults or cross-installation sync.
- Mobile uses the same selected fields and order without horizontal document overflow.
- The existing desktop card view remains unchanged.
- German and English, light and dark themes, and Viewer access must work.
- Reuse `/api/v1/profile/settings`; do not add a migration or a dedicated preference endpoint.
- Do not change accessory column ordering, storage, or width behavior.
- Do not publish, push, close issues, or release before local user review.

---

## File Map

- `backend/internal/application/vehicles.go`: extended list projection and search predicates.
- `backend/internal/application/vehicles_test.go`: search and list-projection regression tests.
- `backend/internal/application/settings_test.go`: existing profile-setting persistence and user-isolation coverage.
- `openapi/railkeeper.yaml`: accurate vehicle fields and search description.
- `frontend/src/features/vehicles/vehicleTableColumns.ts`: column registry, defaults, normalization, movement, formatting, and sorting values.
- `frontend/src/features/vehicles/vehicleTableColumns.test.ts`: pure column-model tests.
- `frontend/src/features/vehicles/useVehicleColumnPreferences.ts`: profile-setting load/save queue and column commands.
- `frontend/src/features/vehicles/useVehicleColumnPreferences.test.tsx`: server persistence tests.
- `frontend/src/features/vehicles/VehicleColumnPicker.tsx`: grouped selection, visible order, movement, and reset controls.
- `frontend/src/features/vehicles/VehicleColumnPicker.test.tsx`: keyboard and command tests.
- `frontend/src/features/vehicles/VehicleInventoryTable.tsx`: dynamic desktop table.
- `frontend/src/features/vehicles/VehicleInventoryTable.test.tsx`: dynamic desktop column tests.
- `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx`: dynamic compact mobile fields.
- `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx`: ordered mobile field tests.
- `frontend/src/features/vehicles/VehicleInventoryPanel.tsx`: integrates picker, table, mobile list, and new filters.
- `frontend/src/features/vehicles/useVehicleInventoryController.ts`: new filters and sort fallback.
- `frontend/src/features/vehicles/useVehicleInventoryController.test.tsx`: combined-filter and sort-fallback tests.
- `frontend/src/features/vehicles/VehiclesView.tsx`: preference/controller wiring.
- `frontend/src/features/vehicles/VehiclesView.test.tsx`: integration coverage for profile loading and visible columns.
- `frontend/src/features/accessories/articleTableColumns.ts`: accessory reset helper.
- `frontend/src/features/accessories/ArticleColumnPicker.tsx`: accessory reset action.
- `frontend/src/features/accessories/AccessoriesView.tsx`: reset wiring and persistence.
- `frontend/src/features/accessories/articleTableColumns.test.ts`: reset-model test.
- `frontend/src/features/accessories/ArticleColumnPicker.test.tsx`: reset UI test.
- `frontend/src/shared/i18n/de.ts`, `frontend/src/shared/i18n/en.ts`: labels and messages.
- `frontend/src/styles/vehicle-inventory.css`: picker and dynamic mobile/table presentation.
- `frontend/src/app/styles.css`: imports the focused stylesheet.
- `frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts`: bounded responsive CSS checks.

---

### Task 1: Extend server-side vehicle search and list fields

**Files:**
- Modify: `backend/internal/application/vehicles.go`
- Modify: `backend/internal/application/vehicles_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: `VehicleService.List(ctx context.Context, query string) ([]Vehicle, error)`.
- Produces: the same method signature, now searching `series`, `vehicle_number`, and `decoder_type` and populating every short field approved by the design.

- [ ] **Step 1: Write failing search and projection tests**

Add table-driven cases to `backend/internal/application/vehicles_test.go`:

```go
func TestListVehiclesSearchesExtendedFields(t *testing.T) {
	db := testDB(t)
	service := application.NewVehicleService(db)
	ctx := context.Background()

	created, err := service.Create(ctx, application.CreateVehicleInput{
		Manufacturer: "Roco", Name: "Diesellok", Gauge: "H0",
		Category: "Lokomotive", Gattung: "Diesellok",
		Series: "BR 218", VehicleNumber: "218 217-8", DecoderType: "LokSound 5",
		RailwayCompany: "DB", Epoch: "IV", Adapter: "PluX22",
		AcquisitionType: "Kauf", AcquiredFrom: "Fachhandel", PurchasePrice: "199,90",
		PurchaseDate: "2026-08-15", StorageLocation: "Vitrine 1", Condition: "Sehr gut",
		Packaging: "OVP", LengthMM: "189", WeightG: "420", Color: "ozeanblau/beige",
		Lettering: "DB", Load: "", Interior: "Führerstand", Axles: "Bo'Bo'",
		AxleCount: "4", TractionTireCount: "2", Wheelset: "AC",
		CouplingFront: "Kurzkupplung", CouplingRear: "Kurzkupplung",
		PowerPickup: "Schleifer", Digital: true, DigitalDecoderNumber: "21",
		DTDecoder: true, DTDecoderNumber: "88", ABCBrakes: true,
		DriveEnabled: true, HeadlightsEnabled: true, LightingEnabled: true,
		SoundGeneratorEnabled: true, SmokeGeneratorEnabled: false, ExhibitionReady: true,
	}, "actor-1")
	if err != nil { t.Fatal(err) }

	for _, query := range []string{"br 218", "217-8", "loksound"} {
		items, listErr := service.List(ctx, "  "+query+"  ")
		if listErr != nil { t.Fatal(listErr) }
		if len(items) != 1 || items[0].ID != created.ID {
			t.Fatalf("query %q returned %#v", query, items)
		}
	}

	items, err := service.List(ctx, "218")
	if err != nil { t.Fatal(err) }
	got := items[0]
	if got.Adapter != "PluX22" || got.PurchasePrice != "199,90" ||
		got.StorageLocation != "Vitrine 1" || got.AxleCount != "4" ||
		!got.DriveEnabled || !got.ExhibitionReady {
		t.Fatalf("short list fields missing: %#v", got)
	}
}
```

- [ ] **Step 2: Run the backend test and verify RED**

Run:

```powershell
cd backend
go test ./internal/application -run TestListVehiclesSearchesExtendedFields -count=1
```

Expected: FAIL because the three new fields are absent from the `WHERE` clause and approved short fields are not scanned by `List`.

- [ ] **Step 3: Expand the SQL and scan targets**

In `VehicleService.List`, append the approved columns after `list_price` and before timestamps:

```sql
COALESCE(acquisition_type, ''), COALESCE(acquired_from, ''),
COALESCE(purchase_price, ''), COALESCE(purchase_date, ''),
COALESCE(storage_location, ''), COALESCE(condition, ''), COALESCE(packaging, ''),
COALESCE(length_mm, ''), COALESCE(weight_g, ''), COALESCE(color, ''),
COALESCE(lettering, ''), COALESCE(load, ''), COALESCE(interior, ''),
COALESCE(axles, ''), COALESCE(axle_count, ''), COALESCE(traction_tire_count, ''),
COALESCE(wheelset, ''), COALESCE(coupling_front, ''), COALESCE(coupling_rear, ''),
COALESCE(power_pickup, ''), COALESCE(adapter, ''), drive_enabled,
headlights_enabled, lighting_enabled, sound_generator_enabled,
smoke_generator_enabled
```

Add matching `rows.Scan` targets and convert integer flags to booleans. Extend the predicates and arguments exactly as follows:

```sql
OR series LIKE ? COLLATE NOCASE
OR vehicle_number LIKE ? COLLATE NOCASE
OR decoder_type LIKE ? COLLATE NOCASE
```

Use the same trimmed `%query%` value for all eight search predicates. Preserve `ORDER BY updated_at DESC, inventory_number ASC`.

- [ ] **Step 4: Align the OpenAPI contract**

Add the currently missing `Vehicle` and `CreateVehicleRequest` properties used by the list:

```yaml
        decoderType:
          type: string
        exhibition:
          type: boolean
        acquisitionType:
          type: string
        acquiredFrom:
          type: string
        purchasePrice:
          type: string
        purchaseDate:
          type: string
          format: date
        storageLocation:
          type: string
        condition:
          type: string
        packaging:
          type: string
```

Give the `q` parameter this description:

```yaml
          description: Case-insensitive substring search across inventory number, manufacturer, article number, designation, series, vehicle number, and decoder type.
```

- [ ] **Step 5: Verify GREEN and commit**

Run:

```powershell
cd backend
gofmt -w internal/application/vehicles.go internal/application/vehicles_test.go
go test ./internal/application -run 'TestListVehicles(FiltersByQuery|SearchesExtendedFields)' -count=1
cd ..
git add backend/internal/application/vehicles.go backend/internal/application/vehicles_test.go openapi/railkeeper.yaml
git commit -m "feat: expand vehicle inventory search"
```

Expected: PASS; commit contains only backend search/list and contract changes.

---

### Task 2: Define and normalize the vehicle column model

**Files:**
- Create: `frontend/src/features/vehicles/vehicleTableColumns.ts`
- Create: `frontend/src/features/vehicles/vehicleTableColumns.test.ts`
- Modify: `frontend/src/features/vehicles/vehicleViewModel.ts`

**Interfaces:**
- Produces: `VehicleTableColumn`, `VehicleColumnGroup`, `vehicleTableColumns`, `defaultVehicleTableColumns`, `normalizeVehicleTableColumns`, `parseVehicleTableColumns`, `serializeVehicleTableColumns`, `toggleVehicleTableColumn`, `moveVehicleTableColumn`, `sortableVehicleColumn`, and `vehicleColumnSortValue`.
- Consumes later: preference hook, inventory controller, picker, table, and mobile list.

- [ ] **Step 1: Write failing pure-model tests**

Create `vehicleTableColumns.test.ts` with these behaviors:

```ts
import { describe, expect, it } from "vitest";
import {
  defaultVehicleTableColumns,
  moveVehicleTableColumn,
  normalizeVehicleTableColumns,
  parseVehicleTableColumns,
  serializeVehicleTableColumns,
  toggleVehicleTableColumn
} from "./vehicleTableColumns";

describe("vehicle table columns", () => {
  it("keeps the current desktop columns as defaults", () => {
    expect(defaultVehicleTableColumns).toEqual([
      "image", "inventoryNumber", "manufacturer", "articleNumber",
      "name", "gauge", "epoch", "exhibition"
    ]);
  });

  it("preserves saved order, removes unknown and duplicate keys, and does not append new keys", () => {
    expect(normalizeVehicleTableColumns([
      "series", "manufacturer", "futureColumn", "series"
    ])).toEqual(["series", "manufacturer"]);
  });

  it("restores inventory number when only presentation columns remain", () => {
    expect(normalizeVehicleTableColumns(["image", "exhibition"]))
      .toEqual(["image", "exhibition", "inventoryNumber"]);
  });

  it("uses defaults for missing or malformed settings", () => {
    expect(parseVehicleTableColumns(undefined)).toEqual(defaultVehicleTableColumns);
    expect(parseVehicleTableColumns("not-json")).toEqual(defaultVehicleTableColumns);
  });

  it("toggles, moves, and serializes a normalized ordered list", () => {
    const shown = toggleVehicleTableColumn(defaultVehicleTableColumns, "series");
    const moved = moveVehicleTableColumn(shown, "series", "up");
    expect(moved.at(-2)).toBe("series");
    expect(parseVehicleTableColumns(serializeVehicleTableColumns(moved))).toEqual(moved);
  });
});
```

- [ ] **Step 2: Run the test and verify RED**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleTableColumns.test.ts
```

Expected: FAIL because `vehicleTableColumns.ts` does not exist.

- [ ] **Step 3: Implement the typed registry and helpers**

Create a registry containing exactly the approved keys:

```ts
export const vehicleTableColumnKeys = [
  "image", "inventoryNumber", "manufacturer", "articleNumber", "name", "gauge",
  "epoch", "exhibition", "railwayCompany", "category", "gattung", "series",
  "vehicleNumber", "ean", "productionPeriod", "digital", "digitalDecoderNumber",
  "dtDecoder", "dtDecoderNumber", "decoderType", "adapter", "abcBrakes",
  "listPrice", "acquisitionType", "acquiredFrom", "purchasePrice", "purchaseDate",
  "storageLocation", "condition", "packaging", "lengthMm", "weightG", "color",
  "lettering", "load", "interior", "axles", "axleCount", "tractionTireCount",
  "wheelset", "couplingFront", "couplingRear", "powerPickup", "driveEnabled",
  "headlightsEnabled", "lightingEnabled", "soundGeneratorEnabled",
  "smokeGeneratorEnabled", "exhibitionReady"
] as const;

export type VehicleTableColumn = typeof vehicleTableColumnKeys[number];
export type VehicleColumnMove = "up" | "down";
export const defaultVehicleTableColumns: VehicleTableColumn[] = [
  "image", "inventoryNumber", "manufacturer", "articleNumber",
  "name", "gauge", "epoch", "exhibition"
];
```

Define group metadata for `identity`, `digital`, `ownership`, `technical`, and `equipment`. Mark `image` and `exhibition` as presentation columns, boolean keys as `boolean`, `purchaseDate` as `date`, and all other fields as `text`.

Implement normalization without appending absent registry keys:

```ts
export function normalizeVehicleTableColumns(values: Iterable<unknown>) {
  const seen = new Set<VehicleTableColumn>();
  const normalized = [...values].flatMap((value) => {
    if (!isVehicleTableColumn(value) || seen.has(value)) return [];
    seen.add(value);
    return [value];
  });
  if (!normalized.some(isVehicleDataColumn)) normalized.push("inventoryNumber");
  return normalized;
}
```

`vehicleColumnSortValue(vehicle, key)` must return `"1"`/`"0"` for boolean values and a trimmed lowercase string for all others. Replace the closed `SortKey` union in `vehicleViewModel.ts` with the exported sortable column type.

- [ ] **Step 4: Verify GREEN and commit**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleTableColumns.test.ts
cd ..
git add frontend/src/features/vehicles/vehicleTableColumns.ts frontend/src/features/vehicles/vehicleTableColumns.test.ts frontend/src/features/vehicles/vehicleViewModel.ts
git commit -m "feat: define vehicle inventory columns"
```

Expected: PASS with stable ordered normalization.

---

### Task 3: Persist column preferences per user through profile settings

**Files:**
- Create: `backend/internal/application/settings_test.go`
- Create: `frontend/src/features/vehicles/useVehicleColumnPreferences.ts`
- Create: `frontend/src/features/vehicles/useVehicleColumnPreferences.test.tsx`

**Interfaces:**
- Consumes: `api.profileSettings()`, `api.updateProfileSettings(settings)`, and Task 2 column helpers.
- Produces: `useVehicleColumnPreferences(onMessage)` returning `{ columns, loading, toggleColumn, moveColumn, resetColumns }`.

- [ ] **Step 1: Add backend user-isolation characterization coverage**

Create `settings_test.go`:

```go
package application_test

import (
	"context"
	"testing"

	"railkeeper/backend/internal/application"
)

func TestUserSettingsRemainIsolatedAndMergePartialUpdates(t *testing.T) {
	db := testDB(t)
	service := application.NewSettingsService(db)
	ctx := context.Background()

	for _, id := range []string{"user-a", "user-b"} {
		if _, err := db.Exec(`INSERT INTO users(id, username, password_hash, active, created_at, updated_at)
VALUES(?, ?, 'hash', 1, '2026-08-15T00:00:00Z', '2026-08-15T00:00:00Z')`, id, id); err != nil {
			t.Fatal(err)
		}
	}

	_, err := service.UpdateUserSettings(ctx, "user-a", application.SettingsPayload{Settings: map[string]string{
		"railkeeper.vehicles.tableColumns": `["series","inventoryNumber"]`,
		"theme": "dark",
	}})
	if err != nil { t.Fatal(err) }
	_, err = service.UpdateUserSettings(ctx, "user-a", application.SettingsPayload{Settings: map[string]string{
		"railkeeper.vehicles.tableColumns": `["inventoryNumber"]`,
	}})
	if err != nil { t.Fatal(err) }

	a, err := service.UserSettings(ctx, "user-a")
	if err != nil { t.Fatal(err) }
	b, err := service.UserSettings(ctx, "user-b")
	if err != nil { t.Fatal(err) }
	if a.Settings["theme"] != "dark" || a.Settings["railkeeper.vehicles.tableColumns"] != `["inventoryNumber"]` {
		t.Fatalf("unexpected merged settings: %#v", a.Settings)
	}
	if len(b.Settings) != 0 { t.Fatalf("settings leaked to user-b: %#v", b.Settings) }
}
```

Run `go test ./internal/application -run TestUserSettingsRemainIsolatedAndMergePartialUpdates -count=1`; expected PASS because this verifies the existing server facility used by the feature.

- [ ] **Step 2: Write failing hook tests**

Mock the shared API and assert:

```ts
it("loads the ordered server preference without local storage", async () => {
  vi.spyOn(api, "profileSettings").mockResolvedValue({
    settings: { [vehicleTableColumnSettingKey]: '["series","inventoryNumber"]' }
  });
  const { result } = renderHook(() => useVehicleColumnPreferences(vi.fn()));
  await waitFor(() => expect(result.current.columns).toEqual(["series", "inventoryNumber"]));
  expect(window.localStorage.getItem(vehicleTableColumnSettingKey)).toBeNull();
});

it("saves normalized changes as a partial profile update", async () => {
  vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
  const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
  const { result } = renderHook(() => useVehicleColumnPreferences(vi.fn()));
  await waitFor(() => expect(result.current.loading).toBe(false));
  act(() => result.current.toggleColumn("series"));
  await waitFor(() => expect(update).toHaveBeenCalledWith({
    [vehicleTableColumnSettingKey]: expect.stringContaining("series")
  }));
});
```

Run `npm.cmd run test:run -- src/features/vehicles/useVehicleColumnPreferences.test.tsx`; expected FAIL because the hook does not exist.

- [ ] **Step 3: Implement the hook with serialized saves**

Use this stable key and public shape:

```ts
export const vehicleTableColumnSettingKey = "railkeeper.vehicles.tableColumns";

export function useVehicleColumnPreferences(onMessage: (message: string) => void) {
  const [columns, setColumns] = useState(defaultVehicleTableColumns);
  const [loading, setLoading] = useState(true);
  const saveQueue = useRef(Promise.resolve());
  // load once with api.profileSettings(); normalize or keep defaults on error
  // each command calculates next, updates state, and appends api.updateProfileSettings
  return { columns, loading, toggleColumn, moveColumn, resetColumns };
}
```

Implement `queueSave(next)` by chaining from `saveQueue.current.catch(() => undefined)` so rapid moves are sent in order. On load failure call `onMessage(t("vehicles.columns.loadError"))`; on save failure call `onMessage(t("vehicles.columns.saveError"))`. Do not write this preference to `localStorage`.

- [ ] **Step 4: Verify GREEN and commit**

Run both focused suites, then commit:

```powershell
cd backend
go test ./internal/application -run TestUserSettingsRemainIsolatedAndMergePartialUpdates -count=1
cd ..\frontend
npm.cmd run test:run -- src/features/vehicles/useVehicleColumnPreferences.test.tsx src/features/vehicles/vehicleTableColumns.test.ts
cd ..
git add backend/internal/application/settings_test.go frontend/src/features/vehicles/useVehicleColumnPreferences.ts frontend/src/features/vehicles/useVehicleColumnPreferences.test.tsx
git commit -m "feat: persist vehicle columns per user"
```

Expected: both suites PASS and unrelated profile keys remain untouched.

---

### Task 4: Add railway company, epoch, and adapter filters plus sort fallback

**Files:**
- Modify: `frontend/src/features/vehicles/useVehicleInventoryController.ts`
- Modify: `frontend/src/features/vehicles/useVehicleInventoryController.test.tsx`
- Modify: `frontend/src/features/vehicles/vehicleViewModel.ts`

**Interfaces:**
- Consumes: ordered `VehicleTableColumn[]` from Task 3 and `vehicleColumnSortValue` from Task 2.
- Produces: `railwayCompanyFilter`, `epochFilter`, `adapterFilter`, their setters/options, and valid sorting constrained to visible sortable columns.

- [ ] **Step 1: Write failing controller tests**

Extend the hook call to accept visible columns and add:

```ts
it("combines railway company, epoch, and adapter filters", () => {
  const db = vehicleFixture({ id: "db", railwayCompany: "DB", epoch: "IV", adapter: "PluX22" });
  const dr = analogVehicleFixture({ id: "dr", railwayCompany: "DR", epoch: "III", adapter: "NEM 652" });
  const { result } = renderHook(() => useVehicleInventoryController([db, dr], defaultVehicleTableColumns));

  act(() => result.current.setRailwayCompanyFilter("DB"));
  act(() => result.current.setEpochFilter("IV"));
  act(() => result.current.setAdapterFilter("PluX22"));
  expect(result.current.filteredVehicles.map((vehicle) => vehicle.id)).toEqual(["db"]);
  act(() => result.current.resetInventoryFilters());
  expect(result.current.hasActiveInventoryFilters).toBe(false);
});

it("falls back when the active sort column is hidden", () => {
  const { result, rerender } = renderHook(
    ({ columns }) => useVehicleInventoryController([vehicleFixture(), analogVehicleFixture()], columns),
    { initialProps: { columns: ["inventoryNumber", "series"] as VehicleTableColumn[] } }
  );
  act(() => result.current.toggleSort("series"));
  rerender({ columns: ["inventoryNumber"] });
  expect(result.current.sort).toEqual({ key: "inventoryNumber", direction: "asc" });
});
```

Run the focused controller test; expected FAIL because the new state and hook parameter are missing.

- [ ] **Step 2: Implement option extraction, filtering, and reset**

Extend `inventoryFilterOptions` with sorted unique values:

```ts
railwayCompanies: unique(vehicles.map((vehicle) => vehicle.railwayCompany || "")),
epochs: unique(vehicles.map((vehicle) => vehicle.epoch || "")),
adapters: unique(vehicles.map((vehicle) => vehicle.adapter || ""))
```

Add strict-value predicates:

```ts
if (railwayCompanyFilter && vehicle.railwayCompany !== railwayCompanyFilter) return false;
if (epochFilter && vehicle.epoch !== epochFilter) return false;
if (adapterFilter && vehicle.adapter !== adapterFilter) return false;
```

Include all three in URL-preset clearing, active-filter detection, return values, and reset.

- [ ] **Step 3: Implement visible-column sort fallback**

Use `vehicleColumnSortValue` in the comparator. Add an effect that checks `sortableVehicleColumn(sort.key)` and membership in the ordered columns. When hidden, choose `inventoryNumber` if present and sortable; otherwise select the first visible sortable column. Always set fallback direction to `asc`.

- [ ] **Step 4: Verify GREEN and commit**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/useVehicleInventoryController.test.tsx
cd ..
git add frontend/src/features/vehicles/useVehicleInventoryController.ts frontend/src/features/vehicles/useVehicleInventoryController.test.tsx frontend/src/features/vehicles/vehicleViewModel.ts
git commit -m "feat: add vehicle inventory domain filters"
```

Expected: PASS for combined filters, reset, existing presets, sorting, and selection.

---

### Task 5: Build the accessible column picker and dynamic inventory views

**Files:**
- Create: `frontend/src/features/vehicles/VehicleColumnPicker.tsx`
- Create: `frontend/src/features/vehicles/VehicleColumnPicker.test.tsx`
- Create: `frontend/src/features/vehicles/VehicleInventoryTable.tsx`
- Create: `frontend/src/features/vehicles/VehicleInventoryTable.test.tsx`
- Create: `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx`
- Create: `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryPanel.tsx`
- Create: `frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts`
- Create: `frontend/src/styles/vehicle-inventory.css`
- Modify: `frontend/src/app/styles.css`

**Interfaces:**
- Consumes: Task 2 registry and Task 3 commands.
- Produces: a toolbar picker and two renderers using the exact same ordered column list.

- [ ] **Step 1: Write failing picker tests**

Test open/close, grouped checkboxes, movement, and reset:

```tsx
render(
  <VehicleColumnPicker
    columns={["inventoryNumber", "series"]}
    onToggle={onToggle}
    onMove={onMove}
    onReset={onReset}
  />
);
await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
await user.click(screen.getByRole("checkbox", { name: "Baureihe" }));
expect(onToggle).toHaveBeenCalledWith("series");
await user.click(screen.getByRole("button", { name: "Baureihe nach oben" }));
expect(onMove).toHaveBeenCalledWith("series", "up");
await user.click(screen.getByRole("button", { name: "Auf Standard zurücksetzen" }));
expect(onReset).toHaveBeenCalledOnce();
```

Run the picker test; expected FAIL because the component does not exist.

- [ ] **Step 2: Implement the picker**

Use `Columns3`, `ChevronUp`, `ChevronDown`, and `RotateCcw`. Follow the existing accessory picker behavior for outside click, Escape, focus restoration, `aria-haspopup="dialog"`, and `aria-expanded`. Render:

```tsx
<section aria-label={t("vehicles.columns.visibleOrder")}>
  {columns.map((column, index) => (
    <div key={column} className="vehicle-column-order-row">
      <span>{columnLabel(column, t)}</span>
      <button disabled={index === 0} aria-label={t("vehicles.columns.moveUp", { label })} />
      <button disabled={index === columns.length - 1} aria-label={t("vehicles.columns.moveDown", { label })} />
    </div>
  ))}
</section>
```

Below it render grouped checkboxes from the registry and a text reset button. Do not implement drag-and-drop or width controls.

- [ ] **Step 3: Write failing dynamic rendering tests**

Test the new table and mobile components directly with a fixture and columns `series`, `digital`, and `purchaseDate`. Assert desktop header/cell order, localized boolean text, absence of hidden defaults, mobile field order, and no duplicate row actions.

Run:

```powershell
npm.cmd run test:run -- src/features/vehicles/VehicleInventoryTable.test.tsx src/features/vehicles/VehicleInventoryMobileList.test.tsx
```

Expected: FAIL because the components do not exist. Create these two test files beside their components as part of this step.

- [ ] **Step 4: Implement shared ordered rendering**

`VehicleInventoryTable` receives all existing selection/action callbacks plus `columns`, `sort`, and `onToggleSort`. Map `columns` once for headers and once per row. Use special cells for `image`, `name`, and `exhibition`; render all boolean fields as `t("common.yes")` or `t("common.no")`; render missing values as `t("common.placeholder")`.

`VehicleInventoryMobileList` receives the same `columns`. Render selected image only when `image` is present. Render the remaining selected values as an ordered `<dl>` and keep edit/delete/quick-menu actions outside the data sequence. The article `aria-label` uses inventory number and name even if both fields are hidden visually.

Replace the fixed table and fixed mobile markup in `VehicleInventoryPanel.tsx` with these components. Leave the existing desktop card view unchanged. Place `VehicleColumnPicker` in `inventory-toolbar-actions` in all view modes so mobile users can configure the shared preference.

- [ ] **Step 5: Add bounded responsive styling and tests**

Create `vehicle-inventory.css` using existing tokens. Required rules:

```css
.vehicle-column-picker { position: relative; }
.vehicle-column-picker-popover { position: absolute; inset-inline-end: 0; z-index: 30; width: min(680px, calc(100vw - 32px)); max-height: min(72vh, 720px); overflow: auto; }
.vehicle-column-groups { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.vehicle-mobile-fields { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); min-width: 0; }
@media (max-width: 720px) {
  .vehicle-column-picker-popover { position: fixed; inset: auto 12px 12px; width: auto; }
  .vehicle-column-groups, .vehicle-mobile-fields { grid-template-columns: 1fr; }
}
```

Import it from `frontend/src/app/styles.css` before `overrides-responsive.css`. The CSS test reads the stylesheet and asserts the fixed mobile inset, one-column mobile grid, bounded width, and absence of document-width declarations.

- [ ] **Step 6: Verify GREEN and commit**

Run all new component/model tests, then commit the focused UI:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleColumnPicker.test.tsx src/features/vehicles/VehicleInventoryTable.test.tsx src/features/vehicles/VehicleInventoryMobileList.test.tsx src/features/vehicles/vehicleInventoryResponsive.test.ts
cd ..
git add frontend/src/features/vehicles/VehicleColumnPicker.tsx frontend/src/features/vehicles/VehicleColumnPicker.test.tsx frontend/src/features/vehicles/VehicleInventoryTable.tsx frontend/src/features/vehicles/VehicleInventoryTable.test.tsx frontend/src/features/vehicles/VehicleInventoryMobileList.tsx frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx frontend/src/features/vehicles/VehicleInventoryPanel.tsx frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts frontend/src/styles/vehicle-inventory.css frontend/src/app/styles.css
git commit -m "feat: add configurable vehicle inventory columns"
```

Expected: PASS and the fixed card-view markup remains unchanged.

---

### Task 6: Wire preferences, filters, and localized copy into VehiclesView

**Files:**
- Modify: `frontend/src/features/vehicles/VehiclesView.tsx`
- Modify: `frontend/src/features/vehicles/VehiclesView.test.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryPanel.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: Tasks 2 through 5.
- Produces: complete end-to-end frontend behavior for #80 and #81.

- [ ] **Step 1: Write failing integration tests**

In `VehiclesView.test.tsx`, mock `api.profileSettings` with a non-default ordered list and assert the rendered headers appear in that order. Change a column and assert `api.updateProfileSettings` receives only the vehicle column key. Select railway company, epoch, and adapter values and assert the displayed result count narrows while the search request still uses `api.vehicles(query)`.

Also assert the German search placeholder is:

```text
Inventarnummer, Hersteller, Artikel, Bezeichnung, Baureihe, Fahrzeug-Nr. oder Decoder-Typ
```

Run the focused integration test; expected FAIL because `VehiclesView` has not wired the new modules.

- [ ] **Step 2: Wire hooks and props**

In `VehiclesView.tsx`, instantiate preferences before the inventory controller:

```ts
const columnPreferences = useVehicleColumnPreferences(setMessage);
const inventory = useVehicleInventoryController(vehicles, columnPreferences.columns);
```

Pass `columns`, `columnsLoading`, `onToggleColumn`, `onMoveColumn`, and `onResetColumns` to `VehicleInventoryPanel`. Pass the three new filter values, options, and setters. Keep `api.vehicles(query)` as the only server list request.

- [ ] **Step 3: Add exact German and English copy**

Add keys for `vehicles.columns.*`, group labels, load/save errors, movement labels, reset, and new filters. Update the search placeholder in both languages. Use existing `vehicle.field.*` labels for column names; add missing field labels only when no key exists.

German filter labels: `Bahngesellschaft`, `Epoche`, `Adapter/Schnittstelle`.

English filter labels: `Railway company`, `Era`, `Adapter/interface`.

- [ ] **Step 4: Verify GREEN and commit**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehiclesView.test.tsx src/features/vehicles/useVehicleColumnPreferences.test.tsx src/features/vehicles/useVehicleInventoryController.test.tsx
cd ..
git add frontend/src/features/vehicles/VehiclesView.tsx frontend/src/features/vehicles/VehiclesView.test.tsx frontend/src/features/vehicles/VehicleInventoryPanel.tsx frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: integrate vehicle columns and filters"
```

Expected: PASS with server preference, search, and filters working together.

---

### Task 7: Add accessory reset-to-defaults without changing its storage model

**Files:**
- Modify: `frontend/src/features/accessories/articleTableColumns.ts`
- Modify: `frontend/src/features/accessories/articleTableColumns.test.ts`
- Modify: `frontend/src/features/accessories/ArticleColumnPicker.tsx`
- Modify: `frontend/src/features/accessories/ArticleColumnPicker.test.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/article-overview.css`

**Interfaces:**
- Produces: `resetArticleTableColumns()` and an `onReset` picker action.
- Preserves: current browser-local key, visible-column order, identity guard, and table widths.

- [ ] **Step 1: Write failing model and UI tests**

Add:

```ts
expect(resetArticleTableColumns()).toEqual(defaultArticleTableColumns);
```

Render `ArticleColumnPicker` with `onReset`, open it, click `Auf Standard zurücksetzen`, and expect one call. In `AccessoriesView.test.tsx`, hide a column, reset, and assert the local-storage value equals all `articleTableColumns` in their existing stable order.

Run the three accessory tests; expected FAIL because reset support is missing.

- [ ] **Step 2: Implement reset and persistence**

Add:

```ts
export function resetArticleTableColumns() {
  return new Set<ArticleTableColumn>(articleTableColumns);
}
```

Add required `onReset: () => void` to the picker and render a text button after the checkboxes. In `AccessoriesView`, set defaults and call `persistArticleTableColumns(next)` in the reset handler. Add localized copy and a subtle divider/button rule in `article-overview.css` using existing tokens.

- [ ] **Step 3: Verify GREEN and commit**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/articleTableColumns.test.ts src/features/accessories/ArticleColumnPicker.test.tsx src/features/accessories/AccessoriesView.test.tsx
cd ..
git add frontend/src/features/accessories/articleTableColumns.ts frontend/src/features/accessories/articleTableColumns.test.ts frontend/src/features/accessories/ArticleColumnPicker.tsx frontend/src/features/accessories/ArticleColumnPicker.test.tsx frontend/src/features/accessories/AccessoriesView.tsx frontend/src/features/accessories/AccessoriesView.test.tsx frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/article-overview.css
git commit -m "feat: reset accessory columns to defaults"
```

Expected: PASS with unchanged browser-local persistence semantics.

---

### Task 8: Full verification and local visual handoff

**Files:**
- Modify only if verification exposes a defect: files already listed in Tasks 1 through 7.

**Interfaces:**
- Produces: a locally verified feature branch ready for user inspection, not publication.

- [ ] **Step 1: Run complete backend verification**

```powershell
cd backend
gofmt -w internal/application/vehicles.go internal/application/vehicles_test.go internal/application/settings_test.go
go test ./...
```

Expected: all Go packages PASS with no formatting diff.

- [ ] **Step 2: Run complete frontend verification**

```powershell
cd ..\frontend
npm.cmd run test:run
npm.cmd run build
```

Expected: all Vitest suites PASS; TypeScript and Vite build successfully without errors.

- [ ] **Step 3: Check repository hygiene**

```powershell
cd ..
git diff --check
git status --short
git log --oneline --decorate -8
```

Expected: no whitespace errors; only known pre-existing untracked files remain; implementation files are committed on `dev/issues-80-81-inventory`.

- [ ] **Step 4: Start the local app and inspect all required states**

Build frontend assets, start the Go server with repository-local `GOCACHE`, then inspect:

1. German, dark, desktop table: default columns unchanged.
2. Column picker: grouped long list, scroll containment, toggling, movement, reset.
3. Saved preference after reload and sign-out/sign-in.
4. A second user account: independent default preference.
5. Search by series, vehicle number, and decoder type.
6. Combined railway company, epoch, adapter, and existing filters.
7. Hidden active sort column falls back safely.
8. Mobile width: ordered fields, actions, no document overflow.
9. English and light theme.
10. Accessory picker reset.

Capture screenshots for user review only if they help compare dense desktop and mobile states. Do not add screenshots or `frontend/dist` to Git.

- [ ] **Step 5: Stop before publication**

Report the local URL, test/build results, branch, commits, and any visual caveats. Wait for explicit user approval before push, issue closure, version bump, release, or publication.
