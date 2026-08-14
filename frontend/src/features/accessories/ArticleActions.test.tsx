import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { accessoryArticleFixture } from "../../test/fixtures/accessories";
import { ArticleActions } from "./ArticleActions";

const article = accessoryArticleFixture();

describe("ArticleActions", () => {
  it("routes direct view, edit, and archive actions for writers", async () => {
    const user = userEvent.setup();
    const onView = vi.fn();
    const onEdit = vi.fn();
    const onArchive = vi.fn();
    render(<ArticleActions article={article} canEdit onView={onView} onEdit={onEdit}
      onArchive={onArchive} onRestore={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" }));
    await user.click(screen.getByRole("button", { name: "Artikel bearbeiten: Gerades Modellgleis" }));
    await user.click(screen.getByRole("button", { name: "Artikel archivieren: Gerades Modellgleis" }));

    expect(onView).toHaveBeenCalledWith(article);
    expect(onEdit).toHaveBeenCalledWith(article);
    expect(onArchive).toHaveBeenCalledWith(article);
    expect(screen.queryByRole("button", { name: /Weitere Aktionen/ })).not.toBeInTheDocument();
  });

  it("routes a direct restore action for archived articles", async () => {
    const user = userEvent.setup();
    const archived = accessoryArticleFixture({ archived: true });
    const onRestore = vi.fn();
    render(<ArticleActions article={archived} canEdit onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={onRestore} />);

    await user.click(screen.getByRole("button", {
      name: "Artikel wiederherstellen: Gerades Modellgleis"
    }));
    expect(onRestore).toHaveBeenCalledWith(archived);
    expect(screen.queryByRole("button", { name: /Artikel archivieren:/ })).not.toBeInTheDocument();
  });

  it("shows direct delete only when explicitly allowed and routes the selected article", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(<ArticleActions article={article} canEdit canDelete onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} onDelete={onDelete} />);

    const deleteButton = screen.getByRole("button", { name: "Artikel löschen: Gerades Modellgleis" });
    expect(deleteButton).toHaveClass("danger");
    await user.click(deleteButton);

    expect(onDelete).toHaveBeenCalledWith(article);
  });

  it("does not expose delete to editors", () => {
    render(<ArticleActions article={article} canEdit canDelete={false} onView={vi.fn()}
      onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.queryByRole("button", { name: /Artikel löschen:/ })).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Artikel archivieren:/ })).toBeInTheDocument();
  });

  it("keeps mutation actions hidden for read-only users", () => {
    render(<ArticleActions article={article} canEdit={false} onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} onDelete={vi.fn()} />);

    expect(screen.getByRole("button", { name: /Artikel ansehen/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel bearbeiten/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel archivieren:/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel löschen:/ })).not.toBeInTheDocument();
  });
});
