import { describe, expect, it } from "vitest";

import { formatAccessoryMoney, normalizeAccessoryMoney } from "./accessoryMoney";

describe("accessory money", () => {
  it("normalizes a backend-valid leading plus sign", () => {
    expect(normalizeAccessoryMoney("+129.90")).toBe("129.90");
    expect(formatAccessoryMoney("+129.90", "de")).toBe("129,90 €");
  });

  it("rejects a plus sign without an amount", () => {
    expect(normalizeAccessoryMoney("+")).toBeUndefined();
  });
});
