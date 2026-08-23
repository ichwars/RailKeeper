# Responsive ECoS Adoption Actions Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep the ECoS adoption actions fully visible in the locomotive comparison dialog at narrow widths.

**Architecture:** Preserve the existing comparison component and data flow. Conditionally omit the
impossible write-preview action for unassigned locomotives, then make the existing footer wrap its
remaining buttons without changing their desktop sizing.

**Tech Stack:** React 19, TypeScript, CSS, Vitest, Testing Library, Vite

## Global Constraints

- Do not change ECoS read, assignment, vehicle creation, or write APIs.
- Keep German and English action labels unchanged.
- Assigned locomotives retain the existing write-preview flow.
- Use existing RailKeeper controls and design tokens.

---

### Task 1: Make unassigned locomotive actions responsive

**Files:**
- Modify: `frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx`
- Modify: `frontend/src/features/digitalCenters/DigitalCentersView.test.tsx`
- Modify: `frontend/src/styles/digital-centers.css`

**Interfaces:**
- Consumes: `DigitalCenterWorkItem.vehicleId`, existing adoption callbacks, and `.digital-center-button`.
- Produces: An unassigned-state footer containing create, assign, and close actions only, with wrapping.

- [ ] **Step 1: Write failing component and CSS contract assertions**

In the existing `explains why an ECoS-only locomotive cannot be written` test, replace the disabled
write-preview assertion with:

```tsx
expect(screen.queryByRole("button", { name: "Schreibvorschau erstellen" }))
  .not.toBeInTheDocument();
expect(screen.getByRole("button", { name: "Neues Fahrzeug anlegen" })).toBeVisible();
expect(screen.getByRole("button", { name: "Bestehendem Fahrzeug zuordnen" })).toBeVisible();
```

Add this responsive CSS assertion to the existing CSS contract describe block:

```tsx
it("wraps comparison actions instead of clipping them at narrow widths", () => {
  expect(digitalCentersCSS).toMatch(
    /\.digital-comparison-dialog\s*>\s*footer\s*\{[^}]*flex-wrap:\s*wrap;/s
  );
  expect(digitalCentersCSS).toMatch(
    /\.digital-comparison-dialog\s*>\s*footer\s+\.digital-center-button\s*\{[^}]*max-width:\s*100%;/s
  );
});
```

- [ ] **Step 2: Run the focused test and verify the new assertions fail**

Run:

```powershell
cd frontend
npm.cmd test -- DigitalCentersView.test.tsx --run
```

Expected: FAIL because the disabled write-preview button still renders and the footer lacks wrapping.

- [ ] **Step 3: Hide the impossible action and allow wrapping**

Change the write-preview render condition in `LocomotiveComparisonDialog.tsx`:

```tsx
{!preview && !unassigned && <button type="button" className="digital-center-button"
  disabled={!canWrite || loading || changedFields.length === 0}
  onClick={() => void onPreview(changedFields).catch(() => undefined)}>
  {t("digitalCenters.write.createPreview")}
</button>}
```

Extend the existing footer styles in `digital-centers.css`:

```css
.digital-comparison-dialog > footer {
  justify-content: flex-end;
  flex-wrap: wrap;
  border-top: 1px solid var(--line);
}

.digital-comparison-dialog > footer .digital-center-button {
  max-width: 100%;
  white-space: normal;
}
```

- [ ] **Step 4: Run focused tests and the production build**

Run:

```powershell
cd frontend
npm.cmd test -- DigitalCentersView.test.tsx DigitalCentersFlows.test.tsx --run
npm.cmd run build
```

Expected: Both test files pass and the Vite production build exits with code 0.

- [ ] **Step 5: Run browser QA at narrow and desktop widths**

Use the existing in-app Browser tab at `http://127.0.0.1:8081/digital-centers`:

1. Reload after the build.
2. Read ECoS data and open `LokPilot 5 micro` comparison.
3. Verify create, assign, and close remain inside the dialog at the narrow reference width.
4. Verify write preview is absent for this unassigned locomotive.
5. Verify the assigned comparison flow still shows write preview at desktop width.
6. Check browser console errors and warnings.

Expected: No clipping, overlap, framework overlay, or relevant console error.

- [ ] **Step 6: Commit the focused UI fix**

```powershell
git add frontend/src/features/digitalCenters/LocomotiveComparisonDialog.tsx `
  frontend/src/features/digitalCenters/DigitalCentersView.test.tsx `
  frontend/src/styles/digital-centers.css
git commit -m "fix: keep ECoS adoption actions visible"
```
