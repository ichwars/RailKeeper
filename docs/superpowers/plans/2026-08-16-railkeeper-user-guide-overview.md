# RailKeeper User Guide Overview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a complete English and German guide for the dashboard overview, its metrics,
data-quality indicators, maintenance signals, shortcuts, and local layout controls in stable
RailKeeper v0.1.17.6.

**Architecture:** Add one paired user-guide page at the paths already reserved by the coverage
matrix. Link the pair from both user-guide sidebars and landing pages, then change only the
`overview` coverage topic from `planned` to `documented`. The pages explain how the browser derives
dashboard values from vehicle data and distinguish dashboard indicators from the filtered vehicle
lists opened by the data-gap links.

**Tech Stack:** VitePress 2, Markdown with validated frontmatter, Node.js documentation checks,
React/TypeScript sources as the stable behavior contract.

## Global Constraints

- English remains the public root locale; German uses the `/de/` prefix.
- English and German pages use identical relative paths below `docs/site/` and `docs/site/de/`.
- User content documents stable RailKeeper `0.1.17.6`, not unpublished `main` behavior.
- Both pages use `audience: user`, `status: stable`, `reviewedVersion: 0.1.17.6`, and
  `lastReviewed: 2026-08-16`.
- The two language versions must be semantically equivalent, not literal machine translations.
- State that Admin, Editor, Viewer, and Planner users can open the overview. Messe-only users start
  in Exhibition and cannot open it.
- Do not describe the overview as a server-side report. Its values are calculated in the browser
  from the vehicles returned for the current session.
- Do not claim that the overview has a general search or filter bar. Only the four data-gap links
  open the vehicle inventory with the matching filter active.
- Explain percentage denominators precisely. Decoder coverage uses all vehicles, while the
  `Digital without decoder no.` gap counts only digital vehicles without either decoder-number
  field.
- Explain that `Fully documented` requires an article number, EAN, and at least one image. It does
  not certify every vehicle field as complete.
- Explain that widget order and hidden state are local to the current browser storage. The reset
  control appears after at least one widget is hidden and restores the default order and all
  widgets.
- Do not include private inventory data, screenshots with real records, or generated build output.
- Keep `docs/.vitepress/dist` and `docs/node_modules` out of Git.

---

### Task 1: Publish the paired overview and data-quality guide

**Files:**

- Create: `docs/site/guide/overview/index.md`
- Create: `docs/site/de/guide/overview/index.md`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/coverage.json`

**Interfaces:**

- Consumes: the `overview` coverage topic, `OverviewView`, application role routing, vehicle gap
  filters, and stable version metadata.
- Produces: `/guide/overview/`, `/de/guide/overview/`, paired sidebar and landing-page entries, and
  a `documented` coverage status for `overview`.

- [ ] **Step 1: Make the coverage contract fail for the missing pages**

Change only the `overview` topic in `docs/coverage.json` from `planned` to `documented`.

- [ ] **Step 2: Verify the contract rejects the missing destinations**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: failure naming the missing English and German `overview` pages.

- [ ] **Step 3: Write the English page**

Create `docs/site/guide/overview/index.md` with required frontmatter and these content requirements:

1. Title `Overview, metrics, and data quality`, scope, supported roles, and stable version.
2. `Open and refresh the overview`:
   - use **Overview** in the sidebar;
   - the refresh icon reloads the vehicle data and disables itself while loading;
   - a load failure appears above the dashboard and does not silently replace the data.
3. `Read the four summary metrics` with a table defining:
   - total inventory and its distinct top-five category/gauge count;
   - digitalization as digital vehicles divided by all vehicles;
   - recorded list value as the sum of parseable maintained list-price strings in euros, rounded
     to whole euros for display;
   - maintenance as overdue or due-today entries, with separate next-30-days and all-open counts.
4. `Use the dashboard widgets` with exact coverage of:
   - Inventory mix: five most common categories;
   - Data quality: image, decoder-number, article-number, EAN, and fully-documented percentages;
   - Action needed: the four data gaps and their filtered `/vehicles?gap=...` destinations;
   - Manufacturers: five most common manufacturers;
   - Quick actions: Vehicles, Import/Export, and Settings/master-data destinations;
   - Maintenance radar: four closest incomplete due-dated entries, overdue/today/upcoming labels,
     completed total, all maintenance costs, and distinct condition ratings;
   - Next value: deterministic recommendation priority for empty inventory, image coverage below
     70 percent, due maintenance, then the stable-inventory suggestion.
5. `Understand the data-quality percentages`:
   - denominator is the total vehicle count for every percentage;
   - an image means at least one vehicle image;
   - decoder number accepts either decoder-number field and is not limited to digital vehicles;
   - fully documented means article number + EAN + at least one image;
   - zero vehicles produce zero percent rather than an undefined value.
6. `Work from data gaps`:
   - select a gap to open Vehicles with exactly that gap filter;
   - list the four gap meanings;
   - clarify that editing happens in Vehicles, not on the dashboard;
   - clarify that the overview itself has no general search/filter bar.
7. `Arrange the dashboard`:
   - up/down controls move each widget in the order;
   - hide removes a widget;
   - reset restores all seven widgets and default order;
   - preferences are stored in current browser local storage, not as shared account settings.
8. `Empty and exceptional states` covering no vehicles, no maintenance with due date, no major
   data gaps, all widgets hidden, and vehicle-loading errors.
9. `Related pages` linking only to `/guide/`, `/guide/getting-started/`, and `/administration/` so
   no link points to an unpublished guide page.
10. `Documented RailKeeper version` stating stable `v0.1.17.6` and review date 2026-08-16.

Use English UI labels exactly as shown by the stable interface.

- [ ] **Step 4: Write the semantically equivalent German page**

Create `docs/site/de/guide/overview/index.md` with matching frontmatter, facts, formulas, caveats,
states, and destinations. Use the title `Übersicht, Kennzahlen und Datenqualität` and German UI
labels exactly as shown by the stable interface.

- [ ] **Step 5: Add both pages to the guide sidebars**

In `docs/.vitepress/config.mts`, preserve the landing-page and getting-started items, then add:

```ts
{ text: "Dashboard and data quality", link: "/guide/overview/" }
```

```ts
{ text: "Dashboard und Datenqualität", link: "/de/guide/overview/" }
```

- [ ] **Step 6: Add the chapter to both guide landing pages**

Add a short next-step paragraph to `docs/site/guide/index.md` linking to `/guide/overview/` and the
equivalent German paragraph linking to `/de/guide/overview/`. Keep both landing pages semantically
equivalent and set `lastReviewed` to `2026-08-16`.

- [ ] **Step 7: Run the complete documentation check**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: 19 tests pass, coverage validation produces no errors, and VitePress builds both routes
without dead links.

- [ ] **Step 8: Review the language pair and source fidelity**

Check the pair against:

```text
frontend/src/features/overview/OverviewView.tsx
frontend/src/features/vehicles/useVehicleInventoryController.ts
frontend/src/app/App.tsx
frontend/src/app/Shell.tsx
frontend/src/shared/i18n/en.ts
frontend/src/shared/i18n/de.ts
```

Expected: both pages describe the same stable behavior, every formula and threshold matches the
source, and no general dashboard search or filter is claimed.

- [ ] **Step 9: Commit the content package**

```powershell
git add docs/site/guide docs/site/de/guide docs/.vitepress/config.mts docs/coverage.json
git commit -m "docs: add overview user guide"
```
