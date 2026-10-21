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

- Local vehicle images can be prepared for ECoS 50200 onward, CS3, IB3 Configuration and the Z21
  app. Browser checks, crop/preview and a separate download preserve originals and never write to a
  device. See [image export](architecture/locomotive-image-export.md) for evidence and limits.

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
- Decoder definitions are local `decoder_profile` master-data entries, separate from vehicle readings.
  `internal/decoderlibrary/catalog.json` holds reviewed DCC and selected model CV rules for all eight catalogued brands.
  Coverage remains model- and document-specific, not a blanket family guarantee. `inventory.json` separately describes manufacturers, families and products, persisted as
  one validated `decoder_catalog/library` entry. Cataloguing a model never enables another model’s
  interpretation. Optional `documentationProfileId` links branded OEM products to a reference
  without enabling type recognition or transferring the retail brand CV8.
  Definitions can preserve source-array references and page/section locations, factory defaults,
  access type, units and finite allowed-value sets. Sources can carry retrieval dates and document
  SHA-256 hashes. These describe documentation, never substitute missing vehicle readings or
  authorize hardware writes. See [CV reference data](architecture/decoder-cv-reference-data.md).
  Go and TypeScript share these resources. Hierarchy, source URLs, counts and limits
  are validated on seed, master-data writes/import and backup preflight. The settings directory shows
  product coverage separately from reviewed CV rules, including inactive rules as reference.
  Product coverage is explicitly partial for each manufacturer and includes current and older
  generations. Optional sourced `variants` pair each article number with one documented connection;
  summary arrays do not imply positional pairs. `documentStatus` marks historical documentation or
  announcements without asserting availability. Branded retail entries (`brandOnly`) cannot link to
  interpretation profiles or identity rules: a registered brand CV8 does not establish the OEM CV8,
  and the retail name alone must not cause a hardware-manufacturer conflict. The directory covers
  ESU, ZIMO, D&H, Lenz, Kühn, Uhlenbrock, Roco and PIKO; CV interpretation remains separately gated.
  Startup seeds missing entries and upgrades only fingerprinted, unedited bundled metadata from
  known previous revisions. Local labels, ordering, source URLs and activation remain unchanged;
  edited metadata and custom origins are preserved. Restore uses the same seed policy.
  The authenticated `/api/v1/decoder-library` endpoint serves persisted definitions to import
  previews and vehicle views. Matching requires unindexed CV8 plus model/family aliases or a
  documented type identifier (D&H CV261, ZIMO CV250, Lenz CV7). ZIMO identifies the base type only,
  without inferring an interface or borrowing a variant's CV rules. Identified models without
  reviewed profiles use DCC foundations. Conflicts fall back to DCC basics; CV8 alone never
  selects a model. Recognition does not rename raw profile groups or modify readings. Shared
  Go/TypeScript fixtures cover the same evidence and conflict cases. Known older firmware
  suppresses manufacturer-specific rules.
  Individual definitions may require a newer `firmwareMin` than their profile. `cvStride` describes
  sparse addresses within a range; intervening addresses and rows beyond the documented mapping
  layout remain unassigned. Go lookup, TypeScript lookup, directory search and validation apply
  the same stride. ESU constant braking additionally requires its enabling CV254; CV255 alone
  cannot enable the mode. Existing profiles without an enabling-register rule retain their semantics.
  The bounded `rules.lenzStandardV2` evaluator combines CV51 brake modes, CV28/29 RailCom,
  output masks and internal smoothing only for the reviewed Lenz model. See
  [Lenz quality and source conflicts](architecture/lenz-cv-quality.md).
  `rules.kuehnN45` separately enables shifted F0–F12 output masks, SUSI gating, decimal effects,
  motor/DC braking summaries and special internal-curve presentation only for reviewed N45 models.
  It does not infer key mode/inversion or apply these rules to other Kühn decoders. See
  [Kühn quality and source conflicts](architecture/kuehn-cv-quality.md).
  Coverage counts include documented memory ranges, not full semantic decoding of every byte.
  Profile validation applies to master-data writes/imports and backup preflight. Backup version 24
  carries this data; restore fills missing bundled profiles for older backups.
- RailComPlus presentation uses one client-side evaluator (`shared/railComPlus.ts`) for both
  import previews and saved vehicle readings. It requires the reviewed active LokPilot 5 micro DCC
  profile and complete, unique indexed bytes. Known contradictory export dates are not combined.
  The bounded ASCII registration name (page 1/0, CV261–288, zero-terminated) is mutable metadata,
  never decoder identity evidence or an automatic vehicle-name update. Page 1/1 key descriptions
  produce reviewable symbol suggestions through unique numeric ECoS/ESU symbol metadata.
  Unassigned descriptions do not create functions. Explicit application fills missing symbols in
  the function draft, preserving names, existing symbols, mode and direction. The Control tab has
  one editing list and one combined action for safe missing fields, including mode/inversion for
  firmware 5.14.183. All four states of
  F0 (Light), F1 (Physical) and F3 (Logical) were compared against LokProgrammer 5.2.18 exports,
  confirming the same mode/inversion masks independently of category and symbol. F0 changes only
  CV257 on page 1/1 (0, 128, 144, 16), while CV258 stays 3. Other keys use the same descriptor
  layout; they were not individually tested with positive exports. Saved/selected modes and both
  known inversion states are protected; only unknown inversion and unselected new defaults are
  filled. A client-only `modeAssigned` marker distinguishes a selected mode from the new-row
  default and is never serialized to the API. Applied fields and CV/source/version evidence are
  appended to notes. Other firmware remains comparison-only; direction dependency is independent.
  The pinned JMRI transformation documents the byte-layout evidence, with ESU's symbol table
  taking precedence for labels. These interpretations do not add catalogue coverage claims, API
  fields or duplicate stored metadata.
- Control edits, function JSON imports and deletions are staged until the vehicle modal's shared
  save action. `functionChanges` applies distinct F0–F31 upserts/deletions in the existing vehicle
  transaction, including core fields/images/audit. Keys omitted from the request stay unchanged.
  Function `revision` hashes the complete stored record; an empty expected revision requires a new
  key, while stale or unexpectedly existing keys return 409 and roll back the entire transaction.
  Older clients omitting this optional field retain their existing behavior. No schema or backup
  version change is needed. ECoS mapping/CVs and remote media remain separate existing operations.
  Only configured/edited or positively CV-mapped keys appear under the default-on configured filter;
  generic RailComPlus descriptors alone do not establish active mapping. The technical mapping is
  collapsed at the end. Source text remains byte-for-byte in existing notes unless notes are edited;
  presentation splits only explicit generated source lines into a read-only provenance fold. Own-note
  edits preserve those source lines, including historical references, within the existing UTF-8 limit.
- Function inversion and direction dependency are independent nullable booleans. Migration 0071
  retains the former mislabelled checkbox as `legacy_direction_dependent` and clears both confirmed
  meanings to NULL. `flagsVersion: 2` identifies API/function-JSON writes with the corrected meanings;
  older clients/imports cannot overwrite confirmed flags. Backup v24 preserves nullable flags and
  the legacy reference; older backup rows are normalized before restore without guessing meanings.
- Maintainer tests can use `RAILKEEPER_CV_FIXTURE` to read a local LokProgrammer TXT export instead
  of the synthetic full-list fixture. This optional test setting is not runtime configuration;
  private exports stay outside version control.
- SQLite remains the default database because it keeps local installation, backup, and restore simple.
- Attachments are stored on the filesystem below the configured data directory; metadata stays in SQLite.
- CV provenance stores the physical source line and source export wall-clock timestamp without an inferred timezone. Missing provenance on edits/imports preserves existing metadata. Legacy records remain unknown until matched to an original source.
- Backup version 24 includes vehicle sets, vehicles, accessories, storage, maintenance, CVs,
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
Brems- und Mappingdeutungen benötigen jeweils die verifizierte Herstellerkennung und das
passende geprüfte Modellprofil. Die [76420-Auswertung](architecture/uhlenbrock-cv-quality.md)
trennt einheitliche Ausgangsmasken, Analogfreigabe und Bremsbedingungen von der
Funktionstasten-Schaltart und Invertierung.
Abgeleitete Anzeigen ersetzen keine Live-Abfrage und überschreiben keine Fahrzeug- oder
Funktionsangaben ohne ausdrückliche Übernahme im Formular.
