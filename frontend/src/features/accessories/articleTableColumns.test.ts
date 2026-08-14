import { beforeEach, describe, expect, it } from "vitest";

import {
  articleTableColumns,
  articleTableColumnSettingKey,
  persistArticleTableColumns,
  storedArticleTableColumns,
  toggleArticleTableColumn,
  type ArticleTableColumn
} from "./articleTableColumns";

describe("article table columns", () => {
  beforeEach(() => window.localStorage.clear());

  it("defaults missing or malformed preferences to every supported column", () => {
    expect([...storedArticleTableColumns()]).toEqual(articleTableColumns);

    window.localStorage.setItem(articleTableColumnSettingKey, "not-json");
    expect([...storedArticleTableColumns()]).toEqual(articleTableColumns);
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
});
