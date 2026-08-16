import { describe, expect, it } from "vitest";

import type { AccessoryArticle } from "../../shared/api";
import {
  articleEditorWriteInput,
  articleToEditorForm,
  emptyArticleEditorForm,
  validateArticleEditorForm
} from "./articleEditorModel";

const archivedArticle: AccessoryArticle = {
  id: "archived-1",
  inventoryNumber: "RK-ART-000001",
  manufacturer: "Tillig",
  articleNumber: "83101",
  name: "Archiviertes Gleis",
  category: "straight",
  trackingMode: "quantity",
  manufacturerStatus: "available",
  articleType: "track",
  subtype: "straight",
  gauges: ["TT"],
  listPrice: "129.90",
  packageQuantity: 1,
  stockUnit: "piece",
  minimumStock: 0,
  inventoryStrategy: "quantity",
  alternativeNumbers: [],
  keywords: [],
  archived: true,
  attributes: [],
  createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T09:00:00Z"
};

describe("articleEditorModel", () => {
  it("roundtrips and exactly normalizes compatible list-price input", () => {
    expect(articleToEditorForm(archivedArticle).listPrice).toBe("129.90");
    expect(articleEditorWriteInput({
      ...emptyArticleEditorForm(),
      listPrice: "1.299,90"
    }).listPrice).toBe("1299.90");
    expect(articleEditorWriteInput({
      ...emptyArticleEditorForm(),
      listPrice: "0"
    }).listPrice).toBe("0.00");
  });

  it("rejects negative and malformed list-price drafts without changing them", () => {
    for (const listPrice of ["-1", "1.2345", "abc"]) {
      const form = { ...emptyArticleEditorForm(), listPrice };
      const validation = validateArticleEditorForm(form);

      expect(validation.fieldErrors.listPrice).toBeDefined();
      expect(validation.tabErrors.article).toBe(true);
      expect(() => articleEditorWriteInput(form)).toThrow("invalid list price");
      expect(form.listPrice).toBe(listPrice);
    }
  });

  it("keeps archived articles archived through edit mapping while create starts active", () => {
    expect(articleEditorWriteInput(articleToEditorForm(archivedArticle)).archived).toBe(true);
    expect(articleEditorWriteInput(emptyArticleEditorForm()).archived).toBe(false);
  });

  it("attaches invalid numeric subject drafts to the subject tab without losing the draft", () => {
    const form = {
      ...emptyArticleEditorForm(),
      articleType: "track" as const,
      attributeNumberDrafts: { lengthMm: "." }
    };

    const validation = validateArticleEditorForm(form, {
      required: "Pflichtfeld",
      positive: "Positiv",
      nonnegative: "Nicht negativ",
      integer: "Ganzzahlig",
      invalidSubject: "Fachwert ungültig",
      invalidOption: "Auswahl ungültig",
      invalidStep: "Schrittweite ungültig",
      invalidMoney: "Preis ungültig"
    });

    expect(validation.fieldErrors.attributes).toBe("Fachwert ungültig");
    expect(validation.tabErrors.subject).toBe(true);
    expect(form.attributeNumberDrafts.lengthMm).toBe(".");
  });

  it("rejects invalid subject values before serialization and serializes valid decimal steps", () => {
    const invalid = {
      ...emptyArticleEditorForm(),
      manufacturer: "Tillig",
      name: "Gleis",
      articleType: "track" as const,
      subtype: "straight",
      attributeNumberDrafts: { connectionCount: "1.5" }
    };
    expect(() => articleEditorWriteInput(invalid)).toThrow("invalid subject values");

    const valid = {
      ...invalid,
      attributeNumberDrafts: { angleDegrees: "0,3", connectionCount: "2" }
    };
    expect(articleEditorWriteInput(valid).attributes).toEqual([
      { key: "angleDegrees", kind: "number", numberValue: 0.3, unit: "°" },
      { key: "connectionCount", kind: "number", numberValue: 2 }
    ]);
  });

  it("rejects decimal package and minimum quantities without changing their raw drafts", () => {
    const form = {
      ...emptyArticleEditorForm(),
      packageQuantity: "1.5",
      minimumStock: "2,5"
    };

    const validation = validateArticleEditorForm(form, {
      required: "Required",
      positive: "Positive",
      nonnegative: "Nonnegative",
      integer: "Enter a whole number",
      invalidSubject: "Invalid subject",
      invalidOption: "Invalid option",
      invalidStep: "Invalid step",
      invalidMoney: "Invalid money"
    });

    expect(() => articleEditorWriteInput(form)).toThrow("invalid article quantities");
    expect(validation.fieldErrors.packageQuantity).toBe("Enter a whole number");
    expect(validation.fieldErrors.minimumStock).toBe("Enter a whole number");
    expect(validation.tabErrors).toEqual({ article: true, stock: true });
    expect(form.packageQuantity).toBe("1.5");
    expect(form.minimumStock).toBe("2,5");
  });

  it("roundtrips only historical inactive custom attributes owned by the loaded edit snapshot", () => {
    const historical = { key: "legacyMaterial", kind: "text" as const, textValue: "Holz" };
    const form = {
      ...emptyArticleEditorForm(), articleType: "other" as const, attributes: [historical]
    };

    expect(articleEditorWriteInput(form, [], [historical]).attributes).toEqual([historical]);
    expect(validateArticleEditorForm(form, undefined, [], [historical]).fieldErrors.attributes).toBeUndefined();
    expect(() => articleEditorWriteInput(form)).toThrow("invalid subject values");
    expect(() => articleEditorWriteInput({
      ...form, attributes: [{ ...historical, textValue: "Manipuliert" }]
    }, [], [historical])).toThrow("invalid subject values");
  });

  it("consumes each historical inactive attribute match only once", () => {
    const historical = { key: "legacyMaterial", kind: "text" as const, textValue: "Holz" };
    const duplicated = {
      ...emptyArticleEditorForm(), articleType: "other" as const,
      attributes: [historical, { ...historical }]
    };

    expect(validateArticleEditorForm(duplicated, undefined, [], [historical]).fieldErrors.attributes)
      .toBeDefined();
    expect(() => articleEditorWriteInput(duplicated, [], [historical])).toThrow("invalid subject values");
  });
});
