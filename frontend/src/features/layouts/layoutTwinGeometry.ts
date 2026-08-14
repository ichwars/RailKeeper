import type { LayoutTwinUnit } from "../../shared/api";

export type TwinPoint = { xMm: number; yMm: number };
export type TwinViewBox = { x: number; y: number; width: number; height: number };

export function screenToTwinPoint(
  clientX: number,
  clientY: number,
  rect: Pick<DOMRect, "left" | "top" | "width" | "height">,
  viewBox: TwinViewBox
): TwinPoint {
  const scale = Math.min(rect.width / viewBox.width, rect.height / viewBox.height);
  const renderedWidth = viewBox.width * scale;
  const renderedHeight = viewBox.height * scale;
  const offsetX = (rect.width - renderedWidth) / 2;
  const offsetY = (rect.height - renderedHeight) / 2;
  return {
    xMm: viewBox.x + (clientX - rect.left - offsetX) / scale,
    yMm: viewBox.y + (clientY - rect.top - offsetY) / scale
  };
}

export function twinGlobalToLocal(unit: LayoutTwinUnit, point: TwinPoint): TwinPoint {
  const radians = unit.rotationDegrees * Math.PI / 180;
  const x = point.xMm - unit.positionXMm;
  const y = point.yMm - unit.positionYMm;
  return {
    xMm: x * Math.cos(radians) + y * Math.sin(radians),
    yMm: -x * Math.sin(radians) + y * Math.cos(radians)
  };
}

export function twinLocalToGlobal(unit: LayoutTwinUnit, point: TwinPoint): TwinPoint {
  const radians = unit.rotationDegrees * Math.PI / 180;
  return {
    xMm: unit.positionXMm + point.xMm * Math.cos(radians) - point.yMm * Math.sin(radians),
    yMm: unit.positionYMm + point.xMm * Math.sin(radians) + point.yMm * Math.cos(radians)
  };
}

export function twinPointInsideOutline(point: TwinPoint, outline: TwinPoint[]): boolean {
  if (outline.length < 3) return true;
  let inside = false;
  let previous = outline.length - 1;
  for (const [index, current] of outline.entries()) {
    const other = outline[previous];
    const crossProduct = (point.yMm - current.yMm) * (other.xMm - current.xMm) -
      (point.xMm - current.xMm) * (other.yMm - current.yMm);
    const onSegment = Math.abs(crossProduct) < 1e-9 &&
      point.xMm >= Math.min(current.xMm, other.xMm) &&
      point.xMm <= Math.max(current.xMm, other.xMm) &&
      point.yMm >= Math.min(current.yMm, other.yMm) &&
      point.yMm <= Math.max(current.yMm, other.yMm);
    if (onSegment) return true;
    if ((current.yMm > point.yMm) !== (other.yMm > point.yMm)) {
      const intersectionX = (other.xMm - current.xMm) * (point.yMm - current.yMm) /
        (other.yMm - current.yMm) + current.xMm;
      if (point.xMm < intersectionX) inside = !inside;
    }
    previous = index;
  }
  return inside;
}
