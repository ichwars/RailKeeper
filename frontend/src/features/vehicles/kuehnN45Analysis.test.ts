import { describe, expect, it } from "vitest";
import { bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import { kuehnDETranslations } from "../../shared/i18n/kuehn.de";
import { kuehnENTranslations } from "../../shared/i18n/kuehn.en";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";
import { functionCVSuggestion } from "./functionCVSuggestions";
import type { FunctionEdit } from "./vehicleTransforms";

const label = "Kühn N45";
const rows = (pairs: number[][], profile = label) => [[8, 157], ...pairs]
  .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: profile }));
const analyze = (pairs: number[][]) => decoderAnalysis(rows(pairs), label);
const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);

describe("reviewed Kühn N45 summaries", () => {
  it.each([[0, false], [1, true], [2, true], [3, true], [16, false], [48, false]])(
    "uses only ABC polarity bits for CV27=%i", (value, abc) => {
      const a = analyze([[27, value as number], [123, 9], [124, 255]]);
      expect(a.abc).toBe(abc);
      expect(a.constantBrake).toBeUndefined();
    });

  it("combines analog permission with the documented DC braking override", () => {
    expect(analyze([[29, 14], [27, 48]]).dcAnalog).toBe(false);
    expect(analyze([[29, 14], [27, 16]]).dcAnalog).toBe(true);
    expect(analyze([[29, 10]]).dcAnalog).toBe(false);
    expect(analyze([[29, 14]]).dcAnalog).toBeUndefined();
    expect(analyze([[27, 48]]).dcAnalog).toBeUndefined();
    expect(analyze([[29, 255], [27, 48]]).dcAnalog).toBeUndefined();
    expect(analyze([[29, 14], [28, 0]]).railcomEnabled).toBe(true); // No invented CV28 channel bits.
  });

  it("requires model evidence and retains index/protocol/manufacturer boundaries", () => {
    for (const profile of ["N025", "T65", "T125", "T145", "F060", "N45-18", "Zentrale"]) {
      expect(decoderAnalysis(rows([[56, 2]], profile), profile).kuehn).toBeUndefined();
    }
    expect(decoderAnalysis([{ cvNumber: 56, value: 2, decoderProfile: label }], label).kuehn).toBeUndefined();
    expect(decoderAnalysis(rows([[56, 2]]).map(r => ({ ...r, protocol: "MM" })), label).kuehn).toBeUndefined();
    expect(decoderAnalysis(rows([[56, 2]]).map(r => ({ ...r, value: r.cvNumber === 8 ? 151 : r.value })), label).kuehn).toBeUndefined();
    const indexed = decoderAnalysis([...rows([]), ...rows([[27, 3], [56, 2], [58, 72]])
      .map(r => ({ ...r, indexCV31: 0, indexCV32: 0 }))], label);
    expect(indexed.abc).toBeUndefined();
    expect(indexed.kuehn?.loadControl).toBeUndefined();
    expect(decoderMapping(indexed, t)).toEqual([]);
  });

  it("reads shifted masks for F3, F4, F5, F8, F9 and F12", () => {
    const mappings = decoderMapping(analyze([[37, 4], [38, 128], [39, 2], [42, 32], [43, 4], [46, 128], [56, 2]]), t);
    expect(mappings.map(m => [m.functionKey, m.actions, m.warning])).toEqual([
      ["F3", ["Ausgang C"], false], ["F4", ["Ausgang C"], false],
      ["F5", ["Ausgang E (nur Logikpegel)"], false], ["F8", ["Ausgang A"], false],
      ["F9", ["Ausgang A"], false], ["F12", ["Ausgang F (nur Logikpegel)"], false]
    ]);
    expect(decoderMapping(analyze([[35, 128]]), t)[0]).toMatchObject({ warning: true, actions: ["Nicht zugeordnete Bits: 128"] });
  });

  it("keeps SUSI-disabled and unknown E/F availability explicit", () => {
    expect(decoderMapping(analyze([[39, 2], [56, 6]]), t)[0]).toMatchObject({ warning: true, actions: ["Ausgang E (durch SUSI gesperrt)"] });
    expect(decoderMapping(analyze([[39, 2]]), t)[0]).toMatchObject({ warning: true, actions: ["Ausgang E (Verfügbarkeit ohne CV56 ungeklärt)"] });
    expect(decoderMapping(analyze([[33, 1], [34, 2]]), t).map(m => m.direction)).toEqual(["forward", "reverse"]);
    expect(decoderMapping(analyze([[35, 4], [51, 96]]), t)[0].actions).toEqual(["Ausgang C (aus bei Vorwärtsfahrt · aus bei Rückwärtsfahrt)"]);
  });

  it("separates logical function assignments from switching mode or inversion", () => {
    const a = analyze([[58, 72], [59, 1]]);
    expect(a.kuehn?.shuntingKeys).toEqual(["F3"]);
    expect(a.kuehn?.noDelayKeys).toEqual(["F4"]);
    expect(a.kuehn?.dimmingKeys).toEqual(["F1"]);
    expect(decoderMapping(a, t).map(m => [m.functionKey, m.actions])).toEqual([
      ["F3", ["Rangiergang"]], ["F4", ["Beschleunigungs-/Bremsverzögerung aus"]], ["F1", ["Abblendfunktion"]]
    ]);
    expect(a.railComPlus).toBeUndefined();
    expect(a.esu5).toBe(false);
  });

  it("does not invent consist direction, permissions, firmware or foreign registers", () => {
    const a = analyze([[7, 36], [19, 170], [21, 255], [22, 255], [66, 128]]);
    expect(a.consistAddress).toBeUndefined();
    expect(a.consistReverse).toBeUndefined();
    expect(a.functionsComplete).toBe(false);
    expect(a.enabledFunctions).toEqual([]);
    expect(a.firmware).toBeUndefined();
    expect(a.get(66)).toBeUndefined();
    expect(analyze([[19, 42]]).consistAddress).toBe(42);
    expect(analyze([[29, 32], [17, 231], [18, 16]]).address).toBeUndefined();
  });

  it("decodes decimal values, special trim zero and reversed memory permission", () => {
    const a = analyze([[55, 79], [60, 43], [61, 28], [62, 91], [63, 48], [95, 0], [112, 32], [113, 9], [9, 2], [56, 3]]);
    expect(a.kuehn).toMatchObject({ effectCycle: 7, effectDimming: 9, pullInSeconds: 4, holdingLevel: 8,
      reverseFactor: 1, memory: false, randomSeconds: 16, frequency: 32000, loadControl: true, emkDivider: true,
      dimming: [{ units: 3, tens: 4 }, { units: 8, tens: 2 }, { units: 1, tens: 9 }] });
    expect(analyze([[63, 0]]).kuehn?.pullInSeconds).toBe(0.5);
    expect(analyze([[63, 49]]).kuehn?.pullInSeconds).toBeUndefined();
    expect(analyze([[55, 100], [60, -1], [112, 256], [9, 3]]).kuehn)
      .toMatchObject({ effectCycle: undefined, frequency: undefined, memory: undefined });
    expect(analyze([[95, 64]]).kuehn?.reverseFactor).toBe(0.5);
    expect(analyze([[112, 0]]).kuehn?.memory).toBe(true);
  });

  it("uses CV14 bits2–5 for F9–F12 and no low-bit substitution", () => {
    expect(analyze([[13, 5], [14, 36]]).kuehn?.analogKeys).toEqual(["F0", "F1", "F3", "F9", "F12"]);
    expect(analyze([[13, 0], [14, 3]]).kuehn?.analogKeys).toEqual(["F0"]);
    expect(analyze([[13, 5]]).kuehn?.analogKeys).toBeUndefined();
  });

  it("keeps invalid effects unknown without losing the independently documented direction condition", () => {
    const output = cvExplanationRows(analyze([[49, 35], [50, 128]]), "outputs", t);
    expect(output.find(r => r.key === "kuehn-49")?.reading).toBe("Unbekannt · aus bei Rückwärtsfahrt");
    expect(output.find(r => r.key === "kuehn-50")?.reading).toBe("Unbekannt");
  });

  it("keeps raw readings, provenance and manual function settings unchanged", () => {
    const values = rows([[33, 1], [56, 2], [58, 72]]).map(r => ({ ...r, sourceLine: 8, sourceExportedAt: "2026-10-17T17:41:59" }));
    const before = structuredClone(values), a = decoderAnalysis(values, label);
    const current: FunctionEdit = { name: "Manuell", mode: "moment", inverted: true, notes: "Eigener Nachweis" };
    const original = structuredClone(current);
    expect(functionCVSuggestion(a, "F3", current, [], t)).toEqual({ notesFull: false });
    for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const) cvExplanationRows(a, topic, t);
    expect(values).toEqual(before);
    expect(current).toEqual(original);
  });

  it("has translated summaries in both languages, including missing readings", () => {
    expect(Object.keys(kuehnDETranslations).sort()).toEqual(Object.keys(kuehnENTranslations).sort());
    for (const language of ["de", "en"] as const) {
      const tr = (key: string, values?: Record<string, string | number>) => translate(language, key, values);
      for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const) {
        const explanations = cvExplanationRows(analyze([]), topic, tr, language);
        for (const row of explanations) expect([row.name, row.help, row.reading].join(" ")).not.toMatch(/vehicles\.|\{\w+\}/);
      }
    }
  });

  it("handles a reviewed custom profile with a removed optional definition without crashing", () => {
    const catalog = bundledDecoderProfiles.map(p => p.id === "kuehn-n45"
      ? { ...p, definitions: p.definitions.filter(d => d.cv !== 63) } : p);
    expect(() => cvExplanationRows(decoderAnalysis(rows([]), label, catalog), "outputs", t)).not.toThrow();
  });
});

describe("N45 speed characteristic", () => {
  const curve = (pairs: number[][]) => buildSpeedCurveProfiles(rows(pairs), "Unbekannt")[0];
  it("does not draw special CV5/6 values as ordinary support points", () => {
    const p = curve([[29, 14], [2, 3], [5, 0], [6, 0], [66, 128]]);
    expect(p.kuehnCurve).toBe(true);
    expect(p.primaryCurve).toBe("none");
    expect(p.threePoint.map(p => p.value)).toEqual([3, 0, 0]);
    expect(p.forwardTrim).toBeNull();
  });
  it("draws only a selected stored table, preserving the 28/128 ambiguity", () => {
    expect(curve([[29, 30], [67, 1], [94, 255]]).primaryCurve).toBe("speedTable");
    expect(curve([[67, 1], [94, 255], [2, 3]]).primaryCurve).toBe("none");
    expect(curve([[29, 255], [67, 1]]).primaryCurve).toBe("none");
    expect(curve([[29, 30], [67, 256], [94, -1]]).primaryCurve).toBe("none");
    expect(curve([[29, 30], [2, 3], [5, 255]]).primaryCurve).toBe("none");
  });
});
