import { describe, expect, it } from "vitest";

import type { FlexTrackPath } from "../../shared/api";
import { sampleFlexPath } from "./flexTrackGeometry";

describe("sampleFlexPath", () => {
  it("samples deterministic cubic Bezier endpoints for an unsaved preview", () => {
    const path: FlexTrackPath = {
      schemaVersion: 1,
      endXMm: 500,
      endYMm: 100,
      endDirectionDegrees: 20,
      startHandleMm: 180,
      endHandleMm: 170
    };

    const first = sampleFlexPath(path);
    const second = sampleFlexPath(path);

    expect(first).toEqual(second);
    expect(first[0]).toEqual({ xMm: 0, yMm: 0 });
    expect(first.at(-1)).toEqual({ xMm: 500, yMm: 100 });
    expect(first).toHaveLength(33);
  });
});
