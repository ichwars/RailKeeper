import type { StorageLocation } from "../../shared/api";

export function activeStorageLocations(locations: StorageLocation[]) {
  return locations.filter((location) => !location.archived);
}

export function storageLocationPath(location: StorageLocation, locations: StorageLocation[]) {
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

export function availableStorageLocationParents(locations: StorageLocation[], editingID = "") {
  const excluded = new Set<string>(editingID ? [editingID] : []);
  if (editingID) {
    let changed = true;
    while (changed) {
      changed = false;
      locations.forEach((location) => {
        if (location.parentId && excluded.has(location.parentId) && !excluded.has(location.id)) {
          excluded.add(location.id);
          changed = true;
        }
      });
    }
  }
  return activeStorageLocations(locations).filter((location) => !excluded.has(location.id));
}
