import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { emptyOptions } from "./vehicleViewModel";
import { useVehicleEditorController } from "./useVehicleEditorController";

function renderController() {
  const onMessage = vi.fn();
  const onReset = vi.fn();
  const onDetailLoaded = vi.fn();
  const onFormChange = vi.fn();
  const hook = renderHook(() => useVehicleEditorController({
    options: emptyOptions,
    onMessage,
    onReset,
    onDetailLoaded,
    onFormChange
  }));
  return { ...hook, onMessage, onReset, onDetailLoaded, onFormChange };
}

describe("useVehicleEditorController", () => {
  it("owns the create and close lifecycle", () => {
    const { result, onMessage, onReset } = renderController();

    act(() => result.current.commands.openCreate());
    expect(result.current.state).toMatchObject({ modalOpen: true, mode: "create", activeTab: "model" });
    expect(onReset).toHaveBeenCalledWith("create");

    act(() => result.current.commands.closeModal());
    expect(result.current.state.modalOpen).toBe(false);
    expect(onReset).toHaveBeenLastCalledWith("close");
    expect(onMessage).toHaveBeenLastCalledWith("");
  });

  it("loads a vehicle into edit mode", async () => {
    const vehicle = vehicleFixture();
    vi.spyOn(api, "vehicle").mockResolvedValue(vehicle);
    const { result, onDetailLoaded } = renderController();

    act(() => result.current.commands.openEdit(vehicle, "cv"));

    await waitFor(() => expect(result.current.state.modalOpen).toBe(true));
    expect(result.current.state).toMatchObject({ selected: vehicle, mode: "edit", activeTab: "cv" });
    expect(result.current.state.form.name).toBe("BR 106");
    expect(onDetailLoaded).toHaveBeenCalledWith(vehicle);
  });

  it("keeps coupled fields synchronized", () => {
    const { result, onFormChange } = renderController();

    act(() => result.current.commands.update({ couplingSame: true, couplingFront: "NEM" }));
    act(() => result.current.commands.updateCouplingFront("Kurzkupplung"));

    expect(result.current.state.form.couplingRear).toBe("Kurzkupplung");
    expect(onFormChange).toHaveBeenLastCalledWith(expect.objectContaining({
      couplingFront: "Kurzkupplung",
      couplingRear: "Kurzkupplung"
    }));
  });
});
