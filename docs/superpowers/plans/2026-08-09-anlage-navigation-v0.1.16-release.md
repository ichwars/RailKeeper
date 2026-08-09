# Anlage Navigation And v0.1.16 Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Disable the layout menu without removing its routes, publish bilingual release documentation, and ship the verified Stage 1 work as stable version `0.1.16`.

**Architecture:** A small navigation-availability module is the single source of truth for temporarily disabled views. Shell rendering, configured start-view resolution, and the settings selector consume that module while the existing `/layouts` route and backend remain unchanged. Release preparation stays on `dev/stage1-acceptance` until local and GitHub checks pass, then a merge commit is tagged and distributed by the existing GitHub workflows.

**Tech Stack:** React 19, TypeScript 7, Vitest, Testing Library, CSS design tokens, Go 1.26, Git, GitHub CLI, GitHub Actions, Docker/GHCR

## Global Constraints

- Work only inline in the existing worktree. Do not dispatch subagents.
- Preserve all user changes and keep `.superpowers/`, `frontend/dist`, caches, runtime data, and credentials uncommitted.
- Use `Anlage` in German and `Layout` in English for the disabled main-navigation item.
- Keep `/layouts`, the React layout workspace, backend routes, data models, and stored layout data unchanged.
- A stored `layouts` start view must fall back to `overview` without overwriting the stored preference.
- Keep `README.md` in English and add a complete `README.de.md` translation.
- Add equivalent `CHANGELOG.md` and `CHANGELOG.de.md` files beginning with stable version `0.1.16`.
- Do not rewrite the published `dev/stage1-acceptance` history. Integrate `origin/main` with a merge.
- Create `v0.1.16` only after the PR and all required checks pass.
- Never overwrite an existing remote tag or GitHub release.

---

### Task 1: Integrate the Current Remote Main Branch

**Files:**
- Verify only, no source files expected

**Interfaces:**
- Consumes: clean tracked state on `dev/stage1-acceptance`
- Produces: a branch containing the latest `origin/main` without rewritten Stage 1 commits

- [ ] **Step 1: Verify branch and tracked worktree state**

Run:

```powershell
git branch --show-current
git status --short
```

Expected: branch is `dev/stage1-acceptance`; only the known untracked `.superpowers/` directory may be listed.

- [ ] **Step 2: Refresh authoritative remote state**

Run:

```powershell
git fetch origin main --tags
git rev-list --left-right --count origin/main...HEAD
git ls-remote --tags origin refs/tags/v0.1.16
gh release view v0.1.16
```

Expected: the divergence count is reported; both tag and release lookups return no existing `v0.1.16`. Stop rather than overwrite if either exists.

- [ ] **Step 3: Merge `origin/main` without rewriting history**

Run:

```powershell
git merge --no-edit origin/main
```

Expected: merge succeeds. If conflicts occur, inspect every conflicted file, preserve both the remote dependency/security changes and Stage 1 behavior, then finish with a merge commit.

- [ ] **Step 4: Verify the integrated baseline**

Run:

```powershell
Set-Location backend
go test ./...
Set-Location ..\frontend
npm.cmd run test:run
npm.cmd run build
Set-Location ..
git diff --check
```

Expected: all Go and frontend tests pass, the production build succeeds, and `git diff --check` is empty.

### Task 2: Define And Render Temporarily Disabled Navigation

**Files:**
- Create: `frontend/src/app/navigationAvailability.ts`
- Create: `frontend/src/app/navigationAvailability.test.ts`
- Modify: `frontend/src/app/Shell.tsx`
- Modify: `frontend/src/app/Shell.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/base.css`

**Interfaces:**
- Consumes: `AppView` from `frontend/src/app/App.tsx`
- Produces: `isViewTemporarilyDisabled(view: AppView): boolean` and `availableStartView(value: string): AppView`

- [ ] **Step 1: Write failing availability and shell tests**

Create `navigationAvailability.test.ts`:

```ts
import { describe, expect, it } from "vitest";

import { availableStartView, isViewTemporarilyDisabled } from "./navigationAvailability";

describe("navigation availability", () => {
  it("marks only the layout workspace as temporarily disabled", () => {
    expect(isViewTemporarilyDisabled("layouts")).toBe(true);
    expect(isViewTemporarilyDisabled("accessories")).toBe(false);
  });

  it.each([
    ["layouts", "overview"],
    ["inventory", "vehicles"],
    ["accessories", "accessories"],
    ["invalid", "overview"]
  ])("maps configured start view %s to %s", (stored, expected) => {
    expect(availableStartView(stored)).toBe(expected);
  });
});
```

Extend `Shell.test.tsx` with:

```tsx
it("shows the singular layout item but does not expose it as a link", async () => {
  render(
    <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
      <p>Inhalt</p>
    </Shell>
  );

  const label = screen.getByText("Anlage");
  const disabledItem = label.closest("[aria-disabled='true']");
  expect(disabledItem).toHaveClass("disabled");
  expect(disabledItem).toHaveAttribute("title", "Anlage ist vorübergehend nicht verfügbar.");
  expect(screen.queryByRole("link", { name: "Anlage" })).not.toBeInTheDocument();
  await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
});

it("uses the singular English layout label and hint", async () => {
  window.localStorage.setItem("railkeeper.settings.language", "en");
  render(
    <Shell username="editor" roles={["Editor"]} activeView="accessories" onLogout={vi.fn()}>
      <p>Content</p>
    </Shell>
  );

  const disabledItem = screen.getByText("Layout").closest("[aria-disabled='true']");
  expect(disabledItem).toHaveAttribute("title", "Layout is temporarily unavailable.");
  expect(screen.queryByRole("link", { name: "Layout" })).not.toBeInTheDocument();
  await waitFor(() => expect(api.profileSettings).toHaveBeenCalledOnce());
});
```

- [ ] **Step 2: Run tests and verify the intended failure**

Run:

```powershell
Set-Location frontend
npm.cmd run test:run -- src/app/navigationAvailability.test.ts src/app/Shell.test.tsx
```

Expected: FAIL because `navigationAvailability.ts` does not exist and the shell still renders `Anlagen` as a link.

- [ ] **Step 3: Implement the central availability module**

Create `navigationAvailability.ts`:

```ts
import type { AppView } from "./App";

const appViews = ["overview", "vehicles", "accessories", "layouts", "exhibition", "importExport", "settings"] as const;
const temporarilyDisabledViews = new Set<AppView>(["layouts"]);

function isAppView(value: string): value is AppView {
  return appViews.some((view) => view === value);
}

export function isViewTemporarilyDisabled(view: AppView): boolean {
  return temporarilyDisabledViews.has(view);
}

export function availableStartView(value: string): AppView {
  if (value === "inventory") return "vehicles";
  if (!isAppView(value) || isViewTemporarilyDisabled(value)) return "overview";
  return value;
}
```

- [ ] **Step 4: Render the layout item as non-interactive content**

Import `isViewTemporarilyDisabled` in `Shell.tsx`. Replace the navigation map body with this structure:

```tsx
{orderedNavItems.map((item) => {
  const Icon = item.icon;
  const label = t(item.labelKey);
  if (isViewTemporarilyDisabled(item.view)) {
    return (
      <span
        key={item.view}
        className="nav-entry disabled"
        aria-disabled="true"
        title={t("nav.layouts.disabled")}
      >
        <Icon size={16} aria-hidden="true" />
        <span>{label}</span>
      </span>
    );
  }
  return (
    <a
      key={item.view}
      className={`nav-entry${activeView === item.view ? " active" : ""}`}
      href={item.href}
      onClick={() => setMobileMenuOpen(false)}
    >
      <Icon size={16} aria-hidden="true" />
      <span>{label}</span>
    </a>
  );
})}
```

Change the translation values and add the matching hint keys:

```ts
// de.ts
"nav.layouts": "Anlage",
"nav.layouts.disabled": "Anlage ist vorübergehend nicht verfügbar.",

// en.ts
"nav.layouts": "Layout",
"nav.layouts.disabled": "Layout is temporarily unavailable.",
```

- [ ] **Step 5: Style links and disabled items through one navigation class**

In `base.css`, replace the `.nav a` entry selectors with `.nav > .nav-entry`, keep active and hover rules restricted to anchors, and add:

```css
.nav > .nav-entry.disabled {
  color: var(--muted);
  cursor: not-allowed;
  opacity: 0.48;
  user-select: none;
}

.nav > a.nav-entry.active,
.nav > a.nav-entry:hover {
  background: var(--accent-soft);
  color: var(--accent-strong);
}

.nav > a.nav-entry.active {
  box-shadow: inset 3px 0 0 var(--accent);
}
```

Use `.sidebar.collapsed .nav > .nav-entry` and `.sidebar.collapsed .nav > .nav-entry span` for the existing collapsed selectors so the disabled icon follows the same layout without exposing the hidden text.

- [ ] **Step 6: Run targeted tests and commit**

Run:

```powershell
npm.cmd run test:run -- src/app/navigationAvailability.test.ts src/app/Shell.test.tsx src/shared/i18n.test.ts
Set-Location ..
git diff --check
git add frontend/src/app/navigationAvailability.ts frontend/src/app/navigationAvailability.test.ts frontend/src/app/Shell.tsx frontend/src/app/Shell.test.tsx frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts frontend/src/styles/base.css
git commit -m "feat: disable layout navigation"
```

Expected: all targeted tests pass; translation key parity remains intact; commit contains only navigation files.

### Task 3: Enforce The Start-View Fallback Without Disabling Direct Routes

**Files:**
- Create: `frontend/src/app/App.test.tsx`
- Modify: `frontend/src/app/App.tsx`
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`

**Interfaces:**
- Consumes: `availableStartView(value: string): AppView` and `isViewTemporarilyDisabled(view: AppView): boolean`
- Produces: exported `configuredStartView(): AppView` and `currentView(): AppView` for focused routing tests

- [ ] **Step 1: Write failing routing and settings tests**

Create `App.test.tsx`:

```ts
import { beforeEach, describe, expect, it } from "vitest";

import { configuredStartView, currentView } from "./App";

describe("App navigation availability", () => {
  beforeEach(() => {
    window.history.replaceState(null, "", "/");
  });

  it("falls back from a stored layout start view without overwriting it", () => {
    window.localStorage.setItem("railkeeper.settings.defaultView", "layouts");

    expect(configuredStartView()).toBe("overview");
    expect(window.localStorage.getItem("railkeeper.settings.defaultView")).toBe("layouts");
  });

  it("keeps the direct layout route available", () => {
    window.history.replaceState(null, "", "/layouts");
    expect(currentView()).toBe("layouts");
  });
});
```

Extend `SettingsView.test.tsx`:

```tsx
it("disables Layout as a start page and displays Overview for a stored layout preference", async () => {
  window.localStorage.setItem("railkeeper.settings.defaultView", "layouts");
  render(<SettingsView username="viewer" />);

  const select = await screen.findByRole("combobox", { name: "Startseite" });
  expect(select).toHaveValue("overview");
  expect(within(select).getByRole("option", { name: "Anlage" })).toBeDisabled();
  expect(window.localStorage.getItem("railkeeper.settings.defaultView")).toBe("layouts");
});
```

- [ ] **Step 2: Run the tests and verify failure**

Run:

```powershell
Set-Location frontend
npm.cmd run test:run -- src/app/App.test.tsx src/features/settings/SettingsView.test.tsx
```

Expected: FAIL because the App helpers are not exported, `layouts` is still returned as the start view, and the settings option is enabled.

- [ ] **Step 3: Route stored preferences through the availability module**

In `App.tsx`, import `availableStartView`, export the two focused helpers, and keep the direct route check before the configured fallback:

```ts
export function configuredStartView(): AppView {
  return availableStartView(window.localStorage.getItem(defaultViewSettingKey) || "");
}

export function currentView(): AppView {
  if (window.location.pathname.startsWith("/overview")) return "overview";
  if (window.location.pathname.startsWith("/vehicles")) return "vehicles";
  if (window.location.pathname.startsWith("/accessories")) return "accessories";
  if (window.location.pathname.startsWith("/layouts")) return "layouts";
  if (window.location.pathname.startsWith("/exhibition")) return "exhibition";
  if (window.location.pathname.startsWith("/import-export")) return "importExport";
  if (window.location.pathname.startsWith("/settings")) return "settings";
  return configuredStartView();
}
```

- [ ] **Step 4: Disable the settings option and normalize only displayed state**

Import `availableStartView` and `isViewTemporarilyDisabled` in `SettingsView.tsx`. Initialize and hydrate state through `availableStartView`:

```ts
const [defaultView, setDefaultView] = useState(() =>
  availableStartView(readLocalSetting(localSettingKeys.defaultView, "overview"))
);
```

Use the same function for the profile setting callback:

```ts
store(localSettingKeys.defaultView, (value) => setDefaultView(availableStartView(value)));
```

Disable the existing option without removing it:

```tsx
<option value="layouts" disabled={isViewTemporarilyDisabled("layouts")}>
  {t("nav.layouts")}
</option>
```

Do not call `setLocalSetting` during normalization. This preserves a previously stored `layouts` value.

- [ ] **Step 5: Run focused tests and commit**

Run:

```powershell
npm.cmd run test:run -- src/app/App.test.tsx src/app/navigationAvailability.test.ts src/features/settings/SettingsView.test.tsx
Set-Location ..
git diff --check
git add frontend/src/app/App.tsx frontend/src/app/App.test.tsx frontend/src/features/settings/SettingsView.tsx frontend/src/features/settings/SettingsView.test.tsx
git commit -m "fix: fall back from disabled layout start view"
```

Expected: tests pass, `/layouts` remains directly resolvable, and no stored setting is changed by reading it.

### Task 4: Prepare Version And Bilingual Release Documentation

**Files:**
- Modify: `backend/cmd/railkeeper/main.go`
- Modify: `README.md`
- Create: `README.de.md`
- Create: `CHANGELOG.md`
- Create: `CHANGELOG.de.md`
- Create: `docs/releases/v0.1.16.md`

**Interfaces:**
- Consumes: verified Stage 1 commit history and the release specification
- Produces: application version `0.1.16` and bilingual public release text

- [ ] **Step 1: Bump the runtime version**

Change only the version constant:

```go
const (
	version               = "0.1.16"
	defaultUpdateCheckURL = "https://api.github.com/repos/ichwars/RailKeeper/releases/latest"
)
```

Run:

```powershell
gofmt -w backend/cmd/railkeeper/main.go
```

- [ ] **Step 2: Update the English README**

Add this language selector below the badge block:

```html
<p align="center">
  <strong>English</strong> · <a href="README.de.md">Deutsch</a>
</p>
```

Add these exact capabilities to `Highlights`:

```markdown
- Article inventory with generated inventory numbers, sortable selection-first overview, quantity or individual tracking, storage-location stock, reservations, installations, documents and usage history
- Layout, module, setup and plan-revision foundation with version-conflict protection; its main-navigation entry remains temporarily disabled while the workspace is refined
```

Add `Article Overview` to the operational screen list and update the pinned Docker example to:

```env
RAILKEEPER_IMAGE=ghcr.io/ichwars/railkeeper:v0.1.16
```

- [ ] **Step 3: Create the complete German README**

Create `README.de.md` as a section-for-section German translation of `README.md`, preserving every command, environment variable, path, image, link, security statement, and license statement verbatim where it is technical. Use these exact heading mappings:

```text
Overview -> Überblick
Highlights -> Funktionsumfang
Screens -> Ansichten
Quick Start -> Schnellstart
Windows Portable -> Windows Portable
Docker Compose -> Docker Compose
Update an existing Docker installation -> Bestehende Docker-Installation aktualisieren
Optional environment file -> Optionale Umgebungsdatei
Local Development -> Lokale Entwicklung
Architecture -> Architektur
Security -> Sicherheit
Counters And Badges -> Zähler und Badges
License -> Lizenz
Support -> Unterstützung
```

The German language selector must be:

```html
<p align="center">
  <a href="README.md">English</a> · <strong>Deutsch</strong>
</p>
```

Translate the two new feature bullets as:

```markdown
- Artikelbestand mit generierten Inventarnummern, sortierbarer Auswahlspalte, Mengen- oder Einzelverfolgung, Lagerortbeständen, Reservierungen, Einbauten, Dokumenten und Nutzungshistorie
- Fundament für Anlagen, Module, Aufbauten und Planrevisionen mit Versionskonfliktschutz; der Eintrag in der Hauptnavigation bleibt vorübergehend deaktiviert, solange der Arbeitsbereich weiter verfeinert wird
```

- [ ] **Step 4: Create equivalent English and German changelogs**

Create `CHANGELOG.md` and `CHANGELOG.de.md` with matching `0.1.16` sections dated `2026-08-09`. Both must contain equivalent entries for:

```text
Added: article catalogue and inventory, generated inventory numbers, stock locations and transactions,
       reservations/installations, documents/history, layout/module/setup/plan foundation, backup v2.
Changed: sortable selection-first article overview, quieter KPI hierarchy, master-data dropdowns,
         compact two-column article forms, consolidated settings data navigation, singular disabled
         Layout/Anlage navigation.
Fixed: transactional stock/allocation invariants, inactive attribute preservation, backup v1 import
       compatibility, input and document validation, focus restoration after confirmations.
Security: stricter API validation, protected master-data import invariants, conservative backup restore.
```

Use `Added`, `Changed`, `Fixed`, and `Security` headings in English and `Hinzugefügt`, `Geändert`, `Behoben`, and `Sicherheit` in German. Each changelog links to its translated counterpart directly below the title.

- [ ] **Step 5: Add the exact bilingual GitHub release body**

Create `docs/releases/v0.1.16.md` with this structure and no generated commit dump:

```markdown
# RailKeeper v0.1.16

## Deutsch

RailKeeper 0.1.16 führt den neuen Artikelbestand mit Inventarnummern, Lagerorten, Bestandsbuchungen,
Reservierungen, Einbauten, Dokumenten und Nutzungshistorie ein. Die Artikelübersicht und Dialoge
wurden für dichte Werkstattabläufe überarbeitet. Das Anlagenfundament ist technisch enthalten; der
Menüpunkt „Anlage“ bleibt bis zur nächsten Überarbeitungsstufe vorübergehend deaktiviert.

Weitere Details stehen im [deutschen Changelog](../../CHANGELOG.de.md).

## English

RailKeeper 0.1.16 introduces the new article inventory with inventory numbers, storage locations,
stock transactions, reservations, installations, documents, and usage history. The article overview
and dialogs were refined for dense workshop workflows. The layout foundation is included
technically; the “Layout” menu item remains temporarily disabled until the next refinement stage.

See the [English changelog](../../CHANGELOG.md) for details.
```

- [ ] **Step 6: Verify documentation consistency and commit**

Run:

```powershell
rg -n "0\.1\.14|0\.1\.15" README.md README.de.md CHANGELOG.md CHANGELOG.de.md docs/releases/v0.1.16.md backend/cmd/railkeeper/main.go
rg -n "Anlagen|Layouts" README.md README.de.md CHANGELOG.md CHANGELOG.de.md docs/releases/v0.1.16.md
git diff --check
Set-Location backend
go test ./cmd/railkeeper
Set-Location ..
git add backend/cmd/railkeeper/main.go README.md README.de.md CHANGELOG.md CHANGELOG.de.md docs/releases/v0.1.16.md
git commit -m "docs: prepare v0.1.16 release"
```

Expected: old stable-version references remain only where intentionally historical; plural navigation labels do not describe the disabled menu item; Go test and whitespace checks pass.

### Task 5: Run Full Local Release-Candidate Acceptance

**Files:**
- Verify all changed files
- Do not commit: `frontend/dist`, `.cache`, `.superpowers/`, local data, screenshots, or credentials

**Interfaces:**
- Consumes: complete `0.1.16` release candidate
- Produces: local evidence sufficient to open the release PR

- [ ] **Step 1: Run the full automated baseline**

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

Expected: all backend packages pass, all frontend test files pass, Vite builds successfully, no whitespace errors exist, and only `.superpowers/` plus ignored build/runtime output remain uncommitted.

- [ ] **Step 2: Refresh the local server with the built frontend**

Restart the existing RailKeeper process on `127.0.0.1:18083` using the worktree's `backend`, `frontend/dist`, migrations, seeds, and existing isolated data directory. Keep the process hidden. Then run:

```powershell
Invoke-RestMethod http://127.0.0.1:18083/health
Invoke-RestMethod http://127.0.0.1:18083/api/v1/version
```

Expected: `/health` returns HTTP 200 with `status: ok`; `/api/v1/version` reports `0.1.16`.

- [ ] **Step 3: Perform visual and interaction checks**

Use the existing authenticated in-app browser session and verify:

```text
Desktop dark: expanded sidebar shows muted Anlage with no hover highlight and no navigation.
Desktop dark: collapsed sidebar shows the muted icon and localized title.
Desktop light: both expanded and collapsed states preserve contrast without appearing active.
Mobile: opening the navigation shows Anlage disabled and does not close or navigate when touched.
Keyboard: Tab never focuses Anlage; Enter and Space cannot open /layouts from the menu.
Direct route: entering /layouts still loads the existing layout workspace.
Settings: Startseite shows Übersicht when layouts was stored; Anlage option is disabled.
```

Expected: all states match the specification with no browser-console errors.

- [ ] **Step 4: Smoke-test core Stage 1 flows**

In the local browser:

```text
1. Log in and open Artikelübersicht.
2. Sort at least two table columns and verify the selection column remains first.
3. Open an article, switch through Artikel, Bestand, Kauf & Dokumente, and Fachangaben: Gleis.
4. Verify the stock booking and transfer forms retain their compact layout.
5. Open Einstellungen > Import/Export, create a backup, select it for restore validation, and run only
   validation. Do not confirm or execute destructive restore.
6. Confirm the backup reports format version 2 and validates successfully.
```

Expected: no save, loading, layout, console, or validation regressions.

- [ ] **Step 5: Confirm the release-candidate commit set**

Run:

```powershell
git status --short
git log --oneline --decorate origin/main..HEAD
git diff --stat origin/main...HEAD
```

Expected: no tracked changes remain; `.superpowers/` remains untracked; the PR contains the intended Stage 1, navigation, version, and documentation commits.

### Task 6: Push, Review, Merge, And Synchronize Main

**Files:**
- Remote Git refs and GitHub PR only

**Interfaces:**
- Consumes: verified `dev/stage1-acceptance`
- Produces: merged PR and local `main` matching `origin/main`

- [ ] **Step 1: Push the complete branch**

Run:

```powershell
git push -u origin dev/stage1-acceptance
```

Expected: the remote branch points to the locally verified release-candidate commit.

- [ ] **Step 2: Create the release PR**

Create a PR against `main` with title:

```text
feat: ship article inventory stage 1
```

Create it with this exact body:

```powershell
$rkPrBody = @'
## Summary
- add layout, module, setup, and plan-revision foundations
- add complete article catalogue, inventory, stock, allocation, document, and history workflows
- add generated article inventory numbers and the sortable selection-first overview
- consolidate article master data under Settings > Data
- keep Layout/Anlage visible but temporarily disabled in main navigation
- prepare bilingual documentation and version 0.1.16

## Compatibility
- backup exports use format version 2
- backup format version 1 remains importable
- authentication data remains excluded from backups
- existing layout APIs and direct /layouts access remain available

## Verification
- go test ./...
- npm.cmd run test:run
- npm.cmd run build
- desktop/mobile, dark/light visual checks
- login, article, stock, and backup-validation smoke tests
'@
$rkPrUrl = gh pr create --base main --head dev/stage1-acceptance --title "feat: ship article inventory stage 1" --body $rkPrBody
$rkPr = gh pr view dev/stage1-acceptance --json number,url,state,mergeable,statusCheckRollup | ConvertFrom-Json
$rkPr
```

Expected: `$rkPrUrl` and `$rkPr.url` identify the same open PR.

- [ ] **Step 3: Wait for and inspect every required check**

Poll with:

```powershell
$rkPr = gh pr view dev/stage1-acceptance --json number,url,state | ConvertFrom-Json
gh pr checks $rkPr.number
```

Expected: every required check reaches `pass`. If any check fails, capture and inspect the newest failed run:

```powershell
$rkFailedRun = gh run list --branch dev/stage1-acceptance --status failure --limit 1 --json databaseId,url,workflowName | ConvertFrom-Json | Select-Object -First 1
gh run view $rkFailedRun.databaseId --log-failed
```

Fix it on the same branch, rerun all relevant local checks, commit, push, and wait again.

- [ ] **Step 4: Merge only the green PR**

Run:

```powershell
$rkPr = gh pr view dev/stage1-acceptance --json number | ConvertFrom-Json
gh pr merge $rkPr.number --merge
$rkMergedPr = gh pr view $rkPr.number --json state,mergedAt,mergeCommit,url | ConvertFrom-Json
$rkMergedPr
```

Expected: state is `MERGED`, `mergedAt` is set, and `mergeCommit.oid` identifies the exact release commit.

- [ ] **Step 5: Synchronize local main safely**

Run:

```powershell
git fetch origin main
git merge-base --is-ancestor main origin/main
git branch -f main origin/main
git -C C:\Users\droth\Documents\GitHub\RailKeeper status --short
git -C C:\Users\droth\Documents\GitHub\RailKeeper switch main
git -C C:\Users\droth\Documents\GitHub\RailKeeper status --short
git -C C:\Users\droth\Documents\GitHub\RailKeeper rev-parse HEAD
git rev-parse origin/main
```

Expected: the ancestry check succeeds before moving `main`; the primary worktree is clean on `main`; both final OIDs match. Keep `dev/layout-accessory-foundation-bootstrap` intact as a recoverable branch even though its `.gitignore` change is already present in the merged tree.

### Task 7: Tag And Verify The Stable Rollout

**Files:**
- Remote tag `v0.1.16`
- GitHub Release body sourced from `docs/releases/v0.1.16.md`
- Published Windows ZIP and GHCR images

**Interfaces:**
- Consumes: exact merge commit returned by the merged PR
- Produces: stable GitHub and container release `v0.1.16`

- [ ] **Step 1: Recheck release preconditions**

Run:

```powershell
git fetch origin main --tags
git ls-remote --tags origin refs/tags/v0.1.16
gh release view v0.1.16
git rev-parse origin/main
```

Expected: no `v0.1.16` tag or release exists, and `origin/main` still contains the PR merge commit. Stop on collision or if the merge commit is no longer the intended release target.

- [ ] **Step 2: Create and push the annotated tag on the PR merge commit**

Run:

```powershell
$rkMergedPr = gh pr view dev/stage1-acceptance --json state,mergeCommit,url | ConvertFrom-Json
$rkMergeOid = $rkMergedPr.mergeCommit.oid
git tag -a v0.1.16 $rkMergeOid -m "RailKeeper v0.1.16"
git show --no-patch --decorate v0.1.16
git push origin v0.1.16
```

Expected: local inspection shows the intended merge commit and the remote accepts the new tag.

- [ ] **Step 3: Monitor tag workflows without changing the release**

Poll these workflows for the tagged commit:

```powershell
$rkMergedPr = gh pr view dev/stage1-acceptance --json mergeCommit | ConvertFrom-Json
$rkMergeOid = $rkMergedPr.mergeCommit.oid
$rkWindowsRuns = gh run list --commit $rkMergeOid --workflow windows-portable.yml --limit 5 --json databaseId,status,conclusion,url,headBranch,createdAt | ConvertFrom-Json
$rkDockerRuns = gh run list --commit $rkMergeOid --workflow docker-image.yml --limit 5 --json databaseId,status,conclusion,url,headBranch,createdAt | ConvertFrom-Json
$rkWindowsRuns | Format-Table
$rkDockerRuns | Format-Table
$rkFailedReleaseRun = @($rkWindowsRuns) + @($rkDockerRuns) | Where-Object conclusion -eq "failure" | Select-Object -First 1
if ($rkFailedReleaseRun) { gh run view $rkFailedReleaseRun.databaseId --log-failed }
```

Expected: the tag-triggered Windows and Docker runs reach `completed` with conclusion `success`. A failure prints its failed log automatically; do not move or replace the published tag.

- [ ] **Step 4: Apply the reviewed bilingual release notes**

Run:

```powershell
gh release edit v0.1.16 --notes-file docs/releases/v0.1.16.md --title "RailKeeper v0.1.16" --latest
gh release view v0.1.16 --json tagName,name,isDraft,isPrerelease,assets,url
gh release list --limit 20 --json tagName,isLatest,isDraft,isPrerelease
```

Expected: stable, non-draft, non-prerelease release named `RailKeeper v0.1.16`; at least one Windows portable ZIP is attached.

- [ ] **Step 5: Verify published container tags**

Run:

```powershell
docker manifest inspect ghcr.io/ichwars/railkeeper:v0.1.16
docker manifest inspect ghcr.io/ichwars/railkeeper:0.1.16
docker manifest inspect ghcr.io/ichwars/railkeeper:latest
```

Expected: all three manifests exist and resolve to the digest produced by the successful Docker workflow.

- [ ] **Step 6: Smoke-test the released container in isolation**

First confirm that no container already uses the reserved name:

```powershell
docker ps -a --filter name=^/railkeeper-release-smoke-v0116$ --format "{{.Names}}"
```

Expected: empty output. Then run:

```powershell
docker run --rm -d --name railkeeper-release-smoke-v0116 -p 18084:8080 ghcr.io/ichwars/railkeeper:v0.1.16
$rkHealth = $null
1..15 | ForEach-Object {
  if ($rkHealth) { return }
  try { $rkHealth = Invoke-RestMethod http://127.0.0.1:18084/health } catch { Start-Sleep -Seconds 2 }
}
$rkHealth
$rkContainerHealth = ""
1..20 | ForEach-Object {
  if ($rkContainerHealth -eq "healthy") { return }
  $rkContainerHealth = docker inspect --format "{{.State.Health.Status}}" railkeeper-release-smoke-v0116
  if ($rkContainerHealth -ne "healthy") { Start-Sleep -Seconds 2 }
}
$rkContainerHealth
docker stop railkeeper-release-smoke-v0116
```

Expected: `/health` returns HTTP 200, container health becomes `healthy`, and `docker stop` removes the isolated `--rm` container. No persistent volume or credentials are created.

- [ ] **Step 7: Record final release state**

Run:

```powershell
gh release view v0.1.16 --json url,tagName,assets
gh release list --limit 20 --json tagName,isLatest,isDraft,isPrerelease
git status --short
git -C C:\Users\droth\Documents\GitHub\RailKeeper status --short
```

Expected: the release URL, tag, latest flag, and Windows asset are present; both tracked worktrees are clean; only the known `.superpowers/` directory may remain untracked in the feature worktree.
