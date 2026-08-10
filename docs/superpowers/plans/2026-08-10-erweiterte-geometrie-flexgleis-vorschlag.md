# Flexgleis-Grundmodell und Verlaufsvorschlag Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** TILLIG 83125 als objektbezogenes Flexgleis mit serverseitig abgeleiteter Bézier-Geometrie,
explizitem Verlaufsvorschlag und vollständiger Integration in den vorhandenen Gleisplaner bereitstellen.

**Architecture:** Die unveränderliche Geometriebibliothek beschreibt Produktlänge und empfohlenen
Mindestradius. Ein Planobjekt speichert nur kompakte Flexparameter; fokussierte Domänenmodule leiten
effektive Route, Länge, Radius und Vorschläge ab. Bestehende Analysen konsumieren diese gemeinsame
effektive Geometrie, während ein app-eigener Dialog Vorschau und bewusste Übernahme trennt.

**Tech Stack:** Go, SQLite, React, TypeScript, SVG, Vitest, OpenAPI, RailKeeper-i18n und bestehende
app-eigene UI-Komponenten

## Global Constraints

- Alles bleibt auf `dev/issue-36-advanced-geometry`; kein Push, PR, Merge oder Release.
- TILLIG 83125 verwendet 664 mm Planungsgrenze und 543 mm dokumentierte Radius-Empfehlung.
- Ein Flexobjekt entspricht genau einem physischen Stück; Verschnitt und Reststücke bleiben außen vor.
- Ein Vorschlag verändert nichts persistent, bis der Benutzer „Verlauf übernehmen“ auslöst.
- Überlange Verläufe sind nicht übernehmbar; Radiusunterschreitungen bleiben sichtbare Warnungen.
- Bestehende starre G1-Pläne, Lineage, Reservierungen und Backups 1 bis 10 bleiben kompatibel.
- Die Domäne ist alleiniger Owner der effektiven Geometrie und des serverseitigen Vorschlags.
- Kein echter Klothoidenlöser, kein automatisches Mehrgleis-Routing und keine freien Planobjekte.
- Deutsch und Englisch, Backend, Frontend und OpenAPI werden synchron gepflegt.
- Tests folgen red-green-refactor; jeder Task endet mit einem lokalen Commit.

---

### Task 1: Flexprodukt, Anlagenlimit und Backup-Grundlage persistieren

**Files:**
- Create: `backend/migrations/0052_flex_track_paths.sql`
- Create: `backend/internal/infrastructure/flex_track_migration_test.go`
- Modify: `backend/internal/domain/track_planner.go`
- Modify: `backend/internal/domain/track_planner_test.go`
- Modify: `backend/internal/application/layouts.go`
- Modify: `backend/internal/application/layouts_test.go`
- Modify: `backend/internal/infrastructure/layout_repository.go`
- Modify: `backend/internal/infrastructure/layout_repository_test.go`
- Modify: `backend/internal/infrastructure/track_planner_schema_test.go`
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Interfaces:**
- Produces: `TrackGeometryFlex TrackGeometryKind = "flex"`
- Produces: `TrackGeometryDefinition.MinimumRadiusMM *float64`
- Produces: `Layout.MinimumFlexRadiusMM *float64`
- Persists: `track_geometry_definitions.minimum_radius_mm`, `plan_track_objects.flex_path_json`,
  `layouts.minimum_flex_radius_mm`
- Produces: Backup-Version 11; Versionen bis 10 erhalten für neue Spalten `NULL`

- [x] **Step 1: Fehlschlagende Kind-, Layout-, Migrations- und Backuptests schreiben**

Der Domänentest nimmt `flex` in die gültigen Arten auf. Layouttests erstellen 700 mm, aktualisieren
auf 650 mm, entfernen den Wert und lehnen `0`, negative, `NaN` und unendliche Werte ab.

Der Migrationstest kopiert Migrationen bis 0051 in ein temporäres Verzeichnis, legt ein bestehendes
G1-Planobjekt an, führt alle Migrationen aus und prüft anschließend:

```go
var kind string
var length, radius float64
err := db.QueryRow(`
SELECT kind, length_mm, minimum_radius_mm
FROM track_geometry_definitions
WHERE id='tillig-tt-modellgleis-83125-v1'`).Scan(&kind, &length, &radius)
if err != nil || kind != "flex" || length != 664 || radius != 543 {
	t.Fatalf("unexpected flex definition: %q %.2f %.2f, %v", kind, length, radius, err)
}
```

Zusätzlich müssen `PRAGMA foreign_key_check` leer, das bestehende Objekt vorhanden und G1 weiterhin
`straight` sein. Der Backuptest exportiert Layoutlimit, Produkt-Mindestradius und einen Flexpfad,
restauriert Version 11 und entfernt für eine Version-10-Kopie genau die drei neuen Spalten.

- [x] **Step 2: Gezielte Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run:

```powershell
cd backend
go test ./internal/domain ./internal/application ./internal/infrastructure -run "Flex|BackupVersionEleven"
```

Expected: FAIL, weil `flex`, `MinimumRadiusMM`, `MinimumFlexRadiusMM` und Migration 0052 fehlen.

- [x] **Step 3: Migration 0052 implementieren**

Die Migration verwendet `PRAGMA defer_foreign_keys=ON`, baut
`track_geometry_definitions` mit der erweiterten Kind-Constraint und `minimum_radius_mm` neu auf,
kopiert alle Zeilen, stellt den Bibliotheksindex wieder her und ergänzt danach die beiden nullable
Spalten:

```sql
CREATE TABLE track_geometry_definitions_next (
  id TEXT PRIMARY KEY,
  library_id TEXT NOT NULL,
  article_number TEXT NOT NULL,
  name TEXT NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('straight', 'curve', 'turnout', 'crossing', 'flex')),
  length_mm REAL NOT NULL CHECK (length_mm > 0),
  minimum_radius_mm REAL CHECK (minimum_radius_mm IS NULL OR minimum_radius_mm > 0),
  geometry_json TEXT NOT NULL,
  source_url TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('draft', 'verified', 'retired')),
  created_at TEXT NOT NULL,
  FOREIGN KEY (library_id) REFERENCES track_geometry_libraries(id) ON DELETE RESTRICT,
  UNIQUE (library_id, article_number)
);
```

83125 erhält eine gerade Platzierungsgeometrie mit Ports `(0,0,180)` und `(664,0,0)`, eine Route von
0 bis 664, `kind='flex'`, `minimum_radius_mm=543` und die offizielle Produkt-URL. Die Spalten lauten:

```sql
ALTER TABLE plan_track_objects ADD COLUMN flex_path_json TEXT;
ALTER TABLE layouts ADD COLUMN minimum_flex_radius_mm REAL
  CHECK(minimum_flex_radius_mm IS NULL OR minimum_flex_radius_mm > 0);
```

- [x] **Step 4: Go-Modelle, Layoutvalidierung und Repository ergänzen**

```go
type TrackGeometryDefinition struct {
	// bestehende Felder
	MinimumRadiusMM *float64 `json:"minimumRadiusMm,omitempty"`
}

type Layout struct {
	// bestehende Felder
	MinimumFlexRadiusMM *float64 `json:"minimumFlexRadiusMm,omitempty"`
}
```

`CreateLayoutInput` erhält denselben Pointer. `validLayoutInput` akzeptiert `nil` oder einen endlichen
Wert größer als 0. Insert, Update, `layoutSelect` und `scanLayout` führen die Spalte direkt nach
`minimum_track_clearance_mm`. Geometrierepository und Scanner führen `minimum_radius_mm` direkt nach
`length_mm`.

- [x] **Step 5: Backup-Version 11 und Legacy-Normalisierung implementieren**

`backupVersion` wird 11, der Future-Version-Test erwartet 12. Vor dem Restore ergänzt die bestehende
Legacy-Normalisierung bei Versionen bis 10 fehlende Werte:

```go
if doc.Version <= 10 {
	for _, row := range doc.Tables["track_geometry_definitions"] {
		row["minimum_radius_mm"] = nil
	}
	for _, row := range doc.Tables["plan_track_objects"] {
		row["flex_path_json"] = nil
	}
	for _, row := range doc.Tables["layouts"] {
		row["minimum_flex_radius_mm"] = nil
	}
}
```

- [x] **Step 6: Task-1-Suiten ausführen und lokal committen**

Run: `cd backend; go test ./internal/domain ./internal/application ./internal/infrastructure`

Expected: PASS.

```powershell
git add backend/migrations/0052_flex_track_paths.sql `
  backend/internal/infrastructure/flex_track_migration_test.go `
  backend/internal/domain/track_planner.go backend/internal/domain/track_planner_test.go `
  backend/internal/application/layouts.go backend/internal/application/layouts_test.go `
  backend/internal/infrastructure/layout_repository.go `
  backend/internal/infrastructure/layout_repository_test.go `
  backend/internal/infrastructure/track_planner_schema_test.go `
  backend/internal/application/backup.go backend/internal/application/backup_test.go
git commit -m "feat(planner): persist flex track foundation"
```

---

### Task 2: Effektive Bézier-Geometrie ableiten und analysieren

**Files:**
- Create: `backend/internal/domain/track_flex_geometry.go`
- Create: `backend/internal/domain/track_flex_geometry_test.go`
- Modify: `backend/internal/domain/track_planner.go`
- Modify: `backend/internal/domain/track_plan_analysis.go`
- Modify: `backend/internal/domain/track_plan_analysis_test.go`
- Modify: `backend/internal/domain/track_plan_clearance.go`
- Modify: `backend/internal/domain/track_plan_clearance_test.go`
- Modify: `backend/internal/domain/track_plan_revision_diff.go`
- Modify: `backend/internal/domain/track_plan_revision_diff_test.go`

**Interfaces:**
- Produces: `FlexTrackPath`, `EffectiveTrackGeometry`, `BuildFlexTrackGeometry`
- Produces: `PlanTrackObject.EffectiveGeometry`, `EffectiveLengthMM`, `EffectiveMinimumRadiusMM`
- Produces: `TrackPlanIssueFlexRadiusBelowLimit = "flex_radius_below_limit"`
- Consumes: `TrackPlanLimits.MinimumFlexRadiusMM *float64`

- [x] **Step 1: Golden- und Grenztests für Bézier-Ableitung schreiben**

```go
path := FlexTrackPath{
	SchemaVersion: 1, EndXMM: 500, EndYMM: 100, EndDirectionDegrees: 20,
	StartHandleMM: 180, EndHandleMM: 170,
}
effective, err := BuildFlexTrackGeometry(path)
if err != nil {
	t.Fatal(err)
}
if len(effective.Geometry.Ports) != 2 || effective.Geometry.Ports[0].XMM != 0 ||
	len(effective.Geometry.Routes) != 1 || effective.LengthMM <= 500 ||
	effective.MinimumRadiusMM == nil || *effective.MinimumRadiusMM <= 0 {
	t.Fatalf("unexpected effective geometry: %#v", effective)
}
```

Weitere Tests prüfen den geraden 664-mm-Verlauf, S-Kurve, Normalisierung der Endrichtung, nicht
endliche Werte, nicht positive Tangenten, identische Endpunkte, maximale Segmentlänge 5 mm,
Sehnenabweichung 0,05 mm und deterministische Punktfolgen.

- [x] **Step 2: Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/domain -run "FlexTrack|EffectiveTrack"`

Expected: FAIL wegen fehlender Typen und Funktionen.

- [x] **Step 3: Fokussiertes Geometriemodul implementieren**

```go
type FlexTrackPath struct {
	SchemaVersion       int     `json:"schemaVersion"`
	EndXMM              float64 `json:"endXMm"`
	EndYMM              float64 `json:"endYMm"`
	EndDirectionDegrees float64 `json:"endDirectionDegrees"`
	StartHandleMM       float64 `json:"startHandleMm"`
	EndHandleMM         float64 `json:"endHandleMm"`
}

type EffectiveTrackGeometry struct {
	Geometry        TrackGeometry
	LengthMM        float64
	MinimumRadiusMM *float64
}

func BuildFlexTrackGeometry(path FlexTrackPath) (EffectiveTrackGeometry, error)
func EffectiveGeometryForObject(object PlanTrackObject) (EffectiveTrackGeometry, error)
```

`BuildFlexTrackGeometry` berechnet `P0` bis `P3`, unterteilt rekursiv bei Segmentlänge über 5 mm oder
Sehnenfehler über 0,05 mm und bricht oberhalb 4.096 Segmente mit `ErrInvalidFlexTrackPath` ab. Die
Krümmung wird an allen erzeugten Parameterstellen berechnet; Radiuswerte mit Krümmung unter `1e-12`
werden als unbegrenzt übersprungen.

- [x] **Step 4: Planobjekt und alle Analysen auf effektive Geometrie umstellen**

`PlanTrackObject` erhält `FlexPath *FlexTrackPath`, `EffectiveGeometry TrackGeometry`,
`EffectiveLengthMM float64` und `EffectiveMinimumRadiusMM *float64`. Starre Objekte werden mit der
Bibliotheksroute und -länge hydriert; Flexobjekte mit `BuildFlexTrackGeometry`.

`AnalyzeTrackPlanWithLimits`, `FindTrackSnap`, Überlappung, Port-Höhen, Clearance-Routen und Grade
verwenden Hilfsfunktionen für effektive Route, Ports und Länge. Für Flexobjekte gilt:

```go
limit := object.Geometry.MinimumRadiusMM
if limits.MinimumFlexRadiusMM != nil && (limit == nil || *limits.MinimumFlexRadiusMM > *limit) {
	limit = limits.MinimumFlexRadiusMM
}
if limit != nil && object.EffectiveMinimumRadiusMM != nil &&
	*object.EffectiveMinimumRadiusMM+1e-9 < *limit {
	// flex_radius_below_limit warning with RadiusMM and RadiusLimitMM
}
```

`TrackPlanIssue` erhält `RadiusMM` und `RadiusLimitMM`. `trackObjectsDiffer` vergleicht Flexpfade
feldweise mit `1e-9` Toleranz.

- [x] **Step 5: Analyse- und Revisionsfixtures ergänzen**

Tests prüfen: Radiuswarnung exakt einmal, Gleichheit ohne Warnung, verschärftes Anlagenlimit,
Flex-Steigung über effektive Länge, Flex-G1-Snapping, Überlappung, Ebenenkreuzung sowie eine
Flexpfadänderung als `changed` bei unveränderter Lineage.

- [x] **Step 6: Domain-Suite ausführen und lokal committen**

Run: `cd backend; go test ./internal/domain`

Expected: PASS.

```powershell
git add backend/internal/domain/track_flex_geometry.go `
  backend/internal/domain/track_flex_geometry_test.go backend/internal/domain/track_planner.go `
  backend/internal/domain/track_plan_analysis.go backend/internal/domain/track_plan_analysis_test.go `
  backend/internal/domain/track_plan_clearance.go backend/internal/domain/track_plan_clearance_test.go `
  backend/internal/domain/track_plan_revision_diff.go `
  backend/internal/domain/track_plan_revision_diff_test.go
git commit -m "feat(planner): derive effective flex geometry"
```

---

### Task 3: Flexpfad persistieren und serverseitigen Vorschlag anbieten

**Files:**
- Create: `backend/internal/domain/track_flex_suggestion.go`
- Create: `backend/internal/domain/track_flex_suggestion_test.go`
- Modify: `backend/internal/application/track_planner.go`
- Modify: `backend/internal/application/track_planner_test.go`
- Modify: `backend/internal/infrastructure/track_planner_repository.go`
- Modify: `backend/internal/infrastructure/track_planner_repository_test.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository.go`
- Modify: `backend/internal/infrastructure/layout_revision_repository_test.go`
- Modify: `backend/internal/api/router.go`
- Modify: `backend/internal/api/track_planner_handlers.go`
- Modify: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Produces: `SuggestFlexTrackPath(FlexTrackSuggestionInput) FlexTrackSuggestion`
- Produces: `TrackPlannerService.PreviewFlexPath(context.Context, string, FlexTrackPreviewInput)`
- Produces: `POST /api/v1/plan-track-objects/{id}/flex-preview`
- Extends: Create/Update plan object inputs with `FlexPath *domain.FlexTrackPath`

- [ ] **Step 1: Fehlschlagende Vorschlags-, Persistenz-, Klon- und API-Tests schreiben**

Der Domänentest ruft denselben Vorschlag zweimal auf und verlangt identische Parameter und Punkte.
Ein gerader Zielpunkt liefert 664/0/0 und Tangenten 664/3. Ein 700-mm-Ziel liefert
`LengthExceeded=true` und `Applicable=false`. Ein enger, aber kurzer Verlauf liefert
`RadiusBelowLimit=true`, bleibt `Applicable=true`.

Repositorytests erstellen und aktualisieren einen Flexpfad, laden effektive Felder und klonen ihn in
eine neue Revision. API-Tests prüfen Vorschau, Update, Planner-Rolle, Viewer-Verbot, CSRF und 409 bei
veralteter `expectedVersion`.

- [ ] **Step 2: Gezielte Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run:

```powershell
cd backend
go test ./internal/domain ./internal/application ./internal/infrastructure ./internal/api `
  -run "FlexSuggestion|FlexPath|OpenAPI"
```

Expected: FAIL wegen fehlender Vorschlagsfunktion, Persistenzfelder, Route und OpenAPI-Schemas.

- [ ] **Step 3: Deterministischen Grob-zu-fein-Vorschlag implementieren**

```go
type FlexTrackSuggestionInput struct {
	EndXMM, EndYMM, EndDirectionDegrees float64
	MaximumLengthMM, RadiusLimitMM       float64
}

type FlexTrackSuggestion struct {
	Path             FlexTrackPath          `json:"path"`
	Effective        EffectiveTrackGeometry `json:"-"`
	LengthExceeded   bool                   `json:"lengthExceeded"`
	RadiusBelowLimit bool                   `json:"radiusBelowLimit"`
	Applicable       bool                   `json:"applicable"`
}
```

Das Raster startet mit Tangenten in 10-%-Schritten der Sehnenlänge zwischen 10 % und 150 %, verfeinert
dreimal um den besten Kandidaten und verwendet die in der Spec festgelegte lexikografische Bewertung.
Gleichstände vergleichen Start- und Endtangente numerisch.

- [ ] **Step 4: Repository-Persistenz, Hydrierung und Klonen implementieren**

`CreateObject` und `UpdateObject` serialisieren `flex_path_json`. `trackObjectSelect` und
`scanTrackObject` führen die Spalte direkt nach `elevation_end_mm`; der Scanner hydriert danach die
effektive Geometrie. Starre Geometrien verlangen `NULL`, Flexgeometrien einen gültigen Pfad.

`layout_revision_repository.go` liest und kopiert `flex_path_json` zusammen mit Lineage und Höhen.
`GetPlan` liest `layout.minimum_flex_radius_mm` nach dem Clearance-Limit in
`TrackPlanLimits.MinimumFlexRadiusMM`.

- [ ] **Step 5: Application-Service und HTTP-Handler implementieren**

```go
type FlexTrackPreviewInput struct {
	EndXMM              float64 `json:"endXMm"`
	EndYMM              float64 `json:"endYMm"`
	EndDirectionDegrees float64 `json:"endDirectionDegrees"`
	ExpectedVersion     int     `json:"expectedVersion"`
}

type FlexTrackPreview struct {
	Path                     domain.FlexTrackPath `json:"path"`
	EffectiveGeometry        domain.TrackGeometry `json:"effectiveGeometry"`
	EffectiveLengthMM        float64              `json:"effectiveLengthMm"`
	EffectiveMinimumRadiusMM *float64             `json:"effectiveMinimumRadiusMm,omitempty"`
	RadiusLimitMM            float64              `json:"radiusLimitMm"`
	LengthExceeded           bool                 `json:"lengthExceeded"`
	RadiusBelowLimit         bool                 `json:"radiusBelowLimit"`
	Applicable               bool                 `json:"applicable"`
}
```

Der Service lädt das Objekt samt Anlagenlimits, verlangt `flex`, passende Version und endliche Werte,
ruft die Domäne auf und schreibt nichts. Der POST-Handler verlangt Planner/Admin und CSRF. Das
vorhandene Update bleibt der einzige persistente Schreibpfad.

- [ ] **Step 6: OpenAPI vollständig synchronisieren**

Schemas ergänzen `flex`, `minimumRadiusMm`, Layoutlimit, Flexpfad, effektive Felder,
`FlexTrackPreviewInput`, `FlexTrackPreview`, `flex_radius_below_limit`, `radiusMm` und
`radiusLimitMm`. Der neue Pfad dokumentiert 200, 400, 403, 404 und 409.

- [ ] **Step 7: Backend- und Vertragssuiten ausführen und lokal committen**

Run:

```powershell
cd backend
go test ./internal/domain ./internal/application ./internal/infrastructure ./internal/api
```

Expected: PASS.

```powershell
git add backend/internal/domain/track_flex_suggestion.go `
  backend/internal/domain/track_flex_suggestion_test.go backend/internal/application/track_planner.go `
  backend/internal/application/track_planner_test.go `
  backend/internal/infrastructure/track_planner_repository.go `
  backend/internal/infrastructure/track_planner_repository_test.go `
  backend/internal/infrastructure/layout_revision_repository.go `
  backend/internal/infrastructure/layout_revision_repository_test.go backend/internal/api/router.go `
  backend/internal/api/track_planner_handlers.go backend/internal/api/track_planner_handlers_test.go `
  backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml
git commit -m "feat(api): preview and save flex paths"
```

---

### Task 4: Anlagenoberfläche um den Flexradius erweitern

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/layouts/LayoutFormDialog.tsx`
- Modify: `frontend/src/features/layouts/LayoutFormDialog.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutsView.tsx`
- Modify: `frontend/src/features/layouts/LayoutsView.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutWorkspace.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: API `minimumFlexRadiusMm?: number | null`
- Produces: `LayoutFormValue.minimumFlexRadiusMm: string`

- [ ] **Step 1: Fehlschlagende Formular-, Payload- und Profiltests schreiben**

```tsx
const radius = screen.getByRole("spinbutton", {
  name: "Mindest-Flexgleisradius (mm)"
});
await user.clear(radius);
await user.type(radius, "700");
await user.click(screen.getByRole("button", { name: "Anlage speichern" }));
expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ minimumFlexRadiusMm: "700" }));
```

Ein zweiter Test trägt 500 ein und erwartet eine sichtbare Erklärung, dass Werte unter 543 die
Produktempfehlung nicht senken. Das Payload darf 500 senden; das Profil zeigt als wirksame Basis
weiterhin die Produktgrenze nur am Flexobjekt, nicht als umgeschriebenen Layoutwert.

- [ ] **Step 2: Gezielte Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/LayoutFormDialog.test.tsx `
  src/features/layouts/LayoutsView.test.tsx
```

Expected: FAIL wegen fehlendem Feld, Typ und Text.

- [ ] **Step 3: Typen, app-eigenes Zahlenfeld und Datenfluss implementieren**

`Layout`, `LayoutInput` und `LayoutFormValue` erhalten das neue Feld. Leer wird `null`, sonst
`Number(value)`. Das Formular akzeptiert leer oder endlich größer als 0 und rendert:

```tsx
<AppNumberInput label={t("layouts.field.minimumFlexRadiusMm")}
  value={form.minimumFlexRadiusMm} min="0.1" step="0.1" disabled={saving}
  helpText={t("layouts.field.minimumFlexRadiusMmHelp")}
  error={flexRadiusValid ? undefined : t("layouts.field.minimumFlexRadiusMmError")}
  onValueChange={(value) => setForm((current) => ({ ...current, minimumFlexRadiusMm: value }))} />
```

Das Profil zeigt den gespeicherten Wert mit zwei Nachkommastellen oder „Nicht festgelegt“.

- [ ] **Step 4: Deutsche und englische Texte ergänzen**

Deutsch: „Mindest-Flexgleisradius (mm)“, „Ein höherer Anlagenwert verschärft die
Produktempfehlung. Kleinere Werte senken sie nicht.“, „Bitte einen Wert größer als 0 eingeben.“

Englisch: „Minimum flex-track radius (mm)“, „A higher layout value strengthens the product
recommendation. Lower values do not weaken it.“, „Enter a value greater than 0.“

- [ ] **Step 5: Formtests und Build ausführen und lokal committen**

Run: targeted tests from Step 2, then `npm.cmd run build`.

Expected: PASS.

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts `
  frontend/src/features/layouts/LayoutFormDialog.tsx `
  frontend/src/features/layouts/LayoutFormDialog.test.tsx frontend/src/features/layouts/LayoutsView.tsx `
  frontend/src/features/layouts/LayoutsView.test.tsx `
  frontend/src/features/layouts/LayoutWorkspace.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts
git commit -m "feat(layouts): configure flex track radius"
```

---

### Task 5: App-eigenen Flexeditor und effektive Darstellung bauen

**Files:**
- Create: `frontend/src/features/layouts/flexTrackGeometry.ts`
- Create: `frontend/src/features/layouts/flexTrackGeometry.test.ts`
- Create: `frontend/src/features/layouts/FlexTrackEditorDialog.tsx`
- Create: `frontend/src/features/layouts/FlexTrackEditorDialog.test.tsx`
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/layouts/trackPlannerGeometry.ts`
- Modify: `frontend/src/features/layouts/trackPlannerGeometry.test.ts`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.tsx`
- Modify: `frontend/src/features/layouts/TrackPlannerCanvas.test.tsx`
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx`
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `frontend/src/app/styles.css`

**Interfaces:**
- Consumes: `FlexTrackPath`, `FlexTrackPreview`, effektive Geometrie und Radiuswarnung
- Produces: `sampleFlexPath(path: FlexTrackPath): TrackPoint[]` nur für ungespeicherte lokale Vorschau
- Produces: `FlexTrackEditorDialog`

- [ ] **Step 1: Fehlschlagende API-, Geometrie-, Dialog- und Canvas-Tests schreiben**

Die API-Typen verlangen `flex`, Flexpfad, effektive Felder und `previewFlexTrackPath`.
`flexTrackGeometry.test.ts` prüft exakte Endpunkte und eine deterministische lokale Vorschau.

Der Dialogtest öffnet mit einem Flexobjekt, ändert Endpunkt und Richtung, klickt
„Verlauf vorschlagen“, zeigt 620,00 mm und 700,00 mm Radius, warnt bei Radiusunterschreitung,
deaktiviert bei Überlänge und ruft `onApply(preview.path)` erst nach „Verlauf übernehmen“ auf.

Der Canvastest prüft, dass starre Objekte ihre Bibliotheksroute und Flexobjekte
`effectiveGeometry.routes` rendern, der Bearbeiten-Button den Dialog öffnet und ein übernommener Pfad
im vorhandenen Update-Payload landet.

- [ ] **Step 2: Gezielte Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run:

```powershell
cd frontend
npm.cmd test -- --run src/features/layouts/flexTrackGeometry.test.ts `
  src/features/layouts/FlexTrackEditorDialog.test.tsx `
  src/features/layouts/TrackPlannerCanvas.test.tsx `
  src/features/layouts/TrackPlanAnalysisPanel.test.tsx
```

Expected: FAIL wegen fehlendem Modul, Dialog, Typen und Warncode.

- [ ] **Step 3: Frontend-Verträge und lokalen Vorschauhelfer implementieren**

```ts
export type FlexTrackPath = {
  schemaVersion: 1;
  endXMm: number;
  endYMm: number;
  endDirectionDegrees: number;
  startHandleMm: number;
  endHandleMm: number;
};

export type FlexTrackPreview = {
  path: FlexTrackPath;
  effectiveGeometry: TrackGeometry;
  effectiveLengthMm: number;
  effectiveMinimumRadiusMm?: number;
  radiusLimitMm: number;
  lengthExceeded: boolean;
  radiusBelowLimit: boolean;
  applicable: boolean;
};
```

`sampleFlexPath` berechnet ausschließlich die gestrichelte lokale Vorschau. Persistierte Darstellung,
Länge und Radius kommen immer aus der Serverantwort.

- [ ] **Step 4: App-eigenen Dialog implementieren**

Der Dialog verwendet Portal, Fokusfalle, Escape-Behandlung, `AppNumberInput`, vorhandene Modal-Tokens
und einen eingebetteten SVG-Ausschnitt. Formularänderungen verwerfen das letzte Serverergebnis;
„Verlauf übernehmen“ ist nur bei `preview.applicable` aktiv. Radiuswarnung und Überlänge besitzen
Symbol und Text.

- [ ] **Step 5: Canvas, Snapping und Inspector integrieren**

`TrackPlannerCanvas` rendert `object.effectiveGeometry ?? object.geometry.geometry`. Beim Platzieren
einer Flexgeometrie sendet es den geraden Standardpfad. Update und Höhenänderung erhalten den
bestehenden Flexpfad. Der Inspector zeigt effektive Länge und Radius und öffnet den Dialog.

Der Endpunkt-Handle erscheint nur bei geöffnetem Flexeditor. Seine lokale Bewegung aktualisiert die
ungespeicherte Vorschau; `snapFlexEnd` übernimmt bei höchstens 8 mm Abstand Position und
Gegenrichtung eines kompatiblen Ports. Kein Zielobjekt wird verändert.

- [ ] **Step 6: Radiuswarnung und Übersetzungen ergänzen**

`TrackPlanAnalysisPanel` erhält Symbol `⌒`, separaten Zähler und Detailtext:

- Deutsch: „Flexradius {radius} mm unterschreitet Grenzwert {limit} mm“.
- Englisch: „Flex-track radius {radius} mm is below limit {limit} mm“.

Weitere Texte decken Dialogtitel, Felder, Vorschlag, Übernahme, Länge, Radius, Überlänge,
Produktempfehlung und leere Werte ab.

- [ ] **Step 7: Frontend-Suiten und Build ausführen und lokal committen**

Run: targeted tests from Step 2, then `npm.cmd test -- --run` and `npm.cmd run build`.

Expected: alle Tests und Build PASS.

```powershell
git add frontend/src/features/layouts/flexTrackGeometry.ts `
  frontend/src/features/layouts/flexTrackGeometry.test.ts `
  frontend/src/features/layouts/FlexTrackEditorDialog.tsx `
  frontend/src/features/layouts/FlexTrackEditorDialog.test.tsx `
  frontend/src/shared/apiLayoutsAccessories.ts frontend/src/features/layouts/trackPlannerGeometry.ts `
  frontend/src/features/layouts/trackPlannerGeometry.test.ts `
  frontend/src/features/layouts/TrackPlannerCanvas.tsx `
  frontend/src/features/layouts/TrackPlannerCanvas.test.tsx `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/app/styles.css
git commit -m "feat(planner): edit flex track paths"
```

---

### Task 6: Paket G vollständig abnehmen und dokumentieren

**Files:**
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-flexgleis-vorschlag-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-flexgleis-vorschlag.md`
- Modify: `docs/aegis/work/2026-08-10-advanced-geometry-flex-track/20-checkpoint.md`
- Modify: `docs/aegis/work/2026-08-10-advanced-geometry-flex-track/90-evidence.md`
- Create: `docs/aegis/work/2026-08-10-advanced-geometry-flex-track/99-reflection.md`

**Interfaces:**
- Consumes: vollständiges Paket G
- Produces: lokales Abnahmeprotokoll und sauberen lokalen Branch

- [ ] **Step 1: Vollständige automatisierte Prüfung ausführen**

Run:

```powershell
cd backend
go test ./...

cd ..\frontend
npm.cmd test -- --run
npm.cmd run build
```

Expected: alle Go-Pakete, Vitest-Dateien und der Produktionsbuild bestehen.

- [ ] **Step 2: Lokalen Server mit aktuellem Build neu starten**

Der Server läuft weiter auf `127.0.0.1:18083` mit Repository-Daten, Migrationen, Seeds,
`frontend/dist` und repositorylokalem `GOCACHE`. `/health` muss HTTP 200 liefern; die Sitzung bleibt
als `codex-test` angemeldet.

- [ ] **Step 3: Browserabnahme für gültigen Vorschlag durchführen**

Im Anlagenprofil wird ein Flexradius von 700,00 mm gespeichert. In einem leeren Entwurf wird TILLIG
83125 platziert, der Flexeditor geöffnet und ein kurzer glatter Verlauf vorgeschlagen. Vorschau,
gestrichelter Zustand, Länge, Radius und ausdrückliche Übernahme werden geprüft. Nach Neuladen bleiben
Pfad und Werte erhalten; Stückliste zeigt genau ein Stück 83125.

- [ ] **Step 4: Browserabnahme für Warnung und Überlänge durchführen**

Ein enger Verlauf unter 700 mm erzeugt genau eine anklickbare Radiuswarnung und bleibt bewusst
übernehmbar. Ein Verlauf über 664 mm zeigt Überlänge und deaktiviert die Übernahme. Abbrechen ändert
das gespeicherte Objekt nicht. Browserkonsole bleibt frei von Warnungen und Fehlern.

- [ ] **Step 5: Abnahme, Evidence und Drift aktualisieren**

Die Spec erhält Branch, Commitfolge, Testzahlen, Buildzahl, Server-PID, Browserwerte und „kein Push,
PR oder Merge“. Plancheckboxen werden markiert. Aegis-Checkpoint und Evidence dokumentieren die
frischen Befehle; Reflection bewertet Scope, Kompatibilität und Paket-H-Abgrenzung.

- [ ] **Step 6: Dokumentation lokal committen und Branchzustand prüfen**

```powershell
git add docs/superpowers/specs/2026-08-10-erweiterte-geometrie-flexgleis-vorschlag-design.md `
  docs/superpowers/plans/2026-08-10-erweiterte-geometrie-flexgleis-vorschlag.md `
  docs/aegis/work/2026-08-10-advanced-geometry-flex-track/20-checkpoint.md `
  docs/aegis/work/2026-08-10-advanced-geometry-flex-track/90-evidence.md `
  docs/aegis/work/2026-08-10-advanced-geometry-flex-track/99-reflection.md
git commit -m "docs: record flex track acceptance"
git diff --check
git status --short
```

Expected: Commit erfolgreich, `git diff --check` ohne Ausgabe, Arbeitskopie sauber und weiterhin kein
Push, PR oder Merge.
