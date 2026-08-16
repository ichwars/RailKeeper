import { describe, expect, it } from "vitest";

import {
  quickMenuFloatingPosition,
  quickMenuShouldOpenAbove
} from "./vehicleInventoryRenderers";

describe("vehicle inventory quick menu placement", () => {
  it("opens upward when a lower row has insufficient space below", () => {
    expect(quickMenuShouldOpenAbove({
      anchorTop: 900,
      anchorBottom: 930,
      menuHeight: 286,
      viewportHeight: 1011
    })).toBe(true);
  });

  it("keeps opening downward when the menu fits below the row", () => {
    expect(quickMenuShouldOpenAbove({
      anchorTop: 470,
      anchorBottom: 530,
      menuHeight: 286,
      viewportHeight: 1011
    })).toBe(false);
  });

  it("keeps every action accessible in a short viewport", () => {
    expect(quickMenuFloatingPosition({
      anchorTop: 149,
      anchorBottom: 181,
      anchorRight: 464,
      menuWidth: 150,
      menuHeight: 286,
      viewportWidth: 496,
      viewportHeight: 357
    })).toEqual({
      top: 187,
      left: 314,
      maxHeight: 162,
      openAbove: false
    });
  });
});
