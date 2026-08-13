import type { CreateVehicleRequest } from "../../shared/api";
import type {
  ArticleSearchFieldGroup,
  Translate
} from "../../shared/articleSearch/articleSearchModel";
export {
  articleFieldStatus,
  articleResultKey,
  articleSelectionKey,
  imageSelectionKey,
  sourceDisplayName,
  sourceShortLink
} from "../../shared/articleSearch/articleSearchModel";

export type ArticleFieldKey = keyof CreateVehicleRequest;

export const articleFieldLabels: Partial<Record<ArticleFieldKey, string>> = {
  manufacturer: "Hersteller",
  articleNumber: "Artikel-Nr.",
  articleSourceUrl: "Quelle",
  name: "Bezeichnung",
  gauge: "Spurweite",
  epoch: "Epoche",
  railwayCompany: "Bahngesellschaft",
  category: "Kategorie",
  gattung: "Gattung",
  description: "Beschreibung",
  series: "Baureihe",
  vehicleNumber: "Fahrzeug-Nr.",
  digitalDecoderNumber: "Digital / Decoder-Nr.",
  dtDecoderNumber: "DT / Decoder-Nr.",
  decoderType: "Decoder-Typ",
  exhibitionReady: "Messe tauglich",
  exhibition: "Ausstellung",
  abcBrakes: "ABC Bremsen",
  ean: "EAN-Nr.",
  productionPeriod: "Produktionszeit",
  listPrice: "Listenpreis",
  lengthMm: "Länge (mm)",
  weightG: "Gewicht (g)",
  color: "Farbe",
  lettering: "Beschriftung",
  load: "Beladung",
  interior: "Inneneinrichtung",
  axles: "Achsen",
  axleCount: "Anzahl",
  tractionTireCount: "Haftreifen",
  wheelset: "Radsatz",
  couplingFront: "Kupplung vorne",
  couplingRear: "Kupplung hinten",
  powerPickup: "Stromabnahme",
  adapter: "Adapter",
  digital: "Digital",
  soundGeneratorEnabled: "Soundgenerator",
  headlightsEnabled: "Fahrlicht",
  lightingEnabled: "Beleuchtung",
  driveDescription: "Antrieb Beschreibung",
  headlightsDescription: "Fahrlicht Beschreibung",
  lightingDescription: "Beleuchtung Beschreibung",
  soundGeneratorDescription: "Soundgenerator Beschreibung",
  smokeGeneratorDescription: "Rauchgenerator Beschreibung",
  additionalInfo: "Zusatzinformationen"
};

export const articleFieldGroups: { title: string; keys: ArticleFieldKey[] }[] = [
  {
    title: "Modell",
    keys: ["name", "articleNumber", "manufacturer", "gauge", "ean", "railwayCompany", "epoch", "series", "vehicleNumber", "gattung", "category"]
  },
  {
    title: "Masse / Bauart",
    keys: ["lengthMm", "weightG", "color", "lettering", "load", "interior", "axles", "axleCount", "tractionTireCount"]
  },
  {
    title: "Technik",
    keys: ["adapter", "powerPickup", "digital", "digitalDecoderNumber", "dtDecoderNumber", "soundGeneratorEnabled", "headlightsEnabled", "lightingEnabled", "driveDescription", "headlightsDescription", "lightingDescription", "soundGeneratorDescription", "smokeGeneratorDescription"]
  },
  {
    title: "Weitere Daten",
    keys: ["description", "additionalInfo", "productionPeriod", "listPrice", "articleSourceUrl"]
  }
];

export function vehicleArticleSearchGroups(t: Translate): ArticleSearchFieldGroup[] {
  const groupLabels: Record<string, string> = {
    "Modell": t("vehicles.articleSearch.group.model"),
    "Masse / Bauart": t("vehicles.articleSearch.group.mass"),
    "Technik": t("vehicles.articleSearch.group.technology"),
    "Weitere Daten": t("vehicles.articleSearch.group.more")
  };
  return articleFieldGroups.map((group) => ({
    key: group.title,
    label: groupLabels[group.title] || group.title,
    fields: group.keys.map((key) => {
      const translated = t(`vehicle.field.${key}`);
      return {
        key,
        label: translated === `vehicle.field.${key}` ? articleFieldLabels[key] || key : translated
      };
    })
  }));
}

const booleanArticleFields = new Set<ArticleFieldKey>([
  "digital",
  "dtDecoder",
  "exhibitionReady",
  "exhibition",
  "abcBrakes",
  "driveEnabled",
  "headlightsEnabled",
  "lightingEnabled",
  "soundGeneratorEnabled",
  "smokeGeneratorEnabled",
  "qrCodeEnabled"
]);

export function isArticleFieldKey(key: string): key is ArticleFieldKey {
  return key in articleFieldLabels;
}

export function booleanFromArticleValue(value: string) {
  return ["ja", "true", "1", "yes", "vorhanden", "digital"].includes(value.trim().toLocaleLowerCase("de-DE"));
}

export function articleValueForForm(key: ArticleFieldKey, value: string) {
  if (booleanArticleFields.has(key)) {
    return booleanFromArticleValue(value);
  }
  return value;
}

export function currentArticleValue(form: CreateVehicleRequest, key: ArticleFieldKey) {
  const value = form[key];
  if (typeof value === "boolean") {
    return value ? "Ja" : "Nein";
  }
  return String(value || "").trim();
}
