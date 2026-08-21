# RailKeeper Function Symbol Library Design

Date: 2026-08-21
Status: Approved visual direction, pending written-spec review

## Goal

Replace every bundled function-symbol graphic with an independently authored RailKeeper symbol in
the approved **Werkstatt-Linie** style. Remove the previously bundled ESU-derived graphic data and
source metadata from the current repository tree and all future packages while preserving existing
vehicle data, exhibition data, backups, and ECoS interoperability.

This work covers 86 detailed function symbols and the 8 general fallback symbols introduced by the
vehicle-function migration.

## Decisions

- Use the approved visual variant A, **Werkstatt-Linie**.
- Keep existing master-data keys, labels, sort orders, and ECoS function-code mappings stable.
- Replace graphics and graphic-origin metadata only. Do not rename stored symbol keys.
- Remove the previous Base64 graphics from historical migrations in the current source tree.
- Add a forward-only migration for databases that already applied the historical migrations.
- Preserve user-created symbols whose keys are not part of the 94 bundled identities.
- Do not rewrite Git history. A history purge is destructive and requires a separate explicit task.
- Keep ECoS trademark and interoperability notices, but remove claims that RailKeeper distributes
  graphics derived from ESU/ECoS material.

## Considered Approaches

### 1. Preserve identifiers and replace the complete artwork set

This is the selected approach. It avoids rewriting vehicle and exhibition references while removing
the previous artwork from the shipped source tree and installed databases.

### 2. Rename every detailed symbol to a RailKeeper-prefixed key

This would make the identifiers visually neutral, but it would require rewriting relational values,
JSON fields, imports, old backups, tests, and ECoS lookup behavior. The identifiers are internal
compatibility keys, not distributed artwork, so the migration risk is not justified.

### 3. Move bundled symbols to a frontend-only icon registry

This would reduce database image data, but it would weaken the existing data-driven master-data
model and complicate uploaded symbols, settings previews, exhibition printing, and external API
clients. RailKeeper will retain SVG data in master-data metadata.

## Visual System

Every symbol uses a `64 × 64` view box and is independently drawn from the function meaning. The
design must not trace or reuse the geometry of the removed graphics.

The Werkstatt-Linie rules are:

- rounded line caps and joins;
- an optical stroke weight equivalent to 2.6 to 2.8 units at the 64-unit source size;
- a simple silhouette that remains recognizable at 19, 24, and 32 pixels;
- no words, product marks, manufacturer marks, or decorative frames inside the icon;
- a restrained primary contour and one optional RailKeeper-green semantic accent;
- no dependence on color alone to distinguish two different functions;
- consistent optical centering, margins, and baseline behavior across the full library;
- direction-specific symbols use an unmistakable arrow or side marker in addition to geometry;
- paired functions, such as front/rear or up/down, share one base geometry and a consistent modifier.

The application uses a dark symbol tile in both themes. The tile provides a stable background for
the bright Werkstatt-Linie palette and matches the approved comparison image. Focus, selection, and
disabled states remain visible outside the icon through the existing control border and state styles.

## Source Assets and Generation

The editable source of truth lives under:

`assets/function-symbols/workshop-line/`

It contains:

- one manifest with the 94 bundled keys, labels, categories, ECoS codes where applicable, source
  filenames, and sort order;
- one palette-independent SVG geometry file per bundled symbol;
- no copied graphics, embedded external assets, fonts, or raster images.

A repository script at `tools/build_function_symbols.py` validates the manifest and generates the
SQL payload. Generated SQL must not be edited by hand.

The generator must fail when:

- the manifest does not contain exactly 94 unique bundled keys;
- one of the expected 86 detailed or 8 fallback symbols is missing;
- an SVG has a view box other than `0 0 64 64`;
- an SVG contains text, scripts, event handlers, external references, raster images, or unsupported
  elements;
- a detailed symbol lacks its required ECoS compatibility code;
- output contains metadata fields associated with the removed artwork.

For each source geometry, the generator creates three self-contained SVG data URLs:

- `imageData`: dark neutral contour on transparent background for printing and export;
- `activeImageData`: bright contour with the restrained RailKeeper-green accent for application
  display on the dark tile;
- `inactiveImageData`: muted neutral contour for disabled or inactive presentation.

The three outputs share identical geometry. Only the palette changes.

## Metadata Contract

Bundled entries retain their current `type`, `key`, `label`, `active`, and `sort_order` values. Their
new metadata contains only RailKeeper-owned presentation data and factual compatibility fields:

| Field | Required value |
| --- | --- |
| `category` | Semantic category such as `Licht` |
| `description` | Neutral RailKeeper description of the function |
| `library` | `railkeeper-workshop-line` |
| `libraryVersion` | Numeric value `1` |
| `ecosCode` | Numeric ECoS compatibility code when the entry has one |
| `imageMime` | `image/svg+xml` |
| `imageData` | Self-contained Base64 SVG data URL using the print palette |
| `activeImageData` | Self-contained Base64 SVG data URL using the active application palette |
| `inactiveImageData` | Self-contained Base64 SVG data URL using the inactive palette |

`ecosCode` is omitted from general fallback entries that have no distinct ECoS code. Functional
categories are semantic, such as `Licht`, `Sound`, `Dampf`, `Kupplung`, `Fahren`, or `Sonderfunktion`.

The following previous metadata fields must not remain on bundled entries:

- `sourceDocument`;
- `variant`;
- `originalName` or `fileName` when they identify removed source material;
- graphic attribution or descriptions claiming ESU-derived artwork.

The existing `esu-fNNN-*` keys remain stable compatibility identifiers. ECoS lookup should prefer the
explicit numeric `ecosCode` metadata field and continue accepting the numeric portion of legacy keys
for imported and restored data.

## Migration Strategy

### Clean source tree and new installations

- Rewrite migration `0020_esu_function_symbols.sql` so it contains the same 86 stable identities and
  neutral compatibility metadata, but none of the previous SVG data or source-document references.
- Replace the graphic-update body of migration `0025_update_esu_function_symbols_variant2.sql` with
  a compatibility comment explaining that the former artwork payload was retired. Keep the filename
  so already-applied migration versions remain valid.
- Migration `0013_vehicle_functions.sql` may retain its original minimal fallback rows because the
  final forward migration updates all 94 entries on every new installation.
- Add `0064_replace_bundled_function_symbols.sql`, generated from the new asset manifest. It writes
  the new artwork and neutral metadata for all 94 bundled identities.

The migration runner records filenames without checksums. Editing the bodies of already-applied
migrations does not reapply them to existing databases. Migration `0064` is therefore required even
after the historical files are sanitized.

### Existing installations

Migration `0064` updates only the 94 known bundled `(type, key)` identities. It must:

- replace image and source-related metadata with the new metadata contract;
- preserve each row's ID, key, label, active flag, sort order, creation time, and references;
- leave separately created custom symbol keys untouched;
- avoid deleting or recreating rows so vehicle and exhibition references remain stable.

Administrators who replaced the image of a bundled key directly will receive the new official
artwork for that bundled key. Fully custom symbols with their own keys remain untouched. This is an
intentional bundled-library update and must be stated in the release notes.

## Rendering Behavior

The shared symbol renderer remains the single application entry point.

- Normal application rendering prefers `activeImageData`, then `imageData`, then a RailKeeper-owned
  fallback component.
- Disabled or explicitly inactive rendering may request `inactiveImageData`.
- Exhibition and report printing uses `imageData`, which is optimized for a white page.
- User-uploaded SVG and raster images continue to render through the existing isolated `<img>` path.
- The 8 current Lucide fallback icons are replaced with small RailKeeper-owned inline components so
  an unavailable master-data response never reintroduces a third-party visual style.
- The symbol tile uses a dedicated dark design token instead of the current warm paper background.
  It must remain legible in light mode, dark mode, hover, focus, selection, and disabled states.

The symbol picker, vehicle functions, exhibition entry dialog, exhibition workspace, exhibition
printing, settings preview, and vehicle reports must all use the same metadata selection rules.

## Backup and Restore

Application backups contain master-data rows. A backup created by an older RailKeeper release can
therefore contain the retired graphic payload even after the database migration has run.

Restore must protect the currently installed 94 bundled symbol rows:

1. Before destructive restore, capture the current bundled symbol rows from the migrated database.
2. Restore the backup through the existing validation and transaction path.
3. Replace the 94 bundled identities with the captured current-library rows before committing.
4. Restore user-created symbols with other keys normally.

This prevents an old backup from reactivating retired graphics while preserving user-owned symbol
entries. Restore validation must not reject an otherwise compatible old backup merely because its
bundled symbol metadata is obsolete.

## Documentation and Notices

Update the following maintained surfaces:

- `THIRD_PARTY_NOTICES.md`: retain the ECoS trademark and independent-project notice, remove the
  statement that RailKeeper includes graphics derived from ESU/ECoS symbol material;
- `docs/roadmap.md`: describe the RailKeeper Werkstatt-Linie library instead of an ESU symbol set;
- English and German master-data and decoder-function documentation where the bundled library is
  described;
- release notes for the version containing the migration, including the bundled-key customization
  behavior.

The documentation must distinguish factual ECoS interoperability from ownership of the new
RailKeeper graphics.

## Verification

Automated checks must cover:

- exactly 94 bundled symbols after a full migration;
- exactly 86 detailed entries with unique ECoS codes;
- all bundled entries use `library=railkeeper-workshop-line` and `libraryVersion=1`;
- all three SVG data URLs decode to valid, safe SVG with the expected `64 × 64` view box;
- no removed Base64 fragment, source filename, `sourceDocument`, or former variant metadata remains
  in the current repository tree;
- a database migrated through `0025` receives all new images through `0064`;
- an old backup restore retains the current bundled library and restores unrelated custom symbols;
- ECoS code lookup resolves the same semantic symbol before and after the migration;
- the eight fallback keys render RailKeeper components when metadata is missing;
- frontend tests cover active, print, missing-metadata, and custom-upload rendering precedence.

Run the normal project checks:

```powershell
cd backend
go test ./...

cd ..\frontend
npm.cmd run build
```

Visual verification must inspect representative symbols from every category at 19, 24, and 32
pixels in light and dark mode, plus the exhibition print preview on a white page. All 94 symbols must
also be reviewed on a generated contact sheet before completion.

## Acceptance Criteria

- The current repository tree contains no previously bundled ESU-derived SVG payloads.
- Fresh and upgraded installations contain the same 94 RailKeeper-owned symbols.
- Existing vehicle, exhibition, import, and ECoS references continue to work without key rewrites.
- Old backups cannot reactivate the retired bundled graphics.
- Custom symbols with separate keys survive migration and restore.
- Every bundled symbol is recognizable at the application's smallest rendered size.
- Application and print palettes are both legible and use identical icon geometry.
- Notices no longer describe the bundled symbol library as derived from ESU graphics.

## Non-Goals

- Rewriting Git history or force-pushing rewritten branches.
- Changing ECoS network behavior, supported function codes, or device-write scope.
- Renaming existing symbol keys.
- Replacing the master-data symbol upload feature.
- Introducing a new icon framework or general application-wide icon redesign.
