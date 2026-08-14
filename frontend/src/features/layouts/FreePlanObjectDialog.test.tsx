import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { PlanFreeObject } from "../../shared/api";
import { FreePlanObjectDialog } from "./FreePlanObjectDialog";

describe("FreePlanObjectDialog", () => {
  it("creates all shape variants with app-owned selects and explicit confirmation", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<FreePlanObjectDialog initialPosition={{ xMm: 300, yMm: 150 }} saving={false}
      onSubmit={onSubmit} onClose={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Schließen" })).toHaveFocus();
    expect(screen.getByRole("button", { name: "Kategorie" })).toHaveTextContent("Gebäude");
    expect(screen.getByRole("button", { name: "Form" })).toHaveTextContent("Rechteck");
    expect(screen.getByRole("spinbutton", { name: "Position X (mm)" })).toHaveValue(300);
    expect(screen.getByRole("spinbutton", { name: "Breite (mm)" })).toBeInTheDocument();

    await user.type(screen.getByRole("textbox", { name: "Name" }), "Gleisbezeichnung");
    await user.click(screen.getByRole("button", { name: "Form" }));
    await user.click(screen.getByRole("option", { name: "Beschriftung" }));
    expect(screen.queryByRole("spinbutton", { name: "Breite (mm)" })).not.toBeInTheDocument();
    await user.type(screen.getByRole("textbox", { name: "Text" }), "Gleis 1");
    await user.click(screen.getByRole("button", { name: "Planobjekt speichern" }));

    expect(onSubmit).toHaveBeenCalledWith({
      name: "Gleisbezeichnung",
      category: "structure",
      positionXMm: 300,
      positionYMm: 150,
      rotationDegrees: 0,
      shape: { schemaVersion: 1, kind: "label", text: "Gleis 1", fontSizeMm: 8 }
    });
  });

  it("edits existing objects, validates required values and cancels with Escape", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    const onClose = vi.fn();
    const object = {
      id: "free-1", lineageId: "free-1", revisionId: "revision-1", name: "Bahnsteig",
      category: "platform", positionXMm: 120, positionYMm: 60, rotationDegrees: 15,
      shape: { schemaVersion: 1, kind: "ellipse", widthMm: 400, heightMm: 65 },
      version: 3, createdAt: "2026-08-10T00:00:00Z", updatedAt: "2026-08-10T00:00:00Z"
    } satisfies PlanFreeObject;
    render(<FreePlanObjectDialog object={object} saving={false}
      onSubmit={onSubmit} onClose={onClose} />);

    expect(screen.getByRole("button", { name: "Form" })).toHaveTextContent("Ellipse");
    expect(screen.getByRole("button", { name: "Kategorie" })).toHaveTextContent("Bahnsteig");
    expect(screen.getByRole("spinbutton", { name: "Breite (mm)" })).toHaveValue(400);
    await user.clear(screen.getByRole("textbox", { name: "Name" }));
    expect(screen.getByRole("button", { name: "Planobjekt speichern" })).toBeDisabled();
    await user.keyboard("{Escape}");
    expect(onClose).toHaveBeenCalledOnce();
    expect(onSubmit).not.toHaveBeenCalled();
  });
});
