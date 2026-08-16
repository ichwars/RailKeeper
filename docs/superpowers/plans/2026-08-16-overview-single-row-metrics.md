# Overview Single-Row Metrics Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep all four overview metric areas on one row at desktop widths without changing tablet or mobile stacking.

**Architecture:** Retain the existing four children and the valuation card's two-column span. Change only the desktop grid from four to five equal tracks, while the existing `max-width: 900px` and `max-width: 640px` overrides continue to provide tablet and mobile layouts.

**Tech Stack:** React, TypeScript, CSS Grid, Vitest, Testing Library, Vite

## Global Constraints

- Desktop uses five equal grid tracks: one each for inventory, digitalization and maintenance, two for valuation.
- At widths up to 900 px, retain the existing two-column grid and valuation span of two.
- At widths up to 640 px, retain the existing single-column grid and valuation span of one.
- Do not change card content, valuation calculations, API behavior, height, spacing, or issue #82.
- Do not push, open a PR, merge, or publish before local user approval.

---

### Task 1: Keep overview metrics on one desktop row

**Files:**
- Modify: `frontend/src/features/overview/OverviewView.test.tsx`
- Modify: `frontend/src/styles/overview.css`
- Test: `frontend/src/features/overview/OverviewView.test.tsx`

**Interfaces:**
- Consumes: `.overview-hero` and `.overview-valuation-card` from the existing overview markup.
- Produces: a five-track desktop grid with the existing valuation card spanning two tracks.

- [x] **Step 1: Write the failing desktop-grid assertion**

Extend the existing responsive layout test with the exact desktop contract:

```ts
expect(overview).toContain("grid-template-columns: repeat(5, minmax(0, 1fr))");
expect(overview).toContain(".overview-valuation-card");
expect(overview).toContain("grid-column: span 2");
```

- [x] **Step 2: Run the focused test and verify RED**

Run:

```powershell
npm.cmd run test:run -- src/features/overview/OverviewView.test.tsx
```

Expected: FAIL because `overview.css` still contains `repeat(4, minmax(0, 1fr))`.

- [x] **Step 3: Implement the minimal desktop CSS change**

Change only the base desktop rule in `frontend/src/styles/overview.css`:

```css
.overview-hero {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}
```

Do not edit the responsive overrides.

- [x] **Step 4: Verify GREEN and run the frontend regression checks**

Run:

```powershell
npm.cmd run test:run -- src/features/overview/OverviewView.test.tsx
npm.cmd run test:run
npm.cmd run build
```

Expected: focused test PASS, 96 test files and 497 tests or more PASS, production build PASS.

- [x] **Step 5: Perform local visual verification**

Reload `http://127.0.0.1:8083/overview` after the build and verify:

- desktop: inventory, digitalization, valuation and maintenance occupy one row;
- valuation remains two tracks wide with its 2×2 value matrix;
- mobile at 390×844 remains one column without horizontal overflow;
- browser console contains no errors.

- [x] **Step 6: Commit the narrow correction**

```powershell
git add frontend/src/features/overview/OverviewView.test.tsx frontend/src/styles/overview.css docs/superpowers/plans/2026-08-16-overview-single-row-metrics.md
git commit -m "fix: keep overview metrics on one desktop row"
```
