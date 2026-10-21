import { describe, expect, it } from "vitest";
import { translate } from "../../shared/i18n";
import { uhlenbrockDETranslations } from "../../shared/i18n/uhlenbrock.de";
import { uhlenbrockENTranslations } from "../../shared/i18n/uhlenbrock.en";
import { decoderDefinition, resolveDecoderLibrary } from "../../shared/decoderLibrary";
import { bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";
import { functionCVSuggestion } from "./functionCVSuggestions";
import type { FunctionEdit } from "./vehicleTransforms";

const label = "Uhlenbrock 76420";
const rows = (pairs: number[][], profile = label) => [[8, 85], ...pairs]
  .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: profile }));
const analyze = (pairs: number[][]) => decoderAnalysis(rows(pairs), label);
const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);

describe("76420 model-specific interpretation", () => {
  it.each([[0, 1, false, false], [4, 1, false, true], [4, 2, true, false], [4, 3, true, true]])(
    "combines CV29=%i with analog mode CV51=%i", (config, analog, dc, ac) => {
      expect(analyze([[29, Number(config)], [51, Number(analog)]]).uhlenbrock).toMatchObject({ dcAnalog: dc, acAnalog: ac });
    });
  it("retains uncertainty and never interprets CV29 bit3 as RailCom", () => {
    expect(analyze([[29, 4]]).dcAnalog).toBeUndefined();
    expect(analyze([[51, 3]]).dcAnalog).toBeUndefined();
    expect(analyze([[29, 0]]).dcAnalog).toBe(false);
    expect(analyze([[29, 12], [51, 0]]).dcAnalog).toBeUndefined();
    const a = analyze([[29, 14], [27, 3], [28, 3], [254, 10], [255, 20]]);
    expect(a.railcomEnabled).toBeUndefined(); expect(a.abc).toBeUndefined(); expect(a.constantBrake).toBeUndefined();
    for (const cv of [21, 22, 27, 28, 95, 254, 255]) {
      expect(a.get(cv)).toBeUndefined();
      expect(a.definition({ cvNumber: cv, value: 0 })).toBeUndefined();
    }
    const railcom = cvExplanationRows(a, "address", t).find(r => r.key === "railcom")!;
    expect(railcom.reading).toBe("Unbekannt"); expect(railcom.help).toContain("unbenutzt");
  });
  it("requires both Märklin braking conditions and distinguishes an inactive target", () => {
    expect(analyze([[29, 4], [49, 128], [52, 30]]).uhlenbrock).toMatchObject({ dcBrake: true, brakeTarget: 0 });
    expect(analyze([[29, 4], [49, 132], [52, 30]]).uhlenbrock).toMatchObject({ dcBrake: true, brakeTarget: 30 });
    expect(analyze([[29, 0], [49, 128]]).uhlenbrock?.dcBrake).toBe(false);
    expect(analyze([[49, 0]]).uhlenbrock?.dcBrake).toBe(false);
    expect(analyze([[29, 4]]).uhlenbrock?.dcBrake).toBeUndefined();
    expect(analyze([[49, 132]]).uhlenbrock?.brakeTarget).toBeUndefined();
    expect(analyze([[52, 30]]).uhlenbrock?.brakeTarget).toBeUndefined();
  });
  it.each([[0, "both"], [8, "dcc"], [16, "motorola"], [24, "blocked"]])(
    "handles protocol disable mask %i including the both-disabled conflict", (value, protocol) => {
      expect(analyze([[49, Number(value)]]).uhlenbrock?.protocols).toBe(protocol);
    });
  it("separates inverted load-enable polarity, SUSI selection and light swapping", () => {
    expect(analyze([[49, 32]]).uhlenbrock).toMatchObject({ loadControl: true, susi: "sound", dimmingAux: true, lightsSwapped: false });
    expect(analyze([[49, 67]]).uhlenbrock).toMatchObject({ loadControl: false, susi: "lissy", dimmingAux: false, lightsSwapped: true });
    expect(analyze([]).uhlenbrock).toMatchObject({ loadControl: undefined, susi: undefined, lightsSwapped: undefined });
  });
  it.each([[0, []], [1, ["A1"]], [2, ["A2"]], [3, ["A1", "A2"]]])(
    "decodes CV98=%i using the German original", (value, outputs) => {
      expect(analyze([[98, Number(value)], [99, 10]]).uhlenbrock).toMatchObject({ timedOutputs: outputs, durationSeconds: 1 });
    });
  it("leaves missing/invalid timer and protection values unknown", () => {
    expect(analyze([[98, 4], [99, 256], [60, 1]]).uhlenbrock)
      .toMatchObject({ timedOutputs: undefined, durationSeconds: undefined, shortCircuit: undefined });
    expect(analyze([[98, 0], [99, 0], [60, 0]]).uhlenbrock)
      .toMatchObject({ timedOutputs: [], durationSeconds: 0, shortCircuit: false });
    expect(analyze([[99, 255], [60, 9]]).uhlenbrock)
      .toMatchObject({ timedOutputs: undefined, durationSeconds: 25.5, shortCircuit: true });
  });
  it("uses uniform masks F0–F8, without borrowing the shifted Kühn mapping", () => {
    const a = analyze([[33, 1], [34, 2], [35, 4], [36, 8], [37, 16], [38, 32], [39, 5], [42, 63]]);
    const mapping = decoderMapping(a, t);
    expect(mapping.map(m => m.functionKey)).toEqual(["F0", "F0", "F1", "F2", "F3", "F4", "F5", "F8"]);
    expect(mapping.map(m => m.direction).slice(0, 3)).toEqual(["forward", "reverse", "both"]);
    expect(mapping[4].actions).toEqual(["Rangiergang"]);
    expect(mapping[5].actions).toEqual(["Anfahr-/Bremsverzögerung schalten"]);
    expect(mapping[6].actions).toEqual(["Lichtausgang vorn", "Ausgang A1"]);
    expect(mapping[7].actions).toHaveLength(6);
    expect(decoderMapping(analyze([[35, 64], [36, 0], [42, -1]]), t)).toEqual([]);
  });
  it("accepts documented consist reversal but does not invent consist key permissions", () => {
    const a = analyze([[19, 170], [21, 255], [22, 255], [7, 12]]);
    expect(a.consistAddress).toBe(42); expect(a.consistReverse).toBe(true);
    expect(a.functionsComplete).toBe(false); expect(a.enabledFunctions).toEqual([]); expect(a.firmware).toBeUndefined();
    expect(analyze([[29, 32], [17, 199], [18, 208]]).address).toBe(2000);
    expect(analyze([[29, 32], [17, 192], [18, 127]]).address).toBeUndefined();
    expect(analyze([[29, 32], [17, 231], [18, 16]]).address).toBeUndefined();
  });
  it("puts documented braking first and includes separate AC permission", () => {
    const a = analyze([[29, 4], [49, 128], [51, 1]]);
    expect(cvExplanationRows(a, "braking", t).slice(0, 2).map(r => r.key))
      .toEqual(["uhlenbrock-dcBrake", "uhlenbrock-brakeTarget"]);
    expect(cvExplanationRows(a, "address", t).find(r => r.key === "uhlenbrock-acAnalog")?.reading).toBe("ja");
  });
  it("requires the exact model, manufacturer and base DCC readings", () => {
    for (const other of ["76425", "76320", "76400", "IntelliDrive", "Zentrale"]) {
      expect(decoderAnalysis(rows([[49, 32]], other), other).uhlenbrock).toBeUndefined();
    }
    expect(decoderAnalysis([{ cvNumber: 49, value: 32, decoderProfile: label }], label).uhlenbrock).toBeUndefined();
    expect(decoderAnalysis(rows([[8, 151], [49, 32]]), label).uhlenbrock).toBeUndefined();
    expect(decoderAnalysis(rows([[49, 32]]).map(r => ({ ...r, protocol: "MM" })), label).uhlenbrock).toBeUndefined();
    const a = decoderAnalysis([...rows([]), { cvNumber: 49, value: 32, decoderProfile: label, indexCV31: 0, indexCV32: 0 }], label);
    expect(a.uhlenbrock?.loadControl).toBeUndefined();
  });
  it("retains readings, import provenance and manually configured vehicle functions", () => {
    const values = rows([[33, 1], [37, 16], [49, 64], [98, 3], [99, 10]])
      .map(r => ({ ...r, sourceLine: 8, sourceExportedAt: "2026-10-17T17:41:59" }));
    const before = structuredClone(values), a = decoderAnalysis(values, label);
    for (const current of [{}, { name: "Manuell", mode: "moment", inverted: true, notes: "Eigener Nachweis" }] satisfies FunctionEdit[]) {
      const original = structuredClone(current);
      expect(functionCVSuggestion(a, "F3", current, [], t)).toEqual({ notesFull: false });
      expect(current).toEqual(original);
    }
    for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const) cvExplanationRows(a, topic, t);
    expect(values).toEqual(before); expect(a.railComPlus).toBeUndefined(); expect(a.esu5).toBe(false);
    const match = resolveDecoderLibrary(values, label);
    const result = decoderDefinition(bundledDecoderProfiles, match, { cvNumber: 98, value: 3 })!;
    expect(result.definition.reference).toEqual({ sourceIndex: 1, location: "S.6, CV98" });
    expect(result.profile.sources[1].sha256).toBe("afd8ec41bf923ed7bf9a34b9900cdeb289ca44cf456ab48b730bd6cd985c46c3");
  });
  it("has German and English explanations, also for missing readings", () => {
    expect(Object.keys(uhlenbrockDETranslations).sort()).toEqual(Object.keys(uhlenbrockENTranslations).sort());
    for (const lang of ["de", "en"] as const) for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const) {
      const explanations = cvExplanationRows(analyze([]), topic, (key, values) => translate(lang, key, values), lang);
      for (const r of explanations) expect([r.name, r.help, r.reading].join(" ")).not.toMatch(/vehicles\.|\{\w+\}/);
    }
  });
});

describe("76420 speed parameters", () => {
  const curve = (pairs: number[][]) => buildSpeedCurveProfiles(rows(pairs), "Unbekannt")[0];
  it("uses CV47/48 as raw corrections, never CV66/95 or a guessed parameter curve", () => {
    const p = curve([[29, 6], [2, 1], [6, 24], [5, 48], [47, 32], [48, 31], [66, 200], [95, 250], [67, 1], [94, 255]]);
    expect(p.uhlenbrockCurve).toBe(true); expect(p.primaryCurve).toBe("none");
    expect(p.forwardTrim).toBe(32); expect(p.reverseTrim).toBe(31);
    expect(p.threePoint.map(p => p.value)).toEqual([1, 24, 48]);
    expect(p.sourceCount).toBe(8); // CV66/95 are not 76420 speed parameters.
    expect(curve([[29, 6], [2, 0], [5, 64], [6, -1], [47, 64]]).threePoint).toEqual([]);
    expect(curve([[29, 6], [66, 128], [95, 128]]).forwardTrim).toBeNull();
  });
  it("plots only a selected table, with valid byte values", () => {
    expect(curve([[29, 22], [67, 0], [94, 255]]).primaryCurve).toBe("speedTable");
    expect(curve([[29, 255], [67, 1]]).primaryCurve).toBe("speedTable"); // Manual allows CV29 0–255.
    expect(curve([[67, 0], [94, 255]]).primaryCurve).toBe("none");
    expect(curve([[29, 22], [2, 1], [5, 48], [6, 24]]).primaryCurve).toBe("none");
    expect(curve([[29, 22], [67, -1], [94, 256]]).primaryCurve).toBe("none");
    expect(curve([[29, 256], [67, 1]]).primaryCurve).toBe("none");
  });
});
