import { describe, expect, it } from "vitest";

import type { ArticleSearchResponse, MasterDataEntry } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import {
  emptyVehicle, hasArticleSearchCriteria, hasQrPayloadData, inferFunctionTypeFromSymbol,
  isBadArticleValue, sanitizeArticleSearchResponse
} from "./vehicleViewModel";

describe("vehicle suggestions preserve identities and reject unrelated source text", () => {
  it.each([
    ["sound", "sound"], ["horn", "sound"], ["pfiff", "sound"],
    ["licht", "licht"], ["light", "licht"], ["lampe", "licht"],
    ["kupplung", "kupplung"], ["rauch", "rauch"], ["smoke", "rauch"],
    ["warnung", "sonderfunktion"], ["sonder", "sonderfunktion"], ["sifa", "sonderfunktion"]
  ])("recognizes the function hint %s without changing the saved symbol", (key, expected) => {
    expect(inferFunctionTypeFromSymbol(key, [])).toBe(expected);
  });

  it("uses active symbol labels, retains unknown function types and ignores inactive labels", () => {
    const symbol: MasterDataEntry = {
      id: "original", type: "function_symbol", key: "legacy-custom", label: "Horn",
      active: true, sortOrder: 10, metadata: {}, createdAt: "2026-11-08", updatedAt: "2026-11-08"
    };
    expect(inferFunctionTypeFromSymbol(symbol.key, [symbol])).toBe("sound");
    expect(inferFunctionTypeFromSymbol(symbol.key, [{ ...symbol, active: false }], "legacy-type"))
      .toBe("legacy-type");
    expect(inferFunctionTypeFromSymbol("unknown", [], "")).toBe("standard");
    expect(inferFunctionTypeFromSymbol("", [symbol], "legacy-type")).toBe("standard");
    expect(symbol.key).toBe("legacy-custom");
  });

  it("requires a complete article identity unless an explicit EAN is provided", () => {
    const form = { ...emptyVehicle, manufacturer: "Stored maker", gauge: "TT", name: "Stored name" };
    expect(hasArticleSearchCriteria(form)).toBe(true);
    expect(hasArticleSearchCriteria({ ...form, manufacturer: " " })).toBe(false);
    expect(hasArticleSearchCriteria({ ...form, name: "", articleNumber: "" })).toBe(false);
    expect(hasArticleSearchCriteria(form, { fields: { ean: " 1234567890123 " } })).toBe(true);
    expect(hasArticleSearchCriteria(form, {})).toBe(false);
    expect(hasArticleSearchCriteria(form, {
      fields: { manufacturer: "Maker", articleNumber: "0001", gauge: "TT" }
    })).toBe(true);
    expect(hasArticleSearchCriteria(form, {
      manufacturer: "Maker", name: "Name", gauge: "TT", fields: {}
    })).toBe(true);
    expect(hasArticleSearchCriteria(form, { name: "Name", manufacturer: "Maker" })).toBe(false);
    expect(hasArticleSearchCriteria(form, { gauge: "TT", articleNumber: "0001" })).toBe(false);
  });

  it("keeps a QR identity from the stored vehicle when the current form is incomplete", () => {
    expect(hasQrPayloadData(null, emptyVehicle)).toBe(false);
    expect(hasQrPayloadData(null, { ...emptyVehicle, inventoryNumber: "000001" })).toBe(true);
    expect(hasQrPayloadData(null, { ...emptyVehicle, name: "Named draft" })).toBe(true);
    expect(hasQrPayloadData(vehicleFixture(), emptyVehicle)).toBe(true);
    expect(hasQrPayloadData(vehicleFixture({ inventoryNumber: "", name: "" }), emptyVehicle)).toBe(false);
  });

  it.each([
    ["lengthMm", "19", true], ["lengthMm", "20", false], ["lengthMm", "600", false],
    ["lengthMm", "600,1", true], ["lengthMm", "150,5", false], ["lengthMm", "unknown", true],
    ["lightingDescription", "Fahrtrichtung", true], ["lightingDescription", "Lichtwechsel", true],
    ["lightingDescription", "Spitzenlicht", true], ["lightingDescription", "Schlusslicht", true],
    ["lightingDescription", "Innenraumbeleuchtung", false],
    ["headlightsDescription", "Altersempfehlung", true], ["headlightsDescription", "Downloads", true],
    ["headlightsDescription", "Bedienungsanleitung", true], ["headlightsDescription", "weiß/rot", false],
    ["soundGeneratorDescription", "Menu", true], ["soundGeneratorDescription", "Menü", true],
    ["soundGeneratorDescription", "Menue", true], ["soundGeneratorDescription", "Sprunggröße", true],
    ["soundGeneratorDescription", "Sprunggroesse", true], ["soundGeneratorDescription", "Wählen Sie", true],
    ["soundGeneratorDescription", "Waehlen Sie", true], ["soundGeneratorDescription", "Dieselsound", false],
    ["description", "Cookie settings", true], ["name", "   ", true], ["name", "BR 218", false]
  ])("classifies imported %s text %s without altering valid field content", (key, value, rejected) => {
    expect(isBadArticleValue(key, value)).toBe(rejected);
  });

  it("removes unsuitable suggestions without mutating provenance, source objects or valid bytes", () => {
    const response: ArticleSearchResponse = {
      query: "Stored maker 0001", sources: ["manufacturer"], results: [{
        source: "manufacturer", title: "Original source", url: "https://maker.example/0001",
        snippet: "Original snippet", score: 95, fields: {
          name: { label: "Name", value: "  BR 218  ", confidence: 0.9 },
          lengthMm: { label: "Length", value: "9999", confidence: 0.9 },
          description: { label: "Description", value: "Cookie preferences", confidence: 0.9 }
        }, images: [{ url: "https://maker.example/photo.png", title: "Original image", source: "manufacturer" }]
      }]
    };
    const before = structuredClone(response);
    const filtered = sanitizeArticleSearchResponse(response);
    expect(filtered.results[0].fields).toEqual({ name: response.results[0].fields.name });
    expect(filtered.results[0].fields.name.value).toBe("  BR 218  ");
    expect(filtered.results[0]).toMatchObject({ url: response.results[0].url, images: response.results[0].images });
    expect(filtered.query).toBe(response.query);
    expect(response).toEqual(before);
    expect(sanitizeArticleSearchResponse({ query: "No match", results: [] }).results).toEqual([]);
  });
});
