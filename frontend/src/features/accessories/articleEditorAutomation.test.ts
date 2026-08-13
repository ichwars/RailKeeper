import { describe, expect, it } from "vitest";

import type { MasterDataEntry } from "../../shared/api";
import { emptyArticleEditorForm } from "./articleEditorModel";
import { scaleForGauges, suggestedArticleKeywords } from "./articleEditorAutomation";

const timestamp = "2026-08-13T10:00:00Z";
const gaugeEntries: MasterDataEntry[] = [{
  id: "gauge:tt",
  type: "gauge",
  key: "tt",
  label: "TT",
  active: true,
  sortOrder: 10,
  metadata: { scale: "1:120" },
  createdAt: timestamp,
  updatedAt: timestamp
}, {
  id: "gauge:h0",
  type: "gauge",
  key: "h0",
  label: "H0",
  active: true,
  sortOrder: 20,
  metadata: {},
  createdAt: timestamp,
  updatedAt: timestamp
}];

describe("article editor automation", () => {
  it("resolves scale metadata from the first active selected gauge", () => {
    expect(scaleForGauges(["TT"], gaugeEntries)).toBe("1:120");
    expect(scaleForGauges(["tt"], gaugeEntries)).toBe("1:120");
    expect(scaleForGauges(["H0"], gaugeEntries)).toBe("");
    expect(scaleForGauges(["unbekannt"], gaugeEntries)).toBe("");
  });

  it("builds trimmed, case-insensitively deduplicated keyword suggestions", () => {
    const form = {
      ...emptyArticleEditorForm(),
      name: "  Einfache Weiche EW1 links ",
      manufacturer: " Tillig ",
      articleType: "track" as const,
      subtype: "turnout"
    };

    expect(suggestedArticleKeywords(form, "Gleis", "Weiche"))
      .toBe("Einfache Weiche EW1 links, Tillig, Gleis, Weiche");
    expect(suggestedArticleKeywords({ ...form, manufacturer: "gleis" }, "Gleis", "Weiche"))
      .toBe("Einfache Weiche EW1 links, gleis, Weiche");
  });

  it("does not generate keywords before an article has identity data", () => {
    expect(suggestedArticleKeywords(emptyArticleEditorForm(), "Sonstiges", "")).toBe("");
  });
});
