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
  const hook = renderHook(({ form, targetKey }) => useArticleSearchController({
    form, targetKey,
    pendingImageCount: 0,
    replaceForm,
    updateForm,
    addImages,
    onMessage,
    t: (key, values) => translate("de", key, values)
  }), { initialProps: { form, targetKey: "first" } });
  return { ...hook, replaceForm, updateForm, addImages, onMessage };
}

describe("useArticleSearchController", () => {
  it("ignores superseded replies and keeps the latest request loading", async () => {
    let first!: (value: { query: string; results: ArticleSearchResult[] }) => void;
    let second!: typeof first;
    vi.spyOn(api, "articleSearch").mockReturnValueOnce(new Promise(resolve => { first = resolve; }))
      .mockReturnValueOnce(new Promise(resolve => { second = resolve; }));
    const { result: controller } = renderController();
    act(() => controller.current.commands.run()); act(() => controller.current.commands.run());
    await act(async () => { first({ query: "old", results: [result] }); });
    expect(controller.current.state).toMatchObject({ response: null, loading: true });
    await act(async () => { second({ query: "new", results: [result] }); });
    expect(controller.current.state.response?.query).toBe("new");
    expect(controller.current.state.loading).toBe(false);
  });

  it("invalidates another vehicle's pending response and callbacks", async () => {
    let complete!: (value: { query: string; results: ArticleSearchResult[] }) => void;
    const request = vi.spyOn(api, "articleSearch").mockReturnValue(new Promise(resolve => { complete = resolve; }));
    const { result: controller, rerender, updateForm } = renderController();
    const old = controller.current.commands;
    act(() => old.run()); rerender({ form: { ...form, name: "Other" }, targetKey: "other" });
    act(() => old.run()); expect(request).toHaveBeenCalledOnce();
    await act(async () => { complete({ query: "old", results: [result] }); });
    expect(controller.current.state).toMatchObject({ response: null, open: false, loading: false });
    act(() => old.applyResult(result)); expect(updateForm).not.toHaveBeenCalled();
  });

  it("deselects changed target fields without applying any fields or images until reviewed again", async () => {
    vi.spyOn(api, "articleSearch").mockResolvedValue({ query: "source", results: [result] });
    const { result: controller, rerender, updateForm, addImages } = renderController();
    act(() => controller.current.commands.run()); await waitFor(() => expect(controller.current.state.loading).toBe(false));
    act(() => controller.current.commands.toggleImage(result, 0, result.images![0], true));
    const old = controller.current.commands;
    rerender({ form: { ...form, articleNumber: "Own newer number" }, targetKey: "first" });
    act(() => old.applyResult(result));
    expect(updateForm).not.toHaveBeenCalled(); expect(addImages).not.toHaveBeenCalled();
    expect(controller.current.state.error).toContain("abgewählt");
    expect(controller.current.state.selectedFields[articleSelectionKey(result, "articleNumber", 0)]).toBe(false);
    act(() => controller.current.commands.toggleField(result, 0, "articleNumber", true));
    act(() => controller.current.commands.applyResult(result));
    expect(updateForm).toHaveBeenCalledExactlyOnceWith({ articleNumber: "12345" });
    expect(addImages).toHaveBeenCalledOnce();
  });

  it("applies the exact result when URLs repeat and consumes a confirmation only once", async () => {
    const second = { ...result, title: "Another variant", fields: { articleNumber: { label: "Article", value: "222", confidence: 1 } } };
    vi.spyOn(api, "articleSearch").mockResolvedValue({ query: "source", results: [result, second] });
    const { result: controller, updateForm } = renderController();
    act(() => controller.current.commands.run()); await waitFor(() => expect(controller.current.state.loading).toBe(false));
    act(() => controller.current.commands.applyResult({ ...second, title: "Unreviewed" }));
    expect(updateForm).not.toHaveBeenCalled();
    act(() => { controller.current.commands.applyResult(second); controller.current.commands.applyResult(second); });
    expect(updateForm).toHaveBeenCalledExactlyOnceWith({ articleNumber: "222" });
  });

  it("ignores a reply after closing the preview", async () => {
    let complete!: (value: { query: string; results: ArticleSearchResult[] }) => void;
    vi.spyOn(api, "articleSearch").mockReturnValue(new Promise(resolve => { complete = resolve; }));
    const { result: controller } = renderController();
    act(() => controller.current.commands.run()); act(() => controller.current.setters.setOpen(false));
    await act(async () => { complete({ query: "closed", results: [result] }); });
    expect(controller.current.state).toMatchObject({ open: false, response: null, loading: false });
  });

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
