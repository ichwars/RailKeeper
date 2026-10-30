import { expect, it } from "vitest";
import cases from "../../../backend/internal/decoderlibrary/ecos_profile_cases_test.json";
import { bundledDecoderInventory, bundledDecoderProfiles, type DecoderReading } from "./decoderLibraryModel";
import { decoderDefinition, resolveDecoderLibrary } from "./decoderLibrary";

for (const fixture of cases) it(fixture.name, () => {
  const profiles = structuredClone(bundledDecoderProfiles);
  for (const profile of [...profiles]) {
    profile.active = !fixture.deactivate.some(id => id === profile.id);
    if (profile.id === fixture.duplicate) profiles.push({ ...profile, id: profile.id + "-ambiguous" });
  }
  const rows: DecoderReading[] = fixture.rows.map(row => ({ ...row, decoderProfile: fixture.label }));
  const before = structuredClone(rows);
  const match = resolveDecoderLibrary(rows, fixture.label, profiles, fixture.inventoryMissing ? null : bundledDecoderInventory,
    fixture.modelLabel, fixture.ecosProfile);
  expect(match.status).toBe(fixture.expected.status);
  expect(match.profileId).toBe(fixture.expected.profileId);
  expect(match.matchedBy || "").toBe(fixture.expected.matchedBy);
  expect(match.ecosProfile?.status ?? null).toBe(fixture.expected.sourceStatus);
  if (fixture.expected.manufacturerId !== "conflicting") {
    expect(match.manufacturerId ?? null).toBe(fixture.expected.manufacturerId);
  }
  expect(Boolean(decoderDefinition(profiles, match, { ...fixture.probe, decoderProfile: fixture.label })))
    .toBe(fixture.definition);
  expect(rows).toEqual(before);
});
