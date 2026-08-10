import type { TrackGeometryDefinition } from "../../shared/api";

export function trackGeometryManufacturer(geometry: Pick<TrackGeometryDefinition, "manufacturer">) {
  return geometry.manufacturer.trim();
}

export function trackGeometryLabel(
  geometry: Pick<TrackGeometryDefinition, "manufacturer" | "articleNumber" | "name">
) {
  return `${trackGeometryManufacturer(geometry)} ${geometry.articleNumber} · ${geometry.name}`;
}
