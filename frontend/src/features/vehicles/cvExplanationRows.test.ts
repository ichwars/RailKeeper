import { describe, expect, it } from "vitest";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { deTranslations } from "../../shared/i18n/de";
import { decoderDETranslations } from "../../shared/i18n/decoder.de";

const t = (key: string) => decoderDETranslations[key] || deTranslations[key] || key;
const values = [[1, 3], [8, 151], [17, 192], [18, 128], [19, 0], [21, 0], [22, 0], [27, 24], [29, 14], [254, 0], [255, 0]]
  .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: "LokPilot 5 micro DCC" }));

describe("CV explanations", () => {
  it("uses DH characteristic and brake sources and explains only stored output registers", () => {
    const dh = [[8, 97], [29, 30], [48, 5], [115, 2], [145, 1], [154, 24], [155, 0], [254, 0]]
      .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: "DH18A" }));
    const a = decoderAnalysis(dh, "DH18A");
    const motor = cvExplanationRows(a, "motor", t);
    expect(motor[0].numbers).toEqual([48]);
    expect(motor[0].reading).toBe(5);
    expect(motor.some(row => row.key === "curve")).toBe(false);
    expect(cvExplanationRows(a, "braking", t).find(row => row.key === "constantBrake")?.numbers).toEqual([154, 155]);
    expect(a.constantBrake).toBe(true);
    expect(cvExplanationRows(a, "outputs", t).map(row => row.key)).toEqual(["115", "145"]);
    expect(cvExplanationRows(a, "outputs", t)[0].help).toContain("keine Dauer-/Moment-Einstellung");
    expect(a.railComPlus).toBeUndefined();
    const reverseOnly = decoderAnalysis(dh.map(v => v.cvNumber === 154 ? { ...v, value: 0 }
      : v.cvNumber === 155 ? { ...v, value: 24 } : v), "DH18A");
    expect(reverseOnly.constantBrake).toBe(false);
  });
  it("distinguishes the active address, inactive long address and consist permissions", () => {
    const rows = cvExplanationRows(decoderAnalysis(values, "LokPilot 5 micro DCC"), "address", t);
    expect(rows.find(row => row.key === "address")?.reading).toBe(3);
    expect(rows.find(row => row.key === "longAddress")?.reading).toBe(128);
    expect(rows.find(row => row.key === "longAddress")?.note).toBe("Nicht aktiv");
    expect(rows.find(row => row.key === "steps")?.reading).toBe("28/128");
    expect(rows.find(row => row.key === "steps")?.note).toBe("Richtung normal");
    expect(rows.find(row => row.key === "consistFunctions")?.reading).toBe("Keine freigegeben");
    expect(rows.find(row => row.key === "address")?.numbers).toEqual([1, 29]);
  });
  it("keeps missing and indexed-only values unknown and does not mix profiles", () => {
    const a = decoderAnalysis([...values, { cvNumber: 29, value: 14, decoderProfile: "Other", indexCV31: 0, indexCV32: 0 }], "Other");
    const rows = cvExplanationRows(a, "address", t);
    expect(rows.find(row => row.key === "address")?.reading).toBe("Unbekannt");
    expect(rows.find(row => row.key === "steps")?.reading).toBe("Unbekannt");
    expect(cvExplanationRows(a, "braking", t).find(row => row.key === "abc")?.reading).toBe("Unbekannt");
  });
  it("shows ESU brake states without treating DC brake bits as ABC", () => {
    const rows = cvExplanationRows(decoderAnalysis(values, "LokPilot 5 micro DCC"), "braking", t);
    expect(rows.find(row => row.key === "abc")?.reading).toBe("nein");
    expect(rows.find(row => row.key === "constantBrake")?.reading).toBe("nein");
  });
});
