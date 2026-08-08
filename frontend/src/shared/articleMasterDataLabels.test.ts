import { describe, expect, it } from "vitest";

import type { MasterDataEntry } from "./api";
import { masterDataDisplayLabel, masterDataPersistedLabel } from "./articleMasterDataLabels";
import { translate } from "./i18n";

const entry = (type: string, key: string, label: string): MasterDataEntry => ({
  id: `${type}-${key}`,
  type,
  key,
  label,
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: "2026-08-08T08:00:00Z",
  updatedAt: "2026-08-08T08:00:00Z"
});

const de = (key: string, values?: Record<string, string | number>) => translate("de", key, values);
const en = (key: string, values?: Record<string, string | number>) => translate("en", key, values);

describe("article master-data labels", () => {
  it("localizes unchanged standard units, article types and subtypes", () => {
    expect(masterDataDisplayLabel(entry("stock_unit", "piece", "Piece"), de)).toBe("Stück");
    expect(masterDataDisplayLabel(entry("stock_unit", "piece", "Piece"), en)).toBe("Piece");
    expect(masterDataDisplayLabel(
      entry("article_type", "electrical_control", "Electrical control"), de
    )).toBe("Elektrik & Steuerung");
    expect(masterDataDisplayLabel(
      entry("accessory_subtype", "track:straight", "Straight"), de
    )).toBe("Gerade");
  });

  it("preserves manufacturers, custom fields, custom subtypes and renamed standards", () => {
    expect(masterDataDisplayLabel(entry("manufacturer", "tillig", "Tillig"), de)).toBe("Tillig");
    expect(masterDataDisplayLabel(
      entry("accessory_custom_field", "material", "Materialqualität"), en
    )).toBe("Materialqualität");
    expect(masterDataDisplayLabel(
      entry("accessory_subtype", "track:club_profile", "Vereinsprofil"), en
    )).toBe("Vereinsprofil");
    expect(masterDataDisplayLabel(entry("article_type", "track", "Gleismaterial"), en))
      .toBe("Gleismaterial");
  });

  it("keeps the canonical stored label when a localized edit remains unchanged", () => {
    const standard = entry("article_type", "track", "Track");

    expect(masterDataPersistedLabel(standard, "Gleis", de)).toBe("Track");
    expect(masterDataPersistedLabel(standard, "Gleismaterial", de)).toBe("Gleismaterial");
  });
});
