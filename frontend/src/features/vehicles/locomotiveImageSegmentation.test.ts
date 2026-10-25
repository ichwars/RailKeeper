import { expect, it } from "vitest";
import { locomotiveImageBackground } from "./locomotiveImagePixels";
import { segmentLocomotiveForeground } from "./locomotiveImageSegmentation";

const width = 48, height = 40;
function photo() { return new Uint8ClampedArray(width * height * 4).fill(255); }
function segment(data: Uint8ClampedArray) {
  const background = locomotiveImageBackground(data, width, height);
  expect(background).not.toBeNull();
  return segmentLocomotiveForeground(data, width, height, background!, [255, 255, 255]);
}

it("finds the same contour when the photograph is mirrored", () => {
  const data = photo(), mirrored = photo();
  for (let y = 27; y < 38; y++) for (let x = 4; x < 44; x++) {
    const value = Math.round(255 - 90 * Math.exp(-((x - 24) ** 2) / 100 - ((y - 31) ** 2) / 8));
    data.set([value, value, value, 255], (y * width + x) * 4);
  }
  for (let y = 8; y < 26; y++) for (let x = 11; x < 35; x++) {
    data.set([60, 90, 65, 255], (y * width + x) * 4);
  }
  for (let y = 0; y < height; y++) for (let x = 0; x < width; x++) {
    const offset = (y * width + x) * 4;
    mirrored.set(data.subarray(offset, offset + 4), (y * width + width - x - 1) * 4);
  }
  const originalMask = segment(data), mirroredMask = segment(mirrored);
  for (let y = 0; y < height; y++) for (let x = 0; x < width; x++) {
    expect(mirroredMask[y * width + width - x - 1]).toBe(originalMask[y * width + x]);
  }
});

it("retains a faint subject if there are no reliable foreground anchors", () => {
  const data = photo();
  for (let y = 10; y < 30; y++) for (let x = 12; x < 36; x++) {
    data.set([232, 232, 232, 255], (y * width + x) * 4);
  }
  const mask = segment(data);
  expect(mask[20 * width + 24]).toBe(1);
  expect(mask[0]).toBe(0);
  expect(mask.reduce((count, value) => count + value, 0)).toBe(24 * 20);
});

it("rejects invalid dimensions, masks and background colours before searching", () => {
  const data = new Uint8ClampedArray(4), mask = new Uint8Array(1);
  expect(() => segmentLocomotiveForeground(data, 0, 1, mask, [255, 255, 255])).toThrow("invalid_bitmap");
  expect(() => segmentLocomotiveForeground(data, 1, 1, new Uint8Array(0), [255, 255, 255]))
    .toThrow("invalid_bitmap");
  expect(() => segmentLocomotiveForeground(data, 1, 1, mask, [255, NaN, 255])).toThrow("invalid_bitmap");
});
