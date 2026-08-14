# Limited ECoS Integration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Retain reviewed ECoS locomotive master-data read/write, CV values, static function keys, and symbols while removing runtime state, images, and non-locomotive object managers.

**Architecture:** Centralize production ECoS commands in explicit allowlisted builders in the application service. Narrow backend DTOs and parsers first, then align frontend types and views so excluded fields cannot reappear through raw responses or stored draft payloads. Preserve existing admin authorization, CSRF enforcement, dry-run, and confirmation behavior.

**Tech Stack:** Go, TCP ECoS client, React, TypeScript, Vitest, OpenAPI YAML, German/English i18n

## Global Constraints

- Keep connection tests and locomotive manager object `10` for on-demand locomotive list and detail access.
- Keep reads for object ID, name, address, protocol, profile, `funcdesc`, documented static function mapping, and CV values.
- Keep writes limited to name, address, and protocol with dry run and explicit confirmation.
- Do not introduce CV programming, function commands, driving commands, or STOP/GO commands.
- Do not query, store, expose, or display speed, speed step, current direction, `funcset`, active function state, ECoS locomotive images, switch objects, routes, S88, boosters, or other object managers.
- Keep locally shipped function symbols, mark their third-party origin, and ask ESU to review them.
- Do not migrate or delete existing RailKeeper vehicle data, CV data, function definitions, or local vehicle images.

---

### Task 1: Backend Command Allowlist and State-Free Parsing

**Files:**
- Modify: `backend/internal/application/ecos_test.go`
- Modify: `backend/internal/application/ecos.go`

**Interfaces:**
- Consumes: `ECoSService.exchange`, `exchangeRequestedGet`, and `exchangeRequestedCommands`
- Produces: `eCoSLocomotiveListCommand`, `eCoSLocomotiveDetailCommand(int)`, `eCoSLiveSubscriptionCommands()`, state-free `ECoSLocomotive`, `ECoSRawLocomotive`, and `ECoSFunction`

- [ ] **Step 1: Add failing command-allowlist tests**

Add tests asserting:

```go
func TestECoSCommandsExcludeRuntimeAndLayoutState(t *testing.T) {
	commands := append([]string{
		eCoSLocomotiveListCommand,
		eCoSLocomotiveDetailCommand(1001),
	}, eCoSLiveSubscriptionCommands()...)
	joined := strings.ToLower(strings.Join(commands, "\n"))
	for _, forbidden := range []string{
		"speed", "speedstep", " dir", "funcset", "switching",
		"queryobjects(11", "queryobjects(26", "request(11", "request(26",
		"icon", "image", "picture", "pic", "userimage",
	} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("forbidden ECoS field or manager %q in %q", forbidden, joined)
		}
	}
}

func TestECoSLiveCommandsOnlyMonitorBaseObject(t *testing.T) {
	want := []string{"request(1, view)", "get(1, info, status)"}
	if diff := cmp.Diff(want, eCoSLiveSubscriptionCommands()); diff != "" {
		t.Fatalf("unexpected live commands (-want +got):\n%s", diff)
	}
}
```

Use the project's existing comparison style instead of adding `cmp` if `ecos_test.go` currently uses `reflect.DeepEqual`.

- [ ] **Step 2: Add a failing parser regression test**

Parse a reply containing `speed`, `speedstep`, `dir`, `funcset`, `func`, and `funcdesc`, marshal the result to JSON, and assert that only the static `funcdesc` entry survives:

```go
payload, err := json.Marshal(parseECoSLocomotives([]string{
	`1001 name["BR 218"] speed[22] speedstep[4] dir[1] funcset[10] func[0,1] funcdesc[0,3]`,
}))
if err != nil { t.Fatal(err) }
text := string(payload)
for _, forbidden := range []string{"speed", "direction", "functionSet", "active"} {
	if strings.Contains(text, forbidden) { t.Fatalf("forbidden JSON field %q in %s", forbidden, text) }
}
if !strings.Contains(text, `"description":3`) { t.Fatalf("missing static function description: %s", text) }
```

- [ ] **Step 3: Run the focused backend tests and confirm failure**

Run:

```powershell
Set-Location backend
go test ./internal/application -run "TestECoS(CommandsExcludeRuntimeAndLayoutState|LiveCommandsOnlyMonitorBaseObject|ParserIgnoresRuntimeState)" -count=1
```

Expected: FAIL because command builders and DTOs still include excluded fields.

- [ ] **Step 4: Implement explicit production commands**

Define and use:

```go
const eCoSLocomotiveListCommand = "queryObjects(10, addr, name, protocol)"

func eCoSLocomotiveDetailCommand(objectID int) string {
	return fmt.Sprintf("get(%d, profile, protocol, name, addr, funcdesc)", objectID)
}

func eCoSLiveSubscriptionCommands() []string {
	return []string{"request(1, view)", "get(1, info, status)"}
}

func eCoSRawProbeFields() []string {
	return []string{"cv", "cvs", "cvlist", "functionmapping"}
}
```

Replace every duplicated locomotive list/detail string with these builders. Keep the existing targeted CV probes.

- [ ] **Step 5: Remove runtime and image DTO fields**

Reduce `ECoSFunction` to `Index` and `Description`. Remove speed, speed step, direction, function set, active state, and image candidates from both locomotive DTOs. Remove `ECoSImageCandidate` and image parsing helpers plus their now-unused imports (`encoding/base64`, `mime`, `net/http`, and `path/filepath`) when no longer referenced.

- [ ] **Step 6: Make the parser ignore excluded fields**

Keep raw attribute collection only for allowed keys. `applyECoSArgument` must handle `name`, `protocol`, `profile`, `addr`, and `funcdesc`; cases for `speed`, `speedstep`, `dir`, `funcset`, and `func` must be absent. Introduce an allowlist:

```go
var eCoSAllowedLocomotiveAttributes = map[string]struct{}{
	"name": {}, "protocol": {}, "profile": {}, "addr": {},
	"funcdesc": {}, "functionmapping": {}, "cv": {}, "cvs": {}, "cvlist": {},
}
```

Only keys in this set may enter the `Attributes` response map. Preserve `parseECoSCVValues` and function-description parsing.

- [ ] **Step 7: Run focused and full backend tests**

Run:

```powershell
gofmt -w internal/application/ecos.go internal/application/ecos_test.go
go test ./internal/application -run ECoS -count=1
go test ./...
Set-Location ..
```

Expected: all ECoS and backend tests pass.

- [ ] **Step 8: Commit the backend restriction**

```powershell
git add backend/internal/application/ecos.go backend/internal/application/ecos_test.go
git commit -m "fix: limit ECoS data and command scope"
```

### Task 2: Frontend Contract and State-Free Import Helpers

**Files:**
- Modify: `frontend/src/shared/api.ts`
- Modify: `frontend/src/features/importExport/importExportHelpers.test.tsx`
- Modify: `frontend/src/features/importExport/importExportHelpers.tsx`
- Modify: `frontend/src/features/importExport/ImportExportView.tsx`
- Modify: `frontend/src/features/vehicles/vehicleViewModel.ts`
- Modify: `frontend/src/features/vehicles/useVehicleECoSDraftController.ts`

**Interfaces:**
- Consumes: state-free backend `ECoSRawLocomotive` JSON from Task 1
- Produces: state-free TypeScript types, static function suggestions, CV-only value summaries, and draft payloads without ECoS image suggestions

- [ ] **Step 1: Add failing static-function tests**

Extend `importExportHelpers.test.tsx` imports with `ecosFunctionSuggestions` and add:

```tsx
it("creates ECoS function suggestions only from static descriptions", () => {
  const suggestions = ecosFunctionSuggestions({
    objectId: 1001,
    functions: [{ index: 0, description: 3 }]
  }, []);
  expect(suggestions).toHaveLength(1);
  expect(suggestions[0]).toMatchObject({ functionKey: "F0", active: undefined });
  expect(suggestions[0].notes).toBe("Aus ECoS funcdesc 3.");
});
```

Add a rendering assertion using `renderToStaticMarkup` that the raw probe output contains CV/function information but no active marker, speed, direction, or image section.

- [ ] **Step 2: Run the frontend helper test and confirm failure**

Run:

```powershell
Set-Location frontend
npm.cmd run test:run -- src/features/importExport/importExportHelpers.test.tsx
```

Expected: FAIL because active/image state is still part of suggestions and rendering.

- [ ] **Step 3: Narrow shared ECoS types**

Remove `ECoSImageCandidate`, `ECoSImageSuggestion`, and the `ECoSRawLocomotive` properties `speed`, `speedStep`, `direction`, `functionSet`, and `imageCandidates`. Change functions to:

```ts
functions?: Array<{
  index: number;
  description?: number;
}>;
```

- [ ] **Step 4: Remove image suggestions from draft interfaces and controllers**

Remove `imageSuggestions` from `ImportRow`, `ECoSVehicleDraftPayload`, and `buildECoSVehicleDraftRow`. In `ImportExportView.tsx`, stop placing it in the session payload. In `useVehicleECoSDraftController.ts`, remove the ECoS image conversion and make `mergeImages` return existing vehicle images unchanged:

```ts
const mergeImages = (current: PendingArticleImage[]) => current;
```

- [ ] **Step 5: Remove runtime and image helper behavior**

Delete `formatECoSDirection`, active-state formatting, function-set fallback, ECoS image formatting/suggestion helpers, and `renderECoSImagePreview`. Make `ecosFunctionSuggestions` filter only functions with a numeric description and return notes exactly as `Aus ECoS funcdesc <code>.` without an `active` property.

Keep runtime and image field names in a separate ignored-key set so they are not shown as “unknown” if an older cached probe contains them:

```ts
const ignoredECoSRuntimeAttributes = new Set([
  "dir", "func", "funcset", "speed", "speedstep",
  "icon", "image", "pic", "picture", "userimage"
]);
```

`rawECoSUnknownAttributes` must exclude both known allowed keys and ignored excluded keys.

- [ ] **Step 6: Simplify worklist and raw probe rendering**

Change the value summary to accept only `{ cvs: number }`, remove the image preview call, and render static functions using index and description only. Preserve CV review, matching, skip, dry-run preview, and confirmed write controls.

- [ ] **Step 7: Run helper tests and the TypeScript build**

Run:

```powershell
npm.cmd run test:run -- src/features/importExport/importExportHelpers.test.tsx
npm.cmd run build
Set-Location ..
```

Expected: the targeted tests and production build succeed without image or active-state type references.

- [ ] **Step 8: Commit the frontend contract restriction**

```powershell
git add frontend/src/shared/api.ts frontend/src/features/importExport/importExportHelpers.test.tsx frontend/src/features/importExport/importExportHelpers.tsx frontend/src/features/importExport/ImportExportView.tsx frontend/src/features/vehicles/vehicleViewModel.ts frontend/src/features/vehicles/useVehicleECoSDraftController.ts
git commit -m "fix: remove ECoS runtime state from UI"
```

### Task 3: ECoS UI Copy, API Description, and Public Scope

**Files:**
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/shared/i18n.test.ts`
- Modify: `openapi/railkeeper.yaml`
- Modify: `README.md`
- Modify: `README.de.md`
- Modify: `THIRD_PARTY_NOTICES.md`
- Modify: `CHANGELOG.md`

**Interfaces:**
- Consumes: narrowed backend/frontend behavior from Tasks 1 and 2
- Produces: user-facing German/English scope statements and API summaries matching production behavior

- [ ] **Step 1: Add translation parity checks for the new scope text**

Add keys in both locales for a scope note and ensure `i18n.test.ts` continues to assert identical key sets. The German scope text must say:

```text
RailKeeper liest Lokstammdaten, CV-Werte und statische Funktionstasten. Geschwindigkeit, Richtung, aktive Funktionszustände und Anlagenobjekte werden nicht überwacht.
```

The English scope text must say:

```text
RailKeeper reads locomotive master data, CV values and static function keys. Speed, direction, active function states and layout objects are not monitored.
```

- [ ] **Step 2: Remove excluded translation keys and image wording**

Remove control-state, speed, direction, active/inactive, ECoS image, and image-count keys. Change the raw probe note to mention CV and static function fields only. Change `worklist.valueSummary` to `{cvs} CVs` in both languages.

- [ ] **Step 3: Update OpenAPI endpoint summaries**

Describe `/ecos/locomotives/raw` as reading allowed locomotive master data, CVs, and static function definitions. Describe live start/status as base-station connection monitoring without locomotive or layout-state subscriptions. Keep the existing admin response documentation.

- [ ] **Step 4: Align READMEs, notices, and changelog**

In both READMEs state the retained read/write fields and excluded runtime/object-manager scope. In `THIRD_PARTY_NOTICES.md`, retain the independent-project disclaimer and explicitly identify the local function-symbol references for ESU review. In `CHANGELOG.md`, record the removal of runtime state, images, and non-locomotive managers.

- [ ] **Step 5: Run translation, scope, and build checks**

Run:

```powershell
Set-Location frontend
npm.cmd run test:run -- src/shared/i18n.test.ts src/features/importExport/importExportHelpers.test.tsx
npm.cmd run build
Set-Location ..
rg -n -i "queryObjects\(11|queryObjects\(26|request\(11|request\(26|speedstep|funcset|userimage" backend/internal/application frontend/src openapi/railkeeper.yaml
git diff --check
```

Expected: tests and build succeed. The search may find only negative-test fixtures or ignored compatibility keys, never production ECoS commands, API properties, or visible UI copy.

- [ ] **Step 6: Commit public ECoS scope documentation**

```powershell
git add frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/shared/i18n.test.ts openapi/railkeeper.yaml README.md README.de.md THIRD_PARTY_NOTICES.md CHANGELOG.md
git commit -m "docs: clarify limited ECoS integration"
```

### Task 4: Full Verification and ESU Letter Drafts

**Files:**
- Create: `docs/legal/esu-integration-review-request.de.md`
- Create: `docs/legal/esu-integration-review-request.en.md`

**Interfaces:**
- Consumes: verified production command set, AGPL notice files, public repository URL
- Produces: unsent German and English review-request drafts describing the exact integration

- [ ] **Step 1: Run full project verification**

Run:

```powershell
Set-Location backend
go test ./...
Set-Location ..\frontend
npm.cmd run test:run
npm.cmd run build
Set-Location ..
git diff --check
git status --short
```

Expected: all Go and frontend tests pass, the production build succeeds, whitespace check is clean, and status lists only intended letter files if earlier tasks were committed.

- [ ] **Step 2: Audit the actual ECoS command surface**

Run:

```powershell
rg -n "queryObjects\(|request\(|get\(|set\(" backend/internal/application/ecos.go backend/internal/ecos
rg -n -i "speed|speedstep|dir|funcset|queryObjects\(11|queryObjects\(26|request\(11|request\(26|icon|userimage" backend/internal/application/ecos.go frontend/src/features/importExport frontend/src/shared/api.ts
```

Expected: the first command shows only connection, locomotive list/detail/CV/static function reads and the name/address/protocol write path. The second command shows no production query or response property for excluded data; test/ignore references are documented separately.

- [ ] **Step 3: Draft the German ESU review request**

Create a send-ready but unsent letter addressed to ESU electronic solutions ulm GmbH & Co. KG. Include the repository URL, AGPL-3.0-only, exact allowed read fields, exact write fields, excluded commands/managers, local symbol origin, the independent-project disclaimer, and a request for concrete objections or required removals. Do not claim compatibility certification or permission.

- [ ] **Step 4: Draft the English ESU review request**

Create a faithful English equivalent with the same field lists, limitations, and willingness to change or remove objected components. Mark both documents `Draft, not sent` at the top.

- [ ] **Step 5: Cross-check both letters against production code**

Run:

```powershell
rg -n "name|address|protocol|CV|funcdesc|speed|direction|S88|booster|switch|route|AGPL-3.0-only|Draft" docs/legal/esu-integration-review-request.de.md docs/legal/esu-integration-review-request.en.md
git diff --check
```

Expected: both letters contain the same retained and excluded capabilities and are clearly marked as drafts.

- [ ] **Step 6: Commit the letter drafts**

```powershell
git add docs/legal/esu-integration-review-request.de.md docs/legal/esu-integration-review-request.en.md
git commit -m "docs: draft ESU integration review request"
```
