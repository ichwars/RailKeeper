import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import {
  accessoryArticleFixture,
  accessoryArticleTypes,
  accessorySubtypes
} from "../../test/fixtures/accessories";
import { ArticleCompactList } from "./ArticleCompactList";

describe("ArticleCompactList", () => {
  beforeEach(() => setLanguage("de"));

  it("renders compact article identity, classification, gauge, and stock", () => {
    const article = accessoryArticleFixture({ primaryImageUrl: "/api/v1/accessory-products/article-1/image" });
    const { container } = render(<ArticleCompactList items={[article]}
      articleTypeEntries={accessoryArticleTypes} subtypeEntries={accessorySubtypes}
      canEdit onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);

    const list = screen.getByRole("list");
    expect(list).toHaveClass("article-mobile-list");
    expect(within(list).getByText("RK-ART-000001")).toBeInTheDocument();
    expect(within(list).getByText("Gerades Modellgleis")).toBeInTheDocument();
    expect(within(list).getByText("Tillig · 83101")).toBeInTheDocument();
    expect(within(list).getByText("TT")).toBeInTheDocument();
    expect(within(list).getByText("Gerade")).toBeInTheDocument();
    expect(within(list).getByText("18 gesamt")).toBeInTheDocument();
    expect(container.querySelector(".article-mobile-media img")).toHaveAttribute(
      "src", "/api/v1/accessory-products/article-1/image"
    );
  });

  it("opens the article and shows the image fallback", async () => {
    const user = userEvent.setup();
    const article = accessoryArticleFixture({ primaryImageUrl: undefined });
    const onView = vi.fn();
    render(<ArticleCompactList items={[article]} articleTypeEntries={accessoryArticleTypes}
      subtypeEntries={accessorySubtypes} canEdit onView={onView} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getByText("Keine Vorschau")).toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" })[0]!);
    expect(onView).toHaveBeenCalledWith(article);
  });

  it("keeps mutation actions hidden for read-only users", () => {
    render(<ArticleCompactList items={[accessoryArticleFixture()]}
      articleTypeEntries={accessoryArticleTypes} subtypeEntries={accessorySubtypes}
      canEdit={false} onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getAllByRole("button", { name: /Artikel ansehen/ }).length).toBeGreaterThan(0);
    expect(screen.queryByRole("button", { name: /Artikel bearbeiten/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel archivieren:/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel löschen:/ })).not.toBeInTheDocument();
  });

  it("renders archive and admin delete as direct compact actions", () => {
    const article = accessoryArticleFixture();
    render(<ArticleCompactList items={[article]} articleTypeEntries={accessoryArticleTypes}
      subtypeEntries={accessorySubtypes} canEdit canDelete onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Artikel archivieren: Gerades Modellgleis" }))
      .toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Artikel löschen: Gerades Modellgleis" }))
      .toBeInTheDocument();
  });
});
