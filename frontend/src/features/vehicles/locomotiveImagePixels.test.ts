import { expect, it } from "vitest";
import { locomotiveImageBounds, quantizeECoSImage } from "./locomotiveImagePixels";

it("suggests a tighter crop for a side-view on uniform margins without changing the source", () => {
  const pixels = new Uint8ClampedArray(100 * 100 * 4).fill(255);
  for (let y = 40; y < 60; y++) for (let x = 0; x < 100; x++) pixels[(y * 100 + x) * 4] = 20;
  const original = pixels.slice();
  expect(locomotiveImageBounds(pixels, 100, 100)).toEqual({ x: 0, y: 40, width: 100, height: 20 });
  expect(pixels).toEqual(original);
});

it("keeps ambiguous photographic edges and blank images while respecting transparency", () => {
  const pixels = new Uint8ClampedArray(20 * 20 * 4).fill(255);
  expect(locomotiveImageBounds(pixels, 20, 20)).toEqual({ x: 0, y: 0, width: 20, height: 20 });
  pixels[0] = 0;
  expect(locomotiveImageBounds(pixels, 20, 20)).toEqual({ x: 0, y: 0, width: 20, height: 20 });
  pixels.fill(0);
  expect(locomotiveImageBounds(pixels, 20, 20)).toBeNull();
  pixels[(10 * 20 + 10) * 4 + 3] = 1;
  expect(locomotiveImageBounds(pixels, 20, 20)).toEqual({ x: 10, y: 10, width: 1, height: 1 });
});

it("does not trim a large non-uniform image border merely because its corners agree", () => {
  const pixels = new Uint8ClampedArray(20 * 20 * 4).fill(255);
  for (let x = 1; x < 19; x++) pixels[x * 4] = 0;
  expect(locomotiveImageBounds(pixels, 20, 20)).toEqual({ x: 0, y: 0, width: 20, height: 20 });
});

it("matches independently calculated palette vectors including grey reset, ties and boundaries", () => {
  const pixels = new Uint8ClampedArray(12 * 8 * 4);
  for (let y = 0; y < 8; y++) for (let x = 0; x < 12; x++) {
    let color = [(x * 47 + y * 13 + 17) % 256, (x * 29 + y * 61 + 91) % 256, (x * 83 + y * 7 + 43) % 256];
    if (x === 4 || x === 7 && y === 3) color = [182, 182, 182];
    if (y === 0 && x < 4) color = [[18, 18, 36], [255, 255, 255], [0, 0, 0], [181, 182, 183]][x];
    pixels.set([...color, 255], (y * 12 + x) * 4);
  }
  const original = pixels.slice(), output = quantizeECoSImage(pixels, 12, 8);
  const rows = ["00ff00b6b6fa21658aaccd32", "31567ab9b606494d8fb1f53b", "397d83a4b62e5171b7d81e26",
    "41468aadb63355b6bec5054a", "496eb3d5b63a7d86a6c92e52", "5196b6d9b6476989cff1355b",
    "759aa3c5b64b6db2b7195a62", "6186a8edb65395bade21466b"];
  const rg = [0, 36, 73, 109, 146, 182, 219, 255], blue = [0, 73, 182, 255];
  rows.forEach((row, y) => {
    for (let x = 0; x < 12; x++) {
      const index = parseInt(row.slice(x * 2, x * 2 + 2), 16);
      expect([...output.slice((y * 12 + x) * 4, (y * 12 + x + 1) * 4)])
        .toEqual([rg[index >> 5], rg[(index >> 2) & 7], blue[index & 3], 255]);
    }
  });
  expect(pixels).toEqual(original);
  expect(quantizeECoSImage(output, 12, 8)).toEqual(output);
});

it("rejects unsupported dimensions and transparency rather than silently flattening it", () => {
  expect(() => quantizeECoSImage(new Uint8ClampedArray(4), 1, 1)).toThrow("background");
  expect(() => quantizeECoSImage(new Uint8ClampedArray(4), 191, 1)).toThrow("invalid");
});
