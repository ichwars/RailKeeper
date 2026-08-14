import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { PlanFreeObject } from "../../shared/api";
import { FreePlanObjectInspector } from "./FreePlanObjectInspector";

const object = {
  id: "free-1", lineageId: "free-1", revisionId: "revision-1", name: "Bahnsteig 1",
  category: "platform", positionXMm: 120, positionYMm: 60, rotationDegrees: 15,
  shape: { schemaVersion: 1, kind: "rectangle", widthMm: 400, heightMm: 65 },
  version: 3, createdAt: "now", updatedAt: "now"
} satisfies PlanFreeObject;

describe("FreePlanObjectInspector", () => {
  it("shows shape facts and exposes focused edit, rotate and delete actions", async () => {
    const user = userEvent.setup();
    const onEdit = vi.fn();
    const onRotate = vi.fn();
    const onDelete = vi.fn();
    render(<FreePlanObjectInspector object={object} editable saving={false}
      onEdit={onEdit} onRotate={onRotate} onDelete={onDelete} />);

    expect(screen.getByRole("heading", { name: "Bahnsteig 1" })).toBeInTheDocument();
    expect(screen.getByText("400,00 × 65,00 mm")).toBeInTheDocument();
    expect(screen.getByText("120,00 / 60,00 mm")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Planobjekt bearbeiten" }));
    await user.click(screen.getByRole("button", { name: "+15°" }));
    await user.click(screen.getByRole("button", { name: "Planobjekt löschen" }));
    expect(onEdit).toHaveBeenCalledOnce();
    expect(onRotate).toHaveBeenCalledWith(15);
    expect(onDelete).toHaveBeenCalledOnce();
  });

  it("keeps published plan objects read-only", () => {
    render(<FreePlanObjectInspector object={object} editable={false} saving={false}
      onEdit={vi.fn()} onRotate={vi.fn()} onDelete={vi.fn()} />);
    expect(screen.queryByRole("button", { name: "Planobjekt bearbeiten" })).not.toBeInTheDocument();
    expect(screen.getByText("Bahnsteig")).toBeInTheDocument();
  });
});
