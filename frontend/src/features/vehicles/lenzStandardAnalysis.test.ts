import { describe, expect, it } from "vitest";
import { bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";

const label = "STANDARD+ V2";
const rows = (pairs: number[][]) => [[8, 99], ...pairs].map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: label }));
const analyze = (pairs: number[][]) => decoderAnalysis(rows(pairs), label);
const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);

describe("reviewed Lenz STANDARD+ V2 interpretation", () => {
  it.each([[0, false, false], [1, false, false], [2, true, false], [3, true, true],
    [128, false, true], [129, false, true], [131, true, true]])(
    "separates ABC and independent speed-zero braking in CV51=%i", (value, abc, enabled) => {
      const a = analyze([[51, value as number], [52, 0]]);
      expect(a.abc).toBe(abc);
      expect(a.constantBrake).toBe(enabled);
      expect(a.constantBrakeCVs).toEqual([51]);
    });

  it("does not enable constant braking from a distance alone or undocumented bits", () => {
    expect(analyze([[52, 50]]).constantBrake).toBeUndefined();
    expect(analyze([[51, 64], [52, 50]]).constantBrake).toBe(false);
    expect(analyze([[51, 256]]).constantBrake).toBeUndefined();
    expect(analyze([[51, -1]]).abc).toBeUndefined();
  });

  it("evaluates polarity-independent DC stops only with analog operation disabled", () => {
    expect(analyze([[29, 10], [51, 32]]).lenz?.dcStop).toBe(true);
    expect(analyze([[29, 14], [51, 32]]).lenz?.dcStop).toBe(false);
    expect(analyze([[51, 32]]).lenz?.dcStop).toBeUndefined();
    expect(analyze([[29, 10]]).lenz?.dcStop).toBeUndefined();
  });

  it("keeps conflicting shuttle modes and absent motor types unknown", () => {
    expect(analyze([[51, 26]]).lenz?.shuttleMode).toBe("conflict");
    expect(analyze([[51, 10]]).lenz?.shuttleMode).toBe("withoutStop");
    expect(analyze([[51, 18]]).lenz?.shuttleMode).toBe("withStop");
    expect(analyze([[51, 8]]).lenz?.shuttleMode).toBe("requiresABC");
    expect(analyze([[50, 5 + 16 + 32 + 64 + 128]]).lenz)
      .toMatchObject({ motorType: 5, functionDecoder: true, emkDivider: true, loadControl: false, lowFrequency: true });
    expect(analyze([[50, 6]]).lenz?.motorType).toBeUndefined();
    expect(analyze([]).lenz?.motorType).toBeUndefined();
  });

  it("requires RailCom permission and at least one documented channel", () => {
    expect(analyze([[29, 14], [28, 3]]).railcomEnabled).toBe(true);
    expect(analyze([[29, 14], [28, 0]]).railcomEnabled).toBe(false);
    expect(analyze([[29, 6]]).railcomEnabled).toBe(false);
    expect(analyze([[29, 14]]).railcomEnabled).toBeUndefined();
    expect(analyze([[29, 14], [28, 4]]).railcomEnabled).toBeUndefined();
    expect(analyze([[29, 255], [28, 3]]).railcomEnabled).toBeUndefined();
  });

  it("preserves address limits without applying an unproven consist direction bit", () => {
    expect(analyze([[19, 42]]).consistAddress).toBe(42);
    expect(analyze([[19, 170]]).consistAddress).toBeUndefined();
    expect(analyze([[19, 42]]).consistReverse).toBeUndefined();
    expect(analyze([[19, 42], [21, 255], [22, 255]]).enabledFunctions).toEqual([]);
    expect(analyze([[19, 42], [21, 255], [22, 255]]).functionsComplete).toBe(false);
    expect(analyze([[29, 32], [17, 231], [18, 15]]).address).toBe(9999);
    expect(analyze([[29, 32], [17, 231], [18, 16]]).address).toBeUndefined();
    expect(analyze([[29, 32], [17, 192], [18, 0]]).address).toBeUndefined();
    const reading = cvExplanationRows(analyze([[19, 42]]), "address", t).find(r => r.key === "consist");
    expect(reading?.reading).toBe(42);
  });

  it("reads separate F0/F1 directions and F28 without inventing output E or key flags", () => {
    const a = analyze([[33, 3], [34, 2], [35, 4], [47, 8], [39, 16], [144, 9], [58, 4], [59, 8]]);
    const mappings = decoderMapping(a, t);
    expect(mappings.map(m => [m.functionKey, m.direction, m.actions, m.warning])).toEqual([
      ["F0", "forward", ["Ausgang A", "Ausgang B"], false], ["F0", "reverse", ["Ausgang B"], false],
      ["F1", "forward", ["Ausgang C"], false], ["F5", "both", ["Bit4: Ausgang nicht eindeutig belegt"], true],
      ["F1", "reverse", ["Ausgang D"], false], ["F28", "both", ["Ausgang A", "Ausgang D"], false]
    ]);
    expect(mappings[5].sourceCVs).toEqual([144]);
    expect(a.lenz?.shuntingKeys).toEqual(["F3"]);
    expect(a.lenz?.noDelayKeys).toEqual(["F4"]);
    expect(a.railComPlus).toBeUndefined();
    const explained = cvExplanationRows(a, "outputs", t);
    expect(explained.find(r => r.key === "lenz-shuntingKeys")?.reading).toBe("F3");
    expect(explained.find(r => r.key === "mapping-6")?.numbers).toEqual([144]);
  });

  it("decodes decimal effects and located timing formulas without using factory defaults", () => {
    const a = analyze([[60, 43], [62, 21], [63, 32], [147, 30], [148, 80], [149, 30], [30, 5]]);
    expect(a.lenz).toMatchObject({ effectA: 3, effectB: 4, effectC: 1, effectD: 2, faultBits: [0, 2] });
    expect(a.lenz?.flashHz).toBeCloseTo(1.0101);
    expect(analyze([[60, 53]]).lenz?.effectA).toBeUndefined();
    expect(analyze([[62, 13]]).lenz?.effectD).toBeUndefined();
    expect(cvExplanationRows(a, "outputs", t).find(r => r.key === "lenz-time149")?.reading).toBe("0,48 s");
    expect(cvExplanationRows(analyze([]), "outputs", t).find(r => r.key === "lenz-time147")?.reading).toBe("Unbekannt");
    expect(cvExplanationRows(a, "communication", t).find(r => r.key === "lenz-faults")?.reading)
      .toBe("Lampen-Kurzschluss · Motor-/Gleis-Kurzschluss");
  });

  it("does not mix profiles, index pages or unsupported model states", () => {
    const values = [...rows([[29, 14]]), { cvNumber: 51, value: 3, decoderProfile: label, indexCV31: 0, indexCV32: 0 },
      { cvNumber: 51, value: 131, decoderProfile: "Other" }];
    const before = structuredClone(values);
    expect(decoderAnalysis(values, label).constantBrake).toBeUndefined();
    expect(decoderAnalysis(values.filter(r => r.cvNumber !== 8), label).lenz).toBeUndefined();
    expect(decoderAnalysis(values, label, bundledDecoderProfiles.map(p => ({ ...p, active: p.id !== "lenz-standard-plus-v2" }))).lenz).toBeUndefined();
    expect(decoderAnalysis([...values, { cvNumber: 7, value: 91, decoderProfile: label }], label).lenz).toBeUndefined();
    expect(values).toEqual(before);
  });

  it("plots only a selected table and keeps Lenz support parameters distinct from ZIMO special values", () => {
    const source = rows([[29, 14], [2, 1], [6, 1], [5, 0], [67, 200]]);
    const before = structuredClone(source);
    const [smooth] = buildSpeedCurveProfiles(source, "Unbekannt");
    expect(smooth.lenzCurve).toBe(true); expect(smooth.zimoCurve).toBeUndefined();
    expect(smooth.primaryCurve).toBe("none"); expect(smooth.threePoint.map(p => p.value)).toEqual([1, 1, 0]);
    const [table] = buildSpeedCurveProfiles(rows([[29, 30], [67, 12], [94, 255]]), "Unbekannt");
    expect(table.primaryCurve).toBe("speedTable");
    for (const pairs of [[[29, 30], [2, 1], [6, 48], [5, 254]], [[67, 12]], [[29, 255], [67, 12]]]) {
      expect(buildSpeedCurveProfiles(rows(pairs), "Unbekannt")[0].primaryCurve).toBe("none");
    }
    expect(source).toEqual(before);
  });
});
