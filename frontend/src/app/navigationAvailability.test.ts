import { describe, expect, it } from "vitest";

import { availableStartView, isViewTemporarilyDisabled } from "./navigationAvailability";

describe("navigation availability", () => {
  it("marks only the layout workspace as temporarily disabled", () => {
    expect(isViewTemporarilyDisabled("layouts")).toBe(true);
    expect(isViewTemporarilyDisabled("accessories")).toBe(false);
  });

  it.each([
    ["layouts", "overview"],
    ["inventory", "vehicles"],
    ["accessories", "accessories"],
    ["invalid", "overview"]
  ])("maps configured start view %s to %s", (stored, expected) => {
    expect(availableStartView(stored)).toBe(expected);
  });
});
