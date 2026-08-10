# Konfigurierbare Steigungsgrenzen Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Eine optionale maximale Steigung pro Anlage speichern und Überschreitungen im Gleisplan als
präzise, anklickbare Warnung anzeigen.

**Architecture:** Das Layout trägt den nullable Grenzwert durch SQLite, Anwendung, API und Frontend.
Der Track-Plan-Repository-Leseweg liefert die aktuelle Anlagenkonfiguration als nicht serialisierte
Analyseoption. Die Domänenanalyse bleibt ohne Option rückwärtskompatibel und ergänzt bei gesetzter
Grenze den Warncode `grade_limit_exceeded`.

**Tech Stack:** Go, SQLite, React, TypeScript, Vitest, OpenAPI, RailKeeper-i18n und gemeinsame UI-Komponenten

## Global Constraints

- Zulässig sind endliche Werte größer als 0 und höchstens 100 Prozent.
- Ein leerer Wert deaktiviert ausschließlich die Grenzwert-Warnung.
- `abs(gradePercent) > maxGradePercent` warnt; Gleichheit ist zulässig.
- Die Warnung korrigiert nichts automatisch und blockiert keine Veröffentlichung.
- Der Wert gilt pro Anlage, nicht pro Einheit, Gleis oder Revision.
- Backup-Version 9 muss Backups bis Version 8 ohne erfundenen Standardwert wiederherstellen.
- Deutsch und Englisch werden gemeinsam gepflegt.
- Alle Änderungen bleiben lokal, ohne Push, PR oder Merge.

---

### Task 1: Layout-Persistenz, Validierung und Backup

**Files:**
- Create: `backend/migrations/0050_layout_max_grade_percent.sql`
- Modify: `backend/internal/application/layouts.go`
- Test: `backend/internal/application/layouts_test.go`
- Modify: `backend/internal/infrastructure/layout_repository.go`
- Test: `backend/internal/infrastructure/layout_repository_test.go`
- Modify: `backend/internal/application/backup.go`
- Test: `backend/internal/application/backup_test.go`

**Interfaces:**
- Produces: `Layout.MaxGradePercent *float64`, `CreateLayoutInput.MaxGradePercent *float64`
- Persists: `layouts.max_grade_percent REAL NULL`

- [ ] **Step 1: Fehlende Layout-Validierung und Persistenz als fehlschlagende Tests ergänzen**

```go
limit := 3.5
created, err := service.CreateLayout(ctx, CreateLayoutInput{
	Name: "Steigungsanlage", Kind: domain.LayoutPrivate, Gauge: "TT", Scale: "1:120",
	MaxGradePercent: &limit,
}, "planner-1")
if err != nil || created.MaxGradePercent == nil || *created.MaxGradePercent != limit {
	t.Fatalf("grade limit not persisted: layout=%#v err=%v", created, err)
}

for _, invalid := range []float64{0, -1, 100.1, math.NaN(), math.Inf(1)} {
	input := CreateLayoutInput{Name: "Ungültig", Kind: domain.LayoutPrivate, Gauge: "TT", Scale: "1:120",
		MaxGradePercent: &invalid}
	if _, err := service.CreateLayout(ctx, input, "planner-1"); !errors.Is(err, ErrLayoutValidation) {
		t.Fatalf("expected validation error for %v, got %v", invalid, err)
	}
}
```

- [ ] **Step 2: Gezielte Layouttests ausführen und den erwarteten Compile-Fehler bestätigen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure -run "Layout.*Grade|Grade.*Layout"`

Expected: FAIL, `MaxGradePercent` ist noch nicht definiert.

- [ ] **Step 3: Migration, Modell, Validierung und Repository implementieren**

```sql
ALTER TABLE layouts ADD COLUMN max_grade_percent REAL
    CHECK(max_grade_percent IS NULL OR (max_grade_percent > 0 AND max_grade_percent <= 100));
```

```go
type Layout struct {
	// bestehende Felder
	MaxGradePercent *float64 `json:"maxGradePercent,omitempty"`
}

type CreateLayoutInput struct {
	// bestehende Felder
	MaxGradePercent *float64 `json:"maxGradePercent"`
}

func validLayoutInput(input CreateLayoutInput) bool {
	return input.Name != "" && input.Gauge != "" && input.Scale != "" && input.Kind.Valid() &&
		(input.MaxGradePercent == nil || finite(*input.MaxGradePercent) &&
			*input.MaxGradePercent > 0 && *input.MaxGradePercent <= 100)
}
```

`layoutSelect`, `scanLayout`, `INSERT` und `UPDATE` führen `max_grade_percent` an derselben Position
zwischen `description` und `version`. Nullable Werte werden direkt über `*float64` an `database/sql`
übergeben und in ein lokales `sql.NullFloat64` gescannt.

- [ ] **Step 4: Backup-Version und Legacy-Abnahme ergänzen**

`backupVersion` wird 9. Der vorhandene Future-Version-Test erwartet Version 10. Ein neuer Test
exportiert eine Anlage mit 3,5 Prozent, stellt sie wieder her und prüft den Wert. Derselbe Export wird
auf Version 8 zurückgestuft, `max_grade_percent` aus den Layoutzeilen entfernt und muss nach Restore
`NULL` liefern.

- [ ] **Step 5: Layout-, Infrastruktur- und Backuptests ausführen**

Run: `cd backend; go test ./internal/application ./internal/infrastructure`

Expected: PASS.

- [ ] **Step 6: Persistenz lokal committen**

```powershell
git add backend/migrations/0050_layout_max_grade_percent.sql backend/internal/application/layouts.go `
  backend/internal/application/layouts_test.go backend/internal/infrastructure/layout_repository.go `
  backend/internal/infrastructure/layout_repository_test.go backend/internal/application/backup.go `
  backend/internal/application/backup_test.go
git commit -m "feat(layouts): configure maximum grade"
```

### Task 2: Domänenanalyse und Plan-Kontext

**Files:**
- Modify: `backend/internal/domain/track_plan_analysis.go`
- Test: `backend/internal/domain/track_plan_analysis_test.go`
- Modify: `backend/internal/application/track_planner.go`
- Test: `backend/internal/application/track_planner_test.go`
- Modify: `backend/internal/infrastructure/track_planner_repository.go`
- Test: `backend/internal/infrastructure/track_planner_repository_test.go`

**Interfaces:**
- Consumes: `layouts.max_grade_percent`
- Produces: `TrackPlanLimits{MaxGradePercent *float64}` und `grade_limit_exceeded`

- [ ] **Step 1: Grenzfalltests für Aufstieg, Abstieg, Gleichheit und fehlende Grenze schreiben**

```go
func TestAnalyzeTrackPlanWithLimitsReportsAbsoluteGradeExcess(t *testing.T) {
	ascending := testG1Object("ascending", 0, 0, 0)
	ascending.ElevationEndMM = 6.64 // 4.00 % bei 166 mm
	descending := testG1Object("descending", 300, 0, 0)
	descending.ElevationStartMM, descending.ElevationEndMM = 10, 3.36
	limit := 3.0
	issues := AnalyzeTrackPlanWithLimits([]PlanTrackObject{ascending, descending},
		TrackPlanLimits{MaxGradePercent: &limit}).Issues
	if got := filterTrackIssues(issues, TrackPlanIssueGradeLimitExceeded); len(got) != 2 {
		t.Fatalf("expected two grade warnings, got %#v", got)
	}
}

func TestAnalyzeTrackPlanWithLimitsAllowsBoundaryAndUnsetLimit(t *testing.T) {
	track := testG1Object("track", 0, 0, 0)
	track.ElevationEndMM = 4.98 // 3.00 % bei 166 mm
	limit := 3.0
	if issues := filterTrackIssues(AnalyzeTrackPlanWithLimits([]PlanTrackObject{track},
		TrackPlanLimits{MaxGradePercent: &limit}).Issues, TrackPlanIssueGradeLimitExceeded); len(issues) != 0 {
		t.Fatalf("boundary produced warning: %#v", issues)
	}
	if issues := filterTrackIssues(AnalyzeTrackPlan([]PlanTrackObject{track}).Issues,
		TrackPlanIssueGradeLimitExceeded); len(issues) != 0 {
		t.Fatalf("unset limit produced warning: %#v", issues)
	}
}
```

- [ ] **Step 2: Domänentest ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd backend; go test ./internal/domain -run GradeLimit`

Expected: FAIL wegen fehlender Limits und Warncodes.

- [ ] **Step 3: Optionsfähige Domänenanalyse minimal implementieren**

```go
const TrackGradeLimitTolerancePercent = 1e-9

type TrackPlanLimits struct { MaxGradePercent *float64 }

const TrackPlanIssueGradeLimitExceeded TrackPlanIssueCode = "grade_limit_exceeded"

type TrackPlanIssue struct {
	// bestehende Felder
	GradePercent      *float64 `json:"gradePercent,omitempty"`
	GradeLimitPercent *float64 `json:"gradeLimitPercent,omitempty"`
}

func AnalyzeTrackPlan(objects []PlanTrackObject) TrackPlanAnalysis {
	return AnalyzeTrackPlanWithLimits(objects, TrackPlanLimits{})
}

func AnalyzeTrackPlanWithLimits(objects []PlanTrackObject, limits TrackPlanLimits) TrackPlanAnalysis {
	// vorhandene Analyse; nach der Grade-Berechnung optional Warnung anhängen
}
```

Die berechnete Steigung wird in eine lokale Variable gelegt. Wenn die Grenze gesetzt ist und
`math.Abs(grade)-*limits.MaxGradePercent > TrackGradeLimitTolerancePercent`, wird ein Warning-Issue
mit einer Objekt-ID sowie Kopien von Steigung und Grenze angehängt.

- [ ] **Step 4: Anlagenlimit mit jedem Plan laden und in Analyse sowie Preview verwenden**

`TrackPlan` erhält `Limits domain.TrackPlanLimits \`json:"-"\``. `GetPlan` verbindet Revision,
Variante, Einheit und Anlage und scannt `layout.max_grade_percent` als `sql.NullFloat64`. `AnalyzePlan`
ruft `AnalyzeTrackPlanWithLimits(plan.Objects, plan.Limits)` auf. `ChangePreview` verwendet
`current.Limits` für Basis- und Arbeitsrevision, auch wenn noch keine Basisrevision existiert.

- [ ] **Step 5: Domain-, Application- und Repositorytests ausführen**

Run: `cd backend; go test ./internal/domain ./internal/application ./internal/infrastructure -run "TrackPlan|GradeLimit"`

Expected: PASS.

- [ ] **Step 6: Analyse lokal committen**

```powershell
git add backend/internal/domain/track_plan_analysis.go backend/internal/domain/track_plan_analysis_test.go `
  backend/internal/application/track_planner.go backend/internal/application/track_planner_test.go `
  backend/internal/infrastructure/track_planner_repository.go `
  backend/internal/infrastructure/track_planner_repository_test.go
git commit -m "feat(planner): warn on excessive grades"
```

### Task 3: API- und OpenAPI-Vertrag

**Files:**
- Test: `backend/internal/api/layout_handlers_test.go`
- Test: `backend/internal/api/track_planner_handlers_test.go`
- Modify: `backend/internal/api/openapi_contract_test.go`
- Modify: `openapi/railkeeper.yaml`

**Interfaces:**
- Produces: JSON `maxGradePercent`, `gradePercent`, `gradeLimitPercent`
- Extends: `TrackPlanIssue.code` und `TrackPlanIssueChange.code` um `grade_limit_exceeded`

- [ ] **Step 1: Handler-Vertrag mit 3-Prozent-Anlage und 4-Prozent-Gleis testen**

Der Layout-Handler-Test sendet `"maxGradePercent":3.0` bei Create und Update und prüft die Antwort.
Der Track-Plan-Handler-Test erzeugt ein 166-mm-Gleis mit 6,64 mm Höhenänderung und erwartet:

```go
if issue.Code != domain.TrackPlanIssueGradeLimitExceeded || issue.GradePercent == nil ||
	issue.GradeLimitPercent == nil || math.Abs(*issue.GradePercent-4) > 1e-9 ||
	*issue.GradeLimitPercent != 3 {
	t.Fatalf("unexpected grade limit issue: %#v", issue)
}
```

- [ ] **Step 2: API-Tests ausführen und den fehlenden OpenAPI-Vertrag bestätigen**

Run: `cd backend; go test ./internal/api -run "Layout|TrackPlanner|OpenAPI"`

Expected: FAIL, die neuen Felder und Enumwerte fehlen im Schema.

- [ ] **Step 3: OpenAPI-Schemas synchron erweitern**

`Layout` und `LayoutInput` erhalten `maxGradePercent` als nullable number mit `exclusiveMinimum: 0`
und `maximum: 100`. `TrackPlanIssue` erhält den Enumwert sowie nullable `gradePercent` und
`gradeLimitPercent`; letzterer hat `exclusiveMinimum: 0` und `maximum: 100`.
`TrackPlanIssueChange.code` erhält denselben Enumwert.

- [ ] **Step 4: API- und OpenAPI-Tests ausführen**

Run: `cd backend; go test ./internal/api`

Expected: PASS.

- [ ] **Step 5: API-Vertrag lokal committen**

```powershell
git add backend/internal/api/layout_handlers_test.go backend/internal/api/track_planner_handlers_test.go `
  backend/internal/api/openapi_contract_test.go openapi/railkeeper.yaml
git commit -m "feat(api): expose grade limit warnings"
```

### Task 4: Anlagenformular und Profil

**Files:**
- Modify: `frontend/src/shared/apiLayoutsAccessories.ts`
- Modify: `frontend/src/features/layouts/LayoutFormDialog.tsx`
- Test: `frontend/src/features/layouts/LayoutFormDialog.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutsView.tsx`
- Test: `frontend/src/features/layouts/LayoutsView.test.tsx`
- Modify: `frontend/src/features/layouts/LayoutWorkspace.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`

**Interfaces:**
- Consumes: `Layout.maxGradePercent?: number`, `LayoutInput.maxGradePercent?: number | null`
- Produces: `LayoutFormValue.maxGradePercent: string`

- [ ] **Step 1: Fehlschlagende Formulartests für Setzen, Leeren und Validieren schreiben**

```tsx
const gradeInput = screen.getByRole("spinbutton", { name: "Maximale Steigung (%)" });
await user.clear(gradeInput);
await user.type(gradeInput, "3.5");
await user.click(screen.getByRole("button", { name: "Anlage speichern" }));
expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ maxGradePercent: "3.5" }));

await user.clear(gradeInput);
await user.type(gradeInput, "101");
expect(screen.getByRole("button", { name: "Anlage speichern" })).toBeDisabled();
```

- [ ] **Step 2: Gezielte Frontendtests ausführen und Compile-/Assertionfehler bestätigen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutFormDialog.test.tsx src/features/layouts/LayoutsView.test.tsx`

Expected: FAIL wegen fehlendem Feld und fehlender Beschriftung.

- [ ] **Step 3: Frontend-Typen und app-eigenes Zahlenfeld implementieren**

`Layout` und `LayoutInput` erhalten `maxGradePercent?: number | null`. `LayoutFormValue` erhält
`maxGradePercent: string`. `LayoutFormDialog` rendert:

```tsx
<AppNumberInput label={t("layouts.field.maxGradePercent")} value={form.maxGradePercent}
  min="0.1" max="100" step="0.1" disabled={saving}
  helpText={t("layouts.field.maxGradePercentHelp")}
  error={gradeLimitValid ? undefined : t("layouts.field.maxGradePercentError")}
  onValueChange={(value) => setForm((current) => ({ ...current, maxGradePercent: value }))} />
```

`gradeLimitValid` ist bei leerem String wahr, sonst nur bei endlicher Zahl `> 0 && <= 100`.
Create und Update senden bei leerem Feld `null`, sonst `Number(form.maxGradePercent)`. Das Profil
zeigt `Intl.NumberFormat` mit zwei Nachkommastellen und `%` oder den vorhandenen Nicht-festgelegt-Text.

- [ ] **Step 4: Deutsche und englische Texte ergänzen**

Deutsch: „Maximale Steigung (%)“, „Leer lassen, um die Warnung zu deaktivieren.“,
„Bitte einen Wert über 0 bis einschließlich 100 eingeben.“
Englisch: „Maximum grade (%)“, „Leave empty to disable the warning.“,
„Enter a value greater than 0 and up to 100.“

- [ ] **Step 5: Formulartests und Build ausführen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/LayoutFormDialog.test.tsx src/features/layouts/LayoutsView.test.tsx`

Run: `cd frontend; npm.cmd run build`

Expected: PASS.

- [ ] **Step 6: Anlagenoberfläche lokal committen**

```powershell
git add frontend/src/shared/apiLayoutsAccessories.ts frontend/src/features/layouts/LayoutFormDialog.tsx `
  frontend/src/features/layouts/LayoutFormDialog.test.tsx frontend/src/features/layouts/LayoutsView.tsx `
  frontend/src/features/layouts/LayoutsView.test.tsx frontend/src/features/layouts/LayoutWorkspace.tsx `
  frontend/src/shared/i18n/de.ts frontend/src/shared/i18n/en.ts
git commit -m "feat(layouts): edit maximum grade limit"
```

### Task 5: Planerwarnung, Gesamtprüfung und Abnahme

**Files:**
- Modify: `frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx`
- Test: `frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx`
- Modify: `frontend/src/shared/i18n/de.ts`
- Modify: `frontend/src/shared/i18n/en.ts`
- Modify: `docs/superpowers/specs/2026-08-10-erweiterte-geometrie-steigungsgrenzen-design.md`
- Modify: `docs/superpowers/plans/2026-08-10-erweiterte-geometrie-steigungsgrenzen.md`

**Interfaces:**
- Consumes: `grade_limit_exceeded`, `gradePercent`, `gradeLimitPercent`
- Produces: lokalisierter Zähler, Detailtext und vorhandenes `onSelectObject`-Fokusverhalten

- [ ] **Step 1: UI-Test für Zähler, Detail und Fokus schreiben**

```tsx
const issue = {
  code: "grade_limit_exceeded" as const,
  severity: "warning" as const,
  objectIds: ["track-2"],
  gradePercent: -5.51,
  gradeLimitPercent: 3
};
expect(screen.getByText("1 Steigungsüberschreitung")).toBeInTheDocument();
const warning = screen.getByRole("button", {
  name: "Warnung: Steigung -5,51 % überschreitet Grenzwert 3,00 %"
});
await user.click(warning);
expect(selectObject).toHaveBeenCalledWith("track-2");
```

- [ ] **Step 2: Paneltest ausführen und erwartetes Fehlschlagen bestätigen**

Run: `cd frontend; npm.cmd test -- --run src/features/layouts/TrackPlanAnalysisPanel.test.tsx`

Expected: FAIL wegen fehlendem Code, Zähler und Detailtext.

- [ ] **Step 3: Zähler, Detail und Übersetzungen implementieren**

`issueSymbols` erhält `grade_limit_exceeded`. Der Header zählt den Code. Die Ausgabe formatiert beide
Werte mit zwei Nachkommastellen und verwendet:

```text
layouts.trackAnalysis.gradeLimitExceededOne
layouts.trackAnalysis.gradeLimitExceededMany
layouts.trackAnalysis.issue.grade_limit_exceeded
layouts.trackAnalysis.issueGradeLimitDetail
```

Deutsch verwendet „{count} Steigungsüberschreitung(en)“ und
„Steigung {grade} % überschreitet Grenzwert {limit} %“. Englisch verwendet
„{count} grade limit exceedance(s)“ und „Grade {grade}% exceeds limit {limit}%“.

- [ ] **Step 4: Vollständige automatisierte Prüfung ausführen**

Run: `cd backend; go test ./...`

Run: `cd frontend; npm.cmd test -- --run`

Run: `cd frontend; npm.cmd run build`

Expected: Alle Go-Pakete, Vitest-Dateien und der Produktionsbuild bestehen.

- [ ] **Step 5: Lokalen Server mit aktuellem Build neu starten und Browser prüfen**

Die Browserabnahme setzt die Anlage auf 3,00 Prozent, erzeugt an einem 166-mm-Gleis eine betragsmäßig
größere Steigung, prüft Zähler, Detailtext und Fokus, gleicht das Höhenprofil anschließend auf einen
zulässigen Wert an und lädt die Seite neu. `/health` muss HTTP 200 liefern, die Sitzung bleibt als
`codex-test` angemeldet und die Browserkonsole bleibt fehlerfrei.

- [ ] **Step 6: Abnahme dokumentieren und lokal committen**

Die Spec erhält Branch, Testzahlen, Buildzahl, Browserwerte, Server-PID und die Bestätigung „kein
Push, PR oder Merge“. Dieser Plan erhält den Status „lokal vollständig umgesetzt“ und alle erledigten
Checkboxen.

```powershell
git add frontend/src/features/layouts/TrackPlanAnalysisPanel.tsx `
  frontend/src/features/layouts/TrackPlanAnalysisPanel.test.tsx frontend/src/shared/i18n/de.ts `
  frontend/src/shared/i18n/en.ts frontend/src/shared/apiLayoutsAccessories.ts
git commit -m "feat(planner): show grade limit warnings"

git add docs/superpowers/specs/2026-08-10-erweiterte-geometrie-steigungsgrenzen-design.md `
  docs/superpowers/plans/2026-08-10-erweiterte-geometrie-steigungsgrenzen.md
git commit -m "docs: record grade limit acceptance"
```
