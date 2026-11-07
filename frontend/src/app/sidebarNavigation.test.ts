import { describe, expect, it } from "vitest";
import { applySidebarProfile, defaultSidebarOrder, normalizeSidebarPrefs, sidebarProfile } from "./sidebarPreferences";
import { sidebarNavigation } from "./sidebarNavigation";

describe("personal navigation profiles and permissions", () => {
  it("preserves a custom order and does not mutate existing preferences", () => {
    const prefs = normalizeSidebarPrefs({ order: ["vehicles", "settings", "overview"], hidden: ["accessories"] });
    const before = structuredClone(prefs);
    const collection = applySidebarProfile(prefs, "collection");
    expect(prefs).toEqual(before);
    expect(collection.order).toEqual(before.order);
    expect(sidebarProfile(prefs)).toBe("custom");
    expect(sidebarProfile(collection)).toBe("collection");
    const all = applySidebarProfile(collection, "all");
    expect(all.order).toEqual(before.order);
    expect(all.hidden).toEqual([]);
    expect(sidebarProfile(all)).toBe("all");
  });

  it("keeps every allowed tool accessible when the collection preset is selected", () => {
    const prefs = applySidebarProfile({ order: defaultSidebarOrder, hidden: [] }, "collection");
    const navigation = sidebarNavigation(["Admin"], prefs);
    expect(navigation.visible.map((item) => item.view)).toEqual(["overview", "vehicles", "accessories", "settings"]);
    expect(navigation.additional.map((item) => item.view)).toEqual(["exhibition", "importExport", "digitalCenters"]);
    expect([...navigation.visible, ...navigation.additional]).toHaveLength(defaultSidebarOrder.length);
  });

  it.each(["Editor", "Viewer"])("retains inventory permissions for %s without granting exhibition", (role) => {
    const prefs = applySidebarProfile({ order: [], hidden: [] }, "collection");
    const navigation = sidebarNavigation([role], prefs);
    expect(navigation.visible.map((item) => item.view)).toEqual(["overview", "vehicles", "accessories", "settings"]);
    expect(navigation.additional.map((item) => item.view)).toEqual(["importExport", "digitalCenters"]);
  });

  it("never exposes inventory through additional functions to Messe users", () => {
    const prefs = applySidebarProfile({ order: defaultSidebarOrder, hidden: [] }, "collection");
    const navigation = sidebarNavigation(["Messe"], prefs);
    expect(navigation.visible).toEqual([]);
    expect(navigation.additional.map((item) => item.view)).toEqual(["exhibition", "importExport"]);
    expect(sidebarNavigation([], prefs)).toEqual({ visible: [], additional: [] });
  });

  it("keeps settings visible even if malformed preferences try to hide it", () => {
    expect(sidebarNavigation(["Admin"], { order: [], hidden: ["settings"] }).visible)
      .toContainEqual(expect.objectContaining({ view: "settings" }));
  });
});
