import { useRef, useState } from "react";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AppSelect } from "./AppSelect";

function VehicleSelection({ onEscape, containMenu = false }: { onEscape: () => void; containMenu?: boolean }) {
  const [value, setValue] = useState("one");
  const dialogRef = useRef<HTMLDivElement | null>(null);
  return <div ref={dialogRef} role="dialog" onKeyDown={(event) => {
    if (event.key === "Escape") onEscape();
  }}>
    <label>Fahrzeug<AppSelect value={value} portalContainerRef={containMenu ? dialogRef : undefined}
      onChange={(event) => setValue(event.target.value)}>
      <option value="one">RK-1 · Lok</option>
      <option value="two">RK-2 · Wagen</option>
    </AppSelect></label>
  </div>;
}

describe("AppSelect in a dialog", () => {
  it("renders the menu in its supplied dialog layer", async () => {
    const user = userEvent.setup();
    render(<VehicleSelection containMenu onEscape={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: "Fahrzeug" }));
    expect(screen.getByRole("dialog")).toContainElement(screen.getByRole("listbox"));
    await user.click(screen.getByRole("option", { name: "RK-2 · Wagen" }));
    expect(screen.getByRole("button", { name: "Fahrzeug" })).toHaveTextContent("RK-2 · Wagen");
  });

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

describe("AppSelect keyboard and option safety", () => {
  function Selection({ disabled = false }: { disabled?: boolean }) {
    const [value, setValue] = useState("roco");
    return <label>Manufacturer<AppSelect value={value} disabled={disabled} onChange={event => setValue(event.target.value)}>
      <><option value="disabled" disabled>Restricted</option><option value="roco">Roco</option></>
      <option value="riv">Rivarossi</option><option value="esu"><span>ESU</span> LokPilot</option>
    </AppSelect></label>;
  }

  it("skips disabled options, supports boundaries and commits only on Enter", async () => {
    const user = userEvent.setup(); render(<Selection />);
    const trigger = screen.getByRole("button", { name: "Manufacturer" });
    trigger.focus(); await user.keyboard("{ArrowDown}{End}");
    expect(trigger).toHaveTextContent("Roco");
    await user.keyboard("{Enter}");
    expect(trigger).toHaveTextContent("ESU LokPilot");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    await user.keyboard("{Home}{Enter}");
    expect(trigger).toHaveTextContent("Roco");
    await user.keyboard("{ArrowUp}{Enter}");
    expect(trigger).toHaveTextContent("ESU LokPilot");
  });

  it("cycles repeated typeahead letters without selecting disabled matching options", async () => {
    const user = userEvent.setup(); render(<Selection />);
    const trigger = screen.getByRole("button", { name: "Manufacturer" });
    trigger.focus(); await user.keyboard("r");
    expect(trigger).toHaveTextContent("Rivarossi");
    await user.keyboard("r");
    expect(trigger).toHaveTextContent("Roco");
  });

  it("finds a substring inside a rich option label and ignores accelerator keys", async () => {
    const user = userEvent.setup(); render(<Selection />);
    const trigger = screen.getByRole("button", { name: "Manufacturer" });
    trigger.focus(); await user.keyboard("pilot");
    expect(trigger).toHaveTextContent("ESU LokPilot");
    await user.keyboard("{Control>}r{/Control}");
    expect(trigger).toHaveTextContent("ESU LokPilot");
  });

  it("opens with Space and closes on an outside pointer without changing selection", async () => {
    const user = userEvent.setup(); render(<><Selection /><button>Outside</button></>);
    const trigger = screen.getByRole("button", { name: "Manufacturer" });
    trigger.focus(); await user.keyboard(" ");
    expect(screen.getByRole("listbox")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Outside" }));
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(trigger).toHaveTextContent("Roco");
  });

  it("keeps disabled and empty selectors inert", async () => {
    const change = vi.fn(), user = userEvent.setup();
    const { rerender } = render(<AppSelect aria-label="Empty" value="" onChange={change}>Unusable text</AppSelect>);
    await user.click(screen.getByRole("button", { name: "Empty" }));
    await user.keyboard("{ArrowDown}{ArrowUp}{Home}{End}{Enter}z");
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
    expect(change).not.toHaveBeenCalled();
    rerender(<Selection disabled />);
    expect(screen.getByRole("button", { name: "Manufacturer" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Manufacturer" }));
    expect(screen.queryByRole("listbox")).not.toBeInTheDocument();
  });
});
