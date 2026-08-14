import { beforeEach, describe, expect, it } from "vitest";

import {
  articleViewSettingKey,
  persistArticleViewMode,
  storedArticleViewMode
} from "./articleViewMode";

describe("article view mode", () => {
  beforeEach(() => window.localStorage.clear());

  it("defaults missing and unknown values to table", () => {
    expect(storedArticleViewMode()).toBe("table");
    window.localStorage.setItem(articleViewSettingKey, "unknown");
    expect(storedArticleViewMode()).toBe("table");
  });

  it("persists and restores cards", () => {
    persistArticleViewMode("cards");
    expect(window.localStorage.getItem(articleViewSettingKey)).toBe("cards");
    expect(storedArticleViewMode()).toBe("cards");
  });
});
