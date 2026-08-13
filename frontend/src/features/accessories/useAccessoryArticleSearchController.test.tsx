import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type ArticleSearchResult, type MasterDataEntry } from "../../shared/api";
import { articleSelectionKey, imageSelectionKey } from "../../shared/articleSearch/articleSearchModel";
import { emptyArticleEditorForm } from "./articleEditorModel";
import { useAccessoryArticleSearchController } from "./useAccessoryArticleSearchController";

const entries = (type: string, key: string, label: string): MasterDataEntry[] => [{
  id: `${type}:${key}`,
  type,
  key,
  label,
  active: true,
  sortOrder: 10,
  metadata: {},
  createdAt: "2026-08-13T08:00:00Z",
  updatedAt: "2026-08-13T08:00:00Z"
}];

const searchResult: ArticleSearchResult = {
  source: "manufacturer",
  title: "Tillig 83101",
  url: "https://www.tillig.com/83101",
  snippet: "Gerades Gleis",
  score: 100,
  fields: {
    manufacturer: { label: "Hersteller", value: "Tillig", confidence: 1 },
    articleNumber: { label: "Artikelnummer", value: "83101", confidence: 1 },
    name: { label: "Bezeichnung", value: "Gerades Gleis", confidence: 1 },
    gauge: { label: "Spurweite", value: "TT", confidence: 1 },
    articleType: { label: "Artikelart", value: "track", confidence: 1 }
  },
  images: [{
    url: "https://www.tillig.com/83101.jpg",
    title: "Tillig 83101",
    source: "https://www.tillig.com/83101"
  }]
};

function renderController(readOnly = false) {
  const form = {
    ...emptyArticleEditorForm(),
    manufacturer: "Tillig",
    articleNumber: "83101",
    gauges: ["TT"],
    name: "Eigener Name"
  };
  const updateForm = vi.fn();
  const addImages = vi.fn();
  const hook = renderHook(() => useAccessoryArticleSearchController({
    form,
    readOnly,
    pendingImageCount: 0,
    manufacturers: entries("manufacturer", "tillig", "Tillig"),
    gauges: entries("gauge", "tt", "TT"),
    updateForm,
    addImages
  }));
  return { ...hook, updateForm, addImages };
}

describe("useAccessoryArticleSearchController", () => {
  beforeEach(() => window.localStorage.clear());

  it("preselects only usable empty accessory fields and applies selected images", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue({ query: "Tillig 83101 TT", results: [searchResult] });
    const { result, updateForm, addImages } = renderController();

    act(() => result.current.commands.run());
    await waitFor(() => expect(result.current.state.loading).toBe(false));

    expect(result.current.state.selectedFields[articleSelectionKey(searchResult, "name", 0)]).toBe(false);
    expect(result.current.state.selectedFields[articleSelectionKey(searchResult, "articleType", 0)]).toBeUndefined();
    act(() => result.current.commands.toggleField(searchResult, 0, "name", true));
    act(() => result.current.commands.toggleImage(searchResult, 0, searchResult.images![0], true));
    expect(result.current.state.selectedImages[imageSelectionKey(searchResult, searchResult.images![0], 0)])
      .toBe(true);

    act(() => result.current.commands.applyResult(searchResult));
    expect(updateForm).toHaveBeenCalledWith({ name: "Gerades Gleis" });
    expect(addImages).toHaveBeenCalledWith([expect.objectContaining({
      url: searchResult.images![0].url,
      isPrimary: true
    })]);
    expect(result.current.state.open).toBe(false);
  });

  it("uses a scanned barcode as the only search field", async () => {
    const articleSearch = vi.spyOn(api, "articleSearch")
      .mockResolvedValue({ query: "4012278000011", results: [] });
    const { result, updateForm } = renderController();

    act(() => result.current.commands.openBarcode());
    act(() => result.current.setters.setBarcodeValue(" 4012278000011 "));
    act(() => result.current.commands.submitBarcode({ preventDefault: vi.fn() } as never));
    await waitFor(() => expect(result.current.state.loading).toBe(false));

    expect(updateForm).toHaveBeenCalledWith({ ean: "4012278000011" });
    expect(articleSearch).toHaveBeenCalledWith(expect.objectContaining({ fields: { ean: "4012278000011" } }));
    expect(result.current.state.barcodeOpen).toBe(false);
  });

  it("does not offer unknown or inactive master-data values for selection", async () => {
    const unknownResult: ArticleSearchResult = {
      ...searchResult,
      fields: {
        ...searchResult.fields,
        manufacturer: { label: "Hersteller", value: "Unbekannt GmbH", confidence: 1 },
        gauge: { label: "Spurweite", value: "IIm", confidence: 1 }
      }
    };
    vi.spyOn(api, "articleSearch").mockResolvedValue({ query: "4012278000011", results: [unknownResult] });
    const updateForm = vi.fn();
    const { result } = renderHook(() => useAccessoryArticleSearchController({
      form: { ...emptyArticleEditorForm(), ean: "4012278000011" },
      readOnly: false,
      pendingImageCount: 0,
      manufacturers: entries("manufacturer", "legacy", "Unbekannt GmbH")
        .map((entry) => ({ ...entry, active: false })),
      gauges: [],
      updateForm,
      addImages: vi.fn()
    }));

    act(() => result.current.commands.run());
    await waitFor(() => expect(result.current.state.loading).toBe(false));

    const manufacturerKey = articleSelectionKey(unknownResult, "manufacturer", 0);
    const gaugeKey = articleSelectionKey(unknownResult, "gauge", 0);
    expect(result.current.state.selectedFields[manufacturerKey]).toBeUndefined();
    expect(result.current.state.selectedFields[gaugeKey]).toBeUndefined();
    expect(result.current.commands.canSelectField("manufacturer", "Unbekannt GmbH")).toBe(false);
    expect(result.current.commands.canSelectField("gauge", "IIm")).toBe(false);

    act(() => result.current.commands.toggleField(unknownResult, 0, "manufacturer", true));
    expect(result.current.state.selectedFields[manufacturerKey]).toBeUndefined();
    act(() => result.current.commands.applyResult(unknownResult));
    expect(updateForm).toHaveBeenCalledWith(expect.not.objectContaining({
      manufacturer: expect.anything(),
      gauges: expect.anything()
    }));
  });

  it("does not open searches in read-only mode", () => {
    const articleSearch = vi.spyOn(api, "articleSearch");
    const { result } = renderController(true);

    act(() => result.current.commands.run());
    act(() => result.current.commands.openBarcode());

    expect(articleSearch).not.toHaveBeenCalled();
    expect(result.current.state.open).toBe(false);
    expect(result.current.state.barcodeOpen).toBe(false);
  });
});
