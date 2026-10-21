import { expect, it } from "vitest";
import { bundledDecoderInventory, bundledDecoderProfiles, type DecoderReading } from "./decoderLibraryModel";
import { decoderDefinition, resolveDecoderLibrary } from "./decoderLibrary";

it.each([
  ["LokPilot 5 micro DCC", 151, 9, 40], ["DH18A", 97, 56, 3],
  ["MN180N18", 145, 18, 128], ["STANDARD+ V2", 99, 147, 30],
  ["Kühn N45", 157, 54, 16], ["Uhlenbrock 76420", 85, 50, 32],
  ["Roco 10882", 145, 2, 4], ["PIKO 56500", 162, 141, 30]
] as const)("keeps the documented default and source for %s", (label, maker, cv, factory) => {
  const rows: DecoderReading[] = [{ cvNumber: 8, value: maker, decoderProfile: label }];
  const match = resolveDecoderLibrary(rows, label);
  const result = decoderDefinition(bundledDecoderProfiles, match, { cvNumber: cv, value: 0 });
  expect(result?.definition.factoryDefault).toBe(factory);
  expect(result?.definition.reference?.location).toBeTruthy();
  expect(result?.profile.sources[result.definition.reference!.sourceIndex].retrievedAt).toBe("2026-10-20");
  expect(rows).toEqual([{ cvNumber: 8, value: maker, decoderProfile: label }]);
});

it("does not confuse PIKO effect banks with RailComPlus or infer rules from the Roco brand", () => {
  const label = "PIKO 56500";
  const match = resolveDecoderLibrary([{ cvNumber: 8, value: 162, decoderProfile: label }], label);
  const row = { cvNumber: 267, value: 0, indexCV31: 18, indexCV32: 0 };
  expect(decoderDefinition(bundledDecoderProfiles, match, row)?.definition.reference).toBeDefined();
  expect(decoderDefinition(bundledDecoderProfiles, match, { ...row, indexCV31: 1, indexCV32: 1 })).toBeUndefined();
  const roco = resolveDecoderLibrary([{ cvNumber: 8, value: 161, decoderProfile: "Roco 10882" }], "Roco 10882");
  expect(roco.status).toBe("conflict");
  const reference = bundledDecoderInventory.models.find(m => m.documentationProfileId === "roco-10882");
  expect(reference?.brandOnly).toBe(true);
  expect(reference?.profileId).toBeUndefined();
  expect(reference?.identification).toBeUndefined();
});

it("keeps reserved ESU registers readable without counting them as interpreted functions", () => {
  const decoderProfile = "LokPilot 5 micro DCC";
  const rows: DecoderReading[] = [{ cvNumber: 8, value: 151, decoderProfile },
    ...[57, 58, 141, 142, 143, 250].map(cvNumber => ({ cvNumber, value: 125, decoderProfile }))];
  const before = structuredClone(rows);
  const match = resolveDecoderLibrary(rows, decoderProfile);
  expect(match.knownValues).toBe(1);
  expect(match.totalValues).toBe(7);
  for (const row of rows.slice(1)) {
    const result = decoderDefinition(bundledDecoderProfiles, match, row);
    expect(result?.definition.reserved).toBe(true);
    expect(result?.definition.factoryDefault).toBeUndefined();
    expect(result?.definition.bits).toBeUndefined();
    expect(decoderDefinition(bundledDecoderProfiles, match, { ...row, indexCV31: 1, indexCV32: 1 })).toBeUndefined();
  }
  expect(rows).toEqual(before);
});

it("does not borrow DCC curve point meanings for the verified DH48 characteristic", () => {
  const decoderProfile = "DH18A";
  const rows = [{ cvNumber: 8, value: 97, decoderProfile },
    ...[6, 67, 80, 94].map(cvNumber => ({ cvNumber, value: 100, decoderProfile }))];
  const match = resolveDecoderLibrary(rows, decoderProfile);
  expect(match.knownValues).toBe(1);
  for (const row of rows.slice(1)) expect(decoderDefinition(bundledDecoderProfiles, match, row)).toBeUndefined();
  expect(decoderDefinition(bundledDecoderProfiles, match, { cvNumber: 48, value: 5 })?.definition.name.de)
    .toBe("Krümmung der Motorkennlinie");
});

it("keeps reviewed ZIMO data separate from interface identification and raw import values", () => {
  const decoderProfile = "MN180N18";
  const profile = bundledDecoderProfiles.find(p => p.id === "zimo-mn180n18")!;
  expect(profile.firmwareMin).toBe("5.29");
  expect(profile.definitions).toHaveLength(68);
  const rows: DecoderReading[] = [{ cvNumber: 8, value: 145, decoderProfile },
    { cvNumber: 250, value: 127, decoderProfile }, { cvNumber: 5, value: 1, decoderProfile },
    { cvNumber: 6, value: 1, decoderProfile }, { cvNumber: 140, value: 4, decoderProfile },
    { cvNumber: 830, value: 128, decoderProfile }];
  const before = structuredClone(rows);
  const match = resolveDecoderLibrary(rows, decoderProfile);
  expect(match.status).toBe("matched");
  expect(match.firmware).toBeUndefined();
  expect(match.firmwareCompatible).toBeUndefined();
  for (const row of rows) {
    const result = decoderDefinition(bundledDecoderProfiles, match, row)!;
    expect(result.definition.reference?.sourceIndex).toBe(1);
    expect(result.definition.reference?.location).toMatch(/^S\./);
    expect(result.profile.sources[1].sha256).toBe("2e40762b6f4dcd5d761fe1507d70c8b997657b2fbe98d480ce63993f01f5ec43");
    expect(decoderDefinition(bundledDecoderProfiles, match, { ...row, indexCV31: 1, indexCV32: 1 })).toBeUndefined();
  }
  expect(decoderDefinition(bundledDecoderProfiles, match, rows[4])?.definition.allowedValues)
    .toEqual([0, 1, 2, 3, 11, 12, 13]);
  expect(rows).toEqual(before); // Invalid readings and special values stay unchanged, not repaired to defaults.
  const generic = resolveDecoderLibrary(rows.map(row => ({ ...row, decoderProfile: "Zentrale" })), "Zentrale");
  expect(generic.profileId).toBe("dcc-basics");
  expect(generic.identification?.scope).toBe("family");
  expect(decoderDefinition(bundledDecoderProfiles, generic, { cvNumber: 147, value: 100 })).toBeUndefined();
  const older = { ...match, firmware: "4.229", firmwareCompatible: false };
  expect(decoderDefinition(bundledDecoderProfiles, older, { cvNumber: 147, value: 100 })).toBeUndefined();
  expect(profile.mapping).toBeUndefined();
  expect(profile.rules?.constantBrake).toBeUndefined();
  for (const cvNumber of [58, 123, 151, 267, 309, 430]) {
    expect(decoderDefinition(bundledDecoderProfiles, match, { cvNumber, value: 1 })).toBeUndefined();
  }
});
