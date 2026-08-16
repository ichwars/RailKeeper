# RailKeeper Accessories Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `aegis:executing-plans` to implement this plan
> task by task. Keep all Git mutations with the coordinating agent.

**Goal:** Publish a complete four-page English and German user-guide section for the stable
RailKeeper accessories workspace.

**Architecture:** Reuse the existing `accessories` coverage owner and VitePress user-guide
structure. Add one bilingual overview pair as the coverage destination plus three mirrored
subpage pairs for records, stock, and allocations. Keep every behavior claim pinned to tag
`v0.1.17.6` and link only to already published documentation owners.

**Tech Stack:** Markdown, JSON, TypeScript VitePress configuration, Node.js documentation tests,
VitePress 2.0.0-alpha.19, Git, GitHub Actions, GitHub Pages.

**Baseline/Authority Refs:**

- `AGENTS.md`;
- `docs/superpowers/specs/2026-08-16-railkeeper-accessories-guide-design.md`;
- `docs/superpowers/plans/2026-08-15-railkeeper-documentation-foundation.md`;
- `docs/coverage.json`;
- tag `v0.1.17.6` and the stable source files named under each task;
- existing English/German guide pages and `docs/.vitepress/config.mts`.

**Compatibility Boundary:** Documentation only. Do not change runtime behavior, frontend or
backend source, API contracts, schemas, migrations, dependencies, screenshots, generated output,
or coverage ownership outside the existing `accessories` topic. Later `main` behavior is excluded.

**TDD Route:**

- Mode: off
- Decision: skipped
- Strict authority: not applicable
- Test posture: post-change documentation regression and stable-source audit
- Reason: the approved change creates documentation and navigation, not executable product logic
- Verification: intentional coverage failure, focused validation, full documentation check,
  semantic parity review, source-fidelity review, pull-request checks, and live-page checks

**Verification:** `node scripts/validate-docs.mjs`, `npm.cmd run check`, `git diff --check`, focused
metadata/link/line-length scans, stable-tag source audit, pull-request checks, and post-merge HTTP
checks for every English and German page.

## Plan basis and readiness

**Aegis Visibility:** Planning is useful because eight new stable pages must remain semantically
paired while describing permissions, stock arithmetic, immediate writes, and recovery without
leaking unpublished layout or later-development behavior.

**BaselineUsageDraft:**

- Required baseline refs: approved design, documentation foundation, coverage contract,
  `v0.1.17.6`, current VitePress configuration
- Delivered context refs: repository `AGENTS.md` and current worktree state
- Acknowledged before plan refs: all required refs above
- Cited in plan refs: all required refs above
- Missing refs: none
- Decision: continue

**Requirement Ready Check:**

- Requirement source refs: approved accessories-guide design
- Goals and scope refs: design Objective, Scope, Non-goals, Information architecture
- User/scenario refs: overview, article record, stock/document, and allocation lifecycles
- Requirement item refs: roles, trust/persistence, errors, coverage, navigation, verification
- Acceptance/verification criteria refs: design Verification and Expected outcome
- Open blocker questions: none
- Decision: ready

**Change Necessity:**

- User-visible need: complete bilingual accessories documentation that users can navigate
- No-change/non-code option: insufficient because the stable workflows are absent from the site
- Why code change is necessary: no product-code change is necessary
- Minimum change boundary: Markdown pages, coverage JSON, VitePress sidebar, and published links
- Decision: docs/config-only

**Existence Check:**

- Proposed new surface: four English and four German pages under `guide/accessories`
- Existing owner/reuse candidate: existing `accessories` topic and VitePress user-guide section
- Why existing surface is insufficient: its configured destination pair does not exist
- Creation proof: approved four-page information architecture and current `planned` coverage state
- Entropy/retirement impact: one bounded topic group, no duplicate runtime or documentation owner
- Decision: add-with-proof

**Architecture Integrity Lens:**

- Invariant: one coverage owner for all stable accessories user workflows
- Canonical owner/contract: `accessories` in `docs/coverage.json`
- Responsibility overlap: vehicle article search stays with `vehicle-search-spares`; layout design
  stays `not-published`; master-data administration remains planned
- Higher-level simplification: use one overview hub instead of one oversized page or UI-tab pages
- Retirement/falsifier: no temporary fallback or duplicate page path is introduced
- Verdict: proceed

**Plan Pressure Test:**

- Owner/contract/retirement: one existing owner, no retirement track required
- Architecture integrity/higher-level path: four task-oriented pages fit the approved structure
- Verification scope: stable source, bilingual parity, coverage, build, PR, and live deployment
- Task executability: each page pair and navigation slice has exact files, headings, and checks
- Pressure result: proceed

**Complexity Budget:**

- Artifact class: maintained user documentation
- Target files/artifacts: eight new pages, coverage JSON, VitePress config, four existing pages
- Current pressure: accessories behavior is broad and unsuitable for one page
- Projected post-change pressure: four focused pages with mirrored sections
- Budget result: within-budget
- Planned governance: keep each page task-focused and avoid repeating full field/lifecycle tables

**Plan-Time Complexity Check:**

- Target files: new accessory pages plus narrow navigation and related-link owners
- Existing size/shape signals: existing vehicle chapters are long but focused; config is centralized
- Owner fit: new content belongs under the existing `accessories` topic
- Add-in-place risk: one-page design would be overlong; broad edits to unrelated pages would drift
- Better file boundary: overview, records, stock/documents, and allocations/history page pairs
- Recommendation: add owner files and keep existing-file edits wiring-only

## Execution Readiness View

- Intent Lock: document every stable user-facing accessories workflow in English and German
- Scope Fence: documentation and navigation only, no product or API changes
- Baseline Lock: behavior claims must resolve to tag `v0.1.17.6`
- Approved Behavior: four-page task-oriented structure and role/safety model from the design
- Owner/Contract Constraints: promote only `accessories`; preserve all other coverage owners
- Compatibility Boundary: do not publish layout operation, planned administration, or later `main`
- Retirement Boundary: no old path, adapter, fallback, or temporary compatibility layer
- Task Batches: overview, records, stock/documents, allocations/history, navigation, closeout
- Test Obligations: intentional gate failure, pair parity, coverage validation, VitePress build
- Review Gates: source fidelity, language parity, permissions, persistence, partial writes, recovery
- Drift/Rewind Rules: stop and return to the design if stable behavior requires another owner or
  the four-page boundary no longer fits
- Evidence Required Before Completion: clean diff, full docs check, reviewed head SHA, green PR
  checks, merged PR, and successful live English/German routes
- Advisory Boundary: method-pack execution guidance only, not authoritative completion

## Global constraints

- Work only in `.worktrees/docs-user-guide-accessories` on
  `dev/docs-user-guide-accessories`. Do not modify or push local `main`.
- Preserve the user's untracked files in the main checkout.
- Use stable release `v0.1.17.6` and review date `2026-08-16` on every new page.
- Keep English and German pages in the same semantic section order.
- Use UI labels from the matching stable locale. Preserve identifiers and enum values only where
  they help users interpret stored behavior.
- State whether every operation changes a dialog draft, writes immediately, or only reads data.
- State confirmations, partial-write behavior, refresh behavior, and recovery precisely.
- External search values and URLs are suggestions, not authoritative product data.
- Do not document the layout workspace as published. Allocation targets may be named only where
  the stable accessories dialog exposes them.
- Link only to published pages. Do not activate planned master-data, Settings, backup, layout, or
  administration destinations.
- Preserve LF, UTF-8, approximately 100 to 120 character lines, and established documentation
  style. Do not use em dashes or unfinished-content markers.
- Do not commit `docs/node_modules`, `docs/.vitepress/dist`, caches, or other generated output.
- Make one scoped commit after each coherent task passes its verification.
- Merge only after the exact reviewed head has all required GitHub checks green.

## File map

| File | Responsibility |
| --- | --- |
| `docs/coverage.json` | Promote the existing accessories topic. |
| `docs/site/guide/accessories/index.md` | English overview and catalogue workflow. |
| `docs/site/de/guide/accessories/index.md` | German overview counterpart. |
| `docs/site/guide/accessories/article-records.md` | English record and technical-data workflow. |
| `docs/site/de/guide/accessories/article-records.md` | German record counterpart. |
| `docs/site/guide/accessories/stock-purchases-documents.md` | English inventory and document workflow. |
| `docs/site/de/guide/accessories/stock-purchases-documents.md` | German inventory counterpart. |
| `docs/site/guide/accessories/allocations-history.md` | English reservation and installation workflow. |
| `docs/site/de/guide/accessories/allocations-history.md` | German allocation counterpart. |
| `docs/.vitepress/config.mts` | Add mirrored Accessories/Zubehör sidebar groups. |
| `docs/site/guide/index.md` | Add the English guide transition. |
| `docs/site/de/guide/index.md` | Add the German guide transition. |
| `docs/site/guide/overview/index.md` | Add the published accessories related link. |
| `docs/site/de/guide/overview/index.md` | Add the German related link. |
| `docs/site/guide/vehicles/search-and-spares.md` | Clarify the accessory-search boundary. |
| `docs/site/de/guide/vehicles/search-and-spares.md` | German boundary clarification. |

---

### Task 1: Publish the coverage destination and accessories overview

**Files:**

- Modify: `docs/coverage.json`
- Create: `docs/site/guide/accessories/index.md`
- Create: `docs/site/de/guide/accessories/index.md`
- Reference: approved design and stable overview sources

**Why:** Users need a reliable entry point that explains the article catalogue before opening
write-heavy detail workflows.

**Change Necessity:** The destination pair does not exist. The minimum boundary is the existing
coverage status plus its exact English/German overview paths.

**Impact/Compatibility:** No ownership changes. `accessories` becomes `documented` only after both
configured paths exist and pass validation.

**Verification:** Intentional `node scripts/validate-docs.mjs` failure before page creation, then a
successful focused validation, `git diff --check`, and clean task-owned status before commit.

**Stable sources:**

- `frontend/src/features/accessories/AccessoriesView.tsx`;
- `ArticleMetrics.tsx`, `ArticleToolbar.tsx`, `ArticleTable.tsx`, `ArticleCardGrid.tsx`,
  `ArticleCompactList.tsx`, `ArticleActions.tsx`;
- `articleTableColumns.ts`, `articleViewMode.ts`, `useArticleOverview.ts`;
- `backend/internal/application/accessory_overview.go`;
- stable overview tests, route access, and `accessories.*` translations.

- [ ] **Step 1: Audit the stable overview before writing claims**

Run from the worktree root:

```powershell
git show v0.1.17.6:frontend/src/features/accessories/AccessoriesView.tsx
git show v0.1.17.6:frontend/src/features/accessories/ArticleMetrics.tsx
git show v0.1.17.6:frontend/src/features/accessories/ArticleToolbar.tsx
git show v0.1.17.6:frontend/src/features/accessories/ArticleTable.tsx
git show v0.1.17.6:frontend/src/features/accessories/useArticleOverview.ts
git show v0.1.17.6:backend/internal/application/accessory_overview.go
```

Confirm these stable facts before prose is added:

- metrics cover article/type counts, available quantity/location count, allocated quantity split
  into reserved/installed, and non-clickable care hints;
- article, available, and allocated metric cards apply the stable reset/status filters;
- server query filters are text, article type, manufacturer, gauge, status, and storage location;
- `allocated` expands to `reserved` plus `installed`;
- initial sort is inventory number ascending, and selecting an active sortable column reverses it;
- table and card preference and visible table columns use browser local storage;
- image, inventory number, manufacturer, article number, name, type, gauge, stock, and storage are
  the configurable columns; inventory number and name cannot both be hidden;
- compact width always uses the mobile list;
- table selection has no bulk command in `v0.1.17.6`;
- Viewer is read-only, Planner can later reserve, Editor can write/archive/restore, and only Admin
  receives permanent delete;
- an earlier result can remain visible with an error, so users must reload before trusting it.

- [ ] **Step 2: Turn the coverage topic into an intentional failing gate**

Change only the `accessories` topic status:

```json
"status": "documented"
```

Run from `docs`:

```powershell
node scripts/validate-docs.mjs
```

Expected: failure names both missing `guide/accessories/index.md` paths. Any ownership, schema, or
unrelated failure is unexpected and must be resolved before continuing.

- [ ] **Step 3: Create the English overview page**

Use this exact frontmatter and heading order:

```markdown
---
title: Accessories overview
description: Find, filter, inspect, and safely manage accessory articles.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Accessories overview

## Access rights and workspace model

## Read the overview metrics

## Search and filter articles

## Choose columns, view, and sorting

## Select, open, and manage an article

## Resolve loading, empty, and error states

## Continue with an article workflow

## Related pages

## Documented RailKeeper version
```

Include every confirmed overview fact from Step 1. Explain all stable filters and status values,
every configurable column, local-browser persistence, compact behavior, per-row actions, archive
and restore, Admin-only permanent deletion, and the inert selection boundary. Do not duplicate the
field, stock, or allocation instructions from later pages.

- [ ] **Step 4: Create the German semantic counterpart**

Use matching structure and this frontmatter:

```markdown
---
title: Zubehörübersicht
description: Zubehörartikel finden, filtern, prüfen und sicher verwalten.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Zubehörübersicht
```

Translate the stable UI meaning, not the English sentence shape. Use the same tables, warnings,
role boundaries, persistence statements, error states, and related destinations.

- [ ] **Step 5: Verify and commit Task 1**

Run:

```powershell
cd docs
node scripts/validate-docs.mjs
cd ..
git diff --check
git status --short
```

Expected: validation succeeds, whitespace check is clean, and only the three Task 1 files plus the
already committed design/plan history belong to the branch. Commit only Task 1 files:

```powershell
git add docs/coverage.json docs/site/guide/accessories/index.md `
  docs/site/de/guide/accessories/index.md
git commit -m "docs: add accessories overview guide"
```

---

### Task 2: Document article records and technical data

**Files:**

- Create: `docs/site/guide/accessories/article-records.md`
- Create: `docs/site/de/guide/accessories/article-records.md`

**Why:** Users need one complete path for product identity, validation, type-specific data, search,
duplicate handling, and lifecycle actions before they change stock.

**Change Necessity:** The approved overview is intentionally a hub. The focused pair prevents core
record rules from being buried inside stock or allocation instructions.

**Impact/Compatibility:** This pair stays under the existing accessories owner. It may name
configured master data but must not document its planned administration page.

**Verification:** Focused documentation validation, mirrored heading scan, `git diff --check`, and
stable-source readback for every field, role, confirmation, and lifecycle claim.

**Stable sources:**

- `ArticleEditorDialog.tsx`, `ArticleCoreTab.tsx`, `ArticleSubjectTab.tsx`;
- `articleEditorModel.ts`, `articleEditorAutomation.ts`, `articleTypes.ts`, `articleSubtypes.ts`,
  `articleTypeFields.ts`;
- `accessoryArticleSearch.ts`, `useAccessoryArticleSearchController.ts`;
- shared barcode/article-search dialogs and stable translations;
- `useArticleEditorController.ts`, application/accessory validation, duplicate, archive, restore,
  and deletion services plus route-access tests.

- [ ] **Step 1: Audit stable record behavior and field contracts**

Read the files above from tag `v0.1.17.6` with `git show`. Build the page only after confirming:

- required fields are manufacturer, name, subtype, stock unit, and valid positive package quantity;
- minimum stock is a non-negative whole number;
- article type, subtype, gauges, and stock units are configured master data with historical-value
  preservation where stable code permits it;
- common optional fields include article number, EAN, manufacturer status, scale, description,
  URLs, alternative numbers, keywords, compatibility notes, and internal notes;
- inventory strategies are quantity, individual, and quantity-later-individual;
- the exact stable article types and their supported type-specific fields/options come from
  `articleTypeFields.ts` and the stable master-data entries;
- scale and keywords are auto-managed only until the user manually changes them;
- changing article type can discard subtype and incompatible technical values and requires the
  stable confirmation;
- create mode cannot persist stock or related resources before the article exists;
- duplicate check uses manufacturer/article number and requires deliberate confirmation to create
  a candidate duplicate;
- external accessory search and barcode results are untrusted draft suggestions;
- close with dirty article or subdraft state requires confirmation;
- archive/restore require Editor; permanent delete requires Admin and can be blocked by references.

- [ ] **Step 2: Create the English records page**

Use this frontmatter and heading order:

```markdown
---
title: Article records and technical data
description: Create accessory records, maintain their identity, and verify technical data.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Article records and technical data

## Access rights and persistence

## Create an article

## Common field reference

## Choose article type and technical data

## Use barcode and article-data search

## Review duplicate candidates

## Edit without losing data

## Archive, restore, or delete an article

## Resolve validation and resource errors

## Related pages

## Documented RailKeeper version
```

Use concise tables for common fields, manufacturer status, inventory strategy, article types, and
type-specific field groups. State draft versus immediate persistence for every action. Cover the
exact confirmations, validation errors, historical inactive choices, stale resources, retry
actions, unsaved-close behavior, and permanent-delete consequences.

- [ ] **Step 3: Create the German semantic counterpart**

Use matching structure and this frontmatter:

```markdown
---
title: Artikelstammdaten und Fachangaben
description: Zubehörartikel anlegen, ihre Identität pflegen und Fachangaben prüfen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Artikelstammdaten und Fachangaben
```

Use the stable German UI labels and preserve every validation, trust, permission, confirmation,
and persistence statement from the English page.

- [ ] **Step 4: Verify and commit Task 2**

Run validation and the whitespace check, then compare the English/German heading sequences:

```powershell
cd docs
node scripts/validate-docs.mjs
cd ..
rg -n "^## " docs/site/guide/accessories/article-records.md
rg -n "^## " docs/site/de/guide/accessories/article-records.md
git diff --check
```

Expected: validation succeeds and both pages have the same section count and semantic order.
Commit only the pair:

```powershell
git add docs/site/guide/accessories/article-records.md `
  docs/site/de/guide/accessories/article-records.md
git commit -m "docs: explain accessory article records"
```

---

### Task 3: Document stock, purchases, assets, and documents

**Files:**

- Create: `docs/site/guide/accessories/stock-purchases-documents.md`
- Create: `docs/site/de/guide/accessories/stock-purchases-documents.md`

**Why:** Stock mutations have materially different effects for counted units and individual
assets. Users need those effects, confirmations, and recovery rules in one focused chapter.

**Change Necessity:** The stock and document lifecycle is too detailed for the overview and shares
one resource-refresh boundary in the stable editor.

**Impact/Compatibility:** Describe storage locations only as resources consumed by the published
accessories UI. Do not present backend-only storage-location CRUD as a visible workflow.

**Verification:** Focused documentation validation, mirrored heading scan, `git diff --check`, and
stable-source readback for inventory arithmetic, transactions, files, refreshes, and backup.

**Stable sources:**

- `ArticleStockTab.tsx`, `AccessoryStockPanel.tsx`, `AccessoryAssetForm.tsx`;
- `ArticlePurchaseDocumentsTab.tsx` and its tests;
- `articleEditorResources.ts`, `useArticleEditorController.ts`;
- accessory inventory, purchase, and document application services and handlers;
- remote accessory image import, upload security, stable translations, tests, and backup code.

- [ ] **Step 1: Audit stock and document state transitions**

Confirm from tag `v0.1.17.6`:

- quantity supports adjustments and transfers; individual supports assets; hybrid supports both
  counted stock and individualization;
- positive and negative adjustments use an active location and reject zero;
- transfer requires different active source/target locations and positive available quantity;
- every accepted adjustment, transfer, individualization, purchase booking, installation, and
  removal creates the exact stable stock movement entries;
- reserved or installed assets cannot be edited from the asset table;
- asset fields and allowed non-reserved/non-installed lifecycle values match stable code;
- purchase fields, quantity/date rules, `bookToStock`, and required location behavior are exact;
- purchase creation plus stock booking has the transaction/partial-write behavior proven by the
  stable service, not inferred from separate UI requests;
- document categories, primary-image rules, accepted operations, URL import behavior, and
  immediate persistence are exact;
- public URL, redirect, private-address, file type, size, empty-file, and path-confinement rules
  match the stable handlers;
- successful sub-actions refresh all related resources; failed refresh marks resources stale and
  disables further writes until retry succeeds;
- stable backup inclusion is stated only after reading the backup implementation/tests.

- [ ] **Step 2: Create the English stock page**

Use this frontmatter and heading order:

```markdown
---
title: Stock, purchases, and documents
description: Manage accessory quantities, individual items, purchases, images, and documents.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Stock, purchases, and documents

## Choose an inventory strategy

## Read stock and availability

## Adjust quantity stock

## Transfer stock between locations

## Manage individual items

## Record purchases

## Manage images and documents

## Protect data during immediate writes

## Resolve stock and document errors

## Related pages

## Documented RailKeeper version
```

Include a strategy matrix for quantity, individual, and hybrid. Include complete field/operation
tables for assets, purchases, documents, and stock movements. For every button state whether a
confirmation occurs, which record changes, whether a refresh follows, and how to recover if the
write succeeded but resource reload failed.

- [ ] **Step 3: Create the German semantic counterpart**

Use matching structure and this frontmatter:

```markdown
---
title: Bestand, Käufe und Dokumente
description: Zubehörmengen, Einzelstücke, Käufe, Bilder und Dokumente verwalten.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Bestand, Käufe und Dokumente
```

Preserve every arithmetic, lifecycle, confirmation, transaction, security, refresh, and recovery
statement. Use `Einzelstücke` consistently for individually tracked assets.

- [ ] **Step 4: Verify and commit Task 3**

Run:

```powershell
cd docs
node scripts/validate-docs.mjs
cd ..
rg -n "^## " docs/site/guide/accessories/stock-purchases-documents.md
rg -n "^## " docs/site/de/guide/accessories/stock-purchases-documents.md
git diff --check
```

Expected: validation succeeds, the section order is mirrored, and no generated artifact is
tracked. Commit only the pair:

```powershell
git add docs/site/guide/accessories/stock-purchases-documents.md `
  docs/site/de/guide/accessories/stock-purchases-documents.md
git commit -m "docs: explain accessory stock and documents"
```

---

### Task 4: Document reservations, installations, and history

**Files:**

- Create: `docs/site/guide/accessories/allocations-history.md`
- Create: `docs/site/de/guide/accessories/allocations-history.md`

**Why:** Allocation commands change stock, asset lifecycle, reservation state, installation state,
and history. The page must make those coupled effects predictable.

**Change Necessity:** This lifecycle has its own Planner/Editor boundary and cannot be reduced to a
stock-page footnote without hiding target and recovery rules.

**Impact/Compatibility:** Vehicle, layout, and layout-unit targets may be named because the stable
dialog exposes them. Layout creation or editing remains explicitly unpublished.

**Verification:** Focused documentation validation, mirrored heading scan, `git diff --check`, and
stable-source readback for every permission, target, state transition, and usage event.

**Stable sources:**

- `AccessoryReservationsPanel.tsx`, `AccessoryInstallationsPanel.tsx`;
- `AccessoryTargetFields.tsx`, `AccessoryTechnicalFields.tsx`;
- `ArticleUsageHistoryTab.tsx`, `ArticleStockTab.tsx`;
- `useArticleEditorController.ts` permissions/resource refresh;
- `backend/internal/application/accessory_allocations.go` and tests;
- allocation handlers, route access, stable types, and translations.

- [ ] **Step 1: Audit allocation invariants and effects**

Confirm from tag `v0.1.17.6`:

- target union is exactly vehicle, layout, or layout unit;
- technical placement fields are placement, digital address, decoder output, connection, and
  wiring notes;
- Planner can create/cancel reservations but cannot install or manage general stock;
- reservation requires product, target, source location, positive quantity, and an eligible asset
  when individual tracking requires one;
- only active reservations can be cancelled;
- installation can consume an active reservation or use the stable direct path;
- a selected reservation constrains target, location, asset, and quantity exactly as the UI does;
- installation condition defaults and all stable condition values are exact;
- condition update and removal each require their stable confirmation;
- removal disposition is stored, maintenance, defective, or retired; stored requires a location;
- quantity stock and asset lifecycle effects are exact for reservation, cancellation, installation,
  condition change, and removal;
- usage event types and chronology are reservation, installation, condition change, and removal;
- allocation summary meanings for owned, stored, reserved, installed, available, and missing match
  stable service calculations;
- stale-resource and failed-refresh recovery matches the controller.

- [ ] **Step 2: Create the English allocation page**

Use this frontmatter and heading order:

```markdown
---
title: Reservations, installations, and usage
description: Reserve accessory stock, record installations, and understand usage history.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Reservations, installations, and usage

## Roles and allocation targets

## Read the allocation summary

## Reserve stock

## Cancel a reservation

## Record an installation

## Update installation condition

## Remove an installation

## Read usage history

## Keep stock and lifecycle state consistent

## Resolve allocation errors

## Related pages

## Documented RailKeeper version
```

Use effect tables that separately state quantity-stock, asset, reservation, installation, and
history changes. Explain every target and technical-placement field without documenting how to
operate the layout workspace. Include all role, confirmation, stale-state, partial-result, and
recovery rules.

- [ ] **Step 3: Create the German semantic counterpart**

Use matching structure and this frontmatter:

```markdown
---
title: Reservierungen, Einbauten und Verwendung
description: Zubehör reservieren, Einbauten erfassen und die Verwendungshistorie verstehen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Reservierungen, Einbauten und Verwendung
```

Keep `Reservierung`, `Einbau`, `Ausbau`, `Verbleib`, and `Verwendungshistorie` consistent with the
stable UI. Preserve the same effect matrices and warnings.

- [ ] **Step 4: Verify and commit Task 4**

Run validation, heading comparison, and whitespace checks as in Tasks 2 and 3. Expected: all pass.
Commit only the allocation pair:

```powershell
git add docs/site/guide/accessories/allocations-history.md `
  docs/site/de/guide/accessories/allocations-history.md
git commit -m "docs: explain accessory allocations"
```

---

### Task 5: Connect navigation and published cross-links

**Files:**

- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/site/guide/overview/index.md`
- Modify: `docs/site/de/guide/overview/index.md`
- Modify: `docs/site/guide/vehicles/search-and-spares.md`
- Modify: `docs/site/de/guide/vehicles/search-and-spares.md`
- Modify: all eight new accessory pages for final related links

**Why:** Published content must be discoverable from the site hierarchy and existing owner
boundaries without exposing planned destinations.

**Change Necessity:** The pages build without sidebar and landing entries but would remain easy to
miss, contradicting the approved wiki-navigation goal.

**Impact/Compatibility:** Wiring only. Existing vehicle routes, order, titles, and metadata remain
unchanged except for narrow related-page additions.

**Verification:** Full `npm.cmd run check`, route scan for all eight pages, `git diff --check`, and
review of every new link against an existing published Markdown destination.

- [ ] **Step 1: Add mirrored sidebar groups**

After the existing User Guide vehicle group, add an English **Accessories** group with links in this
order:

```text
/guide/accessories/
/guide/accessories/article-records
/guide/accessories/stock-purchases-documents
/guide/accessories/allocations-history
```

Add the German **Zubehör** group in the same position and order under `/de/guide/accessories/...`.
Use the approved page titles as sidebar labels.

- [ ] **Step 2: Add landing and boundary transitions**

Add one concise workflow paragraph to each User Guide landing page after the vehicle search/spares
paragraph. Add the Accessories overview to the related-page lists of the Overview chapter and the
vehicle search/spares chapter. The vehicle link must clarify that vehicle spare parts and accessory
articles are separate inventories.

- [ ] **Step 3: Finalize related links on all accessory pages**

Every accessory page links to:

- User Guide overview;
- Accessories overview;
- the other three accessory pages where it is not the current page.

The overview may additionally link to Overview metrics. The records page may link to vehicle
article search only to explain the boundary. Do not link to planned Settings, master-data,
backup/restore, layouts, or administration pages.

- [ ] **Step 4: Verify navigation and commit Task 5**

Run:

```powershell
cd docs
npm.cmd run check
cd ..
git diff --check
rg -n "/(de/)?guide/accessories" docs/.vitepress/config.mts docs/site
git status --short
```

Expected: 19 tests pass, coverage validation succeeds, VitePress builds every new route, and all
English/German navigation paths appear. Commit only Task 5 wiring files:

```powershell
git add docs/.vitepress/config.mts docs/site/guide/index.md docs/site/de/guide/index.md `
  docs/site/guide/overview/index.md docs/site/de/guide/overview/index.md `
  docs/site/guide/vehicles/search-and-spares.md `
  docs/site/de/guide/vehicles/search-and-spares.md docs/site/guide/accessories `
  docs/site/de/guide/accessories
git commit -m "docs: link accessories user guide"
```

---

### Task 6: Prove source fidelity, parity, and publication readiness

**Files:**

- Review: all task-owned files
- Modify only when a verified finding requires correction

**Why:** A polished guide is unsafe if it misstates permissions, stock effects, persistence, or
stable-version behavior.

**Change Necessity:** Review edits are allowed only for evidence-backed documentation defects and
must remain inside the approved file boundary.

**Impact/Compatibility:** No scope expansion. Any discovery that belongs to another coverage owner
returns to design review instead of being absorbed.

**Verification:** Complete stable-source audit, bilingual/hygiene scans, fresh full documentation
check, focused review, exact-head PR checks, merge workflows, and eight live HTTP checks.

- [ ] **Step 1: Run the complete stable-source audit**

Check each claim against `v0.1.17.6`, not the worktree source. Use `git show` and `git grep` over:

```text
frontend/src/features/accessories/**
frontend/src/shared/apiLayoutsAccessories.ts
frontend/src/shared/i18n/en.ts
frontend/src/shared/i18n/de.ts
backend/internal/api/accessory*_handlers.go
backend/internal/api/routes.go
backend/internal/application/accessor*.go
backend/internal/domain/accessory*.go
backend migrations, backup implementation/tests, and stable OpenAPI paths
```

Build a review checklist covering: metrics, search fields, filters, sort, view persistence, columns,
roles, fields, master-data behavior, type automation, search trust, duplicates, draft persistence,
archive/delete, inventory strategies, arithmetic, movements, assets, purchases, documents, URL/file
safety, targets, reservations, installations, conditions, removal, history, backup, refresh, and
recovery. Correct every unsupported or incomplete statement.

- [ ] **Step 2: Run bilingual and hygiene scans**

Run from the worktree root:

```powershell
rg -n "TBD|TODO|FIXME|coming soon|placeholder" docs/site/guide/accessories `
  docs/site/de/guide/accessories
rg -n "^## " docs/site/guide/accessories docs/site/de/guide/accessories
rg -n "reviewedVersion:|lastReviewed:|status:|audience:" `
  docs/site/guide/accessories docs/site/de/guide/accessories
git diff --check origin/main...HEAD
```

Expected: no unfinished markers; paired pages have matching semantic heading counts; all metadata
uses user/stable/0.1.17.6/2026-08-16; whitespace is clean.

Check line lengths and correct prose lines over 120 characters unless a Markdown table or
unbreakable link makes the line clearer:

```powershell
$files = Get-ChildItem docs/site/guide/accessories,docs/site/de/guide/accessories -Filter *.md
foreach ($file in $files) {
  $lines = Get-Content $file.FullName
  for ($index = 0; $index -lt $lines.Count; $index++) {
    if ($lines[$index].Length -gt 120) {
      "{0}:{1}:{2}" -f $file.FullName, ($index + 1), $lines[$index].Length
    }
  }
}
```

- [ ] **Step 3: Run full documentation verification**

Run fresh:

```powershell
cd docs
npm.cmd run check
cd ..
git diff --check origin/main...HEAD
git status --short --branch
```

Expected: 19 tests pass, coverage validation succeeds, VitePress production build succeeds, diff
check is clean, and the worktree contains no uncommitted task changes.

- [ ] **Step 4: Request focused review and resolve findings**

Use `aegis:requesting-code-review` for a read-only review of `origin/main...HEAD`. Review questions:

1. Does every behavioral claim match `v0.1.17.6` rather than later `main`?
2. Are English and German semantically equivalent?
3. Are Viewer, Planner, Editor, Admin, and Messe boundaries exact?
4. Are search results, drafts, immediate writes, refreshes, and backup states separated?
5. Are quantity, hybrid, and individual inventory effects exact?
6. Are document URL/file protections and path confinement accurate?
7. Are reservation, installation, removal, asset, and history transitions exact?
8. Are partial-write and failed-refresh recovery instructions safe?
9. Are layout and planned-administration boundaries preserved?

Correct valid findings, rerun the focused checks, and obtain a clean re-review. Commit corrections
as one scoped commit:

```powershell
git add docs/coverage.json docs/.vitepress/config.mts `
  docs/site/guide/accessories docs/site/de/guide/accessories `
  docs/site/guide/index.md docs/site/de/guide/index.md `
  docs/site/guide/overview/index.md docs/site/de/guide/overview/index.md `
  docs/site/guide/vehicles/search-and-spares.md `
  docs/site/de/guide/vehicles/search-and-spares.md
git commit -m "docs: refine accessories user guide"
```

If no files changed, do not create an empty commit.

- [ ] **Step 5: Push, open the PR, and verify the reviewed head**

Run:

```powershell
git push -u origin dev/docs-user-guide-accessories
gh pr create --base main --head dev/docs-user-guide-accessories `
  --title "docs: add bilingual accessories guide" `
  --body "Publishes the complete stable v0.1.17.6 accessories guide in English and German."
gh pr checks --watch
```

Record the exact head SHA with `git rev-parse HEAD`. Do not merge if a check is failing, pending, or
was run against another SHA.

- [ ] **Step 6: Merge only when green and verify publication**

After all required PR checks pass for the reviewed head:

```powershell
gh pr merge --merge
```

Wait for merge-triggered CI, CodeQL, Trivy, Documentation Pages, and Docker Image workflows. Verify
their conclusions against the merge commit:

```powershell
$mergeSha = gh pr view --json mergeCommit --jq '.mergeCommit.oid'
gh run list --commit $mergeSha --limit 20 `
  --json databaseId,workflowName,status,conclusion,url
```

Verify all eight live routes return HTTP 200 and contain `v0.1.17.6`:

```text
https://ichwars.github.io/RailKeeper/guide/accessories/
https://ichwars.github.io/RailKeeper/guide/accessories/article-records
https://ichwars.github.io/RailKeeper/guide/accessories/stock-purchases-documents
https://ichwars.github.io/RailKeeper/guide/accessories/allocations-history
https://ichwars.github.io/RailKeeper/de/guide/accessories/
https://ichwars.github.io/RailKeeper/de/guide/accessories/article-records
https://ichwars.github.io/RailKeeper/de/guide/accessories/stock-purchases-documents
https://ichwars.github.io/RailKeeper/de/guide/accessories/allocations-history
```

Use this exact read-only check after Documentation Pages succeeds:

```powershell
$urls = @(
  'https://ichwars.github.io/RailKeeper/guide/accessories/',
  'https://ichwars.github.io/RailKeeper/guide/accessories/article-records',
  'https://ichwars.github.io/RailKeeper/guide/accessories/stock-purchases-documents',
  'https://ichwars.github.io/RailKeeper/guide/accessories/allocations-history',
  'https://ichwars.github.io/RailKeeper/de/guide/accessories/',
  'https://ichwars.github.io/RailKeeper/de/guide/accessories/article-records',
  'https://ichwars.github.io/RailKeeper/de/guide/accessories/stock-purchases-documents',
  'https://ichwars.github.io/RailKeeper/de/guide/accessories/allocations-history'
)
foreach ($url in $urls) {
  $response = Invoke-WebRequest -UseBasicParsing $url
  if ($response.StatusCode -ne 200 -or $response.Content -notmatch 'v0\.1\.17\.6') {
    throw "Live verification failed: $url"
  }
  "OK $($response.StatusCode) $url"
}
```

Preserve the completed worktree for traceability unless the user later requests cleanup. Local
`main` remains untouched.

## Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| Later `main` behavior leaks into stable docs | Resolve every claim against `v0.1.17.6`. |
| One language omits a safety rule | Mirror headings, tables, warnings, and review each pair. |
| Counted and individual stock are conflated | Lead with a three-strategy matrix and effect tables. |
| Planner permissions are overstated | Verify both route access and editor permission derivation. |
| Backend-only storage-location operations look user-facing | Document locations as consumed resources only. |
| Layout work appears published | Name stable targets while explicitly excluding layout operation. |
| Immediate write is mistaken for dialog draft | State persistence and refresh after every action. |
| Partial success causes duplicate retry | Describe reload, reconcile, then retry only missing work. |
| Generated docs output enters Git | Inspect status and stage explicit task-owned paths only. |
| PR merges a stale review | Record head SHA and require checks on that exact head. |

## Retirement and rollback

No runtime path, adapter, compatibility layer, or old documentation owner is introduced. Rollback
is a normal revert of the documentation commits. Reverting must restore `accessories` to `planned`,
remove all eight pages and sidebar/landing links together, and leave other coverage owners intact.

## Expected completion evidence

- `accessories` is `documented` with its existing overview paths;
- all four English/German page pairs exist with stable metadata and mirrored content;
- navigation and published cross-links reach every page;
- every stable claim has a `v0.1.17.6` source owner;
- local documentation checks pass from a clean task worktree;
- focused review has no unresolved findings;
- the exact reviewed PR head has all required checks green;
- the PR is merged and all eight live pages return HTTP 200;
- local `main` and user-owned untracked files remain unchanged.
