import { describe, expect, it } from "vitest";

import { cvValueFixture } from "../../test/fixtures/vehicles";
import {
  buildCVImportPreview,
  cvValuesFromImport,
  functionMappingsFromImport,
  isValidCVValueInput,
  isValidFunctionMapping
} from "./cvImport";

describe("CV imports", () => {
  it("parses JSON and semicolon rows", () => {
    expect(cvValuesFromImport('[{"cvNumber":1,"value":3,"description":"Adresse"}]')).toMatchObject([
      { cvNumber: 1, value: 3, description: "Adresse" }
    ]);
    expect(cvValuesFromImport("CV;Wert;Beschreibung\n29;14;Konfiguration")).toMatchObject([
      { cvNumber: 29, value: 14, description: "Konfiguration" }
    ]);
  });

  it("marks new, changed, same, duplicate and invalid rows", () => {
    const existing = cvValueFixture();
    const preview = buildCVImportPreview(
      "decoder.txt",
      [
        { cvNumber: 1, value: 3, description: "Adresse", decoderProfile: "ESU LokPilot 5" },
        { cvNumber: 2, value: 5, decoderProfile: "ESU LokPilot 5" },
        { cvNumber: 1, value: 4, decoderProfile: "ESU LokPilot 5" },
        { cvNumber: 2, value: 7, decoderProfile: "ESU LokPilot 5" },
        { cvNumber: 0, value: 300 }
      ],
      [existing]
    );

    expect(preview.rows.map((row) => row.status)).toEqual(["same", "new", "invalid", "invalid", "invalid"]);
    expect(preview.rows.map((row) => row.selected)).toEqual([false, true, false, false, false]);
  });

  it("marks an existing CV with a changed value", () => {
    const preview = buildCVImportPreview(
      "decoder.json",
      [{ cvNumber: 1, value: 4, description: "Adresse", decoderProfile: "ESU LokPilot 5" }],
      [cvValueFixture()]
    );

    expect(preview.rows[0]).toMatchObject({ status: "changed", selected: true, message: "ändert Wert" });
  });

  it("validates CV values and function mappings", () => {
    expect(isValidCVValueInput({ cvNumber: 1024, value: 255 })).toBe(true);
    expect(isValidCVValueInput({ cvNumber: 1025, value: 0 })).toBe(false);

    const [mapping] = functionMappingsFromImport(
      '{"functions":[{"functionKey":"f1","name":"Sound","functionType":"sound","mode":"moment"}]}'
    );
    expect(mapping.functionKey).toBe("F1");
    expect(isValidFunctionMapping(mapping)).toBe(true);
    expect(isValidFunctionMapping({ ...mapping, functionKey: "F99" })).toBe(false);
  });

  it("preserves indexed JSON and named CSV columns, including zero indexes", () => {
    const rows = cvValuesFromImport('cv_number;value;cv31;cv32;description\n257;4;16;3;"A; B"\n257;5;0;0;Zero\n1;3;;;Adresse');
    expect(rows).toMatchObject([
      { cvNumber: 257, value: 4, indexCV31: 16, indexCV32: 3, description: "A; B" },
      { cvNumber: 257, value: 5, indexCV31: 0, indexCV32: 0 },
      { cvNumber: 1, value: 3, indexCV31: undefined, indexCV32: undefined }
    ]);
    expect(cvValuesFromImport(JSON.stringify({ cvValues: rows }))).toEqual(rows);
    expect(buildCVImportPreview("pages.csv", rows, []).rows.every((row) => row.status === "new")).toBe(true);
    expect(cvValuesFromImport('[{"cvNumber":1,"value":null}]')[0].value).toBeNaN();
    expect(cvValuesFromImport('[{"cvNumber":1,"value":3,"sourceFileId":"foreign"}]')[0].sourceFileId).toBe("");
  });

  it("preserves provenance in JSON/CSV and detects metadata-only changes", () => {
    const provenance = { sourceLine: 4, sourceExportedAt: "2026-10-17T17:41:59" };
    const existing = cvValueFixture();
    const input = { ...existing, ...provenance };
    expect(cvValuesFromImport(JSON.stringify({ cvValues: [input] }))[0]).toMatchObject(provenance);
    expect(cvValuesFromImport("CV;Wert;sourceLine;sourceExportedAt\n1;3;4;2026-10-17T17:41:59")[0]).toMatchObject(provenance);
    expect(cvValuesFromImport("\nCV;Wert\n\n1;3")[0].sourceLine).toBe(4);
    expect(buildCVImportPreview("data.json", [input], [existing]).rows[0]).toMatchObject({ status: "changed", selected: true, message: "ändert Originalzeile, Exportdatum" });
    expect(buildCVImportPreview("data.json", [input], [input]).rows[0].status).toBe("same");
    expect(isValidCVValueInput({ cvNumber: 1, value: 3, sourceLine: 0 })).toBe(false);
    expect(isValidCVValueInput({ cvNumber: 1, value: 3, sourceExportedAt: "2026-02-30T17:41:59" })).toBe(false);
    expect(isValidCVValueInput({ cvNumber: 1, value: 3, sourceExportedAt: "2026-10-17T17:41:59Z" })).toBe(false);
  });

  it("keeps equal CV numbers separate across index pages and rejects incomplete pairs", () => {
    const existing = cvValueFixture({ cvNumber: 257, indexCV31: 16, indexCV32: 3 });
    const inputs = [
      { cvNumber: 257, value: 3, indexCV31: 16, indexCV32: 4, decoderProfile: existing.decoderProfile },
      { cvNumber: 257, value: 3, indexCV31: 16, decoderProfile: existing.decoderProfile },
      { cvNumber: 1, value: 3, indexCV31: 16, indexCV32: 0 }
    ];
    expect(buildCVImportPreview("pages.json", inputs, [existing]).rows.map((row) => row.status))
      .toEqual(["new", "invalid", "invalid"]);
    expect(isValidCVValueInput({ cvNumber: 257, value: 0, indexCV31: 0, indexCV32: 0 })).toBe(true);
  });
});
