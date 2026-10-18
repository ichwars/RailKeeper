import { describe, expect, it } from "vitest";

import { availableStartView } from "./navigationAvailability";

describe("navigation availability", () => {
  it("rejects retired planning views while preserving accessories", () => {
    expect(availableStartView("layouts")).toBe("overview");
    expect(availableStartView("accessories")).toBe("accessories");
  });

  it.each([
    ["layouts", "overview"],
    ["inventory", "vehicles"],
    ["accessories", "accessories"],
    ["importExport", "importExport"],
    ["digitalCenters", "digitalCenters"],
    ["settings", "settings"],
    ["invalid", "overview"]
  ])("maps configured start view %s to %s", (stored, expected) => {
    expect(availableStartView(stored)).toBe(expected);
  });
});
