import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { AccessoryArticleListItem, MasterDataEntry } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
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
    expect(screen.getByRole("table")).toHaveClass("article-table");
    expect(screen.getByRole("table").parentElement).toHaveClass("article-table-wrap");
    await user.click(screen.getByRole("button", { name: "Nach Bestand sortieren" }));
    expect(onSort).toHaveBeenCalledWith("stock");
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

  it("localizes canonical subtype keys returned by the backend", () => {
    render(<ArticleTable items={[{ ...article, subtype: "track:straight" }]} subtypeEntries={subtypes}
      sort="article" direction="asc" canEdit={false} onSort={vi.fn()} onView={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getByText("Gerade")).toBeInTheDocument();
    expect(screen.queryByText("track:straight")).not.toBeInTheDocument();
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

  it("moves focus into the correct row menu and returns it on Escape", async () => {
    const user = userEvent.setup();
    render(
      <ArticleTable items={[article, secondArticle]} sort="article" direction="asc" canEdit
        onSort={vi.fn()} onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />
    );
    const triggers = screen.getAllByRole("button", { name: /Weitere Aktionen/ });

    await user.click(triggers[1]!);
    const menuItem = screen.getByRole("menuitem", { name: "Artikel archivieren" });
    expect(menuItem).toHaveFocus();
    expect(menuItem).toHaveAttribute("tabindex", "0");
    expect(screen.getByRole("menu").closest(".article-table-wrap")).toHaveClass("menu-open");

    await user.keyboard("{ArrowDown}{ArrowUp}{Home}{End}");
    expect(menuItem).toHaveFocus();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(triggers[1]).toHaveFocus();
  });

  it("closes the menu on Tab or outside click without returning focus to the old trigger", async () => {
    const user = userEvent.setup();
    render(
      <div>
        <ArticleTable items={[article]} sort="article" direction="asc" canEdit onSort={vi.fn()}
          onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />
        <button type="button">Nach Tabelle</button>
      </div>
    );
    const trigger = screen.getByRole("button", { name: /Weitere Aktionen/ });

    await user.click(trigger);
    expect(screen.getByRole("menuitem")).toHaveFocus();
    await user.tab();
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(trigger).not.toHaveFocus();

    await user.click(trigger);
    expect(screen.getByRole("menuitem")).toHaveFocus();
    await user.click(screen.getByRole("button", { name: "Nach Tabelle" }));
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Nach Tabelle" })).toHaveFocus();
  });
});
