import type { PlanTrackObject, TrackPoint, TrackPort } from "../../shared/api";

const SNAP_DISTANCE_MM = 8;
const SNAP_DIRECTION_TOLERANCE_DEGREES = 5;

export type TrackPose = {
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
};

export type TrackSnapPreview = {
  snapped: boolean;
  pose: TrackPose;
  movingPortId?: string;
  targetObjectId?: string;
  targetPortId?: string;
  distanceMm?: number;
};

export function routePolylinePoints(points: TrackPoint[]): string {
  return points.map((point) => `${point.xMm},${point.yMm}`).join(" ");
}

export function trackObjectTransform(object: {
  positionXMm: number;
  positionYMm: number;
  rotationDegrees: number;
}): string {
  return `translate(${object.positionXMm} ${object.positionYMm}) rotate(${object.rotationDegrees})`;
}

export function normalizedRotation(value: number): number {
  return ((value % 360) + 360) % 360;
}

function transformedPort(port: TrackPort, pose: TrackPose): TrackPort {
  const radians = pose.rotationDegrees * Math.PI / 180;
  const cosine = Math.cos(radians);
  const sine = Math.sin(radians);
  return {
    ...port,
    xMm: pose.positionXMm + port.xMm * cosine - port.yMm * sine,
    yMm: pose.positionYMm + port.xMm * sine + port.yMm * cosine,
    directionDegrees: normalizedRotation(port.directionDegrees + pose.rotationDegrees)
  };
}

function angleDifference(first: number, second: number): number {
  const difference = Math.abs(normalizedRotation(first) - normalizedRotation(second));
  return Math.min(difference, 360 - difference);
}

function opposingAngleDifference(first: number, second: number): number {
  return angleDifference(first, normalizedRotation(second + 180));
}

function isBetterSnap(
  distance: number,
  targetObjectId: string,
  targetPortId: string,
  current: TrackSnapPreview
): boolean {
  const currentDistance = current.distanceMm ?? Number.POSITIVE_INFINITY;
  if (Math.abs(distance - currentDistance) > 1e-9) return distance < currentDistance;
  const objectComparison = targetObjectId.localeCompare(current.targetObjectId ?? "");
  if (objectComparison !== 0) return objectComparison < 0;
  return targetPortId.localeCompare(current.targetPortId ?? "") < 0;
}

export function snapTrackPose(moving: PlanTrackObject, objects: PlanTrackObject[]): TrackSnapPreview {
  const basePose: TrackPose = {
    positionXMm: moving.positionXMm,
    positionYMm: moving.positionYMm,
    rotationDegrees: moving.rotationDegrees
  };
  let result: TrackSnapPreview = { snapped: false, pose: basePose };
  for (const movingPort of moving.geometry.geometry.ports) {
    const currentPort = transformedPort(movingPort, basePose);
    for (const targetObject of objects) {
      if (targetObject.id === moving.id || targetObject.geometry.status !== "verified") continue;
      const targetPose: TrackPose = {
        positionXMm: targetObject.positionXMm,
        positionYMm: targetObject.positionYMm,
        rotationDegrees: targetObject.rotationDegrees
      };
      for (const targetLocalPort of targetObject.geometry.geometry.ports) {
        const targetPort = transformedPort(targetLocalPort, targetPose);
        const distance = Math.hypot(currentPort.xMm - targetPort.xMm, currentPort.yMm - targetPort.yMm);
        if (distance > SNAP_DISTANCE_MM ||
          opposingAngleDifference(currentPort.directionDegrees, targetPort.directionDegrees) >
            SNAP_DIRECTION_TOLERANCE_DEGREES ||
          (result.snapped && !isBetterSnap(distance, targetObject.id, targetPort.id, result))) continue;
        const rotationDegrees = normalizedRotation(
          targetPort.directionDegrees + 180 - movingPort.directionDegrees
        );
        const rotatedPort = transformedPort(movingPort, { ...basePose, rotationDegrees });
        result = {
          snapped: true,
          pose: {
            positionXMm: moving.positionXMm + targetPort.xMm - rotatedPort.xMm,
            positionYMm: moving.positionYMm + targetPort.yMm - rotatedPort.yMm,
            rotationDegrees
          },
          movingPortId: movingPort.id,
          targetObjectId: targetObject.id,
          targetPortId: targetPort.id,
          distanceMm: distance
        };
      }
    }
  }
  return result;
}
