import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, type AccessoryArticleListResult } from "../../shared/api";
import { useArticleOverview } from "./useArticleOverview";

const emptyResult: AccessoryArticleListResult = {
  items: [],
  metrics: {
    articleCount: 0,
    articleTypeCount: 0,
    available: 0,
    locationCount: 0,
    reserved: 0,
    installed: 0,
    careHintCount: 0
  },
  filters: { manufacturers: [], articleTypes: [], gauges: [], storageLocations: [] }
};

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((next) => { resolve = next; });
  return { promise, resolve };
}

describe("useArticleOverview", () => {
  beforeEach(() => {
    vi.spyOn(api, "accessoryArticles").mockResolvedValue(emptyResult);
  });

  it("maps instant search and every filter to a stable server query", async () => {
    const { result } = renderHook(() => useArticleOverview());
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      sort: "inventoryNumber",
      direction: "asc"
    }));

    act(() => result.current.setFilter("query", "Tillig 83101"));
    act(() => result.current.setFilter("articleType", "track"));
    act(() => result.current.setFilter("manufacturer", "Tillig"));
    act(() => result.current.setFilter("gauge", "TT"));
    act(() => result.current.setFilter("status", "allocated"));
    act(() => result.current.setFilter("locationId", "location-1"));

    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      query: "Tillig 83101",
      articleTypes: ["track"],
      manufacturer: "Tillig",
      gauges: ["TT"],
      statuses: ["reserved", "installed"],
      locationId: "location-1",
      sort: "inventoryNumber",
      direction: "asc"
    }));

    act(() => result.current.resetFilters());
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      sort: "inventoryNumber",
      direction: "asc"
    }));
    expect(result.current.hasActiveFilters).toBe(false);
  });

  it("keeps the sort key stable and toggles only the active direction", async () => {
    const { result } = renderHook(() => useArticleOverview());
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.setSort("stock"));
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      sort: "stock", direction: "asc"
    }));
    act(() => result.current.setSort("stock"));
    await waitFor(() => expect(api.accessoryArticles).toHaveBeenLastCalledWith({
      sort: "stock", direction: "desc"
    }));
    expect(result.current.sort).toBe("stock");
    expect(result.current.direction).toBe("desc");
  });

  it("ignores stale responses from superseded searches", async () => {
    const first = deferred<AccessoryArticleListResult>();
    const second = deferred<AccessoryArticleListResult>();
    vi.mocked(api.accessoryArticles).mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
    const { result } = renderHook(() => useArticleOverview());

    act(() => result.current.setFilter("query", "new"));
    second.resolve({ ...emptyResult, metrics: { ...emptyResult.metrics, articleCount: 2 } });
    await waitFor(() => expect(result.current.data.metrics.articleCount).toBe(2));
    first.resolve({ ...emptyResult, metrics: { ...emptyResult.metrics, articleCount: 99 } });
    await act(async () => { await Promise.resolve(); });

    expect(result.current.data.metrics.articleCount).toBe(2);
    expect(result.current.loading).toBe(false);
  });
});
