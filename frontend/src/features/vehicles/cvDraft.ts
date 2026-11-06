import type { VehicleCVChange, VehicleCVValue, VehicleCVValueInput } from "../../shared/api";
import { cvValueKey, isValidCVValueInput } from "./cvImport";
import { reconcileDraft } from "./vehicleDraftRefresh";

export const maxManualCVChanges = 500;
export type CVDraftChange = { id: string; original: VehicleCVValue | null; value: VehicleCVValueInput | null };
export type CVDraftError = "invalid" | "occupied" | "limit" | "revision";

// Form and transport records stay separate. Never send a response ID, history or revision as CV fields.
export function cvDraftInput(value: VehicleCVValueInput): VehicleCVValueInput {
  return {
    cvNumber: Number(value.cvNumber), value: Number(value.value),
    decoderProfile: (value.decoderProfile || "").trim(),
    indexCV31: value.indexCV31 ?? null, indexCV32: value.indexCV32 ?? null,
    description: (value.description || "").trim(), category: (value.category || "").trim(),
    protocol: (value.protocol || "").trim(), sourceFileId: value.sourceFileId || "",
    sourceLine: value.sourceLine ?? null, sourceExportedAt: value.sourceExportedAt || ""
  };
}

export function sameCVInput(left: VehicleCVValueInput, right: VehicleCVValueInput) {
  return JSON.stringify(cvDraftInput(left)) === JSON.stringify(cvDraftInput(right));
}

export function cvDraftValues(stored: VehicleCVValue[], changes: CVDraftChange[], vehicleId: string) {
  const byOriginal = new Map(changes.flatMap(change => change.original ? [[change.original.id, change] as const] : []));
  const seen = new Set<string>();
  const pending = (change: CVDraftChange): VehicleCVValue[] => {
    seen.add(change.id);
    return change.value ? [{ ...(change.original || { vehicleId, createdAt: "", updatedAt: "" }),
      ...change.value, id: change.id }] : [];
  };
  const values = stored.flatMap(value => {
    const change = byOriginal.get(value.id);
    return change ? pending(change) : [value];
  });
  return [...values, ...changes.filter(change => !seen.has(change.id)).flatMap(pending)];
}

export function cvDraftCurrent(change: CVDraftChange, stored: VehicleCVValue[], changes: CVDraftChange[]) {
  if (change.original) return stored.find(value => value.id === change.original?.id);
  return cvDraftCollision(change, stored, changes);
}

export function cvDraftCollision(change: CVDraftChange, stored: VehicleCVValue[], changes: CVDraftChange[]) {
  const input = change.value;
  const occupied = input && stored.find(value => value.id !== change.original?.id && cvValueKey(value) === cvValueKey(input));
  if (occupied && changes.some(other => other.original?.id === occupied.id &&
    (!other.value || cvValueKey(other.value) !== cvValueKey(occupied))) &&
    !cvDraftOrder(changes).blocked.some(entry => entry.id === change.id)) return undefined;
  return occupied || undefined;
}

export function cvDraftConflicts(changes: CVDraftChange[], stored: VehicleCVValue[]) {
  return changes.filter(change => {
    const current = cvDraftCurrent(change, stored, changes);
    return (change.original && current?.revision !== change.original.revision) || Boolean(cvDraftCollision(change, stored, changes));
  });
}

export function stageCVInput(changes: CVDraftChange[], stored: VehicleCVValue[], vehicleId: string,
  form: VehicleCVValueInput, editingId: string | null, newID: string): { changes?: CVDraftChange[]; error?: CVDraftError } {
  let input = cvDraftInput(form);
  if (!isValidCVValueInput(input)) return { error: "invalid" };
  const byteLimits = { description: 1000, category: 80, protocol: 80, decoderProfile: 160, sourceFileId: 80 } as const;
  if (Object.entries(byteLimits).some(([key, limit]) =>
    new TextEncoder().encode(String(Reflect.get(input, key) || "")).length > limit)) return { error: "invalid" };
  const working = cvDraftValues(stored, changes, vehicleId);
  const target = editingId ? working.find(value => value.id === editingId)
    : working.find(value => cvValueKey(value) === cvValueKey(input));
  if (editingId && !target) return { error: "revision" };
  if (working.some(value => value.id !== target?.id && cvValueKey(value) === cvValueKey(input))) return { error: "occupied" };
  const id = target?.id || newID, existing = changes.find(change => change.id === id);
  const original = existing ? existing.original : target || null;
  if (original && !/^[0-9a-f]{64}$/.test(original.revision || "")) return { error: "revision" };
  if (!existing && changes.length >= maxManualCVChanges) return { error: "limit" };
  if (!editingId && target) {
    // Entering a known CV number from the add form changes its value without erasing unspecified metadata.
    // To clear an existing field deliberately, use that row's edit action.
    input = { ...cvDraftInput(target), ...Object.fromEntries(Object.entries(input).filter(([, value]) =>
      value !== "" && value != null)), cvNumber: input.cvNumber, value: input.value,
      decoderProfile: input.decoderProfile, indexCV31: input.indexCV31, indexCV32: input.indexCV32 };
  }
  const remaining = changes.filter(change => change.id !== id);
  const next = original && sameCVInput(original, input) ? remaining
    : [...remaining, { id, original, value: input }];
  return { changes: next };
}

export function stageCVDeletion(changes: CVDraftChange[], value: VehicleCVValue): CVDraftChange[] | null {
  const existing = changes.find(change => change.id === value.id);
  const original = existing ? existing.original : value;
  if (original && !/^[0-9a-f]{64}$/.test(original.revision || "")) return null;
  const remaining = changes.filter(change => change.id !== value.id);
  if (!original) return remaining;
  if (!existing && changes.length >= maxManualCVChanges) return null;
  return [...remaining, { id: value.id, original, value: null }];
}

export function cvDraftRequest(changes: CVDraftChange[]): VehicleCVChange[] {
  const plan = cvDraftOrder(changes);
  if (plan.blocked.length) throw new Error("CV draft contains cyclic identity moves");
  return plan.ordered.map(change => ({
      ...(change.original ? { id: change.original.id } : {}), expectedRevision: change.original?.revision || "",
      ...(change.value ? { value: cvDraftInput(change.value) } : { delete: true })
    }));
}

// Apply a move only after the explicitly selected row occupying its target has vacated it.
// Cycles remain conflicts; never clear or overwrite an intermediate identity implicitly.
function cvDraftOrder(changes: CVDraftChange[]) {
  const bySource = new Map(changes.flatMap(change => change.original ? [[cvValueKey(change.original), change] as const] : []));
  const dependencies = new Map(changes.map(change => {
    const occupant = change.value && bySource.get(cvValueKey(change.value));
    return [change.id, occupant && occupant.id !== change.id ? occupant.id : null] as const;
  }));
  const ordered: CVDraftChange[] = [], done = new Set<string>();
  let pending = [...changes];
  while (pending.length) {
    const ready = pending.filter(change => !dependencies.get(change.id) || done.has(dependencies.get(change.id) || ""));
    if (!ready.length) break;
    ready.sort((a, b) => Number(Boolean(a.value)) - Number(Boolean(b.value)) || Number(!a.original) - Number(!b.original));
    for (const change of ready) { ordered.push(change); done.add(change.id); }
    pending = pending.filter(change => !done.has(change.id));
  }
  return { ordered, blocked: pending };
}

export function rebaseCVChange(changes: CVDraftChange[], id: string, current: VehicleCVValue) {
  const change = changes.find(value => value.id === id);
  if (!change || !/^[0-9a-f]{64}$/.test(current.revision || "") ||
    changes.some(value => value.id !== id && value.original?.id === current.id)) return changes;
  const value = change.value && (change.original
    ? reconcileDraft(cvDraftInput(change.value), cvDraftInput(change.original), cvDraftInput(current)).value
    : { ...cvDraftInput(current), ...Object.fromEntries(Object.entries(cvDraftInput(change.value))
      .filter(([, field]) => field !== "" && field != null)), cvNumber: change.value.cvNumber,
      value: change.value.value, decoderProfile: change.value.decoderProfile,
      indexCV31: change.value.indexCV31, indexCV32: change.value.indexCV32 });
  return changes.flatMap(entry => entry.id !== id ? [entry]
    : value && sameCVInput(current, value) ? [] : [{ ...change, original: current, value }]);
}
