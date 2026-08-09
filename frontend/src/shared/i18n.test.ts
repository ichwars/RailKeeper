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

  it("keeps German and English translation keys in parity", async () => {
    const [{ deTranslations }, { enTranslations }] = await Promise.all([import("./i18n/de"), import("./i18n/en")]);
    expect(Object.keys(deTranslations).sort()).toEqual(Object.keys(enTranslations).sort());
  });

  it("uses individual item terminology throughout accessory translations", async () => {
    const [{ deTranslations }, { enTranslations }] = await Promise.all([import("./i18n/de"), import("./i18n/en")]);
    const accessoryValues = (translations: Record<string, string>) => Object.entries(translations)
      .filter(([key]) => key.startsWith("accessories."))
      .map(([, value]) => value)
      .join("\n");

    expect(accessoryValues(enTranslations)).not.toMatch(/individual (?:assets?|tracked devices)/i);
    expect(accessoryValues(deTranslations)).not.toMatch(/Einzelobjekt/i);
  });
});
