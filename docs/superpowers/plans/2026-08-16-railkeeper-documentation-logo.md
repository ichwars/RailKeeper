# RailKeeper Documentation Logo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the documentation placeholder artwork with the existing official RailKeeper logo
assets in navigation, favicon, and both landing-page heroes.

**Architecture:** VitePress continues to serve brand assets from `frontend/public/brand` through its
configured public directory. Configuration owns navigation and favicon branding, while the paired
English and German frontmatter owns hero artwork.

**Tech Stack:** VitePress 2.0.0-alpha.19, Markdown frontmatter, TypeScript configuration, npm.

## Global Constraints

- Navigation and favicon use `/brand/railkeeper-mark.png`.
- English and German hero sections use `/brand/railkeeper-logo.png`.
- Do not change copy, navigation, colors, spacing, or page structure.
- Keep English and German metadata and relative paths paired.
- Load no external image resource.

---

### Task 1: Replace and verify documentation branding

**Files:**

- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/site/index.md`
- Modify: `docs/site/de/index.md`

**Interfaces:**

- Consumes: `frontend/public/brand/railkeeper-mark.png` and
  `frontend/public/brand/railkeeper-logo.png`.
- Produces: official RailKeeper branding at the English and German documentation entry points.

- [ ] **Step 1: Confirm the documentation baseline**

Run:

```powershell
cd docs
npm.cmd ci --cache ..\.cache\npm-docs
npm.cmd run check
```

Expected: dependency installation reports zero vulnerabilities, all documentation tests pass, and
VitePress builds successfully.

- [ ] **Step 2: Replace the configured navigation and favicon asset**

In `docs/.vitepress/config.mts`, use this favicon entry:

```ts
{
  rel: "icon",
  type: "image/png",
  href: "/RailKeeper/brand/railkeeper-mark.png",
}
```

Use this theme logo value:

```ts
logo: "/brand/railkeeper-mark.png",
```

- [ ] **Step 3: Replace both hero assets**

In both `docs/site/index.md` and `docs/site/de/index.md`, use:

```yaml
image:
  src: /brand/railkeeper-logo.png
  alt: RailKeeper
```

- [ ] **Step 4: Run the documentation contract and build**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: 19 tests pass, the document validator exits successfully, and the VitePress build
completes without a dead-link or missing-asset error.

- [ ] **Step 5: Perform browser acceptance**

Run:

```powershell
cd docs
npm.cmd run preview -- --host 127.0.0.1
```

Check `/RailKeeper/` and `/RailKeeper/de/` in light and dark mode at desktop and mobile widths.
Expected: the navigation shows the official train signet; both heroes show the full official logo;
images retain their proportions without clipping or horizontal overflow; all image requests stay on
the local preview origin.

- [ ] **Step 6: Commit the implementation**

Run:

```powershell
git add docs/.vitepress/config.mts docs/site/index.md docs/site/de/index.md
git commit -m "docs: use official RailKeeper logos"
```
