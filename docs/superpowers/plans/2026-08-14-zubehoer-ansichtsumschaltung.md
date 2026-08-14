# Zubehörübersicht View Modes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a persisted desktop table/card switch and an automatic compact mobile article list to the accessory overview.

**Architecture:** Keep `useArticleOverview` as the data owner and add presentation-only view state in
`AccessoriesView`. Extract the accessible article action menu once, then reuse it from the existing
table, a new card grid, and a new compact mobile list. CSS selects the mobile presentation at the
existing 900 pixel breakpoint without JavaScript viewport checks.

**Tech Stack:** React 19, TypeScript strict mode, Vite, Vitest, Testing Library, Lucide React, CSS
design tokens, browser `localStorage`.

## Global Constraints

- The table view remains the default.
- Store the desktop choice under `railkeeper.accessories.view`.
- Below 900 pixels always show the compact mobile list and hide the desktop switch.
- Do not add backend, OpenAPI, API request, filtering, sorting, selection, or permission changes.
- Reuse the current `overview.data.items` result and the existing article callbacks.
- Keep German and English UI text aligned.
- Keep icon buttons transparent with color-only hover and focus feedback.
- Keep table selection state when switching desktop views, but add no card or mobile checkboxes.
- Use LF, UTF-8, strict TypeScript, existing design tokens, and focused files below roughly 500 lines.

## File Map

- Create `frontend/src/features/accessories/articleViewMode.ts`: persisted view-mode type and helpers.
- Create `frontend/src/features/accessories/articleViewMode.test.ts`: persistence fallback tests.
- Create `frontend/src/features/accessories/ArticleActions.tsx`: shared view, edit, archive, and restore controls.
- Create `frontend/src/features/accessories/ArticleActions.test.tsx`: action, focus, and keyboard tests.
- Create `frontend/src/features/accessories/ArticleCardGrid.tsx`: desktop accessory card grid.
- Create `frontend/src/features/accessories/ArticleCardGrid.test.tsx`: card content and role tests.
- Create `frontend/src/features/accessories/ArticleCompactList.tsx`: automatic compact mobile list.
- Create `frontend/src/features/accessories/ArticleCompactList.test.tsx`: mobile content and action tests.
- Create `frontend/src/test/fixtures/accessories.ts`: reusable complete accessory article fixture.
- Modify `frontend/src/features/accessories/ArticleTable.tsx`: consume `ArticleActions` and remove duplicate menu state.
- Modify `frontend/src/features/accessories/ArticleTable.test.tsx`: retain table-specific and shared-action integration tests.
- Modify `frontend/src/features/accessories/ArticleToolbar.tsx`: add the desktop view switch.
- Modify `frontend/src/features/accessories/AccessoriesView.tsx`: own view state and select desktop presentation.
- Modify `frontend/src/features/accessories/AccessoriesView.test.tsx`: integration and persistence coverage.
- Modify `frontend/src/features/accessories/accessoriesResponsive.test.ts`: enforce the desktop/mobile CSS contract.
- Modify `frontend/src/shared/i18n/de.ts`: German view labels.
- Modify `frontend/src/shared/i18n/en.ts`: English view labels.
- Modify `frontend/src/styles/accessories.css`: toolbar, cards, compact list, actions, and responsive rules.

---

### Task 1: Persisted accessory view mode

**Files:**
- Create: `frontend/src/features/accessories/articleViewMode.ts`
- Create: `frontend/src/features/accessories/articleViewMode.test.ts`

**Interfaces:**
- Produces: `ArticleViewMode = "table" | "cards"`.
- Produces: `articleViewSettingKey = "railkeeper.accessories.view"`.
- Produces: `storedArticleViewMode(storage?): ArticleViewMode`.
- Produces: `persistArticleViewMode(mode, storage?): void`.

- [ ] **Step 1: Write the failing persistence tests**

```ts
import { beforeEach, describe, expect, it } from "vitest";

import {
  articleViewSettingKey,
  persistArticleViewMode,
  storedArticleViewMode
} from "./articleViewMode";

describe("article view mode", () => {
  beforeEach(() => window.localStorage.clear());

  it("defaults missing and unknown values to table", () => {
    expect(storedArticleViewMode()).toBe("table");
    window.localStorage.setItem(articleViewSettingKey, "unknown");
    expect(storedArticleViewMode()).toBe("table");
  });

  it("persists and restores cards", () => {
    persistArticleViewMode("cards");
    expect(window.localStorage.getItem(articleViewSettingKey)).toBe("cards");
    expect(storedArticleViewMode()).toBe("cards");
  });
});
```

- [ ] **Step 2: Run the test and confirm the missing-module failure**

Run: `cd frontend; npm.cmd test -- --run src/features/accessories/articleViewMode.test.ts`

Expected: FAIL because `./articleViewMode` does not exist.

- [ ] **Step 3: Add the minimal typed storage owner**

```ts
export type ArticleViewMode = "table" | "cards";

export const articleViewSettingKey = "railkeeper.accessories.view";

type ViewStorage = Pick<Storage, "getItem" | "setItem">;

export function storedArticleViewMode(
  storage: ViewStorage = window.localStorage
): ArticleViewMode {
  return storage.getItem(articleViewSettingKey) === "cards" ? "cards" : "table";
}

export function persistArticleViewMode(
  mode: ArticleViewMode,
  storage: ViewStorage = window.localStorage
) {
  storage.setItem(articleViewSettingKey, mode);
}
```

- [ ] **Step 4: Run the focused tests**

Run: `cd frontend; npm.cmd test -- --run src/features/accessories/articleViewMode.test.ts`

Expected: 1 test file and 2 tests pass.

- [ ] **Step 5: Commit the persistence owner**

```powershell
git add -- frontend/src/features/accessories/articleViewMode.ts frontend/src/features/accessories/articleViewMode.test.ts
git commit -m "feat(accessories): persist overview view mode"
```

---

### Task 2: Shared article actions

**Files:**
- Create: `frontend/src/features/accessories/ArticleActions.tsx`
- Create: `frontend/src/features/accessories/ArticleActions.test.tsx`
- Create: `frontend/src/test/fixtures/accessories.ts`
- Modify: `frontend/src/features/accessories/ArticleTable.tsx`
- Modify: `frontend/src/features/accessories/ArticleTable.test.tsx`
- Modify: `frontend/src/styles/accessories.css`

**Interfaces:**
- Consumes: `AccessoryArticleListItem` and the existing view, edit, archive, and restore callbacks.
- Produces: `ArticleActions` with identical behavior in table, card, and mobile contexts.

- [ ] **Step 1: Write failing shared-action tests**

Create the reusable fixture first:

```ts
import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";

export function accessoryArticleFixture(
  overrides: Partial<AccessoryArticleListItem> = {}
): AccessoryArticleListItem {
  return {
    id: "article-1", inventoryNumber: "RK-ART-000001", manufacturer: "Tillig",
    articleNumber: "83101", name: "Gerades Modellgleis", articleType: "track",
    subtype: "straight", gauges: ["TT"], inventoryStrategy: "quantity", archived: false,
    owned: 18, available: 12, reserved: 4, installed: 2,
    locationNames: ["Werkstatt"], hasUsageHistory: true, careHintCount: 0,
    updatedAt: "2026-08-08T10:00:00Z", attributes: [],
    ...overrides
  };
}

const timestamp = "2026-08-08T10:00:00Z";

export const accessoryArticleTypes: MasterDataEntry[] = [{
  id: "article-type-track", type: "article_type", key: "track", label: "Gleis",
  active: true, sortOrder: 10, metadata: {}, createdAt: timestamp, updatedAt: timestamp
}];

export const accessorySubtypes: MasterDataEntry[] = [{
  id: "article-subtype-track-straight", type: "accessory_subtype", key: "track:straight",
  label: "Gerade", active: true, sortOrder: 10, metadata: {},
  createdAt: timestamp, updatedAt: timestamp
}];
```

Then cover callback routing plus keyboard focus:

```tsx
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { accessoryArticleFixture } from "../../test/fixtures/accessories";
import { ArticleActions } from "./ArticleActions";

const article = accessoryArticleFixture();

describe("ArticleActions", () => {
  it("routes view, edit, and archive for writers", async () => {
    const user = userEvent.setup();
    const onView = vi.fn();
    const onEdit = vi.fn();
    const onArchive = vi.fn();
    render(<ArticleActions article={article} canEdit onView={onView} onEdit={onEdit}
      onArchive={onArchive} onRestore={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" }));
    await user.click(screen.getByRole("button", { name: "Artikel bearbeiten: Gerades Modellgleis" }));
    await user.click(screen.getByRole("button", { name: "Weitere Aktionen: Gerades Modellgleis" }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel archivieren" }));

    expect(onView).toHaveBeenCalledWith(article);
    expect(onEdit).toHaveBeenCalledWith(article);
    expect(onArchive).toHaveBeenCalledWith(article);
  });

  it("moves focus into the menu and restores it on Escape", async () => {
    const user = userEvent.setup();
    render(<ArticleActions article={article} canEdit onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);
    const trigger = screen.getByRole("button", { name: "Weitere Aktionen: Gerades Modellgleis" });

    await user.click(trigger);
    expect(screen.getByRole("menuitem")).toHaveFocus();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });
});
```

- [ ] **Step 2: Run the new test and confirm the missing-component failure**

Run: `cd frontend; npm.cmd test -- --run src/features/accessories/ArticleActions.test.tsx`

Expected: FAIL because `ArticleActions` does not exist.

- [ ] **Step 3: Extract the canonical action implementation**

Create `ArticleActions.tsx` with the existing labels and focus behavior:

```tsx
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { Eye, MoreHorizontal, Pencil } from "lucide-react";

import type { AccessoryArticleListItem } from "../../shared/api";
import { useI18n } from "../../shared/i18n";

type ArticleActionsProps = {
  article: AccessoryArticleListItem;
  canEdit: boolean;
  onView?: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

export function ArticleActions({ article, canEdit, onView, onEdit, onArchive, onRestore }: ArticleActionsProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const { t } = useI18n();

  useLayoutEffect(() => {
    if (open) menuRef.current?.querySelector<HTMLButtonElement>("[role='menuitem']")?.focus();
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const closeOutside = (event: PointerEvent) => {
      if (event.target instanceof Node && rootRef.current?.contains(event.target)) return;
      setOpen(false);
    };
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key !== "Escape") return;
      event.preventDefault();
      setOpen(false);
      triggerRef.current?.focus();
    };
    document.addEventListener("pointerdown", closeOutside);
    document.addEventListener("keydown", closeOnEscape);
    return () => {
      document.removeEventListener("pointerdown", closeOutside);
      document.removeEventListener("keydown", closeOnEscape);
    };
  }, [open]);

  const handleMenuKeyDown = (event: React.KeyboardEvent<HTMLDivElement>) => {
    if (["ArrowDown", "ArrowUp", "Home", "End"].includes(event.key)) {
      event.preventDefault();
      menuRef.current?.querySelector<HTMLButtonElement>("[role='menuitem']")?.focus();
    } else if (event.key === "Tab") {
      setOpen(false);
    }
  };

  return <div ref={rootRef} className="table-actions article-row-actions">
    {onView ? <button type="button" className="icon-button article-action-button"
      onClick={() => onView(article)} aria-label={t("accessories.actions.viewNamed", { name: article.name })}
      title={t("accessories.actions.view")}><Eye size={16} aria-hidden="true" /></button> : null}
    {canEdit && onEdit ? <button type="button" className="icon-button article-action-button"
      onClick={() => onEdit(article)} aria-label={t("accessories.actions.editNamed", { name: article.name })}
      title={t("accessories.actions.edit")}><Pencil size={16} aria-hidden="true" /></button> : null}
    {canEdit ? <div className="article-overflow">
      <button ref={triggerRef} type="button" className="icon-button article-action-button"
        onClick={() => setOpen((current) => !current)}
        aria-label={t("accessories.actions.moreNamed", { name: article.name })}
        title={t("accessories.actions.more")} aria-haspopup="menu" aria-expanded={open}>
        <MoreHorizontal size={17} aria-hidden="true" />
      </button>
      {open ? <div ref={menuRef} className="article-action-menu" role="menu"
        onKeyDown={handleMenuKeyDown}>
        <button type="button" role="menuitem" onClick={() => {
          setOpen(false);
          void (article.archived ? onRestore(article) : onArchive(article));
        }}>{t(article.archived ? "accessories.actions.restore" : "accessories.actions.archive")}</button>
      </div> : null}
    </div> : null}
  </div>;
}
```

- [ ] **Step 4: Replace table-local action state with `ArticleActions`**

Remove the menu hooks and Lucide action imports from `ArticleTable.tsx`, import `ArticleActions`, and
replace the entire current `<div className="table-actions article-row-actions">` block with:

```tsx
<ArticleActions
  article={article}
  canEdit={canEdit}
  onView={onView}
  onEdit={onEdit}
  onArchive={onArchive}
  onRestore={onRestore}
/>
```

Keep the wrapper stable and let CSS detect an open descendant:

```tsx
<div className="table-wrap article-table-wrap">
```

Update the menu reserve selectors in `accessories.css`:

```css
.article-table-wrap:has(.article-action-menu) {
  padding-bottom: 48px;
}

.article-table-wrap:has(.article-action-menu) .article-table {
  overflow: visible;
}
```

- [ ] **Step 5: Keep table integration assertions and remove owner-specific expectations**

In `ArticleTable.test.tsx`, keep the existing view/edit/archive assertions. Replace the
`.menu-open` assertion with:

```tsx
expect(screen.getByRole("menu").closest(".article-table-wrap")).not.toBeNull();
```

Move detailed Arrow, Home, End, Escape, Tab, and outside-click expectations to
`ArticleActions.test.tsx`. Add explicit Arrow/Home/End and Tab cases there before deleting them from
the table test.

- [ ] **Step 6: Run shared actions and table tests**

Run:
`cd frontend; npm.cmd test -- --run src/features/accessories/ArticleActions.test.tsx src/features/accessories/ArticleTable.test.tsx`

Expected: both test files pass and the existing archive/restore behavior remains covered.

- [ ] **Step 7: Commit the action extraction**

```powershell
git add -- frontend/src/features/accessories/ArticleActions.tsx frontend/src/features/accessories/ArticleActions.test.tsx frontend/src/test/fixtures/accessories.ts frontend/src/features/accessories/ArticleTable.tsx frontend/src/features/accessories/ArticleTable.test.tsx frontend/src/styles/accessories.css
git commit -m "refactor(accessories): share article row actions"
```

---

### Task 3: Desktop article card grid

**Files:**
- Create: `frontend/src/features/accessories/ArticleCardGrid.tsx`
- Create: `frontend/src/features/accessories/ArticleCardGrid.test.tsx`
- Modify: `frontend/src/styles/accessories.css`

**Interfaces:**
- Consumes: filtered and sorted `AccessoryArticleListItem[]`, master-data labels, permissions, and existing callbacks.
- Consumes: `ArticleActions` from Task 2.
- Produces: `ArticleCardGrid` for the desktop cards mode.

- [ ] **Step 1: Write failing card content and permission tests**

Import `accessoryArticleFixture`, `accessoryArticleTypes`, and `accessorySubtypes` from
`../../test/fixtures/accessories`, create `const article = accessoryArticleFixture()`, and assert
these signals:

```tsx
render(<ArticleCardGrid items={[article]} articleTypeEntries={accessoryArticleTypes}
  subtypeEntries={accessorySubtypes} canEdit onView={onView} onEdit={vi.fn()}
  onArchive={vi.fn()} onRestore={vi.fn()} />);

expect(screen.getByRole("list", { name: "Artikel-Kachelansicht" })).toBeInTheDocument();
expect(screen.getByText("RK-ART-000001")).toBeInTheDocument();
expect(screen.getByText("Tillig")).toBeInTheDocument();
expect(screen.getByText("Gerades Modellgleis")).toBeInTheDocument();
expect(screen.getByText("18 gesamt")).toBeInTheDocument();
expect(screen.getByText("12 frei · 4 reserviert · 2 eingebaut")).toBeInTheDocument();
await user.click(screen.getByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" }));
expect(onView).toHaveBeenCalledWith(article);
```

Render again with `canEdit={false}` and assert that edit and more-action buttons are absent while
the view button remains.

- [ ] **Step 2: Run the card test and confirm the missing-component failure**

Run: `cd frontend; npm.cmd test -- --run src/features/accessories/ArticleCardGrid.test.tsx`

Expected: FAIL because `ArticleCardGrid` does not exist.

- [ ] **Step 3: Implement the focused card component**

Create a list whose mapped article uses this complete structural shape:

```tsx
import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleActions } from "./ArticleActions";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";

type ArticleCardGridProps = {
  items: AccessoryArticleListItem[];
  articleTypeEntries: MasterDataEntry[];
  subtypeEntries: MasterDataEntry[];
  canEdit: boolean;
  onView: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

export function ArticleCardGrid({
  items, articleTypeEntries, subtypeEntries, canEdit,
  onView, onEdit, onArchive, onRestore
}: ArticleCardGridProps) {
  const { t } = useI18n();

  return (
<div className="article-card-grid" role="list" aria-label={t("accessories.view.cardsLabel")}>
  {items.map((article) => {
    const primaryLocation = article.locationNames[0] || t("common.none");
    const typeLabel = articleTypeLabel(article.articleType, articleTypeEntries, t);
    const subtypeLabel = article.subtype
      ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
      : t("common.none");
    return <article key={article.id} className="article-card" role="listitem">
      <button type="button" className="article-card-media" onClick={() => onView(article)}
        aria-label={t("accessories.actions.viewNamed", { name: article.name })}>
        {article.primaryImageUrl
          ? <img src={article.primaryImageUrl} alt="" />
          : <div className="image-placeholder">{t("exhibition.noPreview")}</div>}
      </button>
      <div className="article-card-body">
        <div className="article-card-title">
          <div><strong title={article.inventoryNumber}>{article.inventoryNumber}</strong>
            <span title={article.manufacturer}>{article.manufacturer}</span></div>
          <span className="article-card-gauge">{article.gauges.join(", ") || t("common.none")}</span>
        </div>
        <button type="button" className="article-name-button" onClick={() => onView(article)}>
          <strong title={article.name}>{article.name}</strong>
        </button>
        <dl>
          <div><dt>{t("accessories.table.articleNumber")}</dt><dd>{article.articleNumber || t("common.none")}</dd></div>
          <div><dt>{t("accessories.table.type")}</dt><dd><strong>{typeLabel}</strong><small>{subtypeLabel}</small></dd></div>
          <div><dt>{t("accessories.table.storage")}</dt>
            <dd title={article.locationNames.join(", ")}>{primaryLocation}</dd>
            {article.locationNames.length > 1
              ? <small>{t("accessories.table.moreLocations", { count: article.locationNames.length - 1 })}</small>
              : null}
          </div>
        </dl>
        <div className="article-card-stock">
          <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
          <small>{t("accessories.table.stockBreakdown", {
            available: article.available, reserved: article.reserved, installed: article.installed
          })}</small>
        </div>
        <ArticleActions article={article} canEdit={canEdit} onView={onView} onEdit={onEdit}
          onArchive={onArchive} onRestore={onRestore} />
      </div>
    </article>;
  })}
</div>
  );
}
```

Import `articleSubtypeLabel` and `articleTypeLabel` for the exact label calls above.

- [ ] **Step 4: Add accessory-owned card styling**

Add token-based styles for `.article-card-grid`, `.article-card`, `.article-card-media`,
`.article-card-body`, `.article-card-title`, `.article-card-gauge`, `.article-card-stock`, and their
text truncation. Use `grid-template-columns: repeat(auto-fill, minmax(260px, 1fr))`, a 150 pixel
media row, `var(--panel)`, `var(--panel-subtle)`, `var(--line)`, and existing font tokens. Add:

```css
.article-card:has(.article-action-menu) {
  z-index: 5;
}

.article-card .article-row-actions {
  margin-top: auto;
}
```

- [ ] **Step 5: Run the card tests**

Run: `cd frontend; npm.cmd test -- --run src/features/accessories/ArticleCardGrid.test.tsx`

Expected: all card content, callback, image fallback, and read-only assertions pass.

- [ ] **Step 6: Commit the desktop card grid**

```powershell
git add -- frontend/src/features/accessories/ArticleCardGrid.tsx frontend/src/features/accessories/ArticleCardGrid.test.tsx frontend/src/styles/accessories.css
git commit -m "feat(accessories): add article card grid"
```

---

### Task 4: Compact mobile article list

**Files:**
- Create: `frontend/src/features/accessories/ArticleCompactList.tsx`
- Create: `frontend/src/features/accessories/ArticleCompactList.test.tsx`
- Modify: `frontend/src/styles/accessories.css`

**Interfaces:**
- Consumes: the same articles, labels, permissions, and callbacks as `ArticleCardGrid`.
- Consumes: `ArticleActions` from Task 2.
- Produces: `ArticleCompactList`, rendered at all times but exposed only below 900 pixels by CSS.

- [ ] **Step 1: Write failing compact-list tests**

Import `accessoryArticleFixture`, `accessoryArticleTypes`, and `accessorySubtypes` from
`../../test/fixtures/accessories`, create `const article = accessoryArticleFixture()`, and assert
the compact signals and shared actions:

```tsx
render(<ArticleCompactList items={[article]} articleTypeEntries={accessoryArticleTypes}
  subtypeEntries={accessorySubtypes} canEdit onView={onView} onEdit={vi.fn()}
  onArchive={vi.fn()} onRestore={vi.fn()} />);

const list = screen.getByRole("list", { name: "Kompakte Artikelliste" });
expect(list).toHaveClass("article-mobile-list");
expect(within(list).getByText("RK-ART-000001")).toBeInTheDocument();
expect(within(list).getByText("Gerades Modellgleis")).toBeInTheDocument();
expect(within(list).getByText(/Tillig.*83101/)).toBeInTheDocument();
expect(within(list).getByText("TT")).toBeInTheDocument();
expect(within(list).getByText("18 gesamt")).toBeInTheDocument();
```

Add a missing-image case and a `canEdit={false}` case.

- [ ] **Step 2: Run the compact-list test and confirm the missing-component failure**

Run: `cd frontend; npm.cmd test -- --run src/features/accessories/ArticleCompactList.test.tsx`

Expected: FAIL because `ArticleCompactList` does not exist.

- [ ] **Step 3: Implement the compact article component**

Create the list with this structure for every article:

```tsx
import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";
import { useI18n } from "../../shared/i18n";
import { ArticleActions } from "./ArticleActions";
import { articleSubtypeLabel } from "./articleSubtypes";
import { articleTypeLabel } from "./articleTypes";

type ArticleCompactListProps = {
  items: AccessoryArticleListItem[];
  articleTypeEntries: MasterDataEntry[];
  subtypeEntries: MasterDataEntry[];
  canEdit: boolean;
  onView: (article: AccessoryArticleListItem) => void;
  onEdit?: (article: AccessoryArticleListItem) => void;
  onArchive: (article: AccessoryArticleListItem) => void | Promise<void>;
  onRestore: (article: AccessoryArticleListItem) => void | Promise<void>;
};

export function ArticleCompactList({
  items, articleTypeEntries, subtypeEntries, canEdit,
  onView, onEdit, onArchive, onRestore
}: ArticleCompactListProps) {
  const { t } = useI18n();

  return (
<div className="article-mobile-list" role="list" aria-label={t("accessories.view.mobileLabel")}>
  {items.map((article) => {
    const typeLabel = articleTypeLabel(article.articleType, articleTypeEntries, t);
    const subtypeLabel = article.subtype
      ? articleSubtypeLabel(article.articleType, article.subtype, subtypeEntries, t)
      : typeLabel;
    return <article key={article.id} className="article-mobile-item" role="listitem">
    <button type="button" className="article-mobile-media" onClick={() => onView(article)}
      aria-label={t("accessories.actions.viewNamed", { name: article.name })}>
      {article.primaryImageUrl
        ? <img src={article.primaryImageUrl} alt="" />
        : <div className="image-placeholder">{t("exhibition.noPreview")}</div>}
    </button>
    <button type="button" className="article-mobile-main" onClick={() => onView(article)}>
      <span>{article.inventoryNumber}</span>
      <strong>{article.name}</strong>
      <small>{article.manufacturer || t("common.none")} · {article.articleNumber || t("common.none")}</small>
    </button>
    <div className="article-mobile-meta">
      <span>{article.gauges.join(", ") || t("common.none")}</span>
      <small>{subtypeLabel}</small>
      <strong>{t("accessories.table.stockOwned", { count: article.owned })}</strong>
    </div>
    <div className="article-mobile-actions">
      <ArticleActions article={article} canEdit={canEdit} onView={onView} onEdit={onEdit}
        onArchive={onArchive} onRestore={onRestore} />
    </div>
  </article>;
  })}
</div>
  );
}
```

Import `articleSubtypeLabel` and `articleTypeLabel` for the exact label calls above.

- [ ] **Step 4: Add compact list base styling**

Keep `.article-mobile-list { display: none; }` in the base rules. Style `.article-mobile-item` as a
dense grid with image, main copy, meta, and actions. Use a 58 pixel image, truncation for long names,
transparent action buttons, and existing panel/line tokens. Add an open-menu z-index rule:

```css
.article-mobile-item:has(.article-action-menu) {
  position: relative;
  z-index: 5;
}
```

- [ ] **Step 5: Run compact-list and shared-action tests**

Run:
`cd frontend; npm.cmd test -- --run src/features/accessories/ArticleCompactList.test.tsx src/features/accessories/ArticleActions.test.tsx`

Expected: both test files pass.

- [ ] **Step 6: Commit the compact mobile list**

```powershell
git add -- frontend/src/features/accessories/ArticleCompactList.tsx frontend/src/features/accessories/ArticleCompactList.test.tsx frontend/src/styles/accessories.css
git commit -m "feat(accessories): add compact mobile article list"
```

---

### Task 5: Integrate toolbar switch and responsive presentations

**Files:**
- Modify: `frontend/src/features/accessories/ArticleToolbar.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/features/accessories/accessoriesResponsive.test.ts`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/accessories.css`

**Interfaces:**
- Consumes: `ArticleViewMode`, `storedArticleViewMode`, and `persistArticleViewMode` from Task 1.
- Consumes: `ArticleCardGrid` from Task 3 and `ArticleCompactList` from Task 4.
- Produces: complete user-visible table/card switching with automatic mobile presentation.

- [ ] **Step 1: Write failing integration and persistence coverage**

Clear the accessory setting in `AccessoriesView.test.tsx` `beforeEach`, then add:

```tsx
it("switches desktop views and restores the persisted choice", async () => {
  const user = userEvent.setup();
  const view = render(<AccessoriesView roles={["Editor"]} />);
  await screen.findByText("Gerades Modellgleis");

  expect(screen.getByRole("button", { name: "Tabellenansicht" })).toHaveClass("active");
  expect(screen.getByRole("table")).toBeInTheDocument();

  await user.click(screen.getByRole("button", { name: "Kachelansicht" }));
  expect(screen.getByRole("button", { name: "Kachelansicht" })).toHaveClass("active");
  expect(screen.getByRole("list", { name: "Artikel-Kachelansicht" })).toBeInTheDocument();
  expect(window.localStorage.getItem("railkeeper.accessories.view")).toBe("cards");

  view.unmount();
  render(<AccessoriesView roles={["Editor"]} />);
  await screen.findByRole("list", { name: "Artikel-Kachelansicht" });
  expect(screen.getByRole("button", { name: "Kachelansicht" })).toHaveClass("active");
});
```

Also assert that switching views does not add an `api.accessoryArticles` call and that
`article-mobile-list` remains rendered beside the selected desktop presentation.

Add the English control check and restore German after the assertion:

```tsx
it("localizes the view controls", async () => {
  setLanguage("en");
  vi.mocked(api.accessoryArticles).mockResolvedValueOnce({
    ...overview,
    items: overview.items.map((item) => ({ ...item, name: "Straight model track" }))
  });
  render(<AccessoriesView roles={["Viewer"]} />);
  await screen.findByText("Straight model track");
  expect(screen.getByRole("button", { name: "Table view" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Card view" })).toBeInTheDocument();
  setLanguage("de");
});
```

- [ ] **Step 2: Add failing responsive CSS assertions**

Extend `accessoriesResponsive.test.ts` with:

```ts
it("switches accessory presentations at the existing mobile breakpoint", () => {
  expect(accessoriesCss).toMatch(/\.article-mobile-list\s*\{[^}]*display:\s*none/s);
  expect(accessoriesCss).toMatch(
    /@media\s*\(max-width:\s*900px\)[\s\S]*?\.article-desktop-content\s*\{[^}]*display:\s*none/s
  );
  expect(accessoriesCss).toMatch(
    /@media\s*\(max-width:\s*900px\)[\s\S]*?\.article-mobile-list\s*\{[^}]*display:\s*grid/s
  );
  expect(accessoriesCss).toMatch(
    /@media\s*\(max-width:\s*900px\)[\s\S]*?\.article-view-tools\s*\{[^}]*display:\s*none/s
  );
});
```

- [ ] **Step 3: Add localized view labels**

Add these German keys beside the accessory toolbar keys:

```ts
"accessories.view.label": "Ansicht wechseln",
"accessories.view.table": "Tabellenansicht",
"accessories.view.cards": "Kachelansicht",
"accessories.view.cardsLabel": "Artikel-Kachelansicht",
"accessories.view.mobileLabel": "Kompakte Artikelliste",
```

Add the corresponding English keys:

```ts
"accessories.view.label": "Change view",
"accessories.view.table": "Table view",
"accessories.view.cards": "Card view",
"accessories.view.cardsLabel": "Article card view",
"accessories.view.mobileLabel": "Compact article list",
```

- [ ] **Step 4: Add the toolbar interface and controls**

Extend `ArticleToolbar` props with:

```ts
viewMode: ArticleViewMode;
onViewModeChange: (mode: ArticleViewMode) => void;
```

Import `Grid2X2` and `Table2`, wrap search and controls in `.article-toolbar-primary`, and add:

```tsx
<span className="inventory-view-tools article-view-tools" aria-label={t("accessories.view.label")}>
  <button type="button" className={viewMode === "table" ? "icon-button active" : "icon-button"}
    onClick={() => onViewModeChange("table")} aria-label={t("accessories.view.table")}
    title={t("accessories.view.table")} aria-pressed={viewMode === "table"}>
    <Table2 size={16} aria-hidden="true" />
  </button>
  <button type="button" className={viewMode === "cards" ? "icon-button active" : "icon-button"}
    onClick={() => onViewModeChange("cards")} aria-label={t("accessories.view.cards")}
    title={t("accessories.view.cards")} aria-pressed={viewMode === "cards"}>
    <Grid2X2 size={16} aria-hidden="true" />
  </button>
</span>
```

- [ ] **Step 5: Integrate state and all three presentations**

In `AccessoriesView`, initialize and update the mode:

```ts
const [viewMode, setViewMode] = useState<ArticleViewMode>(storedArticleViewMode);

const changeViewMode = (mode: ArticleViewMode) => {
  setViewMode(mode);
  persistArticleViewMode(mode);
};
```

Pass `viewMode` and `changeViewMode` to `ArticleToolbar`. Replace the single `ArticleTable` result
branch with both responsive surfaces:

```tsx
<>
  <ArticleCompactList items={overview.data.items} articleTypeEntries={articleTypeEntries}
    subtypeEntries={subtypeEntries} canEdit={canEdit}
    onView={(article) => openArticle(article, "view")}
    onEdit={(article) => openArticle(article, "edit")}
    onArchive={(article) => overview.archiveArticle(article.id)}
    onRestore={(article) => overview.restoreArticle(article.id)} />
  <div className="article-desktop-content">
    {viewMode === "cards" ? <ArticleCardGrid items={overview.data.items}
      articleTypeEntries={articleTypeEntries} subtypeEntries={subtypeEntries} canEdit={canEdit}
      onView={(article) => openArticle(article, "view")}
      onEdit={(article) => openArticle(article, "edit")}
      onArchive={(article) => overview.archiveArticle(article.id)}
      onRestore={(article) => overview.restoreArticle(article.id)} /> : <ArticleTable
      items={overview.data.items} articleTypeEntries={articleTypeEntries} subtypeEntries={subtypeEntries}
      sort={overview.sort} direction={overview.direction} canEdit={canEdit}
      selectedIDs={selectedArticleIDs} onToggleSelection={toggleSelection}
      onToggleAll={toggleAll} onSort={overview.setSort}
      onView={(article) => openArticle(article, "view")}
      onEdit={(article) => openArticle(article, "edit")}
      onArchive={(article) => overview.archiveArticle(article.id)}
      onRestore={(article) => overview.restoreArticle(article.id)} />}
  </div>
</>
```

Extract the existing inline selection callbacks as local `toggleSelection` and `toggleAll`
functions before the JSX so the integration stays readable and preserves current selection behavior:

```ts
const toggleSelection = (id: string) => setSelectedArticleIDs((current) => {
  const next = new Set(current);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  return next;
});

const toggleAll = () => setSelectedArticleIDs((current) => {
  const visibleIDs = overview.data.items.map((item) => item.id);
  const allSelected = visibleIDs.length > 0 && visibleIDs.every((id) => current.has(id));
  return allSelected ? new Set() : new Set(visibleIDs);
});
```

- [ ] **Step 6: Complete toolbar and responsive CSS**

Add `.article-toolbar-primary` as a flex row in grid column 2, let `.article-search-field` grow, and
reuse the transparent `.inventory-toolbar-actions .icon-button` behavior for `.article-view-tools`.
At the existing 920 pixel toolbar breakpoint move `.article-toolbar-primary` to column 1. Add:

```css
@media (max-width: 900px) {
  .article-desktop-content,
  .article-view-tools {
    display: none;
  }

  .article-mobile-list {
    display: grid;
  }
}
```

- [ ] **Step 7: Run the complete focused accessory suite**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/accessories/articleViewMode.test.ts src/features/accessories/ArticleActions.test.tsx src/features/accessories/ArticleCardGrid.test.tsx src/features/accessories/ArticleCompactList.test.tsx src/features/accessories/ArticleTable.test.tsx src/features/accessories/AccessoriesView.test.tsx src/features/accessories/accessoriesResponsive.test.ts
```

Expected: all listed files pass. There is no extra API call when changing views.

- [ ] **Step 8: Commit the integrated feature**

```powershell
git add -- frontend/src/features/accessories/ArticleToolbar.tsx frontend/src/features/accessories/AccessoriesView.tsx frontend/src/features/accessories/AccessoriesView.test.tsx frontend/src/features/accessories/accessoriesResponsive.test.ts frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/accessories.css
git commit -m "feat(accessories): switch article overview views"
```

---

### Task 6: Full verification and browser acceptance

**Files:**
- Verify only. Do not commit `frontend/dist`, `.cache`, local data, screenshots, or browser artifacts.

**Interfaces:**
- Consumes: the complete feature from Tasks 1 through 5.
- Produces: fresh automated and browser evidence for the exact approved design.

- [ ] **Step 1: Run the full frontend test suite**

Run: `cd frontend; npm.cmd test -- --run`

Expected: every frontend test file passes.

- [ ] **Step 2: Build the production frontend**

Run: `cd frontend; npm.cmd run build`

Expected: TypeScript succeeds and Vite completes the production build.

- [ ] **Step 3: Check repository whitespace and task scope**

Run:

```powershell
git diff --check
git status --short
```

Expected: `git diff --check` is empty. Only intentional task changes remain, or the tree is clean
after the task commits. Generated `frontend/dist` stays untracked or ignored.

- [ ] **Step 4: Verify desktop table and card behavior in the browser**

Open `/accessories` on the current local server. At a desktop viewport verify:

1. Table is active after clearing `railkeeper.accessories.view`.
2. Switching to cards changes only the presentation and keeps result count, filters, and order.
3. Reload restores cards.
4. Image, inventory number, manufacturer, article number, type, subtype, gauges, stock, and storage are readable.
5. View and edit open the expected article mode.
6. The three-point menu is fully visible and Escape restores focus.
7. Viewer or Planner access exposes view but not edit/archive controls.
8. Browser console contains no warnings or errors caused by the feature.

- [ ] **Step 5: Verify the automatic mobile list**

At a viewport below 900 pixels verify:

1. The view switch is hidden.
2. Neither the desktop table nor card grid is visible.
3. The compact list shows image, identity, type, gauge, stock, and actions.
4. Long German names truncate without breaking actions.
5. Returning above 900 pixels restores the persisted desktop mode.

- [ ] **Step 6: Verify light and dark themes**

In both themes confirm card borders, placeholder contrast, active view icon, focus rings, hover
colors, and the action menu remain legible without boxed icon-button hover effects.

- [ ] **Step 7: Record the final evidence without another source change**

Capture in the task closeout:

```text
Focused accessory tests: pass/fail with file and test counts
Full frontend suite: pass/fail with file and test counts
Production build: pass/fail with transformed module count
Desktop browser: table, cards, persistence, actions, console
Mobile browser: compact list, hidden switch, restored desktop choice
Themes: light and dark checked
Unverified scope: exact remaining gaps, or none
```

Do not create another commit unless verification finds a real defect that requires a separately
tested repair.
