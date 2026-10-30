import { StrictMode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { buildDigitalCenterVehicleDraft } from "../digitalCenters/digitalCenterVehicleAdoption";
import type { DigitalCenterWorkItem } from "../digitalCenters/digitalCenterModel";
import { useVehicleECoSDraftController } from "./useVehicleECoSDraftController";
import { ecosVehicleDraftStorageKey } from "./vehicleViewModel";
import { api } from "../../shared/api";

const cs3Item: DigitalCenterWorkItem = {
  id: "item-77", sessionId: "session-1", centerObjectId: "77", vehicleId: "",
  name: "BR 106", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
  center: { objectId: 77, name: "BR 106", decoderAddress: 3, protocol: "DCC" },
  railkeeper: {}, proposed: {}, conflicts: [],
  createdAt: "2026-08-23T08:00:00Z", updatedAt: "2026-08-23T08:00:00Z"
};

describe("useVehicleECoSDraftController", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    window.sessionStorage.clear();
    window.history.replaceState(null, "", "/");
  });

  it.each([false, true])("prepares saved function keys without device reads (StrictMode: %s)", async strict => {
    const item: DigitalCenterWorkItem = { ...cs3Item, center: { ...cs3Item.center,
      functionSnapshot: { objectId: 77, host: "ecos.test", readAt: "2026-10-29T08:00:00Z",
        functions: [{ index: 1, description: 9, momentary: true }] } } };
    const draft = buildDigitalCenterVehicleDraft(item, "ecos");
    window.sessionStorage.setItem(ecosVehicleDraftStorageKey, JSON.stringify(draft));
    vi.spyOn(api, "masterData").mockResolvedValue([{ id: "shunt", key: "shunt", type: "symbols",
      label: "Rangieren", active: true, sortOrder: 0, createdAt: "", updatedAt: "", metadata: { ecosCode: 9 } }]);
    vi.spyOn(api, "decoderLibrary").mockResolvedValue({ schemaVersion: 1, profiles: [] });
    const read = vi.spyOn(api, "readVehicleECoSFunctions");
    const readCVs = vi.spyOn(api, "readDigitalCenterCVBatch");
    const onOpenCreate = vi.fn();
    renderHook(() => useVehicleECoSDraftController({ onOpenCreate, onOpenUpdate: vi.fn(),
      onFinishOpen: vi.fn(), onMessage: vi.fn(), t: key => key }), { wrapper: strict ? StrictMode : undefined });
    await waitFor(() => expect(onOpenCreate).toHaveBeenCalledOnce());
    expect(onOpenCreate).toHaveBeenCalledWith(expect.objectContaining({ functionEdits: {
      F1: expect.objectContaining({ name: "Rangieren", symbolKey: "shunt", mode: "moment" })
    } }));
    expect(read).not.toHaveBeenCalled();
    expect(readCVs).not.toHaveBeenCalled();
    expect(window.sessionStorage.getItem(ecosVehicleDraftStorageKey)).toBeNull();
  });

  it("retains an unprepared draft on unmount and opens it once on the next mount", async () => {
    const draft = buildDigitalCenterVehicleDraft({ ...cs3Item, center: { functionSnapshot: {
      objectId: 77, host: "ecos.test", readAt: "2026-10-29T08:00:00Z",
      functions: [{ index: 1, description: 9, momentary: true }]
    } } }, "ecos");
    const stored = JSON.stringify(draft);
    window.sessionStorage.setItem(ecosVehicleDraftStorageKey, stored);
    let resolveSymbols!: (symbols: []) => void;
    vi.spyOn(api, "masterData").mockReturnValue(new Promise(resolve => { resolveSymbols = resolve; }));
    vi.spyOn(api, "decoderLibrary").mockResolvedValue({ schemaVersion: 1, profiles: [] });
    const options = { onOpenCreate: vi.fn(), onOpenUpdate: vi.fn(), onFinishOpen: vi.fn(),
      onMessage: vi.fn(), t: (key: string) => key };
    const first = renderHook(() => useVehicleECoSDraftController(options));
    expect(window.sessionStorage.getItem(ecosVehicleDraftStorageKey)).toBe(stored);
    first.unmount();
    await act(async () => resolveSymbols([]));
    expect(options.onOpenCreate).not.toHaveBeenCalled();
    expect(window.sessionStorage.getItem(ecosVehicleDraftStorageKey)).toBe(stored);
    renderHook(() => useVehicleECoSDraftController(options), { wrapper: StrictMode });
    await waitFor(() => expect(options.onOpenCreate).toHaveBeenCalledOnce());
    expect(options.onFinishOpen).toHaveBeenCalledOnce();
    expect(window.sessionStorage.getItem(ecosVehicleDraftStorageKey)).toBeNull();
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
