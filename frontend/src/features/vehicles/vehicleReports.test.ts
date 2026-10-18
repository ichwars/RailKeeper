import { describe, expect, it } from "vitest";

import { cvValueFixture, vehicleFixture } from "../../test/fixtures/vehicles";
import { inventoryReportHtml } from "./vehicleReports";

describe("vehicle reports operational fields", () => {
  it("renders a prototype operation section with speed and home base", () => {
    const vehicle = vehicleFixture({ maximumSpeedKmh: 120, homeBase: "Bw Leipzig-West" });
    const html = inventoryReportHtml(
      [vehicle],
      "",
      { key: "inventoryNumber", direction: "asc" },
      { mode: "details", title: "Bestand", includeQRCode: false, includeImages: false },
      {}
    );

    expect(html).toContain("Vorbild &amp; Betrieb");
    expect(html).toContain("120 km/h");
    expect(html).toContain("Bw Leipzig-West");
  });
  it("keeps index pages in the vehicle report", () => {
    const vehicle = vehicleFixture({ cvValues: [
      cvValueFixture({ cvNumber: 257, indexCV31: 16, indexCV32: 3 }),
      cvValueFixture({ id: "other", cvNumber: 257, indexCV31: 16, indexCV32: 4 })
    ] });
    const html = inventoryReportHtml([vehicle], "", { key: "inventoryNumber", direction: "asc" },
      { mode: "details", title: "CV", includeQRCode: false, includeImages: false }, {});
    expect(html).toContain("CV 257 [CV31=16, CV32=3]");
    expect(html).toContain("CV 257 [CV31=16, CV32=4]");
  });

});
