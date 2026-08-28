import { beforeEach, describe, expect, it } from "vitest";

import {
  articleTableColumns,
  articleTableColumnSettingKey,
  articleTableColumnWidthDefinitions,
  defaultArticleTableColumns,
  defaultArticleTableLayout,
  moveArticleTableColumn,
  parseArticleTableLayout,
  persistArticleTableColumns,
  resetArticleTableColumns,
  serializeArticleTableLayout,
  storedArticleTableColumns,
  toggleArticleTableColumn,
  type ArticleTableColumn
} from "./articleTableColumns";

describe("article table columns", () => {
  beforeEach(() => window.localStorage.clear());

  it("keeps newly introduced columns hidden for missing, malformed, and older preferences", () => {
    expect(articleTableColumns).toContain("listPrice");
    expect(defaultArticleTableColumns).not.toContain("listPrice");
    expect(storedArticleTableColumns()).not.toContain("listPrice");

    window.localStorage.setItem(articleTableColumnSettingKey, "not-json");
    expect(storedArticleTableColumns()).not.toContain("listPrice");

    window.localStorage.setItem(articleTableColumnSettingKey, JSON.stringify([
      "image", "inventoryNumber", "manufacturer", "articleNumber", "name", "type", "gauge",
      "stock", "storage"
    ]));
    expect(storedArticleTableColumns()).not.toContain("listPrice");
  });

  it("filters stale values, duplicates, and restores an identity column", () => {
    window.localStorage.setItem(
      articleTableColumnSettingKey,
      JSON.stringify(["image", "unknown", "image"])
    );

    expect(storedArticleTableColumns()).toEqual(["image", "inventoryNumber"]);
  });

  it("does not hide the final visible identity column", () => {
    const current: ArticleTableColumn[] = ["name", "stock"];

    expect(toggleArticleTableColumn(current, "name")).toEqual(["name", "stock"]);
  });

  it("allows either identity column to be hidden and preserves the remaining order", () => {
    const current: ArticleTableColumn[] = ["stock", "inventoryNumber", "name"];

    expect(toggleArticleTableColumn(current, "inventoryNumber")).toEqual(["stock", "name"]);
    expect(toggleArticleTableColumn(current, "name")).toEqual(["stock", "inventoryNumber"]);
  });

  it("moves and persists columns in the user-defined order", () => {
    const moved = moveArticleTableColumn(["inventoryNumber", "name", "storage"], "storage", "up");
    expect(moved).toEqual(["inventoryNumber", "storage", "name"]);
    persistArticleTableColumns(moved);
    expect(window.localStorage.getItem(articleTableColumnSettingKey))
      .toBe('["inventoryNumber","storage","name"]');
  });

  it("loads legacy and versioned layouts with hidden bounded widths", () => {
    expect(parseArticleTableLayout('["name","inventoryNumber"]')).toEqual({
      columns: ["name", "inventoryNumber"],
      widths: {}
    });
    expect(parseArticleTableLayout(JSON.stringify({
      version: 1,
      columns: ["inventoryNumber", "name"],
      widths: { name: 9999, storage: 202, unknown: 200 }
    }))).toEqual({
      columns: ["inventoryNumber", "name"],
      widths: {
        name: articleTableColumnWidthDefinitions.name.maxWidth,
        storage: 202
      }
    });
    expect(parseArticleTableLayout(serializeArticleTableLayout(defaultArticleTableLayout)))
      .toEqual(defaultArticleTableLayout);
  });

  it("restores the standard order without the optional list price", () => {
    expect(resetArticleTableColumns()).toEqual(defaultArticleTableColumns);
    expect(resetArticleTableColumns()).not.toContain("listPrice");
  });
});
