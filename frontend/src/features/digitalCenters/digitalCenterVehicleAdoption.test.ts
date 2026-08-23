import { afterEach, describe, expect, it } from "vitest";

import { ecosVehicleDraftStorageKey } from "../vehicles/vehicleViewModel";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import {
  buildDigitalCenterVehicleDraft,
  openDigitalCenterVehicleDraft
} from "./digitalCenterVehicleAdoption";

const workItem: DigitalCenterWorkItem = {
  id: "item-77",
  sessionId: "session-1",
  centerObjectId: "77",
  vehicleId: "",
  name: "BR 106",
  decoderAddress: 3,
  protocol: "DCC",
  compareStatus: "new",
  stationStatus: "read",
  center: { objectId: 77, name: "BR 106", decoderAddress: 3, protocol: "DCC" },
  railkeeper: {},
  proposed: {},
  conflicts: [],
  createdAt: "2026-08-23T08:00:00Z",
  updatedAt: "2026-08-23T08:00:00Z"
};

describe("digitalCenterVehicleAdoption", () => {
  afterEach(() => {
    window.sessionStorage.clear();
    window.history.replaceState(null, "", "/");
  });

  it("builds a bounded ECoS create draft", () => {
    const draft = buildDigitalCenterVehicleDraft(workItem);

    expect(draft).toMatchObject({
      source: "ecos",
      mode: "create",
      vehicle: {
        name: "BR 106",
        category: "Lokomotive",
        digital: true,
        digitalDecoderNumber: "3"
      },
      externalMapping: {
        provider: "ecos",
        externalId: "77",
        externalName: "BR 106",
        externalAddress: "3",
        externalProtocol: "DCC",
        syncStatus: "linked"
      },
      cvValues: [],
      functionValues: [],
      returnToDigitalCenters: { sessionId: "session-1", objectId: "77" }
    });
    expect(draft.vehicle.manufacturer).toBe("");
    expect(draft.vehicle.gauge).toBe("");
    expect(draft.vehicle.gattung).toBe("");
  });

  it("rejects invalid object ids before storing or navigating", () => {
    expect(() => openDigitalCenterVehicleDraft({ ...workItem, centerObjectId: "invalid" })).toThrow();
    expect(window.sessionStorage.getItem(ecosVehicleDraftStorageKey)).toBeNull();
    expect(window.location.pathname).toBe("/");
  });
});
