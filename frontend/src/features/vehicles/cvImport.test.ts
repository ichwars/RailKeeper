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
});
