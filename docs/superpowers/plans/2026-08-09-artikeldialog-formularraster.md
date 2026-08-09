# Artikeldialog Formularraster Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bestand, Reservierung, Einbau und Boolean-Fachangaben im Artikeldialog konsistent und responsiv ausrichten.

**Architecture:** Die bestehende Komponentenstruktur und Datenbindung bleiben unverändert. Betroffene Formulare erhalten gezielte semantische Klassen; `accessories.css` definiert daraus gleich hohe Bestandsformulare, ein responsives Zweispaltenraster und eine kontrollhohe Checkboxzeile.

**Tech Stack:** React 19, TypeScript 7, CSS Grid, Vitest, Testing Library, Vite, RailKeeper Browser-QA.

## Global Constraints

- Bestehende RailKeeper-Komponenten, Design-Tokens und Breakpoints werden wiederverwendet.
- Keine Änderung an API, Datenmodell, Validierung, Feldreihenfolge oder Übersetzungen.
- Keine neue UI-Abhängigkeit und kein neues allgemeines Rastersystem.
- Auf Desktop zwei gleich breite Spalten, unter 920 Pixeln eine Spalte.
- Dark und Light Theme sowie die feste Dialogaktionsleiste bleiben funktionsfähig.
- Ausschließlich inline arbeiten; keine Subagents einsetzen.

---

## File Map

- `frontend/src/features/accessories/AccessoryStockPanel.tsx`: semantische Klassen für die beiden Bestandsformulare.
- `frontend/src/features/accessories/AccessoryReservationsPanel.tsx`: Zweispaltenklasse und breite Formularzeilen.
- `frontend/src/features/accessories/AccessoryInstallationsPanel.tsx`: Zweispaltenklasse und breite Formularzeilen.
- `frontend/src/features/accessories/ArticleSubjectTab.tsx`: getrennte Checkbox-Kontrollzeile.
- `frontend/src/styles/accessories.css`: alle lokalen Raster-, Höhen- und Responsive-Regeln.
- `frontend/src/features/accessories/AccessoryStockPanel.test.tsx`: Strukturtest der Bestandsformulare.
- `frontend/src/features/accessories/AccessoryAllocationPanels.test.tsx`: Strukturtests für Reservieren und Einbauen.
- `frontend/src/features/accessories/ArticleSubjectTab.test.tsx`: Strukturtest der Boolean-Felder.
- `frontend/src/features/accessories/accessoriesResponsive.test.ts`: statischer Schutz der Desktop- und Mobilregeln.

### Task 1: Bestandsformulare symmetrisch ausrichten

**Files:**
- Modify: `frontend/src/features/accessories/AccessoryStockPanel.tsx`
- Modify: `frontend/src/styles/accessories.css`
- Test: `frontend/src/features/accessories/AccessoryStockPanel.test.tsx`

**Interfaces:**
- Consumes: bestehende `AccessoryStockPanel`-Props und `.accessory-form`.
- Produces: `.article-stock-form` als rein lokale Layoutklasse ohne Verhaltensänderung.

- [x] **Step 1: Strukturtest für zwei markierte Bestandsformulare ergänzen**

```tsx
it("aligns stock adjustment and transfer as peer forms", () => {
  const quantityArticle = { ...article, trackingMode: "quantity" as const,
    inventoryStrategy: "quantity" as const };
  const view = render(<AccessoryStockPanel article={quantityArticle} stock={null} movements={[]} assets={[]}
    locations={[location]} canEdit onChanged={vi.fn()} onDirtyChange={vi.fn()} />);
  const forms = view.container.querySelectorAll(".article-stock-commands > .article-stock-form");
  expect(forms).toHaveLength(2);
  expect(forms[0]?.querySelector(".primary-button")).toHaveTextContent("Bestand buchen");
  expect(forms[1]?.querySelector(".primary-button")).toHaveTextContent("Umbuchen");
});
```

- [x] **Step 2: Fokustest ausführen und erwartetes Rot bestätigen**

Run: `npm.cmd test -- --run src/features/accessories/AccessoryStockPanel.test.tsx`
Expected: FAIL, weil `.article-stock-form` noch nicht existiert.

- [x] **Step 3: Beide Formulare markieren und den Label-Selektor auf direkte Kindlabels begrenzen**

```tsx
<form className="accessory-form article-stock-form" onSubmit={submitAdjustment}>
...
<form className="accessory-form article-stock-form" onSubmit={submitTransfer}>
```

```css
.article-stock-commands {
  align-items: stretch;
}

.article-stock-form {
  height: 100%;
}

.article-stock-form > .primary-button {
  margin-top: auto;
}

.accessory-form > label {
  display: grid;
  gap: 6px;
  color: var(--muted);
  font-size: var(--font-size-md);
  font-weight: var(--font-weight-bold);
}
```

Der bestehende Selektor `.accessory-form label` wird durch `.accessory-form > label` ersetzt. Dadurch behandelt er nicht mehr das innere `.app-field-label` von `AppNumberInput` als Raster und Pflichtsterne bleiben im Label.

- [x] **Step 4: Fokustest erneut ausführen**

Run: `npm.cmd test -- --run src/features/accessories/AccessoryStockPanel.test.tsx`
Expected: PASS.

- [x] **Step 5: Task-Änderungen prüfen**

Run: `git diff --check`
Expected: kein Output.

### Task 2: Reservieren und Einbauen zweispaltig anordnen

**Files:**
- Modify: `frontend/src/features/accessories/AccessoryReservationsPanel.tsx`
- Modify: `frontend/src/features/accessories/AccessoryInstallationsPanel.tsx`
- Modify: `frontend/src/styles/accessories.css`
- Test: `frontend/src/features/accessories/AccessoryAllocationPanels.test.tsx`
- Test: `frontend/src/features/accessories/accessoriesResponsive.test.ts`

**Interfaces:**
- Consumes: `AccessoryTargetFields`, `AccessoryTechnicalFields`, bedingte Mengen- und Einzelstückfelder.
- Produces: `.accessory-allocation-form` für das Raster und `.accessory-form-wide` für kontrollierte Vollbreitenzeilen.

- [x] **Step 1: Strukturtests für beide Allokationsformulare ergänzen**

```tsx
it("marks reservation and installation forms for the responsive two-column grid", () => {
  const reservationView = render(<AccessoryReservationsPanel article={article} reservations={[]} assets={[]}
    locations={[location]} vehicles={[]} layouts={[layout]} units={[]} canReserve
    onChanged={vi.fn()} onDirtyChange={vi.fn()} />);
  expect(screen.getByRole("heading", { name: "Zubehör reservieren" }).closest("form"))
    .toHaveClass("accessory-allocation-form");
  reservationView.unmount();

  render(<AccessoryInstallationsPanel article={article} reservations={[]} installations={[]} assets={[]}
    locations={[location]} vehicles={[]} layouts={[layout]} units={[]} canInstall
    onChanged={vi.fn()} onDirtyChange={vi.fn()} />);
  expect(screen.getByRole("heading", { name: "Zubehör einbauen" }).closest("form"))
    .toHaveClass("accessory-allocation-form");
});
```

```ts
expect(css).toContain(".accessory-allocation-form");
expect(css).toMatch(/@media \(max-width: 920px\)[\s\S]*\.accessory-allocation-form[\s\S]*grid-template-columns: 1fr/);
```

- [x] **Step 2: Fokustests ausführen und erwartetes Rot bestätigen**

Run: `npm.cmd test -- --run src/features/accessories/AccessoryAllocationPanels.test.tsx src/features/accessories/accessoriesResponsive.test.ts`
Expected: FAIL wegen fehlender Allokationsklassen und CSS-Regeln.

- [x] **Step 3: Formular- und Vollbreitenklassen in beiden Panels setzen**

```tsx
<form className="accessory-form accessory-allocation-form" onSubmit={submit}>
...
<label className="accessory-form-wide">{t("accessories.field.notes")}<textarea ... /></label>
```

```tsx
<form className="accessory-form accessory-allocation-form" onSubmit={submitInstallation}>
  <h3>{t("accessories.installations.create")}</h3>
  <label className="accessory-form-wide">{t("accessories.field.reservation")}<AppSelect ... /></label>
  ...
  <label className="accessory-form-wide">{t("accessories.field.notes")}<textarea ... /></label>
```

Die Ausbauform `.accessory-removal-form` bleibt bewusst einspaltig.

- [x] **Step 4: Lokale Desktop- und Mobilregeln ergänzen**

```css
.accessory-allocation-form {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  align-items: start;
}

.accessory-allocation-form > h3,
.accessory-allocation-form > .accessory-form-wide,
.accessory-allocation-form > .accessory-target-summary,
.accessory-allocation-form > .primary-button {
  grid-column: 1 / -1;
}

.accessory-allocation-form > .accessory-technical-fields {
  display: contents;
}

@media (max-width: 920px) {
  .accessory-allocation-form {
    grid-template-columns: 1fr;
  }
}
```

- [x] **Step 5: Fokustests erneut ausführen**

Run: `npm.cmd test -- --run src/features/accessories/AccessoryAllocationPanels.test.tsx src/features/accessories/accessoriesResponsive.test.ts`
Expected: PASS.

### Task 3: Checkboxzeilen an Eingabefeldern ausrichten

**Files:**
- Modify: `frontend/src/features/accessories/ArticleSubjectTab.tsx`
- Modify: `frontend/src/styles/accessories.css`
- Test: `frontend/src/features/accessories/ArticleSubjectTab.test.tsx`

**Interfaces:**
- Consumes: bestehende `.article-checkbox`-Felder und `accessories.subject.booleanHint`.
- Produces: `.article-checkbox-control` als 38 Pixel hohe, vertikal zentrierte Kontrollzeile.

- [x] **Step 1: Boolean-Strukturtest ergänzen**

```tsx
const weatherproof = screen.getByRole("checkbox", { name: "Wetterfest" });
expect(weatherproof.closest(".article-checkbox-control")).toBeInTheDocument();
expect(weatherproof.closest(".article-checkbox")).toHaveClass("app-field");
```

- [x] **Step 2: Fokustest ausführen und erwartetes Rot bestätigen**

Run: `npm.cmd test -- --run src/features/accessories/ArticleSubjectTab.test.tsx`
Expected: FAIL, weil `.article-checkbox-control` noch fehlt.

- [x] **Step 3: Kontrollzeile markieren und CSS-Ausrichtung korrigieren**

```tsx
<span className="article-checkbox-control"><input type="checkbox" ... />
  {t("accessories.subject.booleanHint")}
</span>
```

```css
.article-checkbox {
  display: grid;
  gap: 7px;
  color: var(--text);
  font-size: var(--font-size-sm);
  font-weight: var(--font-weight-bold);
}

.article-checkbox-control {
  display: flex;
  min-height: 38px;
  align-items: center;
  gap: 7px;
}
```

- [x] **Step 4: Fokustest erneut ausführen**

Run: `npm.cmd test -- --run src/features/accessories/ArticleSubjectTab.test.tsx`
Expected: PASS.

### Task 4: Gesamtverifikation und Browser-QA

**Files:**
- Modify: `docs/superpowers/specs/2026-08-09-artikeldialog-formularraster-design.md`
- Modify: `docs/superpowers/plans/2026-08-09-artikeldialog-formularraster.md`
- Test: alle oben genannten Frontendtests.

**Interfaces:**
- Consumes: fertig implementierte Layoutklassen und laufenden Server unter `http://127.0.0.1:18083`.
- Produces: verifizierten Build, Desktop-/Mobilnachweis und abgeschlossene Dokumentation.

- [x] **Step 1: Alle fokussierten Tests gemeinsam ausführen**

Run:
```powershell
npm.cmd test -- --run src/features/accessories/AccessoryStockPanel.test.tsx src/features/accessories/AccessoryAllocationPanels.test.tsx src/features/accessories/ArticleSubjectTab.test.tsx src/features/accessories/accessoriesResponsive.test.ts
```
Expected: alle Testdateien PASS.

- [x] **Step 2: Vollständige Frontendtests und Build ausführen**

Run:
```powershell
npm.cmd test -- --run --reporter=default
npm.cmd run build
```
Expected: 49 Testdateien PASS und Vite-Build erfolgreich.

- [x] **Step 3: Browser-QA durchführen**

Flow: `/accessories` → Artikel öffnen → Reiter `Bestand` → Reservierungs- und Einbauabschnitte → Reiter `Fachangaben: Gleis`.

Prüfen:
- Desktop 1008 Pixel oder breiter: zwei Spalten und bündige Bestandsbuttons.
- Mobil 390 × 844 Pixel: eine Spalte, kein Dokumentüberlauf, Dialogfooter erreichbar.
- Dark und Light Theme.
- Checkboxen `Bettung` und `Digitaltauglich` auf Eingabefeldhöhe.
- keine relevanten Browser-Warnungen oder Framework-Overlays.

- [x] **Step 4: Dokumentstatus und Checklisten abschließen**

```markdown
**Status:** Implemented and verified
```

Alle erledigten Planpunkte werden von `[ ]` auf `[x]` gesetzt.

- [x] **Step 5: Ausschließlich Task-Dateien committen**

```powershell
git add -- frontend/src/features/accessories/AccessoryStockPanel.tsx frontend/src/features/accessories/AccessoryReservationsPanel.tsx frontend/src/features/accessories/AccessoryInstallationsPanel.tsx frontend/src/features/accessories/ArticleSubjectTab.tsx frontend/src/styles/accessories.css frontend/src/features/accessories/AccessoryStockPanel.test.tsx frontend/src/features/accessories/AccessoryAllocationPanels.test.tsx frontend/src/features/accessories/ArticleSubjectTab.test.tsx frontend/src/features/accessories/accessoriesResponsive.test.ts docs/superpowers/specs/2026-08-09-artikeldialog-formularraster-design.md docs/superpowers/plans/2026-08-09-artikeldialog-formularraster.md
git commit -m "style: align article dialog forms"
```

Expected: Commit erfolgreich; `.superpowers/`, `frontend/dist`, Caches und lokale Daten bleiben unversioniert.
