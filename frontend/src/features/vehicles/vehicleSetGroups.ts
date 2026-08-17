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

function setSummaryForVehicle(vehicle: Vehicle): VehicleSetSummary | null {
	if (vehicle.vehicleSet) return vehicle.vehicleSet;
	if (!vehicle.vehicleSetId) return null;
	return {
		id: vehicle.vehicleSetId,
		inventoryNumber: "",
		name: vehicle.vehicleSetName || "",
		manufacturer: vehicle.manufacturer,
		articleNumber: vehicle.articleNumber,
		gauge: vehicle.gauge,
		epoch: vehicle.epoch,
		memberCount: vehicle.vehicleSetSize || 1,
		position: vehicle.vehicleSetPosition || 0
	};
}

export function groupVehicleInventory(vehicles: Vehicle[]): VehicleInventoryGroup[] {
  const groups: VehicleInventoryGroup[] = [];
  const setIndexes = new Map<string, number>();

  for (const vehicle of vehicles) {
		const set = setSummaryForVehicle(vehicle);
		if (!set) {
      groups.push({ kind: "single", vehicle });
      continue;
    }
		const existingIndex = setIndexes.get(set.id);
    if (existingIndex !== undefined) {
      const existing = groups[existingIndex];
			if (existing.kind === "set") {
				existing.members.push(vehicle);
				if (vehicle.vehicleSet) {
					existing.set = vehicle.vehicleSet;
					existing.totalMemberCount = vehicle.vehicleSet.memberCount;
				}
			}
      continue;
    }
		setIndexes.set(set.id, groups.length);
    groups.push({
      kind: "set",
			id: set.id,
			set,
			members: [vehicle],
			visibleMemberCount: 1,
			totalMemberCount: set.memberCount
    });
  }

  for (const group of groups) {
    if (group.kind === "set") {
			group.visibleMemberCount = group.members.length;
      group.members.sort((left, right) => (
				(left.vehicleSet?.position ?? left.vehicleSetPosition ?? 0)
				- (right.vehicleSet?.position ?? right.vehicleSetPosition ?? 0)
      ));
    }
  }
  return groups;
}
