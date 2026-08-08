import { describe, expect, it } from "vitest";

import { readSettingsLocation, settingsLocationSearch } from "./settingsDataModel";

describe("settings data navigation model", () => {
  it("uses manufacturers as the default data location", () => {
    expect(readSettingsLocation("?tab=data")).toEqual({
      tab: "data",
      group: "general",
      type: "manufacturer"
    });
  });

  it("maps the removed article-management route to article stock units", () => {
    expect(readSettingsLocation("?tab=articleManagement")).toEqual({
      tab: "data",
      group: "article",
      type: "stock_unit"
    });
  });

  it("keeps a valid article data location", () => {
    expect(readSettingsLocation("?tab=data&group=article&type=locations")).toEqual({
      tab: "data",
      group: "article",
      type: "locations"
    });
  });

  it("falls back when a type does not belong to its group", () => {
    expect(readSettingsLocation("?tab=data&group=general&type=locations").type).toBe("manufacturer");
  });

  it("writes stable data parameters and omits the general tab parameter", () => {
    expect(
      settingsLocationSearch({ tab: "data", group: "article", type: "customFields" })
    ).toBe("?tab=data&group=article&type=customFields");
    expect(
      settingsLocationSearch({ tab: "general", group: "general", type: "manufacturer" })
    ).toBe("");
  });
});
