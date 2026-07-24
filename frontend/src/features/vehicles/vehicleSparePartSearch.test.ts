import { describe, expect, it } from "vitest";

import { buildSparePartImportPlan } from "./vehicleSparePartSearch";

describe("buildSparePartImportPlan", () => {
  it("merges duplicate search results into one complete spare part", () => {
    const plan = buildSparePartImportPlan({ spareParts: [] }, {
      query: "4711",
      results: [
        {
          source: "manufacturer",
          title: "Parts",
          url: "https://example.test/parts",
          snippet: "",
          score: 1,
          fields: {},
          spareParts: [
            { articleNumber: "4711", description: "Motor", price: "", url: "" },
            { articleNumber: "4711", description: "Motor", price: "12,00 EUR", url: "https://example.test/4711" }
          ]
        }
      ]
    });

    expect(Array.from(plan.creates.values())).toEqual([{
      articleNumber: "4711",
      description: "Motor",
      price: "12,00 EUR",
      url: "https://example.test/parts"
    }]);
    expect(plan.updates.size).toBe(0);
  });
});
