import type { DecoderLibraryMatch, DecoderLibraryProfile } from "./decoderLibraryModel";

const normalized = (value: string) => value.trim().replace(/\s+/g, " ").toLowerCase();

export function selectECoSDecoderProfile(profiles: DecoderLibraryProfile[], label: string,
  modelLabel: string, source: string): {
  profile?: DecoderLibraryProfile; evidence?: DecoderLibraryMatch["ecosProfile"]; conflict: boolean;
} {
  if (normalized(label) !== "ecos" || !source.trim()) return { conflict: false };
  const candidates = profiles.filter(profile => profile.active && profile.kind === "model" &&
    [profile.name, profile.nameEn || "", ...profile.aliases].some(name => normalized(name) === normalized(source)));
  const evidence: NonNullable<DecoderLibraryMatch["ecosProfile"]> = { name: source, status: "unknown" };
  if (!candidates.length) return { evidence, conflict: false };
  const profile = candidates[0];
  if (profile.manufacturerId == null) return { evidence, conflict: false };
  const manualCompatible = normalized(modelLabel) === normalized(label) ||
    [profile.name, profile.nameEn || "", ...profile.aliases, ...profile.familyAliases]
      .some(name => normalized(name) === normalized(modelLabel));
  if (candidates.length !== 1 || !manualCompatible) {
    return { evidence: { ...evidence, status: "conflict" }, conflict: true };
  }
  return { profile, evidence: { ...evidence, status: "matched" }, conflict: false };
}
