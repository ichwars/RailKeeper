import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { accessoryArticleFixture } from "../../test/fixtures/accessories";
import { ArticleActions } from "./ArticleActions";

const article = accessoryArticleFixture();

describe("ArticleActions", () => {
  it("routes view, edit, and archive for writers", async () => {
    const user = userEvent.setup();
    const onView = vi.fn();
    const onEdit = vi.fn();
    const onArchive = vi.fn();
    render(<ArticleActions article={article} canEdit onView={onView} onEdit={onEdit}
      onArchive={onArchive} onRestore={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Artikel ansehen: Gerades Modellgleis" }));
    await user.click(screen.getByRole("button", { name: "Artikel bearbeiten: Gerades Modellgleis" }));
    await user.click(screen.getByRole("button", { name: "Weitere Aktionen: Gerades Modellgleis" }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel archivieren" }));

    expect(onView).toHaveBeenCalledWith(article);
    expect(onEdit).toHaveBeenCalledWith(article);
    expect(onArchive).toHaveBeenCalledWith(article);
  });

  it("routes restore for archived articles", async () => {
    const user = userEvent.setup();
    const archived = accessoryArticleFixture({ archived: true });
    const onRestore = vi.fn();
    render(<ArticleActions article={archived} canEdit onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={onRestore} />);

    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel wiederherstellen" }));
    expect(onRestore).toHaveBeenCalledWith(archived);
  });

  it("shows delete only when explicitly allowed and routes the selected article", async () => {
    const user = userEvent.setup();
    const onDelete = vi.fn();
    render(<ArticleActions article={article} canEdit canDelete onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} onDelete={onDelete} />);

    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    await user.click(screen.getByRole("menuitem", { name: "Artikel löschen" }));

    expect(onDelete).toHaveBeenCalledWith(article);
  });

  it("does not expose delete to editors", async () => {
    const user = userEvent.setup();
    render(<ArticleActions article={article} canEdit canDelete={false} onView={vi.fn()}
      onEdit={vi.fn()} onArchive={vi.fn()} onRestore={vi.fn()} onDelete={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Weitere Aktionen/ }));
    expect(screen.queryByRole("menuitem", { name: "Artikel löschen" })).not.toBeInTheDocument();
  });

  it("moves focus through the menu and restores it on Escape", async () => {
    const user = userEvent.setup();
    render(<ArticleActions article={article} canEdit canDelete onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} onDelete={vi.fn()} />);
    const trigger = screen.getByRole("button", { name: "Weitere Aktionen: Gerades Modellgleis" });

    await user.click(trigger);
    const [archiveItem, deleteItem] = screen.getAllByRole("menuitem");
    expect(archiveItem).toHaveFocus();
    await user.keyboard("{ArrowDown}");
    expect(deleteItem).toHaveFocus();
    await user.keyboard("{ArrowDown}");
    expect(archiveItem).toHaveFocus();
    await user.keyboard("{End}");
    expect(deleteItem).toHaveFocus();
    await user.keyboard("{Home}");
    expect(archiveItem).toHaveFocus();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("closes on Tab and outside click without restoring the trigger", async () => {
    const user = userEvent.setup();
    render(<div>
      <ArticleActions article={article} canEdit onView={vi.fn()} onEdit={vi.fn()}
        onArchive={vi.fn()} onRestore={vi.fn()} />
      <button type="button">Nach Aktionen</button>
    </div>);
    const trigger = screen.getByRole("button", { name: /Weitere Aktionen/ });

    await user.click(trigger);
    await user.tab();
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(trigger).not.toHaveFocus();

    await user.click(trigger);
    await user.click(screen.getByRole("button", { name: "Nach Aktionen" }));
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Nach Aktionen" })).toHaveFocus();
  });

  it("keeps mutation actions hidden for read-only users", () => {
    render(<ArticleActions article={article} canEdit={false} onView={vi.fn()} onEdit={vi.fn()}
      onArchive={vi.fn()} onRestore={vi.fn()} />);

    expect(screen.getByRole("button", { name: /Artikel ansehen/ })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Artikel bearbeiten/ })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /Weitere Aktionen/ })).not.toBeInTheDocument();
  });
});
