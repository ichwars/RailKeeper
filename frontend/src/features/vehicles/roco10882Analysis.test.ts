import { describe, expect, it } from "vitest";
import { translate } from "../../shared/i18n";
import { rocoDETranslations } from "../../shared/i18n/roco.de";
import { rocoENTranslations } from "../../shared/i18n/roco.en";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";
import { functionCVSuggestion } from "./functionCVSuggestions";
import type { FunctionEdit } from "./vehicleTransforms";

const label = "Roco 10882";
const rows = (pairs: number[][], profile = label) => [[8, 145], ...pairs]
  .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: profile }));
const analyze = (pairs: number[][]) => decoderAnalysis(rows(pairs), label);
const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);

describe("Roco10882,2012 source scope", () => {
  it("uses technical ZIMO145 and never substitutes brand161 or current ZIMO rules", () => {
    const a = analyze([[7, 30], [65, 2]]);
    expect(a.roco).toBeDefined(); expect(a.zimo).toBeUndefined(); expect(a.firmware).toBeUndefined();
    expect(a.libraryProfile?.manufacturerId).toBe(145);
    const conflict = analyze([[250, 127]]);
    expect(conflict.libraryMatch.status).toBe("conflict"); expect(conflict.roco).toBeUndefined();
    for (const profile of ["10882", "Roco 10882 (ZIMO, Anleitung 2012)"]) {
      expect(decoderAnalysis(rows([], profile), profile).roco).toBeDefined();
    }
    for (const profile of ["Roco 10883", "Roco 10884", "Roco 10895", "MX630", "MN180N18", "Zentrale"]) {
      expect(decoderAnalysis(rows([], profile), profile).roco).toBeUndefined();
    }
    expect(decoderAnalysis(rows([[8, 161]]), label).roco).toBeUndefined();
    expect(decoderAnalysis([{ cvNumber: 29, value: 14, decoderProfile: label }], label).roco).toBeUndefined();
    expect(decoderAnalysis(rows([]).map(r => ({ ...r, protocol: "MM" })), label).roco).toBeUndefined();
  });
  it.each([[0, 0, false], [0, 32, true], [4, 32, false], [8, 32, true], [12, 32, false]])(
    "uses analog bit2, not the erroneous summary bit3: CV29=%i CV124=%i", (config, brake, enabled) => {
      const a = analyze([[29, Number(config)], [124, Number(brake)], [112, 64]]);
      expect(a.roco?.brakeEnabled).toBe(enabled); expect(a.roco?.brakeType).toBe("ac");
      expect(a.abc).toBeUndefined(); expect(a.constantBrake).toBeUndefined();
    });
  it("retains missing braking conditions and distinguishes stored type from activation", () => {
    expect(analyze([[29, 0]]).roco?.brakeEnabled).toBeUndefined();
    expect(analyze([[124, 32]]).roco?.brakeEnabled).toBeUndefined();
    expect(analyze([[29, 4]]).roco?.brakeEnabled).toBe(false);
    expect(analyze([[124, 0]]).roco?.brakeEnabled).toBe(false);
    expect(analyze([[112, 0], [124, 32], [29, 0]]).roco).toMatchObject({ brakeEnabled: true, brakeType: "dc" });
    expect(analyze([[29, 0], [124, 32]]).roco?.brakeType).toBeUndefined();
    const a = decoderAnalysis([...rows([]), { cvNumber: 124, value: 32, decoderProfile: label, indexCV31: 0, indexCV32: 0 }], label);
    expect(a.roco?.brakeEnabled).toBeUndefined();
  });
  it.each([[14, 0, false], [14, 1, true], [14, 2, true], [14, 3, true], [6, 3, false]])(
    "combines RailCom permission with transmit mask: CV29=%i CV28=%i", (config, channels, enabled) => {
      expect(analyze([[29, Number(config)], [28, Number(channels)]]).railcomEnabled).toBe(enabled);
    });
  it("keeps missing/invalid RailCom inputs unknown and exposes independent send options", () => {
    expect(analyze([[29, 14]]).railcomEnabled).toBeUndefined();
    expect(analyze([[29, 14], [28, 4]]).railcomEnabled).toBeUndefined();
    expect(analyze([[28, 3]]).railcomEnabled).toBeUndefined();
    expect(analyze([[29, 14], [28, 1]]).roco).toMatchObject({ broadcast: true, cvData: false });
    expect(analyze([[29, 14], [28, 2]]).roco).toMatchObject({ broadcast: false, cvData: true });
  });
  it("reads the actual shifted columns and marks outputs unavailable on10882", () => {
    const mappings = decoderMapping(analyze([[33, 1], [34, 2], [35, 4], [36, 8], [37, 16], [38, 1], [41, 4], [42, 8]]), t);
    expect(mappings.map(m => [m.functionKey, m.direction, m.warning])).toEqual([
      ["F0", "forward", false], ["F0", "reverse", false], ["F1", "both", false], ["F2", "both", false],
      ["F3", "both", true], ["F4", "both", false], ["F7", "both", true], ["F8", "both", true]
    ]);
    expect(mappings[2].actions).toEqual(["AUX1"]); expect(mappings[3].actions).toEqual(["AUX2"]);
    expect(mappings[4].actions).toEqual(["AUX3 (beim10882 nicht vorhanden)"]);
    expect(mappings[5].actions).toEqual(["AUX2"]);
    expect(mappings[6].actions).toEqual(["AUX7 (beim10882 nicht vorhanden)"]);
    expect(mappings[7].actions).toEqual(["Nicht zugeordnete Bits: 8"]);
    expect(decoderMapping(analyze([[35, 256], [36, -1]]), t)).toEqual([]);
  });
  it("uses only documented positive key assignments, without treating0 as F0", () => {
    expect(analyze([[155, 0], [156, 0]]).roco).toMatchObject({ shuntingKey: undefined, noDelayKey: undefined });
    expect(decoderMapping(analyze([[155, 3], [156, 4]]), t).map(m => [m.functionKey, m.actions])).toEqual([
      ["F3", ["Rangiergang"]], ["F4", ["Beschleunigungs-/Bremsverzögerung aus"]]
    ]);
    expect(analyze([[155, 13], [156, 255]]).roco).toMatchObject({ shuntingKey: undefined, noDelayKey: undefined });
  });
  it("separates dimming exemptions and the DELTA option from key mode/inversion", () => {
    expect(analyze([[60, 0], [114, 133], [51, 8]]).roco).toMatchObject({
      dimmingEnabled: false, dimmingExempt: ["frontLight", "AUX1"], unsupportedDimmingBits: 128
    });
    expect(cvExplanationRows(analyze([[51, 8]]), "outputs", t).find(r => r.key === "roco-delta")?.reading).toBe(8);
    expect(analyze([[60, 100], [114, 0], [51, 0]]).roco).toMatchObject({ dimmingEnabled: true, dimmingExempt: [] });
  });
  it("never borrows undocumented consist, ABC, trims, AUX7 or model-identification registers", () => {
    const a = analyze([[19, 170], [21, 255], [22, 255], [27, 3], [66, 128], [95, 128], [152, 1]]);
    expect(a.consistAddress).toBeUndefined(); expect(a.consistReverse).toBeUndefined();
    expect(a.functionsComplete).toBe(false); expect(a.enabledFunctions).toEqual([]);
    for (const cv of [19, 21, 22, 27, 66, 95, 152, 250]) expect(a.definition({ cvNumber: cv, value: 0 })).toBeUndefined();
    expect(analyze([[29, 32], [17, 231], [18, 16]]).address).toBeUndefined();
    expect(analyze([[29, 32], [17, 192], [18, 128]]).address).toBe(128);
  });
  it("preserves raw imports, provenance and manual function fields", () => {
    const values = rows([[37, 16], [155, 3], [156, 4], [51, 8]])
      .map(r => ({ ...r, sourceLine: 9, sourceExportedAt: "2026-10-20T15:00:00" }));
    const before = structuredClone(values), a = decoderAnalysis(values, label);
    for (const current of [{}, { name: "Mein Rangiergang", mode: "moment", inverted: true, notes: "Manuell" }] satisfies FunctionEdit[]) {
      const original = structuredClone(current);
      expect(functionCVSuggestion(a, "F3", current, [], t)).toEqual({ notesFull: false }); expect(current).toEqual(original);
    }
    for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const) cvExplanationRows(a, topic, t);
    expect(values).toEqual(before); expect(a.railComPlus).toBeUndefined();
    const d = a.definition({ cvNumber: 124, value: 0 })!;
    expect(d.definition.reference).toEqual({ sourceIndex: 0, location: "S.6–7, CV124" });
    expect(d.profile.sources[0].sha256).toBe("a5fe8303c15db927b125ef4dfd15d0ffa7e4f3f404850537971b0cfe669b9883");
  });
  it("has readable localized explanations and puts the source-specific braking first", () => {
    expect(Object.keys(rocoDETranslations).sort()).toEqual(Object.keys(rocoENTranslations).sort());
    for (const lang of ["de", "en"] as const) for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const) {
      for (const r of cvExplanationRows(analyze([]), topic, (key, values) => translate(lang, key, values), lang)) {
        expect([r.name, r.help, r.reading].join(" ")).not.toMatch(/vehicles\.|\{\w+\}/);
      }
    }
    expect(cvExplanationRows(analyze([]), "braking", t).slice(0, 2).map(r => r.key)).toEqual(["roco-brakeEnabled", "roco-brakeType"]);
  });
});

describe("Roco10882 documented speed characteristic", () => {
  const curve = (pairs: number[][]) => buildSpeedCurveProfiles(rows(pairs), "Unbekannt")[0];
  it("uses the explicit three-point mode and its own ranges, with no modern ZIMO special values", () => {
    const p = curve([[29, 14], [2, 4], [5, 252], [6, 85], [65, 2], [66, 128], [95, 128], [67, 1], [94, 255]]);
    expect(p.rocoCurve).toBe(true); expect(p.primaryCurve).toBe("threePoint");
    expect(p.threePoint.map(p => p.value)).toEqual([4, 85, 252]);
    expect(p.forwardTrim).toBeNull(); expect(p.reverseTrim).toBeNull(); expect(p.sourceCount).toBe(6);
    expect(curve([[29, 14], [2, 1], [5, 1], [6, 1]]).threePoint.map(p => p.value)).toEqual([1, 1, 1]);
    expect(curve([[29, 14], [2, 0], [5, 0], [6, 253]]).threePoint).toEqual([]);
  });
  it("never falls back to an inactive or unknown curve", () => {
    expect(curve([[29, 30], [67, 0], [94, 255]]).primaryCurve).toBe("speedTable");
    expect(curve([[29, 30], [2, 4], [5, 252], [6, 85]]).primaryCurve).toBe("none");
    expect(curve([[29, 14], [67, 0], [94, 255]]).primaryCurve).toBe("none");
    expect(curve([[2, 4], [5, 252], [6, 85], [67, 1], [94, 255]]).primaryCurve).toBe("none");
    expect(curve([[29, 256], [67, 1]]).primaryCurve).toBe("none");
    expect(curve([[29, 30], [67, -1], [94, 256]]).primaryCurve).toBe("none");
  });
});
