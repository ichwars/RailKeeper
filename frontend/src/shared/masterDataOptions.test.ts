import { describe, expect, it } from "vitest";

import type { MasterDataEntry } from "./api";
import { masterDataOptions } from "./masterDataOptions";

function entry(key: string, label: string, active: boolean): MasterDataEntry {
  return {
    id: `manufacturer:${key}`,
    type: "manufacturer",
    key,
    label,
    active,
    sortOrder: 0,
    metadata: {},
    createdAt: "2026-08-16T00:00:00Z",
    updatedAt: "2026-08-16T00:00:00Z"
  };
}

describe("masterDataOptions", () => {
  it("keeps only active entries plus current inactive and legacy values", () => {
    const options = masterDataOptions(
      [entry("piko", "Piko", true), entry("roco", "Roco", false), entry("esu", "ESU", false)],
      ["Roco", "Legacy"],
      (item) => item.label
    );

    expect(options.map(({ value, active }) => ({ value, active }))).toEqual([
      { value: "Piko", active: true },
      { value: "Roco", active: false },
      { value: "Legacy", active: false }
    ]);
  });
});
