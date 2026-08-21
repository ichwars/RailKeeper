import { describe, expect, it } from "vitest";

import { functionSymbolImageData } from "./functionSymbolImages";

describe("functionSymbolImageData", () => {
  const metadata = {
    imageData: "print",
    activeImageData: "active",
    inactiveImageData: "inactive",
    svgData: "legacy",
  };

  it("selects the requested palette without changing geometry ownership", () => {
    expect(functionSymbolImageData(metadata, "active")).toBe("active");
    expect(functionSymbolImageData(metadata, "inactive")).toBe("inactive");
    expect(functionSymbolImageData(metadata, "print")).toBe("print");
  });

  it("falls back safely for custom and legacy uploads", () => {
    expect(functionSymbolImageData({ imageData: "custom" }, "active")).toBe("custom");
    expect(functionSymbolImageData({ svgData: "legacy" }, "print")).toBe("legacy");
    expect(functionSymbolImageData(undefined, "active")).toBe("");
  });
});
