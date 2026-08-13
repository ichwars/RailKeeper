import { describe, expect, it } from "vitest";

import type { ArticleSearchResult, MasterDataEntry } from "../../shared/api";
import { articleSelectionKey } from "../../shared/articleSearch/articleSearchModel";
import { emptyArticleEditorForm } from "./articleEditorModel";
import {
  accessorySearchInput,
  applyAccessorySearchResult,
  currentAccessorySearchValue,
  hasAccessorySearchCriteria,
  isUsableAccessorySearchValue
} from "./accessoryArticleSearch";

const timestamp = "2026-08-13T10:00:00Z";
const manufacturer: MasterDataEntry = {
  id: "manufacturer:tillig",
  type: "manufacturer",
  key: "tillig",
  label: "Tillig",
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: timestamp,
  updatedAt: timestamp
};
const gauge: MasterDataEntry = {
  id: "gauge:tt",
  type: "gauge",
  key: "tt",
  label: "TT",
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: timestamp,
  updatedAt: timestamp
};
const result: ArticleSearchResult = {
  source: "manufacturer",
  title: "Tillig 83101",
  url: "https://www.tillig.com/83101.html",
  snippet: "Gerades Modellgleis",
  score: 100,
  fields: {
    manufacturer: { label: "Hersteller", value: "Tillig", confidence: 1 },
    articleNumber: { label: "Artikelnummer", value: "83101", confidence: 1 },
    name: { label: "Bezeichnung", value: "Gerades Modellgleis", confidence: 0.9 },
    ean: { label: "EAN", value: "4012500831012", confidence: 0.9 },
    gauge: { label: "Spurweite", value: "TT", confidence: 1 },
    scale: { label: "Maßstab", value: "1:120", confidence: 0.8 },
    description: { label: "Beschreibung", value: "Bettungsgleis", confidence: 0.8 },
    articleSourceUrl: {
      label: "Produktquelle",
      value: "https://www.tillig.com/83101.html",
      confidence: 1
    },
    category: { label: "Kategorie", value: "Gleis", confidence: 0.7 }
  }
};

describe("accessory article search", () => {
  it("maps article criteria without inferring article type", () => {
    const input = accessorySearchInput({
      ...emptyArticleEditorForm(),
      manufacturer: "Tillig",
      articleNumber: "83101",
      gauges: ["TT"],
      ean: "4012500831012"
    });

    expect(input).toEqual(expect.objectContaining({
      manufacturer: "Tillig",
      articleNumber: "83101",
      gauge: "TT",
      fields: expect.objectContaining({ ean: "4012500831012" })
    }));
    expect(input.fields).not.toHaveProperty("articleType");
    expect(input.fields).not.toHaveProperty("subtype");
  });

  it("accepts EAN alone or manufacturer, gauge, and article identity", () => {
    expect(hasAccessorySearchCriteria({ fields: { ean: "4012500831012" } })).toBe(true);
    expect(hasAccessorySearchCriteria({ manufacturer: "Tillig", gauge: "TT", name: "Gleis" })).toBe(true);
    expect(hasAccessorySearchCriteria({ manufacturer: "Tillig", name: "Gleis" })).toBe(false);
  });

  it("applies selected compatible values while preserving article type and subtype", () => {
    const selectedFields = Object.fromEntries(Object.keys(result.fields).map((key) => [
      articleSelectionKey(result, key, 0), true
    ]));
    const patch = applyAccessorySearchResult({
      form: { ...emptyArticleEditorForm(), articleType: "track", subtype: "straight" },
      result,
      resultIndex: 0,
      selectedFields,
      manufacturers: [manufacturer],
      gauges: [gauge]
    });

    expect(patch).toMatchObject({
      manufacturer: "Tillig",
      articleNumber: "83101",
      name: "Gerades Modellgleis",
      ean: "4012500831012",
      gauges: ["TT"],
      scale: "1:120",
      description: "Bettungsgleis",
      productUrl: "https://www.tillig.com/83101.html"
    });
    expect(patch).not.toHaveProperty("articleType");
    expect(patch).not.toHaveProperty("subtype");
  });

  it("ignores unknown master data, unselected fields, and unsuitable descriptions", () => {
    const unsafeResult: ArticleSearchResult = {
      ...result,
      fields: {
        ...result.fields,
        manufacturer: { label: "Hersteller", value: "Unbekannt", confidence: 1 },
        gauge: { label: "Spurweite", value: "Z", confidence: 1 },
        description: { label: "Beschreibung", value: "Cookie Einstellungen", confidence: 1 }
      }
    };
    const selectedFields = Object.fromEntries(Object.keys(unsafeResult.fields).map((key) => [
      articleSelectionKey(unsafeResult, key, 0), key !== "name"
    ]));

    const patch = applyAccessorySearchResult({
      form: emptyArticleEditorForm(),
      result: unsafeResult,
      resultIndex: 0,
      selectedFields,
      manufacturers: [manufacturer],
      gauges: [gauge]
    });

    expect(patch).not.toHaveProperty("manufacturer");
    expect(patch).not.toHaveProperty("gauges");
    expect(patch).not.toHaveProperty("description");
    expect(patch).not.toHaveProperty("name");
    expect(isUsableAccessorySearchValue("description", "Cookie Einstellungen")).toBe(false);
  });

  it("formats current accessory values for conflict comparison", () => {
    const form = { ...emptyArticleEditorForm(), gauges: ["TT", "H0"], productUrl: result.url };
    expect(currentAccessorySearchValue(form, "gauge")).toBe("TT, H0");
    expect(currentAccessorySearchValue(form, "articleSourceUrl")).toBe(result.url);
  });
});
