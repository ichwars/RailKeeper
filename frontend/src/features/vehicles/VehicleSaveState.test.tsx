import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { VehicleSaveState } from "./VehicleSaveState";

describe("VehicleSaveState", () => {
  it("distinguishes a saved side action from the remaining draft", () => {
    render(<VehicleSaveState tab="uploads" saved="media" conflicts={[]} onResolve={vi.fn()} />);
    expect(screen.getByRole("status")).toHaveTextContent("Abbrechen nimmt diese Aktion nicht zurück");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(screen.getByText(/Bildtitel, Hauptbild/)).toBeInTheDocument();
  });

  it("shows both conflicting values and requires an explicit choice", async () => {
    const user = userEvent.setup(), onResolve = vi.fn();
    render(<VehicleSaveState tab="model" saved={null} onResolve={onResolve}
      conflicts={[{ field: "name", local: "Mein Entwurf", saved: "Neue gespeicherte Bezeichnung" }]} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Mein Entwurf");
    expect(screen.getByRole("alert")).toHaveTextContent("Neue gespeicherte Bezeichnung");
    await user.click(screen.getByRole("button", { name: /Gespeicherte Werte/ }));
    expect(onResolve).toHaveBeenLastCalledWith("saved");
    await user.click(screen.getByRole("button", { name: /Meine Änderungen/ }));
    expect(onResolve).toHaveBeenLastCalledWith("draft");
  });
});
