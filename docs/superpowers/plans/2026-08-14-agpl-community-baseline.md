# AGPL and Community Baseline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the former MIT Self-Hosting license with AGPL-3.0-only and add the approved funding, ownership, contribution, support, conduct, trademark, and third-party notices.

**Architecture:** Keep the official AGPL text isolated in `LICENSE.md` and place project-specific legal boundaries in separate notice files. Repository metadata, Docker metadata, funding configuration, and the German and English READMEs reference the same SPDX identifier and voluntary-support model.

**Tech Stack:** Markdown, YAML, JSON, Dockerfile OCI labels, GitHub community-health files

## Global Constraints

- Use the unmodified official GNU Affero General Public License Version 3 text.
- Publish new RailKeeper versions as `AGPL-3.0-only`; do not describe the change as retroactive.
- State accurately that AGPL permits commercial use.
- Do not add memberships, paid support, SLA, sponsor benefits, a CLA, or a GitHub ruleset.
- Funding contains only `github: ichwars` and `ko_fi: ichwars`.
- Do not claim that RailKeeper or ECoS is a registered trademark unless independently verified.
- Keep `SECURITY.md` as the vulnerability-reporting authority.

---

### Task 1: Official License and Package Metadata

**Files:**
- Modify: `LICENSE.md`
- Modify: `frontend/package.json`
- Modify: `frontend/package-lock.json`
- Modify: `Dockerfile`

**Interfaces:**
- Consumes: official AGPLv3 text from `https://www.gnu.org/licenses/agpl-3.0.txt`
- Produces: repository, npm, lockfile, and OCI metadata consistently identifying `AGPL-3.0-only`

- [ ] **Step 1: Replace the license text with the official AGPLv3 text**

Use the complete response body from `https://www.gnu.org/licenses/agpl-3.0.txt` as `LICENSE.md` without project-specific additions.

- [ ] **Step 2: Verify that the local license matches the official source**

Run:

```powershell
$official = (Invoke-WebRequest -UseBasicParsing https://www.gnu.org/licenses/agpl-3.0.txt).Content
$local = Get-Content -Raw LICENSE.md
if ($local.Replace("`r`n", "`n").TrimEnd() -ne $official.Replace("`r`n", "`n").TrimEnd()) { throw "AGPL text differs" }
```

Expected: command exits successfully without output.

- [ ] **Step 3: Add package identity and license metadata**

Set the beginning of `frontend/package.json` to:

```json
{
  "name": "railkeeper-frontend",
  "private": true,
  "license": "AGPL-3.0-only",
  "scripts": {
```

Add the same `name`, `private`, and `license` properties to the root package entry `packages[""]` in `frontend/package-lock.json`.

- [ ] **Step 4: Add runtime OCI labels**

Immediately after `FROM alpine:3.24 AS runtime`, add:

```dockerfile
LABEL org.opencontainers.image.source="https://github.com/ichwars/RailKeeper" \
  org.opencontainers.image.licenses="AGPL-3.0-only"
```

- [ ] **Step 5: Validate metadata**

Run:

```powershell
Set-Location frontend
npm.cmd install --package-lock-only --ignore-scripts
node -e "const p=require('./package.json'); const l=require('./package-lock.json').packages['']; if(p.license!=='AGPL-3.0-only'||l.license!=='AGPL-3.0-only'||!p.private||!l.private) process.exit(1)"
Set-Location ..
rg -n "org.opencontainers.image.(source|licenses)" Dockerfile
```

Expected: Node exits successfully and both OCI labels are printed.

- [ ] **Step 6: Commit the license and metadata change**

```powershell
git add LICENSE.md frontend/package.json frontend/package-lock.json Dockerfile
git commit -m "legal: adopt AGPL-3.0-only"
```

### Task 2: Funding and Rights Notices

**Files:**
- Modify: `.github/FUNDING.yml`
- Create: `THIRD_PARTY_NOTICES.md`
- Create: `TRADEMARKS.md`
- Create: `SUPPORT.md`

**Interfaces:**
- Consumes: `AGPL-3.0-only` licensing decision and the ECoS scope specification
- Produces: exact funding configuration and clear separation of software, project identity, and third-party rights

- [ ] **Step 1: Replace the funding configuration**

Use exactly:

```yaml
github: ichwars
ko_fi: ichwars
```

- [ ] **Step 2: Create the third-party notice**

`THIRD_PARTY_NOTICES.md` must state that third-party components and content retain their own rights, that the RailKeeper license grants no rights to third-party marks, graphics, documentation, or protocol rights, and include exactly:

```text
ECoS is a trademark of ESU electronic solutions ulm GmbH & Co. KG. RailKeeper is an independent project and is not affiliated with or endorsed by ESU.
```

Document that the locally included ECoS/ESU function-symbol references remain subject to ESU's rights and are being presented to ESU for review.

- [ ] **Step 3: Create the RailKeeper identity notice**

`TRADEMARKS.md` must explain that AGPL-3.0-only licenses the software but does not grant rights to the RailKeeper name or logo, and that modified versions must not imply official status or endorsement.

- [ ] **Step 4: Create support boundaries**

`SUPPORT.md` must direct bugs to GitHub Issues, usage and feature discussion to GitHub Discussions, security reports to `SECURITY.md`, and state that voluntary tips create no entitlement to software, features, support, response times, or special access.

- [ ] **Step 5: Verify funding and notice wording**

Run:

```powershell
$funding = Get-Content -Raw .github/FUNDING.yml
if ($funding -notmatch "github: ichwars" -or $funding -notmatch "ko_fi: ichwars") { throw "Funding entries missing" }
if ($funding -match "paypal|buy_me_a_coffee") { throw "Legacy funding remains" }
rg -n "AGPL-3.0-only|ECoS|independent project|SECURITY.md|no entitlement" THIRD_PARTY_NOTICES.md TRADEMARKS.md SUPPORT.md
```

Expected: both funding entries exist, no legacy funding entry exists, and each required notice is found.

- [ ] **Step 6: Commit the funding and rights notices**

```powershell
git add .github/FUNDING.yml THIRD_PARTY_NOTICES.md TRADEMARKS.md SUPPORT.md
git commit -m "docs: add funding and rights notices"
```

### Task 3: Repository Ownership and Contribution Rules

**Files:**
- Create: `.github/CODEOWNERS`
- Create: `.github/PULL_REQUEST_TEMPLATE.md`
- Create: `CONTRIBUTING.md`
- Create: `CODE_OF_CONDUCT.md`

**Interfaces:**
- Consumes: RailKeeper paths and AGPL-3.0-only contribution policy
- Produces: repository ownership routing and review-ready contribution expectations

- [ ] **Step 1: Create CODEOWNERS**

Use the following RailKeeper-specific ownership map:

```text
* @ichwars
/backend/ @ichwars
/backend/migrations/ @ichwars
/backend/seeds/ @ichwars
/frontend/ @ichwars
/frontend/package-lock.json @ichwars
/openapi/ @ichwars
/deploy/ @ichwars
/Dockerfile @ichwars
/docker-compose.yml @ichwars
/.github/ @ichwars
/LICENSE.md @ichwars
/SECURITY.md @ichwars
/THIRD_PARTY_NOTICES.md @ichwars
/TRADEMARKS.md @ichwars
/CONTRIBUTING.md @ichwars
/CODE_OF_CONDUCT.md @ichwars
/SUPPORT.md @ichwars
```

- [ ] **Step 2: Create contribution guidance**

`CONTRIBUTING.md` must include environment prerequisites, the existing Go and frontend checks, focused-change expectations, no generated/runtime data, a security-reporting link, and this inbound-license statement:

```text
By submitting a contribution, you agree to license it under AGPL-3.0-only and confirm that you have the necessary rights to do so. No copyright assignment or CLA is required.
```

- [ ] **Step 3: Create the pull request template**

Include sections for summary, validation, documentation, security/data impact, screenshots for UI work, and these checkboxes:

```markdown
- [ ] I ran the relevant Go tests and/or frontend tests and build.
- [ ] I updated documentation and the API contract where required.
- [ ] I did not include secrets, runtime data, generated builds, or private backups.
- [ ] I license my contribution under AGPL-3.0-only and have the rights to submit it.
```

- [ ] **Step 4: Create a concise code of conduct**

`CODE_OF_CONDUCT.md` must require respectful technical collaboration, prohibit harassment and disclosure of private information, identify `ichwars` through GitHub private contact as the enforcement contact, and allow proportionate moderation.

- [ ] **Step 5: Validate ownership and required contribution statements**

Run:

```powershell
rg -n "^\* @ichwars$|^/\.github/ @ichwars$|AGPL-3.0-only|No copyright assignment|SECURITY.md" .github/CODEOWNERS .github/PULL_REQUEST_TEMPLATE.md CONTRIBUTING.md CODE_OF_CONDUCT.md
```

Expected: the default and self-ownership rules plus contribution and security references are printed.

- [ ] **Step 6: Commit the community files**

```powershell
git add .github/CODEOWNERS .github/PULL_REQUEST_TEMPLATE.md CONTRIBUTING.md CODE_OF_CONDUCT.md
git commit -m "docs: add repository community baseline"
```

### Task 4: Public Documentation and Final Legal Verification

**Files:**
- Modify: `README.md`
- Modify: `README.de.md`
- Modify: `CHANGELOG.md`

**Interfaces:**
- Consumes: license, funding, support, trademark, and third-party notice files from Tasks 1 to 3
- Produces: aligned public English and German explanations

- [ ] **Step 1: Update license badges and license sections**

Use `AGPL--3.0--only` in shields.io badge URLs and `AGPL-3.0-only` in prose. Both READMEs must explain that changed network-hosted versions must offer corresponding source to their users, that the change is not retroactive, and that AGPL permits commercial use.

- [ ] **Step 2: Replace support promotion**

Remove PayPal and Buy Me a Coffee references. Link to `https://github.com/sponsors/ichwars` and `https://ko-fi.com/ichwars`, describe support as voluntary tips without benefits, and link `SUPPORT.md` for support channels.

- [ ] **Step 3: Update ECoS headline claims and legal links**

Describe ECoS as a reviewed exchange of locomotive master data, CV values, and static function definitions. State that RailKeeper does not monitor speed, direction, active function states, or layout object managers. Link `THIRD_PARTY_NOTICES.md` and `TRADEMARKS.md` from the license/legal section.

- [ ] **Step 4: Add the changelog entry**

Under the current unreleased or newest version section, record the AGPL-3.0-only change, the reason for stronger network copyleft, the GitHub/Ko-fi funding switch, and the new community/legal files.

- [ ] **Step 5: Run consistency checks and frontend build**

Run:

```powershell
rg -n -i "MIT Self-Hosting|buymeacoffee|buy me a coffee|paypal\.me" README.md README.de.md .github LICENSE.md CONTRIBUTING.md SUPPORT.md TRADEMARKS.md THIRD_PARTY_NOTICES.md
rg -n "AGPL-3.0-only|github.com/sponsors/ichwars|ko-fi.com/ichwars" README.md README.de.md CHANGELOG.md frontend/package.json frontend/package-lock.json
Set-Location frontend
npm.cmd run build
Set-Location ..
git diff --check
```

Expected: the first search has no matches, the second finds the aligned new references, the frontend build succeeds, and `git diff --check` reports nothing.

- [ ] **Step 6: Commit public documentation**

```powershell
git add README.md README.de.md CHANGELOG.md
git commit -m "docs: explain AGPL license change"
```
