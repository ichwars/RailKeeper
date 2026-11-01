import { bundledDecoderInventory, bundledDecoderProfiles } from "./decoderLibraryModel";
import type { DecoderDefinition, DecoderInventory, DecoderLibraryMatch, DecoderLibraryProfile, DecoderReading } from "./decoderLibraryModel";
import { identifyDecoder } from "./decoderIdentity";
import { selectECoSDecoderProfile } from "./ecosDecoderProfile";

const normalized = (value: string) => value.trim().replace(/\s+/g, " ").toLowerCase();
const includes = (aliases: string[], label: string) => aliases.some(alias => normalized(alias) === normalized(label));

export function firmwareAtLeast(actual: string, minimum: string) {
  const a = actual.split(".").map(Number), b = minimum.split(".").map(Number);
  for (let i = 0; i < Math.max(a.length, b.length); i++) {
    if ((a[i] || 0) !== (b[i] || 0)) return (a[i] || 0) > (b[i] || 0);
  }
  return true;
}

export function decoderDefinition(profiles: DecoderLibraryProfile[], match: DecoderLibraryMatch, value: DecoderReading) {
  if (match.status === "unsupported" || match.status === "unavailable") return undefined;
  // Legacy ECoS numbered entries carry no index-page identity.
  // Even a documented base definition cannot disambiguate a banked number by itself.
  if (isECoSReading(value) && value.indexCV31 == null && value.indexCV32 == null &&
    decoderNumberIsIndexed(profiles, match, value.cvNumber)) return undefined;
  const ordered = [
    ...profiles.filter(p => p.id === match.profileId && p.kind === "model" &&
      ["matched", "family"].includes(match.status) && match.firmwareCompatible !== false),
    ...profiles.filter(p => p.kind === "standard" && p.id === "dcc-basics")
  ].filter(p => p.active);
  const modelDefinitionsOnly = ordered.some(p => p.kind === "model" && (p.rules?.modelDefinitionsOnly || p.rules?.uhlenbrock76420 || p.rules?.roco10882 || p.rules?.piko56500));
  const customCurve = ordered.some(p => p.kind === "model" && p.rules?.motorCurveCV);
  for (const profile of ordered) {
    if (modelDefinitionsOnly && profile.kind === "standard") continue;
    if (profile.kind === "standard" && customCurve && (value.cvNumber === 6 ||
      (value.cvNumber >= 67 && value.cvNumber <= 94))) continue;
    const definition = profile.definitions.find(d => decoderCVMatches(d, value.cvNumber) && matchesDefinitionIndex(d, value) &&
      (!d.firmwareMin || !match.firmware || firmwareAtLeast(match.firmware, d.firmwareMin)));
    if (definition) return { definition, profile };
  }
  return undefined;
}

export const isECoSReading = (value: Pick<DecoderReading, "decoderProfile">) => normalized(value.decoderProfile || "") === "ecos";

export function decoderNumberIsIndexed(profiles: DecoderLibraryProfile[], match: DecoderLibraryMatch, cv: number) {
  return profiles.some(p => p.active && p.id === match.profileId && p.kind === "model" &&
    p.definitions.some(d => d.indexCV31 != null && decoderCVMatches(d, cv)));
}

export function decoderIndexUnknown(profiles: DecoderLibraryProfile[], match: DecoderLibraryMatch, value: DecoderReading) {
  if (value.indexCV31 != null || value.indexCV32 != null)
    return value.indexCV31 == null || value.indexCV32 == null;
  return isECoSReading(value) && value.cvNumber > 256 && !decoderDefinition(profiles, match, value);
}

export function decoderCVMatches(d: DecoderDefinition, cv: number) {
  return cv >= d.cv && cv <= (d.cvEnd ?? d.cv) && (cv - d.cv) % (d.cvStride || 1) === 0;
}

function matchesDefinitionIndex(d: DecoderDefinition, value: DecoderReading) {
  if (d.indexCV31 == null) return value.indexCV31 == null && value.indexCV32 == null;
  return value.indexCV31 === d.indexCV31 && value.indexCV32 != null &&
    value.indexCV32 >= (d.indexCV32 ?? 0) && value.indexCV32 <= (d.indexCV32End ?? d.indexCV32 ?? 0);
}

export function resolveDecoderLibrary(values: DecoderReading[], label: string, profiles = bundledDecoderProfiles,
  inventory: DecoderInventory | null = bundledDecoderInventory, modelLabel = label, ecosProfile = ""): DecoderLibraryMatch {
  const rows = values.filter(v => (v.decoderProfile || "").trim() === label.trim());
  const base = rows.filter(v => v.indexCV31 == null && v.indexCV32 == null);
  const manufacturers = new Set(base.filter(v => v.cvNumber === 8).map(v => v.value));
  const manufacturerId = base.find(v => v.cvNumber === 8)?.value;
  const firmware = base.find(v => v.cvNumber === 7)?.description?.match(/^LokProgrammer Decoder: .+ \(([\d.]+)\)$/)?.[1];
  const match: DecoderLibraryMatch = { status: "standard", manufacturerId, firmware, totalValues: rows.length, knownValues: 0 };
  if (rows.some(v => v.protocol?.trim() && !/^DCC(?: ?(?:14|28|128))?$/i.test(v.protocol.trim()))) {
    return { ...match, status: "unsupported" };
  }
  let source = selectECoSDecoderProfile(profiles, label, modelLabel, ecosProfile);
  if (!inventory && source.profile) {
    source = { evidence: { name: ecosProfile, status: "unknown" }, conflict: false };
  }
  if (source.evidence) match.ecosProfile = source.evidence;
  const manufacturerContext = manufacturerId ?? source.profile?.manufacturerId;
  const identityLabel = source.profile ? source.profile.model || source.profile.name : modelLabel;
  const { identity, conflict: identityConflict } = identifyDecoder(inventory, identityLabel, rows, manufacturerContext);
  let selected = profiles.find(p => p.active && p.kind === "standard" && p.id === "dcc-basics");
  const models = profiles.filter(p => p.active && p.kind === "model");
  const exact = source.profile ? [source.profile] : models.filter(p => includes(p.aliases, modelLabel));
  const candidates = exact.length ? exact : models.filter(p => includes(p.familyAliases, modelLabel));
  if (manufacturers.size > 1 || source.conflict || identityConflict || candidates.length > 1) match.status = "conflict";
  else if (candidates.length === 1) {
    const p = candidates[0];
    if (manufacturerContext == null) match.status = "suggested";
    else if (p.manufacturerId !== manufacturerContext) match.status = "conflict";
    else if (identity && !identity.profiles.includes(p.id)) match.status = "conflict";
    else { selected = p; match.status = exact.length ? "matched" : "family";
      match.matchedBy = source.profile ? "ecos-profile" : "name"; }
  }
  if (identity && match.status !== "conflict") {
    match.identification = identity.evidence;
    if (!match.matchedBy) {
      match.matchedBy = "cv";
      match.status = identity.evidence.scope === "family" ? "family" : "identified";
      const model = identity.evidence.scope === "model" ? models.find(p => p.id === identity.profileId && p.manufacturerId === manufacturerContext) : undefined;
      if (model) { selected = model; match.status = "matched"; }
    }
  }
  if (!selected) return { ...match, status: "unavailable" };
  if (match.status === "conflict" && source.profile) match.ecosProfile = { name: ecosProfile, status: "conflict" };
  if (selected.kind === "standard" && match.status === "standard" &&
    profiles.some(p => p.active && p.kind === "model" && p.manufacturerId === manufacturerId)) match.status = "manufacturer";
  Object.assign(match, { profileId: selected.id, profileName: match.status === "family" ? selected.family : selected.name, revision: selected.revision });
  if (match.matchedBy === "cv" && match.identification) match.profileName = match.identification.name;
  if (selected.kind === "model" && firmware && selected.firmwareMin) match.firmwareCompatible = firmwareAtLeast(firmware, selected.firmwareMin);
  match.knownValues = rows.filter(row => {
    const result = decoderDefinition(profiles, match, row);
    return result && !result.definition.reserved;
  }).length;
  return match;
}
