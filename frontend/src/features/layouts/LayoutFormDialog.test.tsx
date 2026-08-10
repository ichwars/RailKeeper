import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { describe, expect, it, vi } from "vitest";

import { LayoutFormDialog, type LayoutFormValue } from "./LayoutFormDialog";

const initialValue: LayoutFormValue = {
  name: "meine",
  kind: "private",
  gauge: "TT",
  scale: "1:120",
  maxGradePercent: "",
  minimumTrackClearanceMm: "",
  description: "",
  archived: false
};

function DirtyDialogHarness({ returnFocusTo, onClose }: {
  returnFocusTo: HTMLElement;
  onClose: () => void;
}) {
  const [open, setOpen] = useState(true);
  if (!open) return null;
  return <LayoutFormDialog mode="edit" initialValue={initialValue} saving={false} message=""
    conflict={false} returnFocusTo={returnFocusTo} onSubmit={() => undefined} onClose={() => {
      onClose();
      setOpen(false);
    }} />;
}

describe("LayoutFormDialog", () => {
  it("uses RailKeeper controls, focuses the name, and submits a create draft", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();

    render(
      <LayoutFormDialog
        mode="create"
        initialValue={initialValue}
        saving={false}
        message=""
        conflict={false}
        onSubmit={onSubmit}
        onClose={() => undefined}
      />
    );

    const dialog = screen.getByRole("dialog", { name: "Anlage anlegen" });
    const name = within(dialog).getByRole("textbox", { name: "Bezeichnung" });
    expect(name).toHaveFocus();
    expect(name.closest(".app-text-input")).not.toBeNull();
    expect(within(dialog).getByRole("textbox", { name: "Beschreibung" }).closest(".app-text-area")).not.toBeNull();
    expect(dialog.querySelector("select")).toBeNull();
    expect(within(dialog).queryByRole("checkbox", { name: "Archiviert" })).not.toBeInTheDocument();

    await user.clear(name);
    await user.type(name, "Heimanlage");
    await user.click(within(dialog).getByRole("button", { name: "Anlage speichern" }));

    expect(onSubmit).toHaveBeenCalledWith({ ...initialValue, name: "Heimanlage" });
  });

  it("asks before discarding a dirty draft and restores focus", async () => {
    const trigger = document.createElement("button");
    trigger.textContent = "Bearbeiten";
    document.body.append(trigger);
    trigger.focus();
    const onClose = vi.fn();
    render(<DirtyDialogHarness returnFocusTo={trigger} onClose={onClose} />);

    const description = screen.getByRole("textbox", { name: "Beschreibung" });
    fireEvent.change(description, { target: { value: "Geändert" } });
    fireEvent.keyDown(description, { key: "Escape" });

    const confirm = screen.getByRole("dialog", { name: "Änderungen verwerfen?" });
    expect(confirm).toHaveTextContent("nicht gespeicherten Änderungen");
    fireEvent.click(within(confirm).getByRole("button", { name: "Verwerfen" }));
    await waitFor(() => expect(onClose).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(screen.queryByRole("dialog", { name: "Änderungen verwerfen?" }))
      .not.toBeInTheDocument());

    await waitFor(() => expect(trigger).toHaveFocus());
    trigger.remove();
  });

  it("keeps the draft visible when saving fails and offers conflict reload", async () => {
    const user = userEvent.setup();
    const onReloadConflict = vi.fn();
    const view = render(
      <LayoutFormDialog
        mode="edit"
        initialValue={initialValue}
        saving={false}
        message=""
        conflict={false}
        onSubmit={() => undefined}
        onClose={() => undefined}
      />
    );
    const name = screen.getByRole("textbox", { name: "Bezeichnung" });
    await user.clear(name);
    await user.type(name, "Lokaler Entwurf");

    view.rerender(
      <LayoutFormDialog
        mode="edit"
        initialValue={initialValue}
        saving={false}
        message="Die Anlage wurde zwischenzeitlich geändert. Deine Eingaben bleiben erhalten."
        conflict
        onReloadConflict={onReloadConflict}
        onSubmit={() => undefined}
        onClose={() => undefined}
      />
    );

    expect(screen.getByRole("alert")).toHaveTextContent("zwischenzeitlich geändert");
    expect(name).toHaveValue("Lokaler Entwurf");
    await user.click(screen.getByRole("button", { name: "Serverstand neu laden" }));
    expect(onReloadConflict).toHaveBeenCalledTimes(1);
  });

  it("shows the app-owned archive checkbox only while editing", () => {
    render(
      <LayoutFormDialog
        mode="edit"
        initialValue={initialValue}
        saving={false}
        message=""
        conflict={false}
        onSubmit={() => undefined}
        onClose={() => undefined}
      />
    );

    expect(screen.getByRole("checkbox", { name: "Archiviert" }).closest(".app-checkbox")).not.toBeNull();
  });

  it("uses the app-owned number input and submits an optional maximum grade", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LayoutFormDialog mode="create" initialValue={initialValue} saving={false} message=""
      conflict={false} onSubmit={onSubmit} onClose={() => undefined} />);

    const grade = screen.getByRole("spinbutton", { name: "Maximale Steigung (%)" });
    expect(grade.closest(".app-number-input")).not.toBeNull();
    await user.type(grade, "3.5");
    await user.click(screen.getByRole("button", { name: "Anlage speichern" }));

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ maxGradePercent: "3.5" }));
  });

  it("rejects a maximum grade outside the supported range", async () => {
    const user = userEvent.setup();
    render(<LayoutFormDialog mode="create" initialValue={initialValue} saving={false} message=""
      conflict={false} onSubmit={() => undefined} onClose={() => undefined} />);

    const grade = screen.getByRole("spinbutton", { name: "Maximale Steigung (%)" });
    await user.type(grade, "101");

    expect(screen.getByText("Bitte einen Wert über 0 bis einschließlich 100 eingeben.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Anlage speichern" })).toBeDisabled();
  });

  it("uses the app-owned number input and submits an optional track clearance", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<LayoutFormDialog mode="create" initialValue={initialValue} saving={false} message=""
      conflict={false} onSubmit={onSubmit} onClose={() => undefined} />);

    const clearance = screen.getByRole("spinbutton", { name: "Mindestabstand kreuzender Gleise (mm)" });
    expect(clearance.closest(".app-number-input")).not.toBeNull();
    await user.type(clearance, "40");
    await user.click(screen.getByRole("button", { name: "Anlage speichern" }));

    expect(onSubmit).toHaveBeenCalledWith(expect.objectContaining({ minimumTrackClearanceMm: "40" }));
  });

  it("rejects a non-positive track clearance", async () => {
    const user = userEvent.setup();
    render(<LayoutFormDialog mode="create" initialValue={initialValue} saving={false} message=""
      conflict={false} onSubmit={() => undefined} onClose={() => undefined} />);

    const clearance = screen.getByRole("spinbutton", { name: "Mindestabstand kreuzender Gleise (mm)" });
    await user.type(clearance, "0");

    expect(screen.getByText("Bitte einen Wert größer als 0 eingeben.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Anlage speichern" })).toBeDisabled();
  });
});
