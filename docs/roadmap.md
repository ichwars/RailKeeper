# Roadmap

Reviewed against the current source on 2026-11-03. Stable baseline: v0.1.22. Implemented features, remaining
development and missing hardware acceptance are listed separately below. An open
integration issue does not mean that its entire adapter is missing.

## Completed Foundation

- Go API and static frontend runtime
- SQLite migrations and seed loading
- first-run setup and authentication
- persistent setup and login rate limiting
- CSRF and role checks
- vehicle CRUD
- master data management
- inventory number schemes and history
- article data search review workflow
- master data import/export
- QR code generation
- article-search image suggestions and primary image selection
- local image uploads with drag and drop
- automatic thumbnails for JPEG/PNG/WebP images
- image deletion reference checks
- vehicle attachments
- backup and restore
- backup compatibility preflight
- backup coverage for vehicle, upload, master-data, maintenance, CV and exhibition data
- backup compatibility for older exports without newer optional tables
- typed restore confirmation before destructive backup import
- overview dashboard
- vehicle CSV/TSV/XLSX/XLS/ODS/JSON import and export
- manual column mapping for unknown vehicle import headers
- field-level update preview for vehicle imports
- safe duplicate update mode for vehicle imports
- maintenance and condition history
- images and attachments per maintenance entry
- maintenance radar and dashboard summaries
- digital function mapping
- digital function icon picker
- RailKeeper Werkstatt-Linie function symbols with stored SVG graphics, custom uploads, and neutral
  ECoS compatibility-code metadata
- digital function JSON import/export
- structured CV values and CV files
- CV import comparison preview
- decoder profile suggestions for CV values and files
- visible CV change history
- ESU/LokProgrammer project files as decoder attachments with safe metadata preview and extraction
- safe decoder preview image detection for embedded JPG/PNG/WebP images
- LokProgrammer/decoder CSV, XML and text preview for CV values and function mappings
- decoder preview actions for CV import review and function mapping import
- responsive inventory table/card switch
- compact mobile inventory layout
- compact vehicle quick menu for rows/cards/mobile entries
- icon-supported vehicle quick menu actions
- configurable overview dashboard with hide, reorder and reset controls
- overview statistics export
- login redirect to the overview dashboard
- configurable update-check endpoint with offline-safe UI status
- default GitHub release update endpoint for stable release checks
- optional prerelease update checks for beta channels
- system printer discovery with optional configured printer list
- admin user management for local accounts and Admin, Editor, Viewer and Messe role assignment
- admin session management with targeted local session revocation
- current-user password change with automatic revocation of other sessions
- admin password resets revoke affected user sessions
- token-based password reset flow from login
- optional TOTP two-factor authentication enforced by the backend
- settings security event review for audit logs
- compact backup export storage summary in settings
- exhibition lists for fair/show operations with a dedicated Messe role
- exhibition entry dialogs with shared function symbol picker
- exhibition print view for the current list state
- audit coverage for exhibition list and entry lifecycle actions
- server-side exhibition address and duplicate-vehicle conflict detection, scoped by interface and day
- inventory filter bar with segmented icon filters and master-data dropdowns
- inventory PDF report dialog for overview and detailed vehicle reports
- read-only vehicle detail view with clickable source links and image preview
- camera-assisted EAN scanning where supported by the browser
- active, inactive, and print palettes for the RailKeeper Werkstatt-Linie with stable dark tiles
- accessory catalogue with quantity stock and individually tracked assets
- hierarchical storage locations and stock adjustments
- reservations and installations for vehicles
- accessory condition, removal and allocation history
- vehicle spare-parts research with reviewable sources and extraction from attachments
- inline PDF, text and image attachment previews
- decoder manufacturer/model library with source-backed CV and function suggestions
- unified ECoS reading and comparison of locomotive data, stored CVs and static functions
- confirmed ECoS image upload with target preview and verification
- bounded Z21 address reads for decoder addresses already known to RailKeeper
- separate Intellibox 3 transport capabilities for implemented Z21 UDP and planned LocoNet TCP
- stable Linux AMD64 and ARM64 Docker images
- lossless retirement of former layout planning, preserving vehicle installations and old backups

## Implemented, Awaiting a Stable Release

- Direct in-app bug reports for authenticated users, with operator-configured GitHub delivery.
- Reviewable system/process/SQLite and browser diagnostics, a remembered opt-out, and no background upload.
- Optional screenshot and encrypted-contact capabilities, independent of text reporting.
- Safe report IDs also on local HTTP origins, persistent submission admission and retry reconciliation.

## Current Development

1. Digital command-station adapters
   - [ECoS #132](https://github.com/ichwars/RailKeeper/issues/132): add identified real-device/firmware
     fixtures, extend compatibility evidence and harden interrupted or incomplete workflows
   - [Z21 #133](https://github.com/ichwars/RailKeeper/issues/133): validate the implemented bounded
     address-read path on real Z21/z21 hardware and resolve firmware-dependent protocol ambiguities
   - [Intellibox 3 #134](https://github.com/ichwars/RailKeeper/issues/134): validate Z21-compatible UDP
     on real hardware; implement LocoNet TCP and locomotive import only after protocol evidence exists

## Further Improvements Requiring Scope or Samples

1. Decoder imports and reference data
   - expand supported LokProgrammer export formats when real samples become available
   - proprietary ESUX CV blocks remain undecoded; investigate only if supported exports are insufficient
   - extend manufacturer/model and CV coverage using the documented gaps and primary sources
2. Settings and system integration
   - decide how far print jobs should be automated beyond the browser system dialog
   - keep storage usage, backup and restore visible without making settings feel overloaded
3. Exhibition conflict policy
   - address conflicts are already detected by the backend and shown to users
   - decide separately whether conflicts should prevent saving instead of remaining warnings
4. Ongoing Bambuddy-inspired design polish
   - continue refining dense toolbar/table layouts without boxed hover effects
   - keep icon buttons transparent by default with color-only hover feedback
   - review mobile navigation after the collapsible desktop sidebar work
   - continue checking table, card and modal readability in dark mode
   - refine maintenance placement and visual hierarchy in the inventory area
   - adapt Bambuddy-style row context menus further where they add clear value
5. Vehicle action depth
   - evaluate direct upload shortcuts in the quick menu where useful
   - keep destructive actions separated and confirmed
   - later upload polish: drag sorting for vehicle images
   - later upload polish: bulk actions for attachments

## Acceptance and Maintenance Follow-up

- Existing compatibility evidence and limitations are recorded in
  [command-station compatibility](site/reference/command-station-compatibility.md).
  In particular, the ECoS busy-control image-transfer flow has automated coverage but still lacks
  an additional real-device acceptance run.
- v0.1.22 has native AMD64/ARM64 container checks, but no new physical Raspberry Pi or Windows
  runtime acceptance. The Windows EXE metadata/COFF-header anomaly from release verification
  remains a separate builder follow-up, not a confirmed application runtime failure.
- Dependency updates are tracked as pull requests rather than permanent feature milestones.
  At this review, [PR #251](https://github.com/ichwars/RailKeeper/pull/251) is open.

## Explicitly Deferred

- Intellibox 2 USB/COM support until the Windows-local workflow and Docker limitations are clear
- CV programming and wider device writes without a separate protocol and safety decision
- public sharing by default
- cloud sync
- multi-tenant hosting

Layout/module/track planning was deliberately retired in v0.1.22. Historical worktrees and plans
for that area are retained context, not unfinished work scheduled for reintegration.
