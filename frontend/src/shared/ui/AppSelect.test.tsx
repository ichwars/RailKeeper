import { useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AppSelect } from "./AppSelect";

function VehicleSelection({ onEscape }: { onEscape: () => void }) {
  const [value, setValue] = useState("one");
  return <div role="dialog" onKeyDown={(event) => {
    if (event.key === "Escape") onEscape();
  }}>
    <label>Fahrzeug<AppSelect value={value} onChange={(event) => setValue(event.target.value)}>
      <option value="one">RK-1 · Lok</option>
      <option value="two">RK-2 · Wagen</option>
    </AppSelect></label>
  </div>;
}

describe("AppSelect in a dialog", () => {
  it("returns focus from a floating option to the vehicle selector", async () => {
    const user = userEvent.setup();
    render(<VehicleSelection onEscape={vi.fn()} />);
    const trigger = screen.getByRole("button", { name: "Fahrzeug" });
    await user.click(trigger);
    await user.click(screen.getByRole("option", { name: "RK-2 · Wagen" }));
    expect(trigger).toHaveTextContent("RK-2 · Wagen");
    expect(trigger).toHaveFocus();
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });

  it("closes the open selector before allowing Escape to reach the parent dialog", async () => {
    const user = userEvent.setup();
    const onEscape = vi.fn();
    render(<VehicleSelection onEscape={onEscape} />);
    const trigger = screen.getByRole("button", { name: "Fahrzeug" });
    await user.click(trigger);
    screen.getByRole("option", { name: "RK-2 · Wagen" }).focus();
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
    expect(onEscape).not.toHaveBeenCalled();
    await user.keyboard("{Escape}");
    expect(onEscape).toHaveBeenCalledOnce();
  });
});
