import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createRef, useState } from "react";
import { describe, expect, it, vi } from "vitest";
import { AppNumberControl } from "./AppNumberControl";

describe("AppNumberControl", () => {
  it("steps within native bounds and emits exactly one original change event", async () => {
    const changed = vi.fn();
    function Form() {
      const [value, setValue] = useState("0");
      return <label>CV31<AppNumberControl aria-label="CV31" value={value} min={0} max={1}
        onChange={event => { changed(event.target.value); setValue(event.target.value); }} /></label>;
    }
    render(<Form />);
    const up = screen.getByRole("button", { name: /Wert erhöhen|Increase value/ });
    const down = screen.getByRole("button", { name: /Wert verringern|Decrease value/ });
    await userEvent.click(up);
    expect(screen.getByRole("spinbutton")).toHaveValue(1);
    expect(changed).toHaveBeenLastCalledWith("1");
    await userEvent.click(up);
    expect(changed).toHaveBeenCalledTimes(1);
    await userEvent.click(down);
    expect(changed).toHaveBeenLastCalledWith("0");
    fireEvent.change(screen.getByRole("spinbutton"), { target: { value: "" } });
    expect(changed).toHaveBeenLastCalledWith("");
    expect(screen.getByRole("spinbutton")).toHaveFocus();
  });

  it("preserves raw decimal strings, forwarded refs and free decimal entry", () => {
    const ref = createRef<HTMLInputElement>();
    const changed = vi.fn();
    render(<AppNumberControl ref={ref} aria-label="Exact" step="any" value="12.500001"
      onChange={event => changed(event.target.value)} />);
    expect(ref.current).toBe(screen.getByRole("spinbutton", { name: "Exact" }));
    fireEvent.change(ref.current!, { target: { value: "12.500002" } });
    expect(changed).toHaveBeenCalledWith("12.500002");
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it.each([{ disabled: true }, { readOnly: true }])("prevents step changes for unavailable fields %o", props => {
    render(<AppNumberControl {...props} aria-label="Protected" defaultValue="3" />);
    expect(screen.getAllByRole("button").every(button => button.hasAttribute("disabled"))).toBe(true);
    expect(screen.getByRole("spinbutton")).toHaveValue(3);
  });
});
