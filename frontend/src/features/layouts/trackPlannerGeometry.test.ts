import { describe, expect, it } from "vitest";

import { routePolylinePoints, trackObjectTransform } from "./trackPlannerGeometry";

describe("track planner geometry", () => {
  it("keeps Tillig G1 millimetres exact in SVG point strings", () => {
    expect(routePolylinePoints([{ xMm: 0, yMm: 0 }, { xMm: 166, yMm: 0 }])).toBe("0,0 166,0");
  });

  it("combines the millimetre translation and rotation in one object transform", () => {
    expect(trackObjectTransform({ positionXMm: 517, positionYMm: 250, rotationDegrees: 15 }))
      .toBe("translate(517 250) rotate(15)");
  });
});
