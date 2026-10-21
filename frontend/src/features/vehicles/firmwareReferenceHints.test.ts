import { readFileSync } from "node:fs";
import { expect, it } from "vitest";
import { bundledDecoderInventory as inventory, type DecoderCVReference } from "../../shared/decoderLibraryModel";
import { translate } from "../../shared/i18n";
import { decoderAnalysis } from "./decoderAnalysis";
import { cvValueInfo } from "./cvValueInfo";
import { manufacturerCVHints } from "./manufacturerCVHints";

const references: DecoderCVReference[] = ["esu", "piko", "lenz", "kuehn", "uhlenbrock", "roco"]
  .flatMap(maker => JSON.parse(readFileSync(`../backend/internal/decoderlibrary/references/${maker}.json`, "utf8")));

it("shows all five ESU variants' documented firmware limits without inventing readings or sound facts", () => {
  const reference = references.find(current => current.id === "cv-gap-esu-lokpilot-5")!;
  expect(reference.modelIds).toHaveLength(5);
  for (const id of reference.modelIds) {
    const model = inventory.models.find(current => current.id === id)!;
    const hints = manufacturerCVHints(model, references);
    const value = { cvNumber: 199, value: 3, sourceLine: 17, description: "Meine eigene Angabe" };
    const before = structuredClone(value);
    const hint = hints(value)!;
    expect(hint.row).toMatchObject({ sourceIndex: 0, page: 0, applicability: "manual" });
    expect(hint.row.label).toContain("5.8.156");
    expect(hint.reference.sources[hint.row.sourceIndex].url).toBe("https://www.esu.eu/download/software/lokprogrammer/");
    expect(hint.row.factoryValue).toBeUndefined();
    expect(hints({ cvNumber: 11, value: 1 })?.row.label).toContain("5.14.183");
    for (const cvNumber of [114, 115, 131]) expect(hints({ cvNumber, value: 1 })?.row.label).toContain("5.4.139");
    for (const cvNumber of [1, 196, 197, 245, 257]) expect(hints({ cvNumber, value: 1 })).toBeUndefined();
    expect(hints({ ...value, indexCV31: 0, indexCV32: 0 })).toBeUndefined();
    const analysis = decoderAnalysis([{ cvNumber: 8, value: 151 }, value], "");
    expect(cvValueInfo(value, analysis, key => translate("de", key), "de", hint))
      .toMatchObject({ known: false, definition: undefined });
    expect(value).toEqual(before);
  }
});

it("keeps six PIKO projects' common firmware evidence separate from project defaults and version conflicts", () => {
  const reference = references.find(current => current.id === "cv-gap-piko-sound-products")!;
  expect(reference.modelIds).toHaveLength(6);
  expect(reference.note.de).toContain("Historie führt Bit4 erst mit 1.20");
  for (const id of reference.modelIds) {
    const model = inventory.models.find(current => current.id === id)!;
    const hints = manufacturerCVHints(model, references);
    expect(hints({ cvNumber: 48, value: 0 })?.row.label).toContain("1,0 V ab FW 1.16; 0,5 V ab 1.18");
    expect(hints({ cvNumber: 16, value: 1 })?.row).toMatchObject({ cv: 15, cvEnd: 16 });
    expect(hints({ cvNumber: 18, value: 128 })?.row).toMatchObject({ cv: 17, cvEnd: 18 });
    for (const cvNumber of [152, 153]) {
      expect(hints({ cvNumber, value: 10 })?.row).toMatchObject({ cv: 152, cvEnd: 153 });
    }
    expect(hints({ cvNumber: 244, value: 1 })?.row).toMatchObject({ cv: 240, cvEnd: 245 });
    for (const cvNumber of [191, 224, 236, 251, 257]) expect(hints({ cvNumber, value: 1 })).toBeUndefined();
    const value = { cvNumber: 254, value: 1, sourceExportedAt: "2026-10-17T17:41:59", description: "Meine Rampe" };
    const before = structuredClone(value);
    const hint = hints(value)!;
    expect(hint.row.label).toContain("1.20");
    expect(hint.row.factoryValue).toBeUndefined();
    expect(hint.reference.sources[hint.row.sourceIndex].url).toContain("decoder-firmwarestand-122.html");
    expect(hints({ ...value, indexCV31: 16, indexCV32: 0 })).toBeUndefined();
    const fault = hints({ cvNumber: 30, value: 16 })!;
    expect(fault.row.label).toContain("Bit4-Versionskonflikt");
    expect(fault.reference.sources[fault.row.sourceIndex].url).toContain("fragen-zu-h0-57.html");
    const analysis = decoderAnalysis([{ cvNumber: 8, value: 162 }, value], "");
    expect(cvValueInfo(value, analysis, key => translate("de", key), "de", hint))
      .toMatchObject({ known: false, definition: undefined });
    expect(value).toEqual(before);
  }
});

it("leaves all 15 unresolved exact models without borrowed CV tables", () => {
  const ids = ["cv-gap-lenz-legacy", "cv-gap-lenz-silver", "cv-gap-kuehn-t",
    "cv-gap-uhlenbrock-intellisound-8", "cv-gap-roco-motor"];
  const unresolved = references.filter(reference => ids.includes(reference.id));
  expect(unresolved.flatMap(reference => reference.modelIds)).toHaveLength(15);
  for (const reference of unresolved) for (const id of reference.modelIds) {
    const model = inventory.models.find(current => current.id === id)!;
    expect(reference.status).toBe("unavailable");
    expect(manufacturerCVHints(model, references)({ cvNumber: 50, value: 42 })).toBeUndefined();
  }
});
