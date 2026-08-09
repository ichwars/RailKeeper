# Artikelübersicht Bestandslayout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Die vorhandene Artikelübersicht übernimmt ohne neue Inhalte oder Funktionen die visuelle Hierarchie der Fahrzeugbestandsseite.

**Architecture:** Bestehende Bestandsklassen werden direkt auf Kopf, Kennzahlen, Werkzeugbereich und Tabelle angewendet. Zubehörspezifische Klassen bleiben nur für Spaltenbreiten, Menüs und responsive Tabellenführung zuständig. Datenfluss, Rollen und Interaktionen bleiben unverändert.

**Tech Stack:** React 19, TypeScript, Vitest, Testing Library, bestehende RailKeeper-CSS-Tokens und Lucide-Icons.

## Global Constraints

- Bild 2 der Nutzerfreigabe ist die verbindliche visuelle Referenz.
- Keine neuen Inhalte, Funktionen, Filter, Ansichten oder Aktionen.
- Keine neuen Abhängigkeiten, Assets oder UI-Komponenten.
- Dark/Light, Deutsch/Englisch, Desktop/Mobile und Tastaturbedienung bleiben erhalten.
- `.superpowers/`, `frontend/dist`, Caches und lokale Daten bleiben uncommitted.

---

### Task 1: Vorhandene Artikelübersicht auf Bestandsprimitiven ausrichten

**Files:**
- Modify: `frontend/src/features/accessories/AccessoriesView.test.tsx`
- Modify: `frontend/src/features/accessories/AccessoriesView.tsx`
- Modify: `frontend/src/features/accessories/ArticleOverviewHeader.tsx`
- Modify: `frontend/src/features/accessories/ArticleMetrics.tsx`
- Modify: `frontend/src/features/accessories/ArticleToolbar.tsx`
- Modify: `frontend/src/features/accessories/ArticleTable.tsx`
- Modify: `frontend/src/styles/accessories.css`

**Interfaces:**
- Consumes: bestehende `inventory-head`, `inventory-status-row`, `inventory-status-card`, `inventory-panel`, `inventory-list-head`, `inventory-filter-result` und `inventory-table` Klassen.
- Produces: unveränderte Artikel-API und Bedienlogik mit an den Fahrzeugbestand angeglichener DOM- und CSS-Hierarchie.

- [ ] **Step 1: Add the failing structural regression test**

Erweitere den ersten Test in `AccessoriesView.test.tsx` um:

```tsx
expect(screen.queryByText("WERKSTATT UND SAMMLUNG")).not.toBeInTheDocument();

const metrics = screen.getByLabelText("Artikelkennzahlen");
expect(metrics).toHaveClass("inventory-status-row");
expect(screen.getAllByTestId("article-metric")).toHaveLength(4);
for (const metric of screen.getAllByTestId("article-metric")) {
  expect(metric).toHaveClass("inventory-status-card");
}

const list = screen.getByRole("region", { name: "Artikel" });
expect(list).toHaveClass("inventory-panel");
expect(within(list).getByRole("searchbox", { name: "Artikel suchen" }))
  .toBeInTheDocument();
expect(within(list).getByRole("table")).toHaveClass("inventory-table");
```

- [ ] **Step 2: Run RED**

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/AccessoriesView.test.tsx
```

Expected: FAIL, weil Oberzeile, Bestandskarten-, Bestandspanel- und Bestandstabellenklassen noch fehlen.

- [ ] **Step 3: Apply the existing vehicle inventory hierarchy**

In `ArticleOverviewHeader.tsx` die Eyebrow-Zeile entfernen und Titel, Beschreibung und Aktion unverändert lassen.

In `ArticleMetrics.tsx` die vorhandenen vier Karten so rendern:

```tsx
<section className="inventory-status-row article-metrics" aria-label={t("accessories.metrics.label")}>
  {cards.map(({ key, icon: Icon, label, value, active, action, actionLabel }) => (
    <article key={key} data-testid="article-metric"
      className={active ? "inventory-status-card article-metric active" : "inventory-status-card article-metric"}>
      {action ? (
        <button type="button" onClick={action} aria-pressed={active} aria-label={actionLabel}>
          <span><Icon size={16} aria-hidden="true" /></span>
          <small>{label}</small>
          <strong>{value}</strong>
        </button>
      ) : (
        <>
          <span><Icon size={16} aria-hidden="true" /></span>
          <small>{label}</small>
          <strong>{value}</strong>
        </>
      )}
    </article>
  ))}
</section>
```

In `AccessoriesView.tsx` Werkzeugleiste in die Bestandskopfzeile ziehen und die Region benennen:

```tsx
<section className="panel inventory-panel article-overview-panel"
  aria-label={t("accessories.overview.listTitle")} aria-busy={overview.loading}>
  <div className="panel-head inventory-list-head article-list-head">
    <div>
      <h2>{t("accessories.overview.listTitle")}</h2>
      <p>{t("accessories.overview.listSubtitle")}</p>
    </div>
    <ArticleToolbar ... />
  </div>
  ...
</section>
```

In `ArticleToolbar.tsx` nur die Trefferklasse ergänzen:

```tsx
<span className="inventory-filter-result article-result-count" aria-live="polite">
```

In `ArticleTable.tsx` die vorhandene Tabelle ergänzen:

```tsx
<table className="inventory-table article-table">
```

In `accessories.css` die Bestandsklassen wirken lassen: vier gleich breite Kennzahlenkarten, `article-toolbar` als `display: contents`, Suche in Spalte 2, Filter über beide Spalten, transparentes Panel und ausschließlich zubehörspezifische Tabellenbreiten beibehalten. Keine Farben, Schatten oder Schriftgrößen hart codieren.

- [ ] **Step 4: Run GREEN and complete frontend verification**

```powershell
cd frontend
npm.cmd run test:run -- src/features/accessories/AccessoriesView.test.tsx src/features/accessories/ArticleTable.test.tsx
npm.cmd run test:run
npm.cmd run build
```

Expected: alle fokussierten Tests, alle Frontendtests und der Produktionsbuild PASS.

- [ ] **Step 5: Run browser comparison**

Prüfe in der In-App-Browserseite `/accessories` bei derselben Breite wie Bild 1:

- kein Eyebrow,
- Kopfzeile und Aktion wie Fahrzeugbestand,
- vier Bestandskarten mit unveränderten Artikelwerten,
- transparenter Listenbereich,
- Suche rechts neben dem Listentitel,
- Filter und Trefferzahl in einer Zeile,
- Tabelle im Fahrzeugbestandsstil,
- 390 × 844 px ohne Dokumentüberlauf,
- Dark und Light ohne neue Warnungen oder Fehler.

- [ ] **Step 6: Close docs and commit**

Setze den Designstatus auf `Implemented and verified`, markiere alle Planpunkte `[x]`, dann:

```powershell
git diff --check
git add docs/superpowers/specs/2026-08-09-artikeluebersicht-bestandslayout-design.md `
  docs/superpowers/plans/2026-08-09-artikeluebersicht-bestandslayout.md `
  frontend/src/features/accessories/AccessoriesView.test.tsx `
  frontend/src/features/accessories/AccessoriesView.tsx `
  frontend/src/features/accessories/ArticleOverviewHeader.tsx `
  frontend/src/features/accessories/ArticleMetrics.tsx `
  frontend/src/features/accessories/ArticleToolbar.tsx `
  frontend/src/features/accessories/ArticleTable.tsx `
  frontend/src/styles/accessories.css
git commit -m "style: align article overview with inventory"
```
