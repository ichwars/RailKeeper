import { expect, it } from "vitest";
import { locomotiveImageBackground } from "./locomotiveImagePixels";
import { refineLocomotiveImageMatte, unmixLocomotiveImageBackground } from "./locomotiveImageMatte";

const width = 80, height = 64;
const offset = (x: number, y: number) => (y * width + x) * 4;
function photo() { return new Uint8ClampedArray(width * height * 4).fill(255); }
function paint(data: Uint8ClampedArray, x: number, y: number, colour: number[], alpha = 1) {
  data.set([...colour.map(value => Math.round(value * alpha + 255 * (1 - alpha))), 255], offset(x, y));
}
function foreground(data: Uint8ClampedArray, colour = [40, 90, 50]) {
  for (let y = 10; y < 40; y++) for (let x = 20; x < 60; x++) paint(data, x, y, colour);
}
function separate(data: Uint8ClampedArray) {
  const background = locomotiveImageBackground(data, width, height);
  expect(background).not.toBeNull();
  const matte = refineLocomotiveImageMatte(data, width, height, background!);
  return unmixLocomotiveImageBackground(data, matte);
}

it("recovers mixed foreground edge colours instead of leaving a faded white fringe", () => {
  const data = photo(), colour = [40, 90, 50]; foreground(data, colour);
  for (let y = 10; y < 40; y++) paint(data, 19, y, colour, 0.25);
  const original = data.slice(), result = separate(data), edge = offset(19, 25);
  expect(result[edge + 3]).toBeGreaterThanOrEqual(62);
  expect(result[edge + 3]).toBeLessThanOrEqual(66);
  for (let channel = 0; channel < 3; channel++) {
    expect(Math.abs(result[edge + channel] - colour[channel])).toBeLessThanOrEqual(3);
    const greyComposite = result[edge + channel] * result[edge + 3] / 255 + 182 * (1 - result[edge + 3] / 255);
    expect(Math.abs(greyComposite - (colour[channel] * 0.25 + 182 * 0.75))).toBeLessThan(2);
  }
  expect(data).toEqual(original);
});

it("preserves enclosed white details, body colours and narrow dark handrails", () => {
  const data = photo(); foreground(data);
  for (let y = 15; y < 26; y++) for (let x = 30; x < 41; x++) paint(data, x, y, [255, 255, 255]);
  for (let y = 10; y < 40; y++) paint(data, 10, y, [30, 30, 30]);
  const result = separate(data);
  for (const [x, y] of [[35, 20], [45, 25], [10, 25]]) {
    expect([...result.slice(offset(x, y), offset(x, y) + 4)])
      .toEqual([...data.slice(offset(x, y), offset(x, y) + 4)]);
  }
});

it("keeps a light grey body even when its antialiased edge resembles a soft shadow", () => {
  const data = photo(); foreground(data, [210, 210, 210]);
  for (let y = 10; y < 40; y++) {
    for (const [x, value] of [[17, 250], [18, 244], [19, 235], [20, 224], [21, 214]]) {
      paint(data, x, y, [value, value, value]);
    }
  }
  const result = separate(data);
  expect([...result.slice(offset(35, 25), offset(35, 25) + 4)]).toEqual([210, 210, 210, 255]);
});

it("removes gradual neutral floor shadows connected to the background", () => {
  const data = photo(); foreground(data);
  const shades = [250, 245, 237, 229, 221, 213, 207, 213, 221, 229, 237, 245, 250];
  for (let row = 0; row < shades.length; row++) for (let x = 20; x < 60; x++) {
    paint(data, x, 48 + row, [shades[row], shades[row], shades[row]]);
  }
  const result = separate(data);
  expect(result[offset(40, 54) + 3]).toBe(0);
  expect([...result.slice(offset(40, 25), offset(40, 25) + 4)]).toEqual([40, 90, 50, 255]);
});

it.each([
  { name: "coloured metal rim", rim: [150, 138, 130] },
  { name: "neutral metal rim", rim: [150, 150, 150] }
])(
  "removes a touching floor shadow while retaining the wheel and its $name", ({ rim }) => {
    const data = photo();
    for (let y = 0; y < height; y++) for (let x = 0; x < width; x++) {
      const shade = Math.round(255 - 120 * Math.exp(-((x - 40) ** 2) / 200 - ((y - 55) ** 2) / 12.5));
      paint(data, x, y, [shade, shade, shade]);
    }
    for (let y = 10; y < 31; y++) for (let x = 20; x < 60; x++) paint(data, x, y, [40, 90, 50]);
    for (let y = 28; y <= 52; y++) for (let x = 28; x <= 52; x++) {
      const radius = Math.hypot(x - 40, y - 40);
      if (radius <= 12) paint(data, x, y, radius >= 10 ? rim : [35, 35, 35]);
    }
    const result = separate(data);
    for (const [x, y] of [[40, 40], [40, 51], [29, 40], [51, 40]]) {
      expect([...result.slice(offset(x, y), offset(x, y) + 4)])
        .toEqual([...data.slice(offset(x, y), offset(x, y) + 4)]);
    }
    for (const [x, y] of [[40, 55], [35, 56], [45, 56], [30, 57], [50, 57]]) {
      expect(result[offset(x, y) + 3]).toBe(0);
    }
  }
);

it("unmixes a known compositing vector, retaining source alpha and opaque details", () => {
  const data = new Uint8ClampedArray([201, 211, 206, 255, 70, 90, 80, 128, 255, 255, 255, 255]);
  const matte = new Uint8ClampedArray([255, 255, 255, 191, 255, 255, 255, 0, 255, 255, 255, 255]);
  const result = unmixLocomotiveImageBackground(data, matte);
  expect([...result]).toEqual([40, 80, 60, 64, 70, 90, 80, 128, 0, 0, 0, 0]);
  expect([...data.slice(0, 4)]).toEqual([201, 211, 206, 255]);
});

it("rejects mismatched masks and oversized processing buffers", () => {
  expect(() => unmixLocomotiveImageBackground(new Uint8ClampedArray(4), new Uint8ClampedArray(3))).toThrow("invalid_bitmap");
  expect(() => refineLocomotiveImageMatte(new Uint8ClampedArray(4), 2, 1, new Uint8Array(2))).toThrow("invalid_bitmap");
  expect(() => refineLocomotiveImageMatte(new Uint8ClampedArray(4), 1, 1, new Uint8Array(1))).toThrow("invalid_bitmap");
  const pixels = 1201 * 1200;
  expect(() => refineLocomotiveImageMatte(new Uint8ClampedArray(pixels * 4), 1201, 1200, new Uint8Array(pixels)))
    .toThrow("invalid_bitmap");
});
