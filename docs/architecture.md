# Architecture

RailKeeper is a small modular monolith. It is deployed as one process, but the code is separated by
responsibility.

## Boundaries

- `api`: HTTP transport, request validation, response mapping
- `application`: compatibility composition and remaining legacy use cases/transactions
- `authentication`: independent login and account/session/reset/2FA decisions behind typed ports
- `exhibition`: planning, lifecycle, readiness and conflict rules behind typed repository ports
- `infrastructure/exhibitiondata`: list/entry/exception SQLite storage and guarded transactions
- `configuration`: personal/digital/SMTP settings and maintenance rules behind typed storage ports
- `infrastructure/settingsdata`: setting queries/transactions and concrete SQLite maintenance/file sizes
- `valuation`: exact manual valuation rules behind a transactional storage port
- `transfer`: shared exchange contracts, compatibility encoding and preview identities
- `inventory/numbering`, `inventory/money`, `inventory/overview`: numbering and collection-value rules behind typed storage ports
- `inventory/masterdata`: master-data use cases, reconciliation and cache ownership behind semantic storage ports
- `inventory/workshop`: maintenance and spare-part rules behind vehicle lookup and storage ports
- `inventory/catalog`: immutable article revisions, selected-field adoption and legacy comparison rules behind typed transactions
- `inventory/vehicle`: shared vehicle/set/media records, input rules and compatible draft revision encodings
- `ecosmodel`: pure ECoS read projections, comparison plans and copied telemetry records
- `backup`: backup format, compatibility checks and restore orchestration behind storage/rule ports
- `decoder`: physical decoder/installation and CV identity, scope, conflict and partial-state rules
- `revision`: compatible JSON/SHA-256 comparison tokens without persistence or transport
- `domain`: vehicle inventory model and domain rules
- `infrastructure`: SQLite, migrations, seed loading and filesystem adapters

The first independent login slice, its compatibility facade and explicit transaction owners
are described in [authentication boundaries](architecture/authentication-boundaries.md).
The independent [transfer contracts](architecture/transfer-boundaries.md) keep storage adapters
from depending on application-owned transport models. Compatibility aliases preserve callers.
The [inventory boundaries](architecture/inventory-boundaries.md) preserve numbering transaction
ownership and recorded collection prices while separating those rules from SQLite.
The [ECoS boundary](architecture/ecos-boundaries.md) separates protocol decisions from the
existing connection, monitoring and explicitly confirmed transport operations.
The [backup boundary](architecture/backup-boundaries.md) separates format checks, domain validation,
the SQLite/file restore owner and legacy compatibility work within that same transaction.
The [decoder boundary](architecture/decoder-boundaries.md) separates identity and conflict decisions
from caller-transaction SQL while preserving CV histories, source bytes and original profile labels.

The [exhibition boundary](architecture/exhibition-boundaries.md) preserves planning revisions,
conflict identities, lifecycle timestamps and the existing transaction ownership.
The [configuration boundary](architecture/configuration-boundaries.md) preserves personal settings,
SMTP credential/fallback rules and native SQLite maintenance through typed storage ports.

## API Contract

`openapi/railkeeper.yaml` is the public contract. The existing request adapter in
`frontend/src/shared/api.ts` retains authentication, CSRF, errors, timeouts and runtime validation.
Transport types are derived from the contract without replacing that adapter; editable UI drafts
remain separate from transport records.

`npm ci` in `frontend` installs the pinned generator toolchain from `tools/api-types`.
Its TypeScript 5 dependency is isolated from the application's TypeScript 7 compiler.
`npm run generate:api` produces the ignored `frontend/src/generated/api.d.ts`; development,
typecheck, test and build scripts generate it automatically before use. A fresh checkout requires
no generated source checked into Git. Docker copies the contract and generator manifests into its
build stage; neither Node nor the generator is part of the production runtime.
Generator flags follow the [openapi-typescript CLI](https://openapi-ts.dev/cli), with
`default-non-nullable` disabled so schema defaults do not make optional input fields required.

Vehicle, CV and transfer transport records use generated schema aliases. `vehicleFormModel` derives
editable drafts while retaining incomplete control values until the existing save validation runs.
Stored transfer previews retain their legacy optional alias list through a small view projection;
their existing runtime parser is unchanged. JSON fields and required response fields are checked
against the actual Go records, including optional media metadata and inherited schema properties.

## Vehicle editing and persistence

`VehiclesView` renders inventory and auxiliary dialogs; `VehicleEditorHost` maps the existing
controllers to the editor tabs. `useVehicleWorkspace` connects those controllers once per mounted
vehicle workspace, retaining their ownership, lifetime and reset order. It adds no request layer or
second state store. The editor host stays mounted while its dialog opens and closes; its conditional
dialog keeps the previous mounting behavior. Collection loading, filtering input
and master-data reads live in `useVehicleCollectionController`; the existing editor controller owns form
and selection state and refreshes the selected aggregate. `useVehicleSetController` owns set dialog state,
duplication prefills and creation. Existing function, maintenance, decoder-file and document controllers
retain their request boundaries.

Image draft state lives in `useVehicleImageDraftController` and makes no API calls. The existing media
controller prepares independent confirmed file actions through `useVehicleUploads` and
`useVehicleMediaDeletion`. `vehicleImagePersistence` partitions
new remote images from the core save payload; `vehicleImportAdoption` handles local mapping/CV adoption
and return navigation. The core save acknowledges and selects the saved vehicle before ancillary imports,
so retrying a failed image import uses the saved vehicle. Each successful remote image is acknowledged
with its persisted ID before the next import, including when a later import fails.

Independent actions refresh the selected aggregate without reinitializing its open drafts. Core fields
use a three-way comparison with the previous response; concurrent edits to the same refreshed field
remain visible until the user chooses a value. Dirty functions keep their original revision, image
metadata and attachment edits are reconciled, and unrelated workshop/CV forms remain open. Only opening
another editor resets all sections. Closing is blocked during a save. Bilingual guidance distinguishes
pending drafts from actions already stored. See [vehicle save boundaries](architecture/vehicle-save-boundaries.md)
for the individual write paths and cancellation semantics. These local actions add no physical device writes.

Vehicle create/update orchestration lives in `inventory/vehicle`, including validation, number
reservation, optimistic revisions, image replacement, function conflicts and the single commit.
`vehicledata` owns concrete SQL and creates that transaction; the core use case commits or rolls it
back through a typed port. Set creation composes a narrower insertion port in its own transaction. Set validation, defaults,
member order and post-commit response rules also live in `inventory/vehicle`; `vehicledata` owns
set rows, membership and audit SQL. Canonical set edits retain individual member data. Number counters, core fields, images, functions, source mapping, history and audit stay
atomic. File localization happens before the transaction; compensation runs after commit/rollback.
Separate completion ports retain the existing response subsets and order after commit, including
separately inventoried decoder/CV enrichment. No application models are imported by these adapters.
Canonical records and compatible public aliases preserve response and stored values.

Vehicle deletion keeps membership cleanup, cascading rows and its audit in one transaction behind
its own narrow port. Exhibition expiry retains the existing two-step behavior and SQLite local-date
rules. A failed legacy step can be retried without rewriting timestamps of already-reset vehicles.
The application SQL ratchet distinguishes a directly inferred `net/url.URL.Query` from SQL;
the original URL-parameter candidate was not a persistence call. Real SQL on other or shadowed
receivers remains counted. Concrete vehicle lifecycle, set, numbering, catalog and workshop SQL is owned by focused
storage adapters. Separately inventoried decoder and seed work remains explicit.

`media` owns blob encoding, content hashes, read semantics and legacy-file migration behind
repository/file ports. `infrastructure/blobdata` owns concrete blob/reference SQL and confined
filesystem reads. Original files and stored path text remain intact; Windows separator normalization
applies only to reading. Immutable decoder sources can insert blobs into their caller-owned
transaction. Background cleanup retains every existing media, article and decoder-snapshot reference.

Historical snapshot reading/decoding lives in `decoder/workflow` behind a read-only transaction
port; `vehicledata` owns its actual queries. These reads require no current vehicle/workspace.
Recorded labels, missing versus zero index pages, source bytes and unknown hardware remain exact.
Configured submodules receive their own dependency checks; every independent domain module also
rejects concrete database, filesystem and network imports.

Snapshot capture and legacy adoption also live in `decoder/workflow`. Typed storage/file ports
retain the original transaction while freezing original bytes and recording partial state. Restore
composes the same rules without a nested commit. Archived variants are compared through a streaming
cursor; conflicting or unreadable originals stay unresolved. Foreign legacy references retain their
original owner metadata and never become silently adopted local content. Profile/NULL asymmetries,
original dates and history ordering are preserved.

Local decoder-instance, installation/date and working-group use cases also use
`decoder/workflow` transaction ports. Their revision check, mutation, optional checkpoint,
audit and response readback remain atomic. File roots are configured only by storage composition;
no physical command-station operation is introduced. CV working-group touches share the existing
CV transaction and preserve recorded profile text and origin.

CV writes similarly use caller-owned transactions through `vehicledata`. The additive
[atomic CV changes API](architecture/cv-drafts.md) supports bounded manual batches, exact profile/index
identity checks and computed row revisions. Existing single-value and import routes share the writer;
their existing requests remain valid. The database schema and backup format are unchanged.

## Frontend settings

Import previews use shared [source, target, effect and comparison vocabulary](architecture/import-review.md)
while keeping their specialized parsers and confirmed write boundaries.

The active transfer workspace uses `dataTransferModel`, the backend-owned file parser/preview,
`transferReviewState` and `useTransferImportConfirmation`. ECoS adoption belongs to the digital-center
workspace and vehicle adoption controllers. The former `importExportHelpers.tsx` had no production
importers and was reachable only from its own test. That obsolete parallel implementation and its
self-tests were removed together. No active parser, route, import format or CV suggestion path was
replaced. Existing transfer API, review, ECoS and vehicle tests remain the behavior checks.

`SettingsView` composes the existing settings areas and coordinates shared confirmations. Its controllers
stay mounted across tabs so navigation does not discard pending form state. Existing load points, save
timing and server-side authorization remain unchanged.

| Controller | Responsibility |
| --- | --- |
| `useSettingsNavigation` | Tab, data-group and type selection, including existing URL normalization |
| `useSettingsPreferences` | Profile preferences, local fallback, themes, sidebar and printer selection |
| `useSettingsGeneral` | Version checks and inventory-number scheme drafts and persistence |
| `useSettingsAuth` | Current user, accounts, password changes, sessions and audit reloads |
| `useSettingsSMTP`, `useSettingsTwoFactor` | Existing mail and personal 2FA request/validation flows |
| `useSettingsMasterData` | Shared master-data cache, edit drafts, filtering, selection and writes |
| `useSettingsStorage`, `useSettingsTransfer` | Storage, backup preflight/restore and master-data transfer |

The existing settings cards and tabs render these controllers. Master-data forms and table presentation
are separate components; the table retains its existing columns, capabilities and sort rules. Successful
imports invalidate the shared master-data cache; restores also refresh storage usage. Decoder, article
and digital-center settings retain their existing feature components.

## Runtime

The [regression baseline](testing-baseline.md) documents the feature matrix, synthetic reference collection,
content comparisons, app/installation restore exercises and process/browser checks used for incremental changes.
Its reports distinguish source coverage, actual runtime checks and remaining hardware/platform acceptance.

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
  app. Browser checks, crop/preview and a separate download preserve originals. Admins can additionally
  preview and explicitly confirm an ECoS image upload and locomotive assignment. See
  [image export](architecture/locomotive-image-export.md) and
  [ECoS image transfer](architecture/ecos-image-transfer.md) for evidence and limits.

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
- ECoS function keys can be read into a protected vehicle-function draft with decoder-library
  comparison and explicit selection; see [ECoS functions](architecture/ecos-functions.md).
- ECoS activation requires a successful end marker and complete version fields. Firmware 4.3.3 is the
  currently recorded real-device acceptance; other complete versions remain explicitly unverified. See
  `docs/architecture/ecos-compatibility.md` for the status model and evidence boundary.
- Z21 locomotive reads are targeted status polls for decoder addresses already known to RailKeeper, not a
  roster import. Only the confirmed address enters the comparison workspace; driving and function state is
  discarded. See `docs/architecture/z21-read-scope.md` for protocol evidence, bounds, and hardware gaps.
- The additive [decoder installation and CV snapshot model](architecture/decoder-configurations.md)
  preserves physical decoder identity separately from working profile labels and recorded partial
  states. Backup support precedes use; existing mutable CV rows remain authoritative.
- Decoder CV identity is vehicle + exact trimmed decoder profile + CV number + optional CV31/CV32 pair.
  Unindexed values and zero index pages remain distinct. Reviewed CV selections commit atomically.
  ECoS CVs use a bounded database XML export with explicit index pages and additive adoption with
  provenance; see [stored ECoS CV values](architecture/ecos-stored-cvs.md) for cache and identity limits.
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
  previews and vehicle views. Matching uses unindexed CV8 plus model/family aliases or a
  documented type identifier (D&H CV261, ZIMO CV250, Lenz CV7). For ECoS source groups, an exact,
  unique saved ECoS model profile can also select an active library model without inventing CV8.
  Manual type and supplied identity conflicts block this path; source context stays separate
  from inventory fields. See [stored ECoS CVs](architecture/ecos-stored-cvs.md).
  ZIMO identifies the base type only,
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
  Profile validation applies to master-data writes/imports and backup preflight. Backup version 26
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
- External mapping normalization and ownership/rebind decisions live in `inventory/vehicle`.
  `vehicledata` executes owner-conditional writes and the existing atomic consolidation. Vehicle
  creation binds its source mapping using the same caller-owned vehicle transaction.
- `inventory/vehicle` owns function CRUD rules, legacy flag decisions and revision-checked bulk
  changes. `vehicledata` owns their SQL. Bulk changes use the caller-owned vehicle transaction
  without beginning or committing another transaction. Existing function IDs and creation times
  survive upserts; new IDs retain the deterministic vehicle-ID/function-key format.
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
- Backup version 26 includes optional catalog articles, immutable revisions, vehicle/set links and
  individual catalog overrides alongside vehicle sets, vehicles, accessories, storage, maintenance, CVs,
  exhibitions, transfer profiles, conflict exceptions, uploads, and archived legacy planning data.
  Older supported backups remain importable. Legacy layout allocations are archived and retired
  transactionally during restore; vehicle allocations remain unchanged. Restore defers foreign-key
  checks and validates them at commit.
- Consistent catalog relationships become authoritative with transactional legacy projections for
  every existing inventory writer. Unresolved proposals retain legacy authority and their original
  marker. Its field map, backup boundary and authority rules are defined in
  [vehicle catalog relationships](architecture/vehicle-catalog.md).
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

Formularelemente erben dieselbe Schrift wie der umgebende Text. Tabellen und bestehende
Beschriftungen verwenden `AppChoiceInput` für Kontrollkästchen und Optionsfelder sowie
`AppNumberControl` für Zahlen, ohne Ereignisse, Formularwerte oder Eingabereferenzen umzuwandeln.
`AppCheckbox` und `AppNumberInput` ergänzen diese Bausteine um ihre eigenen Beschriftungen.
Dateiauswahlen verwenden `AppFilePicker` oder bei unmittelbaren Uploads `AppFileButton`;
versteckte native Eingaben bleiben für die Dateiübergabe erhalten. Schalter und Themenkarten
behalten ihre vorhandene Darstellung mit sichtbarem Tastaturfokus.

Form controls inherit the surrounding typography. `AppChoiceInput` and `AppNumberControl`
preserve native events, form values and input refs inside existing labels and tables;
`AppCheckbox` and `AppNumberInput` add field labels. `AppFilePicker` handles selected files,
while `AppFileButton` retains complete native file-change events for immediate uploads.
Hidden native inputs remain transport elements. Existing switches and theme cards keep their
visual treatment and expose keyboard focus. Typography tests check token definitions and usage.

Mobile Eingaben, Auswahloptionen und Dateischaltflächen verwenden das gemeinsame
44-Pixel-Bedienziel. Kalender und Auswahlmenüs bleiben auch bei wenig Platz innerhalb des
Bildschirms; Kalender können vertikal scrollen. Die kompakte Desktop-Geometrie bleibt erhalten.

Mobile inputs, select options and file buttons share a 44-pixel touch target. Calendars and
selection menus remain within the viewport, including short screens; calendars can scroll
vertically. Desktop controls retain their compact geometry.

Die globale Meldeaktion liegt außerhalb der Sammlungsfläche: auf dem Desktop in einem
reservierten Bereich der Seitenleiste, mobil im Kopfbereich. Eingeklappte Navigation und
schmale Übersichtsaktionen behalten ihren eigenen Platz.

The global report action stays outside collection content, in reserved sidebar space on desktop
and in the mobile header. Collapsed navigation and narrow overview shortcuts retain their space.
