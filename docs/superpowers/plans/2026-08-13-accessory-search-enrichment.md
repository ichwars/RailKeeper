# Accessory Search Enrichment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add safe structured track-data suggestions to accessory web search and automatically manage scale and keyword defaults without overwriting manual edits.

**Architecture:** Extend the existing generic article-search `fields` map with canonical track keys, parsed by a focused backend helper and converted by the accessory frontend into existing typed attributes. Keep scale and keyword automation in the article-editor session, where current form values, gauge metadata, and localized type/subtype labels are available.

**Tech Stack:** Go 1.24 backend, React 19, TypeScript, Vitest, Testing Library, Vite 8, SQLite-backed master data.

## Global Constraints

- Every web-search value remains a suggestion and requires explicit selection before import.
- Existing URL, timeout, redirect, MIME, and size protections remain unchanged.
- Manual scale and keyword values always win over automatic defaults.
- Non-track article types retain their current article-search behavior.
- No database migration or new API response shape is introduced.
- Keep all work local on `dev/fix-accessory-search-dialog`; do not push or merge.

---

### Task 1: Extract canonical track fields in article search

**Files:**
- Create: `backend/internal/application/article_search_track_fields.go`
- Modify: `backend/internal/application/article_search_ranking.go`
- Test: `backend/internal/application/article_search_test.go`

**Interfaces:**
- Consumes: `ArticleSearchInput.Fields["articleType"]` and cleaned page/search text.
- Produces: `buildTrackArticleFields(input ArticleSearchInput, text string) map[string]ArticleSearchField`.

- [ ] **Step 1: Write failing extraction tests**

Add table-driven tests that call `buildArticleFields` with `Fields: map[string]string{"articleType": "track"}` and representative labelled specifications. Assert canonical values for text, decimal numbers, booleans, connection count, and direction:

```go
input := ArticleSearchInput{Fields: map[string]string{"articleType": "track"}}
fields := buildArticleFields(input, "Tillig EW1", "https://www.tillig.com/83329", `
Gleissystem: TT Modellgleis
Länge: 129,5 mm
Radius: 353 mm
Winkel: 15°
Richtung: links
Herzstückwinkel: 12°
Schwellenart: Holz
Profilhöhe: 2,07 mm
Bettung: nein
Anzahl Anschlüsse: 3
Digitaltauglich: ja`)
expectArticleField(t, fields, "trackSystem", "TT Modellgleis")
expectArticleField(t, fields, "lengthMm", "129.5")
expectArticleField(t, fields, "direction", "left")
expectArticleField(t, fields, "roadbed", "false")
expectArticleField(t, fields, "digitalReady", "true")
```

Also assert that the same text does not add track fields when `articleType` is not `track`, and that ambiguous boolean prose produces no boolean field.

- [ ] **Step 2: Run the backend test and verify RED**

Run: `go test ./internal/application -run 'Test.*Track.*Article.*Fields' -count=1`

Expected: FAIL because canonical track fields are absent.

- [ ] **Step 3: Implement the focused parser**

Create `article_search_track_fields.go` with explicit label maps and conservative normalizers:

```go
func buildTrackArticleFields(input ArticleSearchInput, text string) map[string]ArticleSearchField {
    if !strings.EqualFold(strings.TrimSpace(input.Fields["articleType"]), "track") {
        return nil
    }
    fields := map[string]ArticleSearchField{}
    addTextTrackField(fields, "trackSystem", "Gleissystem", text,
        []string{"Gleissystem", "Gleissortiment", "Track system"})
    addDecimalTrackField(fields, "lengthMm", "Länge (mm)", text,
        []string{"Länge", "Laenge", "Length"})
    addDecimalTrackField(fields, "radiusMm", "Radius (mm)", text, []string{"Radius"})
    addDecimalTrackField(fields, "angleDegrees", "Winkel (°)", text,
        []string{"Winkel", "Angle"})
    addDirectionTrackField(fields, text)
    addDecimalTrackField(fields, "frogAngleDegrees", "Herzstückwinkel (°)", text,
        []string{"Herzstückwinkel", "Herzstueckwinkel", "Frog angle"})
    addTextTrackField(fields, "sleeperType", "Schwellenart", text,
        []string{"Schwellenart", "Schwellen", "Sleeper type"})
    addDecimalTrackField(fields, "railHeightMm", "Profilhöhe (mm)", text,
        []string{"Profilhöhe", "Profilhoehe", "Rail height"})
    addBooleanTrackField(fields, "roadbed", "Bettung", text,
        []string{"Bettung", "Gleisbettung", "Roadbed"})
    addIntegerTrackField(fields, "connectionCount", "Anzahl Anschlüsse", text,
        []string{"Anzahl Anschlüsse", "Anzahl Anschluesse", "Connections"})
    addBooleanTrackField(fields, "digitalReady", "Digitaltauglich", text,
        []string{"Digitaltauglich", "Digital geeignet", "Digital ready"})
    return fields
}
```

Normalize decimal commas to dots, accept only finite nonnegative values, map `links/rechts/symmetrisch` and English equivalents to `left/right/symmetric`, and accept booleans only for explicit `ja/nein`, `yes/no`, or `true/false` labelled values.

Merge these fields at the end of `buildArticleFields`, retaining the higher-confidence value if a generic field already exists.

- [ ] **Step 4: Run focused and full backend verification**

Run:

```powershell
go test ./internal/application -run 'Test.*Track.*Article.*Fields' -count=1
go test ./...
go vet ./...
```

Expected: all commands PASS.

---

### Task 2: Convert track search suggestions into typed accessory subject values

**Files:**
- Modify: `frontend/src/features/accessories/accessoryArticleSearch.ts`
- Modify: `frontend/src/features/accessories/useAccessoryArticleSearchController.ts`
- Test: `frontend/src/features/accessories/accessoryArticleSearch.test.ts`
- Test: `frontend/src/features/accessories/useAccessoryArticleSearchController.test.tsx`

**Interfaces:**
- Consumes: canonical track search fields from Task 1 and the existing `articleTypeFieldRegistry.track` definitions.
- Produces: `accessorySearchInput`, `accessorySearchFieldGroups`, `currentAccessorySearchValue`, `isSelectableAccessorySearchValue`, and `applyAccessorySearchResult` with typed track support.

- [ ] **Step 1: Write failing frontend mapping tests**

Create a track form containing existing attributes and assert that:

```ts
expect(accessorySearchInput(form).fields).toMatchObject({
  articleType: "track",
  subtype: "turnout",
  trackSystem: "TT Modellgleis",
  lengthMm: "129.5",
  roadbed: "false"
});
expect(currentAccessorySearchValue(form, "lengthMm")).toBe("129.5");
expect(isSelectableAccessorySearchValue("direction", "left", manufacturers, gauges, "track")).toBe(true);
expect(isSelectableAccessorySearchValue("direction", "up", manufacturers, gauges, "track")).toBe(false);
```

Apply a selected result and assert that text/boolean/single-select attributes are replaced by key while numbers populate `attributeNumberDrafts` with the normalized value and preserve the registry unit.

- [ ] **Step 2: Run focused tests and verify RED**

Run: `npm.cmd test -- --run src/features/accessories/accessoryArticleSearch.test.ts src/features/accessories/useAccessoryArticleSearchController.test.tsx`

Expected: FAIL because track fields are neither exposed nor applied.

- [ ] **Step 3: Implement typed mapping**

Add helpers based on `fieldDefinitionsForType(form.articleType)`:

```ts
function subjectSearchValue(form: ArticleEditorForm, key: string): string {
  const draft = form.attributeNumberDrafts[key];
  if (draft !== undefined) return draft.trim();
  const attribute = form.attributes.find((item) => item.key === key);
  if (!attribute) return "";
  if (attribute.kind === "text") return attribute.textValue;
  if (attribute.kind === "number") return String(attribute.numberValue);
  if (attribute.kind === "boolean") return String(attribute.booleanValue);
  return attribute.optionValues.join(", ");
}
```

Include `articleType`, `subtype`, and nonempty subject values in the search request. Add a localized `Fachangaben: Gleis` group using existing subject-field translations. Validate each suggestion against its definition before selection. Convert selected values into existing `AccessoryAttributeValue` variants without changing the API response type.

Update `canSelectField` calls in the controller to pass the current article type.

- [ ] **Step 4: Run focused frontend tests**

Run: `npm.cmd test -- --run src/features/accessories/accessoryArticleSearch.test.ts src/features/accessories/useAccessoryArticleSearchController.test.tsx`

Expected: PASS.

---

### Task 3: Auto-manage scale and keyword defaults per editor session

**Files:**
- Create: `frontend/src/features/accessories/articleEditorAutomation.ts`
- Create: `frontend/src/features/accessories/articleEditorAutomation.test.ts`
- Modify: `frontend/src/features/accessories/ArticleEditorDialog.tsx`
- Modify: `frontend/src/features/accessories/ArticleCoreTab.tsx`
- Test: `frontend/src/features/accessories/ArticleEditorDialog.test.tsx`

**Interfaces:**
- Produces: `scaleForGauges(gauges: string[], entries: MasterDataEntry[]): string` and `suggestedArticleKeywords(form: ArticleEditorForm, articleTypeLabel: string, subtypeLabel: string): string`.
- Consumes: gauge `metadata.scale`, localized type/subtype labels, and the existing `onChange` form patch seam.

- [ ] **Step 1: Write failing pure helper tests**

Assert that TT resolves to `1:120`, missing metadata returns an empty string, and keyword values are trimmed and deduplicated case-insensitively:

```ts
expect(scaleForGauges(["TT"], gaugeEntries)).toBe("1:120");
expect(suggestedArticleKeywords(form, "Gleis", "Weiche"))
  .toBe("Einfache Weiche EW1 links, Tillig, Gleis, Weiche");
```

- [ ] **Step 2: Write failing editor-session behavior tests**

Use an `ArticleEditorDialog` harness whose `onChange` updates form state. Verify:

1. selecting TT fills `1:120`;
2. selecting another gauge updates an untouched auto scale;
3. manually typing `1:100` prevents later gauge changes from overwriting it;
4. name/manufacturer/type/subtype changes update generated keywords;
5. manually editing keywords prevents later source changes from overwriting them;
6. a persisted nonempty keyword field starts in manual mode.

- [ ] **Step 3: Run tests and verify RED**

Run: `npm.cmd test -- --run src/features/accessories/articleEditorAutomation.test.ts src/features/accessories/ArticleEditorDialog.test.tsx`

Expected: FAIL because automatic defaults do not exist.

- [ ] **Step 4: Implement helpers and session ownership**

Implement pure derivation helpers. In `ArticleEditorDialog`, initialize refs once per keyed session:

```ts
const scaleAutoManagedRef = useRef(!props.form.scale.trim());
const keywordAutoManagedRef = useRef(!props.form.keywords.trim());
const lastAutoScaleRef = useRef("");
```

Wrap user-originated core changes so direct edits to `scale` or `keywords` disable their respective automatic mode. Use effects keyed to gauges and keyword source fields to call `props.onChange` only when the derived value differs. The editor is already keyed by `sessionKey`, so refs reset for each create/edit session. Keep `ArticleCoreTab` controlled and route scale/keyword edits through the wrapper.

- [ ] **Step 5: Run focused frontend tests**

Run: `npm.cmd test -- --run src/features/accessories/articleEditorAutomation.test.ts src/features/accessories/ArticleEditorDialog.test.tsx`

Expected: PASS.

---

### Task 4: Preserve and verify popup stacking and gauge typeahead fixes

**Files:**
- Modify: `frontend/src/styles/vehicle-dialogs.css`
- Modify: `frontend/src/shared/ui/AppMultiSelect.tsx`
- Test: `frontend/src/features/accessories/accessoriesResponsive.test.ts`
- Test: `frontend/src/shared/ui/AppMultiSelect.test.tsx`

**Interfaces:**
- Preserves the already implemented local behavior: search layers above editor layer and label typeahead for multi-selects.

- [ ] **Step 1: Re-run the existing focused regression tests**

Run: `npm.cmd test -- --run src/features/accessories/accessoriesResponsive.test.ts src/shared/ui/AppMultiSelect.test.tsx`

Expected: 13 tests PASS.

- [ ] **Step 2: Review integration with new scale behavior**

Confirm the gauge typeahead test selects `TT`, the editor receives `gauges: ["TT"]`, and Task 3 derives `scale: "1:120"` without treating the selection as a manual scale edit.

---

### Task 5: Full verification and local server update

**Files:**
- Verify only; do not commit generated `frontend/dist`, `.cache`, or local `data`.

**Interfaces:**
- Produces a locally testable `v0.1.17` server with the unpushed fixes.

- [ ] **Step 1: Run all checks**

Run:

```powershell
cd backend
go test ./...
go vet ./...
cd ..\frontend
npm.cmd test -- --run
npm.cmd run build
cd ..
git diff --check
git status --short
```

Expected: all tests/builds pass; only intended source, test, spec, and plan files are changed or committed.

- [ ] **Step 2: Restart the local server from this worktree**

Identify the exact PID listening on `18083`, stop only that process, and start `go run ./cmd/railkeeper` with:

```powershell
$env:RAILKEEPER_ADDR=':18083'
$env:RAILKEEPER_DATA_DIR='C:\Users\droth\Documents\GitHub\RailKeeper\data'
$env:RAILKEEPER_MIGRATIONS_DIR='C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-search-interactions\backend\migrations'
$env:RAILKEEPER_SEEDS_DIR='C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-search-interactions\backend\seeds'
$env:RAILKEEPER_STATIC_DIR='C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-search-interactions\frontend\dist'
$env:GOCACHE='C:\Users\droth\Documents\GitHub\RailKeeper\.worktrees\accessory-search-interactions\.cache\go-build'
```

Expected: `/health` returns `{"status":"ok"}`, `/api/v1/version` returns `0.1.17`, `/accessories` returns HTTP 200, and its HTML references the newly built asset.

- [ ] **Step 3: Report local-only state**

Report exact test counts, build result, changed files, server state, and that no implementation commit, push, PR, merge, or release was performed.
