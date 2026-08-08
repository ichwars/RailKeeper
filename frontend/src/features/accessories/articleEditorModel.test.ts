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
  manufacturer: "Tillig",
  articleNumber: "83101",
  name: "Archiviertes Gleis",
  category: "straight",
  trackingMode: "quantity",
  manufacturerStatus: "available",
  articleType: "track",
  subtype: "straight",
  gauges: ["TT"],
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
      invalidSubject: "Fachwert ungültig"
    });

    expect(validation.fieldErrors.attributes).toBe("Fachwert ungültig");
    expect(validation.tabErrors.subject).toBe(true);
    expect(form.attributeNumberDrafts.lengthMm).toBe(".");
  });
});
