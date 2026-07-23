import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type ArticleSearchResponse } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleDocumentsController } from "./useVehicleDocumentsController";

const response: ArticleSearchResponse = {
  query: "Piko 50600",
  results: [{
    source: "Piko",
    title: "BR 106",
    url: "https://www.piko-shop.de/50600",
    snippet: "Dokumente",
    score: 100,
    fields: {},
    documents: [
      {
        title: "Ersatzteilliste BR 106",
        url: "https://www.piko-shop.de/50600-et.pdf",
        source: "Piko",
        kind: "spare-parts"
      },
      {
        title: "Ersatzteilliste Duplikat",
        url: "https://www.piko-shop.de/50600-et.pdf",
        source: "Piko"
      }
    ]
  }]
};

function renderController() {
  const setSaving = vi.fn();
  const onMessage = vi.fn();
  const refreshSelectedVehicle = vi.fn().mockResolvedValue(undefined);
  const t = (key: string, values?: Record<string, string | number>) =>
    values?.count === undefined ? key : `${key}:${values.count}`;
  const hook = renderHook(() => useVehicleDocumentsController({
    selected: vehicleFixture({ manufacturer: "Piko", articleNumber: "50600", attachments: [] }),
    setSaving,
    onMessage,
    refreshSelectedVehicle,
    t
  }));
  return { ...hook, setSaving, onMessage, refreshSelectedVehicle };
}

describe("useVehicleDocumentsController", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
  });

  it("deduplicates found documents and imports the selected rows", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue(response);
    vi.spyOn(api, "importVehicleAttachmentFromUrl").mockResolvedValue({
      id: "attachment-1",
      vehicleId: "vehicle-1",
      fileName: "50600-et.pdf",
      originalName: "50600-et.pdf",
      sizeBytes: 100,
      createdAt: "2026-07-23T08:00:00Z",
      updatedAt: "2026-07-23T08:00:00Z"
    });
    const { result, onMessage, refreshSelectedVehicle } = renderController();

    await act(() => result.current.commands.search());
    expect(result.current.state.documents).toHaveLength(1);

    act(() => result.current.commands.toggleAll(true));
    expect(Object.values(result.current.state.selectedDocuments)).toEqual([true]);

    act(() => result.current.commands.importSelected());
    await waitFor(() => expect(api.importVehicleAttachmentFromUrl).toHaveBeenCalledWith(
      "vehicle-1",
      expect.objectContaining({
        url: "https://www.piko-shop.de/50600-et.pdf",
        category: "Ersatzteilliste"
      })
    ));
    expect(refreshSelectedVehicle).toHaveBeenCalledWith("vehicle-1");
    expect(onMessage).toHaveBeenLastCalledWith("vehicles.uploads.webDocumentImported:1");
  });

  it("reports when web search is disabled", async () => {
    window.localStorage.setItem("railkeeper.articleSearchEnabled", "false");
    const { result } = renderController();

    await act(() => result.current.commands.search());

    expect(result.current.state.ran).toBe(true);
    expect(result.current.state.error).toContain("deaktiviert");
  });
});
