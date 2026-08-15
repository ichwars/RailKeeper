import { describe, expect, it } from "vitest";

import { quickMenuShouldOpenAbove } from "./vehicleInventoryRenderers";

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
});
