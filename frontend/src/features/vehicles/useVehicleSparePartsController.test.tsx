import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

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

  afterEach(() => {
    vi.unstubAllGlobals();
    document.body.replaceChildren();
  });

  it("avoids smooth scrolling when reduced motion is requested", async () => {
    const matchMedia = vi.fn().mockReturnValue({ matches: true });
    vi.stubGlobal("matchMedia", matchMedia);
    const editor = document.createElement("details");
    editor.id = "vehicle-spare-parts-editor";
    editor.scrollIntoView = vi.fn();
    document.body.append(editor);
    const part = sparePart();
    const { result } = renderController();

    act(() => result.current.commands.edit(part));

    await waitFor(() => expect(editor.scrollIntoView).toHaveBeenCalledWith({
      behavior: "auto",
      block: "start"
    }));
    expect(matchMedia).toHaveBeenCalledWith("(prefers-reduced-motion: reduce)");
  });

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

describe("spare part failure and retry contracts", () => {
  it("blocks remote searches when disabled and retains manual input", async () => {
    window.localStorage.setItem("railkeeper.articleSearchEnabled", "false");
    const search = vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse());
    const { result, onMessage } = renderController();
    act(() => result.current.commands.updateForm({ description: "Manual retained input" }));
    await act(() => result.current.commands.searchSingle(sparePart()));
    await act(() => result.current.commands.runSearch());
    await act(() => result.current.commands.importAll());
    expect(search).not.toHaveBeenCalled();
    expect(result.current.state.searchError).toBe("vehicles.spareParts.searchDisabledSettings");
    expect(result.current.state.lookupErrors["spare-part-1"]).toBe("vehicles.spareParts.searchDisabledSettings");
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.spareParts.searchDisabledSettings");
    expect(result.current.state.form.description).toBe("Manual retained input");
  });

  it("rejects an empty manual row and retains edits through a failed save", async () => {
    const create = vi.spyOn(api, "createVehicleSparePart").mockRejectedValueOnce(new Error("save unavailable")).mockResolvedValue(sparePart());
    const { result, onMessage, refreshSelectedVehicle } = renderController();
    act(() => result.current.commands.save());
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.spareParts.incomplete");
    expect(create).not.toHaveBeenCalled();
    act(() => result.current.commands.updateForm({ articleNumber: " ET-700 ", description: " Keep this text ", url: " https://example.test/part ", price: " 5.00 " }));
    act(() => result.current.commands.save());
    await waitFor(() => expect(onMessage).toHaveBeenLastCalledWith("save unavailable"));
    expect(result.current.state.form.description).toBe(" Keep this text ");
    expect(refreshSelectedVehicle).not.toHaveBeenCalled();
    act(() => result.current.commands.save());
    await waitFor(() => expect(result.current.state.form.description).toBe(""));
    expect(create).toHaveBeenLastCalledWith("vehicle-1", { articleNumber: "ET-700", description: "Keep this text", url: "https://example.test/part", price: "5.00" });
    expect(refreshSelectedVehicle).toHaveBeenCalledOnce();
  });

  it("preserves article identity and individual descriptions while retrying price adoption", async () => {
    const part = sparePart({ description: "Individual description" });
    vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse());
    const update = vi.spyOn(api, "updateVehicleSparePart").mockRejectedValueOnce(new Error("concurrent conflict")).mockResolvedValue(part);
    const { result, onMessage } = renderController();
    await act(() => result.current.commands.searchSingle(part));
    const suggestion = { articleNumber: "wrong suggestion identity", description: "Catalog name" };
    act(() => result.current.commands.applyLookup(part, suggestion));
    await waitFor(() => expect(onMessage).toHaveBeenLastCalledWith("concurrent conflict"));
    expect(result.current.state.lookupResults[part.id]?.length).toBeGreaterThan(0);
    act(() => result.current.commands.applyLookup(part, suggestion));
    await waitFor(() => expect(result.current.state.lookupResults[part.id]).toBeUndefined());
    expect(update).toHaveBeenLastCalledWith("vehicle-1", part.id, { articleNumber: part.articleNumber, description: part.description, price: "12,99 EUR", url: "https://www.piko-shop.de/56123" });
  });

  it("retains candidates after a partial search failure and requires explicit selection", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValueOnce(searchResponse()).mockRejectedValueOnce("lookup offline");
    const parts = [sparePart(), sparePart({ id: "second", articleNumber: "ET-99999" })];
    const { result } = renderController(vehicleFixture({ manufacturer: "Piko", articleNumber: "50600", spareParts: parts }));
    await act(() => result.current.commands.runSearch());
    expect(result.current.state.foundParts.length).toBeGreaterThan(0);
    expect(result.current.state.searchError).toBe("vehicles.spareParts.searchPartialFailed:1");
    expect(result.current.commands.selectedInputs()).toEqual([]);
    act(() => result.current.commands.toggleAllFound(true));
    expect(result.current.commands.selectedInputs().length).toBe(result.current.state.foundParts.length);
    const key = Object.keys(result.current.state.selectedFoundParts)[0];
    act(() => result.current.commands.toggleFound(key, false));
    expect(result.current.commands.selectedInputs()).toHaveLength(result.current.state.foundParts.length - 1);
    act(() => result.current.commands.clearSelectedFound());
    expect(result.current.commands.selectedInputs()).toEqual([]);
    act(() => result.current.commands.toggleSort("price"));
    expect(result.current.state.sort).toEqual({ key: "price", direction: "asc" });
    act(() => result.current.commands.toggleSort("price"));
    expect(result.current.state.sort.direction).toBe("desc");
    act(() => result.current.commands.resetDetail());
    expect(result.current.state.foundParts).toEqual([]);
    expect(result.current.state.searchRan).toBe(false);
  });

  it("distinguishes an empty inventory from complete remote and single-item failures", async () => {
    const empty = renderController(vehicleFixture({ spareParts: [] }));
    await act(() => empty.result.current.commands.runSearch());
    expect(empty.result.current.state.searchError).toBe("vehicles.spareParts.searchDisabledEmpty");
    const failed = renderController();
    vi.spyOn(api, "articleSearch").mockRejectedValue("lookup offline");
    await act(() => failed.result.current.commands.runSearch());
    expect(failed.result.current.state.searchError).toBe("vehicles.spareParts.searchFailed");
    await act(() => failed.result.current.commands.searchSingle(sparePart()));
    expect(failed.result.current.state.lookupErrors["spare-part-1"]).toBe("lookup offline");
    expect(failed.result.current.state.lookupLoadingId).toBe("");
  });

  it("deduplicates local attachment suggestions and excludes private download URLs", async () => {
    const attachment = { id: "attachment", vehicleId: "vehicle-1", fileName: "stored.pdf", originalName: "manual.pdf", sizeBytes: 3, createdAt: timestamp, updatedAt: timestamp };
    const saveAttachment = vi.spyOn(api, "updateVehicleAttachment").mockResolvedValue(attachment);
    vi.spyOn(api, "vehicleSparePartSuggestions").mockResolvedValue([
      { articleNumber: "ET-56123", description: "Radsatz komplett" },
      { articleNumber: "ET-700", description: " New part ", url: "/api/v1/vehicles/private/attachments/local", price: " 7.50 " },
      { articleNumber: "ET-700", description: "New part" }, { articleNumber: "", description: "" }
    ]);
    const create = vi.spyOn(api, "createVehicleSparePart").mockResolvedValue(sparePart({ id: "new-part" }));
    const { result, onMessage, onOpenSpareParts, refreshSelectedVehicle } = renderController();
    await act(() => result.current.commands.extractAttachment(attachment));
    expect(saveAttachment).toHaveBeenCalledWith("vehicle-1", attachment.id, { description: "", category: "", maintenanceId: "" });
    expect(create).toHaveBeenCalledExactlyOnceWith("vehicle-1", { articleNumber: "ET-700", description: "New part", url: "", price: "7.50" });
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.uploads.extractSparePartsDone:1");
    expect(onOpenSpareParts).toHaveBeenCalledOnce();
    expect(refreshSelectedVehicle).toHaveBeenCalledOnce();
  });

  it("keeps extraction failures recoverable and shows existing suggestions without duplicate writes", async () => {
    const attachment = { id: "attachment", vehicleId: "vehicle-1", fileName: "stored.pdf", originalName: "manual.pdf", sizeBytes: 3, createdAt: timestamp, updatedAt: timestamp };
    vi.spyOn(api, "updateVehicleAttachment").mockRejectedValueOnce(new Error("description save failed")).mockResolvedValue(attachment);
    vi.spyOn(api, "vehicleSparePartSuggestions").mockResolvedValue([{ articleNumber: "ET-56123", description: "Radsatz komplett" }]);
    const { result, onMessage, onOpenSpareParts } = renderController();
    await act(() => result.current.commands.extractAttachment(attachment));
    expect(api.vehicleSparePartSuggestions).not.toHaveBeenCalled();
    expect(onMessage).toHaveBeenLastCalledWith("description save failed");
    expect(onOpenSpareParts).not.toHaveBeenCalled();
    await act(() => result.current.commands.extractAttachment(attachment));
    expect(result.current.state.searchRan).toBe(true);
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.uploads.extractSparePartsEmpty");
    expect(result.current.state.foundParts).toHaveLength(1);
  });

  it("clears busy flags after failed import without altering existing part data", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse([{ articleNumber: "ET-700", description: "New catalog part" }]));
    vi.spyOn(api, "createVehicleSparePart").mockRejectedValue("write unavailable");
    const vehicle = vehicleFixture({ manufacturer: "Piko", articleNumber: "50600", spareParts: [sparePart()] });
    const before = structuredClone(vehicle);
    const { result, onMessage, refreshSelectedVehicle } = renderController(vehicle);
    await act(() => result.current.commands.importAll());
    expect(onMessage).toHaveBeenLastCalledWith("write unavailable");
    expect(result.current.state.importAllLoading).toBe(false);
    expect(vehicle).toEqual(before);
    expect(refreshSelectedVehicle).not.toHaveBeenCalled();
    vi.mocked(api.articleSearch).mockResolvedValue(searchResponse([]));
    await act(() => result.current.commands.importAll());
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.spareParts.importAllNone");
  });

  it("guards unsupported manufacturers, missing articles and absent selections", async () => {
    const remote = vi.spyOn(api, "articleSearch").mockResolvedValue(searchResponse());
    const unsupported = renderController(vehicleFixture({ manufacturer: "Unknown manufacturer" }));
    expect(unsupported.result.current.state.canImportAll).toBe(false);
    expect(unsupported.result.current.state.importAllTitle).toBe("vehicles.spareParts.importAllUnsupported");
    await act(() => unsupported.result.current.commands.importAll());
    const missingArticle = renderController(vehicleFixture({ manufacturer: "Piko", articleNumber: "" }));
    expect(missingArticle.result.current.state.importAllTitle).toBe("vehicles.spareParts.importAllMissingArticle");
    const absent = renderController(null);
    expect(absent.result.current.state.importAllTitle).toBe("vehicles.spareParts.importAllNoVehicle");
    act(() => { absent.result.current.commands.save(); absent.result.current.commands.remove(sparePart()); absent.result.current.commands.applyLookup(sparePart(), { articleNumber: "", description: "" }); });
    await act(() => absent.result.current.commands.runSearch());
    await act(() => absent.result.current.commands.importAll());
    await act(() => absent.result.current.commands.searchSingle(sparePart()));
    expect(remote).not.toHaveBeenCalled();
  });
});
