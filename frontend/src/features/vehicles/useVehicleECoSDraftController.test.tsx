import { renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { buildDigitalCenterVehicleDraft } from "../digitalCenters/digitalCenterVehicleAdoption";
import type { DigitalCenterWorkItem } from "../digitalCenters/digitalCenterModel";
import { useVehicleECoSDraftController } from "./useVehicleECoSDraftController";
import { ecosVehicleDraftStorageKey } from "./vehicleViewModel";

const cs3Item: DigitalCenterWorkItem = {
  id: "item-77", sessionId: "session-1", centerObjectId: "77", vehicleId: "",
  name: "BR 106", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
  center: { objectId: 77, name: "BR 106", decoderAddress: 3, protocol: "DCC" },
  railkeeper: {}, proposed: {}, conflicts: [],
  createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
};

describe("useVehicleECoSDraftController", () => {
  afterEach(() => {
    window.sessionStorage.clear();
    window.history.replaceState(null, "", "/");
  });

  it("loads a CS3 vehicle draft and removes its navigation marker", async () => {
    const draft = buildDigitalCenterVehicleDraft(cs3Item, "cs3");
    window.sessionStorage.setItem(ecosVehicleDraftStorageKey, JSON.stringify(draft));
    window.history.replaceState(null, "", "/vehicles?source=cs3");
    const onOpenCreate = vi.fn();
    const onFinishOpen = vi.fn();

    renderHook(() => useVehicleECoSDraftController({
      onOpenCreate,
      onOpenUpdate: vi.fn(),
      onFinishOpen,
      onMessage: vi.fn(),
      t: (key) => key
    }));

    await waitFor(() => expect(onOpenCreate).toHaveBeenCalledOnce());
    expect(onFinishOpen).toHaveBeenCalledWith(expect.objectContaining({
      source: "cs3",
      externalMapping: expect.objectContaining({ provider: "cs3", externalId: "77" })
    }));
    expect(window.sessionStorage.getItem(ecosVehicleDraftStorageKey)).toBeNull();
    expect(window.location.pathname + window.location.search).toBe("/vehicles");
  });
});
