import { vehicleCatalogDE } from "./i18n/vehicle-catalog.de";
import { vehicleCatalogEN } from "./i18n/vehicle-catalog.en";
import { vehicleMediaDE } from "./i18n/vehicle-media.de";
import { importReviewDE } from "./i18n/import-review.de";
import { importReviewEN } from "./i18n/import-review.en";
import { vehicleMediaEN } from "./i18n/vehicle-media.en";
import { ecosFunctionsDE } from "./i18n/ecos-functions.de";
import { ecosFunctionsEN } from "./i18n/ecos-functions.en";
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
  de: { ...vehicleCatalogDE, ...importReviewDE, ...vehicleMediaDE, ...ecosFunctionsDE, ...locomotiveImageDETranslations, ...deTranslations, ...decoderDETranslations, ...libraryDETranslations, ...railcomDETranslations, ...kuehnDETranslations, ...uhlenbrockDETranslations, ...rocoDETranslations, ...pikoDETranslations },
  en: { ...vehicleCatalogEN, ...importReviewEN, ...vehicleMediaEN, ...ecosFunctionsEN, ...locomotiveImageENTranslations, ...enTranslations, ...decoderENTranslations, ...libraryENTranslations, ...railcomENTranslations, ...kuehnENTranslations, ...uhlenbrockENTranslations, ...rocoENTranslations, ...pikoENTranslations }
};
