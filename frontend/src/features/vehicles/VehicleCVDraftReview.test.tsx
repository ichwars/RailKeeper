import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { cvValueFixture } from "../../test/fixtures/vehicles";
import { cvDraftInput } from "./cvDraft";
import { VehicleCVDraftReview } from "./VehicleCVDraftReview";

describe("VehicleCVDraftReview", () => {
  const original = cvValueFixture();
  const changes = [{ id: original.id, original, value: { ...cvDraftInput(original), value: 5 } }];

  it("offers a single explicit save and shows both sides of a stale-row conflict", async () => {
    const user = userEvent.setup(), rebase = vi.fn();
    render(<VehicleCVDraftReview draft={{ changes, values: [], conflicts: [original.id] }}
      stored={[{ ...original, value: 7, revision: "d".repeat(64) }]} files={[]} disabled={false}
      save={vi.fn()} discard={vi.fn()} discardChange={vi.fn()} rebase={rebase} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Ihr Entwurf bleibt erhalten");
    expect(screen.getByRole("cell", { name: /7 Seit Beginn/ })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "CV-Entwurf speichern" })).toBeDisabled();
    await user.click(screen.getByRole("button", { name: "Diesen CV als neue Entwurfsbasis verwenden" }));
    expect(rebase).toHaveBeenCalledWith(original.id);
  });

  it("keeps every draft write action disabled for read-only or busy editors", () => {
    render(<VehicleCVDraftReview draft={{ changes, values: [], conflicts: [] }} stored={[original]} files={[]}
      disabled save={vi.fn()} discard={vi.fn()} discardChange={vi.fn()} rebase={vi.fn()} />);
    for (const button of screen.getAllByRole("button")) expect(button).toBeDisabled();
  });
});
