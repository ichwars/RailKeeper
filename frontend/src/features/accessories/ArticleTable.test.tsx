import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { AccessoryArticleListItem } from "../../shared/api";
import { ArticleTable } from "./ArticleTable";

const article: AccessoryArticleListItem = {
  id: "article-1",
  manufacturer: "Ein sehr langer Herstellername für Modellbahnartikel",
  articleNumber: "83101",
  name: "Gerades Modellgleis mit besonders langer Bezeichnung",
  articleType: "track",
  subtype: "straight",
  gauges: ["TT"],
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

describe("ArticleTable", () => {
  it("renders the exact approved columns and semantic sortable headers", async () => {
    const onSort = vi.fn();
    const user = userEvent.setup();
    render(
      <ArticleTable
        items={[article]}
        sort="article"
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
      "Artikel", "Art / Unterart", "Spur", "Bestand", "Lagerung", "Aktionen"
    ]);
    expect(screen.getByRole("columnheader", { name: "Artikel" })).toHaveAttribute("aria-sort", "ascending");
    expect(screen.getByRole("columnheader", { name: "Bestand" })).toHaveAttribute("aria-sort", "none");
    await user.click(screen.getByRole("button", { name: "Nach Bestand sortieren" }));
    expect(onSort).toHaveBeenCalledWith("stock");
  });

  it("uses accessible transparent row actions and archive or restore in overflow", async () => {
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
    const more = within(row).getByRole("button", { name: "Weitere Aktionen: Gerades Modellgleis mit besonders langer Bezeichnung" });
    expect(view).toHaveClass("icon-button", "article-action-button");
    expect(view).toHaveAttribute("title", "Artikel ansehen");
    await user.click(view);
    await user.click(edit);
    await user.click(more);
    await user.click(screen.getByRole("menuitem", { name: "Artikel archivieren" }));
    expect(onView).toHaveBeenCalledWith(article);
    expect(onEdit).toHaveBeenCalledWith(article);
    expect(onArchive).toHaveBeenCalledWith(article);

    rerender(
      <ArticleTable items={[{ ...article, archived: true }]} sort="article" direction="asc" canEdit
        onSort={vi.fn()} onView={onView} onEdit={onEdit} onArchive={onArchive} onRestore={onRestore} />
    );
    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel wiederherstellen" }));
    expect(onRestore).toHaveBeenCalledWith(expect.objectContaining({ id: "article-1", archived: true }));
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
    expect(screen.queryByRole("button", { name: /Weitere Aktionen/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Artikel ansehen/ })).toBeInTheDocument();
  });
});
