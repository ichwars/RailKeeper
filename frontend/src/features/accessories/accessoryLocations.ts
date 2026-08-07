import type { StorageLocation } from "../../shared/api";

export function activeAccessoryLocations(locations: StorageLocation[]) {
  return locations.filter((location) => !location.archived);
}

export function accessoryLocationPath(location: StorageLocation, locations: StorageLocation[]) {
  const byID = new Map(locations.map((item) => [item.id, item]));
  const names: string[] = [];
  const visited = new Set<string>();
  let current: StorageLocation | undefined = location;
  while (current && !visited.has(current.id)) {
    visited.add(current.id);
    names.unshift(current.name);
    current = current.parentId ? byID.get(current.parentId) : undefined;
  }
  return names.join(" / ");
}
