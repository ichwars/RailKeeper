import { describe, expect, it } from "vitest";

import type { LayoutTwinUnit } from "../../shared/api";
import {
  screenToTwinPoint,
  twinGlobalToLocal,
  twinLocalToGlobal,
  twinPointInsideOutline
} from "./layoutTwinGeometry";

const unit = {
  positionXMm: 100, positionYMm: 200, rotationDegrees: 90
} as LayoutTwinUnit;

describe("layout twin geometry", () => {
  it("round-trips local and global coordinates through a rotated unit", () => {
    const global = twinLocalToGlobal(unit, { xMm: 10, yMm: 20 });
    expect(global.xMm).toBeCloseTo(80);
    expect(global.yMm).toBeCloseTo(210);
    const local = twinGlobalToLocal(unit, global);
    expect(local.xMm).toBeCloseTo(10);
    expect(local.yMm).toBeCloseTo(20);
  });

  it("detects positions outside a custom outline", () => {
    const outline = [
      { xMm: 0, yMm: 0 },
      { xMm: 100, yMm: 0 },
      { xMm: 100, yMm: 50 },
      { xMm: 0, yMm: 50 }
    ];
    expect(twinPointInsideOutline({ xMm: 25, yMm: 25 }, outline)).toBe(true);
    expect(twinPointInsideOutline({ xMm: 125, yMm: 25 }, outline)).toBe(false);
    expect(twinPointInsideOutline({ xMm: 0, yMm: 25 }, outline)).toBe(true);
  });

  it("accounts for letterboxing when mapping pointer coordinates into the SVG viewBox", () => {
    const point = screenToTwinPoint(300, 200,
      { left: 0, top: 0, width: 600, height: 400 },
      { x: 0, y: 0, width: 1000, height: 400 });
    expect(point.xMm).toBeCloseTo(500);
    expect(point.yMm).toBeCloseTo(200);
  });
});
