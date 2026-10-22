import type {
  DecoderCVReference, DecoderCVReferenceRow, DecoderInventory, DecoderModel, DecoderReading
} from "../../shared/decoderLibraryModel";
import type { DecoderAnalysis } from "./decoderAnalysis";
import { isECoSReading } from "../../shared/decoderLibrary";

export type ManufacturerCVHint = {
  model: DecoderModel; reference: DecoderCVReference; row: DecoderCVReferenceRow;
  unindexedAddressConfirmed?: boolean;
};

const normalized = (text: string) => text.trim().replace(/\s+/g, " ").toLowerCase();
const key = (cv: number, cv31?: number | null, cv32?: number | null) =>
  cv31 == null && cv32 == null ? `${cv}/base` : `${cv}/${cv31}/${cv32}`;

// A document association is deliberately weaker than an executable decoder profile.
export function manufacturerReferenceModel(analysis: DecoderAnalysis, inventory: DecoderInventory | undefined,
  label: string, manualType: string): DecoderModel | undefined {
  if (!inventory || ["conflict", "unsupported", "unavailable"].includes(analysis.libraryMatch.status)) return;
  const id = analysis.libraryMatch.identification?.modelId;
  const identified = inventory.models.find(model => model.id === id);
  if (identified) return identified;
  const matches = inventory.models.filter(model => {
    const family = inventory.families.find(family => family.id === model.familyId);
    const maker = inventory.manufacturers.find(maker => maker.id === family?.manufacturerId);
    if (!maker) return false;
    const cv8 = analysis.libraryMatch.manufacturerId;
    if (cv8 != null && maker.cv8 !== cv8) return false;
    const type = model.identification;
    if (type && analysis.rows.some(row => row.cvNumber === type.cv && row.indexCV31 == null &&
      row.indexCV32 == null && !type.values.includes(row.value))) return false;
    return [label, manualType].some(text => text.trim() &&
      [model.name, `${maker.name} ${model.name}`, ...model.articleNumbers].some(name => normalized(name) === normalized(text)));
  });
  return matches.length === 1 ? matches[0] : undefined;
}

export function manufacturerCVHints(model: DecoderModel, references: DecoderCVReference[]) {
  const hints = new Map<string, ManufacturerCVHint>();
  const conflicts = new Set<string>();
  const nonBaseNumbers = new Set<number>();
  for (const reference of references) {
    if (!reference.active || reference.status === "unavailable" || !reference.modelIds.includes(model.id)) continue;
    for (const row of reference.rows) {
      if (!reference.sources[row.sourceIndex]) continue;
      const end = row.cvEnd ?? row.cv;
      if (!Number.isInteger(row.cv) || !Number.isInteger(end) || row.cv < 1 || end > 1024 || end < row.cv) continue;
      if (row.indexScope !== "base") for (let cv = row.cv; cv <= end; cv++) nonBaseNumbers.add(cv);
      if (row.indexScope === "unspecified") continue;
      if (row.indexScope === "indexed" && (row.indexCV31 == null || row.indexCV32 == null)) continue;
      if (row.indexScope === "base" && (row.indexCV31 != null || row.indexCV32 != null)) continue;
      for (let cv = row.cv; cv <= end; cv++) {
        const identity = key(cv, row.indexCV31, row.indexCV32);
        const previous = hints.get(identity);
        // Do not resolve contradictory table headings or variant/default qualifications by order.
        if (previous && (["label", "valueRange", "factoryValue", "applicability"].some(field =>
          previous.row[field as keyof DecoderCVReferenceRow] !== row[field as keyof DecoderCVReferenceRow]) ||
          previous.reference.note.de !== reference.note.de || previous.reference.note.en !== reference.note.en)) {
          conflicts.add(identity);
        }
        if (!previous) hints.set(identity, { model, reference, row });
      }
    }
  }
  for (const identity of conflicts) hints.delete(identity);
  return (value: DecoderReading) => {
    const hint = hints.get(key(value.cvNumber, value.indexCV31, value.indexCV32));
    if (!hint || !isECoSReading(value) || value.indexCV31 != null || value.indexCV32 != null) return hint;
    if (nonBaseNumbers.has(value.cvNumber)) return undefined;
    return { ...hint, unindexedAddressConfirmed: hint.row.indexScope === "base" && hint.row.applicability === "model" };
  };
}
