import type { DecoderInventory, DecoderLibraryMatch, DecoderReading } from "./decoderLibraryModel";

const normalized = (value: string) => value.trim().replace(/\s+/g, " ").toLowerCase();
type Evidence = NonNullable<DecoderLibraryMatch["identification"]>;
export type IdentityMatch = { evidence: Evidence; profileId?: string; profiles: Array<string | undefined> };

export function identifyDecoder(inventory: DecoderInventory | null, label: string, rows: DecoderReading[], manufacturer?: number): {
  identity?: IdentityMatch; conflict: boolean;
} {
  if (!inventory || manufacturer == null) return { conflict: false };
  const base = new Map<number, number>(), conflicts = new Set<number>();
  for (const row of rows) {
    if (row.indexCV31 != null || row.indexCV32 != null) continue;
    if (base.has(row.cvNumber) && base.get(row.cvNumber) !== row.value) conflicts.add(row.cvNumber);
    base.set(row.cvNumber, row.value);
  }
  const candidates: IdentityMatch[] = [];
  for (const model of inventory.models) {
    const family = inventory.families.find(f => f.id === model.familyId);
    const maker = inventory.manufacturers.find(m => m.id === family?.manufacturerId);
    if (!maker) continue;
    const named = [model.name, `${maker.name} ${model.name}`].some(name => normalized(name) === normalized(label));
    if (named && !model.brandOnly && maker.cv8 !== manufacturer) return { conflict: true };
    const rule = model.identification;
    if (!rule || maker.cv8 !== manufacturer) continue;
    if (conflicts.has(rule.cv)) return { conflict: true };
    const value = base.get(rule.cv);
    if (value == null) continue;
    if (!rule.values.includes(value)) {
      if (named) return { conflict: true };
      continue;
    }
    candidates.push({ evidence: { scope: rule.scope, name: rule.name, cvNumber: rule.cv, value,
      ...(rule.scope === "model" ? { modelId: model.id } : {}) }, profileId: model.profileId, profiles: [] });
  }
  if (!candidates.length) return { conflict: false };
  const first = candidates[0];
  if (candidates.slice(1).some(({ evidence: candidate }) => first.evidence.scope !== "family" || candidate.scope !== "family" ||
    candidate.name !== first.evidence.name || candidate.cvNumber !== first.evidence.cvNumber || candidate.value !== first.evidence.value)) {
    return { conflict: true };
  }
  first.profiles = candidates.map(candidate => candidate.profileId);
  return { identity: first, conflict: false };
}
