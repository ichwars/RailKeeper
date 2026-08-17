import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { api, type ArticleSearchResult, type CreateVehicleRequest } from "../../shared/api";
import { translate } from "../../shared/i18n";
import { articleSelectionKey, imageSelectionKey } from "./articleSearch";
import { useArticleSearchController } from "./useArticleSearchController";

const form: CreateVehicleRequest = {
  manufacturer: "ESU",
  articleNumber: "",
  name: "BR 106",
  gauge: "H0"
};

const result: ArticleSearchResult = {
  source: "manufacturer",
  title: "BR 106",
  url: "https://www.esu.eu/br-106",
  snippet: "Diesellokomotive",
  score: 100,
  fields: {
    articleNumber: { label: "Artikelnummer", value: "12345", confidence: 1 }
  },
  images: [{
    url: "https://www.esu.eu/br-106.jpg",
    title: "BR 106",
    source: "https://www.esu.eu/br-106"
  }]
};

function renderController() {
  const replaceForm = vi.fn();
  const updateForm = vi.fn();
  const addImages = vi.fn();
  const onMessage = vi.fn();
  const hook = renderHook(() => useArticleSearchController({
    form,
    pendingImageCount: 0,
    replaceForm,
    updateForm,
    addImages,
    onMessage,
    t: (key, values) => translate("de", key, values)
  }));
  return { ...hook, replaceForm, updateForm, addImages, onMessage };
}

describe("useArticleSearchController", () => {
  it("loads suggestions and applies explicitly selected fields and images", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue({ query: "ESU 12345", results: [result] });
    const { result: controller, updateForm, addImages } = renderController();

    act(() => controller.current.commands.run());
    await waitFor(() => expect(controller.current.state.loading).toBe(false));
    expect(controller.current.state.response?.results).toHaveLength(1);

    expect(controller.current.state.selectedFields[articleSelectionKey(result, "articleNumber", 0)]).toBe(true);
    act(() => controller.current.commands.toggleImage(result, 0, result.images![0], true));
    expect(controller.current.state.selectedImages[imageSelectionKey(result, result.images![0], 0)]).toBe(true);

    act(() => controller.current.commands.applyResult(result));
    expect(updateForm).toHaveBeenCalledWith({ articleNumber: "12345" });
    expect(addImages).toHaveBeenCalledWith([expect.objectContaining({
      url: result.images![0].url,
      isPrimary: true
    })]);
    expect(controller.current.state.open).toBe(false);
  });

  it("reports missing search criteria without calling the API", () => {
    const articleSearch = vi.spyOn(api, "articleSearch");
    const onMessage = vi.fn();
    const { result: controller } = renderHook(() => useArticleSearchController({
      form: { manufacturer: "", name: "", gauge: "" },
      pendingImageCount: 0,
      replaceForm: vi.fn(),
      updateForm: vi.fn(),
      addImages: vi.fn(),
      onMessage,
      t: (key, values) => translate("de", key, values)
    }));

    act(() => controller.current.commands.run());

    expect(articleSearch).not.toHaveBeenCalled();
    expect(onMessage).toHaveBeenCalledWith(
      "Bitte Artikel-Nr. oder Bezeichnung sowie Hersteller und Spurweite erfassen."
    );
  });
});
