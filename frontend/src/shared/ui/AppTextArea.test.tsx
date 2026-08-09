import { render, screen } from "@testing-library/react";
import { createRef } from "react";
import { describe, expect, it } from "vitest";

import { AppTextArea } from "./AppTextArea";

describe("AppTextArea", () => {
  it("connects its label, help, error, invalid state, and forwarded ref", () => {
    const ref = createRef<HTMLTextAreaElement>();

    render(
      <AppTextArea
        ref={ref}
        label="Beschreibung"
        helpText="Interne Notiz"
        error="Beschreibung fehlt"
        value=""
        onChange={() => undefined}
      />
    );

    const field = screen.getByRole("textbox", { name: "Beschreibung" });
    expect(ref.current).toBe(field);
    expect(field).toHaveAttribute("aria-invalid", "true");
    expect(field).toHaveAccessibleDescription("Interne Notiz Beschreibung fehlt");
    expect(screen.getByRole("alert")).toHaveTextContent("Beschreibung fehlt");
  });

  it("preserves native disabled and read-only text semantics", () => {
    const { rerender } = render(<AppTextArea label="Beschreibung" disabled value="Text" />);
    expect(screen.getByRole("textbox", { name: "Beschreibung" })).toBeDisabled();

    rerender(<AppTextArea label="Beschreibung" readOnly value="Text" />);
    expect(screen.getByRole("textbox", { name: "Beschreibung" })).toHaveAttribute("readonly");
  });
});
