# RailKeeper Vehicle Inventory User Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish complete English and German user documentation for the stable vehicle inventory,
core vehicle records, reports, QR labels, and destructive record handling in RailKeeper v0.1.17.6.

**Architecture:** Add one paired chapter at the paths reserved by the coverage matrix. The chapter
documents the stable release tag as its behavior contract because current `main` has additional
inventory columns and filters. It covers the inventory workspace and every core form field while
leaving media, maintenance, decoder/CV, article search, spare parts, and detailed exhibition work
for their dedicated chapters.

**Tech Stack:** VitePress 2, validated Markdown frontmatter, Node.js documentation checks,
React/TypeScript and Go sources from tag `v0.1.17.6` as the behavior contract.

## Global Constraints

- English remains the public root locale; German uses the `/de/` prefix.
- English and German pages use identical relative paths below `docs/site/` and `docs/site/de/`.
- Both pages use `audience: user`, `status: stable`, `reviewedVersion: 0.1.17.6`, and
  `lastReviewed: 2026-08-16`.
- The language pair must be semantically equivalent and use the stable UI labels for each locale.
- Describe tag `v0.1.17.6`, not the additional table-column, mobile-card, railway-company, epoch, or
  adapter filter behavior now present on `main`.
- Admin, Editor, Viewer, and Planner can inspect the vehicle inventory. Vehicle create, update, and
  delete APIs require Admin or Editor. In v0.1.17.6, write controls can still remain visible to
  Viewer and Planner users, but the server rejects unauthorized writes.
- The inventory search sends the current text to the server and searches only inventory number,
  manufacturer, article number, and designation. Client filters then narrow the returned rows.
- State that filters combine. The stable filter set is inventory state, due-maintenance state,
  manufacturer, category, subtype, exhibition readiness, and an optional quality-gap filter opened
  from Overview.
- Explain that stable **Without maintenance** means no due maintenance entry, not no maintenance
  history.
- The desktop view-mode preference is local to the current browser. Do not claim account sync.
- The report's **All** scope means all currently searched and filtered vehicles in their current
  sort order. **Selection** includes only selected vehicles that remain visible.
- A blank inventory number on create requires an active inventory-number scheme for the selected
  category. A manual inventory number must be unique. A blank number on update retains the existing
  value.
- Manufacturer, designation, gauge, category, and subtype are required. Surrounding whitespace is
  removed when the server saves the record.
- Changing category clears a subtype that is not permitted by the configured category-subtype
  relations. **Coupling (F=R)** keeps the rear value synchronized to the front value.
- Static select values are German strings in both locales in v0.1.17.6. Document that release
  behavior instead of translating the stored choices.
- Deleting a vehicle is permanent and cascades through vehicle-owned database records. References
  from layout/accessory allocation data can block deletion. Do not claim physical file cleanup.
- Do not expand media, maintenance, decoder/CV, article-data search, spare-parts, or exhibition
  administration beyond the entry points and boundaries needed to understand the core inventory.
- Do not link to planned pages that do not exist yet.
- Keep `docs/.vitepress/dist`, `docs/node_modules`, private inventory data, and generated reports out
  of Git.

---

### Task 1: Publish the paired vehicle inventory and core-record chapter

**Files:**

- Create: `docs/site/guide/vehicles/index.md`
- Create: `docs/site/de/guide/vehicles/index.md`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/coverage.json`

**Interfaces:**

- Consumes: the `vehicles-core` coverage topic, stable vehicle list and mutation routes,
  `VehiclesView`, `VehicleInventoryPanel`, core form components, output controller, and stable i18n.
- Produces: `/guide/vehicles/`, `/de/guide/vehicles/`, paired navigation and landing-page links, and
  `documented` coverage for `vehicles-core`.

- [ ] **Step 1: Make the coverage contract fail for the missing pages**

Change only the `vehicles-core` topic in `docs/coverage.json` from `planned` to `documented`.

- [ ] **Step 2: Verify the missing paired pages are rejected**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: 19 tests pass, then validation fails only for missing
`guide/vehicles/index.md` and `de/guide/vehicles/index.md`.

- [ ] **Step 3: Write the English chapter**

Create `docs/site/guide/vehicles/index.md` with the required frontmatter and these sections:

1. `Vehicle inventory and core records`: scope, stable version, role access, and the v0.1.17.6
   visible-write-control limitation for Viewer and Planner.
2. `Read the inventory status`: define Total inventory, Digitalization, Maintenance, Next
   appointment, and Image care; explain which cards activate filters or open the vehicle view.
3. `Search the inventory`: exact four server-side search fields, immediate reload behavior, AND
   combination with client filters, and load-error behavior.
4. `Filter the inventory`: exact stable inventory, maintenance, master-data, exhibition-ready, and
   Overview quality-gap filters; category-subtype dependency; **Without maintenance** caveat;
   **Clear filters** behavior and result count.
5. `Choose a view and sort`: desktop table/card, compact mobile list, browser-local view mode,
   table fields, sortable fields, initial natural inventory-number sort, and repeated-header sort
   direction.
6. `Select vehicles`: table checkboxes, select-all-visible behavior, hidden selection caveat, and
   report selection using only visible selected rows.
7. `Open and inspect a vehicle`: name/image/view actions, read-only sections, omission of empty
   fields, quick-menu destinations, and the role boundary for Edit.
8. `Create a vehicle`: Admin/Editor workflow, required fields, automatic or manual inventory
   number, inventory-number errors, creation switching to edit mode, and later tabs becoming
   available only after the base record exists.
9. `Core field reference` with compact tables that name every stable core field and explain:
   - Model/product identity: Inventory number, Article no., Manufacturer, Gauge, Designation,
     Railway company, Epoch, Category, Subtype, Description, Series, Vehicle no., EAN, Production
     period, List price, Digital, both decoder switches/numbers, Decoder type, Exhibition ready,
     Exhibition, and ABC brakes.
   - Details: Length, Weight, Color, Lettering, Load, Interior, Axles, Axle count, Traction tire
     count, Wheelset, Power pickup, Adapter/interface, Coupling front/rear/same, Drive, Headlights,
     Lighting, Sound generator, Smoke generator, and QR-code toggle.
   - Vehicle/ownership: Acquisition, source, Price, Date, Location, location details, Condition,
     condition details, Packaging, and Additional information.
   - Explain switches that enable adjacent description fields and that the QR toggle controls the
     edit-form QR button.
10. `Stable select choices`: list exact German stored choices for wheelset, coupling, power pickup,
    adapter/interface, acquisition, source, storage, condition, and packaging; explain that
    manufacturer/gauge/epoch/railway-company/category/subtype come from configured master data.
11. `Edit a vehicle`: load full detail, required validation, inventory-number uniqueness and history,
    category/subtype reset, coupling synchronization, save/reload behavior, and server-side
    permission enforcement.
12. `Create QR labels`: identity requirement, exact text payload (inventory number, designation,
    optional decoder number), PNG/SVG download, print, and distinction between form toggle and
    quick-menu/read-view access.
13. `Create inventory reports`: Overview list versus Detail list, title, current filtered/sorted
    scope versus visible selection, QR/image options, individual detail report, browser print/PDF,
    and popup/empty-result failure guidance.
14. `Exhibition switch boundary`: eligibility uses Digital plus primary digital decoder number,
    enabling requires an unlocked list and Admin or combined Editor+Messe permissions, duplicate
    checks, and disabling only resets the vehicle flag. Keep the full exhibition workflow deferred.
15. `Delete a vehicle`: confirmation, permanent cascade, possible allocation-reference block,
    audit log, backup recommendation, and Admin/Editor requirement.
16. `Empty, loading, and error states`: initial load, empty inventory, empty filter result, failed
    list/detail/save/delete/report, missing inventory scheme, duplicate number, and authorization.
17. `Related pages`: link only to the published guide landing, Getting Started, Overview, and
    Administration landing pages.
18. `Documented RailKeeper version`: stable v0.1.17.6 and review date 2026-08-16.

- [ ] **Step 4: Write the semantically equivalent German chapter**

Create `docs/site/de/guide/vehicles/index.md` with the same facts, caveats, field coverage, stored
choice values, permission boundaries, destructive effects, and related destinations. Use the title
`Fahrzeugbestand und Grunddaten` and exact stable German UI labels.

- [ ] **Step 5: Add the paired sidebar entries**

Add the English entry after the dashboard chapter:

```ts
{ text: "Vehicle inventory", link: "/guide/vehicles/" }
```

Add the German counterpart:

```ts
{ text: "Fahrzeugbestand", link: "/de/guide/vehicles/" }
```

- [ ] **Step 6: Add the chapter to both guide landing pages**

Append one semantically paired paragraph linking to the new chapter and describing search, filters,
record creation/editing, reports, QR labels, and deletion. Keep `lastReviewed: 2026-08-16`.

- [ ] **Step 7: Run the complete documentation check**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: 19 tests pass, coverage validation succeeds, and VitePress builds every locale route.

- [ ] **Step 8: Verify stable-source fidelity and language parity**

Compare both pages line by line against the `v0.1.17.6` versions of:

```text
frontend/src/features/vehicles/VehicleInventoryPanel.tsx
frontend/src/features/vehicles/useVehicleInventoryController.ts
frontend/src/features/vehicles/VehicleModelTab.tsx
frontend/src/features/vehicles/VehicleFormFields.tsx
frontend/src/features/vehicles/VehicleReadOnlyView.tsx
frontend/src/features/vehicles/useVehicleOutputController.ts
frontend/src/features/vehicles/vehicleOptions.ts
frontend/src/features/vehicles/vehicleQr.ts
frontend/src/features/vehicles/vehicleMutationCommands.ts
backend/internal/api/routes.go
backend/internal/api/vehicle_handlers.go
backend/internal/application/vehicle_mutation.go
backend/internal/application/vehicle_repository.go
```

Expected: every stable field is named, EN and DE are semantically equivalent, role and destructive
effects are explicit, and no post-v0.1.17.6 inventory filters or columns are documented.

- [ ] **Step 9: Commit the content package**

```powershell
git add docs/site/guide docs/site/de/guide docs/.vitepress/config.mts docs/coverage.json
git commit -m "docs: add vehicle inventory user guide"
```
