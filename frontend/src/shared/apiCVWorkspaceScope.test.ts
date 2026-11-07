import { afterEach, describe, expect, it, vi } from "vitest";
import { api, type VehicleCVValueInput } from "./api";

describe("explicit CV working group transport", () => {
  afterEach(() => { vi.unstubAllGlobals(); document.cookie = "rk_csrf=; Max-Age=0; path=/"; });

  it("compares row and binding in deletions while preserving legacy calls", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 204 }); vi.stubGlobal("fetch", fetch);
    document.cookie = "rk_csrf=scoped-delete-csrf; path=/";
    const revision = "a".repeat(64), scope = { workspaceId: "raw/group +", expectedWorkspaceRevision: "b".repeat(64) };
    await api.deleteVehicleCVValue("vehicle", "cv");
    await api.deleteVehicleCVFile("vehicle", "file");
    await api.deleteVehicleCVValue("vehicle", "cv", revision, scope);
    await api.deleteVehicleCVFile("vehicle", "file", revision, scope);
    expect(fetch.mock.calls[0][0]).toBe("/api/v1/vehicles/vehicle/cv-values/cv");
    expect(fetch.mock.calls[1][0]).toBe("/api/v1/vehicles/vehicle/cv-files/file");
    for (const [url, options] of fetch.mock.calls.slice(2)) {
      const parsed = new URL(url, "http://localhost");
      expect(Object.fromEntries(parsed.searchParams)).toEqual({ expectedRevision: revision, ...scope });
      expect(options.method).toBe("DELETE"); expect(options.headers["X-CSRF-Token"]).toBe("scoped-delete-csrf");
    }
    await api.applyVehicleCVChanges("vehicle", [{ id: "cv", expectedRevision: revision, delete: true, ...scope }]);
    expect(JSON.parse(fetch.mock.calls[4][1].body).changes[0]).toEqual({ id: "cv", expectedRevision: revision, delete: true, ...scope });
  });

  it("preserves exact scope comparisons in single, batch and import writes", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 200, json: async () => ({}) });
    vi.stubGlobal("fetch", fetch);
    document.cookie = "rk_csrf=scoped-cv-csrf; path=/";
    const rowRevision = "b".repeat(64);
    const input: VehicleCVValueInput = { cvNumber: 257, value: 3, indexCV31: 0, indexCV32: 0,
      decoderProfile: "Legacy", workspaceId: "raw-group", expectedWorkspaceRevision: "a".repeat(64),
      expectedRevision: rowRevision };
    await api.createVehicleCVValue("vehicle", input);
    await api.updateVehicleCVValue("vehicle", "cv", input);
    await api.importVehicleCVValues("vehicle", [input]);
    const { expectedRevision, ...value } = input;
    await api.applyVehicleCVChanges("vehicle", [{ id: "cv", expectedRevision: rowRevision, value }]);
    const bodies = fetch.mock.calls.map(([, options]) => JSON.parse(options.body));
    expect(bodies).toEqual([input, input, { values: [input] }, { changes: [{ id: "cv", expectedRevision, value }] }]);
    for (const [, options] of fetch.mock.calls) expect(options.headers["X-CSRF-Token"]).toBe("scoped-cv-csrf");
  });

  it("adds upload scope only when supplied and retains the caller's working label", async () => {
    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 201, json: async () => ({}) });
    vi.stubGlobal("fetch", fetch);
    const file = new File(["CV;Value\n1;3\n"], "scope.csv", { type: "text/csv" });
    await api.uploadVehicleCVFile("vehicle", file, "Legacy", "Original description");
    await api.uploadVehicleCVFile("vehicle", file, "Legacy", "Original description", "raw-group", "c".repeat(64));
    const legacy: FormData = fetch.mock.calls[0][1].body, scoped: FormData = fetch.mock.calls[1][1].body;
    expect(legacy.has("workspaceId")).toBe(false);
    expect(legacy.has("expectedWorkspaceRevision")).toBe(false);
    expect(scoped.get("workspaceId")).toBe("raw-group");
    expect(scoped.get("expectedWorkspaceRevision")).toBe("c".repeat(64));
    expect(scoped.get("decoderProfile")).toBe("Legacy");
    expect(scoped.get("description")).toBe("Original description");
  });
});
