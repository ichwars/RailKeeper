import { act, renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";

import { api, type VehicleSet } from "../../shared/api";
import { vehicleFixture } from "../../test/fixtures/vehicles";
import { useVehicleSetController } from "./useVehicleSetController";
import { emptyVehicle } from "./vehicleViewModel";

afterEach(() => { vi.restoreAllMocks(); localStorage.clear(); });

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
