import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import { ArticleColumnPicker } from "./ArticleColumnPicker";
import { defaultArticleTableColumns, type ArticleTableColumn } from "./articleTableColumns";

describe("ArticleColumnPicker", () => {
  beforeEach(() => setLanguage("de"));

  it("opens a localized checkbox popover and routes column toggles", async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();
    render(<ArticleColumnPicker visibleColumns={defaultArticleTableColumns} onToggle={onToggle} onReset={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    expect(screen.getByRole("group", { name: "Tabellenspalten" })).toBeInTheDocument();
    expect(screen.getAllByRole("checkbox")).toHaveLength(9);

    await user.click(screen.getByRole("checkbox", { name: "Hersteller" }));
    expect(onToggle).toHaveBeenCalledWith("manufacturer");
  });

  it("locks the final visible identity column", async () => {
    const user = userEvent.setup();
    const visible = new Set<ArticleTableColumn>(["name", "stock"]);
    render(<ArticleColumnPicker visibleColumns={visible} onToggle={vi.fn()} onReset={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    expect(screen.getByRole("checkbox", { name: "Name" })).toBeDisabled();
    expect(screen.getByRole("checkbox", { name: "Inventarnummer" })).not.toBeDisabled();
  });

  it("closes on Escape and restores focus to its trigger", async () => {
    const user = userEvent.setup();
    render(<ArticleColumnPicker visibleColumns={defaultArticleTableColumns} onToggle={vi.fn()} onReset={vi.fn()} />);
    const trigger = screen.getByRole("button", { name: "Tabellenspalten auswählen" });

    await user.click(trigger);
    await user.keyboard("{Escape}");

    expect(screen.queryByRole("group", { name: "Tabellenspalten" })).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("closes on an outside pointer action", async () => {
    const user = userEvent.setup();
    render(<div>
      <ArticleColumnPicker visibleColumns={defaultArticleTableColumns} onToggle={vi.fn()} onReset={vi.fn()} />
      <button type="button">Außerhalb</button>
    </div>);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    await user.click(screen.getByRole("button", { name: "Außerhalb" }));

    expect(screen.queryByRole("group", { name: "Tabellenspalten" })).not.toBeInTheDocument();
  });

  it("routes the localized reset action", async () => {
    const user = userEvent.setup();
    const onReset = vi.fn();
    render(<ArticleColumnPicker
      visibleColumns={new Set<ArticleTableColumn>(["inventoryNumber"])}
      onToggle={vi.fn()}
      onReset={onReset}
    />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    await user.click(screen.getByRole("button", { name: "Auf Standard zurücksetzen" }));

    expect(onReset).toHaveBeenCalledOnce();
  });
});
