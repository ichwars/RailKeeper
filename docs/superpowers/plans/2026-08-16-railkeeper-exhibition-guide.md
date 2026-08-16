# RailKeeper Exhibition Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `aegis:executing-plans` to implement this plan
> task by task. Keep all Git mutations with the coordinating agent.

**Goal:** Publish a complete three-page English and German user-guide section for the stable
RailKeeper exhibition workspace.

**Architecture:** Reuse the existing `exhibition` coverage owner and VitePress user-guide
structure. Add one bilingual overview pair as the coverage destination plus two mirrored subpage
pairs for list administration and entry operation. Keep every behavior claim pinned to tag
`v0.1.17.6` and link only to already published documentation owners.

**Tech Stack:** Markdown, JSON, TypeScript VitePress configuration, Node.js documentation tests,
VitePress 2.0.0-alpha.19, Git, GitHub Actions, GitHub Pages.

**Baseline/Authority Refs:**

- `AGENTS.md`;
- `docs/superpowers/specs/2026-08-16-railkeeper-exhibition-guide-design.md`;
- `docs/superpowers/plans/2026-08-15-railkeeper-documentation-foundation.md`;
- `docs/coverage.json`;
- tag `v0.1.17.6` and the stable source files named under each task;
- existing English/German guide pages and `docs/.vitepress/config.mts`.

**Compatibility Boundary:** Documentation only. Do not change runtime behavior, frontend or
backend source, API contracts, schemas, migrations, dependencies, screenshots, generated output,
or coverage ownership outside the existing `exhibition` topic. Later `main` behavior is excluded.

**TDD Route:**

- Mode: off
- Decision: skipped
- Strict authority: not applicable
- Test posture: post-change documentation regression and stable-source audit
- Reason: the approved change creates documentation and navigation, not executable product logic
- Verification: intentional coverage failure, focused validation, full documentation check,
  bilingual review, source-fidelity review, pull-request checks, and live-page checks

**Verification:** `node scripts/validate-docs.mjs`, `npm.cmd run check`, `git diff --check`, focused
metadata/link/line-length scans, stable-tag source audit, pull-request checks, and post-merge HTTP
checks for every English and German page.

## Plan basis and readiness

**Aegis Visibility:** Planning is useful because six new pages must remain semantically paired
while describing Messe and Admin roles, list locks, duplicate addresses, entry persistence, and print
behavior without leaking planned administration or later-development behavior.

**BaselineUsageDraft:**

- Required baseline refs: approved design, documentation foundation, coverage contract,
  `v0.1.17.6`, current VitePress configuration
- Acknowledged before plan refs: all required refs above
- Cited in plan refs: all required refs above
- Missing refs: none
- Decision: continue

**Requirement Ready Check:**

- Requirement source refs: approved exhibition-guide design
- Goals and scope refs: design Objective, Scope, Non-goals, Information architecture
- User/scenario refs: Messe operation, Admin list management, entry maintenance, printing
- Requirement item refs: roles, locks, validation, persistence, errors, navigation, verification
- Acceptance/verification criteria refs: design Verification and Expected outcome
- Open blocker questions: none
- Decision: ready

**Change Necessity:**

- User-visible need: complete bilingual exhibition documentation that users can navigate
- No-change/non-code option: insufficient because the stable workflows are absent from the site
- Why code change is necessary: no product-code change is necessary
- Minimum change boundary: Markdown pages, coverage JSON, VitePress sidebar, and published links
- Decision: docs/config-only

**Existence Check:**

- Proposed new surface: three English and three German pages under `guide/exhibition`
- Existing owner/reuse candidate: existing `exhibition` topic and VitePress user-guide section
- Why existing surface is insufficient: its configured destination pair does not exist
- Creation proof: approved three-page information architecture and current `planned` coverage state
- Entropy/retirement impact: one bounded topic group, no duplicate runtime or documentation owner
- Decision: add-with-proof

**Architecture Integrity Lens:**

- Invariant: one coverage owner for all stable exhibition-list workflows
- Canonical owner/contract: `exhibition` in `docs/coverage.json`
- Responsibility overlap: role administration, master-data administration, vehicle editing,
  backup operation, deployment, and layouts remain with their existing owners
- Higher-level simplification: use one overview hub instead of one oversized page or role copies
- Retirement/falsifier: no temporary path, fallback, or duplicate owner is introduced
- Verdict: proceed

**Plan Pressure Test:**

- Owner/contract/retirement: one existing owner, no retirement track required
- Architecture integrity/higher-level path: three task-oriented pages fit the approved structure
- Verification scope: stable source, bilingual parity, coverage, build, PR, and live deployment
- Task executability: each page pair and navigation slice has exact files, headings, and checks
- Pressure result: proceed

**Complexity Budget:**

- Artifact class: maintained user documentation
- Target files/artifacts: six new pages, coverage JSON, VitePress config, four existing pages
- Current pressure: roles, entry fields, locks, and printing are unsuitable for one page
- Projected post-change pressure: three focused pages with mirrored sections
- Budget result: within-budget
- Planned governance: keep each page task-focused and avoid repeating field and permission tables

**Plan-Time Complexity Check:**

- Target files: new exhibition pages plus narrow navigation and related-link owners
- Existing size/shape signals: vehicle and accessories chapters are long but focused; config is
  centralized
- Owner fit: new content belongs under the existing `exhibition` topic
- Add-in-place risk: one-page design would obscure role and lock boundaries
- Better file boundary: overview, list/locking, and entry/printing page pairs
- Recommendation: add owner files and keep existing-file edits wiring-only

## Execution Readiness View

- Intent Lock: document every stable user-facing exhibition workflow in English and German
- Scope Fence: documentation and navigation only, no product or API changes
- Baseline Lock: behavior claims must resolve to tag `v0.1.17.6`
- Approved Behavior: three-page task-oriented structure and role model from the corrected design
- Owner/Contract Constraints: promote only `exhibition`; preserve all other coverage owners
- Compatibility Boundary: do not publish planned administration, layouts, or later `main`
- Retirement Boundary: no old path, adapter, fallback, or compatibility layer
- Task Batches: overview, list administration, entry operation, navigation, closeout
- Test Obligations: intentional gate failure, pair parity, coverage validation, VitePress build
- Review Gates: source fidelity, language parity, roles, locks, persistence, validation, printing
- Drift/Rewind Rules: return to the design if stable behavior requires another owner or more pages
- Evidence Required Before Completion: clean diff, full docs check, reviewed head SHA, green PR
  checks, merged PR, green merge workflows, and six successful live routes
- Advisory Boundary: method-pack execution guidance only, not completion authority

## Global constraints

- Work only in `.worktrees/docs-user-guide-exhibition` on
  `dev/docs-user-guide-exhibition`. Do not modify or push local `main`.
- Preserve the user's untracked files in the main checkout.
- Use stable release `v0.1.17.6` and review date `2026-08-16` on every new page.
- Keep English and German pages in the same semantic section order.
- Use exact UI labels from the matching stable locale.
- State whether each operation changes dialog state, writes immediately, or only reads data.
- Treat Messe as the isolated operating role and Admin as the direct list-management role.
- State confirmations, lock effects, failed-refresh behavior, and recovery precisely.
- Do not document role assignment, master-data administration, backup operation, deployment, or
  the layout workspace as published in this chapter.
- Link only to published pages. Do not activate planned Settings, administration, layout, backup,
  deployment, or reference destinations.
- Preserve LF, UTF-8, approximately 100 to 120 character lines, and established documentation
  style. Do not use em dashes or unfinished-content markers.
- Do not commit `docs/node_modules`, `docs/.vitepress/dist`, caches, or generated output.
- Make one scoped commit after each coherent task passes its verification.
- Merge only after the exact reviewed head has all required GitHub checks green.

## File map

| File | Responsibility |
| --- | --- |
| `docs/coverage.json` | Promote the existing exhibition topic. |
| `docs/site/guide/exhibition/index.md` | English workspace overview and operating sequence. |
| `docs/site/de/guide/exhibition/index.md` | German overview counterpart. |
| `docs/site/guide/exhibition/lists-and-locking.md` | English list administration workflow. |
| `docs/site/de/guide/exhibition/lists-and-locking.md` | German list counterpart. |
| `docs/site/guide/exhibition/entries-and-printing.md` | English entry and print workflow. |
| `docs/site/de/guide/exhibition/entries-and-printing.md` | German entry counterpart. |
| `docs/.vitepress/config.mts` | Add mirrored Exhibition/Messe sidebar groups. |
| `docs/site/guide/index.md` | Add the English guide transition. |
| `docs/site/de/guide/index.md` | Add the German guide transition. |
| `docs/site/guide/vehicles/index.md` | Link the published exhibition workflow at its boundary. |
| `docs/site/de/guide/vehicles/index.md` | Add the German boundary link. |

---

### Task 1: Publish the coverage destination and exhibition overview

**Files:**

- Modify: `docs/coverage.json`
- Create: `docs/site/guide/exhibition/index.md`
- Create: `docs/site/de/guide/exhibition/index.md`

**Why:** Messe users need a reliable entry point that explains the isolated workspace and the
complete operating sequence before they modify an open list.

**Change Necessity:** The destination pair does not exist. The minimum boundary is the existing
coverage status plus its exact English/German overview paths.

**Impact/Compatibility:** No ownership changes. `exhibition` becomes `documented` only after both
configured paths exist and pass validation.

**Verification:** Intentional validation failure before page creation, then successful focused
validation, `git diff --check`, and a clean task-owned status before commit.

**Stable sources:**

- `frontend/src/app/App.tsx`, `Shell.tsx`, and their role tests;
- `frontend/src/features/exhibition/ExhibitionView.tsx` and tests;
- exhibition translations in `frontend/src/shared/i18n/en.ts` and `de.ts`;
- `backend/internal/api/routes.go` and stable exhibition authorization tests.

- [ ] **Step 1: Capture the implementation start snapshot and audit the stable overview**

Run from the worktree root:

```powershell
git status --short --branch
git rev-parse HEAD
git show v0.1.17.6:frontend/src/app/App.tsx
git show v0.1.17.6:frontend/src/app/Shell.tsx
git show v0.1.17.6:frontend/src/features/exhibition/ExhibitionView.tsx
git grep -n '"exhibition\.' v0.1.17.6 -- frontend/src/shared/i18n/en.ts `
  frontend/src/shared/i18n/de.ts
```

Confirm before writing:

- the route and navigation require the Messe role;
- list management controls additionally require Admin;
- the list table shows designation, date, entry count, status, and actions;
- initial list sort is date descending and stable columns can reverse direction;
- selecting a row controls the adjacent entry table;
- view and print are available without Admin;
- entry creation is disabled when no list is selected or the selected list is locked;
- pure Messe operation remains isolated from general inventory master data;
- loading, no-list, no-selection, no-entry, and request-error states are visible.

- [ ] **Step 2: Turn the coverage topic into an intentional failing gate**

Change only the `exhibition` topic status:

```json
"status": "documented"
```

Run from `docs`:

```powershell
node scripts/validate-docs.mjs
```

Expected: failure names both missing `guide/exhibition/index.md` paths. An ownership, schema, or
unrelated failure is unexpected and must be resolved before continuing.

- [ ] **Step 3: Create the English overview page**

Use this exact frontmatter and heading order:

```markdown
---
title: Exhibition workspace
description: Prepare, maintain, review, and print exhibition lists safely.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Exhibition workspace

## Access rights and isolation

## Understand lists and entries

## Follow the exhibition workflow

## Work with an open or locked list

## Read loading, empty, and error states

## Continue with a focused workflow

## Related pages

## Documented RailKeeper version
```

Include the confirmed Messe/Admin role split, two-panel relationship, sort defaults, action visibility,
lock boundary, empty states, immediate-write boundary, and transitions to the two subpages. Keep
field-level and list-admin detail on their focused pages.

- [ ] **Step 4: Create the German semantic counterpart**

Use matching structure and this frontmatter:

```markdown
---
title: Messearbeitsbereich
description: Messelisten sicher vorbereiten, pflegen, prüfen und drucken.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Messearbeitsbereich
```

Use exact German stable labels such as **Messeliste**, **offen**, **gesperrt**, **Ansehen**, and
**Liste drucken**. Preserve the same role, persistence, error, and recovery statements.

- [ ] **Step 5: Verify and commit Task 1**

Run:

```powershell
cd docs
node scripts/validate-docs.mjs
cd ..
git diff --check
git status --short
```

Expected: validation succeeds and only the three Task 1 files are uncommitted. Commit:

```powershell
git add docs/coverage.json docs/site/guide/exhibition/index.md `
  docs/site/de/guide/exhibition/index.md
git commit -m "docs: add exhibition workspace guide"
```

---

### Task 2: Document list administration and locking

**Files:**

- Create: `docs/site/guide/exhibition/lists-and-locking.md`
- Create: `docs/site/de/guide/exhibition/lists-and-locking.md`

**Why:** Admin accounts need a precise preparation workflow whose lock and
deletion effects remain predictable for operators.

**Change Necessity:** The overview is intentionally a hub. This focused pair prevents Admin-only
actions and lock effects from being hidden in the entry instructions.

**Impact/Compatibility:** The pair stays under the exhibition owner. It explains role requirements
but does not document how roles are assigned.

**Verification:** Focused validation, mirrored heading scan, stable-source readback, whitespace
check, and one scoped commit.

**Stable sources:**

- list handlers in `ExhibitionView.tsx`;
- `backend/internal/application/exhibition.go` list methods and tests;
- `backend/internal/api/exhibition.go`, route definitions, audit tests, and translations;
- migrations `0018_exhibition_lists.sql` and stable backup coverage.

- [ ] **Step 1: Audit list validation, permissions, and consequences**

Use `git show` and `git grep` on tag `v0.1.17.6` to confirm:

- designation and date are required and normalized by the stable service;
- new lists default to the browser-generated UTC calendar date; users must verify it near local
  midnight;
- create, edit, lock, unlock, and list deletion require Admin on the server;
- normal workspace access still requires Messe in the stable UI;
- locking preserves reading, viewing, and printing but rejects entry mutations;
- unlocking restores entry mutation;
- list deletion uses the exact confirmation and removes its entries through the stable database
  relationship;
- successful create, update, lock, unlock, and delete operations write audit records;
- missing, invalid, locked, forbidden, and reload-failure behavior is described exactly.

- [ ] **Step 2: Create the English list page**

Use this exact structure:

```markdown
---
title: Lists and locking
description: Create exhibition lists, control their lock state, and remove them safely.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Lists and locking

## Roles for operation and administration

## Create a list

## Select, sort, and inspect lists

## Edit designation or date

## Lock or unlock a list

## Delete a list

## Resolve list errors

## Related pages

## Documented RailKeeper version
```

State exact actions and confirmations, immediate persistence, cascade consequences, lock effects,
audit ownership at a user-relevant level, and reload-before-retry guidance.

- [ ] **Step 3: Create the German semantic counterpart**

Use matching headings and this frontmatter:

```markdown
---
title: Listen und Sperren
description: Messelisten anlegen, sperren, entsperren und sicher löschen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Listen und Sperren
```

Use the exact stable labels **Neue Liste**, **Messeliste anlegen**, **Messeliste bearbeiten**,
**Sperren**, **Entsperren**, **Löschen**, and **Speichern**.

- [ ] **Step 4: Verify and commit Task 2**

Run:

```powershell
cd docs
node scripts/validate-docs.mjs
cd ..
rg -n "^## " docs/site/guide/exhibition/lists-and-locking.md
rg -n "^## " docs/site/de/guide/exhibition/lists-and-locking.md
git diff --check
```

Expected: validation succeeds and section order is mirrored. Commit:

```powershell
git add docs/site/guide/exhibition/lists-and-locking.md `
  docs/site/de/guide/exhibition/lists-and-locking.md
git commit -m "docs: explain exhibition list management"
```

---

### Task 3: Document entries, conflicts, images, functions, and printing

**Files:**

- Create: `docs/site/guide/exhibition/entries-and-printing.md`
- Create: `docs/site/de/guide/exhibition/entries-and-printing.md`

**Why:** Entry changes combine required identity, operating data, address uniqueness, images,
function mappings, lock state, and print output. Operators must understand the complete write.

**Change Necessity:** The approved overview cannot carry the field and print reference without
becoming difficult to use. This is the minimum focused boundary.

**Impact/Compatibility:** No product behavior changes. Master data and function symbols are
consumed resources only; their administration stays outside this chapter.

**Verification:** Stable-field audit, mirrored page scan, focused validation, whitespace check,
and one scoped commit.

**Stable sources:**

- entry form, duplicate detection, image helpers, function serialization, view, and print code in
  `ExhibitionView.tsx`;
- `FunctionSymbolPicker`, function-symbol helpers, `AppSelect`, and stable API types;
- `backend/internal/application/exhibition.go` entry methods, scanners, and tests;
- entry handlers, route access, migrations 0027, 0028, and 0036, translations, and backup tests.

- [ ] **Step 1: Audit every stable entry and print rule**

Confirm from `v0.1.17.6`:

- required fields are owner and locomotive designation;
- optional general fields are manufacturer, type, class, epoch, railway company, DCC address, SX
  address, decoder type, adapter/interface, analog state, day scope, and notes;
- day scope normalizes no selection or all four individual days to `all`;
- DCC and SX conflicts compare trimmed lowercase values inside the selected list and exclude the
  edited entry;
- a conflict blocks saving and names the conflicting locomotive;
- pure Messe lacks general inventory master-data choices while combined inventory roles can load
  them; the guide must state the actual free-text/select behavior of `AppSelect`;
- one image source can be uploaded as a data URL or entered as a link; replacement and removal are
  dialog changes until **Save** succeeds;
- functions cover F0 through F31, F0 defaults to light, and only configured functions are stored
  and displayed;
- earlier plain-text function values are parsed when stable code supports them;
- entry create/update require an open list; entry deletion additionally requires Admin and
  confirmation;
- initial entry sort is owner ascending; sortable keys are owner, locomotive name, decoder number,
  and function keys;
- detail view and report fetch current list entries before display;
- report output is A4 landscape and optionally includes images, with browser print handling;
- save, stale list, lock, image-read, detail-load, and print-load failures have safe recovery text.

- [ ] **Step 2: Create the English entry page**

Use this exact structure:

```markdown
---
title: Entries and printing
description: Record exhibition locomotives, resolve address conflicts, and print the operating list.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Entries and printing

## Entry rights and persistence

## Create or edit an entry

## Complete general and control data

## Select exhibition days

## Resolve DCC and SX address conflicts

## Add or remove the image

## Configure function keys F0 through F31

## Read, sort, and delete entries

## View and print the report

## Resolve entry and print errors

## Related pages

## Documented RailKeeper version
```

Use field, role, and print-output tables where they improve scanning. State draft versus immediate
persistence and explain that browser or printer settings can change final output.

- [ ] **Step 3: Create the German semantic counterpart**

Use matching structure and this frontmatter:

```markdown
---
title: Einträge und Drucken
description: Messeloks erfassen, Adresskonflikte lösen und die Betriebsliste drucken.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Einträge und Drucken
```

Use exact stable labels for **Allgemein**, **Bilder upload**, **Funktionstasten**, **Adresse DCC**,
**Adresse SX**, **Mit Bildern drucken**, and all save/delete confirmations.

- [ ] **Step 4: Verify and commit Task 3**

Run the focused validator, mirrored heading scan, and `git diff --check`. Expected: all pass and
only the pair is uncommitted. Commit:

```powershell
git add docs/site/guide/exhibition/entries-and-printing.md `
  docs/site/de/guide/exhibition/entries-and-printing.md
git commit -m "docs: explain exhibition entries and printing"
```

---

### Task 4: Connect navigation and published cross-links

**Files:**

- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/site/guide/vehicles/index.md`
- Modify: `docs/site/de/guide/vehicles/index.md`
- Modify: all six new exhibition pages for final related links

**Why:** Published content must be discoverable from the site hierarchy and the existing vehicle
exhibition boundary without exposing planned destinations.

**Change Necessity:** The pages can build without navigation but would remain easy to overlook,
contradicting the approved wiki goal.

**Impact/Compatibility:** Wiring only. Existing routes, titles, and metadata remain unchanged
except for narrow related-page additions.

**Verification:** Full documentation check, route scan for all six pages, link review, whitespace
check, and one scoped commit.

- [ ] **Step 1: Add mirrored sidebar groups after Accessories**

Add English **Exhibition** links in this order:

```text
/guide/exhibition/
/guide/exhibition/lists-and-locking
/guide/exhibition/entries-and-printing
```

Add German **Messe** links in the same position and order under `/de/guide/exhibition/...`. Use the
approved page titles as sidebar labels.

- [ ] **Step 2: Add landing and vehicle-boundary transitions**

Add one concise workflow paragraph to each User Guide landing page after Accessories. It must link
to the new overview and name list preparation, operating entries, locks, address conflicts, and
printing without duplicating the subpages.

Add the exhibition overview to the related links of both vehicle core pages. The nearby boundary
text already explains the vehicle **Exhibition** switch; the new link must lead to operational list
handling without changing that existing behavior claim.

- [ ] **Step 3: Finalize related links on all exhibition pages**

Every exhibition page links to:

- User Guide overview;
- Exhibition workspace;
- the other two exhibition pages where it is not the current page.

The overview and entry page may link to Vehicle inventory only to explain the isolation boundary.
Do not link to planned role administration, master data, backup/restore, Settings, layouts,
deployment, or reference pages.

- [ ] **Step 4: Verify navigation and commit Task 4**

Run:

```powershell
cd docs
npm.cmd run check
cd ..
git diff --check
rg -n "/(de/)?guide/exhibition" docs/.vitepress/config.mts docs/site
git status --short
```

Expected: 19 tests pass, coverage validation succeeds, VitePress builds every route, and all
English/German paths appear. Commit only Task 4 files:

```powershell
git add docs/.vitepress/config.mts docs/site/guide/index.md docs/site/de/guide/index.md `
  docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md `
  docs/site/guide/exhibition docs/site/de/guide/exhibition
git commit -m "docs: publish exhibition user guide"
```

---

### Task 5: Prove source fidelity, parity, and publication readiness

**Files:**

- Review: all task-owned files
- Modify only when a verified finding requires correction

**Why:** A polished guide is unsafe if it misstates roles, lock effects, required fields,
address conflicts, persistence, deletion, or print behavior.

**Change Necessity:** Review edits are allowed only for evidence-backed documentation defects and
must remain inside the approved file boundary.

**Impact/Compatibility:** No scope expansion. Any discovery owned by another coverage topic returns
to design review instead of being absorbed.

**Verification:** Complete stable-source audit, bilingual and hygiene scans, fresh documentation
check, focused independent review, exact-head PR checks, merge workflows, and six live HTTP checks.

- [ ] **Step 1: Run the complete stable-source audit**

Check each claim against `v0.1.17.6` using `git show` and `git grep` over:

```text
frontend/src/app/App.tsx
frontend/src/app/Shell.tsx
frontend/src/features/exhibition/ExhibitionView.tsx
frontend/src/features/exhibition/ExhibitionView.test.tsx
frontend/src/shared/api.ts and related adapters/types
frontend/src/shared/i18n/en.ts
frontend/src/shared/i18n/de.ts
frontend/src/shared/ui/AppSelect.tsx
frontend/src/shared/functionSymbols.tsx
backend/internal/api/exhibition.go
backend/internal/api/routes.go
backend/internal/application/exhibition.go
backend/internal/application/exhibition_test.go
backend migrations 0018, 0027, 0028, 0029, and 0036
backend backup implementation/tests and stable OpenAPI paths
```

Audit route access, roles, list fields, defaults, sorting, lock behavior, deletion, audit events,
entry fields, day normalization, master-data isolation, duplicate comparison, images, functions,
view, printing, persistence, backup, partial refresh, and recovery. Correct every unsupported or
incomplete statement.

- [ ] **Step 2: Run bilingual and hygiene scans**

Run:

```powershell
rg -n -i "\b(todo|tbd|placeholder|coming soon|incomplete)\b|noch zu dokumentieren|platzhalter" `
  docs/site/guide/exhibition docs/site/de/guide/exhibition
rg -n "^## " docs/site/guide/exhibition docs/site/de/guide/exhibition
rg -n "reviewedVersion:|lastReviewed:|status:|audience:" `
  docs/site/guide/exhibition docs/site/de/guide/exhibition
git diff --check origin/main...HEAD
```

Expected: no unfinished markers; paired pages have matching semantic heading counts; all metadata
uses user/stable/0.1.17.6/2026-08-16; whitespace is clean.

Check non-table prose lines over 120 characters and wrap them:

```powershell
$files = Get-ChildItem docs/site/guide/exhibition,docs/site/de/guide/exhibition -Filter *.md
foreach ($file in $files) {
  $lineNumber = 0
  Get-Content -LiteralPath $file.FullName | ForEach-Object {
    $lineNumber++
    if ($_.Length -gt 120 -and -not $_.StartsWith('|')) {
      "{0}:{1}:{2}" -f $file.FullName,$lineNumber,$_.Length
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
check is clean, and the task worktree contains no uncommitted changes.

- [ ] **Step 4: Request focused review and resolve findings**

Use `aegis:requesting-code-review` for a read-only review of `origin/main...HEAD`. Require answers:

1. Does every behavior claim match `v0.1.17.6` rather than later `main`?
2. Are English and German semantically equivalent and UI labels exact?
3. Is Messe versus Admin authority exact in UI and server routes?
4. Are list locks applied to create, update, and delete entry operations correctly?
5. Are required fields, day normalization, and duplicate address comparisons exact?
6. Are image and function edits correctly described as form state until entry save?
7. Are list and entry writes, refresh failures, and retry guidance safe?
8. Are view and print data, image option, layout, and browser-print behavior exact?
9. Are planned-administration, vehicle, backup, and layout boundaries preserved?

Correct valid findings, rerun focused checks, and obtain a clean re-review. Commit corrections only
when files changed:

```powershell
git add docs/coverage.json docs/.vitepress/config.mts `
  docs/site/guide/exhibition docs/site/de/guide/exhibition `
  docs/site/guide/index.md docs/site/de/guide/index.md `
  docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md
git commit -m "docs: refine exhibition user guide"
```

- [ ] **Step 5: Push, open the PR, and verify the reviewed head**

Run:

```powershell
git push -u origin dev/docs-user-guide-exhibition
gh pr create --base main --head dev/docs-user-guide-exhibition `
  --title "docs: add bilingual exhibition guide" `
  --body "Publishes the complete stable v0.1.17.6 exhibition guide in English and German."
gh pr checks --watch
```

Record `git rev-parse HEAD`. Do not merge if a check is failing, pending, or ran against another
SHA.

- [ ] **Step 6: Merge only when green and verify publication**

After all required checks pass for the reviewed head:

```powershell
gh pr merge --merge
$mergeSha = gh pr view --json mergeCommit --jq '.mergeCommit.oid'
gh run list --commit $mergeSha --limit 20 `
  --json databaseId,workflowName,status,conclusion,url
```

Wait for merge-triggered CI, CodeQL, Trivy, Documentation Pages, and Docker Image. All must finish
successfully. After Documentation Pages succeeds, verify these routes return HTTP 200 and contain
`v0.1.17.6`:

```text
https://ichwars.github.io/RailKeeper/guide/exhibition/
https://ichwars.github.io/RailKeeper/guide/exhibition/lists-and-locking
https://ichwars.github.io/RailKeeper/guide/exhibition/entries-and-printing
https://ichwars.github.io/RailKeeper/de/guide/exhibition/
https://ichwars.github.io/RailKeeper/de/guide/exhibition/lists-and-locking
https://ichwars.github.io/RailKeeper/de/guide/exhibition/entries-and-printing
```

Preserve the completed worktree for traceability unless the user later requests cleanup. Local
`main` and its user-owned untracked files remain untouched.

## Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| Later `main` behavior leaks into stable docs | Resolve every claim against `v0.1.17.6`. |
| Admin is incorrectly described as needing Messe | State Admin's stable direct UI and server access. |
| Locked lists appear partially writable | Audit all entry mutations in UI and application service. |
| Pure Messe appears to access general inventory | Explain the isolated master-data boundary exactly. |
| Duplicate warnings are described as advisory | Verify and state their save-blocking behavior. |
| Image or function edits appear immediately stored | Separate form state from successful entry save. |
| Browser print output is overpromised | Document report layout and browser/printer variability. |
| Planned administration pages become visible | Link only to existing published destinations. |
| Generated output enters Git | Inspect status and stage explicit task-owned paths only. |
| PR merges a stale review | Record head SHA and require checks on that exact head. |

## Retirement and rollback

No runtime path, adapter, compatibility layer, or old documentation owner is introduced. Rollback
is a normal revert of the documentation commits. Reverting must restore `exhibition` to `planned`,
remove all six pages and sidebar/landing links together, and leave other coverage owners intact.

## Expected completion evidence

- `exhibition` is `documented` with its existing overview paths;
- all three English/German page pairs exist with stable metadata and mirrored content;
- navigation and published cross-links reach every page;
- every stable claim has a `v0.1.17.6` source owner;
- local documentation checks pass from a clean task worktree;
- focused review has no unresolved findings;
- the exact reviewed PR head has all required checks green;
- the PR is merged and all six live pages return HTTP 200;
- local `main` and user-owned untracked files remain unchanged.

## Execution Route

- Decision: inline
- Evidence: the tasks share one documentation owner and several pages are linked by later wiring;
  current policy does not authorize subagent delegation for this request
- Fallback: pause at the failing task boundary, preserve the worktree, and report exact evidence
- User confirmation required: no
