import { fireEvent, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { ArticleTable } from "./ArticleTable";
import {
  articleTableColumnWidthDefinitions,
  type ArticleTableColumn
} from "./articleTableColumns";

const article: AccessoryArticleListItem = {
  id: "article-1",
  inventoryNumber: "RK-ART-000001",
  manufacturer: "Ein sehr langer Herstellername für Modellbahnartikel",
  articleNumber: "83101",
  name: "Gerades Modellgleis mit besonders langer Bezeichnung",
  articleType: "track",
  subtype: "straight",
  gauges: ["TT"],
  listPrice: "1299.90",
  inventoryStrategy: "quantity",
  archived: false,
  owned: 18,
  available: 12,
  reserved: 4,
  installed: 2,
  locationNames: ["Werkstatt / Schrank A / Fach mit langem Namen"],
  hasUsageHistory: true,
  careHintCount: 0,
  updatedAt: "2026-08-08T10:00:00Z",
  attributes: []
};

const secondArticle: AccessoryArticleListItem = {
  ...article,
  id: "article-2",
  articleNumber: "5220",
  name: "Lichtsignal"
};
const subtypes: MasterDataEntry[] = [{
  id: "straight", type: "accessory_subtype", key: "track:straight", label: "Straight", active: true,
  sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z"
}, {
  id: "custom", type: "accessory_subtype", key: "track:club_profile", label: "Club profile", active: true,
  sortOrder: 20, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z"
}];
const articleTypes: MasterDataEntry[] = [{
  id: "track", type: "article_type", key: "track", label: "Gleismaterial", active: false,
  sortOrder: 10, metadata: {}, createdAt: "2026-08-08T08:00:00Z", updatedAt: "2026-08-08T08:00:00Z"
}];

describe("ArticleTable", () => {
  it("renders the exact approved columns and semantic sortable headers", async () => {
    const onSort = vi.fn();
    const user = userEvent.setup();
    render(
      <ArticleTable
        items={[article]}
        sort="inventoryNumber"
        direction="asc"
        canEdit
        onSort={onSort}
        onView={vi.fn()}
        onEdit={vi.fn()}
        onArchive={vi.fn()}
        onRestore={vi.fn()}
      />
    );

    expect(screen.getAllByRole("columnheader").map((header) => header.textContent?.trim())).toEqual([
      "", "Bild", "Inventarnummer", "Hersteller", "Artikelnummer", "Name",
      "Art / Unterart", "Spur", "Bestand", "Lagerung", "Aktionen"
    ]);
    expect(screen.getByRole("columnheader", { name: "Inventarnummer" })).toHaveAttribute("aria-sort", "ascending");
    expect(screen.getByRole("columnheader", { name: "Bestand" })).toHaveAttribute("aria-sort", "none");
    expect(screen.getByRole("table")).toHaveClass("article-table");
    expect(screen.getByRole("table").parentElement).toHaveClass("article-table-wrap");
    await user.click(screen.getByRole("button", { name: "Nach Bestand sortieren" }));
    expect(onSort).toHaveBeenCalledWith("stock");
  });

  it("selects individual and all visible rows without exposing a bulk action", async () => {
    const user = userEvent.setup();
    const onToggleSelection = vi.fn();
    const onToggleAll = vi.fn();
    render(<ArticleTable items={[article, secondArticle]} selectedIDs={new Set([article.id])}
      sort="inventoryNumber" direction="asc" canEdit onSort={vi.fn()}
      onToggleSelection={onToggleSelection} onToggleAll={onToggleAll}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    const selectAll = screen.getByRole("checkbox", { name: "Alle sichtbaren Artikel auswählen" });
    expect(selectAll).not.toBeChecked();
    expect(selectAll).toHaveProperty("indeterminate", true);
    await user.click(screen.getByRole("checkbox", { name: `Artikel auswählen: ${secondArticle.name}` }));
    expect(onToggleSelection).toHaveBeenCalledWith(secondArticle.id);
    await user.click(selectAll);
    expect(onToggleAll).toHaveBeenCalledOnce();
    expect(screen.queryByRole("button", { name: /ausgewählte artikel/i })).not.toBeInTheDocument();
  });

  it("renders localized built-in subtype labels and preserves configured custom labels", () => {
    setLanguage("de");
    const view = render(<ArticleTable items={[article]} subtypeEntries={subtypes} sort="article" direction="asc"
      canEdit={false} onSort={vi.fn()} onView={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);
    expect(screen.getByText("Gerade")).toBeInTheDocument();

    view.rerender(<ArticleTable items={[{ ...article, subtype: "club_profile" }]} subtypeEntries={subtypes}
      sort="article" direction="asc" canEdit={false} onSort={vi.fn()} onView={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);
    expect(screen.getByText("Club profile")).toBeInTheDocument();

    setLanguage("en");
    view.rerender(<ArticleTable items={[article]} subtypeEntries={subtypes} sort="article" direction="asc"
      canEdit={false} onSort={vi.fn()} onView={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);
    expect(screen.getByText("Straight")).toBeInTheDocument();
    setLanguage("de");
  });

  it("honors administrator-renamed built-in subtype labels in every language", () => {
    const renamed = [{ ...subtypes[0]!, label: "Werkstattgerade" }, subtypes[1]!];
    setLanguage("de");
    const view = render(<ArticleTable items={[article]} subtypeEntries={renamed} sort="article" direction="asc"
      canEdit={false} onSort={vi.fn()} onView={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);
    expect(screen.getByText("Werkstattgerade")).toBeInTheDocument();

    setLanguage("en");
    view.rerender(<ArticleTable items={[article]} subtypeEntries={renamed} sort="article" direction="asc"
      canEdit={false} onSort={vi.fn()} onView={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);
    expect(screen.getByText("Werkstattgerade")).toBeInTheDocument();
    setLanguage("de");
  });

  it("uses the configured article type label even when the result type is inactive", () => {
    render(<ArticleTable items={[article]} articleTypeEntries={articleTypes} sort="article" direction="asc"
      canEdit={false} onSort={vi.fn()} onView={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getByText("Gleismaterial")).toBeInTheDocument();
    expect(screen.queryByText("Gleis")).not.toBeInTheDocument();
  });

  it("localizes canonical subtype keys returned by the backend", () => {
    render(<ArticleTable items={[{ ...article, subtype: "track:straight" }]} subtypeEntries={subtypes}
      sort="article" direction="asc" canEdit={false} onSort={vi.fn()} onView={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getByText("Gerade")).toBeInTheDocument();
    expect(screen.queryByText("track:straight")).not.toBeInTheDocument();
  });

  it("uses accessible transparent direct row actions for archive and restore", async () => {
    const onView = vi.fn();
    const onEdit = vi.fn();
    const onArchive = vi.fn();
    const onRestore = vi.fn();
    const user = userEvent.setup();
    const { rerender } = render(
      <ArticleTable items={[article]} sort="article" direction="asc" canEdit onSort={vi.fn()}
        onView={onView} onEdit={onEdit} onArchive={onArchive} onRestore={onRestore} />
    );

    const row = screen.getByRole("row", { name: /Gerades Modellgleis/ });
    const view = within(row).getByRole("button", { name: "Artikel ansehen: Gerades Modellgleis mit besonders langer Bezeichnung" });
    const edit = within(row).getByRole("button", { name: "Artikel bearbeiten: Gerades Modellgleis mit besonders langer Bezeichnung" });
    const archive = within(row).getByRole("button", { name: "Artikel archivieren: Gerades Modellgleis mit besonders langer Bezeichnung" });
    expect(view).toHaveClass("icon-button", "article-action-button");
    expect(view).toHaveAttribute("title", "Artikel ansehen");
    await user.click(view);
    await user.click(edit);
    await user.click(archive);
    expect(onView).toHaveBeenCalledWith(article);
    expect(onEdit).toHaveBeenCalledWith(article);
    expect(onArchive).toHaveBeenCalledWith(article);

    rerender(
      <ArticleTable items={[{ ...article, archived: true }]} sort="article" direction="asc" canEdit
        onSort={vi.fn()} onView={onView} onEdit={onEdit} onArchive={onArchive} onRestore={onRestore} />
    );
    await user.click(screen.getByRole("button", { name: /Artikel wiederherstellen:/ }));
    expect(onRestore).toHaveBeenCalledWith(expect.objectContaining({ id: "article-1", archived: true }));
    expect(screen.queryByRole("button", { name: /Weitere Aktionen/ })).not.toBeInTheDocument();
  });

  it("keeps full long values accessible while visually truncating them", () => {
    render(
      <ArticleTable items={[article]} sort="article" direction="asc" canEdit={false} onSort={vi.fn()}
        onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />
    );

    expect(screen.getByText(article.name)).toHaveAttribute("title", article.name);
    expect(screen.getByText(article.manufacturer)).toHaveAttribute("title", article.manufacturer);
    expect(screen.getByText(article.locationNames[0]!)).toHaveAttribute("title", article.locationNames[0]);
    expect(screen.queryByRole("button", { name: /bearbeiten/i })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel archivieren:/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Artikel ansehen/ })).toBeInTheDocument();
  });

  it("renders direct admin delete in the correct table row", () => {
    render(
      <ArticleTable items={[article, secondArticle]} sort="article" direction="asc" canEdit canDelete
        onSort={vi.fn()} onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()}
        onDelete={vi.fn()} />
    );

    const rows = screen.getAllByRole("row");
    expect(within(rows[1]!).getByRole("button", { name: /Artikel löschen: Gerades/ })).toBeInTheDocument();
    expect(within(rows[2]!).getByRole("button", { name: /Artikel löschen: Lichtsignal/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Weitere Aktionen/ })).not.toBeInTheDocument();
  });

  it("hides matching headers and cells while retaining selection and actions", () => {
    const columns: ArticleTableColumn[] = ["inventoryNumber", "name"];
    render(
      <ArticleTable items={[article]} columns={columns}
        sort="article" direction="asc" canEdit onSort={vi.fn()} onView={vi.fn()}
        onArchive={vi.fn()} onRestore={vi.fn()} />
    );

    expect(screen.queryByRole("columnheader", { name: "Hersteller" })).not.toBeInTheDocument();
    expect(screen.queryByText(article.manufacturer)).not.toBeInTheDocument();
    expect(screen.getByText(article.inventoryNumber)).toBeInTheDocument();
    expect(screen.getByText(article.name)).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Auswahl" })).toBeInTheDocument();
    expect(screen.getByRole("columnheader", { name: "Aktionen" })).toBeInTheDocument();
    expect(screen.getByRole("table")).toHaveStyle("--article-table-min-width: 532px");
  });

  it("renders an optional nonsortable list-price column with exact locale formatting", () => {
    setLanguage("de");
    render(
      <ArticleTable items={[article]}
        columns={["inventoryNumber", "listPrice"]}
        sort="article" direction="asc" canEdit={false} onSort={vi.fn()}
        onArchive={vi.fn()} onRestore={vi.fn()} />
    );

    const header = screen.getByRole("columnheader", { name: "Listenpreis pro Stück" });
    expect(within(header).queryByRole("button")).not.toBeInTheDocument();
    expect(screen.getByText("1.299,90 €")).toBeInTheDocument();
  });

  it("renders user order, configured widths, and resizers only for data columns", () => {
    const onCommitColumnWidth = vi.fn();
    render(
      <ArticleTable items={[article]} columns={["manufacturer", "inventoryNumber"]}
        columnWidths={{ manufacturer: 196 }} sort="article" direction="asc" canEdit={false}
        onSort={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()}
        onPreviewColumnWidth={vi.fn()} onCommitColumnWidth={onCommitColumnWidth} />
    );

    expect(screen.getAllByRole("columnheader").map((header) => header.textContent?.trim()))
      .toEqual(["", "Hersteller", "Inventarnummer", "Aktionen"]);
    const expectedWidth = 44 + 136 + 196 +
      articleTableColumnWidthDefinitions.inventoryNumber.defaultWidth;
    const table = screen.getByRole("table");
    expect(table).toHaveStyle(`--article-table-min-width: ${expectedWidth}px`);
    expect(table.querySelector('col[data-column="manufacturer"]')).toHaveStyle("width: 196px");
    expect(table.querySelector('col[data-column="inventoryNumber"]')).toHaveStyle(
      `width: ${articleTableColumnWidthDefinitions.inventoryNumber.defaultWidth}px`
    );
    expect(table.querySelector("col.table-fill-cell")).toHaveStyle(
      `width: max(0px, calc(100% - ${expectedWidth}px))`
    );
    expect(table.querySelector("col.select-cell")).toHaveStyle({
      width: "44px",
      minWidth: "44px",
      maxWidth: "44px"
    });
    expect(table.querySelector("col.actions-cell")).toHaveStyle({
      width: "136px",
      minWidth: "136px",
      maxWidth: "136px"
    });
    expect(screen.getAllByRole("separator")).toHaveLength(2);
    fireEvent.keyDown(screen.getByRole("separator", {
      name: "Breite von Hersteller ändern"
    }), { key: "ArrowRight" });
    expect(onCommitColumnWidth).toHaveBeenCalledWith("manufacturer", 204);
  });

  it("hides column resizers while profile widths are loading", () => {
    render(
      <ArticleTable items={[article]} columns={["manufacturer", "inventoryNumber"]}
        columnWidthsLoading sort="article" direction="asc" canEdit={false}
        onSort={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()}
        onPreviewColumnWidth={vi.fn()} onCommitColumnWidth={vi.fn()} />
    );

    expect(screen.queryAllByRole("separator")).toHaveLength(0);
  });
});
