import type { VehicleFunctionChange, VehicleFunctionInput } from "../../shared/api";
import type { FunctionEdit, FunctionEditState } from "./vehicleTransforms";

export function functionInput(edit: FunctionEdit): VehicleFunctionInput {
  return {
    name: edit.name || "", symbolKey: edit.symbolKey || "", functionType: edit.functionType || "standard",
    mode: edit.mode || "dauer", directionDependent: edit.directionDependent ?? null,
    inverted: edit.inverted ?? null, legacyDirectionDependent: edit.legacyDirectionDependent ?? null,
    flagsVersion: 2, notes: edit.notes || ""
  };
}

export function functionHasContent(edit: FunctionEdit) {
  return Boolean(edit.persisted || edit.name?.trim() || edit.symbolKey || edit.notes?.trim() ||
    edit.modeAssigned || edit.inverted != null || edit.directionDependent != null);
}

export function functionDraftChanges(edits: FunctionEditState): VehicleFunctionChange[] {
  return Object.entries(edits).flatMap<VehicleFunctionChange>(([functionKey, edit]) => {
    if (edit.deleted) return edit.persisted
      ? [{ functionKey, expectedRevision: edit.revision || "", delete: true }] : [];
    if ((edit.persisted && !edit.dirty) || !functionHasContent(edit)) return [];
    return [{ functionKey, expectedRevision: edit.revision || "", value: functionInput(edit) }];
  });
}

export function invalidFunctionChanges(changes: VehicleFunctionChange[]) {
  const bytes = (text = "") => new TextEncoder().encode(text).length;
  return changes.filter(change => change.value && (bytes(change.value.name) > 120 ||
    bytes(change.value.symbolKey) > 80 || bytes(change.value.notes) > 1000)).map(change => change.functionKey);
}

// Presentation only: retain every original note byte until the user edits their own text.
// Match only the two explicit source-line formats emitted by RailKeeper for this key.
export function functionNoteParts(notes = "", functionKey: string) {
  const key = /^F(\d+)$/.exec(functionKey), cv = key ? 257 + 2 * Number(key[1]) : -1;
  const isSource = (line: string) => (line.startsWith("RailComPlus · ") &&
    line.includes("CV31=1 / CV32=1") && new RegExp(`CV${cv}=\\d+`).test(line) &&
    new RegExp(`CV${cv + 1}=\\d+`).test(line)) || line.startsWith(`CV-Mapping · ${functionKey} · `);
  const lines = notes.split("\n");
  return { own: lines.filter(line => !isSource(line)).join("\n"), sources: lines.filter(isSource) };
}

export function withFunctionOwnNote(notes: string | undefined, functionKey: string, own: string) {
  const { sources } = functionNoteParts(notes, functionKey);
  return [own, ...sources].filter(Boolean).join("\n");
}
