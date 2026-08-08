import { render, screen } from "@testing-library/react";
import { createRef } from "react";
import { describe, expect, it } from "vitest";

import { AppTextInput } from "./AppTextInput";

describe("AppTextInput", () => {
  it("connects its label, help, error, invalid state, and forwarded ref", () => {
    const ref = createRef<HTMLInputElement>();

    render(
      <AppTextInput
        ref={ref}
        label="Artikelnummer"
        helpText="Nummer des Herstellers"
        error="Artikelnummer fehlt"
        value=""
        onChange={() => undefined}
      />
    );

    const input = screen.getByRole("textbox", { name: "Artikelnummer" });
    expect(ref.current).toBe(input);
    expect(input).toHaveAttribute("aria-invalid", "true");
    expect(input).toHaveAccessibleDescription("Nummer des Herstellers Artikelnummer fehlt");
    expect(screen.getByRole("alert")).toHaveTextContent("Artikelnummer fehlt");
  });

  it("preserves native disabled and read-only text semantics", () => {
    const { rerender } = render(<AppTextInput label="Name" disabled value="Gleis" />);
    expect(screen.getByRole("textbox", { name: "Name" })).toBeDisabled();

    rerender(<AppTextInput label="Name" readOnly value="Gleis" />);
    expect(screen.getByRole("textbox", { name: "Name" })).toHaveAttribute("readonly");
  });
});
