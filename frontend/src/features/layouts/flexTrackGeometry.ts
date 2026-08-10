import type { FlexTrackPath, TrackPoint } from "../../shared/api";

const PREVIEW_SEGMENTS = 32;

export function sampleFlexPath(path: FlexTrackPath): TrackPoint[] {
  const endRadians = path.endDirectionDegrees * Math.PI / 180;
  const control = [
    { xMm: 0, yMm: 0 },
    { xMm: path.startHandleMm, yMm: 0 },
    {
      xMm: path.endXMm - path.endHandleMm * Math.cos(endRadians),
      yMm: path.endYMm - path.endHandleMm * Math.sin(endRadians)
    },
    { xMm: path.endXMm, yMm: path.endYMm }
  ];
  return Array.from({ length: PREVIEW_SEGMENTS + 1 }, (_, index) => {
    const t = index / PREVIEW_SEGMENTS;
    const inverse = 1 - t;
    return {
      xMm: inverse ** 3 * control[0].xMm + 3 * inverse ** 2 * t * control[1].xMm +
        3 * inverse * t ** 2 * control[2].xMm + t ** 3 * control[3].xMm,
      yMm: inverse ** 3 * control[0].yMm + 3 * inverse ** 2 * t * control[1].yMm +
        3 * inverse * t ** 2 * control[2].yMm + t ** 3 * control[3].yMm
    };
  });
}
