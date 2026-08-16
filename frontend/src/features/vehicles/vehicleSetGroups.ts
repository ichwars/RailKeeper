import type { Vehicle } from "../../shared/api";

export type VehicleInventoryGroup =
  | { kind: "single"; vehicle: Vehicle }
  | { kind: "set"; id: string; name: string; members: Vehicle[] };

export function groupVehicleInventory(vehicles: Vehicle[]): VehicleInventoryGroup[] {
  const groups: VehicleInventoryGroup[] = [];
  const setIndexes = new Map<string, number>();

  for (const vehicle of vehicles) {
    if (!vehicle.vehicleSetId) {
      groups.push({ kind: "single", vehicle });
      continue;
    }
    const existingIndex = setIndexes.get(vehicle.vehicleSetId);
    if (existingIndex !== undefined) {
      const existing = groups[existingIndex];
      if (existing.kind === "set") existing.members.push(vehicle);
      continue;
    }
    setIndexes.set(vehicle.vehicleSetId, groups.length);
    groups.push({
      kind: "set",
      id: vehicle.vehicleSetId,
      name: vehicle.vehicleSetName || vehicle.name,
      members: [vehicle]
    });
  }

  for (const group of groups) {
    if (group.kind === "set") {
      group.members.sort((left, right) => (
        (left.vehicleSetPosition || 0) - (right.vehicleSetPosition || 0)
      ));
    }
  }
  return groups;
}
