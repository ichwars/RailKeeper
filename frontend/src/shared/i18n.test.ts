import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  languageChangedEvent,
  languageSettingKey,
  readLanguage,
  setLanguage,
  translate
} from "./i18n";

describe("i18n", () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.lang = "";
  });

  it("uses German when no language is stored", () => {
    expect(readLanguage()).toBe("de");
  });

  it("stores English and announces the change", () => {
    const listener = vi.fn();
    window.addEventListener(languageChangedEvent, listener);

    setLanguage("en");

    expect(window.localStorage.getItem(languageSettingKey)).toBe("en");
    expect(document.documentElement.lang).toBe("en");
    expect(listener).toHaveBeenCalledOnce();
    window.removeEventListener(languageChangedEvent, listener);
  });

  it("interpolates translated values", () => {
    expect(translate("de", "common.roles", { count: 3 })).toBe("3 Rollen");
  });

  it("returns the key when no translation exists", () => {
    expect(translate("en", "missing.translation.key")).toBe("missing.translation.key");
  });
});
