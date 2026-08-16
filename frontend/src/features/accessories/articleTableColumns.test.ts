import { beforeEach, describe, expect, it } from "vitest";

import {
  articleTableColumns,
  articleTableColumnSettingKey,
  persistArticleTableColumns,
  resetArticleTableColumns,
  storedArticleTableColumns,
  toggleArticleTableColumn,
  defaultArticleTableColumns,
  type ArticleTableColumn
} from "./articleTableColumns";

describe("article table columns", () => {
  beforeEach(() => window.localStorage.clear());

  it("keeps newly introduced columns hidden for missing, malformed, and older preferences", () => {
    expect(articleTableColumns).toContain("listPrice");
    expect(defaultArticleTableColumns.has("listPrice")).toBe(false);
    expect(storedArticleTableColumns().has("listPrice")).toBe(false);

    window.localStorage.setItem(articleTableColumnSettingKey, "not-json");
    expect(storedArticleTableColumns().has("listPrice")).toBe(false);

    window.localStorage.setItem(articleTableColumnSettingKey, JSON.stringify([
      "image", "inventoryNumber", "manufacturer", "articleNumber", "name", "type", "gauge", "stock", "storage"
    ]));
    expect(storedArticleTableColumns().has("listPrice")).toBe(false);
  });

  it("filters stale values and restores an identity column", () => {
    window.localStorage.setItem(
      articleTableColumnSettingKey,
      JSON.stringify(["image", "unknown"])
    );

    expect([...storedArticleTableColumns()]).toEqual(["image", "inventoryNumber"]);
  });

  it("does not hide the final visible identity column", () => {
    const current = new Set<ArticleTableColumn>(["name", "stock"]);

    expect([...toggleArticleTableColumn(current, "name")]).toEqual(["name", "stock"]);
  });

  it("allows either identity column to be hidden when the other remains visible", () => {
    const current = new Set<ArticleTableColumn>(["inventoryNumber", "name", "stock"]);

    expect([...toggleArticleTableColumn(current, "inventoryNumber")]).toEqual(["name", "stock"]);
    expect([...toggleArticleTableColumn(current, "name")]).toEqual(["inventoryNumber", "stock"]);
  });

  it("persists columns in stable table order", () => {
    persistArticleTableColumns(new Set<ArticleTableColumn>(["storage", "name"]));

    expect(window.localStorage.getItem(articleTableColumnSettingKey)).toBe('["name","storage"]');
  });

  it("restores the standard columns without the optional list price", () => {
    expect(resetArticleTableColumns()).toEqual(defaultArticleTableColumns);
    expect(resetArticleTableColumns().has("listPrice")).toBe(false);
  });
});
