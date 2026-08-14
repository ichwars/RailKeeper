import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef } from "react";
import { describe, expect, it, vi } from "vitest";

import { AppCheckbox } from "./AppCheckbox";

describe("AppCheckbox", () => {
  it("connects its label, checked state, change event, and forwarded ref", async () => {
    const user = userEvent.setup();
    const ref = createRef<HTMLInputElement>();
    const onChange = vi.fn();

    render(<AppCheckbox ref={ref} label="Archiviert" checked={false} onChange={onChange} />);

    const checkbox = screen.getByRole("checkbox", { name: "Archiviert" });
    expect(ref.current).toBe(checkbox);
    expect(checkbox).not.toBeChecked();
    await user.click(checkbox);
    expect(onChange).toHaveBeenCalledTimes(1);
  });

  it("preserves native disabled and required semantics", () => {
    render(<AppCheckbox label="Freigabe" checked={false} disabled required readOnly />);

    const checkbox = screen.getByRole("checkbox", { name: "Freigabe" });
    expect(checkbox).toBeDisabled();
    expect(checkbox).toBeRequired();
  });
});
