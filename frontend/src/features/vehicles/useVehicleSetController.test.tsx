import { act, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";

import { api, type VehicleSet } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleSetController } from "./useVehicleSetController";
import { emptyVehicle } from "./vehicleViewModel";

afterEach(() => { vi.restoreAllMocks(); localStorage.clear(); });

it("invalidates captured and rendered set scopes when the same set reopens or closes", () => {
  const options = { username: "tester", form: emptyVehicle, vehicles: [], pendingArticleImages: [],
    openDetail: vi.fn(), setSaving: vi.fn(), setSelectedDetail: vi.fn(), setMode: vi.fn(),
    setActiveTab: vi.fn(), closeModal: vi.fn(), openCreate: vi.fn(), load: vi.fn(), onMessage: vi.fn(),
    t: (key: string) => key };
  const { result, unmount } = renderHook(() => useVehicleSetController(options));
  act(() => result.current.openVehicleSet("set", "edit"));
  const scope = result.current.captureVehicleSetScope(), session = result.current.vehicleSetSessionKey;
  expect(scope()).toBe(true);
  act(() => result.current.openVehicleSet("set", "edit"));
  expect(scope()).toBe(false); expect(result.current.captureVehicleSetScope(session)()).toBe(false);
  const next = result.current.captureVehicleSetScope();
  act(() => result.current.closeVehicleSetDialog()); expect(next()).toBe(false);
  act(() => result.current.openVehicleSet("set", "view"));
  const last = result.current.captureVehicleSetScope(); unmount(); expect(last()).toBe(false);
});

it("does not replace a new editor or clear its draft when an older set creation finishes", async () => {
  let finish!: (set: VehicleSet) => void, current = true;
  vi.spyOn(api, "createVehicleSet").mockReturnValue(new Promise(resolve => { finish = resolve; }));
  const options = { username: "tester", form: emptyVehicle, vehicles: [], pendingArticleImages: [],
    openDetail: vi.fn(), setSaving: vi.fn(), setSelectedDetail: vi.fn(), setMode: vi.fn(),
    setActiveTab: vi.fn(), closeModal: vi.fn(), openCreate: vi.fn(), load: vi.fn(), onMessage: vi.fn(),
    t: (key: string) => key, captureScope: () => () => current };
  const { result } = renderHook(() => useVehicleSetController(options));
  localStorage.setItem("railkeeper.vehicleCreateDraft.v2:tester", "new draft");
  await act(async () => {
    const pending = result.current.submitVehicleSet([], {}); current = false;
    finish({ id: "set", members: [vehicleFixture()] } as VehicleSet); await pending;
  });
  expect(options.setSelectedDetail).not.toHaveBeenCalled();
  expect(options.setMode).not.toHaveBeenCalled();
  expect(options.setSaving.mock.calls).toEqual([[true]]);
  expect(localStorage.getItem("railkeeper.vehicleCreateDraft.v2:tester")).toBe("new draft");
});
