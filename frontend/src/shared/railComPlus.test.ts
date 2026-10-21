import { describe, expect, it } from "vitest";
import { railComPlus } from "./railComPlus";
import { resolveDecoderLibrary } from "./decoderLibrary";
import { bundledDecoderProfiles, type DecoderReading } from "./decoderLibraryModel";

const label = "LokPilot 5 micro DCC";
const base = { cvNumber: 8, value: 151, decoderProfile: label };
const indexed = (cvNumber: number, value: number, indexCV32 = 1) =>
  ({ cvNumber, value, indexCV31: 1, indexCV32, decoderProfile: label });
const rows = [base, ...[..."Testlok", "\0"].map((c, i) => indexed(261 + i, c.charCodeAt(0), 0)),
  indexed(257, 0), indexed(258, 3), indexed(259, 32), indexed(260, 2)];
const analyze = (values: DecoderReading[] = rows, profiles = bundledDecoderProfiles) => {
  const match = resolveDecoderLibrary(values, label, profiles);
  return railComPlus(values, profiles.find(p => p.id === match.profileId), match);
};

describe("reviewed RailComPlus metadata", () => {
  it.each([
    [32, false, false], [160, true, false], [176, true, true], [48, false, true]
  ])("matches the LokProgrammer F1 comparison export %i", (value, momentary, inverted) => {
    const firmware = { cvNumber: 7, value: 255, decoderProfile: label,
      description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.14.183)" };
    const result = analyze([base, firmware, indexed(259, value), indexed(260, 2)]);
    expect(result?.flagsVerification).toEqual({ firmware: "5.14.183", programmerVersion: "5.2.18" });
    expect(result?.functions[0]).toMatchObject({ functionKey: "F1", iconCode: 2, category: 1, momentary, inverted });
    expect(analyze([base, { ...firmware, description: firmware.description.replace("183", "184") },
      indexed(259, value), indexed(260, 2)])?.flagsVerification).toBeUndefined();
    expect(analyze([base, firmware, firmware, indexed(259, value), indexed(260, 2)])?.flagsVerification).toBeUndefined();
    expect(analyze([base, indexed(259, value), indexed(260, 2)])?.flagsVerification).toBeUndefined();
  });
  it("reads a terminated name and key descriptions without mutating readings", () => {
    const before = structuredClone(rows), result = analyze();
    expect(result?.name).toBe("Testlok");
    expect(result?.functions.map(row => [row.functionKey, row.iconCode])).toEqual([["F0", 3], ["F1", 2]]);
    expect(result?.incompleteFunctions).toBe(30);
    expect(rows).toEqual(before);
  });
  it.each([
    [96, false, false, "2026-10-20T02:52:38"], [224, true, false, "2026-10-20T02:53:04"],
    [240, true, true, "2026-10-20T02:54:19"], [112, false, true, "2026-10-20T02:54:43"]
  ])("matches the LokProgrammer F3 Logical comparison export %i independently of F1", (value, momentary, inverted, exportedAt) => {
    const values = [...rows, { cvNumber: 7, value: 255, decoderProfile: label,
      description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.14.183)" },
      { ...indexed(263, value), sourceLine: 181, sourceExportedAt: exportedAt },
      { ...indexed(264, 9), sourceLine: 182, sourceExportedAt: exportedAt }];
    const before = structuredClone(values), result = analyze(values);
    expect(result?.flagsVerification).toEqual({ firmware: "5.14.183", programmerVersion: "5.2.18" });
    expect(result?.functions.find(row => row.functionKey === "F3")).toMatchObject({
      firstCV: 263, firstValue: value, secondValue: 9, iconCode: 9, category: 3, momentary, inverted,
      sourceLine: 181, secondSourceLine: 182, sourceExportedAt: exportedAt
    });
    expect(result?.functions.find(row => row.functionKey === "F1"))
      .toMatchObject({ firstValue: 32, secondValue: 2, category: 1, momentary: false, inverted: false });
    expect(values).toEqual(before);
  });
  it.each([
    [0, false, false, "2026-10-20T03:14:07"], [128, true, false, "2026-10-20T03:14:39"],
    [144, true, true, "2026-10-20T03:15:02"], [16, false, true, "2026-10-20T03:15:37"]
  ])("matches the LokProgrammer F0 Light comparison export %i independently of F1 and F3",
    (value, momentary, inverted, exportedAt) => {
      const values = [base, { cvNumber: 7, value: 255, decoderProfile: label,
        description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.14.183)" },
        { ...indexed(257, value), sourceLine: 175, sourceExportedAt: exportedAt },
        { ...indexed(258, 3), sourceLine: 176, sourceExportedAt: exportedAt },
        indexed(259, 32), indexed(260, 2), indexed(263, 96), indexed(264, 9)];
      const before = structuredClone(values), result = analyze(values);
      expect(result?.flagsVerification).toEqual({ firmware: "5.14.183", programmerVersion: "5.2.18" });
      expect(result?.functions.find(row => row.functionKey === "F0")).toMatchObject({
        firstCV: 257, firstValue: value, secondValue: 3, iconCode: 3, category: 0, momentary, inverted,
        sourceLine: 175, secondSourceLine: 176, sourceExportedAt: exportedAt
      });
      expect(result?.functions.find(row => row.functionKey === "F1"))
        .toMatchObject({ firstValue: 32, secondValue: 2, category: 1, momentary: false, inverted: false });
      expect(result?.functions.find(row => row.functionKey === "F3"))
        .toMatchObject({ firstValue: 96, secondValue: 9, category: 3, momentary: false, inverted: false });
      expect(values).toEqual(before);
    });
  it("separates a 12-bit symbol from category, momentary mode and inversion", () => {
    const value = analyze([base, indexed(257, 0xf3), indexed(258, 5)])?.functions[0];
    expect(value).toMatchObject({ iconCode: 773, category: 3, momentary: true, inverted: true });
    expect(analyze([base, indexed(257, 96), indexed(258, 9)])?.functions[0])
      .toMatchObject({ iconCode: 9, category: 3, momentary: false, inverted: false });
  });
  it("does not create functions from unassigned descriptors", () => {
    const result = analyze([base, ...Array.from({ length: 64 }, (_, i) => indexed(257 + i, i % 2))]);
    expect(result?.functions).toEqual([]);
    expect(result?.unassignedFunctions).toBe(32);
  });
  it("does not substitute zero for missing or duplicate bytes", () => {
    expect(analyze([base, indexed(257, 32)])?.functions).toEqual([]);
    expect(analyze([base, indexed(257, 32), indexed(257, 32), indexed(258, 3)])?.functions).toEqual([]);
    expect(analyze(rows.filter(row => row.cvNumber !== 264))?.name).toBeUndefined();
  });
  it("keeps unverified encoding and an unterminated name unknown", () => {
    expect(analyze([base, indexed(261, 196, 0)])?.nameIssue).toBe("encoding");
    expect(analyze([base, indexed(261, 65, 0)])?.nameIssue).toBe("incomplete");
  });
  it("does not combine bytes with contradictory export dates", () => {
    const values = [base, { ...indexed(257, 32), sourceExportedAt: "2026-10-01T12:00:00" },
      { ...indexed(258, 3), sourceExportedAt: "2026-10-02T12:00:00" }];
    expect(analyze(values)?.functions).toEqual([]);
    const name = [base, { ...indexed(261, 65, 0), sourceExportedAt: "2026-10-01T12:00:00" },
      { ...indexed(262, 0, 0), sourceExportedAt: "2026-10-02T12:00:00" }];
    expect(analyze(name)).toMatchObject({ nameIssue: "incomplete" });
    expect(analyze(name)?.name).toBeUndefined();
  });
  it("requires the active supported model, manufacturer and protocol", () => {
    expect(analyze(rows.map(row => row.cvNumber === 8 ? { ...row, value: 99 } : row))).toBeUndefined();
    expect(analyze(rows.map(row => ({ ...row, protocol: "Motorola" })))).toBeUndefined();
    expect(analyze(rows, bundledDecoderProfiles.map(p => ({ ...p, active: false })))).toBeUndefined();
    const older = [...rows, { cvNumber: 7, value: 255, decoderProfile: label,
      description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.2.100)" }];
    expect(analyze(older)).toBeUndefined();
  });
  it("never borrows bytes from a different profile or index page", () => {
    expect(analyze([base, ...rows.slice(1).map(row => ({ ...row, indexCV31: 16 }))])?.functions).toEqual([]);
    expect(analyze([...rows, { ...indexed(258, 9), decoderProfile: "Other" }])).toBeUndefined();
  });
});
