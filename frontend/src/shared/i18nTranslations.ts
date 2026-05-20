import type { Language } from "./i18n";
import { deTranslations } from "./i18n/de";
import { enTranslations } from "./i18n/en";

export const translations: Record<Language, Record<string, string>> = {
  de: deTranslations,
  en: enTranslations
};
