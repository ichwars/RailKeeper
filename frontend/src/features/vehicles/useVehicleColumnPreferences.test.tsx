import { act, renderHook, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api } from "../../shared/api";
import { defaultVehicleTableColumns } from "./vehicleTableColumns";
import {
  useVehicleColumnPreferences,
  vehicleTableColumnSettingKey
} from "./useVehicleColumnPreferences";

describe("useVehicleColumnPreferences", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("loads the ordered server preference without local storage", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({
      settings: { [vehicleTableColumnSettingKey]: '["series","inventoryNumber"]' }
    });

    const { result } = renderHook(() => useVehicleColumnPreferences(vi.fn()));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.columns).toEqual(["series", "inventoryNumber"]);
    expect(window.localStorage.getItem(vehicleTableColumnSettingKey)).toBeNull();
  });

  it("uses defaults when the server has no column preference", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });

    const { result } = renderHook(() => useVehicleColumnPreferences(vi.fn()));

    await waitFor(() => expect(result.current.loading).toBe(false));
    expect(result.current.columns).toEqual(defaultVehicleTableColumns);
		expect(result.current.columns[0]).toBe("type");
  });

  it("saves normalized changes as a partial profile update", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({ settings: {} });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
    const { result } = renderHook(() => useVehicleColumnPreferences(vi.fn()));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.toggleColumn("series"));

    await waitFor(() => expect(update).toHaveBeenCalledWith({
      [vehicleTableColumnSettingKey]: expect.stringContaining("series")
    }));
    expect(result.current.columns.at(-1)).toBe("series");
  });

  it("previews widths locally, commits them in the versioned layout, and resets all widths", async () => {
    vi.spyOn(api, "profileSettings").mockResolvedValue({
      settings: {
        [vehicleTableColumnSettingKey]: JSON.stringify({
          version: 1,
          columns: ["inventoryNumber", "name"],
          widths: { name: 286 }
        })
      }
    });
    const update = vi.spyOn(api, "updateProfileSettings").mockResolvedValue({ settings: {} });
    const { result } = renderHook(() => useVehicleColumnPreferences(vi.fn()));
    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.widths).toEqual({ name: 286 });
    act(() => result.current.previewColumnWidth("name", 300));
    expect(result.current.widths).toEqual({ name: 300 });
    expect(update).not.toHaveBeenCalled();

    act(() => result.current.commitColumnWidth("name", 304));
    await waitFor(() => expect(update).toHaveBeenCalledOnce());
    expect(JSON.parse(Object.values(update.mock.calls[0]![0])[0]!)).toEqual({
      version: 1,
      columns: ["inventoryNumber", "name"],
      widths: { name: 304 }
    });

    act(() => result.current.resetColumns());
    await waitFor(() => expect(update).toHaveBeenCalledTimes(2));
    expect(result.current.widths).toEqual({});
    expect(result.current.columns).toEqual(defaultVehicleTableColumns);
  });

  it("reports load and save failures without discarding the visible selection", async () => {
    const onMessage = vi.fn();
    vi.spyOn(api, "profileSettings").mockRejectedValue(new Error("offline"));
    vi.spyOn(api, "updateProfileSettings").mockRejectedValue(new Error("offline"));
    const { result } = renderHook(() => useVehicleColumnPreferences(onMessage));
    await waitFor(() => expect(result.current.loading).toBe(false));

    act(() => result.current.toggleColumn("series"));

    await waitFor(() => expect(onMessage).toHaveBeenCalledTimes(2));
    expect(result.current.columns).toContain("series");
  });
});
