import { describe, expect, it, vi } from "vitest";

import { emptyVehicle } from "./vehicleViewModel";
import { printQrSvgLabel } from "./vehicleQr";

describe("QR label printing", () => {
  it("escapes vehicle fields before writing printable HTML", () => {
    let written = "";
    const printWindow = {
      document: {
        write: (value: string) => { written = value; },
        close: vi.fn()
      },
      focus: vi.fn(),
      print: vi.fn()
    };
    vi.spyOn(window, "open").mockReturnValue(printWindow as unknown as Window);

    printQrSvgLabel("<svg></svg>", {
      ...emptyVehicle,
      inventoryNumber: "<img src=x onerror=alert(1)>",
      name: "BR & <script>alert(1)</script>",
      digitalDecoderNumber: "\"quoted\""
    });

    expect(written).not.toContain("<script>");
    expect(written).not.toContain("<img src=x");
    expect(written).toContain("&lt;script&gt;");
    expect(written).toContain("&amp;");
    expect(written).toContain("&quot;quoted&quot;");
  });
});
