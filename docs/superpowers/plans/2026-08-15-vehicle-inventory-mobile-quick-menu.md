# Compact Mobile Vehicle Cards Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the noisy mobile vehicle list with compact expandable cards, keep every vehicle quick menu visible, and add the missing PluX12 option from Issue #85.

**Architecture:** `VehicleInventoryMobileList` owns only the set of expanded vehicle IDs. A new presentational `VehicleInventoryMobileCard` renders the fixed compact hierarchy, remaining selected columns, and actions from the existing vehicle column definitions. Focused vehicle inventory CSS fixes menu stacking, while the existing vehicle adapter option module becomes the single source for vehicle and exhibition dropdowns.

**Tech Stack:** React 19, TypeScript 7, Testing Library, Vitest, CSS, existing RailKeeper i18n and vehicle column helpers.

## Global Constraints

- Keep search, filters, persisted column selection, column order, data model, API, and permissions unchanged.
- Respect the normalized visible columns on desktop and mobile; do not introduce a second mobile preference.
- Show the image in the collapsed mobile header only when the image column is visible.
- Keep every card collapsed initially; expanded state is local to the current list view and is not persisted.
- Do not repeat compact-header fields in the expanded details.
- Keep long values inside the card without horizontal document overflow.
- Keep View, Edit, and Quick menu directly available; Delete remains available through the existing quick menu.
- Add `PluX12` to vehicle and exhibition dropdowns from one shared option list; do not change accessories or article-source parsing.
- Use existing design tokens and German/English i18n.
- Do not push, publish, close issues, or change backend code as part of this plan.

---

## File Structure

- Create `frontend/src/features/vehicles/VehicleInventoryMobileCard.tsx`: one accessible expandable vehicle card with compact summary, ordered detail fields, and actions.
- Modify `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx`: own expanded vehicle IDs, remove stale IDs when the result list changes, and render cards.
- Modify `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx`: exercise collapsed summary, independent expansion, column order, action isolation, hidden image, and stale-state cleanup.
- Modify `frontend/src/shared/i18n/de.ts`: German expand/collapse labels.
- Modify `frontend/src/shared/i18n/en.ts`: English expand/collapse labels.
- Modify `frontend/src/styles/vehicle-inventory.css`: compact card layout, detail layout, quick-menu stacking, and table action-cell overflow exception.
- Modify `frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts`: lock the responsive card and overflow rules with focused CSS regression tests.
- Modify `frontend/src/features/vehicles/vehicleOptions.ts`: add `PluX12` in numeric PluX order and serve as the shared adapter option source.
- Create `frontend/src/features/vehicles/vehicleOptions.test.ts`: lock the shared adapter option order.
- Modify `frontend/src/features/exhibition/ExhibitionView.tsx`: import the shared adapter options instead of keeping a duplicate list.
- Modify `frontend/src/features/exhibition/ExhibitionView.test.tsx`: verify that the exhibition entry dropdown exposes `PluX12`.

---

### Task 1: Build the expandable mobile vehicle card

**Files:**
- Create: `frontend/src/features/vehicles/VehicleInventoryMobileCard.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx`
- Modify: `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: `Vehicle`, `VehicleTableColumn`, `vehicleColumnText`, `vehicleColumnLabel`, `primaryImage`, `previewImageUrl`, and the existing action callbacks.
- Produces: `VehicleInventoryMobileCard(props)` with `expanded: boolean` and `onToggleExpanded: () => void`; `VehicleInventoryMobileList` remains API-compatible with its current parent except that its unused direct `onDelete` prop is removed.

- [ ] **Step 1: Replace the mobile-list test with failing behavior tests**

Replace `frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx` with:

```tsx
import { fireEvent, render, screen, within } from "@testing-library/react";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
import { VehicleInventoryMobileList } from "./VehicleInventoryMobileList";

function renderList(overrides: Partial<ComponentProps<typeof VehicleInventoryMobileList>> = {}) {
  const first = vehicleFixture({
    id: "vehicle-1",
    inventoryNumber: "RK-LOK-000001",
    name: "BR 106",
    manufacturer: "ESU",
    articleNumber: "12345",
    gauge: "H0",
    epoch: "III",
    series: "BR 106",
    purchaseDate: "2026-08-15"
  });
  const second = vehicleFixture({
    id: "vehicle-2",
    inventoryNumber: "RK-LOK-000002",
    name: "BR 218",
    manufacturer: "Piko",
    articleNumber: "57903",
    gauge: "H0",
    epoch: "IV",
    series: "BR 218",
    images: []
  });
  const props: ComponentProps<typeof VehicleInventoryMobileList> = {
    vehicles: [first, second],
    columns: [
      "image",
      "inventoryNumber",
      "name",
      "manufacturer",
      "articleNumber",
      "gauge",
      "epoch",
      "series",
      "purchaseDate"
    ],
    onOpenDetail: vi.fn(),
    onOpenEdit: vi.fn(),
    renderQuickMenu: (vehicle) => <button type="button">Menü {vehicle.inventoryNumber}</button>,
    ...overrides
  };
  return { ...render(<VehicleInventoryMobileList {...props} />), first, second, props };
}

describe("VehicleInventoryMobileList", () => {
  it("shows compact summaries and independently expands ordered remaining fields", () => {
    renderList();

    const firstToggle = screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    });
    const secondToggle = screen.getByRole("button", {
      name: "Details für RK-LOK-000002 anzeigen"
    });
    expect(firstToggle).toHaveAttribute("aria-expanded", "false");
    expect(secondToggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.queryByRole("term")).not.toBeInTheDocument();

    fireEvent.click(firstToggle);

    expect(firstToggle).toHaveAttribute("aria-expanded", "true");
    expect(secondToggle).toHaveAttribute("aria-expanded", "false");
    expect(screen.getAllByRole("term").map((term) => term.textContent)).toEqual([
      "Baureihe",
      "Datum"
    ]);
    expect(screen.getAllByRole("definition").map((value) => value.textContent)).toEqual([
      "BR 106",
      "15.08.2026"
    ]);
  });

  it("hides an unselected image and keeps View, Edit, and Quick menu outside the toggle", () => {
    const onOpenDetail = vi.fn();
    const onOpenEdit = vi.fn();
    renderList({
      vehicles: [vehicleFixture({ images: [] })],
      columns: ["inventoryNumber", "series"],
      onOpenDetail,
      onOpenEdit
    });

    expect(screen.queryByText("Keine Vorschau")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Anzeigen" }));
    fireEvent.click(screen.getByRole("button", { name: "Bearbeiten" }));

    expect(onOpenDetail).toHaveBeenCalledTimes(1);
    expect(onOpenEdit).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    })).toHaveAttribute("aria-expanded", "false");
    expect(screen.getByRole("button", { name: "Menü RK-LOK-000001" })).toBeVisible();
  });

  it("forgets expanded IDs when a filtered vehicle leaves the result list", () => {
    const { first, props, rerender } = renderList({ vehicles: [vehicleFixture()] });
    fireEvent.click(screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    }));

    rerender(<VehicleInventoryMobileList {...props} vehicles={[]} />);
    rerender(<VehicleInventoryMobileList {...props} vehicles={[first]} />);

    expect(screen.getByRole("button", {
      name: "Details für RK-LOK-000001 anzeigen"
    })).toHaveAttribute("aria-expanded", "false");
  });

  it("uses the neutral image treatment when the visible image has no file", () => {
    renderList({
      vehicles: [vehicleFixture({ images: [] })],
      columns: ["image", "inventoryNumber"]
    });

    const card = screen.getByRole("article");
    expect(within(card).getByText("Keine Vorschau")).toBeVisible();
  });
});
```

- [ ] **Step 2: Run the focused test and verify the new contract fails**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleInventoryMobileList.test.tsx
```

Expected: FAIL because the current list opens the detail page instead of expanding, still exposes
Delete directly, and has no localized expand/collapse controls.

- [ ] **Step 3: Add exact German and English labels**

Add beside `vehicles.mobileList` in `frontend/src/shared/i18n/de.ts`:

```ts
    "vehicles.mobile.expand": "Details für {inventoryNumber} anzeigen",
    "vehicles.mobile.collapse": "Details für {inventoryNumber} ausblenden",
```

Add beside `vehicles.mobileList` in `frontend/src/shared/i18n/en.ts`:

```ts
    "vehicles.mobile.expand": "Show details for {inventoryNumber}",
    "vehicles.mobile.collapse": "Hide details for {inventoryNumber}",
```

- [ ] **Step 4: Create the presentational card**

Create `frontend/src/features/vehicles/VehicleInventoryMobileCard.tsx` with:

```tsx
import type { ReactNode } from "react";
import { ChevronDown, ChevronUp, Eye, Pencil } from "lucide-react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { previewImageUrl, primaryImage } from "./vehicleTransforms";
import {
  vehicleColumnLabel,
  vehicleColumnText,
  type VehicleTableColumn
} from "./vehicleTableColumns";

const mobileSummaryColumns = new Set<VehicleTableColumn>([
  "inventoryNumber",
  "name",
  "manufacturer",
  "articleNumber",
  "gauge",
  "epoch"
]);

type VehicleInventoryMobileCardProps = {
  vehicle: Vehicle;
  columns: readonly VehicleTableColumn[];
  expanded: boolean;
  onToggleExpanded: () => void;
  onOpenDetail: (vehicle: Vehicle) => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryMobileCard({
  vehicle,
  columns,
  expanded,
  onToggleExpanded,
  onOpenDetail,
  onOpenEdit,
  renderQuickMenu
}: VehicleInventoryMobileCardProps) {
  const { language, t } = useI18n();
  const image = primaryImage(vehicle.images);
  const showsImage = columns.includes("image");
  const shows = (column: VehicleTableColumn) => columns.includes(column);
  const text = (column: VehicleTableColumn) => vehicleColumnText(vehicle, column, language, t);
  const detailColumns = columns.filter((column) => (
    column !== "image" && !mobileSummaryColumns.has(column)
  ));
  const makerLine = [
    shows("manufacturer") ? text("manufacturer") : "",
    shows("articleNumber") ? text("articleNumber") : ""
  ].filter(Boolean).join(" · ");
  const metaLine = [
    shows("gauge") ? text("gauge") : "",
    shows("epoch") ? text("epoch") : ""
  ].filter(Boolean).join(" · ");
  const detailsID = `vehicle-mobile-details-${vehicle.id}`;

  return (
    <article className={`inventory-mobile-item vehicle-mobile-item${showsImage ? "" : " no-image"}`}>
      <button
        type="button"
        className="vehicle-mobile-toggle"
        onClick={onToggleExpanded}
        aria-expanded={expanded}
        aria-controls={detailsID}
        aria-label={t(expanded ? "vehicles.mobile.collapse" : "vehicles.mobile.expand", {
          inventoryNumber: vehicle.inventoryNumber
        })}
      >
        {showsImage ? (
          <span className="vehicle-mobile-image" aria-hidden="true">
            {image
              ? <img src={previewImageUrl(image)} alt="" />
              : <span className="image-placeholder">{t("exhibition.noPreview")}</span>}
          </span>
        ) : null}
        <span className="vehicle-mobile-summary">
          {shows("inventoryNumber") ? <small>{text("inventoryNumber")}</small> : null}
          {shows("name") ? <strong>{text("name")}</strong> : null}
          {makerLine ? <span>{makerLine}</span> : null}
          {metaLine ? <em>{metaLine}</em> : null}
        </span>
        {expanded
          ? <ChevronUp size={18} aria-hidden="true" />
          : <ChevronDown size={18} aria-hidden="true" />}
      </button>

      {expanded && detailColumns.length > 0 ? (
        <dl className="vehicle-mobile-fields" id={detailsID}>
          {detailColumns.map((column) => (
            <div key={column}>
              <dt>{vehicleColumnLabel(column, t)}</dt>
              <dd>{text(column)}</dd>
            </div>
          ))}
        </dl>
      ) : null}

      <div className="inventory-mobile-actions">
        <button
          type="button"
          className="icon-button"
          onClick={() => onOpenDetail(vehicle)}
          aria-label={t("vehicles.view")}
          title={t("vehicles.view")}
        >
          <Eye size={16} aria-hidden="true" />
        </button>
        <button
          type="button"
          className="icon-button"
          onClick={() => onOpenEdit(vehicle)}
          aria-label={t("vehicles.edit")}
          title={t("vehicles.edit")}
        >
          <Pencil size={16} aria-hidden="true" />
        </button>
        {renderQuickMenu(vehicle)}
      </div>
    </article>
  );
}
```

- [ ] **Step 5: Make the list own and clean expanded state**

Replace `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx` with:

```tsx
import { useEffect, useState, type ReactNode } from "react";

import type { Vehicle } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { VehicleInventoryMobileCard } from "./VehicleInventoryMobileCard";
import type { VehicleTableColumn } from "./vehicleTableColumns";

type VehicleInventoryMobileListProps = {
  vehicles: Vehicle[];
  columns: readonly VehicleTableColumn[];
  onOpenDetail: (vehicle: Vehicle) => void;
  onOpenEdit: (vehicle: Vehicle) => void;
  renderQuickMenu: (vehicle: Vehicle) => ReactNode;
};

export function VehicleInventoryMobileList({
  vehicles,
  columns,
  onOpenDetail,
  onOpenEdit,
  renderQuickMenu
}: VehicleInventoryMobileListProps) {
  const { t } = useI18n();
  const [expandedVehicleIDs, setExpandedVehicleIDs] = useState<Set<string>>(() => new Set());

  useEffect(() => {
    const visibleIDs = new Set(vehicles.map((vehicle) => vehicle.id));
    setExpandedVehicleIDs((current) => {
      const next = new Set([...current].filter((vehicleID) => visibleIDs.has(vehicleID)));
      return next.size === current.size ? current : next;
    });
  }, [vehicles]);

  const toggleExpanded = (vehicleID: string) => {
    setExpandedVehicleIDs((current) => {
      const next = new Set(current);
      if (next.has(vehicleID)) next.delete(vehicleID);
      else next.add(vehicleID);
      return next;
    });
  };

  return (
    <div className="inventory-mobile-list" aria-label={t("vehicles.mobileList")}>
      {vehicles.map((vehicle) => (
        <VehicleInventoryMobileCard
          key={vehicle.id}
          vehicle={vehicle}
          columns={columns}
          expanded={expandedVehicleIDs.has(vehicle.id)}
          onToggleExpanded={() => toggleExpanded(vehicle.id)}
          onOpenDetail={onOpenDetail}
          onOpenEdit={onOpenEdit}
          renderQuickMenu={renderQuickMenu}
        />
      ))}
    </div>
  );
}
```

Remove `onDelete={onDelete}` from the `VehicleInventoryMobileList` call in
`frontend/src/features/vehicles/VehicleInventoryPanel.tsx`. Keep `onDelete` on the panel itself,
the desktop table, desktop cards, and existing quick menu.

- [ ] **Step 6: Run the focused test and TypeScript build**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleInventoryMobileList.test.tsx
npm.cmd run build
```

Expected: all mobile-list tests PASS; TypeScript and Vite build PASS with only the already-known
Vite native-config and chunk-size warnings.

- [ ] **Step 7: Commit the independently working mobile behavior**

```powershell
git add -- frontend/src/features/vehicles/VehicleInventoryMobileCard.tsx frontend/src/features/vehicles/VehicleInventoryMobileList.tsx frontend/src/features/vehicles/VehicleInventoryMobileList.test.tsx frontend/src/features/vehicles/VehicleInventoryPanel.tsx frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat: add expandable mobile vehicle cards"
```

---

### Task 2: Apply compact styling and fix quick-menu clipping

**Files:**
- Modify: `frontend/src/styles/vehicle-inventory.css`
- Modify: `frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts`

**Interfaces:**
- Consumes: class names emitted by `VehicleInventoryMobileCard` and the existing `.quick-menu` markup.
- Produces: a two-column compact detail grid above 420px, one-column details at very narrow widths, visible action-cell overflow, and raised rows/cards while a quick menu is open.

- [ ] **Step 1: Write failing CSS regression expectations**

Replace the final test in `frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts` and
add two focused tests:

```ts
  it("uses an expandable compact mobile card with a narrow fallback", () => {
    expect(css).toMatch(/\.vehicle-mobile-toggle\s*\{[^}]*grid-template-columns:\s*64px\s+minmax\(0,\s*1fr\)\s+24px/s);
    expect(css).toMatch(/\.vehicle-mobile-fields\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/s);
    expect(css).toMatch(/@media\s*\(max-width:\s*420px\)[\s\S]*\.vehicle-mobile-fields\s*\{[^}]*grid-template-columns:\s*1fr/s);
  });

  it("keeps open mobile quick menus above following cards", () => {
    expect(css).toMatch(/\.vehicle-mobile-item:has\(\.quick-menu\)\s*\{[^}]*z-index:\s*5/s);
    expect(css).toMatch(/\.vehicle-mobile-item\s*\{[^}]*overflow:\s*visible/s);
  });

  it("clips long data cells without clipping the desktop action menu", () => {
    expect(css).toMatch(/\.vehicle-inventory-table td:not\(\.actions-cell\)\s*\{[^}]*overflow:\s*hidden/s);
    expect(css).toMatch(/\.vehicle-inventory-table td\.actions-cell\s*\{[^}]*overflow:\s*visible/s);
    expect(css).toMatch(/\.vehicle-inventory-table tbody tr:has\(\.quick-menu\)\s*\{[^}]*z-index:\s*5/s);
  });
```

- [ ] **Step 2: Run the focused CSS test and verify it fails**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleInventoryResponsive.test.ts
```

Expected: FAIL because data-cell overflow currently also applies to `.actions-cell`, and the old
mobile layout has no expandable header or open-menu stacking rule.

- [ ] **Step 3: Replace the current vehicle mobile and table-overflow blocks**

In `frontend/src/styles/vehicle-inventory.css`, keep the existing column-picker styles and replace
the rules beginning at `.vehicle-inventory-table td` through the current mobile media query with:

```css
.vehicle-inventory-table td:not(.actions-cell) {
  overflow: hidden;
  text-overflow: ellipsis;
}

.vehicle-inventory-table tbody tr {
  position: relative;
}

.vehicle-inventory-table tbody tr:has(.quick-menu) {
  z-index: 5;
}

.vehicle-inventory-table td.actions-cell {
  overflow: visible;
}

.vehicle-mobile-item {
  position: relative;
  display: block;
  overflow: visible;
  padding: 0;
}

.vehicle-mobile-item:has(.quick-menu) {
  z-index: 5;
}

.vehicle-mobile-toggle {
  display: grid;
  grid-template-columns: 64px minmax(0, 1fr) 24px;
  align-items: center;
  gap: 10px;
  width: 100%;
  border: 0;
  padding: 10px;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.vehicle-mobile-item.no-image .vehicle-mobile-toggle {
  grid-template-columns: minmax(0, 1fr) 24px;
}

.vehicle-mobile-toggle:hover,
.vehicle-mobile-toggle:focus-visible {
  background: var(--accent-soft);
  outline: none;
}

.vehicle-mobile-toggle:focus-visible {
  box-shadow: inset 0 0 0 2px color-mix(in srgb, var(--accent) 48%, transparent);
}

.vehicle-mobile-image {
  display: grid;
  width: 64px;
  height: 52px;
  place-items: center;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 7px;
  background: #fff;
}

.vehicle-mobile-image img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center;
}

.vehicle-mobile-summary {
  display: grid;
  min-width: 0;
  gap: 2px;
}

.vehicle-mobile-summary small,
.vehicle-mobile-summary span,
.vehicle-mobile-summary em {
  overflow: hidden;
  color: var(--muted);
  font-size: var(--font-size-xs);
  font-style: normal;
  font-weight: var(--font-weight-extra-bold);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.vehicle-mobile-summary strong {
  overflow: hidden;
  color: var(--text);
  font-size: var(--font-size-base);
  line-height: var(--line-height-control-md);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.vehicle-mobile-summary em {
  color: var(--accent-strong);
}

.vehicle-mobile-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 14px;
  min-width: 0;
  margin: 0;
  border-top: 1px solid var(--line);
  padding: 10px;
}

.vehicle-mobile-fields > div {
  min-width: 0;
}

.vehicle-mobile-fields dt {
  color: var(--muted);
  font-size: var(--font-size-xs);
}

.vehicle-mobile-fields dd {
  overflow-wrap: anywhere;
  margin: 2px 0 0;
  color: var(--text);
  font-size: var(--font-size-sm);
}

.vehicle-mobile-item .inventory-mobile-actions {
  display: flex;
  justify-content: flex-end;
  gap: 4px;
  border-top: 1px solid var(--line);
  padding: 6px 8px;
}

@media (max-width: 720px) {
  .vehicle-column-picker-popover {
    position: fixed;
    inset: auto 12px 12px;
    width: auto;
    max-height: calc(100dvh - 24px);
  }

  .vehicle-column-groups {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 420px) {
  .vehicle-mobile-fields {
    grid-template-columns: 1fr;
  }
}
```

- [ ] **Step 4: Run focused component and CSS tests**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/VehicleInventoryMobileList.test.tsx src/features/vehicles/vehicleInventoryResponsive.test.ts
```

Expected: both files PASS, including independent expansion, hidden image, exact detail order,
mobile stacking, and the desktop action-cell overflow exception.

- [ ] **Step 5: Commit the layout and clipping fix**

```powershell
git add -- frontend/src/styles/vehicle-inventory.css frontend/src/features/vehicles/vehicleInventoryResponsive.test.ts
git commit -m "fix: refine mobile inventory and quick menus"
```

---

### Task 3: Add PluX12 from one shared adapter option list

**Files:**
- Modify: `frontend/src/features/vehicles/vehicleOptions.ts`
- Create: `frontend/src/features/vehicles/vehicleOptions.test.ts`
- Modify: `frontend/src/features/exhibition/ExhibitionView.tsx`
- Modify: `frontend/src/features/exhibition/ExhibitionView.test.tsx`

**Interfaces:**
- Consumes: the existing exported `adapterOptions: string[]` used by `VehicleFormFields`.
- Produces: `adapterOptions` ordered as `NEM 651`, `NEM 652`, `PluX12`, `PluX16`, `PluX22`, `MTC21`, `Next18`, `8-polig`, `21-polig`, imported by both vehicle and exhibition forms.

- [ ] **Step 1: Add failing shared-option and exhibition-dropdown tests**

Create `frontend/src/features/vehicles/vehicleOptions.test.ts` with:

```ts
import { describe, expect, it } from "vitest";

import { adapterOptions } from "./vehicleOptions";

describe("vehicle adapter options", () => {
  it("includes PluX12 in numeric PluX order", () => {
    expect(adapterOptions).toEqual([
      "NEM 651",
      "NEM 652",
      "PluX12",
      "PluX16",
      "PluX22",
      "MTC21",
      "Next18",
      "8-polig",
      "21-polig"
    ]);
  });
});
```

Replace the existing Testing Library import in
`frontend/src/features/exhibition/ExhibitionView.test.tsx` and add `userEvent`:

```ts
import userEvent from "@testing-library/user-event";
import { render, screen, waitFor, within } from "@testing-library/react";
```

Add this test inside the existing `describe` block:

```tsx
  it("offers PluX12 from the shared vehicle adapter options", async () => {
    const user = userEvent.setup();
    vi.spyOn(api, "exhibitionLists").mockResolvedValue([{
      id: "list-1",
      designation: "Clubabend",
      date: "2026-08-15",
      locked: false,
      entryCount: 0,
      createdAt: "2026-08-15T18:00:00Z",
      updatedAt: "2026-08-15T18:00:00Z"
    }]);
    vi.spyOn(api, "exhibitionEntries").mockResolvedValue([]);
    vi.spyOn(api, "masterData").mockResolvedValue([]);
    vi.spyOn(api, "masterDataAll").mockResolvedValue({});

    render(<ExhibitionView roles={["Admin"]} />);
    await user.click(await screen.findByRole("button", { name: "Eintrag" }));

    const adapter = screen.getByLabelText("Adapter / Schnittstelle");
    expect(within(adapter).getByRole("option", { name: "PluX12" })).toHaveValue("PluX12");
  });
```

- [ ] **Step 2: Run both focused tests and verify they fail**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleOptions.test.ts src/features/exhibition/ExhibitionView.test.tsx
```

Expected: the option-order test FAILS because `PluX12` is missing, and the exhibition test FAILS
because its local duplicate list does not contain `PluX12`.

- [ ] **Step 3: Add PluX12 to the shared list and remove the exhibition duplicate**

Change the adapter export in `frontend/src/features/vehicles/vehicleOptions.ts` to:

```ts
export const adapterOptions = [
  "NEM 651",
  "NEM 652",
  "PluX12",
  "PluX16",
  "PluX22",
  "MTC21",
  "Next18",
  "8-polig",
  "21-polig"
];
```

In `frontend/src/features/exhibition/ExhibitionView.tsx`, add:

```ts
import { adapterOptions } from "../vehicles/vehicleOptions";
```

Delete the local `const adapterOptions = [...]` declaration. Keep the existing dropdown mapping
unchanged so it now consumes the shared export.

- [ ] **Step 4: Run the focused tests and production build**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/vehicles/vehicleOptions.test.ts src/features/exhibition/ExhibitionView.test.tsx
npm.cmd run build
```

Expected: both test files PASS and the TypeScript/Vite production build PASS.

- [ ] **Step 5: Commit Issue #85 as an independent change**

```powershell
git add -- frontend/src/features/vehicles/vehicleOptions.ts frontend/src/features/vehicles/vehicleOptions.test.ts frontend/src/features/exhibition/ExhibitionView.tsx frontend/src/features/exhibition/ExhibitionView.test.tsx
git commit -m "feat: add PluX12 adapter option"
```

---

### Task 4: Verify the complete frontend behavior locally

**Files:**
- Verify: `frontend/src/features/vehicles/VehicleInventoryMobileCard.tsx`
- Verify: `frontend/src/features/vehicles/VehicleInventoryMobileList.tsx`
- Verify: `frontend/src/styles/vehicle-inventory.css`
- Verify: `frontend/src/features/vehicles/vehicleOptions.ts`
- Verify: `frontend/src/features/exhibition/ExhibitionView.tsx`

**Interfaces:**
- Consumes: the complete implementation from Tasks 1 through 3.
- Produces: automated and visual evidence that the approved mobile hierarchy and quick-menu fix work together without regressions.

- [ ] **Step 1: Run the full frontend test suite**

Run:

```powershell
cd frontend
npm.cmd run test:run
```

Expected: all existing and new frontend tests PASS.

- [ ] **Step 2: Run the production frontend build**

Run:

```powershell
cd frontend
npm.cmd run build
```

Expected: TypeScript and Vite build PASS. Record the known Vite native-config and chunk-size
warnings if they remain; treat any new warning or error as a regression.

- [ ] **Step 3: Run the backend regression suite despite no backend changes**

Run:

```powershell
cd backend
go test ./...
```

Expected: all backend packages PASS.

- [ ] **Step 4: Check the mobile design in the local browser**

At `http://127.0.0.1:8081/vehicles`, use a 390 x 844 viewport and verify:

1. Cards start collapsed and show the selected image, inventory number, designation,
   manufacturer/article number, gauge, and epoch without horizontal scrolling.
2. A vehicle without a file uses the neutral preview only when the image column is visible.
3. Opening one card reveals only remaining selected fields in their persisted order; opening a
   second card does not close or corrupt the first.
4. View and Edit execute their original actions without toggling the card.
5. The quick menu overlays following cards, closes on outside click and Escape, and remains fully
   usable.
6. Long German values wrap only in the expanded detail area; the collapsed card stays compact.
7. Repeat in English and in light and dark themes.

Also open a vehicle form and a new exhibition entry, then verify that both adapter dropdowns show
`PluX12`, `PluX16`, and `PluX22` in that order.

- [ ] **Step 5: Check the desktop quick menu at representative rows**

At a desktop viewport of at least 1440 x 900, verify the first, middle, and last visible table rows:

1. The quick menu is not clipped to the action cell or hidden behind later rows.
2. Long data cells still use ellipsis and do not widen the actions column.
3. View, Edit, QR code, Print, Uploads, Maintenance, Spare parts, and Delete retain their existing
   behavior and permissions.
4. The menu closes on outside click, Escape, and after selecting an action.

- [ ] **Step 6: Review the final diff and working tree**

Run:

```powershell
git diff --check
git status --short
git log -5 --oneline
```

Expected: no whitespace errors; only the intended source/test changes plus the user's existing
untracked design images, Facebook export, and the visual-companion `.superpowers` directory remain.
Do not add or remove those unrelated files.
