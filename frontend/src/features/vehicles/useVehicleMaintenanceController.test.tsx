import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { maintenanceFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { todayISODate } from "./vehicleMaintenance";
import { useVehicleMaintenanceController } from "./useVehicleMaintenanceController";

function renderController() {
  const setSaving = vi.fn();
  const onMessage = vi.fn();
  const refreshSelectedVehicle = vi.fn().mockResolvedValue(undefined);
  const hook = renderHook(() => useVehicleMaintenanceController({
    selected: vehicleFixture(),
    setSaving,
    onMessage,
    refreshSelectedVehicle
  }));
  return { ...hook, setSaving, onMessage, refreshSelectedVehicle };
}

describe("useVehicleMaintenanceController", () => {
  it("loads and resets an existing maintenance entry", () => {
    const entry = maintenanceFixture({ notes: "Prüfen", cost: "12,00 €" });
    const { result } = renderController();

    act(() => result.current.commands.edit(entry));
    expect(result.current.state.editingId).toBe(entry.id);
    expect(result.current.state.form).toMatchObject({ notes: "Prüfen", cost: "12,00 €" });

    act(() => result.current.commands.resetForm());
    expect(result.current.state.editingId).toBeNull();
    expect(result.current.state.form.kind).toBe("Wartung");
  });

  it("creates a completed maintenance entry and refreshes the vehicle", async () => {
    const created = maintenanceFixture({ id: "maintenance-created", status: "erledigt" });
    vi.spyOn(api, "createVehicleMaintenance").mockResolvedValue(created);
    const { result, refreshSelectedVehicle, setSaving } = renderController();

    act(() => result.current.commands.updateForm({ status: "erledigt", completedAt: "", cost: " 12,00 € " }));
    act(() => result.current.commands.save());

    await waitFor(() => expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1"));
    expect(api.createVehicleMaintenance).toHaveBeenCalledWith("vehicle-1", expect.objectContaining({
      status: "erledigt",
      completedAt: todayISODate(),
      cost: "12,00 €"
    }));
    expect(setSaving).toHaveBeenNthCalledWith(1, true);
    expect(setSaving).toHaveBeenLastCalledWith(false);
  });

  it("updates, completes and deletes maintenance entries", async () => {
    const entry = maintenanceFixture({ completedAt: "" });
    vi.spyOn(api, "updateVehicleMaintenance").mockResolvedValue(entry);
    vi.spyOn(api, "deleteVehicleMaintenance").mockResolvedValue(undefined);
    const { result, refreshSelectedVehicle } = renderController();

    act(() => result.current.commands.edit(entry));
    act(() => result.current.commands.save());
    await waitFor(() => expect(api.updateVehicleMaintenance).toHaveBeenCalledWith(
      "vehicle-1",
      entry.id,
      expect.objectContaining({ kind: entry.kind })
    ));

    act(() => result.current.commands.complete(entry));
    await waitFor(() => expect(api.updateVehicleMaintenance).toHaveBeenLastCalledWith(
      "vehicle-1",
      entry.id,
      expect.objectContaining({ status: "erledigt", completedAt: todayISODate() })
    ));

    act(() => result.current.commands.remove(entry));
    await waitFor(() => expect(api.deleteVehicleMaintenance).toHaveBeenCalledWith("vehicle-1", entry.id));
    expect(refreshSelectedVehicle).toHaveBeenCalledTimes(3);
  });
});
