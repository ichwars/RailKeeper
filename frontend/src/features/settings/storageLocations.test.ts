import { describe, expect, it } from "vitest";

import type { StorageLocation } from "../../shared/api";
import { activeStorageLocations, storageLocationPath } from "../../shared/storageLocations";

const base = { archived: false, createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T10:00:00Z" };
const locations: StorageLocation[] = [
  { ...base, id: "room-a", name: "Raum A" },
  { ...base, id: "cabinet-a", parentId: "room-a", name: "Schrank" },
  { ...base, id: "room-b", name: "Raum B" },
  { ...base, id: "cabinet-b", parentId: "room-b", name: "Schrank", archived: true }
];

describe("storage locations", () => {
  it("builds unambiguous hierarchical paths", () => {
    expect(storageLocationPath(locations[1], locations)).toBe("Raum A / Schrank");
    expect(storageLocationPath(locations[3], locations)).toBe("Raum B / Schrank");
  });

  it("excludes archived locations from new operations", () => {
    expect(activeStorageLocations(locations).map((location) => location.id))
      .toEqual(["room-a", "cabinet-a", "room-b"]);
  });

  it("excludes descendants of archived ancestors from new operations", () => {
    const hierarchy: StorageLocation[] = [
      { ...base, id: "archive", name: "Archiv", archived: true },
      { ...base, id: "child", parentId: "archive", name: "Regal" },
      { ...base, id: "leaf", parentId: "child", name: "Fach" },
      { ...base, id: "active", name: "Werkstatt" }
    ];

    expect(activeStorageLocations(hierarchy).map((location) => location.id)).toEqual(["active"]);
  });

  it("defensively excludes locations with missing or cyclic ancestors", () => {
    const invalidHierarchy: StorageLocation[] = [
      { ...base, id: "missing", parentId: "unknown", name: "Verwaist" },
      { ...base, id: "cycle-a", parentId: "cycle-b", name: "A" },
      { ...base, id: "cycle-b", parentId: "cycle-a", name: "B" },
      { ...base, id: "root", name: "Werkstatt" }
    ];

    expect(activeStorageLocations(invalidHierarchy).map((location) => location.id)).toEqual(["root"]);
  });
});
