import type { VehicleCVFilePreview, VehicleFunctionInput } from "../../shared/api";
import { isValidFunctionMapping } from "./cvImport";
import { functionFileSource } from "./functionDraft";
import type { FunctionEditState } from "./vehicleTransforms";
import { emptyFunctionEdit } from "./vehicleViewModel";

export function supplementFileFunctionDraft(current: FunctionEditState, previews: VehicleCVFilePreview[]) {
  const groups = new Map<string, { name: string; functionType: string; files: Set<string>; conflict: boolean }>();
  const blocked = new Set<string>();
  for (const preview of previews) {
    for (const row of preview.suggestedFunctions || []) {
      if (!isValidFunctionMapping(row)) { blocked.add(row.functionKey); continue; }
      const name = row.name.trim(), functionType = row.functionType || "standard";
      const previous = groups.get(row.functionKey);
      if (previous) {
        previous.conflict ||= previous.name !== name || previous.functionType !== functionType;
        previous.files.add(preview.fileName);
      } else groups.set(row.functionKey, { name, functionType, files: new Set([preview.fileName]), conflict: false });
    }
  }
  const edits = { ...current };
  let applied = 0;
  for (const [key, suggestion] of groups) {
    if (suggestion.conflict || blocked.has(key)) { blocked.add(key); continue; }
    const existing = current[key] || { ...emptyFunctionEdit(key), name: "", symbolKey: "", functionType: "standard" };
    // An intentional edit/deletion and all assigned fields remain protected.
    if (existing.dirty || existing.deleted) continue;
    const patch: Partial<VehicleFunctionInput> = {};
    if (!existing.name?.trim() && suggestion.name) patch.name = suggestion.name;
    if (!existing.persisted && (!existing.functionType || existing.functionType === "standard") &&
      suggestion.functionType !== (existing.functionType || "standard")) patch.functionType = suggestion.functionType;
    if (!Object.keys(patch).length) continue;
    const notes = [existing.notes, ...[...suggestion.files].map(file => functionFileSource(file, key))]
      .filter(Boolean).join("\n");
    if (new TextEncoder().encode(notes).length > 1000 || new TextEncoder().encode(patch.name || "").length > 120) {
      blocked.add(key);
      continue;
    }
    edits[key] = { ...existing, ...patch, notes, dirty: true };
    applied++;
  }
  return { edits, applied, blocked: [...blocked] };
}
