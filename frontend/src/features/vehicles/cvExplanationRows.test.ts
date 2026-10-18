import { describe, expect, it } from "vitest";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { deTranslations } from "../../shared/i18n/de";
import { decoderDETranslations } from "../../shared/i18n/decoder.de";

const t = (key: string) => decoderDETranslations[key] || deTranslations[key] || key;
const values = [[1, 3], [8, 151], [17, 192], [18, 128], [19, 0], [21, 0], [22, 0], [27, 24], [29, 14], [254, 0], [255, 0]]
  .map(([cvNumber, value]) => ({ cvNumber, value, decoderProfile: "LokPilot 5 micro DCC" }));

describe("CV explanations", () => {
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
