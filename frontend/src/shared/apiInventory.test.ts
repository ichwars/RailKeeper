import { afterEach, describe, expect, it, vi } from "vitest";

import { vehicleFixture } from "../test/fixtures/vehicles";
import { createInventoryAPI } from "./apiInventory";
import { request } from "./apiTransport";

describe("vehicle report data requests", () => {
  afterEach(() => vi.unstubAllGlobals());

  it("keeps report order across batches within the JSON request limit", async () => {
    const ids = Array.from({ length: 1001 }, (_, i) => i.toString(16).padStart(32, "0")).reverse();
    const fetchMock = vi.fn<typeof fetch>()
      .mockResolvedValueOnce(Response.json(ids.slice(0, 1000).map((id) => vehicleFixture({ id }))))
      .mockResolvedValueOnce(Response.json([vehicleFixture({ id: ids[1000] })]));
    vi.stubGlobal("fetch", fetchMock);

    const report = await createInventoryAPI(request).vehicleReportData(ids);

    expect(fetchMock.mock.calls.map(([path, init]) => [path, init?.method, init?.body])).toEqual([
      ["/api/v1/vehicles/report-data", "POST", JSON.stringify({ vehicleIds: ids.slice(0, 1000) })],
      ["/api/v1/vehicles/report-data", "POST", JSON.stringify({ vehicleIds: ids.slice(1000) })]
    ]);
    expect(report.map((vehicle) => vehicle.id)).toEqual(ids);
  });
});
