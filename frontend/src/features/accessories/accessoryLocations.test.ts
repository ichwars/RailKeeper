import { describe, expect, it } from "vitest";

import type { StorageLocation } from "../../shared/api";
import { accessoryLocationPath, activeAccessoryLocations } from "./accessoryLocations";

const base = { archived: false, createdAt: "2026-08-07T10:00:00Z", updatedAt: "2026-08-07T10:00:00Z" };
const locations: StorageLocation[] = [
  { ...base, id: "room-a", name: "Raum A" },
  { ...base, id: "cabinet-a", parentId: "room-a", name: "Schrank" },
  { ...base, id: "room-b", name: "Raum B" },
  { ...base, id: "cabinet-b", parentId: "room-b", name: "Schrank", archived: true }
];

describe("accessory locations", () => {
  it("builds unambiguous hierarchical paths", () => {
    expect(accessoryLocationPath(locations[1], locations)).toBe("Raum A / Schrank");
    expect(accessoryLocationPath(locations[3], locations)).toBe("Raum B / Schrank");
  });

  it("excludes archived locations from new operations", () => {
    expect(activeAccessoryLocations(locations).map((location) => location.id))
      .toEqual(["room-a", "cabinet-a", "room-b"]);
  });
});
