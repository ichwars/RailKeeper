import { describe, expect, it } from "vitest";
import { bundledDecoderProfiles, type DecoderReading } from "./decoderLibraryModel";
import { decoderDefinition, resolveDecoderLibrary } from "./decoderLibrary";
import { decoderAnalysis } from "../features/vehicles/decoderAnalysis";

const label = "LokPilot 5 micro DCC";
const rows: DecoderReading[] = [
  { cvNumber: 8, value: 151, decoderProfile: label },
  { cvNumber: 1, value: 3, decoderProfile: label },
  { cvNumber: 27, value: 24, decoderProfile: label },
  { cvNumber: 257, value: 20, indexCV31: 16, indexCV32: 3, decoderProfile: label }
];

describe("decoder library resolution", () => {
  it("requires a model name and matching manufacturer, preserving separate profiles", () => {
    expect(resolveDecoderLibrary(rows, label)).toMatchObject({ status: "matched", knownValues: 4 });
    expect(resolveDecoderLibrary(rows.map(r => ({ ...r, decoderProfile: "Unknown" })), "Unknown"))
      .toMatchObject({ status: "manufacturer", profileId: "dcc-basics", knownValues: 2 });
    expect(resolveDecoderLibrary(rows.slice(1), label)).toMatchObject({ status: "suggested", knownValues: 1 });
    expect(resolveDecoderLibrary([...rows, { cvNumber: 8, value: 99, decoderProfile: "Other" }], label))
      .toMatchObject({ status: "matched", totalValues: 4 });
    expect(resolveDecoderLibrary(rows.map(r => ({ ...r, decoderProfile: "LokPilot 5" })), "LokPilot 5"))
      .toMatchObject({ status: "family", profileName: "LokPilot 5" });
  });

  it("rejects conflicting, ambiguous and non-DCC identities", () => {
    expect(resolveDecoderLibrary([{ ...rows[0], value: 99 }, ...rows.slice(1)], label).status).toBe("conflict");
    expect(resolveDecoderLibrary([...rows, { ...rows[0], value: 99 }], label).status).toBe("conflict");
    expect(resolveDecoderLibrary(rows.map(r => ({ ...r, protocol: "MM" })), label))
      .toMatchObject({ status: "unsupported", knownValues: 0 });
    const ambiguous = [...bundledDecoderProfiles, { ...bundledDecoderProfiles[1], id: "another-model" }];
    expect(resolveDecoderLibrary(rows, label, ambiguous)).toMatchObject({ status: "conflict", knownValues: 2 });
  });

  it("keeps base CVs, zero pages, unknown addresses and firmware scope distinct", () => {
    const match = resolveDecoderLibrary(rows, label);
    expect(decoderDefinition(bundledDecoderProfiles, match, { cvNumber: 1, value: 3, indexCV31: 0, indexCV32: 0 })).toBeUndefined();
    expect(decoderDefinition(bundledDecoderProfiles, match, { cvNumber: 291, value: 31, indexCV31: 16, indexCV32: 0 })).toBeDefined();
    expect(decoderDefinition(bundledDecoderProfiles, match, { ...rows[3], indexCV32: 13 })?.definition.name.de)
      .toContain("Zufallsfunktion 1");
    expect(decoderDefinition(bundledDecoderProfiles, match, { ...rows[3], indexCV32: 14 })).toBeUndefined();
    const old = [...rows, { cvNumber: 7, value: 255, decoderProfile: label,
      description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.2.100)" }];
    expect(resolveDecoderLibrary(old, label)).toMatchObject({ firmwareCompatible: false, knownValues: 3 });
    expect(decoderAnalysis(old, label).abc).toBeUndefined();
  });

  it("honors disabled definitions without changing raw readings", () => {
    const copy = JSON.stringify(rows);
    const disabled = bundledDecoderProfiles.map(p => ({ ...p, active: false }));
    expect(resolveDecoderLibrary(rows, label, disabled)).toMatchObject({ status: "unavailable", knownValues: 0 });
    expect(decoderAnalysis([...rows, { cvNumber: 29, value: 14, decoderProfile: label }], label, disabled).address).toBeUndefined();
    expect(JSON.stringify(rows)).toBe(copy);
  });
});
