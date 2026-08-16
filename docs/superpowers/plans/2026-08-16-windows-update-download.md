# Windows Update Download Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement issue #82 by offering a trusted, exact Windows Standalone ZIP download after a
successful update check, without installing or changing local files.

**Architecture:** The version API receives the runtime storage/distribution capability created by
the #84 plan and selects only the exact target-version asset from the configured RailKeeper GitHub
release. A focused frontend component renders the download action or release-page fallback; the
browser performs the actual download.

**Tech Stack:** Go 1.26, React 19, TypeScript, Vitest, OpenAPI, GitHub Releases, VitePress.

## Global Constraints

- This plan starts only after every #84 safety and Windows-package acceptance check passes.
- RailKeeper downloads nothing server-side and never extracts, installs, replaces, migrates, or
  restarts itself.
- The button appears only for Windows Standalone and only for a newer version.
- The selected asset name is exactly
  `fmt.Sprintf("RailKeeper-windows-x64-v%s.zip", normalizedVersion)`.
- The download URL must use HTTPS and point to the configured `ichwars/RailKeeper` GitHub release.
- Docker, server, development, explicit portable-data mode, wrong assets, and untrusted URLs receive
  no Windows download capability.
- The release-page fallback remains available.
- German and English UI, API, docs, and tests change together.
- Do not create a new update service, installer, or background task.

---

## Planned file structure

- `backend/internal/api/version_handlers.go`: trusted asset derivation and response capability.
- `backend/internal/api/version_test.go`: exact asset, trust, mode, and fallback tests.
- `frontend/src/features/settings/WindowsUpdateDownload.tsx`: focused download/fallback panel.
- `frontend/src/features/settings/WindowsUpdateDownload.test.tsx`: label, URL, absence, and fallback.
- `frontend/src/features/settings/SettingsView.tsx`: composes the focused update component.
- `openapi/railkeeper.yaml` and `frontend/src/shared/api.ts`: aligned `windowsPackage` contract.

### Task 1: Select only the trusted Windows Standalone asset

**Files:**
- Modify: `backend/internal/api/version_handlers.go`
- Modify: `backend/internal/api/version_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/cmd/railkeeper/startup.go`
- Modify: `openapi/railkeeper.yaml`
- Modify: `frontend/src/shared/api.ts`

**Interfaces:**
- Consumes: `api.Config.WindowsStandaloneDownload`, configured update URL, release version, release page,
  and GitHub asset list.
- Produces:

```go
type windowsPackageResponse struct {
    Version string `json:"version"`
    Name    string `json:"name"`
    URL     string `json:"url"`
}

type versionInfoResponse struct {
    Version         string                  `json:"version"`
    LatestVersion   string                  `json:"latestVersion,omitempty"`
    UpdateAvailable bool                    `json:"updateAvailable"`
    SourceURL       string                  `json:"sourceUrl,omitempty"`
    ReleaseURL      string                  `json:"releaseUrl,omitempty"`
    ReleaseNotes    string                  `json:"releaseNotes,omitempty"`
    WindowsPackage *windowsPackageResponse `json:"windowsPackage,omitempty"`
    CheckedAt       string                  `json:"checkedAt"`
    Status          string                  `json:"status"`
    Message         string                  `json:"message"`
}
```

```ts
export type VersionInfo = {
  version: string;
  latestVersion?: string;
  updateAvailable: boolean;
  sourceUrl?: string;
  releaseUrl?: string;
  releaseNotes?: string;
  windowsPackage?: {
    version: string;
    name: string;
    url: string;
  };
  checkedAt: string;
  status: "local" | "not_configured" | "current" | "update_available" |
    "unavailable" | "no_release";
  message: string;
};
```

- [ ] **Step 1: Replace arbitrary first-asset expectations with failing trust tests**

Add table-driven tests with release version `v0.2.0` and these exact outcomes:

| Mode | Asset | URL | Windows package |
| --- | --- | --- | --- |
| standalone | `RailKeeper-windows-x64-v0.2.0.zip` | trusted GitHub HTTPS | present |
| standalone | first asset is source tarball, second is exact ZIP | trusted GitHub HTTPS | exact ZIP |
| standalone | `RailKeeper-windows-x64-v0.1.9.zip` | trusted GitHub HTTPS | absent |
| standalone | exact name | HTTP | absent |
| standalone | exact name | other owner/repo | absent |
| standalone | exact name | `github.com.evil.test` | absent |
| server | exact name | trusted GitHub HTTPS | absent |
| configured portable-data | exact name | trusted GitHub HTTPS | absent |

Use the concrete trusted URL
`https://github.com/ichwars/RailKeeper/releases/download/v0.2.0/RailKeeper-windows-x64-v0.2.0.zip`.
In every absent case, assert `ReleaseURL` remains available and no arbitrary asset URL is returned.

- [ ] **Step 2: Run version tests and verify RED**

Run: `cd backend; go test ./internal/api -run 'VersionInfo|WindowsPackage' -count=1`

Expected: FAIL because the response still returns the first arbitrary asset and lacks runtime gating.

- [ ] **Step 3: Implement exact name and GitHub HTTPS validation**

Add `WindowsStandaloneDownload bool` to `api.Config` and `App`. Wire it to true only when #84
resolved a bundled Windows runtime with `StorageModeWindowsStandalone`; an explicitly configured
portable-data directory keeps it false.
Normalize `v0.2.0` to `0.2.0` only for composing the expected filename, preserving prerelease text.
Require exact case-sensitive filename equality.

Parse the update API URL and require host `api.github.com` with path prefix
`/repos/ichwars/RailKeeper/releases`. Parse the asset URL and require scheme `https`, host
`github.com`, and path prefix `/ichwars/RailKeeper/releases/download/<tag>/`. Match the final path
segment to the exact expected filename. Return the nested capability only after
`UpdateAvailable=true` and `WindowsStandaloneDownload=true`.

Remove `AssetURL` and `AssetName` from the public response instead of retaining an unsafe parallel
field.

- [ ] **Step 4: Update OpenAPI and TypeScript types**

Add optional `windowsPackage` with required `version`, `name`, and URI-formatted `url`. Remove the
old asset fields if present. Mirror the exact optional nested type in `frontend/src/shared/api.ts`.

- [ ] **Step 5: Run backend and contract tests**

Run:

```powershell
cd backend
gofmt -w internal/api/version_handlers.go internal/api/version_test.go `
  internal/api/router.go cmd/railkeeper/startup.go
go test ./internal/api ./cmd/railkeeper -count=1
go test ./... -count=1
```

Expected: PASS.

- [ ] **Step 6: Commit trusted asset selection**

```powershell
git add backend/internal/api/version_handlers.go backend/internal/api/version_test.go `
  backend/internal/api/router.go backend/cmd/railkeeper/startup.go openapi/railkeeper.yaml `
  frontend/src/shared/api.ts
git commit -m "feat: expose trusted Windows update package"
```

### Task 2: Render the explicit browser download action

**Files:**
- Create: `frontend/src/features/settings/WindowsUpdateDownload.tsx`
- Create: `frontend/src/features/settings/WindowsUpdateDownload.test.tsx`
- Modify: `frontend/src/features/settings/SettingsView.tsx`
- Modify: `frontend/src/features/settings/SettingsView.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/styles/settings.css`

**Interfaces:**
- Consumes: `VersionInfo` and translation function.
- Produces:

```tsx
type WindowsUpdateDownloadProps = {
  info: VersionInfo;
};

export function WindowsUpdateDownload({ info }: WindowsUpdateDownloadProps): JSX.Element;
```

- [ ] **Step 1: Write failing component tests**

For a package at version `v0.2.0`, assert a link with accessible name
`Version v0.2.0 herunterladen` and exact trusted `href`. Assert its explanatory copy says the ZIP is
downloaded only and is not installed. Assert there is no confirmation dialog after clicking.

For an available update without `windowsPackage`, assert no download button and a visible
release-page fallback. For Docker/server-shaped responses, assert the Windows wording is absent.
Cover English `Download version v0.2.0`, long prerelease versions, offline message, and missing
release page.

- [ ] **Step 2: Run focused frontend tests and verify RED**

Run:

```powershell
cd frontend
npm.cmd test -- --run WindowsUpdateDownload SettingsView
```

Expected: FAIL because the component and nested capability rendering do not exist.

- [ ] **Step 3: Implement the focused download component**

Render a normal `<a>` styled with the existing primary-button tokens. Use `href` directly from the
validated backend response, `rel="noreferrer"`, and no JavaScript fetch, filesystem call, or status
mutation. Do not set “installed”, “completed”, or equivalent state. Keep the existing release link
outside or immediately below the component as fallback.

Add exact translations:

```text
de: Version {{version}} herunterladen
de help: Das ZIP wird nur heruntergeladen. RailKeeper installiert oder ersetzt keine Dateien.
en: Download version {{version}}
en help: This only downloads the ZIP. RailKeeper does not install or replace files.
```

Use compact settings layout tokens and allow the button label to wrap on narrow screens without
overflow.

- [ ] **Step 4: Run focused, complete frontend tests, and build**

Run:

```powershell
cd frontend
npm.cmd test -- --run WindowsUpdateDownload SettingsView
npm.cmd test -- --run
npm.cmd run build
```

Expected: PASS.

- [ ] **Step 5: Commit the download UI**

```powershell
git add frontend/src/features/settings/WindowsUpdateDownload.tsx `
  frontend/src/features/settings/WindowsUpdateDownload.test.tsx `
  frontend/src/features/settings/SettingsView.tsx `
  frontend/src/features/settings/SettingsView.test.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/styles/settings.css
git commit -m "feat: download Windows standalone updates"
```

### Task 3: Document the safe download boundary

**Files:**
- Modify: `README.md`
- Modify: `deploy/windows/README-Windows.txt`
- Modify: `docs/coverage.json`
- Modify: `docs/site/administration/index.md`
- Modify: `docs/site/de/administration/index.md`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`

**Interfaces:**
- Consumes: completed #84 behavior and Tasks 1-2.
- Produces: bilingual download and update instructions with no installer implication.

- [ ] **Step 1: Update German and English user/operator documentation**

Document that the button starts the matching GitHub ZIP download, the version appears in the button,
no files are extracted or replaced, the program must be stopped before manual replacement, the safe
default data path remains untouched, the release page is the fallback, and Docker continues with
`docker compose pull` plus `docker compose up -d`.

- [ ] **Step 2: Run documentation terminology and contract checks**

Run:

```powershell
rg -n "Update abgeschlossen|automatisch installiert|Portable" README.md deploy/windows docs/site
cd docs
npm.cmd run check
```

Expected: no installer claims; “Portable” appears only in explicit legacy/portable-data warnings;
documentation tests and VitePress build pass.

- [ ] **Step 3: Commit download documentation**

```powershell
git add README.md deploy/windows/README-Windows.txt docs/coverage.json `
  docs/site/administration/index.md docs/site/de/administration/index.md `
  docs/site/guide/index.md docs/site/de/guide/index.md
git commit -m "docs: explain Windows update downloads"
```

### Task 4: Verify #82 without touching real user data

**Files:**
- None. This is a verification-only task. If a defect appears, stop this task, add its regression
  test to `backend/internal/api/version_test.go`,
  `frontend/src/features/settings/WindowsUpdateDownload.test.tsx`, or
  `frontend/src/features/settings/SettingsView.test.tsx`, fix it in the corresponding production
  file, commit it separately, and restart Task 4 from Step 1.

**Interfaces:**
- Consumes: the exact Windows package produced by #84 and the version API/UI from this plan.
- Produces: local acceptance evidence and a clean implementation branch.

- [ ] **Step 1: Run the complete automated matrix**

Run:

```powershell
cd backend
go test ./... -count=1
cd ..\frontend
npm.cmd test -- --run
npm.cmd run build
cd ..\docs
npm.cmd run check
cd ..
pwsh -NoProfile -File tools/test_windows_package_validation.ps1
pwsh -NoProfile -File tools/build_windows_standalone.ps1
git diff --check origin/main...HEAD
```

Expected: every test/build succeeds and the ZIP validator reports no user-data entries.

- [ ] **Step 2: Exercise update states in an isolated browser instance**

Use an isolated temporary data directory and a local fake release server. Verify exact Windows ZIP,
missing ZIP, source archive before ZIP, wrong version, HTTP asset, foreign repository, offline state,
current version, prerelease version, and server/Docker mode. Confirm only the trusted standalone case
has a button and that its click requests the exact asset URL without changing the temporary DB or
files.

- [ ] **Step 3: Check visual states**

Verify desktop and mobile widths, dark and light themes, long German prerelease labels, loading,
offline, release-only fallback, and available-download states. The update card must remain an
operational settings panel and must not grow into a marketing banner.

- [ ] **Step 4: Verify clean status**

Run: `git status -sb`

Expected: branch is clean. Do not publish or merge until the user reviews the local acceptance
result.
