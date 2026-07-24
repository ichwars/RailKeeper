import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import {
  api,
  type ArticleSearchResponse,
  type ArticleSearchSparePart,
  type Vehicle,
  type VehicleSparePart
} from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleSparePartsController } from "./useVehicleSparePartsController";

const timestamp = "2026-07-23T08:00:00Z";

function sparePart(overrides: Partial<VehicleSparePart> = {}): VehicleSparePart {
  return {
    id: "spare-part-1",
    vehicleId: "vehicle-1",
    articleNumber: "ET-56123",
    description: "Radsatz komplett",
    price: "",
    url: "https://www.piko-shop.de/56123",
    createdAt: timestamp,
    updatedAt: timestamp,
    ...overrides
  };
}

function searchResponse(spareParts: ArticleSearchSparePart[] = [{
  articleNumber: "ET-56123",
  description: "Radsatz komplett",
  price: "12,99 EUR",
  url: "https://www.piko-shop.de/56123",
  availability: "lieferbar"
}]): ArticleSearchResponse {
  return {
    query: "ET-56123",
    results: [{
      source: "Piko",
      title: "Piko Ersatzteile",
      url: "https://www.piko-shop.de/ersatzteile",
      snippet: "ET-56123 Radsatz komplett",
      score: 100,
      fields: {},
      spareParts
    }]
  };
}

function renderController(selected: Vehicle | null = vehicleFixture({
  manufacturer: "Piko",
  articleNumber: "50600",
  spareParts: [sparePart()]
}), active = false) {
  const setSaving = vi.fn();
  const onMessage = vi.fn();
  const onOpenSpareParts = vi.fn();
  const refreshSelectedVehicle = vi.fn().mockResolvedValue(undefined);
  const t = (key: string, values?: Record<string, string | number>) =>
    values?.count === undefined ? key : `${key}:${values.count}`;
  const hook = renderHook(() => useVehicleSparePartsController({
    selected,
    active,
    attachmentEdits: {},
    setSaving,
    onMessage,
    onOpenSpareParts,
    refreshSelectedVehicle,
    t
  }));
  return { ...hook, setSaving, onMessage, onOpenSpareParts, refreshSelectedVehicle };
}

describe("useVehicleSparePartsController", () => {
  beforeEach(() => window.localStorage.clear());

  it("edits, creates and deletes spare parts", async () => {
    const part = sparePart();
    vi.spyOn(api, "createVehicleSparePart").mockResolvedValue(part);
    vi.spyOn(api, "updateVehicleSparePart").mockResolvedValue(part);
    vi.spyOn(api, "deleteVehicleSparePart").mockResolvedValue(undefined);
    const { result, refreshSelectedVehicle, setSaving } = renderController();

    act(() => result.current.commands.edit(part));
    act(() => result.current.commands.updateForm({ price: " 13,49 EUR " }));
    act(() => result.current.commands.save());
    await waitFor(() => expect(api.updateVehicleSparePart).toHaveBeenCalledWith(
      "vehicle-1",
      part.id,
      expect.objectContaining({ price: "13,49 EUR" })
    ));

    act(() => result.current.commands.updateForm({ articleNumber: "ET-99999", description: "Kupplung" }));
    act(() => result.current.commands.save());
    await waitFor(() => expect(api.createVehicleSparePart).toHaveBeenCalledWith(
      "vehicle-1",
      expect.objectContaining({ articleNumber: "ET-99999", description: "Kupplung" })
    ));

    act(() => result.current.commands.remove(part));
    await waitFor(() => expect(api.deleteVehicleSparePart).toHaveBeenCalledWith("vehicle-1", part.id));
    expect(refreshSelectedVehicle).toHaveBeenCalledTimes(3);
    expect(setSaving).toHaveBeenLastCalledWith(false);
  });

  it("finds candidates for one stored spare part", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse());
    const part = sparePart();
    const { result } = renderController();

    await act(() => result.current.commands.searchSingle(part));

    expect(api.articleSearch).toHaveBeenCalledWith(expect.objectContaining({
      manufacturer: "Piko",
      articleNumber: "ET-56123",
      fields: expect.objectContaining({ sparePartLookup: "piko", vehicleArticleNumber: "50600" })
    }));
    expect(result.current.state.lookupResults[part.id]?.[0]).toMatchObject({
      price: "12,99 EUR",
      availability: "lieferbar"
    });
  });

  it("updates existing entries and imports new catalog entries", async () => {
    const existing = sparePart({ articleNumber: "ET-56123", price: "-", url: "" });
    vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse([
      {
        articleNumber: "ET-56123",
        description: "Radsatz komplett",
        price: "12,99 EUR",
        url: "https://www.piko-shop.de/56123"
      },
      {
        articleNumber: "ET-99999",
        description: "Kupplung",
        price: "4,99 EUR",
        url: "https://www.piko-shop.de/99999"
      }
    ]));
    vi.spyOn(api, "updateVehicleSparePart").mockResolvedValue(existing);
    vi.spyOn(api, "createVehicleSparePart").mockResolvedValue(sparePart({ id: "spare-part-2" }));
    const { result, onMessage, refreshSelectedVehicle } = renderController(vehicleFixture({
      manufacturer: "Piko",
      articleNumber: "50600",
      spareParts: [existing]
    }));

    await act(() => result.current.commands.importAll());

    expect(api.updateVehicleSparePart).toHaveBeenCalledWith("vehicle-1", existing.id, expect.objectContaining({
      price: "12,99 EUR",
      url: "https://www.piko-shop.de/56123"
    }));
    expect(api.createVehicleSparePart).toHaveBeenCalledWith("vehicle-1", expect.objectContaining({
      articleNumber: "ET-99999"
    }));
    expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1");
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.spareParts.importAllDone:2");
  });

  it("refreshes linked-part status when the tab becomes active", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse());
    const { result } = renderController(undefined, true);

    await waitFor(() => expect(result.current.state.statuses["spare-part-1"]).toMatchObject({
      availability: "lieferbar"
    }));
    expect(result.current.state.statusLoading).toEqual({});
  });
});
