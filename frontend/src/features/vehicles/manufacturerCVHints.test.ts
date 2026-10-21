import { readFileSync } from "node:fs";
import { describe, expect, it } from "vitest";
import { bundledDecoderInventory as inventory, type DecoderCVReference, type DecoderCVReferenceRow } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvValueInfo } from "./cvValueInfo";
import { manufacturerCVHints, manufacturerReferenceModel } from "./manufacturerCVHints";

const references: DecoderCVReference[] = ["esu", "zimo", "doehler-haass", "lenz", "kuehn", "uhlenbrock", "roco", "piko"]
  .flatMap(maker => JSON.parse(readFileSync(`../backend/internal/decoderlibrary/references/${maker}.json`, "utf8")))
  .map(reference => ({ ...reference, active: true }));
const model = inventory.models.find(model => model.name === "N025")!;
const source = { title: "Handbuch", url: "https://example.test/manual.pdf", version: "1", scope: "Testmodell" };
const row: DecoderCVReferenceRow = { cv: 54, label: "Dokumentierter Parameter", page: 4, sourceIndex: 0,
  indexScope: "base", applicability: "manual", factoryValue: "15" };
const reference = (rows: DecoderCVReferenceRow[], patch: Partial<DecoderCVReference> = {}): DecoderCVReference => ({
  id: "test", revision: "test", modelIds: [model.id], title: { de: "Titel", en: "Title" },
  note: { de: "Firmware beachten", en: "Check firmware" }, status: "partial", sources: [source], rows, active: true, ...patch
});
const analysis = (label = "Kühn N025", manufacturer = 157) => decoderAnalysis([
  { cvNumber: 8, value: manufacturer, decoderProfile: label }, { cvNumber: 54, value: 42, decoderProfile: label }
], label);

it("requires an exact unique model or technical identity and rejects conflicts and family guesses", () => {
  expect(manufacturerReferenceModel(analysis(), inventory, "Kühn N025", "Eigener Name")).toEqual(model);
  expect(manufacturerReferenceModel(analysis("", 97), inventory, "DH21B", "")).toBeDefined();
  expect(manufacturerReferenceModel(analysis("", 151), inventory, "Kühn N025", "")).toBeUndefined();
  expect(manufacturerReferenceModel(analysis(""), inventory, "N025", "N45")).toBeUndefined();
  expect(manufacturerReferenceModel(analysis("", 99), inventory, "10104", "")).toBeUndefined();
  expect(manufacturerReferenceModel(analysis("", 145), inventory, "ZIMO MN180", "")).toBeUndefined();
  const unknownType = decoderAnalysis([{ cvNumber: 8, value: 97 }, { cvNumber: 261, value: 255 }], "");
  expect(manufacturerReferenceModel(unknownType, inventory, "", "DH21B")).toBeUndefined();
  const identified = decoderAnalysis([{ cvNumber: 8, value: 97 }, { cvNumber: 261, value: 201 }], "");
  expect(manufacturerReferenceModel(identified, inventory, "", "Eigener Name")?.name).toBe("DH21B");
  expect(manufacturerReferenceModel({ ...identified, libraryMatch: { ...identified.libraryMatch, status: "conflict" } }, inventory, "DH21B", "")).toBeUndefined();
});

it("never promotes a reference heading to a functional rule or changes raw/manual data", () => {
  const values = [{ cvNumber: 54, value: 42, decoderProfile: "Kühn N025", description: "Meine Beschreibung",
    sourceLine: 17, sourceExportedAt: "2026-10-17T17:41:59" }];
  const before = structuredClone(values);
  const hint = manufacturerCVHints(model, [reference([row])])(values[0]);
  const info = cvValueInfo(values[0], analysis(), key => translate("de", key), "de", hint);
  expect(info).toMatchObject({ known: false, name: row.label, definition: undefined, profile: undefined });
  expect(info.help).toContain("kein Betriebszustand");
  expect(analysis().get(54)).toBeUndefined();
  expect(values).toEqual(before);
  const known = cvValueInfo({ cvNumber: 1, value: 3 }, analysis(), key => translate("de", key), "de", hint);
  expect(known.name).toBe("Kurze DCC-Adresse"); expect(known.manufacturerHint).toBeUndefined();
});

it("isolates base/indexed/unknown pages, unindexed high CVs, conflicts and active model scope", () => {
  const indexed = { ...row, label: "Indiziert", indexScope: "indexed" as const, indexCV31: 18, indexCV32: 0 };
  const hints = manufacturerCVHints(model, [reference([row, indexed,
    { ...row, cv: 260 }, { ...row, cv: 55, indexScope: "unspecified" },
    { ...row, cv: 56 }, { ...row, cv: 56, label: "Widerspruch" }])]);
  expect(hints({ cvNumber: 54, value: 1 })?.row.label).toBe(row.label);
  expect(hints({ cvNumber: 54, value: 1, indexCV31: 18, indexCV32: 0 })?.row.label).toBe("Indiziert");
  for (const cvNumber of [55, 56]) expect(hints({ cvNumber, value: 1 })).toBeUndefined();
  expect(hints({ cvNumber: 260, value: 1 })).toBeDefined();
  expect(hints({ cvNumber: 260, value: 1, indexCV31: 0, indexCV32: 0 })).toBeUndefined();
  expect(hints({ cvNumber: 54, value: 1, indexCV31: 18, indexCV32: 1 })).toBeUndefined();
  expect(hints({ cvNumber: 54, value: 1, indexCV31: 18 })).toBeUndefined();
  for (const patch of [{ active: false }, { modelIds: ["another-model"] }, { status: "unavailable" as const }])
    expect(manufacturerCVHints(model, [reference([row], patch)])({ cvNumber: 54, value: 1 })).toBeUndefined();
});

it("uses the reviewed BR491 effect and sound pages independently without changing vehicle values", () => {
  const br491 = inventory.models.find(current => current.id === "piko-sound-products-36520")!;
  const hints = manufacturerCVHints(br491, references);
  const values = [
    { cvNumber: 257, value: 3, indexCV31: 18, indexCV32: 0, description: "Meine Lichtangabe" },
    { cvNumber: 257, value: 150, indexCV31: 16, indexCV32: 0 }
  ];
  const before = structuredClone(values);
  const effect = hints(values[0])!;
  expect(effect).toMatchObject({ row: { label: "Ausgang F0v Effekt A", page: 4 } });
  expect(effect.reference.sources[effect.row.sourceIndex].url).toBe("https://www.piko-shop.de/is.php?id=38019");
  expect(hints(values[1])).toMatchObject({ row: { label: "Gesamtlautstärke", page: 5 } });
  expect(hints({ cvNumber: 257, value: 3 })).toBeUndefined();
  expect(hints({ cvNumber: 257, value: 3, indexCV31: 18, indexCV32: 1 })).toBeUndefined();
  const info = cvValueInfo(values[0], analysis("PIKO 36520", 162), key => translate("de", key), "de", hints(values[0]));
  expect(info).toMatchObject({ known: false, definition: undefined });
  expect(values).toEqual(before);
});

it("keeps reused Lenz articles ambiguous and the V2 documentation gap visible", () => {
  expect(manufacturerReferenceModel(analysis("", 99), inventory, "10104", "")).toBeUndefined();
  const v2 = inventory.models.find(current => current.id === "lenz-silver-silver-mini-v2")!;
  const v2Refs = references.filter(reference => reference.modelIds.includes(v2.id));
  expect(v2Refs).toHaveLength(1);
  expect(v2Refs[0].status).toBe("unavailable");
  expect(manufacturerCVHints(v2, v2Refs)({ cvNumber: 51, value: 1 })).toBeUndefined();
  const original = inventory.models.find(current => current.id === "lenz-legacy-lf100xf")!;
  expect(manufacturerCVHints(original, references)({ cvNumber: 56, value: 1 })?.row.label).toBe("Blinkfrequenz");
});

it("keeps legacy Lenz lighting, motor and SUSI facts separate with their original provenance", () => {
  const hintsFor = (id: string) => manufacturerCVHints(inventory.models.find(model => model.id === id)!, references);
  const lf = hintsFor("lenz-legacy-lf101xf");
  const direct = hintsFor("lenz-legacy-silver-direct");
  const mini = hintsFor("lenz-legacy-silver-mini");
  const values = [{ cvNumber: 50, value: 42, description: "Meine Helligkeit", sourceLine: 17 }];
  const before = structuredClone(values);
  const light = lf(values[0])!;
  expect(light.row).toMatchObject({ label: "Dimming value for output A", factoryValue: "255", page: 10 });
  expect(light.reference.sources[light.row.sourceIndex]).toMatchObject({
    version: "Art. 10104, englische Ausgabe Februar 2007",
    url: "https://www.yumpu.com/en/document/view/34965399/lf101xf-six-function-dcc-function-only-decoder-lenz-usa"
  });
  expect(direct(values[0])?.row.label).toContain("Motor configuration");
  expect(lf({ cvNumber: 56, value: 1 })?.row.label).toBe("Function mapping: dimming output C");
  for (const cvNumber of [43, 62, 63, 64]) expect(mini({ cvNumber, value: 1 })).toBeUndefined();
  expect(mini({ cvNumber: 80, value: 42 })?.row).toMatchObject({ cv: 67, cvEnd: 94, page: 21 });
  const gold = hintsFor("lenz-legacy-gold-maxi");
  expect(gold({ cvNumber: 126, value: 102 })?.row.label).toContain("SUSI");
  expect(gold({ cvNumber: 126, value: 102, indexCV31: 0, indexCV32: 0 })).toBeUndefined();
  expect(cvValueInfo(values[0], analysis("", 99), key => translate("de", key), "de", light))
    .toMatchObject({ known: false, definition: undefined, name: "Dimming value for output A" });
  expect(values).toEqual(before);
});

it("uses original LE030 and LE0511A/D tables with exact boundaries and source pages", () => {
  const hintsFor = (id: string) => manufacturerCVHints(inventory.models.find(model => model.id === id)!, references);
  const early = hintsFor("lenz-legacy-le030");
  expect(early({ cvNumber: 81, value: 12 })?.row).toMatchObject({ cv: 67, cvEnd: 81, page: 10, valueRange: "0-63" });
  for (const cvNumber of [6, 82, 94]) expect(early({ cvNumber, value: 12 })).toBeUndefined();
  expect(early({ cvNumber: 60, value: 3 })?.reference.sources[0].version).toContain("Juni 1997");
  for (const id of ["lenz-legacy-le0511a", "lenz-legacy-le0511d"]) {
    const hints = hintsFor(id);
    const value = { cvNumber: 52, value: 42, sourceLine: 17, description: "Meine Lichtangabe" };
    const before = structuredClone(value);
    const hint = hints(value)!;
    expect(hint.row).toMatchObject({ label: "Helligkeit von Funktionsausgang A", page: 4, factoryValue: "64" });
    expect(hint.reference.sources[hint.row.sourceIndex]).toMatchObject({
      version: "Art. 10111/10112, 3. Auflage 07/03", url: "https://www.lenz-elektronik.de/src/pdf/b_10111.pdf"
    });
    expect(hints({ cvNumber: 94, value: 12 })?.row).toMatchObject({ cv: 67, cvEnd: 94, page: 4 });
    for (const cvNumber of [5, 6, 60, 95]) expect(hints({ cvNumber, value: 12 })).toBeUndefined();
    expect(hints({ ...value, indexCV31: 0, indexCV32: 0 })).toBeUndefined();
    expect(cvValueInfo(value, analysis("", 99), key => translate("de", key), "de", hint))
      .toMatchObject({ known: false, definition: undefined, name: "Helligkeit von Funktionsausgang A" });
    expect(value).toEqual(before);
  }
});

it("uses the MX622 OEM CV source for the four proven Roco articles without inferring firmware or defaults", () => {
  for (const article of ["685404", "685504", "686101", "686201"]) {
    const current = inventory.models.find(model => model.id.startsWith(`roco-motor-${article}-`))!;
    const hints = manufacturerCVHints(current, references);
    const value = { cvNumber: 57, value: 42, sourceLine: 17, description: "Meine Motorspannung" };
    const before = structuredClone(value);
    const hint = hints(value)!;
    expect(hint.row).toMatchObject({ label: "Motorregelung: Referenzspannung", page: 69, sourceIndex: 1, applicability: "manual" });
    expect(hint.row.factoryValue).toBeUndefined();
    expect(hint.reference.sources[hint.row.sourceIndex]).toMatchObject({
      url: "https://www.zimo.at/web2010/documents/MX-KleineDecoder.pdf",
      version: "Stand 19.08.2024, 79-seitige Herstellerfassung"
    });
    expect(hint.reference.sources[0].url).toContain("20220627_FLM_Zubehoer_DE.pdf");
    expect(hints({ cvNumber: 18, value: 128 })?.row).toMatchObject({ cv: 17, cvEnd: 18 });
    expect(hints({ cvNumber: 18, value: 128 })?.row.valueRange).toBeUndefined();
    expect(hints({ cvNumber: 94, value: 255 })?.row).toMatchObject({ cv: 67, cvEnd: 94 });
    for (const cvNumber of [11, 20, 97, 100, 260]) expect(hints({ cvNumber, value: 42 })).toBeUndefined();
    expect(hints({ ...value, indexCV31: 0, indexCV32: 0 })).toBeUndefined();
    expect(cvValueInfo(value, analysis("", 145), key => translate("de", key), "de", hint))
      .toMatchObject({ known: false, definition: undefined, name: "Motorregelung: Referenzspannung" });
    expect(value).toEqual(before);
  }
  const other = inventory.models.find(model => model.id.startsWith("roco-motor-10734-"))!;
  expect(manufacturerCVHints(other, references)({ cvNumber: 57, value: 42 })).toBeUndefined();
});

describe("sequential catalogue-wide reference audit", () => {
  it("checks every model without upgrading uncertain documentary facts to automatic interpretation", () => {
    let covered = 0, gaps = 0, hintsCount = 0;
    for (const current of inventory.models) {
      const refs = references.filter(reference => reference.modelIds.includes(current.id));
      expect(refs.length, current.id).toBeGreaterThan(0);
      const before = JSON.stringify(refs);
      const hints = manufacturerCVHints(current, refs);
      if (refs.some(reference => reference.rows.length)) covered++; else gaps++;
      const seen = new Set<string>();
      for (const ref of refs) for (const entry of ref.rows) {
        for (let cvNumber = entry.cv; cvNumber <= (entry.cvEnd ?? entry.cv); cvNumber++) {
          const value = { cvNumber, value: 42, indexCV31: entry.indexCV31, indexCV32: entry.indexCV32 };
          if (entry.indexScope === "unspecified") continue;
          const hint = hints(value);
          if (!hint) continue; // Contradictory rows are intentionally omitted.
          expect(hint.reference.modelIds, current.id).toContain(current.id);
          expect(hint.reference.sources[hint.row.sourceIndex]?.url).toMatch(/^https?:\/\//);
          expect(hint.row.indexScope).not.toBe("unspecified");
          seen.add(`${cvNumber}/${value.indexCV31}/${value.indexCV32}`);
          if (entry.indexScope === "base") expect(hints({ ...value, indexCV31: 255, indexCV32: 255 })).toBeUndefined();
        }
      }
      hintsCount += seen.size;
      expect(JSON.stringify(refs)).toBe(before);
    }
    expect(inventory.models).toHaveLength(599); expect(covered).toBe(580); expect(gaps).toBe(19);
    expect(hintsCount).toBeGreaterThan(60000);
  });
});
