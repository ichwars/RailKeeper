import { describe, expect, it } from "vitest";

import { availableStartView, isViewTemporarilyDisabled } from "./navigationAvailability";

describe("navigation availability", () => {
  it("keeps the layout workspace available", () => {
    expect(isViewTemporarilyDisabled("layouts")).toBe(false);
    expect(isViewTemporarilyDisabled("accessories")).toBe(false);
  });

  it.each([
    ["layouts", "layouts"],
    ["inventory", "vehicles"],
    ["accessories", "accessories"],
    ["importExport", "importExport"],
    ["settings", "settings"],
    ["invalid", "overview"]
  ])("maps configured start view %s to %s", (stored, expected) => {
    expect(availableStartView(stored)).toBe(expected);
  });
});
