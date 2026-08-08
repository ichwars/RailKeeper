import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { SettingsTabList } from "./SettingsTabList";

const options = [
  { id: "general", label: "Allgemeine Stammdaten" },
  { id: "article", label: "Artikelstammdaten" }
] as const;

describe("SettingsTabList", () => {
  const scrollIntoView = vi.fn();

  beforeEach(() => {
    scrollIntoView.mockReset();
    Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
      configurable: true,
      value: scrollIntoView
    });
  });

  it("renders one selected tab and keeps it visible", () => {
    render(
      <SettingsTabList
        ariaLabel="Datengruppe"
        options={options}
        value="general"
        onChange={vi.fn()}
      />
    );

    const tablist = screen.getByRole("tablist", { name: "Datengruppe" });
    const general = screen.getByRole("tab", { name: "Allgemeine Stammdaten" });
    expect(tablist).toContainElement(general);
    expect(general).toHaveAttribute("aria-selected", "true");
    expect(screen.getAllByRole("tab").filter((tab) => tab.tabIndex === 0)).toEqual([general]);
    expect(scrollIntoView).toHaveBeenCalledWith({ block: "nearest", inline: "nearest" });
  });

  it("moves and wraps with arrow keys", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <SettingsTabList
        ariaLabel="Datengruppe"
        options={options}
        value="general"
        onChange={onChange}
      />
    );

    const general = screen.getByRole("tab", { name: "Allgemeine Stammdaten" });
    const article = screen.getByRole("tab", { name: "Artikelstammdaten" });
    general.focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenLastCalledWith("article");
    expect(article).toHaveFocus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenLastCalledWith("general");
    expect(general).toHaveFocus();
    await user.keyboard("{ArrowLeft}");
    expect(onChange).toHaveBeenLastCalledWith("article");
    expect(article).toHaveFocus();
  });

  it("supports Home and End", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <SettingsTabList
        ariaLabel="Datengruppe"
        options={options}
        value="article"
        onChange={onChange}
      />
    );

    const general = screen.getByRole("tab", { name: "Allgemeine Stammdaten" });
    const article = screen.getByRole("tab", { name: "Artikelstammdaten" });
    article.focus();
    await user.keyboard("{Home}");
    expect(onChange).toHaveBeenLastCalledWith("general");
    expect(general).toHaveFocus();
    await user.keyboard("{End}");
    expect(onChange).toHaveBeenLastCalledWith("article");
    expect(article).toHaveFocus();
  });
});
