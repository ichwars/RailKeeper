# RailKeeper Vehicle Media User Guide Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a complete, source-verified English and German user-guide chapter for vehicle
images and general attachments in stable RailKeeper v0.1.17.6.

**Architecture:** Add one paired VitePress Markdown chapter under the existing vehicle guide, keep
the operational workflow in a single page per language, and update the coverage contract so web
document search and remote imports remain owned by the later search-and-spares chapter. Navigation
and cross-links are added only after the content pair validates successfully.

**Tech Stack:** VitePress 2, Markdown, JSON coverage manifest, Node.js documentation tests, GitHub
Actions.

## Global Constraints

- Document only stable RailKeeper **v0.1.17.6**. Do not describe later `main` behavior as stable.
- Create both `docs/site/guide/vehicles/media.md` and
  `docs/site/de/guide/vehicles/media.md` with semantically equivalent content.
- Use exact stable UI labels in the respective language and stored German attachment-category values.
- Cover local images, general attachments, preview/download behavior, metadata persistence,
  maintenance links, deletion, formats, default limits, partial uploads, roles, and errors.
- Mention web document search, remote image/document imports, CV files, maintenance workflows, and
  spare-part extraction only as boundaries. Do not explain those specialist workflows here.
- Do not add links to unpublished target pages.
- Keep lines readable, avoid em dashes, and do not introduce `TODO`, `TBD`, or `FIXME` markers.
- Do not commit `docs/.vitepress/dist`, dependency caches, local screenshots, or other generated output.
- Preserve the dirty local `main`; all work stays in the isolated
  `dev/docs-user-guide-vehicle-media` worktree.
- Merge the GitHub pull request only after CI, Trivy, and CodeQL all conclude successfully on the
  unchanged expected head SHA.

---

### Task 1: Sharpen the coverage contract and prove the missing-page gate

**Files:**
- Modify: `docs/coverage.json`
- Reference: `docs/scripts/validate-docs.mjs`
- Reference: `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-media-guide-design.md`

**Interfaces:**
- Consumes: longest-prefix ownership implemented by `longestPrefixOwner()` in
  `docs/scripts/validate-docs.mjs`.
- Produces: an honest `vehicle-media` coverage contract whose only missing artifacts are the paired
  media pages.

- [ ] **Step 1: Confirm the stable source surfaces before changing ownership**

Run from the repository root:

```powershell
git grep -n "vehicles.uploads" v0.1.17.6 -- frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git grep -n -E 'images/import-url|attachments/import-url' v0.1.17.6 -- backend/internal/api/routes.go
```

Expected: web-document and spare-part extraction labels are nested below `vehicles.uploads`, and
the remote import routes are nested below the general images/attachments API prefixes.

- [ ] **Step 2: Route excluded specialist surfaces to `vehicle-search-spares`**

Add these exact entries to `owners.translationPrefixes`, before the general
`vehicles.uploads` entry for readability:

```json
"vehicles.uploads.extractSpareParts": "vehicle-search-spares",
"vehicles.uploads.webDocument": "vehicle-search-spares",
"vehicles.uploads.webDocuments": "vehicle-search-spares"
```

Add these exact entries to `owners.apiPrefixes`, before the general images and attachments entries:

```json
"/api/v1/vehicles/{id}/images/import-url": "vehicle-search-spares",
"/api/v1/vehicles/{id}/attachments/import-url": "vehicle-search-spares"
```

Keep local image and attachment upload, metadata, download, and deletion owned by `vehicle-media`.

- [ ] **Step 3: Mark the media topic documented before its pages exist**

Change only this topic status:

```json
{
  "id": "vehicle-media",
  "audience": "user",
  "status": "documented",
  "englishPath": "guide/vehicles/media.md",
  "germanPath": "de/guide/vehicles/media.md"
}
```

- [ ] **Step 4: Run the documentation check and verify the intentional failure**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: the 19 unit tests pass, then coverage validation fails only because
`guide/vehicles/media.md` and `de/guide/vehicles/media.md` do not exist. Any ownership, JSON, or
unrelated validation failure must be fixed before continuing.

Do not commit the intentionally red state.

---

### Task 2: Create the complete English and German media chapter

**Files:**
- Create: `docs/site/guide/vehicles/media.md`
- Create: `docs/site/de/guide/vehicles/media.md`
- Modify: `docs/coverage.json` (already changed and uncommitted in Task 1)
- Reference: `frontend/src/features/vehicles/VehicleUploadsTab.tsx` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/useVehicleMediaController.ts` at tag `v0.1.17.6`
- Reference: `frontend/src/features/vehicles/vehicleFiles.ts` at tag `v0.1.17.6`
- Reference: `backend/internal/api/vehicle_image_handlers.go` at tag `v0.1.17.6`
- Reference: `backend/internal/api/vehicle_attachment_handlers.go` at tag `v0.1.17.6`
- Reference: `backend/internal/application/vehicle_media.go` at tag `v0.1.17.6`

**Interfaces:**
- Consumes: the `vehicle-media` English/German paths from `docs/coverage.json`.
- Produces: a validated language pair at `/guide/vehicles/media` and
  `/de/guide/vehicles/media`.

- [ ] **Step 1: Create the English page with stable metadata**

Start `docs/site/guide/vehicles/media.md` with exactly:

```markdown
---
title: Vehicle images and attachments
description: Upload, organize, preview, download, and safely remove vehicle images and attachments.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Vehicle images and attachments
```

Use this complete section order:

```markdown
## Open the Uploads tab
## Upload local images
## Organize image metadata
## Link an image to maintenance
## Upload general attachments
## Attachment formats, limits, and categories
## Preview, open, and download attachments
## Delete images and attachments
## Roles, storage, and backup boundaries
## Empty, partial, and error states
## Related pages
## Documented RailKeeper version
```

The prose must explicitly state all of the following:

- a vehicle must be saved before local media can be added;
- Admin, Editor, Viewer, and Planner can inspect media, while server writes require Admin or Editor;
- local image formats are JPG/JPEG, PNG, and WebP, with a default 10 MB server limit per image;
- multiple files upload sequentially; when none exists initially, the last successfully uploaded
  image becomes the main image;
- upload and persisted-image deletion are immediate operations;
- description, order, main-image selection, and maintenance link require **Save changes**;
- deleting a linked image requires clearing its maintenance link and saving first;
- deleting the main image promotes the next sorted image;
- attachment upload supports file selection and drag-and-drop after the vehicle exists;
- default attachment formats are PDF, TXT, CSV, JSON, XML, ZIP, JPG/JPEG, PNG, and WebP, and the
  server can restrict this extension set;
- the default attachment limit is 25 MB per file and the server can enforce stricter configuration;
- empty, executable, disallowed, content-blocked, and oversized files are rejected;
- one pre-upload category/note applies to all files in that upload;
- stored categories are `Anleitung`, `Rechnung`, `Decoder-Datei`, `Dokumentation`,
  `Ersatzteilliste`, `Zertifikat`, and `Sonstiges`;
- automatic category priority exactly matches `attachmentCategoryForFile()`;
- each stored attachment shows original name, category, MIME type, and size;
- category and note changes persist only through the row save action; saving reloads the vehicle,
  so pending image metadata must be saved first and attachment rows must be edited and saved one at
  a time;
- PDF, image, TXT, CSV, JSON, and XML previews are supported, while ZIP has no inline preview;
- Preview, Open file, and Download file are distinct actions;
- image removal is immediate without an extra confirmation, attachment removal asks for confirmation;
- successful earlier files remain stored when a later file in a batch fails, and later files are not
  attempted in that run; reload the vehicle and compare stored media before retrying only files that
  are still missing;
- uploads are local/private RailKeeper data and belong to the application backup scope; v0.1.17.6
  accepts backup files only up to 250 MiB for validation and restore, so the exported backup must be
  accepted as compatible before destructive cleanup;
- web-document import, remote article images, spare-part extraction, CV files, and maintenance
  editing are specialist boundaries, not workflows explained on this page.

- [ ] **Step 2: Create the German page with semantic parity and exact UI labels**

Start `docs/site/de/guide/vehicles/media.md` with exactly:

```markdown
---
title: Fahrzeugbilder und Beilagen
description: Fahrzeugbilder und Beilagen hochladen, ordnen, anzeigen, herunterladen und sicher entfernen.
audience: user
status: stable
reviewedVersion: 0.1.17.6
lastReviewed: 2026-08-16
---

# Fahrzeugbilder und Beilagen
```

Use the semantically equivalent German headings:

```markdown
## Tab Uploads öffnen
## Lokale Bilder hochladen
## Bildmetadaten ordnen
## Bild mit einer Wartung verknüpfen
## Allgemeine Beilagen hochladen
## Formate, Grenzen und Kategorien für Beilagen
## Beilagen anzeigen, öffnen und herunterladen
## Bilder und Beilagen löschen
## Rollen-, Speicher- und Sicherungsgrenzen
## Leere, teilweise und fehlerhafte Zustände
## Verwandte Seiten
## Dokumentierte RailKeeper-Version
```

Use exact visible stable labels including **Uploads**, **Bild hochladen**, **Hauptbild**,
**Bildbeschreibung**, **Keine Wartung**, **Beilage hochladen**, **Dateien hier ablegen**,
**Kategorie automatisch**, **Anzeigen**, **Datei öffnen**, **Datei herunterladen**, and
**Beilage löschen?**. Keep stored category values identical to the English page.

- [ ] **Step 3: Run structural and unfinished-content checks**

Run from the repository root:

```powershell
rg -n "^#{2,3} " docs/site/guide/vehicles/media.md docs/site/de/guide/vehicles/media.md
rg -n "TODO|TBD|FIXME|—" docs/site/guide/vehicles/media.md docs/site/de/guide/vehicles/media.md
git diff --check
```

Expected: both pages have 12 level-two headings in matching order, no unfinished marker or em dash,
and no whitespace errors.

- [ ] **Step 4: Run the full documentation check**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: all 19 tests pass, documentation validation passes, and the VitePress production build
completes successfully.

- [ ] **Step 5: Commit the coverage contract and paired pages**

Run from the repository root:

```powershell
git add docs/coverage.json docs/site/guide/vehicles/media.md docs/site/de/guide/vehicles/media.md
git commit -m "docs: add vehicle media user guide"
```

---

### Task 3: Add navigation, landing links, and vehicle-guide cross-links

**Files:**
- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/site/guide/vehicles/index.md`
- Modify: `docs/site/de/guide/vehicles/index.md`

**Interfaces:**
- Consumes: the validated routes `/guide/vehicles/media` and `/de/guide/vehicles/media` from Task 2.
- Produces: discoverable navigation without linking to any unpublished specialist page.

- [ ] **Step 1: Add the sidebar items after the core vehicle chapter**

Add these exact entries immediately after the existing vehicle-inventory entries:

```ts
{ text: "Vehicle images and attachments", link: "/guide/vehicles/media" }
```

```ts
{ text: "Fahrzeugbilder und Beilagen", link: "/de/guide/vehicles/media" }
```

- [ ] **Step 2: Add one concise landing-page link per language**

The English paragraph must say the page covers local images, main/alternative image handling,
general attachments, preview/download, and safe removal. The German paragraph must convey the same
scope without adding specialist web-document, CV, maintenance, or spare-parts instructions.

- [ ] **Step 3: Add the media page to both core vehicle related-page lists**

Add these links under the existing overview link:

```markdown
- [Vehicle images and attachments](/guide/vehicles/media)
```

```markdown
- [Fahrzeugbilder und Beilagen](/de/guide/vehicles/media)
```

- [ ] **Step 4: Re-run the complete documentation check**

Run:

```powershell
Set-Location docs
npm.cmd run check
```

Expected: 19 tests pass, validation passes, and VitePress builds every linked page.

- [ ] **Step 5: Commit navigation and cross-links**

Run from the repository root:

```powershell
git add docs/.vitepress/config.mts docs/site/guide/index.md docs/site/de/guide/index.md docs/site/guide/vehicles/index.md docs/site/de/guide/vehicles/index.md
git commit -m "docs: link vehicle media guide"
```

---

### Task 4: Perform source-fidelity audit and independent review

**Files:**
- Review: every file changed since `origin/main`
- Reference: `docs/superpowers/specs/2026-08-16-railkeeper-vehicle-media-guide-design.md`
- Reference: stable source files listed in Task 2

**Interfaces:**
- Consumes: the complete committed documentation diff.
- Produces: a review-cleared head commit with no Critical or Important findings.

- [ ] **Step 1: Audit the highest-risk claims directly against v0.1.17.6**

Run targeted searches:

```powershell
git grep -n -E 'maxImageBytes|maxAttachmentBytes|RAILKEEPER_MAX_IMAGE_MB|RAILKEEPER_MAX_ATTACHMENT_MB' v0.1.17.6 -- backend
git show v0.1.17.6:frontend/src/features/vehicles/vehicleFiles.ts
git show v0.1.17.6:frontend/src/features/vehicles/useVehicleMediaController.ts
git show v0.1.17.6:frontend/src/features/vehicles/VehicleUploadsTab.tsx
git show v0.1.17.6:backend/internal/application/vehicle_media.go
```

Confirm formats, default limits, category priority, immediate versus deferred persistence,
maintenance-linked deletion protection, primary-image promotion, sequential batches, preview kinds,
attachment confirmation, and cleanup claims.

- [ ] **Step 2: Inspect the full diff and repository state**

Run:

```powershell
git diff --check origin/main..HEAD
git diff --stat origin/main..HEAD
git status --short
git log --oneline origin/main..HEAD
```

Expected: no whitespace errors, only the specification, coverage, media pages, navigation, and
cross-link changes are present, and the worktree is clean.

- [ ] **Step 3: Run an independent read-only review**

Dispatch the repository review workflow with:

- base SHA from `git rev-parse origin/main`;
- head SHA from `git rev-parse HEAD`;
- the confirmed specification path;
- explicit checks for version fidelity, EN/DE semantic parity, coverage ownership, role boundaries,
  persistence timing, partial batch behavior, preview support, maintenance links, and deletion.

The reviewer must not mutate the worktree. Fix every Critical and Important finding before
continuing. Apply valid Minor corrections when they improve exact UI terminology or completeness.

- [ ] **Step 4: Verify any review corrections and commit them**

After corrections, run:

```powershell
Set-Location docs
npm.cmd run check
Set-Location ..
git diff --check origin/main..HEAD
git status --short
```

Expected: 19 tests pass, VitePress builds, no diff errors remain, and the worktree is clean after a
correction commit such as:

```powershell
git add docs
git commit -m "docs: refine vehicle media guide"
```

Request a focused read-only re-review of the corrections. Do not publish while any Critical or
Important finding remains.

---

### Task 5: Publish the reviewed branch and merge only when GitHub is green

**Files:**
- No new source files expected.
- Verify: committed branch `dev/docs-user-guide-vehicle-media`.

**Interfaces:**
- Consumes: a clean, independently reviewed branch with fresh local verification.
- Produces: a merged pull request on `main`, guarded by the exact reviewed head SHA.

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

Expected: 19 tests pass, VitePress builds, no diff errors or uncommitted files exist, and the merge
base is the expected `origin/main` commit.

- [ ] **Step 2: Push only the feature branch**

Run:

```powershell
git push -u origin dev/docs-user-guide-vehicle-media
```

Do not push the locally divergent `main` branch.

- [ ] **Step 3: Create and ready the pull request**

Create a draft pull request against `main` titled:

```text
docs: add bilingual vehicle media guide
```

The body must list the paired pages, coverage-boundary correction, navigation updates, stable
v0.1.17.6 source audit, `npm.cmd run check`, independent review, and the absence of runtime/API
behavior changes. Mark it ready only after confirming the remote head SHA equals the locally
reviewed SHA.

- [ ] **Step 4: Monitor all required GitHub checks**

Poll workflow runs for the exact head SHA until all three complete:

```text
CI: success
Trivy: success
CodeQL: success
```

If a check fails, inspect the failing job, fix the root cause on the feature branch, rerun local
verification, push the new commit, and restart the expected-head verification. Do not merge a stale
or partially green SHA.

- [ ] **Step 5: Merge with expected-head protection and verify closure**

Immediately before merging, confirm the pull request is open, non-draft, mergeable, and still points
to the reviewed SHA. Merge with `expected_head_sha` set to that exact SHA. Then fetch pull-request
metadata again and require:

```text
state: closed
merged: true
merge_commit_sha: non-empty
```

Leave the isolated worktree in place for traceability and do not modify or push local `main`.
