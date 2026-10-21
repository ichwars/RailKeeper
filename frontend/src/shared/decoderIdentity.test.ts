import { expect, it } from "vitest";
import cases from "../../../backend/internal/decoderlibrary/identity_cases_test.json";
import { bundledDecoderInventory, bundledDecoderProfiles, type DecoderReading } from "./decoderLibraryModel";
import { resolveDecoderLibrary, decoderDefinition } from "./decoderLibrary";

for (const test of cases) it(test.name, () => {
  const profiles = structuredClone(bundledDecoderProfiles).map(profile => ({ ...profile,
    active: !test.deactivate?.includes(profile.id) }));
  const rows: DecoderReading[] = test.rows.map((row: DecoderReading) => ({ ...row, decoderProfile: row.decoderProfile ?? test.label }));
  const before = structuredClone(rows);
  const result = resolveDecoderLibrary(rows, test.label, profiles, bundledDecoderInventory);
  const { profileId, ...expected } = test.expected;
  expect(result.profileId || "").toBe(profileId);
  expect(result).toMatchObject(expected);
  if (!test.expected.identification) expect(result.identification).toBeUndefined();
  if (test.expected.profileId === "dcc-basics") expect(decoderDefinition(profiles, result, { cvNumber: 27, value: 3 })).toBeUndefined();
  expect(rows).toEqual(before);
});
