import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef, FormEvent } from "react";
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
        name="gauges"
        options={options}
        value={["tt"]}
        onValueChange={onValueChange}
        placeholder="Spurweiten auswählen"
      />
    );

    const trigger = screen.getByRole("button", { name: "Spurweiten TT" });
    expect(ref.current).toBe(trigger);
    expect(trigger).toHaveTextContent("TT");
    const nativeSelect = document.querySelector("select[multiple]");
    expect(nativeSelect).toHaveAttribute("hidden");
    expect(nativeSelect).not.toHaveClass("visually-hidden");

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
        placeholder="Protokolle auswählen"
      />
    );

    const trigger = screen.getByRole("button", { name: "Protokolle TT" });
    trigger.focus();
    await user.keyboard("{ArrowDown}");
    expect(screen.queryByRole("listbox", { name: "Protokolle" })).not.toBeInTheDocument();
    expect(onValueChange).not.toHaveBeenCalled();
    expect(trigger).toHaveAttribute("aria-readonly", "true");
    expect(trigger).toHaveAttribute("aria-invalid", "true");
  });

  it("uses one roving option focus and supports listbox keyboard navigation", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();
    const onParentKeyDown = vi.fn();

    render(
      <div onKeyDown={onParentKeyDown}>
        <AppMultiSelect
          label="Spurweiten"
          options={options}
          value={[]}
          onValueChange={onValueChange}
          placeholder="Spurweiten auswählen"
        />
        <button type="button">Danach</button>
      </div>
    );

    const trigger = screen.getByRole("button", { name: "Spurweiten Spurweiten auswählen" });
    trigger.focus();
    await user.keyboard("{ArrowDown}");
    const h0 = screen.getByRole("option", { name: "H0" });
    const tt = screen.getByRole("option", { name: "TT" });
    const disabled = screen.getByRole("option", { name: "N" });
    expect(h0).toHaveFocus();
    expect(h0).toHaveAttribute("tabindex", "0");
    expect(tt).toHaveAttribute("tabindex", "-1");
    expect(disabled).toHaveAttribute("tabindex", "-1");

    await user.keyboard("{ArrowDown} ");
    expect(tt).toHaveFocus();
    expect(onValueChange).toHaveBeenCalledWith(["tt"]);
    await user.keyboard("{Home}");
    expect(h0).toHaveFocus();
    await user.keyboard("{End}");
    expect(tt).toHaveFocus();
    onParentKeyDown.mockClear();
    await user.keyboard("{Escape}");
    expect(trigger).toHaveFocus();
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(onParentKeyDown).not.toHaveBeenCalled();

    await user.keyboard("{Enter}");
    expect(screen.getByRole("option", { name: "H0" })).toHaveFocus();
    await user.tab();
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Danach" })).toHaveFocus();
  });

  it("opens and moves focus to an option when its label is typed", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();

    render(
      <AppMultiSelect
        label="Spurweite"
        options={options}
        value={[]}
        onValueChange={onValueChange}
        placeholder="Spurweite auswählen"
      />
    );

    const trigger = screen.getByRole("button", { name: "Spurweite Spurweite auswählen" });
    trigger.focus();
    await user.keyboard("tt");

    expect(screen.getByRole("listbox", { name: "Spurweite" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "TT" })).toHaveFocus();
    expect(onValueChange).not.toHaveBeenCalled();

    await user.keyboard(" ");
    expect(onValueChange).toHaveBeenCalledWith(["tt"]);
  });

  it("submits selected values and redirects required validation to the visible control", () => {
    const { container, rerender } = render(
      <form>
        <AppMultiSelect
          label="Spurweiten"
          name="gauges"
          options={options}
          value={["h0", "tt"]}
          placeholder="Spurweiten auswählen"
        />
      </form>
    );
    const form = container.querySelector("form");
    if (!form) throw new Error("expected form");
    expect(new FormData(form).getAll("gauges")).toEqual(["h0", "tt"]);

    rerender(
      <form>
        <AppMultiSelect
          label="Spurweiten"
          name="gauges"
          options={options}
          value={[]}
          placeholder="Spurweiten auswählen"
          required
        />
      </form>
    );
    const nativeSelect = container.querySelector<HTMLSelectElement>('select[name="gauges"]');
    if (!nativeSelect) throw new Error("expected native form control");
    const trigger = screen.getByRole("button", { name: "Spurweiten Spurweiten auswählen" });
    expect(trigger).toHaveAttribute("aria-invalid", "true");
    fireEvent.invalid(nativeSelect);
    expect(trigger).toHaveFocus();
  });

  it("does not constrain an empty read-only or disabled required selection", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn((event: FormEvent) => event.preventDefault());
    const { container, rerender } = render(
      <form onSubmit={onSubmit}>
        <AppMultiSelect
          label="Gauges"
          name="gauges"
          options={options}
          value={[]}
          placeholder="Select gauges"
          required
          readOnly
        />
        <button type="submit">Save</button>
      </form>
    );
    const select = container.querySelector<HTMLSelectElement>('select[name="gauges"]');
    if (!select) throw new Error("expected native form control");
    const trigger = screen.getByRole("button", { name: "Gauges Select gauges" });

    expect(select).not.toBeRequired();
    expect(select.checkValidity()).toBe(true);
    await user.click(screen.getByRole("button", { name: "Save" }));
    expect(onSubmit).toHaveBeenCalledOnce();
    expect(trigger).not.toHaveFocus();

    rerender(
      <form onSubmit={onSubmit}>
        <AppMultiSelect
          label="Gauges"
          name="gauges"
          options={options}
          value={[]}
          placeholder="Select gauges"
          required
          disabled
        />
      </form>
    );
    expect(select).not.toBeRequired();
    expect(select.checkValidity()).toBe(true);
  });
});
