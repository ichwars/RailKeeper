# Architecture

RailKeeper is a small modular monolith. It is deployed as one process, but the code is separated by
responsibility.

## Boundaries

- `api`: HTTP transport, request validation, response mapping
- `application`: use cases, transactions, authorization decisions
- `domain`: vehicle inventory model and domain rules
- `infrastructure`: SQLite, migrations, seed loading and filesystem adapters

## API Contract

`openapi/railkeeper.yaml` is the public contract. The frontend currently uses a small hand-written adapter
in `frontend/src/shared/api.ts`; generated types can replace it later once the API stops changing rapidly.

## Runtime

The production runtime is a Go binary that serves:

- `/api/v1/*` for JSON APIs
- `/health` for container health checks
- static frontend files from `RAILKEEPER_STATIC_DIR`

Node.js is only used to build the frontend.

## Persistent storage and startup safety

- Windows Standalone resolves its default data directory to `%LOCALAPPDATA%\RailKeeper\data`.
- Docker keeps the fixed `/data` volume. Explicit `RAILKEEPER_DATA_DIR` configuration takes
  precedence and disables automatic legacy-directory migration.
- A legacy Windows `data` directory beside the executable is copied into a verified sibling staging
  directory. SQLite state is captured through a consistent snapshot, ordinary files are compared by
  path, type, size, and SHA-256, and the complete staging directory is promoted atomically. The
  source is never modified or removed.
- Two different databases produce a typed conflict outcome. Only a loopback safety page starts;
  migrations, seeds, file services, authentication routes, and the normal frontend remain inactive.
- Pending schema migrations on an existing database require a validated SQLite snapshot under
  `safety-backups` before any migration statement runs.
- The Admin storage API exposes only the already resolved path and capabilities. It never accepts a
  filesystem path from the browser.

## Scope Decisions

- Vehicles are the core inventory aggregate. A vehicle set owns shared catalogue and acquisition data,
  while its ordered members remain ordinary vehicles with their own inventory numbers, technical data,
  maintenance, CV values, uploads, and exhibition assignments.
- Accessories are a separate catalogue and inventory area. Quantity stock and individually tracked assets
  share storage locations, reservations, installations, condition, and history workflows.
- Accessories may be reserved and installed in vehicles. The layout workspace and track planner
  have been removed. The retired Planner role becomes Viewer without gaining write access.
  Legacy planning tables are dormant archives for backup compatibility; see
  [Retired layout planning](architecture/retired-layout-planning.md).
- Article data web search is a core module. It already uses an adapter boundary so provider-specific logic
  can be replaced later. It remains a synchronous request with one truthful in-progress state; simulated
  phases and polling are not part of the current contract. See `docs/architecture/article-search-progress.md`.
- Intellibox 3 keeps Z21-compatible UDP and LocoNet over TCP as separate transports. Only the bounded,
  read-only Z21 connection test and diagnostics are available today. See
  `docs/architecture/intellibox3-transports.md` for the validation status and hardware test boundary.
- ECoS activation requires a successful end marker and complete version fields. Firmware 4.3.3 is the
  currently recorded real-device acceptance; other complete versions remain explicitly unverified. See
  `docs/architecture/ecos-compatibility.md` for the status model and evidence boundary.
- Z21 locomotive reads are targeted status polls for decoder addresses already known to RailKeeper, not a
  roster import. Only the confirmed address enters the comparison workspace; driving and function state is
  discarded. See `docs/architecture/z21-read-scope.md` for protocol evidence, bounds, and hardware gaps.
- Decoder CV identity is vehicle + exact trimmed decoder profile + CV number + optional CV31/CV32 pair.
  Unindexed values and zero index pages remain distinct. Reviewed CV selections commit atomically.
- Maintainer tests can use `RAILKEEPER_CV_FIXTURE` to read a local LokProgrammer TXT export instead
  of the synthetic full-list fixture. This optional test setting is not runtime configuration;
  private exports stay outside version control.
- SQLite remains the default database because it keeps local installation, backup, and restore simple.
- Attachments are stored on the filesystem below the configured data directory; metadata stays in SQLite.
- CV provenance stores the physical source line and source export wall-clock timestamp without an inferred timezone. Missing provenance on edits/imports preserves existing metadata. Legacy records remain unknown until matched to an original source.
- Backup version 23 includes vehicle sets, vehicles, accessories, storage, maintenance, CVs,
  exhibitions, transfer profiles, conflict exceptions, uploads, and archived legacy planning data.
  Older supported backups remain importable. Legacy layout allocations are archived and retired
  transactionally during restore; vehicle allocations remain unchanged. Restore defers foreign-key
  checks and validates them at commit.
- Backup and restore intentionally exclude local authentication and installation data such as users,
  roles, sessions, password hashes, audit logs, SMTP settings, and user UI settings.


### Gemeinsame Typografie und Decoderanzeige

Die Oberfläche bevorzugt wie der bestätigte Entwurf Segoe UI. Für Systeme ohne diese Schrift
wird Microsoft Selawik lokal als `RailKeeper Sans` gebündelt (OFL-Lizenz unter
`frontend/public/fonts/`, Quelle: https://github.com/microsoft/Selawik). Schriftgrößen, Gewichte und Abstände bleiben über die gemeinsamen
Design-Tokens steuerbar; externe Schriftanbieter sind nicht erforderlich. Der Fahrzeugdialog
behält seine Popup-Struktur und ergänzt unter Modell den Bereich Digital. Die Auswertung in
`decoderAnalysis.ts` trennt Decoderprofile, Basis-CVs und Indexseiten. Herstellerspezifische
Brems- und Mappingdeutungen benötigen die verifizierte ESU-Kennung und ein LokPilot-5-Profil.
Abgeleitete Anzeigen ersetzen keine Live-Abfrage und überschreiben keine Fahrzeug- oder
Funktionsangaben ohne ausdrückliche Übernahme im Formular.
