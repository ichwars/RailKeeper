# Einstellungen: Daten-Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the complete manufacturer database under Settings → Data, move all article master data into the same area, and derive the settings page header from the active main tab.

**Architecture:** `SettingsView` remains the owner of the active main tab and composes the existing data and article settings panels. A small pure query model owns backward-compatible `tab`, `group`, and `type` normalization. One reusable app-owned tab-list component provides semantic and keyboard behavior while the existing manufacturer data path remains the only manufacturer editor.

**Tech Stack:** React 19, TypeScript strict mode, Vite, Vitest, Testing Library, existing RailKeeper i18n and CSS tokens.

## Global Constraints

- Work inline in the existing worktree; do not dispatch subagents.
- Do not add a database migration, API route, adapter, or second manufacturer data model.
- Preserve `master_data_entries` type `manufacturer` and its complete editor fields.
- Remove the reduced manufacturer editor from article master data.
- Keep German and English complete and preserve technical keys and custom labels.
- Keep Viewer read-only, Admin and Editor writable, and Messe without additional access.
- Keep the document overflow-free at 390 px; tab rows may scroll horizontally.
- Do not stage `.superpowers/`, `frontend/dist`, caches, local data, credentials, or generated output.

---

### Task 1: Normalize settings and data query state

**Files:**

- Create: `frontend/src/features/settings/settingsDataModel.ts`
- Create: `frontend/src/features/settings/settingsDataModel.test.ts`
- Modify: `frontend/src/features/settings/settingsModel.ts`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`

**Interfaces:**

- Produces: `DataGroup = "general" | "article"`.
- Produces: closed `GeneralDataType` and `ArticleDataType` unions.
- Produces: `readSettingsLocation(search: string): SettingsLocation`.
- Produces: `settingsLocationSearch(location: SettingsLocation): string`.
- Preserves: legacy `?tab=articleManagement` by normalizing it to Data/article/stock-unit.

- [x] **Step 1: Add failing query-model tests**

```ts
expect(readSettingsLocation("?tab=data")).toEqual({
  tab: "data", group: "general", type: "manufacturer"
});
expect(readSettingsLocation("?tab=articleManagement")).toEqual({
  tab: "data", group: "article", type: "stock_unit"
});
expect(readSettingsLocation("?tab=data&group=article&type=locations")).toEqual({
  tab: "data", group: "article", type: "locations"
});
expect(readSettingsLocation("?tab=data&group=general&type=locations").type).toBe("manufacturer");
```

- [x] **Step 2: Run RED**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/settingsDataModel.test.ts
```

Expected: FAIL because `settingsDataModel` does not exist.

- [x] **Step 3: Implement closed navigation types and compatibility**

```ts
export const generalDataTypes = [
  "manufacturer", "vehicle_category", "vehicle_gattung", "epoch", "gauge",
  "railway_company", "cv8_manufacturer", "symbols"
] as const;
export const articleDataTypes = ["stock_unit", "types", "customFields", "locations"] as const;

export type DataGroup = "general" | "article";
export type GeneralDataType = typeof generalDataTypes[number];
export type ArticleDataType = typeof articleDataTypes[number];
export type SettingsLocation = {
  tab: SettingsTab;
  group: DataGroup;
  type: GeneralDataType | ArticleDataType;
};
```

`readSettingsLocation` normalizes unknown group/type combinations to the group default. The removed
`articleManagement` value maps to `data/article/stock_unit`. `settingsLocationSearch` writes stable
`tab=data&group=...&type=...` parameters and omits `tab` for General.

- [x] **Step 4: Remove the visible Article Management tab and restore manufacturer**

```ts
export type SettingsTab = "general" | "data" | "digital" | "importExport" | "appearance" | "auth";

export const settingsTabs = [
  { id: "general", labelKey: "settings.tabs.general", descriptionKey: "settings.general.subtitle" },
  { id: "data", labelKey: "settings.tabs.data", descriptionKey: "settings.data.pageSubtitle" },
  { id: "digital", labelKey: "settings.tabs.digital", descriptionKey: "settings.digital.subtitle" },
  { id: "importExport", labelKey: "settings.tabs.importExport", descriptionKey: "importExport.subtitle" },
  { id: "appearance", labelKey: "settings.tabs.appearance", descriptionKey: "settings.appearance.subtitle" },
  { id: "auth", labelKey: "settings.tabs.auth", descriptionKey: "settings.auth.subtitle" }
] satisfies Array<{ id: SettingsTab; labelKey: string; descriptionKey: string }>;
```

Add `{ type: "manufacturer" }` as the first `masterDataTypes` item. Replace tests that asserted
manufacturer removal with manufacturer-presence and Article-Management-absence assertions.

- [x] **Step 5: Verify and commit Task 1**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/settingsDataModel.test.ts src/features/settings/SettingsView.test.tsx
git add src/features/settings/settingsDataModel.ts src/features/settings/settingsDataModel.test.ts src/features/settings/settingsModel.ts src/features/settings/SettingsView.test.tsx
git commit -m "refactor: define settings data navigation"
```

Expected: focused tests PASS.

---

### Task 2: Add one semantic tab-list owner

**Files:**

- Create: `frontend/src/features/settings/SettingsTabList.tsx`
- Create: `frontend/src/features/settings/SettingsTabList.test.tsx`
- Modify: `frontend/src/styles/settings.css`

**Interfaces:**

- Produces: `SettingsTabOption<T> = { id: T; label: string }`.
- Produces: `SettingsTabList<T>({ ariaLabel, options, value, onChange, className })`.
- Guarantees: one `tablist`, roving `tabIndex`, ArrowLeft/Right, Home/End, and selected-tab visibility.

- [x] **Step 1: Add failing semantics and keyboard tests**

```tsx
render(<SettingsTabList ariaLabel="Datengruppe" options={options}
  value="general" onChange={onChange} />);
const general = screen.getByRole("tab", { name: "Allgemeine Stammdaten" });
expect(general).toHaveAttribute("aria-selected", "true");
general.focus();
await user.keyboard("{ArrowRight}");
expect(onChange).toHaveBeenCalledWith("article");
expect(screen.getByRole("tab", { name: "Artikelstammdaten" })).toHaveFocus();
```

Also assert ArrowLeft wrapping, Home, End, exactly one `tabIndex=0`, and a guarded
`scrollIntoView({ block: "nearest", inline: "nearest" })` call.

- [x] **Step 2: Run RED**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/SettingsTabList.test.tsx
```

Expected: FAIL because `SettingsTabList` does not exist.

- [x] **Step 3: Implement the reusable component**

```tsx
export function SettingsTabList<T extends string>(props: SettingsTabListProps<T>) {
  const refs = useRef<Array<HTMLButtonElement | null>>([]);
  const selectAt = (index: number) => {
    props.onChange(props.options[index].id);
    refs.current[index]?.focus();
  };
  return <div className={`settings-secondary-tabs ${props.className || ""}`.trim()}
    role="tablist" aria-label={props.ariaLabel}>
    {props.options.map((option, index) => <button key={option.id} type="button" role="tab"
      ref={(element) => { refs.current[index] = element; }}
      aria-selected={option.id === props.value} tabIndex={option.id === props.value ? 0 : -1}
      className={option.id === props.value ? "active" : ""}
      onClick={() => props.onChange(option.id)}
      onKeyDown={(event) => handleTabKey(event, index, props.options.length, selectAt)}>
      {option.label}
    </button>)}
  </div>;
}
```

Add `settings-data-groups` for the compact grouped switch and `settings-data-types` for the
underlined horizontally scrollable row. Use existing tokens and no boxed data-tab cards.

- [x] **Step 4: Verify and commit Task 2**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/SettingsTabList.test.tsx
git add src/features/settings/SettingsTabList.tsx src/features/settings/SettingsTabList.test.tsx src/styles/settings.css
git commit -m "feat: add semantic settings data tabs"
```

Expected: focused tests PASS.

---

### Task 3: Integrate article master data under Data

**Files:**

- Modify: `frontend/src/features/settings/ArticleManagementSettings.tsx`
- Modify: `frontend/src/features/settings/ArticleManagementSettings.test.tsx`
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/settings.css`

**Interfaces:**

- Consumes: `SettingsTabList`, `DataGroup`, `GeneralDataType`, `ArticleDataType`, and query helpers.
- Produces: controlled article-data rendering with `activeSection` and `onSectionChange`.
- Retires: Article Management main-tab branch, component page heading, and generic manufacturer section.
- Preserves: manufacturer editor, roles, lazy loading, request guards, storage hierarchy, and localized labels.

- [x] **Step 1: Add failing integration tests**

```tsx
window.history.replaceState(null, "", "/settings?tab=data");
render(<SettingsView username="viewer" />);
expect(await screen.findByRole("heading", { level: 1, name: "Daten" })).toBeInTheDocument();
expect(screen.queryByRole("heading", { level: 2, name: "Daten" })).not.toBeInTheDocument();
expect(screen.getByRole("tab", { name: "Hersteller" })).toHaveAttribute("aria-selected", "true");
expect(screen.getByText("Herstellerseite")).toBeInTheDocument();
expect(screen.getByText("Suchdomains")).toBeInTheDocument();
```

Add assertions for exactly two group tabs, all four article-data tabs, no manufacturer tab inside
the Article group, stable query changes, legacy-link normalization, Viewer read-only behavior, and
Editor mutations. Keep the existing deferred-request regression when switching to Locations.

- [x] **Step 2: Run RED**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/SettingsView.test.tsx src/features/settings/ArticleManagementSettings.test.tsx
```

Expected: FAIL on the visible Article Management main tab, absent groups, missing manufacturer type,
and static/duplicated headings.

- [x] **Step 3: Make ArticleManagementSettings article-only and controlled**

```ts
export type ArticleDataSection = "stock_unit" | "types" | "customFields" | "locations";
export type ArticleManagementSettingsProps = {
  roles: string[];
  activeSection: ArticleDataSection;
  onSectionChange: (section: ArticleDataSection) => void;
};
```

Remove `manufacturers` and `manufacturer` from the component section/type definitions, loading,
rendering, and tests. Remove its page-level title and description. Render the read-only note,
controlled `SettingsTabList`, alerts, and active panel. Preserve canonical label persistence.

- [x] **Step 4: Compose groups and data tabs in SettingsView**

```tsx
<SettingsTabList ariaLabel={t("settings.data.groups")}
  options={dataGroupOptions} value={dataGroup} onChange={selectDataGroup}
  className="settings-data-groups" />

{dataGroup === "general" ? <>
  <SettingsTabList ariaLabel={t("settings.data.nav")}
    options={generalTypeOptions} value={activeGeneralType} onChange={selectGeneralType}
    className="settings-data-types" />
  {renderExistingGeneralMasterDataPanel()}
</> : <ArticleManagementSettings roles={currentSession?.roles || []}
  activeSection={activeArticleType} onSectionChange={selectArticleType} />}
```

Restore Manufacturer as the first general type. Leave its website, domains, aliases, scales,
warnings, sorting, search, source links, form mapping, and API calls unchanged. Remove the standalone
`activeSettingsTab === "articleManagement"` branch.

- [x] **Step 5: Add DE/EN group copy and remove only unused main-tab copy**

Add exact translations for `settings.data.pageSubtitle`, `settings.data.groups`,
`settings.data.group.general`, and `settings.data.group.article`. Keep article section labels used by
the embedded panel. Use `rg` before removing any old key.

- [x] **Step 6: Verify and commit Task 3**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/SettingsView.test.tsx src/features/settings/ArticleManagementSettings.test.tsx src/features/settings/SettingsTabList.test.tsx src/features/settings/settingsDataModel.test.ts
git add src/features/settings src/shared/i18n/de.ts src/shared/i18n/en.ts src/styles/settings.css
git commit -m "feat: consolidate settings master data"
```

Expected: focused tests PASS.

---

### Task 4: Dynamic page header and complete verification

**Files:**

- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/settings.css`
- Modify: `docs/superpowers/specs/2026-08-08-einstellungen-daten-navigation-design.md`
- Modify: `docs/superpowers/plans/2026-08-08-einstellungen-daten-navigation.md`

**Interfaces:**

- Consumes: `settingsTabs[*].labelKey` and `descriptionKey`.
- Produces: exactly one level-one heading and one active-tab description.
- Retires: fixed Settings title/subtitle, title-adjacent version, and duplicate page headings.

- [x] **Step 1: Add failing header tests**

```tsx
for (const [tab, title, description] of [
  ["general", "Allgemein", "Sprache, Startseite, Datumsformat und Druckausgabe."],
  ["data", "Daten", "Stammdaten für Fahrzeuge, Artikel und Anlagen zentral pflegen."],
  ["appearance", "Darstellung", "Erscheinungsbild und Anzeigeoptionen festlegen."]
] as const) {
  window.history.replaceState(null, "", tab === "general" ? "/settings" : `/settings?tab=${tab}`);
  const view = render(<SettingsView username="viewer" />);
  expect(await screen.findByRole("heading", { level: 1, name: title })).toBeInTheDocument();
  expect(screen.getByText(description)).toBeInTheDocument();
  view.unmount();
}
```

Assert that no version is rendered beside the heading and no active-tab title/description is
duplicated inside its content. Preserve genuine subsection headings.

- [x] **Step 2: Run RED**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/SettingsView.test.tsx
```

Expected: FAIL because the fixed Settings title and adjacent version still render.

- [x] **Step 3: Derive the page header from the active tab**

```tsx
const activeDefinition = settingsTabs.find((tab) => tab.id === activeSettingsTab) ?? settingsTabs[0];

<section className="settings-head">
  <h1>{t(activeDefinition.labelKey)}</h1>
  <p>{t(activeDefinition.descriptionKey)}</p>
</section>
```

Remove the adjacent version and only duplicated page-level heading/description blocks. Keep
operational headings such as Inventory numbers, Manufacturer database, Backup, and Users.

- [x] **Step 4: Run automated gates**

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/SettingsView.test.tsx src/features/settings/ArticleManagementSettings.test.tsx src/features/settings/SettingsTabList.test.tsx src/features/settings/settingsDataModel.test.ts
npm.cmd run test:run
npm.cmd run build
```

Expected: focused PASS, complete Vitest PASS, and `tsc -b && vite build` PASS.

- [x] **Step 5: Run browser gates**

Verify German/English, light/dark, desktop and 390 × 844 px, all six headers, group/type URL state,
complete manufacturer columns and edit fields, article tabs without Manufacturer, keyboard behavior,
selected-tab visibility, zero new console warnings/errors, and no document overflow.

- [x] **Step 6: Close documentation and commit**

Set the design status to `Implemented and verified`, mark this plan complete, then run:

```powershell
git diff --check
git status --short
```

Stage only scoped source, tests, i18n, CSS, spec, and plan files. Commit with:

```powershell
git commit -m "fix: restore settings manufacturer data"
```

## Final Delivery Gate

- [x] Complete manufacturer editing is reachable under Data and uses the existing records.
- [x] No generic manufacturer editor or Article Management main tab remains.
- [x] All article master data is reachable under Data → Article master data.
- [x] The active main tab supplies the only page title and description.
- [x] Legacy links normalize without errors or blank content.
- [x] German, English, roles, errors, loading, themes, mobile, and keyboard behavior pass.
- [x] Full tests, build, browser console, overflow check, and `git diff --check` pass.
- [x] `.superpowers/`, generated output, caches, local data, and credentials remain uncommitted.
