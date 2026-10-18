import type { Language } from "./i18n";
import { deTranslations } from "./i18n/de";
import { enTranslations } from "./i18n/en";

import { decoderDETranslations } from "./i18n/decoder.de";
import { decoderENTranslations } from "./i18n/decoder.en";

export const translations: Record<Language, Record<string, string>> = {
  de: { ...deTranslations, ...decoderDETranslations },
  en: { ...enTranslations, ...decoderENTranslations }
};
