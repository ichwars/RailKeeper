import type {
  ArticleSearchResponse,
  ArticleSearchSparePart,
  Vehicle,
  VehicleSparePart,
  VehicleSparePartInput
} from "../../shared/api";
import { sourceDisplayName } from "./articleSearch";
import { normalizedText } from "./vehicleTransforms";
import { articleSearchSources, sanitizeArticleSearchResponse } from "./vehicleViewModel";

export function sparePartImportKey(part: VehicleSparePartInput) {
  const articleNumber = part.articleNumber?.trim() || "";
  const description = part.description?.trim() || "";
  const url = part.url?.trim() || "";
  return (articleNumber || description ? `${articleNumber}|${description}` : `|${url}`).toLocaleLowerCase("de-DE");
}

export function normalizedSparePartUrl(value?: string) {
  return String(value || "").trim().replace(/\/+$/, "").toLocaleLowerCase("de-DE");
}

export function sparePartCatalogKey(
  part: VehicleSparePartInput | VehicleSparePart | ArticleSearchSparePart
) {
  const article = normalizedText(part.articleNumber || "").replace(/^et(?=\d)/, "").replace(/[^a-z0-9]/g, "");
  if (article) return `article:${article}`;
  const description = normalizedText(part.description || "").replace(/[^a-z0-9]/g, "");
  if (description) return `description:${description}`;
  const urlValue = "url" in part ? part.url : "";
  return `url:${normalizedSparePartUrl(urlValue)}`;
}

export function sparePartSearchCandidateMatches(part: VehicleSparePart, signal: string) {
  const normalizedSignal = normalizedText(signal);
  const compactSignal = normalizedSignal.replace(/[^a-z0-9]/g, "");
  const articleNumber = normalizedText(part.articleNumber || "");
  const compactArticleNumber = articleNumber.replace(/[^a-z0-9]/g, "");
  if (articleNumber) {
    return normalizedSignal.includes(articleNumber) ||
      Boolean(compactArticleNumber && compactSignal.includes(compactArticleNumber));
  }
  const descriptionTokens = normalizedText(part.description || "")
    .split(/\s+/)
    .filter((token) => token.length >= 4);
  return descriptionTokens.length > 0 &&
    descriptionTokens.slice(0, 4).some((token) => normalizedSignal.includes(token));
}

export function sparePartSearchSourcesForLookup() {
  const sources = articleSearchSources().filter((source) => source === "manufacturer" || source === "catalogs");
  return sources.length > 0 ? sources : ["manufacturer", "catalogs"];
}

export function sparePartLookupMode(vehicle: Vehicle) {
  const manufacturer = vehicle.manufacturer.trim().toLocaleLowerCase("de-DE");
  if (manufacturer.includes("piko")) return "piko";
  if (manufacturer.includes("roco")) return "roco";
  return "";
}

export function visibleSparePartUrl(part: VehicleSparePart) {
  return part.url && !part.url.startsWith("/api/v1/vehicles/") ? part.url : "";
}

export function sparePartLookupCandidates(part: VehicleSparePart, response: ArticleSearchResponse) {
  const sanitized = sanitizeArticleSearchResponse(response);
  const parts = new Map<string, ArticleSearchSparePart>();

  sanitized.results.forEach((result) => {
    const resultSignal = `${result.title || ""} ${result.snippet || ""} ${result.url || ""}`
      .toLocaleLowerCase("de-DE");
    (result.spareParts || []).forEach((candidate) => {
      const candidateSignal = `${candidate.articleNumber || ""} ${candidate.description || ""} ${candidate.url || ""}`
        .toLocaleLowerCase("de-DE");
      if (!sparePartSearchCandidateMatches(part, candidateSignal)) return;
      const key = `${candidate.price || ""}|${candidate.url || ""}`.toLocaleLowerCase("de-DE");
      if (key !== "|" && !parts.has(key)) parts.set(key, candidate);
    });
    if (sparePartSearchCandidateMatches(part, resultSignal) && result.url) {
      const fallback: ArticleSearchSparePart = {
        articleNumber: part.articleNumber || "",
        description: part.description || result.title || "",
        price: "",
        url: result.url,
        source: result.source || sourceDisplayName(result.url)
      };
      const key = `${fallback.price || ""}|${fallback.url || ""}`.toLocaleLowerCase("de-DE");
      if (key !== "|" && !parts.has(key)) parts.set(key, fallback);
    }
  });

  return Array.from(parts.values())
    .filter((candidate) => candidate.price || candidate.url)
    .sort((left, right) => {
      const priceRank = Number(Boolean(right.price)) - Number(Boolean(left.price));
      if (priceRank !== 0) return priceRank;
      const availabilityRank = Number(Boolean(right.availability)) - Number(Boolean(left.availability));
      if (availabilityRank !== 0) return availabilityRank;
      return Number(Boolean(right.url)) - Number(Boolean(left.url));
    })
    .slice(0, 5);
}

export function sparePartStatusCandidate(part: VehicleSparePart, response: ArticleSearchResponse) {
  const candidates = sanitizeArticleSearchResponse(response).results.flatMap((result) => result.spareParts || []);
  const partKey = sparePartCatalogKey(part);
  const partUrl = normalizedSparePartUrl(part.url);
  return candidates.find((candidate) => sparePartCatalogKey(candidate) === partKey) ||
    candidates.find((candidate) => partUrl && normalizedSparePartUrl(candidate.url) === partUrl) ||
    sparePartLookupCandidates(part, response)[0];
}
