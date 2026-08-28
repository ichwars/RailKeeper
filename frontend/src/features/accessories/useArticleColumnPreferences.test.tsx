import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import {
  articleTableColumnSettingKey,
  defaultArticleTableColumns
} from "./articleTableColumns";
import { useArticleColumnPreferences } from "./useArticleColumnPreferences";

describe("useArticleColumnPreferences", () => {
  beforeEach(() => window.localStorage.clear());
  afterEach(() => vi.restoreAllMocks());

  it("prefers the profile over a legacy browser value", async () => {
    window.localStorage.setItem(articleTableColumnSettingKey, '["name"]');
    vi.spyOn(api, "profileSettings").mockResolvedValue({
      settings: { [articleTableColumnSettingKey]: '["storage","inventoryNumber"]' }
    });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });

    const { result } = renderHook(() => useArticleColumnPreferences(vi.fn()));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.columns).toEqual(["storage", "inventoryNumber"]);
    expect(update).not.toHaveBeenCalled();
  });

  it("migrates legacy browser columns when the profile has no layout", async () => {
    window.localStorage.setItem(articleTableColumnSettingKey, '["name","storage"]');
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });

    const { result } = renderHook(() => useArticleColumnPreferences(vi.fn()));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.columns).toEqual(["name", "storage"]);
    await waitFor(() => expect(update).toHaveBeenCalledOnce());
    expect(JSON.parse(Object.values(update.mock.calls[0]![0])[0]!)).toEqual({
      version: 1,
      columns: ["name", "storage"],
      widths: {}
    });
  });

  it("orders, resizes, and resets the complete profile layout", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({
      settings: {
        [articleTableColumnSettingKey]: JSON.stringify({
          version: 1,
          columns: ["inventoryNumber", "name", "storage"],
          widths: { name: 280 }
        })
      }
    });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
    const { result } = renderHook(() => useArticleColumnPreferences(vi.fn()));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.moveColumn("storage", "up"));
    expect(result.current.columns).toEqual(["inventoryNumber", "storage", "name"]);

    act(() => result.current.previewColumnWidth("name", 300));
    expect(result.current.widths).toEqual({ name: 300 });
    act(() => result.current.commitColumnWidth("name", 304));
    act(() => result.current.resetColumns());

    await waitFor(() => expect(update).toHaveBeenCalledTimes(3));
    expect(result.current.columns).toEqual(defaultArticleTableColumns);
    expect(result.current.widths).toEqual({});
  });
});
