import { afterEach, describe, expect, it } from "vitest";

import { ecosVehicleDraftStorageKey } from "../vehicles/vehicleViewModel";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import type { DigitalCenterWorkItem } from "./digitalCenterModel";
import {
  buildDigitalCenterVehicleDraft,
  openDigitalCenterVehicleDraft,
  rankDigitalCenterVehicleCandidates
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

  it("ranks address and name matches before ordinary vehicles without mutating input", () => {
    const ordinary = vehicleFixture({ id: "ordinary", inventoryNumber: "RK-LOK-000003", name: "V 200" });
    const nameMatch = vehicleFixture({
      id: "name-match", inventoryNumber: "RK-LOK-000002", name: "BR-106", digitalDecoderNumber: "8"
    });
    const addressMatch = vehicleFixture({
      id: "address-match", inventoryNumber: "RK-LOK-000001", name: "Andere Lok", digitalDecoderNumber: "3"
    });
    const vehicles = [ordinary, nameMatch, addressMatch];

    expect(rankDigitalCenterVehicleCandidates(workItem, vehicles, "").map((vehicle) => vehicle.id))
      .toEqual([addressMatch.id, nameMatch.id, ordinary.id]);
    expect(vehicles.map((vehicle) => vehicle.id)).toEqual([ordinary.id, nameMatch.id, addressMatch.id]);
  });

  it("filters candidates by searchable vehicle identity", () => {
    const ordinary = vehicleFixture({ id: "ordinary", name: "V 200" });
    const nameMatch = vehicleFixture({ id: "name-match", name: "BR 106" });

    expect(rankDigitalCenterVehicleCandidates(workItem, [ordinary, nameMatch], "br 106")
      .map((vehicle) => vehicle.id)).toEqual([nameMatch.id]);
  });
});
