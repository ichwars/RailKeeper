import {
  ArticleSearchImage,
  ArticleSearchResult,
  CreateVehicleRequest
} from "../../shared/api";

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

export function articleResultKey(result: ArticleSearchResult, index = 0) {
  return `${result.url || result.title}-${index}`;
}

export function articleSelectionKey(result: ArticleSearchResult, key: string, index = 0) {
  return `${articleResultKey(result, index)}::${key}`;
}

export function imageSelectionKey(result: ArticleSearchResult, image: ArticleSearchImage, index = 0) {
  return `${articleResultKey(result, index)}::image::${image.url}`;
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

export function articleFieldStatus(current: string, found: string) {
  if (!current) return "empty";
  if (current.toLocaleLowerCase("de-DE") === found.toLocaleLowerCase("de-DE")) return "same";
  return "conflict";
}

export function sourceDisplayName(rawUrl: string) {
  try {
    const host = new URL(rawUrl).hostname.replace(/^www\./, "");
    const [name] = host.split(".");
    return name ? name.charAt(0).toUpperCase() + name.slice(1) : host;
  } catch {
    return "Quelle";
  }
}

export function sourceShortLink(rawUrl?: string) {
  const value = String(rawUrl || "").trim();
  if (!value) return "";
  try {
    const url = new URL(value);
    const host = url.hostname.replace(/^www\./, "");
    const path = `${url.pathname}${url.search}`.replace(/\/$/, "");
    if (!path || path === "/") return host;
    const shortenedPath = path.length > 44 ? `${path.slice(0, 24)}...${path.slice(-16)}` : path;
    return `${host}${shortenedPath}`;
  } catch {
    return value.length > 54 ? `${value.slice(0, 32)}...${value.slice(-18)}` : value;
  }
}
