# RailKeeper User Guide Getting Started Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish a complete English and German guide for first setup, sign-in, sign-out, optional
two-factor sign-in, and password recovery in stable RailKeeper v0.1.17.6.

**Architecture:** Add one paired user-guide page at the paths already reserved by the coverage
matrix. Link the pair from both user-guide sidebars and landing pages, then change only the
`setup-auth` coverage topic from `planned` to `documented`. The pages describe verified user-facing
behavior and link to existing overview pages for administrative prerequisites that belong to later
documentation stages.

**Tech Stack:** VitePress 2, Markdown with validated frontmatter, Node.js documentation checks,
React/TypeScript and Go sources as the behavior contract.

## Global Constraints

- English remains the public root locale; German uses the `/de/` prefix.
- English and German pages use identical relative paths below `docs/site/` and `docs/site/de/`.
- User content documents stable RailKeeper `0.1.17.6`, not unpublished `main` behavior.
- Both pages use `audience: user`, `status: stable`, `reviewedVersion: 0.1.17.6`, and
  `lastReviewed: 2026-08-16`.
- The two language versions must be semantically equivalent, not literal machine translations.
- Do not document user management, two-factor setup, session administration, or SMTP setup as part
  of this chapter. Mention these only as prerequisites or related administrative work.
- Do not claim that a password-reset email is always sent. The visible confirmation is identical
  for known and unknown addresses, but stable v0.1.17.6 includes `expiresAt` only for known accounts
  in the HTTP response. Document this account-enumeration limitation explicitly.
- Do not include credentials, reset tokens, real email addresses, private paths, or productive
  data.
- Keep generated output, `docs/.vitepress/dist`, and `docs/node_modules` out of Git.

---

### Task 1: Publish the paired first-setup and sign-in guide

**Files:**

- Create: `docs/site/guide/getting-started/index.md`
- Create: `docs/site/de/guide/getting-started/index.md`
- Modify: `docs/site/guide/index.md`
- Modify: `docs/site/de/guide/index.md`
- Modify: `docs/.vitepress/config.mts`
- Modify: `docs/coverage.json`

**Interfaces:**

- Consumes: the `setup-auth` coverage topic, `SetupView`, `LoginView`, application startup routing,
  setup/auth API handlers, and stable version metadata.
- Produces: `/guide/getting-started/`, `/de/guide/getting-started/`, paired sidebar entries, and a
  `documented` coverage status for `setup-auth`.

- [ ] **Step 1: Make the coverage contract fail for the missing pages**

Change only the `setup-auth` topic in `docs/coverage.json`:

```json
{
  "id": "setup-auth",
  "audience": "user",
  "status": "documented",
  "englishPath": "guide/getting-started/index.md",
  "germanPath": "de/guide/getting-started/index.md"
}
```

- [ ] **Step 2: Verify the contract rejects the missing destinations**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: failure naming the missing English and German `setup-auth` pages.

- [ ] **Step 3: Write the English page**

Create `docs/site/guide/getting-started/index.md` with the required frontmatter and these exact
content requirements:

1. Title `First setup and sign-in` and a short scope paragraph.
2. `Before you start`:
   - RailKeeper must already be installed and reachable in a browser.
   - The setup form appears only while the database has no user.
   - There are no default credentials.
3. `Create the first administrator`:
   - no existing role is required;
   - username is required and at least 3 characters after surrounding whitespace is removed;
   - email is required and valid because password recovery uses it;
   - password and repetition are required, identical, and at least 12 characters;
   - submit with `Create admin`, then use the normal sign-in page;
   - the first account receives Admin, Editor, and Viewer roles;
   - setup is one-time and limited to five attempts per client address in ten minutes.
4. `Sign in`:
   - enter username and password;
   - when two-factor authentication is enabled, the code field appears after the first submission;
   - enter the current authenticator code and submit again;
   - a successful session lasts up to 12 hours;
   - Admin, Editor, Viewer, and Planner users start on Overview, Messe-only users on Exhibition.
5. `Sign out`:
   - use the log-out icon in the sidebar footer;
   - signing out revokes the current server-side session and returns to sign-in.
6. `Recover a forgotten password`:
   - open `Forgot password?`, enter the account email, and request the reset;
   - the visible confirmation stays identical for known and unknown addresses;
   - the v0.1.17.6 HTTP response can still reveal a known account through its optional `expiresAt`
     field, which must be identified as a security limitation;
   - configured SMTP sends a time-limited link; without SMTP, an operator can obtain the local
     recovery URL from the server log;
   - contact an administrator if no email arrives;
   - the newest link invalidates earlier open reset requests, expires after 30 minutes, and is
     single-use;
   - set and repeat a password of at least 12 characters;
   - completion revokes every existing session for that user.
7. `Troubleshooting` table covering absent setup form, rejected setup, invalid sign-in, missing
   two-factor code, missing reset email, invalid/expired reset link, and rate limiting.
8. `Security notes` covering HTTPS for network-exposed instances, unique credentials, and the
   incomplete account-enumeration protection of the v0.1.17.6 reset response.
9. `Related pages` linking to `/guide/` and `/administration/` only, so no link points to an
   unpublished page.
10. `Documented RailKeeper version` stating stable `v0.1.17.6` and review date 2026-08-16.

Use interface labels exactly as shown by the English UI: `Create admin`, `Sign in`,
`Forgot password?`, and `Request reset`.

- [ ] **Step 4: Write the semantically equivalent German page**

Create `docs/site/de/guide/getting-started/index.md` with matching frontmatter and the same facts,
order, limits, security effects, troubleshooting cases, and related destinations. Use these German
UI labels verbatim: `Admin erstellen`, `Anmelden`, `Passwort vergessen?`, and `Reset anfordern`.
Use the heading `Ersteinrichtung und Anmeldung`.

- [ ] **Step 5: Add both pages to the guide sidebars**

In `docs/.vitepress/config.mts`, preserve the existing overview item and add the paired second item:

```ts
{ text: "Getting started", link: "/guide/getting-started/" }
```

```ts
{ text: "Erste Schritte", link: "/de/guide/getting-started/" }
```

- [ ] **Step 6: Add a clear entry point to both guide landing pages**

Append a `Start here` section to `docs/site/guide/index.md` that links to
`/guide/getting-started/` and states that the chapter covers first administrator creation, sign-in,
sign-out, two-factor sign-in, and password recovery. Add the equivalent `Hier beginnen` section to
`docs/site/de/guide/index.md` linking to `/de/guide/getting-started/`. Update both landing-page
`lastReviewed` values to `2026-08-16`.

- [ ] **Step 7: Run the complete documentation check**

Run:

```powershell
cd docs
npm.cmd run check
```

Expected: 19 tests pass, coverage validation produces no errors, and VitePress builds both new
routes without dead links.

- [ ] **Step 8: Review the language pair and source fidelity**

Check the pair against this source checklist:

```text
frontend/src/features/setup/SetupView.tsx
frontend/src/features/auth/LoginView.tsx
frontend/src/app/App.tsx
backend/internal/application/setup.go
backend/internal/application/auth.go
backend/internal/api/auth_handlers.go
```

Expected: the two pages contain the same behavior; every numeric limit and security effect matches
the sources; later administrative features are not expanded here.

- [ ] **Step 9: Commit the content package**

```powershell
git add docs/site/guide docs/site/de/guide docs/.vitepress/config.mts docs/coverage.json
git commit -m "docs: add getting started user guide"
```
