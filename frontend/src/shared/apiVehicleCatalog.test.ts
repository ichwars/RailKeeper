import { afterEach, describe, expect, it, vi } from "vitest";
import { api, ApiError, type VehicleCatalogFields, type VehicleCatalogRevisionInput } from "./api";

const fields: VehicleCatalogFields = {
  name: "Model", manufacturer: "Piko", articleNumber: "123", articleSourceUrl: "", gauge: "H0", epoch: "",
  railwayCompany: "", category: "Lokomotive", gattung: "Diesellok", description: "", ean: "", productionPeriod: "", listPrice: ""
};

describe("catalog API preserves explicit confirmation", () => {
  afterEach(() => { vi.unstubAllGlobals(); document.cookie = "rk_csrf=; Max-Age=0; path=/"; });

  it("preserves stable revision identity, explicit empty selection and CSRF on repeated requests", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 201, json: async () => ({}) });
    vi.stubGlobal("fetch", fetch);
    document.cookie = "rk_csrf=catalog-csrf; path=/";
    const input: VehicleCatalogRevisionInput = {
      id: "a".repeat(32), expectedRevision: "b".repeat(64), fields, independent: true, selectedFields: []
    };
    await api.createCatalogRevision("vehicle", "owned/1", input);
    await api.createCatalogRevision("vehicle", "owned/1", input);
    for (const [path, options] of fetch.mock.calls) {
      expect(path).toBe("/api/v1/catalog/vehicle/owned%2F1/revisions");
      expect(options.headers["X-CSRF-Token"]).toBe("catalog-csrf");
      expect(JSON.parse(options.body)).toEqual(input);
    }
    expect(input.selectedFields).toEqual([]);
  });

  it("reads set context and submits only the selected adoption fields", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
    vi.stubGlobal("fetch", fetch);
    await api.catalogContext("set", "set/1");
    const input = { expectedRevision: "c".repeat(64), revisionId: "legacy-set:set/1:r1", selectedFields: ["name" as const] };
    await api.adoptCatalogRevision("set", "set/1", input);
    expect(fetch.mock.calls[0][0]).toBe("/api/v1/catalog/set/set%2F1");
    expect(fetch.mock.calls[1][0]).toBe("/api/v1/catalog/set/set%2F1/adoption");
    expect(JSON.parse(fetch.mock.calls[1][1].body)).toEqual(input);
  });

  it("exposes a catalog conflict without automatically replaying the confirmation", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: false, status: 409, statusText: "Conflict",
      json: async () => ({ error: "catalog_conflict", message: "Review the saved assignment." }) });
    vi.stubGlobal("fetch", fetch);
    await expect(api.adoptCatalogRevision("vehicle", "owned", {
      expectedRevision: "d".repeat(64), revisionId: "revision", selectedFields: []
    })).rejects.toMatchObject<Partial<ApiError>>({ status: 409, code: "catalog_conflict" });
    expect(fetch).toHaveBeenCalledTimes(1);
  });
});
