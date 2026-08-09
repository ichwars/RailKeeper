import type { TrackPoint } from "../../shared/api";

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
