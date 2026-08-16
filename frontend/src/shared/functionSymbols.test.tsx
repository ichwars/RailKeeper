import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { MasterDataEntry } from "./api";
import { FunctionSymbolPicker, functionSymbolMetadata } from "./functionSymbols";

function symbol(key: string, label: string, active: boolean): MasterDataEntry {
  return {
    id: `symbols:${key}`,
    type: "symbols",
    key,
    label,
    active,
    sortOrder: 0,
    metadata: { marker: key },
    createdAt: "2026-08-16T00:00:00Z",
    updatedAt: "2026-08-16T00:00:00Z"
  };
}

describe("FunctionSymbolPicker", () => {
  it("keeps only the current inactive symbol and does not resurrect fallback symbols", async () => {
    const user = userEvent.setup();
    const symbols = [
      symbol("light", "Licht", true),
      symbol("sound", "Sound", false),
      symbol("horn", "Horn", false)
    ];
    render(
      <FunctionSymbolPicker
        value="sound"
        symbols={symbols}
        label="Funktionssymbol"
        onChange={vi.fn()}
      />
    );

    await user.click(screen.getByLabelText("Funktionssymbol"));
    expect(screen.getByRole("button", { name: "Sound (inaktiv)" })).toBeDisabled();
    expect(screen.getByRole("button", { name: "Licht" })).toBeEnabled();
    expect(screen.queryByRole("button", { name: "Horn" })).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Kupplung" })).not.toBeInTheDocument();
    expect(functionSymbolMetadata(symbols, "sound")).toEqual({ marker: "sound" });
  });

  it("shows an unmatched current key as inactive", async () => {
    const user = userEvent.setup();
    render(
      <FunctionSymbolPicker
        value="legacy"
        symbols={[symbol("light", "Licht", true)]}
        label="Funktionssymbol"
        onChange={vi.fn()}
      />
    );

    await user.click(screen.getByLabelText("Funktionssymbol"));
    expect(screen.getByRole("button", { name: "legacy (inaktiv)" })).toBeDisabled();
  });
});
