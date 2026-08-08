import { useState } from "react";
import { render, screen, within } from "@testing-library/react";
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

  it("closes a committed command before a failed resource refresh so it cannot run twice", async () => {
    const user = userEvent.setup();
    const run = vi.fn().mockResolvedValue(undefined);
    const afterSuccess = vi.fn().mockRejectedValue(new Error("Aktualisierung fehlgeschlagen"));
    function Harness() {
      const [open, setOpen] = useState(true);
      return <AccessoryConfirmDialog action={open ? {
        title: "Bestand buchen", body: "Buchung bestätigen", run, afterSuccess
      } : null} onClose={() => setOpen(false)} />;
    }
    render(<Harness />);

    await user.click(screen.getByRole("button", { name: "Bestätigen" }));

    expect(run).toHaveBeenCalledOnce();
    expect(afterSuccess).toHaveBeenCalledOnce();
    expect(screen.queryByRole("dialog", { name: "Bestand buchen" })).not.toBeInTheDocument();
  });

  it("renders structured confirmation content without nesting paragraphs", () => {
    const view = render(<AccessoryConfirmDialog action={{
      title: "Mögliche Dublette",
      body: <><p>Treffer prüfen</p><ul><li>Tillig 83101</li></ul></>,
      run: vi.fn()
    }} onClose={vi.fn()} />);

    expect(document.body.querySelector("p p")).toBeNull();
    expect(screen.getByText("Tillig 83101")).toBeInTheDocument();
  });

  it("ports nested confirmations outside and makes only the child modal accessible", async () => {
    const user = userEvent.setup();
    function Harness() {
      const [open, setOpen] = useState(false);
      return <div role="dialog" aria-modal="true" aria-label="Artikel bearbeiten">
        <button type="button" onClick={() => setOpen(true)}>Aktion öffnen</button>
        <AccessoryConfirmDialog action={open ? {
          title: "Aktion bestätigen", body: "Wirklich ausführen?", run: vi.fn()
        } : null} onClose={() => setOpen(false)} />
      </div>;
    }
    render(<Harness />);
    const parent = screen.getByRole("dialog", { name: "Artikel bearbeiten" });
    const invoker = screen.getByRole("button", { name: "Aktion öffnen" });

    await user.click(invoker);

    const child = screen.getByRole("dialog", { name: "Aktion bestätigen" });
    expect(parent).toHaveAttribute("aria-hidden", "true");
    expect(parent).toHaveAttribute("inert");
    expect(parent).not.toContainElement(child);
    expect(within(child).getByRole("button", { name: "Abbrechen" })).toHaveFocus();

    await user.keyboard("{Escape}");
    expect(parent).not.toHaveAttribute("aria-hidden");
    expect(parent).not.toHaveAttribute("inert");
    expect(invoker).toHaveFocus();
  });
});
