import { act, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { api } from "../../shared/api";
import { useVehicleCollectionController } from "./useVehicleCollectionController";

function setup() {
  vi.useFakeTimers();
  const vehicles = vi.spyOn(api, "vehicles").mockResolvedValue([]);
  vi.spyOn(api, "masterDataAll").mockResolvedValue({});
  vi.spyOn(api, "masterDataRelations").mockResolvedValue([]);
  return vehicles;
}

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("useVehicleCollectionController", () => {
  it("loads on mount, coalesces typing and ignores a superseded response", async () => {
    const vehicles = setup();
    let rejectInitial: (error: Error) => void = () => undefined;
    vehicles.mockImplementationOnce(() => new Promise((_, reject) => { rejectInitial = reject; }));
    const { result } = renderHook(useVehicleCollectionController);
    expect(vehicles.mock.calls).toEqual([[""]]);

    act(() => result.current.setQuery("B"));
    await act(async () => { await vi.advanceTimersByTimeAsync(100); });
    act(() => result.current.setQuery("BR"));
    expect(result.current.query).toBe("BR");
    await act(async () => {
      rejectInitial(new Error("Stale failure"));
      await vi.advanceTimersByTimeAsync(299);
    });
    expect(vehicles).toHaveBeenCalledTimes(1);
    expect(result.current.message).toBe("");

    await act(async () => { await vi.advanceTimersByTimeAsync(1); });
    expect(vehicles.mock.calls).toEqual([[""], ["BR"]]);
    expect(result.current.loading).toBe(false);
  });

  it("refreshes the latest query immediately and clears a pending search on unmount", async () => {
    const vehicles = setup();
    const { result, unmount } = renderHook(useVehicleCollectionController);
    act(() => result.current.setQuery("BR 218"));
    await act(async () => result.current.load());
    expect(vehicles.mock.calls).toEqual([[""], ["BR 218"]]);
    expect(result.current.loading).toBe(false);
    await act(async () => { await vi.advanceTimersByTimeAsync(300); });
    expect(vehicles).toHaveBeenCalledTimes(2);

    act(() => result.current.setQuery("other"));
    unmount();
    await vi.advanceTimersByTimeAsync(300);
    expect(vehicles).toHaveBeenCalledTimes(2);
  });
});
