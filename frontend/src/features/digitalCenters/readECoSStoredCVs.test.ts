import { afterEach, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import { readECoSStoredCVs } from "./readECoSStoredCVs";

afterEach(() => vi.restoreAllMocks());

it("uses the completed comparison snapshot for adoption without a separate device request", async () => {
  const item: DigitalCenterWorkItem = {
    id: "item", sessionId: "session", centerObjectId: "1002", vehicleId: "",
    name: "Test", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
    center: { cvs: [{ number: 257, value: 0, indexCV31: 0, indexCV32: 0 }], cvRead: {
      source: "database", nextCV: 1025, unavailable: 0, sourceHost: "ecos.test", readAt: "2026-10-29T08:00:00Z" } },
    railkeeper: {}, proposed: {}, conflicts: [], createdAt: "", updatedAt: ""
  };
  const read = vi.spyOn(api, "readDigitalCenterCVBatch");
  const result = await readECoSStoredCVs(item, vi.fn(), { useSaved: true });
  expect(result).toBe(item);
  expect(result?.center.cvs?.[0]).toMatchObject({ indexCV31: 0, indexCV32: 0 });
  expect(read).not.toHaveBeenCalled();
  expect(await readECoSStoredCVs(item, vi.fn(), { useSaved: true, isCancelled: () => true })).toBeNull();
});

it("does not block vehicle adoption or retry a completed comparison attempt with incomplete CV data", async () => {
  const item: DigitalCenterWorkItem = {
    id: "item", sessionId: "session", centerObjectId: "1002", vehicleId: "",
    name: "Test", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "incomplete",
    center: { cvs: [{ number: 8, value: 151 }], cvSources: {
      sources: [], backupStatus: "failed", backupAdded: 0, conflicts: [], warning: "Unavailable" } },
    railkeeper: {}, proposed: {}, conflicts: [], createdAt: "", updatedAt: ""
  };
  const before = structuredClone(item), read = vi.spyOn(api, "readDigitalCenterCVBatch");
  const result = await readECoSStoredCVs(item, vi.fn(), { useSaved: true });
  expect(result?.center.cvs).toEqual([]);
  expect(result?.name).toBe("Test");
  expect(item).toEqual(before);
  expect(read).not.toHaveBeenCalled();
});

it("distinguishes refreshing saved progress from explicitly restarting a scan", async () => {
  const item: DigitalCenterWorkItem = {
    id: "item", sessionId: "session", centerObjectId: "1002", vehicleId: "",
    name: "Test", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
    center: { cvs: [], cvRead: { source: "database", nextCV: 1025, unavailable: 0,
      sourceHost: "ecos.test", readAt: "2026-10-22T11:00:00Z" } },
    railkeeper: {}, proposed: {}, conflicts: [], createdAt: "", updatedAt: ""
  };
  const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockResolvedValue(item);
  await readECoSStoredCVs(item, () => undefined, { refresh: true });
  expect(read).toHaveBeenLastCalledWith("session", "item", 1, { source: "database", restart: false });
  await readECoSStoredCVs(item, () => undefined, { restart: true });
  expect(read).toHaveBeenLastCalledWith("session", "item", 1, { source: "database", restart: true });
});
