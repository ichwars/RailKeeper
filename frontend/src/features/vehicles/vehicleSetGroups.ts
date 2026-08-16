import type { Vehicle, VehicleSetSummary } from "../../shared/api";

export type VehicleInventoryGroup =
  | { kind: "single"; vehicle: Vehicle }
	| {
		kind: "set";
		id: string;
		set: VehicleSetSummary;
		members: Vehicle[];
		visibleMemberCount: number;
		totalMemberCount: number;
	};

export type VehicleInventorySetGroup = Extract<VehicleInventoryGroup, { kind: "set" }>;

export function groupVehicleInventory(vehicles: Vehicle[]): VehicleInventoryGroup[] {
  const groups: VehicleInventoryGroup[] = [];
  const setIndexes = new Map<string, number>();

  for (const vehicle of vehicles) {
		if (!vehicle.vehicleSet) {
      groups.push({ kind: "single", vehicle });
      continue;
    }
		const existingIndex = setIndexes.get(vehicle.vehicleSet.id);
    if (existingIndex !== undefined) {
      const existing = groups[existingIndex];
      if (existing.kind === "set") existing.members.push(vehicle);
      continue;
    }
		setIndexes.set(vehicle.vehicleSet.id, groups.length);
    groups.push({
      kind: "set",
			id: vehicle.vehicleSet.id,
			set: vehicle.vehicleSet,
			members: [vehicle],
			visibleMemberCount: 1,
			totalMemberCount: vehicle.vehicleSet.memberCount
    });
  }

  for (const group of groups) {
    if (group.kind === "set") {
			group.visibleMemberCount = group.members.length;
      group.members.sort((left, right) => (
				(left.vehicleSet?.position || 0) - (right.vehicleSet?.position || 0)
      ));
    }
  }
  return groups;
}
