import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";

import { AppMultiSelect } from "./AppMultiSelect";

const options = [
  { value: "h0", label: "H0" },
  { value: "tt", label: "TT" },
  { value: "n", label: "N", disabled: true }
];

describe("AppMultiSelect", () => {
  it("renders an app-owned listbox and reports selected values accessibly", async () => {
    const user = userEvent.setup();
    const ref = createRef<HTMLButtonElement>();
    const onValueChange = vi.fn();

    render(
      <AppMultiSelect
        ref={ref}
        label="Spurweiten"
        helpText="Mehrere Werte möglich"
        options={options}
        value={["tt"]}
        onValueChange={onValueChange}
      />
    );

    const trigger = screen.getByRole("button", { name: "Spurweiten" });
    expect(ref.current).toBe(trigger);
    expect(trigger).toHaveTextContent("TT");
    expect(document.querySelector("select[multiple]")).not.toBeInTheDocument();

    await user.click(trigger);
    const listbox = screen.getByRole("listbox", { name: "Spurweiten" });
    expect(listbox).toHaveAttribute("aria-multiselectable", "true");
    expect(screen.getByRole("option", { name: "TT" })).toHaveAttribute("aria-selected", "true");
    await user.click(screen.getByRole("option", { name: "H0" }));
    expect(onValueChange).toHaveBeenCalledWith(["tt", "h0"]);
  });

  it("opens from the keyboard and prevents changes while read-only", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();

    render(
      <AppMultiSelect
        label="Protokolle"
        options={options}
        value={["tt"]}
        onValueChange={onValueChange}
        readOnly
        error="Auswahl prüfen"
      />
    );

    const trigger = screen.getByRole("button", { name: "Protokolle" });
    trigger.focus();
    await user.keyboard("{ArrowDown}");
    expect(screen.getByRole("listbox", { name: "Protokolle" })).toBeVisible();
    await user.click(screen.getByRole("option", { name: "H0" }));
    expect(onValueChange).not.toHaveBeenCalled();
    expect(trigger).toHaveAttribute("aria-readonly", "true");
    expect(trigger).toHaveAttribute("aria-invalid", "true");
  });
});
