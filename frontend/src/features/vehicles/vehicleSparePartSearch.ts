import {
  api,
  type ArticleSearchResponse,
  type ArticleSearchSparePart,
  type Vehicle,
  type VehicleSparePart,
  type VehicleSparePartInput
} from "../../shared/api";
import { sourceDisplayName } from "./articleSearch";
import { strictCleanSparePartDescription } from "./VehicleSparePartsTab";
import { sparePartCatalogKey, sparePartSearchCandidateMatches } from "./vehicleSpareParts";
import { sanitizeArticleSearchResponse } from "./vehicleViewModel";

type SparePartImportPlan = {
  creates: Map<string, VehicleSparePartInput>;
  updates: Map<string, { id: string; input: VehicleSparePartInput }>;
};

export function buildSparePartImportPlan(
  vehicle: Pick<Vehicle, "spareParts">,
  response: ArticleSearchResponse
): SparePartImportPlan {
  const existingPartsByKey = new Map((vehicle.spareParts || []).map((part) => [sparePartCatalogKey(part), part]));
  const creates = new Map<string, VehicleSparePartInput>();
  const updates = new Map<string, { id: string; input: VehicleSparePartInput }>();

  sanitizeArticleSearchResponse(response).results.forEach((searchResult) => {
    (searchResult.spareParts || []).forEach((part) => {
      const input: VehicleSparePartInput = {
        articleNumber: part.articleNumber?.trim() || "",
        description: strictCleanSparePartDescription(part.description || "") || part.description || "",
        price: part.price?.trim() || "",
        url: part.url?.trim() || searchResult.url || ""
      };
      const key = sparePartCatalogKey(input);
      if (key === "url:") return;
      const existingPart = existingPartsByKey.get(key);
      if (existingPart) {
        const existingPrice = existingPart.price && existingPart.price !== "-" ? existingPart.price : "";
        const existingUrl = existingPart.url?.startsWith("/api/v1/vehicles/") ? "" : existingPart.url || "";
        const nextPrice = existingPrice || input.price || "";
        const nextUrl = existingUrl || input.url || "";
        if ((nextPrice && nextPrice !== existingPrice) || (nextUrl && nextUrl !== existingUrl)) {
          updates.set(existingPart.id, {
            id: existingPart.id,
            input: {
              articleNumber: existingPart.articleNumber || input.articleNumber || "",
              description: existingPart.description || input.description || "",
              price: nextPrice,
              url: nextUrl
            }
          });
        }
        return;
      }
      const current = creates.get(key);
      creates.set(key, current ? {
        articleNumber: current.articleNumber || input.articleNumber || "",
        description: current.description || input.description || "",
        price: current.price || input.price || "",
        url: current.url || input.url || ""
      } : input);
    });
  });

  return { creates, updates };
}

export async function searchStoredSpareParts(
  vehicle: Pick<Vehicle, "manufacturer">,
  storedParts: VehicleSparePart[],
  searchSources: string[]
) {
  const responses: { part: VehicleSparePart; response: ArticleSearchResponse }[] = [];
  let failedSearches = 0;
  for (let index = 0; index < storedParts.length; index += 3) {
    const batchResults = await Promise.allSettled(storedParts.slice(index, index + 3).map((part) =>
      api.articleSearch({
        manufacturer: vehicle.manufacturer,
        articleNumber: part.articleNumber || "",
        name: "",
        gauge: "",
        searchSources,
        fields: { manufacturer: vehicle.manufacturer || "", articleNumber: part.articleNumber || "" }
      }).then((response) => ({ part, response }))
    ));
    batchResults.forEach((result) => {
      if (result.status === "fulfilled") responses.push(result.value);
      else failedSearches += 1;
    });
  }

  const parts = new Map<string, ArticleSearchSparePart>();
  responses.forEach(({ part, response }) => {
    sanitizeArticleSearchResponse(response).results.forEach((searchResult) => {
      const resultSignal = `${searchResult.title || ""} ${searchResult.snippet || ""} ${searchResult.url || ""}`
        .toLocaleLowerCase("de-DE");
      (searchResult.spareParts || []).forEach((candidate) => {
        const candidateSignal = `${candidate.articleNumber || ""} ${candidate.description || ""} ${candidate.url || ""}`
          .toLocaleLowerCase("de-DE");
        if (!sparePartSearchCandidateMatches(part, candidateSignal)) return;
        const key = `${candidate.articleNumber || ""}|${candidate.description || ""}|${candidate.url || ""}`
          .toLocaleLowerCase("de-DE");
        if (key !== "||" && !parts.has(key)) parts.set(key, candidate);
      });
      if (sparePartSearchCandidateMatches(part, resultSignal) && searchResult.url) {
        const fallback: ArticleSearchSparePart = {
          articleNumber: part.articleNumber || "",
          description: part.description || searchResult.title || "",
          price: part.price || "",
          url: searchResult.url,
          source: searchResult.source || sourceDisplayName(searchResult.url)
        };
        const key = `${fallback.articleNumber || ""}|${fallback.description || ""}|${fallback.url || ""}`
          .toLocaleLowerCase("de-DE");
        if (key !== "||" && !parts.has(key)) parts.set(key, fallback);
      }
    });
  });

  return { parts: Array.from(parts.values()), failedSearches };
}
