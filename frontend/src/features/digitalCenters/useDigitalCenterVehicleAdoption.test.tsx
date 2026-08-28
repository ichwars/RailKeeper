import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, ApiError } from "../../shared/api";
import { setLanguage } from "../../shared/i18n";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import { digitalCenterExternalMapping } from "./digitalCenterVehicleAdoption";
import { useDigitalCenterVehicleAdoption } from "./useDigitalCenterVehicleAdoption";

const item: DigitalCenterWorkItem = {
  id: "item-77", sessionId: "session-1", centerObjectId: "77", vehicleId: "",
  name: "BR 106", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
  center: { objectId: 77, name: "BR 106", decoderAddress: 3, protocol: "DCC" },
  railkeeper: {}, proposed: {}, conflicts: [],
  createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
};

describe("useDigitalCenterVehicleAdoption", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    setLanguage("de");
  });

  it("loads vehicles and assigns the explicitly selected vehicle", async () => {
    const vehicles = [vehicleFixture({ id: "vehicle-1" }), vehicleFixture({ id: "vehicle-2" })];
    const onAssigned = vi.fn();
    vi.spyOn(api, "vehicles").mockResolvedValue(vehicles);
    vi.spyOn(api, "upsertVehicleExternalMapping").mockResolvedValue({
      id: "mapping-1", vehicleId: "vehicle-2", provider: "ecos", externalId: "77",
      syncStatus: "linked", createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
    });
    const { result } = renderHook(() => useDigitalCenterVehicleAdoption({ onAssigned }));

    await act(async () => result.current.commands.load(item));
    expect(api.vehicles).toHaveBeenCalledWith("");
    expect(result.current.state.vehicles).toEqual(vehicles);

    act(() => result.current.setters.setSelectedVehicleId("vehicle-2"));
    await act(async () => result.current.commands.assign(item, "ecos", "vehicle-2"));
    expect(api.upsertVehicleExternalMapping).toHaveBeenCalledWith(
      "vehicle-2",
      digitalCenterExternalMapping(item, "ecos")
    );
    expect(onAssigned).toHaveBeenCalledOnce();
  });

  it("keeps the selection and localizes an ownership conflict", async () => {
    vi.spyOn(api, "upsertVehicleExternalMapping").mockRejectedValue(
      new ApiError("conflict", "external_mapping_conflict", 409)
    );
    const onAssigned = vi.fn();
    const { result } = renderHook(() => useDigitalCenterVehicleAdoption({ onAssigned }));
    act(() => result.current.setters.setSelectedVehicleId("vehicle-2"));

    await act(async () => result.current.commands.assign(item, "ecos", "vehicle-2"));

    expect(result.current.state.selectedVehicleId).toBe("vehicle-2");
    expect(result.current.state.error).toBe(
      "Diese Lok der Digitalzentrale ist bereits einem anderen Fahrzeug zugeordnet. Bitte erneut auslesen."
    );
    expect(onAssigned).not.toHaveBeenCalled();
  });

  it("assigns a CS3 item with a CS3 external mapping", async () => {
    vi.spyOn(api, "upsertVehicleExternalMapping").mockResolvedValue({
      id: "mapping-1", vehicleId: "vehicle-2", provider: "cs3", externalId: "77",
      syncStatus: "linked", createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
    });
    const onAssigned = vi.fn();
    const { result } = renderHook(() => useDigitalCenterVehicleAdoption({ onAssigned }));

    await act(async () => result.current.commands.assign(item, "cs3", "vehicle-2"));

    expect(api.upsertVehicleExternalMapping).toHaveBeenCalledWith(
      "vehicle-2",
      expect.objectContaining({ provider: "cs3", externalId: "77" })
    );
    expect(onAssigned).toHaveBeenCalledOnce();
  });
});
