import { describe, expect, it } from "vitest";

import type { Vehicle } from "../../shared/api";
import { groupVehicleInventory } from "./vehicleSetGroups";

const vehicle = (id: string, patch: Partial<Vehicle> = {}): Vehicle => ({
  id,
  inventoryNumber: id,
  manufacturer: "Roco",
  name: id,
  gauge: "H0",
  digital: false,
  dtDecoder: false,
  exhibitionReady: false,
  exhibition: false,
  abcBrakes: false,
  couplingSame: false,
  driveEnabled: false,
  headlightsEnabled: false,
  lightingEnabled: false,
  soundGeneratorEnabled: false,
  smokeGeneratorEnabled: false,
  qrCodeEnabled: false,
  createdAt: "2026-08-16T00:00:00Z",
  updatedAt: "2026-08-16T00:00:00Z",
  ...patch
});

describe("groupVehicleInventory", () => {
  it("keeps singles in order and collects ordered set members", () => {
    const grouped = groupVehicleInventory([
      vehicle("member-2", { vehicleSetId: "set-1", vehicleSetName: "TEE", vehicleSetPosition: 2 }),
      vehicle("single"),
      vehicle("member-1", { vehicleSetId: "set-1", vehicleSetName: "TEE", vehicleSetPosition: 1 })
    ]);

    expect(grouped).toHaveLength(2);
    expect(grouped[0]).toMatchObject({ kind: "set", id: "set-1", name: "TEE" });
    if (grouped[0].kind === "set") expect(grouped[0].members.map((member) => member.id)).toEqual(["member-1", "member-2"]);
    expect(grouped[1]).toMatchObject({ kind: "single", vehicle: { id: "single" } });
  });
});
