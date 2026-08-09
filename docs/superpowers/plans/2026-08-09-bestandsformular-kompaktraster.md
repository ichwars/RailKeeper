# Bestandsformular Kompaktraster Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Bestandskorrektur und Umbuchung im Artikeldialog als kompaktes 1:2-Raster ohne gestreckte Leerflächen anordnen.

**Architecture:** Die vorhandenen Formulare, Komponenten und Submit-Handler bleiben unverändert. Ein lokaler Container gruppiert die vier Umbuchungsfelder; `accessories.css` verteilt den äußeren Bereich mit 1:2-Gewichtung und den Umbuchungscontainer als responsives 2×2-Raster.

**Tech Stack:** React 19, TypeScript 7, CSS Grid, Vitest, Testing Library, Vite, RailKeeper Browser-QA.

## Global Constraints

- Die Bestandskorrektur belegt auf Desktop ungefähr ein Drittel der verfügbaren Breite.
- Die Umbuchung belegt auf Desktop ungefähr zwei Drittel der verfügbaren Breite.
- Beide Desktop-Schaltflächen schließen auf derselben unteren Linie ab.
- Unterhalb von 920 Pixeln stehen Bestandskorrektur und Umbuchung untereinander.
- Unterhalb von 560 Pixeln wird auch die Umbuchung einspaltig.
- Keine Änderung an API, Datenmodell, Validierung, Übersetzungen oder Feldinhalten.
- Keine Änderung der Dialoggröße oder der festen Aktionsleiste.
- Bestehende RailKeeper-Komponenten, Abstände und Design-Tokens werden wiederverwendet.
- Ausschließlich inline arbeiten; keine Subagents einsetzen.

---

## File Map

- `frontend/src/features/accessories/AccessoryStockPanel.tsx`: gruppiert ausschließlich die vier
  vorhandenen Umbuchungsfelder und markiert das Umbuchungsformular lokal.
- `frontend/src/styles/accessories.css`: definiert das gewichtete äußere Raster, das interne
  Umbuchungsraster und den 560-Pixel-Breakpoint; entfernt die bisherige Höhenstreckung.
- `frontend/src/features/accessories/AccessoryStockPanel.test.tsx`: sichert Formular- und
  Feldcontainerstruktur.
- `frontend/src/features/accessories/accessoriesResponsive.test.ts`: sichert 1:2-Aufteilung,
  2×2-Umbuchungsraster, mobile Einspaltigkeit und das Entfernen der Streckregeln.
- `docs/superpowers/specs/2026-08-09-bestandsformular-kompaktraster-design.md`: erhält nach
  erfolgreicher Verifikation den Status `Implemented and verified`.
- `docs/superpowers/plans/2026-08-09-bestandsformular-kompaktraster.md`: verfolgt die erledigten
  Schritte.
- `design-qa.md`: dokumentiert Referenzvergleich, Viewports, Themes und offene Findings.

### Task 1: Gewichtetes Kompaktraster implementieren

**Files:**
- Modify: `frontend/src/features/accessories/AccessoryStockPanel.tsx:131-154`
- Modify: `frontend/src/styles/accessories.css:475-495`
- Modify: `frontend/src/styles/accessories.css:652-664`
- Modify: `frontend/src/styles/accessories.css:666-720`
- Test: `frontend/src/features/accessories/AccessoryStockPanel.test.tsx:26-45`
- Test: `frontend/src/features/accessories/accessoriesResponsive.test.ts:33-44`

**Interfaces:**
- Consumes: vorhandene `AccessoryStockPanel`-Props, `.article-stock-commands`,
  `.article-stock-form`, `LocationSelect`, `AppNumberInput` und `AppTextInput`.
- Produces: `.article-transfer-form` als lokale Formularmarkierung und
  `.article-transfer-fields` als rein visuellen Container für vier bestehende Felder.

- [x] **Step 1: Strukturtest um den Umbuchungscontainer ergänzen**

Den bestehenden Test `aligns stock adjustment and transfer as peer forms` erweitern:

```tsx
const commands = view.container.querySelector(".article-stock-commands");
const forms = commands?.querySelectorAll(":scope > .article-stock-form");
expect(forms).toHaveLength(2);
expect(forms?.[0]?.querySelector(".primary-button")).toHaveTextContent("Bestand buchen");
expect(forms?.[1]).toHaveClass("article-transfer-form");
expect(forms?.[1]?.querySelector(".primary-button")).toHaveTextContent("Umbuchen");

const transferFields = forms?.[1]?.querySelector(".article-transfer-fields");
expect(transferFields).toBeInTheDocument();
expect(transferFields?.children).toHaveLength(4);
```

- [x] **Step 2: Responsive-Test um die neuen Rasterregeln ergänzen**

In `accessoriesResponsive.test.ts` ergänzen:

```ts
it("uses the compact weighted stock grid and stacks transfer fields on narrow screens", () => {
  expect(accessoriesCss).toMatch(
    /\.article-stock-commands\s*\{[^}]*grid-template-columns:\s*minmax\(240px,\s*1fr\)\s*minmax\(0,\s*2fr\)/s,
  );
  expect(accessoriesCss).toMatch(
    /\.article-transfer-fields\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\)/s,
  );
  expect(accessoriesCss).toMatch(
    /@media\s*\(max-width:\s*560px\)[\s\S]*\.article-transfer-fields\s*\{[^}]*grid-template-columns:\s*1fr/s,
  );
  expect(accessoriesCss).not.toMatch(/\.article-stock-form\s*\{[^}]*height:\s*100%/s);
  expect(accessoriesCss).not.toContain(".article-stock-form > .primary-button");
});
```

- [x] **Step 3: Beide Fokustests ausführen und erwartetes Rot bestätigen**

Run:

```powershell
npm.cmd run test:run -- src/features/accessories/AccessoryStockPanel.test.tsx src/features/accessories/accessoriesResponsive.test.ts
```

Expected: FAIL wegen fehlender `.article-transfer-form`- und `.article-transfer-fields`-Struktur
sowie der noch vorhandenen Gleichverteilungs- und Streckregeln.

- [x] **Step 4: Die vier Umbuchungsfelder lokal gruppieren**

In `AccessoryStockPanel.tsx` ausschließlich das zweite Formular anpassen:

```tsx
<form className="accessory-form article-stock-form article-transfer-form" onSubmit={submitTransfer}>
  <h4><ArrowRightLeft size={15} aria-hidden="true" /> {t("accessories.editor.stock.transfer")}</h4>
  <div className="article-transfer-fields">
    <LocationSelect label={t("accessories.editor.stock.fromLocation")} value={effectiveLocationId}
      locations={activeLocations} allLocations={locations} onChange={setLocationId} />
    <LocationSelect label={t("accessories.editor.stock.toLocation")} value={effectiveTransferToId}
      locations={targetLocations} allLocations={locations} onChange={setTransferToId} />
    <AppNumberInput label={t("accessories.field.quantity")} min="1" required value={transferQuantity}
      onValueChange={setTransferQuantity} />
    <AppTextInput label={t("accessories.field.notes")} value={transferNote}
      onChange={(event) => setTransferNote(event.target.value)} />
  </div>
  <button type="submit" className="primary-button"
    disabled={!effectiveLocationId || !effectiveTransferToId || Number(transferQuantity) <= 0}>
    {t("accessories.editor.stock.transfer")}
  </button>
</form>
```

Die vorhandene `disabled`-Bedingung und alle Handler bleiben unverändert.

- [x] **Step 5: Gleichverteilung und Höhenstreckung durch das Kompaktraster ersetzen**

In `accessories.css` den aktuellen lokalen Block ersetzen:

```css
.article-stock-commands {
  grid-template-columns: minmax(240px, 1fr) minmax(0, 2fr);
  align-items: start;
  margin-top: 11px;
}

.article-transfer-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 11px;
}
```

Die Regeln `.article-stock-form { ... height: 100%; }` und
`.article-stock-form > .primary-button { margin-top: auto; }` vollständig entfernen.

Im vorhandenen 560-Pixel-Media-Block ergänzen:

```css
.article-transfer-fields {
  grid-template-columns: 1fr;
}
```

Der vorhandene 920-Pixel-Block für `.article-stock-commands { grid-template-columns: 1fr; }`
bleibt unverändert.

- [x] **Step 6: Fokustests erneut ausführen**

Run:

```powershell
npm.cmd run test:run -- src/features/accessories/AccessoryStockPanel.test.tsx src/features/accessories/accessoriesResponsive.test.ts
```

Expected: beide Testdateien PASS.

- [x] **Step 7: Task-Änderungen statisch prüfen**

Run: `git diff --check`

Expected: kein Output.

- [x] **Step 8: Implementierung committen**

```powershell
git add -- frontend/src/features/accessories/AccessoryStockPanel.tsx frontend/src/styles/accessories.css frontend/src/features/accessories/AccessoryStockPanel.test.tsx frontend/src/features/accessories/accessoriesResponsive.test.ts
git commit -m "style: compact article stock forms"
```

Expected: ausschließlich die vier Task-Dateien sind im Commit enthalten.

### Task 2: Gesamtverifikation und Dokumentationsabschluss

**Files:**
- Modify: `docs/superpowers/specs/2026-08-09-bestandsformular-kompaktraster-design.md:1-6`
- Modify: `docs/superpowers/plans/2026-08-09-bestandsformular-kompaktraster.md`
- Modify: `design-qa.md`
- Test: vollständiges Frontend und Go-Backend.

**Interfaces:**
- Consumes: das fertig implementierte `.article-transfer-fields`-Raster und den lokalen Server
  unter `http://127.0.0.1:18083`.
- Produces: vollständige Test-, Build- und Browser-Evidenz sowie abgeschlossene Spezifikation.

- [x] **Step 1: Vollständige Frontendtests ausführen**

Run:

```powershell
cd frontend
npm.cmd run test:run
```

Expected: mindestens die aktuelle Baseline von 49 Testdateien und 291 Tests PASS.

- [x] **Step 2: Produktionsbuild ausführen**

Run:

```powershell
cd frontend
npm.cmd run build
```

Expected: TypeScript-Build und Vite-Build erfolgreich.

- [x] **Step 3: Backend-Regressionssuite ausführen**

Run:

```powershell
cd backend
go test ./...
```

Expected: alle Backend-Packages PASS.

- [x] **Step 4: Browser-QA im Artikeldialog durchführen**

Flow: `/accessories` → Tillig 83101 bearbeiten → Reiter `Bestand`.

Prüfen:

- Desktop ab 920 Pixel: äußere Aufteilung ungefähr 1:2, Umbuchung intern 2×2.
- Bestandskorrektur: normale Feldabstände und keine gestreckte Restfläche.
- Beide Desktop-Schaltflächen: gleiche Unterkante.
- Mobil 390 × 844 Pixel: Vorgänge untereinander, Umbuchungsfelder einspaltig, Footer erreichbar.
- Dark und Light Theme: keine Überlagerung oder abgeschnittene Beschriftung.
- Browserkonsole: keine neuen Fehler oder Framework-Overlays.

Die Desktopaufnahme zusammen mit
`C:\Users\droth\AppData\Local\Temp\codex-clipboard-94dbed73-3e14-488c-bb24-8cf53ae27812.png`
und der ausgewählten Variante B vergleichen.

- [x] **Step 5: Design-QA und Spezifikationsstatus aktualisieren**

In der Spezifikation setzen:

```markdown
**Status:** Implemented and verified
```

In `design-qa.md` einen Abschnitt `Nachtrag: Bestandsformular Kompaktraster` ergänzen. Der Abschnitt
muss Referenz, Desktop-/Mobil-Viewports, Dark/Light, Vergleichsergebnis, Findings P0 bis P3 und den
abschließenden Wert `passed` enthalten.

- [x] **Step 6: Planchecklisten abschließen und Dokumente prüfen**

Alle erledigten Schritte von `[ ]` auf `[x]` setzen.

Run: `git diff --check`

Expected: kein Output.

- [x] **Step 7: Abschlussdokumentation committen**

```powershell
git add -- design-qa.md docs/superpowers/specs/2026-08-09-bestandsformular-kompaktraster-design.md docs/superpowers/plans/2026-08-09-bestandsformular-kompaktraster.md
git commit -m "docs: verify compact stock form grid"
```

Expected: `.superpowers/`, `frontend/dist`, Caches und lokale Daten bleiben unversioniert.
