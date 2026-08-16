# RailKeeper Vehicle Search and Spare Parts Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a complete English and German stable user-guide chapter for vehicle article
search, barcode/EAN search, web-document import, attachment extraction, and spare-part workflows.

**Architecture:** Add one paired VitePress chapter and promote its existing coverage topic from
`planned` to `documented`. Keep the chapter workflow-first, connect it only to already published
user-guide owners, and verify every behavior claim against tag `v0.1.17.6` before publication.

**Tech Stack:** Markdown, JSON, TypeScript VitePress configuration, Node.js documentation tests,
VitePress 2.0.0-alpha.19, GitHub Actions, GitHub Pages.

## Global Constraints

- Document stable RailKeeper `v0.1.17.6`, never later behavior from `main`.
- Use review date `2026-08-16` and the established user-guide frontmatter contract.
- Keep English and German pages in the same semantic section order with equivalent safety guidance.
- Treat search results, prices, availability, documents, and extracted values as untrusted
  suggestions until the user verifies them.
- Clearly separate transient search state, unsaved vehicle-editor state, immediate server writes,
  and saved vehicle data.
- State every sequential, non-atomic workflow and its reload, compare, retry recovery sequence.
- Link only to published pages. Do not activate links to planned settings, master-data, accessory,
  administration, or digital-center chapters.
- Do not change runtime behavior, APIs, schemas, dependencies, screenshots, or generated output.
- Work only in `.worktrees/docs-user-guide-search-spares` on
  `dev/docs-user-guide-search-spares`. Do not modify or push local `main`.
- Preserve LF, UTF-8, approximately 100 to 120 character lines, and the existing documentation
  style. Do not use em dashes or unfinished-content markers.
- Merge only after local checks, independent review, and every required GitHub check pass for the
  exact reviewed head SHA.

---

## File map

| File | Responsibility |
| --- | --- |
| `docs/coverage.json` | Promote `vehicle-search-spares` to a published bilingual topic. |
| `docs/site/guide/vehicles/search-and-spares.md` | English workflow and recovery reference. |
| `docs/site/de/guide/vehicles/search-and-spares.md` | German semantic counterpart. |
| `docs/.vitepress/config.mts` | Add both routes after the decoder/CV chapter. |
| `docs/site/guide/index.md` | Add the English User Guide landing transition. |
| `docs/site/de/guide/index.md` | Add the German User Guide landing transition. |
| `docs/site/guide/vehicles/index.md` | Link from vehicle core data to search and spare parts. |
| `docs/site/de/guide/vehicles/index.md` | German core-vehicle link. |
| `docs/site/guide/vehicles/media.md` | Replace the specialist boundary with a published transition. |
| `docs/site/de/guide/vehicles/media.md` | German media transition. |
| `docs/site/guide/vehicles/maintenance.md` | Add the published related-page link. |
| `docs/site/de/guide/vehicles/maintenance.md` | German maintenance link. |
| `docs/site/guide/vehicles/decoder-cv.md` | Add the published related-page link. |
| `docs/site/de/guide/vehicles/decoder-cv.md` | German decoder/CV link. |
| `docs/aegis/work/2026-08-16-vehicle-search-spares-guide/*` | Maintain execution and evidence state. |

---

### Task 1: Publish the coverage contract and bilingual page pair

**Files:**
- Modify: `docs/coverage.json`
- Create: `docs/site/guide/vehicles/search-and-spares.md`
- Create: `docs/site/de/guide/vehicles/search-and-spares.md`
- Reference: `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-search-spares-guide-design.md`

**Interfaces:**
- Consumes: existing topic ID `vehicle-search-spares`, its assigned frontend/backend ownership,
  and the stable-source facts in the approved specification.
- Produces: the validated routes `/guide/vehicles/search-and-spares` and
  `/de/guide/vehicles/search-and-spares` for Task 2.

- [ ] **Step 1: Turn the existing coverage topic into an intentional failing gate**

Change only this field on the `vehicle-search-spares` topic:

```json
"status": "documented"
```

Run from `docs`:

```powershell
node scripts/validate-docs.mjs
```

Expected: validation fails because both configured Markdown pages are absent. The failure must name
the `vehicle-search-spares` paths. Any ownership or schema error is unexpected and must be fixed
before continuing.

- [ ] **Step 2: Create the English page with the exact contract and section order**

Create `docs/site/guide/vehicles/search-and-spares.md` with this frontmatter and heading structure:

```markdown
---
title: Article search, web documents, and spare parts
description: Verify external article data, import web documents, and maintain vehicle spare parts.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Article search, web documents, and spare parts

## Prerequisites, settings, and access rights

## Search for article data

## Search by barcode or camera

## Review and apply one result

## Find and import web documents

## Extract spare parts from an attachment

## Maintain spare parts manually

## Look up one stored spare part

## Import a Piko or Roco manufacturer overview

## Protect data during multi-step actions

## Resolve empty, partial, and error states

## Related pages

## Documented RailKeeper version
```

Write complete operational prose under those headings. The page must include every item below,
using the exact stable values shown:

```text
Prerequisites and roles
- Article data web search is enabled by default and its enabled flag and source choices are local
  to the current browser.
- Default sources: manufacturer sites, Modellbahn-Fokus catalogs, dealer sites, and general web;
  Modellbau Wiki is optional.
- Normal search requires manufacturer, gauge, and article number or designation.
- EAN-only search is the exception.
- Admin and Editor can write. Viewer and Planner inspect stored data. Messe is excluded.
- The Viewer-level search API does not make the read-only vehicle UI editable.

Article and barcode search
- Model-tab actions are Search barcode and Search article data.
- Search sends the stable identity and non-empty searchable fields, uses a ten-second service
  timeout, deduplicates by URL, sorts by score, and returns at most ten results.
- Result review covers final query, active source classes, preferred domains, planned queries,
  score, snippet, detail-page state, source, images, field status, and conflicts.
- Cover all four field groups from the specification without inventing stable fields.
- Camera requires HTTPS or localhost, browser support, and permission. It requests the
  environment-facing camera, strips non-digits, requires at least eight digits, fills the field,
  and never auto-submits.
- Manual input and keyboard scanners remain the fallback.

Review, apply, and vehicle persistence
- Only fields whose current value is empty start selected. Equal/conflicting values do not.
- Images never start selected. Failed previews disappear from selection.
- Apply selected fields updates one local editor draft only. Boolean text is normalized.
- Selected images become pending remote images and the first can become the main candidate.
- Create or Save changes is required. Closing or reloading first loses draft changes.
- Vehicle save writes the core vehicle first, then remote images sequentially.
- A later image failure keeps the vehicle and earlier images. Reload, compare, retry only missing
  images.

Web documents
- Found documents requires a saved vehicle and normal identity values.
- Results deduplicate by URL or title. Imported detection uses the source URL stored in an
  attachment description.
- A stored row offers local download/open and cannot be selected again.
- Imports become real local attachments and record source plus URL.
- Category inference is Ersatzteilliste, Anleitung, or Dokumentation.
- Only public HTTP(S) URLs are accepted. Redirects are rechecked. Private/internal targets, empty
  content, disallowed types, and oversized content are rejected. Remote timeout is ten seconds.
- Multi-document imports are sequential. Earlier success survives later failure.

Attachment extraction
- Extract spare parts first saves the selected attachment metadata, then analyzes only that
  attachment and creates accepted suggestions. Only full success reloads the vehicle and switches
  tabs. A failure skips both final actions.
- A saved vehicle article number is required.
- Analysis reads at most 12 MiB and returns at most 80 unique suggestions.
- PDFs use direct text first; optional OCR may help scanned PDFs.
- Stable extraction has no row preview or confirmation and creates all remaining candidates
  sequentially.
- Existing duplicates, empty rows, and document-title-like rows are removed.
- An internal RailKeeper attachment-download URL is not stored as an external part link.
- Metadata and earlier created rows survive a later failure.

Manual spare parts
- Fields: article number, description, free-text price, and external link.
- Article number, description, or link is required. Price alone is invalid.
- Initial sort is article number ascending. All four columns toggle ascending/descending.
- Create identity uses normalized article number first, then normalized URL, then description.
- Article normalization ignores ET prefix, punctuation, spaces, and case.
- Create merge preserves existing non-empty values and fills only missing values.
- Editing a selected row writes exactly its four submitted fields.
- Save, apply, and delete write immediately and reload the full vehicle. Delete has no prompt.
- A full reload replaces unrelated unsaved editor state.

Lookup and manufacturer overview
- Single-part search requires a stored part article number and returns at most five candidates.
- Sort priority: has price, then has availability, then has link.
- Apply price and link persists only price and URL. Availability is display-only.
- If the selected result lacks price, stable code may apply another priced candidate's price and
  URL from the same result list.
- Opening the tab checks linked rows once. Piko/Roco use one overview request. Other manufacturers
  inspect only the first four linked rows that have article numbers.
- Availability remains an unverified suggestion.
- Find available spare parts requires a saved Piko/Roco vehicle, vehicle article number, and
  writable UI. With article search disabled, selecting the enabled button stops before the request,
  shows the Settings message, and writes nothing.
- Overview import consolidates duplicates, updates existing rows first, then creates missing rows.
- Existing article number, description, price, and URL win. Only missing price/URL are filled.
- The sequence is non-atomic and has no preview. No missing spare parts found can mean every
  suggestion already matched.

Storage and recovery
- Include the complete persistence/refresh matrix from the approved specification.
- State that no later request rolls back an earlier successful sequential write.
- Saved article-derived fields, imported images, attachment blobs, and spare parts are included in
  application backup scope. Search responses, selections, and unsaved drafts are not.
- Recommend a validated application backup before large imports, extraction, or cleanup.
- Include all 17 empty, partial, and error situations from the approved specification.
- Mention that stable English mode can still show some German backend/frontend messages.
```

End the English page with exactly these published links and version statement:

```markdown
## Related pages

- [User Guide overview](/guide/)
- [Vehicle inventory and basic data](/guide/vehicles/)
- [Vehicle images and attachments](/guide/vehicles/media)
- [Vehicle maintenance and condition](/guide/vehicles/maintenance)
- [Decoder, functions, and CV data](/guide/vehicles/decoder-cv)

## Documented RailKeeper version

This page documents stable RailKeeper **v0.1.17.6** and was last reviewed on 2026-08-16.
```

- [ ] **Step 3: Create the German semantic counterpart**

Create `docs/site/de/guide/vehicles/search-and-spares.md` with this frontmatter and matching
section order:

```markdown
---
title: Artikelsuche, Web-Dokumente und Ersatzteile
description: Externe Artikeldaten prüfen, Web-Dokumente importieren und Fahrzeugersatzteile pflegen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Artikelsuche, Web-Dokumente und Ersatzteile

## Voraussetzungen, Einstellungen und Zugriffsrechte

## Nach Artikeldaten suchen

## Mit Barcode oder Kamera suchen

## Ein Suchergebnis prüfen und übernehmen

## Web-Dokumente finden und importieren

## Ersatzteile aus einer Beilage extrahieren

## Ersatzteile manuell pflegen

## Ein gespeichertes Ersatzteil nachschlagen

## Eine Piko- oder Roco-Herstellerübersicht importieren

## Daten bei mehrstufigen Aktionen schützen

## Leere, teilweise und fehlerhafte Zustände beheben

## Verwandte Seiten

## Dokumentierte RailKeeper-Version
```

Translate every English requirement from Step 2 semantically, including every numeric bound,
role, ordering rule, persistence boundary, partial-failure warning, and recovery action. Use the
real German UI labels from stable translations where they exist. Keep technical stored category
values `Ersatzteilliste`, `Anleitung`, and `Dokumentation` unchanged. Do not translate source names,
API concepts, or Piko/Roco identity rules into behavior that stable code does not implement.

End the German page with exactly:

```markdown
## Verwandte Seiten

- [Übersicht des Benutzerhandbuchs](/de/guide/)
- [Fahrzeugbestand und Grunddaten](/de/guide/vehicles/)
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
- [Fahrzeugwartung und Zustand](/de/guide/vehicles/maintenance)
- [Decoder, Funktionen und CV-Daten](/de/guide/vehicles/decoder-cv)

## Dokumentierte RailKeeper-Version

Diese Seite dokumentiert die stabile RailKeeper-Version **v0.1.17.6** und wurde zuletzt am
16.08.2026 geprüft.
```

- [ ] **Step 4: Verify page parity and make the coverage gate pass**

Run from the repository root:

```powershell
rg -n '^## ' docs/site/guide/vehicles/search-and-spares.md
rg -n '^## ' docs/site/de/guide/vehicles/search-and-spares.md
rg -n '10|12 MiB|80|first four|v0.1.17.6|sequential' docs/site/guide/vehicles/search-and-spares.md
rg -n '10|12 MiB|80|ersten vier|v0.1.17.6|nacheinander' docs/site/de/guide/vehicles/search-and-spares.md
$markers = @('TO' + 'DO', 'T' + 'BD', 'FIX' + 'ME', [char]0x2014)
$pages = @(
  'docs/site/guide/vehicles/search-and-spares.md',
  'docs/site/de/guide/vehicles/search-and-spares.md'
)
Select-String -Path $pages -Pattern $markers
git diff --check
Set-Location docs
npm.cmd run check
```

Expected: both pages have the same 13-section semantic order, key-fact scans find the stable
bounds and sequential-write guidance, the marker scan and `git diff --check` print nothing, all 19
tests pass, coverage validation succeeds, and VitePress builds.

- [ ] **Step 5: Commit the coverage contract and paired chapter**

Run from the repository root:

```powershell
git add docs/coverage.json docs/site/guide/vehicles/search-and-spares.md
git add docs/site/de/guide/vehicles/search-and-spares.md
git commit -m "docs: add vehicle search and spare parts guide"
```

---

### Task 2: Connect sidebars, landing pages, and published related pages

**Files:**
- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/site/guide/vehicles/index.md`
- Modify: `docs/site/de/guide/vehicles/index.md`
- Modify: `docs/site/guide/vehicles/media.md`
- Modify: `docs/site/de/guide/vehicles/media.md`
- Modify: `docs/site/guide/vehicles/maintenance.md`
- Modify: `docs/site/de/guide/vehicles/maintenance.md`
- Modify: `docs/site/guide/vehicles/decoder-cv.md`
- Modify: `docs/site/de/guide/vehicles/decoder-cv.md`

**Interfaces:**
- Consumes: the two validated routes from Task 1.
- Produces: discoverable paired pages with links only among published user-guide owners.

- [ ] **Step 1: Add the sidebar routes after decoder/CV**

Add to the German sidebar immediately after the decoder/CV entry:

```ts
{
  text: "Artikelsuche, Web-Dokumente und Ersatzteile",
  link: "/de/guide/vehicles/search-and-spares",
},
```

Add to the English sidebar immediately after its decoder/CV entry:

```ts
{
  text: "Article search, web documents, and spare parts",
  link: "/guide/vehicles/search-and-spares",
},
```

- [ ] **Step 2: Add one landing-page transition per language**

Append after the decoder/CV paragraph in `docs/site/guide/index.md`:

```markdown

[Article search, web documents, and spare parts](/guide/vehicles/search-and-spares) explains how to
verify external suggestions, save selected article data and images, import documents, and maintain
spare parts without overlooking partial writes.
```

Append the counterpart to `docs/site/de/guide/index.md`:

```markdown

[Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares) erklärt, wie du
externe Vorschläge prüfst, ausgewählte Artikeldaten und Bilder speicherst, Dokumente importierst und
Ersatzteile pflegst, ohne teilweise Schreibvorgänge zu übersehen.
```

- [ ] **Step 3: Add the new page to eight existing related-page lists**

Add this English item after decoder/CV in the core vehicle, media, and maintenance pages, and after
maintenance in the decoder/CV page:

```markdown
- [Article search, web documents, and spare parts](/guide/vehicles/search-and-spares)
```

Add this German item at the equivalent position in all four German pages:

```markdown
- [Artikelsuche, Web-Dokumente und Ersatzteile](/de/guide/vehicles/search-and-spares)
```

In the media chapter's role/boundary paragraph, replace the claim that **Found documents**,
**Extract spare parts**, web-document import, and remote article images are unexplained specialist
boundaries. Keep CV files and maintenance editing as boundaries, and add a direct sentence linking
article search/document/spare-part workflows to the new page. Make the same semantic correction in
German.

- [ ] **Step 4: Verify route counts, link boundaries, and the full documentation build**

Run:

```powershell
rg -n 'vehicles/search-and-spares' docs/.vitepress/config.mts docs/site/guide docs/site/de/guide
rg -n 'settings|master.data|accessor|digital cent|administration' `
  docs/site/guide/vehicles/search-and-spares.md `
  docs/site/de/guide/vehicles/search-and-spares.md
git diff --check
Set-Location docs
npm.cmd run check
```

Expected: the route appears in two sidebars, two landing transitions, eight existing related-page
lists, and five related links on each new page. No active link targets a planned owner. All 19
tests pass, validation succeeds, and VitePress builds.

- [ ] **Step 5: Commit navigation and cross-links**

Run from the repository root:

```powershell
git add docs/.vitepress/config.mts docs/site/guide/index.md docs/site/de/guide/index.md
git add docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md
git add docs/site/guide/vehicles/media.md docs/site/de/guide/vehicles/media.md
git add docs/site/guide/vehicles/maintenance.md docs/site/de/guide/vehicles/maintenance.md
git add docs/site/guide/vehicles/decoder-cv.md docs/site/de/guide/vehicles/decoder-cv.md
git commit -m "docs: link vehicle search and spare parts guide"
```

---

### Task 3: Audit stable-source fidelity and clear independent review

**Files:**
- Review: every file changed in `origin/main..HEAD`
- Reference: `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-search-spares-guide-design.md`
- Reference: stable frontend, backend, route, OpenAPI, translation, migration, and backup sources.

**Interfaces:**
- Consumes: the committed coverage, page pair, navigation, and cross-link diff.
- Produces: a review-cleared exact head with no Critical, Important, or valid completeness finding.

- [ ] **Step 1: Recheck the highest-risk behavior directly at the stable tag**

Run from the repository root:

```powershell
git show v0.1.17.6:frontend/src/features/vehicles/useArticleSearchController.ts
git show v0.1.17.6:frontend/src/shared/articleSearch/ArticleSearchDialog.tsx
git show v0.1.17.6:frontend/src/shared/articleSearch/BarcodeSearchDialog.tsx
git show v0.1.17.6:frontend/src/shared/articleSearch/articleSearchPreferences.ts
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleDocumentsController.ts
git show v0.1.17.6:frontend/src/features/vehicles/vehicleDocuments.ts
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleSparePartsController.ts
git show v0.1.17.6:frontend/src/features/vehicles/VehicleSparePartsTab.tsx
git show v0.1.17.6:frontend/src/features/vehicles/vehicleSparePartSearch.ts
git show v0.1.17.6:frontend/src/features/vehicles/vehicleSpareParts.ts
git show v0.1.17.6:backend/internal/application/article_search.go
git show v0.1.17.6:backend/internal/application/article_search_spare_parts.go
git show v0.1.17.6:backend/internal/application/vehicle_spare_parts_service.go
git show v0.1.17.6:backend/internal/api/vehicle_operation_handlers.go
git show v0.1.17.6:backend/internal/api/vehicle_attachment_handlers.go
git show v0.1.17.6:backend/internal/api/vehicle_image_handlers.go
git show v0.1.17.6:backend/internal/api/routes.go
git show v0.1.17.6:backend/internal/application/backup.go
git show v0.1.17.6:openapi/railkeeper.yaml
```

Confirm prerequisites, EAN exception, browser-local sources, search limits, selection defaults,
camera behavior, write order, URL safety, categories, extraction bounds, OCR boundary, identity and
merge rules, immediate refreshes, lookup sorting/application, availability checks, Piko/Roco import
order, role boundaries, storage, and backup scope.

- [ ] **Step 2: Inspect the complete diff and repository state**

Run:

```powershell
git diff --check origin/main..HEAD
git diff --stat origin/main..HEAD
git status --short
git log --oneline origin/main..HEAD
```

Expected: no whitespace error, no runtime file, no generated output, and no unrelated file appears.
The worktree is clean after committed documentation changes.

- [ ] **Step 3: Request an independent read-only review**

Use the `requesting-code-review` workflow with this exact review brief:

```text
Base: output of git rev-parse origin/main
Head: output of git rev-parse HEAD
Specification: docs/superpowers/specs/2026-08-16-railkeeper-vehicle-search-spares-guide-design.md
Focus: stable v0.1.17.6 fidelity, English/German parity, real labels, search prerequisites and
limits, source trust, result selection defaults, transient versus stored state, vehicle/image save
order, document import URL safety, extraction bounds and OCR, spare-part validation and identity,
lookup application, availability checks, Piko/Roco conservative import, roles, immediate refreshes,
partial sequential writes, backup scope, coverage, navigation, and unpublished-link boundaries.
The reviewer must not mutate the worktree.
```

Fix every Critical and Important finding. Apply valid Minor findings that improve stable fidelity,
parity, safety, or completeness.

- [ ] **Step 4: Verify and commit review corrections**

After each correction batch, run:

```powershell
Set-Location docs
npm.cmd run check
Set-Location ..
git diff --check origin/main..HEAD
git status --short
```

Expected: all 19 tests pass, validation and VitePress build succeed, and no diff error remains.
Commit real corrections with:

```powershell
git add docs
git commit -m "docs: refine vehicle search and spare parts guide"
```

Request a focused read-only re-review of every correction. Do not publish while a Critical,
Important, or valid completeness finding remains.

---

### Task 4: Publish and merge only when the reviewed head is green

**Files:**
- No new source files expected.
- Verify: committed branch `dev/docs-user-guide-search-spares`.

**Interfaces:**
- Consumes: a clean, independently reviewed branch with fresh local verification.
- Produces: a merged pull request on `main`, guarded by the exact reviewed head SHA and confirmed
  successful documentation publication.

- [ ] **Step 1: Run fresh pre-push verification**

Run:

```powershell
Set-Location docs
npm.cmd run check
Set-Location ..
git diff --check origin/main..HEAD
git status --short
git rev-parse HEAD
git merge-base HEAD origin/main
```

Expected: all 19 tests pass, validation and VitePress build succeed, no diff error or uncommitted
file exists, and the merge base is the expected `origin/main` commit.

- [ ] **Step 2: Push only the feature branch**

Run:

```powershell
git push -u origin dev/docs-user-guide-search-spares
```

Do not modify or push local `main`.

- [ ] **Step 3: Create and ready the pull request**

Create a draft pull request against `main` titled:

```text
docs: add bilingual vehicle search and spare parts guide
```

Use this body:

```markdown
## Summary

- add complete English and German vehicle search and spare-parts chapters for stable v0.1.17.6
- document barcode/EAN search, result review, remote images, web documents, extraction, and lookup
- explain immediate and partial writes, recovery, roles, storage, and backup boundaries
- mark search/spare-parts coverage documented and connect published guide navigation

## Verification

- `npm.cmd run check`
- stable-tag source audit against `v0.1.17.6`
- independent English/German fidelity, persistence, and safety review

No runtime or API behavior changes are included.
```

Mark the pull request ready only after its remote head SHA equals the locally reviewed SHA.

- [ ] **Step 4: Monitor required checks and review threads for the exact SHA**

Require these checks for the exact head:

```text
CI: success
Trivy: success
CodeQL: success
```

Inspect every review conversation. Resolve a thread only after correction or a source-backed reason
that it is inapplicable. Any pushed correction invalidates the prior head review and check result.
Rerun local verification, re-review the correction, and wait for all exact-head checks again.

- [ ] **Step 5: Merge with expected-head protection and verify publication**

Immediately before merge, confirm the pull request is open, non-draft, mergeable, has no unresolved
review thread, and still points to the reviewed SHA. Merge with expected-head protection. Fetch PR
metadata again and require `state: closed`, `merged: true`, and a non-empty merge commit SHA.

Then monitor the merge-triggered workflows and require successful Documentation Pages publication
for the merge commit. Verify both live routes return HTTP 200 and display `v0.1.17.6`. Leave the
isolated worktree in place for traceability and do not modify or push local `main`.
