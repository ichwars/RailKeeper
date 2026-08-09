import { fireEvent, render, screen } from "@testing-library/react";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";

import { AppNumberInput } from "./AppNumberInput";

describe("AppNumberInput", () => {
  it("forwards the native number input and emits unparsed string values", () => {
    const ref = createRef<HTMLInputElement>();
    const onValueChange = vi.fn();

    render(
      <AppNumberInput
        ref={ref}
        label="Länge"
        helpText="In Millimetern"
        value="12.50"
        onValueChange={onValueChange}
      />
    );

    const input = screen.getByRole("spinbutton", { name: "Länge" });
    expect(ref.current).toBe(input);
    expect(input).toHaveValue(12.5);
    fireEvent.change(input, { target: { value: "12.75" } });
    expect(onValueChange).toHaveBeenCalledWith("12.75");
  });

  it("wires errors and read-only state without parsing at render time", () => {
    render(<AppNumberInput label="Preis" value="" readOnly error="Preis ist ungültig" />);

    const input = screen.getByRole("spinbutton", { name: "Preis" });
    expect(input).toHaveAttribute("readonly");
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input).toHaveAccessibleDescription("Preis ist ungültig");
  });
});
