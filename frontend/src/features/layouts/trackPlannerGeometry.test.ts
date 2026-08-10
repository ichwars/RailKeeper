import { describe, expect, it } from "vitest";

import type { PlanTrackObject, TrackGeometryDefinition } from "../../shared/api";
import { routePolylinePoints, snapFlexEnd, snapTrackPose, trackObjectTransform } from "./trackPlannerGeometry";

const geometry: TrackGeometryDefinition = {
  id: "g1", libraryId: "tillig-v1", articleNumber: "83101", name: "Gleisstück G1",
  kind: "straight", lengthMm: 166, sourceUrl: "https://example.test/83101", status: "verified",
  createdAt: "2026-08-09T10:00:00Z",
  geometry: {
    schemaVersion: 1,
    ports: [
      { id: "a", xMm: 0, yMm: 0, directionDegrees: 180 },
      { id: "b", xMm: 166, yMm: 0, directionDegrees: 0 }
    ],
    routes: [{ id: "main", points: [{ xMm: 0, yMm: 0 }, { xMm: 166, yMm: 0 }] }]
  }
};

function track(id: string, positionXMm: number, positionYMm = 0, rotationDegrees = 0): PlanTrackObject {
  return {
    id, lineageId: id, revisionId: "revision-1", geometryId: geometry.id, geometry,
    positionXMm, positionYMm, rotationDegrees, elevationStartMm: 0, elevationEndMm: 0, version: 1,
    effectiveGeometry: geometry.geometry, effectiveLengthMm: geometry.lengthMm,
    createdAt: "2026-08-09T10:00:00Z", updatedAt: "2026-08-09T10:00:00Z"
  };
}

describe("track planner geometry", () => {
  it("keeps Tillig G1 millimetres exact in SVG point strings", () => {
    expect(routePolylinePoints([{ xMm: 0, yMm: 0 }, { xMm: 166, yMm: 0 }])).toBe("0,0 166,0");
  });

  it("combines the millimetre translation and rotation in one object transform", () => {
    expect(trackObjectTransform({ positionXMm: 517, positionYMm: 250, rotationDegrees: 15 }))
      .toBe("translate(517 250) rotate(15)");
  });

  it("snaps the nearest compatible endpoint within eight millimetres", () => {
    const snap = snapTrackPose(track("moving", 172, 2, 2), [track("target", 0)]);

    expect(snap).toMatchObject({
      snapped: true,
      movingPortId: "a",
      targetObjectId: "target",
      targetPortId: "b",
      pose: { positionXMm: 166, positionYMm: 0, rotationDegrees: 0 }
    });
  });

  it("does not snap outside the distance or direction tolerance", () => {
    expect(snapTrackPose(track("far", 174.01), [track("target", 0)]).snapped).toBe(false);
    expect(snapTrackPose(track("angled", 172, 0, 5.01), [track("target", 0)]).snapped).toBe(false);
  });

  it("snaps flex track through its effective endpoint instead of the catalogue endpoint", () => {
    const moving = track("moving", 0);
    moving.geometry = { ...geometry, kind: "flex", lengthMm: 664 };
    moving.effectiveGeometry = {
      schemaVersion: 1,
      ports: [
        { id: "a", xMm: 0, yMm: 0, directionDegrees: 180 },
        { id: "b", xMm: 500, yMm: 0, directionDegrees: 0 }
      ],
      routes: []
    };
    const target = track("target", 506);

    expect(snapTrackPose(moving, [target])).toMatchObject({
      snapped: true,
      movingPortId: "b",
      targetObjectId: "target",
      pose: { positionXMm: 6, positionYMm: 0, rotationDegrees: 0 }
    });
  });

  it("snaps a dragged flex endpoint to a compatible target without changing the target", () => {
    const moving = track("moving", 0);
    const target = track("target", 506);
    const path = {
      schemaVersion: 1 as const,
      endXMm: 500,
      endYMm: 0,
      endDirectionDegrees: 0,
      startHandleMm: 170,
      endHandleMm: 170
    };

    const result = snapFlexEnd(moving, path, [moving, target]);

    expect(result).toMatchObject({
      snapped: true,
      targetObjectId: "target",
      targetPortId: "a",
      path: { endXMm: 506, endYMm: 0, endDirectionDegrees: 0 }
    });
    expect(target.positionXMm).toBe(506);
  });
});
