import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { setLanguage } from "../../shared/i18n";
import { ArticleColumnPicker } from "./ArticleColumnPicker";
import { defaultArticleTableColumns, type ArticleTableColumn } from "./articleTableColumns";

describe("ArticleColumnPicker", () => {
  beforeEach(() => setLanguage("de"));

  it("disables configuration until profile preferences are loaded", () => {
    render(<ArticleColumnPicker columns={defaultArticleTableColumns} loading
      onToggle={vi.fn()} onMove={vi.fn()} onReset={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Tabellenspalten auswählen" })).toBeDisabled();
  });

  it("opens a localized checkbox popover and routes column toggles", async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();
    render(<ArticleColumnPicker columns={defaultArticleTableColumns} onToggle={onToggle}
      onMove={vi.fn()} onReset={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    expect(screen.getByRole("group", { name: "Tabellenspalten" })).toBeInTheDocument();
    expect(screen.getAllByRole("checkbox")).toHaveLength(10);
    expect(screen.getByRole("checkbox", { name: "Listenpreis pro Stück" })).not.toBeChecked();

    await user.click(screen.getByRole("checkbox", { name: "Hersteller" }));
    expect(onToggle).toHaveBeenCalledWith("manufacturer");
  });

  it("prevents changes while profile preferences are loading", () => {
    render(<ArticleColumnPicker columns={defaultArticleTableColumns} loading
      onToggle={vi.fn()} onMove={vi.fn()} onReset={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Tabellenspalten auswählen" })).toBeDisabled();
  });

  it("locks the final visible identity column", async () => {
    const user = userEvent.setup();
    const columns: ArticleTableColumn[] = ["name", "stock"];
    render(<ArticleColumnPicker columns={columns} onToggle={vi.fn()} onMove={vi.fn()}
      onReset={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    expect(screen.getByRole("checkbox", { name: "Name" })).toBeDisabled();
    expect(screen.getByRole("checkbox", { name: "Inventarnummer" })).not.toBeDisabled();
  });

  it("closes on Escape and restores focus to its trigger", async () => {
    const user = userEvent.setup();
    render(<ArticleColumnPicker columns={defaultArticleTableColumns} onToggle={vi.fn()}
      onMove={vi.fn()} onReset={vi.fn()} />);
    const trigger = screen.getByRole("button", { name: "Tabellenspalten auswählen" });

    await user.click(trigger);
    await user.keyboard("{Escape}");

    expect(screen.queryByRole("group", { name: "Tabellenspalten" })).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it("closes on an outside pointer action", async () => {
    const user = userEvent.setup();
    render(<div>
      <ArticleColumnPicker columns={defaultArticleTableColumns} onToggle={vi.fn()}
        onMove={vi.fn()} onReset={vi.fn()} />
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
      columns={["inventoryNumber"]}
      onToggle={vi.fn()}
      onMove={vi.fn()}
      onReset={onReset}
    />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    await user.click(screen.getByRole("button", { name: "Auf Standard zurücksetzen" }));

    expect(onReset).toHaveBeenCalledOnce();
  });

  it("routes visible column order changes", async () => {
    const user = userEvent.setup();
    const onMove = vi.fn();
    render(<ArticleColumnPicker columns={["inventoryNumber", "name", "storage"]}
      onToggle={vi.fn()} onMove={onMove} onReset={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    await user.click(screen.getByRole("button", { name: "Lagerung nach oben" }));

    expect(onMove).toHaveBeenCalledWith("storage", "up");
  });
});
