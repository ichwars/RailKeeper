import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { AccessoryConfirmDialog } from "./AccessoryConfirmDialog";

describe("AccessoryConfirmDialog", () => {
  it("traps focus, cancels only itself on Escape, and restores the invoker", async () => {
    const user = userEvent.setup();
    const invoker = document.createElement("button");
    document.body.append(invoker);
    invoker.focus();
    const onClose = vi.fn();
    const action = { title: "Bestand buchen", body: "Buchung bestätigen", run: vi.fn() };
    const view = render(<AccessoryConfirmDialog action={action} onClose={onClose} />);

    const cancel = screen.getByRole("button", { name: "Abbrechen" });
    const confirm = screen.getByRole("button", { name: "Bestätigen" });
    expect(cancel).toHaveFocus();
    confirm.focus();
    await user.tab();
    expect(cancel).toHaveFocus();

    await user.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalledOnce();
    expect(action.run).not.toHaveBeenCalled();
    view.rerender(<AccessoryConfirmDialog action={null} onClose={onClose} />);
    expect(invoker).toHaveFocus();
    invoker.remove();
  });

  it("does not leak a failed command error into the next confirmation", async () => {
    const user = userEvent.setup();
    const first = { title: "Erste Aktion", body: "Fehler", run: vi.fn().mockRejectedValue(new Error("Kaputt")) };
    const second = { title: "Zweite Aktion", body: "Neu", run: vi.fn() };
    const view = render(<AccessoryConfirmDialog action={first} onClose={vi.fn()} />);
    await user.click(screen.getByRole("button", { name: "Bestätigen" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Kaputt");

    view.rerender(<AccessoryConfirmDialog action={null} onClose={vi.fn()} />);
    view.rerender(<AccessoryConfirmDialog action={second} onClose={vi.fn()} />);

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
});
