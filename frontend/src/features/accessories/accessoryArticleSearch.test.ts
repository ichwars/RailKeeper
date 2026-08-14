import { describe, expect, it } from "vitest";

import type { ArticleSearchResult, MasterDataEntry } from "../../shared/api";
import { articleSelectionKey } from "../../shared/articleSearch/articleSearchModel";
import { emptyArticleEditorForm } from "./articleEditorModel";
import {
  accessorySearchInput,
  accessorySearchFieldGroups,
  applyAccessorySearchResult,
  currentAccessorySearchValue,
  hasAccessorySearchCriteria,
  isSelectableAccessorySearchValue,
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
  it("maps article criteria and current typed subject values as search context", () => {
    const input = accessorySearchInput({
      ...emptyArticleEditorForm(),
      manufacturer: "Tillig",
      articleNumber: "83101",
      gauges: ["TT"],
      ean: "4012500831012",
      articleType: "track",
      subtype: "turnout",
      attributes: [
        { key: "trackSystem", kind: "text", textValue: "TT Modellgleis" },
        { key: "roadbed", kind: "boolean", booleanValue: false }
      ],
      attributeNumberDrafts: { lengthMm: "129,5" }
    });

    expect(input).toEqual(expect.objectContaining({
      manufacturer: "Tillig",
      articleNumber: "83101",
      gauge: "TT",
      fields: expect.objectContaining({
        ean: "4012500831012",
        articleType: "track",
        subtype: "turnout",
        trackSystem: "TT Modellgleis",
        roadbed: "false",
        lengthMm: "129.5"
      })
    }));
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
    const form = {
      ...emptyArticleEditorForm(),
      articleType: "track" as const,
      gauges: ["TT", "H0"],
      productUrl: result.url,
      attributes: [
        { key: "trackSystem", kind: "text" as const, textValue: "TT Modellgleis" },
        { key: "roadbed", kind: "boolean" as const, booleanValue: false }
      ],
      attributeNumberDrafts: { lengthMm: "129,5" }
    };
    expect(currentAccessorySearchValue(form, "gauge")).toBe("TT, H0");
    expect(currentAccessorySearchValue(form, "articleSourceUrl")).toBe(result.url);
    expect(currentAccessorySearchValue(form, "trackSystem")).toBe("TT Modellgleis");
    expect(currentAccessorySearchValue(form, "roadbed")).toBe("false");
    expect(currentAccessorySearchValue(form, "lengthMm")).toBe("129.5");
  });

  it("shows and validates the track subject field group", () => {
    const t = (key: string, values?: Record<string, string | number>) =>
      key === "accessories.editor.tabs.subject" ? `Fachangaben: ${values?.type}` : key;
    const groups = accessorySearchFieldGroups(t, "track");

    expect(groups.at(-1)).toMatchObject({
      key: "subject",
      label: "Fachangaben: accessories.articleType.track"
    });
    expect(groups.at(-1)?.fields.map((field) => field.key)).toEqual(expect.arrayContaining([
      "trackSystem", "lengthMm", "radiusMm", "direction", "roadbed", "digitalReady"
    ]));
    expect(isSelectableAccessorySearchValue("direction", "left", [manufacturer], [gauge], "track"))
      .toBe(true);
    expect(isSelectableAccessorySearchValue("direction", "up", [manufacturer], [gauge], "track"))
      .toBe(false);
    expect(isSelectableAccessorySearchValue("lengthMm", "129,5", [manufacturer], [gauge], "track"))
      .toBe(true);
    expect(isSelectableAccessorySearchValue("lengthMm", "-2", [manufacturer], [gauge], "track"))
      .toBe(false);
  });

  it("applies selected track subject fields as typed values", () => {
    const trackResult: ArticleSearchResult = {
      ...result,
      fields: {
        trackSystem: { label: "Gleissystem", value: "TT Modellgleis", confidence: 1 },
        lengthMm: { label: "Länge", value: "129,5", confidence: 1 },
        direction: { label: "Richtung", value: "left", confidence: 1 },
        roadbed: { label: "Bettung", value: "false", confidence: 1 },
        digitalReady: { label: "Digitaltauglich", value: "true", confidence: 1 }
      }
    };
    const selectedFields = Object.fromEntries(Object.keys(trackResult.fields).map((key) => [
      articleSelectionKey(trackResult, key, 0), true
    ]));
    const patch = applyAccessorySearchResult({
      form: {
        ...emptyArticleEditorForm(),
        articleType: "track",
        attributes: [{ key: "trackSystem", kind: "text", textValue: "Alt" }]
      },
      result: trackResult,
      resultIndex: 0,
      selectedFields,
      manufacturers: [manufacturer],
      gauges: [gauge]
    });

    expect(patch.attributeNumberDrafts).toEqual({ lengthMm: "129.5" });
    expect(patch.attributes).toEqual(expect.arrayContaining([
      { key: "trackSystem", kind: "text", textValue: "TT Modellgleis" },
      { key: "direction", kind: "single_select", optionValues: ["left"] },
      { key: "roadbed", kind: "boolean", booleanValue: false },
      { key: "digitalReady", kind: "boolean", booleanValue: true }
    ]));
    expect(patch.attributes?.filter((attribute) => attribute.key === "trackSystem")).toHaveLength(1);
  });
});
