import { describe, expect, it } from "vitest";

import { adapterOptions } from "./vehicleOptions";

describe("vehicle adapter options", () => {
  it("includes PluX12 in numeric PluX order", () => {
    expect(adapterOptions).toEqual([
      "NEM 651",
      "NEM 652",
      "PluX12",
      "PluX16",
      "PluX22",
      "MTC21",
      "Next18",
      "8-polig",
      "21-polig"
    ]);
  });
});
