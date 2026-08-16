# Vehicle Creation And Inventory UI Correction

**Status:** Approved for specification on 2026-08-17

## Goal

Correct the vehicle-set inventory and vehicle-creation experience so the implemented structure,
hierarchy, responsive behavior, and interaction states match the visual direction approved in
GitHub discussion #96 and issue #102.

This correction keeps the existing vehicle-set persistence and article-search engine. It replaces
the UI components that reduced the approved concepts to generic forms, and it adds a persistent
inventory number for every set.

## Authority And References

- GitHub discussion #96, especially:
  - inventory overview concept in discussion comment `18041755`;
  - creation-flow concepts in discussion comment `18041939`.
- GitHub issue #102, including the three-step flow, responsive direction, and initial acceptance
  criteria.
- The reference screenshots reviewed on 2026-08-17 for:
  - desktop inventory overview;
  - desktop creation step 1;
  - desktop and mobile article-data review;
  - desktop and mobile vehicle details.
- Existing RailKeeper design tokens, controls, permissions, filters, sorting, inventory selection,
  article-search ranking, and local-first behavior.

When prose and a reference screenshot differ on visual structure, the approved screenshot controls.
When a screenshot contains illustrative data that has no RailKeeper domain meaning, the existing
domain controls and this specification must not invent a misleading field.

## Confirmed Decisions

- Use a focused dialog on desktop and a full-screen flow on mobile.
- Use the targeted component-replacement approach, not a CSS-only patch or full vehicle-area
  rewrite.
- Give every vehicle set a real, persistent, automatically assigned inventory number.
- Use the default set number scheme `RK-SET` with six digits, for example `RK-SET-000001`.
- Keep individual member vehicles and their inventory numbers.
- Preserve the existing article-search providers, ranking, enrichment, extraction, selection, and
  safety behavior.
- Replace the existing separate article-search review dialog in the create flow with embedded wizard
  states.
- Require visual comparison at fixed desktop and mobile viewports before merge.

## Non-Goals

- No standalone article-management area.
- No replacement of SQLite, the local-first model, or the existing search providers.
- No redesign of vehicle detail, maintenance, decoder/CV, media, or exhibition screens outside the
  create flow and inventory rows.
- No new generic UI framework.
- No fictional lifecycle or availability state. The `Status` values shown in concepts are
  illustrative. RailKeeper continues to display real configured columns, including exhibition and
  condition data, until a separate lifecycle model is specified.
- No complete-set deletion in this correction. The destructive semantics for deleting a set and all
  members require a separate explicit decision.

## Domain Model

### Persistent set inventory number

Add `inventory_number TEXT NOT NULL` to `vehicle_sets` through migration
`0060_vehicle_set_inventory_number.sql`.

The migration must:

1. Create an active inventory-number scheme with category `Set`, prefix `RK-SET`, next number `1`,
   and padding `6` when the category does not already exist.
2. Backfill existing sets deterministically in `created_at, id` order.
3. Advance the scheme by the number of values consumed during backfill.
4. Add a unique index for `vehicle_sets.inventory_number`.
5. Prevent empty set inventory numbers on insert and update.

Set creation reserves the inventory number inside the same transaction as the set and its members.
A rollback must not consume a number. Reservation reuses the existing inventory-number service and
checks conflicts against `vehicle_sets`.

The `Set` scheme appears in settings with the existing number-scheme controls. Administrators can
change its prefix, padding, next number, and active state using the established behavior.

### Backup and restore

Raise the app-backup format to version 18. Version 18 includes the new set inventory number. Older
backups remain importable; restored sets without a number receive deterministic numbers during the
restore transaction. Authentication data remains excluded.

### Set actions

The inventory set row supports these real actions:

- **View:** open a read-only set summary with shared data and ordered members.
- **Edit:** edit shared set, article, acquisition, storage, condition, and packaging fields.
- **Duplicate:** open the create wizard prefilled from the set. Set and member inventory numbers are
  cleared and assigned only when the duplicate is created.
- **Quick menu:** expose View, Edit, and Duplicate when space is constrained.

No action may be displayed as active unless it works. Complete-set deletion is not exposed.

Editing a set updates the canonical `vehicle_sets` row and the compatibility snapshots of all
members in one transaction. Member-specific fields stay unchanged.

## API Contract

### Vehicle set

Add `inventoryNumber` to `VehicleSet` responses. It is server-assigned and is not accepted as a
client-selected value during ordinary set creation.

Add:

- `GET /api/v1/vehicle-sets/{id}` for the shared set and ordered members;
- `PATCH /api/v1/vehicle-sets/{id}` for shared fields.

Both routes use the existing role and CSRF conventions: Viewer may read, Editor and Admin may edit.

### Vehicle list projection

Each vehicle-list entry that belongs to a set receives an optional `vehicleSet` summary object with:

- `id`
- `inventoryNumber`
- `name`
- `manufacturer`
- `articleNumber`
- `gauge`
- `epoch`
- `acquisitionType`
- `purchaseDate`
- `purchasePrice`
- `condition`
- `memberCount`
- `position`

The summary is read from the canonical set row, not reconstructed from the first member. Existing
flat membership fields remain temporarily for API compatibility and are deprecated only in a later
contract version.

## Inventory Experience

### Desktop hierarchy

Replace the full-width group separator with a real set row aligned to the inventory columns.

The default desktop presentation contains:

- selection control with indeterminate state when only some visible members are selected;
- expand/collapse control;
- type badge identifying the row as a set or acquisition;
- set inventory number;
- manufacturer;
- article number;
- set designation and member-count badge;
- gauge and epoch;
- the real configured operational column, not an invented status;
- set actions.

A secondary line inside the set row shows acquisition date and purchase price when present. Missing
values are omitted rather than rendered as empty labels.

Expanded members appear below the set with visible tree connectors. Member rows retain their own:

- selection control;
- type badge and primary image;
- inventory number;
- vehicle number as compact secondary text;
- manufacturer, article number, designation, gauge, epoch, operational value, and actions.

The hierarchy must remain understandable without relying on green color alone.

### Filtering, sorting, and selection

- Filtering and searching continue to evaluate physical vehicle records.
- When at least one member matches, its set row is shown.
- Only matching members are shown while a filter is active; the member badge distinguishes visible
  matches from total members when the values differ.
- Sorting physical vehicles determines group order from the first matching member. Members retain
  their explicit set position inside a group.
- Selecting a set selects all currently visible member vehicles. The set checkbox is checked,
  unchecked, or indeterminate according to its visible members.
- Reports and exhibition actions remain vehicle-based.

### Configurable columns

Add a `type` presentation column and include it in the new default column set. Existing saved column
preferences remain valid. Set rows render meaningful values only for columns backed by canonical
set data and leave member-only values blank.

### Mobile hierarchy

Create a dedicated mobile set card instead of shrinking the desktop row. It shows the set number,
name, manufacturer/article, member count, acquisition summary, expand/collapse control, and compact
actions. Expanded member cards retain their images, inventory and vehicle numbers, and existing
actions. Tree indentation and borders communicate membership without requiring color perception.

## Creation Wizard

### Shell and progress

Create a focused wizard shell shared by single vehicles and sets.

- Desktop uses the approved vertical progress rail with step title, summary, completion mark, and
  the main work area beside it.
- Mobile uses a compact top progress indicator, step title, and a full-screen scroll area.
- Header and footer remain visible while the step body scrolls.
- Back, cancel, save-draft, and primary actions use explicit labels that describe the destination or
  result.

The wizard state is an explicit reducer/state machine rather than three unrelated conditional
blocks. It stores creation kind, current step, article-search substate, selected fields and images,
shared set draft, ordered member drafts, and active detail tab.

### Step 1: type and basic data

The desktop screen follows the approved structure:

- Single vehicle and Set cards at the top.
- Manufacturer, designation, article number, gauge, and category in the basic-data section.
- Set-only vehicle count with a minimum of two.
- Read-only preview of the next set inventory number, labelled as provisional until creation.

Changing the member count regenerates only untouched member placeholders. It never silently removes
members containing entered data; reducing the count below populated members requires explicit
confirmation.

### Step 2: article data

Step 2 contains distinct substates inside the wizard:

1. **Search input:** barcode, manufacturer/article search, or manual continuation.
2. **Results:** ranked result cards with source, summary, image preview, and explicit selection.
3. **Review:** grouped field and image review before values are applied.

Review groups are:

- Identification and master data;
- Railway and epoch;
- Technical data;
- Description;
- Images and documents.

Each group shows found count and conflict count. Desktop may expand a group into a compact table.
Mobile renders field/value cards without a horizontally scrolling table. Every field shows current
value, found value, selection, and conflict status. Users can continue without importing values.

The shared article-search model and controller remain the source of truth. The existing standalone
dialog remains available for edit/detail contexts outside creation until those contexts are
separately redesigned.

### Step 3: set and vehicle details

For sets, step 3 provides these tabs:

- `Set & acquisition`
- one tab per ordered vehicle
- `Add vehicle`

The set tab owns shared article, acquisition, storage, condition, and packaging data. Member tabs own
designation, vehicle number, technical data, equipment, coupling/power pickup, decoder/CV previews,
notes, and other physical-vehicle fields. Shared fields are visible as inherited context but are not
duplicated as editable member controls.

Desktop uses the approved tab row and grouped accordions. Mobile uses a set/member selector and the
same grouped sections in a single column. Section headers show field counts. Set and member inventory
numbers remain read-only previews until creation.

### Draft persistence

`Save draft` stores the complete create-wizard state in local storage with a schema version and
timestamp. No server record or inventory number is reserved. Opening create offers to resume or
discard the saved draft. Successfully creating the vehicle or set removes the draft. Invalid or
incompatible draft data is ignored with a useful message.

## Component Boundaries

Introduce focused components instead of growing `VehiclesView.tsx` or the existing model tab:

- `VehicleCreateWizardShell`
- `VehicleCreateStepBasics`
- `VehicleCreateStepArticle`
- `VehicleCreateArticleResults`
- `VehicleCreateArticleReview`
- `VehicleCreateStepDetails`
- `VehicleSetDetailsTabs`
- `VehicleSetInventoryRow`
- `VehicleSetInventoryMobileCard`

Keep orchestration in the vehicle feature and article-search selection logic in the shared search
model. `VehiclesView.tsx` wires controllers and dialogs only.

## Error Handling

- Number-scheme absence produces the established inventory-scheme error and no partial set.
- Inventory-number conflicts retry through the existing bounded reservation mechanism.
- A failed set update rolls back the set row and every member compatibility snapshot.
- Article-search failure stays inside step 2 and preserves entered values and selections.
- Draft-storage failure does not block creation and reports that the draft could not be saved.
- Validation moves focus to the first invalid field and names the affected step and section.

## Accessibility

- The wizard remains one labelled modal dialog on desktop and one labelled full-screen dialog on
  mobile.
- Progress uses an ordered list with current-step semantics; visual labels may not disappear without
  an accessible equivalent.
- Set and creation-kind cards use radio semantics.
- Set/member tabs follow tab-list, tab, and tab-panel semantics with keyboard navigation.
- Accordions expose expanded state and retain logical heading order.
- Set selection communicates checked, unchecked, and mixed states.
- All icon-only actions have visible tooltips and accessible names.
- Hierarchy, completion, conflicts, and validation never rely on color alone.
- Touch targets remain at least 44 by 44 CSS pixels on mobile.

## Testing And Visual Acceptance

### Automated behavior

Add focused tests for:

- deterministic migration backfill and scheme advancement;
- transaction-safe set-number reservation and rollback;
- backup version 18 and older-backup compatibility;
- set read/update authorization and OpenAPI alignment;
- canonical set summaries in filtered vehicle lists;
- group sorting, visible-member filtering, and tri-state selection;
- wizard reducer transitions and draft restore;
- search input, results, review, conflict, and manual paths;
- set/member tabs and payload construction;
- keyboard and accessible-state behavior.

Run the full backend suite, frontend suite, frontend production build, documentation validation, and
Windows package validation before release.

### Visual contract

Capture and compare at minimum:

- inventory desktop at 1365 by 768 or the exact approved-reference viewport;
- inventory mobile at 390 by 844;
- wizard step 1 desktop and mobile;
- article results desktop and mobile;
- article review desktop and mobile;
- step 3 set tab and member tab desktop and mobile;
- empty, loading, search-error, validation-error, and long-German-text states;
- light and dark themes.

For every approved reference, capture the same viewport and state. Compare source and implementation
together, correct visible structural mismatches, and repeat. A successful build or a screenshot of a
different state is not visual acceptance.

Before merge, the user receives a running local checkpoint for visual inspection. Merge and release
occur only after that checkpoint and all required checks are approved.

## Delivery Sequence

Use one correction branch and one pull request with reviewable commits:

1. Set inventory-number migration, API projection, set read/update, backup, and tests.
2. Desktop/mobile set hierarchy in the inventory.
3. Wizard shell, basics, article-search states, and review.
4. Set/member detail tabs and local draft persistence.
5. Visual comparison corrections, full verification, documentation, and release notes.

This sequence exposes the overview correction early while keeping one final integration and release
gate.
