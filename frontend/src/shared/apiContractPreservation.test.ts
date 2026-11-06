import { afterEach, describe, expect, expectTypeOf, it, vi } from "vitest";
import type { components } from "../generated/api";
import { api, type CreateVehicleRequest, type VehicleCVValueInput } from "./api";

describe("generated API transport preserves write semantics", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    document.cookie = "rk_csrf=; Max-Age=0; path=/";
  });

  it.each([
    { label: "omitted", images: undefined },
    { label: "explicitly empty", images: [] }
  ])("keeps images $label", async ({ images }) => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({ id: "vehicle-1" }) });
    vi.stubGlobal("fetch", fetch);
    document.cookie = "rk_csrf=contract-csrf; path=/";
    const draft: CreateVehicleRequest = { name: "Reference", manufacturer: "Reference", gauge: "TT" };
    if (images !== undefined) draft.images = images;

    await api.updateVehicle("vehicle-1", draft);

    const [url, options] = fetch.mock.calls[0];
    expect(url).toBe("/api/v1/vehicles/vehicle-1");
    expect(options.headers["X-CSRF-Token"]).toBe("contract-csrf");
    const body = JSON.parse(options.body);
    expect(Object.hasOwn(body, "images")).toBe(images !== undefined);
    if (images !== undefined) expect(body.images).toEqual([]);
    expect(body).not.toHaveProperty("functionChanges");
  });

  it("preserves unindexed and zero-page CV identity and source provenance", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 201, json: async () => ({ id: "cv-1" }) });
    vi.stubGlobal("fetch", fetch);
    const common: VehicleCVValueInput = {
      cvNumber: 257, value: 12, decoderProfile: "Reference", sourceFileId: "source-1",
      sourceLine: 7, sourceExportedAt: "2026-11-06T10:00:00"
    };
    await api.createVehicleCVValue("vehicle-1", common);
    await api.createVehicleCVValue("vehicle-1", { ...common, indexCV31: 0, indexCV32: 0 });
    const unindexed = JSON.parse(fetch.mock.calls[0][1].body);
    const indexed = JSON.parse(fetch.mock.calls[1][1].body);
    expect(unindexed).not.toHaveProperty("indexCV31");
    expect(unindexed).not.toHaveProperty("indexCV32");
    expect(indexed).toEqual({ ...common, indexCV31: 0, indexCV32: 0 });
    expectTypeOf<components["schemas"]["VehicleFunction"]["directionDependent"]>()
      .toEqualTypeOf<boolean | null>();
    expectTypeOf<components["schemas"]["VehicleCVValueInput"]["indexCV31"]>()
      .toEqualTypeOf<number | null | undefined>();
  });
});
