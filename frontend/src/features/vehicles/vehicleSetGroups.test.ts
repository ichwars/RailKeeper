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
	it("keeps legacy set membership when a partial update omits the canonical summary", () => {
		const grouped = groupVehicleInventory([
			vehicle("member-2", {
				vehicleSetId: "set-1",
				vehicleSetName: "TEE Roland",
				vehicleSetPosition: 2,
				vehicleSetSize: 4
			})
		]);

		expect(grouped[0]).toMatchObject({
			kind: "set",
			id: "set-1",
			set: { name: "TEE Roland", memberCount: 4, position: 2 },
			members: [{ id: "member-2" }]
		});
	});

	it("shows a canonical set once when only one member matches a filter", () => {
		const grouped = groupVehicleInventory([vehicle("member-3", { vehicleSet: {
			id: "set-1", inventoryNumber: "RK-SET-000001", name: "TEE Roland", manufacturer: "Märklin",
			gauge: "H0", memberCount: 4, position: 3
		} })]);
		expect(grouped).toHaveLength(1);
		expect(grouped[0]).toMatchObject({
			kind: "set", visibleMemberCount: 1, totalMemberCount: 4, members: [{ id: "member-3" }]
		});
	});

	it("keeps explicit member positions after sorted input determines group order", () => {
		const set = (id: string, position: number, memberCount: number) => ({
			id, inventoryNumber: `RK-SET-${id}`, name: id, manufacturer: "Roco", gauge: "H0", memberCount, position
		});
		const grouped = groupVehicleInventory([
			vehicle("member-2", { vehicleSet: set("set-b", 2, 2) }),
			vehicle("member-a", { vehicleSet: set("set-a", 1, 1) }),
			vehicle("member-1", { vehicleSet: set("set-b", 1, 2) })
		]);
		expect(grouped.map((group) => group.kind === "set" ? group.id : group.vehicle.id))
			.toEqual(["set-b", "set-a"]);
		expect(grouped[0].kind === "set" ? grouped[0].members.map((member) => member.id) : [])
			.toEqual(["member-1", "member-2"]);
	});

	it("uses canonical set data and distinguishes visible from total members", () => {
		const set = {
			id: "set-1",
			inventoryNumber: "RK-SET-000001",
			name: "Rheingold",
			manufacturer: "Roco",
			articleNumber: "45923",
			gauge: "H0",
			memberCount: 4,
			position: 2
		};
		const grouped = groupVehicleInventory([
			vehicle("member-2", { vehicleSet: set }),
			vehicle("member-1", { vehicleSet: { ...set, position: 1 } })
		]);

		expect(grouped[0]).toMatchObject({
			kind: "set",
			id: "set-1",
			set: { id: "set-1", inventoryNumber: "RK-SET-000001", memberCount: 4 },
			visibleMemberCount: 2,
			totalMemberCount: 4
		});
		if (grouped[0].kind === "set") {
			expect(grouped[0].members.map((member) => member.id)).toEqual(["member-1", "member-2"]);
		}
	});

  it("keeps singles in order and collects ordered set members", () => {
		const set = {
			id: "set-1", inventoryNumber: "RK-SET-000001", name: "TEE", manufacturer: "Roco",
			gauge: "H0", memberCount: 2, position: 1
		};
    const grouped = groupVehicleInventory([
			vehicle("member-2", { vehicleSet: { ...set, position: 2 } }),
      vehicle("single"),
			vehicle("member-1", { vehicleSet: set })
    ]);

    expect(grouped).toHaveLength(2);
		expect(grouped[0]).toMatchObject({ kind: "set", id: "set-1", set: { name: "TEE" } });
    if (grouped[0].kind === "set") expect(grouped[0].members.map((member) => member.id)).toEqual(["member-1", "member-2"]);
    expect(grouped[1]).toMatchObject({ kind: "single", vehicle: { id: "single" } });
  });
});
