import type { Vehicle } from "../../shared/api";
import { vehicleImagesToPending, type PendingArticleImage } from "./vehicleTransforms";

export type DraftRefreshConflict = { field: string; local: unknown; saved: unknown };

// Core-write responses contain saved images/functions but omit independent side records.
// Copy only those side records; absent optional core fields must still clear their old values.
export function coreSavedDetail(detail: Vehicle, previous: Vehicle): Vehicle {
  return {
    ...detail,
    attachments: detail.attachments ?? previous.attachments,
    maintenance: detail.maintenance ?? previous.maintenance,
    spareParts: detail.spareParts ?? previous.spareParts,
    cvValues: detail.cvValues ?? previous.cvValues,
    cvFiles: detail.cvFiles ?? previous.cvFiles,
    externalMappings: detail.externalMappings ?? previous.externalMappings,
    ecosDecoderProfile: detail.ecosDecoderProfile ?? previous.ecosDecoderProfile,
    vehicleSet: detail.vehicleSet ?? previous.vehicleSet
  };
}

export function draftValue(value: object, field: string): unknown {
  return Object.hasOwn(value, field) ? Reflect.get(value, field) : undefined;
}

export function sameDraftValue(left: unknown, right: unknown) {
  return Object.is(left, right) || (left !== null && right !== null &&
    typeof left === "object" && typeof right === "object" && JSON.stringify(left) === JSON.stringify(right));
}

// A side action may refresh the aggregate while the user is still typing. Adopt server changes
// only for untouched fields and keep both values when the same field changed on both sides.
export function reconcileDraft<T extends object>(current: T, previous: T, incoming: T) {
  let value = { ...incoming };
  const conflicts: DraftRefreshConflict[] = [];
  for (const [field, local] of Object.entries(current)) {
    const before = draftValue(previous, field), saved = draftValue(incoming, field);
    if (sameDraftValue(local, before)) continue;
    value = { ...value, [field]: local };
    if (!sameDraftValue(saved, before) && !sameDraftValue(local, saved)) {
      conflicts.push({ field, local, saved });
    }
  }
  return { value, conflicts };
}

export function retainRefreshConflicts(
  previous: DraftRefreshConflict[], current: object, incoming: object, conflicts: DraftRefreshConflict[]
) {
  const retained = new Map(previous.map(conflict => [conflict.field, {
    field: conflict.field, local: draftValue(current, conflict.field), saved: draftValue(incoming, conflict.field)
  }]));
  for (const conflict of conflicts) retained.set(conflict.field, conflict);
  return [...retained.values()].filter(conflict => !sameDraftValue(conflict.local, conflict.saved));
}

export function reconcileImageDraft(current: PendingArticleImage[], previous: Vehicle, incoming: Vehicle) {
  const base = vehicleImagesToPending(previous), latest = vehicleImagesToPending(incoming);
  const baseByID = new Map(base.map(image => [image.id, image]));
  const currentByID = new Map(current.map(image => [image.id, image]));
  const latestByID = new Map(latest.map(image => [image.id, image]));
  const shared = new Set(current.filter(image => baseByID.has(image.id) && latestByID.has(image.id)).map(image => image.id));
  const reordered = !sameDraftValue(current.filter(image => shared.has(image.id)).map(image => image.id),
    base.filter(image => shared.has(image.id)).map(image => image.id));
  const merged = latest.map(image => {
    const local = currentByID.get(image.id), before = baseByID.get(image.id);
    if (!local) return image;
    return before ? reconcileDraft(local, before, image).value : { ...image, ...local };
  });
  const byID = new Map(merged.map(image => [image.id, image]));
  const retained = current.filter(image => !latestByID.has(image.id) &&
    (!image.persisted || !baseByID.has(image.id) || !sameDraftValue(image, baseByID.get(image.id))));
  const ordered = reordered
    ? [...current.flatMap(image => { const value = byID.get(image.id); return value ? [value] : []; }),
      ...merged.filter(image => !currentByID.has(image.id))]
    : merged;
  const result = [...ordered, ...retained];
  const localPrimary = current.find(image => image.isPrimary)?.id;
  const basePrimary = base.find(image => image.isPrimary)?.id;
  const primary = localPrimary !== basePrimary && result.some(image => image.id === localPrimary)
    ? localPrimary : latest.find(image => image.isPrimary)?.id || result.find(image => image.isPrimary)?.id;
  return result.map(image => ({ ...image, isPrimary: image.id === primary }));
}
