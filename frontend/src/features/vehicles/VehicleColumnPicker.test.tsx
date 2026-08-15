import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { languageSettingKey } from "../../shared/i18n";
import { VehicleColumnPicker } from "./VehicleColumnPicker";

describe("VehicleColumnPicker", () => {
  afterEach(() => {
    window.localStorage.removeItem(languageSettingKey);
  });

  it("toggles, moves, and resets columns", async () => {
    window.localStorage.setItem(languageSettingKey, "de");
    const user = userEvent.setup();
    const onToggle = vi.fn();
    const onMove = vi.fn();
    const onReset = vi.fn();
    render(
      <VehicleColumnPicker
        columns={["inventoryNumber", "series"]}
        loading={false}
        onToggle={onToggle}
        onMove={onMove}
        onReset={onReset}
      />
    );

    await user.click(screen.getByRole("button", { name: "Tabellenspalten auswählen" }));
    await user.click(screen.getByRole("checkbox", { name: "Baureihe" }));
    expect(onToggle).toHaveBeenCalledWith("series");

    await user.click(screen.getByRole("button", { name: "Baureihe nach oben" }));
    expect(onMove).toHaveBeenCalledWith("series", "up");

    await user.click(screen.getByRole("button", { name: "Auf Standard zurücksetzen" }));
    expect(onReset).toHaveBeenCalledOnce();
  });

  it("closes with Escape and restores focus", async () => {
    const user = userEvent.setup();
    render(
      <VehicleColumnPicker
        columns={["inventoryNumber"]}
        loading={false}
        onToggle={vi.fn()}
        onMove={vi.fn()}
        onReset={vi.fn()}
      />
    );

    const trigger = screen.getByRole("button", { name: "Tabellenspalten auswählen" });
    await user.click(trigger);
    await user.keyboard("{Escape}");

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });
});
