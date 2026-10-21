import { describe, expect, it } from "vitest";
import { decoderAnalysis, esuMapping, type DecoderCV } from "./decoderAnalysis";
import { translate } from "../../shared/i18n";

const profile = "LokPilot 5 micro DCC";
const base = (cvNumber: number, value: number): DecoderCV => ({ cvNumber, value, decoderProfile: profile });
const analyze = (values: DecoderCV[]) => decoderAnalysis(values, profile);
const standard = [base(8, 151), base(29, 14), base(1, 3)];
const t = (key: string) => translate("de", key);

function mappingRow(row: number, a: number, b: number, k: number, n = 0, o = 0): DecoderCV[] {
  const start = 257 + (row % 16) * 16;
  return [
    ...[a, b, 0, 0, 0, 0, 0, 0, 0].map((value, i) => ({ ...base(start + i, value), indexCV31: 16, indexCV32: 3 + Math.floor(row / 16) })),
    ...[k, undefined, 0, n, o, undefined, 0].flatMap((value, i) => value == null ? [] : [{ ...base(start + i, value), indexCV31: 16, indexCV32: 8 + Math.floor(row / 16) }])
  ];
}

describe("profile-aware decoder analysis", () => {
  it("never merges profiles, base values and page zero", () => {
    const values = [...standard, { ...base(1, 55), decoderProfile: "Other" },
      { ...base(1, 99), indexCV31: 0, indexCV32: 0 }];
    expect(analyze(values).address).toBe(3);
    expect(decoderAnalysis(values, "Other").address).toBeUndefined();
    expect(analyze([]).abc).toBeUndefined();
  });
  it("selects the active short/long address without guessing missing bytes", () => {
    expect(analyze([...standard, base(17, 192), base(18, 128)]).longAddress).toBe(128);
    expect(analyze([base(29, 32), base(17, 195), base(18, 232)]).address).toBe(1000);
    expect(analyze([base(29, 32), base(18, 232)]).address).toBeUndefined();
  });
  it("separates inactive/reversed consists from their function permissions", () => {
    const a = analyze([...standard, base(19, 170), base(21, 5), base(22, 1)]);
    expect(a.consistAddress).toBe(42); expect(a.consistReverse).toBe(true);
    expect(a.enabledFunctions).toEqual(["F0 →", "F1", "F3"]);
    expect(analyze([base(19, 128)]).consistAddress).toBe(0);
    expect(analyze([base(21, 0)]).functionsComplete).toBe(false);
    expect(analyze([base(21, 0), base(22, 0)]).enabledFunctions).toEqual([]);
  });
  it("does not call DC braking ABC or infer missing braking values as zero", () => {
    expect(analyze([...standard, base(27, 24), base(254, 0), base(255, 0)]))
      .toMatchObject({ abc: false, constantBrake: false });
    expect(analyze([...standard, base(27, 1), base(254, 20)]))
      .toMatchObject({ abc: true, constantBrake: true });
    expect(analyze([...standard, base(254, 0)]).constantBrake).toBe(false);
    expect(analyze([...standard, base(254, 0), base(255, 50)]).constantBrake).toBe(false);
    expect(analyze([...standard, base(255, 50)]).constantBrake).toBeUndefined();
    expect(decoderAnalysis([base(27, 1)], profile).abc).toBeUndefined();
    expect(decoderAnalysis([...standard], "Other").esu5).toBe(false);
  });
  it("combines ESU DC analog permission with the CV29 master permission", () => {
    expect(analyze([...standard, base(50, 2)]).dcAnalog).toBe(true);
    expect(analyze([...standard, base(50, 0)]).dcAnalog).toBe(false);
    expect(analyze([base(8, 151), base(29, 10), base(50, 2)]).dcAnalog).toBe(false);
    expect(analyze(standard).dcAnalog).toBeUndefined();
    expect(decoderAnalysis([base(29, 14), base(50, 2)], profile).dcAnalog).toBeUndefined();
  });
  it("takes a firmware version only from the preserved file header", () => {
    expect(analyze([base(7, 255)]).firmware).toBeUndefined();
    expect(analyze([{ ...base(7, 255), description: "LokProgrammer Decoder: LokPilot 5 micro DCC (5.14.183)" }]).firmware).toBe("5.14.183");
  });
});

describe("ESU indexed mapping", () => {
  it("reads direction and actions from separate index pages", () => {
    const rows = esuMapping(analyze([...standard, ...mappingRow(0, 20, 0, 1), ...mappingRow(1, 24, 0, 2),
      ...mappingRow(4, 0, 4, 0, 2), ...mappingRow(5, 0, 16, 0, 0, 8)]), t);
    expect(rows.map((row) => [row.functionKey, row.direction, row.actions])).toEqual([
      ["F0", "forward", ["Licht vorne"]], ["F0", "reverse", ["Licht hinten"]],
      ["F3", "both", ["Rangiergang"]], ["F4", "both", ["Beschleunigungs-/Bremsverzögerung aus"]]
    ]);
  });
  it("keeps PowerPack control distinct and handles the next mapping page", () => {
    const a = analyze([...standard, ...mappingRow(16, 0, 64, 16),
      { ...base(291, 31), indexCV31: 16, indexCV32: 0 }]);
    expect(esuMapping(a, t)[0]).toMatchObject({ row: 17, functionKey: "F5", warning: true,
      actions: ["AUX3 · PowerPack Control"] });
  });
  it("does not guess conditions or actions from missing, negative or compound fields", () => {
    for (const values of [mappingRow(0, 48, 0, 1), mappingRow(0, 80, 0, 1),
      mappingRow(0, 20, 0, 1).filter((row) => !(row.indexCV32 === 3 && row.cvNumber === 258))]) {
      expect(esuMapping(analyze([...standard, ...values]), t)).toEqual([]);
    }
    expect(esuMapping(analyze(mappingRow(0, 20, 0, 1)), t)).toEqual([]);
  });
});
