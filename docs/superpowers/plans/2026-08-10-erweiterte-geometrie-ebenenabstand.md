# Ebenenabstand an Gleiskreuzungen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Einen optionalen Mindestabstand kreuzender Gleise pro Anlage speichern und dessen
Unterschreitung als präzise, anklickbare Planwarnung anzeigen.

**Architecture:** Die Anlagenkonfiguration trägt den nullable Grenzwert durch SQLite, Backup,
Anwendung, API und Frontend. Eine neue fokussierte Domänendatei analysiert ausschließlich eindeutig
interpolierbare zweipolige Routen, liefert pro Objektpaar den kleinsten Abstand und wird von der
vorhandenen Plananalyse mit den aktuellen Anlagenlimits aufgerufen.

**Tech Stack:** Go, SQLite, React, TypeScript, Vitest, OpenAPI, RailKeeper-i18n und gemeinsame UI-Komponenten

## Global Constraints

- Alles bleibt auf `dev/issue-36-advanced-geometry`; kein Push, PR oder Merge.
- Der Wert beschreibt den vertikalen Schienenoberkantenabstand, keine physische Bauwerksfreigabe.
- Ein leerer Wert deaktiviert die Prüfung; gültige Werte sind endlich und größer als 0.
- Nur Geometrien mit zwei Ports, einer Route und eindeutig zugeordneten Routenenden werden geprüft.
- Echte Schnittpunkte liegen mehr als 0,25 mm von allen Segmentenden entfernt; kollineare Segmente
  bleiben ausschließlich in der vorhandenen Überlappungsprüfung.
- Der Grenzwert ist zulässig; nur eine echte Unterschreitung erzeugt `insufficient_clearance`.
- Mehrere Kreuzungen desselben Objektpaars erzeugen höchstens einen Hinweis für den kleinsten Abstand.
- Hinweise bleiben Warnungen, verändern keine Geometrie und blockieren keine Veröffentlichung.
- Deutsch und Englisch, Backend, Frontend und OpenAPI werden synchron geändert.

---

### Task 1: Anlagenlimit persistieren und sichern

**Files:**
- Create: `backend/migrations/0051_layout_minimum_track_clearance.sql`
- Modify: `backend/internal/application/layouts.go`
- Modify: `backend/internal/application/layouts_test.go`
- Modify: `backend/internal/infrastructure/layout_repository.go`
- Modify: `backend/internal/infrastructure/layout_repository_test.go`
- Modify: `backend/internal/application/backup.go`
- Modify: `backend/internal/application/backup_test.go`

**Interfaces:**
- Produces: `Layout.MinimumTrackClearanceMM *float64`
- Produces: `CreateLayoutInput.MinimumTrackClearanceMM *float64`
- Persists: `layouts.minimum_track_clearance_mm REAL NULL`
- Produces: Backup-Version 10; Versionen bis 9 stellen den Wert als `NULL` wieder her

- [x] **Step 1: Fehlende Validierung und Persistenz als fehlschlagende Tests ergänzen**

In `layouts_test.go` Create und Update mit `MinimumTrackClearanceMM: pointer(40.0)` prüfen. Für die
Servervalidierung dieselbe Tabelle mit `0`, `-1`, `math.NaN()` und `math.Inf(1)` verwenden und jeweils
`ErrValidation` erwarten. In `layout_repository_test.go` den Wert `40` erstellen, als `25.5`
aktualisieren, erneut laden und anschließend mit `nil` entfernen.

```go
clearance := 40.0
layout, err := service.CreateLayout(ctx, application.CreateLayoutInput{
	Name: "Clearance", Kind: domain.LayoutPrivate, Gauge: "TT", Scale: "1:120",
	MinimumTrackClearanceMM: &clearance,
}, "planner")
if err != nil || layout.MinimumTrackClearanceMM == nil || *layout.MinimumTrackClearanceMM != 40 {
	t.Fatalf("minimum clearance not persisted: %#v, %v", layout, err)
}
```

- [x] **Step 2: Gezielte Tests ausführen und den erwarteten Compile-Fehler bestätigen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "Layout.*Clearance|Clearance.*Layout"`

Expected: FAIL, `MinimumTrackClearanceMM` ist noch nicht definiert.

- [x] **Step 3: Migration, Modell, Validierung und Repository implementieren**

```sql
ALTER TABLE layouts ADD COLUMN minimum_track_clearance_mm REAL
    CHECK(minimum_track_clearance_mm IS NULL OR minimum_track_clearance_mm > 0);
```

`Layout` und `CreateLayoutInput` erhalten `MinimumTrackClearanceMM *float64` mit dem JSON-Namen
`minimumTrackClearanceMm`. `validLayoutInput` akzeptiert `nil` oder `finite(value) && value > 0`.
Insert, Update, `layoutSelect` und `scanLayout` führen die Spalte direkt nach `max_grade_percent`;
der Scan verwendet ein eigenes `sql.NullFloat64`.

- [x] **Step 4: Backup-Version und Legacy-Roundtrip ergänzen**

`backupVersion` wird 10, der Future-Version-Test erwartet 11. Ein Test exportiert und restauriert
eine Anlage mit 40 mm. Derselbe Export wird auf Version 9 zurückgestuft und
`minimum_track_clearance_mm` aus jeder Layoutzeile entfernt; nach Restore muss die Spalte `NULL`
enthalten.

- [x] **Step 5: Application-, Repository- und Backuptests ausführen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure`

Expected: PASS.

- [x] **Step 6: Persistenz lokal committen**

```powershell
git add backend/migrations/0051_layout_minimum_track_clearance.sql `
  backend/internal/application/layouts.go backend/internal/application/layouts_test.go `
  backend/internal/infrastructure/layout_repository.go `
  backend/internal/infrastructure/layout_repository_test.go backend/internal/application/backup.go `
  backend/internal/application/backup_test.go
git commit -m "feat(layouts): configure track clearance"
```

---

### Task 2: Schnittpunkte und Ebenenabstände analysieren

**Files:**
- Create: `backend/internal/domain/track_plan_clearance.go`
- Create: `backend/internal/domain/track_plan_clearance_test.go`
- Modify: `backend/internal/domain/track_plan_analysis.go`

**Interfaces:**
- Consumes: `PlanTrackObject`, `TransformTrackPoint`, `TrackConnectionDistanceMM`
- Produces: `TrackPlanLimits.MinimumTrackClearanceMM *float64`
- Produces: `TrackPlanIssueInsufficientClearance = "insufficient_clearance"`
- Produces: `analyzeTrackClearances(objects []PlanTrackObject, limit float64) []TrackPlanIssue`

- [x] **Step 1: Fehlschlagende Domänentests für Kreuzung, Interpolation und Grenzen schreiben**

Die Testfixture kreuzt ein waagerechtes G1 in dessen Mitte mit einem um 90 Grad gedrehten G1:

```go
lower := testG1Object("lower", 0, 0, 0)
upper := testG1Object("upper", 83, -83, 90)
lower.ElevationStartMM, lower.ElevationEndMM = 0, 20
upper.ElevationStartMM, upper.ElevationEndMM = 45, 45
limit := 40.0

issues := filterTrackIssues(AnalyzeTrackPlanWithLimits(
	[]PlanTrackObject{upper, lower},
	TrackPlanLimits{MinimumTrackClearanceMM: &limit},
).Issues, TrackPlanIssueInsufficientClearance)
if len(issues) != 1 || issues[0].ClearanceMM == nil ||
	math.Abs(*issues[0].ClearanceMM-35) > 1e-9 ||
	issues[0].ClearanceLimitMM == nil || *issues[0].ClearanceLimitMM != 40 ||
	issues[0].IntersectionXMM == nil || math.Abs(*issues[0].IntersectionXMM-83) > 1e-9 ||
	issues[0].IntersectionYMM == nil || math.Abs(*issues[0].IntersectionYMM) > 1e-9 {
	t.Fatalf("unexpected clearance issue: %#v", issues)
}
```

Weitere Tests prüfen: Abstand 40 ohne Hinweis, kein Limit ohne Hinweis, Schnitt am gemeinsamen
Endpunkt ohne Hinweis, kollineare Überlappung ohne Clearance-Hinweis, umgekehrte Route mit korrekter
Höhe, Mehrport- und Mehrfachroutengeometrien ohne Hinweis sowie zwei Kreuzungen desselben Objektpaars
mit genau einem Hinweis für das kleinere Ergebnis.

- [x] **Step 2: Domänentests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/domain -run Clearance`

Expected: FAIL wegen fehlendem Limit, Warncode und Detailfeldern.

- [x] **Step 3: Fokussierte Clearance-Analyse implementieren**

`track_plan_clearance.go` enthält nur die neue Geometrieauswertung. Die Kernstrukturen lauten:

```go
type clearanceRoute struct {
	ObjectID string
	Points []TrackPoint
	CumulativeLengths []float64
	TotalLength float64
	ElevationStartMM float64
	ElevationEndMM float64
}

type clearanceCandidate struct {
	ObjectIDs []string
	ClearanceMM float64
	Intersection TrackPoint
}

func analyzeTrackClearances(objects []PlanTrackObject, limit float64) []TrackPlanIssue
func clearanceRouteForObject(object PlanTrackObject) (clearanceRoute, bool)
func properTrackSegmentIntersection(first, second trackSegment) (TrackPoint, float64, float64, bool)
```

`clearanceRouteForObject` verlangt exakt zwei Ports und eine Route. Es vergleicht Routenanfang und
-ende mit den Ports über `trackPointDistance <= TrackConnectionDistanceMM`, transformiert alle
Routenpunkte und führt kumulierte Segmentlängen. Bei umgekehrter Zuordnung tauscht es die beiden
Höhen. `properTrackSegmentIntersection` verwendet Kreuzprodukte, verwirft parallele/kollineare
Segmente und akzeptiert nur Parameter, deren Abstand zu allen Segmentenden größer als 0,25 mm ist.

Für jeden Schnittpunkt wird der kumulierte Weganteil beider Routen berechnet und die Höhe linear
interpoliert. Pro sortiertem Objektpaar bleibt der Kandidat mit kleinstem Abstand; bei Gleichstand
gewinnen kleineres X und danach kleineres Y. Nur
`candidate.ClearanceMM+1e-9 < limit` erzeugt einen Warnhinweis mit Kopien aller Detailwerte.

- [x] **Step 4: Analyse in den vorhandenen Ablauf einhängen**

`TrackPlanLimits` erhält `MinimumTrackClearanceMM *float64`. `TrackPlanIssue` erhält
`ClearanceMM`, `ClearanceLimitMM`, `IntersectionXMM` und `IntersectionYMM` als optionale Pointer.
Nach Verbindungs-, Höhen- und Überlappungsprüfung hängt `AnalyzeTrackPlanWithLimits` bei gesetztem
Limit `analyzeTrackClearances(ordered, *limits.MinimumTrackClearanceMM)` an. Die bestehende stabile
Sortierung läuft anschließend über alle Hinweise.

- [x] **Step 5: Domain-Suite ausführen**

Run: `cd backend; go test ./internal/domain`

Expected: PASS.

- [x] **Step 6: Analyse lokal committen**

```powershell
git add backend/internal/domain/track_plan_clearance.go `
  backend/internal/domain/track_plan_clearance_test.go `
  backend/internal/domain/track_plan_analysis.go
git commit -m "feat(planner): detect insufficient track clearance"
```

---

### Task 3: Anlagenlimit zum Plan und in den API-Vertrag führen

**Files:**
- Modify: `backend/internal/infrastructure/track_planner_repository.go`
- Modify: `backend/internal/infrastructure/track_planner_repository_test.go`
- Modify: `backend/internal/application/track_planner_test.go`
- Modify: `backend/internal/api/layout_handlers_test.go`
- Modify: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Consumes: `Layout.MinimumTrackClearanceMM`, `TrackPlanLimits.MinimumTrackClearanceMM`
- Produces: JSON `minimumTrackClearanceMm`, `clearanceMm`, `clearanceLimitMm`, `intersectionXMm`,
  `intersectionYMm`

- [x] **Step 1: Repository-, Service- und Handlertests fehlschlagend erweitern**

Der Repositorytest erstellt eine Anlage mit 40 mm und prüft, dass `GetPlan` und ein geklonter Plan
`Limits.MinimumTrackClearanceMM == 40` liefern. Der Servicetest verwendet zwei kreuzende G1 und prüft
einen Hinweis in Analyse sowie einen in `Added` und später `Resolved` in der Änderungsvorschau.
Layout-Handler testen Create 40 und Update 25.5; Track-Plan-Handler prüfen Code und vier Detailfelder.

- [x] **Step 2: API- und Repositorytests ausführen und das erwartete Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure ./internal/api -run "Clearance|OpenAPI"`

Expected: FAIL, Repositoryabfrage und OpenAPI führen das neue Limit noch nicht.

- [x] **Step 3: Plan-Repository um das zweite Limit ergänzen**

Die `GetPlan`-Abfrage liest nach `layout.max_grade_percent` zusätzlich
`layout.minimum_track_clearance_mm`. Ein zweites `sql.NullFloat64` wird bei `Valid` nach
`plan.Limits.MinimumTrackClearanceMM` kopiert. Klonen und Änderungsvorschau verwenden bereits das
komplette `TrackPlanLimits`-Struct und benötigen keine parallele Sonderlogik.

- [x] **Step 4: OpenAPI synchron erweitern**

`Layout` und `LayoutInput` erhalten `minimumTrackClearanceMm` als nullable number mit
`exclusiveMinimum: 0`. `TrackPlanIssue.code` und `TrackPlanIssueChange.code` erhalten
`insufficient_clearance`. `TrackPlanIssue` erhält die vier nullable number-Felder; nur
`clearanceLimitMm` besitzt `exclusiveMinimum: 0`.

- [x] **Step 5: Backend- und OpenAPI-Suiten ausführen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure ./internal/api`

Expected: PASS.

- [x] **Step 6: Vertrag lokal committen**

```powershell
git add backend/internal/infrastructure/track_planner_repository.go `
  backend/internal/infrastructure/track_planner_repository_test.go `
  backend/internal/application/track_planner_test.go backend/internal/api/layout_handlers_test.go `
  backend/internal/api/track_planner_handlers_test.go backend/internal/api/openapi_contract_test.go `
  openapi/railkeeper.yaml
git commit -m "feat(api): expose track clearance warnings"
```

---

### Task 4: Anlagenoberfläche um den Mindestabstand erweitern

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
- Consumes: API `minimumTrackClearanceMm?: number | null`
- Produces: `LayoutFormValue.minimumTrackClearanceMm: string`

- [ ] **Step 1: Fehlschlagende Form- und Profiltests schreiben**

```tsx
const clearance = screen.getByRole("spinbutton", {
  name: "Mindestabstand kreuzender Gleise (mm)"
});
expect(clearance.closest(".app-number-input")).not.toBeNull();
await user.type(clearance, "40");
await user.click(screen.getByRole("button", { name: "Anlage speichern" }));
expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ minimumTrackClearanceMm: "40" }));
```

Ein zweiter Test trägt `0` ein und erwartet einen deaktivierten Speichern-Button. `LayoutsView` prüft
Create- und Update-Payload sowie die Profilanzeige `40,00 mm`; Leeren sendet `null` und zeigt
„Nicht festgelegt“.

- [ ] **Step 2: Gezielte Tests ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutFormDialog.test.tsx src/features/layouts/LayoutsView.test.tsx`

Expected: FAIL wegen fehlendem Typ, Feld und Text.

- [ ] **Step 3: Typen, app-eigenes Feld und Datenfluss implementieren**

`Layout` und `LayoutInput` erhalten `minimumTrackClearanceMm?: number | null`.
`LayoutFormValue` erhält `minimumTrackClearanceMm: string`. `LayoutFormDialog` trimmt den Wert,
akzeptiert leer oder `Number.isFinite(value) && value > 0` und rendert `AppNumberInput` mit
`min="0.1"` sowie `step="0.1"`. Create und Update senden leer als `null`, sonst `Number(value)`.
`LayoutWorkspace` formatiert den Wert mit zwei Nachkommastellen und `mm`.

- [ ] **Step 4: Deutsche und englische Texte ergänzen**

Deutsch: „Mindestabstand kreuzender Gleise (mm)“, „Gemessen zwischen Schienenoberkanten. Leer lassen,
um die Warnung zu deaktivieren.“, „Bitte einen Wert größer als 0 eingeben.“

Englisch: „Minimum crossing track separation (mm)“, „Measured between rail tops. Leave empty to
disable the warning.“, „Enter a value greater than 0.“

- [ ] **Step 5: Formtests und Produktionsbuild ausführen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutFormDialog.test.tsx src/features/layouts/LayoutsView.test.tsx`

Run: `cd frontend; npm.cmd run build`

Expected: PASS.

- [ ] **Step 6: Anlagenoberfläche lokal committen**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts frontend/src/features/layouts/LayoutFormDialog.tsx `
  frontend/src/features/layouts/LayoutFormDialog.test.tsx frontend/src/features/layouts/LayoutsView.tsx `
  frontend/src/features/layouts/LayoutsView.test.tsx frontend/src/features/layouts/LayoutWorkspace.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat(layouts): edit track clearance limit"
```

---

### Task 5: Warnhinweis darstellen und Paket vollständig abnehmen

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx`
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-ebenenabstand-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-ebenenabstand.md`

**Interfaces:**
- Consumes: `insufficient_clearance` und die vier numerischen Detailfelder
- Produces: lokalisierter Zähler, Detailtext und vorhandenes `onSelectObject`-Fokusverhalten

- [ ] **Step 1: Fehlschlagenden Paneltest für Zähler, Detail und Fokus schreiben**

```tsx
const issue = {
  code: "insufficient_clearance" as const,
  severity: "warning" as const,
  objectIds: ["lower", "upper"],
  clearanceMm: 25,
  clearanceLimitMm: 40,
  intersectionXMm: 83,
  intersectionYMm: 0
};

expect(screen.getByText("1 Abstandsunterschreitung")).toBeInTheDocument();
const warning = screen.getByRole("button", {
  name: "Warnung: Ebenenabstand 25,00 mm unterschreitet Grenzwert 40,00 mm"
});
await user.click(warning);
expect(selectObject).toHaveBeenCalledWith("lower");
```

- [ ] **Step 2: Paneltest ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/TrackPlanAnalysisPanel.test.tsx`

Expected: FAIL wegen fehlendem Code, Zähler und Detailtext.

- [ ] **Step 3: Typ, Zähler, Detailtext und Übersetzungen implementieren**

Der TypeScript-Union und `TrackPlanIssue` erhalten Code und Felder. `issueSymbols` erhält `↕`.
`TrackPlanAnalysisPanel` zählt den Code separat und formatiert Abstand sowie Grenze mit zwei
Nachkommastellen. Deutsch verwendet „{count} Abstandsunterschreitung(en)“ und „Ebenenabstand
{clearance} mm unterschreitet Grenzwert {limit} mm“. Englisch verwendet „{count} clearance
violation(s)“ und „Track separation {clearance} mm is below limit {limit} mm“.

- [ ] **Step 4: Vollständige automatisierte Prüfung ausführen**

Run: `cd backend; go test ./...`

Run: `cd frontend; npm.cmd test -- --run`

Run: `cd frontend; npm.cmd run build`

Expected: Alle Go-Pakete, Vitest-Dateien und der Produktionsbuild bestehen.

- [ ] **Step 5: Lokalen Server neu starten und Browserabnahme durchführen**

Die Anlage erhält 40,00 mm Mindestabstand. Zwei G1 werden rechtwinklig mittig gekreuzt; das untere
liegt bei 0 mm, das obere bei 25 mm. Die Planprüfung muss genau eine Abstandsunterschreitung anzeigen,
der Detailtext muss 25,00/40,00 mm nennen und der Klick das erste G1 fokussieren. Nach Anheben des
oberen Gleises auf 40 mm verschwindet der Hinweis und bleibt nach Neuladen entfernt. `/health`
antwortet 200, die Sitzung bleibt als `codex-test` angemeldet und die Browserkonsole fehlerfrei.

- [ ] **Step 6: Abnahme dokumentieren und lokal committen**

Die Spec erhält Branch, Testzahlen, Buildzahl, Browserwerte, Server-PID und die Bestätigung „kein
Push, PR oder Merge“. Dieser Plan erhält den Status „lokal vollständig umgesetzt“ und alle erledigten
Checkboxen werden markiert.

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts `
  docs/superpowers/specs/2026-08-10-erweiterte-geometrie-ebenenabstand-design.md `
  docs/superpowers/plans/2026-08-10-erweiterte-geometrie-ebenenabstand.md
git commit -m "feat(planner): show track clearance warnings"
```
