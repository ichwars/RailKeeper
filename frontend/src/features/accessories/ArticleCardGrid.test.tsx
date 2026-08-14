import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import {
  accessoryArticleFixture,
  accessoryArticleTypes,
  accessorySubtypes
} from "../../test/fixtures/accessories";
import { ArticleCardGrid } from "./ArticleCardGrid";

describe("ArticleCardGrid", () => {
  beforeEach(() => setLanguage("de"));

  it("renders dense article identity, classification, stock, and storage", () => {
    const article = accessoryArticleFixture({
      primaryImageUrl: "/api/v1/accessory-products/article-1/image",
      locationNames: ["Werkstatt", "Vitrine"]
    });
    const { container } = render(<ArticleCardGrid items={[article]}
      articleTypeEntries={accessoryArticleTypes} subtypeEntries={accessorySubtypes}
      canEdit onView={vi.fn()} onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} />);

    const list = screen.getByRole("list");
    expect(list).toHaveClass("article-card-grid");
    expect(within(list).getByText("RK-ART-000001")).toBeInTheDocument();
    expect(within(list).getByText("Tillig")).toBeInTheDocument();
    expect(within(list).getByText("Gerades Modellgleis")).toBeInTheDocument();
    expect(within(list).getByText("83101")).toBeInTheDocument();
    expect(within(list).getByText("Gleis")).toBeInTheDocument();
    expect(within(list).getByText("Gerade")).toBeInTheDocument();
    expect(within(list).getByText("TT")).toBeInTheDocument();
    expect(within(list).getByText("18 gesamt")).toBeInTheDocument();
    expect(within(list).getByText("12 frei · 4 reserviert · 2 eingebaut")).toBeInTheDocument();
    expect(within(list).getByText("Werkstatt")).toHaveAttribute("title", "Werkstatt, Vitrine");
    expect(within(list).getByText("+ 1 weitere")).toBeInTheDocument();
    expect(container.querySelector(".article-card-media img")).toHaveAttribute(
      "src", "/api/v1/accessory-products/article-1/image"
    );
  });

  it("opens the article and falls back when no image exists", async () => {
    const user = userEvent.setup();
    const article = accessoryArticleFixture({ primaryImageUrl: undefined });
    const onView = vi.fn();
    render(<ArticleCardGrid items={[article]} articleTypeEntries={accessoryArticleTypes}
      subtypeEntries={accessorySubtypes} canEdit onView={onView} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getByText("Keine Vorschau")).toBeInTheDocument();
    await user.click(screen.getAllByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" })[0]!);
    expect(onView).toHaveBeenCalledWith(article);
  });

  it("keeps writer actions hidden for read-only users", () => {
    const article = accessoryArticleFixture();
    render(<ArticleCardGrid items={[article]} articleTypeEntries={accessoryArticleTypes}
      subtypeEntries={accessorySubtypes} canEdit={false} onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getAllByRole("button", { name: /Artikel ansehen/ }).length).toBeGreaterThan(0);
    expect(screen.queryByRole("button", { name: /Artikel bearbeiten/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Weitere Aktionen/ })).not.toBeInTheDocument();
  });
});
