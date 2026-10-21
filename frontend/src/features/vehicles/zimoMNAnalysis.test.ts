import { describe, expect, it } from "vitest";
import { bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";
import { decoderDETranslations } from "../../shared/i18n/decoder.de";
import { deTranslations } from "../../shared/i18n/de";

const label = "MN180N18";
const rows = (pairs: number[][]) => [[8, 145], ...pairs].map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: label }));
const analyze = (pairs: number[][]) => decoderAnalysis(rows(pairs), label);
const t = (key: string) => decoderDETranslations[key] || deTranslations[key] || key;

describe("reviewed ZIMO composed interpretation", () => {
  it.each([
    [34, 12, 1234, false], [0, 100, 10000, false], [34, 130, 234, false],
    [162, 12, 1234, true], [34, 128, 34, false], [0, 128, 0, false], [39, 102, 10239, false]
  ])("combines CV19=%i and CV20=%i without mixing address, RailCom or direction", (cv19, cv20, address, reverse) => {
    const a = analyze([[19, cv19 as number], [20, cv20 as number]]);
    expect(a.consistAddress).toBe(address);
    expect(a.consistReverse).toBe(reverse);
    expect(a.consistCVs).toEqual([19, 20]);
  });

  it.each([[[19, 34]], [[20, 12]], [[19, 128], [20, 0]], [[19, 34], [20, 131]], [[19, 40], [20, 102]]])(
    "does not invent an address from missing, undocumented or overflowing bytes: %j", (...pairs) => {
      expect(analyze(pairs).consistAddress).toBeUndefined();
    });

  it("separates grouped permissions and auto-consist from the unresolved F9 bit", () => {
    const a = analyze([[21, 1], [22, 128 + 64 + 8 + 4 + 1]]);
    expect(a.enabledFunctions).toEqual(["F0 →", "F1", "F10", ...Array.from({ length: 15 }, (_, i) => `F${13 + i}`)]);
    expect(a.enabledFunctions).not.toContain("F9");
    expect(a.functionsComplete).toBe(false);
    expect(a.zimo?.autoConsist).toBe(true);
    const explained = cvExplanationRows(a, "address", t);
    expect(explained.find(row => row.key === "consistFunctions")?.reading).toContain("F10");
    expect(explained.find(row => row.key === "consistFunctions")?.note).toContain("widersprüchlich");
    expect(analyze([[21, 0], [22, 64]]).functionsComplete).toBe(true);
    expect(analyze([[21, 0], [22, 64]]).enabledFunctions).toEqual([]);
  });

  it.each([0, 1, 2, 3, 11, 12, 13])("interprets CV140=%i as a mode, independently of a stored distance", mode => {
    const a = analyze([[140, mode], [141, 20]]);
    expect(a.constantBrake).toBe(mode !== 0);
    expect(a.constantBrakeCVs).toEqual([140]);
    expect(a.zimo?.brakeDistance.forward).toBe(20);
    expect(a.zimo?.brakeDistance.reverse).toBe(20);
    expect(a.zimo?.brakeDistance.fine).toBe(false);
  });

  it("keeps each direction unknown until both fine-resolution bytes exist", () => {
    const a = analyze([[140, 11], [141, 0], [830, 1], [831, 200], [833, 255]]);
    expect(a.zimo?.brakeDistance.forward).toBe(456);
    expect(a.zimo?.brakeDistance.reverse).toBeUndefined();
    const explained = cvExplanationRows(a, "braking", t);
    expect(explained.find(row => row.key === "brakeForward")?.numbers).toEqual([141, 830, 831]);
    expect(explained.find(row => row.key === "brakeForward")?.note).toContain("CV830=1 · CV831=200");
    expect(explained.find(row => row.key === "brakeReverse")?.note).toContain("CV832=? · CV833=255");
    expect(explained.find(row => row.key === "brakeReverse")?.reading).toBe("Unbekannt");
    const missingSelector = analyze([[140, 1], [830, 1], [831, 200]]);
    expect(missingSelector.constantBrake).toBe(true);
    expect(missingSelector.zimo?.brakeDistance.forward).toBeUndefined();
    expect(analyze([[140, 4], [141, 20]]).constantBrake).toBeUndefined();
    expect(analyze([[141, 20]]).constantBrake).toBeUndefined();
  });

  it("requires CV29 and CV12 analog permissions without treating zero as stable format disablement", () => {
    expect(analyze([[29, 14], [12, 117]]).dcAnalog).toBe(true);
    expect(analyze([[29, 14], [12, 116]]).dcAnalog).toBe(false);
    expect(analyze([[29, 10]]).dcAnalog).toBe(false);
    expect(analyze([[29, 14]]).dcAnalog).toBeUndefined();
    expect(analyze([[29, 14], [12, 0]]).dcAnalog).toBeUndefined();
    expect(cvExplanationRows(analyze([[29, 14], [12, 116]]), "address", t)
      .find(row => row.key === "analog")?.numbers).toEqual([12, 29]);
  });

  it("does not borrow values from index pages, other profiles, inactive models or older firmware", () => {
    const values = [...rows([[19, 34], [140, 1], [141, 0]]),
      { cvNumber: 20, value: 12, decoderProfile: label, indexCV31: 0, indexCV32: 0 },
      { cvNumber: 830, value: 1, decoderProfile: label, indexCV31: 1, indexCV32: 1 },
      { cvNumber: 831, value: 200, decoderProfile: "Other" }];
    const before = structuredClone(values);
    const a = decoderAnalysis(values, label);
    expect(a.consistAddress).toBeUndefined();
    expect(a.zimo?.brakeDistance.forward).toBeUndefined();
    expect(values).toEqual(before);
    const older = [...values, { cvNumber: 7, value: 4, decoderProfile: label,
      description: "LokProgrammer Decoder: MN180N18 (4.229)" }];
    expect(decoderAnalysis(older, label).zimo).toBeUndefined();
    expect(decoderAnalysis(values, label, bundledDecoderProfiles.map(p => ({ ...p, active: p.id !== "zimo-mn180n18" }))).zimo).toBeUndefined();
  });

  it("keeps smoothed support parameters raw and plots only a selected 28-point table", () => {
    const values = rows([[29, 14], [2, 1], [5, 1], [6, 1], [67, 22]]);
    const before = structuredClone(values);
    const [smoothed] = buildSpeedCurveProfiles(values, "Ohne Profil");
    expect(smoothed.zimoCurve).toEqual({ maximum: 255, middle: null, defaultMiddle: true });
    expect(smoothed.threePoint.map(p => p.value)).toEqual([1, 1, 1]);
    expect(smoothed.primaryCurve).toBe("none");
    expect(values).toEqual(before);
    const [table] = buildSpeedCurveProfiles(rows([[29, 30], [5, 0], [6, 1], [67, 22], [94, 255]]), "Ohne Profil");
    expect(table.primaryCurve).toBe("speedTable");
    expect(table.speedTable.map(p => p.value)).toEqual([22, 255]);
    const [missingTable] = buildSpeedCurveProfiles(rows([[29, 30], [2, 1], [5, 1], [6, 1]]), "Ohne Profil");
    expect(missingTable.primaryCurve).toBe("none");
    expect(missingTable.speedTableActive).toBe(true);
    const [missingMode] = buildSpeedCurveProfiles(values.filter(v => v.cvNumber !== 29), "Ohne Profil");
    expect(missingMode.primaryCurve).toBe("none");
    const [invalidMode] = buildSpeedCurveProfiles(rows([[29, 255], [67, 22]]), "Ohne Profil");
    expect(invalidMode.primaryCurve).toBe("none");
    expect(invalidMode.speedTableActive).toBeNull();
    const [invalidMiddle] = buildSpeedCurveProfiles(rows([[29, 14], [5, 0], [6, 0]]), "Ohne Profil");
    expect(invalidMiddle.zimoCurve).toEqual({ maximum: 255, middle: null, defaultMiddle: false });
  });
});
