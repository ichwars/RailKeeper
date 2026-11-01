import { describe, expect, it } from "vitest";
import type { ECoSDecoderProfileEvidence, VehicleCVValueInput } from "../../shared/api";
import { bundledDecoderInventory, bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import type { FunctionEdit } from "./vehicleTransforms";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { functionCVSuggestion } from "./functionCVSuggestions";

const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);
const base = (cvNumber: number, value: number): VehicleCVValueInput => ({ cvNumber, value, decoderProfile: "ECoS" });
const indexed = (cvNumber: number, value: number, indexCV32: number) => ({ ...base(cvNumber, value), indexCV31: 16, indexCV32 });
const evidence = (name: string): ECoSDecoderProfileEvidence => ({ name, objectId: "1001", host: "192.0.2.10", readAt: "2026-10-27T10:00:00Z" });
const analyze = (name: string, rows: VehicleCVValueInput[], manual = "") =>
  decoderAnalysis(rows, "ECoS", bundledDecoderProfiles, bundledDecoderInventory, manual, evidence(name));

describe("reviewed ESU model rules", () => {
  it.each([
    ["LokPilot 5 nano DCC", "esu-lokpilot-5-nano-dcc"],
    ["LokPilot V4.0 micro DCC", "esu-lokpilot-v4-micro-dcc"],
    ["LokPilot micro V4.0", "esu-lokpilot-v4-micro"],
    ["LokPilot Standard", "esu-lokpilot-standard"]
  ])("uses the exact saved %s profile without inventing a manufacturer CV", (name, id) => {
    const rows = [base(29, 14)];
    const a = analyze(name, rows);
    expect(a.libraryMatch.profileId).toBe(id);
    expect(a.libraryMatch.matchedBy).toBe("ecos-profile");
    expect(a.libraryMatch.manufacturerId).toBeUndefined();
    expect(rows).toEqual([base(29, 14)]);
    expect(analyze(name, rows, "D&H DH18A").modelRules).toBe(false);
    expect(analyze(name, [base(8, 145), ...rows]).modelRules).toBe(false);
  });

  it("keeps variant CV meanings and undocumented output pages separate", () => {
    const standard = analyze("LokPilot Standard", [base(114, 31)]);
    const nano = analyze("LokPilot 5 nano DCC", [base(114, 31), indexed(259, 1, 0)]);
    expect(standard.definition(base(114, 31))?.definition.name.en).toContain("Rear light");
    expect(nano.definition(base(114, 31))?.definition.name.en).toContain("fade-in");
    expect(nano.definition(indexed(259, 1, 0))).toBeUndefined();
    expect(nano.railComPlus?.flagsVerification).toBeUndefined();
  });

  it("uses the Standard direction registers and logical output bits", () => {
    const a = analyze("LokPilot Standard", [base(33, 1), base(34, 2), base(35, 4), base(37, 64),
      base(38, 128), base(47, 8), base(129, 64), base(135, 128)]);
    const mappings = decoderMapping(a, t);
    expect(mappings.map(m => [m.functionKey, m.direction])).toEqual([
      ["F0", "forward"], ["F0", "reverse"], ["F1", "forward"], ["F3", "both"],
      ["F4", "forward"], ["F1", "reverse"], ["F13", "both"], ["F4", "reverse"]
    ]);
    expect(mappings.find(m => m.functionKey === "F3")?.actions).toEqual(["Rangiergang"]);
    expect(mappings.find(m => m.functionKey === "F4")?.actions).toEqual(["Beschleunigungs-/Bremsverzögerung aus"]);
    expect(decoderMapping(analyze("LokPilot Standard", [indexed(37, 64, 0), base(38, 256)]), t)).toEqual([]);
  });

  it("requires all V4 condition and action bytes and never reads the V5 layout", () => {
    const values = [...Array.from({ length: 9 }, (_, i) => indexed(257 + i, i === 1 ? 4 : 0, 2)),
      ...[0, 0, 2, 0, 0].map((value, i) => indexed(266 + i, value, 2))];
    expect(decoderMapping(analyze("LokPilot V4.0 micro DCC", values), t)).toMatchObject([
      { functionKey: "F3", actions: ["Rangiergang"], warning: false }
    ]);
    const noDelay = values.map(r => r.cvNumber === 268 ? { ...r, value: 1 } : r);
    expect(decoderMapping(analyze("LokPilot V4.0 micro DCC", noDelay), t)[0].actions)
      .toEqual(["Beschleunigungs-/Bremsverzögerung aus"]);
    for (const cv of [265, 266, 267, 268, 269, 270]) {
      expect(decoderMapping(analyze("LokPilot V4.0 micro DCC", values.filter(r => r.cvNumber !== cv)), t)).toEqual([]);
    }
    expect(decoderMapping(analyze("LokPilot V4.0 micro DCC", values.map(r => ({ ...r, indexCV32: 8 }))), t)).toEqual([]);
    expect(decoderMapping(analyze("LokPilot 5 nano DCC", values), t)).toEqual([]);
    expect(decoderMapping(analyze("LokPilot V4.0 micro DCC", [...values, values[0]]), t)).toEqual([]);
  });

  it("fills an empty name but keeps all manual function information", () => {
    const a = analyze("LokPilot Standard", [base(37, 64)]);
    const empty: FunctionEdit = { functionType: "standard", mode: "toggle", symbolKey: "", name: "", notes: "" };
    expect(functionCVSuggestion(a, "F3", empty, [], t).patch?.name).toBe("Rangiergang");
    const manual: FunctionEdit = { ...empty, persisted: true, name: "Eigene Rangierfunktion", symbolKey: "club",
      mode: "momentary", inverted: true, notes: "Eigene Zuordnung" };
    const before = structuredClone(manual);
    expect(functionCVSuggestion(a, "F3", manual, [], t).patch).toBeUndefined();
    expect(manual).toEqual(before);
  });

  it("does not interpret omitted Nano conditions or output bytes as zero", () => {
    const values = [...Array.from({ length: 10 }, (_, i) => indexed(257 + i, i === 1 ? 4 : 0, 3)),
      ...[0, 0, 0, 2, 0, 0, 0].map((value, i) => indexed(257 + i, value, 8))];
    expect(decoderMapping(analyze("LokPilot 5 nano DCC", values), t)).toMatchObject([
      { functionKey: "F3", actions: ["Rangiergang"], warning: false }
    ]);
    for (const [cv, page] of [[266, 3], [258, 8], [262, 8]]) {
      expect(decoderMapping(analyze("LokPilot 5 nano DCC", values.filter(r =>
        r.cvNumber !== cv || r.indexCV32 !== page)), t)).toEqual([]);
    }
  });
});
