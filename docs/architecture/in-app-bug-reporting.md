# RailKeeper In-App Bug Reporting Design

Date: 2026-08-29
Status: Approved design, pending written-spec review

## Goal

Replace the existing floating GitHub issue link with a RailKeeper-native bug-report dialog on
configured installations. An authenticated RailKeeper user can describe a problem, optionally add
an encrypted contact address and a screenshot, review the diagnostic scope, and submit the report
without leaving RailKeeper or holding a GitHub account.

RailKeeper creates a labeled issue in the configured GitHub repository. Screenshots are stored in a
separate public GitHub asset repository and referenced from the issue. Installations without the
required credentials retain the current external GitHub link as a fallback.

## Product Decisions

- Submission happens directly from RailKeeper. A GitHub login is not required for the reporter.
- GitHub access uses two separately scoped fine-grained personal access tokens.
- The issue token has `Issues: write` only for the issue repository.
- The asset token has `Contents: write` only for a separate public screenshot repository.
- Neither token is shipped with RailKeeper, exposed to the browser, stored in SQLite, or included in
  backups.
- The optional screenshot is uploaded automatically and is publicly visible.
- The optional email address is encrypted with an `age` recipient before it is added to the public
  issue. Only the maintainer holding the corresponding private identity can decrypt it.
- Automatic diagnostics are deliberately data-minimal. They contain no inventory counts, vehicle
  data, connected-device statistics, addresses, credentials, request content, or local paths.
- A small local submission record provides idempotency without retaining the report description,
  email address, diagnostics, or screenshot.
- New issues receive `bug`, `source:in-app`, and one area label derived from the selected RailKeeper
  product area.
- User-authored issue content is not translated. Issue section headings are bilingual, with German
  first and English second.

## Considered Approaches

### 1. Separate fine-grained tokens per target repository

This is the selected approach. It keeps issue creation and screenshot storage fully automatic while
preventing the screenshot credential from writing to the RailKeeper source repository. It is
appropriate for RailKeeper installations operated by the repository maintainer or another operator
who can provide valid GitHub credentials.

### 2. GitHub App installation

A GitHub App would provide short-lived installation tokens and centralized revocation, but it would
require an App ID, installation ID, private key handling, and more authentication code. Separating
issue and asset permissions cleanly would still require additional operational complexity. This is
not justified for the first implementation.

### 3. Central RailKeeper report relay

A hosted relay would remove GitHub configuration from every installation, but it would introduce a
new cloud service, availability requirement, abuse surface, and privacy boundary. That conflicts
with RailKeeper's local-first, self-hosted operating model and is out of scope.

## User Experience

### Entry point

The existing floating bug button remains in its current position and visual style.

- When direct reporting is configured and healthy, the button opens the in-app dialog.
- When direct reporting is unavailable, the button retains its current external GitHub issue link.
- The button remains keyboard reachable and exposes a localized accessible name and tooltip.

The shell loads the reporting capability after session initialization. A temporary capability-check
failure does not block the rest of the application.

### Dialog fields

The dialog follows the existing RailKeeper modal, input, button, focus, and density conventions. It
contains:

1. **Area**, required. The current route selects the initial value, and the user can change it.
2. **What happened?**, required multiline description.
3. **What did you expect?**, optional multiline text.
4. **How can the problem be reproduced?**, optional multiline text.
5. **Email**, optional. Help text states that RailKeeper encrypts it before submission.
6. **Screenshot**, optional. It supports file selection, drag and drop, clipboard paste, and mobile
   image capture where the browser provides it.
7. **Diagnostic disclosure**, collapsed by default. It lists included and excluded categories and
   can show the exact values prepared for this report.
8. **Security notice**, linking to GitHub private vulnerability reporting and warning against
   submitting vulnerabilities, secrets, tokens, or private data as a normal bug.

The capability response provides the safe server diagnostic values, and the frontend adds its own
browser values. The disclosure therefore shows the exact combined payload before submission rather
than only naming diagnostic categories.

The screenshot control shows a preview, filename, encoded size, and remove action. Adding a
screenshot reveals a required confirmation that the user checked the image for private data and
understands that the image will be public.

The dialog provides `Cancel` and a single primary `Send bug report` action. It does not use the
reference application's `Start debug logging` wording because RailKeeper maintains its bounded
diagnostic buffer automatically for the current browser session.

### Title generation

The user does not enter a separate issue title. RailKeeper takes the first non-empty sentence from
the required description, collapses whitespace, removes control characters, and truncates the
summary so the complete title stays within 120 Unicode code points.

The title format is:

```text
[Bug][<area label>] <summary>
```

Example:

```text
[Bug][Fahrzeuge] Vorschaubild verschwindet nach dem Speichern
```

The area name follows the active RailKeeper language. The machine-readable area key remains in the
issue body and GitHub label.

### Progress, success, and failure

While submission is running, the primary button shows progress and cannot be invoked twice. Closing
the dialog requires explicit cancellation while a request is active.

On complete success, the dialog shows the issue number and a link to the created issue. On partial
success, it shows the issue number, explains which screenshot step failed, and offers a retry that
continues the same report rather than creating another issue. On total failure, all fields and the
screenshot preview remain available for retry.

The form uses inline validation, a summary for submission-level errors, and focus placement on the
first actionable error. Escape, backdrop click, tab order, and focus restoration follow the shared
modal behavior.

## Product Areas and Labels

The supported area keys and labels are:

| Area key | Route source | GitHub label |
| --- | --- | --- |
| `overview` | `/overview` | `area:overview` |
| `vehicles` | `/vehicles` | `area:vehicles` |
| `accessories` | `/accessories` | `area:accessories` |
| `layouts` | `/layouts` | `area:layouts` |
| `exhibition` | `/exhibition` | `area:exhibition` |
| `importExport` | `/import-export` | `area:import-export` |
| `digitalCenters` | `/digital-centers` | `area:digital-centers` |
| `settings` | `/settings` | `area:settings` |
| `other` | fallback | `area:other` |

Every issue also receives the configured base labels. The default and documented base labels are
`bug` and `source:in-app`. All base and area labels must already exist in the issue repository.
RailKeeper validates them during its capability check and does not create or mutate repository label
taxonomy automatically.

## Architecture

### Frontend components

The feature lives in a focused `frontend/src/features/bugReporting` module rather than growing the
application shell.

- `BugReportDialog` owns form state, validation, progress, success, and retry presentation.
- `BugReportScreenshotField` owns selection, paste, drop, preview, removal, and public-image consent.
- `bugReportDiagnostics` collects approved browser and session diagnostics.
- `clientErrorBuffer` maintains a memory-only ring buffer of sanitized frontend events.
- The shared API adapter exposes capability and multipart submission calls.
- `Shell` owns only capability loading, the floating-button mode, and dialog open state.

The dialog is lazy loaded so the reporting feature does not increase the initial application view
bundle unnecessarily.

### Backend components

- Thin handlers parse authenticated requests, enforce multipart limits, map known errors to problem
  responses, and return the submission result.
- `application.BugReportService` validates the domain input, composes the bilingual GitHub body,
  encrypts contact data, coordinates idempotency, and advances the submission state.
- A diagnostic collector reads safe runtime and database facts without querying inventory tables.
- A screenshot sanitizer decodes, bounds, resizes when required, and re-encodes image content.
- `GitHubIssueClient` creates issues, reads issues for reconciliation, and creates screenshot comments.
- `GitHubAssetClient` writes sanitized images to the configured public asset repository.
- Both clients use a bounded HTTP client, fixed `https://api.github.com` endpoints, the documented
  GitHub API version header, and redacted errors.

The POST handler uses a bounded streaming multipart reader. It does not call a parser that may spill
form fields or uploaded files into operating-system temporary storage. The complete bounded request
is held only for the duration of validation and submission.

GitHub interfaces are injected into the application service. Tests use local fakes and never require
real GitHub credentials or network access.

## API Contract

### Get reporting capability

```http
GET /api/v1/bug-reports/config
```

This endpoint requires any authenticated RailKeeper role. It returns no repository credentials,
contact key, or operational error details.

```json
{
  "enabled": true,
  "fallbackUrl": "https://github.com/ichwars/RailKeeper/issues/new",
  "maxScreenshotBytes": 5242880,
  "acceptedScreenshotTypes": ["image/png", "image/jpeg", "image/webp"],
  "includedDiagnosticCategories": [
    "railkeeper",
    "runtime",
    "browser",
    "presentation",
    "navigation",
    "roles",
    "enabledDigitalProviderTypes",
    "clientErrors"
  ],
  "excludedDiagnosticCategories": [
    "inventoryData",
    "inventoryCounts",
    "connectedDevices",
    "addresses",
    "credentials",
    "requestContent",
    "localPaths"
  ],
  "serverDiagnostics": {
    "railkeeperVersion": "0.1.20",
    "deploymentType": "container",
    "goVersion": "go1.26.0",
    "operatingSystem": "linux",
    "architecture": "amd64",
    "schemaVersion": 63,
    "enabledDigitalProviderTypes": ["ecos", "z21"]
  }
}
```

`enabled=false` is a normal response for an unconfigured installation. The reason remains available
to operators through a safe startup warning, not to ordinary clients.

### Submit or continue a report

```http
POST /api/v1/bug-reports
Content-Type: multipart/form-data
```

This endpoint requires any authenticated role and CSRF protection. The multipart fields are:

| Field | Required | Contract |
| --- | --- | --- |
| `reportId` | yes | Client-generated UUID, stable across retries |
| `area` | yes | One supported product-area key |
| `description` | yes | 20 to 10,000 Unicode code points after trimming |
| `expectedBehavior` | no | At most 5,000 Unicode code points |
| `reproductionSteps` | no | At most 10,000 Unicode code points |
| `email` | no | Valid address, at most 254 bytes |
| `clientDiagnostics` | yes | JSON matching the bounded diagnostic schema |
| `screenshotConsent` | conditional | Must be `true` when a screenshot exists |
| `screenshot` | no | PNG, JPEG, or WebP, at most 5 MiB before decoding |

A successful first submission returns `201 Created`. A completed idempotent retry returns `200 OK`.

```json
{
  "reportId": "8b1d6340-0f66-472e-9954-58f1e9d88c88",
  "issueNumber": 184,
  "issueUrl": "https://github.com/ichwars/RailKeeper/issues/184",
  "status": "complete",
  "warning": ""
}
```

`status` is `complete` or `partial`. A partial result always includes a valid issue number and URL.
Known error codes include:

- `bug_reporting_unavailable`;
- `bug_report_validation_failed`;
- `bug_report_screenshot_invalid`;
- `bug_report_screenshot_too_large`;
- `bug_report_rate_limited`;
- `bug_report_github_unavailable`;
- `bug_report_submission_conflict`.

The OpenAPI document, backend route, frontend types, and tests must remain aligned.

## Diagnostic Data Contract

### Included server values

- RailKeeper version;
- deployment form: Windows portable, container, or native process;
- Go runtime version;
- backend operating-system family and architecture;
- current SQLite migration number;
- enabled digital provider type names only, such as ECoS, Z21, Intellibox 3, or CS3.

The digital-provider diagnostic must not include enabled-device counts, connection status, hostnames,
IP addresses, ports, credentials, mappings, locomotive data, or live telemetry.

### Included browser values

- browser family and major version;
- operating-system family and major version when reliably available;
- viewport width and height;
- RailKeeper language;
- IANA time-zone name;
- effective light or dark theme;
- current pathname without query string or fragment;
- selected product area;
- current role names without username or user ID;
- up to 30 sanitized client-error events from the current page session.

The implementation uses User-Agent Client Hints when available and a bounded local parser fallback.
It does not send the raw User-Agent header as a diagnostic field.

### Client-error buffer

The buffer exists only in JavaScript memory and clears on reload. It records:

- timestamp;
- source category such as runtime, unhandled promise, or API response;
- error class or RailKeeper problem code;
- sanitized, 500-character maximum message;
- HTTP method, status, and abstract API route group when applicable.

Sanitization removes query strings, fragments, URLs, email addresses, IP addresses, UUIDs, long
numeric identifiers, Windows and POSIX paths, and control characters. API headers, cookies, tokens,
request bodies, response bodies, form values, database values, stack traces, and concrete record IDs
are never added to the buffer.

### Explicit exclusions

The report contains no:

- vehicle, accessory, layout, exhibition, maintenance, decoder, CV, upload, or attachment data;
- inventory or monetary counts;
- usernames, display names, user IDs, or session identifiers;
- connected-device counts or device status;
- serial numbers, network addresses, hostnames, ports, or local filesystem paths;
- passwords, tokens, cookies, CSRF values, API keys, webhook URLs, or SMTP values;
- request or response bodies;
- raw application logs, database files, backups, or dependency inventories.

The disclosure UI must reflect the implemented fields exactly. Adding a diagnostic field later is an
API and privacy-contract change that requires tests and documentation.

## Contact Encryption

When an email address is present, the backend validates it and immediately encrypts its UTF-8 bytes
to the configured `age` X25519 recipient. The output is ASCII armored and placed in a collapsed
GitHub issue section labeled as encrypted maintainer contact.

The plaintext address exists only in request memory. It is not stored in the submission table,
logged, included in error messages, or sent to GitHub. The private `age` identity is never configured
in RailKeeper. Operator documentation includes the exact local `age --decrypt` workflow.

Publicly reachable RailKeeper installations must use HTTPS. Contact encryption protects the value
after it reaches the trusted RailKeeper backend; it does not replace transport security.

## Screenshot Processing and Storage

The server treats screenshots as untrusted input.

1. Enforce the 5 MiB multipart limit before decoding.
2. Verify the declared media type and decoded format independently.
3. Reject malformed, truncated, animated, or unsupported input.
4. Reject decoded dimensions above 16 megapixels.
5. Normalize orientation during decoding when applicable.
6. Remove metadata by rendering decoded pixels into a new image.
7. Downscale proportionally when a dimension exceeds 4,096 pixels or the normalized output exceeds
   5 MiB.
8. Re-encode original PNG input and any decoded image with transparency as PNG. Re-encode original
   JPEG and opaque WebP input as JPEG at fixed quality 88. Never preserve the submitted file bytes.

The asset path is deterministic and contains no user data:

```text
reports/YYYY/MM/<report-id>.<normalized-extension>
```

The GitHub commit message contains only the report ID and issue number. The issue receives a separate
screenshot comment containing an immutable raw URL pinned to the returned asset commit SHA. The
comment also states that the screenshot is public and includes a hidden report marker for retry
reconciliation.

The dedicated asset repository is intentionally append-only for this workflow. Automatic retention
or deletion is out of scope because deleting an asset would break the issue evidence. Keeping assets
outside the source repository isolates repository growth and prevents screenshot commits from
polluting RailKeeper's code history.

## Issue Body

The generated issue body uses this stable order:

1. security notice;
2. bug description;
3. expected behavior;
4. reproduction steps;
5. selected product area;
6. automatically collected diagnostics;
7. sanitized client-error events;
8. encrypted maintainer contact, when present;
9. public-screenshot status, when present;
10. hidden RailKeeper report marker.

Section headings are German and English, for example `Fehlerbeschreibung / Bug description`. Empty
optional sections are omitted. Diagnostics use a fenced text or JSON block so GitHub does not render
diagnostic content as Markdown. User text is escaped or fenced where necessary to prevent it from
altering the generated section structure, mentioning users unexpectedly, or injecting hidden HTML.

The issue marker contains only the report UUID and schema version:

```html
<!-- railkeeper-bug-report:v1:8b1d6340-0f66-472e-9954-58f1e9d88c88 -->
```

## Idempotency and Submission State

A forward-only migration adds a `bug_report_submissions` table. It stores only:

- report UUID;
- actor user ID for ownership and rate-limit auditing;
- state;
- GitHub issue number and URL when available;
- normalized screenshot asset URL and commit SHA when available;
- last safe error category, never the remote response body;
- created and updated timestamps.

The table is operational metadata and is excluded from application backup export and restore.

States are `reserved`, `issue_created`, `asset_uploaded`, `complete`, and `partial`. The service
advances states monotonically and serializes work per report UUID. A retry by another user is
rejected even if that user knows the UUID.

The external API operations cannot participate in a SQLite transaction. Reconciliation therefore
uses the hidden report markers:

- before creating an issue for a previously reserved report, list the bounded set of issues created
  after the reservation timestamp and inspect their bodies for the exact issue marker;
- before creating a screenshot comment after an uncertain response, inspect the issue comments for
  the exact screenshot marker;
- use the deterministic asset path to detect an upload that succeeded before the local state write.

This prevents duplicate issues, comments, and files across client retries, timeouts, and process
restarts. A failed screenshot never causes the already created issue to be deleted.

## Rate Limiting and Authorization

Both API routes require a valid RailKeeper session for any existing application role: Admin, Editor,
Viewer, Planner, or Messe. The POST route is protected by the existing CSRF middleware.

Submission is limited to five new report IDs per actor user per rolling hour. Continuing an existing
report ID does not consume another quota unit. The limit uses the persistent rate-limit facility so a
process restart does not reset abuse protection. Rate-limit responses include a safe retry time and
do not reveal GitHub state.

Server authorization is authoritative. Hiding or disabling the frontend button is not a security
boundary.

## Configuration and Capability Checks

Direct reporting requires:

```text
RAILKEEPER_GITHUB_ISSUE_TOKEN
RAILKEEPER_GITHUB_ISSUE_REPOSITORY
RAILKEEPER_GITHUB_ASSET_TOKEN
RAILKEEPER_GITHUB_ASSET_REPOSITORY
RAILKEEPER_BUG_REPORT_CONTACT_AGE_RECIPIENT
RAILKEEPER_BUG_REPORT_LABELS
```

`RAILKEEPER_GITHUB_ISSUE_REPOSITORY` and `RAILKEEPER_GITHUB_ASSET_REPOSITORY` use the exact
`owner/repository` form and must identify different repositories on `github.com`.
`RAILKEEPER_BUG_REPORT_LABELS` is a comma-separated list and defaults to `bug,source:in-app` when
unset. Tokens are read once during process startup and held only in memory.

The capability check verifies with short timeouts that:

- both repository identifiers are syntactically valid and distinct;
- the issue repository is reachable with the issue token;
- every configured base and area label exists;
- the asset repository is reachable with the asset token and reports public visibility;
- the configured branch exists or the repository has a usable default branch;
- the contact value parses as one valid `age` X25519 recipient.

Capability results are cached briefly and refreshed after a failed submission. Startup and runtime
logs report only a safe category such as `issue token rejected` or `asset repository is not public`.
They never include a token, remote response body, contact address, report text, or screenshot URL.

This design works only on installations whose operator configures valid credentials. RailKeeper does
not embed a maintainer token in source, binaries, container images, frontend assets, or sample files.

## Failure Handling

- Invalid local input returns a field-specific validation response without contacting GitHub.
- An unavailable or rate-limited GitHub issue API leaves the report reserved and safe to retry.
- A successful issue followed by a failed asset upload returns `partial` with the issue link.
- A successful asset upload followed by an uncertain comment response triggers marker-based
  reconciliation before another comment is attempted.
- An invalid or revoked configuration disables new direct submissions and leaves the external issue
  link available.
- GitHub error bodies are bounded, classified, and discarded. They are not forwarded to the client
  or application logs.
- A browser network timeout does not cancel an external operation that GitHub already completed.
  Retrying the still-open draft continues through its report UUID. Explicit cancellation discards
  an unsent draft and its UUID after confirmation.

No automatic background retry runs after the browser session ends. The GitHub issue remains the
authoritative report once its number has been recorded.

## Documentation and Operational Setup

Update maintained English and German documentation with:

- creation of the separate public screenshot repository;
- exact fine-grained token repository selection and minimum permissions;
- creation of all required GitHub labels;
- generation and secure backup of the `age` identity;
- environment configuration for Docker and native operation;
- HTTPS requirement for publicly reachable installations;
- the exact included and excluded diagnostic fields;
- public screenshot behavior and permanent asset retention;
- local contact decryption instructions;
- troubleshooting for disabled capability, rejected credentials, missing labels, and GitHub limits.

The production runbook must state that these credentials are not part of RailKeeper backups and must
be restored separately after a server migration.

## Verification

### Backend automated checks

- capability responses expose no secrets and handle complete, incomplete, and invalid configuration;
- every authenticated role can read capability and submit, while anonymous requests are rejected;
- CSRF protection rejects an otherwise valid multipart submission without a valid token;
- all text lengths, area keys, UUIDs, email addresses, and multipart boundaries are validated;
- five new report IDs per hour succeed and the next is rate limited, while retrying an existing ID
  remains allowed;
- issue titles, bilingual bodies, hidden markers, and label mappings are deterministic;
- plaintext email never appears in persisted state, issue payloads, logs, errors, or test snapshots;
- the encrypted contact decrypts correctly only with the paired `age` identity;
- PNG, JPEG, and WebP inputs decode and normalize successfully;
- MIME mismatches, animation, malformed files, oversized input, excessive pixel counts, and output
  that cannot be bounded safely are rejected;
- normalized images contain no submitted metadata and use bounded dimensions and size;
- issue, asset, and comment failures produce the documented states and safe errors;
- retries after uncertain responses reconcile issue, file, and comment markers without duplicates;
- submission metadata is excluded from backup export and restore;
- OpenAPI contract tests cover both routes, success variants, and problem responses.

### Frontend automated checks

- the shell opens the dialog when capability is enabled and uses the external fallback otherwise;
- route-to-area detection chooses every supported default;
- required fields, screenshot consent, length limits, and email validation expose accessible errors;
- file selection, drop, paste, preview, removal, and mobile-compatible input behavior work;
- disclosure values match the actual diagnostic payload;
- the error buffer stays bounded and removes every prohibited value class;
- API calls use multipart data, include CSRF through the shared adapter, and preserve the report UUID
  across retries;
- complete, partial, validation, rate-limit, configuration, and network states preserve the correct
  draft data and focus behavior;
- German and English translations cover all visible and accessible strings.

### Manual and visual checks

Inspect the affected flow at representative desktop and mobile widths in light and dark themes.
Check keyboard-only operation, focus trapping and restoration, Escape behavior, long German text,
screen-reader labels, empty and populated disclosure, invalid fields, screenshot preview, active
submission, complete success, partial success, and offline failure. Verify that no page-level
horizontal clipping appears and that mobile controls meet the project's touch-target guidance.

Run the normal project checks:

```powershell
cd backend
go test ./...

cd ..\frontend
npm.cmd run build
```

## Acceptance Criteria

- A configured installation lets any authenticated RailKeeper role submit a bug without leaving the
  application or logging into GitHub.
- Exactly one labeled issue is created for one report UUID across retries and process restarts.
- An optional screenshot is sanitized, stored only in the separate public repository, and referenced
  from the issue without granting the asset token access to the source repository.
- An optional contact address appears on GitHub only as decryptable `age` ciphertext.
- The exact diagnostic contract is visible before submission and contains none of the excluded data.
- The application stores no report body, plaintext contact, diagnostic payload, or screenshot after
  submission processing.
- Failure and partial-success states preserve user work and provide an idempotent retry path.
- Unconfigured installations retain the existing external GitHub issue link.
- API, frontend adapter, translations, operator documentation, and automated tests remain aligned.

## Non-Goals

- A central hosted RailKeeper reporting relay.
- Shipping GitHub credentials or a contact private key with RailKeeper.
- Letting reporters choose arbitrary GitHub hosts, repositories, labels, or asset paths.
- Collecting backend logs, raw stack traces, request or response content, dependency inventories,
  inventory statistics, or connected-device statistics.
- General telemetry, analytics, crash reporting, or background upload outside an explicit report.
- Automatic security-vulnerability submission through the normal issue endpoint.
- Automatic translation or rewriting of the reporter's description.
- Automatic screenshot deletion or asset-repository retention management.
- Migrating existing manually created issues or the existing GitHub issue templates.
