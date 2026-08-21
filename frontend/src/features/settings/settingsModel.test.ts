import { describe, expect, it } from "vitest";

import type { MasterDataEntry } from "../../shared/api";
import { masterDataImage } from "./settingsModel";

describe("masterDataImage", () => {
  it("uses the active function-symbol palette in the settings UI", () => {
    const entry: MasterDataEntry = {
      id: "symbols:light",
      type: "symbols",
      key: "light",
      label: "Licht",
      active: true,
      sortOrder: 10,
      metadata: { imageData: "print-data", activeImageData: "active-data" },
      createdAt: "2026-08-21T00:00:00Z",
      updatedAt: "2026-08-21T00:00:00Z",
    };

    expect(masterDataImage(entry)).toBe("active-data");
  });
});
