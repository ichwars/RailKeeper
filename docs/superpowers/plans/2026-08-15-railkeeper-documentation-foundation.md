# RailKeeper Documentation Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the bilingual VitePress foundation, automated documentation checks, coverage
inventory, and GitHub Pages deployment for the RailKeeper handbook.

**Architecture:** VitePress 1.6.4 runs as an isolated Node 24 workspace under `docs/`, with authored
content in `docs/site/`. English uses the site root and German uses `/de/`. Small dependency-free
Node scripts enforce paired pages and map repository surfaces into a reviewed coverage manifest.

**Tech Stack:** Node.js 24, npm, VitePress 1.6.4, Markdown, CSS, Node's built-in test runner,
GitHub Actions, GitHub Pages.

## Global Constraints

- English is the public default; German is complete and equal, with identical paths below `/de/`.
- User and administration pages describe stable RailKeeper v0.1.17.5.
- Development and architecture pages describe `main`.
- Every published page has `audience`, `status`, `reviewedVersion`, and `lastReviewed` metadata.
- Search remains local in the browser. Do not add analytics, cookies, or an external search service.
- Reuse the existing RailKeeper SVG brand assets and the existing color tokens.
- Do not publish generated output, caches, local credentials, private data, or `.superpowers/`.
- Preserve the current `docs/*.md`, `docs/releases/`, and `docs/screenshots/` files until their
  content is deliberately migrated in later stages.
- Do not document the hidden layout workspace as a stable user feature.
- This plan implements Etappe 1 only. Benutzerhandbuch, Administration, Entwicklung, and final
  content audit receive separate plans after this foundation is accepted.

---

## File Map

| Path | Responsibility |
| --- | --- |
| `docs/package.json` | Isolated documentation commands and pinned VitePress dependency |
| `docs/package-lock.json` | Reproducible dependency graph |
| `docs/.vitepress/config.mts` | Routing, locales, navigation, local search, shared brand assets |
| `docs/.vitepress/theme/index.ts` | Default-theme extension |
| `docs/.vitepress/theme/custom.css` | Restrained RailKeeper visual tokens |
| `docs/site/` | Published English pages and mirrored `de/` tree |
| `docs/versions.json` | Canonical stable and development documentation versions |
| `docs/coverage.json` | Reviewed mapping from repository surfaces to documentation topics |
| `docs/scripts/validate-docs.mjs` | Pairing, metadata, and coverage validation CLI |
| `docs/scripts/validate-docs.test.mjs` | Validator regression tests |
| `docs/scripts/source-inventory.mjs` | Read-only source discovery for routes, API, i18n, config, and docs |
| `docs/scripts/source-inventory.test.mjs` | Inventory parser regression tests |
| `.github/workflows/ci.yml` | Pull-request documentation quality gate |
| `.github/workflows/docs-pages.yml` | GitHub Pages build and deployment |

## Reference Documents

- Design: `docs/superpowers/specs/2026-08-15-railkeeper-bilingual-documentation-design.md`
- Existing overview: `README.md` and `README.de.md`
- Existing operations: `docs/production-runbook.md`
- Existing architecture: `docs/architecture.md`
- Existing security model: `docs/security.md`
- VitePress i18n: <https://vitepress.dev/guide/i18n>
- VitePress local search: <https://vitepress.dev/reference/default-theme-search>
- VitePress GitHub Pages deployment: <https://vitepress.dev/guide/deploy>

### Task 1: Scaffold the isolated bilingual VitePress site

**Files:**

- Modify: `.gitignore`
- Create: `docs/package.json`
- Create: `docs/package-lock.json`
- Create: `docs/.vitepress/config.mts`
- Create: `docs/.vitepress/theme/index.ts`
- Create: `docs/.vitepress/theme/custom.css`
- Create: `docs/site/index.md`
- Create: `docs/site/de/index.md`
- Create: `docs/site/guide/index.md`
- Create: `docs/site/de/guide/index.md`
- Create: `docs/site/administration/index.md`
- Create: `docs/site/de/administration/index.md`
- Create: `docs/site/development/index.md`
- Create: `docs/site/de/development/index.md`
- Create: `docs/site/reference/index.md`
- Create: `docs/site/de/reference/index.md`

**Interfaces:**

- Consumes: `frontend/public/brand/railkeeper-mark.svg` and repository Git history.
- Produces: `npm run dev`, `npm run build`, and `npm run preview` in `docs/`; stable paths `/`,
  `/guide/`, `/administration/`, `/development/`, `/reference/`, and their `/de/` counterparts.

- [ ] **Step 1: Ignore only documentation dependencies and generated build state**

Append this block to `.gitignore`:

```gitignore
# Documentation site
docs/node_modules/
docs/.vitepress/cache/
docs/.vitepress/dist/
.superpowers/
```

- [ ] **Step 2: Create the isolated package manifest**

Create `docs/package.json`:

```json
{
  "name": "railkeeper-documentation",
  "private": true,
  "license": "AGPL-3.0-only",
  "scripts": {
    "dev": "vitepress dev",
    "build": "vitepress build",
    "preview": "vitepress preview",
    "test": "node --test scripts/*.test.mjs",
    "check": "npm test && node scripts/validate-docs.mjs && vitepress build"
  },
  "devDependencies": {
    "vitepress": "1.6.4"
  }
}
```

- [ ] **Step 3: Generate the lockfile without sharing frontend dependencies**

Run:

```powershell
cd docs
npm.cmd install
```

Expected: `docs/package-lock.json` is created, `npm audit` reports no unresolved high or critical
finding, and no root-level `package.json` is introduced.

- [ ] **Step 4: Create the VitePress configuration**

Create `docs/.vitepress/config.mts` with these exact behaviors:

```ts
import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vitepress";

const repositoryUrl = "https://github.com/ichwars/RailKeeper";

export default defineConfig({
  title: "RailKeeper",
  description: "Documentation for the local-first model railway inventory and operations tool.",
  base: "/RailKeeper/",
  srcDir: "site",
  cleanUrls: true,
  lastUpdated: true,
  vite: {
    publicDir: fileURLToPath(new URL("../../frontend/public", import.meta.url)),
  },
  locales: {
    root: { label: "English", lang: "en" },
    de: {
      label: "Deutsch",
      lang: "de",
      link: "/de/",
      description: "Dokumentation für RailKeeper.",
    },
  },
  themeConfig: {
    logo: "/brand/railkeeper-mark.svg",
    siteTitle: "RailKeeper Docs",
    socialLinks: [{ icon: "github", link: repositoryUrl }],
    editLink: {
      pattern: `${repositoryUrl}/edit/main/docs/site/:path`,
      text: "Edit this page on GitHub",
    },
    search: {
      provider: "local",
      options: {
        locales: {
          de: {
            translations: {
              button: { buttonText: "Suchen", buttonAriaLabel: "Dokumentation durchsuchen" },
              modal: {
                displayDetails: "Detaillierte Liste anzeigen",
                resetButtonTitle: "Suche zurücksetzen",
                backButtonTitle: "Suche schließen",
                noResultsText: "Keine Ergebnisse gefunden für",
                footer: {
                  selectText: "Auswählen",
                  selectKeyAriaLabel: "Eingabetaste",
                  navigateText: "Navigieren",
                  navigateUpKeyAriaLabel: "Pfeil nach oben",
                  navigateDownKeyAriaLabel: "Pfeil nach unten",
                  closeText: "Schließen",
                  closeKeyAriaLabel: "Escape-Taste",
                },
              },
            },
          },
        },
      },
    },
    locales: {
      root: {
        nav: [
          { text: "User Guide", link: "/guide/" },
          { text: "Administration", link: "/administration/" },
          { text: "Development", link: "/development/" },
          { text: "Reference", link: "/reference/" },
        ],
        sidebar: {
          "/guide/": [{ text: "User Guide", items: [{ text: "Overview", link: "/guide/" }] }],
          "/administration/": [
            { text: "Administration", items: [{ text: "Overview", link: "/administration/" }] },
          ],
          "/development/": [
            { text: "Development", items: [{ text: "Overview", link: "/development/" }] },
          ],
          "/reference/": [{ text: "Reference", items: [{ text: "Overview", link: "/reference/" }] }],
        },
        outlineTitle: "On this page",
        lastUpdated: { text: "Last reviewed" },
      },
      de: {
        nav: [
          { text: "Benutzerhandbuch", link: "/de/guide/" },
          { text: "Administration", link: "/de/administration/" },
          { text: "Entwicklung", link: "/de/development/" },
          { text: "Referenz", link: "/de/reference/" },
        ],
        sidebar: {
          "/de/guide/": [
            { text: "Benutzerhandbuch", items: [{ text: "Überblick", link: "/de/guide/" }] },
          ],
          "/de/administration/": [
            { text: "Administration", items: [{ text: "Überblick", link: "/de/administration/" }] },
          ],
          "/de/development/": [
            { text: "Entwicklung", items: [{ text: "Überblick", link: "/de/development/" }] },
          ],
          "/de/reference/": [
            { text: "Referenz", items: [{ text: "Überblick", link: "/de/reference/" }] },
          ],
        },
        editLink: {
          pattern: `${repositoryUrl}/edit/main/docs/site/:path`,
          text: "Diese Seite auf GitHub bearbeiten",
        },
        outlineTitle: "Auf dieser Seite",
        lastUpdated: { text: "Zuletzt geprüft" },
      },
    },
  },
});
```

- [ ] **Step 5: Extend the default theme with RailKeeper tokens**

Create `docs/.vitepress/theme/index.ts`:

```ts
import DefaultTheme from "vitepress/theme";
import "./custom.css";

export default DefaultTheme;
```

Create `docs/.vitepress/theme/custom.css`:

```css
:root {
  --vp-c-brand-1: #419310;
  --vp-c-brand-2: #1c621b;
  --vp-c-brand-3: #154d15;
  --vp-c-brand-soft: #e9f8df;
  --vp-c-bg: #edf2f1;
  --vp-c-bg-soft: #f5f8f6;
  --vp-c-divider: #d5dfdc;
  --vp-c-text-1: #0b1e26;
  --vp-c-text-2: #4f6869;
  --vp-font-family-base: "Segoe UI", Arial, sans-serif;
}

.dark {
  --vp-c-brand-1: #a5ec60;
  --vp-c-brand-2: #8ad449;
  --vp-c-brand-3: #70ba32;
  --vp-c-brand-soft: rgb(165 236 96 / 13%);
  --vp-c-bg: #0f1314;
  --vp-c-bg-soft: #171c1f;
  --vp-c-divider: #273235;
  --vp-c-text-1: #f5f8f6;
  --vp-c-text-2: #a8b7b3;
}

.VPNavBarTitle .logo {
  height: 28px;
  width: 28px;
}

.VPHero .name {
  color: var(--vp-c-text-1);
}
```

- [ ] **Step 6: Create paired landing pages with real scope statements**

Create the page pairs below. Every file starts with the listed metadata. Use `status: stable` and
`reviewedVersion: 0.1.17.5` for home, guide, and administration. Use `status: development` and
`reviewedVersion: main` for development and reference. All use `lastReviewed: 2026-08-15`.

- `site/index.md` and `site/de/index.md`, audience `reference`: use the headings
  `RailKeeper Documentation` and `RailKeeper-Dokumentation`, link all three audiences, and state the
  stable version.
- `site/guide/index.md` and `site/de/guide/index.md`, audience `user`: use `User Guide` and
  `Benutzerhandbuch`, and define the stable end-user scope.
- `site/administration/index.md` and `site/de/administration/index.md`, audience `admin`: use
  `Installation and Administration` and `Installation und Administration`, and define the operator
  scope.
- `site/development/index.md` and `site/de/development/index.md`, audience `developer`: use
  `Development and Architecture` and `Entwicklung und Architektur`, and state that content follows
  `main`.
- `site/reference/index.md` and `site/de/reference/index.md`, audience `reference`: use `Reference`
  and `Referenz`, and name coverage, glossary, troubleshooting, and releases as later destinations.

Use this exact frontmatter shape on each non-home page:

```yaml
---
title: User Guide
description: Learn every stable RailKeeper workflow.
audience: user
status: stable
reviewedVersion: 0.1.17.5
lastReviewed: 2026-08-15
---
```

Use VitePress `layout: home`, a single restrained hero, and three audience feature cards on both
home pages. Do not use marketing claims, oversized typography, telemetry, or remote images.

- [ ] **Step 7: Build and preview the site**

Run:

```powershell
cd docs
npm.cmd run build
npm.cmd run preview -- --host 127.0.0.1
```

Expected: the build exits 0; the preview serves under `/RailKeeper/`; all five English paths and
five German paths return HTML; the language switch preserves the matching page path.

- [ ] **Step 8: Commit the foundation**

```powershell
git add .gitignore docs/package.json docs/package-lock.json docs/.vitepress docs/site
git commit -m "docs: scaffold bilingual VitePress site"
```

### Task 2: Enforce language pairing, metadata, and version consistency

**Files:**

- Create: `docs/versions.json`
- Create: `docs/scripts/validate-docs.mjs`
- Create: `docs/scripts/validate-docs.test.mjs`
- Modify: `docs/package.json`

**Interfaces:**

- Consumes: Markdown files below `docs/site/` and `docs/versions.json`.
- Produces: `validateDocumentTree(root, versions): Promise<string[]>`; CLI exit 0 for a valid tree,
  exit 1 with one line per defect otherwise.

- [ ] **Step 1: Add the canonical documentation versions**

Create `docs/versions.json`:

```json
{
  "stable": "0.1.17.5",
  "development": "main"
}
```

- [ ] **Step 2: Write validator tests first**

Create `docs/scripts/validate-docs.test.mjs`. Use `node:test`, `node:assert/strict`,
`fs/promises.mkdtemp`, and `os.tmpdir`. Cover these exact cases:

```js
test("accepts a complete matching language pair", async () => {
  const root = await fixture({
    "index.md": page("reference", "stable", "0.1.17.5"),
    "de/index.md": page("reference", "stable", "0.1.17.5"),
  });
  assert.deepEqual(await validateDocumentTree(root, versions), []);
});

test("reports a missing German counterpart", async () => {
  const root = await fixture({ "guide/index.md": page("user", "stable", "0.1.17.5") });
  assert.deepEqual(await validateDocumentTree(root, versions), [
    "guide/index.md: missing counterpart de/guide/index.md",
  ]);
});

test("reports missing and mismatched metadata", async () => {
  const root = await fixture({
    "index.md": page("reference", "stable", "0.1.17.5"),
    "de/index.md": page("developer", "development", "main"),
  });
  const errors = await validateDocumentTree(root, versions);
  assert(errors.includes("index.md: audience differs from de/index.md"));
  assert(errors.includes("index.md: status differs from de/index.md"));
  assert(errors.includes("index.md: reviewedVersion differs from de/index.md"));
});

test("enforces the canonical version for each status", async () => {
  const root = await fixture({
    "index.md": page("reference", "stable", "0.1.14"),
    "de/index.md": page("reference", "stable", "0.1.14"),
  });
  assert((await validateDocumentTree(root, versions)).some((error) =>
    error.includes("reviewedVersion must be 0.1.17.5"),
  ));
});
```

The local `fixture` helper writes exactly the supplied files. The local `page` helper emits valid
frontmatter with `lastReviewed: 2026-08-15` and a `# Test` body.

- [ ] **Step 3: Run tests and verify the expected failure**

Run:

```powershell
cd docs
npm.cmd test
```

Expected: FAIL because `validate-docs.mjs` does not exist.

- [ ] **Step 4: Implement the validator**

Implement these exported functions in `docs/scripts/validate-docs.mjs`:

```js
export function parseFrontmatter(source) // returns Record<string, string>
export async function markdownFiles(root) // returns sorted POSIX-style relative paths
export async function validateDocumentTree(root, versions) // returns sorted string[]
```

`validateDocumentTree` must:

1. pair every English path with `de/<path>` and every German path with its English path;
2. require `audience`, `status`, `reviewedVersion`, and `lastReviewed`;
3. allow audiences `user`, `admin`, `developer`, and `reference`;
4. allow statuses `stable` and `development`;
5. require ISO date syntax `YYYY-MM-DD` for `lastReviewed`;
6. require `versions.stable` for stable pages and `versions.development` for development pages;
7. require matching audience, status, reviewed version, and review date across each pair;
8. reject `TODO`, `TBD`, and `FIXME` in published Markdown;
9. print each error to stderr and set `process.exitCode = 1` when invoked as a CLI.

Do not add a YAML dependency. Parse only the flat scalar frontmatter contract above and fail on
duplicate keys or an unterminated block.

- [ ] **Step 5: Run validator tests and the real tree check**

```powershell
cd docs
npm.cmd test
node scripts/validate-docs.mjs
```

Expected: all validator tests PASS and the real-tree command exits 0 without stderr.

- [ ] **Step 6: Make validation part of the existing check command**

Keep this exact script in `docs/package.json`:

```json
"check": "npm test && node scripts/validate-docs.mjs && vitepress build"
```

- [ ] **Step 7: Commit the contract**

```powershell
git add docs/versions.json docs/scripts/validate-docs.mjs docs/scripts/validate-docs.test.mjs docs/package.json
git commit -m "test: enforce bilingual documentation contract"
```

### Task 3: Discover repository documentation surfaces deterministically

**Files:**

- Create: `docs/scripts/source-inventory.mjs`
- Create: `docs/scripts/source-inventory.test.mjs`

**Interfaces:**

- Consumes: `frontend/src/app/App.tsx`, `frontend/src/app/Shell.tsx`, both translation maps,
  `backend/internal/api/routes.go`, `openapi/railkeeper.yaml`, tracked source files, and legacy docs.
- Produces: `buildSourceInventory(repositoryRoot): Promise<SourceInventory>` and deterministic JSON
  on stdout. It writes no repository file.

- [ ] **Step 1: Write parser tests first**

Create tests that assert these exact transformations:

```js
assert.deepEqual(extractFrontendRoutes('startsWith("/vehicles")\nhref: "/overview"'), [
  "/overview",
  "/vehicles",
]);
assert.deepEqual(extractTranslationKeys('  "vehicles.cv.title": "CV",'), ["vehicles.cv.title"]);
assert.deepEqual(
  extractApiRoutes('{http.MethodGet, "/api/v1/vehicles", routeAccessViewer, handler, nil}'),
  [{ access: "Viewer", method: "GET", path: "/api/v1/vehicles" }],
);
assert.deepEqual(extractOpenApiPaths("paths:\n  /vehicles:\n    get:\n    post:\n"), [
  "GET /vehicles",
  "POST /vehicles",
]);
assert.deepEqual(extractEnvironmentVariables('env("RAILKEEPER_ADDR", ":8080")'), [
  "RAILKEEPER_ADDR",
]);
```

Add one integration fixture proving that English and German translation key sets must match and
that the returned object is stable across two runs.

- [ ] **Step 2: Run tests and verify the expected failure**

```powershell
cd docs
npm.cmd test
```

Expected: FAIL because `source-inventory.mjs` does not exist.

- [ ] **Step 3: Implement the read-only inventory builder**

Export these functions from `docs/scripts/source-inventory.mjs`:

```js
export function extractFrontendRoutes(source)
export function extractTranslationKeys(source)
export function extractApiRoutes(source)
export function extractOpenApiPaths(source)
export function extractEnvironmentVariables(source)
export async function buildSourceInventory(repositoryRoot)
```

Return this exact top-level shape, with every array deduplicated and sorted and without timestamps:

```json
{
  "schemaVersion": 1,
  "frontendRoutes": [],
  "translationKeys": [],
  "translationNamespaces": [],
  "apiRoutes": [],
  "openApiOperations": [],
  "environmentVariables": [],
  "legacyDocuments": []
}
```

Implementation rules:

- Collect routes from both `App.tsx` `startsWith()` calls and `Shell.tsx` `href` values.
- Keep full translation keys and derive top-level namespaces from the first dot-separated segment.
- Throw `English and German translation keys differ` and list both missing sets if they diverge.
- Convert Go method constants to uppercase HTTP names and strip `routeAccess` from access names.
- Parse only method keys below OpenAPI path keys; ignore schema properties named like methods.
- Discover `RAILKEEPER_*` names from tracked text files while excluding `.git`, `node_modules`,
  `dist`, `.cache`, `data`, and `.superpowers`.
- List the root README files, stable `docs/*.md`, and `docs/releases/*.md` as legacy documents.
- When invoked directly, resolve the repository root two levels above the script and print formatted
  JSON to stdout. Do not write a generated inventory into the repository.

- [ ] **Step 4: Run the parser and inspect real-source invariants**

```powershell
cd docs
npm.cmd test
node scripts/source-inventory.mjs
```

Expected: tests PASS; English and German translation counts are equal and nonzero, 185 API route
specifications are present at the current committed baseline, all 140 OpenAPI path groups are
represented after method expansion, and every visible/direct app route and tracked
`RAILKEEPER_*` variable appears. Do not encode observed counts in the implementation.

- [ ] **Step 5: Commit the inventory tool**

```powershell
git add docs/scripts/source-inventory.mjs docs/scripts/source-inventory.test.mjs
git commit -m "test: inventory documentation source surfaces"
```

### Task 4: Add the reviewed coverage manifest and public coverage view

**Files:**

- Create: `docs/coverage.json`
- Modify: `docs/scripts/validate-docs.mjs`
- Modify: `docs/scripts/validate-docs.test.mjs`
- Create: `docs/site/reference/coverage.md`
- Create: `docs/site/de/reference/coverage.md`
- Modify: `docs/.vitepress/config.mts`

**Interfaces:**

- Consumes: `buildSourceInventory()` from Task 3 and ownership maps from
  `docs/coverage.json`.
- Produces: `validateCoverage(inventory, manifest): string[]`; public coverage status pages at
  `/reference/coverage` and `/de/reference/coverage`.

- [ ] **Step 1: Write failing coverage tests**

Add tests proving that `validateCoverage` reports:

```text
frontend route /unmapped is not covered
translation key settings.newArea.title is not covered
API route GET /api/v1/unmapped is not covered
environment variable RAILKEEPER_UNMAPPED is not covered
coverage topic vehicles references missing English page guide/vehicles/index.md
```

Also add one passing fixture in which each source item matches exactly one topic.

- [ ] **Step 2: Run the focused test and verify failure**

```powershell
cd docs
node --test scripts/validate-docs.test.mjs
```

Expected: FAIL because `validateCoverage` is not exported.

- [ ] **Step 3: Create the initial reviewed manifest**

Create `docs/coverage.json` with `schemaVersion: 1` and these exact topic IDs:

```text
setup-auth, overview, vehicles-core, vehicle-media, vehicle-maintenance,
vehicle-decoder-cv, vehicle-search-spares, accessories, exhibition,
import-export, settings-general, master-data, users-sessions-security,
backup-restore, digital-centers, system-operations, layouts-unpublished,
deployment-configuration, development-architecture, releases-support,
shared-navigation
```

Each topic object contains only ownership metadata and its paired destination:

```json
{
  "id": "vehicles-core",
  "audience": "user",
  "status": "planned",
  "englishPath": "guide/vehicles/index.md",
  "germanPath": "de/guide/vehicles/index.md"
}
```

Add a top-level `owners` object. Use exact matching for frontend routes, top-level translation
prefixes, and environment variables. Use longest-prefix matching for API paths and detailed
translation prefixes. Include these complete owner assignments:

```json
{
  "frontendRoutes": {
    "/overview": "overview",
    "/vehicles": "vehicles-core",
    "/accessories": "accessories",
    "/layouts": "layouts-unpublished",
    "/exhibition": "exhibition",
    "/import-export": "import-export",
    "/settings": "settings-general"
  },
  "translationPrefixes": {
    "accessories": "accessories",
    "app": "setup-auth",
    "auth": "setup-auth",
    "common": "shared-navigation",
    "exhibition": "exhibition",
    "importExport": "import-export",
    "layouts": "layouts-unpublished",
    "nav": "shared-navigation",
    "overview": "overview",
    "settings.articleManagement": "master-data",
    "settings.articleSearch": "vehicle-search-spares",
    "settings.audit": "users-sessions-security",
    "settings.auditAction": "users-sessions-security",
    "settings.auth": "users-sessions-security",
    "settings.backup": "backup-restore",
    "settings.digital": "digital-centers",
    "settings.inventoryNumbers": "master-data",
    "settings.master": "master-data",
    "settings.masterTransfer": "master-data",
    "settings.password": "users-sessions-security",
    "settings.role": "users-sessions-security",
    "settings.roles": "users-sessions-security",
    "settings.session": "users-sessions-security",
    "settings.sessions": "users-sessions-security",
    "settings.smtp": "users-sessions-security",
    "settings.storage": "system-operations",
    "settings.updates": "releases-support",
    "settings.users": "users-sessions-security",
    "settings": "settings-general",
    "setup": "setup-auth",
    "vehicle": "vehicles-core",
    "vehicles.articleSearch": "vehicle-search-spares",
    "vehicles.barcode": "vehicle-search-spares",
    "vehicles.cv": "vehicle-decoder-cv",
    "vehicles.functionMode": "vehicle-decoder-cv",
    "vehicles.functionType": "vehicle-decoder-cv",
    "vehicles.functions": "vehicle-decoder-cv",
    "vehicles.image": "vehicle-media",
    "vehicles.imageCare": "vehicle-media",
    "vehicles.imagePreview": "vehicle-media",
    "vehicles.maintenance": "vehicle-maintenance",
    "vehicles.search": "vehicle-search-spares",
    "vehicles.spareParts": "vehicle-search-spares",
    "vehicles.speedCurve": "vehicle-decoder-cv",
    "vehicles.uploads": "vehicle-media",
    "vehicles": "vehicles-core"
  },
  "apiPrefixes": {
    "/health": "system-operations",
    "/api/v1/version": "releases-support",
    "/api/v1/system/digital-settings": "digital-centers",
    "/api/v1/system/smtp": "users-sessions-security",
    "/api/v1/system": "system-operations",
    "/api/v1/setup": "setup-auth",
    "/api/v1/auth/two-factor": "users-sessions-security",
    "/api/v1/auth/password": "users-sessions-security",
    "/api/v1/auth": "setup-auth",
    "/api/v1/profile": "settings-general",
    "/api/v1/roles": "users-sessions-security",
    "/api/v1/users": "users-sessions-security",
    "/api/v1/sessions": "users-sessions-security",
    "/api/v1/ecos": "digital-centers",
    "/api/v1/digital-centers": "digital-centers",
    "/api/v1/vehicles/{id}/images": "vehicle-media",
    "/api/v1/vehicles/{id}/attachments": "vehicle-media",
    "/api/v1/vehicles/{id}/maintenance": "vehicle-maintenance",
    "/api/v1/vehicles/{id}/spare-parts": "vehicle-search-spares",
    "/api/v1/vehicles/{id}/functions": "vehicle-decoder-cv",
    "/api/v1/vehicles/{id}/cv-values": "vehicle-decoder-cv",
    "/api/v1/vehicles/{id}/cv-files": "vehicle-decoder-cv",
    "/api/v1/cv-files": "vehicle-decoder-cv",
    "/api/v1/article-search": "vehicle-search-spares",
    "/api/v1/vehicles": "vehicles-core",
    "/api/v1/accessory": "accessories",
    "/api/v1/storage-locations": "accessories",
    "/api/v1/layout": "layouts-unpublished",
    "/api/v1/plan-": "layouts-unpublished",
    "/api/v1/track-": "layouts-unpublished",
    "/api/v1/inventory-number-schemes": "master-data",
    "/api/v1/master-data": "master-data",
    "/api/v1/backup": "backup-restore",
    "/api/v1/exhibition-lists": "exhibition"
  },
  "environmentVariables": {
    "RAILKEEPER_ADDR": "deployment-configuration",
    "RAILKEEPER_ALLOWED_ATTACHMENT_EXTENSIONS": "deployment-configuration",
    "RAILKEEPER_COOKIE_SECURE": "deployment-configuration",
    "RAILKEEPER_DATA_DIR": "deployment-configuration",
    "RAILKEEPER_DEFAULT_PRINTER": "deployment-configuration",
    "RAILKEEPER_IMAGE": "deployment-configuration",
    "RAILKEEPER_MAX_ATTACHMENT_MB": "deployment-configuration",
    "RAILKEEPER_MAX_ATTACHMENT_BYTES": "deployment-configuration",
    "RAILKEEPER_MAX_IMAGE_MB": "deployment-configuration",
    "RAILKEEPER_MAX_IMAGE_BYTES": "deployment-configuration",
    "RAILKEEPER_MIGRATIONS_DIR": "deployment-configuration",
    "RAILKEEPER_OPEN_BROWSER": "deployment-configuration",
    "RAILKEEPER_PDF_OCR": "deployment-configuration",
    "RAILKEEPER_PDF_OCR_MAX_PAGES": "deployment-configuration",
    "RAILKEEPER_PORTABLE": "deployment-configuration",
    "RAILKEEPER_PRINTERS": "deployment-configuration",
    "RAILKEEPER_PUBLIC_URL": "deployment-configuration",
    "RAILKEEPER_SEEDS_DIR": "deployment-configuration",
    "RAILKEEPER_SMTP_FROM": "deployment-configuration",
    "RAILKEEPER_SMTP_HOST": "deployment-configuration",
    "RAILKEEPER_SMTP_PASSWORD": "deployment-configuration",
    "RAILKEEPER_SMTP_PORT": "deployment-configuration",
    "RAILKEEPER_SMTP_TLS": "deployment-configuration",
    "RAILKEEPER_SMTP_USER": "deployment-configuration",
    "RAILKEEPER_STATIC_DIR": "deployment-configuration",
    "RAILKEEPER_UPDATE_CHECK_URL": "deployment-configuration"
  }
}
```

The implementation must prefer the longest matching translation or API prefix. Every discovered
frontend route, full translation key, API route, and environment variable must resolve to exactly
one topic. Assign `layouts-unpublished` the status `not-published` and all other not-yet-authored
destinations `planned`.

- [ ] **Step 4: Implement coverage validation**

Add and export:

```js
export function validateCoverage(inventory, manifest, contentRoot)
```

It must reject unknown status/audience values, duplicate topic IDs, owner references to unknown
topics, source items with no owner, and `documented` topics whose paired target files do not exist.
Allowed coverage statuses are `planned`, `documented`, `internal`, and `not-published`. Call this
validation from the CLI after `validateDocumentTree`.

- [ ] **Step 5: Create paired human-readable coverage pages**

The English and German pages explain:

- what source surfaces are scanned;
- the difference between `planned`, `documented`, `internal`, and `not-published`;
- that a topic becomes `documented` only after both target pages pass the page standard;
- that layout planning remains `not-published` for stable users;
- how contributors run `npm.cmd run check` and update `docs/coverage.json`.

Both pages link to `docs/coverage.json` on GitHub and show the 20 topic IDs grouped under user,
admin, developer, and shared headings. Add the page to both reference sidebars.

- [ ] **Step 6: Run the complete documentation check**

```powershell
cd docs
npm.cmd run check
```

Expected: all tests PASS, coverage has no unmapped or multiply mapped source item, and VitePress
builds without dead links.

- [ ] **Step 7: Commit the coverage contract**

```powershell
git add docs/coverage.json docs/scripts/validate-docs.mjs docs/scripts/validate-docs.test.mjs
git add docs/site/reference docs/site/de/reference docs/.vitepress/config.mts
git commit -m "docs: add repository coverage matrix"
```

### Task 5: Add CI and GitHub Pages deployment

**Files:**

- Modify: `.github/workflows/ci.yml`
- Create: `.github/workflows/docs-pages.yml`

**Interfaces:**

- Consumes: `docs/package-lock.json` and `npm run check` from Tasks 1 through 4.
- Produces: required PR check `Documentation`; deployment environment `github-pages`.

- [ ] **Step 1: Add the documentation job to CI**

Append this job to `.github/workflows/ci.yml`:

```yaml
  documentation:
    name: Documentation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
      - uses: actions/setup-node@v7
        with:
          node-version: 24
          cache: npm
          cache-dependency-path: docs/package-lock.json
      - name: Install documentation dependencies
        working-directory: docs
        run: npm ci
      - name: Check documentation
        working-directory: docs
        run: npm run check
```

- [ ] **Step 2: Create the Pages workflow**

Create `.github/workflows/docs-pages.yml`:

```yaml
name: Documentation Pages

on:
  push:
    branches: [main]
    paths:
      - "docs/**"
      - "frontend/public/brand/**"
      - ".github/workflows/docs-pages.yml"
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: documentation-pages
  cancel-in-progress: false

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v7
        with:
          fetch-depth: 0
      - uses: actions/setup-node@v7
        with:
          node-version: 24
          cache: npm
          cache-dependency-path: docs/package-lock.json
      - uses: actions/configure-pages@v4
      - name: Install documentation dependencies
        working-directory: docs
        run: npm ci
      - name: Build documentation
        working-directory: docs
        run: npm run check
      - uses: actions/upload-pages-artifact@v3
        with:
          path: docs/.vitepress/dist

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    needs: build
    runs-on: ubuntu-latest
    steps:
      - name: Deploy documentation
        id: deployment
        uses: actions/deploy-pages@v4
```

- [ ] **Step 3: Validate workflow syntax and the local equivalent**

Run:

```powershell
cd docs
npm.cmd ci
npm.cmd run check
```

Expected: clean install succeeds, all documentation checks pass, and
`docs/.vitepress/dist/index.html` plus `docs/.vitepress/dist/de/index.html` exist.

- [ ] **Step 4: Commit workflows**

```powershell
git add .github/workflows/ci.yml .github/workflows/docs-pages.yml
git commit -m "ci: publish RailKeeper documentation pages"
```

### Task 6: Add repository entry points and contributor rules

**Files:**

- Modify: `README.md`
- Modify: `README.de.md`
- Modify: `CONTRIBUTING.md`
- Modify: `SUPPORT.md`

**Interfaces:**

- Consumes: published routes from Task 1.
- Produces: discoverable handbook links and a same-PR documentation rule.

- [ ] **Step 1: Add language-correct handbook links**

Add `Documentation` to the top link row in `README.md` and `Dokumentation` to `README.de.md`:

```text
https://ichwars.github.io/RailKeeper/
https://ichwars.github.io/RailKeeper/de/
```

Add one sentence near each feature overview stating that the handbook contains complete user,
operator, and developer guidance. Do not expand the README into a second handbook.

- [ ] **Step 2: Document the contribution workflow**

Add a `Documentation changes` section to `CONTRIBUTING.md` with these rules:

```markdown
## Documentation changes

- Update the English and German page in the same pull request when behavior changes.
- Keep identical relative paths below `docs/site/` and `docs/site/de/`.
- Run `cd docs` followed by `npm run check` before opening the pull request.
- Update `docs/coverage.json` when a route, API area, translation namespace, configuration value,
  or documentation topic is introduced.
- A one-language-only change is allowed only for a wording correction that does not change meaning.
```

- [ ] **Step 3: Route support requests without duplicating troubleshooting content**

Add English and German handbook links to `SUPPORT.md`. Keep issue-reporting and security-reporting
instructions in `SUPPORT.md`; link to handbook troubleshooting instead of copying it.

- [ ] **Step 4: Run link/build checks and commit**

```powershell
cd docs
npm.cmd run check
cd ..
git add README.md README.de.md CONTRIBUTING.md SUPPORT.md
git commit -m "docs: link the RailKeeper handbook"
```

Expected: documentation checks pass and the only duplicated URLs are the intended language-specific
entry points.

### Task 7: Correct the stable GitHub release marker and perform end-to-end acceptance

**Files:**

- Modify only if evidence requires correction: `docs/versions.json`
- Modify only if evidence requires correction: paired home pages
- No release artifact is created locally or committed.

**Interfaces:**

- Consumes: tag `v0.1.17.5`, `.github/workflows/windows-portable.yml`, public GitHub Releases, and
  the deployed Pages site.
- Produces: verified public latest release `v0.1.17.5` and an accepted Etappe 1 deployment.

- [ ] **Step 1: Verify the local and remote tag before changing GitHub state**

Run:

```powershell
git show -s --format="%H %D %cs %s" v0.1.17.5
git ls-remote --tags origin refs/tags/v0.1.17.5
gh release view v0.1.17.5 --repo ichwars/RailKeeper --json tagName,isDraft,isPrerelease,assets,url
```

Expected: the local tag resolves to commit `9e81856313c759f733d97405bb9fa606d47e74b5`; the remote tag
exists; the release is non-draft and non-prerelease with the portable ZIP asset. If the remote tag
is absent, confirm the local tag still points to that exact reviewed commit, then run:

```powershell
git push origin refs/tags/v0.1.17.5
```

Stop instead of pushing if the commit differs.

- [ ] **Step 2: Recover a missing release only through the existing release workflow**

If the release or portable ZIP is missing, run:

```powershell
gh workflow run windows-portable.yml --repo ichwars/RailKeeper --ref v0.1.17.5
$releaseRunId = gh run list --repo ichwars/RailKeeper `
  --workflow windows-portable.yml --limit 1 --json databaseId --jq '.[0].databaseId'
gh run watch --repo ichwars/RailKeeper $releaseRunId --exit-status
```

Expected: the workflow succeeds and publishes the ZIP to a non-prerelease v0.1.17.5 release. Do
not upload an unverified local ZIP manually.

- [ ] **Step 3: Mark the verified release as latest and confirm the public API**

```powershell
gh release edit v0.1.17.5 --repo ichwars/RailKeeper --latest
gh api repos/ichwars/RailKeeper/releases/latest --jq .tag_name
```

Expected: `v0.1.17.5`.

- [ ] **Step 4: Enable GitHub Pages from Actions if not already enabled**

In repository settings, set Pages source to GitHub Actions. Then run the workflow manually once:

```powershell
gh workflow run docs-pages.yml --repo ichwars/RailKeeper --ref main
$pagesRunId = gh run list --repo ichwars/RailKeeper `
  --workflow docs-pages.yml --limit 1 --json databaseId --jq '.[0].databaseId'
gh run watch --repo ichwars/RailKeeper $pagesRunId --exit-status
```

Expected: the deployment succeeds and reports the Pages URL.

- [ ] **Step 5: Perform visual and functional acceptance**

Check the deployed site at 390 px, 768 px, and 1440 px widths in both light and dark modes:

- English opens at `/RailKeeper/`; German opens at `/RailKeeper/de/`.
- Switching language from every landing and coverage page preserves the corresponding path.
- Search finds `backup`, `vehicle`, `Wartung`, and `Sicherung` only in the appropriate locale.
- Navigation exposes the three audience areas and reference without clipped German labels.
- The logo loads from the repository asset and no external request is made for fonts, search, or
  analytics.
- Missing pages return the VitePress 404 page, not the RailKeeper application shell.
- All page edit links target the correct file below `docs/site/`.

- [ ] **Step 6: Run the complete repository baseline**

```powershell
cd backend
go test ./...
cd ..\frontend
npm.cmd run build
cd ..\docs
npm.cmd run check
```

Expected: all Go tests pass, the frontend production build succeeds, and every documentation test
and build succeeds.

- [ ] **Step 7: Record Etappe 1 acceptance**

Update the implementation pull request description with:

```text
- Documentation check: passed
- Backend tests: passed
- Frontend build: passed
- GitHub Pages deployment: passed
- English/German path parity: passed
- Light/dark and 390/768/1440 px visual checks: passed
- Latest stable GitHub Release: v0.1.17.5
- Remaining content work: Etappen 2 through 5 from the approved design specification
```

No additional source commit is required unless verification finds and fixes a defect.
