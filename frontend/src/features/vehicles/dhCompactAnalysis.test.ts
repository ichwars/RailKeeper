import { describe, expect, it } from "vitest";
import { bundledDecoderProfiles } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import { decoderAnalysis, decoderMapping } from "./decoderAnalysis";
import { cvExplanationRows } from "./cvExplanationRows";
import { buildSpeedCurveProfiles } from "./speedCurve";
import { functionCVSuggestion } from "./functionCVSuggestions";

const t = (key: string, values?: Record<string, string | number>) => translate("de", key, values);
const rows = (code: number, pairs: number[][] = [], decoderProfile = "") => [[8, 97], [261, code], ...pairs]
  .map(([cvNumber, value], i) => ({ cvNumber, value, decoderProfile, sourceLine: i + 4, sourceExportedAt: "2026-10-20T15:00:00" }));

describe.each([["DH05C", 52], ["DH10C", 102], ["DH16A", 160], ["DH12A", 120], ["DH14B", 141], ["DH21A", 200], ["DH18A", 180], ["DH21B", 201], ["DH22A", 202], ["DH22B", 203], ["DH24A", 204], ["DHSP10A", 110]] as const)("%s documented DCC scope", (model, code) => {
  it("recognizes the exact model, retains unknown firmware/generation and validates its own definitions", () => {
    const a = decoderAnalysis(rows(code, [[29, 6], [27, 3], [154, 0], [155, 50]]), "");
    expect(a.libraryMatch).toMatchObject({ status: "matched", profileId: `dh-${model.toLowerCase()}`, matchedBy: "cv" });
    expect(a.libraryMatch.identification?.scope).toBe("model");
    expect(a.firmware).toBeUndefined(); expect(a.libraryMatch.firmwareCompatible).toBeUndefined();
    expect(a.abc).toBe(true); expect(a.constantBrake).toBe(false); expect(a.esu5).toBe(false); expect(a.railComPlus).toBeUndefined();
    for (const cvNumber of [6, 67, 94, 98, 101, 131, 158].filter(cv =>
      !(["DH21B", "DH22B", "DH24A"].includes(model) && [98, 158].includes(cv)) && !(model === "DH24A" && cv === 101))) expect(a.definition({ cvNumber, value: 1 })).toBeUndefined();
    if (model !== "DH21A") expect(a.definition({ cvNumber: 144, value: 1 })).toBeUndefined();
    expect(decoderAnalysis(rows(code, [[27, 4], [48, 8], [137, 128], [154, -1]]), "").get(27)).toBeUndefined();
    for (const [cv, value] of [[48,8],[137,128],[154,-1]]) expect(decoderAnalysis(rows(code,[[cv,value]]),"").get(cv)).toBeUndefined();
  });
  it("uses the manufacturer's address, consist and source facts, without factory fallbacks", () => {
    const a = decoderAnalysis(rows(code, [[29, 38], [17, 195], [18, 232], [19, 170], [21, 5], [22, 5], [154, 80], [155, 0]]), "");
    expect(a.address).toBe(1000); expect(a.consistAddress).toBe(42); expect(a.consistReverse).toBe(true);
    expect(a.enabledFunctions).toEqual(["F0 →", "F1", "F3", "F9"]); expect(a.functionsComplete).toBe(true);
    expect(a.constantBrake).toBe(true); expect(a.definition({ cvNumber: 19, value: 0 })?.definition.reference?.location).toBe("S.67, CV19");
    expect(decoderAnalysis(rows(code,[[19,128]]),"").consistAddress).toBe(0);
    expect(decoderAnalysis(rows(code),"").longAddress).toBeUndefined();
    expect(decoderAnalysis(rows(code,[[29,38],[17,231],[18,255]]),"").address).toBeUndefined();
  });
  it("keeps registers above256 unindexed, and never derives key mode/inversion or edits manual data", () => {
    const values = rows(code, [[35, 4], [137, 1], [401, 29], [115, 1], [119, 10]]);
    const before = structuredClone(values), a = decoderAnalysis(values, "");
    expect(a.definition(values[3])?.definition).toBeDefined();
    expect(a.definition({ cvNumber: 401, value: 29, indexCV31: 1, indexCV32: 1 })).toBeUndefined();
    expect(decoderMapping(a, t)).toEqual([]);
    const manual = { name: "Mein Licht", mode: "moment", inverted: true, notes: "Eigener Nachweis" };
    expect(functionCVSuggestion(a, "F1", manual, [], t)).toEqual({ notesFull: false });
    expect(manual).toEqual({ name: "Mein Licht", mode: "moment", inverted: true, notes: "Eigener Nachweis" });
    for (const topic of ["address", "motor", "braking", "outputs", "communication"] as const)
      for (const row of cvExplanationRows(a, topic, t)) expect([row.name, row.help, row.reading].join(" ")).not.toMatch(/vehicles\./);
    expect(values).toEqual(before);
  });
  it("shows the documented characteristic parameter without invented DCC points", () => {
    const curve = buildSpeedCurveProfiles(rows(code, [[29, 6], [48, 5], [2, 0], [5, 92], [6, 50], [67, 100]]), "Decoder")[0];
    expect(curve.motorCharacteristic).toEqual({ cvNumber: 48, value: 5 }); expect(curve.primaryCurve).toBe("none");
    const missing = buildSpeedCurveProfiles(rows(code, [[29, 6], [6, 50], [67, 100]]), "Decoder")[0];
    expect(missing.motorCharacteristic?.value).toBeNull(); expect(missing.primaryCurve).toBe("none");
    expect(buildSpeedCurveProfiles(rows(code, [[48, 8]]), "Decoder")[0].motorCharacteristic?.value).toBeNull();
  });
});

it("keeps the DH18A-specific SUSI default out of the compact models and records generation limits", () => {
  for (const id of ["dh-dh05c", "dh-dh10c", "dh-dh18a"]) {
    const p = bundledDecoderProfiles.find(p => p.id === id)!;
    expect(p.definitions).toHaveLength(120);
    expect(p.definitions.find(d => d.cv === 137)?.factoryDefault).toBe(id === "dh-dh18a" ? 1 : 0);
    if (id !== "dh-dh18a") {
      for (const cv of [137, 129, 151, 465]) expect(p.definitions.find(d => d.cv === cv)?.help.de.toLowerCase()).toContain("generation");
    }
  }
});

it("does not interpret an older verified firmware, a different model, or a manufacturer code alone", () => {
  const old = rows(102, [[7,16],[48,5]], "DH10C").map(r => r.cvNumber === 7
    ? { ...r, description: "LokProgrammer Decoder: DH10C (3.14.095)" } : r);
  expect(decoderAnalysis(old,"DH10C").modelRules).toBe(false);
  expect(decoderAnalysis(old,"DH10C").get(48)).toBeUndefined();
  expect(decoderAnalysis(rows(102, [], "DH05C"), "DH05C").libraryMatch.status).toBe("conflict");
  expect(decoderAnalysis([{ cvNumber:8, value:97 }], "").libraryMatch.profileId).toBe("dcc-basics");
});


it("uses the DH16A-specific hardware facts without borrowing compact-generation restrictions", () => {
  const a = decoderAnalysis(rows(160, [[137,0],[35,4],[129,10]]), "");
  expect(a.libraryMatch.profileId).toBe("dh-dh16a");
  const susi = a.definition({ cvNumber:137, value:0 })!.definition;
  expect(susi.factoryDefault).toBe(0);
  expect(susi.reference?.location).toBe("S.20–21/75/105, CV137");
  expect(susi.help.de).toContain("in beiden Generationen verstärkt");
  for (const cv of [127,129,151]) {
    const d = a.definition({cvNumber:cv,value:0})!.definition;
    expect(d.help.de).not.toContain("Generation2");
    expect(d.help.de).toContain("mit SUSI");
  }
  const identity = a.definition({cvNumber:261,value:160})!.definition;
  expect(identity.help.de).toContain("160");
  expect(identity.help.de).toContain("keine Bestimmung der Hardwaregeneration");
});


it("keeps DH12A AUX3–6 at logic level and uses its own pin-use and type references", () => {
  const a = decoderAnalysis(rows(120, [[137,0],[35,4],[129,10]]), "");
  expect(a.libraryMatch.profileId).toBe("dh-dh12a");
  const susi = a.definition({ cvNumber:137, value:0 })!.definition;
  expect(susi.factoryDefault).toBe(0);
  expect(susi.reference?.location).toBe("S.18/75/105, CV137");
  expect(susi.help.de).toContain("AUX3–6 sind unverstärkte Logikausgänge");
  expect(susi.help.de).toContain("AUX3/4 haben eigene Anschlüsse");
  for (const cv of [35,153,156]) {
    const definition = a.definition({cvNumber:cv,value:0})!.definition;
    expect(definition.bits?.filter(bit => bit.bit === 4 || bit.bit === 5).every(bit => bit.name.de.includes("Logikpegel"))).toBe(true);
  }
  for (const cv of [127,129,151]) expect(a.definition({cvNumber:cv,value:0})!.definition.help.de).toContain("mit SUSI");
  expect(a.definition({cvNumber:261,value:120})!.definition.help.de).toContain("120 steht für DH12A");
});


it("keeps DH14B logical AUX5/6 on the shared pins and never identifies SUSI from bit0 alone", () => {
  const a = decoderAnalysis(rows(141, [[137,33],[35,4],[129,10]]), "");
  expect(a.libraryMatch.profileId).toBe("dh-dh14b");
  const susi = a.definition({ cvNumber:137, value:33 })!.definition;
  expect(susi.factoryDefault).toBe(0);
  expect(susi.reference?.location).toBe("S.19/75/105, CV137");
  expect(susi.help.de).toContain("33 deaktiviert SUSI");
  expect(susi.help.de).toContain("49 alternativ AUX5/6 auf denselben Pins");
  expect(susi.bits?.[0].on).toEqual(susi.bits?.[0].off);
  for (const cv of [35,127,129,151,153,465]) {
    const d = a.definition({ cvNumber:cv, value:0 })!.definition;
    expect(d.help.de).toContain("dieselben Pins");
    expect(d.help.de).toContain("keine zusätzlichen Anschlüsse");
  }
  const old = rows(141, [[7,16],[48,5]], "DH14B").map(row => row.cvNumber === 7
    ? { ...row, description: "LokProgrammer Decoder: DH14B (3.14.095)" } : row);
  expect(decoderAnalysis(old, "DH14B").modelRules).toBe(false);
  expect(a.definition({ cvNumber:261, value:141 })!.definition.help.de).toContain("141 steht für DH14B");
});


it("keeps DH21A generation and AUX5/6 levels unresolved and explains only corroborated CV144 bits", () => {
  const a = decoderAnalysis(rows(200, [[137,0],[144,25]]), "");
  expect(a.libraryMatch.profileId).toBe("dh-dh21a");
  expect(a.firmware).toBeUndefined();
  const susi = a.definition({ cvNumber:137, value:0 })!.definition;
  expect(susi.reference?.location).toBe("S.24–27/75/105, CV137");
  expect(susi.help.de).toContain("-4 in beiden Generationen verstärkt");
  expect(susi.help.de).toContain("-5 dagegen Logikpegel");
  expect(susi.help.de).toContain("Pegel bleibt ungeklärt");
  const partial = a.definition({ cvNumber:144, value:25 })!.definition;
  expect(partial.bits?.map(bit => bit.bit)).toEqual([1,2,3,4,5]);
  expect(partial.help.de).toContain("Bit0 bleibt ungeklärt");
  expect(partial.help.de).toContain("nur für Generation2");
  expect(a.abc).toBeUndefined(); expect(a.constantBrake).toBeUndefined();
  expect(a.definition({ cvNumber:159, value:0 })).toBeUndefined();
  expect(decoderAnalysis(rows(200, [[144,64]]), "").get(144)).toBeUndefined();
  const alias = decoderAnalysis(rows(200, [], "DH21A-4"), "DH21A-4");
  expect(alias.libraryMatch).toMatchObject({ profileId:"dh-dh21a", status:"matched", matchedBy:"name" });
  expect(alias.libraryMatch.identification?.scope).toBe("model");
  expect(bundledDecoderProfiles.find(profile => profile.id === "dh-dh21a")?.definitions).toHaveLength(121);
});
