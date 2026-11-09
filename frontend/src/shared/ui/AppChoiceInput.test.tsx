import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";
import { AppChoiceInput } from "./AppChoiceInput";

describe("AppChoiceInput", () => {
  it("preserves partial table selection, labels, keyboard input and the native ref", async () => {
    const ref = createRef<HTMLInputElement>();
    const change = vi.fn();
    render(<label>Select rows<AppChoiceInput type="checkbox" ref={ref} onChange={change} /></label>);
    const input = screen.getByRole("checkbox", { name: "Select rows" });
    expect(ref.current).toBe(input);
    ref.current!.indeterminate = true;
    expect(input).toBePartiallyChecked();
    input.focus();
    await userEvent.keyboard(" ");
    expect(input).toBeChecked();
    expect(input).not.toBePartiallyChecked();
    expect(change).toHaveBeenCalledTimes(1);
  });

  it("keeps a radio group exclusive and submits its selected value", async () => {
    const { container } = render(<form>
      <label>All<AppChoiceInput type="radio" name="scope" value="all" defaultChecked /></label>
      <label>Selected<AppChoiceInput type="radio" name="scope" value="selected" /></label>
      <AppChoiceInput type="checkbox" disabled aria-label="Unavailable" />
    </form>);
    await userEvent.click(screen.getByRole("radio", { name: "Selected" }));
    expect(screen.getByRole("radio", { name: "All" })).not.toBeChecked();
    expect(new FormData(container.querySelector("form")!).get("scope")).toBe("selected");
    expect(screen.getByRole("checkbox", { name: "Unavailable" })).toBeDisabled();
  });
});
