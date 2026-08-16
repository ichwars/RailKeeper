import { describe, expect, it } from "vitest";

import { vehicleFixture } from "../../test/fixtures/vehicles";
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
});
