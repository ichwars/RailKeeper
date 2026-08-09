import type { StorageLocation } from "./api";

export function activeStorageLocations(locations: StorageLocation[]) {
  const byID = new Map(locations.map((location) => [location.id, location]));
  return locations.filter((location) => {
    const visited = new Set<string>();
    let current: StorageLocation | undefined = location;
    while (current) {
      if (current.archived || visited.has(current.id)) return false;
      visited.add(current.id);
      if (!current.parentId) return true;
      current = byID.get(current.parentId);
      if (!current) return false;
    }
    return false;
  });
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
