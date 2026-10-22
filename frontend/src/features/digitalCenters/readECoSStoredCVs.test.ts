import { afterEach, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import { readECoSStoredCVs } from "./readECoSStoredCVs";

afterEach(() => vi.restoreAllMocks());

it("distinguishes refreshing saved progress from explicitly restarting a scan", async () => {
  const item: DigitalCenterWorkItem = {
    id: "item", sessionId: "session", centerObjectId: "1002", vehicleId: "",
    name: "Test", decoderAddress: 3, protocol: "DCC", compareStatus: "new", stationStatus: "read",
    center: { cvs: [], cvRead: { nextCV: 1025, unavailable: 1024,
      sourceHost: "ecos.test", readAt: "2026-10-22T11:00:00Z" } },
    railkeeper: {}, proposed: {}, conflicts: [], createdAt: "", updatedAt: ""
  };
  const read = vi.spyOn(api, "readDigitalCenterCVBatch").mockResolvedValue(item);
  await readECoSStoredCVs(item, () => undefined, { refresh: true });
  expect(read).toHaveBeenLastCalledWith("session", "item", 1, { restart: false });
  await readECoSStoredCVs(item, () => undefined, { restart: true });
  expect(read).toHaveBeenLastCalledWith("session", "item", 1, { restart: true });
});
