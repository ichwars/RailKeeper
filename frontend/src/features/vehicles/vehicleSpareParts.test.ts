import { describe, expect, it } from "vitest";

import type { ArticleSearchResponse, VehicleSparePart } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import {
  normalizedSparePartUrl,
  sparePartCatalogKey,
  sparePartImportKey,
  sparePartLookupCandidates,
  sparePartLookupMode,
  sparePartSearchCandidateMatches,
  sparePartStatusCandidate,
  visibleSparePartUrl
} from "./vehicleSpareParts";

const timestamp = "2026-07-23T08:00:00Z";

function sparePart(overrides: Partial<VehicleSparePart> = {}): VehicleSparePart {
  return {
    id: "spare-part-1",
    vehicleId: "vehicle-1",
    articleNumber: "ET-56123",
    description: "Radsatz komplett",
    price: "",
    url: "https://www.piko-shop.de/56123/",
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

function searchResponse(): ArticleSearchResponse {
  return {
    query: "ET-56123",
    results: [{
      source: "Piko",
      title: "Piko Ersatzteil ET-56123",
      url: "https://www.piko-shop.de/56123",
      snippet: "Radsatz komplett",
      score: 100,
      fields: {},
      spareParts: [{
        articleNumber: "ET-56123",
        description: "Radsatz komplett",
        price: "12,99 EUR",
        url: "https://www.piko-shop.de/56123",
        availability: "lieferbar"
      }]
    }]
  };
}

describe("vehicleSpareParts", () => {
  it("builds stable catalog and import keys", () => {
    expect(sparePartImportKey({ articleNumber: " ET-56123 ", description: " Radsatz " }))
      .toBe("et-56123|radsatz");
    expect(sparePartImportKey({ url: "https://example.test/part" }))
      .toBe("|https://example.test/part");
    expect(sparePartCatalogKey(sparePart())).toBe("article:et56123");
    expect(normalizedSparePartUrl("HTTPS://EXAMPLE.TEST/PART///")).toBe("https://example.test/part");
  });

  it("detects supported manufacturer lookups and visible URLs", () => {
    expect(sparePartLookupMode(vehicleFixture({ manufacturer: "Piko Spielwaren GmbH" }))).toBe("piko");
    expect(sparePartLookupMode(vehicleFixture({ manufacturer: "Roco" }))).toBe("roco");
    expect(sparePartLookupMode(vehicleFixture({ manufacturer: "ESU" }))).toBe("");
    expect(visibleSparePartUrl(sparePart())).toContain("piko-shop.de");
    expect(visibleSparePartUrl(sparePart({ url: "/api/v1/vehicles/vehicle-1/spare-parts/1" }))).toBe("");
  });

  it("matches identifiers and ranks priced lookup candidates first", () => {
    const part = sparePart();
    expect(sparePartSearchCandidateMatches(part, "Passend fuer ET 56123")).toBe(true);
    expect(sparePartSearchCandidateMatches(part, "Kupplung 99999")).toBe(false);

    const candidates = sparePartLookupCandidates(part, searchResponse());
    expect(candidates[0]).toMatchObject({ price: "12,99 EUR", availability: "lieferbar" });
    expect(sparePartStatusCandidate(part, searchResponse())).toEqual(candidates[0]);
  });
});
