import { describe, expect, it } from "vitest";

import type { MasterDataEntry } from "../../shared/api";
import { translate } from "../../shared/i18n";
import { articleTypeLabel, articleTypeOptions } from "./articleTypes";

const entry = (key: string, label: string, active = true): MasterDataEntry => ({
  id: `article-type-${key}`,
  type: "article_type",
  key,
  label,
  active,
  sortOrder: 10,
  metadata: {},
  createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
});

describe("article type labels", () => {
  it("localizes unchanged seed labels and preserves configured renames", () => {
    const entries = [entry("track", "Track"), entry("signal", "Formsignal")];

    expect(articleTypeLabel("track", entries, (key) => translate("de", key))).toBe("Gleis");
    expect(articleTypeLabel("track", entries, (key) => translate("en", key))).toBe("Track");
    expect(articleTypeLabel("signal", entries, (key) => translate("de", key))).toBe("Formsignal");
  });

  it("offers active types and only the unchanged inactive historical type", () => {
    const entries = [entry("track", "Track"), entry("signal", "Signal", false), entry("other", "Other")];
    const t = (key: string) => translate("de", key);

    expect(articleTypeOptions(entries, null, t).map((option) => option.value)).toEqual(["track", "other"]);
    expect(articleTypeOptions(entries, "signal", t)).toEqual([
      { value: "track", label: "Gleis", active: true },
      { value: "signal", label: "Signal", active: false },
      { value: "other", label: "Sonstiger Artikel", active: true }
    ]);
  });
});
