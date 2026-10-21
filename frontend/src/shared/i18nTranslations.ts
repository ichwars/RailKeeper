import { pikoDETranslations } from "./i18n/piko.de";
import { pikoENTranslations } from "./i18n/piko.en";
import type { Language } from "./i18n";
import { deTranslations } from "./i18n/de";
import { enTranslations } from "./i18n/en";
import { locomotiveImageDETranslations } from "./i18n/locomotive-image.de";
import { locomotiveImageENTranslations } from "./i18n/locomotive-image.en";

import { decoderDETranslations } from "./i18n/decoder.de";
import { decoderENTranslations } from "./i18n/decoder.en";

import { libraryDETranslations } from "./i18n/library.de";
import { libraryENTranslations } from "./i18n/library.en";

import { railcomDETranslations } from "./i18n/railcom.de";
import { railcomENTranslations } from "./i18n/railcom.en";
import { kuehnDETranslations } from "./i18n/kuehn.de";
import { kuehnENTranslations } from "./i18n/kuehn.en";

import { uhlenbrockDETranslations } from "./i18n/uhlenbrock.de";
import { uhlenbrockENTranslations } from "./i18n/uhlenbrock.en";

import { rocoDETranslations } from "./i18n/roco.de";
import { rocoENTranslations } from "./i18n/roco.en";

export const translations: Record<Language, Record<string, string>> = {
  de: { ...locomotiveImageDETranslations, ...deTranslations, ...decoderDETranslations, ...libraryDETranslations, ...railcomDETranslations, ...kuehnDETranslations, ...uhlenbrockDETranslations, ...rocoDETranslations, ...pikoDETranslations },
  en: { ...locomotiveImageENTranslations, ...enTranslations, ...decoderENTranslations, ...libraryENTranslations, ...railcomENTranslations, ...kuehnENTranslations, ...uhlenbrockENTranslations, ...rocoENTranslations, ...pikoENTranslations }
};
