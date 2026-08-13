import type { ArticleSearchImage, ArticleSearchResult } from "../api";

export type ArticleSearchFieldDefinition = {
  key: string;
  label: string;
};

export type ArticleSearchFieldGroup = {
  key: string;
  label: string;
  fields: ArticleSearchFieldDefinition[];
};

export type Translate = (key: string, values?: Record<string, string | number>) => string;

export function articleResultKey(result: ArticleSearchResult, index = 0) {
  return `${result.url || result.title}-${index}`;
}

export function articleSelectionKey(result: ArticleSearchResult, key: string, index = 0) {
  return `${articleResultKey(result, index)}::${key}`;
}

export function imageSelectionKey(result: ArticleSearchResult, image: ArticleSearchImage, index = 0) {
  return `${articleResultKey(result, index)}::image::${image.url}`;
}

export function articleFieldStatus(current: string, found: string) {
  if (!current) return "empty" as const;
  if (current.toLocaleLowerCase("de-DE") === found.toLocaleLowerCase("de-DE")) return "same" as const;
  return "conflict" as const;
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

export function isBadCommonArticleValue(key: string, value: string) {
  const normalized = value.trim();
  if (!normalized) return true;
  if (key !== "description") return false;
  const lower = normalized.toLocaleLowerCase("de-DE");
  return [
    "die absicht ist",
    "anzeigen zu zeigen",
    "personalisierte anzeigen",
    "cookie",
    "google_analytics",
    "altersempfehlung",
    "downloads",
    "bedienungsanleitung"
  ].some((token) => lower.includes(token));
}
