import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api, type Vehicle } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { emptyOptions } from "./vehicleViewModel";
import { useVehicleEditorController } from "./useVehicleEditorController";

function renderController() {
  const onMessage = vi.fn();
  const onReset = vi.fn();
  const onDetailLoaded = vi.fn();
  const onDetailRefreshed = vi.fn();
  const onCoreSaved = vi.fn();
  const onFormChange = vi.fn();
  const onCollectionChanged = vi.fn();
  const hook = renderHook(() => useVehicleEditorController({
    options: emptyOptions,
    onMessage,
    onReset,
    onDetailLoaded,
    onDetailRefreshed,
    onCoreSaved,
    onFormChange,
    onCollectionChanged
  }));
  return { ...hook, onMessage, onReset, onDetailLoaded, onDetailRefreshed, onCoreSaved, onFormChange, onCollectionChanged };
}

describe("useVehicleEditorController", () => {
  it("retains the stable identity of React-compatible setters", () => {
    const { result } = renderController();
    const setters = result.current.setters;
    act(() => result.current.commands.update({ name: "Changed" }));
    expect(result.current.setters.setForm).toBe(setters.setForm);
    expect(result.current.setters.setSelected).toBe(setters.setSelected);
    expect(result.current.setters.setModalOpen).toBe(setters.setModalOpen);
  });

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

  it("refreshes the selected aggregate and collection without changing the active editor mode", async () => {
    const vehicle = vehicleFixture();
    const refreshed = vehicleFixture({ name: "Refreshed" });
    vi.spyOn(api, "vehicle").mockResolvedValueOnce(vehicle).mockResolvedValueOnce(refreshed);
    const { result, onCollectionChanged, onDetailRefreshed } = renderController();
    act(() => result.current.commands.openEdit(vehicle, "uploads"));
    await waitFor(() => expect(result.current.state.modalOpen).toBe(true));
    await act(() => result.current.commands.refreshSelectedVehicle());
    expect(result.current.state).toMatchObject({ selected: refreshed, mode: "edit", activeTab: "uploads" });
    expect(onDetailRefreshed).toHaveBeenLastCalledWith(refreshed, vehicle);
    expect(onCollectionChanged).toHaveBeenCalledOnce();
  });

  it("keeps typing during a side-action refresh while adopting untouched server fields", async () => {
    const vehicle = vehicleFixture(), refreshed = vehicleFixture({ description: "Stored note" });
    let finish: ((value: Vehicle) => void) | undefined;
    const response = new Promise<Vehicle>(resolve => { finish = resolve; });
    vi.spyOn(api, "vehicle").mockResolvedValueOnce(vehicle).mockReturnValueOnce(response);
    const { result, onReset, onDetailLoaded } = renderController();
    act(() => result.current.commands.openEdit(vehicle));
    await waitFor(() => expect(result.current.state.modalOpen).toBe(true));
    let pending: Promise<void> | undefined;
    act(() => { pending = result.current.commands.refreshSelectedVehicle(vehicle.id, "media"); });
    act(() => result.current.commands.update({ name: "Still typing" }));
    await act(async () => { finish?.(refreshed); await pending; });
    expect(result.current.state.form).toMatchObject({ name: "Still typing", description: "Stored note" });
    expect(result.current.state.lastSavedAction).toBe("media");
    expect(onDetailLoaded).toHaveBeenCalledOnce();
    expect(onReset).not.toHaveBeenCalled();
  });

  it("retains conflicting fields across repeated refreshes until explicitly resolved", async () => {
    const vehicle = vehicleFixture(), stored = vehicleFixture({ name: "Other editor" });
    vi.spyOn(api, "vehicle").mockResolvedValueOnce(vehicle).mockResolvedValue(stored);
    const { result } = renderController();
    act(() => result.current.commands.openEdit(vehicle));
    await waitFor(() => expect(result.current.state.modalOpen).toBe(true));
    act(() => result.current.commands.update({ name: "My draft" }));
    await act(() => result.current.commands.refreshSelectedVehicle());
    await act(() => result.current.commands.refreshSelectedVehicle());
    expect(result.current.state.refreshConflicts).toEqual([{ field: "name", local: "My draft", saved: "Other editor" }]);
    act(() => result.current.commands.resolveRefreshConflicts("saved"));
    expect(result.current.state.form.name).toBe("Other editor");
    expect(result.current.state.refreshConflicts).toEqual([]);
  });

  it("ignores a response arriving after the editor was closed", async () => {
    const vehicle = vehicleFixture();
    let finish: ((value: Vehicle) => void) | undefined;
    const response = new Promise<Vehicle>(resolve => { finish = resolve; });
    vi.spyOn(api, "vehicle").mockResolvedValueOnce(vehicle).mockReturnValueOnce(response);
    const { result, onDetailRefreshed } = renderController();
    act(() => result.current.commands.openEdit(vehicle));
    await waitFor(() => expect(result.current.state.modalOpen).toBe(true));
    const pending = result.current.commands.refreshSelectedVehicle();
    act(() => result.current.commands.closeModal());
    await act(async () => { finish?.(vehicle); await pending; });
    expect(result.current.state).toMatchObject({ selected: null, modalOpen: false });
    expect(result.current.state.form.name).toBe("");
    expect(onDetailRefreshed).not.toHaveBeenCalled();
  });

  it("acknowledges cleared core fields without dropping independent records", async () => {
    const original = vehicleFixture({ description: "Before" });
    vi.spyOn(api, "vehicle").mockResolvedValue(original);
    const { result, onCoreSaved } = renderController();
    act(() => result.current.commands.openEdit(original));
    await waitFor(() => expect(result.current.state.modalOpen).toBe(true));
    act(() => result.current.commands.update({ description: "" }));
    const submitted = result.current.state.form;
    const response = vehicleFixture({ description: undefined, cvValues: undefined, images: undefined });
    act(() => result.current.commands.acknowledgeSavedVehicle(response, submitted, "core"));
    expect(result.current.state.form.description).toBe("");
    expect(result.current.state.selected?.cvValues).toEqual(original.cvValues);
    expect(result.current.state.selected?.images).toBeUndefined();
    expect(onCoreSaved).toHaveBeenCalledOnce();
  });
});
