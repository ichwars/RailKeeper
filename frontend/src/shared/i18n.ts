import { useEffect, useState } from "react";
import { translations } from "./i18nTranslations";

export type Language = "de" | "en";

export const languageSettingKey = "railkeeper.settings.language";
export const languageChangedEvent = "railkeeper-language-changed";


export function readLanguage(): Language {
  const stored = window.localStorage.getItem(languageSettingKey);
  return stored === "en" || stored === "de" ? stored : "de";
}

export function setLanguage(language: Language) {
  window.localStorage.setItem(languageSettingKey, language);
  document.documentElement.lang = language;
  window.dispatchEvent(new CustomEvent(languageChangedEvent, { detail: language }));
}

export function translate(language: Language, key: string, values: Record<string, string | number> = {}) {
  const template = translations[language][key] || translations.de[key] || key;
  return Object.entries(values).reduce((text, [name, value]) => text.replaceAll(`{${name}}`, String(value)), template);
}

const locale = (language: Language) => language === "de" ? "de-DE" : "en-GB";

export function formatDate(value: string | Date, language: Language): string {
  return new Intl.DateTimeFormat(locale(language), { dateStyle: "short" }).format(new Date(value));
}

export function formatDateTime(value: string | Date, language: Language): string {
  return new Intl.DateTimeFormat(locale(language), { dateStyle: "short", timeStyle: "short" }).format(new Date(value));
}

export function useI18n() {
  const [language, setCurrentLanguage] = useState<Language>(() => readLanguage());

  useEffect(() => {
    const syncLanguage = () => setCurrentLanguage(readLanguage());
    window.addEventListener(languageChangedEvent, syncLanguage);
    window.addEventListener("storage", syncLanguage);
    return () => {
      window.removeEventListener(languageChangedEvent, syncLanguage);
      window.removeEventListener("storage", syncLanguage);
    };
  }, []);

  return {
    language,
    setLanguage,
    t: (key: string, values?: Record<string, string | number>) => translate(language, key, values)
  };
}
