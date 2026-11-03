# In-app bug reports

Reviewed: 2026-11-02. This replaces the earlier implementation plan with the current contract.

RailKeeper offers an authenticated report dialog through the global bug action. A configured
instance submits directly to its fixed GitHub issue repository using operator credentials.
Unconfigured instances retain the external GitHub issue link. There is no central reporting relay.

## Reference and scope

The user requested comparison with [Bambuddy's report dialog](https://github.com/maziggy/bambuddy/blob/98ecaf5020e3a20a5de49859b8ad5a754d5ad85c/frontend/src/components/BugReportBubble.tsx)
and [support collector](https://github.com/maziggy/bambuddy/blob/98ecaf5020e3a20a5de49859b8ad5a754d5ad85c/backend/app/api/routes/support.py).
That implementation collects system, process, database and device information plus sanitized logs
and submits through its own relay. RailKeeper adapts the report/diagnostic workflow to its local
architecture. No Bambuddy source code is copied, no relay is introduced and no device probes are
triggered by opening a report.

## User control

The form contains area, description, expected behavior and reproduction steps. Screenshots and
encrypted contact are optional capabilities. All visible text is maintained in German and English.

The diagnostic checkbox defaults to enabled and remembers the user's choice in this browser.
Disabling it clears the memory-only error buffer, stops subsequent capture, omits browser diagnostics
from the multipart request and tells the server to skip fresh diagnostics. The server independently
discards any supplied client diagnostics when inclusion is false. Manual text and attachments remain
separate, explicit user input. No background upload occurs.

When enabled, a disclosure shows browser values and a server snapshot. Server runtime/memory values
are refreshed at submission, so their numerical values may differ from the preview. Capturing errors
begins only inside the authenticated shell and stops on logout. Stored preference changes propagate
between tabs; the current-session choice still applies if browser storage fails.

## Diagnostic allowlist

Server diagnostics are collected only for an explicit preview or an enabled submission:

- application version, deployment shape, Go version, OS and architecture;
- collection time, process uptime, CPU count, goroutine count, Go heap and runtime memory;
- SQLite version, schema version, database size and journal mode;
- enabled digital-provider type names, without device addresses or network calls.

Browser diagnostics contain family/major versions, viewport, language, time zone, theme, an
allowlisted top-level route, product area, role names and at most 30 sanitized error events.
Events keep timestamps, classification, method, abstract route, status and allowlisted codes.
Raw error messages/stacks, request/response bodies and console output are never retained.

Inventory records/counts, device identities/counts/status, IPs, hosts, ports, paths, usernames,
credentials, cookies, tokens, emails and database contents are excluded. Operational values use
explicit structs/allowlists rather than dumping environment variables or arbitrary settings.

## API and authorization

- `GET /api/v1/bug-reports/config` returns capability and limits, with no server diagnostics by default.
- `GET /api/v1/bug-reports/config?diagnostics=true` explicitly requests a fresh diagnostic snapshot.
- `POST /api/v1/bug-reports` accepts bounded multipart input and is protected by authentication/CSRF.
- All four current roles (Admin, Editor, Viewer and Messe) can report. No inventory rights are added.
- `includeDiagnostics=false` skips collection and ignores client diagnostics. For compatibility with
  the earlier isolated implementation, omitted inclusion defaults to true and requires valid diagnostics.
- New reports are limited to five IDs per actor/hour; submission attempts have a separate 15/hour limit.

The server creates bilingual section headings, quotes user content safely and adds fixed labels.
Request bodies, credentials and GitHub error bodies are not logged or returned to the browser.

## Idempotency and private persistence

A random report UUID survives retries in the open form. The private submission table contains only
actor, state, issue/asset identifiers, timestamps and bounded error category. Report text, diagnostics,
contact and screenshot bytes are not persisted. This table is excluded from application backups.

The service serializes each report, enforces actor ownership and reconciles the public UUID marker
before creating an issue. If reconciliation cannot be read, creation stops rather than risking a
duplicate after an uncertain previous response. Screenshot/comment stages reconcile separately.
Each reconciliation reads at most 1,000 GitHub entries, including pull requests in the issue list.
Reaching that bound without proving the list complete stops the creation stage, even when exactly
1,000 entries exist. The same rule prevents duplicate screenshot comments after an incomplete scan.
Partial results retain the existing issue link and permit retry without intentional duplication.

## Optional screenshot and contact capabilities

PNG/JPEG/static WebP screenshots require separate operator asset credentials and a public repository
that differs from the issue/source repository. Images are size/pixel bounded, decoded and re-encoded
to remove metadata. The UI requires separate public-upload consent. Without asset configuration the
text report remains available and the screenshot field is hidden. Missing screenshot capability is
also enforced by the API before an issue is created.

An optional age recipient enables contact input. The server encrypts the email before submitting it;
only ciphertext enters the issue. Without that recipient the contact field is hidden. No private
identity is shipped with the application.

## Operations and verification

Operator credentials remain outside source/images/data backups. Prefer restricted issue-write and
separate asset-write tokens. Using an existing broader credential requires an explicit operator
choice and does not expand the feature's fixed repository/route scope. Operator setup and recovery
are documented in the production runbook and both handbook languages.

Tests cover opt-out, data exclusion, bounded parsing/images, all-role access, CSRF, persistent rate
limits, actor ownership, retry/reconciliation, encryption, backup exclusion and API/frontend contracts.
Browser QA covers desktop/mobile, light/dark, keyboard focus, invalid input, diagnostic preview,
remembered opt-out, submission outcomes and overflow using isolated synthetic data. Such checks do
not establish live GitHub delivery or hardware compatibility.
