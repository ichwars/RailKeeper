# RailKeeper Function Symbol Library Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development
> (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use
> checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace all 94 bundled function symbols with the independently authored RailKeeper
Werkstatt-Linie library while preserving stored keys, ECoS mappings, custom symbols, printing, and
safe backup restore.

**Architecture:** Palette-independent SVG source files and a standard-library Python generator
produce validated active, inactive, and print SVG data URLs plus migration SQL. The existing
master-data model remains authoritative. A small pure TypeScript helper selects the correct image
variant, and backup restore preserves the currently installed bundled library instead of
reintroducing retired artwork from old backups.

**Tech Stack:** SVG 1.1-compatible markup, Python 3 standard library, SQLite migrations, Go 1.26.6,
React 19, TypeScript 7, Vitest 4, CSS design tokens.

## Global Constraints

- Keep all existing master-data keys, labels, sort orders, and ECoS function-code mappings stable.
- The library contains exactly 86 detailed symbols and 8 general fallback symbols.
- Every editable SVG uses `viewBox="0 0 64 64"`, contains no text, script, event handler, external
  reference, font, or raster image, and is independently drawn.
- Werkstatt-Linie uses rounded joins and caps, a 2.6 to 2.8 unit optical stroke, a primary contour,
  and at most one RailKeeper-green semantic accent.
- Symbols must remain recognizable at 19, 24, and 32 pixels without relying on color alone.
- `imageData` is the white-page print variant, `activeImageData` is the bright application variant,
  and `inactiveImageData` is the muted variant. All three share identical geometry.
- Historical migrations must contain none of the retired Base64 graphics or source-document metadata.
- Migration `0064` updates only the 94 bundled identities and leaves custom keys untouched.
- Old backups must not reactivate retired graphics for bundled keys.
- Do not rewrite Git history, rename symbol keys, change ECoS device behavior, replace the upload
  feature, or introduce a new icon framework.
- Preserve unrelated user changes and the existing untracked `facebook-*` paths.

---

## File Structure

### New source and tooling

- `assets/function-symbols/workshop-line/manifest.json`: authoritative identity, label, category, ECoS
  code, sort order, and SVG filename catalog.
- `assets/function-symbols/workshop-line/*.svg`: 94 palette-independent Werkstatt-Linie geometry
  files.
- `tools/build_function_symbols.py`: validates the catalog and SVG safety, renders three palettes,
  writes migrations, and writes an SVG contact sheet.
- `tools/test_build_function_symbols.py`: standard-library tests for validation, palette rendering,
  data URLs, and deterministic SQL generation.
- `backend/migrations/0064_replace_bundled_function_symbols.sql`: generated forward migration for
  fresh and existing databases.
- `backend/internal/infrastructure/function_symbol_migration_test.go`: migration count, metadata,
  preservation, and historical-upgrade coverage.
- `frontend/src/shared/functionSymbolImages.ts`: pure variant-selection helper shared by application and print paths.
- `frontend/src/shared/functionSymbolImages.test.ts`: selection-precedence tests.
- `frontend/src/shared/functionSymbolFallbackIcons.tsx`: eight RailKeeper-owned inline fallback components.

### Existing files to modify

- `backend/migrations/0020_esu_function_symbols.sql`: retain identities and factual ECoS metadata,
  remove retired graphic payloads.
- `backend/migrations/0025_update_esu_function_symbols_variant2.sql`: replace the retired update body
  with a compatibility comment.
- `backend/internal/application/master_data_test.go`: assert the RailKeeper library contract instead
  of ESU graphic provenance.
- `backend/internal/application/backup.go`: capture and restore current bundled function-symbol rows
  during backup import.
- `backend/internal/application/backup_test.go`: prove old bundled payloads cannot return and custom
  symbols still restore.
- `frontend/src/shared/functionSymbols.tsx`: use active image selection and RailKeeper-owned fallbacks.
- `frontend/src/shared/functionSymbols.test.tsx`: cover active image precedence and fallback rendering.
- `frontend/src/features/settings/settingsModel.ts`: use the shared active-image selector for settings preview.
- `frontend/src/features/exhibition/ExhibitionView.tsx`: use the shared print-image selector.
- `frontend/src/features/importExport/importExportHelpers.test.tsx`: prove explicit `ecosCode` compatibility lookup.
- `frontend/src/styles/base.css`: add theme-independent symbol-tile tokens.
- `frontend/src/styles/vehicle-functions-cv-maintenance.css`: apply the dark tile, bright fallback
  contour, accent, and disabled state.
- `THIRD_PARTY_NOTICES.md`: remove the derived-graphics statement while keeping the ECoS trademark notice.
- `docs/roadmap.md`: describe the RailKeeper Werkstatt-Linie library.
- `docs/site/administration/master-data-general.md`: document the bundled RailKeeper library and custom uploads.
- `docs/site/de/administration/master-data-general.md`: German equivalent.
- `docs/site/guide/vehicles/decoder-cv.md`: distinguish RailKeeper symbols from factual ECoS codes.
- `docs/site/de/guide/vehicles/decoder-cv.md`: German equivalent.
- `docs/releases/v0.1.20.md`: record the replacement and bundled-key customization behavior.

---

### Task 1: Build the deterministic symbol generator

**Files:**

- Create: `tools/build_function_symbols.py`
- Create: `tools/test_build_function_symbols.py`

**Interfaces:**

- Produces: `validate_svg(svg_text: str, source: str) -> None`
- Produces: `render_svg(svg_text: str, palette: Palette) -> str`
- Produces: `encode_svg(svg_text: str) -> str`
- Produces: `load_library(root: Path) -> Library`
- Produces: `build_sql(library: Library) -> str`
- Produces CLI: `python tools/build_function_symbols.py --check`
- Produces CLI:
  `python tools/build_function_symbols.py --write-migration`
  `backend/migrations/0064_replace_bundled_function_symbols.sql`
- Produces CLI: `python tools/build_function_symbols.py --contact-sheet .cache/function-symbols/contact-sheet.svg`

- [ ] **Step 1: Write failing generator unit tests**

Create tests using temporary SVG and manifest fixtures. The safety assertions must be explicit:

```python
class FunctionSymbolGeneratorTest(unittest.TestCase):
    def test_rejects_unsafe_svg_content(self) -> None:
        unsafe = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"><script>alert(1)</script></svg>'
        with self.assertRaisesRegex(ValueError, "script"):
            validate_svg(unsafe, "unsafe.svg")

    def test_renders_geometry_identically_for_all_palettes(self) -> None:
        source = (
            '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">'
            '<path data-rk-role="primary" fill="none" stroke="#111111" d="M8 32h48"/>'
            '<path data-rk-role="accent" fill="none" stroke="#419310" d="M32 8v48"/>'
            '</svg>'
        )
        active = render_svg(source, ACTIVE_PALETTE)
        printed = render_svg(source, PRINT_PALETTE)
        self.assertIn('#f2f5f6', active)
        self.assertIn('#111111', printed)
        self.assertEqual(geometry_signature(active), geometry_signature(printed))

    def test_data_url_is_base64_svg(self) -> None:
        svg = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64"/>'
        value = encode_svg(svg)
        self.assertTrue(value.startswith("data:image/svg+xml;base64,"))
        self.assertEqual(base64.b64decode(value.split(",", 1)[1]).decode("utf-8"), svg)
```

Define the geometry comparison helper in the test file so the palette assertion ignores only
presentation attributes:

```python
def geometry_signature(svg_text: str) -> str:
    root = ET.fromstring(svg_text)
    for element in root.iter():
        element.attrib.pop("stroke", None)
        element.attrib.pop("fill", None)
        element.attrib.pop("data-rk-role", None)
    return ET.tostring(root, encoding="unicode", short_empty_elements=True)
```

- [ ] **Step 2: Run the unit tests and verify the missing module failure**

Run:

```powershell
python -m unittest tools.test_build_function_symbols -v
```

Expected: FAIL because `tools.build_function_symbols` does not exist.

- [ ] **Step 3: Implement the generator types, safety checks, palette rendering, and deterministic SQL**

Use only the Python standard library. Define immutable palettes and reject unsafe XML before any
output is written:

```python
@dataclass(frozen=True)
class Palette:
    primary: str
    accent: str


ACTIVE_PALETTE = Palette(primary="#f2f5f6", accent="#a5ec60")
INACTIVE_PALETTE = Palette(primary="#879398", accent="#66736c")
PRINT_PALETTE = Palette(primary="#111111", accent="#1c621b")


FORBIDDEN_ELEMENTS = {"script", "text", "image", "foreignObject", "style"}
FORBIDDEN_ATTRIBUTE_PREFIXES = ("on",)
ALLOWED_ELEMENTS = {"svg", "g", "path", "line", "polyline", "polygon", "rect", "circle", "ellipse"}


def render_svg(svg_text: str, palette: Palette) -> str:
    root = ET.fromstring(svg_text)
    for element in root.iter():
        role = element.attrib.pop("data-rk-role", "")
        color = palette.primary if role == "primary" else palette.accent if role == "accent" else ""
        if color:
            if element.get("stroke") not in (None, "none"):
                element.set("stroke", color)
            if element.get("fill") not in (None, "none"):
                element.set("fill", color)
    return ET.tostring(root, encoding="unicode", short_empty_elements=True)


def encode_svg(svg_text: str) -> str:
    payload = base64.b64encode(svg_text.encode("utf-8")).decode("ascii")
    return "data:image/svg+xml;base64," + payload
```

Implement `--check` without filesystem writes, atomic migration output through a sibling temporary
file plus `Path.replace`, and stable ordering by `sortOrder`, then `key`.

- [ ] **Step 4: Run generator tests and syntax validation**

Run:

```powershell
python -m unittest tools.test_build_function_symbols -v
python -m py_compile tools/build_function_symbols.py tools/test_build_function_symbols.py
```

Expected: all unit tests PASS and both files compile without output.

- [ ] **Step 5: Commit the generator**

```powershell
git add tools/build_function_symbols.py tools/test_build_function_symbols.py
git commit -m "build: add function symbol generator"
```

---

### Task 2: Author and validate the complete Werkstatt-Linie asset library

**Files:**

- Create: `assets/function-symbols/workshop-line/manifest.json`
- Create: `assets/function-symbols/workshop-line/*.svg`
- Modify: `tools/test_build_function_symbols.py`

**Interfaces:**

- Consumes: `load_library`, `validate_svg`, and `--check` from Task 1.
- Produces: one `Library` containing exactly 94 validated symbols.
- Produces: one SVG geometry file matching every manifest `file` value.

- [ ] **Step 1: Add the failing repository-library completeness test**

```python
def test_repository_library_is_complete(self) -> None:
    root = Path("assets/function-symbols/workshop-line")
    library = load_library(root)
    self.assertEqual(len(library.symbols), 94)
    self.assertEqual(sum(item.ecos_code is not None for item in library.symbols), 86)
    self.assertEqual(len({item.key for item in library.symbols}), 94)
    self.assertEqual(len({item.ecos_code for item in library.symbols if item.ecos_code is not None}), 86)
```

- [ ] **Step 2: Run the completeness test and verify it fails**

Run:

```powershell
python -m unittest tools.test_build_function_symbols.FunctionSymbolGeneratorTest.test_repository_library_is_complete -v
```

Expected: FAIL because the asset directory does not exist.

- [ ] **Step 3: Create the manifest with all stable identities**

Use this exact manifest shape:

```json
{
  "version": 1,
  "library": "railkeeper-workshop-line",
  "symbols": [
    {
      "key": "light",
      "label": "Licht",
      "category": "Licht",
      "description": "RailKeeper Funktionssymbol: Licht.",
      "sortOrder": 10,
      "file": "light.svg"
    },
    {
      "key": "esu-f003-stirnbeleuchtung",
      "label": "Stirnbeleuchtung",
      "category": "Licht",
      "description": "RailKeeper Funktionssymbol: Stirnbeleuchtung.",
      "sortOrder": 103,
      "ecosCode": 3,
      "file": "f003-stirnbeleuchtung.svg"
    }
  ]
}
```

The final manifest must include these exact identity groups:

- General fallbacks: `light`, `sound`, `horn`, `coupling`, `smoke`, `drive`, `warning`, `standard`.
- General detailed: `002`, `021`, `101`, `107`.
- Lighting and displays: `003`, `004`, `005`, `014`, `022`, `026`, `036`, `050`, `051`, `054`,
  `055`, `060`, `061`, `078`, `079`, `080`, `081`, `084`, `086`, `087`, `088`.
- Sound and announcements: `006`, `007`, `008`, `015`, `016`, `017`, `029`, `030`, `031`, `034`,
  `035`, `038`, `039`, `065`, `066`, `074`, `106`.
- Driving and braking: `009`, `010`, `032`, `033`, `043`, `057`, `059`, `076`, `077`.
- Steam and auxiliaries: `012`, `019`, `020`, `028`, `040`, `041`, `042`, `044`, `045`, `046`,
  `047`, `048`, `058`, `062`, `063`, `068`, `069`, `124`.
- Coupling, doors, and pantographs: `011`, `013`, `018`, `023`, `089`, `090`, `093`, `094`.
- Crane functions: `024`, `025`, `027`, `095`, `096`, `097`, `098`, `099`, `114`.

Derive each detailed key, German label, and current sort order from the stable identities in the
sanitized `0020` input. Do not derive or reuse SVG geometry, source filenames, descriptions, or
other graphic metadata from the retired payload.

- [ ] **Step 4: Author all 94 SVG files using the approved geometry rules**

Use semantic roles rather than final palette colors. A complete headlight source follows this form:

```svg
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 64 64">
  <g data-rk-role="primary" fill="none" stroke="#111111" stroke-width="2.8"
     stroke-linecap="round" stroke-linejoin="round">
    <path d="M18 18h14c9 0 16 6 16 14s-7 14-16 14H18z"/>
    <path d="M23 23v18"/>
  </g>
  <g data-rk-role="accent" fill="none" stroke="#419310" stroke-width="2.8"
     stroke-linecap="round" stroke-linejoin="round">
    <path d="M12 23l-6-4M12 32H5M12 41l-6 4"/>
  </g>
</svg>
```

Use a shared base geometry for paired functions, then add these unambiguous modifiers:

- front and rear: right- and left-facing side marker;
- up and down: vertical arrow outside the moving part;
- left and right: horizontal arrow outside the crane body;
- sound on/off: speaker plus waves or speaker plus slash;
- light variants: lamp body plus distinct beam direction or location marker;
- mechanical motion: object silhouette plus arrow, never an arrow alone;
- named sounds: source object such as bell, horn, whistle, rail joint, brake, pump, or radio;
- general function: a neutral circular `F`-free control glyph, because SVG text is forbidden.

- [ ] **Step 5: Validate the complete library and generate a contact sheet**

Run:

```powershell
python tools/build_function_symbols.py --check
python tools/build_function_symbols.py --contact-sheet .cache/function-symbols/contact-sheet.svg
```

Expected: `validated 94 symbols (86 ECoS mappings, 8 fallbacks)` and a contact sheet containing all
94 labels and active symbols. Inspect the sheet at 19, 24, and 32 pixel rows before committing.

- [ ] **Step 6: Run generator unit tests**

Run:

```powershell
python -m unittest tools.test_build_function_symbols -v
```

Expected: all tests PASS.

- [ ] **Step 7: Commit the source library**

```powershell
git add assets/function-symbols/workshop-line tools/test_build_function_symbols.py
git commit -m "feat: add Werkstatt-Linie symbol sources"
```

Do not add `.cache/function-symbols/contact-sheet.svg`.

---

### Task 3: Replace historical payloads and migrate installed databases

**Files:**

- Modify: `backend/migrations/0020_esu_function_symbols.sql`
- Modify: `backend/migrations/0025_update_esu_function_symbols_variant2.sql`
- Create: `backend/migrations/0064_replace_bundled_function_symbols.sql`
- Create: `backend/internal/infrastructure/function_symbol_migration_test.go`
- Modify: `backend/internal/application/master_data_test.go`

**Interfaces:**

- Consumes: deterministic SQL from Task 1 and all source assets from Task 2.
- Produces: `0064_replace_bundled_function_symbols.sql` with 94 targeted upserts.
- Preserves: `(type, key)`, row IDs, labels, active flags, sort orders, timestamps, and stored references.

- [ ] **Step 1: Write the failing historical-upgrade migration test**

Create a database migrated through `0063`, replace one bundled image with a retired marker, delete a
second bundled entry, apply `0064`, and assert that both identities receive the new contract:

```go
func TestFunctionSymbolMigrationReplacesBundledArtworkAndPreservesCustomRows(t *testing.T) {
    root := t.TempDir()
    migrationsDir := filepath.Join(root, "migrations")
    if err := os.Mkdir(migrationsDir, 0o700); err != nil {
        t.Fatal(err)
    }
    copyMigrationsThrough(t, filepath.Join("..", "..", "migrations"), migrationsDir,
        "0063_exhibition_workspace.sql")
    db, err := infrastructure.OpenSQLite(filepath.Join(root, "data"))
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { _ = db.Close() })
    if err := infrastructure.Migrate(db, migrationsDir); err != nil {
        t.Fatal(err)
    }
    if _, err := db.Exec(`UPDATE master_data_entries
        SET metadata_json='{"sourceDocument":"retired.zip","imageData":"retired"}'
        WHERE type='symbols' AND key='esu-f003-stirnbeleuchtung'`); err != nil {
        t.Fatal(err)
    }
    if _, err := db.Exec(`DELETE FROM master_data_entries
        WHERE type='symbols' AND key='esu-f004-innenraumbeleuchtung'`); err != nil {
        t.Fatal(err)
    }
    if _, err := db.Exec(`INSERT INTO master_data_entries(
        id,type,key,label,active,sort_order,metadata_json,created_at,updated_at,origin
    ) VALUES('symbols:club','symbols','club','Club',1,900,
        '{"imageData":"custom"}','now','now','custom')`); err != nil {
        t.Fatal(err)
    }

    applyMigrationFile(t, db, "0064_replace_bundled_function_symbols.sql")

    assertWerkstattSymbol(t, db, "esu-f003-stirnbeleuchtung", 3)
    assertWerkstattSymbol(t, db, "esu-f004-innenraumbeleuchtung", 4)
    var customMetadata string
    if err := db.QueryRow(`SELECT metadata_json FROM master_data_entries
        WHERE type='symbols' AND key='club'`).Scan(&customMetadata); err != nil {
        t.Fatal(err)
    }
    if customMetadata != `{"imageData":"custom"}` {
        t.Fatalf("custom metadata changed: %s", customMetadata)
    }
}
```

Add the assertion helper in the same test file:

```go
func assertWerkstattSymbol(t *testing.T, db *sql.DB, key string, wantCode int) {
    t.Helper()
    var library string
    var code int
    var imageData string
    if err := db.QueryRow(`SELECT
        json_extract(metadata_json, '$.library'),
        json_extract(metadata_json, '$.ecosCode'),
        json_extract(metadata_json, '$.imageData')
        FROM master_data_entries WHERE type='symbols' AND key=?`, key).
        Scan(&library, &code, &imageData); err != nil {
        t.Fatal(err)
    }
    if library != "railkeeper-workshop-line" || code != wantCode {
        t.Fatalf("symbol %s library=%q code=%d", key, library, code)
    }
    if !strings.HasPrefix(imageData, "data:image/svg+xml;base64,") {
        t.Fatalf("symbol %s has invalid imageData", key)
    }
}
```

- [ ] **Step 2: Run the migration test and verify the missing migration failure**

Run:

```powershell
cd backend
go test ./internal/infrastructure -run TestFunctionSymbolMigrationReplacesBundledArtworkAndPreservesCustomRows -count=1
```

Expected: FAIL because migration `0064_replace_bundled_function_symbols.sql` is missing.

- [ ] **Step 3: Sanitize historical migrations**

Replace `0020` with compact `INSERT OR IGNORE` rows for the same 86 detailed keys and neutral factual
metadata containing `category`, `description`, and numeric `ecosCode`. Do not include `imageData`,
`activeImageData`, `inactiveImageData`, `sourceDocument`, `variant`, `originalName`, or graphic source
filenames.

Replace the body of `0025` with:

```sql
-- The former bundled function-symbol artwork update was retired.
-- Migration 0064 installs the independently authored RailKeeper Werkstatt-Linie library.
SELECT 1;
```

- [ ] **Step 4: Generate migration 0064**

Run:

```powershell
python tools/build_function_symbols.py --write-migration backend/migrations/0064_replace_bundled_function_symbols.sql
```

Expected: `wrote migration for 94 bundled symbols`.

Each statement must use `INSERT ... ON CONFLICT(type, key) DO UPDATE`. The insert branch restores a
missing bundled identity with its canonical ID, label, active flag, sort order, empty source URL,
new metadata, and `origin='bundled'`. The conflict branch updates only `source_url`, `metadata_json`,
`updated_at`, and `origin`; it must not overwrite the existing ID, label, active flag, sort order, or
creation time.

- [ ] **Step 5: Update the master-data contract test**

Rename `TestESUFunctionSymbolsSeededWithImages` to
`TestRailKeeperFunctionSymbolsSeededWithImages`. Assert 94 library entries, 86 numeric ECoS codes,
the `railkeeper-workshop-line` library marker, version `1`, and valid data URLs for all three image
fields. Keep the existing Fahrgeräusch label assertion.

- [ ] **Step 6: Run migration and master-data tests**

Run:

```powershell
cd backend
go test ./internal/infrastructure -run FunctionSymbol -count=1
go test ./internal/application -run RailKeeperFunctionSymbols -count=1
```

Expected: both commands PASS.

- [ ] **Step 7: Prove retired source data is absent from the current tree**

Run from the repository root:

```powershell
rg -n -S "ESU_Funktionssymbole|Variante 2 Feinlinien|sourceDocument.*ESU" backend assets tools
rg -n -S "derived from ESU/ECoS function-symbol" THIRD_PARTY_NOTICES.md docs
```

Expected: no matches. Factual ECoS trademarks, protocol references, compatibility codes, and stable
`esu-fNNN-*` keys are allowed and are not part of this scan.

- [ ] **Step 8: Commit the migration**

```powershell
git add backend/migrations/0020_esu_function_symbols.sql
git add backend/migrations/0025_update_esu_function_symbols_variant2.sql
git add backend/migrations/0064_replace_bundled_function_symbols.sql
git add backend/internal/infrastructure/function_symbol_migration_test.go
git add backend/internal/application/master_data_test.go
git commit -m "feat: replace bundled function symbol artwork"
```

---

### Task 4: Centralize image-variant selection and replace fallback icons

**Files:**

- Create: `frontend/src/shared/functionSymbolImages.ts`
- Create: `frontend/src/shared/functionSymbolImages.test.ts`
- Create: `frontend/src/shared/functionSymbolFallbackIcons.tsx`
- Modify: `frontend/src/shared/functionSymbols.tsx`
- Modify: `frontend/src/shared/functionSymbols.test.tsx`

**Interfaces:**

- Produces: `type FunctionSymbolImageVariant = "active" | "inactive" | "print"`
- Produces: `functionSymbolImageData(metadata?: Record<string, unknown>, variant?: FunctionSymbolImageVariant): string`
- Produces: `RailKeeperFunctionSymbolFallback({ symbolKey, functionType }: Props): JSX.Element`
- Consumes: metadata fields written by migration `0064`.

- [ ] **Step 1: Write failing image-selection tests**

```typescript
describe("functionSymbolImageData", () => {
  const metadata = {
    imageData: "print",
    activeImageData: "active",
    inactiveImageData: "inactive",
    svgData: "legacy"
  };

  it("selects the requested palette without changing geometry ownership", () => {
    expect(functionSymbolImageData(metadata, "active")).toBe("active");
    expect(functionSymbolImageData(metadata, "inactive")).toBe("inactive");
    expect(functionSymbolImageData(metadata, "print")).toBe("print");
  });

  it("falls back safely for custom and legacy uploads", () => {
    expect(functionSymbolImageData({ imageData: "custom" }, "active")).toBe("custom");
    expect(functionSymbolImageData({ svgData: "legacy" }, "print")).toBe("legacy");
    expect(functionSymbolImageData(undefined, "active")).toBe("");
  });
});
```

- [ ] **Step 2: Run the targeted test and verify it fails**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/shared/functionSymbolImages.test.ts
```

Expected: FAIL because `functionSymbolImages.ts` does not exist.

- [ ] **Step 3: Implement the pure selection helper**

```typescript
export type FunctionSymbolImageVariant = "active" | "inactive" | "print";

function metadataString(metadata: Record<string, unknown> | undefined, key: string) {
  const value = metadata?.[key];
  return typeof value === "string" ? value : "";
}

export function functionSymbolImageData(
  metadata?: Record<string, unknown>,
  variant: FunctionSymbolImageVariant = "active"
) {
  const keys = variant === "active"
    ? ["activeImageData", "imageData", "svgData"]
    : variant === "inactive"
      ? ["inactiveImageData", "imageData", "activeImageData", "svgData"]
      : ["imageData", "activeImageData", "svgData"];
  return keys.map((key) => metadataString(metadata, key)).find(Boolean) || "";
}
```

- [ ] **Step 4: Write failing fallback-renderer tests**

Extend `functionSymbols.test.tsx` to render `functionSymbolIcon` with no metadata for `light`,
`sound`, `horn`, `coupling`, `smoke`, `drive`, `warning`, and `standard`. Assert each output has
`data-rk-function-symbol` equal to the resolved fallback key and contains no Lucide class name.

- [ ] **Step 5: Implement the eight RailKeeper fallback components**

Use a typed geometry registry and one shared SVG shell. Define aliases explicitly so unknown keys
resolve to `standard`:

```tsx
import type { ReactNode } from "react";

type FallbackKey =
  | "light" | "sound" | "horn" | "coupling"
  | "smoke" | "drive" | "warning" | "standard";

type Props = { symbolKey?: string; functionType?: string };

const primary = "function-symbol-primary";
const accent = "function-symbol-accent";

const fallbackGeometry: Record<FallbackKey, ReactNode> = {
  light: <><path className={primary} d="M18 18h14c9 0 16 6 16 14s-7 14-16 14H18zM23 23v18" />
    <path className={accent} d="M12 23l-6-4M12 32H5M12 41l-6 4" /></>,
  sound: <><path className={primary} d="M11 27h10l12-10v30L21 37H11z" />
    <path className={accent} d="M41 25c4 4 4 10 0 14M47 19c8 8 8 18 0 26" /></>,
  horn: <><path className={primary} d="M12 27h13l24-9v28l-24-9H12zM25 37l-3 10h-7l2-10" />
    <path className={accent} d="M8 25v14" /></>,
  coupling: <><path className={primary} d="M7 25h12l7 7-7 7H7M57 25H45l-7 7 7 7h12" />
    <path className={accent} d="M24 32h16M29 27l-5 5 5 5M35 27l5 5-5 5" /></>,
  smoke: <><path className={primary} d="M18 44h29l-3 9H21zM23 44V32h17v12" />
    <path className={accent} d="M26 27c-5-5 5-7 0-12M34 27c-5-6 6-8 1-15M42 28c-4-4 4-6 1-10" /></>,
  drive: <><circle className={primary} cx="32" cy="34" r="20" />
    <path className={primary} d="M17 44h30M32 34l11-9" />
    <circle className={accent} cx="32" cy="34" r="3" /></>,
  warning: <><path className={primary} d="M32 9l25 44H7zM32 23v15" />
    <circle className={accent} cx="32" cy="46" r="2" /></>,
  standard: <circle className={primary} cx="32" cy="32" r="22" />
};

function resolveFallbackKey(symbolKey?: string, functionType?: string): FallbackKey {
  const value = symbolKey || functionType || "standard";
  if (value === "licht") return "light";
  if (value === "kupplung") return "coupling";
  if (value === "rauch") return "smoke";
  return value in fallbackGeometry ? value as FallbackKey : "standard";
}

export function RailKeeperFunctionSymbolFallback({ symbolKey, functionType }: Props) {
  const key = resolveFallbackKey(symbolKey, functionType);
  return (
    <svg viewBox="0 0 64 64" data-rk-function-symbol={key} aria-hidden="true">
      {fallbackGeometry[key]}
    </svg>
  );
}
```

Keep this registry geometrically synchronized with the eight RailKeeper-owned fallback SVG source
files. Remove the eight Lucide imports from `functionSymbols.tsx`, including the Lucide circle used
for “Kein Symbol”.

- [ ] **Step 6: Make application rendering prefer active data**

Replace the local metadata helper in `functionSymbols.tsx` with
`functionSymbolImageData(metadata, "active")`. Keep uploaded data in the isolated `<img>` path and
use `RailKeeperFunctionSymbolFallback` only when no image data exists.

- [ ] **Step 7: Run targeted frontend tests**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/shared/functionSymbolImages.test.ts src/shared/functionSymbols.test.tsx
```

Expected: all tests PASS.

- [ ] **Step 8: Commit shared rendering**

```powershell
git add frontend/src/shared/functionSymbolImages.ts frontend/src/shared/functionSymbolImages.test.ts
git add frontend/src/shared/functionSymbolFallbackIcons.tsx
git add frontend/src/shared/functionSymbols.tsx frontend/src/shared/functionSymbols.test.tsx
git commit -m "feat: render RailKeeper function symbols"
```

---

### Task 5: Apply active, print, and tile behavior across every consumer

**Files:**

- Modify: `frontend/src/features/settings/settingsModel.ts`
- Create: `frontend/src/features/settings/settingsModel.test.ts`
- Modify: `frontend/src/features/exhibition/ExhibitionView.tsx`
- Modify: `frontend/src/features/exhibition/ExhibitionWorkspaceView.test.tsx`
- Modify: `frontend/src/styles/base.css`
- Modify: `frontend/src/styles/vehicle-functions-cv-maintenance.css`
- Modify: `frontend/src/styles/exhibition.css`

**Interfaces:**

- Consumes: `functionSymbolImageData` from Task 4.
- Produces: consistent active images in the app and print images on white pages.
- Produces CSS tokens: `--function-symbol-bg`, `--function-symbol-border`,
  `--function-symbol-primary`, `--function-symbol-accent`, and `--function-symbol-muted`.

- [ ] **Step 1: Write failing consumer tests**

Add a settings-model test proving `masterDataImage` selects `activeImageData` before `imageData`.
Extend the exhibition workspace print test with metadata containing both values and assert the
generated print iframe contains `src="print-data"` and not `src="active-data"`.

- [ ] **Step 2: Run targeted consumer tests and verify they fail**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/settingsModel.test.ts
npm.cmd run test:run -- src/features/exhibition/ExhibitionWorkspaceView.test.tsx
```

Expected: FAIL because settings still prefers `imageData` and exhibition uses a local selector.

- [ ] **Step 3: Replace duplicated metadata selection**

Import the shared helper in both consumers:

```typescript
export function masterDataImage(entry: MasterDataEntry) {
  return functionSymbolImageData(entry.metadata, "active");
}
```

In `ExhibitionView.tsx`, remove `symbolImageDataFromMetadata` and call
`functionSymbolImageData(metadata, "print")` inside `printFunctionChips`.

- [ ] **Step 4: Add stable dark-tile tokens for both themes**

Use these exact token roles and values:

```css
:root {
  --function-symbol-bg: #111719;
  --function-symbol-border: rgba(11, 30, 38, 0.34);
  --function-symbol-primary: #f2f5f6;
  --function-symbol-accent: #419310;
  --function-symbol-muted: #879398;
}

:root[data-theme="dark"] {
  --function-symbol-bg: #0d1417;
  --function-symbol-border: rgba(165, 236, 96, 0.28);
  --function-symbol-primary: #f2f5f6;
  --function-symbol-accent: #a5ec60;
  --function-symbol-muted: #879398;
}
```

Update fallback paths to use the primary and accent tokens. Keep the tile visible in light mode,
dark mode, selected menu rows, disabled rows, and 24-pixel exhibition usage. Print chips continue to
use transparent `imageData` on white paper and must not print the dark tile.

- [ ] **Step 5: Run targeted tests and the frontend build**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/settings/settingsModel.test.ts
npm.cmd run test:run -- src/features/exhibition/ExhibitionWorkspaceView.test.tsx
npm.cmd run test:run -- src/shared/functionSymbols.test.tsx
npm.cmd run build
```

Expected: tests PASS and Vite finishes with `built in` output.

- [ ] **Step 6: Commit consumer integration**

```powershell
git add frontend/src/features/settings/settingsModel.ts
git add frontend/src/features/settings/settingsModel.test.ts
git add frontend/src/features/exhibition/ExhibitionView.tsx
git add frontend/src/features/exhibition/ExhibitionWorkspaceView.test.tsx
git add frontend/src/styles/base.css frontend/src/styles/vehicle-functions-cv-maintenance.css
git add frontend/src/styles/exhibition.css
git commit -m "feat: integrate function symbol palettes"
```

---

### Task 6: Protect the installed symbol library during backup restore

**Files:**

- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Interfaces:**

- Produces: `readBundledFunctionSymbols(ctx context.Context, tx *sql.Tx) ([]bundledFunctionSymbol, error)`
- Produces: `restoreBundledFunctionSymbols(ctx context.Context, tx *sql.Tx, symbols []bundledFunctionSymbol) error`
- Consumes: the current migrated bundled rows before backup tables are cleared.

- [ ] **Step 1: Write the failing old-backup restore regression test**

Create a current database, set the installed `light` metadata to the Werkstatt library marker, export
a backup, mutate the backup's bundled `light` row to retired metadata, and append a separate custom
symbol. After import, assert the bundled row is current and the custom row is restored:

```go
func TestBackupRestoreKeepsCurrentBundledFunctionSymbols(t *testing.T) {
    dataDir := t.TempDir()
    db := backupTestDB(t, dataDir)
    service := application.NewBackupService(db, dataDir)
    doc, err := service.Export(t.Context())
    if err != nil {
        t.Fatal(err)
    }
    for _, row := range doc.Tables["master_data_entries"] {
        if row["type"] == "symbols" && row["key"] == "light" {
            row["metadata_json"] = `{"sourceDocument":"retired.zip","imageData":"retired"}`
        }
    }
    doc.Tables["master_data_entries"] = append(doc.Tables["master_data_entries"], map[string]any{
        "id": "symbols:club", "type": "symbols", "key": "club", "label": "Club",
        "active": 1, "sort_order": 900, "source_url": "",
        "metadata_json": `{"imageData":"custom"}`, "created_at": "now", "updated_at": "now",
    })

    if _, err := service.Import(t.Context(), doc); err != nil {
        t.Fatal(err)
    }
    assertSymbolLibrary(t, db, "light", "railkeeper-workshop-line")
    assertSymbolImage(t, db, "club", "custom")
}
```

Add focused assertion helpers in the same test file:

```go
func assertSymbolLibrary(t *testing.T, db *sql.DB, key, wantLibrary string) {
    t.Helper()
    var library string
    if err := db.QueryRow(`SELECT json_extract(metadata_json, '$.library')
        FROM master_data_entries WHERE type='symbols' AND key=?`, key).Scan(&library); err != nil {
        t.Fatal(err)
    }
    if library != wantLibrary {
        t.Fatalf("symbol %s library=%q want=%q", key, library, wantLibrary)
    }
}

func assertSymbolImage(t *testing.T, db *sql.DB, key, wantImage string) {
    t.Helper()
    var image string
    if err := db.QueryRow(`SELECT json_extract(metadata_json, '$.imageData')
        FROM master_data_entries WHERE type='symbols' AND key=?`, key).Scan(&image); err != nil {
        t.Fatal(err)
    }
    if image != wantImage {
        t.Fatalf("symbol %s image=%q want=%q", key, image, wantImage)
    }
}
```

- [ ] **Step 2: Run the regression test and verify it fails**

Run:

```powershell
cd backend
go test ./internal/application -run TestBackupRestoreKeepsCurrentBundledFunctionSymbols -count=1
```

Expected: FAIL because restored backup metadata replaces the installed `light` row.

- [ ] **Step 3: Capture complete bundled symbol rows before destructive restore**

Define a focused row type containing all `master_data_entries` columns. Query only
`type='symbols' AND origin='bundled' AND json_extract(metadata_json, '$.library')='railkeeper-workshop-line'`.
Call this reader next to `readBundledMasterDataIdentities` before the table-clear loop.

- [ ] **Step 4: Restore captured rows before origin reconciliation**

Upsert the captured rows after backup rows and legacy article data are restored, but before
`reconcileBundledMasterDataIdentities`. On `(type, key)` conflict, restore the captured ID, label,
active flag, sort order, source URL, metadata, timestamps, and `origin='bundled'`. Return contextual
errors naming the affected key.

- [ ] **Step 5: Run restore and backup suites**

Run:

```powershell
cd backend
go test ./internal/application -run "BackupRestore|BackupExportsAndRestores" -count=1
```

Expected: all selected tests PASS, including origin reconciliation and custom-symbol restoration.

- [ ] **Step 6: Commit restore protection**

```powershell
git add backend/internal/application/backup.go backend/internal/application/backup_test.go
git commit -m "fix: preserve bundled symbols during restore"
```

---

### Task 7: Preserve ECoS mapping and update maintained documentation

**Files:**

- Modify: `frontend/src/features/importExport/importExportHelpers.test.tsx`
- Modify: `THIRD_PARTY_NOTICES.md`
- Modify: `docs/roadmap.md`
- Modify: `docs/site/administration/master-data-general.md`
- Modify: `docs/site/de/administration/master-data-general.md`
- Modify: `docs/site/guide/vehicles/decoder-cv.md`
- Modify: `docs/site/de/guide/vehicles/decoder-cv.md`
- Create: `docs/releases/v0.1.20.md`

**Interfaces:**

- Consumes: numeric `metadata.ecosCode` written by migration `0064`.
- Preserves: existing `findSymbolByECoSCode(code, symbols)` public behavior.
- Produces: notices that distinguish RailKeeper-owned graphics from ECoS interoperability.

- [ ] **Step 1: Add an explicit ECoS metadata lookup test**

```typescript
it("maps ECoS codes through neutral compatibility metadata", () => {
  const symbols = [{
    id: "symbols:headlight",
    type: "symbols",
    key: "stable-headlight",
    label: "Stirnbeleuchtung",
    active: true,
    sortOrder: 103,
    metadata: { ecosCode: 3, library: "railkeeper-workshop-line" },
    createdAt: "2026-08-21T00:00:00Z",
    updatedAt: "2026-08-21T00:00:00Z"
  }];
  const suggestions = ecosFunctionSuggestions({
    objectId: 1001,
    functions: [{ index: 0, description: 3 }]
  }, symbols);
  expect(suggestions[0]).toMatchObject({
    functionKey: "F0",
    symbolKey: "stable-headlight",
    name: "Stirnbeleuchtung"
  });
});
```

- [ ] **Step 2: Run the ECoS helper test**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/importExport/importExportHelpers.test.tsx
```

Expected: PASS because `symbolCodeSignal` already reads `metadata.ecosCode`. If it fails, make the
minimal correction in `importExportHelpers.tsx` and include that file in this task's commit.

- [ ] **Step 3: Update third-party notices**

Keep this factual text under the ESU and ECoS heading:

```markdown
ECoS is a trademark of ESU electronic solutions ulm GmbH & Co. KG. RailKeeper is an independent
project and is not affiliated with or endorsed by ESU.

RailKeeper's bundled function-symbol graphics are independently authored as part of the RailKeeper
project. Numeric ECoS function-description codes are retained solely for interoperability.
```

Remove the prior statement that locally stored function-symbol graphics are derived from ESU/ECoS
material.

- [ ] **Step 4: Update roadmap, guides, and release notes**

State that the bundled library is RailKeeper Werkstatt-Linie, that administrators can still upload
custom symbols, and that ECoS codes are compatibility metadata. In the actual target release file,
state that direct image customizations made to one of the 94 bundled keys are replaced during this
library update, while custom keys remain untouched.

- [ ] **Step 5: Verify documentation and compatibility tests**

Run:

```powershell
cd frontend
npm.cmd run test:run -- src/features/importExport/importExportHelpers.test.tsx
cd ..
rg -n -S "derived from ESU/ECoS function-symbol" THIRD_PARTY_NOTICES.md docs
rg -n -S "ESU_Funktionssymbole|Variante 2 Feinlinien" backend assets tools
```

Expected: the test PASSes and the repository scan has no matches.

- [ ] **Step 6: Commit compatibility documentation**

```powershell
git add frontend/src/features/importExport/importExportHelpers.test.tsx
git add THIRD_PARTY_NOTICES.md docs/roadmap.md
git add docs/site/administration/master-data-general.md
git add docs/site/de/administration/master-data-general.md
git add docs/site/guide/vehicles/decoder-cv.md
git add docs/site/de/guide/vehicles/decoder-cv.md
git add docs/releases/v0.1.20.md
git commit -m "docs: document RailKeeper symbol ownership"
```

Only add `importExportHelpers.tsx` if Step 2 required the minimal compatibility correction.

---

### Task 8: Run full verification and visual acceptance

**Files:**

- Modify only files required to fix failures caused by Tasks 1 through 7.
- Do not commit generated `frontend/dist`, `.cache`, contact sheets, or test artifacts.

**Interfaces:**

- Consumes: the complete implementation.
- Produces: verified backend, frontend, migration, repository-cleanliness, and visual results.

- [ ] **Step 1: Run the generator and repository safety checks**

```powershell
python tools/build_function_symbols.py --check
python -m unittest tools.test_build_function_symbols -v
rg -n -S "ESU_Funktionssymbole|Variante 2 Feinlinien|sourceDocument.*ESU" backend assets tools
rg -n -S "derived from ESU/ECoS function-symbol" THIRD_PARTY_NOTICES.md docs
```

Expected: 94 symbols validate, all Python tests PASS, and the scan has no matches.

- [ ] **Step 2: Run all backend tests**

```powershell
cd backend
go test ./...
```

Expected: all Go packages PASS.

- [ ] **Step 3: Run the full frontend test suite and production build**

```powershell
cd frontend
npm.cmd run test:run
npm.cmd run build
```

Expected: all Vitest files PASS and Vite produces a successful production build.

- [ ] **Step 4: Generate and inspect the final contact sheet**

```powershell
python tools/build_function_symbols.py --contact-sheet .cache/function-symbols/contact-sheet.svg
```

Inspect all 94 symbols in active and print palettes at 19, 24, and 32 pixels. Reject any symbol with
ambiguous direction, clipped strokes, indistinguishable paired variants, lost detail, or text-shaped
geometry. Correct only the affected source SVG, regenerate migration `0064`, and repeat Tasks 1
through 3 verification for any changed asset.

- [ ] **Step 5: Visually verify application consumers**

Check light and dark mode in the vehicle function editor, settings symbol preview, exhibition entry
dialog, exhibition workspace, and exhibition print preview. Also check hover, focus, selected,
disabled, long German labels, and mobile width. Confirm print icons use dark contours on white paper
without a dark tile.

- [ ] **Step 6: Review the final diff and generated-file boundaries**

```powershell
git diff --check
git status --short
git diff --stat
```

Expected: no whitespace errors, no generated `frontend/dist`, `.cache`, data files, secrets, or
unrelated user files in the diff. The existing untracked `facebook-*` paths remain untouched.

- [ ] **Step 7: Commit verification fixes only if needed**

If verification required source corrections, stage only those corrections and their regenerated
migration, then commit:

```powershell
git add assets/function-symbols/workshop-line backend/migrations/0064_replace_bundled_function_symbols.sql
git commit -m "fix: refine function symbol legibility"
```

If no corrections were needed, do not create an empty verification commit.
